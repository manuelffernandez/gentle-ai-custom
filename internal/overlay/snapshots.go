package overlay

import (
	"fmt"
	"os"
	"strings"
)

type snapshotCounters struct {
	New       int
	Changed   int
	Unchanged int
}

// safeSnapshotKey validates that an agent key is safe to use as a file name
// component (no path separators, null bytes, or empty string).
func safeSnapshotKey(key string) (string, error) {
	if key == "" || strings.Contains(key, "/") || strings.Contains(key, "\\") || strings.Contains(key, "..") || strings.ContainsRune(key, '\x00') {
		return "", fmt.Errorf("unsafe agent key for snapshot path: %q", key)
	}
	return key, nil
}

// writeSnapshotWithStatus writes content to path and returns a status string
// ("new", "changed", or "unchanged") plus updates the given counters.
func writeSnapshotWithStatus(path, content string, counters *snapshotCounters) (string, error) {
	normalized := normalizeLFTerminated(content)
	status := "new"
	if pathExists(path) {
		existing, err := readText(path)
		if err != nil {
			return "", err
		}
		if normalizeLFTerminated(existing) == normalized {
			status = "unchanged"
		} else {
			status = "changed"
		}
	}
	switch status {
	case "new":
		counters.New++
	case "changed":
		counters.Changed++
	case "unchanged":
		counters.Unchanged++
	}
	return status, writeTextFile(path, content)
}

// migrateRepoSnapshotToLocal copies a versioned repo snapshot to the local
// operational snapshot directory when the local copy is absent. No-op if the
// local copy already exists or the repo snapshot is missing.
func (s *applyPolicyState) migrateRepoSnapshotToLocal(agentKey, repoSnapshotPath, localSnapshotPath string) {
	if pathExists(localSnapshotPath) || !pathExists(repoSnapshotPath) {
		return
	}
	content, err := readText(repoSnapshotPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  WARNING: cannot read repo snapshot for migration (%s): %v\n", agentKey, err)
		return
	}
	if err := writeTextFile(localSnapshotPath, strings.TrimRight(content, "\r\n")); err != nil {
		fmt.Fprintf(os.Stderr, "  WARNING: cannot write local snapshot during migration (%s): %v\n", agentKey, err)
		return
	}
	s.localSnapshotMigrate++
	fmt.Printf("  migrated snapshot %s -> %s (from repo versioned snapshot)\n", agentKey, localSnapshotPath)
}

// backfillRepoSnapshotFromLocal copies the local operational snapshot back into
// the versioned repo snapshot directory when the repo copy is absent. No-op if
// the agent key should not be written to the repo, or if either condition is
// already satisfied.
func (s *applyPolicyState) backfillRepoSnapshotFromLocal(agentKey, localSnapshotPath, repoSnapshotPath string) {
	if !s.shouldWriteRepoSnapshot(agentKey) || pathExists(repoSnapshotPath) || !pathExists(localSnapshotPath) {
		return
	}
	content, err := readText(localSnapshotPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  WARNING: cannot read local snapshot for backfill (%s): %v\n", agentKey, err)
		return
	}
	if err := writeTextFile(repoSnapshotPath, strings.TrimRight(content, "\r\n")); err != nil {
		fmt.Fprintf(os.Stderr, "  WARNING: cannot write repo snapshot during backfill (%s): %v\n", agentKey, err)
		return
	}
	s.repoSnapshotBackfill++
	fmt.Printf("  backfilled repo snapshot %s -> %s (from local operational snapshot)\n", agentKey, repoSnapshotPath)
}
