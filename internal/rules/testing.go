package rules

import (
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionTesting = "Testing"

func init() {
	register(Rule{
		Code:     "BSG060",
		Section:  sectionTesting,
		Severity: lint.SeverityError,
		Doc:      "Give every library a matching `test/<area>.bats`",
		Check:    checkLibraryHasTest,
	})
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
