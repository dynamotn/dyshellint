package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const (
	sectionBuiltins   = "Calling Commands > Builtin Commands vs External Commands"
	sectionShellCheck = "Features and Bugs > Use ShellCheck"
)

func init() {
	register(
		Rule{
			Code:     "BSG108",
			Section:  sectionBuiltins,
			Severity: lint.SeverityWarning,
			Doc:      "Read a whole file with `$(< file)`, not `$(cat file)`",
			Check:    checkCatSubstitution,
		},
		Rule{
			Code:     "BSG109",
			Section:  sectionShellCheck,
			Severity: lint.SeverityWarning,
			Doc:      "Point ShellCheck at a sourced file with `# shellcheck source=`, rather than disabling SC1091",
			Check:    checkSourceDisabled,
		},
	)
}

func checkCatSubstitution(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		cs, ok := node.(*syntax.CmdSubst)
		if !ok || len(cs.Stmts) != 1 {
			return true
		}
		stmt := cs.Stmts[0]
		call, ok := stmt.Cmd.(*syntax.CallExpr)
		if !ok || stmt.Negated || stmt.Background || len(stmt.Redirs) > 0 || len(call.Assigns) > 0 ||
			callName(call) != "cat" || len(call.Args) != 2 || strings.HasPrefix(wordSource(call.Args[1]), "-") {
			return true
		}
		r.At(cs.Pos(), "`$(cat %s)` starts a process to read a file; `$(< %s)` does the same in the shell", wordSource(call.Args[1]), wordSource(call.Args[1]))
		return true
	})
}

// sourceDisabled matches a ShellCheck directive that disables SC1091.
var sourceDisabled = regexp.MustCompile(`^\s*#\s*shellcheck\s+disable=[^#]*\bSC1091\b`)

func checkSourceDisabled(f *File, r *Reporter) {
	for i, line := range f.Lines {
		if sourceDisabled.MatchString(line) {
			r.AtLine(i+1, "disabling SC1091 hides every sourced file from ShellCheck; name the file with `# shellcheck source=<path>`, relative to the script with `source-path=SCRIPTDIR`")
		}
	}
}
