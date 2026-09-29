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
	orchestratorBefore := remoteAuthorizationFixturePrompt(t, opaqueUpstreamOrchestratorFixture)
	orchestratorAfter := strings.Replace(orchestratorBefore, mustRemoteAuthorizationSection(t, false), mustRemoteAuthorizationSection(t, true), 1)
	pluginPath := filepath.Join(home, ".config", "opencode", "plugins", "review-result-artifacts.ts")
	phasePath := filepath.Join(home, ".config", "opencode", "prompts", "sdd", "sdd-init.md")
	contractPath := filepath.Join(home, ".config", "opencode", "skills", "_shared", "review-ledger-contract.md")
	researchSkill := filepath.Join(home, ".config", "opencode", "skills", "sdd-research", "SKILL.md")
	preserved := map[string]testFileSnapshot{
		pluginPath:    snapshotTestFile(t, pluginPath),
		contractPath:  snapshotTestFile(t, contractPath),
		researchSkill: snapshotTestFile(t, researchSkill),
	}

	options := applyPolicyOptions{
		recorder:     newVerboseRecorder(false),
		skillTargets: []string{filepath.Join(home, ".config", "opencode", "skills")},
	}
	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("runApplyPolicyWithOptions() code = %d, want 0", code)
	}

	agents := readFixtureAgents(t, configPath)
	if got := jsonString(mustAgent(t, agents, "gentle-orchestrator")["prompt"]); got != orchestratorAfter {
		t.Fatalf("base orchestrator prompt does not contain the local remote-authorization section\nwant: %q\ngot:  %q", orchestratorAfter, got)
	}
	for path, before := range preserved {
		assertTestFileSnapshot(t, path, before)
	}
	assertRemoteAuthorizationSection(t, mustReadFile(t, phasePath), true)
	onboardPath := filepath.Join(home, ".config", "opencode", "prompts", "sdd", "sdd-onboard.md")
	assertRemoteAuthorizationSection(t, mustReadFile(t, onboardPath), true)
	if mode := snapshotTestFile(t, phasePath).mode.Perm(); mode != 0o640 {
		t.Fatalf("phase prompt mode = %o, want preserved 640", mode)
	}
	for _, key := range []string{"general", "explore", "jd-judge-a", "review-risk", "sdd-init-cheap"} {
		t.Run("remote authorization carrier "+key, func(t *testing.T) {
			assertAgentRemoteAuthorizationSection(t, jsonString(mustAgent(t, agents, key)["prompt"]), true)
		})
	}
	assertRemoteAuthorizationSection(t, jsonString(mustAgent(t, agents, "custom-agent")["prompt"]), false)
	for _, key := range []string{"gentle-orchestrator", "sdd-orchestrator-cheap"} {
		if got := jsonString(mustAgent(t, agents, key)["prompt"]); got != orchestratorAfter {
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
	workUnitSkill := filepath.Join(home, ".config", "opencode", "skills", "work-unit-commits", "SKILL.md")
	for path, marker := range map[string]string{
		judgmentSkill:  "invoke the repo-owned `judgment-retrospective` skill",
		judgmentPrompt: "## Terminal Retrospective Handoff",
		workUnitSkill:  "Every ODD task closes with verified work-unit evidence; create a work-unit commit only when the user explicitly requests committing.",
	} {
		if count := strings.Count(mustReadFile(t, path), marker); count != 1 {
			t.Fatalf("patched %s marker count = %d, want 1", path, count)
		}
	}
	firstSkill := mustReadFile(t, judgmentSkill)
	firstPrompt := mustReadFile(t, judgmentPrompt)
	firstWorkUnitSkill := mustReadFile(t, workUnitSkill)
	firstConfig := mustReadFile(t, configPath)
	assertRuntimePatchIsExactDelta(t, repoRoot, home, runtimePatchBaselines)
	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("idempotent run code = %d, want 0", code)
	}
	if mustReadFile(t, judgmentSkill) != firstSkill || mustReadFile(t, judgmentPrompt) != firstPrompt || mustReadFile(t, workUnitSkill) != firstWorkUnitSkill {
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
	baseJudge = mustAgent(t, agents, "jd-judge-a")
	namedJudge := mustAgent(t, agents, "jd-judge-a-cheap")
	if jsonString(namedJudge["model"]) != "profile/judge" {
		t.Fatalf("named judge model = %q, want profile/judge", jsonString(namedJudge["model"]))
	}
	assertRemoteAuthorizationSection(t, jsonString(namedJudge["prompt"]), true)
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

func TestRemoteAuthorizationSectionReplacementIsExactAndIdempotent(t *testing.T) {
	before := mustRemoteAuthorizationSection(t, false)
	after := mustRemoteAuthorizationSection(t, true)
	definition := runtimePatchSection{Before: before, After: after}
	original := "unrelated leading bytes\n\n" + before + "\n\n## Next section\nkeep this too\n"
	want := "unrelated leading bytes\n\n" + after + "\n\n## Next section\nkeep this too\n"
	updated, err := replaceRemoteAuthorizationSection(original, definition)
	if err != nil {
		t.Fatal(err)
	}
	if updated != want {
		t.Fatalf("replacement changed unrelated content\nwant: %q\ngot:  %q", want, updated)
	}
	second, err := replaceRemoteAuthorizationSection(updated, definition)
	if err != nil || second != updated {
		t.Fatalf("revised section was not idempotent: err=%v", err)
	}
	markedOriginal := "unrelated prefix\r\n<!-- gentle-ai:remote-authorization -->\r\n" + before + "\r\n<!-- /gentle-ai:remote-authorization -->"
	markedWant := "unrelated prefix\r\n<!-- gentle-ai:remote-authorization -->\r\n" + after + "\r\n<!-- /gentle-ai:remote-authorization -->"
	markedUpdated, err := replaceRemoteAuthorizationSection(markedOriginal, definition)
	if err != nil || markedUpdated != markedWant {
		t.Fatalf("marked EOF section replacement changed markers or unrelated bytes: err=%v\nwant: %q\ngot:  %q", err, markedWant, markedUpdated)
	}
	if second, err := replaceRemoteAuthorizationSection(markedUpdated, definition); err != nil || second != markedUpdated {
		t.Fatalf("marked section was not idempotent: err=%v", err)
	}
	markedBeforeHeading := "<!-- gentle-ai:remote-authorization -->\n" + before + "\n<!-- /gentle-ai:remote-authorization -->\n\n## Next section\nkeep this too\n"
	markedAfterHeading := "<!-- gentle-ai:remote-authorization -->\n" + after + "\n<!-- /gentle-ai:remote-authorization -->\n\n## Next section\nkeep this too\n"
	if got, err := replaceRemoteAuthorizationSection(markedBeforeHeading, definition); err != nil || got != markedAfterHeading {
		t.Fatalf("marked section before later heading changed unrelated bytes: err=%v\nwant: %q\ngot:  %q", err, markedAfterHeading, got)
	}
	for _, invalid := range []struct {
		name    string
		content string
	}{
		{name: "missing", content: "no remote section"},
		{name: "duplicate", content: before + "\n\n" + before},
		{name: "unknown heading", content: strings.Replace(before, "## Remote operation authorization", "### Remote authorization", 1)},
		{name: "changed body", content: strings.Replace(before, "Before remote work", "Before any remote work", 1)},
		{name: "mixed old and new", content: after + "\n\n" + before},
		{name: "duplicate opening marker", content: "<!-- gentle-ai:remote-authorization -->\n<!-- gentle-ai:remote-authorization -->\n" + before + "\n<!-- /gentle-ai:remote-authorization -->"},
		{name: "missing closing marker", content: "<!-- gentle-ai:remote-authorization -->\n" + before},
		{name: "reversed markers", content: "<!-- /gentle-ai:remote-authorization -->\n" + before + "\n<!-- gentle-ai:remote-authorization -->"},
		{name: "unexpected text before heading", content: "<!-- gentle-ai:remote-authorization -->\nunexpected\n" + before + "\n<!-- /gentle-ai:remote-authorization -->"},
	} {
		t.Run(invalid.name, func(t *testing.T) {
			if _, err := replaceRemoteAuthorizationSection(invalid.content, definition); err == nil {
				t.Fatal("replaceRemoteAuthorizationSection() error = nil, want fail-closed anchor rejection")
			}
		})
	}
}

func TestRemoteAuthorizationSkipsAbsentOptionalSharedFilesButRejectsMissingReferences(t *testing.T) {
	repoRoot := repoRootForTest(t)
	runtimeRoot := t.TempDir()
	state := applyPolicyState{
		configPath: filepath.Join(runtimeRoot, "opencode.json"),
		managedTarget: ManagedAssetsTarget{
			RuntimePatches:     []RuntimePatchRef{{Key: remoteAuthorizationPatchKey, PatchPath: "overlay/gentle-ai/patches/opencode/remote-authorization-scope.json"}},
			OwnedOverlayAssets: []OwnedOverlayAsset{{RuntimeTargets: []string{"~/.config/opencode/prompts/sdd/optional.md"}}},
		},
		agents: map[string]any{},
	}
	state.policy.OpenCode.BaseOrchestratorKey = "base"
	if err := state.prepareRemoteAuthorizationCustomization(repoRoot); err != nil {
		t.Fatalf("absent optional shared prompt failed preflight: %v", err)
	}
	state.agents["base"] = map[string]any{"prompt": "{file:" + filepath.ToSlash(filepath.Join(runtimeRoot, "prompts", "missing.md")) + "}"}
	if err := state.prepareRemoteAuthorizationCustomization(repoRoot); err == nil {
		t.Fatal("missing referenced prompt passed preflight; want failure")
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
		if len(definition.Sections) > 0 {
			continue
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
		if len(definition.Sections) > 0 {
			continue
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
			name: "incompatible ODD work-unit commit anchor",
			breakRuntime: func(t *testing.T, home string) {
				t.Helper()
				path := filepath.Join(home, ".config", "opencode", "skills", "work-unit-commits", "SKILL.md")
				mustWriteFile(t, path, []byte("incompatible upstream work-unit skill\n"))
			},
		},
		{
			name: "duplicate remote authorization section",
			breakRuntime: func(t *testing.T, home string) {
				t.Helper()
				configPath := filepath.Join(home, ".config", "opencode", "opencode.json")
				data, err := readJSONAny(configPath)
				if err != nil {
					t.Fatal(err)
				}
				agents, _ := jsonObject(data["agent"])
				prompt := jsonString(mustAgent(t, agents, "gentle-orchestrator")["prompt"])
				mustAgent(t, agents, "gentle-orchestrator")["prompt"] = prompt + "\n\n" + mustRemoteAuthorizationSection(t, false)
				writeJSONFixture(t, configPath, data)
			},
		},
		{
			name: "modified remote authorization section",
			breakRuntime: func(t *testing.T, home string) {
				t.Helper()
				configPath := filepath.Join(home, ".config", "opencode", "opencode.json")
				data, err := readJSONAny(configPath)
				if err != nil {
					t.Fatal(err)
				}
				agents, _ := jsonObject(data["agent"])
				prompt := jsonString(mustAgent(t, agents, "gentle-orchestrator")["prompt"])
				mustAgent(t, agents, "gentle-orchestrator")["prompt"] = strings.Replace(prompt, "Before remote work", "Before any remote work", 1)
				writeJSONFixture(t, configPath, data)
			},
		},
		{
			name: "malformed remote authorization marker pair",
			breakRuntime: func(t *testing.T, home string) {
				t.Helper()
				path := filepath.Join(home, ".config", "opencode", "prompts", "sdd", "sdd-onboard.md")
				content := mustReadFile(t, path)
				content = strings.Replace(content, "\n<!-- /gentle-ai:remote-authorization -->", "", 1)
				mustWriteFile(t, path, []byte(content))
			},
		},
		{
			name: "remote-bearing unmanaged named profile",
			breakRuntime: func(t *testing.T, home string) {
				t.Helper()
				configPath := filepath.Join(home, ".config", "opencode", "opencode.json")
				data, err := readJSONAny(configPath)
				if err != nil {
					t.Fatal(err)
				}
				agents, _ := jsonObject(data["agent"])
				agents["sdd-orchestrator-unmanaged"] = map[string]any{"prompt": remoteAuthorizationFixturePrompt(t, "unmanaged")}
				writeJSONFixture(t, configPath, data)
			},
		},
		{
			name: "symlink prompt reference",
			breakRuntime: func(t *testing.T, home string) {
				t.Helper()
				configPath := filepath.Join(home, ".config", "opencode", "opencode.json")
				data, err := readJSONAny(configPath)
				if err != nil {
					t.Fatal(err)
				}
				agents, _ := jsonObject(data["agent"])
				promptPath := filepath.Join(home, ".config", "opencode", "prompts", "sdd", "sdd-init.md")
				linkPath := filepath.Join(home, ".config", "opencode", "prompts", "sdd", "linked-sdd-init.md")
				if err := os.Symlink(promptPath, linkPath); err != nil {
					t.Fatal(err)
				}
				mustAgent(t, agents, "sdd-init")["prompt"] = "{file:" + filepath.ToSlash(linkPath) + "}"
				writeJSONFixture(t, configPath, data)
			},
		},
		{
			name: "out-of-root prompt reference",
			breakRuntime: func(t *testing.T, home string) {
				t.Helper()
				configPath := filepath.Join(home, ".config", "opencode", "opencode.json")
				data, err := readJSONAny(configPath)
				if err != nil {
					t.Fatal(err)
				}
				agents, _ := jsonObject(data["agent"])
				outsidePath := filepath.Join(t.TempDir(), "outside-prompt.md")
				mustWriteFile(t, outsidePath, []byte(remoteAuthorizationFixturePrompt(t, "outside prompt")))
				mustAgent(t, agents, "sdd-init")["prompt"] = "{file:" + filepath.ToSlash(outsidePath) + "}"
				writeJSONFixture(t, configPath, data)
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
			"mode": "primary", "prompt": remoteAuthorizationFixturePrompt(t, opaqueUpstreamOrchestratorFixture),
			"tools": map[string]any{"task": true}, "model": "upstream/base",
		},
	}
	baseTaskPermissions := map[string]any{"*": "deny", "general": "allow", "explore": "allow"}
	defaultPhases := map[string]any{}
	profilePhases := map[string]any{}
	for _, phase := range policy.OpenCode.SDDPhases {
		promptPath := filepath.Join(promptDir, phase+".md")
		mustWriteFile(t, promptPath, []byte(remoteAuthorizationFixturePrompt(t, "upstream rendered phase prompt for "+phase+"\n")))
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
	for _, key := range []string{"general", "explore"} {
		agents[key] = map[string]any{"mode": "subagent", "prompt": remoteAuthorizationFixturePrompt(t, "upstream "+key)}
	}
	agents["custom-agent"] = map[string]any{"prompt": remoteAuthorizationFixturePrompt(t, "unrelated custom agent")}
	actorKeys := append(append([]string{}, judgmentDayActorKeys...), reviewActorKeys...)
	for _, key := range actorKeys {
		baseTaskPermissions[key] = "allow"
		agents[key] = map[string]any{"mode": "subagent", "hidden": true, "prompt": remoteAuthorizationFixturePrompt(t, "upstream "+key), "tools": map[string]any{"read": true}}
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
	workUnitSkill := filepath.Join(home, ".config", "opencode", "skills", "work-unit-commits", "SKILL.md")
	mustWriteFile(t, workUnitSkill, []byte("## ODD Relationship\n\nOrganic Driven Development (ODD) closes every substantial task with a work-unit commit on the feature branch, not only SDD tasks:\n\n- Every ODD task closes with at least one work-unit commit, branch first when on the default branch, with tests and docs alongside the behavior and a Conventional Commit message.\n- The native review candidate is that commit.\n"))
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

func remoteAuthorizationFixturePrompt(t *testing.T, prompt string) string {
	t.Helper()
	repoRoot := repoRootForTest(t)
	var definition runtimePatchDefinition
	patchPath := filepath.Join(repoRoot, "overlay", "gentle-ai", "patches", "opencode", "remote-authorization-scope.json")
	if err := readJSONFile(patchPath, &definition); err != nil {
		t.Fatal(err)
	}
	if len(definition.Sections) != 1 {
		t.Fatalf("remote patch sections = %d, want one", len(definition.Sections))
	}
	return strings.TrimRight(prompt, "\n") + "\n\n<!-- gentle-ai:remote-authorization -->\n" + definition.Sections[0].Before + "\n<!-- /gentle-ai:remote-authorization -->"
}

func assertRemoteAuthorizationSection(t *testing.T, content string, revised bool) {
	t.Helper()
	want := mustRemoteAuthorizationSection(t, revised)
	if !strings.Contains(content, want) || revised && strings.Contains(content, mustRemoteAuthorizationSection(t, false)) {
		t.Fatalf("remote-authorization section does not match revised=%v; content=%q", revised, content)
	}
}

func assertAgentRemoteAuthorizationSection(t *testing.T, prompt string, revised bool) {
	t.Helper()
	if strings.HasPrefix(prompt, "{file:") && strings.HasSuffix(prompt, "}") {
		path := strings.TrimSuffix(strings.TrimPrefix(prompt, "{file:"), "}")
		prompt = mustReadFile(t, path)
	}
	assertRemoteAuthorizationSection(t, prompt, revised)
}

func mustRemoteAuthorizationSection(t *testing.T, revised bool) string {
	t.Helper()
	repoRoot := repoRootForTest(t)
	var definition runtimePatchDefinition
	patchPath := filepath.Join(repoRoot, "overlay", "gentle-ai", "patches", "opencode", "remote-authorization-scope.json")
	if err := readJSONFile(patchPath, &definition); err != nil {
		t.Fatal(err)
	}
	if len(definition.Sections) != 1 {
		t.Fatalf("remote patch sections = %d, want one", len(definition.Sections))
	}
	if revised {
		return definition.Sections[0].After
	}
	return definition.Sections[0].Before
}
