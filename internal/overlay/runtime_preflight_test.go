package overlay

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const opaqueUpstreamOrchestratorFixture = "opaque upstream orchestrator materialization\n"

var opaquePluginFixture = []byte{0x00, 0xff, 0x10, 'o', 'p', 'a', 'q', 'u', 'e'}

func TestApplyPreservesUpstreamRuntimeAndReusesBasePromptsForProfiles(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home, configPath := writeCompatibleRuntimeFixture(t, repoRoot)
	runtimePatchBaselines := snapshotRuntimePatchTargets(t, repoRoot, home)
	orchestratorBefore := opaqueUpstreamOrchestratorFixture
	pluginPath := filepath.Join(home, ".config", "opencode", "plugins", "review-result-artifacts.ts")
	phasePath := filepath.Join(home, ".config", "opencode", "prompts", "sdd", "sdd-init.md")
	contractPath := filepath.Join(home, ".config", "opencode", "skills", "_shared", "review-ledger-contract.md")
	preserved := map[string]testFileSnapshot{
		pluginPath:   snapshotTestFile(t, pluginPath),
		phasePath:    snapshotTestFile(t, phasePath),
		contractPath: snapshotTestFile(t, contractPath),
	}

	options := applyPolicyOptions{
		recorder:     newVerboseRecorder(false),
		skillTargets: []string{filepath.Join(home, ".config", "opencode", "skills")},
	}
	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("runApplyPolicyWithOptions() code = %d, want 0", code)
	}

	agents := readFixtureAgents(t, configPath)
	if got := jsonString(mustAgent(t, agents, "gentle-orchestrator")["prompt"]); got != orchestratorBefore {
		t.Fatalf("base orchestrator prompt changed\nwant: %q\ngot:  %q", orchestratorBefore, got)
	}
	for path, before := range preserved {
		assertTestFileSnapshot(t, path, before)
	}
	for _, key := range []string{"gentle-orchestrator", "sdd-orchestrator-cheap"} {
		if got := jsonString(mustAgent(t, agents, key)["prompt"]); got != orchestratorBefore {
			t.Fatalf("%s prompt does not reuse the rendered upstream base prompt", key)
		}
	}
	for _, phase := range mustLoadPolicyPhases(t, repoRoot) {
		basePrompt := jsonString(mustAgent(t, agents, phase)["prompt"])
		profilePrompt := jsonString(mustAgent(t, agents, phase+"-cheap")["prompt"])
		if profilePrompt != basePrompt {
			t.Fatalf("profile phase %s prompt = %q, want upstream base prompt %q", phase, profilePrompt, basePrompt)
		}
	}
	if got := jsonString(mustAgent(t, agents, "sdd-orchestrator-cheap")["model"]); got != "profile/orchestrator" {
		t.Fatalf("profile model = %q, want profile/orchestrator", got)
	}
	if !reflect.DeepEqual(mustAgent(t, agents, "sdd-init-cheap")["tools"], mustAgent(t, agents, "sdd-init")["tools"]) {
		t.Fatal("new managed phase agent did not preserve upstream base behavior")
	}
	assertNamedProfileTaskPermissions(t, agents, mustLoadPolicyPhases(t, repoRoot), "cheap")
	if strings.Contains(mustReadFile(t, configPath), ".overlay.md") {
		t.Fatal("custom apply left an obsolete .overlay.md reference")
	}
	rawSnapshot := mustReadFile(t, filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "prompts", "orchestrators", "gentle-orchestrator.md"))
	if jsonString(mustAgent(t, agents, "gentle-orchestrator")["prompt"]) == rawSnapshot {
		t.Fatal("custom apply replaced the upstream-materialized orchestrator with the raw audit snapshot")
	}
	obsoletePath := filepath.Join(home, ".config", "opencode", "prompts", "sdd", "orchestrators", "gentle-orchestrator.overlay.md")
	if pathExists(obsoletePath) {
		t.Fatal("custom apply left the obsolete managed .overlay.md file")
	}

	judgmentSkill := filepath.Join(home, ".config", "opencode", "skills", "judgment-day", "SKILL.md")
	judgmentPrompt := filepath.Join(home, ".config", "opencode", "skills", "judgment-day", "references", "prompts-and-formats.md")
	researchSkill := filepath.Join(home, ".config", "opencode", "skills", "sdd-research", "SKILL.md")
	for path, marker := range map[string]string{
		judgmentSkill:  "invoke the repo-owned `judgment-retrospective` skill",
		judgmentPrompt: "## Terminal Retrospective Handoff",
		researchSkill:  "Inline launcher instructions cannot override this output-only boundary.",
	} {
		if count := strings.Count(mustReadFile(t, path), marker); count != 1 {
			t.Fatalf("patched %s marker count = %d, want 1", path, count)
		}
	}
	firstSkill := mustReadFile(t, judgmentSkill)
	firstPrompt := mustReadFile(t, judgmentPrompt)
	firstResearch := mustReadFile(t, researchSkill)
	firstConfig := mustReadFile(t, configPath)
	assertRuntimePatchIsExactDelta(t, repoRoot, home, runtimePatchBaselines)
	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("idempotent run code = %d, want 0", code)
	}
	if mustReadFile(t, judgmentSkill) != firstSkill || mustReadFile(t, judgmentPrompt) != firstPrompt || mustReadFile(t, researchSkill) != firstResearch {
		t.Fatal("runtime patches are not byte-idempotent")
	}
	if mustReadFile(t, configPath) != firstConfig {
		t.Fatal("profile reconciliation is not byte-idempotent")
	}
}

func TestReconcileNamedProfileCreatesUpstreamTaskPermissionContract(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home, configPath := writeCompatibleRuntimeFixture(t, repoRoot)
	data, err := readJSONAny(configPath)
	if err != nil {
		t.Fatal(err)
	}
	agents, _ := jsonObject(data["agent"])
	delete(agents, "sdd-orchestrator-cheap")
	writeJSONFixture(t, configPath, data)
	options := applyPolicyOptions{
		recorder:     newVerboseRecorder(false),
		skillTargets: []string{filepath.Join(home, ".config", "opencode", "skills")},
	}
	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("runApplyPolicyWithOptions() code = %d, want 0", code)
	}
	agents = readFixtureAgents(t, configPath)
	assertNamedProfileTaskPermissions(t, agents, mustLoadPolicyPhases(t, repoRoot), "cheap")
	first := mustReadFile(t, configPath)
	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("idempotent run code = %d, want 0", code)
	}
	if got := mustReadFile(t, configPath); got != first {
		t.Fatal("created named orchestrator permission reconciliation is not idempotent")
	}
}

func TestApplyReconcilesGlobalActorAssignments(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home, configPath := writeCompatibleRuntimeFixture(t, repoRoot)
	localConfigPath := filepath.Join(home, ".config", "gentle-ai-custom", "opencode-local-config.json")
	localConfig, err := readJSONAny(localConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	defaultProfile := localConfig["default_profile"].(map[string]any)
	defaultProfile["judgment_day"] = map[string]any{
		"jd-judge-a": map[string]any{"model": "custom/judge"},
	}
	defaultProfile["review"] = map[string]any{
		"review-risk":        map[string]any{"inherit": true},
		"review-readability": map[string]any{"inherit": true},
		"review-validator":   map[string]any{"inherit": true},
	}
	writeJSONFixture(t, localConfigPath, localConfig)
	runtimeConfig, err := readJSONAny(configPath)
	if err != nil {
		t.Fatal(err)
	}
	agents := runtimeConfig["agent"].(map[string]any)
	mustAgent(t, agents, "jd-judge-a")["model"] = "stale/model"
	mustAgent(t, agents, "jd-judge-a")["variant"] = "stale"
	for _, key := range []string{"review-risk", "review-readability", "review-validator"} {
		mustAgent(t, agents, key)["model"] = "stale/model"
		mustAgent(t, agents, key)["variant"] = "stale"
	}
	writeJSONFixture(t, configPath, runtimeConfig)

	options := applyPolicyOptions{recorder: newVerboseRecorder(false), skillTargets: []string{filepath.Join(home, ".config", "opencode", "skills")}}
	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("runApplyPolicyWithOptions() code = %d, want 0", code)
	}
	agents = readFixtureAgents(t, configPath)
	judge := mustAgent(t, agents, "jd-judge-a")
	if jsonString(judge["model"]) != "custom/judge" {
		t.Fatalf("explicit model = %q, want custom/judge", jsonString(judge["model"]))
	}
	if _, exists := judge["variant"]; exists {
		t.Fatal("explicit model with omitted variant left a stale variant")
	}
	for _, key := range []string{"review-risk", "review-readability", "review-validator"} {
		actor := mustAgent(t, agents, key)
		if _, exists := actor["model"]; exists {
			t.Fatalf("inherited actor %s retained model", key)
		}
		if _, exists := actor["variant"]; exists {
			t.Fatalf("inherited actor %s retained variant", key)
		}
	}
}

func TestApplyReconcilesNamedJudgmentDayActorsAndPreservesUnmanagedAgents(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home, configPath := writeCompatibleRuntimeFixture(t, repoRoot)
	localConfigPath := filepath.Join(home, ".config", "gentle-ai-custom", "opencode-local-config.json")
	localConfig, err := readJSONAny(localConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	profile := localConfig["profiles"].([]any)[0].(map[string]any)
	profile["judgment_day"] = map[string]any{
		"jd-judge-a": map[string]any{"model": "profile/judge"},
	}
	writeJSONFixture(t, localConfigPath, localConfig)
	runtimeConfig, err := readJSONAny(configPath)
	if err != nil {
		t.Fatal(err)
	}
	agents := runtimeConfig["agent"].(map[string]any)
	baseJudge := mustAgent(t, agents, "jd-judge-a")
	baseJudge["hidden"] = true
	baseJudge["tools"] = map[string]any{"read": true, "task": false}
	agents["jd-judge-b-cheap"] = map[string]any{"stale": true}
	agents["custom-agent"] = map[string]any{"sentinel": "preserve"}
	agents["sdd-orchestrator-unmanaged"] = map[string]any{"sentinel": "profile-preserve"}
	agents["jd-judge-a-unmanaged"] = map[string]any{"sentinel": "jd-preserve"}
	writeJSONFixture(t, configPath, runtimeConfig)

	options := applyPolicyOptions{recorder: newVerboseRecorder(false), skillTargets: []string{filepath.Join(home, ".config", "opencode", "skills")}}
	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("runApplyPolicyWithOptions() code = %d, want 0", code)
	}
	agents = readFixtureAgents(t, configPath)
	namedJudge := mustAgent(t, agents, "jd-judge-a-cheap")
	if jsonString(namedJudge["model"]) != "profile/judge" {
		t.Fatalf("named judge model = %q, want profile/judge", jsonString(namedJudge["model"]))
	}
	if jsonString(namedJudge["prompt"]) != jsonString(baseJudge["prompt"]) || !reflect.DeepEqual(namedJudge["tools"], baseJudge["tools"]) || namedJudge["hidden"] != true {
		t.Fatal("named judge did not preserve upstream prompt/tools/hidden structure")
	}
	if _, exists := namedJudge["variant"]; exists {
		t.Fatal("named explicit model with omitted variant retained a variant")
	}
	if _, exists := agents["jd-judge-b-cheap"]; exists {
		t.Fatal("stale managed named JD suffix was not removed")
	}
	assertNamedProfileTaskPermissionsWithJD(t, agents, mustLoadPolicyPhases(t, repoRoot), "cheap", map[string]bool{"jd-judge-a": true})
	for key, field := range map[string]string{"custom-agent": "sentinel", "sdd-orchestrator-unmanaged": "sentinel", "jd-judge-a-unmanaged": "sentinel"} {
		if jsonString(mustAgent(t, agents, key)[field]) == "" {
			t.Fatalf("unmanaged agent %s was not preserved", key)
		}
	}
	first := mustReadFile(t, configPath)
	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("idempotent run code = %d, want 0", code)
	}
	if got := mustReadFile(t, configPath); got != first {
		t.Fatal("named Judgment Day reconciliation is not byte-idempotent")
	}
}

func TestApplyNamedJudgmentDayOmissionDelegatesToGlobal(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home, configPath := writeCompatibleRuntimeFixture(t, repoRoot)
	options := applyPolicyOptions{recorder: newVerboseRecorder(false), skillTargets: []string{filepath.Join(home, ".config", "opencode", "skills")}}
	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("runApplyPolicyWithOptions() code = %d, want 0", code)
	}
	agents := readFixtureAgents(t, configPath)
	assertNamedProfileTaskPermissionsWithJD(t, agents, mustLoadPolicyPhases(t, repoRoot), "cheap", nil)
	for _, key := range judgmentDayActorKeys {
		if _, exists := agents[key+"-cheap"]; exists {
			t.Fatalf("omitted named JD actor %s unexpectedly exists", key+"-cheap")
		}
	}
}

func TestApplyMissingOmissionDelegatedJudgmentDayActorFailsBeforeWrites(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home, configPath := writeCompatibleRuntimeFixture(t, repoRoot)
	runtimeConfig, err := readJSONAny(configPath)
	if err != nil {
		t.Fatal(err)
	}
	delete(runtimeConfig["agent"].(map[string]any), "jd-judge-b")
	writeJSONFixture(t, configPath, runtimeConfig)
	before := snapshotTree(t, home)
	options := applyPolicyOptions{recorder: newVerboseRecorder(false), skillTargets: []string{filepath.Join(home, ".config", "opencode", "skills")}}
	if code := runApplyPolicyWithOptions(repoRoot, options); code == 0 {
		t.Fatal("runApplyPolicyWithOptions() code = 0, want missing omission-delegated Judgment Day actor failure")
	}
	after := snapshotTree(t, home)
	if !reflect.DeepEqual(after, before) {
		t.Fatal("runtime changed after missing omission-delegated Judgment Day actor preflight failure")
	}
}

func TestApplyMissingReferencedActorFailsBeforeWrites(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home, configPath := writeCompatibleRuntimeFixture(t, repoRoot)
	localConfigPath := filepath.Join(home, ".config", "gentle-ai-custom", "opencode-local-config.json")
	localConfig, err := readJSONAny(localConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	defaultProfile := localConfig["default_profile"].(map[string]any)
	defaultProfile["review"] = map[string]any{"review-risk": map[string]any{"inherit": true}}
	writeJSONFixture(t, localConfigPath, localConfig)
	runtimeConfig, err := readJSONAny(configPath)
	if err != nil {
		t.Fatal(err)
	}
	delete(runtimeConfig["agent"].(map[string]any), "review-risk")
	writeJSONFixture(t, configPath, runtimeConfig)
	before := snapshotTree(t, home)
	options := applyPolicyOptions{recorder: newVerboseRecorder(false), skillTargets: []string{filepath.Join(home, ".config", "opencode", "skills")}}
	if code := runApplyPolicyWithOptions(repoRoot, options); code == 0 {
		t.Fatal("runApplyPolicyWithOptions() code = 0, want missing referenced actor failure")
	}
	after := snapshotTree(t, home)
	if !reflect.DeepEqual(after, before) {
		t.Fatal("runtime changed after missing referenced actor preflight failure")
	}
}

func TestApplyPreservesAbsentUnreferencedActor(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home, configPath := writeCompatibleRuntimeFixture(t, repoRoot)
	runtimeConfig, err := readJSONAny(configPath)
	if err != nil {
		t.Fatal(err)
	}
	delete(runtimeConfig["agent"].(map[string]any), "review-risk")
	writeJSONFixture(t, configPath, runtimeConfig)
	options := applyPolicyOptions{recorder: newVerboseRecorder(false), skillTargets: []string{filepath.Join(home, ".config", "opencode", "skills")}}
	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("runApplyPolicyWithOptions() code = %d, want 0 for absent unreferenced actor", code)
	}
	if _, exists := readFixtureAgents(t, configPath)["review-risk"]; exists {
		t.Fatal("custom apply fabricated an absent unreferenced review actor")
	}
}

func TestApplyDoesNotRequireUpstreamPlugin(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home, _ := writeCompatibleRuntimeFixture(t, repoRoot)
	pluginPath := filepath.Join(home, ".config", "opencode", "plugins", "review-result-artifacts.ts")
	if err := os.Remove(pluginPath); err != nil {
		t.Fatal(err)
	}
	options := applyPolicyOptions{recorder: newVerboseRecorder(false), skillTargets: []string{filepath.Join(home, ".config", "opencode", "skills")}}
	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("apply without upstream plugin code = %d, want 0", code)
	}
	if pathExists(pluginPath) {
		t.Fatal("custom apply created an upstream-owned plugin")
	}
}

func TestPrepareRuntimePatchesRejectsMixedPatchedAndUnpatchedAnchors(t *testing.T) {
	repoRoot := t.TempDir()
	home := t.TempDir()
	configPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	targetPath := filepath.Join(home, ".config", "opencode", "skills", "example.md")
	mustWriteFile(t, targetPath, []byte("anchor patched\nanchor\n"))
	patchPath := filepath.Join(repoRoot, "patch.json")
	writeJSONFixture(t, patchPath, map[string]any{
		"version": 1,
		"targets": []any{map[string]any{
			"path": "skills/example.md",
			"replacements": []any{map[string]any{
				"before": "anchor",
				"after":  "anchor patched",
			}},
		}},
	})
	state := applyPolicyState{
		configPath: configPath,
		managedTarget: ManagedAssetsTarget{RuntimePatches: []RuntimePatchRef{
			{Key: "example", PatchPath: "patch.json"},
		}},
	}
	if err := state.prepareRuntimePatches(repoRoot); err == nil {
		t.Fatal("prepareRuntimePatches() error = nil, want mixed anchor rejection")
	}
}

func snapshotRuntimePatchTargets(t *testing.T, repoRoot, home string) map[string]string {
	t.Helper()
	target, _, err := loadManagedAssetsTarget(repoRoot, "opencode")
	if err != nil {
		t.Fatal(err)
	}
	baselines := map[string]string{}
	for _, patch := range target.RuntimePatches {
		var definition runtimePatchDefinition
		if err := readJSONFile(filepath.Join(repoRoot, filepath.FromSlash(patch.PatchPath)), &definition); err != nil {
			t.Fatalf("load runtime patch %s: %v", patch.Key, err)
		}
		for _, patchTarget := range definition.Targets {
			if _, exists := baselines[patchTarget.Path]; !exists {
				baselines[patchTarget.Path] = mustReadFile(t, filepath.Join(home, ".config", "opencode", filepath.FromSlash(patchTarget.Path)))
			}
		}
	}
	return baselines
}

func assertRuntimePatchIsExactDelta(t *testing.T, repoRoot, home string, baselines map[string]string) {
	t.Helper()
	target, _, err := loadManagedAssetsTarget(repoRoot, "opencode")
	if err != nil {
		t.Fatal(err)
	}
	reverted := map[string]string{}
	for path := range baselines {
		reverted[path] = mustReadFile(t, filepath.Join(home, ".config", "opencode", filepath.FromSlash(path)))
	}
	for patchIndex := len(target.RuntimePatches) - 1; patchIndex >= 0; patchIndex-- {
		patch := target.RuntimePatches[patchIndex]
		var definition runtimePatchDefinition
		if err := readJSONFile(filepath.Join(repoRoot, filepath.FromSlash(patch.PatchPath)), &definition); err != nil {
			t.Fatalf("load runtime patch %s: %v", patch.Key, err)
		}
		for _, patchTarget := range definition.Targets {
			content, exists := reverted[patchTarget.Path]
			if !exists {
				t.Fatalf("runtime patch %s target %s has no pre-apply baseline", patch.Key, patchTarget.Path)
			}
			for index := len(patchTarget.Replacements) - 1; index >= 0; index-- {
				replacement := patchTarget.Replacements[index]
				if strings.Count(content, replacement.After) != 1 {
					t.Fatalf("runtime patch %s target %s does not contain exact declared delta %d", patch.Key, patchTarget.Path, index+1)
				}
				content = strings.Replace(content, replacement.After, replacement.Before, 1)
			}
			reverted[patchTarget.Path] = content
		}
	}
	for path, baseline := range baselines {
		if reverted[path] != baseline {
			t.Fatalf("runtime patch target %s differs from its pre-apply baseline beyond the declared replacements", path)
		}
	}
}

type runtimeBreakFixture struct {
	name         string
	breakRuntime func(t *testing.T, home string)
}

type testFileSnapshot struct {
	content string
	mode    os.FileMode
}

func snapshotTestFile(t *testing.T, path string) testFileSnapshot {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return testFileSnapshot{content: mustReadFile(t, path), mode: info.Mode()}
}

func assertTestFileSnapshot(t *testing.T, path string, want testFileSnapshot) {
	t.Helper()
	got := snapshotTestFile(t, path)
	if got != want {
		t.Fatalf("upstream-owned file changed across custom apply: %s", path)
	}
}

func TestApplyPreflightFailureLeavesRuntimeUnchanged(t *testing.T) {
	tests := []runtimeBreakFixture{
		{
			name: "malformed profile base prompt reference",
			breakRuntime: func(t *testing.T, home string) {
				t.Helper()
				configPath := filepath.Join(home, ".config", "opencode", "opencode.json")
				data, err := readJSONAny(configPath)
				if err != nil {
					t.Fatal(err)
				}
				agents, _ := jsonObject(data["agent"])
				mustAgent(t, agents, "sdd-init")["prompt"] = "{file:prompts/sdd/sdd-init.md"
				writeJSONFixture(t, configPath, data)
			},
		},
		{
			name: "missing profile base prompt file",
			breakRuntime: func(t *testing.T, home string) {
				t.Helper()
				if err := os.Remove(filepath.Join(home, ".config", "opencode", "prompts", "sdd", "sdd-init.md")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "obsolete base orchestrator prompt",
			breakRuntime: func(t *testing.T, home string) {
				t.Helper()
				configPath := filepath.Join(home, ".config", "opencode", "opencode.json")
				data, err := readJSONAny(configPath)
				if err != nil {
					t.Fatal(err)
				}
				agents, _ := jsonObject(data["agent"])
				mustAgent(t, agents, "gentle-orchestrator")["prompt"] = "{file:prompts/sdd/orchestrators/gentle-orchestrator.overlay.md}"
				writeJSONFixture(t, configPath, data)
			},
		},
		{
			name: "missing base orchestrator permission",
			breakRuntime: func(t *testing.T, home string) {
				t.Helper()
				mutateFixtureBaseOrchestrator(t, home, func(agent map[string]any) { delete(agent, "permission") })
			},
		},
		{
			name: "missing base orchestrator permission task",
			breakRuntime: func(t *testing.T, home string) {
				t.Helper()
				mutateFixtureBaseOrchestrator(t, home, func(agent map[string]any) {
					agent["permission"] = map[string]any{"question": "allow"}
				})
			},
		},
		{
			name: "base orchestrator permission wildcard is not deny",
			breakRuntime: func(t *testing.T, home string) {
				t.Helper()
				mutateFixtureBaseOrchestrator(t, home, func(agent map[string]any) {
					agent["permission"] = map[string]any{"task": map[string]any{"*": "allow"}}
				})
			},
		},
		{
			name: "incompatible Judgment Day anchors",
			breakRuntime: func(t *testing.T, home string) {
				t.Helper()
				path := filepath.Join(home, ".config", "opencode", "skills", "judgment-day", "SKILL.md")
				mustWriteFile(t, path, []byte("incompatible upstream skill\n"))
			},
		},
		{
			name: "incompatible sdd-research output-only anchor",
			breakRuntime: func(t *testing.T, home string) {
				t.Helper()
				path := filepath.Join(home, ".config", "opencode", "skills", "sdd-research", "SKILL.md")
				mustWriteFile(t, path, []byte("incompatible upstream research skill\n"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoRoot := repoRootForTest(t)
			home, _ := writeCompatibleRuntimeFixture(t, repoRoot)
			tt.breakRuntime(t, home)
			before := snapshotTree(t, home)
			options := applyPolicyOptions{
				recorder:     newVerboseRecorder(false),
				skillTargets: []string{filepath.Join(home, ".config", "opencode", "skills")},
			}
			if code := runApplyPolicyWithOptions(repoRoot, options); code == 0 {
				t.Fatal("runApplyPolicyWithOptions() code = 0, want preflight failure")
			}
			after := snapshotTree(t, home)
			if !reflect.DeepEqual(after, before) {
				t.Fatalf("runtime changed after failed preflight\nbefore: %v\nafter:  %v", before, after)
			}
		})
	}
}

func mutateFixtureBaseOrchestrator(t *testing.T, home string, mutate func(map[string]any)) {
	t.Helper()
	configPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	data, err := readJSONAny(configPath)
	if err != nil {
		t.Fatal(err)
	}
	agents, _ := jsonObject(data["agent"])
	mutate(mustAgent(t, agents, "gentle-orchestrator"))
	writeJSONFixture(t, configPath, data)
}

func writeReviewPluginFixture(t *testing.T, home string, content []byte) {
	t.Helper()
	mustWriteFile(t, filepath.Join(home, ".config", "opencode", "plugins", "review-result-artifacts.ts"), content)
}

func writeCompatibleRuntimeFixture(t *testing.T, repoRoot string) (string, string) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	policy, _, err := loadPolicy(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(home, ".config", "opencode", "opencode.json")
	promptDir := filepath.Join(home, ".config", "opencode", "prompts", "sdd")
	agents := map[string]any{
		"gentle-orchestrator": map[string]any{
			"mode": "primary", "prompt": opaqueUpstreamOrchestratorFixture,
			"tools": map[string]any{"task": true}, "model": "upstream/base",
		},
	}
	baseTaskPermissions := map[string]any{"*": "deny", "general": "allow", "explore": "allow"}
	defaultPhases := map[string]any{}
	profilePhases := map[string]any{}
	for _, phase := range policy.OpenCode.SDDPhases {
		promptPath := filepath.Join(promptDir, phase+".md")
		mustWriteFile(t, promptPath, []byte("upstream rendered phase prompt for "+phase+"\n"))
		if err := os.Chmod(promptPath, 0o640); err != nil {
			t.Fatal(err)
		}
		agents[phase] = map[string]any{
			"mode": "subagent", "prompt": "{file:" + filepath.ToSlash(promptPath) + "}",
			"tools": map[string]any{"read": true}, "model": "upstream/phase",
		}
		baseTaskPermissions[phase] = "allow"
		defaultPhases[phase] = map[string]any{"model": "default/" + phase, "variant": "high"}
		profilePhases[phase] = map[string]any{"model": "profile/" + phase, "variant": "low"}
	}
	actorKeys := append(append([]string{}, judgmentDayActorKeys...), reviewActorKeys...)
	for _, key := range actorKeys {
		baseTaskPermissions[key] = "allow"
		agents[key] = map[string]any{"mode": "subagent", "hidden": true, "prompt": "upstream " + key, "tools": map[string]any{"read": true}}
	}
	mustAgent(t, agents, "gentle-orchestrator")["permission"] = map[string]any{
		"question": "allow",
		"task":     map[string]any{"__replace__": baseTaskPermissions},
	}
	obsoletePath := filepath.Join(home, ".config", "opencode", "prompts", "sdd", "orchestrators", "gentle-orchestrator.overlay.md")
	mustWriteFile(t, obsoletePath, []byte("#### Review Execution Contract\n\n**Sweep budget.** obsolete raw section\n"))
	agents["sdd-orchestrator-cheap"] = map[string]any{"prompt": "{file:" + filepath.ToSlash(obsoletePath) + "}", "model": "old"}
	writeJSONFixture(t, configPath, map[string]any{"agent": agents})

	localConfigPath := filepath.Join(home, ".config", "gentle-ai-custom", "opencode-local-config.json")
	writeJSONFixture(t, localConfigPath, map[string]any{
		"version": 1,
		"default_profile": map[string]any{
			"orchestrator": map[string]any{"model": "default/orchestrator", "variant": "high"},
			"phases":       defaultPhases,
		},
		"profiles": []any{map[string]any{
			"name": "cheap", "orchestrator": map[string]any{"model": "profile/orchestrator", "variant": "low"},
			"phases": profilePhases,
		}},
	})
	writeReviewPluginFixture(t, home, opaquePluginFixture)
	pluginPath := filepath.Join(home, ".config", "opencode", "plugins", "review-result-artifacts.ts")
	if err := os.Chmod(pluginPath, 0o600); err != nil {
		t.Fatal(err)
	}
	contractPath := filepath.Join(home, ".config", "opencode", "skills", "_shared", "review-ledger-contract.md")
	mustWriteFile(t, contractPath, []byte("opaque upstream contract bytes\n"))
	if err := os.Chmod(contractPath, 0o640); err != nil {
		t.Fatal(err)
	}
	upstreamRoot := filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", "skills", "judgment-day")
	mustWriteFile(t, filepath.Join(home, ".config", "opencode", "skills", "judgment-day", "SKILL.md"), []byte(mustReadFile(t, filepath.Join(upstreamRoot, "SKILL.md"))))
	mustWriteFile(t, filepath.Join(home, ".config", "opencode", "skills", "judgment-day", "references", "prompts-and-formats.md"), []byte(mustReadFile(t, filepath.Join(upstreamRoot, "references", "prompts-and-formats.md"))))
	mustWriteFile(t, filepath.Join(home, ".config", "opencode", "skills", "sdd-research", "SKILL.md"), []byte("## Hard Rules\n\n- Do not read local artifacts or call persistence tools. Do not read or mutate repository or Engram state.\n"))
	mustWriteFile(t, filepath.Join(home, ".config", "opencode", "skills", "branch-pr", "SKILL.md"), []byte("pruned sentinel\n"))
	return home, configPath
}

func assertNamedProfileTaskPermissions(t *testing.T, agents map[string]any, phases []string, profileName string) {
	t.Helper()
	assertNamedProfileTaskPermissionsWithJD(t, agents, phases, profileName, nil)
}

func assertNamedProfileTaskPermissionsWithJD(t *testing.T, agents map[string]any, phases []string, profileName string, suffixedJD map[string]bool) {
	t.Helper()
	orchestrator := mustAgent(t, agents, "sdd-orchestrator-"+profileName)
	permission, ok := jsonObject(orchestrator["permission"])
	if !ok {
		t.Fatal("named orchestrator permission object missing")
	}
	tasks, ok := jsonObject(permission["task"])
	if !ok {
		t.Fatal("named orchestrator permission.task object missing")
	}
	if _, leaked := tasks["__replace__"]; leaked {
		t.Fatal("named orchestrator permission.task leaked internal __replace__ marker")
	}
	if got := jsonString(tasks["*"]); got != "deny" {
		t.Fatalf("wildcard permission = %q, want deny", got)
	}
	for _, key := range append([]string{"general", "explore"}, reviewActorKeys...) {
		if got := jsonString(tasks[key]); got != "allow" {
			t.Errorf("task permission %q = %q, want allow", key, got)
		}
	}
	for _, key := range judgmentDayActorKeys {
		wantKey := key
		if suffixedJD[key] {
			wantKey = key + "-" + profileName
			if _, exists := tasks[key]; exists {
				t.Errorf("named orchestrator still authorizes global JD actor %q", key)
			}
		}
		if got := jsonString(tasks[wantKey]); got != "allow" {
			t.Errorf("task permission %q = %q, want allow", wantKey, got)
		}
	}
	for _, phase := range phases {
		if got := jsonString(tasks[phase+"-"+profileName]); got != "allow" {
			t.Errorf("task permission %q = %q, want allow", phase+"-"+profileName, got)
		}
		if _, exists := tasks[phase]; exists {
			t.Errorf("named orchestrator accidentally authorizes unsuffixed phase %q", phase)
		}
	}
}

func readFixtureAgents(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := readJSONAny(path)
	if err != nil {
		t.Fatal(err)
	}
	agents, ok := jsonObject(data["agent"])
	if !ok {
		t.Fatal("fixture config has no agent map")
	}
	return agents
}

func mustAgent(t *testing.T, agents map[string]any, key string) map[string]any {
	t.Helper()
	agent, ok := jsonObject(agents[key])
	if !ok {
		t.Fatalf("agent %q missing", key)
	}
	return agent
}

func mustLoadPolicyPhases(t *testing.T, repoRoot string) []string {
	t.Helper()
	policy, _, err := loadPolicy(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	return policy.OpenCode.SDDPhases
}

func snapshotTree(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		hash := sha256.Sum256(raw)
		result[filepath.ToSlash(rel)] = fmt.Sprintf("%s:%s", hex.EncodeToString(hash[:]), info.Mode())
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return result
}
