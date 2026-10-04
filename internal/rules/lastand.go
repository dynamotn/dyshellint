package rules

import (
	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(
		Rule{
			Code:     "BSG110",
			Section:  sectionReturnVals,
			Severity: lint.SeverityError,
			Doc:      "Do not end a function or a script with `test && action`: a false test becomes its status",
			Check:    checkTrailingAnd,
		},
		Rule{
			Code:     "BSG120",
			Section:  sectionReturnVals,
			Severity: lint.SeverityWarning,
			Doc:      "Silence the one command that may fail, not a whole block with `2> /dev/null`",
			Check:    checkBlockSilenced,
		},
	)
}

// checkTrailingAnd reports a function, or a script, whose last statement is
// `test && action`. When the test is false the `&&` list fails, that status
// becomes the function's, and `set -e` stops the caller of a function that did
// everything it was asked.
func checkTrailingAnd(f *File, r *Reporter) {
	if !stopsOnError(f) {
		return
	}
	report := func(stmts []*syntax.Stmt, what string) {
		if len(stmts) == 0 {
			return
		}
		last := stmts[len(stmts)-1]
		bin, ok := last.Cmd.(*syntax.BinaryCmd)
		// `test && test` is a predicate on purpose: its status is the answer.
		if !ok || last.Negated || bin.Op != syntax.AndStmt || !isTest(bin.X) || isTest(bin.Y) {
			return
		}
		r.At(last.Pos(), "%s ends with `test && action`, so a false test becomes its status and `set -e` stops the caller; write `if test; then action; fi`, or `test || return 0` first", what)
	}
	eachFunc(f, func(decl *syntax.FuncDecl) {
		if block, ok := decl.Body.Cmd.(*syntax.Block); ok {
			report(block.Stmts, "`"+decl.Name.Value+"`")
		}
	})
	if f.Role == RoleEntrypoint {
		report(f.Syntax.Stmts, "the script")
	}
}

// isTest reports whether a statement is a test whose false answer is not a
// failure: `[[ ]]`, `(( ))`, `[`, `test`, or the left side of such a chain.
func isTest(stmt *syntax.Stmt) bool {
	switch cmd := stmt.Cmd.(type) {
	case *syntax.TestClause, *syntax.ArithmCmd:
		return true
	case *syntax.BinaryCmd:
		return isTest(cmd.X)
	case *syntax.CallExpr:
		switch callName(cmd) {
		case "[", "test", "dybatpho::is", "dybatpho::string_is_blank":
			return true
		}
	}
	return false
}

// checkBlockSilenced reports a block, loop, conditional or subshell whose
// standard error goes to /dev/null as a whole: every error inside is lost,
// not only the one that was expected.
func checkBlockSilenced(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		stmt, ok := node.(*syntax.Stmt)
		if !ok {
			return true
		}
		// A block of one command silences that command alone, which is the
		// point; a loop or a conditional always runs more than its test.
		switch cmd := stmt.Cmd.(type) {
		case *syntax.Block:
			if workingStmts(cmd.Stmts) < 2 {
				return true
			}
		case *syntax.Subshell:
			if workingStmts(cmd.Stmts) < 2 {
				return true
			}
		case *syntax.WhileClause, *syntax.ForClause, *syntax.IfClause, *syntax.CaseClause:
		default:
			return true
		}
		for _, redir := range stmt.Redirs {
			if silencesStderr(redir) {
				r.At(redir.OpPos, "this hides the errors of every command in the block, not only the expected one; redirect the one command that may fail, and say why")
			}
		}
		return true
	})
}

// silencesStderr reports whether a redirection sends standard error to
// /dev/null, alone or with standard output.
func silencesStderr(redir *syntax.Redirect) bool {
	if wordLiteral(redir.Word) != "/dev/null" {
		return false
	}
	switch redir.Op {
	case syntax.RdrAll, syntax.AppAll:
		return true
	case syntax.RdrOut, syntax.AppOut:
		return redir.N != nil && redir.N.Value == "2"
	}
	return false
}

// workingStmts counts the statements of a block that do the work, leaving out
// the `set` and `shopt` lines that only prepare it, as in the exclusive
// creation `( set -C; : > "${name}" ) 2> /dev/null`.
func workingStmts(stmts []*syntax.Stmt) int {
	count := 0
	for _, stmt := range stmts {
		if call, ok := stmt.Cmd.(*syntax.CallExpr); ok {
			if name := callName(call); name == "set" || name == "shopt" {
				continue
			}
		}
		count++
	}
	return count
}
