package overlay

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCategorizeGitDiffTreatsOnlyOpenCodeAndEngramSourcesAsManaged(t *testing.T) {
	target := testOpenCodeAgentsTarget()
	managed, unmanaged := categorizeGitDiff([]GitDiffEntry{
		{Status: "M", Path: "internal/assets/opencode/persona-gentleman.md"},
		{Status: "M", Path: "internal/assets/engram/protocol.md"},
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
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "engram", "protocol.md"), engramContent)
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

func TestFormatUpstreamAssetSyncNoteRecordsPreviousAndNewCommit(t *testing.T) {
	got := formatUpstreamAssetSyncNote(
		"b6737129a295980962ec5e8daf41c971fac700b9",
		"b47f8af5d1cc9f6368bff17ac9e1c5694713d2c5",
	)
	want := "Updated approved upstream asset snapshots via sync-upstream-assets from b6737129a295980962ec5e8daf41c971fac700b9 to b47f8af5d1cc9f6368bff17ac9e1c5694713d2c5."
	if got != want {
		t.Fatalf("formatUpstreamAssetSyncNote() = %q, want %q", got, want)
	}
}

func TestFormatUpstreamAssetSyncNoteHandlesUnknownPreviousCommit(t *testing.T) {
	got := formatUpstreamAssetSyncNote("", "b47f8af5d1cc9f6368bff17ac9e1c5694713d2c5")
	want := "Updated approved upstream asset snapshots via sync-upstream-assets to b47f8af5d1cc9f6368bff17ac9e1c5694713d2c5."
	if got != want {
		t.Fatalf("formatUpstreamAssetSyncNote() = %q, want %q", got, want)
	}
}

func TestRunSyncUpstreamAssetsRecordsPreviousMaintainedCommitInStateNote(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	upstreamRepo := t.TempDir()
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "persona-gentleman.md"), []byte("## Rules\n\n- Persona v1.\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "engram", "protocol.md"), []byte("## Engram Persistent Memory — Protocol\n\n- v1\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "sdd-overlay-single.json"), []byte("{\n  \"name\": \"single\"\n}\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "sdd-overlay-multi.json"), []byte("{\n  \"name\": \"multi\"\n}\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "plugins", "model-variants.ts"), []byte("export const modelVariants = [];\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "plugins", "skill-registry.ts"), []byte("export const skillRegistry = [];\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "plugins", "background-agents.ts"), []byte("export const backgroundAgents = [];\n"))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "components", "sdd", "profiles.go"), []byte(`package opencode

var profilePhaseOrder = []string{
	"sdd-init",
	"sdd-apply",
}

func DetectProfiles() string {
	const orchPrefix = "sdd-orchestrator-"
	return orchPrefix
}

func ProfileAgentKeys(suffix string) []string {
	keys := []string{}
	keys = append(keys, "sdd-orchestrator"+suffix)
	return keys
}

func wiring(suffix string) {
	profileMap := map[string]any{}
	orchestratorMap := map[string]any{
		"tools": map[string]any{
			"task": true,
		},
	}
	_ = orchestratorMap
	taskPerms := map[string]any{
		"*":       "deny",
		"general": "allow",
	}
	for _, phase := range profilePhaseOrder {
		taskPerms[phase+suffix] = "allow"
	}
	for _, jd := range JDPhases() {
		taskPerms[jd] = "allow"
	}
	for _, reviewAgent := range ReviewPhases() {
		taskPerms[reviewAgent] = "allow"
	}
	profileMap["permission"] = map[string]any{
		"bash": true,
		"task": taskPerms,
	}
}
`))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "components", "sdd", "inject.go"), []byte(`package opencode

func inject() {
	orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
	_ = orchestratorRaw
	_ = ok
	orchestratorMap["prompt"] = renderSDDOrchestratorAsset(agent)
	taskRaw, ok := permissionRaw["task"].(map[string]any)
	_ = taskRaw
	_ = ok
}
`))
	mustInitGitRepo(t, upstreamRepo)

	previousCommit, err := runGit(upstreamRepo, false, "rev-parse", "HEAD")
	if err != nil {
		t.Fatalf("git rev-parse previous HEAD: %v", err)
	}

	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "assets", "opencode", "persona-gentleman.md"), []byte("## Rules\n\n- Persona v2.\n"))
	if _, err := runGit(upstreamRepo, false, "add", "."); err != nil {
		t.Fatalf("git add updated upstream fixture: %v", err)
	}
	if _, err := runGit(upstreamRepo, false, "commit", "-m", "update persona snapshot"); err != nil {
		t.Fatalf("git commit updated upstream fixture: %v", err)
	}

	upstreamHead, err := runGit(upstreamRepo, false, "rev-parse", "HEAD")
	if err != nil {
		t.Fatalf("git rev-parse upstream HEAD: %v", err)
	}

	repoRoot := t.TempDir()
	t.Setenv(upstreamRepoEnvVar, upstreamRepo)

	policy := Policy{}
	policy.Version = 1
	policy.Name = "test-overlay"
	policy.Description = "test policy"
	policy.Maintenance.IntentFile = "overlay/gentle-ai/policy/maintenance-intent.md"
	policy.Maintenance.StateFile = "overlay/gentle-ai/state/upstream-state.json"
	policy.Maintenance.LogFile = "overlay/gentle-ai/logs/update-log.md"
	policy.Upstream.RepoName = "gentle-ai"
	policy.Upstream.OrchestratorPromptPath = "internal/assets/opencode/sdd-orchestrator.md"
	policy.OpenCode.ConfigPath = "~/.config/opencode/opencode.json"
	policy.OpenCode.LocalConfigPath = "~/.config/gentle-ai-custom/opencode-local-config.json"
	policy.OpenCode.BaseOrchestratorKey = "gentle-orchestrator"
	policy.OpenCode.OrchestratorAgentKeys = []string{"gentle-orchestrator"}
	policy.OpenCode.OrchestratorAgentPrefixes = []string{"sdd-orchestrator"}
	policy.OpenCode.ProfileOrchestratorPrefix = "sdd-orchestrator-"
	policy.OpenCode.SDDPhases = []string{"sdd-init", "sdd-apply"}
	writeJSONFixture(t, filepath.Join(repoRoot, "overlay", "gentle-ai", "policy", "gentle-ai-policy.json"), policy)

	manifest := ManagedAssetsManifest{
		Version:     1,
		Status:      "active_runtime_canonical",
		Description: "test manifest",
		Targets: map[string]ManagedAssetsTarget{
			"opencode": {
				WatchRoots:                 testOpenCodeAgentsTarget().WatchRoots,
				StructuralInvariantSources: []string{"internal/components/sdd/profiles.go", "internal/components/sdd/inject.go"},
				OwnedOverlayAssets:         testOpenCodeAgentsTarget().OwnedOverlayAssets,
			},
		},
	}
	writeJSONFixture(t, filepath.Join(repoRoot, "overlay", "gentle-ai", "policy", "managed-assets.json"), manifest)

	statePath := filepath.Join(repoRoot, "overlay", "gentle-ai", "state", "upstream-state.json")
	writeJSONFixture(t, statePath, UpstreamState{
		LastMaintainedVersion: previousCommit,
		LastMaintainedCommit:  previousCommit,
		Notes:                 "before sync",
	})

	if exitCode := RunSyncUpstreamAssets(repoRoot, nil); exitCode != 0 {
		t.Fatalf("RunSyncUpstreamAssets() exit code = %d, want 0", exitCode)
	}

	var updatedState UpstreamState
	if err := readJSONFile(statePath, &updatedState); err != nil {
		t.Fatalf("read updated state: %v", err)
	}

	if updatedState.LastMaintainedCommit != upstreamHead {
		t.Fatalf("updated last_maintained_commit = %q, want %q", updatedState.LastMaintainedCommit, upstreamHead)
	}

	wantNote := formatUpstreamAssetSyncNote(previousCommit, upstreamHead)
	if updatedState.Notes != wantNote {
		t.Fatalf("updated notes = %q, want %q", updatedState.Notes, wantNote)
	}

	regressionNote := formatUpstreamAssetSyncNote(upstreamHead, upstreamHead)
	if updatedState.Notes == regressionNote {
		t.Fatalf("updated notes regressed to head->head note %q", regressionNote)
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

func TestUpstreamRuntimeSnapshotsAreNeverOwnedInstallSources(t *testing.T) {
	repoRoot := repoRootForTest(t)
	target, _, err := loadManagedAssetsTarget(repoRoot, "opencode")
	if err != nil {
		t.Fatalf("loadManagedAssetsTarget() error = %v", err)
	}
	foundSnapshots := 0
	for _, asset := range target.OwnedOverlayAssets {
		if asset.RepoOwnedPath != "" && asset.Class != "repo_owned_runtime" {
			t.Fatalf("non-owned asset %q declares repo_owned_path", asset.Key)
		}
		if asset.Class != "upstream_runtime_snapshot" {
			continue
		}
		foundSnapshots++
		if asset.RepoOwnedPath != "" {
			t.Fatalf("upstream runtime snapshot %q must not declare repo_owned_path", asset.Key)
		}
		if asset.RepoUpstreamPath == "" {
			t.Fatalf("upstream runtime snapshot %q must retain an audit snapshot", asset.Key)
		}
	}
	if foundSnapshots == 0 {
		t.Fatal("managed-assets does not declare upstream runtime snapshots")
	}
	if len(target.RuntimePatches) != 1 || target.RuntimePatches[0].Key != "judgment-retrospective-handoff" {
		t.Fatalf("runtime patches = %#v, want only judgment-retrospective-handoff", target.RuntimePatches)
	}
	ownedRoot := filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "owned", "opencode")
	var ownedFiles []string
	if err := filepath.Walk(ownedRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ownedFiles = append(ownedFiles, filepath.Base(path))
		}
		return nil
	}); err != nil {
		t.Fatalf("walk owned runtime tree: %v", err)
	}
	if len(ownedFiles) != 1 || ownedFiles[0] != "AGENTS.md" {
		t.Fatalf("owned OpenCode runtime files = %v, want only AGENTS.md", ownedFiles)
	}
}

func TestValidateRuntimeOwnershipBoundaryRejectsUpstreamInstallSource(t *testing.T) {
	target := ManagedAssetsTarget{OwnedOverlayAssets: []OwnedOverlayAsset{
		{Key: "audit-only", Class: "upstream_runtime_snapshot", RepoOwnedPath: "overlay/gentle-ai/assets/owned/opencode/forbidden.md"},
	}}
	if err := validateRuntimeOwnershipBoundary(target); err == nil {
		t.Fatal("validateRuntimeOwnershipBoundary() error = nil, want upstream/custom ownership overlap rejection")
	}
}

func TestValidateRuntimeOwnershipBoundaryAcceptsCurrentManifest(t *testing.T) {
	target, _, err := loadManagedAssetsTarget(repoRootForTest(t), "opencode")
	if err != nil {
		t.Fatal(err)
	}
	if err := validateRuntimeOwnershipBoundary(target); err != nil {
		t.Fatalf("validateRuntimeOwnershipBoundary() error = %v", err)
	}
}

func testOpenCodeAgentsTarget() ManagedAssetsTarget {
	return ManagedAssetsTarget{
		WatchRoots: []string{
			"internal/assets/opencode",
			"internal/assets/engram/protocol.md",
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
				UpstreamPath:     "internal/assets/engram/protocol.md",
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

func writeJSONFixture(t *testing.T, path string, value any) {
	t.Helper()
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal json fixture %s: %v", path, err)
	}
	mustWriteFile(t, path, append(raw, '\n'))
}

func TestProfileTaskScopingInvariantOKAcceptsEquivalentDenyFormatting(t *testing.T) {
	text := `
profileMap := map[string]any{}
taskPerms := map[string]any{
	"*":       "deny",
	"general": "allow",
	"explore": "allow",
}
for _, phase := range profilePhaseOrder {
	taskPerms[phase+suffix] = "allow"
}
for _, jd := range opencode.JDPhases() {
	taskPerms[jd] = "allow"
}
for _, reviewAgent := range opencode.ReviewPhases() {
	taskPerms[reviewAgent] = "allow"
}
profileMap["permission"] = map[string]any{
	"bash": true,
	"task": taskPerms,
}
`
	if !profileTaskScopingInvariantOK(text) {
		t.Fatal("profileTaskScopingInvariantOK() = false, want true for equivalent deny-by-default map formatting")
	}
}

func TestProfileTaskScopingInvariantOKRejectsUnwiredDenyMap(t *testing.T) {
	text := `
taskPerms := map[string]any{
	"*": "deny",
}
for _, phase := range profilePhaseOrder {
	taskPerms[phase+suffix] = "allow"
}
for _, jd := range opencode.JDPhases() {
	taskPerms[jd] = "allow"
}
for _, reviewAgent := range opencode.ReviewPhases() {
	taskPerms[reviewAgent] = "allow"
}
otherMap["permission"] = map[string]any{
	"task": randomPerms,
}
`
	if profileTaskScopingInvariantOK(text) {
		t.Fatal("profileTaskScopingInvariantOK() = true, want false when deny map is not wired to named profile permissions")
	}
}

func TestProfileTaskScopingInvariantOKRejectsCommentAndStringDecoys(t *testing.T) {
	text := `
// taskPerms := map[string]any{"*": "deny"}
message := "taskPerms[phase+suffix] = \"allow\"; taskPerms[jd] = \"allow\"; taskPerms[reviewAgent] = \"allow\"; profileMap[\"permission\"] = map[string]any{\"task\": taskPerms}"
_ = message
`
	if profileTaskScopingInvariantOK(text) {
		t.Fatal("profileTaskScopingInvariantOK() = true, want false for comment/string decoys")
	}
}

func TestProfileTaskToolInvariantOKAcceptsDirectProfileToolsWiring(t *testing.T) {
	text := `
profileMap := map[string]any{
	"tools": map[string]any{
		"task": true,
	},
}
`
	if !profileTaskToolInvariantOK(text) {
		t.Fatal("profileTaskToolInvariantOK() = false, want true for direct profile tools wiring")
	}
}

func TestProfileTaskToolInvariantOKAcceptsV2ReplacementWiring(t *testing.T) {
	text := `
orchEntry := map[string]any{
	"name": "sdd-orchestrator-demo",
	"tools": map[string]any{
		"__replace__": map[string]any{
			"task": true,
			"bash": true,
		},
	},
}
`
	if !profileTaskToolInvariantOK(text) {
		t.Fatal("profileTaskToolInvariantOK() = false, want true for v2 __replace__ tools wiring")
	}
}

func TestProfileTaskToolInvariantOKRejectsUnrelatedLiveTaskDecoy(t *testing.T) {
	text := `
decoy := map[string]any{
	"task": true,
}
otherMap["metadata"] = map[string]any{
	"task": true,
}
`
	if profileTaskToolInvariantOK(text) {
		t.Fatal("profileTaskToolInvariantOK() = true, want false for unrelated live task decoys")
	}
}

func TestProfileTaskScopingInvariantOKAcceptsV2ReplacementWiring(t *testing.T) {
	text := `
taskPerms := map[string]any{
	"*": "deny",
}
for _, phase := range profilePhaseOrder {
	taskPerms[phase+suffix] = "allow"
}
for _, jd := range opencode.JDPhases() {
	if hasProfileAssignment(profile, jd) {
		taskPerms[jd+suffix] = "allow"
	} else {
		taskPerms[jd] = "allow"
	}
}
for _, reviewAgent := range opencode.ReviewPhases() {
	taskPerms[reviewAgent] = "allow"
}
orchEntry := map[string]any{
	"permission": map[string]any{
		"task": map[string]any{
			"__replace__": taskPerms,
		},
	},
}
_ = orchEntry
`
	if !profileTaskScopingInvariantOK(text) {
		t.Fatal("profileTaskScopingInvariantOK() = false, want true for v2 __replace__ permission wiring")
	}
}

func TestEvaluateStructuralInvariantsRejectsProfileTaskScopingCommentDecoy(t *testing.T) {
	upstreamRepo := t.TempDir()
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "profiles.go"), []byte(`package opencode

var profilePhaseOrder = []string{
	"sdd-init",
	"sdd-apply",
}

func DetectProfiles() string {
	const orchPrefix = "sdd-orchestrator-"
	return orchPrefix
}

func ProfileAgentKeys(suffix string) []string {
	keys := []string{}
	keys = append(keys, "sdd-orchestrator"+suffix)
	return keys
}

func wiring() {
	profileMap := map[string]any{
		"tools": map[string]any{"task": true},
	}
	_ = profileMap
	// taskPerms := map[string]any{"*": "deny"}
	message := "taskPerms[phase+suffix] = \"allow\"; taskPerms[jd] = \"allow\"; taskPerms[reviewAgent] = \"allow\"; profileMap[\"permission\"] = map[string]any{\"task\": taskPerms}"
	_ = message
}
`))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "inject.go"), []byte(`package opencode

func inject() {
	orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
	_ = orchestratorRaw
	_ = ok
	orchestratorMap["prompt"] = renderSDDOrchestratorAsset(model.AgentOpenCode)
	taskRaw, ok := permissionRaw["task"].(map[string]any)
	_ = taskRaw
	_ = ok
}
`))
	mustInitGitRepo(t, upstreamRepo)

	target := ManagedAssetsTarget{StructuralInvariantSources: []string{"internal/profiles.go", "internal/inject.go"}}
	policy := Policy{}
	policy.OpenCode.SDDPhases = []string{"sdd-init", "sdd-apply"}

	result, failures := evaluateStructuralInvariants(upstreamRepo, target, policy)
	if result.TaskScopingOK {
		t.Fatal("evaluateStructuralInvariants() TaskScopingOK = true, want false for profile task-scoping comment/string decoy")
	}
	if len(failures) == 0 {
		t.Fatal("evaluateStructuralInvariants() failures = none, want task-scoping invariant failure")
	}
	if !strings.Contains(strings.Join(failures, "\n"), "deny-by-default task permission map") {
		t.Fatalf("evaluateStructuralInvariants() failures = %v, want task-scoping failure", failures)
	}
}

func TestEvaluateStructuralInvariantsRejectsProfileTaskToolDecoy(t *testing.T) {
	upstreamRepo := t.TempDir()
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "profiles.go"), []byte(`package opencode

var profilePhaseOrder = []string{
	"sdd-init",
	"sdd-apply",
}

func DetectProfiles() string {
	const orchPrefix = "sdd-orchestrator-"
	return orchPrefix
}

func ProfileAgentKeys(suffix string) []string {
	keys := []string{}
	keys = append(keys, "sdd-orchestrator"+suffix)
	return keys
}

func wiring(suffix string) {
	decoy := map[string]any{
		"task": true,
	}
	_ = decoy
	taskPerms := map[string]any{"*": "deny"}
	for _, phase := range profilePhaseOrder {
		taskPerms[phase+suffix] = "allow"
	}
	for _, jd := range JDPhases() {
		taskPerms[jd] = "allow"
	}
	for _, reviewAgent := range ReviewPhases() {
		taskPerms[reviewAgent] = "allow"
	}
	profileMap := map[string]any{}
	profileMap["permission"] = map[string]any{
		"bash": true,
		"task": taskPerms,
	}
}
`))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "inject.go"), []byte(`package opencode

func inject() {
	orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
	_ = orchestratorRaw
	_ = ok
	orchestratorMap["prompt"] = renderSDDOrchestratorAsset(model.AgentOpenCode)
	taskRaw, ok := permissionRaw["task"].(map[string]any)
	_ = taskRaw
	_ = ok
}
`))
	mustInitGitRepo(t, upstreamRepo)

	target := ManagedAssetsTarget{StructuralInvariantSources: []string{"internal/profiles.go", "internal/inject.go"}}
	policy := Policy{}
	policy.OpenCode.SDDPhases = []string{"sdd-init", "sdd-apply"}

	result, failures := evaluateStructuralInvariants(upstreamRepo, target, policy)
	if result.TaskScopingOK {
		t.Fatal("evaluateStructuralInvariants() TaskScopingOK = true, want false for task-tool decoy without profile tool wiring")
	}
	if len(failures) == 0 {
		t.Fatal("evaluateStructuralInvariants() failures = none, want task-tool invariant failure")
	}
	if !strings.Contains(strings.Join(failures, "\n"), "enables the task tool") {
		t.Fatalf("evaluateStructuralInvariants() failures = %v, want task-tool failure", failures)
	}
}

func TestBaseAssetInjectionInvariantOKAcceptsCanonicalRenderArguments(t *testing.T) {
	tests := []struct {
		name     string
		argument string
	}{
		{name: "current adapter-local agent", argument: "agent"},
		{name: "historical OpenCode selector", argument: "model.AgentOpenCode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := `
orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
orchestratorMap["prompt"] = renderSDDOrchestratorAsset(` + tt.argument + `)
taskRaw, ok := permissionRaw["task"].(map[string]any)
`
			if !baseAssetInjectionInvariantOK(text) {
				t.Fatalf("baseAssetInjectionInvariantOK() = false, want true for argument %q", tt.argument)
			}
		})
	}
}

func TestBaseAssetInjectionInvariantOKAcceptsLegacyAssetRead(t *testing.T) {
	text := `
orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
orchestratorMap["prompt"] = assets.MustRead(sddOrchestratorAsset(model.AgentOpenCode))
taskRaw, ok := permissionRaw["task"].(map[string]any)
`
	if !baseAssetInjectionInvariantOK(text) {
		t.Fatal("baseAssetInjectionInvariantOK() = false, want true for legacy asset-read binding")
	}
}

func TestBaseAssetInjectionInvariantOKRejectsDecoyRenderCall(t *testing.T) {
	text := `
orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
_ = renderSDDOrchestratorAsset(model.AgentOpenCode)
taskRaw, ok := permissionRaw["task"].(map[string]any)
`
	if baseAssetInjectionInvariantOK(text) {
		t.Fatal("baseAssetInjectionInvariantOK() = true, want false for decoy render call without prompt assignment")
	}
}

func TestBaseAssetInjectionInvariantOKRejectsWrongAssignmentTarget(t *testing.T) {
	text := `
orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
profileMap["prompt"] = renderSDDOrchestratorAsset(model.AgentOpenCode)
taskRaw, ok := permissionRaw["task"].(map[string]any)
`
	if baseAssetInjectionInvariantOK(text) {
		t.Fatal("baseAssetInjectionInvariantOK() = true, want false when render path is assigned to the wrong map")
	}
}

func TestBaseAssetInjectionInvariantOKRejectsArbitraryRenderArgument(t *testing.T) {
	text := `
orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
orchestratorMap["prompt"] = renderSDDOrchestratorAsset(candidate)
taskRaw, ok := permissionRaw["task"].(map[string]any)
`
	if baseAssetInjectionInvariantOK(text) {
		t.Fatal("baseAssetInjectionInvariantOK() = true, want false for arbitrary render argument")
	}
}

func TestBaseAssetInjectionInvariantOKRejectsMissingPromptAssignment(t *testing.T) {
	text := `
orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
taskRaw, ok := permissionRaw["task"].(map[string]any)
`
	if baseAssetInjectionInvariantOK(text) {
		t.Fatal("baseAssetInjectionInvariantOK() = true, want false when prompt assignment is missing")
	}
}

func TestBaseAssetInjectionInvariantOKRejectsCommentAndStringDecoys(t *testing.T) {
	text := `
// orchestratorMap["prompt"] = renderSDDOrchestratorAsset(model.AgentOpenCode)
message := "orchestratorMap[\"prompt\"] = assets.MustRead(sddOrchestratorAsset(model.AgentOpenCode))"
orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
taskRaw, ok := permissionRaw["task"].(map[string]any)
_ = message
`
	if baseAssetInjectionInvariantOK(text) {
		t.Fatal("baseAssetInjectionInvariantOK() = true, want false for comment/string decoys")
	}
}

func TestEvaluateStructuralInvariantsAcceptsUpstreamV220Shape(t *testing.T) {
	upstreamRepo := t.TempDir()
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "profiles.go"), []byte(`package opencode

var profilePhaseOrder = []string{
	"sdd-init",
	"sdd-apply",
}

func DetectProfiles() string {
	const orchPrefix = "sdd-orchestrator-"
	return orchPrefix
}

func ProfileAgentKeys(suffix string) []string {
	keys := []string{}
	keys = append(keys, "sdd-orchestrator"+suffix)
	return keys
}

func wiring(suffix string) {
	profileMap := map[string]any{}
	orchestratorMap := map[string]any{
		"tools": map[string]any{
			"task": true,
		},
	}
	_ = orchestratorMap
	taskPerms := map[string]any{
		"*":       "deny",
		"general": "allow",
	}
	for _, phase := range profilePhaseOrder {
		taskPerms[phase+suffix] = "allow"
	}
	for _, jd := range JDPhases() {
		taskPerms[jd] = "allow"
	}
	for _, reviewAgent := range ReviewPhases() {
		taskPerms[reviewAgent] = "allow"
	}
	profileMap["permission"] = map[string]any{
		"bash": true,
		"task": taskPerms,
	}
}
`))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "inject.go"), []byte(`package opencode

func inject() {
	orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
	_ = orchestratorRaw
	_ = ok
	orchestratorMap["prompt"] = renderSDDOrchestratorAsset(agent)
	taskRaw, ok := permissionRaw["task"].(map[string]any)
	_ = taskRaw
	_ = ok
}
`))
	mustInitGitRepo(t, upstreamRepo)

	target := ManagedAssetsTarget{StructuralInvariantSources: []string{"internal/profiles.go", "internal/inject.go"}}
	policy := Policy{}
	policy.OpenCode.SDDPhases = []string{"sdd-init", "sdd-apply"}

	result, failures := evaluateStructuralInvariants(upstreamRepo, target, policy)
	if len(failures) != 0 {
		t.Fatalf("evaluateStructuralInvariants() failures = %v, want none for upstream v2.2.0 shape", failures)
	}
	if !result.TaskScopingOK {
		t.Fatal("evaluateStructuralInvariants() TaskScopingOK = false, want true for equivalent wiring")
	}
	if !result.BaseAssetInjectionOK {
		t.Fatal("evaluateStructuralInvariants() BaseAssetInjectionOK = false, want true")
	}
}

func TestEvaluateStructuralInvariantsRejectsBaseAssetCommentDecoy(t *testing.T) {
	upstreamRepo := t.TempDir()
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "profiles.go"), []byte(`package opencode

var profilePhaseOrder = []string{
	"sdd-init",
	"sdd-apply",
}

func DetectProfiles() string {
	const orchPrefix = "sdd-orchestrator-"
	return orchPrefix
}

func ProfileAgentKeys(suffix string) []string {
	keys := []string{}
	keys = append(keys, "sdd-orchestrator"+suffix)
	return keys
}

func wiring(suffix string) {
	profileMap := map[string]any{}
	orchestratorMap := map[string]any{
		"task": true,
	}
	_ = orchestratorMap
	taskPerms := map[string]any{"*": "deny"}
	for _, phase := range profilePhaseOrder {
		taskPerms[phase+suffix] = "allow"
	}
	for _, jd := range JDPhases() {
		taskPerms[jd] = "allow"
	}
	for _, reviewAgent := range ReviewPhases() {
		taskPerms[reviewAgent] = "allow"
	}
	profileMap["permission"] = map[string]any{
		"bash": true,
		"task": taskPerms,
	}
}
`))
	mustWriteFile(t, filepath.Join(upstreamRepo, "internal", "inject.go"), []byte(`package opencode

func inject() {
	orchestratorRaw, ok := agentsMap["gentle-orchestrator"]
	_ = orchestratorRaw
	_ = ok
	// orchestratorMap["prompt"] = renderSDDOrchestratorAsset(model.AgentOpenCode)
	message := "orchestratorMap[\"prompt\"] = assets.MustRead(sddOrchestratorAsset(model.AgentOpenCode))"
	_ = message
	taskRaw, ok := permissionRaw["task"].(map[string]any)
	_ = taskRaw
	_ = ok
}
`))
	mustInitGitRepo(t, upstreamRepo)

	target := ManagedAssetsTarget{StructuralInvariantSources: []string{"internal/profiles.go", "internal/inject.go"}}
	policy := Policy{}
	policy.OpenCode.SDDPhases = []string{"sdd-init", "sdd-apply"}

	result, failures := evaluateStructuralInvariants(upstreamRepo, target, policy)
	if result.BaseAssetInjectionOK {
		t.Fatal("evaluateStructuralInvariants() BaseAssetInjectionOK = true, want false for comment/string decoy")
	}
	if len(failures) == 0 {
		t.Fatal("evaluateStructuralInvariants() failures = none, want invariant failure")
	}
	if !strings.Contains(strings.Join(failures, "\n"), "canonical SDD asset render path") {
		t.Fatalf("evaluateStructuralInvariants() failures = %v, want base asset injection failure", failures)
	}
}

func mustInitGitRepo(t *testing.T, repo string) {
	t.Helper()
	if _, err := runGit(repo, false, "init"); err != nil {
		t.Fatalf("git init %s: %v", repo, err)
	}
	if _, err := runGit(repo, false, "config", "user.name", "Test User"); err != nil {
		t.Fatalf("git config user.name %s: %v", repo, err)
	}
	if _, err := runGit(repo, false, "config", "user.email", "test@example.com"); err != nil {
		t.Fatalf("git config user.email %s: %v", repo, err)
	}
	if _, err := runGit(repo, false, "add", "."); err != nil {
		t.Fatalf("git add %s: %v", repo, err)
	}
	if _, err := runGit(repo, false, "commit", "-m", "test fixture"); err != nil {
		t.Fatalf("git commit %s: %v", repo, err)
	}
}
