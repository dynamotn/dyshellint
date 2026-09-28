package directive_test

import (
	"bytes"
	"strings"
	"testing"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/directive"
)

// parse builds a Set the way the linter does, from the same source twice: the
// lines a directive is read from and the tree its scope is measured against.
func parse(t *testing.T, src string) *directive.Set {
	t.Helper()
	prog, err := syntax.NewParser(syntax.KeepComments(true), syntax.Variant(syntax.LangBash)).
		Parse(bytes.NewReader([]byte(src)), "test.sh")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return directive.Parse(strings.Split(src, "\n"), prog)
}

func TestTopOfFileCoversWholeFile(t *testing.T) {
	set := parse(t, strings.Join([]string{
		"#!/usr/bin/env bash",
		"# dyshellint disable=BSG020 # the header lives in the wrapper",
		"",
		"echo one",
		"echo two",
	}, "\n"))

	for _, line := range []int{1, 4, 5} {
		if !set.Suppressed(line, "BSG020") {
			t.Errorf("line %d: BSG020 should be silenced for the whole file", line)
		}
	}
	if set.Suppressed(4, "BSG021") {
		t.Error("a directive should silence only the codes it names")
	}
}

func TestAboveACommandCoversItsBlock(t *testing.T) {
	set := parse(t, strings.Join([]string{
		"#!/usr/bin/env bash",
		"echo before",
		"# dyshellint disable=BSG010",
		"function demo() {",
		"  echo inside",
		"}",
		"echo after",
	}, "\n"))

	if set.Suppressed(2, "BSG010") {
		t.Error("a directive should not reach backwards")
	}
	for _, line := range []int{4, 5, 6} {
		if !set.Suppressed(line, "BSG010") {
			t.Errorf("line %d: the directive should cover the whole function", line)
		}
	}
	if set.Suppressed(7, "BSG010") {
		t.Error("the directive should stop at the end of the function")
	}
}

func TestTrailingCommentCoversItsLineOnly(t *testing.T) {
	set := parse(t, strings.Join([]string{
		"#!/usr/bin/env bash",
		"echo one",
		"echo two # dyshellint disable=SC2086,FMT001",
		"echo three",
	}, "\n"))

	for _, code := range []string{"SC2086", "FMT001"} {
		if !set.Suppressed(3, code) {
			t.Errorf("%s should be silenced on its own line", code)
		}
		if set.Suppressed(4, code) {
			t.Errorf("%s should not reach the next line", code)
		}
	}
}

func TestDisableAllAndCaseFolding(t *testing.T) {
	set := parse(t, strings.Join([]string{
		"#!/usr/bin/env bash",
		"echo one",
		"# DyShellint Disable=all",
		"echo two",
	}, "\n"))

	if !set.Suppressed(4, "BSG001") || !set.Suppressed(4, "SC2086") {
		t.Error("disable=all should silence every code in its scope")
	}
	if set.Suppressed(2, "BSG001") {
		t.Error("disable=all should still respect its scope")
	}
}

func TestShellCheckCommentSilencesItsOwnCodes(t *testing.T) {
	set := parse(t, strings.Join([]string{
		"#!/usr/bin/env bash",
		"echo one",
		"# shellcheck disable=SC2086 # the pattern is a word list",
		"grep -r $pattern .",
		"echo three # shellcheck disable=SC2154",
	}, "\n"))

	if !set.Suppressed(4, "SC2086") {
		t.Error("a shellcheck comment should silence the code it names")
	}
	if !set.Suppressed(5, "SC2154") {
		t.Error("a trailing shellcheck comment should silence its own line")
	}
	if set.Suppressed(4, "BSG001") || set.Suppressed(4, "FMT001") {
		t.Error("a shellcheck comment should never reach a rule of the guide")
	}
}

func TestShellCheckCommentCannotDisableTheGuide(t *testing.T) {
	set := parse(t, strings.Join([]string{
		"#!/usr/bin/env bash",
		"# shellcheck disable=all",
		"# shellcheck disable=BSG002",
		"echo one",
	}, "\n"))

	if !set.Suppressed(4, "SC2086") {
		t.Error("shellcheck disable=all should silence every SC code in scope")
	}
	for _, code := range []string{"BSG002", "FMT001"} {
		if set.Suppressed(4, code) {
			t.Errorf("%s should survive a shellcheck comment", code)
		}
	}
}

func TestPlainCommentIsNotADirective(t *testing.T) {
	set := parse(t, "#!/usr/bin/env bash\n# dyshellint is a linter\necho one\n")
	if set.Suppressed(3, "BSG001") {
		t.Error("a comment that only mentions the linter should silence nothing")
	}
}
