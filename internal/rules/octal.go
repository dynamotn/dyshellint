package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG129",
		Section:  sectionArithmetic,
		Severity: lint.SeverityWarning,
		Doc:      "Read a number checked only for digits in base ten: `10#${n}`, or normalise it once with `n=$((10#${n}))`",
		Check:    checkOctalAfterDigitCheck,
	})
}

// digitRegex matches a `=~` pattern that accepts a run of decimal digits,
// leading zeros included, such as `^[0-9]+$` or `^-?[[:digit:]]{1,18}$`.
var digitRegex = regexp.MustCompile(`^\^(-\?|\[-\+?\]\?|\[\+-\]\?)?\[(0-9|\[:digit:\])\](\+|\*|\{[0-9,]+\})\$$`)

// nonDigitPattern matches a `case` pattern that refuses anything but digits,
// such as `*[!0-9]*`.
var nonDigitPattern = regexp.MustCompile(`^\*\[[!^](0-9|\[:digit:\])\]\*$`)

// checkOctalAfterDigitCheck reports a variable checked against a digits-only
// pattern and then used bare in arithmetic. The check lets `08` and `010`
// through, and arithmetic reads them as octal: the first stops the script, the
// second is eight.
func checkOctalAfterDigitCheck(f *File, r *Reporter) {
	check := func(scope syntax.Node) {
		checked, normalised := map[string]uint{}, map[string]bool{}
		var handed []namedAt
		walkScope(scope, true, func(node syntax.Node) {
			switch n := node.(type) {
			case *syntax.BinaryTest:
				re, ok := n.Y.(*syntax.Word)
				if n.Op != syntax.TsReMatch || !ok || !digitRegex.MatchString(wordLiteral(re)) {
					return
				}
				if word, ok := n.X.(*syntax.Word); ok {
					markChecked(checked, arithmOperand(word), n.Pos().Offset())
				}
			case *syntax.CaseClause:
				for _, item := range n.Items {
					for _, pattern := range item.Patterns {
						if nonDigitPattern.MatchString(wordLiteral(pattern)) {
							markChecked(checked, arithmOperand(n.Word), n.Pos().Offset())
						}
					}
				}
			case *syntax.Assign:
				// `n=$((10#${n}))` reads the number in base ten once for all.
				if n.Name != nil && n.Value != nil && strings.Contains(wordSource(n.Value), "10#") {
					normalised[n.Name.Value] = true
				}
			case *syntax.BinaryArithm:
				// So does `((n = 10#${n}))`.
				word, ok := n.X.(*syntax.Word)
				if ok && n.Op == syntax.Assgn && strings.Contains(f.Text(n.Y.Pos(), n.Y.End()), "10#") {
					if name := arithmOperand(word); name != "" {
						normalised[name] = true
					}
				}
			case *syntax.CallExpr:
				// A helper handed the name after the check, `decimal_into n`,
				// may rewrite it through a nameref.
				if len(n.Args) == 0 {
					return
				}
				for _, arg := range n.Args[1:] {
					if len(arg.Parts) != 1 {
						continue
					}
					if lit, ok := arg.Parts[0].(*syntax.Lit); ok {
						handed = append(handed, namedAt{lit.Value, arg.Pos().Offset()})
					}
				}
			}
		})
		for _, h := range handed {
			if at, ok := checked[h.name]; ok && h.offset > at {
				normalised[h.name] = true
			}
		}
		reported := map[string]bool{}
		walkScope(scope, true, func(node syntax.Node) {
			var expr syntax.ArithmExpr
			switch n := node.(type) {
			case *syntax.ArithmExp:
				expr = n.X
			case *syntax.ArithmCmd:
				expr = n.X
			default:
				return
			}
			// The target of `n = ...` is written, not read.
			targets := map[*syntax.Word]bool{}
			syntax.Walk(expr, func(inner syntax.Node) bool {
				if bin, ok := inner.(*syntax.BinaryArithm); ok && bin.Op == syntax.Assgn {
					if word, ok := bin.X.(*syntax.Word); ok {
						targets[word] = true
					}
				}
				word, ok := inner.(*syntax.Word)
				if !ok {
					return true
				}
				if targets[word] || strings.HasPrefix(wordSource(word), "10#") {
					return false
				}
				name := arithmOperand(word)
				// A use before the check reads a value the check has not seen.
				at, ok := checked[name]
				if name == "" || !ok || word.Pos().Offset() < at || normalised[name] || reported[name] {
					return false
				}
				reported[name] = true
				r.At(word.Pos(), "%q passed a digits-only check, which lets `08` and `010` through, and arithmetic reads them as octal: `08` stops the script and `010` is eight; write `10#${%s}`, or normalise it once with `%s=$((10#${%s}))`",
					name, name, name, name)
				return false
			})
		})
	}
	eachFunc(f, func(decl *syntax.FuncDecl) { check(decl.Body) })
	check(f.Syntax)
}

// namedAt is a variable name handed to a command, and where.
type namedAt struct {
	name   string
	offset uint
}

// markChecked records the first place a variable is checked for digits.
func markChecked(checked map[string]uint, name string, offset uint) {
	if at, ok := checked[name]; name != "" && (!ok || offset < at) {
		checked[name] = offset
	}
}
