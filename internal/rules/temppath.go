package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG082",
		Section:  sectionTempFiles,
		Severity: lint.SeverityError,
		Doc:      "Do not create a file at a path built from `$$`, `$BASHPID` or `$RANDOM` with `>`, `touch` or `mkdir -p`",
		Check:    checkGuessablePathWrite,
	})
}

// guessableParams are the expansions anyone on the machine can guess or try.
var guessableParams = map[string]bool{"$": true, "BASHPID": true, "RANDOM": true, "PPID": true}

func checkGuessablePathWrite(f *File, r *Reporter) {
	report := func(word *syntax.Word, how string) {
		// A path under /tmp is BSG045's, whatever creates it.
		if word == nil || !hasGuessablePart(word) || strings.Contains(wordSource(word), "/tmp/") {
			return
		}
		r.At(word.Pos(), "%s is guessable, and %s follows a link planted at it or reuses what is there; create it with `mktemp`, `dybatpho::create_temp`, or under `set -C`",
			wordSource(word), how)
	}
	eachFunc(f, func(decl *syntax.FuncDecl) {
		if setsNoclobber(decl) {
			return
		}
		syntax.Walk(decl.Body, func(node syntax.Node) bool {
			stmt, ok := node.(*syntax.Stmt)
			if !ok {
				return true
			}
			for _, redir := range stmt.Redirs {
				switch redir.Op {
				case syntax.RdrOut, syntax.AppOut, syntax.RdrAll, syntax.AppAll:
					report(redir.Word, "`"+redir.Op.String()+"`")
				}
			}
			call, ok := stmt.Cmd.(*syntax.CallExpr)
			if !ok {
				return true
			}
			switch callName(call) {
			case "touch":
				for _, arg := range call.Args[1:] {
					report(arg, "`touch`")
				}
			case "mkdir":
				// Without `-p`, mkdir fails on a name that exists, which is the
				// exclusive creation a temporary directory needs.
				if hasFlag(call, 'p') {
					for _, arg := range call.Args[1:] {
						report(arg, "`mkdir -p`")
					}
				}
			}
			return true
		})
	})
}

func hasGuessablePart(word *syntax.Word) bool {
	found := false
	syntax.Walk(word, func(node syntax.Node) bool {
		if pe, ok := node.(*syntax.ParamExp); ok && pe.Param != nil && guessableParams[pe.Param.Value] {
			found = true
		}
		return !found
	})
	return found
}

// hasFlag reports whether a call passes a short option letter, alone or in a
// cluster such as `-pv`.
func hasFlag(call *syntax.CallExpr, letter byte) bool {
	for _, arg := range call.Args[1:] {
		lit := wordLiteral(arg)
		if strings.HasPrefix(lit, "-") && !strings.HasPrefix(lit, "--") && strings.IndexByte(lit[1:], letter) >= 0 {
			return true
		}
	}
	return false
}

// setsNoclobber reports whether a function turns on `set -C`, under which `>`
// refuses a name that already exists.
func setsNoclobber(decl *syntax.FuncDecl) bool {
	found := false
	allCalls(decl.Body, func(call *syntax.CallExpr, name string) {
		if name != "set" {
			return
		}
		for _, arg := range call.Args[1:] {
			lit := wordLiteral(arg)
			if lit == "noclobber" || strings.HasPrefix(lit, "-") && strings.Contains(lit, "C") {
				found = true
			}
		}
	})
	return found
}
