package rules_test

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"sync"
	"testing"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
	"gitlab.com/dynamo-tools/dyshellint/internal/rules"
)

// source is one in-memory file of a run, its path and its content.
type source struct{ path, src string }

// linkAndRun parses the sources, links them in the order given and returns
// the findings of one rule, as "path:line:column", for every file.
func linkAndRun(t *testing.T, code string, sources ...source) []string {
	t.Helper()
	var files []*rules.File
	for _, s := range sources {
		file, err := rules.NewFile(s.path, []byte(s.src), true)
		if err != nil {
			t.Fatalf("parse %s: %v", s.path, err)
		}
		files = append(files, file)
	}
	rules.Link(files)
	var selected []rules.Rule
	for _, rule := range rules.All() {
		if rule.Code == code {
			selected = append(selected, rule)
		}
	}
	var out []string
	for _, file := range files {
		for _, finding := range rules.Run(file, selected) {
			out = append(out, fmt.Sprintf("%s:%d:%d", finding.File, finding.Line, finding.Column))
		}
	}
	slices.Sort(out)
	return out
}

// usageDies defines a `_usage` that ends the script.
const usageDies = `#!/usr/bin/env bash
function _usage {
  printf 'usage\n'
  exit 2
}
_usage
`

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

// usageCalls calls a `_usage` it does not define.
const usageCalls = `#!/usr/bin/env bash
function _main {
  local text
  text="$(_usage)"
  printf '%s\n' "${text}"
}
_main
`

// TestOwnDefinitionFirst expects a call to reach the definition of its own
// file before the one of another file, whatever order the files come in.
func TestOwnDefinitionFirst(t *testing.T) {
	a, b := source{"a.sh", usageDies}, source{"b.sh", usagePrints}
	if got := linkAndRun(t, "BSG047", b, a); len(got) != 0 {
		t.Errorf("b.sh a.sh: b.sh calls its own _usage, which cannot die, got %v", got)
	}
	if got := linkAndRun(t, "BSG047", a, b); len(got) != 0 {
		t.Errorf("a.sh b.sh: b.sh calls its own _usage, which cannot die, got %v", got)
	}
}

// TestOtherDefinitionIsStable expects a call to a function the file does not
// define to reach the same definition whatever order the files come in.
func TestOtherDefinitionIsStable(t *testing.T) {
	a, b, c := source{"a.sh", usageDies}, source{"b.sh", usagePrints}, source{"c.sh", usageCalls}
	want := linkAndRun(t, "BSG047", a, b, c)
	for _, order := range [][]source{{a, c, b}, {b, a, c}, {b, c, a}, {c, a, b}, {c, b, a}} {
		if got := linkAndRun(t, "BSG047", order...); !reflect.DeepEqual(got, want) {
			t.Errorf("order %s %s %s: got %v, want %v", order[0].path, order[1].path, order[2].path, got, want)
		}
	}
}

// TestLongCallChain follows a call through thousands of functions to the one
// that dies, which a pass over every edge per level made take seconds.
func TestLongCallChain(t *testing.T) {
	const depth = 8000
	var b strings.Builder
	b.WriteString("#!/usr/bin/env bash\n")
	for i := range depth - 1 {
		fmt.Fprintf(&b, "function _f%d {\n  _f%d\n}\n", i, i+1)
	}
	fmt.Fprintf(&b, "function _f%d {\n  exit 1\n}\n", depth-1)
	b.WriteString("function _main {\n  local text\n  text=\"$(_f0)\"\n  printf '%s\\n' \"${text}\"\n}\n_main\n")
	if got := linkAndRun(t, "BSG047", source{"chain.sh", b.String()}); len(got) != 1 {
		t.Errorf("want the substitution of _f0 reported once, got %v", got)
	}
}

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
