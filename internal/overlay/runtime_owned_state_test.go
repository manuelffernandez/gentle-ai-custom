package overlay

import (
	"encoding/json"
	"path/filepath"
	"reflect"
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

func TestCollectOwnedRuntimePathsFiltersBySyncMode(t *testing.T) {
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
		name       string
		asset      OwnedOverlayAsset
		wantPaths  []string // expected tracked paths; nil means none
	}{
		{
			name: "file-mode asset with no RuntimeSyncMode is NOT tracked",
			asset: OwnedOverlayAsset{
				Key:             "opencode-agents",
				Class:           "repo_owned_runtime",
				Kind:            "agent_instruction_file",
				RepoOwnedPath:  "overlay/gentle-ai/assets/owned/opencode/AGENTS.md",
				RuntimeSyncMode: "",
				RuntimeTargets:  []string{"~/.config/opencode/AGENTS.md"},
			},
			wantPaths: nil,
		},
		{
			name: "merge-mode directory asset IS tracked",
			asset: OwnedOverlayAsset{
				Key:             "sdd-commands",
				Class:           "repo_owned_runtime",
				Kind:            "command_directory",
				RepoOwnedPath:  "overlay/gentle-ai/assets/owned/opencode/commands",
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
				RepoOwnedPath:  "overlay/gentle-ai/assets/owned/opencode/skills/test-prune-skill",
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
			if tt.wantPaths == nil {
				if len(paths) != 0 {
					t.Errorf("expected no tracked paths, got %v", paths)
				}
				return
			}
			if !reflect.DeepEqual(paths, tt.wantPaths) {
				t.Errorf("tracked paths = %v, want %v", paths, tt.wantPaths)
			}
		})
	}
}

func TestReconcileOwnedRuntimeOutputsStalePathsOnlyForSyncModeAssets(t *testing.T) {
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

	// The saved state must contain only paths for mode-carrying assets (prune/merge).
	// No-mode assets (e.g. AGENTS.md) must never appear in state.
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

func TestCollectOwnedRuntimePathsNoModeAssetIsNotTracked(t *testing.T) {
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
				RuntimeSyncMode: "", // no sync mode → must not enter the deletion surface
				RuntimeTargets:  []string{"~/.config/opencode/AGENTS.md"},
			},
		},
	}

	paths, err := collectOwnedRuntimePaths(basePath, repoRoot, target, sharedSkillsRoot)
	if err != nil {
		t.Fatalf("collectOwnedRuntimePaths() error = %v", err)
	}
	for _, p := range paths {
		if p == noModeTarget {
			t.Errorf("no-mode asset path %q must not appear in tracked runtime paths", noModeTarget)
		}
	}
}

