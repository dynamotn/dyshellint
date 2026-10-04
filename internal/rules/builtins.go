package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const (
	sectionBuiltins   = "Calling Commands > Builtin Commands vs External Commands"
	sectionShellCheck = "Features and Bugs > Use ShellCheck"
)

func init() {
	register(
		Rule{
			Code:     "BSG108",
			Section:  sectionBuiltins,
			Severity: lint.SeverityWarning,
			Doc:      "Read a whole file with `$(< file)`, not `$(cat file)`",
			Check:    checkCatSubstitution,
		},
		Rule{
			Code:     "BSG126",
			Section:  sectionBuiltins,
			Severity: lint.SeverityWarning,
			Doc:      "Split a string into fields with `IFS=: read -r a b _ <<<` or parameter expansion, not `echo | cut`",
			Check:    checkFieldSplit,
		},
		Rule{
			Code:     "BSG109",
			Section:  sectionShellCheck,
			Severity: lint.SeverityWarning,
			Doc:      "Point ShellCheck at a sourced file with `# shellcheck source=`, rather than disabling SC1091",
			Check:    checkSourceDisabled,
		},
	)
}

func checkCatSubstitution(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		cs, ok := node.(*syntax.CmdSubst)
		if !ok || len(cs.Stmts) != 1 {
			return true
		}
		stmt := cs.Stmts[0]
		call, ok := stmt.Cmd.(*syntax.CallExpr)
		if !ok || stmt.Negated || stmt.Background || len(stmt.Redirs) > 0 || len(call.Assigns) > 0 ||
			callName(call) != "cat" || len(call.Args) != 2 || strings.HasPrefix(wordSource(call.Args[1]), "-") {
			return true
		}
		r.At(cs.Pos(), "`$(cat %s)` starts a process to read a file; `$(< %s)` does the same in the shell", wordSource(call.Args[1]), wordSource(call.Args[1]))
		return true
	})
}

// awkField matches an awk program that only prints one field, `{print $2}`.
var awkField = regexp.MustCompile(`^\s*\{\s*print\s+\$[0-9]+\s*;?\s*\}\s*$`)

// checkFieldSplit reports a string echoed into `cut`, or into an awk program
// that prints one field, and `cut` fed by a here-string: `read` splits it
// into named fields in the shell. A script that loads dybatpho is also pointed
// at `dybatpho::split`, which takes a delimiter of several characters as it is.
func checkFieldSplit(f *File, r *Reporter) {
	msg := withDybatpho(f,
		"a process to split a string into fields; `IFS=%s read -r first second _ <<< \"${value}\"` splits it in the shell, or `${value%%%%%s*}` keeps the first field",
		"`mapfile -t fields < <(dybatpho::split \"${value}\" '<delimiter>')` splits on a delimiter of several characters and keeps empty fields")
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
			if sep, ok := fieldSplitter(right); ok {
				r.At(n.OpPos, msg, sep, sep)
			}
		case *syntax.Stmt:
			call, ok := n.Cmd.(*syntax.CallExpr)
			if !ok || callName(call) != "cut" {
				return true
			}
			sep, ok := fieldSplitter(call)
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

// fieldSplitter reports whether a call only cuts fields out of its input,
// `cut -d -f` or `awk -F` with a program that prints one field, and returns
// the separator it splits on. `cut -c` and `cut -b` take characters, not
// fields, and are left alone.
func fieldSplitter(call *syntax.CallExpr) (string, bool) {
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

// sourceDisabled matches a ShellCheck directive that disables SC1091.
var sourceDisabled = regexp.MustCompile(`^\s*#\s*shellcheck\s+disable=[^#]*\bSC1091\b`)

func checkSourceDisabled(f *File, r *Reporter) {
	for i, line := range f.Lines {
		if sourceDisabled.MatchString(line) {
			r.AtLine(i+1, "disabling SC1091 hides every sourced file from ShellCheck; name the file with `# shellcheck source=<path>`, relative to the script with `source-path=SCRIPTDIR`")
		}
	}
}
