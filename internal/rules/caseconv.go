package rules

import (
	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionCaseConversion = "Features and Bugs > Case Conversion"

func init() {
	register(
		Rule{
			Code:     "BSG145",
			Section:  sectionCaseConversion,
			Severity: lint.SeverityWarning,
			Doc:      "Change the case of a variable with `${var,,}` or `${var^^}`, not `tr`",
			Check:    checkTrCase,
		},
		Rule{
			Code:     "BSG146",
			Section:  sectionCaseConversion,
			Severity: lint.SeverityWarning,
			Doc:      "Turn `shopt -s nocasematch` off again with `shopt -u`, or set it in a subshell",
			Check:    checkNocaseScope,
		},
	)
}

var (
	upperSets = map[string]bool{"[:upper:]": true, "[[:upper:]]": true, "A-Z": true, "[A-Z]": true}
	lowerSets = map[string]bool{"[:lower:]": true, "[[:lower:]]": true, "a-z": true, "[a-z]": true}
)

// checkTrCase reports `tr` changing the case of one value: fed a here-string,
// or piped from a single `echo` or `printf`. A `tr` that reads a file or a
// stream does what Bash expansion cannot, and is left alone.
func checkTrCase(f *File, r *Reporter) {
	up := f.parents()
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "tr" || !changesCase(call) {
			return
		}
		stmt, ok := up[call].(*syntax.Stmt)
		if !ok || !(hereString(stmt) || pipedFromPrint(stmt, up)) {
			return
		}
		reportWithDybatpho(f, r, call.Pos(), "or `dybatpho::lower`, `dybatpho::upper` where a command fits better", "`tr` starts a process and a pipe to change the case of one value; write `${var,,}` for lower case or `${var^^}` for upper case")
	})
}

// changesCase reports whether a `tr` maps an upper case set to a lower case
// one, or the reverse.
func changesCase(call *syntax.CallExpr) bool {
	var sets []string
	for _, arg := range call.Args[1:] {
		if lit := wordLiteral(arg); lit != "" && lit[0] != '-' {
			sets = append(sets, lit)
		}
	}
	if len(sets) != 2 {
		return false
	}
	return upperSets[sets[0]] && lowerSets[sets[1]] || lowerSets[sets[0]] && upperSets[sets[1]]
}

func hereString(stmt *syntax.Stmt) bool {
	for _, redir := range stmt.Redirs {
		if redir.Op == syntax.WordHdoc {
			return true
		}
	}
	return false
}

// pipedFromPrint reports whether a statement is the right side of a pipe
// whose left side is one `echo` or `printf`.
func pipedFromPrint(stmt *syntax.Stmt, up map[syntax.Node]syntax.Node) bool {
	bin, ok := up[stmt].(*syntax.BinaryCmd)
	if !ok || bin.Y != stmt || (bin.Op != syntax.Pipe && bin.Op != syntax.PipeAll) {
		return false
	}
	left, ok := bin.X.Cmd.(*syntax.CallExpr)
	if !ok {
		return false
	}
	name := callName(left)
	return name == "echo" || name == "printf"
}

var nocaseOptions = map[string]bool{"nocasematch": true, "nocaseglob": true}

// checkNocaseScope reports `shopt -s nocasematch` (or `nocaseglob`) in an
// entrypoint, in a function or at the top level, that nothing turns off
// again: every `[[ == ]]`, `case` and glob after it compares without case.
// BSG092 covers a library.
func checkNocaseScope(f *File, r *Reporter) {
	if f.Role == RoleLibrary {
		return
	}
	up := f.parents()
	check := func(scope syntax.Node) {
		// `local -` restores only `set` options, not `shopt` ones; saving
		// them with `shopt -p`, to restore later, is the other way out.
		saved := false
		var set []*syntax.CallExpr
		var setOption []string
		unset := map[string][]syntax.Pos{}
		allCalls(scope, func(call *syntax.CallExpr, name string) {
			if name != "shopt" {
				return
			}
			saved = saved || hasFlag(call, 'p')
			if inSubshell(call, up) {
				return
			}
			for _, arg := range call.Args[1:] {
				option := wordLiteral(arg)
				if !nocaseOptions[option] {
					continue
				}
				switch {
				case hasFlag(call, 's'):
					set = append(set, call)
					setOption = append(setOption, option)
				case hasFlag(call, 'u'):
					unset[option] = append(unset[option], call.Pos())
				}
			}
		})
		for i, call := range set {
			restored := saved
			for _, pos := range unset[setOption[i]] {
				if pos.After(call.Pos()) {
					restored = true
				}
			}
			if !restored {
				r.At(call.Pos(), "`shopt -s %s` changes every comparison after it, not only the one it was meant for, and `local -` does not put it back; turn it off with `shopt -u %s` right after the comparison, or set it in a `( ... )` subshell",
					setOption[i], setOption[i])
			}
		}
	}
	eachFunc(f, func(decl *syntax.FuncDecl) { check(decl) })
	check(f.Syntax)
}
