package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/dynamo-tools/dyshellint/internal/rules"
)

// offline keeps the external tools out of a test, so that it reads only what
// the rules of the guide report.
var offline = []string{"--no-shellcheck", "--no-shfmt"}

func runWith(t *testing.T, stdin string, args ...string) (stdout, stderr string, err error) {
	t.Helper()
	var out, errOut bytes.Buffer
	err = run(args, strings.NewReader(stdin), &out, &errOut)
	return out.String(), errOut.String(), err
}

func TestHelpIsNotAFailure(t *testing.T) {
	_, stderr, err := runWith(t, "", "-h")
	if !errors.Is(err, flag.ErrHelp) {
		t.Errorf("got %v, want flag.ErrHelp, which main exits 0 on", err)
	}
	if !strings.Contains(stderr, "usage: dyshellint") {
		t.Errorf("the usage is missing:\n%s", stderr)
	}
}

func TestUnknownFormatIsRejected(t *testing.T) {
	_, _, err := runWith(t, "", append(offline, "--format", "sarif", ".")...)
	if err == nil || errors.Is(err, errFindings) {
		t.Errorf("got %v, want a usage error", err)
	}
}

func TestStdinCannotBeMixedWithPaths(t *testing.T) {
	_, _, err := runWith(t, "", append(offline, "-", "main.go")...)
	if err == nil || errors.Is(err, errFindings) {
		t.Errorf("got %v, want a usage error", err)
	}
}

func TestStdinIsReportedUnderItsName(t *testing.T) {
	script := "#!/usr/bin/env bash\nfunction Demo() {\n  echo hi\n}\n"
	stdout, _, err := runWith(t, script, append(offline, "--format", "json", "--stdin-filename", "scripts/demo.sh", "-")...)
	if !errors.Is(err, errFindings) {
		t.Fatalf("got %v, want findings", err)
	}
	if !strings.Contains(stdout, `"file": "scripts/demo.sh"`) {
		t.Errorf("findings are not under the given name:\n%s", stdout)
	}
}

func TestCleanRunPrintsAnEmptyList(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("not a script\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, stderr, err := runWith(t, "", append(offline, dir)...)
	if err != nil {
		t.Fatalf("got %v", err)
	}
	if !strings.Contains(stderr, "no shell files found") {
		t.Errorf("got %q", stderr)
	}
}

func TestUnknownRuleIsRejected(t *testing.T) {
	_, _, err := runWith(t, "", append(offline, "--rules", "BSG999", ".")...)
	if err == nil || !strings.Contains(err.Error(), "BSG999") {
		t.Errorf("got %v, want an unknown rule error", err)
	}
}

func TestListRulesRejectsUnknownFormat(t *testing.T) {
	_, _, err := runWith(t, "", "--list-rules", "--format", "sarif")
	if err == nil || !strings.Contains(err.Error(), "sarif") {
		t.Errorf("got %v, want an unknown format error", err)
	}
}

func TestListRulesJSON(t *testing.T) {
	var out bytes.Buffer
	if err := listRules(&out, "json"); err != nil {
		t.Fatal(err)
	}
	var listed []listedRule
	if err := json.Unmarshal(out.Bytes(), &listed); err != nil {
		t.Fatalf("not json: %v\n%s", err, out.String())
	}
	if len(listed) != len(rules.All()) {
		t.Fatalf("listed %d rules, registry has %d", len(listed), len(rules.All()))
	}
	for i, rule := range listed {
		if i > 0 && rule.Code <= listed[i-1].Code {
			t.Errorf("%s is listed after %s", rule.Code, listed[i-1].Code)
		}
		if rule.Section == "" || rule.Doc == "" || (rule.Severity != "error" && rule.Severity != "warning") {
			t.Errorf("%s is incomplete: %+v", rule.Code, rule)
		}
	}
}

func TestListRulesText(t *testing.T) {
	var out bytes.Buffer
	if err := listRules(&out, "text"); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out.String(), "BSG001  error") {
		t.Fatalf("unexpected first line: %q", strings.SplitN(out.String(), "\n", 2)[0])
	}
}
