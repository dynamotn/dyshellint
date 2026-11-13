package rules

import (
	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionDates = "Features and Bugs > Dates and Durations"

func init() {
	register(Rule{
		Code:     "BSG150",
		Section:  sectionDates,
		Severity: lint.SeverityWarning,
		Doc:      "Read the clock with `SECONDS` or `printf '%(%s)T' -1`, not by forking `date +%s`",
		Check:    checkDateEpoch,
	})
}

// checkDateEpoch reports `date +%s` asked for the current time and nothing
// else. Bash keeps the clock itself: `SECONDS` counts the seconds a script
// has run, and `printf -v now '%(%s)T' -1` reads the epoch without a fork.
// A call that parses or reformats a date, `date -d ...` or `date -r ...`,
// is left alone: it does more than read the clock.
func checkDateEpoch(f *File, r *Reporter) {
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "date" || len(call.Args) < 2 {
			return
		}
		epoch := false
		for _, arg := range call.Args[1:] {
			switch wordLiteral(arg) {
			case "+%s":
				epoch = true
			case "-u", "--utc", "--universal":
			default:
				return
			}
		}
		if epoch {
			reportWithDybatpho(f, r, call.Pos(), "`dybatpho::timer_start <name>` and `dybatpho::timer_end <name>` time a step and log how long it took", "`date +%%s` forks a process to read a clock Bash already keeps; use `SECONDS` for a duration, or `printf -v now '%%(%%s)T' -1` (Bash 4.2) for the epoch")
		}
	})
}
