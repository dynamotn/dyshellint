package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG084",
		Section:  sectionValidationInSubst,
		Severity: lint.SeverityWarning,
		Doc:      "Do not call a function that sets globals inside `$(...)`, where the change is lost",
		Check:    checkLostSideEffects,
	})
}

// registrars are the commands whose whole effect is a change to the shell's
// own state.
var registrars = map[string]bool{"dybatpho::secret_register": true}

// globalWrite returns the first global a function sets in its own shell: an
// assignment to a name it did not declare local, `declare -g`, or a target of
// `read`, `mapfile` or `printf -v` with a literal name. A call to a registrar
// counts as one.
func globalWrite(decl *syntax.FuncDecl) string {
	names, _ := funcLocals(decl)
	local := map[string]bool{}
	for _, name := range names {
		local[name] = true
	}
	found := ""
	note := func(name string) {
		if found == "" && name != "" && name != "_" && !local[name] {
			found = name
		}
	}
	syntax.Walk(decl.Body, func(node syntax.Node) bool {
		if found != "" {
			return false
		}
		switch n := node.(type) {
		case *syntax.FuncDecl, *syntax.CmdSubst, *syntax.ProcSubst, *syntax.Subshell:
			return node == syntax.Node(decl)
		case *syntax.DeclClause:
			if n.Variant != nil && strings.Contains(declFlags(n), "g") {
				for _, arg := range n.Args {
					if arg.Name != nil {
						found = arg.Name.Value
					}
				}
			}
			return false
		case *syntax.Assign:
			if n.Name != nil {
				note(n.Name.Value)
			}
		case *syntax.CallExpr:
			if len(n.Args) > 0 {
				// `IFS=, read` or `TZ=UTC date` set the variable for that
				// command alone, so its assignments are not walked.
				builtinTargets(n, func(name string, _ syntax.Pos) { note(name) })
				if registrars[callName(n)] {
					found = callName(n)
				}
				return false
			}
			if name := callName(n); registrars[name] {
				found = name
				return false
			}
			builtinTargets(n, func(name string, _ syntax.Pos) { note(name) })
		}
		return true
	})
	return found
}

func checkLostSideEffects(f *File, r *Reporter) {
	p := f.project()
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		cs, ok := node.(*syntax.CmdSubst)
		if !ok {
			return true
		}
		for _, stmt := range cs.Stmts {
			for _, call := range leadingCalls(stmt) {
				name := callName(call)
				decl := p.funcs[name]
				if decl == nil {
					continue
				}
				if global := globalWrite(decl); global != "" {
					r.At(cs.Pos(), "%q sets %q, but inside `$(...)` it runs in a subshell and the change is lost when it ends; call it in the caller's shell and return the value through a nameref",
						name, global)
					return true
				}
			}
		}
		return true
	})
}
