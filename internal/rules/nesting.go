package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const (
	sectionNestingDepth = "Formatting > Nesting Depth"
	sectionFunctionSize = "Formatting > Function Size"
)

// maxNesting is how many levels of `if`, loop and `case` one function may
// stack before the innermost body stops fitting in a reader's head.
const maxNesting = 3

// maxFunctionLines is the number of code lines, blank lines and comments left
// out, past which a function is likely doing more than one job. 50 lines fit
// on one screen, and about one function in twenty of dybatpho goes past it.
const maxFunctionLines = 50

func init() {
	register(
		Rule{
			Code:     "BSG142",
			Section:  sectionNestingDepth,
			Severity: lint.SeverityWarning,
			Doc:      "Keep `if`, loops and `case` nested at most 3 levels deep in a function",
			Check:    checkNestingDepth,
		},
		Rule{
			Code:     "BSG143",
			Section:  sectionFunctionSize,
			Severity: lint.SeverityWarning,
			Doc:      "Split a function longer than 50 lines of code",
			Check:    checkFunctionSize,
		},
	)
}

// checkNestingDepth reports the outermost compound command that sits deeper
// than maxNesting in a function body or at the top level, once: what is nested
// inside it is the same problem.
func checkNestingDepth(f *File, r *Reporter) {
	report := func(node syntax.Node, what string) {
		r.At(node.Pos(), "this `%s` is nested more than %d levels deep; return early from a guard clause, `continue` early in a loop, or move the inner part into a function",
			what, maxNesting)
	}
	eachFunc(f, func(decl *syntax.FuncDecl) { nestWalk(decl.Body, 0, report) })
	nestWalk(f.Syntax, 0, report)
}

// nestWalk visits the children of node, which sits depth levels deep, and
// reports a compound command that would be one level too many. The `elif`
// and `else` of an `if` stay on the level of the `if`, and a nested function
// is a scope of its own, checked by itself.
func nestWalk(node syntax.Node, depth int, report func(syntax.Node, string)) {
	syntax.Walk(node, func(n syntax.Node) bool {
		if n == node {
			return true
		}
		var what string
		switch c := n.(type) {
		case *syntax.FuncDecl:
			return false
		case *syntax.IfClause:
			what = "if"
		case *syntax.ForClause:
			what = "for"
			if c.Select {
				what = "select"
			}
		case *syntax.WhileClause:
			what = "while"
			if c.Until {
				what = "until"
			}
		case *syntax.CaseClause:
			what = "case"
		default:
			return true
		}
		if depth+1 > maxNesting && !isGuardClause(n) {
			report(n, what)
			return false
		}
		if clause, ok := n.(*syntax.IfClause); ok {
			for ; clause != nil; clause = clause.Else {
				for _, stmt := range append(append([]*syntax.Stmt{}, clause.Cond...), clause.Then...) {
					nestWalk(stmt, depth+1, report)
				}
			}
			return false
		}
		nestWalk(n, depth+1, report)
		return false
	})
}

// isGuardClause reports an `if` with no `else` that only leaves: `continue`,
// `return` and the like. It is the cure the rule asks for, not more nesting.
func isGuardClause(node syntax.Node) bool {
	clause, ok := node.(*syntax.IfClause)
	if !ok || clause.Else != nil || len(clause.Then) != 1 {
		return false
	}
	call, ok := clause.Then[0].Cmd.(*syntax.CallExpr)
	return ok && statusGuards[callName(call)]
}

// checkFunctionSize reports a function whose body holds more than
// maxFunctionLines lines of code.
func checkFunctionSize(f *File, r *Reporter) {
	eachFunc(f, func(decl *syntax.FuncDecl) {
		count := 0
		for line := int(decl.Body.Pos().Line()) + 1; line < int(decl.Body.End().Line()); line++ {
			if text := strings.TrimSpace(f.Line(line)); text != "" && !strings.HasPrefix(text, "#") {
				count++
			}
		}
		if count > maxFunctionLines {
			r.At(decl.Position, "%q has %d lines of code, more than %d; split the decision from the side effects and move each step into a function of its own",
				decl.Name.Value, count, maxFunctionLines)
		}
	})
}
