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

// versionName matches a variable that holds a version, and versionPart one
// that holds a single number taken from it, such as `version_major`: splitting
// a version and comparing its parts as numbers is what the guide recommends.
var (
	versionName = regexp.MustCompile(`(?i)(^|_)(version|ver|vers)(_|$)|version$`)
	versionPart = regexp.MustCompile(`(?i)_(major|minor|patch|count)$`)
)

// dottedVersion matches a literal version such as `1.10` or `2.3.4`.
var dottedVersion = regexp.MustCompile(`^v?[0-9]+\.[0-9]+`)

func checkVersionComparison(f *File, r *Reporter) {
	report := func(node syntax.Node) {
		r.At(node.Pos(), "versions do not compare as text, where `1.10` sorts before `1.9`, nor as numbers, which stop at the first dot; use `dybatpho::semver_compare`; `sort -V` only orders plain dotted numbers, and puts `2.0.0-rc1` after `2.0.0`")
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
	return versionParam(wordParam(word))
}

// arithVersion reports whether an arithmetic operand is a whole version
// string; an element such as `BASH_VERSINFO[0]` is a single number.
func arithVersion(expr syntax.ArithmExpr) bool {
	word, ok := expr.(*syntax.Word)
	if !ok {
		return false
	}
	if name := arithmName(word); name != "" {
		return isVersionName(name)
	}
	return versionParam(wordParam(word))
}

// versionParam reports whether an expansion yields a whole version string.
// An element or a length such as `${#version}` is a plain number.
func versionParam(pe *syntax.ParamExp) bool {
	return pe != nil && pe.Param != nil && pe.Index == nil && !pe.Length && isVersionName(pe.Param.Value)
}

func isVersionName(name string) bool {
	return versionName.MatchString(name) && !versionPart.MatchString(name)
}
