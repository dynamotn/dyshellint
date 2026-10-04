package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG094",
		Section:  sectionPipeWhile,
		Severity: lint.SeverityWarning,
		Doc:      "Keep the last line of input that has no newline: `while read -r line || [[ -n ${line} ]]`",
		Check:    checkReadLoopLastLine,
	})
}

func checkReadLoopLastLine(f *File, r *Reporter) {
	up := parents(f.Syntax)
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		loop, ok := node.(*syntax.WhileClause)
		if !ok || loop.Until || len(loop.Cond) != 1 {
			return true
		}
		call, ok := loop.Cond[0].Cmd.(*syntax.CallExpr)
		if !ok || callName(call) != "read" || hasFlag(call, 'd') {
			return true
		}
		stmt, ok := up[loop].(*syntax.Stmt)
		if !ok || !readsExternalInput(stmt, up) {
			return true
		}
		target := "REPLY"
		builtinTargets(call, func(name string, _ syntax.Pos) { target = name })
		r.At(call.Pos(), "`read` fails on a last line that has no newline, so the loop drops it; write `while read -r ... || [[ -n \"${%s}\" ]]`",
			target)
		return true
	})
}

// readsExternalInput reports whether a loop reads input that may end without
// a newline: a file, the output of `cat`, or a `printf` whose format does not
// end in one. A here document, a here string and the output of the usual
// line-printing tools always end with a newline.
func readsExternalInput(stmt *syntax.Stmt, up map[syntax.Node]syntax.Node) bool {
	for _, redir := range stmt.Redirs {
		if redir.Op != syntax.RdrIn || redir.Word == nil {
			continue
		}
		if len(redir.Word.Parts) == 1 {
			if ps, ok := redir.Word.Parts[0].(*syntax.ProcSubst); ok {
				for _, inner := range ps.Stmts {
					if mayLackNewline(firstCall(inner)) {
						return true
					}
				}
				return false
			}
		}
		return true
	}
	if bin, ok := up[stmt].(*syntax.BinaryCmd); ok && (bin.Op == syntax.Pipe || bin.Op == syntax.PipeAll) && bin.Y == stmt {
		return mayLackNewline(lastCall(bin.X))
	}
	return false
}

// mayLackNewline reports whether a command passes on data that may not end in
// a newline: `cat` of a file, or `printf` with a format that does not.
func mayLackNewline(call *syntax.CallExpr) bool {
	switch callName(call) {
	case "cat":
		return true
	case "printf":
		for _, arg := range call.Args[1:] {
			lit := wordLiteral(arg)
			if lit == "-v" {
				return false
			}
			if strings.HasPrefix(lit, "-") {
				continue
			}
			return !strings.HasSuffix(wordSource(arg), `\n'`) && !strings.HasSuffix(wordSource(arg), `\n"`)
		}
	}
	return false
}

// lastCall returns the last simple command of a pipeline.
func lastCall(stmt *syntax.Stmt) *syntax.CallExpr {
	calls := leadingCalls(stmt)
	if len(calls) == 0 {
		return nil
	}
	return calls[len(calls)-1]
}
