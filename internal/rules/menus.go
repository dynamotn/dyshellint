package rules

import (
	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG161",
		Section:  "Environment > Menus and Choices",
		Severity: lint.SeverityWarning,
		Doc:      "Check for a terminal before a menu, and take the same answer from a flag or an environment variable without one",
		Check:    checkUnguardedMenu,
	})
}

// menuTools are the programs that draw a menu or a question on a terminal.
var menuTools = map[string]bool{"fzf": true, "whiptail": true, "dialog": true, "gum": true}

// gumAsks are the `gum` subcommands that wait for the user; `gum style`,
// `gum format` or `gum spin` only print.
var gumAsks = map[string]bool{"choose": true, "filter": true, "confirm": true, "input": true, "write": true, "file": true, "pager": true}

// checkUnguardedMenu reports a `select` loop, or a menu program, in a file
// that never tests for a terminal. The test is looked for with the pattern of
// BSG098 anywhere in the file, since the menu usually sits in a helper that
// its caller guards.
func checkUnguardedMenu(f *File, r *Reporter) {
	if ttyCheck.Match(f.Src) {
		return
	}
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		if loop, ok := node.(*syntax.ForClause); ok && loop.Select {
			reportWithDybatpho(f, r, loop.Pos(), "`dybatpho::tui_menu` draws the menu on a terminal and answers with `DYBATPHO_TUI_DEFAULT` without one", "`select` ends at the first end of input, or reads the next line of a pipe meant for something else, when no one is at a terminal; test `[[ -t 0 ]]` first and take the choice from a flag or an environment variable otherwise")
		}
		return true
	})
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if !menuTools[name] {
			return
		}
		switch {
		case name == "fzf" && (hasFlag(call, 'f') || hasLong(call, "--filter")):
			// `fzf --filter` matches without asking.
			return
		case name == "gum" && (len(call.Args) < 2 || !gumAsks[wordLiteral(call.Args[1])]):
			return
		}
		reportWithDybatpho(f, r, call.Pos(), "`dybatpho::tui_menu` draws the menu on a terminal and answers with `DYBATPHO_TUI_DEFAULT` without one", "`%s` needs a terminal: in CI, cron or a pipe it fails with an unclear error or waits for input that never comes; test `[[ -t 0 ]]` first and take the choice from a flag or an environment variable otherwise", name)
	})
}
