package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG085",
		Section:  sectionArrays,
		Severity: lint.SeverityWarning,
		Doc:      "Walk an array the function did not build with `\"${!a[@]}\"`, not by counting to `${#a[@]}`",
		Check:    checkCountedIndexLoop,
	})
}

func checkCountedIndexLoop(f *File, r *Reporter) {
	eachFunc(f, func(decl *syntax.FuncDecl) {
		namerefs, built := arrayOrigins(decl)
		syntax.Walk(decl.Body, func(node syntax.Node) bool {
			clause, ok := node.(*syntax.ForClause)
			if !ok {
				return true
			}
			loop, ok := clause.Loop.(*syntax.CStyleLoop)
			if !ok || loop.Cond == nil {
				return true
			}
			syntax.Walk(loop.Cond, func(n syntax.Node) bool {
				pe, ok := n.(*syntax.ParamExp)
				if !ok || !pe.Length || pe.Param == nil || pe.Index == nil {
					return true
				}
				name := pe.Param.Value
				if !namerefs[name] && built[name] {
					return true
				}
				r.At(pe.Pos(), "%q comes from the caller, and an array with a gap in its indexes makes `${%s[i]}` unset for some i below `${#%s[@]}`; walk `\"${!%s[@]}\"` instead",
					name, name, name, name)
				return false
			})
			return true
		})
	})
}

// arrayOrigins returns the namerefs a function binds, and the arrays it builds
// itself as locals, which are dense unless the function unsets an element.
func arrayOrigins(decl *syntax.FuncDecl) (namerefs, built map[string]bool) {
	namerefs, built = map[string]bool{}, map[string]bool{}
	syntax.Walk(decl.Body, func(node syntax.Node) bool {
		clause, ok := node.(*syntax.DeclClause)
		if !ok || clause.Variant == nil {
			return true
		}
		flags := declFlags(clause)
		for _, arg := range clause.Args {
			if arg.Name == nil {
				continue
			}
			switch {
			case strings.Contains(flags, "n"):
				namerefs[arg.Name.Value] = true
			case strings.Contains(flags, "a") || arg.Array != nil:
				built[arg.Name.Value] = true
			}
		}
		return true
	})
	return namerefs, built
}
