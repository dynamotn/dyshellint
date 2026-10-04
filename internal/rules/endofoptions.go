package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionEndOfOptions = "Calling Commands > End of Options"

func init() {
	register(Rule{
		Code:     "BSG057",
		Section:  sectionEndOfOptions,
		Severity: lint.SeverityWarning,
		Doc:      "Put `--` before an operand that is a single expansion, which may start with `-`",
		Check:    checkEndOfOptions,
	})
}

// optionTakers are the commands that read a leading `-` in an operand as an
// option, with the options of each that take a value.
var optionTakers = map[string]string{
	"rm": "", "mv": "tS", "cp": "tS", "ln": "tS", "touch": "dtr", "mkdir": "m",
	"cat": "", "ls": "IwT", "chmod": "", "chown": "", "grep": "efmABCd", "sed": "efl",
}

// skipsFirst are the commands whose first operand is not a path: a mode or an
// owner.
var skipsFirst = map[string]bool{"chmod": true, "chown": true}

func checkEndOfOptions(f *File, r *Reporter) {
	inputs := map[*syntax.FuncDecl]map[string]bool{}
	up := parents(f.Syntax)
	eachCall(f, func(call *syntax.CallExpr, name string) {
		valued, ok := optionTakers[name]
		if !ok {
			return
		}
		operand := 0
		for i := 1; i < len(call.Args); i++ {
			arg := call.Args[i]
			lit := wordLiteral(arg)
			if lit == "--" {
				return
			}
			if strings.HasPrefix(lit, "-") && len(lit) > 1 {
				// `grep -e PATTERN` names the pattern, so no operand can be
				// mistaken for it; an option with a value takes the next word.
				if (name == "grep" || name == "sed") && (strings.ContainsAny(lit[1:], "ef") || strings.HasPrefix(lit, "--regexp") || strings.HasPrefix(lit, "--expression")) {
					return
				}
				if !strings.HasPrefix(lit, "--") && strings.ContainsAny(lit[len(lit)-1:], valued) {
					i++
				}
				continue
			}
			operand++
			if operand == 1 && skipsFirst[name] {
				continue
			}
			// Only the first operand can still be taken for an option once
			// the command has seen a non-option word on GNU, and it is the one
			// a value starting with `-` is most often handed in.
			if isScalarExpansion(arg) && fromInput(wordParam(arg).Param.Value, enclosingFunc(call, up), inputs) {
				r.At(arg.Pos(), "`%s` reads a value that starts with `-` as an option; write `%s -- %s`",
					name, name, wordSource(arg))
			}
			return
		}
	})
}

// isScalarExpansion reports whether a word is one scalar expansion and nothing
// else, such as `"$file"` or `"${path}"`, so its whole value decides how it
// starts.
func isScalarExpansion(word *syntax.Word) bool {
	pe := wordParam(word)
	if pe == nil || pe.Param == nil || pe.Length || pe.Excl || pe.Exp != nil {
		return false
	}
	switch pe.Param.Value {
	case "@", "*", "#", "?", "$", "!", "0":
		return false
	}
	if pe.Index != nil {
		if w, ok := pe.Index.(*syntax.Word); ok {
			if lit := wordLiteral(w); lit == "@" || lit == "*" {
				return false
			}
		}
	}
	return true
}

// fromInput reports whether a variable holds a value the caller or the input
// supplied, which is where a leading `-` comes from: a positional parameter, a
// variable filled from one or bound by `dybatpho::expect_args`, a `read`
// target, or the variable of a loop over `"$@"`.
func fromInput(name string, decl *syntax.FuncDecl, cache map[*syntax.FuncDecl]map[string]bool) bool {
	if len(name) == 1 && name[0] >= '1' && name[0] <= '9' {
		return true
	}
	if decl == nil {
		return false
	}
	set, ok := cache[decl]
	if !ok {
		set = map[string]bool{}
		for v := range fedIndex(decl) {
			set[v] = true
		}
		allCalls(decl.Body, func(call *syntax.CallExpr, callee string) {
			if callee == "read" {
				builtinTargets(call, func(target string, _ syntax.Pos) { set[target] = true })
			}
		})
		syntax.Walk(decl.Body, func(node syntax.Node) bool {
			clause, ok := node.(*syntax.ForClause)
			if !ok {
				return true
			}
			if iter, ok := clause.Loop.(*syntax.WordIter); ok && iter.Name != nil {
				if !iter.InPos.IsValid() {
					set[iter.Name.Value] = true
				}
				for _, item := range iter.Items {
					if pe := wordParam(item); pe != nil && pe.Param != nil && (pe.Param.Value == "@" || pe.Param.Value == "*") {
						set[iter.Name.Value] = true
					}
				}
			}
			return true
		})
		cache[decl] = set
	}
	return set[name]
}
