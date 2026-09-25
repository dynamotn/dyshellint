// Package version provides build and version information.
//
// The values are empty in an ordinary `go build` and are filled in at release
// time by the ldflags of the Makefile and of .goreleaser.yaml, so a bug report
// names the exact code that ran.
//
// Example usage:
//
//	info := version.GetBuildInfo()
//	fmt.Printf("Version: %s\n", info.Version)
//	fmt.Printf("Git Commit: %s\n", info.GitCommit)
package version

import (
	"fmt"
	"runtime"
	"strings"
)

var (
	// version is the release this binary was built from. A build that is not a
	// release says so rather than claiming a version it does not have.
	version      = "dev"
	gitCommit    = ""
	gitTreeState = ""
	buildTime    = ""
)

// BuildInfo contains version and build information.
type BuildInfo struct {
	Version      string
	GitCommit    string
	GitTreeState string
	GoVersion    string
	BuildTime    string
}

// GetVersionStr returns the version string.
func GetVersionStr() string {
	return version
}

// GetBuildInfo returns the build information.
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:      GetVersionStr(),
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		GoVersion:    runtime.Version(),
		BuildTime:    buildTime,
	}
}

// String renders the build information as one line per field, leaving out the
// fields an unreleased build has nothing to say about.
func (b BuildInfo) String() string {
	lines := []string{fmt.Sprintf("dyshellint %s", b.Version)}
	for _, field := range []struct{ name, value string }{
		{"commit", b.GitCommit},
		{"tree", b.GitTreeState},
		{"built", b.BuildTime},
		{"go", b.GoVersion},
	} {
		if field.value != "" {
			lines = append(lines, fmt.Sprintf("  %-7s %s", field.name, field.value))
		}
	}
	return strings.Join(lines, "\n")
}
