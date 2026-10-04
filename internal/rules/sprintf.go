package rules

import "fmt"

// sprintf keeps the Reporter helpers on one formatting path. A message is
// always a format string, so a literal percent sign is written `%%` even when
// the message takes no argument, the way `go vet` checks it.
func sprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
