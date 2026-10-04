package rules

import (
	"strconv"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG054",
		Section:  sectionErrorHandler,
		Severity: lint.SeverityWarning,
		Doc:      "Name the function at fault explicitly rather than reading `FUNCNAME[2]` or deeper",
		Check:    checkDeepFuncname,
	})
}

// checkDeepFuncname reports `FUNCNAME[N]` with a literal N of two or more. How
// deep the function that matters sits depends on every call between, so a
// helper reached one frame further than expected names the wrong function.
func checkDeepFuncname(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		pe, ok := node.(*syntax.ParamExp)
		if !ok || pe.Param == nil || pe.Param.Value != "FUNCNAME" || pe.Index == nil {
			return true
		}
		word, ok := pe.Index.(*syntax.Word)
		if !ok {
			return true
		}
		depth, err := strconv.Atoi(wordLiteral(word))
		if err != nil || depth < 2 {
			return true
		}
		r.At(pe.Pos(), "`FUNCNAME[%d]` names whatever sits %d frames up, which shifts when a helper is called from another place; have the public function pass its own `${FUNCNAME[0]}` down instead",
			depth, depth)
		return true
	})
}
