package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const (
	sectionWaiting     = "Calling Commands > Waiting for a Condition"
	sectionParallelism = "Calling Commands > Bounded Parallelism"
)

func init() {
	register(
		Rule{
			Code:     "BSG154",
			Section:  sectionWaiting,
			Severity: lint.SeverityWarning,
			Doc:      "Give every wait a deadline: a loop that sleeps until a condition holds must also stop after a time or a number of tries",
			Check:    checkUnboundedWait,
		},
		Rule{
			Code:     "BSG155",
			Section:  sectionParallelism,
			Severity: lint.SeverityWarning,
			Doc:      "Limit how many jobs run at once: a loop that starts a job with `&` waits for one before starting the next",
			Check:    checkUnboundedFanOut,
		},
	)
}

// exitCalls are the commands that leave a loop or the script.
var exitCalls = map[string]bool{"break": true, "return": true, "exit": true, "dybatpho::die": true}

// deadlineVars are the clocks a deadline is measured against.
var deadlineVars = map[string]bool{"SECONDS": true, "EPOCHSECONDS": true, "EPOCHREALTIME": true}

// numericTests are the integer comparisons of `test` and `[[ ]]`.
var numericTests = map[string]bool{"-lt": true, "-le": true, "-gt": true, "-ge": true, "-eq": true, "-ne": true}

// checkUnboundedWait reports a `while` or `until` loop that sleeps between
// two checks and has no way to give up: nothing in it leaves the loop, reads
// a clock, or compares a counter. A loop on `true` or `:` that never leaves
// is a service loop rather than a wait, and a loop on `read` ends with its
// input; both are left alone.
func checkUnboundedWait(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		loop, ok := node.(*syntax.WhileClause)
		if !ok || !sleepsIn(loop.Do) || endlessCond(loop) || readsCond(loop) || bounded(loop) {
			return true
		}
		reportWithDybatpho(f, r, loop.Pos(), "`dybatpho::retry_until <tries> <delay> -- <check>` stops after a number of tries, and `dybatpho::wait_port <host> <port> <seconds>` keeps to a deadline", "this loop sleeps until its condition holds and waits forever when it never does; stop at a deadline, `deadline=$((SECONDS + 60))` and `((SECONDS < deadline)) || return 1`, or after a number of tries")
		return true
	})
}

// sleepsIn reports whether a list of statements runs `sleep`.
func sleepsIn(list []*syntax.Stmt) bool {
	found := false
	for _, stmt := range list {
		allCalls(stmt, func(_ *syntax.CallExpr, callee string) {
			found = found || callee == "sleep"
		})
	}
	return found
}

// endlessCond reports a `while true` or `while :` loop.
func endlessCond(loop *syntax.WhileClause) bool {
	if loop.Until || len(loop.Cond) != 1 {
		return false
	}
	call, ok := loop.Cond[0].Cmd.(*syntax.CallExpr)
	if !ok || len(call.Args) != 1 {
		return false
	}
	name := callName(call)
	return name == "true" || name == ":"
}

// readsCond reports a loop whose condition reads its input, which ends with
// that input, or checks that a process is alive with `kill -0`, which ends
// with that process: the owner of the input or of the process owns the
// deadline.
func readsCond(loop *syntax.WhileClause) bool {
	found := false
	for _, stmt := range loop.Cond {
		allCalls(stmt, func(call *syntax.CallExpr, callee string) {
			found = found || callee == "read" || callee == "kill" && len(call.Args) > 1 && wordLiteral(call.Args[1]) == "-0"
		})
	}
	return found
}

// bounded reports whether a loop can give up on its own: it leaves with
// `break`, `return`, `exit` or `dybatpho::die`, reads a clock, or compares
// numbers, which is how a counter of tries is checked.
func bounded(loop *syntax.WhileClause) bool {
	found := false
	syntax.Walk(loop, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.FuncDecl:
			return false
		case *syntax.CallExpr:
			name := callName(n)
			if exitCalls[name] {
				found = true
			}
			if name == "test" || name == "[" {
				for _, arg := range n.Args[1:] {
					found = found || numericTests[wordLiteral(arg)]
				}
			}
		case *syntax.ParamExp:
			found = found || n.Param != nil && deadlineVars[n.Param.Value]
		case *syntax.ArithmCmd:
			found = found || comparesNumbers(n.X)
		case *syntax.ArithmExp:
			found = found || comparesNumbers(n.X)
		case *syntax.BinaryTest:
			found = found || numericTests[n.Op.String()]
		}
		return !found
	})
	return found
}

// comparesNumbers reports whether an arithmetic expression compares two
// values, or names a clock, as in `((SECONDS < deadline))`.
func comparesNumbers(expr syntax.ArithmExpr) bool {
	found := false
	syntax.Walk(expr, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.BinaryArithm:
			switch n.Op {
			case syntax.Lss, syntax.Gtr, syntax.Leq, syntax.Geq, syntax.Eql, syntax.Neq:
				found = true
			}
		case *syntax.Word:
			found = found || deadlineVars[wordLiteral(n)]
		}
		return !found
	})
	return found
}

// checkUnboundedFanOut reports a job started with `&` in a loop when no loop
// around it waits: one job per item all at once, which on a list of a
// thousand hosts or files is a thousand processes. A `wait` or a `jobs` count
// in the loop, or in a loop around it that runs the inner one in batches,
// limits it.
func checkUnboundedFanOut(f *File, r *Reporter) {
	up := f.parents()
	reported := map[syntax.Node]bool{}
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		stmt, ok := node.(*syntax.Stmt)
		if !ok || !stmt.Background {
			return true
		}
		inner := enclosingLoop(stmt, up)
		if inner == nil || reported[inner] {
			return true
		}
		for loop := inner; loop != nil; loop = enclosingLoopOf(loop, up) {
			if throttles(loop) {
				return true
			}
		}
		reported[inner] = true
		reportWithDybatpho(f, r, stmt.Pos(), "`dybatpho::parallel_map <jobs> <function> <items...>` runs the items a few at a time and reports each status; `dybatpho::cpu_count` gives a sensible limit", "this loop starts one job per item, all at once, with no limit; count the running jobs and `wait -n` for one when the count reaches a limit, or use `xargs -P`")
		return true
	})
}

// enclosingLoopOf returns the loop around a loop, or nil.
func enclosingLoopOf(loop syntax.Node, up map[syntax.Node]syntax.Node) syntax.Node {
	for n := up[loop]; n != nil; n = up[n] {
		if stmt, ok := n.(*syntax.Stmt); ok {
			return enclosingLoop(stmt, up)
		}
	}
	return nil
}

// throttles reports whether the body of a loop waits for a job or counts the
// running ones, as a nested `while ((running >= max))` does.
func throttles(loop syntax.Node) bool {
	var body []*syntax.Stmt
	switch l := loop.(type) {
	case *syntax.ForClause:
		body = l.Do
	case *syntax.WhileClause:
		body = l.Do
	}
	found := false
	for _, stmt := range body {
		allCalls(stmt, func(_ *syntax.CallExpr, callee string) {
			found = found || callee == "wait" || callee == "jobs" || strings.HasPrefix(callee, "dybatpho::wait")
		})
		syntax.Walk(stmt, func(node syntax.Node) bool {
			if inner, ok := node.(*syntax.WhileClause); ok {
				for _, cond := range inner.Cond {
					syntax.Walk(cond, func(n syntax.Node) bool {
						if arith, ok := n.(*syntax.ArithmCmd); ok && comparesNumbers(arith.X) {
							found = true
						}
						return !found
					})
				}
			}
			return !found
		})
	}
	return found
}
