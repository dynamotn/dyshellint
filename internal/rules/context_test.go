package rules_test

import (
	"testing"

	"gitlab.com/dynamo-tools/dyshellint/internal/rules"
)

// TestBaseNameDropsTheExtension expects a library to be known by its name
// without `.sh` or `.bash`, which is the namespace its functions take.
func TestBaseNameDropsTheExtension(t *testing.T) {
	for path, want := range map[string]string{
		"scripts/lib/package_manager.sh": "package_manager",
		"lib/util.bash":                  "util",
		"lib/util":                       "util",
	} {
		if got := rules.BaseName(path); got != want {
			t.Errorf("BaseName(%q) = %q, want %q", path, got, want)
		}
		file, err := rules.NewFile(path, []byte("function hello {\n  :\n}\n"), false)
		if err != nil {
			t.Fatal(err)
		}
		if file.Namespace != want {
			t.Errorf("NewFile(%q).Namespace = %q, want %q", path, file.Namespace, want)
		}
	}
}
