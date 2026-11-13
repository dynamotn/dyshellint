package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionSystemLogging = "Environment > System Logging"

func init() {
	register(Rule{
		Code:     "BSG138",
		Section:  sectionSystemLogging,
		Severity: lint.SeverityWarning,
		Doc:      "Leave `/var/log` to the system: write to `STDERR` or use `logger` instead of appending to a file there",
		Check:    checkVarLogWrite,
	})
}

// underVarLog reports whether a word is a path under `/var/log`, as written.
func underVarLog(word *syntax.Word) bool {
	return strings.HasPrefix(strings.ReplaceAll(wordSource(word), `"`, ""), "/var/log/")
}

// checkVarLogWrite reports a write into `/var/log`. The script needs root for
// it, nothing rotates the file, and under systemd or cron the journal or the
// mail of the supervisor already collects what the script prints on standard
// error.
func checkVarLogWrite(f *File, r *Reporter) {
	const advice = "`dybatpho::log_to_file <path> rotate:10M keep:3` writes a log file that rotates itself"
	const msg = "writing to %s needs root and leaves the file to grow, since nothing rotates it; print to STDERR for journald or cron to collect, or send it with `logger -t <tag>`"
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.Redirect:
			switch n.Op {
			case syntax.RdrOut, syntax.AppOut, syntax.RdrAll, syntax.AppAll, syntax.RdrClob:
				if n.Word != nil && underVarLog(n.Word) {
					reportWithDybatpho(f, r, n.Pos(), advice, msg, wordSource(n.Word))
				}
			}
		case *syntax.CallExpr:
			if callName(n) != "tee" {
				return true
			}
			for _, arg := range n.Args[1:] {
				if underVarLog(arg) {
					reportWithDybatpho(f, r, arg.Pos(), advice, msg, wordSource(arg))
				}
			}
		}
		return true
	})
}
