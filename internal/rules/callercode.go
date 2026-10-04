package rules

import (
	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG015",
		Section:  sectionCallerNames,
		Severity: lint.SeverityWarning,
		Doc:      "Prefix the locals of a function that runs code its caller passed",
		Check:    checkCallerCodeLocals,
	})
}

func checkCallerCodeLocals(f *File, r *Reporter) {
	p := f.project()
	eachFunc(f, func(decl *syntax.FuncDecl) {
		// The last place the function runs its caller's code: a local declared
		// after it can no longer be reached by that code.
		var last syntax.Pos
		fed := callerFed(decl)
		allCalls(decl.Body, func(call *syntax.CallExpr, callee string) {
			if runsCallerCode(call, fed) || (callee != "" && p.RunsCallerCode(callee) && passesCallerCode(call, fed)) {
				if call.Pos().After(last) {
					last = call.Pos()
				}
			}
		})
		if !last.IsValid() {
			return
		}
		names, at := funcLocals(decl)
		for _, name := range names {
			if isPrivateName(name) || at[name].After(last) {
				continue
			}
			r.At(at[name], "%q runs code its caller passed, and that code sees %q and can change it; prefix the local, as in `__%s_%s`",
				decl.Name.Value, name, privatePrefix(f, decl), name)
		}
	})
}
