package lint

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ShellCheck runs the external `shellcheck` binary and translates its findings.
// The repository's `.shellcheckrc` is what selects the checks, so the runner
// passes no rule flags of its own.
type ShellCheck struct {
	// Binary is the program to run, `shellcheck` unless overridden.
	Binary string
	// RCFile is the configuration to load. It is only needed when the file on
	// disk sits outside the repository, as a buffer read from standard input
	// does, because ShellCheck looks for `.shellcheckrc` next to the file.
	RCFile string
	// SourceDir is where a `source=` directive is resolved from, for the same
	// reason.
	SourceDir string
}

type shellCheckReport struct {
	Comments []struct {
		File    string `json:"file"`
		Line    int    `json:"line"`
		Column  int    `json:"column"`
		Level   string `json:"level"`
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"comments"`
}

// ErrToolMissing is returned when an external tool is not on PATH. The caller
// decides whether that is fatal or only worth a note.
var ErrToolMissing = errors.New("tool not found")

// Run checks every file in one shellcheck invocation.
func (s ShellCheck) Run(files []string) ([]Finding, error) {
	binary := s.Binary
	if binary == "" {
		binary = "shellcheck"
	}
	if _, err := exec.LookPath(binary); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrToolMissing, binary)
	}
	if len(files) == 0 {
		return nil, nil
	}
	args := []string{"--format=json1", "--external-sources"}
	if s.RCFile != "" {
		args = append(args, "--rcfile="+s.RCFile)
	}
	if s.SourceDir != "" {
		args = append(args, "--source-path="+s.SourceDir)
	}
	args = append(args, files...)
	out, err := exec.Command(binary, args...).Output()
	// shellcheck exits non-zero as soon as it reports something, so the exit
	// status alone says nothing; only an unparsable body is a real failure.
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		return nil, fmt.Errorf("run %s: %w", binary, err)
	}
	var report shellCheckReport
	if err := json.Unmarshal(out, &report); err != nil {
		return nil, fmt.Errorf("decode %s output: %w", binary, err)
	}
	findings := make([]Finding, 0, len(report.Comments))
	for _, comment := range report.Comments {
		severity := severityForShellCheck(comment.Level)
		findings = append(findings, Finding{
			File:     comment.File,
			Line:     comment.Line,
			Column:   comment.Column,
			Rule:     fmt.Sprintf("SC%d", comment.Code),
			Severity: severity,
			Level:    severity.String(),
			Message:  strings.TrimSpace(comment.Message),
			Section:  "Features and Bugs > Use ShellCheck",
			Source:   "shellcheck",
		})
	}
	return findings, nil
}

// severityForShellCheck follows the guide: warnings and above must be resolved,
// while info and style are the ⚠️ CONSIDER tier.
func severityForShellCheck(level string) Severity {
	switch level {
	case "error", "warning":
		return SeverityError
	default:
		return SeverityWarning
	}
}
