package overlay

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGitShowPreservesExactBlobBytes(t *testing.T) {
	upstream := t.TempDir()
	files := map[string]string{
		"trailing-newline.md":    "Codex orchestrator\n",
		"trailing-whitespace.md": "contract \n\n",
		"empty.md":               "",
	}
	for name, content := range files {
		mustWriteFile(t, filepath.Join(upstream, "sources", name), []byte(content))
	}
	if err := os.Symlink("trailing-newline.md", filepath.Join(upstream, "sources", "symlink.md")); err != nil {
		t.Skipf("symlink fixture unavailable: %v", err)
	}
	mustInitGitRepo(t, upstream)
	for name, want := range files {
		t.Run(name, func(t *testing.T) {
			got, err := runGitShow(upstream, "HEAD", "sources/"+name)
			if err != nil || got != want {
				t.Fatalf("runGitShow() = %q, %v; want exact bytes %q", got, err, want)
			}
		})
	}
	if normalized, err := runGit(upstream, false, "show", "HEAD:sources/trailing-whitespace.md"); err != nil || normalized != "contract" {
		t.Fatalf("runGit text normalization changed: %q, %v", normalized, err)
	}
	if _, err := runGitShow(upstream, "HEAD", "sources"); err == nil || !strings.Contains(err.Error(), "not a blob") {
		t.Fatalf("tree path accepted as a blob: %v", err)
	}
	if _, err := runGitShow(upstream, "HEAD", "sources/symlink.md"); err == nil || !strings.Contains(err.Error(), "not a regular blob") {
		t.Fatalf("symlink path accepted as a regular blob: %v", err)
	}
	if _, err := runGitShow(upstream, "HEAD", "sources/missing.md"); err == nil {
		t.Fatal("missing upstream blob accepted")
	}
}

func TestRunGitShowCodexSnapshotParityUsesExactBlobBytes(t *testing.T) {
	upstream, repo := t.TempDir(), t.TempDir()
	target := ManagedAssetsTarget{OwnedOverlayAssets: []OwnedOverlayAsset{
		{Key: "codex-orchestrator-source", Class: "upstream_source", Kind: "upstream_source_file", UpstreamPath: "internal/assets/codex/orchestrator.md", RepoUpstreamPath: "overlay/gentle-ai/assets/upstream/codex/orchestrator.md"},
		{Key: "codex-remote-authorization-source", Class: "upstream_source", Kind: "upstream_source_file", UpstreamPath: "internal/assets/generic/remote-authorization-contract.md", RepoUpstreamPath: "overlay/gentle-ai/assets/upstream/codex/remote-authorization-contract.md"},
	}}
	for _, asset := range target.OwnedOverlayAssets {
		content := []byte("upstream source: " + asset.Key + "\n")
		mustWriteFile(t, filepath.Join(upstream, filepath.FromSlash(asset.UpstreamPath)), content)
		mustWriteFile(t, filepath.Join(repo, filepath.FromSlash(asset.RepoUpstreamPath)), content)
	}
	mustInitGitRepo(t, upstream)
	if failures := checkCodexApprovedSnapshots(repo, upstream, target); len(failures) != 0 {
		t.Fatalf("byte-identical newline-terminated Codex snapshots reported drift: %v", failures)
	}
	changed := target.OwnedOverlayAssets[0].RepoUpstreamPath
	mustWriteFile(t, filepath.Join(repo, filepath.FromSlash(changed)), []byte("upstream source: codex-orchestrator-source"))
	if failures := checkCodexApprovedSnapshots(repo, upstream, target); len(failures) != 1 || !strings.Contains(failures[0], "differs from verified upstream source") {
		t.Fatalf("nonmatching approved snapshot did not fail closed: %v", failures)
	}
}
