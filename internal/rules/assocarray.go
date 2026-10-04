package rules

import (
	"regexp"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionAssocArrays = "Features and Bugs > Associative Arrays"

func init() {
	register(Rule{
		Code:     "BSG103",
		Section:  sectionAssocArrays,
		Severity: lint.SeverityError,
		Doc:      "Declare a map with `-A` before assigning a string key to it",
		Check:    checkUndeclaredMap,
	})
}

// arithKey matches a subscript Bash can read as arithmetic: a number or a
// variable name. Anything else is a string key.
var arithKey = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

// checkUndeclaredMap reports `name["key"]=value` on a name that no file of the
// run declares with `-A`. Bash then treats `name` as an indexed array and the
// key as arithmetic, so every string key lands on index 0. A bare word such as
// `name[key]` is left alone: it is a valid arithmetic index too.
func checkUndeclaredMap(f *File, r *Reporter) {
	maps := map[string]bool{}
	for _, file := range f.project().Files() {
		collectMaps(file.Syntax, maps)
	}
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		assign, ok := node.(*syntax.Assign)
		if !ok || assign.Name == nil || assign.Index == nil || maps[assign.Name.Value] {
			return true
		}
		key, ok := assign.Index.(*syntax.Word)
		if !ok || !stringKey(key) {
			return true
		}
		r.At(assign.Pos(), "%q is not declared with `-A`, so Bash reads the key %s as arithmetic and stores it at index 0; declare it with `local -A %s=()` or `declare -A %s=()`",
			assign.Name.Value, wordSource(key), assign.Name.Value, assign.Name.Value)
		return true
	})
}

// collectMaps adds every name a file declares with `-A`, and every nameref,
// whose type is the type of the variable it is bound to.
func collectMaps(root syntax.Node, out map[string]bool) {
	syntax.Walk(root, func(node syntax.Node) bool {
		decl, ok := node.(*syntax.DeclClause)
		if !ok {
			return true
		}
		if flags := declFlags(decl); !containsFlag(flags, 'A') && !containsFlag(flags, 'n') {
			return true
		}
		for _, arg := range decl.Args {
			if arg.Name != nil {
				out[arg.Name.Value] = true
			} else if lit := wordLiteral(arg.Value); arg.Naked && arithKey.MatchString(lit) {
				out[lit] = true
			}
		}
		return true
	})
}

// numberKey matches a subscript that is a plain number.
var numberKey = regexp.MustCompile(`^[0-9]+$`)

// stringKey reports whether a subscript is a literal string that arithmetic
// cannot use as an index: quoted text that is not a number, or a bare word
// with a character no variable name holds. A key that expands a variable may
// well be a number, so it is left alone.
func stringKey(word *syntax.Word) bool {
	lit := wordLiteral(word)
	if lit == "" {
		return false
	}
	for _, part := range word.Parts {
		switch part.(type) {
		case *syntax.SglQuoted, *syntax.DblQuoted:
			return !numberKey.MatchString(lit)
		}
	}
	return !arithKey.MatchString(lit)
}
