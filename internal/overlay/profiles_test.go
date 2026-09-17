package overlay

import "testing"

func TestReconcileProfileAgentAccountsForPromptReferences(t *testing.T) {
	tests := []struct {
		name          string
		existing      map[string]any
		wantUpdated   int
		wantUnchanged int
	}{
		{
			name:        "created profile agent assigns upstream prompt",
			wantUpdated: 1,
		},
		{
			name:          "stale profile prompt is updated",
			existing:      map[string]any{"model": "profile/model", "prompt": "stale prompt"},
			wantUpdated:   1,
			wantUnchanged: 0,
		},
		{
			name:          "matching profile prompt is unchanged",
			existing:      map[string]any{"model": "profile/model", "prompt": "upstream prompt"},
			wantUpdated:   0,
			wantUnchanged: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agents := map[string]any{
				"sdd-init": map[string]any{"model": "upstream/model", "prompt": "upstream prompt"},
			}
			if tt.existing != nil {
				agents["sdd-init-cheap"] = tt.existing
			}
			state := applyPolicyState{
				agents:             agents,
				expectedPromptRefs: map[string]string{},
				recorder:           newVerboseRecorder(false),
			}
			profile := validatedProfile{Name: "cheap"}

			if err := state.reconcileProfileAgent(profile, "sdd-init-cheap", profileAssignment{Model: "profile/model"}, false); err != nil {
				t.Fatalf("reconcileProfileAgent() error = %v", err)
			}

			if state.metrics.promptRefsUpdated != tt.wantUpdated {
				t.Fatalf("promptRefsUpdated = %d, want %d", state.metrics.promptRefsUpdated, tt.wantUpdated)
			}
			if state.metrics.promptRefsUnchanged != tt.wantUnchanged {
				t.Fatalf("promptRefsUnchanged = %d, want %d", state.metrics.promptRefsUnchanged, tt.wantUnchanged)
			}
			if got := jsonString(mustAgent(t, state.agents, "sdd-init-cheap")["prompt"]); got != "upstream prompt" {
				t.Fatalf("profile prompt = %q, want upstream prompt", got)
			}
			if got := state.expectedPromptRefs["sdd-init-cheap"]; got != "upstream prompt" {
				t.Fatalf("expected prompt reference = %q, want upstream prompt", got)
			}
		})
	}
}
