package rules

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG080",
		Section:  sectionEval,
		Severity: lint.SeverityWarning,
		Doc:      "Do not splice a variable into a string that `eval`, `dybatpho::dry_run` or `bash -c` will parse",
		Check:    checkSplicedCode,
	})
}

// shellRunners are the commands that take a whole program as one string after
// `-c`.
var shellRunners = map[string]bool{"bash": true, "sh": true, "dash": true, "zsh": true, "ksh": true}

func checkSplicedCode(f *File, r *Reporter) {
	eachFunc(f, func(decl *syntax.FuncDecl) {
		quoted := quotedVars(decl.Body)
		allCalls(decl.Body, func(call *syntax.CallExpr, name string) {
			for _, word := range codeStrings(call, name) {
				if v := splicedVar(word, quoted); v != "" {
					r.At(word.Pos(), "%q is spliced into a string that `%s` parses as code, so a space splits it and `$(...)` or `;` in it runs; pass the arguments as separate words, or quote the value with `printf %%q` or `${%s@Q}`",
						v, name, v)
				}
			}
		})
	})
}

// codeStrings returns the words a call parses as shell code.
func codeStrings(call *syntax.CallExpr, name string) []*syntax.Word {
	args := call.Args[1:]
	switch {
	case name == "eval":
		return args
	case name == "dybatpho::dry_run":
		// With one argument the helper evaluates it; with several it runs them as
		// words, which is the safe form.
		if len(args) == 1 {
			return args
		}
	case shellRunners[name]:
		for i, arg := range args {
			lit := wordLiteral(arg)
			if strings.HasPrefix(lit, "-") && !strings.HasPrefix(lit, "--") && strings.Contains(lit, "c") && i+1 < len(args) {
				return args[i+1 : i+2]
			}
		}
	}
	return nil
}

// splicedVar returns the first variable a word splices into literal text, or
// the empty string when the word is a single expansion, plain text, or only
// splices values quoted for the shell.
func splicedVar(word *syntax.Word, quoted map[string]bool) string {
	hasText := false
	var spliced string
	syntax.Walk(word, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.Lit:
			if strings.TrimSpace(n.Value) != "" {
				hasText = true
			}
		case *syntax.SglQuoted:
			if strings.TrimSpace(n.Value) != "" {
				hasText = true
			}
		case *syntax.CmdSubst:
			return false
		case *syntax.ParamExp:
			if n.Param == nil || isShellQuoted(n) || quoted[n.Param.Value] {
				return false
			}
			if spliced == "" {
				spliced = n.Param.Value
			}
			return false
		}
		return true
	})
	if !hasText {
		return ""
	}
	return spliced
}

// isShellQuoted reports whether an expansion quotes its value for the shell,
// as `${x@Q}` does.
func isShellQuoted(pe *syntax.ParamExp) bool {
	return pe.Exp != nil && pe.Exp.Op == syntax.OtherParamOps && pe.Exp.Word != nil && wordLiteral(pe.Exp.Word) == "Q"
}

// quotedVars returns the variables a function body, or a file, fills with
// `printf %q`, which are safe to splice into code.
func quotedVars(root syntax.Node) map[string]bool {
	out := map[string]bool{}
	syntax.Walk(root, func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.CallExpr:
			if callName(n) != "printf" {
				return true
			}
			for i := 1; i+2 < len(n.Args); i++ {
				if wordLiteral(n.Args[i]) == "-v" && strings.Contains(wordLiteral(n.Args[i+2]), "%q") {
					out[wordLiteral(n.Args[i+1])] = true
				}
			}
		case *syntax.Assign:
			if n.Name != nil && n.Value != nil && strings.Contains(wordSource(n.Value), "%q") {
				out[n.Name.Value] = true
			}
		}
		return true
	})
	return out
}

// wordSource renders a word back to shell source.
func wordSource(word *syntax.Word) string {
	var b strings.Builder
	syntax.NewPrinter().Print(&b, word)
	return b.String()
}
