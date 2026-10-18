package rules_test

import "testing"

// TestStrictModeSpellings expects BSG031 to take errexit, nounset and
// pipefail in any order and any spelling `set` accepts, and to report a file
// that leaves one of them out.
func TestStrictModeSpellings(t *testing.T) {
	for _, tc := range []struct {
		set    string
		report bool
	}{
		{"set -euo pipefail", false},
		{"set -Eeuo pipefail", false},
		{"set -ueo pipefail", false},
		{"set -eu -o pipefail", false},
		{"set -o errexit -o nounset -o pipefail", false},
		{"set -e\nset -u\nset -o pipefail", false},
		{"set -eu", true},
		{"set -eo pipefail", true},
		{"set -o pipefail", true},
		{"# set -euo pipefail", true},
	} {
		src := "#!/usr/bin/env bash\n" + tc.set + "\nprintf 'ok\\n'\n"
		got := linkAndRun(t, "BSG031", source{"strict.sh", src})
		if reported := len(got) > 0; reported != tc.report {
			t.Errorf("%q: reported %v, want %v", tc.set, reported, tc.report)
		}
	}
}
