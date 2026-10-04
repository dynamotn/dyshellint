package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const (
	sectionPrintingData  = "Features and Bugs > Printing Data"
	sectionHereDocuments = "Features and Bugs > Here Documents"
)

func init() {
	register(
		Rule{
			Code:     "BSG100",
			Section:  sectionPrintingData,
			Severity: lint.SeverityWarning,
			Doc:      "Print data with `printf '%s\\n'`, not `echo -e`, `echo -n` or an `echo` that starts with a variable",
			Check:    checkEchoData,
		},
		Rule{
			Code:     "BSG101",
			Section:  sectionHereDocuments,
			Severity: lint.SeverityError,
			Doc:      "Do not indent a here document with `<<-`, which strips tabs only",
			Check:    checkDashHeredoc,
		},
		Rule{
			Code:     "BSG102",
			Section:  sectionHereDocuments,
			Severity: lint.SeverityWarning,
			Doc:      "Quote the delimiter of a here document that expands nothing, instead of escaping each `$`",
			Check:    checkEscapedHeredoc,
		},
	)
}

// echoOption matches the options `echo` takes, alone or bundled.
var echoOption = regexp.MustCompile(`^-[neE]+$`)

// checkEchoData reports an `echo` that reads its first argument as an option,
// and one whose first argument starts with an expansion: a value of `-n` then
// prints nothing, and `-e` turns the backslashes of the data into escapes.
func checkEchoData(f *File, r *Reporter) {
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if name != "echo" || len(call.Args) < 2 {
			return
		}
		first := call.Args[1]
		if lit := wordLiteral(first); echoOption.MatchString(lit) {
			r.At(first.Pos(), "`echo %s` depends on the shell and its options, and expands the backslashes of the data; use `printf`, with the data in `%%s`", lit)
			return
		}
		if startsWithExpansion(first) {
			r.At(first.Pos(), "`echo` reads a value that starts with `-n` or `-e` as an option, and prints nothing or rewrites it; use `printf '%%s\\n'`")
		}
	})
}

// numericSpecials are the special parameters that always hold a number, so
// they cannot start with a dash.
var numericSpecials = map[string]bool{"#": true, "?": true, "$": true, "!": true}

// startsWithExpansion reports whether the first thing a word produces comes
// from a parameter expansion or a command substitution that may start with a
// dash.
func startsWithExpansion(word *syntax.Word) bool {
	if len(word.Parts) == 0 {
		return false
	}
	part := word.Parts[0]
	if dq, ok := part.(*syntax.DblQuoted); ok {
		if len(dq.Parts) == 0 {
			return false
		}
		part = dq.Parts[0]
	}
	switch p := part.(type) {
	case *syntax.ParamExp:
		return p.Param == nil || p.Length || !numericSpecials[p.Param.Value]
	case *syntax.CmdSubst:
		return true
	}
	return false
}

func checkDashHeredoc(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		if redir, ok := node.(*syntax.Redirect); ok && redir.Op == syntax.DashHdoc {
			r.At(redir.OpPos, "`<<-` strips leading tabs only, which the guide does not indent with, and breaks when an editor turns them into spaces; write the body and the delimiter at the start of the line")
		}
		return true
	})
}

// checkEscapedHeredoc reports a here document whose delimiter is unquoted,
// that expands nothing, and that escapes `$` or a backtick by hand: one escape
// forgotten later runs as code. A quoted delimiter keeps the body literal.
func checkEscapedHeredoc(f *File, r *Reporter) {
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		redir, ok := node.(*syntax.Redirect)
		if !ok || (redir.Op != syntax.Hdoc && redir.Op != syntax.DashHdoc) || redir.Hdoc == nil || quotedDelimiter(redir.Word) {
			return true
		}
		escaped := false
		for _, part := range redir.Hdoc.Parts {
			lit, ok := part.(*syntax.Lit)
			if !ok {
				return true
			}
			if strings.Contains(lit.Value, `\$`) || strings.Contains(lit.Value, "\\`") {
				escaped = true
			}
		}
		if escaped {
			r.At(redir.OpPos, "this here document expands nothing but escapes `$` by hand; quote the delimiter, `<< 'EOF'`, and write the text as it is")
		}
		return true
	})
}

// quotedDelimiter reports whether a here-document delimiter carries quotes or
// a backslash, which makes the body literal.
func quotedDelimiter(word *syntax.Word) bool {
	if word == nil {
		return false
	}
	for _, part := range word.Parts {
		switch p := part.(type) {
		case *syntax.SglQuoted, *syntax.DblQuoted:
			return true
		case *syntax.Lit:
			if strings.Contains(p.Value, `\`) {
				return true
			}
		}
	}
	return false
}
