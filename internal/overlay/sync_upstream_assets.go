package overlay

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type syncUpstreamAssetsOptions struct {
	verbose  bool
	recorder *verboseRecorder
}

type syncUpstreamAssetsStats struct {
	FilesNew       int
	FilesChanged   int
	FilesUnchanged int
	FilesDeleted   int
	DirsSynced     int
}

func RunSyncUpstreamAssets(repoRoot string, args []string) int {
	options, exitCode := normalizeSyncUpstreamAssetsArgs(args)
	if exitCode >= 0 {
		return exitCode
	}
	options.recorder = newVerboseRecorder(options.verbose)
	defer options.recorder.print()

	policy, _, err := loadPolicy(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	manifest, manifestPath, err := loadManagedAssetsManifest(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	upstreamRepo, upstreamRepoSource, err := resolveUpstreamRepo(repoRoot, policy)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	statePath := filepath.Join(repoRoot, policy.Maintenance.StateFile)
	var state UpstreamState
	if err := readJSONFile(statePath, &state); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: state file is not valid JSON at %s: %v\n", statePath, err)
		return 1
	}

	target, ok := manifest.Targets["opencode"]
	if !ok {
		fmt.Fprintf(os.Stderr, "ERROR: managed-assets manifest at %s does not define target %q\n", manifestPath, "opencode")
		return 1
	}

	upstreamHead, err := runGit(upstreamRepo, false, "rev-parse", "HEAD")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot inspect upstream git state: %v\n", err)
		return 1
	}
	upstreamDescribe, _ := runGit(upstreamRepo, true, "describe", "--tags", "--always")
	upstreamExactTag, _ := runGit(upstreamRepo, true, "describe", "--tags", "--exact-match")
	diffEntries, err := runGitDiff(upstreamRepo, state.LastMaintainedCommit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot diff upstream git state from %s: %v\n", state.LastMaintainedCommit, err)
		return 1
	}
	managedEntries, unmanagedEntries := categorizeGitDiff(diffEntries, buildManagedAssetCatalog(target))
	if len(unmanagedEntries) > 0 {
		fmt.Fprintln(os.Stderr, "ERROR: sync-upstream-assets refused to advance the maintained boundary because watched upstream drift is still unmapped.")
		for _, entry := range unmanagedEntries {
			fmt.Fprintf(os.Stderr, "  - %s\n", formatUnmanagedDiffEntry(entry))
		}
		fmt.Fprintln(os.Stderr, "Resolve managed-assets.json first, then re-run the sync.")
		return 1
	}
	structuralResult, structuralFailures := evaluateStructuralInvariants(upstreamRepo, target, policy)
	if len(structuralFailures) > 0 {
		fmt.Fprintln(os.Stderr, "ERROR: sync-upstream-assets refused to advance the maintained boundary because structural invariants failed.")
		for _, failure := range structuralFailures {
			fmt.Fprintf(os.Stderr, "  - %s\n", failure)
		}
		return 1
	}
	if !structuralResult.PhaseOrderOK || !structuralResult.ProfileNamingOK || !structuralResult.TaskScopingOK || !structuralResult.BaseAssetInjectionOK {
		fmt.Fprintln(os.Stderr, "ERROR: sync-upstream-assets refused to advance the maintained boundary because structural invariants are not all ok.")
		return 1
	}

	stats := &syncUpstreamAssetsStats{}
	if err := syncManagedTarget(repoRoot, upstreamRepo, target, options.recorder, stats); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}

	updatedState := state
	boundaryChanged := state.LastMaintainedCommit != upstreamHead || state.LastMaintainedVersion != firstNonEmpty(upstreamExactTag, upstreamDescribe, upstreamHead) || state.LastMaintainedTag != firstNonEmpty(upstreamExactTag, upstreamDescribe)
	if boundaryChanged {
		updatedState.LastMaintainedCommit = upstreamHead
		updatedState.LastMaintainedVersion = firstNonEmpty(upstreamExactTag, upstreamDescribe, upstreamHead)
		updatedState.LastMaintainedTag = firstNonEmpty(upstreamExactTag, upstreamDescribe)
		updatedState.LastReviewedAt = time.Now().UTC().Format(time.RFC3339)
		updatedState.Notes = fmt.Sprintf("Updated approved upstream asset snapshots via sync-upstream-assets from %s.", upstreamHead)
	}
	status, err := writeJSONIndentedWithStatus(statePath, updatedState)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot update %s: %v\n", statePath, err)
		return 1
	}
	if shouldRecordWriteStatus(status) {
		options.recorder.record(statePath, fmt.Sprintf("updated upstream state after asset sync (%s)", describeWriteStatus(status)))
	}

	fmt.Println("Syncing approved upstream assets...")
	fmt.Printf("- Repo root: %s\n", repoRoot)
	fmt.Printf("- Upstream repo: %s\n", upstreamRepo)
	fmt.Printf("- Upstream source: %s\n", upstreamRepoSource)
	fmt.Printf("- Managed assets manifest: %s\n", manifestPath)
	fmt.Printf("- Upstream HEAD: %s\n", upstreamHead)
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  managed files changed in range: %d\n", len(managedEntries))
	fmt.Printf("  files new: %d\n", stats.FilesNew)
	fmt.Printf("  files changed: %d\n", stats.FilesChanged)
	fmt.Printf("  files unchanged: %d\n", stats.FilesUnchanged)
	fmt.Printf("  files deleted: %d\n", stats.FilesDeleted)
	fmt.Printf("  directories synced: %d\n", stats.DirsSynced)
	fmt.Printf("  upstream state write: %s\n", status)
	fmt.Println()
	fmt.Println("Done. Approved upstream asset snapshots now match the current manifest and upstream boundary.")
	return 0
}

func normalizeSyncUpstreamAssetsArgs(args []string) (syncUpstreamAssetsOptions, int) {
	var options syncUpstreamAssetsOptions
	for _, arg := range args {
		switch arg {
		case "-h", "--help":
			printSyncUpstreamAssetsUsage(os.Stdout)
			return options, 0
		case "--verbose":
			options.verbose = true
		default:
			if strings.HasPrefix(arg, "-") {
				fmt.Fprintf(os.Stderr, "Unknown sync-upstream-assets flag: %s\n", arg)
			} else {
				fmt.Fprintf(os.Stderr, "sync-upstream-assets does not accept positional argument: %s\n", arg)
			}
			printSyncUpstreamAssetsUsage(os.Stderr)
			return options, 1
		}
	}
	return options, -1
}

func printSyncUpstreamAssetsUsage(out *os.File) {
	fmt.Fprintf(out, "Usage: %s [--verbose]\n", usageCommandName("sync-upstream-assets"))
}

func syncManagedTarget(repoRoot, upstreamRepo string, target ManagedAssetsTarget, recorder *verboseRecorder, stats *syncUpstreamAssetsStats) error {
	for _, asset := range target.OwnedOverlayAssets {
		if asset.UpstreamPath == "" {
			// Apply-only owned asset: no upstream source to sync; installed by apply only.
			continue
		}
		dst := filepath.Join(repoRoot, filepath.FromSlash(asset.RepoUpstreamPath))
		src := filepath.Join(upstreamRepo, filepath.FromSlash(asset.UpstreamPath))
		if strings.HasSuffix(asset.Kind, "_directory") {
			if err := syncDirectoryWithStatus(src, dst, recorder, stats); err != nil {
				return fmt.Errorf("cannot sync directory asset %q: %w", asset.Key, err)
			}
			continue
		}
		status, err := copyFileWithStatus(src, dst)
		if err != nil {
			return fmt.Errorf("cannot sync asset %q from %s to %s: %w", asset.Key, src, dst, err)
		}
		trackSyncFileStatus(status, stats)
		if shouldRecordWriteStatus(status) {
			recorder.record(dst, fmt.Sprintf("synced upstream asset %s (%s)", asset.Key, describeWriteStatus(status)))
		}
	}
	for _, asset := range target.RetainedUpstreamSkills {
		if err := syncReferencedUpstreamSkill(repoRoot, upstreamRepo, asset, recorder, stats); err != nil {
			return err
		}
	}
	for _, asset := range target.PrunedUpstreamSkills {
		if err := syncReferencedUpstreamSkill(repoRoot, upstreamRepo, asset, recorder, stats); err != nil {
			return err
		}
	}
	return nil
}

func syncReferencedUpstreamSkill(repoRoot, upstreamRepo string, asset UpstreamSkillRef, recorder *verboseRecorder, stats *syncUpstreamAssetsStats) error {
	if strings.TrimSpace(asset.RepoUpstreamPath) == "" {
		return nil
	}
	src := filepath.Join(upstreamRepo, filepath.FromSlash(asset.UpstreamPath))
	dst := filepath.Join(repoRoot, filepath.FromSlash(asset.RepoUpstreamPath))
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("cannot stat referenced upstream skill %q at %s: %w", asset.Key, src, err)
	}
	if info.IsDir() {
		if err := syncDirectoryWithStatus(src, dst, recorder, stats); err != nil {
			return fmt.Errorf("cannot sync referenced upstream skill directory %q from %s to %s: %w", asset.Key, src, dst, err)
		}
		return nil
	}
	status, err := copyFileWithStatus(src, dst)
	if err != nil {
		return fmt.Errorf("cannot sync referenced upstream skill %q from %s to %s: %w", asset.Key, src, dst, err)
	}
	trackSyncFileStatus(status, stats)
	if shouldRecordWriteStatus(status) {
		recorder.record(dst, fmt.Sprintf("synced referenced upstream skill %s (%s)", asset.Key, describeWriteStatus(status)))
	}
	return nil
}

func syncDirectoryWithStatus(srcRoot, dstRoot string, recorder *verboseRecorder, stats *syncUpstreamAssetsStats) error {
	if !pathExists(srcRoot) {
		return fmt.Errorf("source directory missing: %s", srcRoot)
	}
	seen := map[string]bool{}
	if err := filepath.Walk(srcRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			seen[rel] = true
			return ensureDir(dstRoot)
		}
		seen[rel] = true
		dstPath := filepath.Join(dstRoot, rel)
		if info.IsDir() {
			return ensureDir(dstPath)
		}
		status, err := copyFileWithStatus(path, dstPath)
		if err != nil {
			return err
		}
		trackSyncFileStatus(status, stats)
		if shouldRecordWriteStatus(status) {
			recorder.record(dstPath, fmt.Sprintf("synced upstream directory file (%s)", describeWriteStatus(status)))
		}
		return nil
	}); err != nil {
		return err
	}
	stats.DirsSynced++
	return pruneMissingDirectoryEntries(dstRoot, seen, recorder, stats)
}

func pruneMissingDirectoryEntries(dstRoot string, seen map[string]bool, recorder *verboseRecorder, stats *syncUpstreamAssetsStats) error {
	if !pathExists(dstRoot) {
		return nil
	}
	var paths []string
	err := filepath.Walk(dstRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dstRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		paths = append(paths, rel)
		return nil
	})
	if err != nil {
		return err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(paths)))
	for _, rel := range paths {
		if seen[rel] {
			continue
		}
		full := filepath.Join(dstRoot, rel)
		if err := os.RemoveAll(full); err != nil {
			return err
		}
		stats.FilesDeleted++
		recorder.record(full, "removed stale upstream snapshot entry during directory sync")
	}
	return nil
}

func trackSyncFileStatus(status string, stats *syncUpstreamAssetsStats) {
	switch status {
	case "new":
		stats.FilesNew++
	case "changed":
		stats.FilesChanged++
	case "unchanged":
		stats.FilesUnchanged++
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
