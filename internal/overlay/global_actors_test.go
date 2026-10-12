package overlay

import "testing"

func TestReconcileGlobalActorsLeavesUnrelatedAgentsUntouched(t *testing.T) {
	state := applyPolicyState{
		configPath: "opencode.json",
		resolvedGlobalActors: map[string]actorAssignment{
			"jd-judge-a":  {Model: "openai/judge", Variant: "high"},
			"review-risk": {Inherit: true},
		},
		agents: map[string]any{
			"jd-judge-a":  map[string]any{"model": "old", "variant": "low"},
			"review-risk": map[string]any{"model": "old", "variant": "low"},
			"general":     map[string]any{"model": "custom/model", "sentinel": "keep"},
		},
		recorder: newVerboseRecorder(false),
	}

	state.reconcileGlobalActors()

	judge := mustAgent(t, state.agents, "jd-judge-a")
	if jsonString(judge["model"]) != "openai/judge" || jsonString(judge["variant"]) != "high" {
		t.Fatalf("Judgment Day assignment = %#v", judge)
	}
	reviewer := mustAgent(t, state.agents, "review-risk")
	if _, exists := reviewer["model"]; exists {
		t.Fatalf("inherited reviewer still has model: %#v", reviewer)
	}
	if _, exists := reviewer["variant"]; exists {
		t.Fatalf("inherited reviewer still has variant: %#v", reviewer)
	}
	general := mustAgent(t, state.agents, "general")
	if jsonString(general["model"]) != "custom/model" || jsonString(general["sentinel"]) != "keep" {
		t.Fatalf("unrelated agent changed: %#v", general)
	}
	if state.configChanged != true {
		t.Fatal("global actor reconciliation did not mark config changed")
	}
}
