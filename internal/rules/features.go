package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const (
	sectionEval       = "Features and Bugs > Eval is Evil"
	sectionPipeWhile  = "Features and Bugs > Pipes to While"
	sectionForLoops   = "Features and Bugs > For Loops"
	sectionExpansion  = "Formatting > Variable Expansion"
	sectionReturnVals = "Calling Commands > Checking Return Values"
	sectionTempFiles  = "Script Stabilization > Safely Creating Temporary Files"
)

func init() {
	register(
		Rule{
			Code:     "BSG040",
			Section:  sectionEval,
			Severity: lint.SeverityError,
			Doc:      "Do not use `eval`",
			Check:    checkEval,
		},
		Rule{
			Code:     "BSG041",
			Section:  sectionPipeWhile,
			Severity: lint.SeverityError,
			Doc:      "Do not pipe into a `while` loop; feed it with `< <(command)`",
			Check:    checkPipeToWhile,
		},
		Rule{
			Code:     "BSG042",
			Section:  sectionForLoops,
			Severity: lint.SeverityError,
			Doc:      "Do not loop over the unquoted output of a command",
			Check:    checkForOverCommand,
		},
		Rule{
			Code:     "BSG043",
			Section:  sectionExpansion,
			Severity: lint.SeverityError,
			Doc:      "Do not brace positional or special parameters unless it is necessary",
			Check:    checkBracedSpecialParams,
		},
		Rule{
			Code:     "BSG044",
			Section:  sectionReturnVals,
			Severity: lint.SeverityError,
			Doc:      "Do not inspect `$?` in a separate statement, and do not rely on `PIPESTATUS`",
			Check:    checkExitStatusInspection,
		},
		Rule{
			Code:     "BSG045",
			Section:  sectionTempFiles,
			Severity: lint.SeverityError,
			Doc:      "Do not build a temporary path from `$$`, a timestamp or a fixed name",
			Check:    checkHandmadeTempPath,
		},
		Rule{
			Code:     "BSG046",
			Section:  sectionTempFiles,
			Severity: lint.SeverityWarning,
			Doc:      "Create temporary files with `dybatpho::create_temp`, which registers its own cleanup",
			Check:    checkMktemp,
		},
	)
}

func checkEval(f *File, r *Reporter) {
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "eval" {
			return
		}
		r.At(call.Pos(), "`eval` hides what will run; build the command in an array, or use a nameref for an indirect variable")
	})
}

func checkPipeToWhile(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		binary, ok := node.(*syntax.BinaryCmd)
		if !ok || (binary.Op != syntax.Pipe && binary.Op != syntax.PipeAll) {
			return true
		}
		if _, ok := binary.Y.Cmd.(*syntax.WhileClause); !ok {
			return true
		}
		r.At(binary.Y.Pos(), "the loop runs in a subshell, so every variable it sets is lost; write `while read -r line; do ...; done < <(command)`")
		return true
	})
}

func checkForOverCommand(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		clause, ok := node.(*syntax.ForClause)
		if !ok {
			return true
		}
		iter, ok := clause.Loop.(*syntax.WordIter)
		if !ok {
			return true
		}
		for _, item := range iter.Items {
			for _, part := range item.Parts {
				switch part.(type) {
				case *syntax.CmdSubst:
					r.At(item.Pos(), "`for %s in $(...)` splits the output on every space and then globs it; read it into an array with `readarray -t` and loop over that",
						iter.Name.Value)
					return true
				}
			}
		}
		return true
	})
}

// specialParams are the parameters the guide says to leave unbraced.
var specialParams = regexp.MustCompile(`^([0-9]|[@*#?$!_-])$`)

func checkBracedSpecialParams(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		exp, ok := node.(*syntax.ParamExp)
		if !ok || exp.Short || exp.Param == nil {
			return true
		}
		// Braces earn their place as soon as the expansion does something:
		// a default, a slice, a replacement, a length, or a two-digit index.
		if exp.Exp != nil || exp.Slice != nil || exp.Repl != nil || exp.Index != nil ||
			exp.Length || exp.Excl || exp.Width || exp.Names != 0 {
			return true
		}
		if !specialParams.MatchString(exp.Param.Value) {
			return true
		}
		r.At(exp.Pos(), "`${%s}` does not need braces; write `$%s`", exp.Param.Value, exp.Param.Value)
		return true
	})
}

func checkExitStatusInspection(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		stmt, ok := node.(*syntax.Stmt)
		if !ok {
			return true
		}
		test, ok := stmt.Cmd.(*syntax.TestClause)
		if ok {
			reportStatusRef(f, r, test)
			return true
		}
		return true
	})
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		exp, ok := node.(*syntax.ParamExp)
		if !ok || exp.Param == nil {
			return true
		}
		if exp.Param.Value == "PIPESTATUS" {
			r.At(exp.Pos(), "`PIPESTATUS` is easy to read at the wrong moment; test the command itself, or split the pipeline")
		}
		return true
	})
}

// reportStatusRef flags `[[ $? -ne 0 ]]`, the form the guide rejects.
func reportStatusRef(f *File, r *Reporter, test *syntax.TestClause) {
	syntax.Walk(test, func(node syntax.Node) bool {
		exp, ok := node.(*syntax.ParamExp)
		if !ok || exp.Param == nil || exp.Param.Value != "?" {
			return true
		}
		r.At(exp.Pos(), "`$?` read in a separate statement goes stale as soon as a line is added between the two; test the command directly with `if ! command; then`")
		return true
	})
}

// handmadeTempPath matches a temporary path built by hand, which is both a
// collision and a symlink attack.
var handmadeTempPath = regexp.MustCompile(`/tmp/[^"'\s$]*(\$\$|\$\{?RANDOM|\$\(date)|/tmp/[A-Za-z0-9_.-]+$`)

func checkHandmadeTempPath(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		word, ok := node.(*syntax.Word)
		if !ok {
			return true
		}
		text := f.Text(word.Pos(), word.End())
		if !strings.Contains(text, "/tmp/") || !handmadeTempPath.MatchString(text) {
			return true
		}
		r.At(word.Pos(), "%s is a predictable path in a world-writable directory; use `dybatpho::create_temp`, or `mktemp` with a `trap` that removes it", text)
		return false
	})
}

func checkMktemp(f *File, r *Reporter) {
	if !f.UsesDybatpho {
		return
	}
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "mktemp" {
			return
		}
		r.At(call.Pos(), "`mktemp` leaves the cleanup to you; `dybatpho::create_temp` and `dybatpho::create_temp_dir` register it at creation time")
	})
}
