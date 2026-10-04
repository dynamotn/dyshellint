package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG106",
		Section:  sectionTempFiles,
		Severity: lint.SeverityWarning,
		Doc:      "Clean up on `EXIT` only, or end an `INT` or `TERM` handler with `exit`",
		Check:    checkCleanupTrap,
	})
}

// interrupts are the signals whose default action, ending the script, a
// handler replaces.
var interrupts = map[string]bool{"INT": true, "TERM": true, "SIGINT": true, "SIGTERM": true, "2": true, "15": true}

// checkCleanupTrap reports a cleanup handler installed on `INT` or `TERM` that
// never exits. The handler replaces the default action, so Ctrl-C deletes the
// temporary files and the script carries on without them, then exits 0. The
// `EXIT` trap already runs when Bash dies of either signal.
func checkCleanupTrap(f *File, r *Reporter) {
	funcs := f.project().funcs
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "trap" || len(call.Args) < 3 {
			return
		}
		handler := call.Args[1]
		if lit := wordLiteral(handler); lit == "-" || lit == "" && len(handler.Parts) > 0 && isEmptyQuote(handler) {
			return
		}
		signal := ""
		for _, arg := range call.Args[2:] {
			if interrupts[wordLiteral(arg)] {
				signal = wordLiteral(arg)
				break
			}
		}
		if signal == "" {
			return
		}
		names := splitWords(wordSource(handler))
		if len(names) == 1 && funcs[names[0]] != nil {
			body := funcs[names[0]].Body
			names = nil
			allCalls(body, func(_ *syntax.CallExpr, callee string) {
				names = append(names, callee)
			})
		}
		cleans, ends := false, false
		for _, word := range names {
			switch {
			case word == "rm" || word == "rmdir" || strings.Contains(strings.ToLower(word), "clean"):
				cleans = true
			case word == "exit" || word == "kill" || word == "dybatpho::die":
				ends = true
			}
		}
		if cleans && !ends {
			r.At(call.Pos(), "this cleanup replaces what `%s` does, so the script carries on after Ctrl-C and exits 0; trap `EXIT` alone, which Bash runs on `INT` and `TERM` too, or end the handler with `exit`", signal)
		}
	})
}

// isEmptyQuote reports whether a word is an empty quoted string, `”` or `""`,
// the way to ignore a signal on purpose.
func isEmptyQuote(word *syntax.Word) bool {
	src := wordSource(word)
	return src == "''" || src == `""`
}
