package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionCallerNames = "Naming Conventions > Local Names in Nameref and Callback Functions"

func init() {
	register(
		Rule{
			Code:     "BSG014",
			Section:  sectionCallerNames,
			Severity: lint.SeverityError,
			Doc:      "Prefix every local of a function that writes to a variable its caller names",
			Check:    checkCallerNamedLocals,
		},
	)
}

// writesCallerName reports whether a function writes to a variable whose name
// its caller chose: a nameref bound to an expansion, `printf -v`, `read`,
// `mapfile` or `readarray` into a name that is not a literal.
func writesCallerName(decl *syntax.FuncDecl) bool {
	found := false
	syntax.Walk(decl.Body, func(node syntax.Node) bool {
		if found {
			return false
		}
		switch n := node.(type) {
		case *syntax.FuncDecl:
			return false
		case *syntax.DeclClause:
			if n.Variant == nil || !strings.Contains(declFlags(n), "n") {
				return true
			}
			for _, arg := range n.Args {
				if arg.Name != nil && arg.Value != nil && wordLiteral(arg.Value) == "" {
					found = true
				}
			}
		case *syntax.CallExpr:
			found = intoCallerName(n)
		}
		return true
	})
	return found
}

// intoCallerName reports whether a builtin call stores into a variable named by
// an expansion rather than by a literal.
func intoCallerName(call *syntax.CallExpr) bool {
	switch callName(call) {
	case "printf":
		for i := 1; i+1 < len(call.Args); i++ {
			if wordLiteral(call.Args[i]) == "-v" {
				return wordParam(call.Args[i+1]) != nil
			}
		}
	case "read", "mapfile", "readarray":
		if len(call.Args) < 2 {
			return false
		}
		return wordParam(call.Args[len(call.Args)-1]) != nil
	}
	return false
}

func checkCallerNamedLocals(f *File, r *Reporter) {
	eachFunc(f, func(decl *syntax.FuncDecl) {
		if !isPublicFunc(decl) || !writesCallerName(decl) {
			return
		}
		names, at := funcLocals(decl)
		for _, name := range names {
			if isPrivateName(name) {
				continue
			}
			r.At(at[name], "%q writes to a variable its caller names, so a caller whose variable is called %q gets this local instead; prefix it, as in `__%s_%s`",
				decl.Name.Value, name, privatePrefix(f, decl), name)
		}
	})
}

// privatePrefix suggests the prefix a function's locals take: the namespace of
// the file, or of the function, followed by the rest of its name.
func privatePrefix(f *File, decl *syntax.FuncDecl) string {
	name := strings.TrimLeft(decl.Name.Value, "_")
	if i := strings.LastIndex(name, "::"); i >= 0 {
		return strings.ReplaceAll(name[:i], "::", "_") + "_" + name[i+2:]
	}
	if f.Namespace != "" && !strings.HasPrefix(name, f.Namespace) {
		return f.Namespace + "_" + name
	}
	return name
}
