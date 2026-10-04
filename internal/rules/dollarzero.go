package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG039",
		Section:  sectionCommonScripts,
		Severity: lint.SeverityError,
		Doc:      "Find a library's own path with `${BASH_SOURCE[0]}`, not `$0`",
		Check:    checkDollarZeroInLibrary,
	})
}

func checkDollarZeroInLibrary(f *File, r *Reporter) {
	if f.Role != RoleLibrary {
		return
	}
	up := parents(f.Syntax)
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		pe, ok := node.(*syntax.ParamExp)
		if !ok || pe.Param == nil || pe.Param.Value != "0" || isMainGuard(pe, up, f) {
			return true
		}
		r.At(pe.Pos(), "a sourced file's `$0` is the script that sourced it, not this file; use `${BASH_SOURCE[0]}` to find the library's own path")
		return true
	})
}

// isMainGuard reports whether `$0` sits in the test a file uses to run only
// when it is executed: `[[ "${BASH_SOURCE[0]}" == "$0" ]]`.
func isMainGuard(pe *syntax.ParamExp, up map[syntax.Node]syntax.Node, f *File) bool {
	for n := up[pe]; n != nil; n = up[n] {
		if test, ok := n.(*syntax.TestClause); ok {
			return strings.Contains(f.Text(test.Pos(), test.End()), "BASH_SOURCE")
		}
		if _, ok := n.(*syntax.Stmt); ok {
			return false
		}
	}
	return false
}
