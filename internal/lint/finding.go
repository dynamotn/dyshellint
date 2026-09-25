// Package lint holds the reporting types shared by every checker, and the
// runners for the external tools the linter orchestrates.
package lint

import (
	"fmt"
	"sort"
	"strings"
)

// Severity mirrors how the style guide classifies its rules: the ✔️ SHOULD and
// ❌ AVOID bullets are mandatory, while ⚠️ CONSIDER bullets are advisory.
type Severity int

const (
	// SeverityWarning marks a ⚠️ CONSIDER rule. It never fails a run on its own.
	SeverityWarning Severity = iota
	// SeverityError marks a ✔️ SHOULD or ❌ AVOID rule.
	SeverityError
)

func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	default:
		return "warning"
	}
}

// ParseSeverity is the inverse of String, used when decoding external tools.
func ParseSeverity(name string) Severity {
	if strings.EqualFold(name, "error") {
		return SeverityError
	}
	return SeverityWarning
}

// Finding is one rule violation at one place in one file.
type Finding struct {
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Column   int      `json:"column"`
	Rule     string   `json:"rule"`
	Severity Severity `json:"-"`
	Level    string   `json:"severity"`
	Message  string   `json:"message"`
	// Section names the heading of the style guide the rule comes from, so a
	// report points at the prose rather than only at the rule code.
	Section string `json:"section,omitempty"`
	// Source is the tool that produced the finding: dyshellint, shellcheck or shfmt.
	Source string `json:"source"`
}

func (f Finding) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s:%d:%d: %s: [%s] %s", f.File, f.Line, f.Column, f.Level, f.Rule, f.Message)
	if f.Section != "" {
		fmt.Fprintf(&b, " (%s)", f.Section)
	}
	return b.String()
}

// Sort orders findings the way a reader walks a report: by file, then by
// position, then by rule so that the output is stable between runs.
func Sort(findings []Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		switch {
		case a.File != b.File:
			return a.File < b.File
		case a.Line != b.Line:
			return a.Line < b.Line
		case a.Column != b.Column:
			return a.Column < b.Column
		default:
			return a.Rule < b.Rule
		}
	})
}
