package rules

import (
	"bytes"
	"strconv"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const (
	sectionErrexit   = "Calling Commands > Errexit in Conditions"
	sectionExitCodes = "Calling Commands > Exit Codes"
)

func init() {
	register(
		Rule{
			Code:     "BSG104",
			Section:  sectionErrexit,
			Severity: lint.SeverityWarning,
			Doc:      "Turn on `shopt -s inherit_errexit` in an entrypoint that stops on error and uses `$(...)`",
			Check:    checkInheritErrexit,
		},
		Rule{
			Code:     "BSG105",
			Section:  sectionExitCodes,
			Severity: lint.SeverityError,
			Doc:      "Exit with a status in 0-255 that the shell does not reserve: not 126, 127 or a negative number",
			Check:    checkExitStatus,
		},
	)
}

// checkInheritErrexit reports an entrypoint that relies on `set -e` and runs
// command substitutions without `inherit_errexit`: inside `$(a; b)` errexit is
// off, so a failing `a` is ignored and the substitution goes on to `b`.
// dybatpho turns on the strict mode but leaves this option to the script.
func checkInheritErrexit(f *File, r *Reporter) {
	if f.Role != RoleEntrypoint || !stopsOnError(f) || setsShopt(f, "inherit_errexit") {
		return
	}
	// A substitution of one external command fails as a whole, which
	// errexit sees; one that runs several commands, or a function of the run,
	// goes on past the first failure.
	funcs := f.project().funcs
	substitutes := false
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		cs, ok := node.(*syntax.CmdSubst)
		if !ok {
			return !substitutes
		}
		if len(cs.Stmts) > 1 {
			substitutes = true
		}
		for _, stmt := range cs.Stmts {
			for _, call := range leadingCalls(stmt) {
				if funcs[callName(call)] != nil {
					substitutes = true
				}
			}
		}
		return !substitutes
	})
	if !substitutes {
		return
	}
	line := 1
	for i, text := range f.Lines {
		if errexitPattern.MatchString(text) || dybatphoSource.MatchString(text) && bytes.Contains([]byte(text), []byte("init.sh")) {
			line = i + 1
			break
		}
	}
	r.AtLine(line, "errexit is off inside `$(...)`, so a failing command there is ignored and the substitution carries on; add `shopt -s inherit_errexit` next to `set -euo pipefail`")
}

// setsShopt reports whether a file turns on a `shopt` option.
func setsShopt(f *File, option string) bool {
	found := false
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name == "shopt" && hasFlag(call, 's') && hasLong(call, option) {
			found = true
		}
	})
	return found
}

// reservedStatus explains the statuses the shell gives a meaning of its own.
var reservedStatus = map[int]string{
	126: "the shell's \"found but not executable\"",
	127: "the shell's \"command not found\"",
}

func checkExitStatus(f *File, r *Reporter) {
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "exit" && name != "return" || len(call.Args) != 2 {
			return
		}
		lit := wordLiteral(call.Args[1])
		status, err := strconv.Atoi(lit)
		if err != nil {
			return
		}
		switch {
		case status < 0 || status > 255:
			r.At(call.Args[1].Pos(), "`%s %d` is taken modulo 256 and becomes %d; keep a status between 0 and 255", name, status, (status%256+256)%256)
		case reservedStatus[status] != "":
			r.At(call.Args[1].Pos(), "`%s %d` reads as %s, and sends the caller down the wrong branch; use another status and document it with `@exitcode`", name, status, reservedStatus[status])
		}
	})
}
