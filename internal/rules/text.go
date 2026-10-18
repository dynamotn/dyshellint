package rules

import (
	"strings"
	"unicode/utf8"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const (
	sectionTabsSpaces = "Formatting > Tabs and Spaces"
	sectionLineLength = "Formatting > Line Length and Long Strings"
)

// maxLineLength is the limit the guide sets for a source line.
const maxLineLength = 120

func init() {
	register(
		Rule{
			Code:     "BSG070",
			Section:  sectionLineLength,
			Severity: lint.SeverityError,
			Doc:      "Keep lines at 120 characters or fewer",
			Check:    checkLineLength,
			Bats:     true,
		},
		Rule{
			Code:     "BSG071",
			Section:  sectionTabsSpaces,
			Severity: lint.SeverityError,
			Doc:      "Indent with two spaces, never with tabs",
			Check:    checkTabIndent,
			Bats:     true,
		},
		Rule{
			Code:     "BSG072",
			Section:  sectionTabsSpaces,
			Severity: lint.SeverityError,
			Doc:      "Do not leave trailing whitespace",
			Check:    checkTrailingSpace,
			Bats:     true,
		},
	)
}

func checkLineLength(f *File, r *Reporter) {
	for i, line := range f.Lines {
		length := utf8.RuneCountInString(line)
		if length <= maxLineLength {
			continue
		}
		r.AtLine(i+1, "line is %d characters long; keep it under %d, with a here document or an embedded newline for a long string",
			length, maxLineLength)
	}
}

func checkTabIndent(f *File, r *Reporter) {
	inHeredoc := heredocLines(f)
	for i, line := range f.Lines {
		if inHeredoc[i+1] {
			continue
		}
		indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		if !strings.Contains(indent, "\t") {
			continue
		}
		r.AtLine(i+1, "line is indented with a tab; use two spaces")
	}
}

// heredocLines returns the one-based lines that hold the body of a here
// document. Those lines are data the script writes out, such as a Makefile
// whose recipes need their tab, so their whitespace is not the script's own.
// A `<<-` here document strips leading tabs, which is what the operator is
// for, so the line that opens it and the line that closes it count as well.
func heredocLines(f *File) map[int]bool {
	out := map[int]bool{}
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		redir, ok := node.(*syntax.Redirect)
		if !ok || (redir.Op != syntax.Hdoc && redir.Op != syntax.DashHdoc) || redir.Hdoc == nil {
			return true
		}
		first, last := int(redir.Hdoc.Pos().Line()), int(redir.Hdoc.End().Line())
		// The body ends with a newline, so its end sits at the start of the
		// line that holds the closing delimiter.
		if redir.Hdoc.End().Col() == 1 && redir.Op == syntax.Hdoc {
			last--
		}
		if redir.Op == syntax.DashHdoc {
			first = int(redir.OpPos.Line())
		}
		for line := first; line <= last; line++ {
			out[line] = true
		}
		return true
	})
	return out
}

func checkTrailingSpace(f *File, r *Reporter) {
	inHeredoc := heredocLines(f)
	for i, line := range f.Lines {
		if inHeredoc[i+1] || line == "" || strings.TrimRight(line, " \t") == line {
			continue
		}
		r.AtLine(i+1, "line ends with whitespace")
	}
}
