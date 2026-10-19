package overlay

import (
	"strings"
	"testing"
)

func TestApplyRefusesUnmappedRegisteredAgentBeforeWrites(t *testing.T) {
	repo := repoRootForTest(t)
	home, _ := writeCompatibleRuntimeFixture(t, repo)
	registry := make(map[string]registeredAgent, len(agentRegistry)+1)
	for name, entry := range agentRegistry {
		registry[name] = entry
	}
	registry["future-agent"] = registeredAgent{agent: fakeAgent{name: "future-agent"}, skillTargets: []string{"~/.future/skills"}}
	setTestAgentRegistry(t, registry)
	before := snapshotTree(t, home)
	if code := RunApplyCustom(repo, []string{"opencode"}); code == 0 {
		t.Fatal("apply accepted an unclassified registered agent")
	}
	if after := snapshotTree(t, home); !equalStringMaps(before, after) {
		t.Fatal("apply wrote runtime files before target parity refusal")
	}
}

func TestTargetParityCoversEveryRegisteredAgent(t *testing.T) {
	manifest, _, err := loadManagedAssetsManifest(repoRootForTest(t))
	if err != nil {
		t.Fatal(err)
	}
	if err := validateTargetParity(manifest, registeredAgentNames()); err != nil {
		t.Fatalf("current targets lack declared parity: %v", err)
	}
	if err := validateTargetParity(manifest, append(registeredAgentNames(), "future-agent")); err == nil || !strings.Contains(err.Error(), "future-agent") {
		t.Fatalf("newly registered agent bypassed parity gate: %v", err)
	}
}

func TestTargetParityRejectsOmittedSharedCustomization(t *testing.T) {
	manifest, _, err := loadManagedAssetsManifest(repoRootForTest(t))
	if err != nil {
		t.Fatal(err)
	}
	codex := manifest.Targets["codex"]
	codex.RepoOwnedSharedSkills.Allowlist = codex.RepoOwnedSharedSkills.Allowlist[1:]
	manifest.Targets["codex"] = codex
	if err := validateTargetParity(manifest, registeredAgentNames()); err == nil || !strings.Contains(err.Error(), "codex") {
		t.Fatalf("unequal skill installation was accepted: %v", err)
	}
}

func TestTargetParityRejectsMissingJudgmentDayCustomization(t *testing.T) {
	manifest, _, err := loadManagedAssetsManifest(repoRootForTest(t))
	if err != nil {
		t.Fatal(err)
	}
	codex := manifest.Targets["codex"]
	codex.RuntimePatches = codex.RuntimePatches[:1]
	manifest.Targets["codex"] = codex
	if err := validateTargetParity(manifest, registeredAgentNames()); err == nil || !strings.Contains(err.Error(), "judgment-retrospective-handoff") {
		t.Fatalf("missing shared runtime hook was accepted: %v", err)
	}
}

func TestTargetParityRequiresSkillEntryForHostSpecificCommands(t *testing.T) {
	manifest, _, err := loadManagedAssetsManifest(repoRootForTest(t))
	if err != nil {
		t.Fatal(err)
	}
	opencode := manifest.Targets["opencode"]
	changed := false
	for i := range opencode.RepoOwnedCommandBodies.Entries {
		if opencode.RepoOwnedCommandBodies.Entries[i].Key == "pr-create" {
			opencode.RepoOwnedCommandBodies.Entries[i].OwnerSkill = "unmapped-workflow"
			changed = true
		}
	}
	if !changed {
		t.Fatal("command owner fixture is missing pr-create")
	}
	manifest.Targets["opencode"] = opencode
	if err := validateTargetParity(manifest, registeredAgentNames()); err == nil || !strings.Contains(err.Error(), "unmapped-workflow") {
		t.Fatalf("unmapped command owner escaped cross-target entrypoint check: %v", err)
	}
}

func TestTargetParityRejectsUnclassifiedRuntimeCustomizations(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*ManagedAssetsManifest)
		want   string
	}{
		{"new Codex patch", func(m *ManagedAssetsManifest) {
			target := m.Targets["codex"]
			target.RuntimePatches = append(target.RuntimePatches, RuntimePatchRef{Key: "new-shared-rule", PatchPath: "overlay/gentle-ai/patches/codex/new-rule.json"})
			m.Targets["codex"] = target
		}, "new-shared-rule"},
		{"missing Codex instructions", func(m *ManagedAssetsManifest) {
			target := m.Targets["codex"]
			target.RuntimePatches = target.RuntimePatches[1:]
			m.Targets["codex"] = target
		}, "codex-local-instructions"},
		{"new OpenCode owned asset", func(m *ManagedAssetsManifest) {
			target := m.Targets["opencode"]
			target.OwnedOverlayAssets = append(target.OwnedOverlayAssets, OwnedOverlayAsset{Key: "new-runtime-asset", Class: "repo_owned_runtime", RepoOwnedPath: "overlay/gentle-ai/assets/owned/opencode/new.md"})
			m.Targets["opencode"] = target
		}, "new-runtime-asset"},
		{"new command with known owner", func(m *ManagedAssetsManifest) {
			target := m.Targets["opencode"]
			target.RepoOwnedCommandBodies.Entries = append(target.RepoOwnedCommandBodies.Entries, RepoOwnedCommandEntry{Key: "new-commit-command", OwnerSkill: "commit-planner"})
			m.Targets["opencode"] = target
		}, "new-commit-command"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			manifest, _, err := loadManagedAssetsManifest(repoRootForTest(t))
			if err != nil {
				t.Fatal(err)
			}
			tc.change(&manifest)
			if err := validateTargetParity(manifest, registeredAgentNames()); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("unclassified customization was accepted: %v", err)
			}
		})
	}
}
