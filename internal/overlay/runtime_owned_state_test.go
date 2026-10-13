package overlay

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type runtimeOwnedStateAgent struct {
	basePath string
}

func (a runtimeOwnedStateAgent) Name() string              { return "test-runtime-owned-state" }
func (a runtimeOwnedStateAgent) BasePath() (string, error) { return a.basePath, nil }
func (a runtimeOwnedStateAgent) BuildCommandContent(cmd customCommand, body string) string {
	return body
}
func (a runtimeOwnedStateAgent) ApplyOverlay(repoRoot string, options applyPolicyOptions) int {
	return 0
}

func TestReconcileOwnedRuntimeOutputsTracksPruneModeDirectoryAssets(t *testing.T) {
	repoRoot := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	basePath := filepath.Join(homeDir, ".config", "opencode")
	sharedSkillsRoot := filepath.Join(repoRoot, "shared", "skills")

	ownedSourceRoot := filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "owned", "opencode", "skills", "judgment-day")
	mustWriteFile(t, filepath.Join(ownedSourceRoot, "SKILL.md"), []byte("owned skill\n"))
	mustWriteFile(t, filepath.Join(ownedSourceRoot, "references", "prompts-and-formats.md"), []byte("owned prompt\n"))

	stalePath := filepath.Join(basePath, "skills", "judgment-day", "legacy.md")
	mustWriteFile(t, stalePath, []byte("stale\n"))
	state := ownedRuntimeState{Version: 1, Paths: []string{stalePath}}
	encoded, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		t.Fatalf("marshal state: %v", err)
	}
	mustWriteFile(t, filepath.Join(basePath, ".gentle-ai-custom-owned-state.json"), encoded)

	target := ManagedAssetsTarget{
		OwnedOverlayAssets: []OwnedOverlayAsset{
			{
				Key:             "judgment-day-runtime",
				Class:           "repo_owned_runtime",
				Kind:            "skill_directory",
				RepoOwnedPath:   "overlay/gentle-ai/assets/owned/opencode/skills/judgment-day",
				RuntimeSyncMode: "prune",
				RuntimeTargets:  []string{"~/.config/opencode/skills/judgment-day"},
			},
		},
	}

	if err := reconcileOwnedRuntimeOutputs(runtimeOwnedStateAgent{basePath: basePath}, repoRoot, target, sharedSkillsRoot, newVerboseRecorder(false)); err != nil {
		t.Fatalf("reconcileOwnedRuntimeOutputs() error = %v", err)
	}

	if pathExists(stalePath) {
		t.Fatalf("stale runtime path still exists: %s", stalePath)
	}

	gotStateRaw := mustReadFile(t, filepath.Join(basePath, ".gentle-ai-custom-owned-state.json"))
	var gotState ownedRuntimeState
	if err := json.Unmarshal([]byte(gotStateRaw), &gotState); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
	wantPaths := []string{
		filepath.Join(basePath, "skills", "judgment-day", "SKILL.md"),
		filepath.Join(basePath, "skills", "judgment-day", "references", "prompts-and-formats.md"),
	}
	if !reflect.DeepEqual(gotState.Paths, wantPaths) {
		t.Fatalf("owned runtime paths = %v, want %v", gotState.Paths, wantPaths)
	}
}

func TestCollectOwnedRuntimePathsTracksEveryRepoOwnedAsset(t *testing.T) {
	repoRoot := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	basePath := filepath.Join(homeDir, ".config", "opencode")
	sharedSkillsRoot := filepath.Join(repoRoot, "shared", "skills")

	// Write source files for merge-mode and no-mode assets.
	mergeSourcePath := filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "owned", "opencode", "commands")
	mustWriteFile(t, filepath.Join(mergeSourcePath, "cmd.md"), []byte("merge\n"))

	noModeSourcePath := filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "owned", "opencode", "AGENTS.md")
	mustWriteFile(t, noModeSourcePath, []byte("no-mode\n"))

	// Write source directory for prune-mode asset (matches production shape: skill_directory).
	pruneSourcePath := filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "owned", "opencode", "skills", "test-prune-skill")
	mustWriteFile(t, filepath.Join(pruneSourcePath, "SKILL.md"), []byte("prune skill\n"))
	mustWriteFile(t, filepath.Join(pruneSourcePath, "refs", "ref.md"), []byte("ref\n"))

	tests := []struct {
		name      string
		asset     OwnedOverlayAsset
		wantPaths []string // expected tracked paths; nil means none
	}{
		{
			name: "single-file asset without sync mode is tracked",
			asset: OwnedOverlayAsset{
				Key:             "opencode-agents",
				Class:           "repo_owned_runtime",
				Kind:            "agent_instruction_file",
				RepoOwnedPath:   "overlay/gentle-ai/assets/owned/opencode/AGENTS.md",
				RuntimeSyncMode: "",
				RuntimeTargets:  []string{"~/.config/opencode/AGENTS.md"},
			},
			wantPaths: []string{filepath.Join(basePath, "AGENTS.md")},
		},
		{
			name: "merge-mode directory asset IS tracked",
			asset: OwnedOverlayAsset{
				Key:             "sdd-commands",
				Class:           "repo_owned_runtime",
				Kind:            "command_directory",
				RepoOwnedPath:   "overlay/gentle-ai/assets/owned/opencode/commands",
				RuntimeSyncMode: "merge",
				RuntimeTargets:  []string{"~/.config/opencode/commands"},
			},
			wantPaths: []string{filepath.Join(basePath, "commands", "cmd.md")},
		},
		{
			name: "prune-mode directory asset IS tracked with dir-walk",
			asset: OwnedOverlayAsset{
				Key:             "test-prune-skill",
				Class:           "repo_owned_runtime",
				Kind:            "skill_directory",
				RepoOwnedPath:   "overlay/gentle-ai/assets/owned/opencode/skills/test-prune-skill",
				RuntimeSyncMode: "prune",
				RuntimeTargets:  []string{"~/.config/opencode/skills/test-prune-skill"},
			},
			wantPaths: []string{
				filepath.Join(basePath, "skills", "test-prune-skill", "SKILL.md"),
				filepath.Join(basePath, "skills", "test-prune-skill", "refs", "ref.md"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := ManagedAssetsTarget{
				OwnedOverlayAssets: []OwnedOverlayAsset{tt.asset},
			}
			paths, err := collectOwnedRuntimePaths(basePath, repoRoot, target, sharedSkillsRoot)
			if err != nil {
				t.Fatalf("collectOwnedRuntimePaths() error = %v", err)
			}
			if !reflect.DeepEqual(paths, tt.wantPaths) {
				t.Errorf("tracked paths = %v, want %v", paths, tt.wantPaths)
			}
		})
	}
}

func TestApplyChainedReferenceCleanupRequiresSanitizedSource(t *testing.T) {
	repo := t.TempDir()
	source := filepath.Join(repo, "shared", "skills", "chained-pr", "SKILL.md")
	asset := filepath.Join(repo, "shared", "skills", "chained-pr", "assets", "chaining-details.md")
	target := ManagedAssetsTarget{RepoOwnedSharedSkills: RepoOwnedSharedSkills{
		Root: "shared/skills", RuntimeRoot: "~/.codex/skills", Allowlist: []string{"chained-pr"},
	}}
	mustWriteFile(t, source, []byte("[references/chaining-details.md](references/chaining-details.md)\n"))
	mustWriteFile(t, asset, []byte("sanitized asset\n"))
	if err := requireSanitizedChainedSkill(repo, "codex", target); err == nil {
		t.Fatal("upstream-binding source accepted")
	}
	mustWriteFile(t, source, []byte("[assets/chaining-details.md](assets/chaining-details.md)\n"))
	if err := requireSanitizedChainedSkill(repo, "codex", target); err != nil {
		t.Fatalf("sanitized source rejected: %v", err)
	}
	target.RepoOwnedSharedSkills.Allowlist = nil
	if err := requireSanitizedChainedSkill(repo, "codex", target); err == nil {
		t.Fatal("missing sanitized skill selection accepted")
	}
}

func TestReconcileOwnedRuntimeOutputsDeletesMissingCustomOutputs(t *testing.T) {
	repoRoot := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	basePath := filepath.Join(homeDir, ".config", "opencode")
	sharedSkillsRoot := filepath.Join(repoRoot, "shared", "skills")

	// Write the prune-mode source (still active; its files are current).
	pruneSourceRoot := filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "owned", "opencode", "skills", "judgment-day")
	mustWriteFile(t, filepath.Join(pruneSourceRoot, "SKILL.md"), []byte("owned skill\n"))

	// stalePrunePath: a prune-mode file that no longer appears in the source tree.
	// It must be deleted during reconciliation.
	stalePrunePath := filepath.Join(basePath, "skills", "judgment-day", "legacy.md")
	mustWriteFile(t, stalePrunePath, []byte("stale\n"))

	// currentPrunePath: a prune-mode file that IS in the source tree.
	// It must NOT be deleted.
	currentPrunePath := filepath.Join(basePath, "skills", "judgment-day", "SKILL.md")
	mustWriteFile(t, currentPrunePath, []byte("current\n"))

	// Seed the previous state with both paths so reconciliation can compute the delta.
	state := ownedRuntimeState{Version: 1, Paths: []string{currentPrunePath, stalePrunePath}}
	encoded, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		t.Fatalf("marshal state: %v", err)
	}
	mustWriteFile(t, filepath.Join(basePath, ".gentle-ai-custom-owned-state.json"), encoded)

	target := ManagedAssetsTarget{
		OwnedOverlayAssets: []OwnedOverlayAsset{
			{
				Key:             "judgment-day-runtime",
				Class:           "repo_owned_runtime",
				Kind:            "skill_directory",
				RepoOwnedPath:   "overlay/gentle-ai/assets/owned/opencode/skills/judgment-day",
				RuntimeSyncMode: "prune",
				RuntimeTargets:  []string{"~/.config/opencode/skills/judgment-day"},
			},
		},
	}

	if err := reconcileOwnedRuntimeOutputs(runtimeOwnedStateAgent{basePath: basePath}, repoRoot, target, sharedSkillsRoot, newVerboseRecorder(false)); err != nil {
		t.Fatalf("reconcileOwnedRuntimeOutputs() error = %v", err)
	}

	// The stale prune-mode path must have been deleted.
	if pathExists(stalePrunePath) {
		t.Errorf("stale prune-mode path still exists: %s", stalePrunePath)
	}

	// The current prune-mode path must still exist.
	if !pathExists(currentPrunePath) {
		t.Errorf("current prune-mode path was incorrectly deleted: %s", currentPrunePath)
	}

	// The saved state contains exactly the currently custom-owned output.
	gotStateRaw := mustReadFile(t, filepath.Join(basePath, ".gentle-ai-custom-owned-state.json"))
	var gotState ownedRuntimeState
	if err := json.Unmarshal([]byte(gotStateRaw), &gotState); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
	wantPaths := []string{currentPrunePath}
	if !reflect.DeepEqual(gotState.Paths, wantPaths) {
		t.Errorf("saved state paths = %v, want %v", gotState.Paths, wantPaths)
	}
}

func TestCollectOwnedRuntimePathsTracksSingleFileAssetWithoutSyncMode(t *testing.T) {
	repoRoot := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	basePath := filepath.Join(homeDir, ".config", "opencode")
	sharedSkillsRoot := filepath.Join(repoRoot, "shared", "skills")

	// Write the source file for the no-mode asset (e.g. AGENTS.md).
	noModeSrc := filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "owned", "opencode", "AGENTS.md")
	mustWriteFile(t, noModeSrc, []byte("agents\n"))

	noModeTarget := filepath.Join(basePath, "AGENTS.md")

	target := ManagedAssetsTarget{
		OwnedOverlayAssets: []OwnedOverlayAsset{
			{
				Key:             "opencode-agents",
				Class:           "repo_owned_runtime",
				Kind:            "agent_instruction_file",
				RepoOwnedPath:   "overlay/gentle-ai/assets/owned/opencode/AGENTS.md",
				RuntimeSyncMode: "",
				RuntimeTargets:  []string{"~/.config/opencode/AGENTS.md"},
			},
		},
	}

	paths, err := collectOwnedRuntimePaths(basePath, repoRoot, target, sharedSkillsRoot)
	if err != nil {
		t.Fatalf("collectOwnedRuntimePaths() error = %v", err)
	}
	if !reflect.DeepEqual(paths, []string{noModeTarget}) {
		t.Fatalf("tracked runtime paths = %v, want %s", paths, noModeTarget)
	}
}

func TestReconcileOwnedRuntimeOutputsPreservesUpstreamManagedFilesFromOldState(t *testing.T) {
	repoRoot := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	basePath := filepath.Join(homeDir, ".config", "opencode")
	sharedSkillsRoot := filepath.Join(repoRoot, "shared", "skills")
	upstreamCommand := filepath.Join(basePath, "commands", "sdd-apply.md")
	mustWriteFile(t, upstreamCommand, []byte("upstream materialized command\n"))
	state := ownedRuntimeState{Version: 1, Paths: []string{upstreamCommand}}
	encoded, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, filepath.Join(basePath, ".gentle-ai-custom-owned-state.json"), encoded)
	target := ManagedAssetsTarget{OwnedOverlayAssets: []OwnedOverlayAsset{{
		Key: "sdd-commands", Class: "upstream_runtime_snapshot", Kind: "command_directory",
		RuntimeTargets: []string{"~/.config/opencode/commands"},
	}}}

	if err := reconcileOwnedRuntimeOutputs(runtimeOwnedStateAgent{basePath: basePath}, repoRoot, target, sharedSkillsRoot, newVerboseRecorder(false)); err != nil {
		t.Fatal(err)
	}
	if got := mustReadFile(t, upstreamCommand); got != "upstream materialized command\n" {
		t.Fatalf("upstream-managed command changed: %q", got)
	}
}

func TestReconcileOwnedRuntimeOutputsMigrationPreservesUpstreamAndTracksOwnedFile(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home, configPath := writeCompatibleRuntimeFixture(t, repoRoot)
	basePath := filepath.Dir(configPath)
	target, _, err := loadManagedAssetsTarget(repoRoot, "opencode")
	if err != nil {
		t.Fatal(err)
	}
	// Patch targets must be protected independently of retained-skill protection.
	target.RetainedUpstreamSkills = nil

	judgmentSkill := filepath.Join(basePath, "skills", "judgment-day", "SKILL.md")
	judgmentPrompt := filepath.Join(basePath, "skills", "judgment-day", "references", "prompts-and-formats.md")
	agentsPath := filepath.Join(basePath, "AGENTS.md")
	upstreamPrompt := filepath.Join(basePath, "prompts", "sdd", "sdd-init.md")
	sharedContract := filepath.Join(basePath, "skills", "_shared", "review-ledger-contract.md")
	upstreamSkill := filepath.Join(basePath, "skills", "_shared", "research-lifecycle.md")
	obsoleteCustom := filepath.Join(basePath, "legacy-custom", "obsolete.md")
	mustWriteFile(t, agentsPath, []byte("stale agents\n"))
	mustWriteFile(t, upstreamPrompt, []byte("upstream prompt snapshot\n"))
	mustWriteFile(t, upstreamSkill, []byte("upstream retained skill\n"))
	mustWriteFile(t, obsoleteCustom, []byte("obsolete custom output\n"))
	writeOwnedRuntimeStateFixture(t, basePath, []string{
		judgmentSkill,
		judgmentPrompt,
		agentsPath,
		upstreamPrompt,
		sharedContract,
		upstreamSkill,
		obsoleteCustom,
	})

	options := applyPolicyOptions{
		recorder:     newVerboseRecorder(false),
		skillTargets: []string{filepath.Join(basePath, "skills")},
	}
	agent := runtimeOwnedStateAgent{basePath: basePath}
	sharedSkillsRoot := filepath.Join(repoRoot, filepath.FromSlash(target.RepoOwnedSharedSkills.Root))
	commandBodiesRoot := filepath.Join(repoRoot, filepath.FromSlash(target.RepoOwnedCommandBodies.Root))
	commandBodies := mapCommandBodiesByKey(commandBodiesRoot, target.RepoOwnedCommandBodies.Entries)
	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("runApplyPolicyWithOptions() code = %d, want 0", code)
	}
	if err := installAgentAssets(agent, target, sharedSkillsRoot, commandBodies, options.recorder); err != nil {
		t.Fatal(err)
	}
	if err := reconcileOwnedRuntimeOutputs(agent, repoRoot, target, sharedSkillsRoot, options.recorder); err != nil {
		t.Fatal(err)
	}

	for path, marker := range map[string]string{
		judgmentSkill:  "invoke the repo-owned `judgment-retrospective` skill",
		judgmentPrompt: "## Terminal Retrospective Handoff",
	} {
		if !strings.Contains(mustReadFile(t, path), marker) {
			t.Fatalf("patched upstream file %s did not survive stale reconciliation", path)
		}
	}
	for _, path := range []string{upstreamPrompt, sharedContract, upstreamSkill} {
		if !pathExists(path) {
			t.Fatalf("upstream-returned runtime path was deleted: %s", path)
		}
	}
	if pathExists(obsoleteCustom) {
		t.Fatalf("truly obsolete custom output survived: %s", obsoleteCustom)
	}
	if got := mustReadFile(t, agentsPath); got != mustReadFile(t, filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "owned", "opencode", "AGENTS.md")) {
		t.Fatal("freshly installed repo-owned AGENTS.md did not survive reconciliation")
	}
	state, err := loadOwnedRuntimeState(filepath.Join(basePath, ".gentle-ai-custom-owned-state.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !stringSliceContains(state.Paths, agentsPath) {
		t.Fatal("current state does not record repo-owned AGENTS.md")
	}
	for _, upstreamPath := range []string{judgmentSkill, judgmentPrompt, upstreamPrompt, sharedContract, upstreamSkill} {
		if stringSliceContains(state.Paths, upstreamPath) {
			t.Fatalf("current state incorrectly claims upstream-owned path %s", upstreamPath)
		}
	}

	first := snapshotTree(t, home)
	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("repeat runApplyPolicyWithOptions() code = %d, want 0", code)
	}
	if err := installAgentAssets(agent, target, sharedSkillsRoot, commandBodies, options.recorder); err != nil {
		t.Fatal(err)
	}
	if err := reconcileOwnedRuntimeOutputs(agent, repoRoot, target, sharedSkillsRoot, options.recorder); err != nil {
		t.Fatal(err)
	}
	if second := snapshotTree(t, home); !reflect.DeepEqual(second, first) {
		t.Fatal("repeat apply and owned-state reconciliation is not idempotent")
	}
}

func TestReconcileOwnedRuntimeOutputsUsesSafeProtectionContainment(t *testing.T) {
	t.Run("retained skill descendant is preserved", func(t *testing.T) {
		basePath, repoRoot, sharedSkillsRoot := runtimeStateTestPaths(t)
		path := filepath.Join(basePath, "skills", "judgment-day", "references", "format.md")
		mustWriteFile(t, path, []byte("upstream\n"))
		writeOwnedRuntimeStateFixture(t, basePath, []string{path})
		target := ManagedAssetsTarget{RetainedUpstreamSkills: []UpstreamSkillRef{{Key: "judgment-day"}}}
		if err := reconcileOwnedRuntimeOutputs(runtimeOwnedStateAgent{basePath: basePath}, repoRoot, target, sharedSkillsRoot, newVerboseRecorder(false)); err != nil {
			t.Fatal(err)
		}
		if !pathExists(path) {
			t.Fatal("retained upstream skill descendant was deleted")
		}
	})

	t.Run("prefix collision is deleted", func(t *testing.T) {
		basePath, repoRoot, sharedSkillsRoot := runtimeStateTestPaths(t)
		path := filepath.Join(basePath, "skills", "judgment-day-copy", "legacy.md")
		mustWriteFile(t, path, []byte("obsolete\n"))
		writeOwnedRuntimeStateFixture(t, basePath, []string{path})
		target := ManagedAssetsTarget{RetainedUpstreamSkills: []UpstreamSkillRef{{Key: "judgment-day"}}}
		if err := reconcileOwnedRuntimeOutputs(runtimeOwnedStateAgent{basePath: basePath}, repoRoot, target, sharedSkillsRoot, newVerboseRecorder(false)); err != nil {
			t.Fatal(err)
		}
		if pathExists(path) {
			t.Fatal("prefix-collision path was incorrectly protected")
		}
	})

	t.Run("outside base path is rejected", func(t *testing.T) {
		basePath, repoRoot, sharedSkillsRoot := runtimeStateTestPaths(t)
		path := filepath.Join(filepath.Dir(basePath), "outside", "legacy.md")
		mustWriteFile(t, path, []byte("outside\n"))
		writeOwnedRuntimeStateFixture(t, basePath, []string{path})
		if err := reconcileOwnedRuntimeOutputs(runtimeOwnedStateAgent{basePath: basePath}, repoRoot, ManagedAssetsTarget{}, sharedSkillsRoot, newVerboseRecorder(false)); err == nil {
			t.Fatal("reconciliation accepted an outside-root stale path")
		}
		if !pathExists(path) {
			t.Fatal("outside-root path was mutated")
		}
	})

	t.Run("outside upstream protection root is rejected", func(t *testing.T) {
		basePath, repoRoot, sharedSkillsRoot := runtimeStateTestPaths(t)
		outsideRoot := filepath.Join(filepath.Dir(basePath), "outside-upstream")
		target := ManagedAssetsTarget{OwnedOverlayAssets: []OwnedOverlayAsset{{
			Key: "unsafe-upstream", Class: "upstream_runtime_snapshot", RuntimeTargets: []string{outsideRoot},
		}}}
		if err := reconcileOwnedRuntimeOutputs(runtimeOwnedStateAgent{basePath: basePath}, repoRoot, target, sharedSkillsRoot, newVerboseRecorder(false)); err == nil {
			t.Fatal("reconciliation accepted an outside-root upstream protection path")
		}
	})
}

func runtimeStateTestPaths(t *testing.T) (basePath, repoRoot, sharedSkillsRoot string) {
	t.Helper()
	repoRoot = t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	basePath = filepath.Join(home, ".config", "opencode")
	sharedSkillsRoot = filepath.Join(repoRoot, "shared", "skills")
	return basePath, repoRoot, sharedSkillsRoot
}

func writeOwnedRuntimeStateFixture(t *testing.T, basePath string, paths []string) {
	t.Helper()
	raw, err := json.MarshalIndent(ownedRuntimeState{Version: 1, Paths: paths}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	mustWriteFile(t, filepath.Join(basePath, ".gentle-ai-custom-owned-state.json"), raw)
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
