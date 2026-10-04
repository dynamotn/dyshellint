package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionArrays = "Features and Bugs > Arrays"

func init() {
	register(Rule{
		Code:     "BSG049",
		Section:  sectionArrays,
		Severity: lint.SeverityWarning,
		Doc:      "Guard the expansion of an array that may be empty with `${a[@]+\"${a[@]}\"}`",
		Check:    checkEmptyArrayExpansion,
	})
}

// checkEmptyArrayExpansion reports `"${a[@]}"` of a local array that may still
// be empty where it is expanded. Bash before 4.4 treats that expansion as an
// unset variable, so a `set -u` script stops on it; dybatpho supports 4.3.
func checkEmptyArrayExpansion(f *File, r *Reporter) {
	if !f.UsesDybatpho {
		return
	}
	eachFunc(f, func(decl *syntax.FuncDecl) {
		empty := emptyArrays(decl)
		if len(empty) == 0 {
			return
		}
		filled := filledAt(decl, empty)
		syntax.Walk(decl.Body, func(node syntax.Node) bool {
			switch n := node.(type) {
			case *syntax.FuncDecl:
				return false
			case *syntax.ParamExp:
				// `${a[@]+"${a[@]}"}` is the guarded form: the inner expansion
				// only runs when the array is set, so nothing below it counts.
				if n.Exp != nil {
					return false
				}
				name := arrayExpansion(n)
				if name == "" || !empty[name] {
					return true
				}
				if at, ok := filled[name]; ok && n.Pos().After(at) {
					return true
				}
				r.At(n.Pos(), "%q may be empty here, and Bash 4.3 stops a `set -u` script on `${%s[@]}` of an empty array; write `${%s[@]+\"${%s[@]}\"}`",
					name, name, name, name)
			}
			return true
		})
	})
}

// arrayExpansion returns the array an expansion lists whole, as `${a[@]}` or
// `${a[*]}`, and the empty string for anything else, such as `${#a[@]}`.
func arrayExpansion(pe *syntax.ParamExp) string {
	if pe.Param == nil || pe.Length || pe.Excl || pe.Slice != nil || pe.Repl != nil || pe.Index == nil {
		return ""
	}
	word, ok := pe.Index.(*syntax.Word)
	if !ok {
		return ""
	}
	switch wordLiteral(word) {
	case "@", "*":
		return pe.Param.Value
	}
	return ""
}

// emptyArrays returns the local arrays a function declares without elements.
func emptyArrays(decl *syntax.FuncDecl) map[string]bool {
	out := map[string]bool{}
	syntax.Walk(decl.Body, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.FuncDecl:
			return false
		case *syntax.DeclClause:
			if n.Variant == nil || n.Variant.Value != "local" && n.Variant.Value != "declare" {
				return true
			}
			isArray := strings.ContainsAny(declFlags(n), "aA")
			for _, arg := range n.Args {
				if arg.Name == nil {
					continue
				}
				switch {
				case arg.Array != nil:
					if len(arg.Array.Elems) == 0 {
						out[arg.Name.Value] = true
					}
				case isArray && arg.Value == nil:
					out[arg.Name.Value] = true
				}
			}
		}
		return true
	})
	return out
}

// filledAt returns, for each array, where a statement at the top level of the
// function first gives it at least one element. An assignment inside a branch
// or a loop may never run, and `mapfile` or `read -a` can produce nothing, so
// neither counts.
func filledAt(decl *syntax.FuncDecl, arrays map[string]bool) map[string]syntax.Pos {
	out := map[string]syntax.Pos{}
	block, ok := decl.Body.Cmd.(*syntax.Block)
	if !ok {
		return out
	}
	for _, stmt := range block.Stmts {
		call, ok := stmt.Cmd.(*syntax.CallExpr)
		if !ok || len(call.Args) > 0 {
			continue
		}
		for _, assign := range call.Assigns {
			if assign.Name == nil || !arrays[assign.Name.Value] || !hasSureElement(assign.Array) {
				continue
			}
			if _, seen := out[assign.Name.Value]; !seen {
				out[assign.Name.Value] = stmt.End()
			}
		}
	}
	return out
}

// hasSureElement reports whether an array literal yields at least one element
// whatever its expansions hold: one of its elements is not itself the
// expansion of a list, as `"${other[@]}"` or `"$@"` are.
func hasSureElement(array *syntax.ArrayExpr) bool {
	if array == nil {
		return false
	}
	for _, elem := range array.Elems {
		pe := wordParam(elem.Value)
		if pe == nil {
			return true
		}
		if arrayExpansion(pe) == "" && pe.Param.Value != "@" && pe.Param.Value != "*" {
			return true
		}
	}
	return false
}
