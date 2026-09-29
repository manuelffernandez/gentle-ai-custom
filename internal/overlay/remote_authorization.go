package overlay

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const remoteAuthorizationPatchKey = "remote-authorization-scope"

var remoteAuthorizationHeading = regexp.MustCompile(`(?im)^##[ \t]+Remote operation authorization[ \t]*$`)
var anyMarkdownHeading = regexp.MustCompile(`(?m)^(#{1,6})[ \t]+(.+?)[ \t]*#*[ \t]*$`)
var remoteAuthorizationOpenMarker = regexp.MustCompile(`(?m)^<!-- gentle-ai:remote-authorization -->[ \t]*\r?$`)
var remoteAuthorizationCloseMarker = regexp.MustCompile(`(?m)^<!-- /gentle-ai:remote-authorization -->[ \t]*\r?$`)

func (s *applyPolicyState) prepareRemoteAuthorizationCustomization(repoRoot string) error {
	var ref *RuntimePatchRef
	for index := range s.managedTarget.RuntimePatches {
		candidate := &s.managedTarget.RuntimePatches[index]
		if candidate.Key == remoteAuthorizationPatchKey {
			if ref != nil {
				return fmt.Errorf("managed-assets declares duplicate runtime patch %q", remoteAuthorizationPatchKey)
			}
			ref = candidate
		}
	}
	if ref == nil {
		return fmt.Errorf("managed-assets does not declare required runtime patch %q", remoteAuthorizationPatchKey)
	}
	definitionPath := filepath.Join(repoRoot, filepath.FromSlash(ref.PatchPath))
	var definition runtimePatchDefinition
	if err := readJSONFile(definitionPath, &definition); err != nil {
		return fmt.Errorf("runtime patch %q is invalid at %s: %w", ref.Key, definitionPath, err)
	}
	if definition.Version != 1 || len(definition.Targets) != 0 || len(definition.Sections) != 1 {
		return fmt.Errorf("runtime patch %q must declare exactly one version-1 section and no fixed-file targets", ref.Key)
	}
	section := definition.Sections[0]
	if strings.TrimSpace(section.Before) == "" || strings.TrimSpace(section.After) == "" || section.Before == section.After {
		return fmt.Errorf("runtime patch %q has an invalid remote-authorization section definition", ref.Key)
	}

	fileChanges := map[string]preparedRuntimePatch{}
	maintainedPromptKeys := s.maintainedRemoteAuthorizationAgentKeys()
	for key, raw := range s.agents {
		agent, ok := jsonObject(raw)
		if !ok {
			continue
		}
		prompt := jsonString(agent["prompt"])
		if strings.TrimSpace(prompt) == "" {
			continue
		}
		maintained := maintainedPromptKeys[key]
		if !maintained {
			if s.isNamedProfileAgentKey(key) {
				contains, err := s.namedProfileHasRemoteAuthorization(prompt, filepath.Dir(s.configPath), section)
				if err != nil {
					return upstreamRuntimeError("cannot assess unmanaged named profile agent %q: %v", key, err)
				}
				if contains {
					return upstreamRuntimeError("unmanaged named profile agent %q contains remote-authorization instructions; declare or remove that profile before retrying", key)
				}
			}
			continue
		}
		if s.isManagedNamedProfileAgent(key) && strings.Contains(prompt, ".overlay.md") {
			continue
		}
		path, isFile, err := s.runtimePromptFilePath(prompt)
		if err != nil {
			return upstreamRuntimeError("maintained agent %q has an invalid prompt reference: %v", key, err)
		}
		if !isFile {
			updated, err := replaceRemoteAuthorizationSection(prompt, section)
			if err != nil {
				return upstreamRuntimeError("maintained agent %q prompt is incompatible: %v", key, err)
			}
			if updated != prompt {
				agent["prompt"] = updated
				s.configChanged = true
			}
			s.expectedPromptRefs[key] = updated
			continue
		}
		patch, err := prepareRemoteAuthorizationFile(path, filepath.Dir(s.configPath), section)
		if err != nil {
			return upstreamRuntimeError("maintained agent %q prompt file is incompatible: %v", key, err)
		}
		fileChanges[path] = patch
		s.expectedPromptRefs[key] = prompt
	}

	for _, path := range s.sharedSDDPromptPaths() {
		if _, exists := fileChanges[path]; exists {
			continue
		}
		if _, err := os.Lstat(path); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return upstreamRuntimeError("cannot inspect optional shared SDD prompt %s: %v", path, err)
		}
		patch, err := prepareRemoteAuthorizationFile(path, filepath.Dir(s.configPath), section)
		if err != nil {
			return upstreamRuntimeError("shared SDD prompt %s is incompatible: %v", path, err)
		}
		fileChanges[path] = patch
	}

	paths := make([]string, 0, len(fileChanges))
	for path := range fileChanges {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		patch := fileChanges[path]
		s.runtimePatches = append(s.runtimePatches, patch)
		s.removePreservedUpstreamSnapshot(path)
	}
	return nil
}

func (s *applyPolicyState) maintainedRemoteAuthorizationAgentKeys() map[string]bool {
	keys := map[string]bool{
		s.policy.OpenCode.BaseOrchestratorKey: true,
		"general":                             true,
		"explore":                             true,
	}
	for _, phase := range s.policy.OpenCode.SDDPhases {
		keys[phase] = true
	}
	for _, actor := range append(append([]string{}, judgmentDayActorKeys...), reviewActorKeys...) {
		keys[actor] = true
	}
	for key := range s.agents {
		if s.isNamedProfileAgentKey(key) {
			keys[key] = s.isManagedNamedProfileAgent(key)
		}
	}
	return keys
}

func (s *applyPolicyState) isNamedProfileAgentKey(key string) bool {
	if strings.HasPrefix(key, s.policy.OpenCode.ProfileOrchestratorPrefix) && len(key) > len(s.policy.OpenCode.ProfileOrchestratorPrefix) {
		return true
	}
	for _, phase := range s.policy.OpenCode.SDDPhases {
		if strings.HasPrefix(key, phase+"-") && len(key) > len(phase)+1 {
			return true
		}
	}
	for _, actor := range judgmentDayActorKeys {
		if strings.HasPrefix(key, actor+"-") && len(key) > len(actor)+1 {
			return true
		}
	}
	return false
}

func (s *applyPolicyState) isManagedNamedProfileAgent(key string) bool {
	for _, profile := range s.resolvedProfiles {
		if key == s.policy.OpenCode.ProfileOrchestratorPrefix+profile.Name {
			return true
		}
		for _, phase := range s.policy.OpenCode.SDDPhases {
			if key == phase+"-"+profile.Name {
				return true
			}
		}
		for _, actor := range judgmentDayActorKeys {
			if key == actor+"-"+profile.Name {
				return true
			}
		}
	}
	return false
}

func (s *applyPolicyState) namedProfileHasRemoteAuthorization(prompt, runtimeRoot string, section runtimePatchSection) (bool, error) {
	content := prompt
	path, isFile, err := s.runtimePromptFilePath(prompt)
	if err != nil {
		return false, err
	}
	if isFile {
		if err := validatePromptFilePath(path, runtimeRoot); err != nil {
			return false, err
		}
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return false, err
		}
		content = string(contentBytes)
	}
	return remoteAuthorizationSectionState(content, section) != remoteAuthorizationAbsent, nil
}

func (s *applyPolicyState) sharedSDDPromptPaths() []string {
	const prefix = "~/.config/opencode/prompts/sdd/"
	paths := map[string]bool{}
	for _, asset := range s.managedTarget.OwnedOverlayAssets {
		for _, target := range asset.RuntimeTargets {
			if !strings.HasPrefix(target, prefix) {
				continue
			}
			rel := strings.TrimPrefix(target, "~/.config/opencode/")
			paths[filepath.Join(filepath.Dir(s.configPath), filepath.FromSlash(rel))] = true
		}
	}
	result := make([]string, 0, len(paths))
	for path := range paths {
		result = append(result, path)
	}
	sort.Strings(result)
	return result
}

func prepareRemoteAuthorizationFile(path, runtimeRoot string, section runtimePatchSection) (preparedRuntimePatch, error) {
	if err := validatePromptFilePath(path, runtimeRoot); err != nil {
		return preparedRuntimePatch{}, err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return preparedRuntimePatch{}, fmt.Errorf("cannot inspect referenced file: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return preparedRuntimePatch{}, fmt.Errorf("prompt target is not a regular non-symlink file")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return preparedRuntimePatch{}, err
	}
	updated, err := replaceRemoteAuthorizationSection(string(raw), section)
	if err != nil {
		return preparedRuntimePatch{}, err
	}
	return preparedRuntimePatch{path: path, content: updated, mode: info.Mode()}, nil
}

func validatePromptFilePath(path, runtimeRoot string) error {
	root, err := filepath.Abs(filepath.Clean(runtimeRoot))
	if err != nil {
		return err
	}
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(root, absolute)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("prompt file %s is outside OpenCode runtime root %s", path, root)
	}
	for current := absolute; ; current = filepath.Dir(current) {
		if info, err := os.Lstat(current); err == nil && info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("prompt path traverses symlink %s", current)
		} else if err != nil && !os.IsNotExist(err) {
			return err
		}
		if parent := filepath.Dir(current); parent == current {
			break
		}
	}
	return nil
}

type remoteAuthorizationSectionStatus int

const (
	remoteAuthorizationAbsent remoteAuthorizationSectionStatus = iota
	remoteAuthorizationOriginal
	remoteAuthorizationRevised
)

func remoteAuthorizationSectionState(content string, section runtimePatchSection) remoteAuthorizationSectionStatus {
	updated, err := replaceRemoteAuthorizationSection(content, section)
	if err == nil && updated != content {
		return remoteAuthorizationOriginal
	}
	for _, match := range anyMarkdownHeading.FindAllStringSubmatchIndex(content, -1) {
		heading := strings.TrimSpace(content[match[4]:match[5]])
		if strings.Contains(strings.ToLower(heading), "remote") && strings.Contains(strings.ToLower(heading), "authoriz") {
			return remoteAuthorizationRevised
		}
	}
	if strings.Contains(strings.ToLower(content), "remote operation authorization") || strings.Contains(strings.ToLower(content), "remote authorization") {
		return remoteAuthorizationRevised
	}
	return remoteAuthorizationAbsent
}

func replaceRemoteAuthorizationSection(content string, section runtimePatchSection) (string, error) {
	var starts []int
	for _, match := range anyMarkdownHeading.FindAllStringSubmatchIndex(content, -1) {
		heading := strings.TrimSpace(content[match[4]:match[5]])
		if strings.Contains(strings.ToLower(heading), "remote") && strings.Contains(strings.ToLower(heading), "authoriz") {
			if !strings.EqualFold(heading, "Remote operation authorization") || len(content[match[2]:match[3]]) != 2 {
				return "", fmt.Errorf("unknown remote-authorization heading anchor %q", heading)
			}
			starts = append(starts, match[0])
		}
	}
	if len(starts) != 1 || !remoteAuthorizationHeading.MatchString(content) {
		return "", fmt.Errorf("expected exactly one complete remote-authorization section, found %d", len(starts))
	}
	start := starts[0]
	openMarkers := remoteAuthorizationOpenMarker.FindAllStringIndex(content, -1)
	closeMarkers := remoteAuthorizationCloseMarker.FindAllStringIndex(content, -1)
	openMarker := "<!-- gentle-ai:remote-authorization -->"
	closeMarker := "<!-- /gentle-ai:remote-authorization -->"
	if strings.Contains(content, openMarker) || strings.Contains(content, closeMarker) {
		if strings.Count(content, openMarker) != 1 || len(openMarkers) != 1 || strings.Count(content, closeMarker) != 1 || len(closeMarkers) != 1 {
			return "", fmt.Errorf("expected exactly one well-formed remote-authorization marker pair")
		}
		openEnd := openMarkers[0][1]
		if openMarkers[0][0] >= start || closeMarkers[0][0] <= start || closeMarkers[0][0] <= openEnd {
			return "", fmt.Errorf("remote-authorization markers are missing, duplicated, or out of order")
		}
		bodyStart := openEnd
		if strings.HasPrefix(content[bodyStart:], "\r\n") {
			bodyStart += 2
		} else if strings.HasPrefix(content[bodyStart:], "\n") {
			bodyStart++
		} else {
			return "", fmt.Errorf("remote-authorization opening marker is not followed by a newline")
		}
		closeStart := closeMarkers[0][0]
		if strings.TrimSpace(content[bodyStart:start]) != "" {
			return "", fmt.Errorf("unexpected content before remote-authorization heading")
		}
		for _, match := range anyMarkdownHeading.FindAllStringSubmatchIndex(content, -1) {
			if match[0] > start && match[0] < closeStart && len(content[match[2]:match[3]]) <= 2 {
				return "", fmt.Errorf("unexpected heading inside marked remote-authorization section")
			}
		}
		oldSection := content[start:closeStart]
		trimmed := strings.TrimSpace(oldSection)
		if trimmed == strings.TrimSpace(section.After) {
			return content, nil
		}
		if trimmed != strings.TrimSpace(section.Before) {
			return "", fmt.Errorf("remote-authorization section does not match the declared original or revised section")
		}
		withoutTrailingSpace := strings.TrimRight(oldSection, " \t\r\n")
		separator := oldSection[len(withoutTrailingSpace):]
		return content[:start] + section.After + separator + content[closeStart:], nil
	}
	end := len(content)
	for _, match := range anyMarkdownHeading.FindAllStringSubmatchIndex(content, -1) {
		if match[0] <= start || len(content[match[2]:match[3]]) > 2 {
			continue
		}
		end = match[0]
		break
	}
	oldSection := content[start:end]
	trimmed := strings.TrimSpace(oldSection)
	if trimmed == strings.TrimSpace(section.After) {
		return content, nil
	}
	if trimmed != strings.TrimSpace(section.Before) {
		return "", fmt.Errorf("remote-authorization section does not match the declared original or revised section")
	}
	withoutTrailingSpace := strings.TrimRight(oldSection, " \t\r\n")
	separator := oldSection[len(withoutTrailingSpace):]
	return content[:start] + section.After + separator + content[end:], nil
}

func (s *applyPolicyState) removePreservedUpstreamSnapshot(path string) {
	kept := s.preservedUpstreamFiles[:0]
	for _, snapshot := range s.preservedUpstreamFiles {
		if filepath.Clean(snapshot.path) != filepath.Clean(path) {
			kept = append(kept, snapshot)
		}
	}
	s.preservedUpstreamFiles = kept
}
