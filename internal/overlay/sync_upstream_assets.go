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
	release  string
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
	release, err := verifyUpstreamReleaseParity(upstreamRepo, options.release)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: upstream release parity preflight failed: %v\n", err)
		return 1
	}
	if err := validateTargetParity(manifest, registeredAgentNames()); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	statePath := filepath.Join(repoRoot, policy.Maintenance.StateFile)
	var state UpstreamState
	if err := readJSONFile(statePath, &state); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: state file is not valid JSON at %s: %v\n", statePath, err)
		return 1
	}

	for _, name := range []string{"opencode", "codex"} {
		if _, ok := manifest.Targets[name]; !ok {
			fmt.Fprintf(os.Stderr, "ERROR: managed-assets manifest at %s does not define target %q\n", manifestPath, name)
			return 1
		}
	}
	target := manifest.Targets["opencode"]
	codex := manifest.Targets["codex"]

	upstreamHead := release.Commit
	diffEntries, err := runGitDiff(upstreamRepo, state.LastMaintainedCommit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot diff upstream git state from %s: %v\n", state.LastMaintainedCommit, err)
		return 1
	}
	managedEntries, unmanagedEntries := categorizeGitDiff(diffEntries, buildManagedAssetCatalog(target))
	codexManaged, codexUnmanaged := categorizeGitDiff(diffEntries, buildManagedAssetCatalog(codex))
	unmanagedEntries = append(unmanagedEntries, codexUnmanaged...)
	if len(unmanagedEntries) > 0 {
		fmt.Fprintln(os.Stderr, "ERROR: sync-upstream-assets refused to advance the maintained boundary because watched upstream drift is still unmapped.")
		for _, entry := range unmanagedEntries {
			fmt.Fprintf(os.Stderr, "  - %s\n", formatUnmanagedDiffEntry(entry))
		}
		fmt.Fprintln(os.Stderr, "Resolve managed-assets.json first, then re-run the sync.")
		return 1
	}
	structuralResult, structuralFailures := evaluateStructuralInvariants(upstreamRepo, target, policy)
	for _, failure := range evaluateCodexStructuralInvariants(upstreamRepo, codex) {
		structuralFailures = append(structuralFailures, "Codex: "+failure)
	}
	if len(structuralFailures) > 0 {
		fmt.Fprintln(os.Stderr, "ERROR: sync-upstream-assets refused to advance the maintained boundary because structural invariants failed.")
		for _, failure := range structuralFailures {
			fmt.Fprintf(os.Stderr, "  - %s\n", failure)
		}
		return 1
	}
	if !structuralResult.OrchestratorAssetOK || !structuralResult.RoutingInjectionOK || !structuralResult.RemoteAuthorizationOK || !structuralResult.RoleTopologyOK || !structuralResult.ODDTopologyOK || !structuralResult.ODDRegistrationOK {
		fmt.Fprintln(os.Stderr, "ERROR: sync-upstream-assets refused to advance the maintained boundary because structural invariants are not all ok.")
		return 1
	}
	if err := validateRetiredUpstreamSnapshots(repoRoot, target); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: sync-upstream-assets refused unsafe retired snapshot cleanup: %v\n", err)
		return 1
	}
	if !codexSnapshotMappingsComplete(codex) {
		fmt.Fprintln(os.Stderr, "ERROR: Codex approved snapshot mappings incomplete; orchestrator and remote-authorization sources are required")
		return 1
	}
	for _, asset := range codex.OwnedOverlayAssets {
		if !validCodexSnapshotMapping(asset) {
			fmt.Fprintf(os.Stderr, "ERROR: unsafe Codex approved snapshot mapping %q\n", asset.Key)
			return 1
		}
		if _, err := runGitShow(upstreamRepo, "HEAD", asset.UpstreamPath); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Codex approved source %s cannot be read before sync writes: %v\n", asset.UpstreamPath, err)
			return 1
		}
	}

	stats := &syncUpstreamAssetsStats{}
	// Both targets must pass all read-only gates before the first snapshot write.
	if err := syncManagedTarget(repoRoot, upstreamRepo, target, options.recorder, stats); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	if err := syncManagedTarget(repoRoot, upstreamRepo, codex, options.recorder, stats); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		return 1
	}
	if err := removeRetiredUpstreamSnapshots(repoRoot, target, options.recorder, stats); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot remove retired upstream snapshots: %v\n", err)
		return 1
	}

	updatedState := state
	boundaryChanged := state.LastMaintainedCommit != release.Commit || state.LastMaintainedVersion != release.Version || state.LastMaintainedTag != release.Tag
	if boundaryChanged {
		previousCommit := state.LastMaintainedCommit
		updatedState.LastMaintainedCommit = release.Commit
		updatedState.LastMaintainedVersion = release.Version
		updatedState.LastMaintainedTag = release.Tag
		updatedState.LastReviewedAt = time.Now().UTC().Format(time.RFC3339)
		updatedState.Notes = formatUpstreamAssetSyncNote(previousCommit, upstreamHead)
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
	fmt.Printf("- Installed binary version: %s\n", release.InstalledVersion)
	fmt.Printf("- Selected sync release: %s (%s)\n", release.Tag, release.Version)
	fmt.Printf("- Verified release commit: %s\n", release.Commit)
	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("  managed files changed in range: %d\n", len(managedEntries))
	fmt.Printf("  Codex managed files changed in range: %d\n", len(codexManaged))
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
	help := false
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "-h", "--help":
			help = true
		case "--verbose":
			options.verbose = true
		case "--release":
			if options.release != "" {
				fmt.Fprintln(os.Stderr, "ERROR: --release may be specified only once")
				printSyncUpstreamAssetsUsage(os.Stderr)
				return options, 1
			}
			if index+1 >= len(args) || strings.HasPrefix(args[index+1], "-") || !isValidReleaseTag(args[index+1]) {
				fmt.Fprintln(os.Stderr, "ERROR: --release requires an exact v<V> tag, such as v2.9.1")
				printSyncUpstreamAssetsUsage(os.Stderr)
				return options, 1
			}
			options.release = args[index+1]
			index++
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
	if help {
		printSyncUpstreamAssetsUsage(os.Stdout)
		return options, 0
	}
	return options, -1
}

func printSyncUpstreamAssetsUsage(out *os.File) {
	fmt.Fprintf(out, "Usage: %s [--verbose] [--release v<V>]\n", usageCommandName("sync-upstream-assets"))
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

const upstreamSnapshotRoot = "overlay/gentle-ai/assets/upstream/opencode"

func removeRetiredUpstreamSnapshots(repoRoot string, target ManagedAssetsTarget, recorder *verboseRecorder, stats *syncUpstreamAssetsStats) error {
	paths, err := retiredUpstreamSnapshotFiles(repoRoot, target)
	if err != nil {
		return err
	}
	for _, full := range paths {
		if err := os.Remove(full); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return err
		}
		stats.FilesDeleted++
		recorder.record(full, "removed manifest-retired upstream snapshot file")
	}
	return nil
}

func validateRetiredUpstreamSnapshots(repoRoot string, target ManagedAssetsTarget) error {
	_, err := retiredUpstreamSnapshotFiles(repoRoot, target)
	return err
}

func retiredUpstreamSnapshotFiles(repoRoot string, target ManagedAssetsTarget) ([]string, error) {
	active := make([]string, 0, len(target.OwnedOverlayAssets)+len(target.RetainedUpstreamSkills)+len(target.PrunedUpstreamSkills))
	var files []string
	for _, asset := range target.OwnedOverlayAssets {
		active = append(active, snapshotRelativePath(asset.RepoUpstreamPath))
	}
	for _, asset := range target.RetainedUpstreamSkills {
		active = append(active, snapshotRelativePath(asset.RepoUpstreamPath))
	}
	for _, asset := range target.PrunedUpstreamSkills {
		active = append(active, snapshotRelativePath(asset.RepoUpstreamPath))
	}

	for _, retired := range target.RetiredUpstreamSnapshots {
		raw := filepath.ToSlash(strings.TrimSpace(retired))
		invalidSegment := false
		for _, segment := range strings.Split(raw, "/") {
			if segment == "" || segment == "." || segment == ".." {
				invalidSegment = true
				break
			}
		}
		if raw == "" || strings.HasPrefix(raw, "/") || filepath.IsAbs(retired) || invalidSegment {
			return nil, fmt.Errorf("unsafe retired snapshot path %q", retired)
		}
		clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(raw)))
		for _, activePath := range active {
			activeClean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(activePath)))
			if activeClean == clean || strings.HasPrefix(activeClean, clean+"/") || strings.HasPrefix(clean, activeClean+"/") {
				return nil, fmt.Errorf("retired snapshot path %q overlaps active snapshot path %q", clean, activeClean)
			}
		}
		full := filepath.Join(repoRoot, filepath.FromSlash(upstreamSnapshotRoot), filepath.FromSlash(clean))
		if err := ensureNoSymlinkParents(repoRoot, full); err != nil {
			return nil, fmt.Errorf("retired snapshot path %q: %w", clean, err)
		}
		info, err := os.Lstat(full)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if info.IsDir() {
			return nil, fmt.Errorf("retired snapshot path %q resolves to a directory; refusing recursive deletion", clean)
		}
		files = append(files, full)
	}
	return files, nil
}

func snapshotRelativePath(path string) string {
	path = filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	return strings.TrimPrefix(path, upstreamSnapshotRoot+"/")
}

func ensureNoSymlinkParents(root, target string) error {
	rootInfo, err := os.Lstat(root)
	if err == nil && rootInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("snapshot root %q is a symlink", root)
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return fmt.Errorf("target is outside snapshot root")
	}
	current := root
	parts := strings.Split(filepath.Dir(rel), string(filepath.Separator))
	for _, part := range parts {
		if part == "." || part == "" {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("parent %q is a symlink", current)
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

func formatUpstreamAssetSyncNote(previousCommit, newCommit string) string {
	previousCommit = strings.TrimSpace(previousCommit)
	newCommit = strings.TrimSpace(newCommit)
	if previousCommit == "" {
		return fmt.Sprintf("Updated approved upstream asset snapshots via sync-upstream-assets to %s.", newCommit)
	}
	return fmt.Sprintf("Updated approved upstream asset snapshots via sync-upstream-assets from %s to %s.", previousCommit, newCommit)
}
