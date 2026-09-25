package rules_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
	"gitlab.com/dynamo-tools/dyshellint/internal/rules"
)

// update rewrites the `.want` files from the current findings. Run
// `go test ./internal/rules -update` after adding a rule or a fixture, then
// read the diff: it is the documentation of what the rule now reports.
var update = flag.Bool("update", false, "rewrite the .want files from the current findings")

// TestRules runs every rule over each fixture in testdata and compares the
// findings with the `.want` file next to it. A fixture without a `.want` file
// is expected to be clean.
func TestRules(t *testing.T) {
	var fixtures []string
	err := filepath.WalkDir("testdata", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && filepath.Ext(path) == ".sh" {
			fixtures = append(fixtures, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk testdata: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatal("no fixtures found")
	}

	for _, fixture := range fixtures {
		t.Run(fixture, func(t *testing.T) {
			got := check(t, fixture)
			if *update {
				writeWant(t, fixture, got)
				return
			}
			want := readWant(t, fixture)
			if strings.Join(got, "\n") == strings.Join(want, "\n") {
				return
			}
			t.Errorf("findings do not match %s.want\n got: %s\nwant: %s",
				strings.TrimSuffix(fixture, ".sh"), format(got), format(want))
		})
	}
}

// check parses one fixture and returns its findings as "line:column:CODE".
func check(t *testing.T, path string) []string {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	file, err := rules.NewFile(path, src, info.Mode().Perm()&0o111 != 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	findings := rules.Run(file, rules.All())
	lint.Sort(findings)
	out := make([]string, 0, len(findings))
	for _, finding := range findings {
		out = append(out, fmt.Sprintf("%d:%d:%s", finding.Line, finding.Column, finding.Rule))
	}
	return out
}

// readWant loads the expected findings, treating a missing file as "clean".
func readWant(t *testing.T, fixture string) []string {
	t.Helper()
	path := strings.TrimSuffix(fixture, ".sh") + ".want"
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var want []string
	for _, line := range strings.Split(string(content), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			want = append(want, line)
		}
	}
	return want
}

// writeWant records the current findings as the expectation of a fixture.
func writeWant(t *testing.T, fixture string, got []string) {
	t.Helper()
	path := strings.TrimSuffix(fixture, ".sh") + ".want"
	if len(got) == 0 {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			t.Fatalf("remove %s: %v", path, err)
		}
		return
	}
	content := strings.Join(got, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func format(list []string) string {
	if len(list) == 0 {
		return "\n  (none)"
	}
	return "\n  " + strings.Join(list, "\n  ")
}

// TestAllRulesAreCovered fails when a rule has no fixture, so a new rule comes
// with the example that documents it.
func TestAllRulesAreCovered(t *testing.T) {
	covered := map[string]bool{}
	err := filepath.WalkDir("testdata", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".want" {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(content), "\n") {
			parts := strings.Split(strings.TrimSpace(line), ":")
			if len(parts) == 3 {
				covered[parts[2]] = true
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk testdata: %v", err)
	}
	for _, rule := range rules.All() {
		if !covered[rule.Code] {
			t.Errorf("%s has no fixture in testdata", rule.Code)
		}
	}
}
