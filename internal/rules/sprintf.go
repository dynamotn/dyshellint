package rules

import "fmt"

// sprintf keeps the Reporter helpers on one formatting path.
func sprintf(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}
