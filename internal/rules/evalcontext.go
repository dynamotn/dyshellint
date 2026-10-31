package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG130",
		Section:  sectionEval,
		Severity: lint.SeverityWarning,
		Doc:      "Remove an array element with a single-quoted subscript, `unset -v 'map[$key]'`, so the key is expanded once",
		Check:    checkUnsetSubscript,
	})
}

// checkUnsetSubscript reports `unset "map[${key}]"` and `unset arr[$i]`. The
// shell expands the subscript once as a word, and `unset` evaluates it again:
// a key holding `$(cmd)` runs the command, and an indexed array reads it as
// arithmetic.
func checkUnsetSubscript(f *File, r *Reporter) {
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "unset" {
			return
		}
		for _, arg := range call.Args[1:] {
			if strings.HasPrefix(wordLiteral(arg), "-") {
				continue
			}
			if pos, ok := expandedSubscript(arg); ok {
				r.At(pos, "`unset` evaluates the subscript of %s a second time, so a key holding `$(cmd)` runs it; single-quote the whole word, `unset -v 'name[$key]'`, so it is expanded once",
					wordSource(arg))
			}
		}
	})
}

// expandedSubscript returns the first expansion that follows a `[` outside
// single quotes in a word, as in `"map[${key}]"`.
func expandedSubscript(word *syntax.Word) (syntax.Pos, bool) {
	var parts []syntax.WordPart
	for _, part := range word.Parts {
		if dq, ok := part.(*syntax.DblQuoted); ok {
			parts = append(parts, dq.Parts...)
			continue
		}
		parts = append(parts, part)
	}
	bracket := false
	for _, part := range parts {
		switch p := part.(type) {
		case *syntax.Lit:
			bracket = bracket || strings.Contains(p.Value, "[")
		case *syntax.ParamExp, *syntax.CmdSubst, *syntax.ArithmExp:
			if bracket {
				return p.Pos(), true
			}
		}
	}
	return syntax.Pos{}, false
}
