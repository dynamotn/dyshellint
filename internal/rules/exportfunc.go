package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionChildFunctions = "Calling Commands > Functions in Child Processes"

func init() {
	register(Rule{
		Code:     "BSG141",
		Section:  sectionChildFunctions,
		Severity: lint.SeverityWarning,
		Doc:      "Run the script itself with a subcommand, as in `xargs -n1 \"$0\" _worker`, instead of handing a function to a child with `export -f`",
		Check:    checkExportFunctions,
	})
}

// checkExportFunctions reports `export -f`, `declare -fx` and `typeset -fx`.
// An exported function travels as a `BASH_FUNC_name%%` variable into every
// child process, not only the one meant to run it, and without the variables,
// options and other functions it calls.
func checkExportFunctions(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		clause, ok := node.(*syntax.DeclClause)
		if !ok || clause.Variant == nil {
			return true
		}
		flags := declFlags(clause)
		if !strings.Contains(flags, "f") || strings.Contains(flags, "n") {
			return true
		}
		switch clause.Variant.Value {
		case "export":
		case "declare", "typeset":
			if !strings.Contains(flags, "x") {
				return true
			}
		default:
			return true
		}
		reportWithDybatpho(f, r, clause.Pos(), "when a child Bash needs the library, `dybatpho::export_functions` lets it source the whole library through `BASH_ENV`", "an exported function reaches every child process, and runs there without the variables, options and helpers it relies on; call the script itself with a subcommand, as in `xargs -n1 \"$0\" _worker`")
		return true
	})
}
