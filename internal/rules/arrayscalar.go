package rules

import (
	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG113",
		Section:  sectionArrays,
		Severity: lint.SeverityError,
		Doc:      "Assign an array with `name=(...)`: a plain value replaces element 0 only",
		Check:    checkScalarIntoArray,
	})
}

// checkScalarIntoArray reports `name=value` or `name+=value` on a name the
// same function, or the top level of the file, uses as an array. The plain
// assignment writes element 0 and keeps every other element, so code that
// meant to start over still finds the old entries.
func checkScalarIntoArray(f *File, r *Reporter) {
	check := func(root syntax.Node, skipFuncs bool) {
		arrays, declared, namerefs := map[string]bool{}, map[*syntax.Assign]bool{}, map[string]bool{}
		walkScope(root, skipFuncs, func(node syntax.Node) {
			switch n := node.(type) {
			case *syntax.DeclClause:
				for _, arg := range n.Args {
					declared[arg] = true
					if arg.Name != nil && containsFlag(declFlags(n), 'n') {
						namerefs[arg.Name.Value] = true
					}
				}
				if flags := declFlags(n); containsFlag(flags, 'a') || containsFlag(flags, 'A') {
					for _, arg := range n.Args {
						if arg.Name != nil {
							arrays[arg.Name.Value] = true
						}
					}
				}
			case *syntax.Assign:
				if n.Name != nil && n.Array != nil {
					arrays[n.Name.Value] = true
				}
			}
		})
		walkScope(root, skipFuncs, func(node syntax.Node) {
			assign, ok := node.(*syntax.Assign)
			// A declaration sets up the name rather than writing to an array,
			// and a nameref takes the type of whatever it is bound to.
			if !ok || declared[assign] || assign.Name == nil || namerefs[assign.Name.Value] || !arrays[assign.Name.Value] || assign.Array != nil ||
				assign.Index != nil || assign.Naked || assign.Value == nil {
				return
			}
			op := "="
			if assign.Append {
				op = "+="
			}
			r.At(assign.Pos(), "%q is an array, and `%s%s...` writes only its element 0, keeping the others; write `%s%s(...)`",
				assign.Name.Value, assign.Name.Value, op, assign.Name.Value, op)
		})
	}
	eachFunc(f, func(decl *syntax.FuncDecl) { check(decl.Body, true) })
	check(f.Syntax, true)
}

// walkScope visits the nodes under root, leaving out the bodies of nested
// function declarations when skipFuncs is set, as they are a scope of their
// own.
func walkScope(root syntax.Node, skipFuncs bool, fn func(syntax.Node)) {
	syntax.Walk(root, func(node syntax.Node) bool {
		if _, ok := node.(*syntax.FuncDecl); ok && skipFuncs && node != root {
			return false
		}
		if node != nil {
			fn(node)
		}
		return true
	})
}
