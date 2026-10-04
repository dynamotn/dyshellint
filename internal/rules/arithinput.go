package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionArithmetic = "Features and Bugs > Arithmetic"

func init() {
	register(Rule{
		Code:     "BSG087",
		Section:  sectionArithmetic,
		Severity: lint.SeverityWarning,
		Doc:      "Validate a number read from the arguments or input before using it in arithmetic",
		Check:    checkUncheckedArithmetic,
	})
}

// numberCheck matches the name of a command that validates a number.
var numberCheck = regexp.MustCompile(`(?i)(^|::|_)(is|valid|validate|expect_int|is_int|is_number|number|integer|math_)`)

func checkUncheckedArithmetic(f *File, r *Reporter) {
	eachFunc(f, func(decl *syntax.FuncDecl) {
		inputs := map[string]bool{}
		// A private helper of a library is handed what its public caller has
		// already checked; only what it reads itself is input there.
		fromCaller := f.Role != RoleLibrary || isPublicFunc(decl)
		if fromCaller {
			for name := range fedIndex(decl) {
				inputs[name] = true
			}
		}
		allCalls(decl.Body, func(call *syntax.CallExpr, name string) {
			if name == "read" {
				builtinTargets(call, func(target string, _ syntax.Pos) { inputs[target] = true })
			}
		})
		checked := validatedNames(decl)
		reported := map[string]bool{}
		syntax.Walk(decl.Body, func(node syntax.Node) bool {
			var expr syntax.ArithmExpr
			switch n := node.(type) {
			case *syntax.FuncDecl:
				return false
			case *syntax.ArithmExp:
				expr = n.X
			case *syntax.ArithmCmd:
				expr = n.X
			default:
				return true
			}
			syntax.Walk(expr, func(inner syntax.Node) bool {
				word, ok := inner.(*syntax.Word)
				if !ok {
					return true
				}
				if strings.HasPrefix(wordSource(word), "10#") {
					return false
				}
				name := arithmOperand(word)
				if name == "" || checked[name] || reported[name] {
					return false
				}
				if !inputs[name] && !(fromCaller && len(name) == 1 && name[0] >= '1' && name[0] <= '9') {
					return false
				}
				reported[name] = true
				reportWithDybatpho(f, r, word.Pos(), "`dybatpho::is int \"${value}\"` accepts a decimal integer without a leading zero", "%q comes from the caller or the input, and arithmetic reads `08` as a bad octal number and `a[$(cmd)]` as a command to run; check it with a regular expression first, or write `10#$%s`",
					name, name)
				return false
			})
			return true
		})
	})
}

// arithmOperand returns the variable an arithmetic word reads: a bare name, or
// a single expansion such as `${count}` or `$1`.
func arithmOperand(word *syntax.Word) string {
	if name := arithmName(word); name != "" {
		return name
	}
	if pe := wordParam(word); pe != nil && pe.Param != nil && pe.Exp == nil && !pe.Length && pe.Index == nil {
		return pe.Param.Value
	}
	return ""
}

// validatedNames returns the variables a function checks: the left side of a
// `=~` test, the subject of a `case`, or an argument of a command whose name
// says it validates a number.
func validatedNames(decl *syntax.FuncDecl) map[string]bool {
	out := map[string]bool{}
	mark := func(word *syntax.Word) {
		if pe := wordParam(word); pe != nil && pe.Param != nil {
			out[pe.Param.Value] = true
		}
	}
	syntax.Walk(decl.Body, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.BinaryTest:
			if n.Op == syntax.TsReMatch {
				if word, ok := n.X.(*syntax.Word); ok {
					mark(word)
				}
			}
		case *syntax.CaseClause:
			mark(n.Word)
		case *syntax.CallExpr:
			if numberCheck.MatchString(callName(n)) {
				for _, arg := range n.Args[1:] {
					mark(arg)
				}
			}
		}
		return true
	})
	return out
}
