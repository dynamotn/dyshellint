package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionFunctionNames = "Naming Conventions > Function Names"

// funcNamePattern is the whole vocabulary the guide allows in a function name:
// lowercase words joined by underscores, at most one `::` namespace separator,
// and an optional `_` or `__` privacy prefix.
var funcNamePattern = regexp.MustCompile(`^_{0,2}[a-z0-9]+[a-z0-9_]*(::[a-z0-9][a-z0-9_]*)?$`)

func init() {
	register(
		Rule{
			Code:     "BSG001",
			Section:  sectionFunctionNames,
			Severity: lint.SeverityError,
			Doc:      "Declare functions with the `function` keyword",
			Check:    checkFunctionKeyword,
		},
		Rule{
			Code:     "BSG002",
			Section:  sectionFunctionNames,
			Severity: lint.SeverityError,
			Doc:      "Do not write `()` after the name when using the `function` keyword",
			Check:    checkFunctionParens,
		},
		Rule{
			Code:     "BSG003",
			Section:  sectionFunctionNames,
			Severity: lint.SeverityError,
			Doc:      "Write function names in lowercase with underscores, never camelCase or PascalCase",
			Check:    checkFunctionNameStyle,
		},
		Rule{
			Code:     "BSG004",
			Section:  sectionFunctionNames,
			Severity: lint.SeverityError,
			Doc:      "Name library functions after the file: `<file>::name`, or `__<file>_name` when private",
			Check:    checkFunctionNamespace,
		},
	)
}

// eachFunc walks every function declaration in source order.
func eachFunc(f *File, fn func(*syntax.FuncDecl)) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		if decl, ok := node.(*syntax.FuncDecl); ok && decl.Name != nil {
			fn(decl)
		}
		return true
	})
}

func checkFunctionKeyword(f *File, r *Reporter) {
	eachFunc(f, func(decl *syntax.FuncDecl) {
		if !decl.RsrvWord {
			r.At(decl.Position, "%q is declared as `%s()`; use `function %s {` so the declaration stays greppable",
				decl.Name.Value, decl.Name.Value, decl.Name.Value)
		}
	})
}

func checkFunctionParens(f *File, r *Reporter) {
	eachFunc(f, func(decl *syntax.FuncDecl) {
		if decl.RsrvWord && decl.Parens {
			r.At(decl.Position, "`function %s()` carries redundant parentheses; write `function %s {`",
				decl.Name.Value, decl.Name.Value)
		}
	})
}

func checkFunctionNameStyle(f *File, r *Reporter) {
	eachFunc(f, func(decl *syntax.FuncDecl) {
		name := decl.Name.Value
		if funcNamePattern.MatchString(name) {
			return
		}
		switch {
		case strings.ContainsAny(name, "ABCDEFGHIJKLMNOPQRSTUVWXYZ"):
			r.At(decl.Name.Pos(), "%q uses capital letters; write function names in lowercase with underscores", name)
		case strings.Contains(name, "-"):
			r.At(decl.Name.Pos(), "%q uses `-`; separate words with underscores", name)
		default:
			r.At(decl.Name.Pos(), "%q does not read as `namespace::name`, `name`, `_name` or `__namespace_name`", name)
		}
	})
}

func checkFunctionNamespace(f *File, r *Reporter) {
	eachFunc(f, func(decl *syntax.FuncDecl) {
		name := decl.Name.Value
		switch f.Role {
		case RoleLibrary:
			if f.Namespace == "" {
				return
			}
			switch {
			case strings.HasPrefix(name, f.Namespace+"::"):
			case strings.HasPrefix(name, "__"+f.Namespace+"_"):
			default:
				r.At(decl.Name.Pos(), "%q does not belong to this library; name it `%s::%s` when it is public, or `__%s_%s` when it is private",
					name, f.Namespace, strings.TrimLeft(trimNamespace(name), "_"), f.Namespace, strings.TrimLeft(trimNamespace(name), "_"))
			}
		case RoleEntrypoint:
			if strings.Contains(name, "::") || strings.HasPrefix(name, "_") {
				return
			}
			r.At(decl.Name.Pos(), "%q is private to this script; prefix it with `_`, as in `_%s`", name, name)
		}
	})
}

// trimNamespace drops an existing namespace so a suggestion does not repeat it.
func trimNamespace(name string) string {
	if _, rest, ok := strings.Cut(name, "::"); ok {
		return rest
	}
	return name
}
