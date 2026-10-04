package rules

import (
	"os"
	"path/filepath"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionTesting = "Testing"

const sectionStrictAssertions = "Testing > Strict Output Assertions"

func init() {
	register(
		Rule{
			Code:     "BSG060",
			Section:  sectionTesting,
			Severity: lint.SeverityError,
			Doc:      "Give every library a matching `test/<area>.bats`",
			Check:    checkLibraryHasTest,
		},
		Rule{
			Code:     "BSG061",
			Section:  sectionStrictAssertions,
			Severity: lint.SeverityError,
			Doc:      "Pass `-` to an output assertion that reads its expectation from a here document",
			Check:    checkStdinAssertion,
			Bats:     true,
		},
	)
}

// stdinAssertions are the bats-assert helpers that compare against standard
// input only when they are given `-`; without it they ignore the input and
// only check that there was output at all.
var stdinAssertions = map[string]bool{
	"assert_output": true, "refute_output": true,
	"assert_stderr": true, "refute_stderr": true,
}

func checkStdinAssertion(f *File, r *Reporter) {
	if !f.Bats {
		return
	}
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		stmt, ok := node.(*syntax.Stmt)
		if !ok {
			return true
		}
		call, ok := stmt.Cmd.(*syntax.CallExpr)
		if !ok || !stdinAssertions[callName(call)] || !readsStdinDoc(stmt) {
			return true
		}
		for _, arg := range call.Args[1:] {
			if wordLiteral(arg) == "-" {
				return true
			}
		}
		r.At(call.Pos(), "`%s` ignores the here document unless it is given `-`, and then only checks that there was output; write `%s -`",
			callName(call), callName(call))
		return true
	})
}

// readsStdinDoc reports whether a statement feeds a here document or a here
// string to its command.
func readsStdinDoc(stmt *syntax.Stmt) bool {
	for _, redir := range stmt.Redirs {
		switch redir.Op {
		case syntax.Hdoc, syntax.DashHdoc, syntax.WordHdoc:
			return true
		}
	}
	return false
}

func checkLibraryHasTest(f *File, r *Reporter) {
	if f.Role != RoleLibrary || !isLibraryPath(f.Path) {
		return
	}
	// `scripts/lib/<area>.sh` is tested by `scripts/test/<area>.bats`. Without a
	// test folder next to the library folder there is no suite to belong to yet,
	// and the rule stays quiet rather than inventing a layout.
	libDir := filepath.Dir(f.Path)
	testDir := filepath.Join(filepath.Dir(libDir), "test")
	if info, err := os.Stat(testDir); err != nil || !info.IsDir() {
		return
	}
	// The layout follows the name of the file, not the namespace its header may
	// declare: a reader looks the test up by the file they are reading.
	test := filepath.Join(testDir, BaseName(f.Path)+".bats")
	if _, err := os.Stat(test); err == nil {
		return
	}
	r.AtLine(1, "no test file for this library; add %s, written with bats", filepath.ToSlash(strings.TrimPrefix(test, "./")))
}
