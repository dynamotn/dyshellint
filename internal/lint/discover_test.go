package lint

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

func TestDiscoverLeavesOutWhatGitIgnores(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init", "--quiet").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	for name, content := range map[string]string{
		".gitignore":            "dist/\ncoverage/\n",
		"src/lib.sh":            "#!/usr/bin/env bash\n",
		"dist/bundle.sh":        "#!/usr/bin/env bash\n",
		"coverage/kcov-hook.sh": "eval $x\n",
		"untracked.bash":        "#!/usr/bin/env bash\n",
		"test/lib.bats":         "@test \"x\" {\n  true\n}\n",
		"src/notes.txt":         "not a script\n",
	} {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	files, err := Discover([]string{dir})
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	want := []string{filepath.Join(dir, "src", "lib.sh"), filepath.Join(dir, "test", "lib.bats"), filepath.Join(dir, "untracked.bash")}
	slices.Sort(files)
	if !slices.Equal(files, want) {
		t.Errorf("got %v, want %v", files, want)
	}

	// A file named on the command line is checked even when git ignores it.
	bundle := filepath.Join(dir, "dist", "bundle.sh")
	files, err = Discover([]string{bundle})
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if !slices.Equal(files, []string{bundle}) {
		t.Errorf("got %v, want %v", files, []string{bundle})
	}
}

func TestDiscoverWalksASymlinkedDirectory(t *testing.T) {
	real := t.TempDir()
	if err := os.WriteFile(filepath.Join(real, "lib.sh"), []byte("#!/usr/bin/env bash\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlink: %v", err)
	}
	// The files keep the name the caller gave, not the target of the link.
	files, err := Discover([]string{link})
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	want := []string{filepath.Join(link, "lib.sh")}
	if !slices.Equal(files, want) {
		t.Errorf("got %v, want %v", files, want)
	}
}

func TestDiscoverSkipsWhatIsNotARegularFile(t *testing.T) {
	if _, err := exec.LookPath("mkfifo"); err != nil {
		t.Skip("mkfifo is not installed")
	}
	dir := t.TempDir()
	// Opening a pipe nobody writes to blocks for good, which hung the walk on
	// a FIFO without an extension, and reading a device never ends.
	if out, err := exec.Command("mkfifo", filepath.Join(dir, "pipe"), filepath.Join(dir, "pipe.sh")).CombinedOutput(); err != nil {
		t.Skipf("mkfifo: %v: %s", err, out)
	}
	if err := os.Symlink(os.DevNull, filepath.Join(dir, "null.sh")); err != nil {
		t.Skipf("symlink: %v", err)
	}
	script := filepath.Join(dir, "run.sh")
	if err := os.WriteFile(script, []byte("#!/usr/bin/env bash\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	files, err := Discover([]string{dir})
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if !slices.Equal(files, []string{script}) {
		t.Errorf("got %v, want %v", files, []string{script})
	}
}
