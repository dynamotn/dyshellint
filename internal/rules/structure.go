package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const (
	sectionFunctionDecl  = "Formatting > Function Declaration"
	sectionCommonScripts = "Environment > Common Function Scripts"
	sectionStdio         = "Environment > STDOUT and STDERR"
	sectionSuid          = "Shell Files and Interpreter Invocation > SUID/SGID"
	sectionFileExt       = "Shell Files and Interpreter Invocation > File Extensions"
	sectionDebugMode     = "Environment > Debug and Dry-run Mode"
)

func init() {
	register(
		Rule{
			Code:     "BSG030",
			Section:  sectionWhichShell,
			Severity: lint.SeverityError,
			Doc:      "Start executable files with `#!/usr/bin/env bash`",
			Check:    checkShebang,
		},
		Rule{
			Code:     "BSG031",
			Section:  sectionWhichShell,
			Severity: lint.SeverityError,
			Doc:      "Set `set -euo pipefail`, unless dybatpho is sourced",
			Check:    checkStrictMode,
		},
		Rule{
			Code:     "BSG032",
			Section:  sectionCommonScripts,
			Severity: lint.SeverityError,
			Doc:      "Source libraries with `.`, not with `source`",
			Check:    checkSourceBuiltin,
		},
		Rule{
			Code:     "BSG033",
			Section:  sectionFunctionDecl,
			Severity: lint.SeverityError,
			Doc:      "Do not place executable code between function declarations",
			Check:    checkCodeBetweenFunctions,
		},
		Rule{
			Code:     "BSG034",
			Section:  sectionDebugMode,
			Severity: lint.SeverityError,
			Doc:      "Trace with `dybatpho::start_trace`, not with an inline `set -x`",
			Check:    checkInlineTrace,
		},
		Rule{
			Code:     "BSG035",
			Section:  sectionSuid,
			Severity: lint.SeverityWarning,
			Doc:      "Do not call `sudo` from inside a script",
			Check:    checkSudo,
		},
		Rule{
			Code:     "BSG036",
			Section:  sectionStdio,
			Severity: lint.SeverityError,
			Doc:      "Send error and fatal messages to STDERR",
			Check:    checkErrorToStderr,
		},
		Rule{
			Code:     "BSG037",
			Section:  sectionFileExt,
			Severity: lint.SeverityError,
			Doc:      "Keep library scripts non-executable and entrypoints executable",
			Check:    checkExecutableBit,
		},
	)
}

func checkShebang(f *File, r *Reporter) {
	if f.Shebang == "" {
		return
	}
	shebang := strings.TrimSpace(f.Shebang)
	switch {
	case shebang == "#!/usr/bin/env bash":
	case strings.HasPrefix(shebang, "#!/usr/bin/env bash "):
		r.AtLine(1, "the shebang carries options; set them with `set -euo pipefail` instead, they are dropped when the script is run as `bash ./script.sh`")
	default:
		r.AtLine(1, "%q is not `#!/usr/bin/env bash`", shebang)
	}
}

// strictOptions are the options the guide asks for, by their long name.
var strictOptions = []string{"errexit", "nounset", "pipefail"}

// shortOptions maps the letters of `set` that turn on a strict option to its
// long name.
var shortOptions = map[byte]string{'e': "errexit", 'u': "nounset"}

func checkStrictMode(f *File, r *Reporter) {
	if f.Role != RoleEntrypoint || f.UsesDybatpho {
		return
	}
	on := map[string]bool{}
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name == "set" {
			setOptions(call, on)
		}
	})
	for _, option := range strictOptions {
		if !on[option] {
			r.AtLine(1, "no `set -euo pipefail`; add it under the shebang, or source dybatpho, which sets it")
			return
		}
	}
}

// setOptions adds the long name of every option a `set` call turns on to on.
// The letters of a cluster come in any order, as in `-Eeuo pipefail`, and each
// `o` among them takes the next word as the name of an option, the way
// `-o errexit -o nounset` spells them out.
func setOptions(call *syntax.CallExpr, on map[string]bool) {
	args := call.Args[1:]
	for i := 0; i < len(args); i++ {
		lit := wordLiteral(args[i])
		if len(lit) < 2 || lit[0] != '-' || lit == "--" {
			// The rest are positional parameters.
			return
		}
		for _, letter := range []byte(lit[1:]) {
			if letter == 'o' {
				if i+1 < len(args) {
					i++
					on[wordLiteral(args[i])] = true
				}
				continue
			}
			if option := shortOptions[letter]; option != "" {
				on[option] = true
			}
		}
	}
}

func checkSourceBuiltin(f *File, r *Reporter) {
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "source" {
			return
		}
		r.At(call.Pos(), "`source` is a bashism; use `.`, which is POSIX")
	})
}

func checkCodeBetweenFunctions(f *File, r *Reporter) {
	stmts := f.Syntax.Stmts
	firstFunc := -1
	for i, stmt := range stmts {
		if _, ok := stmt.Cmd.(*syntax.FuncDecl); ok {
			firstFunc = i
			break
		}
	}
	if firstFunc < 0 {
		return
	}
	// Everything after the first declaration has to be another declaration,
	// except the single call on the last line that starts the script.
	for i := firstFunc + 1; i < len(stmts)-1; i++ {
		if _, ok := stmts[i].Cmd.(*syntax.FuncDecl); ok {
			continue
		}
		r.At(stmts[i].Pos(), "this runs when the file is sourced; move it above the first function, or into the entrypoint function")
	}
}

func checkInlineTrace(f *File, r *Reporter) {
	if !f.UsesDybatpho {
		return
	}
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "set" {
			return
		}
		for _, arg := range call.Args[1:] {
			if lit := wordLiteral(arg); strings.HasPrefix(lit, "-x") || lit == "-o" {
				if strings.Contains(lit, "x") {
					r.At(call.Pos(), "`set -x` cannot be turned off by the caller; use `dybatpho::start_trace` and `dybatpho::end_trace`")
					return
				}
			}
		}
	})
}

func checkSudo(f *File, r *Reporter) {
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "sudo" && name != "su" {
			return
		}
		r.At(call.Pos(), "%q inside a script; elevate at the call site instead, and never in CI", name)
	})
}

// severityWord matches the messages that belong on STDERR.
var severityWord = regexp.MustCompile(`^(?i)(error|fatal|fail(ed|ure)?|warning)\b`)

func checkErrorToStderr(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		stmt, ok := node.(*syntax.Stmt)
		if !ok {
			return true
		}
		call, ok := stmt.Cmd.(*syntax.CallExpr)
		if !ok || len(call.Args) < 2 {
			return true
		}
		switch wordLiteral(call.Args[0]) {
		case "echo", "printf":
		default:
			return true
		}
		message := wordLiteral(call.Args[len(call.Args)-1])
		if !severityWord.MatchString(strings.TrimSpace(message)) {
			return true
		}
		for _, redirect := range stmt.Redirs {
			if word := wordLiteral(redirect.Word); word == "2" || strings.Contains(word, "&2") {
				return true
			}
			if redirect.N != nil && redirect.N.Value == "2" {
				return true
			}
		}
		r.At(stmt.Pos(), "this message goes to STDOUT; append `>&2`, or use `dybatpho::error`")
		return true
	})
}

func checkExecutableBit(f *File, r *Reporter) {
	if !f.ModeKnown {
		return
	}
	switch f.Role {
	case RoleLibrary:
		if f.Executable {
			r.AtLine(1, "a library script is sourced, never run; `chmod -x` it")
		}
	case RoleEntrypoint:
		if !f.Executable {
			r.AtLine(1, "an entrypoint script should carry the executable bit; `chmod +x` it")
		}
	}
}

// eachCall walks every simple command whose name is a plain literal.
func eachCall(f *File, fn func(call *syntax.CallExpr, name string)) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		call, ok := node.(*syntax.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		if name := wordLiteral(call.Args[0]); name != "" {
			fn(call, name)
		}
		return true
	})
}
