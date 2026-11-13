package rules

import (
	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionCurrentUser = "Environment > Current User"

func init() {
	register(
		Rule{
			Code:     "BSG136",
			Section:  sectionCurrentUser,
			Severity: lint.SeverityWarning,
			Doc:      "Check for root with `(( EUID == 0 ))`, not by comparing `$USER`, `$(whoami)` or `$(id -u)` as text",
			Check:    checkRootByName,
		},
		Rule{
			Code:     "BSG137",
			Section:  sectionCurrentUser,
			Severity: lint.SeverityWarning,
			Doc:      "Find a home directory with `$HOME` or `getent passwd`, not by building `/home/${user}`",
			Check:    checkBuiltHome,
		},
	)
}

// userProbe tells what a word asks about the user: "name" for `$USER`,
// `$LOGNAME`, `$(whoami)` and `$(id -un)`, "uid" for `$(id -u)`, and the
// empty string for anything else.
func userProbe(word *syntax.Word) string {
	if word == nil || len(word.Parts) != 1 {
		return ""
	}
	part := word.Parts[0]
	if dq, ok := part.(*syntax.DblQuoted); ok {
		if len(dq.Parts) != 1 {
			return ""
		}
		part = dq.Parts[0]
	}
	switch p := part.(type) {
	case *syntax.ParamExp:
		if p.Param != nil && (p.Param.Value == "USER" || p.Param.Value == "LOGNAME") && p.Index == nil && p.Repl == nil && !p.Length {
			return "name"
		}
	case *syntax.CmdSubst:
		if len(p.Stmts) != 1 {
			return ""
		}
		call, ok := p.Stmts[0].Cmd.(*syntax.CallExpr)
		if !ok {
			return ""
		}
		switch callName(call) {
		case "whoami":
			return "name"
		case "id":
			if !hasFlag(call, 'u') || hasFlag(call, 'r') || len(call.Args) != len(callFlags(call))+1 {
				// `id -u someone` asks about another user.
				return ""
			}
			if hasFlag(call, 'n') {
				return "name"
			}
			return "uid"
		}
	}
	return ""
}

// callFlags returns the option words of a call.
func callFlags(call *syntax.CallExpr) []*syntax.Word {
	var out []*syntax.Word
	for _, arg := range call.Args[1:] {
		if lit := wordLiteral(arg); len(lit) > 1 && lit[0] == '-' {
			out = append(out, arg)
		}
	}
	return out
}

// comparesToRoot reports whether one side of a comparison asks about the user
// and the other is the answer that means root.
func comparesToRoot(x, y *syntax.Word) bool {
	for _, pair := range [][2]*syntax.Word{{x, y}, {y, x}} {
		switch userProbe(pair[0]) {
		case "name":
			if wordLiteral(pair[1]) == "root" {
				return true
			}
		case "uid":
			if wordLiteral(pair[1]) == "0" {
				return true
			}
		}
	}
	return false
}

// checkRootByName reports a root check made by comparing a user name or a
// `$(id -u)` as text. `USER` and `LOGNAME` are variables anyone can set, and
// cron, `sudo` and containers leave them unset or stale; the name `root` is
// only a convention. Bash already knows the effective uid in `EUID`.
func checkRootByName(f *File, r *Reporter) {
	report := func(pos syntax.Pos) {
		reportWithDybatpho(f, r, pos, "`dybatpho::is_root` asks the same question", "this checks for root through a name or a command's text: `USER` can be set to anything and is often missing under cron and in containers; test `(( EUID == 0 ))`")
	}
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.BinaryTest:
			x, okX := n.X.(*syntax.Word)
			y, okY := n.Y.(*syntax.Word)
			switch n.Op {
			case syntax.TsMatch, syntax.TsMatchShort, syntax.TsNoMatch, syntax.TsEql, syntax.TsNeq:
				if okX && okY && comparesToRoot(x, y) {
					report(n.Pos())
				}
			}
		case *syntax.BinaryArithm:
			x, okX := n.X.(*syntax.Word)
			y, okY := n.Y.(*syntax.Word)
			if (n.Op == syntax.Eql || n.Op == syntax.Neq) && okX && okY && comparesToRoot(x, y) {
				report(n.Pos())
			}
		case *syntax.CallExpr:
			if name := callName(n); name != "[" && name != "test" {
				return true
			}
			args := n.Args[1:]
			for i := 0; i+2 < len(args); i++ {
				switch wordLiteral(args[i+1]) {
				case "=", "==", "!=", "-eq", "-ne":
					if comparesToRoot(args[i], args[i+2]) {
						report(n.Pos())
						return true
					}
				}
			}
		}
		return true
	})
}

// checkBuiltHome reports a path that puts a user name under `/home/` or
// `/Users/`. Home directories live under `/Users` on macOS, `/root` for root,
// `/var/lib/<service>` for many service accounts, and anywhere an
// administrator chose: the password database is the only one that knows.
func checkBuiltHome(f *File, r *Reporter) {
	up := f.parents()
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		word, ok := node.(*syntax.Word)
		if !ok {
			return true
		}
		// `dscl . -read /Users/<name>` names a directory service record, which
		// is how macOS answers the question in the first place.
		if call, ok := up[word].(*syntax.CallExpr); ok && callName(call) == "dscl" {
			return true
		}
		parts := word.Parts
		if len(parts) > 0 {
			if dq, ok := parts[0].(*syntax.DblQuoted); ok && len(parts) == 1 {
				parts = dq.Parts
			}
		}
		if len(parts) < 2 {
			return true
		}
		lit, ok := parts[0].(*syntax.Lit)
		if !ok || (lit.Value != "/home/" && lit.Value != "/Users/") {
			return true
		}
		switch parts[1].(type) {
		case *syntax.ParamExp, *syntax.CmdSubst:
		default:
			return true
		}
		r.At(word.Pos(), "%s guesses where a home directory is, which is `/Users` on macOS, `/root` for root and anywhere for a service account; read it with `getent passwd \"${user}\" | cut -d: -f6`, or use `${HOME}` for the current user",
			wordSource(word))
		return true
	})
}
