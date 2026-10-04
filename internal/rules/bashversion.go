package rules

import (
	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionBashVersion = "Background > Bash Version"

func init() {
	register(Rule{
		Code:     "BSG099",
		Section:  sectionBashVersion,
		Severity: lint.SeverityWarning,
		Doc:      "Check `BASH_VERSINFO` in an entrypoint that uses a feature newer than Bash 4.2",
		Check:    checkBashVersionGuard,
	})
}

// checkBashVersionGuard reports the first use of a Bash 4.3 or 4.4 feature in
// an entrypoint that never looks at `BASH_VERSINFO`. On the Bash 3.2 macOS
// still ships, such a script fails far from the cause, with `invalid option`
// or `command not found`. dybatpho checks the version itself when sourced.
func checkBashVersionGuard(f *File, r *Reporter) {
	if f.Role != RoleEntrypoint || f.UsesDybatpho || readsParam(f, "BASH_VERSINFO") {
		return
	}
	var at syntax.Node
	var feature string
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		if at != nil {
			return false
		}
		switch n := node.(type) {
		case *syntax.DeclClause:
			if n.Variant == nil {
				return true
			}
			flags := declFlags(n)
			switch {
			case containsFlag(flags, 'n'):
				at, feature = n, "a nameref (`-n`, Bash 4.3)"
			case n.Variant.Value == "local" && hasNakedDash(n):
				at, feature = n, "`local -` (Bash 4.4)"
			}
		case *syntax.CallExpr:
			switch name := wordLiteral(firstWord(n)); name {
			case "mapfile", "readarray":
				if hasFlag(n, 'd') {
					at, feature = n, "`"+name+" -d` (Bash 4.4)"
				}
			case "wait":
				if hasFlag(n, 'n') {
					at, feature = n, "`wait -n` (Bash 4.3)"
				}
			case "shopt":
				if hasLong(n, "inherit_errexit") {
					at, feature = n, "`inherit_errexit` (Bash 4.4)"
				}
			}
		case *syntax.ParamExp:
			if n.Exp != nil && n.Exp.Op == syntax.OtherParamOps {
				at, feature = n, "`${var@...}` (Bash 4.4)"
			}
		}
		return at == nil
	})
	if at == nil {
		return
	}
	r.At(at.Pos(), "%s does not exist on the Bash 3.2 that macOS ships, where the script fails later with a confusing error; check `BASH_VERSINFO` at the top and stop with a message naming the version found", feature)
}

// readsParam reports whether a file expands a parameter anywhere.
func readsParam(f *File, name string) bool {
	found := false
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		if pe, ok := node.(*syntax.ParamExp); ok && pe.Param != nil && pe.Param.Value == name {
			found = true
		}
		if arith, ok := node.(*syntax.Lit); ok && arith.Value == name {
			found = true
		}
		return !found
	})
	return found
}

// containsFlag reports whether a run of short options holds a letter.
func containsFlag(flags string, letter byte) bool {
	for i := 0; i < len(flags); i++ {
		if flags[i] == letter && i > 0 {
			return true
		}
	}
	return false
}

// hasNakedDash reports whether a declaration carries a lone `-`, as in
// `local -`.
func hasNakedDash(decl *syntax.DeclClause) bool {
	for _, arg := range decl.Args {
		if arg.Name == nil && arg.Naked && arg.Value != nil && wordLiteral(arg.Value) == "-" {
			return true
		}
	}
	return false
}

// firstWord returns the command word of a call, or nil for an assignment-only
// call.
func firstWord(call *syntax.CallExpr) *syntax.Word {
	if len(call.Args) == 0 {
		return nil
	}
	return call.Args[0]
}
