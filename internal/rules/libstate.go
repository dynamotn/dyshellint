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
// runs.
const scopedOptions = "eCfuxa"

func checkLibraryShellState(f *File, r *Reporter) {
	if f.Role != RoleLibrary {
		return
	}
	up := parents(f.Syntax)
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
		allCalls(decl.Body, func(call *syntax.CallExpr, name string) {
			switch name {
			case "set":
				if localOptions || strings.Contains(text, "$-") || strings.Contains(text, "set +o\n") {
					return
				}
				for _, arg := range call.Args[1:] {
					lit := wordLiteral(arg)
					if lit == "-o" || lit == "+o" || (len(lit) > 1 && (lit[0] == '-' || lit[0] == '+') && strings.ContainsAny(lit[1:], scopedOptions)) {
						report(call, "`set "+lit+"`")
						return
					}
				}
			case "shopt":
				if strings.Contains(text, "shopt -p") || hasFlag(call, 'q') || hasFlag(call, 'p') {
					return
				}
				if hasFlag(call, 's') || hasFlag(call, 'u') {
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
		if localIFS {
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
