package overlay

import "fmt"

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
	fmt.Printf("  skills pruned this run: %d\n", s.prunedCount)
	fmt.Printf("  orchestrators generated (fresh): %d\n", s.generatedCount)
	fmt.Printf("  orchestrators recovered from snapshot: %d\n", s.recoveredCount)
	fmt.Printf("  orchestrators kept (already applied): %d\n", s.keptCount)
	fmt.Printf("  orchestrators skipped: %d\n", s.skippedCount)
	fmt.Printf("  repo snapshots - new: %d, changed: %d, unchanged: %d\n", s.repoSnapshots.New, s.repoSnapshots.Changed, s.repoSnapshots.Unchanged)
	fmt.Printf("  local snapshots - new: %d, changed: %d, unchanged: %d\n", s.localSnapshots.New, s.localSnapshots.Changed, s.localSnapshots.Unchanged)
	fmt.Printf("  local snapshot migrations from repo: %d\n", s.localSnapshotMigrate)
	fmt.Printf("  repo snapshot backfills from local: %d\n", s.repoSnapshotBackfill)
	fmt.Printf("  topology warnings: %d\n", len(s.topologyWarnings))
	fmt.Printf("  SDD profiles in local config: %d\n", s.profilesManagedCount)
	fmt.Printf("  SDD profile agents created: %d\n", s.profileAgentsCreated)
	fmt.Printf("  SDD profile agents updated: %d\n", s.profileAgentsUpdated)
	fmt.Printf("  SDD profile agents unchanged: %d\n", s.profileAgentsSame)
	fmt.Printf("  SDD profiles unmanaged (present in opencode.json, absent from local config): %d\n", len(s.unmanagedProfiles))
	fmt.Println("  audited base baseline verification: ok")

	if len(s.unmanagedProfiles) > 0 {
		fmt.Println()
		fmt.Println("WARNING - unmanaged SDD profiles left untouched (add them to the local SDD profile config to manage):")
		for _, entry := range s.unmanagedProfiles {
			fmt.Printf("  - %s\n", entry)
		}
	}
	if len(s.missingKeepSummary) > 0 {
		fmt.Println()
		fmt.Println("WARNING - keep skills missing (expected but absent):")
		for _, entry := range s.missingKeepSummary {
			fmt.Printf("  - %s\n", entry)
		}
	}
	if s.repoSnapshots.Changed > 0 {
		fmt.Println()
		fmt.Println("NOTE: versioned orchestrator snapshots drifted. Review with:")
		fmt.Println("  git diff overlay/gentle-ai/snapshots/")
	}
	if s.localSnapshots.Changed > 0 {
		fmt.Println()
		fmt.Println("NOTE: local operational orchestrator snapshots drifted under:")
		fmt.Printf("  %s\n", s.localSnapshotDir)
	}
	if s.localSnapshotMigrate > 0 {
		fmt.Println()
		fmt.Printf("NOTE: migrated %d legacy snapshot(s) from the repo into the local operational snapshot dir.\n", s.localSnapshotMigrate)
	}
	if s.repoSnapshotBackfill > 0 {
		fmt.Println()
		fmt.Printf("NOTE: backfilled %d versioned repo snapshot(s) from local operational snapshots.\n", s.repoSnapshotBackfill)
	}
	if s.recoveredCount > 0 {
		fmt.Println()
		fmt.Printf("NOTE: %d orchestrator(s) recovered from snapshot.\n", s.recoveredCount)
		fmt.Println("  The snapshot content may pre-date the current upstream version.")
		fmt.Println("  Run `gentle-ai sync` then re-run this script to capture fresh upstream.")
	}
	if len(s.topologyWarnings) > 0 {
		fmt.Println()
		fmt.Println("NOTE: topology drift detected. Review the topology: warnings above and update policy/intent if needed.")
	}
	s.recorder.print()
	fmt.Println()
	fmt.Println("Done. Restart OpenCode if opencode.json changed.")
}

// printMissingOpenCodeSummary prints an abbreviated summary used when the
// OpenCode config file is absent (only skill pruning ran).
func (s *applyPolicyState) printMissingOpenCodeSummary() {
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  skills pruned this run: %d\n", s.prunedCount)
	if len(s.missingKeepSummary) > 0 {
		fmt.Println("  WARNING - keep skills missing (expected but absent):")
		for _, entry := range s.missingKeepSummary {
			fmt.Printf("    - %s\n", entry)
		}
	}
	s.recorder.print()
	fmt.Println("Done.")
}
