package rules

import (
	"regexp"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG088",
		Section:  sectionArithmetic,
		Severity: lint.SeverityError,
		Doc:      "Do not run `((x++))` as a statement under `set -e`; write `((x += 1))` or `x=$((x + 1))`",
		Check:    checkPostIncrementStatement,
	})
}

// errexitPattern matches a file that turns on `set -e`.
var errexitPattern = regexp.MustCompile(`(?m)^\s*set\s+(-[a-zA-Z]*e|-o\s+errexit)`)

// stopsOnError reports whether a file stops on a failing command.
func stopsOnError(f *File) bool {
	return f.UsesDybatpho || errexitPattern.Match(f.Src)
}

func checkPostIncrementStatement(f *File, r *Reporter) {
	if !stopsOnError(f) {
		return
	}
	up := f.parents()
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		stmt, ok := node.(*syntax.Stmt)
		if !ok || stmt.Negated {
			return true
		}
		cmd, ok := stmt.Cmd.(*syntax.ArithmCmd)
		if list := listOf(stmt, up); !ok || !hasPostStep(cmd.X) || list.Negated || statusUsed(list, up) {
			return true
		}
		r.At(cmd.Pos(), "`((x++))` returns the old value, so it fails when that value is 0 and `set -e` ends the script; write `((x += 1))` or `x=$((x + 1))`")
		return true
	})
}

// hasPostStep reports whether an arithmetic expression ends in a postfix
// increment or decrement, alone or as the last of a comma list.
func hasPostStep(expr syntax.ArithmExpr) bool {
	switch e := expr.(type) {
	case *syntax.UnaryArithm:
		return e.Post && (e.Op == syntax.Inc || e.Op == syntax.Dec)
	case *syntax.BinaryArithm:
		if e.Op == syntax.Comma {
			return hasPostStep(e.Y)
		}
	case *syntax.ParenArithm:
		return hasPostStep(e.X)
	}
	return false
}

// listOf climbs from a statement that ends an `&&` or `||` list to the
// statement of the whole list: `set -e` ignores the status of every command
// of a list but the last, whose status becomes the list's own, as in
// `[[ -n ${1:-} ]] && ((count++))`.
func listOf(stmt *syntax.Stmt, up map[syntax.Node]syntax.Node) *syntax.Stmt {
	for {
		bin, ok := up[stmt].(*syntax.BinaryCmd)
		if !ok || bin.Y != stmt || bin.Op != syntax.AndStmt && bin.Op != syntax.OrStmt {
			return stmt
		}
		outer, ok := up[bin].(*syntax.Stmt)
		if !ok {
			return stmt
		}
		stmt = outer
	}
}

// statusUsed reports whether a statement's status is read rather than left to
// `set -e`: it is a condition, or part of an `&&` or `||` list.
func statusUsed(stmt *syntax.Stmt, up map[syntax.Node]syntax.Node) bool {
	switch p := up[stmt].(type) {
	case *syntax.BinaryCmd:
		return p.Op == syntax.AndStmt || p.Op == syntax.OrStmt
	case *syntax.IfClause:
		return inStmts(p.Cond, stmt)
	case *syntax.WhileClause:
		return inStmts(p.Cond, stmt)
	}
	return false
}
