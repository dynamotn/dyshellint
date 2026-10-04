package rules

import (
	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG059",
		Section:  "Calling Commands > Deprecated Commands",
		Severity: lint.SeverityWarning,
		Doc:      "Use the maintained replacement of a deprecated command",
		Check:    checkDeprecatedCommands,
	})
}

// replacements maps a deprecated command to what replaces it.
var replacements = map[string]string{
	"apt-key":  "a keyring under `/etc/apt/keyrings` named by `signed-by=` in the source",
	"egrep":    "`grep -E`",
	"fgrep":    "`grep -F`",
	"which":    "`command -v`",
	"ifconfig": "`ip addr`",
	"tempfile": "`mktemp`",
	"netstat":  "`ss`",
}

func checkDeprecatedCommands(f *File, r *Reporter) {
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if replacement, ok := replacements[name]; ok {
			r.At(call.Pos(), "`%s` is deprecated and missing from current systems; use %s", name, replacement)
		}
	})
}
