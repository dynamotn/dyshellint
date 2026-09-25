package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStdinSourceKeepsTheNameAndTheExtension(t *testing.T) {
	source, err := StdinSource(strings.NewReader("#!/usr/bin/env bash\n"), "scripts/demo.sh")
	if err != nil {
		t.Fatalf("StdinSource: %v", err)
	}
	defer source.Close()

	if source.Name != "scripts/demo.sh" {
		t.Errorf("name = %q, want scripts/demo.sh", source.Name)
	}
	if filepath.Base(source.Disk) != "demo.sh" {
		t.Errorf("the copy lost the base name: %q", source.Disk)
	}
	if source.ModeKnown {
		t.Error("content read from standard input has no file mode")
	}
	content, err := os.ReadFile(source.Disk)
	if err != nil {
		t.Fatalf("read the copy: %v", err)
	}
	if string(content) != "#!/usr/bin/env bash\n" {
		t.Errorf("the copy holds %q", content)
	}

	disk := source.Disk
	source.Close()
	if _, err := os.Stat(disk); !os.IsNotExist(err) {
		t.Errorf("the copy outlived Close: %v", err)
	}
}

func TestRenameMovesFindingsBackToTheReportedName(t *testing.T) {
	source := Source{Name: "scripts/demo.sh", Disk: "/tmp/dyshellint123/demo.sh"}
	findings := source.Rename([]Finding{
		{File: "/tmp/dyshellint123/demo.sh", Rule: "SC2086"},
		{File: "other.sh", Rule: "SC2086"},
	})
	if findings[0].File != "scripts/demo.sh" {
		t.Errorf("got %q, want scripts/demo.sh", findings[0].File)
	}
	if findings[1].File != "other.sh" {
		t.Errorf("an unrelated file was rewritten to %q", findings[1].File)
	}
}

func TestFindRCFileWalksUp(t *testing.T) {
	root := t.TempDir()
	rc := filepath.Join(root, ".shellcheckrc")
	if err := os.WriteFile(rc, []byte("shell=bash\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	nested := filepath.Join(root, "scripts", "lib")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if got := FindRCFile(nested); got != rc {
		t.Errorf("got %q, want %q", got, rc)
	}
	if got := FindRCFile(t.TempDir()); got != "" {
		t.Errorf("a tree without a configuration returned %q", got)
	}
}
