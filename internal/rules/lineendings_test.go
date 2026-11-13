package rules_test

import (
	"slices"
	"testing"
)

// TestLineEndings expects BSG140 to report a byte order mark on line 1 and
// the first CRLF line once, and to leave a file of LF lines alone. The CRLF
// cases live here rather than in a fixture, which an editor or Git could
// silently convert to LF.
func TestLineEndings(t *testing.T) {
	const body = "set -euo pipefail\nprintf 'ok\\n'\n"
	for _, tc := range []struct {
		name, src string
		want      []string
	}{
		{"lf", "#!/usr/bin/env bash\n" + body, nil},
		{"crlf", "#!/usr/bin/env bash\r\nset -euo pipefail\r\nprintf 'ok\\n'\r\n", []string{"crlf.sh:1:1"}},
		{"late crlf", "#!/usr/bin/env bash\n" + body + "if true; then\r\n  :\r\nfi\r\n", []string{"late crlf.sh:4:1"}},
		{"bom", "\xef\xbb\xbf#!/usr/bin/env bash\n" + body, []string{"bom.sh:1:1"}},
		{"bom and crlf", "\xef\xbb\xbf#!/usr/bin/env bash\r\n" + body, []string{"bom and crlf.sh:1:1", "bom and crlf.sh:1:1"}},
	} {
		got := linkAndRun(t, "BSG140", source{tc.name + ".sh", tc.src})
		if !slices.Equal(got, tc.want) {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}
