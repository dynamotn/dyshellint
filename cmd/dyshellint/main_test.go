package main

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
