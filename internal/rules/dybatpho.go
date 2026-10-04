package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const (
	sectionArguments    = "Environment > Script Argument Control"
	sectionErrorHandler = "Calling Commands > Error Handling"
)

func init() {
	register(
		Rule{
			Code:     "BSG050",
			Section:  sectionArguments,
			Severity: lint.SeverityError,
			Doc:      "Declare the arguments of a function with `dybatpho::expect_args`",
			Check:    checkExpectArgs,
		},
		Rule{
			Code:     "BSG051",
			Section:  sectionArguments,
			Severity: lint.SeverityError,
			Doc:      "Parse the command line from a `_spec_*` function, not by hand",
			Check:    checkHandRolledParsing,
		},
		Rule{
			Code:     "BSG052",
			Section:  sectionArguments,
			Severity: lint.SeverityWarning,
			Doc:      "Always offer `--help` in a spec function",
			Check:    checkSpecHelp,
		},
		Rule{
			Code:     "BSG053",
			Section:  sectionErrorHandler,
			Severity: lint.SeverityError,
			Doc:      "Install the dybatpho handlers once, at the top of an entrypoint",
			Check:    checkErrorHandlers,
		},
	)
}

// positionalDigit matches a read of `$1` … `$9`, braced or not.
var positionalDigit = regexp.MustCompile(`\$\{?[1-9]\}?`)

func checkExpectArgs(f *File, r *Reporter) {
	if !f.UsesDybatpho {
		return
	}
	eachFunc(f, func(decl *syntax.FuncDecl) {
		body := f.Text(decl.Body.Pos(), decl.Body.End())
		if strings.Contains(body, "expect_args") || !positionalDigit.MatchString(body) {
			return
		}
		r.At(decl.Position, "%q reads its arguments positionally; name them with `dybatpho::expect_args name... -- \"$@\"`",
			decl.Name.Value)
	})
}

func checkHandRolledParsing(f *File, r *Reporter) {
	if f.Role != RoleEntrypoint {
		return
	}
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "getopts" {
			return
		}
		r.At(call.Pos(), "`getopts` reimplements what `dybatpho::opts::*` already does; declare the interface in a `_spec_*` function and run it with `dybatpho::generate_from_spec`")
	})
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		while, ok := node.(*syntax.WhileClause)
		if !ok {
			return true
		}
		condition := f.Text(while.WhilePos, while.DoPos)
		if !strings.Contains(condition, "$#") {
			return true
		}
		body := f.Text(while.DoPos, while.DonePos)
		if !strings.Contains(body, "shift") {
			return true
		}
		r.At(while.Pos(), "hand-rolled option parsing drifts from every other script in the repository; declare the interface in a `_spec_*` function instead")
		return true
	})
}

func checkSpecHelp(f *File, r *Reporter) {
	eachFunc(f, func(decl *syntax.FuncDecl) {
		body := f.Text(decl.Body.Pos(), decl.Body.End())
		if !strings.Contains(body, "dybatpho::opts::setup") {
			return
		}
		if strings.Contains(body, "--help") {
			return
		}
		r.At(decl.Position, "%q declares a command line without `--help`; add `dybatpho::opts::disp \"Show help\" --help action:\"dybatpho::generate_help %s\"`",
			decl.Name.Value, decl.Name.Value)
	})
}

func checkErrorHandlers(f *File, r *Reporter) {
	if f.Role != RoleEntrypoint || !f.UsesDybatpho {
		return
	}
	src := string(f.Src)
	if strings.Contains(src, "register_common_handlers") || strings.Contains(src, "register_err_handler") {
		return
	}
	r.AtLine(1, "no error handler is installed; call `dybatpho::register_common_handlers` right after sourcing dybatpho")
}

// withDybatpho appends the dybatpho helper that solves a finding to its
// message, when the file loads dybatpho, so the advice names a function the
// script can already call. msg is a format string; advice is taken literally.
func withDybatpho(f *File, msg, advice string) string {
	if !f.UsesDybatpho {
		return msg
	}
	return msg + "; " + strings.ReplaceAll(advice, "%", "%%")
}

// reportWithDybatpho reports a finding at pos, adding the dybatpho advice when
// the file loads dybatpho. The format stays a constant at the call site, so
// `go vet` still checks it against the arguments.
func reportWithDybatpho(f *File, r *Reporter, pos syntax.Pos, advice, format string, args ...any) {
	r.At(pos, withDybatpho(f, format, advice), args...)
}
