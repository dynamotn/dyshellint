package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG107",
		Section:  sectionNetwork,
		Severity: lint.SeverityWarning,
		Doc:      "Bound a `curl` call with `--max-time`, or run it under `timeout`",
		Check:    checkCurlTimeout,
	})
}

// checkCurlTimeout reports a `curl` with no time limit. It waits as long as
// the server keeps the connection open, and a CI job hangs until the runner
// kills it, with nothing saying which call stalled. Under `timeout`, the call
// is an argument and is not seen here; options kept in an array or a config
// file may carry the limit, so such calls are left alone.
func checkCurlTimeout(f *File, r *Reporter) {
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "curl" || hasFlag(call, 'm') || hasFlag(call, 'K') || hasFlag(call, 'V') || hasFlag(call, 'h') {
			return
		}
		for _, arg := range call.Args[1:] {
			src := wordSource(arg)
			if strings.HasPrefix(src, "--max-time") || strings.HasPrefix(src, "--config") ||
				src == "--version" || src == "--help" || strings.Contains(src, "[@]") {
				return
			}
		}
		reportWithDybatpho(f, r, call.Pos(), "`dybatpho::curl_timeout <url> <output> <connect> <total>` sets both limits for one request", "`curl` with no time limit waits as long as the server keeps the connection open; add `--connect-timeout` and `--max-time`, or run it under `timeout`")
	})
}
