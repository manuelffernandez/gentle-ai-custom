package overlay

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type ownedRuntimeState struct {
	Version int      `json:"version"`
	Paths   []string `json:"paths"`
}

func reconcileOwnedRuntimeOutputs(agent Agent, repoRoot string, target ManagedAssetsTarget, sharedSkillsRoot string, recorder *verboseRecorder) error {
	basePath, err := agent.BasePath()
	if err != nil {
		return fmt.Errorf("cannot resolve agent base path: %w", err)
	}
	statePath := filepath.Join(basePath, ".gentle-ai-custom-owned-state.json")
	previous, err := loadOwnedRuntimeState(statePath)
	if err != nil {
		return err
	}
	current, err := collectOwnedRuntimePaths(basePath, repoRoot, target, sharedSkillsRoot)
	if err != nil {
		return err
	}
	currentSet := make(map[string]bool, len(current))
	for _, path := range current {
		currentSet[path] = true
	}
	for _, path := range previous.Paths {
		if currentSet[path] {
			continue
		}
		if err := removeOwnedRuntimePath(basePath, path, recorder); err != nil {
			return err
		}
	}
	state := ownedRuntimeState{Version: 1, Paths: current}
	if _, err := writeJSONIndentedWithStatus(statePath, state); err != nil {
		return err
	}
	return nil
}

func loadOwnedRuntimeState(path string) (ownedRuntimeState, error) {
	if !pathExists(path) {
		return ownedRuntimeState{Version: 1, Paths: nil}, nil
	}
	var state ownedRuntimeState
	if err := readJSONFile(path, &state); err != nil {
		return ownedRuntimeState{}, fmt.Errorf("owned runtime state at %s is not valid JSON: %v", path, err)
	}
	return state, nil
}

func collectOwnedRuntimePaths(basePath, repoRoot string, target ManagedAssetsTarget, sharedSkillsRoot string) ([]string, error) {
	paths := map[string]struct{}{}
	for _, asset := range target.OwnedOverlayAssets {
		if asset.RepoOwnedPath == "" {
			// Upstream-source-only asset: no owned file; skipped by apply.
			continue
		}
		if asset.RuntimeSyncMode != "merge" {
			continue
		}
		source := ownedAssetSourcePath(repoRoot, asset)
		for _, runtimeTarget := range asset.RuntimeTargets {
			targetPaths, err := mapSourceFilesToRuntimeTargets(source, expandUser(runtimeTarget))
			if err != nil {
				return nil, err
			}
			for _, path := range targetPaths {
				paths[path] = struct{}{}
			}
		}
	}
	for _, skillName := range customSkillNames(target) {
		source := filepath.Join(sharedSkillsRoot, skillName)
		targetPaths, err := mapSourceFilesToRuntimeTargets(source, filepath.Join(basePath, "skills", skillName))
		if err != nil {
			return nil, err
		}
		for _, path := range targetPaths {
			paths[path] = struct{}{}
		}
	}
	for _, entry := range target.RepoOwnedCommandBodies.Entries {
		paths[filepath.Join(basePath, "commands", entry.Key+".md")] = struct{}{}
	}
	result := make([]string, 0, len(paths))
	for path := range paths {
		result = append(result, path)
	}
	sort.Strings(result)
	return result, nil
}

func mapSourceFilesToRuntimeTargets(srcRoot, dstRoot string) ([]string, error) {
	info, err := os.Stat(srcRoot)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{dstRoot}, nil
	}
	var result []string
	err = filepath.Walk(srcRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		result = append(result, filepath.Join(dstRoot, rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(result)
	return result, nil
}

func removeOwnedRuntimePath(basePath, path string, recorder *verboseRecorder) error {
	cleanBase := filepath.Clean(basePath)
	cleanPath := filepath.Clean(path)
	if cleanPath == cleanBase || filepath.Dir(cleanPath) == cleanPath {
		return fmt.Errorf("refusing to remove unsafe owned runtime path %s", path)
	}
	rel, err := filepath.Rel(cleanBase, cleanPath)
	if err != nil || rel == ".." || len(rel) >= 3 && rel[:3] == ".."+string(filepath.Separator) {
		return fmt.Errorf("owned runtime path %s is outside base path %s", path, basePath)
	}
	if pathExists(cleanPath) {
		if err := os.Remove(cleanPath); err != nil && !os.IsNotExist(err) {
			return err
		}
		recorder.record(cleanPath, "removed stale repo-owned runtime file")
	}
	return removeEmptyParents(cleanBase, filepath.Dir(cleanPath))
}

func removeEmptyParents(stopAt, current string) error {
	stopAt = filepath.Clean(stopAt)
	current = filepath.Clean(current)
	for current != stopAt && current != "." && current != string(filepath.Separator) {
		entries, err := os.ReadDir(current)
		if err != nil {
			if os.IsNotExist(err) {
				current = filepath.Dir(current)
				continue
			}
			return err
		}
		if len(entries) > 0 {
			return nil
		}
		if err := os.Remove(current); err != nil && !os.IsNotExist(err) {
			return err
		}
		current = filepath.Dir(current)
	}
	return nil
}
