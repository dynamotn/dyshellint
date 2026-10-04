package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionLocalVariables = "Features and Bugs > Local Variables"

func init() {
	register(Rule{
		Code:     "BSG016",
		Section:  sectionLocalVariables,
		Severity: lint.SeverityWarning,
		Doc:      "Declare loop, arithmetic and `read` targets with `local` inside a function",
		Check:    checkUndeclaredTargets,
	})
}

// setTag matches an shdoc `@set NAME` line, which documents a variable the
// function sets on purpose for its caller.
var setTag = regexp.MustCompile(`@set\s+([A-Za-z_][A-Za-z0-9_]*)`)

// readValueOptions are the options of `read` that take a value, which is not a
// variable name; `-a` takes the name of the array to fill.
const readValueOptions = "dnNptui"

// mapfileValueOptions are the options of `mapfile` and `readarray` that take a
// value.
const mapfileValueOptions = "dnOsuCc"

// assignedTargets calls fn for every variable a body sets without an
// assignment statement: a `for` loop variable, an arithmetic assignment, or the
// variable `read`, `mapfile`, `readarray`, `printf -v` or `getopts` fills.
// Plain assignments are BSG011's.
func assignedTargets(body syntax.Node, fn func(name string, pos syntax.Pos)) {
	syntax.Walk(body, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.FuncDecl:
			return false
		case *syntax.ForClause:
			if iter, ok := n.Loop.(*syntax.WordIter); ok && iter.Name != nil {
				fn(iter.Name.Value, iter.Name.Pos())
			}
		case *syntax.BinaryArithm:
			if isArithmAssign(n.Op) {
				if name := arithmName(n.X); name != "" {
					fn(name, n.X.Pos())
				}
			}
		case *syntax.UnaryArithm:
			if n.Op == syntax.Inc || n.Op == syntax.Dec {
				if name := arithmName(n.X); name != "" {
					fn(name, n.X.Pos())
				}
			}
		case *syntax.CallExpr:
			builtinTargets(n, fn)
		}
		return true
	})
}

func isArithmAssign(op syntax.BinAritOperator) bool {
	switch op {
	case syntax.Assgn, syntax.AddAssgn, syntax.SubAssgn, syntax.MulAssgn,
		syntax.QuoAssgn, syntax.RemAssgn, syntax.AndAssgn, syntax.OrAssgn,
		syntax.XorAssgn, syntax.ShlAssgn, syntax.ShrAssgn:
		return true
	}
	return false
}

// arithmName returns the variable an arithmetic operand names, when it is a
// plain name rather than an expression or an array element.
func arithmName(expr syntax.ArithmExpr) string {
	word, ok := expr.(*syntax.Word)
	if !ok {
		return ""
	}
	name := wordLiteral(word)
	if !isVarName(name) {
		return ""
	}
	return name
}

var varNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func isVarName(name string) bool { return varNamePattern.MatchString(name) }

// builtinTargets reports the variables a builtin fills, by name.
func builtinTargets(call *syntax.CallExpr, fn func(name string, pos syntax.Pos)) {
	args := call.Args
	switch callName(call) {
	case "read":
		for i := 1; i < len(args); i++ {
			word := wordLiteral(args[i])
			if strings.HasPrefix(word, "-") && len(word) > 1 {
				opts := word[1:]
				if strings.ContainsAny(opts[len(opts)-1:], readValueOptions+"a") && i+1 < len(args) {
					if strings.HasSuffix(opts, "a") {
						report(args[i+1], fn)
					}
					i++
				}
				continue
			}
			report(args[i], fn)
		}
	case "mapfile", "readarray":
		for i := 1; i < len(args); i++ {
			word := wordLiteral(args[i])
			if strings.HasPrefix(word, "-") && len(word) > 1 {
				if strings.ContainsAny(word[len(word)-1:], mapfileValueOptions) {
					i++
				}
				continue
			}
			report(args[i], fn)
		}
	case "printf":
		for i := 1; i+1 < len(args); i++ {
			if wordLiteral(args[i]) == "-v" {
				report(args[i+1], fn)
				return
			}
		}
	case "getopts":
		if len(args) >= 3 {
			report(args[2], fn)
		}
	}
}

// report passes on a word when it is a plain variable name; an array element
// such as `parts[1]` reports its array.
func report(word *syntax.Word, fn func(name string, pos syntax.Pos)) {
	name := wordLiteral(word)
	if i := strings.IndexByte(name, '['); i > 0 {
		name = name[:i]
	}
	if isVarName(name) {
		fn(name, word.Pos())
	}
}

func checkUndeclaredTargets(f *File, r *Reporter) {
	comments := funcComments(f)
	eachFunc(f, func(decl *syntax.FuncDecl) {
		names, _ := funcLocals(decl)
		declared := map[string]bool{}
		for _, name := range names {
			declared[name] = true
		}
		for _, global := range globalDecls(decl) {
			declared[global] = true
		}
		for _, match := range setTag.FindAllStringSubmatch(commentText(comments[decl]), -1) {
			declared[match[1]] = true
		}
		reported := map[string]bool{}
		assignedTargets(decl.Body, func(name string, pos syntax.Pos) {
			// UPPERCASE is the guide's marker for a value shared on purpose, and
			// `_` is the shell's own scratch variable.
			if declared[name] || reported[name] || name == "_" || name == strings.ToUpper(name) {
				return
			}
			reported[name] = true
			r.At(pos, "%q is set inside %q without `local`, so it leaks out of the function and overwrites a caller's variable of that name; declare it `local`, or document it with `@set`",
				name, decl.Name.Value)
		})
	})
}

// globalDecls returns the variables a function declares global on purpose,
// with `declare -g`.
func globalDecls(decl *syntax.FuncDecl) []string {
	var names []string
	syntax.Walk(decl.Body, func(node syntax.Node) bool {
		clause, ok := node.(*syntax.DeclClause)
		if !ok || clause.Variant == nil || !strings.Contains(declFlags(clause), "g") {
			return true
		}
		for _, arg := range clause.Args {
			if arg.Name != nil {
				names = append(names, arg.Name.Value)
			}
		}
		return true
	})
	return names
}
