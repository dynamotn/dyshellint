package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionWildcards = "Features and Bugs > Wildcard Expansion of Filenames"

func init() {
	register(Rule{
		Code:     "BSG093",
		Section:  sectionWildcards,
		Severity: lint.SeverityWarning,
		Doc:      "Check that each match of a glob loop exists, unless `nullglob` is on",
		Check:    checkGlobLoop,
	})
}

func checkGlobLoop(f *File, r *Reporter) {
	if setsNullglob(f) {
		return
	}
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		clause, ok := node.(*syntax.ForClause)
		if !ok {
			return true
		}
		iter, ok := clause.Loop.(*syntax.WordIter)
		if !ok || iter.Name == nil {
			return true
		}
		var glob *syntax.Word
		for _, item := range iter.Items {
			if hasUnquotedGlob(item) {
				glob = item
				break
			}
		}
		if glob == nil || checksExistence(clause.Do, iter.Name.Value) {
			return true
		}
		r.At(glob.Pos(), "a glob that matches nothing stays as the literal %s, so the loop runs once on a path that does not exist; test `[[ -e \"${%s}\" ]] || continue`, or turn on `shopt -s nullglob`",
			wordSource(glob), iter.Name.Value)
		return true
	})
}

// hasUnquotedGlob reports whether a word has a `*`, `?` or `[` the shell
// expands as a pattern.
func hasUnquotedGlob(word *syntax.Word) bool {
	for _, part := range word.Parts {
		if lit, ok := part.(*syntax.Lit); ok && strings.ContainsAny(lit.Value, "*?[") {
			return true
		}
	}
	return false
}

// checksExistence reports whether a loop body tests that its variable names a
// file, directory or link that exists.
func checksExistence(body []*syntax.Stmt, name string) bool {
	found := false
	for _, stmt := range body {
		syntax.Walk(stmt, func(node syntax.Node) bool {
			switch n := node.(type) {
			case *syntax.UnaryTest:
				switch n.Op {
				case syntax.TsExists, syntax.TsRegFile, syntax.TsDirect, syntax.TsSmbLink, syntax.TsNoEmpty, syntax.TsRead:
					if w, ok := n.X.(*syntax.Word); ok && expandsName(w, name) {
						found = true
					}
				}
			case *syntax.CallExpr:
				if c := callName(n); c == "dybatpho::is" || strings.HasSuffix(c, "_is_snapshot") || c == "test" || c == "[" {
					for _, arg := range n.Args[1:] {
						if expandsName(arg, name) {
							found = true
						}
					}
				}
			}
			return !found
		})
	}
	return found
}

func expandsName(word *syntax.Word, name string) bool {
	found := false
	syntax.Walk(word, func(node syntax.Node) bool {
		if pe, ok := node.(*syntax.ParamExp); ok && pe.Param != nil && pe.Param.Value == name {
			found = true
		}
		return !found
	})
	return found
}

// setsNullglob reports whether a file turns on `nullglob` or `failglob`.
func setsNullglob(f *File) bool {
	found := false
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "shopt" || !hasFlag(call, 's') {
			return
		}
		for _, arg := range call.Args[1:] {
			if lit := wordLiteral(arg); lit == "nullglob" || lit == "failglob" {
				found = true
			}
		}
	})
	return found
}
