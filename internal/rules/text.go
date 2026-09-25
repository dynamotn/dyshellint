package rules

import (
	"strings"
	"unicode/utf8"

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
		},
		Rule{
			Code:     "BSG071",
			Section:  sectionTabsSpaces,
			Severity: lint.SeverityError,
			Doc:      "Indent with two spaces, never with tabs",
			Check:    checkTabIndent,
		},
		Rule{
			Code:     "BSG072",
			Section:  sectionTabsSpaces,
			Severity: lint.SeverityError,
			Doc:      "Do not leave trailing whitespace",
			Check:    checkTrailingSpace,
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
	inHeredoc := false
	for i, line := range f.Lines {
		// A tab inside a `<<-` here document is what makes the operator work,
		// so indentation is only checked outside one.
		if strings.Contains(line, "<<-") {
			inHeredoc = true
		}
		if inHeredoc {
			if strings.TrimSpace(line) != "" && !strings.HasPrefix(strings.TrimSpace(line), "#") &&
				!strings.ContainsAny(line, "\t") {
				inHeredoc = false
			}
			continue
		}
		indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		if !strings.Contains(indent, "\t") {
			continue
		}
		r.AtLine(i+1, "line is indented with a tab; use two spaces")
	}
}

func checkTrailingSpace(f *File, r *Reporter) {
	for i, line := range f.Lines {
		if line == "" || strings.TrimRight(line, " \t") == line {
			continue
		}
		r.AtLine(i+1, "line ends with whitespace")
	}
}
