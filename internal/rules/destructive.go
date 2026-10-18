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
	check := func(root syntax.Node) {
		guarded := guardedVars(root)
		allCalls(root, func(call *syntax.CallExpr, name string) {
			if !destroysRecursively(call, name) {
				return
			}
			for _, arg := range call.Args[1:] {
				if strings.HasPrefix(wordLiteral(arg), "-") {
					continue
				}
				if v := unguardedVar(arg, guarded); v != "" {
					reportWithDybatpho(f, r, arg.Pos(), "`dybatpho::safe_rm -r` refuses an empty or protected path, and one outside `DYBATPHO_SAFE_ROOTS`", "`%s` runs recursively on a path built from %q, which nothing checks; when it is empty the command reaches `/` or the current directory. Test it with `[[ -n \"${%s}\" ]]`, or write `${%s:?}`",
						name, v, v, v)
					return
				}
			}
		})
	}
	eachFunc(f, func(decl *syntax.FuncDecl) { check(decl.Body) })
	// A script deletes at its top level as often as inside a function;
	// allCalls leaves the function bodies out there, so nothing is reported
	// twice.
	check(f.Syntax)
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
		// `${dir:-build}` and `${dir:=build}` fall back to their text when
		// the variable is empty, so they never expand to nothing. The forms
		// without a colon do not count: a variable set to "" stays empty.
		if pe.Exp != nil && (pe.Exp.Op == syntax.DefaultUnsetOrNull || pe.Exp.Op == syntax.AssignUnsetOrNull) &&
			pe.Exp.Word != nil && hasLiteralText(pe.Exp.Word) {
			return false
		}
		found = pe.Param.Value
		return false
	})
	return found
}

// guardedVars returns the variables a function body, or the top level of a
// script, checks before using: any it tests with `[[ ]]`, any a
// temporary-file helper fills, and any it passes to a path-safety check. A
// `:?` guards only the expansion that carries it. The bodies of functions
// declared under root are a scope of their own and are left out.
func guardedVars(root syntax.Node) map[string]bool {
	out := map[string]bool{}
	walkScope(root, true, func(node syntax.Node) {
		switch n := node.(type) {
		case *syntax.TestClause:
			syntax.Walk(n, func(inner syntax.Node) bool {
				if pe, ok := inner.(*syntax.ParamExp); ok && pe.Param != nil {
					out[pe.Param.Value] = true
				}
				return true
			})
		case *syntax.Assign:
			// A value with literal text in it, such as `"${dir}/.partial"`, is
			// never empty, and neither is a fresh temporary path.
			if n.Name != nil && n.Value != nil && (tempMaker.MatchString(wordSource(n.Value)) || hasLiteralText(n.Value)) {
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
	})
	return out
}

// hasLiteralText reports whether a word has text of its own besides its
// expansions, which keeps its value from ever being empty.
func hasLiteralText(word *syntax.Word) bool {
	found := false
	syntax.Walk(word, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.Lit:
			if n.Value != "" {
				found = true
			}
		case *syntax.SglQuoted:
			if n.Value != "" {
				found = true
			}
		case *syntax.ParamExp, *syntax.CmdSubst:
			return false
		}
		return !found
	})
	return found
}
