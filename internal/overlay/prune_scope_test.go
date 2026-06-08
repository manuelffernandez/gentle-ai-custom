package overlay

import (
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

func setTestAgentRegistry(t *testing.T, registry map[string]registeredAgent) {
	t.Helper()
	original := agentRegistry
	agentRegistry = registry
	t.Cleanup(func() {
		agentRegistry = original
	})
}
