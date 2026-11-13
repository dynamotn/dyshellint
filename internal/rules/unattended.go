package rules

import (
	"path/filepath"
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

const sectionUnattendedRuns = "Environment > Unattended Runs"

func init() {
	register(Rule{
		Code:     "BSG134",
		Section:  sectionUnattendedRuns,
		Severity: lint.SeverityError,
		Doc:      "Do not source `~/.bashrc`, `~/.profile` or another login profile from a script",
		Check:    checkSourcedProfile,
	})
}

// profilePath matches the start-up files of an interactive or login shell,
// once quotes are dropped and `$HOME` is written `~`.
var profilePath = regexp.MustCompile(`^(~/\.(bashrc|bash_profile|bash_login|profile)|/etc/(profile|bash\.bashrc|bashrc))$`)

// homePrefix matches the spellings of the home directory at the start of a
// path.
var homePrefix = regexp.MustCompile(`^(\$\{HOME\}|\$HOME)/`)

// checkSourcedProfile reports a script that sources a shell profile. Under
// cron, systemd or CI the file may be missing or belong to another user, it
// is written for a terminal (prompts, aliases, `stty`), and it makes the
// script depend on one person's setup. The start-up files themselves are left
// alone: `.bash_profile` sourcing `.bashrc` is how they are meant to work.
func checkSourcedProfile(f *File, r *Reporter) {
	base := strings.ToLower(filepath.Base(f.Path))
	if strings.Contains(base, "bashrc") || strings.Contains(base, "profile") || strings.Contains(base, "bash_login") {
		return
	}
	eachCall(f, func(call *syntax.CallExpr, name string) {
		if (name != "." && name != "source") || len(call.Args) < 2 {
			return
		}
		path := strings.ReplaceAll(wordSource(call.Args[1]), `"`, "")
		path = homePrefix.ReplaceAllString(path, "~/")
		if !profilePath.MatchString(path) {
			return
		}
		r.At(call.Pos(), "%s is written for an interactive login, and under cron, systemd or CI it may be missing, belong to another user or wait on a terminal; set `PATH` and the variables the script needs in the script itself",
			wordSource(call.Args[1]))
	})
}
