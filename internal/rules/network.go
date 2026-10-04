package rules

import (
	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionNetwork = "Calling Commands > Network Requests"

func init() {
	register(Rule{
		Code:     "BSG058",
		Section:  sectionNetwork,
		Severity: lint.SeverityError,
		Doc:      "Do not pipe a download straight into a shell",
		Check:    checkDownloadIntoShell,
	})
}

// downloaders fetch a URL to standard output.
var downloaders = map[string]bool{"curl": true, "wget": true}

func checkDownloadIntoShell(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.BinaryCmd:
			if n.Op != syntax.Pipe && n.Op != syntax.PipeAll {
				return true
			}
			if downloads(n.X) && runsStdin(firstCall(n.Y)) {
				r.At(n.Pos(), "a download piped into a shell runs whatever the server sends, and half of it when the connection drops; save it to a file, check it, then run it")
			}
		case *syntax.CallExpr:
			name := callName(n)
			if !shellRunners[name] && name != "source" && name != "." {
				return true
			}
			for _, arg := range n.Args[1:] {
				for _, part := range arg.Parts {
					if ps, ok := part.(*syntax.ProcSubst); ok && procDownloads(ps) {
						r.At(arg.Pos(), "`%s <(curl ...)` runs whatever the server sends; save it to a file, check it, then run it", name)
					}
				}
			}
		}
		return true
	})
}

// downloads reports whether a statement, or the last command of a pipeline in
// it, starts with a downloader.
func downloads(stmt *syntax.Stmt) bool {
	for _, call := range leadingCalls(stmt) {
		if downloaders[callName(call)] {
			return true
		}
	}
	return false
}

func procDownloads(ps *syntax.ProcSubst) bool {
	for _, stmt := range ps.Stmts {
		if downloads(stmt) {
			return true
		}
	}
	return false
}

// runsStdin reports whether a command runs the script it reads on stdin.
func runsStdin(call *syntax.CallExpr) bool {
	name := callName(call)
	if shellRunners[name] {
		for _, arg := range call.Args[1:] {
			if lit := wordLiteral(arg); lit != "-s" && lit != "-" && lit != "--" && !(len(lit) > 1 && lit[0] == '-') {
				return false
			}
		}
		return true
	}
	if name == "source" || name == "." {
		return len(call.Args) > 1 && wordLiteral(call.Args[1]) == "/dev/stdin"
	}
	return false
}
