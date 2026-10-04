package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG038",
		Section:  sectionCommonScripts,
		Severity: lint.SeverityError,
		Doc:      "Guard a library against being sourced twice before it declares a `readonly`",
		Check:    checkUnguardedReadonly,
	})
}

func checkUnguardedReadonly(f *File, r *Reporter) {
	if f.Role != RoleLibrary {
		return
	}
	guarded := false
	for _, stmt := range f.Syntax.Stmts {
		if !guarded && returnsEarly(stmt) {
			guarded = true
			continue
		}
		clause, ok := stmt.Cmd.(*syntax.DeclClause)
		if !ok || clause.Variant == nil || guarded {
			continue
		}
		switch clause.Variant.Value {
		case "readonly":
		case "declare", "typeset":
			if !strings.Contains(declFlags(clause), "r") {
				continue
			}
		default:
			continue
		}
		r.At(stmt.Pos(), "sourcing this library a second time fails on its `readonly`; return early when it is already loaded, as in `[[ -n \"${_LIB_LOADED-}\" ]] && return 0`")
	}
}

// returnsEarly reports whether a top-level statement can end the sourcing of
// a file: it holds a `return`.
func returnsEarly(stmt *syntax.Stmt) bool {
	found := false
	syntax.Walk(stmt, func(node syntax.Node) bool {
		if _, ok := node.(*syntax.FuncDecl); ok {
			return false
		}
		if call, ok := node.(*syntax.CallExpr); ok && callName(call) == "return" {
			found = true
		}
		return !found
	})
	return found
}
