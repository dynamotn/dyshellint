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

// writerParams returns the positions of the arguments a function treats as
// the name of a variable to write: a nameref bound to `$N`, or `printf -v`,
// `read`, `mapfile` or `readarray` into it. Position 0 stands for a name the
// function chose from an expansion it cannot trace back to one argument. A
// name a `case` arm has already pinned to literal words is the function's own.
func writerParams(decl *syntax.FuncDecl) map[int]bool {
	out := map[int]bool{}
	fed := fedIndex(decl)
	var up map[syntax.Node]syntax.Node
	syntax.Walk(decl.Body, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.FuncDecl:
			return false
		case *syntax.DeclClause:
			if n.Variant == nil || !strings.Contains(declFlags(n), "n") {
				return true
			}
			for _, arg := range n.Args {
				if arg.Name != nil && arg.Value != nil && wordLiteral(arg.Value) == "" {
					out[argPosition(arg.Value, fed)] = true
				}
			}
		case *syntax.CallExpr:
			target := callerNamedTarget(n)
			if target == nil {
				return true
			}
			if up == nil {
				up = parents(decl.Body)
			}
			if !pinnedByCase(n, target, up) {
				out[paramPosition(target, fed)] = true
			}
		}
		return true
	})
	return out
}

// fedIndex maps the variables a function binds to one argument to that
// argument's position: the names `dybatpho::expect_args` binds in order, and
// assignments of a single `$N`.
func fedIndex(decl *syntax.FuncDecl) map[string]int {
	out := map[string]int{}
	syntax.Walk(decl.Body, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.FuncDecl:
			return false
		case *syntax.Assign:
			if n.Name == nil {
				return true
			}
			if pe := wordParam(n.Value); pe != nil && pe.Param != nil && isPositionalDigit(pe) {
				out[n.Name.Value] = int(pe.Param.Value[0] - '0')
			}
		case *syntax.CallExpr:
			if callName(n) != "dybatpho::expect_args" {
				return true
			}
			for i, arg := range n.Args[1:] {
				name := wordLiteral(arg)
				if name == "--" {
					break
				}
				out[name] = i + 1
			}
		}
		return true
	})
	return out
}

// argPosition returns the argument a word names a variable through, or 0.
func argPosition(word *syntax.Word, fed map[string]int) int {
	pe := wordParam(word)
	if pe == nil {
		return 0
	}
	return paramPosition(pe, fed)
}

func paramPosition(pe *syntax.ParamExp, fed map[string]int) int {
	if pe == nil || pe.Param == nil {
		return 0
	}
	if isPositionalDigit(pe) {
		return int(pe.Param.Value[0] - '0')
	}
	return fed[pe.Param.Value]
}

// callerNamedTarget returns the expansion a builtin stores into, when the
// variable is named by an expansion rather than by a literal.
func callerNamedTarget(call *syntax.CallExpr) *syntax.ParamExp {
	switch callName(call) {
	case "printf":
		for i := 1; i+1 < len(call.Args); i++ {
			if wordLiteral(call.Args[i]) == "-v" {
				return wordParam(call.Args[i+1])
			}
		}
	case "read", "mapfile", "readarray":
		if len(call.Args) >= 2 {
			return wordParam(call.Args[len(call.Args)-1])
		}
	}
	return nil
}

// pinnedByCase reports whether a call sits in a `case` arm over the same
// variable whose patterns are all literal names, as in
// `case "${name}" in min | max) printf -v "${name}" ...`: the target can then
// only be one of those names.
func pinnedByCase(call *syntax.CallExpr, target *syntax.ParamExp, up map[syntax.Node]syntax.Node) bool {
	for n := up[call]; n != nil; n = up[n] {
		item, ok := n.(*syntax.CaseItem)
		if !ok {
			continue
		}
		clause, ok := up[item].(*syntax.CaseClause)
		if !ok {
			return false
		}
		subject := wordParam(clause.Word)
		if subject == nil || subject.Param == nil || target.Param == nil || subject.Param.Value != target.Param.Value {
			return false
		}
		for _, pattern := range item.Patterns {
			if !isVarName(wordLiteral(pattern)) {
				return false
			}
		}
		return true
	}
	return false
}

func checkCallerNamedLocals(f *File, r *Reporter) {
	eachFunc(f, func(decl *syntax.FuncDecl) {
		if !isPublicFunc(decl) || !f.project().WritesCallerName(decl.Name.Value) {
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
