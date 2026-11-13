package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const (
	sectionTrace    = "Environment > Trace Output"
	sectionProgress = "Environment > Progress Output"
)

func init() {
	register(
		Rule{
			Code:     "BSG156",
			Section:  sectionTrace,
			Severity: lint.SeverityWarning,
			Doc:      "Make `PS4` show where each traced line comes from: file, `${LINENO}` and function",
			Check:    checkPS4Line,
		},
		Rule{
			Code:     "BSG157",
			Section:  sectionProgress,
			Severity: lint.SeverityWarning,
			Doc:      "Draw a progress line with `\\r` or an erase-line escape only when STDERR is a terminal, `[[ -t 2 ]]`",
			Check:    checkUnguardedProgress,
		},
	)
}

// checkPS4Line reports a `PS4` that leaves out `LINENO`. A trace whose lines
// do not say where they come from cannot be traced back to the code. A value
// that builds on the current `PS4` is left alone: the part it adds to is
// where the line number would be.
func checkPS4Line(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		assign, ok := node.(*syntax.Assign)
		if !ok || assign.Name == nil || assign.Name.Value != "PS4" || assign.Naked || assign.Append {
			return true
		}
		value := ""
		if assign.Value != nil {
			value = wordSource(assign.Value)
		}
		if strings.Contains(value, "LINENO") || strings.Contains(value, "PS4") {
			return true
		}
		reportWithDybatpho(f, r, assign.Pos(), "`dybatpho::start_trace` sets a `PS4` with the file, the line and the function", "this `PS4` does not say where a traced line comes from; include `${LINENO}`, as in `PS4='+ ${BASH_SOURCE[0]##*/}:${LINENO}:${FUNCNAME[0]:-main}: '`")
		return true
	})
}

// redrawEscape matches a carriage return, or an escape that erases the line,
// in shell source: `\r`, `\033[K`, `\e[2K`, `\x1b[K`.
var redrawEscape = regexp.MustCompile(`\\r|\\(?:033|e|E|x1b|x1B)\[[0-2]?K`)

// terminalGuard matches a file that tests for a terminal before drawing, or
// hands the drawing to a dybatpho helper that does.
var terminalGuard = regexp.MustCompile(`(?:\[\[?|\btest)\s+(?:!\s+)?-t\s|\btput\b|dybatpho::(?:is_tty|color_supported|spinner|tui_)`)

// checkUnguardedProgress reports `printf` or `echo -e` drawing over the
// current line, with `\r` or an erase-line escape, in a file that never asks
// whether it writes to a terminal. In a log file or a pipe, the redraws pile
// up as one long line of control characters. `\r\n`, the line ending of a
// network protocol, draws nothing and is left alone.
func checkUnguardedProgress(f *File, r *Reporter) {
	if terminalGuard.Match(f.Src) {
		return
	}
	eachCall(f, func(call *syntax.CallExpr, name string) {
		var words []*syntax.Word
		switch name {
		case "printf":
			args := call.Args[1:]
			if len(args) >= 2 && wordLiteral(args[0]) == "-v" {
				return
			}
			if len(args) > 0 && wordLiteral(args[0]) == "--" {
				args = args[1:]
			}
			if len(args) > 0 {
				words = append(words, args[0])
			}
			for _, arg := range args {
				if strings.HasPrefix(wordSource(arg), "$'") {
					words = append(words, arg)
				}
			}
		case "echo":
			escapes := hasFlag(call, 'e')
			for _, arg := range call.Args[1:] {
				if escapes || strings.HasPrefix(wordSource(arg), "$'") {
					words = append(words, arg)
				}
			}
		default:
			return
		}
		for _, word := range words {
			src := strings.ReplaceAll(wordSource(word), `\r\n`, "")
			if redrawEscape.MatchString(src) {
				reportWithDybatpho(f, r, call.Pos(), "`dybatpho::spinner` and `dybatpho::tui_progress_*` draw only on a terminal and log one line otherwise", "this redraws the current line, which in a log file or a pipe piles up as control characters; draw on STDERR and only when `[[ -t 2 ]]`")
				return
			}
		}
	})
}
