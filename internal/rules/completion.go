package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionCompletion = "Environment > Shell Completion"

func init() {
	register(Rule{
		Code:     "BSG133",
		Section:  sectionCompletion,
		Severity: lint.SeverityWarning,
		Doc:      "Generate completion from the `_spec_*` function that parses the arguments, not a hand-written `complete` function",
		Check:    checkHandWrittenCompletion,
	})
}

// checkHandWrittenCompletion reports a `complete` that registers a function, a
// word list or a command (`-F`, `-W`, `-C`) in a file whose command line is
// already declared as a spec. The two lists of options then live apart, and
// the completion keeps offering an option the parser dropped. A file without a
// spec has nothing to generate from and is left alone, as is `complete -p`,
// which only prints.
func checkHandWrittenCompletion(f *File, r *Reporter) {
	if !declaresSpec(f) {
		return
	}
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "complete" || !(hasFlag(call, 'F') || hasFlag(call, 'W') || hasFlag(call, 'C')) {
			return
		}
		reportWithDybatpho(f, r, call.Pos(),
			"print it with `dybatpho::generate_completion <spec> bash`, from a display option of the spec",
			"a hand-written completion drifts from the `_spec_*` function that parses the same options; generate it from the spec")
	})
}

// declaresSpec reports whether a file declares its command line as a spec: it
// defines a `_spec_*` function or runs one with `dybatpho::generate_from_spec`.
func declaresSpec(f *File) bool {
	found := false
	eachFunc(f, func(decl *syntax.FuncDecl) {
		found = found || strings.HasPrefix(decl.Name.Value, "_spec_")
	})
	if found {
		return true
	}
	eachCall(f, func(_ *syntax.CallExpr, name string) {
		found = found || name == "dybatpho::generate_from_spec"
	})
	return found
}
