package rules

import (
	"strconv"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionFileDescriptors = "Features and Bugs > File Descriptors"

func init() {
	register(
		Rule{
			Code:     "BSG152",
			Section:  sectionFileDescriptors,
			Severity: lint.SeverityWarning,
			Doc:      "Let Bash pick the descriptor with `exec {fd}> \"${file}\"`, instead of a fixed number such as `exec 3>`",
			Check:    checkFixedDescriptor,
		},
		Rule{
			Code:     "BSG153",
			Section:  sectionFileDescriptors,
			Severity: lint.SeverityWarning,
			Doc:      "Redirect a loop once, `done >> \"${out}\"`, instead of appending to the same file on every iteration",
			Check:    checkAppendInLoop,
		},
	)
}

// checkFixedDescriptor reports `exec` opening or copying a descriptor of 3 or
// more by number. Any other part of the script, or a library it sources, may
// already use that number, and the second `exec 3>` silently closes the first.
// `exec 3>&-`, closing it, is left alone: the opening is what gets reported.
// A fixed number on another command, as in the swap `3>&1 1>&2 2>&3`, only
// lasts for that command and is fine.
func checkFixedDescriptor(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		stmt, ok := node.(*syntax.Stmt)
		if !ok || len(stmt.Redirs) == 0 {
			return true
		}
		call, ok := stmt.Cmd.(*syntax.CallExpr)
		if !ok || len(call.Args) != 1 || callName(call) != "exec" {
			return true
		}
		for _, redir := range stmt.Redirs {
			if redir.N == nil || wordLiteral(redir.Word) == "-" {
				continue
			}
			if n, err := strconv.Atoi(redir.N.Value); err != nil || n < 3 {
				continue
			}
			r.At(redir.Pos(), "a fixed descriptor %s may already be open elsewhere in the script or a library it sources, and `exec` silently replaces it; write `exec {fd}>...` and let Bash pick a free one (Bash 4.1), then close it with `exec {fd}>&-`", redir.N.Value)
		}
		return true
	})
}

// varyingSpecials are the variables whose value changes on its own between
// two iterations of a loop.
var varyingSpecials = map[string]bool{
	"RANDOM": true, "SRANDOM": true, "SECONDS": true, "EPOCHSECONDS": true,
	"EPOCHREALTIME": true, "LINENO": true, "BASHPID": true,
}

// checkAppendInLoop reports a `>>` in the body of a loop whose target is the
// same on every iteration: it names no variable the loop changes. The file is
// opened and closed again for each line; redirecting the whole loop opens it
// once. A redirect to `/dev/...` costs no file and is left alone. Each loop is
// reported once, at its first such redirect.
func checkAppendInLoop(f *File, r *Reporter) {
	up := f.parents()
	reported := map[syntax.Node]bool{}
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		stmt, ok := node.(*syntax.Stmt)
		if !ok {
			return true
		}
		for _, redir := range stmt.Redirs {
			if redir.Op != syntax.AppOut && redir.Op != syntax.AppAll || redir.N != nil {
				continue
			}
			loop := enclosingLoop(stmt, up)
			if loop == nil || reported[loop] {
				continue
			}
			if !fixedTarget(redir.Word, assignedIn(loop)) || usedTwice(f, loop, redir.Word) {
				continue
			}
			reported[loop] = true
			r.At(redir.Pos(), "this `>>` opens the same file again on every iteration of the loop; redirect the loop once, `done >> %s`, or the lines that write to it as a group, `{ ...; } >> %s`",
				wordSource(redir.Word), wordSource(redir.Word))
		}
		return true
	})
}

// enclosingLoop returns the nearest for, while or until loop whose body holds
// a statement, or nil when there is none before a function boundary. A
// statement in the condition of a `while` is not in its body.
func enclosingLoop(stmt *syntax.Stmt, up map[syntax.Node]syntax.Node) syntax.Node {
	for n := up[stmt]; n != nil; n = up[n] {
		switch loop := n.(type) {
		case *syntax.FuncDecl:
			return nil
		case *syntax.ForClause:
			if stmt.Pos().After(loop.DoPos) {
				return loop
			}
		case *syntax.WhileClause:
			if stmt.Pos().After(loop.DoPos) {
				return loop
			}
		}
	}
	return nil
}

// usedTwice reports whether a loop names the target of a redirect other than
// to append to it, as when it also reads the file back: each write then has to
// land before the next read, and the loop is better left as it is.
func usedTwice(f *File, loop syntax.Node, target *syntax.Word) bool {
	src := wordSource(target)
	appends := 0
	syntax.Walk(loop, func(node syntax.Node) bool {
		if redir, ok := node.(*syntax.Redirect); ok && (redir.Op == syntax.AppOut || redir.Op == syntax.AppAll) &&
			redir.Word != nil && wordSource(redir.Word) == src {
			appends++
		}
		return true
	})
	return strings.Count(f.Text(loop.Pos(), loop.End()), src) > appends
}

// fixedTarget reports whether a redirect target names a regular file that is
// the same on every iteration: it expands no command and no variable the loop
// assigns, no positional parameter and no variable that changes on its own.
func fixedTarget(word *syntax.Word, assigned map[string]bool) bool {
	if word == nil {
		return false
	}
	if lit := wordLiteral(word); len(lit) >= 5 && lit[:5] == "/dev/" {
		return false
	}
	fixed := true
	syntax.Walk(word, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.CmdSubst, *syntax.ArithmExp, *syntax.ProcSubst:
			fixed = false
		case *syntax.ParamExp:
			if n.Param == nil || assigned[n.Param.Value] || varyingSpecials[n.Param.Value] || isPositionalParam(n.Param.Value) {
				fixed = false
			}
		}
		return fixed
	})
	return fixed
}

// assignedIn returns every name a loop may change: its own variable, the
// assignments in it, the variables `read`, `mapfile` and `printf -v` write,
// and every name inside arithmetic. It errs towards more names, which only
// keeps the rule quiet.
func assignedIn(loop syntax.Node) map[string]bool {
	names := map[string]bool{}
	syntax.Walk(loop, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.WordIter:
			names[n.Name.Value] = true
		case *syntax.Assign:
			if n.Name != nil {
				names[n.Name.Value] = true
			}
		case *syntax.ArithmCmd, *syntax.ArithmExp, *syntax.CStyleLoop, *syntax.LetClause:
			syntax.Walk(n, func(inner syntax.Node) bool {
				if w, ok := inner.(*syntax.Word); ok {
					names[wordLiteral(w)] = true
				}
				if pe, ok := inner.(*syntax.ParamExp); ok && pe.Param != nil {
					names[pe.Param.Value] = true
				}
				return true
			})
		case *syntax.CallExpr:
			switch callName(n) {
			case "read", "mapfile", "readarray", "printf", "getopts":
				for _, arg := range n.Args[1:] {
					names[wordLiteral(arg)] = true
				}
			}
		}
		return true
	})
	return names
}
