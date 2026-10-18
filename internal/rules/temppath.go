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
	check := func(scope syntax.Node) {
		if setsNoclobber(scope) {
			return
		}
		syntax.Walk(scope, func(node syntax.Node) bool {
			// A function is checked on its own, under its own `set -C`.
			if decl, ok := node.(*syntax.FuncDecl); ok && decl != scope {
				return false
			}
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
	}
	eachFunc(f, func(decl *syntax.FuncDecl) { check(decl) })
	// The top level is a scope of its own: the code outside every function.
	check(f.Syntax)
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
		if !strings.HasPrefix(lit, "-") || strings.HasPrefix(lit, "--") {
			continue
		}
		cluster := lit[1:]
		if strings.IndexFunc(cluster, notOptionLetter) >= 0 {
			// Only letters make a cluster: `-w'%{http_code}'` or `-o/dev/null`
			// is one option with its value joined on, and the letters of the
			// value are no options.
			cluster = cluster[:1]
		}
		if strings.IndexByte(cluster, letter) >= 0 {
			return true
		}
	}
	return false
}

func notOptionLetter(r rune) bool {
	return !('a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || '0' <= r && r <= '9')
}

// setsNoclobber reports whether a function, or the top level of a file,
// turns on `set -C`, under which `>` refuses a name that already exists.
func setsNoclobber(scope syntax.Node) bool {
	if decl, ok := scope.(*syntax.FuncDecl); ok {
		scope = decl.Body
	}
	found := false
	allCalls(scope, func(call *syntax.CallExpr, name string) {
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
