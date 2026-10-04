package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionTestingStrings = "Features and Bugs > Testing Strings"

func init() {
	register(
		Rule{
			Code:     "BSG116",
			Section:  sectionBashVersion,
			Severity: lint.SeverityError,
			Doc:      "Compare the Bash version per component: not `major >= X && minor >= Y`, nor `BASH_VERSION` as text",
			Check:    checkVersionCompare,
		},
		Rule{
			Code:     "BSG117",
			Section:  sectionNetwork,
			Severity: lint.SeverityWarning,
			Doc:      "Pause between the attempts of a network retry loop, longer each time",
			Check:    checkTightRetry,
		},
		Rule{
			Code:     "BSG118",
			Section:  sectionNetwork,
			Severity: lint.SeverityWarning,
			Doc:      "Give `ssh` a `ConnectTimeout`, or run it under `timeout`",
			Check:    checkSSHTimeout,
		},
		Rule{
			Code:     "BSG119",
			Section:  sectionBuiltins,
			Severity: lint.SeverityWarning,
			Doc:      "End `find -exec` with `+`, which runs the command once for many files",
			Check:    checkFindExecEach,
		},
		Rule{
			Code:     "BSG121",
			Section:  sectionTestingStrings,
			Severity: lint.SeverityWarning,
			Doc:      "Compare a flag variable instead of running it as a command: `if ${flag}; then` executes its value",
			Check:    checkFlagAsCommand,
		},
		Rule{
			Code:     "BSG122",
			Section:  sectionStdio,
			Severity: lint.SeverityWarning,
			Doc:      "Send both streams to a file with `&>` or `&>>`, not `> file 2>&1` or `2>&1 > file`",
			Check:    checkSplitRedirect,
		},
		Rule{
			Code:     "BSG123",
			Section:  sectionShellCheck,
			Severity: lint.SeverityWarning,
			Doc:      "Say why on a `# shellcheck disable=` directive",
			Check:    checkBareDirective,
		},
	)
}

// versionInfo reports whether an arithmetic operand is `BASH_VERSINFO[n]`.
func versionInfo(expr syntax.ArithmExpr, index string) bool {
	word, ok := expr.(*syntax.Word)
	if !ok || len(word.Parts) != 1 {
		return false
	}
	pe, ok := word.Parts[0].(*syntax.ParamExp)
	if !ok || pe.Param == nil || pe.Param.Value != "BASH_VERSINFO" {
		return false
	}
	at, ok := pe.Index.(*syntax.Word)
	return ok && wordLiteral(at) == index
}

func checkVersionCompare(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.BinaryArithm:
			if n.Op != syntax.AndArit {
				return true
			}
			x, okX := n.X.(*syntax.BinaryArithm)
			y, okY := n.Y.(*syntax.BinaryArithm)
			if okX && okY && x.Op == syntax.Geq && versionInfo(x.X, "0") &&
				(y.Op == syntax.Geq || y.Op == syntax.Gtr) && versionInfo(y.X, "1") {
				r.At(n.Pos(), "`major >= X && minor >= Y` refuses a newer major with a lower minor, such as Bash 6.0 for a 5.2 floor; write `BASH_VERSINFO[0] > X || (BASH_VERSINFO[0] == X && BASH_VERSINFO[1] >= Y)`")
			}
		case *syntax.BinaryTest:
			if n.Op != syntax.TsBefore && n.Op != syntax.TsAfter {
				return true
			}
			if x, ok := n.X.(*syntax.Word); ok && strings.Contains(wordSource(x), "BASH_VERSION") {
				r.At(n.Pos(), "`BASH_VERSION` compares as text here, where `5.10` sorts before `5.2`; compare `BASH_VERSINFO` per component")
			}
		}
		return true
	})
}

// networkCommands are the commands a retry loop usually waits on.
var networkCommands = map[string]bool{"curl": true, "wget": true, "ssh": true, "nc": true, "scp": true, "rsync": true}

// checkTightRetry reports a `while` or `until` loop whose condition runs a
// network command and whose body never sleeps: every client of a failing
// service then retries at once, as fast as it can.
func checkTightRetry(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		loop, ok := node.(*syntax.WhileClause)
		if !ok {
			return true
		}
		network := ""
		for _, stmt := range loop.Cond {
			allCalls(stmt, func(_ *syntax.CallExpr, callee string) {
				if networkCommands[callee] {
					network = callee
				}
			})
		}
		if network == "" {
			return true
		}
		sleeps := false
		for _, stmt := range loop.Do {
			allCalls(stmt, func(_ *syntax.CallExpr, callee string) {
				if callee == "sleep" {
					sleeps = true
				}
			})
		}
		if !sleeps {
			reportWithDybatpho(f, r, loop.Pos(), "`dybatpho::retry <times> '<command>'` waits longer after each attempt and stops after the last, and `dybatpho::curl_do` retries a request on its own", "this loop retries `%s` with no pause, hammering a service that is already failing; sleep between attempts, longer each time, and stop after a few", network)
		}
		return true
	})
}

func checkSSHTimeout(f *File, r *Reporter) {
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "ssh" && name != "scp" || hasFlag(call, 'F') || hasFlag(call, 'V') {
			return
		}
		for _, arg := range call.Args[1:] {
			if src := wordSource(arg); strings.Contains(src, "ConnectTimeout") || strings.Contains(src, "[@]") {
				return
			}
		}
		r.At(call.Pos(), "`%s` waits for an unreachable host as long as the kernel does, and for a password forever; add `-o ConnectTimeout=10 -o BatchMode=yes`, and run it under `timeout` to bound the whole call", name)
	})
}

func checkFindExecEach(f *File, r *Reporter) {
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "find" {
			return
		}
		args := call.Args
		for i := 1; i < len(args); i++ {
			if lit := wordLiteral(args[i]); lit != "-exec" && lit != "-execdir" {
				continue
			}
			braces, end := 0, -1
			for j := i + 1; j < len(args); j++ {
				lit := wordLiteral(args[j])
				if lit == "{}" {
					braces++
				}
				if lit == `\;` || lit == ";" || lit == "+" {
					end = j
					break
				}
			}
			if end > 0 && wordLiteral(args[end]) != "+" && braces == 1 && wordLiteral(args[end-1]) == "{}" {
				r.At(args[end].Pos(), "`-exec ... {} \\;` starts one process per file; end it with `+` to pass many files to one command")
			}
		}
	})
}

// commandName matches a variable named for the command it holds, which is
// run on purpose.
var commandName = regexp.MustCompile(`(?i)(cmd|command|handler|callback|func|fn|hook|action|runner|program|binary|bin|tool|editor|pager|shell|probe|exe)$`)

// checkFlagAsCommand reports a condition that runs a variable as a command,
// as in `if ${force}; then`: it works for `true` and `false`, and executes any
// other value.
func checkFlagAsCommand(f *File, r *Reporter) {
	check := func(stmt *syntax.Stmt) {
		call, ok := stmt.Cmd.(*syntax.CallExpr)
		if !ok || len(call.Args) != 1 {
			return
		}
		pe := wordParam(call.Args[0])
		if pe == nil || pe.Param == nil || pe.Index != nil || pe.Exp != nil || isPositionalParam(pe.Param.Value) ||
			commandName.MatchString(pe.Param.Value) {
			return
		}
		r.At(stmt.Pos(), "`%s` runs the value of %q as a command, which executes anything other than `true` or `false`; compare it: `[[ \"${%s}\" == true ]]` or `dybatpho::is true \"${%s}\"`",
			wordSource(call.Args[0]), pe.Param.Value, pe.Param.Value, pe.Param.Value)
	}
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.IfClause:
			for _, stmt := range n.Cond {
				check(stmt)
			}
		case *syntax.WhileClause:
			for _, stmt := range n.Cond {
				check(stmt)
			}
		case *syntax.BinaryCmd:
			if n.Op == syntax.AndStmt || n.Op == syntax.OrStmt {
				check(n.X)
			}
		}
		return true
	})
}

// checkSplitRedirect reports `> file 2>&1`, which `&>` says in one operator,
// and `2>&1 > file`, which still sends errors to the terminal.
func checkSplitRedirect(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		stmt, ok := node.(*syntax.Stmt)
		if !ok {
			return true
		}
		toFile := -1
		for i, redir := range stmt.Redirs {
			switch {
			case (redir.Op == syntax.RdrOut || redir.Op == syntax.AppOut) && (redir.N == nil || redir.N.Value == "1"):
				toFile = i
			case redir.Op == syntax.DplOut && redir.N != nil && redir.N.Value == "2" && wordLiteral(redir.Word) == "1":
				if toFile >= 0 {
					r.At(stmt.Redirs[toFile].OpPos, "`> file 2>&1` sends both streams to the file; write `&>` (or `&>>` to append), which cannot be put in the wrong order")
				} else if i+1 < len(stmt.Redirs) && stmt.Redirs[i+1].Op == syntax.RdrOut || i+1 < len(stmt.Redirs) && stmt.Redirs[i+1].Op == syntax.AppOut {
					r.At(redir.OpPos, "`2>&1 > file` copies standard error to the terminal before standard output moves, so errors still reach the screen; write `&> file`")
				}
			}
		}
		return true
	})
}

// bareDirective matches a ShellCheck disable directive with no reason after
// it on the same line.
var bareDirective = regexp.MustCompile(`^\s*#\s*shellcheck\s+disable=[A-Za-z0-9,]+\s*$`)

func checkBareDirective(f *File, r *Reporter) {
	for i, line := range f.Lines {
		if !bareDirective.MatchString(line) {
			continue
		}
		if i > 0 {
			prev := strings.TrimSpace(f.Lines[i-1])
			if strings.HasPrefix(prev, "#") && !strings.Contains(prev, "shellcheck") && !strings.HasPrefix(prev, "#!") && prev != "#" {
				continue
			}
		}
		r.AtLine(i+1, "this directive does not say why the warning does not apply; add the reason after a second `#` on the same line")
	}
}
