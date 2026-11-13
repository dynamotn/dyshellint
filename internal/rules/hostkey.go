package rules

import (
	"regexp"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG158",
		Section:  "Calling Commands > Host Key Verification",
		Severity: lint.SeverityError,
		Doc:      "Do not turn off host key checking with `StrictHostKeyChecking=no` or `UserKnownHostsFile=/dev/null`",
		Check:    checkHostKeyBypass,
	})
}

// hostKeyBypass matches the ssh options that accept any host key, in either
// spelling ssh reads: `Key=value` or `Key value`, any case, joined to `-o` or
// not.
var hostKeyBypass = regexp.MustCompile(`(?i)StrictHostKeyChecking(\s*=\s*|\s+)["']?(no|off)\b|(User|Global)KnownHostsFile(\s*=\s*|\s+)["']?/dev/null\b`)

// checkHostKeyBypass reports the options wherever they appear: an argument of
// ssh, scp, sftp or git, `rsync -e`, a value of `GIT_SSH_COMMAND` or
// `RSYNC_RSH`, an options array or an ssh config written from a here-document.
// Only the innermost word holding the text is reported, so `$(ssh ...)` inside
// an assignment is reported once.
func checkHostKeyBypass(f *File, r *Reporter) {
	var hits []*syntax.Word
	syntax.Walk(f.Syntax, func(node syntax.Node) bool {
		if word, ok := node.(*syntax.Word); ok && hostKeyBypass.MatchString(f.Text(word.Pos(), word.End())) {
			hits = append(hits, word)
		}
		return true
	})
	for _, word := range hits {
		if holdsAnother(word, hits) {
			continue
		}
		text := f.Text(word.Pos(), word.End())
		at := hostKeyBypass.FindStringIndex(text)
		const msg = "turning off host key checking lets anyone on the path pose as the host and read every password, key and command sent to it; pin its key in a known_hosts file shipped with the script and pass `-o UserKnownHostsFile=<file> -o StrictHostKeyChecking=yes`"
		if lines := strings.Count(text[:at[0]], "\n"); lines > 0 {
			// A here-document: point at the line inside it.
			r.AtLine(int(word.Pos().Line())+lines, msg)
		} else {
			r.At(word.Pos(), msg)
		}
	}
}

// holdsAnother reports whether a word contains one of the other words.
func holdsAnother(word *syntax.Word, words []*syntax.Word) bool {
	for _, other := range words {
		if other != word && word.Pos().Offset() <= other.Pos().Offset() && other.End().Offset() <= word.End().Offset() {
			return true
		}
	}
	return false
}
