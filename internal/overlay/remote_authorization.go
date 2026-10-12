package overlay

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var remoteAuthorizationHeading = regexp.MustCompile(`(?im)^##[ \t]+Remote operation authorization[ \t]*$`)
var anyMarkdownHeading = regexp.MustCompile(`(?m)^(#{1,6})[ \t]+(.+?)[ \t]*#*[ \t]*$`)
var remoteAuthorizationOpenMarker = regexp.MustCompile(`(?m)^<!-- gentle-ai:remote-authorization -->[ \t]*\r?$`)
var remoteAuthorizationCloseMarker = regexp.MustCompile(`(?m)^<!-- /gentle-ai:remote-authorization -->[ \t]*\r?$`)

var retiredRemoteAuthorizationSections = []string{
	`## Remote operation authorization

Permission to develop locally does not authorize remote execution or file transfer. Before remote work, require explicit user authorization for the destination, operation, and credential/session to use. If any part is missing or ambiguous, ask and remain local; do not probe the destination to resolve the ambiguity.

- Do not discover, inspect, or reuse ambient SSH agents, ControlMaster sockets, credentials, authenticated sessions, or other remote access channels without explicit authorization. Their availability is not permission to use them.
- Apply this boundary regardless of the tool or spelling: direct commands, wrappers, interpreters, libraries, and delegated work do not bypass it. Pass the authorized scope to delegates; delegation cannot expand it.
- Explicitly authorized remote work is allowed within that scope. Preserve stricter user instructions and runtime restrictions; do not weaken them or change approval settings to proceed.
- Native ask rules are an additional runtime mechanism, not authorization inferred from local-development access. Automation modes and remembered approvals may suppress prompts. This behavioral contract is not a sandbox and does not guarantee a fresh human prompt for every execution.`,
	`## Remote operation authorization

An explicit user request to perform a remote operation authorizes use of the tool's normally configured authentication only for the requested operation and the destination the user indicates or clearly implies. Do not ask again solely because that authentication is an existing session.

- If the destination, operation, or account is ambiguous, ask before any remote action. Do not probe the destination to resolve ambiguity.
- Do not discover, inspect, select, or switch ambient credentials, agents, sockets, authenticated sessions, or other access channels. Their availability never constitutes authorization.
- Keep authorization limited to the requested operation, destination, and account. Do not expand scope through wrappers, interpreters, libraries, or delegated work; pass the same exact scope to delegates.
- Native ask rules remain an independent runtime mechanism. Automation modes and remembered approvals may suppress prompts, but they do not broaden this authorization contract. This contract is not a sandbox and does not guarantee a fresh human prompt for every execution.`,
}

func (s *applyPolicyState) prepareRemoteAuthorizationRemoval() error {
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
			if isLegacyNamedAgentKey(key) {
				contains, err := s.namedProfileHasRemoteAuthorization(prompt, filepath.Dir(s.configPath))
				if err != nil {
					return upstreamRuntimeError("cannot assess unmanaged named profile agent %q: %v", key, err)
				}
				if contains {
					return upstreamRuntimeError("unmanaged named profile agent %q contains remote-authorization instructions; declare or remove that profile before retrying", key)
				}
			}
			continue
		}
		path, isFile, err := s.runtimePromptFilePath(prompt)
		if err != nil {
			return upstreamRuntimeError("maintained agent %q has an invalid prompt reference: %v", key, err)
		}
		if !isFile {
			updated, err := removeRemoteAuthorizationSection(prompt)
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
		patch, err := prepareRemoteAuthorizationFile(path, filepath.Dir(s.configPath))
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
		patch, err := prepareRemoteAuthorizationFile(path, filepath.Dir(s.configPath))
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
	for _, actor := range append(append([]string{}, judgmentDayActorKeys...), reviewActorKeys...) {
		keys[actor] = true
	}
	return keys
}

func isLegacyNamedAgentKey(key string) bool {
	if isLegacySDDAgentKey(key) {
		return true
	}
	for _, actor := range judgmentDayActorKeys {
		if strings.HasPrefix(key, actor+"-") && len(key) > len(actor)+1 {
			return true
		}
	}
	return false
}

func (s *applyPolicyState) namedProfileHasRemoteAuthorization(prompt, runtimeRoot string) (bool, error) {
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
	updated, err := removeRemoteAuthorizationSection(content)
	return err != nil || updated != content, err
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

func prepareRemoteAuthorizationFile(path, runtimeRoot string) (preparedRuntimePatch, error) {
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
	updated, err := removeRemoteAuthorizationSection(string(raw))
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

func removeRemoteAuthorizationSection(content string) (string, error) {
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
	if strings.Contains(content, "<!-- gentle-ai:remote-authorization -->") != (len(remoteAuthorizationOpenMarker.FindAllStringIndex(content, -1)) > 0) || strings.Contains(content, "<!-- /gentle-ai:remote-authorization -->") != (len(remoteAuthorizationCloseMarker.FindAllStringIndex(content, -1)) > 0) {
		return "", fmt.Errorf("malformed remote-authorization markers")
	}
	openMarkers := remoteAuthorizationOpenMarker.FindAllStringIndex(content, -1)
	closeMarkers := remoteAuthorizationCloseMarker.FindAllStringIndex(content, -1)
	if len(starts) == 0 && len(openMarkers) == 0 && len(closeMarkers) == 0 {
		return content, nil
	}
	if len(starts) != 1 || !remoteAuthorizationHeading.MatchString(content) {
		return "", fmt.Errorf("expected at most one complete remote-authorization section, found %d", len(starts))
	}
	start := starts[0]
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
		if err := validateRetiredRemoteAuthorizationSection(content[start:closeStart]); err != nil {
			return "", err
		}
		return content[:openMarkers[0][0]] + content[closeMarkers[0][1]:], nil
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
	if err := validateRetiredRemoteAuthorizationSection(oldSection); err != nil {
		return "", err
	}
	withoutTrailingSpace := strings.TrimRight(oldSection, " \t\r\n")
	separator := oldSection[len(withoutTrailingSpace):]
	return content[:start] + separator + content[end:], nil
}

func validateRetiredRemoteAuthorizationSection(section string) error {
	trimmed := strings.TrimSpace(section)
	for _, known := range retiredRemoteAuthorizationSections {
		if trimmed == strings.TrimSpace(known) {
			return nil
		}
	}
	return fmt.Errorf("remote-authorization section does not match a known upstream or previously relaxed section")
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
