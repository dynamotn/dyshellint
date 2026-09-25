package version

import (
	"runtime"
	"strings"
	"testing"
)

func TestGetVersionStr(t *testing.T) {
	got := GetVersionStr()
	if got == "" {
		t.Error("GetVersionStr() returned empty string")
	}
	if got != version {
		t.Errorf("GetVersionStr() = %v, want %v", got, version)
	}
}

func TestGetBuildInfo(t *testing.T) {
	info := GetBuildInfo()
	if info.Version == "" {
		t.Error("BuildInfo.Version is empty")
	}
	if info.GoVersion != runtime.Version() {
		t.Errorf("BuildInfo.GoVersion = %v, want %v", info.GoVersion, runtime.Version())
	}
}

// TestStringLeavesOutEmptyFields keeps `--version` readable on a build that
// was not stamped, which is every build that is not a release.
func TestStringLeavesOutEmptyFields(t *testing.T) {
	bare := BuildInfo{Version: "dev", GoVersion: runtime.Version()}
	got := bare.String()
	if strings.Contains(got, "commit") || strings.Contains(got, "built") {
		t.Errorf("an unstamped build printed empty fields:\n%s", got)
	}
	if !strings.HasPrefix(got, "dyshellint dev") {
		t.Errorf("got %q, want it to start with the name and the version", got)
	}

	stamped := BuildInfo{
		Version:      "1.0.0",
		GitCommit:    "abc1234",
		GitTreeState: "clean",
		BuildTime:    "2026-01-02T03:04:05Z",
		GoVersion:    runtime.Version(),
	}
	for _, want := range []string{"abc1234", "clean", "2026-01-02T03:04:05Z"} {
		if !strings.Contains(stamped.String(), want) {
			t.Errorf("%q missing from:\n%s", want, stamped.String())
		}
	}
}
