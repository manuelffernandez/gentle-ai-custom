package overlay

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestReconcilePrunedSkillPermissionsPreservesUnrelatedRulesAndOrdersDenialsLast(t *testing.T) {
	policy := Policy{}
	policy.Skills.Prune = []string{"branch-pr", "issue-creation"}
	config := map[string]any{
		"permission": map[string]any{
			"bash": "ask",
			"skill": map[string]any{
				"*":               "allow",
				"documentation":   "ask",
				"issue-creation*": "allow",
			},
		},
		"agent": map[string]any{
			"gentle-orchestrator": map[string]any{
				"permission": map[string]any{
					"skill": map[string]any{
						"issue-creation":  "allow",
						"issue-creation*": "allow",
						"project-skill":   "ask",
					},
				},
			},
		},
	}
	agents, _ := jsonObject(config["agent"])
	state := applyPolicyState{
		policy:        policy,
		managedTarget: ManagedAssetsTarget{RepoOwnedSharedSkills: RepoOwnedSharedSkills{Allowlist: []string{"branch-pr"}}},
		configData:    config,
		agents:        agents,
		recorder:      newVerboseRecorder(false),
	}

	if err := state.reconcilePrunedSkillPermissions(); err != nil {
		t.Fatalf("reconcilePrunedSkillPermissions() error = %v", err)
	}
	if err := verifyPrunedSkillPermissions(config, agents, state.effectiveSkillDenies); err != nil {
		t.Fatalf("verifyPrunedSkillPermissions() error = %v", err)
	}
	globalPermission, _ := jsonObject(config["permission"])
	globalSkillRules, _ := jsonObject(globalPermission["skill"])
	if got := jsonString(globalSkillRules["documentation"]); got != "ask" {
		t.Fatalf("unrelated global skill permission = %q, want ask", got)
	}
	if _, denied := globalSkillRules["branch-pr"]; denied {
		t.Fatal("repo-owned branch-pr replacement received a deny")
	}
	if got := jsonString(globalPermission["bash"]); got != "ask" {
		t.Fatalf("unrelated global permission = %q, want ask", got)
	}
	agentPermission, _ := jsonObject(mustAgent(t, agents, "gentle-orchestrator")["permission"])
	agentSkillRules, _ := jsonObject(agentPermission["skill"])
	if got := jsonString(agentSkillRules["project-skill"]); got != "ask" {
		t.Fatalf("unrelated agent skill permission = %q, want ask", got)
	}

	raw, err := marshalOpenCodeConfig(config, state.effectiveSkillDenies)
	if err != nil {
		t.Fatalf("marshalOpenCodeConfig() error = %v", err)
	}
	serialized := string(raw)
	if strings.LastIndex(serialized, `"issue-creation*"`) > strings.LastIndex(serialized, `"issue-creation"`) {
		t.Fatalf("policy deny was serialized before a matching wildcard rule:\n%s", serialized)
	}
	first := serialized
	if err := state.reconcilePrunedSkillPermissions(); err != nil {
		t.Fatalf("idempotent reconcilePrunedSkillPermissions() error = %v", err)
	}
	raw, err = marshalOpenCodeConfig(config, state.effectiveSkillDenies)
	if err != nil {
		t.Fatalf("idempotent marshalOpenCodeConfig() error = %v", err)
	}
	if got := string(raw); got != first {
		t.Fatal("pruned skill permission reconciliation is not byte-idempotent")
	}
}

func TestEffectivePrunedSkillDeniesSubtractRepoOwnedReplacements(t *testing.T) {
	got := effectivePrunedSkillDenies(
		[]string{"branch-pr", "issue-creation", "hermes-ephemeral-delegation", "issue-creation"},
		[]string{"branch-pr", "chained-pr"},
	)
	want := []string{"hermes-ephemeral-delegation", "issue-creation"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("effectivePrunedSkillDenies() = %v, want %v", got, want)
	}
}

func TestApplyBlocksPrunedSkillsFromExternalRootsWithoutDeletingThem(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home, configPath := writeCompatibleRuntimeFixture(t, repoRoot)
	externalSkillPath := filepath.Join(home, ".agents", "skills", "issue-creation", "SKILL.md")
	mustWriteFile(t, externalSkillPath, []byte("external runtime skill copy\n"))
	nonPrunedExternalSkillPath := filepath.Join(home, ".claude", "skills", "code-design", "SKILL.md")
	mustWriteFile(t, nonPrunedExternalSkillPath, []byte("non-pruned external runtime skill copy\n"))

	config, err := readJSONAny(configPath)
	if err != nil {
		t.Fatal(err)
	}
	config["permission"] = map[string]any{
		"skill": map[string]any{
			"*":                  "allow",
			"branch-pr":          "allow",
			"unrelated-external": "ask",
		},
	}
	agents, _ := jsonObject(config["agent"])
	agents["custom-agent"] = map[string]any{
		"permission": map[string]any{
			"skill": map[string]any{
				"issue-creation":  "allow",
				"unrelated-agent": "ask",
			},
		},
	}
	writeJSONFixture(t, configPath, config)

	options := applyPolicyOptions{recorder: newVerboseRecorder(false), skillTargets: []string{filepath.Join(home, ".config", "opencode", "skills")}}
	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("runApplyPolicyWithOptions() code = %d, want 0", code)
	}
	if got := mustReadFile(t, externalSkillPath); got != "external runtime skill copy\n" {
		t.Fatalf("external skill copy was modified: %q", got)
	}
	if got := mustReadFile(t, nonPrunedExternalSkillPath); got != "non-pruned external runtime skill copy\n" {
		t.Fatalf("non-pruned external skill copy was modified: %q", got)
	}

	updated, err := readJSONAny(configPath)
	if err != nil {
		t.Fatal(err)
	}
	updatedAgents, _ := jsonObject(updated["agent"])
	policy, _, err := loadPolicy(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	effectiveDenies := effectivePrunedSkillDenies(policy.Skills.Prune, policyRepoOwnedSkills(t, repoRoot))
	if err := verifyPrunedSkillPermissions(updated, updatedAgents, effectiveDenies); err != nil {
		t.Fatalf("external skill leakage guard missing: %v", err)
	}
	globalPermission, _ := jsonObject(updated["permission"])
	globalSkillRules, _ := jsonObject(globalPermission["skill"])
	if got := jsonString(globalSkillRules["unrelated-external"]); got != "ask" {
		t.Fatalf("unrelated global skill permission = %q, want ask", got)
	}
	if got := jsonString(globalSkillRules["branch-pr"]); got != "allow" {
		t.Fatalf("repo-owned branch-pr replacement permission = %q, want allow", got)
	}
	branchSource := filepath.Join(repoRoot, "shared", "skills", "branch-pr", "SKILL.md")
	branchTarget := filepath.Join(home, ".config", "opencode", "skills", "branch-pr", "SKILL.md")
	if err := installSkill(filepath.Join(home, ".config", "opencode"), "branch-pr", branchSource, "opencode", newVerboseRecorder(false)); err != nil {
		t.Fatalf("install sanitized branch-pr replacement: %v", err)
	}
	if got, want := mustReadFile(t, branchTarget), mustReadFile(t, branchSource); got != want {
		t.Fatal("sanitized branch-pr replacement was not installed after pruning")
	}
	customPermission, _ := jsonObject(mustAgent(t, updatedAgents, "custom-agent")["permission"])
	customSkillRules, _ := jsonObject(customPermission["skill"])
	if got := jsonString(customSkillRules["unrelated-agent"]); got != "ask" {
		t.Fatalf("unrelated agent skill permission = %q, want ask", got)
	}
	first := mustReadFile(t, configPath)
	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("idempotent runApplyPolicyWithOptions() code = %d, want 0", code)
	}
	if got := mustReadFile(t, configPath); got != first {
		t.Fatal("external skill leakage guard is not byte-idempotent")
	}
}

func TestApplyNormalizesExistingEffectiveDenyAfterWildcardAllow(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home, configPath := writeCompatibleRuntimeFixture(t, repoRoot)
	options := applyPolicyOptions{recorder: newVerboseRecorder(false), skillTargets: []string{filepath.Join(home, ".config", "opencode", "skills")}}
	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("initial runApplyPolicyWithOptions() code = %d, want 0", code)
	}

	config, err := readJSONAny(configPath)
	if err != nil {
		t.Fatal(err)
	}
	permission, _ := jsonObject(config["permission"])
	permission["skill"] = map[string]any{
		"hermes-ephemeral-delegation": "deny",
		"issue-creation":              "deny",
		"issue-creation*":             "allow",
	}
	writeJSONFixture(t, configPath, config)
	before := mustReadFile(t, configPath)
	assertExactDenyPrecedesWildcard(t, before, "issue-creation")

	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("normalizing runApplyPolicyWithOptions() code = %d, want 0", code)
	}
	normalized := mustReadFile(t, configPath)
	assertWildcardPrecedesExactDeny(t, normalized, "issue-creation")
	if normalized == before {
		t.Fatal("apply did not rewrite the stale effective-deny order")
	}

	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("idempotent runApplyPolicyWithOptions() code = %d, want 0", code)
	}
	if got := mustReadFile(t, configPath); got != normalized {
		t.Fatal("effective deny order normalization is not byte-idempotent")
	}
}

func assertExactDenyPrecedesWildcard(t *testing.T, config, skill string) {
	t.Helper()
	exact := strings.LastIndex(config, `"`+skill+`"`)
	wildcard := strings.LastIndex(config, `"`+skill+`*"`)
	if exact < 0 || wildcard < 0 || exact > wildcard {
		t.Fatalf("expected exact deny before matching wildcard allow for %q:\n%s", skill, config)
	}
}

func assertWildcardPrecedesExactDeny(t *testing.T, config, skill string) {
	t.Helper()
	exact := strings.LastIndex(config, `"`+skill+`"`)
	wildcard := strings.LastIndex(config, `"`+skill+`*"`)
	if exact < 0 || wildcard < 0 || wildcard > exact {
		t.Fatalf("expected matching wildcard allow before exact deny for %q:\n%s", skill, config)
	}
}

func policyRepoOwnedSkills(t *testing.T, repoRoot string) []string {
	t.Helper()
	target, _, err := loadManagedAssetsTarget(repoRoot, "opencode")
	if err != nil {
		t.Fatal(err)
	}
	return target.RepoOwnedSharedSkills.Allowlist
}
