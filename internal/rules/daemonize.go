package rules

import (
	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionLongRunning = "Environment > Long-running Processes"

func init() {
	register(Rule{
		Code:     "BSG139",
		Section:  sectionLongRunning,
		Severity: lint.SeverityWarning,
		Doc:      "Hand a service to systemd, launchd or a process manager instead of detaching it with `nohup` or `setsid`",
		Check:    checkDetach,
	})
}

// checkDetach reports the commands that cut a process loose from the script:
// once detached, nothing restarts it when it dies, stops it on shutdown, or
// collects its output, and its pid is a guess. `setsid --wait` runs the
// command in a new session and waits for it, which detaches nothing.
// `disown` is left alone: a script sends no `SIGHUP` to its jobs anyway, and
// it is mostly used to silence the report of a job about to be killed.
func checkDetach(f *File, r *Reporter) {
	eachCall(f, func(call *syntax.CallExpr, name string) {
		switch name {
		case "nohup":
		case "setsid":
			if hasFlag(call, 'w') {
				return
			}
			for _, arg := range call.Args[1:] {
				if wordLiteral(arg) == "--wait" {
					return
				}
			}
		default:
			return
		}
		reportWithDybatpho(f, r, call.Pos(), "for a job the script waits on, `dybatpho::background_run` and `dybatpho::wait_all` keep track of it", "`%s` leaves a process that nothing restarts, stops on shutdown or logs; run it as a systemd service (or a launchd agent, or under a process manager) instead", name)
	})
}
