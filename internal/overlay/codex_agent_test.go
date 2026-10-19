package overlay

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const codexODDAssignmentTable = "| `odd-explorer` | `gpt-6.1-luna` | `high` |\n| `odd-worker` | `gpt-6.1-luna` | `high` |\n| `odd-verify` | `gpt-6.1-sol` | `medium` |\n| `jd-judge-a` | `gpt-6.1-luna` | `high` |\n| `review-risk` | `gpt-6.1-luna` | `high` |\n"

const upstreamAgreementRule = "- Never agree with user claims without verification. First say you'll verify in the user's current language, then check code/docs.\n"
const localAgreementRule = "- Check consequential factual or technical claims before endorsing them; do not agree merely to be agreeable. Use the available evidence to give a calibrated, actionable judgment, even when it is incomplete.\n"
const upstreamTechnicalRule = "- Verify technical claims before stating them. If unsure, investigate first.\n"
const localUncertaintyRule = "- Distinguish inference from established fact only when that distinction materially affects the decision. State material uncertainty once, briefly; do not repeat obvious caveats or narrate routine verification.\n"
const codexUpstreamPersona = upstreamAgreementRule + upstreamTechnicalRule

func TestCodexApplyCalibratesEvidenceRulesWithoutChangingOtherInstructions(t *testing.T) {
	repo := repoRootForTest(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := filepath.Join(home, ".codex")
	path := filepath.Join(root, "AGENTS.md")
	before := strings.Replace(codexMigrationFixture("", ""), upstreamAgreementRule, upstreamAgreementRule+"- If user is wrong, explain WHY with evidence.\n", 1)
	mustWriteFile(t, path, []byte(before))
	writeCodexMigrationPrerequisites(t, root)
	if code := (&CodexAgent{}).ApplyOverlay(repo, applyPolicyOptions{recorder: newVerboseRecorder(false)}); code != 0 {
		t.Fatalf("Codex apply = %d, want success", code)
	}
	got := mustReadFile(t, path)
	if !strings.Contains(got, localAgreementRule) || !strings.Contains(got, localUncertaintyRule) {
		t.Fatal("Codex apply did not install both calibrated evidence rules")
	}
	if strings.Contains(got, upstreamAgreementRule) || strings.Contains(got, upstreamTechnicalRule) ||
		!strings.Contains(got, "- If user is wrong, explain WHY with evidence.\n") {
		t.Fatal("Codex apply retained an obsolete rule or changed an unrelated instruction")
	}
	first := snapshotTree(t, root)
	if code := (&CodexAgent{}).ApplyOverlay(repo, applyPolicyOptions{recorder: newVerboseRecorder(false)}); code != 0 {
		t.Fatalf("repeat Codex apply = %d, want success", code)
	}
	if second := snapshotTree(t, root); !equalStringMaps(first, second) {
		t.Fatal("repeat Codex apply changed bytes or modes")
	}
}

func TestCodexApplyRejectsChangedEvidenceRuleBeforeWrites(t *testing.T) {
	for _, tc := range []struct {
		name, original, replacement string
	}{
		{"changed agreement", upstreamAgreementRule, strings.Replace(upstreamAgreementRule, "First say you'll verify", "Always say you'll verify", 1)},
		{"duplicate agreement", upstreamAgreementRule, upstreamAgreementRule + upstreamAgreementRule},
		{"mixed agreement", upstreamAgreementRule, upstreamAgreementRule + localAgreementRule},
		{"changed technical rule", upstreamTechnicalRule, strings.Replace(upstreamTechnicalRule, "If unsure", "If uncertain", 1)},
		{"duplicate technical rule", upstreamTechnicalRule, upstreamTechnicalRule + upstreamTechnicalRule},
		{"mixed technical rule", upstreamTechnicalRule, upstreamTechnicalRule + localUncertaintyRule},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := repoRootForTest(t)
			home := t.TempDir()
			t.Setenv("HOME", home)
			root := filepath.Join(home, ".codex")
			path := filepath.Join(root, "AGENTS.md")
			beforeRules := strings.Replace(codexMigrationFixture("", ""), tc.original, tc.replacement, 1)
			mustWriteFile(t, path, []byte(beforeRules))
			writeCodexMigrationPrerequisites(t, root)
			before := snapshotTree(t, root)
			if code := (&CodexAgent{}).ApplyOverlay(repo, applyPolicyOptions{recorder: newVerboseRecorder(false)}); code == 0 {
				t.Fatal("Codex apply accepted an incompatible evidence rule")
			}
			if after := snapshotTree(t, root); !equalStringMaps(before, after) {
				t.Fatal("Codex apply wrote files after evidence-rule preflight failure")
			}
		})
	}
}

const codexCurrentBlock = "<!-- gentle-ai-custom:codex-local-instructions -->\n## Local instructions\n\n- Implement only the requested outcome. Propose optional abstractions or infrastructure separately instead of adding them without approval.\n- When responding, use simple, natural language. Explain what changed, why, and any relevant limits without assuming implementation knowledge.\n- For agent-consumed Markdown, use the smallest clear structure, one H1, shallow headings, short paragraphs, precise paths, explicit stop conditions, and angle-bracket placeholders. Prefer bullets to wide tables.\n- Never read, search, copy, or expose credential files or secret-bearing configuration. Prefer documentation, schemas, examples, and redacted user-run output. If credential access seems necessary, stop and explain why before accessing it.\n- When a Gemini model is active, use skeptical, direct, evidence-first replies; challenge assumptions before validating them and state uncertainty plainly.\n<!-- /gentle-ai-custom:codex-local-instructions -->\n"
const notionKnowledgeBlock = "<!-- gentle-ai-custom:notion-knowledge -->\n## Notion Project Knowledge\n\nUse `notion-project-knowledge` when a request needs knowledge held\nin Notion.\n\nAs the parent agent, also use it to preserve useful project knowledge.\nAt natural progress boundaries, consider Me records for significant\nbehavioral or communicative insight, including subjective impressions\neven without a technical outcome. Technical milestones alone do not\ntrigger a Me record.\n<!-- /gentle-ai-custom:notion-knowledge -->\n"

const legacyNotionKnowledgeBlock = "<!-- gentle-ai-custom:notion-knowledge -->\n## Notion Project Knowledge\n\nUse `notion-project-knowledge` when a request needs knowledge held\nin Notion.\n\nAs the parent agent, also use it to proactively preserve useful\nproject knowledge and personal activity evidence at appropriate\nprogress boundaries.\n<!-- /gentle-ai-custom:notion-knowledge -->\n"

const codexLegacyCommitLine = "- Never commit, push, or create a pull request unless the user explicitly asks. An implementation request, task completion, or review gate does not authorize delivery.\n"

func codexLegacyBlock() string {
	return strings.Replace(codexCurrentBlock, "- When responding,", codexLegacyCommitLine+"- When responding,", 1)
}

func codexMigrationFixture(current, legacy string) string {
	return "upstream prefix\n" + codexODDAssignmentTable + codexUpstreamPersona + "<!-- gentle-ai:engram-protocol -->\nupstream Engram\n<!-- /gentle-ai:engram-protocol -->\n" + current + "\n<!-- gentle-ai:orchestrator -->\nupstream orchestrator\n<!-- /gentle-ai:orchestrator -->\n\n" + legacy + "<!-- gentle-ai:codegraph-guidance -->\nupstream CodeGraph\n"
}

func writeCodexMigrationPrerequisites(t *testing.T, root string) {
	t.Helper()
	mustWriteFile(t, filepath.Join(root, "skills", "judgment-day", "SKILL.md"), []byte("- The only terminal verdicts are `APPROVED | ESCALATED`; never reset or extend an exhausted round budget.\n\n6. Repeat once at most, then run independent final verification and return the terminal verdict.\n\nartifact references, skill resolution, and exactly one final\n"))
	mustWriteFile(t, filepath.Join(root, "skills", "judgment-day", "references", "prompts-and-formats.md"), []byte("Warnings and suggestions are informational. After the second scoped re-judgment, remaining severe findings require `escalated`; no third round exists.\n"))
}

func TestCodexApplyMigratesOnlyRecognizedSecondLocalBlock(t *testing.T) {
	for _, tc := range []struct {
		name, current, legacy string
	}{
		{"legacy and current", "\n" + codexCurrentBlock, codexLegacyBlock() + "\n"},
		{"legacy without current", "", codexLegacyBlock() + "\n"},
		{"current without legacy", "\n" + codexCurrentBlock, ""},
		{"already activated", "\n" + codexCurrentBlock + "\n" + notionKnowledgeBlock, ""},
		{"previous Notion activation", "\n" + codexCurrentBlock + "\n" + legacyNotionKnowledgeBlock, ""},
		{"neither block", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repoRoot := repoRootForTest(t)
			home := t.TempDir()
			t.Setenv("HOME", home)
			root := filepath.Join(home, ".codex")
			path := filepath.Join(root, "AGENTS.md")
			before := codexMigrationFixture(tc.current, tc.legacy)
			mustWriteFile(t, path, []byte(before))
			if err := os.Chmod(path, 0o640); err != nil {
				t.Fatal(err)
			}
			writeCodexMigrationPrerequisites(t, root)
			if code := (&CodexAgent{}).ApplyOverlay(repoRoot, applyPolicyOptions{recorder: newVerboseRecorder(false)}); code != 0 {
				t.Fatalf("Codex apply = %d, want success", code)
			}
			got := mustReadFile(t, path)
			want := codexMigrationFixture("\n"+codexCurrentBlock+"\n"+notionKnowledgeBlock, "")
			want = strings.Replace(want, "| `odd-explorer` | `gpt-6.1-luna` | `high` |", "| `odd-explorer` | `gpt-6.1-sol` | `high` |", 1)
			want = strings.Replace(want, "| `odd-worker` | `gpt-6.1-luna` | `high` |", "| `odd-worker` | `gpt-6.1-sol` | `high` |", 1)
			want = strings.Replace(want, upstreamAgreementRule, localAgreementRule, 1)
			want = strings.Replace(want, upstreamTechnicalRule, localUncertaintyRule, 1)
			if got != want {
				t.Fatalf("Codex AGENTS.md differs beyond declared changes\ngot: %q\nwant: %q", got, want)
			}
			if mode := snapshotTestFile(t, path).mode.Perm(); mode != 0o640 {
				t.Fatalf("Codex AGENTS.md mode = %o, want 640", mode)
			}
			first := snapshotTree(t, root)
			if code := (&CodexAgent{}).ApplyOverlay(repoRoot, applyPolicyOptions{recorder: newVerboseRecorder(false)}); code != 0 {
				t.Fatalf("repeat Codex apply = %d, want success", code)
			}
			if second := snapshotTree(t, root); !equalStringMaps(first, second) {
				t.Fatal("repeat Codex apply changed bytes or modes")
			}
		})
	}
}

func TestCodexApplyRejectsUnexpectedSecondLocalBlockBeforeWrites(t *testing.T) {
	legacy := codexLegacyBlock()
	for _, tc := range []struct{ name, content string }{
		{"duplicate legacy", codexMigrationFixture("\n"+codexCurrentBlock, legacy+"\n"+legacy+"\n")},
		{"partial legacy", codexMigrationFixture("\n"+codexCurrentBlock, strings.Replace(legacy, codexLegacyCommitLine, "", 1)+"\n")},
		{"unknown legacy", codexMigrationFixture("\n"+codexCurrentBlock, strings.Replace(legacy, "- When responding,", "- When writing,", 1)+"\n")},
		{"relocated legacy", codexMigrationFixture("\n"+codexCurrentBlock, "") + "\n" + legacy},
		{"relocated current", codexMigrationFixture("", legacy+"\n") + "\n" + codexCurrentBlock},
		{"unknown Notion activation", codexMigrationFixture("\n"+codexCurrentBlock+"\n"+strings.Replace(legacyNotionKnowledgeBlock, "personal activity evidence", "technical activity evidence", 1), "")},
		{"duplicate Notion activation", codexMigrationFixture("\n"+codexCurrentBlock+"\n"+legacyNotionKnowledgeBlock+legacyNotionKnowledgeBlock, "")},
		{"incomplete Notion activation", codexMigrationFixture("\n"+codexCurrentBlock+"\n"+strings.Replace(legacyNotionKnowledgeBlock, "<!-- /gentle-ai-custom:notion-knowledge -->", "", 1), "")},
		{"relocated Notion activation", codexMigrationFixture("\n"+codexCurrentBlock, "") + "\n" + notionKnowledgeBlock},
		{"incompatible model after legacy", strings.Replace(codexMigrationFixture("\n"+codexCurrentBlock, legacy+"\n"), "| `odd-worker` | `gpt-6.1-luna` | `high` |", "| `odd-worker` | `unknown` | `high` |", 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repoRoot := repoRootForTest(t)
			home := t.TempDir()
			t.Setenv("HOME", home)
			root := filepath.Join(home, ".codex")
			mustWriteFile(t, filepath.Join(root, "AGENTS.md"), []byte(tc.content))
			writeCodexMigrationPrerequisites(t, root)
			before := snapshotTree(t, root)
			if code := (&CodexAgent{}).ApplyOverlay(repoRoot, applyPolicyOptions{recorder: newVerboseRecorder(false)}); code == 0 {
				t.Fatal("Codex apply succeeded with incompatible instructions")
			}
			if after := snapshotTree(t, root); !equalStringMaps(before, after) {
				t.Fatal("Codex apply wrote files after failed preflight")
			}
		})
	}
}

func TestCodexApplyRejectsAmbiguousOrchestratorBoundariesBeforeWrites(t *testing.T) {
	const start = "<!-- gentle-ai:orchestrator -->\n"
	const end = "<!-- /gentle-ai:orchestrator -->\n"
	const codegraph = "<!-- gentle-ai:codegraph-guidance -->\n"
	legacy := codexLegacyBlock() + "\n"
	base := codexMigrationFixture("\n"+codexCurrentBlock, legacy)
	for _, tc := range []struct{ name, content string }{
		{"duplicate orchestrator start", strings.Replace(base, start, start+start, 1)},
		{"duplicate orchestrator end", strings.Replace(base, end, end+end, 1)},
		{"duplicate CodeGraph start", strings.Replace(base, codegraph, codegraph+codegraph, 1)},
		{"orchestrator start after legacy gap", strings.Replace(base, start, "", 1) + start},
		{"decoy gap before real orchestrator", strings.Replace(codexMigrationFixture("\n"+codexCurrentBlock, ""), start, end+"\n"+legacy+codegraph+start, 1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repoRoot := repoRootForTest(t)
			home := t.TempDir()
			t.Setenv("HOME", home)
			root := filepath.Join(home, ".codex")
			mustWriteFile(t, filepath.Join(root, "AGENTS.md"), []byte(tc.content))
			writeCodexMigrationPrerequisites(t, root)
			before := snapshotTree(t, root)
			if code := (&CodexAgent{}).ApplyOverlay(repoRoot, applyPolicyOptions{recorder: newVerboseRecorder(false)}); code == 0 {
				t.Fatal("Codex apply accepted ambiguous legacy block boundaries")
			}
			if after := snapshotTree(t, root); !equalStringMaps(before, after) {
				t.Fatal("Codex apply wrote files after boundary preflight failure")
			}
		})
	}
}

func TestCodexApplyMigratesV4ShapedInstructionsWithoutChangingOtherRows(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := filepath.Join(home, ".codex")
	path := filepath.Join(root, "AGENTS.md")
	before := codexMigrationFixture("\n"+codexCurrentBlock, codexLegacyBlock()+"\n")
	before = strings.Replace(before, "upstream orchestrator\n", "# ODD Orchestrator for Codex\n\n### ODD worker-class assignments\n\nUse the installed routing table for delegated work.\n", 1)
	before = strings.Replace(before, "upstream CodeGraph\n", "## CodeGraph\n\nUse CodeGraph before broad repository searches.\n", 1)
	before = strings.Replace(before, "| `odd-explorer` | `gpt-6.1-luna` | `high` |", "| `odd-explorer` | `gpt-6.1-sol` | `high` |", 1)
	before = strings.Replace(before, "| `odd-worker` | `gpt-6.1-luna` | `high` |", "| `odd-worker` | `gpt-6.1-sol` | `high` |", 1)
	mustWriteFile(t, path, []byte(before))
	if err := os.Chmod(path, 0o640); err != nil {
		t.Fatal(err)
	}
	writeCodexMigrationPrerequisites(t, root)
	if code := (&CodexAgent{}).ApplyOverlay(repoRoot, applyPolicyOptions{recorder: newVerboseRecorder(false)}); code != 0 {
		t.Fatalf("Codex apply = %d, want success", code)
	}
	want := strings.Replace(before, "<!-- /gentle-ai:orchestrator -->\n\n"+codexLegacyBlock()+"\n<!-- gentle-ai:codegraph-guidance -->", "<!-- /gentle-ai:orchestrator -->\n\n<!-- gentle-ai:codegraph-guidance -->", 1)
	want = strings.Replace(want, codexCurrentBlock, codexCurrentBlock+"\n"+notionKnowledgeBlock, 1)
	want = strings.Replace(want, upstreamAgreementRule, localAgreementRule, 1)
	want = strings.Replace(want, upstreamTechnicalRule, localUncertaintyRule, 1)
	if got := mustReadFile(t, path); got != want {
		t.Fatalf("v4-shaped Codex instructions differ beyond the legacy block\ngot: %q\nwant: %q", got, want)
	}
	if mode := snapshotTestFile(t, path).mode.Perm(); mode != 0o640 {
		t.Fatalf("Codex AGENTS.md mode = %o, want 640", mode)
	}
	first := snapshotTree(t, root)
	if code := (&CodexAgent{}).ApplyOverlay(repoRoot, applyPolicyOptions{recorder: newVerboseRecorder(false)}); code != 0 {
		t.Fatalf("repeat Codex apply = %d, want success", code)
	}
	if second := snapshotTree(t, root); !equalStringMaps(first, second) {
		t.Fatal("repeat Codex apply changed v4-shaped fixture bytes or modes")
	}
}

func TestCodexApplyRejectsIncompatibleJudgmentDayAfterLegacyBeforeWrites(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	root := filepath.Join(home, ".codex")
	mustWriteFile(t, filepath.Join(root, "AGENTS.md"), []byte(codexMigrationFixture("\n"+codexCurrentBlock, codexLegacyBlock()+"\n")))
	writeCodexMigrationPrerequisites(t, root)
	judgment := filepath.Join(root, "skills", "judgment-day", "SKILL.md")
	mustWriteFile(t, judgment, []byte("incompatible upstream Judgment Day\n"))
	before := snapshotTree(t, root)
	if code := (&CodexAgent{}).ApplyOverlay(repoRoot, applyPolicyOptions{recorder: newVerboseRecorder(false)}); code == 0 {
		t.Fatal("Codex apply succeeded with incompatible Judgment Day anchor")
	}
	if after := snapshotTree(t, root); !equalStringMaps(before, after) {
		t.Fatal("Codex apply wrote files after Judgment Day preflight failure")
	}
}

func TestRunApplyCustomAllRejectsUnexpectedCodexBlockBeforeSelectedWrites(t *testing.T) {
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
	mustWriteFile(t, openCodeSentinel, []byte("unchanged OpenCode target\n"))
	root := filepath.Join(home, ".codex")
	legacy := strings.Replace(codexLegacyBlock(), "- When responding,", "- When writing,", 1)
	mustWriteFile(t, filepath.Join(root, "AGENTS.md"), []byte(codexMigrationFixture("\n"+codexCurrentBlock, legacy+"\n")))
	writeCodexMigrationPrerequisites(t, root)
	before := snapshotTree(t, home)
	if code := RunApplyCustom(repoRoot, []string{"all"}); code == 0 {
		t.Fatal("RunApplyCustom(all) succeeded with unknown Codex legacy block")
	}
	if after := snapshotTree(t, home); !equalStringMaps(before, after) {
		t.Fatal("selected targets changed after unknown Codex block preflight")
	}
}

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
	agentsBefore := "upstream ODD instructions\n" + codexODDAssignmentTable + codexUpstreamPersona + "<!-- gentle-ai:engram-protocol -->\nupstream Engram body\n<!-- /gentle-ai:engram-protocol -->\nupstream suffix\n"
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
	if !strings.HasPrefix(agentsAfter, "upstream ODD instructions\n| `odd-explorer` | `gpt-6.1-sol` | `high` |\n| `odd-worker` | `gpt-6.1-sol` | `high` |\n| `odd-verify` | `gpt-6.1-sol` | `medium` |\n| `jd-judge-a` | `gpt-6.1-luna` | `high` |\n| `review-risk` | `gpt-6.1-luna` | `high` |\n"+localAgreementRule+localUncertaintyRule+"<!-- gentle-ai:engram-protocol -->\nupstream Engram body\n<!-- /gentle-ai:engram-protocol -->") {
		t.Fatalf("upstream Codex instructions changed: %q", agentsAfter)
	}
	if strings.Contains(agentsAfter, "| `odd-explorer` | `gpt-6.1-luna` | `high` |") || strings.Contains(agentsAfter, "| `odd-worker` | `gpt-6.1-luna` | `high` |") {
		t.Fatal("Codex ODD explorer or worker retained the upstream Luna assignment")
	}
	if !strings.HasSuffix(agentsAfter, "upstream suffix\n") {
		t.Fatalf("upstream Codex suffix changed: %q", agentsAfter)
	}
	for _, marker := range []string{
		"gentle-ai-custom:codex-local-instructions",
		"gentle-ai-custom:notion-knowledge",
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
	if pathExists(upstreamSkillPath) {
		t.Fatal("Codex retained explicitly pruned upstream SDD skill")
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

func TestCodexApplyRejectsIncompatibleODDModelRowsBeforeAnyWrites(t *testing.T) {
	for _, tc := range []struct {
		name string
		rows string
	}{
		{name: "changed explorer", rows: strings.Replace(codexODDAssignmentTable, "| `odd-explorer` | `gpt-6.1-luna` | `high` |", "| `odd-explorer` | `gpt-6.1-other` | `high` |", 1)},
		{name: "missing worker", rows: strings.Replace(codexODDAssignmentTable, "| `odd-worker` | `gpt-6.1-luna` | `high` |\n", "", 1)},
		{name: "duplicate explorer", rows: codexODDAssignmentTable + "| `odd-explorer` | `gpt-6.1-luna` | `high` |\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repoRoot := repoRootForTest(t)
			home := t.TempDir()
			t.Setenv("HOME", home)
			codexRoot := filepath.Join(home, ".codex")
			mustWriteFile(t, filepath.Join(codexRoot, "AGENTS.md"), []byte(tc.rows+codexUpstreamPersona+"<!-- /gentle-ai:engram-protocol -->\n"))
			mustWriteFile(t, filepath.Join(codexRoot, "skills", "judgment-day", "SKILL.md"), []byte("- The only terminal verdicts are `APPROVED | ESCALATED`; never reset or extend an exhausted round budget.\n\n6. Repeat once at most, then run independent final verification and return the terminal verdict.\n\nartifact references, skill resolution, and exactly one final\n"))
			mustWriteFile(t, filepath.Join(codexRoot, "skills", "judgment-day", "references", "prompts-and-formats.md"), []byte("Warnings and suggestions are informational. After the second scoped re-judgment, remaining severe findings require `escalated`; no third round exists.\n"))
			before := snapshotTree(t, codexRoot)
			if code := (&CodexAgent{}).ApplyOverlay(repoRoot, applyPolicyOptions{recorder: newVerboseRecorder(false)}); code == 0 {
				t.Fatal("CodexAgent.ApplyOverlay() = 0, want incompatible model row failure")
			}
			if after := snapshotTree(t, codexRoot); !equalStringMaps(before, after) {
				t.Fatalf("Codex files changed after incompatible model row preflight\nbefore: %v\nafter: %v", before, after)
			}
		})
	}
}

func TestCodexApplyDoesNotRequireSDDMarker(t *testing.T) {
	repoRoot := repoRootForTest(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	codexRoot := filepath.Join(home, ".codex")
	agentsPath := filepath.Join(codexRoot, "AGENTS.md")
	mustWriteFile(t, agentsPath, []byte("upstream ODD instructions without an SDD marker\n"+codexODDAssignmentTable+codexUpstreamPersona+"<!-- /gentle-ai:engram-protocol -->\n"))
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
	if err := os.WriteFile(fakeVersion, []byte("#!/bin/sh\nprintf 'gentle-ai 4.1.1\\n'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	codexRoot := filepath.Join(home, ".codex")
	mustWriteFile(t, filepath.Join(codexRoot, "AGENTS.md"), []byte(codexODDAssignmentTable+codexUpstreamPersona+"<!-- gentle-ai:engram-protocol -->\nupstream Engram\n<!-- /gentle-ai:engram-protocol -->\n"))
	mustWriteFile(t, filepath.Join(codexRoot, "skills", "judgment-day", "SKILL.md"), []byte("- The only terminal verdicts are `APPROVED | ESCALATED`; never reset or extend an exhausted round budget.\n\n6. Repeat once at most, then run independent final verification and return the terminal verdict.\n\nartifact references, skill resolution, and exactly one final\n"))
	mustWriteFile(t, filepath.Join(codexRoot, "skills", "judgment-day", "references", "prompts-and-formats.md"), []byte("Warnings and suggestions are informational. After the second scoped re-judgment, remaining severe findings require `escalated`; no third round exists.\n"))
	mustWriteFile(t, filepath.Join(codexRoot, "skills", "work-unit-commits", "SKILL.md"), []byte("- Every ODD task closes with at least one work-unit commit, branch first when on the default branch, with tests and docs alongside the behavior and a Conventional Commit message.\n"))
	upstreamSDD := filepath.Join(codexRoot, "skills", "sdd-init", "SKILL.md")
	mustWriteFile(t, upstreamSDD, []byte("upstream Codex SDD skill\n"))
	unselectedUpstreamSkill := filepath.Join(codexRoot, "skills", "issue-creation", "SKILL.md")
	mustWriteFile(t, unselectedUpstreamSkill, []byte("unselected Codex skill\n"))
	openCodeSentinel := filepath.Join(home, ".config", "opencode", "skills", "branch-pr", "SKILL.md")
	mustWriteFile(t, openCodeSentinel, []byte("unselected OpenCode sentinel\n"))

	retainedUpstreamSkill := filepath.Join(codexRoot, "skills", "go-testing", "SKILL.md")
	mustWriteFile(t, retainedUpstreamSkill, []byte("retained Codex skill\n"))
	if code := RunApplyCustom(repoRoot, []string{"codex"}); code != 0 {
		t.Fatalf("RunApplyCustom(codex) = %d, want 0", code)
	}
	for _, skill := range []string{"branch-pr", "chained-pr", "code-design", "commit-planner", "judgment-retrospective", "notion-project-knowledge", "package-security", "pr-finalizer"} {
		installed := filepath.Join(codexRoot, "skills", skill, "SKILL.md")
		source := filepath.Join(repoRoot, "shared", "skills", skill, "SKILL.md")
		if got, want := mustReadFile(t, installed), mustReadFile(t, source); got != want {
			t.Errorf("installed Codex skill %q differs from shared source", skill)
		}
	}
	if pathExists(upstreamSDD) || pathExists(unselectedUpstreamSkill) {
		t.Fatal("Codex apply retained explicitly pruned upstream skills")
	}
	if got := mustReadFile(t, retainedUpstreamSkill); got != "retained Codex skill\n" {
		t.Fatalf("unrelated upstream Codex skill changed: %q", got)
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
	mustWriteFile(t, filepath.Join(codexRoot, "AGENTS.md"), []byte(codexODDAssignmentTable+codexUpstreamPersona+"<!-- gentle-ai:engram-protocol -->\nupstream Codex Engram\n<!-- /gentle-ai:engram-protocol -->\n"))
	mustWriteFile(t, filepath.Join(codexRoot, "skills", "judgment-day", "SKILL.md"), []byte("- The only terminal verdicts are `APPROVED | ESCALATED`; never reset or extend an exhausted round budget.\n\n6. Repeat once at most, then run independent final verification and return the terminal verdict.\n\nartifact references, skill resolution, and exactly one final\n"))
	mustWriteFile(t, filepath.Join(codexRoot, "skills", "judgment-day", "references", "prompts-and-formats.md"), []byte("Warnings and suggestions are informational. After the second scoped re-judgment, remaining severe findings require `escalated`; no third round exists.\n"))
	mustWriteFile(t, filepath.Join(codexRoot, "skills", "work-unit-commits", "SKILL.md"), []byte("- Every ODD task closes with at least one work-unit commit, branch first when on the default branch, with tests and docs alongside the behavior and a Conventional Commit message.\n"))
	binDir := filepath.Join(t.TempDir(), "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(binDir, "gentle-ai"), []byte("#!/bin/sh\nprintf 'gentle-ai 4.1.1\\n'\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	if code := RunApplyCustom(repoRoot, []string{"all"}); code != 0 {
		t.Fatalf("RunApplyCustom(all) = %d, want 0", code)
	}
	for _, root := range []string{filepath.Join(home, ".codex"), filepath.Join(home, ".config", "opencode")} {
		for _, skill := range []string{"branch-pr", "notion-project-knowledge"} {
			installed := filepath.Join(root, "skills", skill, "SKILL.md")
			source := filepath.Join(repoRoot, "shared", "skills", skill, "SKILL.md")
			if got, want := mustReadFile(t, installed), mustReadFile(t, source); got != want {
				t.Errorf("%s source mismatch under %s", skill, root)
			}
		}
		content := mustReadFile(t, filepath.Join(root, "AGENTS.md"))
		opening, closing := "<!-- gentle-ai-custom:notion-knowledge -->", "<!-- /gentle-ai-custom:notion-knowledge -->"
		body := mustMarkdownBetweenUniqueMarkers(t, content, opening, closing)
		if opening+body+closing+"\n" != notionKnowledgeBlock {
			t.Errorf("Notion activation differs under %s", root)
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
	first := snapshotTree(t, home)
	if code := RunApplyCustom(repoRoot, []string{"all"}); code != 0 {
		t.Fatalf("repeat RunApplyCustom(all) = %d, want 0", code)
	}
	if second := snapshotTree(t, home); !equalStringMaps(first, second) {
		t.Fatal("repeat all-target apply changed bytes or modes")
	}
}

func chainedCleanupApplyFixture(t *testing.T) (string, string, map[string]string) {
	t.Helper()
	repo := repoRootForTest(t)
	home, _ := writeCompatibleRuntimeFixture(t, repo)
	codexRoot := filepath.Join(home, ".codex")
	mustWriteFile(t, filepath.Join(codexRoot, "AGENTS.md"), []byte(codexODDAssignmentTable+codexUpstreamPersona+"<!-- gentle-ai:engram-protocol -->\nupstream Codex Engram\n<!-- /gentle-ai:engram-protocol -->\n"))
	writeCodexMigrationPrerequisites(t, codexRoot)
	binDir := filepath.Join(t.TempDir(), "bin")
	mustWriteFile(t, filepath.Join(binDir, "gentle-ai"), []byte("#!/bin/sh\nprintf 'gentle-ai 4.1.1\\n'\n"))
	if err := os.Chmod(filepath.Join(binDir, "gentle-ai"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	paths := map[string]string{
		"opencode": filepath.Join(home, ".config", "opencode", "skills", "chained-pr", "references", "chaining-details.md"),
		"codex":    filepath.Join(codexRoot, "skills", "chained-pr", "references", "chaining-details.md"),
		"shared":   filepath.Join(home, ".agents", "skills", "chained-pr", "references", "chaining-details.md"),
	}
	return repo, home, paths
}

func TestApplyChainedReferenceCleanupSelectedTargets(t *testing.T) {
	for _, tc := range []struct {
		name      string
		selected  []string
		populated []string
	}{
		{"opencode only", []string{"opencode"}, []string{"opencode", "codex"}},
		{"codex only", []string{"codex"}, []string{"opencode", "codex"}},
		{"both", []string{"all"}, []string{"opencode", "codex"}},
		{"absent", []string{"all"}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, home, paths := chainedCleanupApplyFixture(t)
			approved := mustReadFile(t, filepath.Join(repo, "overlay", "gentle-ai", "assets", "upstream", "opencode", "skills", "chained-pr", "references", "chaining-details.md"))
			for _, name := range tc.populated {
				mustWriteFile(t, paths[name], []byte(approved))
			}
			mustWriteFile(t, paths["shared"], []byte(approved))
			sharedSkill := filepath.Join(home, ".agents", "skills", "chained-pr", "SKILL.md")
			mustWriteFile(t, sharedSkill, []byte("[references/chaining-details.md](references/chaining-details.md)\n"))
			if code := RunApplyCustom(repo, tc.selected); code != 0 {
				t.Fatalf("apply %v = %d", tc.selected, code)
			}
			for _, name := range []string{"opencode", "codex"} {
				selected := tc.selected[0] == "all" || tc.selected[0] == name
				if selected {
					if _, err := os.Lstat(paths[name]); !os.IsNotExist(err) {
						t.Errorf("selected %s legacy file still present: %v", name, err)
					}
					base := filepath.Join(home, ".codex")
					if name == "opencode" {
						base = filepath.Join(home, ".config", "opencode")
					}
					state, err := loadOwnedRuntimeState(filepath.Join(base, ".gentle-ai-custom-owned-state.json"))
					if err != nil {
						t.Fatal(err)
					}
					if stringSliceContains(state.Paths, paths[name]) {
						t.Errorf("obsolete upstream reference claimed as repo-owned for %s", name)
					}
				} else if len(tc.populated) > 0 && mustReadFile(t, paths[name]) != approved {
					t.Errorf("unselected %s reference changed", name)
				}
			}
			if got := mustReadFile(t, paths["shared"]); got != approved {
				t.Fatal("shared source modified")
			}
			if got := mustReadFile(t, sharedSkill); got != "[references/chaining-details.md](references/chaining-details.md)\n" {
				t.Fatal("shared upstream skill changed")
			}
			if code := RunApplyCustom(repo, tc.selected); code != 0 {
				t.Fatalf("repeat apply %v = %d", tc.selected, code)
			}
		})
	}
}

func TestApplyChainedReferenceCleanupRejectsUnsafeFilesBeforeWrites(t *testing.T) {
	for _, tc := range []struct {
		name, target string
		selected     []string
		seed         func(*testing.T, string, string)
	}{
		{"changed OpenCode file", "opencode", []string{"opencode"}, func(t *testing.T, path, _ string) { mustWriteFile(t, path, []byte("user-authored details\n")) }},
		{"changed Codex file before both", "codex", []string{"all"}, func(t *testing.T, path, _ string) { mustWriteFile(t, path, []byte("user-authored details\n")) }},
		{"symlinked OpenCode file", "opencode", []string{"opencode"}, func(t *testing.T, path, outside string) {
			mustWriteFile(t, outside, []byte("protected user file\n"))
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(outside, path); err != nil {
				t.Skipf("symlink unavailable: %v", err)
			}
		}},
		{"directory at Codex path", "codex", []string{"codex"}, func(t *testing.T, path, _ string) {
			if err := os.MkdirAll(path, 0o755); err != nil {
				t.Fatal(err)
			}
			mustWriteFile(t, filepath.Join(path, "user.txt"), []byte("protected user file\n"))
		}},
		{"symlinked Codex parent", "codex", []string{"all"}, func(t *testing.T, path, outside string) {
			mustWriteFile(t, outside, []byte("protected user file\n"))
			if err := os.MkdirAll(filepath.Dir(filepath.Dir(path)), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Dir(outside), filepath.Dir(path)); err != nil {
				t.Skipf("symlink unavailable: %v", err)
			}
		}},
		{"symlinked selected skill", "codex", []string{"codex"}, func(t *testing.T, path, outside string) {
			mustWriteFile(t, path, []byte(mustReadFile(t, filepath.Join(repoRootForTest(t), filepath.FromSlash(chainedApprovedReference)))))
			mustWriteFile(t, outside, []byte("protected user skill\n"))
			if err := os.Symlink(outside, filepath.Join(filepath.Dir(filepath.Dir(path)), "SKILL.md")); err != nil {
				t.Skipf("symlink unavailable: %v", err)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, home, paths := chainedCleanupApplyFixture(t)
			outside := filepath.Join(t.TempDir(), "chaining-details.md")
			tc.seed(t, paths[tc.target], outside)
			configPath := filepath.Join(home, ".config", "opencode", "opencode.json")
			codexPath := filepath.Join(home, ".codex", "AGENTS.md")
			beforeConfig, beforeCodex := mustReadFile(t, configPath), mustReadFile(t, codexPath)
			if code := RunApplyCustom(repo, tc.selected); code == 0 {
				t.Fatal("unsafe residue accepted by apply")
			}
			if got := mustReadFile(t, configPath); got != beforeConfig {
				t.Fatal("OpenCode config changed before cleanup refusal")
			}
			if got := mustReadFile(t, codexPath); got != beforeCodex {
				t.Fatal("Codex AGENTS changed before cleanup refusal")
			}
			if _, err := os.Lstat(paths[tc.target]); err != nil {
				t.Fatalf("unsafe path was removed: %v", err)
			}
			if tc.name == "symlinked OpenCode file" || tc.name == "symlinked Codex parent" || tc.name == "symlinked selected skill" {
				want := "protected user file\n"
				if tc.name == "symlinked selected skill" {
					want = "protected user skill\n"
				}
				if got := mustReadFile(t, outside); got != want {
					t.Fatalf("outside user file changed: %q", got)
				}
			}
		})
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

func TestCodexApplyRemovesKnownRemoteAuthorizationWithoutChangingOtherInstructions(t *testing.T) {
	for _, tc := range []struct {
		name, section string
	}{
		{"upstream", retiredRemoteAuthorizationSections[0]},
		{"retired relaxation", retiredRemoteAuthorizationSections[1]},
		{"already absent", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := repoRootForTest(t)
			home := t.TempDir()
			t.Setenv("HOME", home)
			root := filepath.Join(home, ".codex")
			path := filepath.Join(root, "AGENTS.md")
			before := codexMigrationFixture("", "")
			if tc.section != "" {
				before += "\n<!-- gentle-ai:remote-authorization -->\n" + tc.section + "\n<!-- /gentle-ai:remote-authorization -->\n"
			}
			mustWriteFile(t, path, []byte(before))
			writeCodexMigrationPrerequisites(t, root)
			if code := (&CodexAgent{}).ApplyOverlay(repo, applyPolicyOptions{recorder: newVerboseRecorder(false)}); code != 0 {
				t.Fatalf("Codex apply = %d, want success", code)
			if err := os.Chmod(path, 0o640); err != nil {
				t.Fatal(err)
			}
			}
			got := mustReadFile(t, path)
			if strings.Contains(got, "gentle-ai:remote-authorization") || strings.Contains(got, "## Remote operation authorization") {
				t.Fatal("Codex retained remote authorization after apply")
			}
			want := codexMigrationFixture("\n"+codexCurrentBlock+"\n"+notionKnowledgeBlock, "")
			want = strings.Replace(want, "| `odd-explorer` | `gpt-6.1-luna` | `high` |", "| `odd-explorer` | `gpt-6.1-sol` | `high` |", 1)
			want = strings.Replace(want, "| `odd-worker` | `gpt-6.1-luna` | `high` |", "| `odd-worker` | `gpt-6.1-sol` | `high` |", 1)
			want = strings.Replace(want, upstreamAgreementRule, localAgreementRule, 1)
			want = strings.Replace(want, upstreamTechnicalRule, localUncertaintyRule, 1)
			if tc.section != "" {
				want += "\n\n"
			}
			if got != want || snapshotTestFile(t, path).mode.Perm() != 0o640 {
				t.Fatalf("Codex apply changed unrelated bytes or file mode\ngot: %q\nwant: %q", got, want)
			}
			first := snapshotTree(t, root)
			if code := (&CodexAgent{}).ApplyOverlay(repo, applyPolicyOptions{recorder: newVerboseRecorder(false)}); code != 0 || !equalStringMaps(first, snapshotTree(t, root)) {
				t.Fatal("repeated Codex apply changed runtime contents")
			}
		})
	}
}

func TestCodexApplyRejectsUnknownRemoteAuthorizationBeforeWrites(t *testing.T) {
	for _, section := range []string{
		strings.Replace(retiredRemoteAuthorizationSections[0], "Before remote work", "Before any work", 1),
		retiredRemoteAuthorizationSections[0] + "\n\n" + retiredRemoteAuthorizationSections[0],
	} {
		repo := repoRootForTest(t)
		home := t.TempDir()
		t.Setenv("HOME", home)
		root := filepath.Join(home, ".codex")
		mustWriteFile(t, filepath.Join(root, "AGENTS.md"), []byte(codexMigrationFixture("", "")+"\n<!-- gentle-ai:remote-authorization -->\n"+section+"\n<!-- /gentle-ai:remote-authorization -->\n"))
		writeCodexMigrationPrerequisites(t, root)
		before := snapshotTree(t, root)
		if code := (&CodexAgent{}).ApplyOverlay(repo, applyPolicyOptions{recorder: newVerboseRecorder(false)}); code == 0 {
			t.Fatal("Codex accepted unknown remote authorization section")
		}
		if after := snapshotTree(t, root); !equalStringMaps(before, after) {
			t.Fatal("Codex wrote files after remote authorization preflight failure")
		}
	}
}

func TestRunApplyCustomCodexRemovesRemoteAuthorization(t *testing.T) {
	repo := repoRootForTest(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	binDir := filepath.Join(t.TempDir(), "bin")
	mustWriteFile(t, filepath.Join(binDir, "gentle-ai"), []byte("#!/bin/sh\nprintf 'gentle-ai 4.1.1\\n'\n"))
	if err := os.Chmod(filepath.Join(binDir, "gentle-ai"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	root := filepath.Join(home, ".codex")
	path := filepath.Join(root, "AGENTS.md")
	mustWriteFile(t, path, []byte(codexMigrationFixture("", "")+"\n<!-- gentle-ai:remote-authorization -->\n"+retiredRemoteAuthorizationSections[0]+"\n<!-- /gentle-ai:remote-authorization -->\n"))
	writeCodexMigrationPrerequisites(t, root)
	if code := RunApplyCustom(repo, []string{"codex"}); code != 0 {
		t.Fatalf("apply-custom codex returned %d, want 0", code)
	}
	if content := mustReadFile(t, path); strings.Contains(content, "## Remote operation authorization") || !strings.Contains(content, "upstream CodeGraph\n") {
		t.Fatal("public Codex apply retained remote authorization or changed unrelated instructions")
	}
	first := snapshotTree(t, root)
	if code := RunApplyCustom(repo, []string{"codex"}); code != 0 || !equalStringMaps(first, snapshotTree(t, root)) {
		t.Fatal("repeat public Codex apply changed runtime contents")
	}
}

func TestRunApplyCustomAllRejectsUnsafeCodexPruneBeforeOpenCodeWrites(t *testing.T) {
	repo := repoRootForTest(t)
	home, openCodeConfig := writeCompatibleRuntimeFixture(t, repo)
	root := filepath.Join(home, ".codex")
	mustWriteFile(t, filepath.Join(root, "AGENTS.md"), []byte(codexMigrationFixture("", "")))
	writeCodexMigrationPrerequisites(t, root)
	outside := filepath.Join(t.TempDir(), "outside-skill")
	mustWriteFile(t, filepath.Join(outside, "SKILL.md"), []byte("private outside skill\n"))
	link := filepath.Join(root, "skills", "issue-creation")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	beforeConfig := mustReadFile(t, openCodeConfig)
	beforeCodex := mustReadFile(t, filepath.Join(root, "AGENTS.md"))
	beforeLink, err := os.Readlink(link)
	if err != nil {
		t.Fatal(err)
	}
	if code := RunApplyCustom(repo, []string{"all"}); code == 0 {
		t.Fatal("unsafe Codex prune path was accepted")
	}
	if mustReadFile(t, openCodeConfig) != beforeConfig || mustReadFile(t, filepath.Join(root, "AGENTS.md")) != beforeCodex {
		t.Fatal("selected target changed before Codex prune preflight refusal")
	}
	if afterLink, err := os.Readlink(link); err != nil || afterLink != beforeLink {
		t.Fatalf("Codex skill link changed after preflight refusal: %q, %v", afterLink, err)
	}
	if got := mustReadFile(t, filepath.Join(outside, "SKILL.md")); got != "private outside skill\n" {
		t.Fatalf("outside skill changed: %q", got)
	}
}
