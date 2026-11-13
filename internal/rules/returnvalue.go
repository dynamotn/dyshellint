package rules

import (
	"strconv"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionReturningValues = "Features and Bugs > Returning Values"

func init() {
	register(Rule{
		Code:     "BSG148",
		Section:  sectionReturningValues,
		Severity: lint.SeverityError,
		Doc:      "Do not return data through the status: `return` takes a number from 0 to 255, never a word or `$(...)`",
		Check:    checkReturnData,
	})
}

// checkReturnData reports `return` given a word that is no number, or the
// output of a command. `return` takes a status from 0 to 255: a word stops
// the function with an error, and data squeezed into the status is cut
// modulo 256. A number out of that range is BSG105's to report.
func checkReturnData(f *File, r *Reporter) {
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "return" || len(call.Args) != 2 {
			return
		}
		arg := call.Args[1]
		if lit := wordLiteral(arg); lit != "" {
			if _, err := strconv.Atoi(lit); err != nil {
				r.At(arg.Pos(), "`return %s` is not a status, and Bash stops the function with \"numeric argument required\"; return 0 for success or 1 for failure, and print data on STDOUT", lit)
			}
			return
		}
		if hasCmdSubst(arg) {
			reportWithDybatpho(f, r, arg.Pos(), "or write it into a variable the caller names, after `dybatpho::expect_ref`", "`return $(...)` passes data through the exit status, which keeps only 0-255; print the data on STDOUT for the caller to capture")
		}
	})
}
