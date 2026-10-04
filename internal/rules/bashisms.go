package rules

import (
	"regexp"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(
		Rule{
			Code:     "BSG124",
			Section:  sectionForLoops,
			Severity: lint.SeverityWarning,
			Doc:      "Generate a sequence with `{1..5}` or `for ((i = 0; i < n; i++))`, not `seq`",
			Check:    checkSeq,
		},
		Rule{
			Code:     "BSG125",
			Section:  sectionWildcards,
			Severity: lint.SeverityError,
			Doc:      "Do not parse the output of `ls`: loop over a glob such as `./*`",
			Check:    checkParsedLs,
		},
		Rule{
			Code:     "BSG126",
			Section:  sectionBuiltins,
			Severity: lint.SeverityWarning,
			Doc:      "Split a string into fields with `IFS=: read -r a b _ <<<` or parameter expansion, not `echo | cut`",
			Check:    checkFieldSplit,
		},
	)
}

// checkSeq reports a call to `seq`, an external process for what brace
// expansion and an arithmetic `for` do in the shell.
func checkSeq(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		if call, ok := node.(*syntax.CallExpr); ok && callName(call) == "seq" {
			r.At(call.Pos(), "`seq` starts a process for a sequence the shell can make; write `{1..5}` for a fixed range, or `for ((i = start; i <= end; i++))` when a bound is a variable")
		}
		return true
	})
}

// checkParsedLs reports `ls` whose output is read by the script: inside a
// command substitution, or on the left of a pipe. `ls` that only shows a
// listing to the user is left alone.
func checkParsedLs(f *File, r *Reporter) {
	const msg = "the output of `ls` is text for people: a name with a space or a newline splits, and a glob character expands; loop over `./*` and test each match, or use `find -print0`"
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.CmdSubst:
			for _, stmt := range n.Stmts {
				if call := firstCall(stmt); callName(call) == "ls" {
					r.At(call.Pos(), msg)
				}
			}
		case *syntax.BinaryCmd:
			if n.Op != syntax.Pipe && n.Op != syntax.PipeAll {
				return true
			}
			// Only the head of a pipeline: `a | b | c` holds `a | b` on its left.
			if inner, ok := n.X.Cmd.(*syntax.BinaryCmd); ok && (inner.Op == syntax.Pipe || inner.Op == syntax.PipeAll) {
				return true
			}
			if call, ok := n.X.Cmd.(*syntax.CallExpr); ok && callName(call) == "ls" {
				r.At(call.Pos(), msg)
			}
		}
		return true
	})
}

// awkField matches an awk program that only prints one field, `{print $2}`.
var awkField = regexp.MustCompile(`^\s*\{\s*print\s+\$[0-9]+\s*;?\s*\}\s*$`)

// checkFieldSplit reports a string echoed into `cut`, or into an awk program
// that prints one field, and `cut` fed by a here-string: `read` splits it
// into named fields in the shell.
func checkFieldSplit(f *File, r *Reporter) {
	const msg = "a process to split a string into fields; `IFS=%s read -r first second _ <<< \"${value}\"` splits it in the shell, or `${value%%%%%s*}` keeps the first field"
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.BinaryCmd:
			if n.Op != syntax.Pipe {
				return true
			}
			left, ok := n.X.Cmd.(*syntax.CallExpr)
			if !ok || (callName(left) != "echo" && callName(left) != "printf") {
				return true
			}
			right, ok := n.Y.Cmd.(*syntax.CallExpr)
			if !ok {
				return true
			}
			if sep, ok := splitter(right); ok {
				r.At(n.OpPos, msg, sep, sep)
			}
		case *syntax.Stmt:
			call, ok := n.Cmd.(*syntax.CallExpr)
			if !ok || callName(call) != "cut" {
				return true
			}
			sep, ok := splitter(call)
			if !ok {
				return true
			}
			for _, redir := range n.Redirs {
				if redir.Op == syntax.WordHdoc {
					r.At(redir.OpPos, msg, sep, sep)
				}
			}
		}
		return true
	})
}

// splitter reports whether a call only cuts fields out of its input, `cut -d
// -f` or `awk -F` with a program that prints one field, and returns the
// separator it splits on. `cut -c` and `cut -b` take characters, not fields,
// and are left alone.
func splitter(call *syntax.CallExpr) (string, bool) {
	sep := " "
	switch callName(call) {
	case "cut":
		delimited, fields := false, false
		for i, arg := range call.Args[1:] {
			lit := wordLiteral(arg)
			switch {
			case len(lit) > 2 && lit[:2] == "-d":
				sep, delimited = lit[2:], true
			case lit == "-d" && i+2 < len(call.Args):
				sep, delimited = wordLiteral(call.Args[i+2]), true
			case len(lit) >= 2 && lit[:2] == "-f":
				fields = true
			}
		}
		return sep, delimited && fields
	case "awk":
		matched := false
		for i, arg := range call.Args[1:] {
			lit := wordLiteral(arg)
			switch {
			case len(lit) > 2 && lit[:2] == "-F":
				sep = lit[2:]
			case lit == "-F" && i+2 < len(call.Args):
				sep = wordLiteral(call.Args[i+2])
			case awkField.MatchString(lit):
				matched = true
			}
		}
		return sep, matched
	}
	return "", false
}
