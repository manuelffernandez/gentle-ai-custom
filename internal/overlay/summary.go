package overlay

import "fmt"

// applyMetrics holds all counter and accumulated-list fields whose sole reason
// to change is "reporting output changed". It is embedded inside applyPolicyState
// as a named field so the two concerns are clearly separated.
type applyMetrics struct {
	prunedCount          int
	missingKeepSummary   []string
	createdOverrides     []string
	ownedAssetWrites     writeCounters
	promptRefsUpdated    int
	promptRefsUnchanged  int
	ownedRuntimeTargetsVerified int
	topologyWarnings     []string
	profilesManagedCount int
	profileAgentsCreated int
	profileAgentsUpdated int
	profileAgentsSame    int
	unmanagedProfiles    []string
}

// printSummary prints the full apply-policy run summary, including counters,
// warnings, and actionable notes for the operator.
func (s *applyPolicyState) printSummary() {
	configStatus := "unchanged"
	if s.configChanged {
		configStatus = "updated"
	}
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  OpenCode config status: %s\n", configStatus)
	fmt.Printf("  skills pruned this run: %d\n", s.metrics.prunedCount)
	fmt.Printf("  owned asset writes - new: %d, changed: %d, unchanged: %d, deleted: %d\n", s.metrics.ownedAssetWrites.New, s.metrics.ownedAssetWrites.Changed, s.metrics.ownedAssetWrites.Unchanged, s.metrics.ownedAssetWrites.Deleted)
	fmt.Printf("  prompt references updated: %d\n", s.metrics.promptRefsUpdated)
	fmt.Printf("  prompt references unchanged: %d\n", s.metrics.promptRefsUnchanged)
	fmt.Printf("  owned runtime targets verified: %d\n", s.metrics.ownedRuntimeTargetsVerified)
	fmt.Printf("  topology warnings: %d\n", len(s.metrics.topologyWarnings))
	fmt.Printf("  SDD profiles managed this run: %d\n", s.metrics.profilesManagedCount)
	fmt.Printf("  SDD profile agents created: %d\n", s.metrics.profileAgentsCreated)
	fmt.Printf("  SDD profile agents updated: %d\n", s.metrics.profileAgentsUpdated)
	fmt.Printf("  SDD profile agents unchanged: %d\n", s.metrics.profileAgentsSame)
	fmt.Printf("  SDD profiles unmanaged (present in opencode.json, absent from local config): %d\n", len(s.metrics.unmanagedProfiles))

	if len(s.metrics.unmanagedProfiles) > 0 {
		fmt.Println()
		fmt.Println("WARNING - unmanaged SDD profiles left untouched (add them to profiles[] in the local OpenCode overlay config to manage):")
		for _, entry := range s.metrics.unmanagedProfiles {
			fmt.Printf("  - %s\n", entry)
		}
	}
	if len(s.metrics.missingKeepSummary) > 0 {
		fmt.Println()
		fmt.Println("WARNING - keep skills missing (expected but absent):")
		for _, entry := range s.metrics.missingKeepSummary {
			fmt.Printf("  - %s\n", entry)
		}
	}
	if len(s.metrics.topologyWarnings) > 0 {
		fmt.Println()
		fmt.Println("NOTE: topology drift detected. Review the topology: warnings above and update policy/intent if needed.")
	}
	fmt.Println()
	fmt.Println("Done. Restart OpenCode if opencode.json changed.")
}

// printMissingOpenCodeSummary prints an abbreviated summary used when the
// OpenCode config file is absent (only skill pruning ran).
func (s *applyPolicyState) printMissingOpenCodeSummary() {
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  skills pruned this run: %d\n", s.metrics.prunedCount)
	fmt.Printf("  owned asset writes - new: %d, changed: %d, unchanged: %d, deleted: %d\n", s.metrics.ownedAssetWrites.New, s.metrics.ownedAssetWrites.Changed, s.metrics.ownedAssetWrites.Unchanged, s.metrics.ownedAssetWrites.Deleted)
	fmt.Printf("  owned runtime targets verified: %d\n", s.metrics.ownedRuntimeTargetsVerified)
	if len(s.metrics.missingKeepSummary) > 0 {
		fmt.Println("  WARNING - keep skills missing (expected but absent):")
		for _, entry := range s.metrics.missingKeepSummary {
			fmt.Printf("    - %s\n", entry)
		}
	}
	fmt.Println("Done.")
}
