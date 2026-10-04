package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionDestructive = "Script Stabilization > Destructive Commands"

func init() {
	register(Rule{
		Code:     "BSG089",
		Section:  sectionDestructive,
		Severity: lint.SeverityWarning,
		Doc:      "Check that a variable is not empty before a recursive `rm`, `chmod`, `chown` or `find -delete` uses it",
		Check:    checkUnguardedDestruction,
	})
}

// tempMaker matches the commands that fill a variable with a fresh temporary
// path, which is never empty.
var tempMaker = regexp.MustCompile(`mktemp|create_temp`)

func checkUnguardedDestruction(f *File, r *Reporter) {
	eachFunc(f, func(decl *syntax.FuncDecl) {
		guarded := guardedVars(decl)
		allCalls(decl.Body, func(call *syntax.CallExpr, name string) {
			if !destroysRecursively(call, name) {
				return
			}
			for _, arg := range call.Args[1:] {
				if strings.HasPrefix(wordLiteral(arg), "-") {
					continue
				}
				if v := unguardedVar(arg, guarded); v != "" {
					r.At(arg.Pos(), "`%s` runs recursively on a path built from %q, which nothing checks; when it is empty the command reaches `/` or the current directory. Test it with `[[ -n \"${%s}\" ]]`, or write `${%s:?}`",
						name, v, v, v)
					return
				}
			}
		})
	})
}

func destroysRecursively(call *syntax.CallExpr, name string) bool {
	switch name {
	case "rm":
		return hasFlag(call, 'r') || hasFlag(call, 'R') || hasLong(call, "--recursive")
	case "chmod", "chown", "chgrp":
		return hasFlag(call, 'R') || hasLong(call, "--recursive")
	case "find":
		return hasLong(call, "-delete")
	}
	return false
}

func hasLong(call *syntax.CallExpr, option string) bool {
	for _, arg := range call.Args[1:] {
		if wordLiteral(arg) == option {
			return true
		}
	}
	return false
}

// unguardedVar returns the first variable a word expands without `:?` that
// the function never checks.
func unguardedVar(word *syntax.Word, guarded map[string]bool) string {
	found := ""
	syntax.Walk(word, func(node syntax.Node) bool {
		pe, ok := node.(*syntax.ParamExp)
		if !ok || found != "" {
			return found == ""
		}
		if pe.Param == nil || guarded[pe.Param.Value] {
			return false
		}
		if pe.Exp != nil && (pe.Exp.Op == syntax.ErrorUnset || pe.Exp.Op == syntax.ErrorUnsetOrNull) {
			return false
		}
		found = pe.Param.Value
		return false
	})
	return found
}

// guardedVars returns the variables a function checks before using: any it
// tests with `[[ ]]`, any a temporary-file helper fills, and any it passes
// to a path-safety check. A `:?` guards only the expansion that carries it.
func guardedVars(decl *syntax.FuncDecl) map[string]bool {
	out := map[string]bool{}
	syntax.Walk(decl.Body, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.TestClause:
			syntax.Walk(n, func(inner syntax.Node) bool {
				if pe, ok := inner.(*syntax.ParamExp); ok && pe.Param != nil {
					out[pe.Param.Value] = true
				}
				return true
			})
		case *syntax.Assign:
			if n.Name != nil && n.Value != nil && tempMaker.MatchString(wordSource(n.Value)) {
				out[n.Name.Value] = true
			}
		case *syntax.CallExpr:
			name := callName(n)
			if tempMaker.MatchString(name) || strings.Contains(name, "safe_path") || strings.Contains(name, "assert") {
				for _, arg := range n.Args[1:] {
					if lit := wordLiteral(arg); isVarName(lit) {
						out[lit] = true
					}
					if pe := wordParam(arg); pe != nil && pe.Param != nil {
						out[pe.Param.Value] = true
					}
				}
			}
		}
		return true
	})
	return out
}
