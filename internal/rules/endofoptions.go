package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionEndOfOptions = "Calling Commands > End of Options"

func init() {
	register(Rule{
		Code:     "BSG057",
		Section:  sectionEndOfOptions,
		Severity: lint.SeverityWarning,
		Doc:      "Put `--` before an operand that is a single expansion, which may start with `-`",
		Check:    checkEndOfOptions,
	})
}

// optionTakers are the commands that read a leading `-` in an operand as an
// option, with the options of each that take a value.
var optionTakers = map[string]string{
	"rm": "", "mv": "tS", "cp": "tS", "ln": "tS", "touch": "dtr", "mkdir": "m",
	"cat": "", "ls": "IwT", "chmod": "", "chown": "", "grep": "efmABCd", "sed": "efl",
}

// skipsFirst are the commands whose first operand is not a path: a mode or an
// owner.
var skipsFirst = map[string]bool{"chmod": true, "chown": true}

func checkEndOfOptions(f *File, r *Reporter) {
	eachCall(f, func(call *syntax.CallExpr, name string) {
		valued, ok := optionTakers[name]
		if !ok {
			return
		}
		operand := 0
		for i := 1; i < len(call.Args); i++ {
			arg := call.Args[i]
			lit := wordLiteral(arg)
			if lit == "--" {
				return
			}
			if strings.HasPrefix(lit, "-") && len(lit) > 1 {
				// `grep -e PATTERN` names the pattern, so no operand can be
				// mistaken for it; an option with a value takes the next word.
				if (name == "grep" || name == "sed") && (strings.ContainsAny(lit[1:], "ef") || strings.HasPrefix(lit, "--regexp") || strings.HasPrefix(lit, "--expression")) {
					return
				}
				if !strings.HasPrefix(lit, "--") && strings.ContainsAny(lit[len(lit)-1:], valued) {
					i++
				}
				continue
			}
			operand++
			if operand == 1 && skipsFirst[name] {
				continue
			}
			// Only the first operand can still be taken for an option once
			// the command has seen a non-option word on GNU, and it is the one
			// a value starting with `-` is most often handed in.
			if isScalarExpansion(arg) {
				r.At(arg.Pos(), "`%s` reads a value that starts with `-` as an option; write `%s -- %s`",
					name, name, wordSource(arg))
			}
			return
		}
	})
}

// isScalarExpansion reports whether a word is one scalar expansion and nothing
// else, such as `"$file"` or `"${path}"`, so its whole value decides how it
// starts.
func isScalarExpansion(word *syntax.Word) bool {
	pe := wordParam(word)
	if pe == nil || pe.Param == nil || pe.Length || pe.Excl || pe.Exp != nil {
		return false
	}
	switch pe.Param.Value {
	case "@", "*", "#", "?", "$", "!", "0":
		return false
	}
	if pe.Index != nil {
		if w, ok := pe.Index.(*syntax.Word); ok {
			if lit := wordLiteral(w); lit == "@" || lit == "*" {
				return false
			}
		}
	}
	return true
}
