package lint

import (
	"bytes"
	"strings"
	"testing"
)

func TestSeverityForShellCheck(t *testing.T) {
	cases := map[string]Severity{
		"error":   SeverityError,
		"warning": SeverityError,
		"info":    SeverityWarning,
		"style":   SeverityWarning,
	}
	for level, want := range cases {
		if got := severityForShellCheck(level); got != want {
			t.Errorf("severityForShellCheck(%q) = %v, want %v", level, got, want)
		}
	}
}

func TestHunkFindings(t *testing.T) {
	diff := strings.Join([]string{
		"--- a/script.sh.orig",
		"+++ b/script.sh",
		"@@ -3,5 +3,5 @@",
		"-if [ -n $name ]",
		"+if [[ -n ${name} ]]",
		"@@ -20 +20 @@",
		"-done",
		"+done",
	}, "\n")
	findings := hunkFindings("script.sh", diff)
	if len(findings) != 2 {
		t.Fatalf("got %d findings, want 2", len(findings))
	}
	if findings[0].Line != 3 || findings[1].Line != 20 {
		t.Errorf("got lines %d and %d, want 3 and 20", findings[0].Line, findings[1].Line)
	}
	if findings[0].Severity != SeverityError {
		t.Errorf("a formatting difference should be an error, got %v", findings[0].Severity)
	}
}

func TestReportFailsOnErrorsOnly(t *testing.T) {
	warning := Finding{File: "a.sh", Line: 1, Rule: "BSG046", Severity: SeverityWarning, Level: "warning", Message: "consider"}
	failure := Finding{File: "a.sh", Line: 2, Rule: "BSG001", Severity: SeverityError, Level: "error", Message: "missing keyword"}

	var out bytes.Buffer
	failed, err := Report{Findings: []Finding{warning}, Format: "text"}.Write(&out)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if failed {
		t.Error("a warning alone should not fail the run")
	}

	out.Reset()
	failed, err = Report{Findings: []Finding{warning}, Format: "text", WarningsAsErrors: true}.Write(&out)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if !failed {
		t.Error("--warnings-as-errors should fail the run on a warning")
	}

	out.Reset()
	failed, err = Report{Findings: []Finding{failure, warning}, Format: "text"}.Write(&out)
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	if !failed {
		t.Error("an error should fail the run")
	}
	if !strings.Contains(out.String(), "1 error(s), 1 warning(s)") {
		t.Errorf("summary missing from report:\n%s", out.String())
	}
}

func TestSortOrdersByFileThenPosition(t *testing.T) {
	findings := []Finding{
		{File: "b.sh", Line: 1},
		{File: "a.sh", Line: 9},
		{File: "a.sh", Line: 2, Column: 5},
		{File: "a.sh", Line: 2, Column: 1},
	}
	Sort(findings)
	want := []string{"a.sh:2:1", "a.sh:2:5", "a.sh:9:0", "b.sh:1:0"}
	for i, finding := range findings {
		got := finding.File + ":" + itoa(finding.Line) + ":" + itoa(finding.Column)
		if got != want[i] {
			t.Errorf("position %d: got %s, want %s", i, got, want[i])
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for ; n > 0; n /= 10 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
	}
	return string(digits)
}
