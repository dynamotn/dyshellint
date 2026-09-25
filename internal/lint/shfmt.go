package lint

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Shfmt runs the external `shfmt` binary in diff mode. Its options encode the
// Formatting chapter of the guide: two-space indentation, indented `case`
// bodies, binary operators at the start of a continued line, and a space after
// a redirection operator.
type Shfmt struct {
	// Binary is the program to run, `shfmt` unless overridden.
	Binary string
}

// Options are the formatting flags the guide implies.
var shfmtOptions = []string{"--language-dialect", "bash", "--indent", "2", "--case-indent", "--binary-next-line", "--space-redirects", "--diff"}

var hunkHeader = regexp.MustCompile(`^@@ -(\d+)(?:,\d+)? \+\d+(?:,\d+)? @@`)

// Run formats every file and reports each hunk that differs.
func (s Shfmt) Run(files []string) ([]Finding, error) {
	binary := s.Binary
	if binary == "" {
		binary = "shfmt"
	}
	if _, err := exec.LookPath(binary); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrToolMissing, binary)
	}
	var findings []Finding
	for _, file := range files {
		out, err := exec.Command(binary, append(shfmtOptions, file)...).Output()
		var exitErr *exec.ExitError
		switch {
		case err == nil:
			continue // The file is already formatted.
		case errors.As(err, &exitErr) && len(out) > 0:
		case errors.As(err, &exitErr):
			return nil, fmt.Errorf("run %s on %s: %s", binary, file, strings.TrimSpace(string(exitErr.Stderr)))
		default:
			return nil, fmt.Errorf("run %s on %s: %w", binary, file, err)
		}
		findings = append(findings, hunkFindings(file, string(out))...)
	}
	return findings, nil
}

// hunkFindings turns a unified diff into one finding per hunk, so the report
// points at the lines to reformat rather than at the file as a whole.
func hunkFindings(file, diff string) []Finding {
	var findings []Finding
	for _, line := range strings.Split(diff, "\n") {
		match := hunkHeader.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		number, _ := strconv.Atoi(match[1])
		findings = append(findings, Finding{
			File:     file,
			Line:     number,
			Column:   1,
			Rule:     "FMT001",
			Severity: SeverityError,
			Level:    SeverityError.String(),
			Message:  "not formatted; run `shfmt -w` with the options in scripts/lint.sh",
			Section:  "Formatting",
			Source:   "shfmt",
		})
	}
	return findings
}
