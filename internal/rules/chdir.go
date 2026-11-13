package rules

import (
	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionChangingDirectory = "Calling Commands > Changing Directory"

func init() {
	register(Rule{
		Code:     "BSG147",
		Section:  sectionChangingDirectory,
		Severity: lint.SeverityWarning,
		Doc:      "Change directory in a subshell, or pass the directory to the tool, instead of `pushd`, `popd` or `cd -`",
		Check:    checkDirStack,
	})
}

// checkDirStack reports `pushd`, `popd` and `cd -`. They print the directory
// on STDOUT, and under `set -e` a failure between the move and the move back
// leaves the script in the wrong place. A `pushd` with no `popd` in a library
// function is BSG091's to report.
func checkDirStack(f *File, r *Reporter) {
	const advice = "run the step in `( cd -P -- \"${dir}\" && ... )`, or pass the directory to the tool (`git -C`, `make -C`, `tar -C`)"
	report := func(call *syntax.CallExpr, what string) {
		r.At(call.Pos(), "`%s` prints the directory on STDOUT and, when a command in between fails, leaves the script in the wrong place; %s", what, advice)
	}
	check := func(scope syntax.Node) {
		pushd, popd := false, false
		allCalls(scope, func(_ *syntax.CallExpr, name string) {
			pushd = pushd || name == "pushd"
			popd = popd || name == "popd"
		})
		allCalls(scope, func(call *syntax.CallExpr, name string) {
			switch name {
			case "pushd":
				if f.Role == RoleLibrary && !popd {
					return
				}
				report(call, "pushd")
			case "popd":
				// A `pushd` in the same scope already carries the finding.
				if !pushd {
					report(call, "popd")
				}
			case "cd":
				if len(call.Args) > 1 && wordLiteral(call.Args[len(call.Args)-1]) == "-" {
					report(call, "cd -")
				}
			}
		})
	}
	eachFunc(f, func(decl *syntax.FuncDecl) { check(decl) })
	check(f.Syntax)
}
