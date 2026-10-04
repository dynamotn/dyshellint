package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionSignalHandlers = "Calling Commands > Signal Handlers"

func init() {
	register(Rule{
		Code:     "BSG055",
		Section:  sectionSignalHandlers,
		Severity: lint.SeverityWarning,
		Doc:      "Do not replace or clear the caller's EXIT, INT or TERM handler from a library",
		Check:    checkLibraryTraps,
	})
}

// callerSignals are the signals a script installs its own cleanup on.
var callerSignals = map[string]bool{
	"EXIT": true, "INT": true, "TERM": true, "HUP": true, "QUIT": true,
	"0": true, "1": true, "2": true, "3": true, "15": true,
}

// saveIdiom matches a function that saves the handlers it is about to change,
// so it can put them back.
var saveIdiom = []string{"trap -p", "traps_save", "dybatpho::trap"}

func checkLibraryTraps(f *File, r *Reporter) {
	if f.Role != RoleLibrary {
		return
	}
	saves := map[*syntax.FuncDecl]bool{}
	eachFunc(f, func(decl *syntax.FuncDecl) {
		text := f.Text(decl.Pos(), decl.End())
		for _, idiom := range saveIdiom {
			if strings.Contains(text, idiom) {
				saves[decl] = true
			}
		}
	})
	up := parents(f.Syntax)
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "trap" || len(call.Args) < 3 {
			return
		}
		first := wordLiteral(call.Args[1])
		if first == "-p" || first == "-l" {
			return
		}
		signal := ""
		for _, arg := range call.Args[2:] {
			if s := strings.TrimPrefix(strings.ToUpper(wordLiteral(arg)), "SIG"); callerSignals[s] {
				signal = s
				break
			}
		}
		if signal == "" {
			return
		}
		if inSubshell(call, up) {
			// A subshell's handlers end with it; the caller's are untouched.
			return
		}
		if decl := enclosingFunc(call, up); decl != nil && saves[decl] {
			return
		}
		what := "replaces"
		if first == "-" {
			what = "clears"
		}
		r.At(call.Pos(), "a library that %s the %s handler throws away the one the calling script installed; compose with `dybatpho::trap`, or save it with `trap -p` and put it back",
			what, signal)
	})
}

// inSubshell reports whether a node runs in a subshell of the function or file
// it sits in: `( ... )`, `$(...)` or `<(...)`.
func inSubshell(node syntax.Node, up map[syntax.Node]syntax.Node) bool {
	for n := up[node]; n != nil; n = up[n] {
		switch n.(type) {
		case *syntax.Subshell, *syntax.CmdSubst, *syntax.ProcSubst:
			return true
		case *syntax.FuncDecl:
			return false
		}
	}
	return false
}
