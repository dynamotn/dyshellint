package lint

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Discover expands the given paths into the list of shell files to check. A
// file named on the command line is taken as given; a directory is walked,
// leaving out what git ignores there, such as a generated bundle.
func Discover(paths []string) ([]string, error) {
	var files []string
	seen := map[string]bool{}
	add := func(path string) {
		clean := filepath.Clean(path)
		if seen[clean] {
			return
		}
		seen[clean] = true
		files = append(files, clean)
	}
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			add(path)
			continue
		}
		kept := notIgnored(path)
		// WalkDir does not follow a root that is a symlink, so the walk goes
		// over the target and puts every name back under the one given.
		root, err := filepath.EvalSymlinks(path)
		if err != nil {
			return nil, err
		}
		err = filepath.WalkDir(root, func(name string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if rel, err := filepath.Rel(root, name); err == nil {
				name = filepath.Join(path, rel)
			}
			if entry.IsDir() {
				if name != filepath.Clean(path) && skipDir(name) {
					return filepath.SkipDir
				}
				return nil
			}
			if kept != nil && !kept[filepath.Clean(name)] {
				return nil
			}
			if isRegular(name, entry) && IsShellFile(name) {
				add(name)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

// notIgnored returns the files under dir that git does not ignore: tracked
// ones, and untracked ones no ignore rule matches. It returns nil outside a
// repository, or without git, and then nothing is left out.
func notIgnored(dir string) map[string]bool {
	// The repository being linted is not trusted: core.fsmonitor in its
	// configuration names a program git would run to list the files.
	out, err := exec.Command("git", "-c", "core.fsmonitor=false", "-C", dir, "ls-files", "-z", "--cached", "--others", "--exclude-standard").Output()
	if err != nil {
		return nil
	}
	kept := map[string]bool{}
	for _, name := range bytes.Split(out, []byte{0}) {
		if len(name) > 0 {
			kept[filepath.Join(dir, string(name))] = true
		}
	}
	return kept
}

// skipDir keeps the walk inside this repository: `.git` itself, and any nested
// checkout or submodule, which carries its own conventions and its own linter.
func skipDir(name string) bool {
	base := filepath.Base(name)
	if base == ".git" || base == ".worktrees" || base == "node_modules" {
		return true
	}
	_, err := os.Stat(filepath.Join(name, ".git"))
	return err == nil
}

// isRegular reports whether a file found by the walk can be read to its end:
// a FIFO blocks the open, and a device such as /dev/zero never ends. A symlink
// counts when its target is a regular file.
func isRegular(name string, entry os.DirEntry) bool {
	if entry.Type().IsRegular() {
		return true
	}
	if entry.Type()&os.ModeSymlink == 0 {
		return false
	}
	info, err := os.Stat(name)
	return err == nil && info.Mode().IsRegular()
}

// IsShellFile reports whether a file is worth parsing: the extension says so,
// `.bats` included, or the first line is a bash shebang.
func IsShellFile(path string) bool {
	switch filepath.Ext(path) {
	case ".sh", ".bash", ".bats":
		return true
	case "":
	default:
		return false
	}
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()
	// A shebang is short, and a binary file has no line to speak of: reading
	// a bounded prefix is enough, and never pulls a large file into memory.
	head := make([]byte, 256)
	n, _ := io.ReadFull(file, head)
	line, _, _ := bytes.Cut(head[:n], []byte{'\n'})
	return bytes.HasPrefix(line, []byte("#!")) && bytes.Contains(line, []byte("sh"))
}
