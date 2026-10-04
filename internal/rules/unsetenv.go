package rules

import (
	"regexp"

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
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		pe, ok := node.(*syntax.ParamExp)
		if !ok || pe.Param == nil || pe.Exp != nil || pe.Excl {
			return true
		}
		name := pe.Param.Value
		if seen[name] || !upperName.MatchString(name) || shellVariables[name] || p.Assigned(name) {
			return true
		}
		seen[name] = true
		r.At(pe.Pos(), "nothing in these files sets %q, so it comes from the environment, and `set -u` stops the script when it is not there; write `${%s:-}` or `${%s-default}`",
			name, name, name)
		return true
	})
}
