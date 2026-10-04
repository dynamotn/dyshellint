package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionPortability = "Features and Bugs > Portability"

func init() {
	register(Rule{
		Code:     "BSG083",
		Section:  sectionPortability,
		Severity: lint.SeverityWarning,
		Doc:      "Do not rely on a GNU-only option without probing for it",
		Check:    checkGnuOnlyFlags,
	})
}

// gnuOption is an option only GNU coreutils, findutils or grep accept. A
// short option is matched inside a cluster such as `-oP`; a long one, and a
// single-dash word such as find's `-printf`, only as itself or with `=value`.
type gnuOption struct {
	short byte
	long  string
}

var gnuOnly = map[string][]gnuOption{
	"date":     {{short: 'd'}, {long: "--date"}},
	"sed":      {{long: "-i"}, {long: "--in-place"}},
	"readlink": {{short: 'f'}, {short: 'e'}, {short: 'm'}, {long: "--canonicalize"}, {long: "--canonicalize-existing"}, {long: "--canonicalize-missing"}},
	"stat":     {{short: 'c'}, {long: "--format"}, {long: "--printf"}},
	"find":     {{long: "-printf"}, {long: "-fprintf"}},
	"grep":     {{short: 'P'}, {long: "--perl-regexp"}},
	"xargs":    {{short: 'r'}, {long: "--no-run-if-empty"}},
	"mktemp":   {{long: "--suffix"}, {long: "--tmpdir"}},
	"sort":     {{short: 'V'}, {long: "--version-sort"}},
	"base64":   {{short: 'w'}, {long: "--wrap"}},
	"cp":       {{long: "--reflink"}},
}

// flavourProbe matches what a function does when it tells the platforms apart
// before choosing an option.
var flavourProbe = regexp.MustCompile(`--version|(?i:gnu|bsd|busybox|darwin|macos|flavou?r|uname|is_linux|is_macos|goos)`)

// linuxOnly matches a comment declaring that a file runs on Linux alone.
var linuxOnly = regexp.MustCompile(`(?i)linux[- ]only`)

func checkGnuOnlyFlags(f *File, r *Reporter) {
	if linuxOnly.Match(f.Src) {
		return
	}
	eachFunc(f, func(decl *syntax.FuncDecl) {
		if flavourProbe.MatchString(f.Text(decl.Pos(), decl.End())) {
			return
		}
		allCalls(decl.Body, func(call *syntax.CallExpr, name string) {
			options, ok := gnuOnly[name]
			if !ok {
				return
			}
			for _, arg := range call.Args[1:] {
				lit := wordLiteral(arg)
				if lit == "--" {
					return
				}
				if opt, ok := matchGnuOption(lit, options); ok {
					r.At(arg.Pos(), "`%s %s` exists only in the GNU tools, so the call fails on macOS, the BSDs and BusyBox; use a portable form, or probe for the flavour first",
						name, opt)
					return
				}
			}
		})
	})
}

func matchGnuOption(lit string, options []gnuOption) (string, bool) {
	if !strings.HasPrefix(lit, "-") {
		return "", false
	}
	for _, opt := range options {
		switch {
		case opt.long != "":
			if lit == opt.long || strings.HasPrefix(lit, opt.long+"=") {
				return opt.long, true
			}
		case !strings.HasPrefix(lit, "--") && strings.IndexByte(lit[1:], opt.short) >= 0:
			return "-" + string(opt.short), true
		}
	}
	return "", false
}
