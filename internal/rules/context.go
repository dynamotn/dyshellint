// Package rules implements the checks that come straight from the Bash coding
// style guide at https://github.com/dynamotn/bash-coding-style and that no
// general-purpose shell linter can express: namespaces, shdoc headers, file
// layout and the dybatpho conventions.
package rules

import (
	"bytes"
	"path/filepath"
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

// Role tells apart the two kinds of file the guide treats differently: a
// library is sourced by other scripts, an entrypoint is run.
type Role int

const (
	// RoleEntrypoint is a runnable script: it has a shebang.
	RoleEntrypoint Role = iota
	// RoleLibrary is a file meant to be sourced, typically under a `lib/` folder.
	RoleLibrary
)

// File is one shell file, parsed once and shared by every rule.
type File struct {
	// Path is the name to print, as it was given on the command line.
	Path string
	// Src is the whole file.
	Src []byte
	// Lines is Src split on newlines, indexed from zero.
	Lines []string
	// Syntax is the parsed program.
	Syntax *syntax.File
	// Role is how the file is meant to be used.
	Role Role
	// Executable reports whether the file carries the executable bit.
	Executable bool
	// ModeKnown reports whether the file mode could be read at all. Content
	// piped in on standard input has no mode of its own, and the rules that
	// depend on one stay quiet rather than guess.
	ModeKnown bool
	// Shebang is the first line when it starts with `#!`, else the empty string.
	Shebang string
	// Namespace is the namespace a library file's functions are expected to
	// use, derived from the base name: `scripts/lib/package_manager.sh` gives
	// `package_manager`. It is empty for entrypoints.
	Namespace string
	// UsesDybatpho reports whether the file sources dybatpho, which relaxes the
	// `set -euo pipefail` rule and enables the dybatpho-specific checks.
	UsesDybatpho bool
}

// Reporter collects the findings of one rule against one file.
type Reporter struct {
	file     *File
	rule     Rule
	findings []lint.Finding
}

// At records a finding at a parsed position.
func (r *Reporter) At(pos syntax.Pos, format string, args ...any) {
	r.report(int(pos.Line()), int(pos.Col()), format, args...)
}

// AtLine records a finding at a line number, one-based, when no node is handy.
func (r *Reporter) AtLine(line int, format string, args ...any) {
	r.report(line, 1, format, args...)
}

func (r *Reporter) report(line, col int, format string, args ...any) {
	r.findings = append(r.findings, lint.Finding{
		File:     r.file.Path,
		Line:     line,
		Column:   col,
		Rule:     r.rule.Code,
		Severity: r.rule.Severity,
		Level:    r.rule.Severity.String(),
		Message:  sprintf(format, args...),
		Section:  r.rule.Section,
		Source:   "dyshellint",
	})
}

// Rule is one check, tied to the heading of the guide that documents it.
type Rule struct {
	// Code is the stable identifier, `BSG` plus three digits.
	Code string
	// Section is the heading of the guide the rule enforces.
	Section string
	// Severity follows the guide: SHOULD and AVOID are errors, CONSIDER warns.
	Severity lint.Severity
	// Doc is a one-line summary, printed by --list-rules.
	Doc string
	// Check runs the rule over one parsed file.
	Check func(f *File, r *Reporter)
}

var registry []Rule

func register(rules ...Rule) {
	registry = append(registry, rules...)
}

// All returns every registered rule, in code order.
func All() []Rule {
	out := make([]Rule, len(registry))
	copy(out, registry)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].Code < out[j-1].Code; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// Run applies the given rules to one file and returns everything they found.
func Run(f *File, rules []Rule) []lint.Finding {
	var findings []lint.Finding
	for _, rule := range rules {
		r := &Reporter{file: f, rule: rule}
		rule.Check(f, r)
		findings = append(findings, r.findings...)
	}
	return findings
}

var dybatphoSource = regexp.MustCompile(`dybatpho/init\.sh|dybatpho::`)

// NewFile parses src and derives everything the rules need from its path.
func NewFile(path string, src []byte, executable bool) (*File, error) {
	parser := syntax.NewParser(syntax.KeepComments(true), syntax.Variant(syntax.LangBash))
	prog, err := parser.Parse(bytes.NewReader(src), path)
	if err != nil {
		return nil, err
	}
	f := &File{
		Path:         path,
		Src:          src,
		Lines:        strings.Split(string(src), "\n"),
		Syntax:       prog,
		Executable:   executable,
		ModeKnown:    true,
		UsesDybatpho: dybatphoSource.Match(src),
	}
	f.Shebang = shebangOf(f.Lines)
	f.Role = RoleEntrypoint
	if f.Shebang == "" || isLibraryPath(path) {
		f.Role = RoleLibrary
	}
	if f.Role == RoleLibrary {
		f.Namespace = strings.TrimSuffix(filepath.Base(path), ".sh")
	}
	return f, nil
}

// shebangOf returns the shebang of a file. A shebang pushed below a comment is
// still a shebang as far as the role of the file goes: BSG024 reports the
// placement, and the remaining rules keep treating the file as an entrypoint
// rather than piling on.
func shebangOf(lines []string) string {
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#!") {
			return trimmed
		}
		if !strings.HasPrefix(trimmed, "#") && trimmed != "" {
			return ""
		}
	}
	return ""
}

// isLibraryPath reports whether a path lives in a `lib` folder, which is where
// the guide puts common function scripts.
func isLibraryPath(path string) bool {
	for _, part := range strings.Split(filepath.ToSlash(filepath.Dir(path)), "/") {
		if part == "lib" {
			return true
		}
	}
	return false
}

// Line returns the one-based line of the file, or the empty string when the
// number is out of range.
func (f *File) Line(n int) string {
	if n < 1 || n > len(f.Lines) {
		return ""
	}
	return f.Lines[n-1]
}

// Text returns the source between two positions, which is how the rules see
// syntax the parser does not keep in the tree, such as the `()` of a function.
func (f *File) Text(from, to syntax.Pos) string {
	start, end := int(from.Offset()), int(to.Offset())
	if start < 0 || end > len(f.Src) || start > end {
		return ""
	}
	return string(f.Src[start:end])
}
