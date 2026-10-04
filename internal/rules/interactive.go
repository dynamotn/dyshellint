package rules

import (
	"regexp"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG098",
		Section:  "Environment > Interactive Input",
		Severity: lint.SeverityWarning,
		Doc:      "Check for a terminal, or give a timeout, before `read` asks the user",
		Check:    checkUnguardedPrompt,
	})
}

// promptName matches the name of a function whose job is to ask the user.
var promptName = regexp.MustCompile(`(?i)prompt|ask|confirm|menu|select|input|question`)

// ttyCheck matches what a function does to find out whether someone can answer.
var ttyCheck = regexp.MustCompile(`-t [0-2]\b|is_tty|is_interactive|interactive|\btty\b`)

func checkUnguardedPrompt(f *File, r *Reporter) {
	up := parents(f.Syntax)
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		stmt, ok := node.(*syntax.Stmt)
		if !ok {
			return true
		}
		call, ok := stmt.Cmd.(*syntax.CallExpr)
		if !ok || callName(call) != "read" || len(stmt.Redirs) > 0 || hasFlag(call, 'u') || hasFlag(call, 't') ||
			hasFlag(call, 'n') || hasFlag(call, 'N') {
			// A one-key read is the inside of a menu or a key loop, whose
			// caller has already made sure there is a terminal.
			return true
		}
		if readsStream(stmt, up) {
			return true
		}
		scope := f.Src
		if decl := enclosingFunc(call, up); decl != nil {
			// A function that is itself the prompt leaves the terminal check to
			// the code that decides to ask.
			if promptName.MatchString(ownName(f, decl.Name.Value)) {
				return true
			}
			scope = []byte(f.Text(decl.Pos(), decl.End()))
		}
		if ttyCheck.Match(scope) {
			return true
		}
		r.At(call.Pos(), "`read` waits for someone to type, which hangs in CI, cron or a pipe; check for a terminal first (`[[ -t 0 ]]` or `dybatpho::is_interactive`), or give it `-t` seconds")
		return true
	})
}

// readsStream reports whether a `read` consumes a stream rather than asking
// the user: the condition of a loop, or the end of a pipeline.
func readsStream(stmt *syntax.Stmt, up map[syntax.Node]syntax.Node) bool {
	var cur syntax.Node = stmt
	for {
		switch p := up[cur].(type) {
		case *syntax.WhileClause:
			return inStmts(p.Cond, cur)
		case *syntax.BinaryCmd:
			if p.Op == syntax.Pipe || p.Op == syntax.PipeAll {
				return syntax.Node(p.Y) == cur
			}
			// `read ... || [[ -n ... ]]` is still the condition of its loop.
			parent, ok := up[p].(*syntax.Stmt)
			if !ok {
				return false
			}
			cur = parent
			continue
		}
		return false
	}
}
