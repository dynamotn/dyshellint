package rules

import (
	"regexp"
	"strings"

	"gitlab.com/dynamo-tools/dyshellint/internal/lint"
)

func init() {
	register(Rule{
		Code:     "BSG149",
		Section:  sectionFileHeader,
		Severity: lint.SeverityWarning,
		Doc:      "Name the license of a distributed script with `# SPDX-License-Identifier: <id>`, not a license paragraph",
		Check:    checkSpdxHeader,
	})
}

// licenseProse matches a header line that states a license in words, in
// either spelling. A bare "licence" is left out: a script may be about
// licences without stating its own.
var licenseProse = regexp.MustCompile(`(?i)\blicen[cs]e\s*:|\blicen[cs]ed\s+under\b`)

// checkSpdxHeader reports a file header that states its license in prose
// with no SPDX line: the prose drifts from the LICENSE file, and no tool can
// read it.
func checkSpdxHeader(f *File, r *Reporter) {
	header := fileHeader(f)
	first := 1
	if f.Shebang != "" && strings.HasPrefix(strings.TrimSpace(f.Line(1)), "#!") {
		first = 2
	}
	at := 0
	for i, line := range header {
		if strings.Contains(line, "SPDX-License-Identifier:") {
			return
		}
		if at == 0 && licenseProse.MatchString(line) {
			at = first + i
		}
	}
	if at > 0 {
		r.AtLine(at, "the header states the license in prose, which no tool reads and which drifts from the LICENSE file; write `# SPDX-License-Identifier: <id>` instead")
	}
}
