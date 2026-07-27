package overlay

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type writeCounters struct {
	New       int
	Changed   int
	Unchanged int
	Deleted   int
}

func trackWriteStatus(status string, counters *writeCounters) {
	if counters == nil {
		return
	}
	switch status {
	case "new":
		counters.New++
	case "changed":
		counters.Changed++
	case "unchanged":
		counters.Unchanged++
	case "deleted":
		counters.Deleted++
	}
}

func (s *applyPolicyState) installOwnedAssets(repoRoot string) error {
	for _, asset := range s.managedTarget.OwnedOverlayAssets {
		if asset.Class != "repo_owned_runtime" {
			if asset.RepoOwnedPath != "" {
				return fmt.Errorf("managed asset %q is class %q but declares repo_owned_path; upstream snapshots cannot be runtime install sources", asset.Key, asset.Class)
			}
			continue
		}
		if asset.RepoOwnedPath == "" {
			return fmt.Errorf("repo-owned runtime asset %q does not declare repo_owned_path", asset.Key)
		}
		source := ownedAssetSourcePath(repoRoot, asset)
		if strings.HasSuffix(asset.Kind, "_directory") {
			if err := s.installOwnedDirectoryAsset(asset, source); err != nil {
				return err
			}
			continue
		}
		if err := s.installOwnedFileAsset(asset, source); err != nil {
			return err
		}
	}
	return nil
}

func (s *applyPolicyState) installOwnedDirectoryAsset(asset OwnedOverlayAsset, source string) error {
	for _, runtimeTarget := range asset.RuntimeTargets {
		destination := expandUser(runtimeTarget)
		prune := asset.RuntimeSyncMode != "merge"
		if err := copyDirectoryContentsWithStatus(source, destination, prune, s.recorder, &s.metrics.ownedAssetWrites); err != nil {
			return fmt.Errorf("cannot install owned directory asset %q from %s to %s: %w", asset.Key, source, destination, err)
		}
		fmt.Printf("  installed owned directory %s -> %s\n", asset.Key, destination)
	}
	return nil
}

func (s *applyPolicyState) installOwnedFileAsset(asset OwnedOverlayAsset, source string) error {
	for _, runtimeTarget := range asset.RuntimeTargets {
		destination := expandUser(runtimeTarget)
		status, err := copyFileWithStatus(source, destination)
		if err != nil {
			return fmt.Errorf("cannot install owned asset %q from %s to %s: %w", asset.Key, source, destination, err)
		}
		trackWriteStatus(status, &s.metrics.ownedAssetWrites)
		if shouldRecordWriteStatus(status) {
			s.recordVerbose(destination, fmt.Sprintf("installed owned asset %s (%s)", asset.Key, describeWriteStatus(status)))
		}
		fmt.Printf("  installed owned asset %s -> %s\n", asset.Key, destination)
	}
	return nil
}

func (s *applyPolicyState) verifyOwnedAssetTargets() error {
	verified := 0
	for _, asset := range s.managedTarget.OwnedOverlayAssets {
		if asset.Class != "repo_owned_runtime" {
			continue
		}
		for _, runtimeTarget := range asset.RuntimeTargets {
			path := expandUser(runtimeTarget)
			info, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("owned runtime target missing for %q at %s: %w", asset.Key, path, err)
			}
			if strings.HasSuffix(asset.Kind, "_directory") && !info.IsDir() {
				return fmt.Errorf("owned runtime target for %q must be a directory: %s", asset.Key, path)
			}
			verified++
		}
	}
	s.metrics.ownedRuntimeTargetsVerified = verified
	return nil
}

func ownedAssetSourcePath(repoRoot string, asset OwnedOverlayAsset) string {
	return filepath.Join(repoRoot, filepath.FromSlash(asset.RepoOwnedPath))
}
