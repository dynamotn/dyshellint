package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG073",
		Section:  sectionExpansion,
		Severity: lint.SeverityWarning,
		Doc:      "Give an environment variable nothing sets a default, as in `${NAME:-}`, under `set -u`",
		Check:    checkBareEnvironment,
	})
}

// upperName matches the UPPERCASE names the guide gives environment variables.
var upperName = regexp.MustCompile(`^[A-Z][A-Z0-9_]*$`)

// nounset matches a file that stops on an unset variable.
var nounset = regexp.MustCompile(`(?m)^\s*set\s+(-[a-zA-Z]*u|-o\s+nounset)`)

// shellVariables are the variables Bash itself, or every login, always sets.
var shellVariables = map[string]bool{
	"BASH": true, "BASHPID": true, "BASHOPTS": true, "BASH_ARGC": true, "BASH_ARGV": true,
	"BASH_COMMAND": true, "BASH_LINENO": true, "BASH_REMATCH": true, "BASH_SOURCE": true,
	"BASH_SUBSHELL": true, "BASH_VERSINFO": true, "BASH_VERSION": true, "BASH_ARGV0": true,
	"DIRSTACK": true, "EPOCHREALTIME": true, "EPOCHSECONDS": true, "EUID": true,
	"FUNCNAME": true, "GROUPS": true, "HOSTNAME": true, "HOSTTYPE": true, "IFS": true,
	"LINENO": true, "MACHTYPE": true, "OPTARG": true, "OPTIND": true, "OSTYPE": true,
	"PIPESTATUS": true, "PPID": true, "PS4": true, "PWD": true, "RANDOM": true, "REPLY": true,
	"SECONDS": true, "SHELLOPTS": true, "SHLVL": true, "SRANDOM": true, "UID": true,
	"HOME": true, "PATH": true,
}

func checkBareEnvironment(f *File, r *Reporter) {
	if !f.UsesDybatpho && !nounset.Match(f.Src) {
		return
	}
	p := f.project()
	seen := map[string]bool{}
	external := externalPrefixes(f, p)
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		pe, ok := node.(*syntax.ParamExp)
		if !ok || pe.Param == nil || pe.Exp != nil || pe.Excl {
			return true
		}
		name := pe.Param.Value
		if seen[name] || !upperName.MatchString(name) || shellVariables[name] || strings.HasPrefix(name, "BATS_") || p.Assigned(name) || hasAnyPrefix(name, external) {
			return true
		}
		seen[name] = true
		r.At(pe.Pos(), "nothing in these files sets %q, so it comes from the environment, and `set -u` stops the script when it is not there; write `${%s:-}` or `${%s-default}`",
			name, name, name)
		return true
	})
}

// externalPrefixes returns the variable prefixes of libraries the file calls
// but the run does not include, such as `DYBATPHO_` for `dybatpho::info` when
// dybatpho itself is not linted: those libraries set their own variables.
func externalPrefixes(f *File, p *Project) []string {
	seen := map[string]bool{}
	var out []string
	eachCall(f, func(_ *syntax.CallExpr, name string) {
		i := strings.Index(name, "::")
		if i <= 0 || p.funcs[name] != nil {
			return
		}
		prefix := strings.ToUpper(name[:i]) + "_"
		if !seen[prefix] {
			seen[prefix] = true
			out = append(out, prefix)
		}
	})
	return out
}

func hasAnyPrefix(name string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
