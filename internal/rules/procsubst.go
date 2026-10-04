package rules

import (
	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionProcessSubst = "Features and Bugs > Process Substitution"

func init() {
	register(Rule{
		Code:     "BSG048",
		Section:  sectionProcessSubst,
		Severity: lint.SeverityWarning,
		Doc:      "Do not feed `<(...)` from a function that can fail; capture its output and check its status first",
		Check:    checkFailingProducer,
	})
}

func checkFailingProducer(f *File, r *Reporter) {
	p := f.project()
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		ps, ok := node.(*syntax.ProcSubst)
		if !ok || ps.Op != syntax.CmdIn {
			return true
		}
		for _, stmt := range ps.Stmts {
			for _, call := range leadingCalls(stmt) {
				name := callName(call)
				if name == "" || !p.CanFail(name) {
					continue
				}
				r.At(ps.Pos(), "%q can fail, but nothing sees the status of `<(...)`: a failure reads as empty output; capture the output in a variable or array first and check the status there",
					name)
				return true
			}
		}
		return true
	})
}
