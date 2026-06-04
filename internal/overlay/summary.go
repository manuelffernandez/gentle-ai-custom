package overlay

import "fmt"

// applyMetrics holds all counter and accumulated-list fields whose sole reason
// to change is "reporting output changed". It is embedded inside applyPolicyState
// as a named field so the two concerns are clearly separated.
type applyMetrics struct {
	prunedCount          int
	missingKeepSummary   []string
	createdOverrides     []string
	generatedCount       int
	recoveredCount       int
	keptCount            int
	skippedCount         int
	repoSnapshots        snapshotCounters
	localSnapshots       snapshotCounters
	localSnapshotMigrate int
	repoSnapshotBackfill int
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
	fmt.Printf("  orchestrators generated (fresh): %d\n", s.metrics.generatedCount)
	fmt.Printf("  orchestrators recovered from snapshot: %d\n", s.metrics.recoveredCount)
	fmt.Printf("  orchestrators kept (already applied): %d\n", s.metrics.keptCount)
	fmt.Printf("  orchestrators skipped: %d\n", s.metrics.skippedCount)
	fmt.Printf("  repo snapshots - new: %d, changed: %d, unchanged: %d\n", s.metrics.repoSnapshots.New, s.metrics.repoSnapshots.Changed, s.metrics.repoSnapshots.Unchanged)
	fmt.Printf("  local snapshots - new: %d, changed: %d, unchanged: %d\n", s.metrics.localSnapshots.New, s.metrics.localSnapshots.Changed, s.metrics.localSnapshots.Unchanged)
	fmt.Printf("  local snapshot migrations from repo: %d\n", s.metrics.localSnapshotMigrate)
	fmt.Printf("  repo snapshot backfills from local: %d\n", s.metrics.repoSnapshotBackfill)
	fmt.Printf("  topology warnings: %d\n", len(s.metrics.topologyWarnings))
	fmt.Printf("  SDD profiles managed this run: %d\n", s.metrics.profilesManagedCount)
	fmt.Printf("  SDD profile agents created: %d\n", s.metrics.profileAgentsCreated)
	fmt.Printf("  SDD profile agents updated: %d\n", s.metrics.profileAgentsUpdated)
	fmt.Printf("  SDD profile agents unchanged: %d\n", s.metrics.profileAgentsSame)
	fmt.Printf("  SDD profiles unmanaged (present in opencode.json, absent from local config): %d\n", len(s.metrics.unmanagedProfiles))
	fmt.Println("  audited base baseline verification: ok")

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
	if s.metrics.repoSnapshots.Changed > 0 {
		fmt.Println()
		fmt.Println("NOTE: versioned orchestrator snapshots drifted. Review with:")
		fmt.Println("  git diff overlay/gentle-ai/snapshots/")
	}
	if s.metrics.localSnapshots.Changed > 0 {
		fmt.Println()
		fmt.Println("NOTE: local operational orchestrator snapshots drifted under:")
		fmt.Printf("  %s\n", s.localSnapshotDir)
	}
	if s.metrics.localSnapshotMigrate > 0 {
		fmt.Println()
		fmt.Printf("NOTE: migrated %d legacy snapshot(s) from the repo into the local operational snapshot dir.\n", s.metrics.localSnapshotMigrate)
	}
	if s.metrics.repoSnapshotBackfill > 0 {
		fmt.Println()
		fmt.Printf("NOTE: backfilled %d versioned repo snapshot(s) from local operational snapshots.\n", s.metrics.repoSnapshotBackfill)
	}
	if s.metrics.recoveredCount > 0 {
		fmt.Println()
		fmt.Printf("NOTE: %d orchestrator(s) recovered from snapshot.\n", s.metrics.recoveredCount)
		fmt.Println("  The snapshot content may pre-date the current upstream version.")
		fmt.Println("  Run `gentle-ai sync` then re-run this script to capture fresh upstream.")
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
	if len(s.metrics.missingKeepSummary) > 0 {
		fmt.Println("  WARNING - keep skills missing (expected but absent):")
		for _, entry := range s.metrics.missingKeepSummary {
			fmt.Printf("    - %s\n", entry)
		}
	}
	fmt.Println("Done.")
}
