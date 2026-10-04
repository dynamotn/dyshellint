package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionLibrarySideEffects = "Environment > Library Side Effects"

func init() {
	register(Rule{
		Code:     "BSG090",
		Section:  sectionLibrarySideEffects,
		Severity: lint.SeverityWarning,
		Doc:      "Return a status from a library function instead of calling `exit`",
		Check:    checkLibraryExit,
	})
}

// stopperName matches the functions whose job is to end the script, and the
// handlers that run on a signal or at exit.
var stopperName = regexp.MustCompile(`(?i)(die|fatal|abort|panic|exit|handler|trap|cleanup|killed|interrupt|on_signal)`)

func checkLibraryExit(f *File, r *Reporter) {
	if f.Role != RoleLibrary {
		return
	}
	handlers := trapHandlers(f)
	up := parents(f.Syntax)
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "exit" || inSubshell(call, up) {
			return
		}
		decl := enclosingFunc(call, up)
		if decl == nil || stopperName.MatchString(ownName(f, decl.Name.Value)) || handlers[decl.Name.Value] {
			return
		}
		r.At(call.Pos(), "`exit` in %q ends the whole script that sourced this library, skipping its cleanup and its own error handling; `return` a status and let the caller decide",
			decl.Name.Value)
	})
}

// trapHandlers returns the functions a file installs as signal handlers.
func trapHandlers(f *File) map[string]bool {
	out := map[string]bool{}
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "trap" && name != "dybatpho::trap" || len(call.Args) < 2 {
			return
		}
		for _, word := range splitWords(wordLiteral(call.Args[1])) {
			out[word] = true
		}
	})
	return out
}

var wordPattern = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_:]*`)

func splitWords(text string) []string { return wordPattern.FindAllString(text, -1) }

// ownName returns a function's name without the namespace of its file, so
// that `libexit::check` is judged by `check`.
func ownName(f *File, name string) string {
	if i := strings.LastIndex(name, "::"); i >= 0 {
		return name[i+2:]
	}
	if f.Namespace != "" {
		return strings.TrimPrefix(strings.TrimLeft(name, "_"), f.Namespace+"_")
	}
	return name
}
