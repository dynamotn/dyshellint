package rules_test

import (
	"reflect"
	"sync"
	"testing"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
	"gitlab.com/dynamo-tools/dyshellint/internal/rules"
)

// usagePrints defines its own `_usage`, which only prints, and captures it.
const usagePrints = `#!/usr/bin/env bash
function _usage {
  printf 'usage\n'
}
function _main {
  local text
  text="$(_usage)"
  printf '%s\n' "${text}"
}
_main
`

// TestUnlinkedFileIsSafeConcurrently runs the rules over a file no run linked
// from several goroutines at once: each builds the project of the file alone,
// which must happen once.
func TestUnlinkedFileIsSafeConcurrently(t *testing.T) {
	file, err := rules.NewFile("b.sh", []byte(usagePrints), true)
	if err != nil {
		t.Fatal(err)
	}
	all := rules.All()
	got := make([][]lint.Finding, 8)
	var wg sync.WaitGroup
	for i := range got {
		wg.Go(func() { got[i] = rules.Run(file, all) })
	}
	wg.Wait()
	for i := range got {
		lint.Sort(got[i])
	}
	for i := range got {
		if !reflect.DeepEqual(got[i], got[0]) {
			t.Errorf("run %d found\n%v\nwant\n%v", i, got[i], got[0])
		}
	}
}
