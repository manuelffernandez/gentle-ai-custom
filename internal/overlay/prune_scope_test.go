package overlay

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type fakeAgent struct {
	name string
}

func (a fakeAgent) Name() string                                                 { return a.name }
func (a fakeAgent) BasePath() (string, error)                                    { return "", nil }
func (a fakeAgent) BuildCommandContent(cmd customCommand, body string) string    { return body }
func (a fakeAgent) ApplyOverlay(repoRoot string, options applyPolicyOptions) int { return 0 }

func TestRegisteredSkillTargetsFollowSelectedRegisteredAgents(t *testing.T) {
	setTestAgentRegistry(t, map[string]registeredAgent{
		"claude": {
			agent:        fakeAgent{name: "claude"},
			skillTargets: []string{"~/.claude/skills"},
		},
		"opencode": {
			agent:        fakeAgent{name: "opencode"},
			skillTargets: []string{"~/.config/opencode/skills"},
		},
	})

	if got, want := registeredSkillTargets([]string{"opencode"}), []string{"~/.config/opencode/skills"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("registeredSkillTargets(opencode) = %v, want %v", got, want)
	}

	if got, want := registeredSkillTargets(registeredAgentNames()), []string{"~/.claude/skills", "~/.config/opencode/skills"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("registeredSkillTargets(all) = %v, want %v", got, want)
	}
}

func TestPruneSkillsUsesSelectedRegisteredTargetsOnly(t *testing.T) {
	setTestAgentRegistry(t, map[string]registeredAgent{
		"claude": {
			agent:        fakeAgent{name: "claude"},
			skillTargets: []string{"~/.claude/skills"},
		},
		"opencode": {
			agent:        fakeAgent{name: "opencode"},
			skillTargets: []string{"~/.config/opencode/skills"},
		},
	})

	tests := []struct {
		name       string
		selected   []string
		wantClaude bool
	}{
		{name: "selected opencode only", selected: []string{"opencode"}, wantClaude: true},
		{name: "all registered agents", selected: registeredAgentNames(), wantClaude: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			homeDir := t.TempDir()
			t.Setenv("HOME", homeDir)

			opencodeSkillsDir := filepath.Join(homeDir, ".config", "opencode", "skills")
			claudeSkillsDir := filepath.Join(homeDir, ".claude", "skills")
			for _, dir := range []string{opencodeSkillsDir, claudeSkillsDir} {
				if err := os.MkdirAll(filepath.Join(dir, "branch-pr"), 0o755); err != nil {
					t.Fatalf("mkdir skill dir %s: %v", dir, err)
				}
			}

			policy := Policy{}
			policy.Skills.Prune = []string{"branch-pr"}

			state := applyPolicyState{
				policy:       policy,
				recorder:     newVerboseRecorder(false),
				skillTargets: registeredSkillTargets(tc.selected),
			}
			if err := state.pruneSkills(); err != nil {
				t.Fatalf("pruneSkills() error = %v", err)
			}

			if pathExists(filepath.Join(opencodeSkillsDir, "branch-pr")) {
				t.Fatalf("selected opencode skills were not pruned")
			}
			if tc.wantClaude {
				if !pathExists(filepath.Join(claudeSkillsDir, "branch-pr")) {
					t.Fatalf("unselected Claude skills were touched")
				}
			} else if pathExists(filepath.Join(claudeSkillsDir, "branch-pr")) {
				t.Fatalf("registered Claude skills were not pruned by all")
			}
		})
	}
}

func TestStandaloneApplyPolicyDoesNotPruneCodexSkills(t *testing.T) {
	repo := repoRootForTest(t)
	home, _ := writeCompatibleRuntimeFixture(t, repo)
	codexSkill := filepath.Join(home, ".codex", "skills", "issue-creation", "SKILL.md")
	mustWriteFile(t, codexSkill, []byte("Codex skill must remain untouched\n"))
	openCodeSkill := filepath.Join(home, ".config", "opencode", "skills", "issue-creation", "SKILL.md")
	mustWriteFile(t, openCodeSkill, []byte("OpenCode skill is pruned\n"))
	if code := RunApplyPolicy(repo, nil); code != 0 {
		t.Fatalf("standalone apply-policy returned %d, want 0", code)
	}
	if pathExists(openCodeSkill) {
		t.Fatal("standalone apply-policy did not prune OpenCode")
	}
	if got := mustReadFile(t, codexSkill); got != "Codex skill must remain untouched\n" {
		t.Fatalf("standalone apply-policy changed unselected Codex skill: %q", got)
	}
}

func TestPruneSkillsReportsOnlyActualRemovals(t *testing.T) {
	tests := []struct {
		name    string
		present []string
	}{
		{name: "all pruned skills absent"},
		{name: "one present and two absent", present: []string{"branch-pr"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			target := filepath.Join(t.TempDir(), "skills")
			if err := os.MkdirAll(target, 0o755); err != nil {
				t.Fatal(err)
			}
			for _, skill := range tc.present {
				if err := os.MkdirAll(filepath.Join(target, skill), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			policy := Policy{}
			policy.Skills.Prune = []string{"branch-pr", "sdd-init", "issue-creation"}
			policy.Skills.Keep = []string{"retained-skill"}
			state := applyPolicyState{
				policy:       policy,
				recorder:     newVerboseRecorder(false),
				skillTargets: []string{target},
			}
			output := capturePruneSkillsOutput(t, func() {
				if err := state.pruneSkills(); err != nil {
					t.Fatalf("pruneSkills() error = %v", err)
				}
			})
			wantOutput := fmt.Sprintf("- pruning unwanted skills in %s\n", target)
			if len(tc.present) > 0 {
				wantOutput += "  removed branch-pr\n"
			}
			if output != wantOutput {
				t.Errorf("pruneSkills() output = %q, want %q", output, wantOutput)
			}
			if state.metrics.prunedCount != len(tc.present) {
				t.Errorf("skills pruned = %d, want %d", state.metrics.prunedCount, len(tc.present))
			}
			if want := []string{target + " -> retained-skill"}; !reflect.DeepEqual(state.metrics.missingKeepSummary, want) {
				t.Errorf("missing keep summary = %v, want %v", state.metrics.missingKeepSummary, want)
			}
			for _, skill := range policy.Skills.Prune {
				if pathExists(filepath.Join(target, skill)) {
					t.Errorf("pruned skill %q remains in selected target", skill)
				}
			}
		})
	}
}

func capturePruneSkillsOutput(t *testing.T, run func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = writer
	t.Cleanup(func() {
		os.Stdout = previous
		reader.Close()
		writer.Close()
	})
	run()
	os.Stdout = previous
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return string(output)
}

func setTestAgentRegistry(t *testing.T, registry map[string]registeredAgent) {
	t.Helper()
	original := agentRegistry
	agentRegistry = registry
	t.Cleanup(func() {
		agentRegistry = original
	})
}
