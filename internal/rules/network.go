package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const (
	sectionNetwork = "Calling Commands > Network Requests"
	sectionRemote  = "Calling Commands > Remote Commands"
)

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

func init() {
	register(Rule{
		Code:     "BSG056",
		Section:  sectionNetwork,
		Severity: lint.SeverityWarning,
		Doc:      "Make `curl` fail on an HTTP error with `--fail`, or check `%{http_code}`",
		Check:    checkCurlWithoutFail,
	})
}

func checkCurlWithoutFail(f *File, r *Reporter) {
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "curl" || hasFlag(call, 'f') || hasLong(call, "--fail") || hasLong(call, "--fail-with-body") ||
			hasLong(call, "--version") || checksHTTPCode(call) {
			return
		}
		reportWithDybatpho(f, r, call.Pos(), "`dybatpho::curl_do` turns an HTTP error into a non-zero status, and retries only a request that may still succeed", "`curl` exits 0 on a 404 or a 500 and hands back the error page as if it were the answer; add `--fail`, or read `-w '%%{http_code}'` and check it")
	})
}

// checksHTTPCode reports whether a call asks curl for the status code, which
// the caller then checks itself.
func checksHTTPCode(call *syntax.CallExpr) bool {
	for i, arg := range call.Args[1:] {
		lit := wordLiteral(arg)
		if (lit == "-w" || lit == "--write-out") && i+2 < len(call.Args) {
			src := wordSource(call.Args[i+2])
			if strings.Contains(src, "http_code") || strings.Contains(src, "response_code") {
				return true
			}
		}
	}
	return false
}

func init() {
	register(Rule{
		Code:     "BSG128",
		Section:  sectionRemote,
		Severity: lint.SeverityWarning,
		Doc:      "Quote every local value in a remote `ssh` command with `${value@Q}`, or send a script to `bash -s`",
		Check:    checkRemoteSplice,
	})
}

// sshArgOptions are the options of `ssh` that take a value.
const sshArgOptions = "BbcDEeFIiJLlmOoPpQRSWw"

// unwrapCall returns the words of the command a call runs, past the wrappers
// that only bound or detach it: `timeout`, `nohup`, `command` and `exec`.
func unwrapCall(call *syntax.CallExpr) []*syntax.Word {
	args := call.Args
	for len(args) > 0 {
		switch wordLiteral(args[0]) {
		case "nohup", "command", "exec":
			args = args[1:]
		case "timeout":
			args = args[1:]
			for len(args) > 0 && strings.HasPrefix(wordLiteral(args[0]), "-") {
				if lit := wordLiteral(args[0]); lit == "-s" || lit == "-k" {
					args = args[1:]
				}
				args = args[1:]
			}
			// The duration.
			if len(args) > 0 {
				args = args[1:]
			}
		default:
			return args
		}
	}
	return args
}

// remoteCommand returns the words `ssh` sends to the remote host as its
// command, given the arguments after `ssh`, or nil when there is none or the
// options cannot be told apart from the host, as when they come from an array.
func remoteCommand(args []*syntax.Word) []*syntax.Word {
	for i := 0; i < len(args); i++ {
		src := wordSource(args[i])
		if strings.Contains(src, "[@]") || strings.Contains(src, "[*]") {
			return nil
		}
		lit := wordLiteral(args[i])
		switch {
		case lit == "--":
			if i+1 < len(args) {
				return args[i+2:]
			}
			return nil
		case len(lit) > 1 && lit[0] == '-':
			// In `-qo value` the last letter takes the next word; in `-p22` the
			// value is attached.
			for j := 1; j < len(lit); j++ {
				if strings.IndexByte(sshArgOptions, lit[j]) >= 0 {
					if j == len(lit)-1 {
						i++
					}
					break
				}
			}
		default:
			return args[i+1:]
		}
	}
	return nil
}

// unquotedVar returns the first variable a word expands without quoting it
// for the shell, leaving command substitutions such as `$(declare -f fn)`
// alone.
func unquotedVar(word *syntax.Word, quoted map[string]bool) string {
	var found string
	syntax.Walk(word, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.CmdSubst:
			return false
		case *syntax.ParamExp:
			if found == "" && n.Param != nil && !isShellQuoted(n) && !quoted[n.Param.Value] {
				found = n.Param.Value
			}
			return false
		}
		return found == ""
	})
	return found
}

// checkRemoteSplice reports a local value that reaches the remote shell of
// `ssh` unquoted. ssh joins its command words with spaces, and the remote
// login shell splits the string again, so the quotes of the local shell are
// lost on the way.
func checkRemoteSplice(f *File, r *Reporter) {
	quoted := quotedVars(f.Syntax)
	eachCall(f, func(call *syntax.CallExpr, _ string) {
		args := unwrapCall(call)
		if len(args) == 0 || wordLiteral(args[0]) != "ssh" {
			return
		}
		words := remoteCommand(args[1:])
		for _, word := range words {
			var v string
			if len(words) == 1 {
				v = splicedVar(word, quoted)
			} else {
				v = unquotedVar(word, quoted)
			}
			if v != "" {
				r.At(word.Pos(), "`ssh` joins its command into one string that the remote shell splits again, so %q loses its quotes there: a space splits it and `$(...)` or `;` in it runs; quote it with `${%s@Q}`, or send the function with `declare -f` to `bash -s`", v, v)
				return
			}
		}
	})
}
