package rules

import (
	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG096",
		Section:  sectionCommonScripts,
		Severity: lint.SeverityWarning,
		Doc:      "Check that a computed library path exists before an entrypoint sources it",
		Check:    checkUncheckedSource,
	})
}

func checkUncheckedSource(f *File, r *Reporter) {
	if f.Role != RoleEntrypoint {
		return
	}
	tested := fileTestedVars(f)
	up := parents(f.Syntax)
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		stmt, ok := node.(*syntax.Stmt)
		if !ok {
			return true
		}
		call, ok := stmt.Cmd.(*syntax.CallExpr)
		if !ok || len(call.Args) < 2 {
			return true
		}
		if name := callName(call); name != "." && name != "source" {
			return true
		}
		path := call.Args[1]
		if !hasExpansion(path) || statusUsed(stmt, up) || anyVarIn(path, tested) {
			return true
		}
		r.At(path.Pos(), "%s may not exist, and a failed `.` only prints an error before the script runs on without the library; test it with `[[ -r ... ]]` first, or add `|| exit 1`",
			wordSource(path))
		return true
	})
}

func hasExpansion(word *syntax.Word) bool {
	found := false
	syntax.Walk(word, func(node syntax.Node) bool {
		switch node.(type) {
		case *syntax.ParamExp, *syntax.CmdSubst:
			found = true
		}
		return !found
	})
	return found
}

// fileTestedVars returns the variables a file tests with a file operator:
// `-e`, `-f`, `-r` or `-s`.
func fileTestedVars(f *File) map[string]bool {
	out := map[string]bool{}
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		test, ok := node.(*syntax.UnaryTest)
		if !ok {
			return true
		}
		switch test.Op {
		case syntax.TsExists, syntax.TsRegFile, syntax.TsRead, syntax.TsNoEmpty:
			syntax.Walk(test.X, func(inner syntax.Node) bool {
				if pe, ok := inner.(*syntax.ParamExp); ok && pe.Param != nil {
					out[pe.Param.Value] = true
				}
				return true
			})
		}
		return true
	})
	return out
}

func anyVarIn(word *syntax.Word, vars map[string]bool) bool {
	found := false
	syntax.Walk(word, func(node syntax.Node) bool {
		if pe, ok := node.(*syntax.ParamExp); ok && pe.Param != nil && vars[pe.Param.Value] {
			found = true
		}
		return !found
	})
	return found
}
