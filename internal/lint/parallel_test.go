package lint

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestEachFileKeepsEveryFinding(t *testing.T) {
	files := []string{"a.sh", "b.sh", "c.sh", "d.sh"}
	findings, err := eachFile(files, 3, func(file string) ([]Finding, error) {
		return []Finding{{File: file}}, nil
	})
	if err != nil {
		t.Fatalf("eachFile: %v", err)
	}
	var got []string
	for _, finding := range findings {
		got = append(got, finding.File)
	}
	if strings.Join(got, ",") != strings.Join(files, ",") {
		t.Errorf("got %v, want %v in the order of files", got, files)
	}
}

func TestEachFileReturnsTheFirstError(t *testing.T) {
	first := errors.New("first")
	_, err := eachFile([]string{"a.sh", "b.sh", "c.sh"}, 2, func(file string) ([]Finding, error) {
		switch file {
		case "b.sh":
			return nil, first
		case "c.sh":
			return nil, errors.New("second")
		}
		return nil, nil
	})
	if !errors.Is(err, first) {
		t.Errorf("got %v, want the error of the earliest file", err)
	}
}

func TestEachFileStartsWithTheLargest(t *testing.T) {
	dir := t.TempDir()
	var files []string
	for name, size := range map[string]int{"small.sh": 1, "large.sh": 100, "medium.sh": 10} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(strings.Repeat("#", size)), 0o600); err != nil {
			t.Fatal(err)
		}
		files = append(files, path)
	}
	var mu sync.Mutex
	var started []string
	_, err := eachFile(files, 1, func(file string) ([]Finding, error) {
		mu.Lock()
		defer mu.Unlock()
		started = append(started, filepath.Base(file))
		return nil, nil
	})
	if err != nil {
		t.Fatalf("eachFile: %v", err)
	}
	if want := "large.sh,medium.sh,small.sh"; strings.Join(started, ",") != want {
		t.Errorf("started %v, want %s", started, want)
	}
}
