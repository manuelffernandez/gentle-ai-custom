package overlay

import (
	"fmt"
	"os"
	"path/filepath"
)

func init() {
	agentRegistry["codex"] = registeredAgent{
		agent:        &CodexAgent{},
		skillTargets: []string{"~/.codex/skills"},
	}
}

// CodexAgent installs portable repo-owned skills and applies exact local deltas
// to files already materialized by Gentle AI. It does not own Codex config,
// SDD prompts, Engram files, or the full AGENTS.md file.
type CodexAgent struct{}

func (a *CodexAgent) Name() string { return "codex" }

func (a *CodexAgent) BasePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot resolve home directory: %w", err)
	}
	return filepath.Join(home, ".codex"), nil
}

func (a *CodexAgent) BuildCommandContent(_ customCommand, body string) string {
	return body
}

func (a *CodexAgent) ApplyOverlay(repoRoot string, options applyPolicyOptions) int {
	state, basePath, err := prepareCodexOverlay(repoRoot, options)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	if err := state.applyRuntimePatches(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	if err := state.verifyRuntimePatches(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	fmt.Printf("Applied Codex instruction overlays -> %s\n", basePath)
	return 0
}

func prepareCodexOverlay(repoRoot string, options applyPolicyOptions) (*applyPolicyState, string, error) {
	policy, _, err := loadPolicy(repoRoot)
	if err != nil {
		return nil, "", err
	}
	if !policy.Codex.PreserveUnselectedSkills || policy.Codex.RenderCommands {
		return nil, "", fmt.Errorf("Codex policy must preserve unselected upstream skills and disable command rendering")
	}
	target, _, err := loadManagedAssetsTarget(repoRoot, "codex")
	if err != nil {
		return nil, "", err
	}
	if err := validateCodexAuditOnlyAssets(target); err != nil {
		return nil, "", err
	}
	if len(target.RepoOwnedCommandBodies.Entries) != 0 {
		return nil, "", fmt.Errorf("Codex target must not render command wrappers")
	}
	if err := validateRuntimeOwnershipBoundary(target); err != nil {
		return nil, "", err
	}
	sharedSkillsRoot := filepath.Join(repoRoot, filepath.FromSlash(target.RepoOwnedSharedSkills.Root))
	commandBodiesRoot := filepath.Join(repoRoot, filepath.FromSlash(target.RepoOwnedCommandBodies.Root))
	commandBodyByKey := mapCommandBodiesByKey(commandBodiesRoot, target.RepoOwnedCommandBodies.Entries)
	if err := validateCustomSources(sharedSkillsRoot, customSkillNames(target), target.RepoOwnedCommandBodies.Entries, commandBodyByKey); err != nil {
		return nil, "", err
	}
	if err := validateOwnedAssetSources(repoRoot, target.OwnedOverlayAssets); err != nil {
		return nil, "", err
	}
	basePath, err := (&CodexAgent{}).BasePath()
	if err != nil {
		return nil, "", err
	}
	state := applyPolicyState{
		managedTarget: target,
		configPath:    filepath.Join(basePath, "config.toml"),
		verbose:       options.verbose,
		recorder:      options.recorder,
	}
	if err := validateCodexInstructionFile(filepath.Join(basePath, "AGENTS.md")); err != nil {
		return nil, "", err
	}
	if err := state.prepareRuntimePatches(repoRoot); err != nil {
		return nil, "", err
	}
	return &state, basePath, nil
}

func validateCodexAuditOnlyAssets(target ManagedAssetsTarget) error {
	if len(target.RetainedUpstreamSkills) != 0 || len(target.PrunedUpstreamSkills) != 0 {
		return fmt.Errorf("Codex target must not install, prune, or claim ownership of upstream runtime assets")
	}
	for _, asset := range target.OwnedOverlayAssets {
		if !validCodexSnapshotMapping(asset) {
			return fmt.Errorf("Codex asset %q must be an audit-only upstream source without a runtime target", asset.Key)
		}
	}
	return nil
}

func validateCodexInstructionFile(path string) error {
	if _, err := os.ReadFile(path); err != nil {
		return fmt.Errorf("Codex upstream instruction file is required at %s: %w", path, err)
	}
	return nil
}
