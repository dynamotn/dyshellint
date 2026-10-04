package rules

import (
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

// readsExternalInput reports whether a loop reads a file, a process
// substitution or a pipe, any of which may end without a newline. A here
// document and a here string always end with one.
func readsExternalInput(stmt *syntax.Stmt, up map[syntax.Node]syntax.Node) bool {
	for _, redir := range stmt.Redirs {
		if redir.Op == syntax.RdrIn {
			return true
		}
	}
	if bin, ok := up[stmt].(*syntax.BinaryCmd); ok && (bin.Op == syntax.Pipe || bin.Op == syntax.PipeAll) {
		return bin.Y == stmt
	}
	return false
}
