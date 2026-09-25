package lint

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Source is one unit to check. It separates the name a finding is reported
// under from the file the external tools are pointed at, which are the same
// thing for a file on disk and different for content read from standard input.
type Source struct {
	// Name is what a finding says, and what the path-based rules look at.
	Name string
	// Disk is the file ShellCheck and shfmt are run against.
	Disk string
	// Content is the text to parse.
	Content []byte
	// Executable is the executable bit of the file.
	Executable bool
	// ModeKnown is false when the content has no file mode of its own.
	ModeKnown bool
	// cleanup removes the temporary copy, when there is one.
	cleanup func()
}

// Close releases the temporary copy of a source read from standard input.
func (s Source) Close() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

// FileSource reads one file from disk.
func FileSource(path string) (Source, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Source{}, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return Source{}, err
	}
	return Source{
		Name:       path,
		Disk:       path,
		Content:    content,
		Executable: info.Mode().Perm()&0o111 != 0,
		ModeKnown:  true,
	}, nil
}

// StdinSource reads the whole of r and stores it in a temporary file, so that
// ShellCheck and shfmt have something to open. The findings are still reported
// under name, which is the buffer an editor is linting.
func StdinSource(r io.Reader, name string) (Source, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return Source{}, fmt.Errorf("read standard input: %w", err)
	}
	if name == "" {
		name = "stdin.sh"
	}
	dir, err := os.MkdirTemp("", "dyshellint")
	if err != nil {
		return Source{}, fmt.Errorf("create a temporary directory: %w", err)
	}
	// The copy keeps the base name, because ShellCheck resolves a `source=`
	// directive relative to the file that carries it, and shfmt picks its
	// dialect from the extension.
	disk := filepath.Join(dir, filepath.Base(name))
	if err := os.WriteFile(disk, content, 0o600); err != nil {
		os.RemoveAll(dir)
		return Source{}, fmt.Errorf("write %s: %w", disk, err)
	}
	return Source{
		Name:    name,
		Disk:    disk,
		Content: content,
		cleanup: func() { os.RemoveAll(dir) },
	}, nil
}

// ConfigDir returns the directory a configuration file should be looked up
// from: the one holding the file the caller named, not the temporary copy.
func (s Source) ConfigDir() string {
	dir := filepath.Dir(s.Name)
	absolute, err := filepath.Abs(dir)
	if err != nil {
		return dir
	}
	return absolute
}

// FindRCFile walks up from dir looking for a `.shellcheckrc`, the way
// ShellCheck itself does when the file is inside the repository.
func FindRCFile(dir string) string {
	for {
		candidate := filepath.Join(dir, ".shellcheckrc")
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// Rename rewrites the file of every finding that points at the temporary copy,
// so that a report never mentions a path the caller has never heard of.
func (s Source) Rename(findings []Finding) []Finding {
	if s.Disk == s.Name {
		return findings
	}
	for i := range findings {
		if findings[i].File == s.Disk {
			findings[i].File = s.Name
		}
	}
	return findings
}
