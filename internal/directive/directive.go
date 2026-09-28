// Package directive reads the comments that silence a finding in place: the
// `# dyshellint disable=CODE,...` of this linter, and the
// `# shellcheck disable=SC####` ShellCheck already understands, which silences
// the ShellCheck codes it names here too. Both scope the same way: at the top
// of a file the comment covers the whole file, above a command it covers that
// command and everything nested in it, and at the end of a line it covers that
// line alone.
package directive

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// All is the code that stands for every rule at once.
const All = "ALL"

// pattern matches the body of a directive comment, from either tool. The
// comment text the parser hands over has already lost its `#`, and anything
// after the codes is free text, which is where the reason for the exception
// goes.
var pattern = regexp.MustCompile(`(?i)^\s*(dyshellint|shellcheck)\s+disable=([A-Za-z0-9,]+)`)

// shellcheckPrefix is what a `# shellcheck disable=` comment is held to: it
// speaks for ShellCheck alone, so it can never silence a rule of the guide,
// not even through `all`.
const shellcheckPrefix = "SC"

// scope is a range of lines, inclusive, and the codes silenced inside it.
type scope struct {
	from, to int
	codes    map[string]bool
	// prefix, when set, narrows the scope to the codes starting with it.
	prefix string
}

// Set is every directive of one file.
type Set struct {
	scopes []scope
}

// Parse collects the directives of one parsed file. lines is the source split
// on newlines, indexed from zero, as rules.File keeps it.
func Parse(lines []string, prog *syntax.File) *Set {
	set := &Set{}
	if prog == nil {
		return set
	}
	comments, starts, ends := collect(prog)
	for _, comment := range comments {
		codes, prefix := codesOf(comment.Text)
		if codes == nil {
			continue
		}
		line := int(comment.Pos().Line())
		from, to := scopeOf(line, int(comment.Pos().Col()), lines, starts, ends)
		if from == 0 {
			continue
		}
		set.scopes = append(set.scopes, scope{from: from, to: to, codes: codes, prefix: prefix})
	}
	return set
}

// Suppressed reports whether a finding at line for code is silenced.
func (s *Set) Suppressed(line int, code string) bool {
	if s == nil {
		return false
	}
	code = strings.ToUpper(code)
	for _, sc := range s.scopes {
		if line < sc.from || line > sc.to {
			continue
		}
		if sc.prefix != "" && !strings.HasPrefix(code, sc.prefix) {
			continue
		}
		if sc.codes[All] || sc.codes[code] {
			return true
		}
	}
	return false
}

// codesOf returns the codes a comment disables and the prefix they are held
// to, or nil when the comment is not a directive.
func codesOf(text string) (codes map[string]bool, prefix string) {
	match := pattern.FindStringSubmatch(text)
	if match == nil {
		return nil, ""
	}
	if strings.EqualFold(match[1], "shellcheck") {
		prefix = shellcheckPrefix
	}
	codes = map[string]bool{}
	for _, code := range strings.Split(match[2], ",") {
		code = strings.ToUpper(strings.TrimSpace(code))
		if code == "" {
			continue
		}
		// A ShellCheck comment that names something other than an SC code is
		// not ours to read: ShellCheck itself decides what to do with it.
		if prefix != "" && code != All && !strings.HasPrefix(code, prefix) {
			continue
		}
		codes[code] = true
	}
	if len(codes) == 0 {
		return nil, ""
	}
	return codes, prefix
}

// scopeOf works out how far a directive reaches. A zero `from` means the
// directive covers nothing, which is what a trailing comment below the last
// command of a file amounts to.
func scopeOf(line, col int, lines []string, starts []int, ends map[int]int) (from, to int) {
	if !standalone(line, col, lines) {
		return line, line
	}
	if len(starts) == 0 || line < starts[0] {
		// Nothing has run yet: the directive belongs to the header, and covers
		// the file the way ShellCheck's own top-of-file directive does.
		return 1, len(lines)
	}
	for _, start := range starts {
		if start > line {
			return start, ends[start]
		}
	}
	return 0, 0
}

// standalone reports whether a comment sits on a line of its own, rather than
// at the end of a command.
func standalone(line, col int, lines []string) bool {
	if line < 1 || line > len(lines) {
		return false
	}
	text := lines[line-1]
	if col < 1 || col > len(text)+1 {
		return false
	}
	return strings.TrimSpace(text[:col-1]) == ""
}

// collect walks the tree once for the comments, the line every statement
// starts on, and how far the outermost statement starting on that line runs.
func collect(prog *syntax.File) (comments []*syntax.Comment, starts []int, ends map[int]int) {
	ends = map[int]int{}
	syntax.Walk(prog, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.Comment:
			comments = append(comments, n)
		case *syntax.Stmt:
			start, end := int(n.Pos().Line()), int(n.End().Line())
			if previous, seen := ends[start]; !seen {
				starts = append(starts, start)
				ends[start] = end
			} else if end > previous {
				ends[start] = end
			}
		}
		return true
	})
	sortInts(starts)
	return comments, starts, ends
}

// sortInts keeps the statement lines in order; the walk visits a redirection
// before the command it belongs to, so the tree order alone is not enough.
func sortInts(values []int) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}
