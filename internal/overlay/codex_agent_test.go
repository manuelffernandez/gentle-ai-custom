package overlay

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeTargetsIncludesCodexAndAll(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want []string
	}{
		{name: "codex", args: []string{"codex"}, want: []string{"codex"}},
		{name: "all", args: []string{"all"}, want: []string{"codex", "opencode"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, got, code := normalizeTargets(tc.args)
			if code >= 0 {
				t.Fatalf("normalizeTargets(%v) exit code = %d", tc.args, code)
			}
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Fatalf("normalizeTargets(%v) = %v, want %v", tc.args, got, tc.want)
			}
		})
	}
}

func TestCodexApplyPatchesOnlyDeclaredUpstreamSectionsAndIsIdempotent(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	codexRoot := filepath.Join(home, ".codex")
	agentsPath := filepath.Join(codexRoot, "AGENTS.md")
	agentsBefore := "upstream ODD instructions\n<!-- gentle-ai:engram-protocol -->\nupstream Engram body\n<!-- /gentle-ai:engram-protocol -->\nupstream suffix\n"
	mustWriteFile(t, agentsPath, []byte(agentsBefore))
	if err := os.Chmod(agentsPath, 0o640); err != nil {
		t.Fatal(err)
	}
	judgmentPath := filepath.Join(codexRoot, "skills", "judgment-day", "SKILL.md")
	judgmentBefore := "- The only terminal verdicts are `APPROVED | ESCALATED`; never reset or extend an exhausted round budget.\n\n6. Repeat once at most, then run independent final verification and return the terminal verdict.\n\nartifact references, skill resolution, and exactly one final\n"
	mustWriteFile(t, judgmentPath, []byte(judgmentBefore))
	if err := os.Chmod(judgmentPath, 0o600); err != nil {
		t.Fatal(err)
	}
	judgmentPromptPath := filepath.Join(codexRoot, "skills", "judgment-day", "references", "prompts-and-formats.md")
	judgmentPromptBefore := "Warnings and suggestions are informational. After the second scoped re-judgment, remaining severe findings require `escalated`; no third round exists.\n"
	mustWriteFile(t, judgmentPromptPath, []byte(judgmentPromptBefore))
	if err := os.Chmod(judgmentPromptPath, 0o640); err != nil {
		t.Fatal(err)
	}
	workUnitPath := filepath.Join(codexRoot, "skills", "work-unit-commits", "SKILL.md")
	workUnitBefore := "- Every ODD task closes with at least one work-unit commit, branch first when on the default branch, with tests and docs alongside the behavior and a Conventional Commit message.\n"
	mustWriteFile(t, workUnitPath, []byte(workUnitBefore))
	if err := os.Chmod(workUnitPath, 0o600); err != nil {
		t.Fatal(err)
	}
	upstreamSkillPath := filepath.Join(codexRoot, "skills", "sdd-init", "SKILL.md")
	mustWriteFile(t, upstreamSkillPath, []byte("upstream SDD skill\n"))
	configPath := filepath.Join(codexRoot, "config.toml")
	mustWriteFile(t, configPath, []byte("model = \"upstream\"\n"))
	configBefore := mustReadFile(t, configPath)

	options := applyPolicyOptions{recorder: newVerboseRecorder(false)}
	agent := &CodexAgent{}
	if code := agent.ApplyOverlay(repoRoot, options); code != 0 {
		t.Fatalf("CodexAgent.ApplyOverlay() = %d, want 0", code)
	}
	agentsAfter := mustReadFile(t, agentsPath)
	if !strings.HasPrefix(agentsAfter, "upstream ODD instructions\n<!-- gentle-ai:engram-protocol -->\nupstream Engram body\n<!-- /gentle-ai:engram-protocol -->") {
		t.Fatalf("upstream Codex instructions changed: %q", agentsAfter)
	}
	if !strings.HasSuffix(agentsAfter, "upstream suffix\n") {
		t.Fatalf("upstream Codex suffix changed: %q", agentsAfter)
	}
	for _, marker := range []string{
		"gentle-ai-custom:codex-local-instructions",
		"Never read, search, copy, or expose credential files",
	} {
		if !strings.Contains(agentsAfter, marker) {
			t.Errorf("Codex AGENTS.md missing local instruction %q", marker)
		}
	}
	if strings.Contains(agentsAfter, "Never commit, push, or create a pull request") {
		t.Fatal("Codex local instructions still override upstream ODD task-close commits")
	}
	if strings.Contains(agentsAfter, "Remote operation authorization") {
		t.Fatal("Codex local instructions added the retired remote-operation authorization rule")
	}
	if got := snapshotTestFile(t, agentsPath).mode.Perm(); got != 0o640 {
		t.Fatalf("Codex AGENTS.md mode = %o, want preserved 640", got)
	}
	for path, marker := range map[string]string{
		judgmentPath:       "invoke the repo-owned `judgment-retrospective` skill",
		judgmentPromptPath: "## Terminal Retrospective Handoff",
	} {
		if strings.Count(mustReadFile(t, path), marker) != 1 {
			t.Errorf("Codex patch marker in %s is not present exactly once", path)
		}
	}
	if got := mustReadFile(t, workUnitPath); got != workUnitBefore {
		t.Errorf("upstream work-unit skill changed without a declared patch: %q", got)
	}
	for path, mode := range map[string]os.FileMode{judgmentPath: 0o600, judgmentPromptPath: 0o640, workUnitPath: 0o600} {
		if got := snapshotTestFile(t, path).mode.Perm(); got != mode {
			t.Errorf("Codex patched skill mode at %s = %o, want preserved %o", path, got, mode)
		}
	}
	if got := mustReadFile(t, upstreamSkillPath); got != "upstream SDD skill\n" {
		t.Fatalf("Codex upstream SDD skill changed: %q", got)
	}
	if got := mustReadFile(t, configPath); got != configBefore {
		t.Fatalf("Codex config.toml changed: %q", got)
	}

	first := snapshotTree(t, codexRoot)
	if code := agent.ApplyOverlay(repoRoot, options); code != 0 {
		t.Fatalf("idempotent CodexAgent.ApplyOverlay() = %d, want 0", code)
	}
	if second := snapshotTree(t, codexRoot); !equalStringMaps(first, second) {
		t.Fatal("Codex patches are not byte- and mode-idempotent")
	}
}

func TestCodexApplyRejectsIncompatibleAnchorBeforeAnyWrites(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	codexRoot := filepath.Join(home, ".codex")
	agentsPath := filepath.Join(codexRoot, "AGENTS.md")
	mustWriteFile(t, agentsPath, []byte("changed upstream Codex instructions\n"))
	judgmentPath := filepath.Join(codexRoot, "skills", "judgment-day", "SKILL.md")
	mustWriteFile(t, judgmentPath, []byte("compatible first patch target\n"))
	workUnitPath := filepath.Join(codexRoot, "skills", "work-unit-commits", "SKILL.md")
	mustWriteFile(t, workUnitPath, []byte("compatible second patch target\n"))
	before := snapshotTree(t, codexRoot)
	if code := (&CodexAgent{}).ApplyOverlay(repoRoot, applyPolicyOptions{recorder: newVerboseRecorder(false)}); code == 0 {
		t.Fatal("CodexAgent.ApplyOverlay() = 0, want incompatible-anchor failure")
	}
	if after := snapshotTree(t, codexRoot); !equalStringMaps(before, after) {
		t.Fatalf("Codex files changed after incompatible-anchor preflight\nbefore: %v\nafter: %v", before, after)
	}
}

func TestCodexApplyDoesNotRequireSDDMarker(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	codexRoot := filepath.Join(home, ".codex")
	agentsPath := filepath.Join(codexRoot, "AGENTS.md")
	mustWriteFile(t, agentsPath, []byte("upstream ODD instructions without an SDD marker\n<!-- /gentle-ai:engram-protocol -->\n"))
	mustWriteFile(t, filepath.Join(codexRoot, "skills", "judgment-day", "SKILL.md"), []byte("- The only terminal verdicts are `APPROVED | ESCALATED`; never reset or extend an exhausted round budget.\n\n6. Repeat once at most, then run independent final verification and return the terminal verdict.\n\nartifact references, skill resolution, and exactly one final\n"))
	mustWriteFile(t, filepath.Join(codexRoot, "skills", "judgment-day", "references", "prompts-and-formats.md"), []byte("Warnings and suggestions are informational. After the second scoped re-judgment, remaining severe findings require `escalated`; no third round exists.\n"))
	mustWriteFile(t, filepath.Join(codexRoot, "skills", "work-unit-commits", "SKILL.md"), []byte("- Every ODD task closes with at least one work-unit commit, branch first when on the default branch, with tests and docs alongside the behavior and a Conventional Commit message.\n"))
	if code := (&CodexAgent{}).ApplyOverlay(repoRoot, applyPolicyOptions{recorder: newVerboseRecorder(false)}); code != 0 {
		t.Fatalf("CodexAgent.ApplyOverlay() = %d, want success without the retired SDD marker", code)
	}
	if !strings.Contains(mustReadFile(t, agentsPath), "upstream ODD instructions without an SDD marker") {
		t.Fatal("Codex apply did not preserve upstream instructions without the SDD marker")
	}
}

func TestRunApplyCustomPreflightsCodexBeforeAnySelectedTargetWrites(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(binDir, "gentle-ai"), []byte("#!/bin/sh\nprintf 'gentle-ai 4.0.0\\n'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	openCodeSentinel := filepath.Join(home, ".config", "opencode", "skills", "branch-pr", "SKILL.md")
	mustWriteFile(t, openCodeSentinel, []byte("must remain unchanged when Codex anchors fail\n"))
	agentsPath := filepath.Join(home, ".codex", "AGENTS.md")
	mustWriteFile(t, agentsPath, []byte("changed upstream Codex instructions\n"))
	before := snapshotTree(t, home)

	if code := RunApplyCustom(repoRoot, []string{"opencode", "codex"}); code == 0 {
		t.Fatal("RunApplyCustom(opencode,codex) = 0, want Codex preflight failure")
	}
	if after := snapshotTree(t, home); !equalStringMaps(before, after) {
		t.Fatalf("selected targets changed after Codex anchor failure\nbefore: %v\nafter: %v", before, after)
	}
}

func TestRunApplyCustomCodexInstallsSharedSkillsWithoutTouchingOtherTargets(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	fakeVersion := filepath.Join(binDir, "gentle-ai")
	if err := os.WriteFile(fakeVersion, []byte("#!/bin/sh\nprintf 'gentle-ai 4.0.0\\n'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	codexRoot := filepath.Join(home, ".codex")
	mustWriteFile(t, filepath.Join(codexRoot, "AGENTS.md"), []byte("<!-- gentle-ai:engram-protocol -->\nupstream Engram\n<!-- /gentle-ai:engram-protocol -->\n"))
	mustWriteFile(t, filepath.Join(codexRoot, "skills", "judgment-day", "SKILL.md"), []byte("- The only terminal verdicts are `APPROVED | ESCALATED`; never reset or extend an exhausted round budget.\n\n6. Repeat once at most, then run independent final verification and return the terminal verdict.\n\nartifact references, skill resolution, and exactly one final\n"))
	mustWriteFile(t, filepath.Join(codexRoot, "skills", "judgment-day", "references", "prompts-and-formats.md"), []byte("Warnings and suggestions are informational. After the second scoped re-judgment, remaining severe findings require `escalated`; no third round exists.\n"))
	mustWriteFile(t, filepath.Join(codexRoot, "skills", "work-unit-commits", "SKILL.md"), []byte("- Every ODD task closes with at least one work-unit commit, branch first when on the default branch, with tests and docs alongside the behavior and a Conventional Commit message.\n"))
	upstreamSDD := filepath.Join(codexRoot, "skills", "sdd-init", "SKILL.md")
	mustWriteFile(t, upstreamSDD, []byte("upstream Codex SDD skill\n"))
	unselectedUpstreamSkill := filepath.Join(codexRoot, "skills", "issue-creation", "SKILL.md")
	mustWriteFile(t, unselectedUpstreamSkill, []byte("unselected Codex skill\n"))
	openCodeSentinel := filepath.Join(home, ".config", "opencode", "skills", "branch-pr", "SKILL.md")
	mustWriteFile(t, openCodeSentinel, []byte("unselected OpenCode sentinel\n"))

	if code := RunApplyCustom(repoRoot, []string{"codex"}); code != 0 {
		t.Fatalf("RunApplyCustom(codex) = %d, want 0", code)
	}
	for _, skill := range []string{"branch-pr", "chained-pr", "code-design", "commit-planner", "judgment-retrospective", "package-security", "pr-finalizer"} {
		installed := filepath.Join(codexRoot, "skills", skill, "SKILL.md")
		source := filepath.Join(repoRoot, "shared", "skills", skill, "SKILL.md")
		if got, want := mustReadFile(t, installed), mustReadFile(t, source); got != want {
			t.Errorf("installed Codex skill %q differs from shared source", skill)
		}
	}
	if got := mustReadFile(t, upstreamSDD); got != "upstream Codex SDD skill\n" {
		t.Fatalf("upstream Codex SDD skill changed: %q", got)
	}
	if got := mustReadFile(t, unselectedUpstreamSkill); got != "unselected Codex skill\n" {
		t.Fatalf("unselected upstream Codex skill changed: %q", got)
	}
	if got := mustReadFile(t, openCodeSentinel); got != "unselected OpenCode sentinel\n" {
		t.Fatalf("unselected OpenCode target changed: %q", got)
	}
	if pathExists(filepath.Join(codexRoot, "prompts", "commit-plan.md")) {
		t.Fatal("Codex apply rendered an OpenCode/shared command as a Codex prompt")
	}
}

func TestRunApplyCustomAllAppliesBothTargetsWithoutCrossTargetCommands(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home, openCodeConfig := writeCompatibleRuntimeFixture(t, repoRoot)
	codexRoot := filepath.Join(home, ".codex")
	mustWriteFile(t, filepath.Join(codexRoot, "AGENTS.md"), []byte("<!-- gentle-ai:engram-protocol -->\nupstream Codex Engram\n<!-- /gentle-ai:engram-protocol -->\n"))
	mustWriteFile(t, filepath.Join(codexRoot, "skills", "judgment-day", "SKILL.md"), []byte("- The only terminal verdicts are `APPROVED | ESCALATED`; never reset or extend an exhausted round budget.\n\n6. Repeat once at most, then run independent final verification and return the terminal verdict.\n\nartifact references, skill resolution, and exactly one final\n"))
	mustWriteFile(t, filepath.Join(codexRoot, "skills", "judgment-day", "references", "prompts-and-formats.md"), []byte("Warnings and suggestions are informational. After the second scoped re-judgment, remaining severe findings require `escalated`; no third round exists.\n"))
	mustWriteFile(t, filepath.Join(codexRoot, "skills", "work-unit-commits", "SKILL.md"), []byte("- Every ODD task closes with at least one work-unit commit, branch first when on the default branch, with tests and docs alongside the behavior and a Conventional Commit message.\n"))
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(binDir, "gentle-ai"), []byte("#!/bin/sh\nprintf 'gentle-ai 4.0.0\\n'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	if code := RunApplyCustom(repoRoot, []string{"all"}); code != 0 {
		t.Fatalf("RunApplyCustom(all) = %d, want 0", code)
	}
	for _, root := range []string{filepath.Join(home, ".codex"), filepath.Join(home, ".config", "opencode")} {
		installed := filepath.Join(root, "skills", "branch-pr", "SKILL.md")
		source := filepath.Join(repoRoot, "shared", "skills", "branch-pr", "SKILL.md")
		if got, want := mustReadFile(t, installed), mustReadFile(t, source); got != want {
			t.Errorf("branch-pr source mismatch under %s", root)
		}
	}
	if !strings.Contains(mustReadFile(t, filepath.Join(codexRoot, "AGENTS.md")), "gentle-ai-custom:codex-local-instructions") {
		t.Fatal("all target did not apply Codex local instructions")
	}
	if !strings.Contains(mustReadFile(t, openCodeConfig), "gentle-orchestrator") {
		t.Fatal("all target did not preserve OpenCode config output")
	}
	if pathExists(filepath.Join(codexRoot, "prompts", "commit-plan.md")) {
		t.Fatal("all target rendered an OpenCode command wrapper for Codex")
	}
}

func equalStringMaps(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}
