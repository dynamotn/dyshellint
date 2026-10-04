package rules

import (
	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionValidationInSubst = "Features and Bugs > Validation in Command Substitution"

func init() {
	register(Rule{
		Code:     "BSG047",
		Section:  sectionValidationInSubst,
		Severity: lint.SeverityWarning,
		Doc:      "Do not call a function that can stop the script inside `$(...)` without checking its status",
		Check:    checkDieInSubst,
	})
}

// statusGuards are the commands that, after `||`, keep the caller from
// carrying on when the substitution failed.
var statusGuards = map[string]bool{
	"return": true, "exit": true, "dybatpho::die": true, "dybatpho::fatal": true,
	"continue": true, "break": true,
}

func checkDieInSubst(f *File, r *Reporter) {
	p := f.project()
	var up map[syntax.Node]syntax.Node
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		cs, ok := node.(*syntax.CmdSubst)
		if !ok {
			return true
		}
		callee := dieCapableCallee(p, cs)
		if callee == "" {
			return true
		}
		if up == nil {
			up = parents(f.Syntax)
		}
		stmt := enclosingStmt(cs, up)
		if stmt == nil || stopsAnyway(cs, up) {
			return true
		}
		if !carriesStatus(stmt, cs, up) {
			r.At(cs.Pos(), "%q can stop the script, but inside `$(...)` it only ends the subshell, and here its status is lost; call it in the caller's shell, returning the value through a nameref",
				callee)
			return true
		}
		if statusChecked(stmt, up) {
			return true
		}
		r.At(cs.Pos(), "%q can stop the script, but inside `$(...)` it only ends the subshell and the caller carries on; check the status right here (`|| return`), or call it in the caller's shell",
			callee)
		return true
	})
}

// dieCapableCallee returns the first command a substitution runs that can stop
// the script, or the empty string.
func dieCapableCallee(p *Project, cs *syntax.CmdSubst) string {
	for _, stmt := range cs.Stmts {
		for _, call := range leadingCalls(stmt) {
			if name := callName(call); name != "" && p.CanDie(name) {
				return name
			}
		}
	}
	return ""
}

// enclosingStmt returns the statement a substitution is part of.
func enclosingStmt(node syntax.Node, up map[syntax.Node]syntax.Node) *syntax.Stmt {
	for n := up[node]; n != nil; n = up[n] {
		if stmt, ok := n.(*syntax.Stmt); ok {
			return stmt
		}
	}
	return nil
}

// carriesStatus reports whether a statement ends with the status of the
// substitution: a plain assignment such as `x=$(f)` does, while a substitution
// used as an argument, in a test, or in `local x=$(f)` loses it.
func carriesStatus(stmt *syntax.Stmt, cs *syntax.CmdSubst, up map[syntax.Node]syntax.Node) bool {
	call, ok := stmt.Cmd.(*syntax.CallExpr)
	if !ok || len(call.Args) > 0 {
		return false
	}
	for n := up[cs]; n != nil && n != syntax.Node(stmt); n = up[n] {
		if _, ok := n.(*syntax.Assign); ok {
			return true
		}
	}
	return false
}

// statusChecked reports whether the status of a statement is read where it
// runs: the condition of an `if` or `while`, a `!`, or a `||` that returns,
// exits, stops the script or records the status.
func statusChecked(stmt *syntax.Stmt, up map[syntax.Node]syntax.Node) bool {
	if stmt.Negated {
		return true
	}
	var cur syntax.Node = stmt
	for {
		parent := up[cur]
		switch p := parent.(type) {
		case *syntax.BinaryCmd:
			if syntax.Node(p.X) != cur {
				return false
			}
			switch p.Op {
			case syntax.OrStmt:
				return guardsStatus(p.Y)
			case syntax.AndStmt:
				cur = upStmt(p, up)
				if cur == nil {
					return false
				}
				continue
			}
			return false
		case *syntax.IfClause:
			return inStmts(p.Cond, cur)
		case *syntax.WhileClause:
			return inStmts(p.Cond, cur)
		}
		return false
	}
}

// upStmt returns the statement that holds a binary command.
func upStmt(cmd *syntax.BinaryCmd, up map[syntax.Node]syntax.Node) syntax.Node {
	if stmt, ok := up[cmd].(*syntax.Stmt); ok {
		return stmt
	}
	return nil
}

func inStmts(list []*syntax.Stmt, node syntax.Node) bool {
	for _, stmt := range list {
		if syntax.Node(stmt) == node {
			return true
		}
	}
	return false
}

// guardsStatus reports whether the right side of a `||` keeps the caller from
// carrying on: it returns, exits or stops the script, records the status in a
// variable, or opens a block that handles the failure.
func guardsStatus(stmt *syntax.Stmt) bool {
	switch cmd := stmt.Cmd.(type) {
	case *syntax.Block:
		return true
	case *syntax.CallExpr:
		if len(cmd.Args) == 0 {
			return len(cmd.Assigns) > 0
		}
		return statusGuards[callName(cmd)]
	case *syntax.BinaryCmd:
		return guardsStatus(cmd.X)
	}
	return false
}

// stopsAnyway reports whether a substitution builds an argument of a command
// that stops the script or returns anyway, such as the message of
// `dybatpho::die`: its status no longer decides anything.
func stopsAnyway(cs *syntax.CmdSubst, up map[syntax.Node]syntax.Node) bool {
	for n := up[cs]; n != nil; n = up[n] {
		if call, ok := n.(*syntax.CallExpr); ok {
			return statusGuards[callName(call)]
		}
	}
	return false
}
