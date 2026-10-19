package overlay

import (
	"fmt"
	"slices"
)

// Target-specific surfaces may differ, but every registered target must have
// an explicit route for the overlay's shared behavior. Adding a target without
// classifying these surfaces is an error, not an implicit OpenCode fallback.
type parityCapabilities struct {
	commandSurface string
	configSurface  string
	permissionRule string
	runtimePatches []string
	ownedAssets    []string
	commandKeys    []string
}

var targetParityCapabilities = map[string]parityCapabilities{
	"opencode": {
		commandSurface: "generated wrappers",
		configSurface:  "OpenCode config overrides",
		permissionRule: "OpenCode skill denies",
		runtimePatches: []string{"judgment-retrospective-handoff"},
		ownedAssets:    []string{"opencode-agents"},
		commandKeys:    []string{"commit-plan", "commit-apply", "commit-fast", "pr-create", "pr-regenerate"},
	},
	"codex": {
		commandSurface: "shared skill entrypoints; no wrapper renderer",
		configSurface:  "upstream-owned config; exact ODD instruction row deltas",
		permissionRule: "upstream-owned permissions; selected rejected-skill pruning",
		runtimePatches: []string{"codex-local-instructions", "judgment-retrospective-handoff"},
	},
}

var expectedSharedSkills = []string{
	"branch-pr", "chained-pr", "code-design", "commit-planner",
	"judgment-retrospective", "notion-project-knowledge", "package-security", "pr-finalizer",
}

func validateTargetParity(manifest ManagedAssetsManifest, registered []string) error {
	if len(registered) == 0 || len(manifest.Targets) != len(registered) || len(targetParityCapabilities) != len(registered) {
		return fmt.Errorf("target parity requires one manifest target and capability declaration for every registered agent (registered=%v)", registered)
	}
	opencode, ok := manifest.Targets["opencode"]
	if !ok {
		return fmt.Errorf("target parity requires an OpenCode command-source target")
	}
	if len(opencode.RepoOwnedCommandBodies.Entries) == 0 {
		return fmt.Errorf("OpenCode command wrappers are missing from the parity source")
	}
	if !sameSkillNames(opencode.RepoOwnedSharedSkills.Allowlist, expectedSharedSkills) {
		return fmt.Errorf("target parity shared skill installation differs from the canonical overlay allowlist for %q", "opencode")
	}
	owners := map[string]bool{}
	for _, command := range opencode.RepoOwnedCommandBodies.Entries {
		owners[command.OwnerSkill] = true
	}
	for _, name := range registered {
		capability, declared := targetParityCapabilities[name]
		if !declared || capability.commandSurface == "" || capability.configSurface == "" || capability.permissionRule == "" {
			return fmt.Errorf("target parity has no explicit host-specific surface declaration for %q", name)
		}
		target, exists := manifest.Targets[name]
		if !exists {
			return fmt.Errorf("target parity manifest does not cover registered agent %q", name)
		}
		patchKeys := make([]string, 0, len(target.RuntimePatches))
		for _, patch := range target.RuntimePatches {
			patchKeys = append(patchKeys, patch.Key)
		}
		if err := checkParityKeys(name, "runtime patch", patchKeys, capability.runtimePatches); err != nil {
			return err
		}
		ownedKeys := []string{}
		for _, asset := range target.OwnedOverlayAssets {
			if asset.Class == "repo_owned_runtime" {
				ownedKeys = append(ownedKeys, asset.Key)
			}
		}
		if err := checkParityKeys(name, "owned runtime asset", ownedKeys, capability.ownedAssets); err != nil {
			return err
		}
		commandKeys := []string{}
		for _, command := range target.RepoOwnedCommandBodies.Entries {
			commandKeys = append(commandKeys, command.Key)
		}
		if err := checkParityKeys(name, "command", commandKeys, capability.commandKeys); err != nil {
			return err
		}
		if target.RepoOwnedSharedSkills.Root != opencode.RepoOwnedSharedSkills.Root ||
			target.RepoOwnedSharedSkills.InstallMode != opencode.RepoOwnedSharedSkills.InstallMode ||
			!sameSkillNames(target.RepoOwnedSharedSkills.Allowlist, expectedSharedSkills) {
			return fmt.Errorf("target parity shared skill installation differs for %q", name)
		}
		if name == "opencode" {
			continue
		}
		if len(target.RepoOwnedCommandBodies.Entries) != 0 {
			return fmt.Errorf("target parity %q must declare its command rendering strategy", name)
		}
		for owner := range owners {
			if !slices.Contains(target.RepoOwnedSharedSkills.Allowlist, owner) {
				return fmt.Errorf("target parity %q lacks shared entrypoint for command owner %q", name, owner)
			}
		}
	}
	return nil
}

func checkParityKeys(target, kind string, actual, declared []string) error {
	seen := map[string]bool{}
	for _, key := range actual {
		if seen[key] || !slices.Contains(declared, key) {
			return fmt.Errorf("target parity %q has unclassified or duplicated %s %q", target, kind, key)
		}
		seen[key] = true
	}
	for _, key := range declared {
		if !seen[key] {
			return fmt.Errorf("target parity %q is missing declared %s %q", target, kind, key)
		}
	}
	return nil
}

func sameSkillNames(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	left, right := slices.Clone(a), slices.Clone(b)
	slices.Sort(left)
	slices.Sort(right)
	return slices.Equal(left, right)
}
