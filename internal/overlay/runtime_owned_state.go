package overlay

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
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
	upstreamProtected, err := collectUpstreamRuntimeProtectionPaths(basePath, repoRoot, target)
	if err != nil {
		return err
	}
	for _, path := range previous.Paths {
		if currentSet[path] {
			continue
		}
		if isProtectedUpstreamRuntimePath(basePath, path, upstreamProtected) {
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

func collectUpstreamRuntimeProtectionPaths(basePath, repoRoot string, target ManagedAssetsTarget) ([]string, error) {
	protected := map[string]struct{}{}
	add := func(path string) error {
		cleanPath, err := cleanRuntimePathWithinBase(basePath, path)
		if err != nil {
			return err
		}
		protected[cleanPath] = struct{}{}
		return nil
	}
	for _, asset := range target.OwnedOverlayAssets {
		if asset.RepoOwnedPath != "" {
			continue
		}
		for _, runtimeTarget := range asset.RuntimeTargets {
			if err := add(expandUser(runtimeTarget)); err != nil {
				return nil, fmt.Errorf("unsafe upstream runtime target for %q: %w", asset.Key, err)
			}
		}
	}
	for _, runtimeTarget := range target.ProtectedRuntimeTargets {
		if err := add(expandUser(runtimeTarget)); err != nil {
			return nil, fmt.Errorf("unsafe explicitly protected upstream runtime target %q: %w", runtimeTarget, err)
		}
	}
	for _, skill := range target.RetainedUpstreamSkills {
		if err := add(filepath.Join(basePath, "skills", skill.Key)); err != nil {
			return nil, fmt.Errorf("unsafe retained upstream skill path for %q: %w", skill.Key, err)
		}
	}
	for _, patch := range target.RuntimePatches {
		definitionPath := filepath.Join(repoRoot, filepath.FromSlash(patch.PatchPath))
		var definition runtimePatchDefinition
		if err := readJSONFile(definitionPath, &definition); err != nil {
			return nil, fmt.Errorf("runtime patch %q is invalid at %s: %w", patch.Key, definitionPath, err)
		}
		for _, patchTarget := range definition.Targets {
			if err := add(filepath.Join(basePath, filepath.FromSlash(patchTarget.Path))); err != nil {
				return nil, fmt.Errorf("unsafe runtime patch target for %q: %w", patch.Key, err)
			}
		}
	}
	result := make([]string, 0, len(protected))
	for path := range protected {
		result = append(result, path)
	}
	sort.Strings(result)
	return result, nil
}

func isProtectedUpstreamRuntimePath(basePath, path string, protectedRoots []string) bool {
	cleanPath, err := cleanRuntimePathWithinBase(basePath, path)
	if err != nil {
		return false
	}
	for _, root := range protectedRoots {
		rel, err := filepath.Rel(root, cleanPath)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return true
		}
	}
	return false
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
	add := func(path string) error {
		cleanPath, err := cleanRuntimePathWithinBase(basePath, path)
		if err != nil {
			return err
		}
		paths[cleanPath] = struct{}{}
		return nil
	}
	for _, asset := range target.OwnedOverlayAssets {
		if asset.Class != "repo_owned_runtime" {
			continue
		}
		source := ownedAssetSourcePath(repoRoot, asset)
		for _, runtimeTarget := range asset.RuntimeTargets {
			targetPaths, err := mapSourceFilesToRuntimeTargets(source, expandUser(runtimeTarget))
			if err != nil {
				return nil, err
			}
			for _, path := range targetPaths {
				if err := add(path); err != nil {
					return nil, fmt.Errorf("unsafe repo-owned runtime target for %q: %w", asset.Key, err)
				}
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
			if err := add(path); err != nil {
				return nil, fmt.Errorf("unsafe repo-owned skill target for %q: %w", skillName, err)
			}
		}
	}
	for _, entry := range target.RepoOwnedCommandBodies.Entries {
		if err := add(filepath.Join(basePath, "commands", entry.Key+".md")); err != nil {
			return nil, fmt.Errorf("unsafe repo-owned command target for %q: %w", entry.Key, err)
		}
	}
	result := make([]string, 0, len(paths))
	for path := range paths {
		result = append(result, path)
	}
	sort.Strings(result)
	return result, nil
}

func cleanRuntimePathWithinBase(basePath, path string) (string, error) {
	cleanBase, err := filepath.Abs(filepath.Clean(basePath))
	if err != nil {
		return "", err
	}
	cleanPath, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(cleanBase, cleanPath)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("runtime path %s is outside or equal to base path %s", path, basePath)
	}
	return cleanPath, nil
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
