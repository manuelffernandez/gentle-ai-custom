package overlay

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyUpstreamReleaseParity(t *testing.T) {
	tests := []struct {
		name          string
		binaryVersion string
		tag           string
		ahead         bool
		wantVersion   string
		wantErr       string
	}{
		{name: "stable exact clean match", binaryVersion: "2.2.2", tag: "v2.2.2", wantVersion: "2.2.2"},
		{name: "HEAD ahead of release", binaryVersion: "2.2.2", tag: "v2.2.2", ahead: true, wantErr: "does not equal exact release"},
		{name: "exact tag missing", binaryVersion: "2.2.2", wantErr: "refs/tags/v2.2.2"},
		{name: "prerelease exact clean match", binaryVersion: "2.3.0-rc.1", tag: "v2.3.0-rc.1", wantVersion: "2.3.0-rc.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newReleaseGuardRepo(t, tt.tag, tt.ahead)
			installFakeGentleAI(t, "#!/bin/sh\necho 'gentle-ai "+tt.binaryVersion+"'\n")

			got, err := verifyUpstreamReleaseParity(repo, "")
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("verifyUpstreamReleaseParity() error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("verifyUpstreamReleaseParity() error = %v", err)
			}
			if got.Version != tt.wantVersion || got.Tag != "v"+tt.wantVersion || got.InstalledVersion != tt.wantVersion {
				t.Fatalf("release identity = %#v, want version %q and exact tag", got, tt.wantVersion)
			}
			head, gitErr := runGit(repo, false, "rev-parse", "HEAD")
			if gitErr != nil || got.Commit != head {
				t.Fatalf("release commit = %q, HEAD = %q, error = %v", got.Commit, head, gitErr)
			}
		})
	}
}

func TestRunSyncUpstreamAssetsRejectsCodexInvariantBeforeWrites(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	upstream := t.TempDir()
	writeV4StructuralFixtureWithModes(t, upstream, "assigned", "connected")
	// Preserve all OpenCode invariants while disconnecting only Codex's adapter.
	mustWriteFile(t, filepath.Join(upstream, "internal", "agents", "codex", "adapter.go"), []byte(`package codex
func (a *Adapter) Agent() model.AgentID { return model.AgentCodex }
func (a *Adapter) SystemPromptFile(homeDir string) string { return filepath.Join(homeDir, ".codex", "wrong.md") }
`))
	mustWriteFile(t, filepath.Join(upstream, "internal", "assets", "codex", "orchestrator.md"), []byte("codex upstream\n"))
	mustWriteFile(t, filepath.Join(upstream, "internal", "assets", "generic", "remote-authorization-contract.md"), []byte("authorization contract\n"))
	mustInitGitRepo(t, upstream)
	if _, err := runGit(upstream, false, "tag", "v2.2.2"); err != nil {
		t.Fatal(err)
	}
	installFakeGentleAI(t, "#!/bin/sh\necho 'gentle-ai 2.2.2'\n")
	t.Setenv(upstreamRepoEnvVar, upstream)
	repo := t.TempDir()
	policy := Policy{Version: 1}
	policy.Maintenance.StateFile = "overlay/gentle-ai/state/upstream-state.json"
	writeJSONFixture(t, filepath.Join(repo, "overlay", "gentle-ai", "policy", "gentle-ai-policy.json"), policy)
	writeJSONFixture(t, filepath.Join(repo, "overlay", "gentle-ai", "policy", "managed-assets.json"), ManagedAssetsManifest{Version: 1, Targets: map[string]ManagedAssetsTarget{
		"opencode": v4StructuralFixtureTarget(),
		"codex": {StructuralInvariantSources: []string{
			"internal/components/agentguidance/orchestrator.go", "internal/components/agentguidance/inject.go",
			"internal/agents/codex/adapter.go", "internal/cli/run.go", "internal/cli/sync.go",
		}, OwnedOverlayAssets: []OwnedOverlayAsset{
			{Key: "codex-orchestrator", Class: "upstream_source", Kind: "upstream_source_file", UpstreamPath: "internal/assets/codex/orchestrator.md", RepoUpstreamPath: "overlay/gentle-ai/assets/upstream/codex/orchestrator.md"},
			{Key: "codex-remote-authorization", Class: "upstream_source", Kind: "upstream_source_file", UpstreamPath: "internal/assets/generic/remote-authorization-contract.md", RepoUpstreamPath: "overlay/gentle-ai/assets/upstream/codex/remote-authorization-contract.md"},
		}},
	}})
	statePath := filepath.Join(repo, "overlay", "gentle-ai", "state", "upstream-state.json")
	head, err := runGit(upstream, false, "rev-parse", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	writeJSONFixture(t, statePath, UpstreamState{LastMaintainedCommit: head})
	before := mustReadFile(t, statePath)
	manifest, _, err := loadManagedAssetsManifest(repo)
	if err != nil || !codexSnapshotMappingsComplete(manifest.Targets["codex"]) {
		t.Fatalf("Codex refusal fixture lacks complete snapshot mappings: %v", err)
	}
	opencode, opencodeFailures := evaluateStructuralInvariants(upstream, manifest.Targets["opencode"], policy)
	if len(opencodeFailures) != 0 || !opencode.OrchestratorAssetOK || !opencode.RoutingInjectionOK ||
		!opencode.RemoteAuthorizationOK || !opencode.RoleTopologyOK || !opencode.ODDTopologyOK || !opencode.ODDRegistrationOK {
		t.Fatalf("OpenCode refusal fixture must pass all six structural gates: %#v, failures=%v", opencode, opencodeFailures)
	}
	if failures := evaluateCodexStructuralInvariants(upstream, manifest.Targets["codex"]); len(failures) == 0 || !strings.Contains(strings.Join(failures, "\n"), "adapter=false") {
		t.Fatalf("Codex refusal fixture must fail its adapter gate: %v", failures)
	}
	if got := RunSyncUpstreamAssets(repo, nil); got == 0 {
		t.Fatal("sync advanced despite Codex invariant mismatch")
	}
	if got := mustReadFile(t, statePath); got != before {
		t.Fatalf("state changed on refusal: %q", got)
	}
	if _, err := os.Lstat(filepath.Join(repo, "overlay", "gentle-ai", "assets", "upstream", "codex", "orchestrator.md")); !os.IsNotExist(err) {
		t.Fatalf("Codex snapshot written on refusal: %v", err)
	}
	if _, err := os.Lstat(filepath.Join(repo, "overlay", "gentle-ai", "assets", "upstream", "codex", "remote-authorization-contract.md")); !os.IsNotExist(err) {
		t.Fatalf("Codex shared snapshot written on refusal: %v", err)
	}
	writeJSONFixture(t, filepath.Join(repo, "overlay", "gentle-ai", "policy", "managed-assets.json"), ManagedAssetsManifest{Version: 1, Targets: map[string]ManagedAssetsTarget{"opencode": v4StructuralFixtureTarget()}})
	if got := RunAuditUpstream(repo, nil); got == 0 {
		t.Fatal("audit accepted missing Codex manifest target")
	}
	if got := RunSyncUpstreamAssets(repo, nil); got == 0 {
		t.Fatal("sync accepted missing Codex manifest target")
	}
	if got := mustReadFile(t, statePath); got != before {
		t.Fatalf("state changed after missing target refusal: %q", got)
	}
}

func TestVerifyUpstreamReleaseParityRejectsBinaryCheckoutVersionMismatch(t *testing.T) {
	tests := []struct {
		name          string
		binaryVersion string
		checkout      string
	}{
		{name: "binary newer", binaryVersion: "2.2.3", checkout: "v2.2.2"},
		{name: "binary older", binaryVersion: "2.2.2", checkout: "v2.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newReleaseGuardRepo(t, "v2.2.2", true)
			if _, err := runGit(repo, false, "tag", "v2.2.3"); err != nil {
				t.Fatalf("tag second release: %v", err)
			}
			if _, err := runGit(repo, false, "checkout", "--detach", tt.checkout); err != nil {
				t.Fatalf("check out %s: %v", tt.checkout, err)
			}
			installFakeGentleAI(t, "#!/bin/sh\necho 'gentle-ai "+tt.binaryVersion+"'\n")

			if _, err := verifyUpstreamReleaseParity(repo, ""); err == nil || !strings.Contains(err.Error(), "does not equal exact release") {
				t.Fatalf("verifyUpstreamReleaseParity() error = %v, want exact release mismatch", err)
			}
		})
	}
}

func TestVerifyUpstreamReleaseParityRejectsSameNamedBranchWithoutTag(t *testing.T) {
	repo := newReleaseGuardRepo(t, "", false)
	if _, err := runGit(repo, false, "branch", "v2.2.2"); err != nil {
		t.Fatalf("create same-named branch: %v", err)
	}
	installFakeGentleAI(t, "#!/bin/sh\necho 'gentle-ai 2.2.2'\n")

	if _, err := verifyUpstreamReleaseParity(repo, ""); err == nil || !strings.Contains(err.Error(), "exact release tag refs/tags/v2.2.2") {
		t.Fatalf("verifyUpstreamReleaseParity() error = %v, want missing exact tag failure", err)
	}
}

func TestVerifyUpstreamReleaseParityRejectsDirtyCheckout(t *testing.T) {
	tests := []struct {
		name  string
		dirty func(*testing.T, string)
	}{
		{name: "staged", dirty: func(t *testing.T, repo string) {
			mustWriteFile(t, filepath.Join(repo, "staged.txt"), []byte("staged\n"))
			if _, err := runGit(repo, false, "add", "staged.txt"); err != nil {
				t.Fatal(err)
			}
		}},
		{name: "unstaged", dirty: func(t *testing.T, repo string) {
			mustWriteFile(t, filepath.Join(repo, "tracked.txt"), []byte("changed\n"))
		}},
		{name: "untracked", dirty: func(t *testing.T, repo string) {
			mustWriteFile(t, filepath.Join(repo, "untracked.txt"), []byte("untracked\n"))
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newReleaseGuardRepo(t, "v2.2.2", false)
			installFakeGentleAI(t, "#!/bin/sh\necho 'gentle-ai 2.2.2'\n")
			tt.dirty(t, repo)
			if _, err := verifyUpstreamReleaseParity(repo, ""); err == nil || !strings.Contains(err.Error(), "checkout is dirty") {
				t.Fatalf("verifyUpstreamReleaseParity() error = %v, want dirty checkout failure", err)
			}
		})
	}
}

func TestVerifyUpstreamReleaseParityRejectsDirtySubmodule(t *testing.T) {
	submodule := newReleaseGuardRepo(t, "", false)
	repo := newReleaseGuardRepo(t, "", false)
	if _, err := runGit(repo, false, "-c", "protocol.file.allow=always", "submodule", "add", submodule, "deps/example"); err != nil {
		t.Fatalf("add submodule: %v", err)
	}
	if _, err := runGit(repo, false, "commit", "-am", "add submodule"); err != nil {
		t.Fatalf("commit submodule: %v", err)
	}
	if _, err := runGit(repo, false, "tag", "v2.2.2"); err != nil {
		t.Fatalf("tag release: %v", err)
	}
	installFakeGentleAI(t, "#!/bin/sh\necho 'gentle-ai 2.2.2'\n")
	mustWriteFile(t, filepath.Join(repo, "deps", "example", "tracked.txt"), []byte("dirty submodule\n"))

	if _, err := verifyUpstreamReleaseParity(repo, ""); err == nil || !strings.Contains(err.Error(), "checkout is dirty") {
		t.Fatalf("verifyUpstreamReleaseParity() error = %v, want dirty submodule failure", err)
	}
}

func TestRunSyncUpstreamAssetsReleaseRefusalDoesNotMutateRepo(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	upstreamRepo := newReleaseGuardRepo(t, "v2.2.2", true)
	t.Setenv(upstreamRepoEnvVar, upstreamRepo)
	installFakeGentleAI(t, "#!/bin/sh\necho 'gentle-ai 2.2.2'\n")

	repoRoot := t.TempDir()
	policy := Policy{}
	policy.Version = 1
	policy.Maintenance.StateFile = "overlay/gentle-ai/state/upstream-state.json"
	writeJSONFixture(t, filepath.Join(repoRoot, "overlay", "gentle-ai", "policy", "gentle-ai-policy.json"), policy)
	writeJSONFixture(t, filepath.Join(repoRoot, "overlay", "gentle-ai", "policy", "managed-assets.json"), ManagedAssetsManifest{
		Version: 1,
		Targets: map[string]ManagedAssetsTarget{"opencode": {}},
	})
	statePath := filepath.Join(repoRoot, "overlay", "gentle-ai", "state", "upstream-state.json")
	snapshotPath := filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "sentinel.txt")
	stateBefore := []byte("{\n  \"last_maintained_version\": \"2.2.1\"\n}\n")
	snapshotBefore := []byte("approved snapshot\n")
	mustWriteFile(t, statePath, stateBefore)
	mustWriteFile(t, snapshotPath, snapshotBefore)

	if exitCode := RunSyncUpstreamAssets(repoRoot, []string{"--release", "v2.2.2"}); exitCode == 0 {
		t.Fatal("RunSyncUpstreamAssets() exit code = 0, want release parity refusal")
	}
	if got := mustReadFile(t, statePath); got != string(stateBefore) {
		t.Fatalf("state changed on refused sync: %q", got)
	}
	if got := mustReadFile(t, snapshotPath); got != string(snapshotBefore) {
		t.Fatalf("snapshot changed on refused sync: %q", got)
	}
}

func TestVerifyUpstreamReleaseParityAllowsNewerTagOnlyWhenExplicit(t *testing.T) {
	repo := newReleaseGuardRepo(t, "v2.2.2", true)
	if _, err := runGit(repo, false, "tag", "v2.2.3"); err != nil {
		t.Fatalf("tag newer release: %v", err)
	}
	installFakeGentleAI(t, "#!/bin/sh\necho 'gentle-ai 2.2.2'\n")

	if _, err := verifyUpstreamReleaseParity(repo, ""); err == nil || !strings.Contains(err.Error(), "does not equal exact release") {
		t.Fatalf("default parity error = %v, want installed-release mismatch", err)
	}
	got, err := verifyUpstreamReleaseParity(repo, "v2.2.3")
	if err != nil {
		t.Fatalf("explicit newer release verification: %v", err)
	}
	if got.InstalledVersion != "2.2.2" || got.Version != "2.2.3" || got.Tag != "v2.2.3" {
		t.Fatalf("release identity = %#v, want installed 2.2.2 and selected v2.2.3", got)
	}
}

func TestVerifyUpstreamReleaseParityRejectsInvalidOrMissingExplicitTag(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want string
	}{
		{name: "missing tag", tag: "v2.2.3", want: "exact release tag refs/tags/v2.2.3"},
		{name: "branch shadow is not a tag", tag: "v2.2.2", want: "exact release tag refs/tags/v2.2.2"},
		{name: "invalid tag syntax", tag: "v2.2.3^{commit}", want: "invalid release"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoTag := "v2.2.2"
			if tt.name == "branch shadow is not a tag" {
				repoTag = "v2.2.1"
			}
			repo := newReleaseGuardRepo(t, repoTag, false)
			if tt.name == "branch shadow is not a tag" {
				if _, err := runGit(repo, false, "branch", tt.tag); err != nil {
					t.Fatalf("create branch shadow: %v", err)
				}
			}
			installFakeGentleAI(t, "#!/bin/sh\necho 'gentle-ai 2.2.2'\n")
			if _, err := verifyUpstreamReleaseParity(repo, tt.tag); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("verifyUpstreamReleaseParity() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestNormalizeAuditUpstreamArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		release string
		exit    int
	}{
		{name: "default", exit: -1},
		{name: "explicit release", args: []string{"--release", "v2.9.1"}, release: "v2.9.1", exit: -1},
		{name: "missing value", args: []string{"--release"}, exit: 1},
		{name: "invalid tag", args: []string{"--release", "release-v2.9.1"}, exit: 1},
		{name: "duplicate", args: []string{"--release", "v2.9.1", "--release", "v2.9.2"}, release: "v2.9.1", exit: 1},
		{name: "unknown flag", args: []string{"--other"}, exit: 1},
		{name: "positional argument", args: []string{"v2.9.1"}, exit: 1},
		{name: "help", args: []string{"--help"}, exit: 0},
		{name: "help does not hide unknown argument", args: []string{"--help", "--other"}, exit: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options, exit := normalizeAuditUpstreamArgs(tt.args)
			if exit != tt.exit || options.release != tt.release {
				t.Fatalf("normalizeAuditUpstreamArgs() = (%#v, %d), want release %q and exit %d", options, exit, tt.release, tt.exit)
			}
		})
	}
}

func TestPrintAuditReleaseSummaryLabelsPreviewWithoutAdoption(t *testing.T) {
	var output bytes.Buffer
	printAuditReleaseSummary(&output, upstreamReleaseIdentity{
		InstalledVersion: "2.9.1",
		Version:          "2.9.2",
		Tag:              "v2.9.2",
		Commit:           "abc123",
	})
	for _, want := range []string{
		"Installed binary version: 2.9.1",
		"Selected audited release: v2.9.2 (2.9.2)",
		"Verified release commit: abc123",
		"Preview only",
		"does not approve adoption, snapshot sync, or runtime refresh",
	} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("audit release output %q does not contain %q", output.String(), want)
		}
	}
}

func TestNormalizeSyncUpstreamAssetsArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		release string
		verbose bool
		exit    int
	}{
		{name: "default", exit: -1},
		{name: "release and verbose", args: []string{"--release", "v2.9.1", "--verbose"}, release: "v2.9.1", verbose: true, exit: -1},
		{name: "missing value", args: []string{"--release"}, exit: 1},
		{name: "invalid tag", args: []string{"--release", "release-v2.9.1"}, exit: 1},
		{name: "duplicate", args: []string{"--release", "v2.9.1", "--release", "v2.9.2"}, release: "v2.9.1", exit: 1},
		{name: "unknown flag", args: []string{"--other"}, exit: 1},
		{name: "help does not hide unknown argument", args: []string{"--help", "--other"}, exit: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options, exit := normalizeSyncUpstreamAssetsArgs(tt.args)
			if exit != tt.exit || options.release != tt.release || options.verbose != tt.verbose {
				t.Fatalf("normalizeSyncUpstreamAssetsArgs() = (%#v, %d), want release %q verbose %t exit %d", options, exit, tt.release, tt.verbose, tt.exit)
			}
		})
	}
}

func newReleaseGuardRepo(t *testing.T, tag string, ahead bool) string {
	t.Helper()
	repo := t.TempDir()
	mustWriteFile(t, filepath.Join(repo, "tracked.txt"), []byte("release\n"))
	mustInitGitRepo(t, repo)
	if tag != "" {
		if _, err := runGit(repo, false, "tag", tag); err != nil {
			t.Fatalf("tag %s: %v", tag, err)
		}
	}
	if ahead {
		mustWriteFile(t, filepath.Join(repo, "tracked.txt"), []byte("ahead\n"))
		if _, err := runGit(repo, false, "commit", "-am", "ahead of release"); err != nil {
			t.Fatalf("commit ahead of release: %v", err)
		}
	}
	return repo
}
