package rules

import (
	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG091",
		Section:  sectionLibrarySideEffects,
		Severity: lint.SeverityWarning,
		Doc:      "Change directory in a library function only inside `( ... )`, or put the old one back",
		Check:    checkLibraryCd,
	})
}

func checkLibraryCd(f *File, r *Reporter) {
	if f.Role != RoleLibrary {
		return
	}
	up := parents(f.Syntax)
	eachFunc(f, func(decl *syntax.FuncDecl) {
		var moves []*syntax.CallExpr
		restores := false
		saved := pwdVars(decl)
		allCalls(decl.Body, func(call *syntax.CallExpr, name string) {
			if inSubshell(call, up) {
				return
			}
			switch name {
			case "popd":
				restores = true
			case "cd", "pushd":
				if len(call.Args) > 1 {
					arg := call.Args[len(call.Args)-1]
					if wordLiteral(arg) == "-" {
						restores = true
						return
					}
					if pe := wordParam(arg); pe != nil && pe.Param != nil && (saved[pe.Param.Value] || pe.Param.Value == "OLDPWD") {
						restores = true
						return
					}
				}
				moves = append(moves, call)
			}
		})
		if restores {
			return
		}
		for _, call := range moves {
			r.At(call.Pos(), "`%s` in a library function moves the whole script that sourced it; run it inside `( ... )`, or save `$PWD` and `cd` back",
				callName(call))
		}
	})
}

// pwdVars returns the variables a function fills with the current directory.
func pwdVars(decl *syntax.FuncDecl) map[string]bool {
	out := map[string]bool{}
	syntax.Walk(decl.Body, func(node syntax.Node) bool {
		assign, ok := node.(*syntax.Assign)
		if !ok || assign.Name == nil || assign.Value == nil {
			return true
		}
		if src := wordSource(assign.Value); src == "$PWD" || src == `"$PWD"` || src == "${PWD}" || src == `"${PWD}"` || src == "$(pwd)" || src == `"$(pwd)"` {
			out[assign.Name.Value] = true
		}
		return true
	})
	return out
}
