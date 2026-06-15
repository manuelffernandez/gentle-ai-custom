package overlay

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCategorizeGitDiffTreatsOnlyOpenCodeAndEngramSourcesAsManaged(t *testing.T) {
	target := testOpenCodeAgentsTarget()
	managed, unmanaged := categorizeGitDiff([]GitDiffEntry{
		{Status: "M", Path: "internal/assets/opencode/persona-gentleman.md"},
		{Status: "M", Path: "internal/assets/claude/engram-protocol.md"},
		{Status: "M", Path: "internal/assets/opencode/sdd-overlay-single.json"},
		{Status: "M", Path: "internal/assets/opencode/sdd-overlay-multi.json"},
		{Status: "M", Path: "internal/assets/opencode/plugins/model-variants.ts"},
		{Status: "M", Path: "internal/assets/claude/agents/jd-fix-agent.md"},
		{Status: "M", Path: "internal/assets/claude/commands/sdd-apply.md"},
		{Status: "A", Path: "internal/assets/claude/output-style-neutral.md"},
	}, buildManagedAssetCatalog(target))

	if len(managed) != 5 {
		t.Fatalf("managed entries = %d, want 5", len(managed))
	}
	if len(unmanaged) != 0 {
		t.Fatalf("unmanaged entries = %d, want 0", len(unmanaged))
	}
	keys := map[string]bool{}
	for _, entry := range managed {
		keys[entry.Record.Key] = true
	}
	for _, key := range []string{"opencode-persona-source", "opencode-engram-source", "opencode-overlay-single", "opencode-overlay-multi", "opencode-plugins"} {
		if !keys[key] {
			t.Fatalf("managed record keys missing %q", key)
		}
	}
	for _, key := range []string{"claude-agents", "claude-commands", "claude-output-style-neutral"} {
		if keys[key] {
			t.Fatalf("managed record keys unexpectedly include %q", key)
		}
	}
}

func TestSyncManagedTargetCopiesOpenCodeAgentsSourceFilesFromUpstream(t *testing.T) {
	upstreamRepo := t.TempDir()
	personaContent := []byte("## Rules\n\n- Upstream persona.\n")
	engramContent := []byte("## Engram Persistent Memory — Protocol\n\n- mem_save\n")
	overlaySingleContent := []byte("{\n  \"name\": \"single\"\n}\n")
	overlayMultiContent := []byte("{\n  \"name\": \"multi\"\n}\n")
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "persona-gentleman.md"), personaContent)
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "claude", "engram-protocol.md"), engramContent)
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "sdd-overlay-single.json"), overlaySingleContent)
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "sdd-overlay-multi.json"), overlayMultiContent)
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "plugins", "model-variants.ts"), []byte("export const modelVariants = [];\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "plugins", "skill-registry.ts"), []byte("export const skillRegistry = [];\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "plugins", "background-agents.ts"), []byte("export const backgroundAgents = [];\n"))

	repoRoot := t.TempDir()
	stats := &syncUpstreamAssetsStats{}
	if err := syncManagedTarget(repoRoot, upstreamRepo, testOpenCodeAgentsTarget(), newVerboseRecorder(false), stats); err != nil {
		t.Fatalf("syncManagedTarget() error = %v", err)
	}

	// Fix #4: verify stats
	if stats.FilesNew != 7 {
		t.Fatalf("stats.FilesNew = %d, want 7", stats.FilesNew)
	}
	if stats.DirsSynced != 1 {
		t.Fatalf("stats.DirsSynced = %d, want 1", stats.DirsSynced)
	}

	// Verify persona-gentleman.md was copied verbatim
	gotPersona, err := os.ReadFile(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "persona-gentleman.md"))
	if err != nil {
		t.Fatalf("read persona mirror: %v", err)
	}
	if string(gotPersona) != string(personaContent) {
		t.Fatalf("persona mirror content mismatch\nwant: %q\ngot:  %q", personaContent, gotPersona)
	}

	// Verify engram-protocol.md was copied verbatim
	gotEngram, err := os.ReadFile(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "engram-protocol.md"))
	if err != nil {
		t.Fatalf("read engram mirror: %v", err)
	}
	if string(gotEngram) != string(engramContent) {
		t.Fatalf("engram mirror content mismatch\nwant: %q\ngot:  %q", engramContent, gotEngram)
	}

	gotSingle, err := os.ReadFile(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "sdd-overlay-single.json"))
	if err != nil {
		t.Fatalf("read single overlay mirror: %v", err)
	}
	if string(gotSingle) != string(overlaySingleContent) {
		t.Fatalf("single overlay mirror content mismatch\nwant: %q\ngot:  %q", overlaySingleContent, gotSingle)
	}

	gotMulti, err := os.ReadFile(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "sdd-overlay-multi.json"))
	if err != nil {
		t.Fatalf("read multi overlay mirror: %v", err)
	}
	if string(gotMulti) != string(overlayMultiContent) {
		t.Fatalf("multi overlay mirror content mismatch\nwant: %q\ngot:  %q", overlayMultiContent, gotMulti)
	}

	for _, name := range []string{"model-variants.ts", "skill-registry.ts", "background-agents.ts"} {
		if _, err := os.Stat(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "plugins", name)); err != nil {
			t.Fatalf("read plugin mirror %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "claude")); !os.IsNotExist(err) {
		t.Fatalf("sync wrote unsupported claude snapshot tree")
	}

	// Verify no AGENTS.md was written (no materialization in sync)
	if _, err := os.Stat(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "AGENTS.md")); err == nil {
		t.Fatalf("sync wrote AGENTS.md but should not materialize during sync")
	}
}

func TestOpenCodeAgentsOwnedSourceExtendsMaterializedBaseline(t *testing.T) {
	if testing.Short() {
		t.Skip("integration: reads committed repo files to verify owned AGENTS.md invariants")
	}
	repoRoot := repoRootForTest(t)
	ownedSnapshot := mustReadFile(t, filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "owned", "opencode", "AGENTS.md"))

	// The owned file must contain the required local overlay sections.
	for _, required := range []string{
		"gentle-ai-custom:no-auto-commit",
		"gentle-ai-custom:gemini-override",
	} {
		if !strings.Contains(ownedSnapshot, required) {
			t.Fatalf("owned AGENTS.md missing required section %q", required)
		}
	}

	// The owned file must not contain stale content from old ownership model.
	if strings.Contains(ownedSnapshot, "Agent Skills Index") {
		t.Fatalf("owned AGENTS.md still contains old root index content")
	}

	// The owned file must start with the upstream persona section marker,
	// confirming the upstream baseline is the starting point.
	if !strings.HasPrefix(strings.TrimLeft(ownedSnapshot, " \t\r\n"), "<!-- gentle-ai:persona -->") {
		t.Fatalf("owned AGENTS.md does not start with the upstream persona section marker")
	}

	// The upstream source mirrors must exist and contain content.
	for _, mirrorPath := range []string{
		filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "persona-gentleman.md"),
		filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "engram-protocol.md"),
	} {
		content := mustReadFile(t, mirrorPath)
		if strings.TrimSpace(content) == "" {
			t.Fatalf("upstream source mirror %s is empty", mirrorPath)
		}
	}
	if _, err := os.Stat(filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "claude")); !os.IsNotExist(err) {
		t.Fatalf("owned AGENTS baseline should not materialize the claude snapshot tree")
	}
}

func testOpenCodeAgentsTarget() ManagedAssetsTarget {
	return ManagedAssetsTarget{
		WatchRoots: []string{
			"internal/assets/opencode",
			"internal/assets/claude/engram-protocol.md",
			"internal/assets/skills",
		},
		OwnedOverlayAssets: []OwnedOverlayAsset{
			{
				Key:              "opencode-persona-source",
				Class:            "upstream_source",
				Kind:             "upstream_source_file",
				UpstreamPath:     "internal/assets/opencode/persona-gentleman.md",
				RepoUpstreamPath: "overlay/gentle-ai/assets/upstream/opencode/persona-gentleman.md",
			},
			{
				Key:              "opencode-engram-source",
				Class:            "upstream_source",
				Kind:             "upstream_source_file",
				UpstreamPath:     "internal/assets/claude/engram-protocol.md",
				RepoUpstreamPath: "overlay/gentle-ai/assets/upstream/opencode/engram-protocol.md",
			},
			{
				Key:              "opencode-overlay-single",
				Class:            "upstream_source",
				Kind:             "upstream_source_file",
				UpstreamPath:     "internal/assets/opencode/sdd-overlay-single.json",
				RepoUpstreamPath: "overlay/gentle-ai/assets/upstream/opencode/sdd-overlay-single.json",
			},
			{
				Key:              "opencode-overlay-multi",
				Class:            "upstream_source",
				Kind:             "upstream_source_file",
				UpstreamPath:     "internal/assets/opencode/sdd-overlay-multi.json",
				RepoUpstreamPath: "overlay/gentle-ai/assets/upstream/opencode/sdd-overlay-multi.json",
			},
			{
				Key:              "opencode-plugins",
				Class:            "upstream_source",
				Kind:             "upstream_source_directory",
				UpstreamPath:     "internal/assets/opencode/plugins",
				RepoUpstreamPath: "overlay/gentle-ai/assets/upstream/opencode/plugins",
			},
			{
				Key:            "opencode-agents",
				Class:          "repo_owned_runtime",
				Kind:           "agent_instruction_file",
				RepoOwnedPath:  "overlay/gentle-ai/assets/owned/opencode/AGENTS.md",
				RuntimeTargets: []string{"~/.config/opencode/AGENTS.md"},
			},
		},
	}
}

func repoRootForTest(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot determine working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("cannot locate repo root: go.mod not found in any parent directory")
		}
		dir = parent
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(raw)
}

func mustWriteFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
