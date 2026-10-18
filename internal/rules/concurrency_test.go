package rules_test

import (
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
	"gitlab.com/dynamo-tools/dyshellint/internal/rules"
)

// TestRunIsSafeConcurrently runs every rule over every fixture of one linked
// project from many goroutines at once, the way the command does, and expects
// what a run one file at a time finds. Under -race it also catches a rule that
// writes to state the files share.
func TestRunIsSafeConcurrently(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("testdata", "*.*"))
	if err != nil {
		t.Fatal(err)
	}
	var files []*rules.File
	for _, path := range paths {
		if ext := filepath.Ext(path); ext != ".sh" && ext != ".bats" {
			continue
		}
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		file, err := rules.NewFile(path, src, true)
		if err != nil {
			continue // A fixture of the parse error itself.
		}
		files = append(files, file)
	}
	rules.Link(files)
	all := rules.All()

	want := make([][]lint.Finding, len(files))
	for i, file := range files {
		want[i] = rules.Run(file, all)
	}
	got := make([][]lint.Finding, len(files))
	var wg sync.WaitGroup
	for i, file := range files {
		wg.Go(func() { got[i] = rules.Run(file, all) })
	}
	wg.Wait()
	for i := range files {
		if !reflect.DeepEqual(got[i], want[i]) {
			t.Errorf("%s: a concurrent run found\n%v\nwant\n%v", files[i].Path, got[i], want[i])
		}
	}
}
