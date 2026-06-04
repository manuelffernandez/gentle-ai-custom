package overlay

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractGentleAIVersion(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    string
		wantErr bool
	}{
		{name: "plain output", output: "gentle-ai 1.34.0", want: "1.34.0"},
		{name: "prefixed output", output: "gentle-ai version v1.34.0", want: "1.34.0"},
		{name: "prerelease output", output: "gentle-ai 1.34.0-rc.1", want: "1.34.0-rc.1"},
		{name: "build metadata output", output: "gentle-ai 1.34.0+build.7", want: "1.34.0+build.7"},
		{name: "missing version", output: "gentle-ai build unknown", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := extractGentleAIVersion(tc.output)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("extractGentleAIVersion(%q) = %q, want error", tc.output, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("extractGentleAIVersion(%q) error = %v", tc.output, err)
			}
			if got != tc.want {
				t.Fatalf("extractGentleAIVersion(%q) = %q, want %q", tc.output, got, tc.want)
			}
		})
	}
}

func TestNormalizeVersionLabel(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "plain", raw: "  1.34.0  ", want: "1.34.0"},
		{name: "prefixed", raw: "v1.34.0", want: "1.34.0"},
		{name: "prerelease preserved", raw: "v1.34.0-rc.1", want: "1.34.0-rc.1"},
		{name: "build metadata preserved", raw: "V1.34.0+build.7", want: "1.34.0+build.7"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeVersionLabel(tc.raw)
			if got != tc.want {
				t.Fatalf("normalizeVersionLabel(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestCompareVersionStrings(t *testing.T) {
	tests := []struct {
		name      string
		installed string
		audited   string
		want      int
		wantErr   bool
	}{
		{name: "match with v prefix", installed: "v1.34.0", audited: "1.34.0", want: 0},
		{name: "installed older", installed: "1.33.9", audited: "1.34.0", want: -1},
		{name: "installed newer", installed: "1.35.0", audited: "1.34.0", want: 1},
		{name: "missing patch treated as zero", installed: "1.34", audited: "1.34.0", want: 0},
		{name: "installed prerelease is uncertain", installed: "1.34.0-rc.1", audited: "1.34.0", wantErr: true},
		{name: "installed build metadata is uncertain", installed: "1.34.0+build.7", audited: "1.34.0", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := compareVersionStrings(tc.installed, tc.audited)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("compareVersionStrings(%q, %q) = %d, want error", tc.installed, tc.audited, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("compareVersionStrings(%q, %q) error = %v", tc.installed, tc.audited, err)
			}
			if got != tc.want {
				t.Fatalf("compareVersionStrings(%q, %q) = %d, want %d", tc.installed, tc.audited, got, tc.want)
			}
		})
	}
}

func TestRunVersionPreflightWarnsOnPrereleaseInstalledVersion(t *testing.T) {
	repoRoot := t.TempDir()
	statePath := filepath.Join(repoRoot, "overlay", "gentle-ai", "state")
	if err := os.MkdirAll(statePath, 0o755); err != nil {
		t.Fatalf("mkdir state path: %v", err)
	}
	if err := os.WriteFile(filepath.Join(statePath, "upstream-state.json"), []byte(`{"last_maintained_version":"v1.34.0"}`), 0o644); err != nil {
		t.Fatalf("write state file: %v", err)
	}
	installFakeGentleAI(t, "#!/bin/sh\necho 'gentle-ai 1.34.0-rc.1'\n")

	var stderr bytes.Buffer
	policy := Policy{}
	policy.Maintenance.StateFile = "overlay/gentle-ai/state/upstream-state.json"

	err := runVersionPreflight(repoRoot, policy, nil, &stderr)
	if err == nil {
		t.Fatal("runVersionPreflight() = nil, want warning error")
	}
	if !strings.Contains(stderr.String(), "could not be compared safely against the audited baseline") {
		t.Fatalf("stderr = %q, want warning about uncertain version comparison", stderr.String())
	}
	if !strings.Contains(stderr.String(), "Prerelease or build-metadata versions should be treated as uncertain") {
		t.Fatalf("stderr = %q, want prerelease explanation", stderr.String())
	}
	if !strings.Contains(err.Error(), "non-interactive mode") {
		t.Fatalf("error = %v, want non-interactive failure", err)
	}
}

func TestDetectGentleAIVersionRejectsFailedCommand(t *testing.T) {
	installFakeGentleAI(t, "#!/bin/sh\necho 'gentle-ai 1.34.0'\nexit 1\n")

	got, err := detectGentleAIVersion()
	if err == nil {
		t.Fatalf("detectGentleAIVersion() = %q, want error", got)
	}
}

func installFakeGentleAI(t *testing.T, script string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "gentle-ai")
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake gentle-ai: %v", err)
	}
	origPath := os.Getenv("PATH")
	if origPath == "" {
		t.Setenv("PATH", dir)
		return
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+origPath)
}
