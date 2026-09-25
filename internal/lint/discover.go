package lint

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Discover expands the given paths into the list of shell files to check. A
// file named on the command line is taken as given; a directory is walked.
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
		err = filepath.WalkDir(path, func(name string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if name != path && skipDir(name) {
					return filepath.SkipDir
				}
				return nil
			}
			if IsShellFile(name) {
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

// IsShellFile reports whether a file is worth parsing: the extension says so,
// or the first line is a bash shebang.
func IsShellFile(path string) bool {
	switch filepath.Ext(path) {
	case ".sh", ".bash":
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
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return false
	}
	line := scanner.Text()
	return strings.HasPrefix(line, "#!") && strings.Contains(line, "sh")
}
