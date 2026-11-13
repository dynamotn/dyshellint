package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG092",
		Section:  sectionLibrarySideEffects,
		Severity: lint.SeverityWarning,
		Doc:      "Scope a shell option, `IFS` or `umask` a library function changes, with `local -`, `local IFS` or a subshell",
		Check:    checkLibraryShellState,
	})
}

// scopedOptions are the `set` letters that change how the caller's own code
// runs. `-x` and `-v` are tracing, which the trace helpers turn on for the
// whole script on purpose.
const scopedOptions = "eCfua"

// optionNames are the long names of the scopedOptions letters, so that
// `set +o noglob` and `set -f` count as one option turned off and on again.
var optionNames = map[byte]string{'e': "errexit", 'C': "noclobber", 'f': "noglob", 'u': "nounset", 'a': "allexport"}

// scopedNames are the options a function has to put back: the scopedOptions
// letters and `pipefail`, which has none.
var scopedNames = map[string]bool{"errexit": true, "noclobber": true, "noglob": true, "nounset": true, "allexport": true, "pipefail": true}

func checkLibraryShellState(f *File, r *Reporter) {
	if f.Role != RoleLibrary {
		return
	}
	up := f.parents()
	eachFunc(f, func(decl *syntax.FuncDecl) {
		text := f.Text(decl.Pos(), decl.End())
		localOptions := hasLocalDash(decl)
		names, _ := funcLocals(decl)
		localIFS := false
		for _, name := range names {
			if name == "IFS" {
				localIFS = true
			}
		}
		report := func(node syntax.Node, what string) {
			if inSubshell(node, up) {
				return
			}
			r.At(node.Pos(), "%s in %q leaks into the script that sourced the library; scope it with `local -`, `local IFS`, a `( ... )` subshell, or save and restore it",
				what, decl.Name.Value)
		}
		restored := restoredState(decl)
		saved := savesOptions(decl)
		allCalls(decl.Body, func(call *syntax.CallExpr, name string) {
			switch name {
			case "set":
				if restored["set"] {
					return
				}
				if localOptions || saved || strings.Contains(text, "$-") {
					return
				}
				for _, arg := range setOptionWords(call) {
					lit := wordLiteral(arg)
					if lit == "-o" || lit == "+o" || (len(lit) > 1 && (lit[0] == '-' || lit[0] == '+') && strings.ContainsAny(lit[1:], scopedOptions)) {
						report(call, "`set "+lit+"`")
						return
					}
				}
			case "shopt":
				if restored["shopt"] || strings.Contains(text, "shopt -p") || hasFlag(call, 'q') || hasFlag(call, 'p') {
					return
				}
				// `shopt -s` with no name lists the options; it changes nothing.
				if (hasFlag(call, 's') || hasFlag(call, 'u')) && hasOperand(call) {
					report(call, "`shopt`")
				}
			case "umask":
				if len(call.Args) > 1 && !strings.Contains(text, "$(umask)") {
					report(call, "`umask`")
				}
			}
		})
		// A bare `IFS=...` statement changes the shell's IFS; one before a
		// command is scoped to that command.
		if localIFS || restored["IFS"] {
			return
		}
		syntax.Walk(decl.Body, func(node syntax.Node) bool {
			call, ok := node.(*syntax.CallExpr)
			if !ok || len(call.Args) > 0 {
				return true
			}
			for _, assign := range call.Assigns {
				if assign.Name != nil && assign.Name.Value == "IFS" {
					report(call, "`IFS=`")
				}
			}
			return true
		})
	})
}

// hasLocalDash reports whether a function declares `local -`, which restores
// the shell options when it returns.
func hasLocalDash(decl *syntax.FuncDecl) bool {
	found := false
	syntax.Walk(decl.Body, func(node syntax.Node) bool {
		clause, ok := node.(*syntax.DeclClause)
		if !ok || clause.Variant == nil || clause.Variant.Value != "local" {
			return true
		}
		for _, arg := range clause.Args {
			if arg.Name == nil && arg.Value != nil && wordLiteral(arg.Value) == "-" {
				found = true
			}
		}
		return true
	})
	return found
}

// restoredState reports which state a function puts back itself: `set` when
// every option it turns off it also turns on again (or the reverse), `shopt`
// when it both sets and unsets, and `IFS` when it saves the old value.
func restoredState(decl *syntax.FuncDecl) map[string]bool {
	on, off := map[string]bool{}, map[string]bool{}
	shoptSet, shoptUnset := false, false
	out := map[string]bool{}
	allCalls(decl.Body, func(call *syntax.CallExpr, name string) {
		switch name {
		case "set":
			words := setOptionWords(call)
			for i := 0; i < len(words); i++ {
				lit := wordLiteral(words[i])
				if len(lit) < 2 || (lit[0] != '-' && lit[0] != '+') {
					continue
				}
				state := on
				if lit[0] == '+' {
					state = off
				}
				for _, letter := range []byte(lit[1:]) {
					option := optionNames[letter]
					if option == "" {
						option = string(letter)
					}
					// `-o NAME` names the option in the next word.
					if letter == 'o' && i+1 < len(words) {
						i++
						option = wordLiteral(words[i])
					}
					state[option] = true
				}
			}
		case "shopt":
			shoptSet = shoptSet || hasFlag(call, 's')
			shoptUnset = shoptUnset || hasFlag(call, 'u')
		}
	})
	paired := len(on)+len(off) > 0
	for option := range on {
		if scopedNames[option] && !off[option] {
			paired = false
		}
	}
	for option := range off {
		if scopedNames[option] && !on[option] {
			paired = false
		}
	}
	out["set"] = paired
	out["shopt"] = shoptSet && shoptUnset
	syntax.Walk(decl.Body, func(node syntax.Node) bool {
		if assign, ok := node.(*syntax.Assign); ok && assign.Value != nil && strings.Contains(wordSource(assign.Value), "IFS") {
			out["IFS"] = true
		}
		return true
	})
	return out
}

// setOptionWords returns the words of a `set` call up to `--` or `-`, after
// which every word is a positional parameter: `set -- -a "$@"` sets `$1` to
// `-a` and turns no option on.
func setOptionWords(call *syntax.CallExpr) []*syntax.Word {
	args := call.Args[1:]
	for i, arg := range args {
		if lit := wordLiteral(arg); lit == "--" || lit == "-" {
			return args[:i]
		}
	}
	return args
}

// savesOptions reports whether a function saves every shell option with a
// bare `set +o`, as in `saved="$(set +o)"; ...; eval "${saved}"`, which
// prints the commands that put them all back.
func savesOptions(decl *syntax.FuncDecl) bool {
	found := false
	allCalls(decl.Body, func(call *syntax.CallExpr, name string) {
		if name == "set" && len(call.Args) == 2 && wordLiteral(call.Args[1]) == "+o" {
			found = true
		}
	})
	return found
}

// hasOperand reports whether a call has a word that is not an option.
func hasOperand(call *syntax.CallExpr) bool {
	for _, arg := range call.Args[1:] {
		if lit := wordLiteral(arg); !strings.HasPrefix(lit, "-") {
			return true
		}
	}
	return false
}
