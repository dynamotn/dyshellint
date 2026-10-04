package rules

import (
	"regexp"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionStructuredOutput = "Features and Bugs > Building Structured Output"

func init() {
	register(Rule{
		Code:     "BSG086",
		Section:  sectionStructuredOutput,
		Severity: lint.SeverityWarning,
		Doc:      "Do not build JSON by putting a raw value between quotes in a format string",
		Check:    checkHandmadeJSON,
	})
}

// jsonKey matches the `"key":` of a JSON object, escaped or not.
var jsonKey = regexp.MustCompile(`\\?"[A-Za-z_][A-Za-z0-9_-]*\\?"\s*:`)

// quotedValue matches a value spliced between JSON quotes: `"%s"` or
// `"${name}"`, escaped or not.
var quotedValue = regexp.MustCompile(`:\s*\\?"(%s|\$\{?[A-Za-z_][A-Za-z0-9_]*\}?)\\?"`)

// escaperHint matches a function that escapes its values before printing them.
var escaperHint = regexp.MustCompile(`(?i)escape|json_string|jq\b|yq\b|@json`)

func checkHandmadeJSON(f *File, r *Reporter) {
	check := func(scope string, word *syntax.Word, command string) {
		text := wordSource(word)
		if !jsonKey.MatchString(text) || !quotedValue.MatchString(text) || escaperHint.MatchString(scope) {
			return
		}
		r.At(word.Pos(), "`%s` puts a value between JSON quotes as it is, so a `\"`, a `\\` or a newline in it breaks the document; escape it first, or build the document with `jq` or the dybatpho JSON helpers",
			command)
	}
	eachFunc(f, func(decl *syntax.FuncDecl) {
		scope := f.Text(decl.Pos(), decl.End())
		allCalls(decl.Body, func(call *syntax.CallExpr, name string) {
			switch name {
			case "printf":
				for i := 1; i < len(call.Args); i++ {
					lit := wordLiteral(call.Args[i])
					if lit == "-v" {
						i++
						continue
					}
					if lit == "--" {
						continue
					}
					check(scope, call.Args[i], name)
					return
				}
			case "echo":
				for _, arg := range call.Args[1:] {
					check(scope, arg, name)
				}
			}
		})
	})
}
