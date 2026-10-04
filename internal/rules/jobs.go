package rules

import (
	"regexp"
	"strconv"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionChildProcesses = "Calling Commands > Child Processes"

func init() {
	register(
		Rule{
			Code:     "BSG111",
			Section:  sectionSignalHandlers,
			Severity: lint.SeverityWarning,
			Doc:      "Keep the script's status in an `EXIT` handler: do not end it with `exit 0` or another fixed status",
			Check:    checkExitTrapStatus,
		},
		Rule{
			Code:     "BSG112",
			Section:  sectionChildProcesses,
			Severity: lint.SeverityWarning,
			Doc:      "Count the failures of `wait` in a loop, instead of stopping at the first one under `set -e`",
			Check:    checkWaitInLoop,
		},
	)
}

// fixedExit matches `exit` followed by a literal status.
var fixedExit = regexp.MustCompile(`\bexit\s+[0-9]+\b`)

// checkExitTrapStatus reports an `EXIT` handler that ends with a literal
// status. The handler decides the status of the script when it calls `exit`,
// so `exit 0` turns a failed run into a success, and `exit 1` the reverse.
func checkExitTrapStatus(f *File, r *Reporter) {
	funcs := f.project().funcs
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "trap" || len(call.Args) < 3 {
			return
		}
		onExit := false
		for _, arg := range call.Args[2:] {
			switch wordLiteral(arg) {
			case "EXIT", "SIGEXIT", "0":
				onExit = true
			}
		}
		if !onExit {
			return
		}
		handler := wordSource(call.Args[1])
		fixed := fixedExit.MatchString(handler)
		if names := splitWords(handler); len(names) == 1 && funcs[names[0]] != nil {
			allCalls(funcs[names[0]].Body, func(inner *syntax.CallExpr, callee string) {
				if callee == "exit" && len(inner.Args) == 2 {
					if _, err := strconv.Atoi(wordLiteral(inner.Args[1])); err == nil {
						fixed = true
					}
				}
			})
		}
		if fixed {
			r.At(call.Pos(), "this `EXIT` handler ends with a fixed status, which replaces the script's own: a failed run reports success; save `$?` first and `exit \"${status}\"`, or end without `exit`")
		}
	})
}

// checkWaitInLoop reports `wait "${pid}"` run as a plain statement in a loop
// under `set -e`. The first job that failed stops the script: the jobs after
// it are never waited for, and their failures never reported.
func checkWaitInLoop(f *File, r *Reporter) {
	if !stopsOnError(f) {
		return
	}
	up := parents(f.Syntax)
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		stmt, ok := node.(*syntax.Stmt)
		if !ok || stmt.Negated {
			return true
		}
		call, ok := stmt.Cmd.(*syntax.CallExpr)
		if !ok || callName(call) != "wait" || len(call.Args) < 2 || hasFlag(call, 'n') {
			return true
		}
		switch loop := up[stmt].(type) {
		case *syntax.ForClause:
			if !inStmts(loop.Do, stmt) {
				return true
			}
		case *syntax.WhileClause:
			if !inStmts(loop.Do, stmt) {
				return true
			}
		default:
			return true
		}
		reportWithDybatpho(f, r, stmt.Pos(), "`dybatpho::background_run <name> <command>` and `dybatpho::wait_all` keep every pid, wait for each job and count the failures", "under `set -e` the first job that failed stops this loop, and the others are neither waited for nor reported; write `wait \"${pid}\" || failed=$((failed + 1))` and check the count after the loop")
		return true
	})
}
