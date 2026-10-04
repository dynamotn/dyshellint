package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG095",
		Section:  sectionDebugMode,
		Severity: lint.SeverityWarning,
		Doc:      "Route a side effect through `dybatpho::dry_run`, or check `DRY_RUN`, in a script that offers a dry run",
		Check:    checkDryRunBypass,
	})
}

// packageManagers change what is installed on the machine.
var packageManagers = map[string]bool{
	"apt": true, "apt-get": true, "dnf": true, "yum": true, "pacman": true, "paru": true,
	"yay": true, "brew": true, "emerge": true, "zypper": true, "apk": true, "port": true,
}

// mutatingUnitVerbs are the systemctl verbs that change the system.
var mutatingUnitVerbs = map[string]bool{
	"enable": true, "disable": true, "start": true, "stop": true, "restart": true,
	"reload": true, "mask": true, "unmask": true, "daemon-reload": true,
}

func checkDryRunBypass(f *File, r *Reporter) {
	src := string(f.Src)
	if !strings.Contains(src, "DRY_RUN") && !strings.Contains(src, "dybatpho::dry_run") {
		return
	}
	// The file that implements the dry run mentions it without offering one.
	if strings.Contains(src, "function dybatpho::dry_run") {
		return
	}
	eachFunc(f, func(decl *syntax.FuncDecl) {
		// A private helper of a library runs for a public function, which is
		// where the dry run is decided.
		if f.Role == RoleLibrary && !isPublicFunc(decl) {
			return
		}
		text := f.Text(decl.Pos(), decl.End())
		if strings.Contains(text, "DRY_RUN") || strings.Contains(text, "dry_run") {
			return
		}
		allCalls(decl.Body, func(call *syntax.CallExpr, name string) {
			if touchesScratch(call) {
				return
			}
			if what := sideEffect(call, name); what != "" {
				r.At(call.Pos(), "%s runs even under `DRY_RUN`, which this script offers; wrap it in `dybatpho::dry_run`, or check `DRY_RUN` first",
					what)
			}
		})
	})
}

// sideEffect describes the change a command makes, or returns the empty
// string for one that only reads.
func sideEffect(call *syntax.CallExpr, name string) string {
	switch {
	case name == "rm" || name == "mv" || name == "install" || name == "wget":
		return "`" + name + "`"
	case name == "cp" && len(call.Args) > 2:
		return "`cp`"
	case name == "ln" && hasFlag(call, 's'):
		return "`ln -s`"
	case name == "curl" && (hasFlag(call, 'o') || hasFlag(call, 'O') || hasLong(call, "--output") || hasLong(call, "--remote-name")):
		return "`curl -o`"
	case packageManagers[name]:
		return "`" + name + "`"
	case name == "systemctl":
		for _, arg := range call.Args[1:] {
			if mutatingUnitVerbs[wordLiteral(arg)] {
				return "`systemctl " + wordLiteral(arg) + "`"
			}
		}
	case name == "chezmoi" && len(call.Args) > 1 && wordLiteral(call.Args[1]) == "apply":
		return "`chezmoi apply`"
	}
	return ""
}

// scratchName matches the variables a script keeps its own working files in.
var scratchName = regexp.MustCompile(`(?i)(tmp|temp|staged|staging|partial|scratch|aside|side)`)

// touchesScratch reports whether every path a call names is one of the
// script's own working files, whose cleanup is no change a dry run must skip.
func touchesScratch(call *syntax.CallExpr) bool {
	seen := false
	for _, arg := range call.Args[1:] {
		if strings.HasPrefix(wordLiteral(arg), "-") {
			continue
		}
		pe := wordParam(arg)
		if pe == nil || pe.Param == nil || !scratchName.MatchString(pe.Param.Value) {
			return false
		}
		seen = true
	}
	return seen
}
