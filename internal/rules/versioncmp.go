package rules

import (
	"regexp"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG097",
		Section:  "Features and Bugs > Comparing Versions",
		Severity: lint.SeverityWarning,
		Doc:      "Compare versions with a version-aware helper, not `<`, `>` or arithmetic",
		Check:    checkVersionComparison,
	})
}

// versionName matches a variable that holds a version.
var versionName = regexp.MustCompile(`(?i)(^|_)(version|ver|vers)(_|$)|version$`)

// dottedVersion matches a literal version such as `1.10` or `2.3.4`.
var dottedVersion = regexp.MustCompile(`^v?[0-9]+\.[0-9]+`)

func checkVersionComparison(f *File, r *Reporter) {
	report := func(node syntax.Node) {
		r.At(node.Pos(), "versions do not compare as text, where `1.10` sorts before `1.9`, nor as numbers, which stop at the first dot; use `dybatpho::semver_compare`, or `sort -V` where GNU sort is certain")
	}
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.BinaryTest:
			if n.Op != syntax.TsBefore && n.Op != syntax.TsAfter {
				return true
			}
			if testsVersion(n.X) || testsVersion(n.Y) {
				report(n)
			}
		case *syntax.BinaryArithm:
			switch n.Op {
			case syntax.Lss, syntax.Gtr, syntax.Leq, syntax.Geq:
				if arithVersion(n.X) || arithVersion(n.Y) {
					report(n)
				}
			}
		}
		return true
	})
}

func testsVersion(expr syntax.TestExpr) bool {
	word, ok := expr.(*syntax.Word)
	if !ok {
		return false
	}
	if dottedVersion.MatchString(wordLiteral(word)) {
		return true
	}
	pe := wordParam(word)
	return pe != nil && pe.Param != nil && pe.Index == nil && versionName.MatchString(pe.Param.Value)
}

// arithVersion reports whether an arithmetic operand is a whole version
// string; an element such as `BASH_VERSINFO[0]` is a single number.
func arithVersion(expr syntax.ArithmExpr) bool {
	word, ok := expr.(*syntax.Word)
	if !ok {
		return false
	}
	if name := arithmName(word); name != "" {
		return versionName.MatchString(name)
	}
	pe := wordParam(word)
	return pe != nil && pe.Param != nil && pe.Index == nil && versionName.MatchString(pe.Param.Value)
}
