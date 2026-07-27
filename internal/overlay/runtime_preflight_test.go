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
	for path, marker := range map[string]string{
		judgmentSkill:  "invoke the repo-owned `judgment-retrospective` skill",
		judgmentPrompt: "## Terminal Retrospective Handoff",
	} {
		if count := strings.Count(mustReadFile(t, path), marker); count != 1 {
			t.Fatalf("patched %s marker count = %d, want 1", path, count)
		}
	}
	firstSkill := mustReadFile(t, judgmentSkill)
	firstPrompt := mustReadFile(t, judgmentPrompt)
	firstConfig := mustReadFile(t, configPath)
	assertRuntimePatchIsExactDelta(t, repoRoot, home)
	if code := runApplyPolicyWithOptions(repoRoot, options); code != 0 {
		t.Fatalf("idempotent run code = %d, want 0", code)
	}
	if mustReadFile(t, judgmentSkill) != firstSkill || mustReadFile(t, judgmentPrompt) != firstPrompt {
		t.Fatal("Judgment Day runtime patch is not byte-idempotent")
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

func assertRuntimePatchIsExactDelta(t *testing.T, repoRoot, home string) {
	t.Helper()
	var definition runtimePatchDefinition
	patchPath := filepath.Join(repoRoot, "overlay", "gentle-ai", "patches", "opencode", "judgment-retrospective.json")
	if err := readJSONFile(patchPath, &definition); err != nil {
		t.Fatal(err)
	}
	for _, target := range definition.Targets {
		runtimePath := filepath.Join(home, ".config", "opencode", filepath.FromSlash(target.Path))
		reverted := mustReadFile(t, runtimePath)
		for index := len(target.Replacements) - 1; index >= 0; index-- {
			replacement := target.Replacements[index]
			if strings.Count(reverted, replacement.After) != 1 {
				t.Fatalf("runtime patch target %s does not contain exact declared delta %d", target.Path, index+1)
			}
			reverted = strings.Replace(reverted, replacement.After, replacement.Before, 1)
		}
		upstreamPath := filepath.Join(repoRoot, "overlay", "gentle-ai", "assets", "upstream", "opencode", filepath.FromSlash(target.Path))
		if reverted != mustReadFile(t, upstreamPath) {
			t.Fatalf("runtime patch target %s differs from upstream beyond the declared replacements", target.Path)
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
	for _, key := range []string{
		"jd-judge-a", "jd-judge-b", "jd-fix-agent",
		"review-risk", "review-readability", "review-reliability", "review-resilience", "review-refuter",
	} {
		baseTaskPermissions[key] = "allow"
		agents[key] = map[string]any{"mode": "subagent", "prompt": "upstream " + key}
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
	mustWriteFile(t, filepath.Join(home, ".config", "opencode", "skills", "branch-pr", "SKILL.md"), []byte("pruned sentinel\n"))
	return home, configPath
}

func assertNamedProfileTaskPermissions(t *testing.T, agents map[string]any, phases []string, profileName string) {
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
	for _, key := range []string{"general", "explore", "jd-judge-a", "jd-judge-b", "jd-fix-agent", "review-risk", "review-readability", "review-reliability", "review-resilience", "review-refuter"} {
		if got := jsonString(tasks[key]); got != "allow" {
			t.Errorf("task permission %q = %q, want allow", key, got)
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
