package overlay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
)

// reconcilePrunedSkillPermissions denies policy-pruned skills that do not have
// a repo-owned replacement intentionally installed in the same runtime.
// Global rules cover native agents; agent rules are required because OpenCode
// merges them after global permissions.
func (s *applyPolicyState) reconcilePrunedSkillPermissions() error {
	s.effectiveSkillDenies = effectivePrunedSkillDenies(s.policy.Skills.Prune, s.managedTarget.RepoOwnedSharedSkills.Allowlist)
	if len(s.effectiveSkillDenies) == 0 {
		return nil
	}

	changed, err := reconcileSkillPermissionRules(s.configData, s.effectiveSkillDenies)
	if err != nil {
		return fmt.Errorf("reconcile global OpenCode skill permissions: %w", err)
	}
	if changed {
		s.configChanged = true
		s.recordVerbose(s.configPath, "permission.skill: enforced denies for policy-pruned skills")
	}

	keys := make([]string, 0, len(s.agents))
	for key := range s.agents {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		agent, ok := jsonObject(s.agents[key])
		if !ok {
			return fmt.Errorf("agent %q is not an object", key)
		}
		changed, err := reconcileSkillPermissionRules(agent, s.effectiveSkillDenies)
		if err != nil {
			return fmt.Errorf("reconcile OpenCode skill permissions for agent %q: %w", key, err)
		}
		if changed {
			s.configChanged = true
			s.recordVerbose(s.configPath, fmt.Sprintf("agent.%s.permission.skill: enforced denies for policy-pruned skills", key))
		}
	}

	s.metrics.prunedSkillPermissionNames = len(s.effectiveSkillDenies)
	s.metrics.prunedSkillPermissionScopes = len(keys) + 1
	return nil
}

func effectivePrunedSkillDenies(pruned, repoOwnedAllowlist []string) []string {
	repoOwned := make(map[string]bool, len(repoOwnedAllowlist))
	for _, skill := range repoOwnedAllowlist {
		repoOwned[skill] = true
	}
	seen := map[string]bool{}
	denies := make([]string, 0, len(pruned))
	for _, skill := range pruned {
		if skill == "" || repoOwned[skill] || seen[skill] {
			continue
		}
		seen[skill] = true
		denies = append(denies, skill)
	}
	sort.Strings(denies)
	return denies
}

func reconcileSkillPermissionRules(container map[string]any, pruned []string) (bool, error) {
	permission, changed, err := normalizePermissionObject(container)
	if err != nil {
		return false, err
	}
	skillRules, skillChanged, err := normalizeSkillPermissionRules(permission)
	if err != nil {
		return false, err
	}
	changed = changed || skillChanged
	for _, skill := range pruned {
		if jsonString(skillRules[skill]) == "deny" {
			continue
		}
		skillRules[skill] = "deny"
		changed = true
	}
	return changed, nil
}

func normalizePermissionObject(container map[string]any) (map[string]any, bool, error) {
	value, exists := container["permission"]
	if !exists {
		permission := map[string]any{}
		container["permission"] = permission
		return permission, true, nil
	}
	if permission, ok := jsonObject(value); ok {
		return permission, false, nil
	}
	if action, ok := value.(string); ok {
		permission := map[string]any{"*": action}
		container["permission"] = permission
		return permission, true, nil
	}
	return nil, false, fmt.Errorf("permission must be an object or action string")
}

func normalizeSkillPermissionRules(permission map[string]any) (map[string]any, bool, error) {
	value, exists := permission["skill"]
	if !exists {
		rules := map[string]any{}
		permission["skill"] = rules
		return rules, true, nil
	}
	if rules, ok := jsonObject(value); ok {
		return rules, false, nil
	}
	if action, ok := value.(string); ok {
		rules := map[string]any{"*": action}
		permission["skill"] = rules
		return rules, true, nil
	}
	return nil, false, fmt.Errorf("permission.skill must be an object or action string")
}

func verifyPrunedSkillPermissions(configData map[string]any, agents map[string]any, pruned []string) error {
	if len(pruned) == 0 {
		return nil
	}
	if err := verifySkillPermissionRules(configData, pruned); err != nil {
		return fmt.Errorf("global OpenCode skill permissions: %w", err)
	}
	for key, value := range agents {
		agent, ok := jsonObject(value)
		if !ok {
			return fmt.Errorf("agent %q is not an object", key)
		}
		if err := verifySkillPermissionRules(agent, pruned); err != nil {
			return fmt.Errorf("agent %q skill permissions: %w", key, err)
		}
	}
	return nil
}

func verifySkillPermissionRules(container map[string]any, pruned []string) error {
	permission, ok := jsonObject(container["permission"])
	if !ok {
		return fmt.Errorf("permission object missing")
	}
	rules, ok := jsonObject(permission["skill"])
	if !ok {
		return fmt.Errorf("permission.skill object missing")
	}
	for _, skill := range pruned {
		if action := jsonString(rules[skill]); action != "deny" {
			return fmt.Errorf("%q is %q, want deny", skill, action)
		}
	}
	return nil
}

// marshalOpenCodeConfig places effective external-skill denies last in each
// skill permission object. OpenCode resolves the last matching permission rule,
// so this preserves unrelated wildcard rules without allowing them to override
// an overlay denial after the config is written.
func marshalOpenCodeConfig(data map[string]any, pruned []string) ([]byte, error) {
	var buffer bytes.Buffer
	if err := writeOpenCodeJSONValue(&buffer, data, 0, openCodeJSONRoot, pruned); err != nil {
		return nil, err
	}
	buffer.WriteByte('\n')
	return buffer.Bytes(), nil
}

type openCodeJSONKind int

const (
	openCodeJSONNormal openCodeJSONKind = iota
	openCodeJSONRoot
	openCodeJSONAgents
	openCodeJSONAgent
	openCodeJSONPermission
	openCodeJSONSkillPermission
)

func writeOpenCodeJSONValue(buffer *bytes.Buffer, value any, depth int, kind openCodeJSONKind, pruned []string) error {
	if object, ok := jsonObject(value); ok {
		return writeOpenCodeJSONObject(buffer, object, depth, kind, pruned)
	}
	if values, ok := value.([]any); ok {
		if len(values) == 0 {
			buffer.WriteString("[]")
			return nil
		}
		buffer.WriteString("[\n")
		for index, item := range values {
			writeOpenCodeIndent(buffer, depth+1)
			if err := writeOpenCodeJSONValue(buffer, item, depth+1, openCodeJSONNormal, pruned); err != nil {
				return err
			}
			if index < len(values)-1 {
				buffer.WriteByte(',')
			}
			buffer.WriteByte('\n')
		}
		writeOpenCodeIndent(buffer, depth)
		buffer.WriteByte(']')
		return nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	buffer.Write(raw)
	return nil
}

func writeOpenCodeJSONObject(buffer *bytes.Buffer, object map[string]any, depth int, kind openCodeJSONKind, pruned []string) error {
	keys := orderedOpenCodeJSONKeys(object, kind, pruned)
	if len(keys) == 0 {
		buffer.WriteString("{}")
		return nil
	}
	buffer.WriteString("{\n")
	for index, key := range keys {
		writeOpenCodeIndent(buffer, depth+1)
		encodedKey, err := json.Marshal(key)
		if err != nil {
			return err
		}
		buffer.Write(encodedKey)
		buffer.WriteString(": ")
		if err := writeOpenCodeJSONValue(buffer, object[key], depth+1, nextOpenCodeJSONKind(kind, key), pruned); err != nil {
			return err
		}
		if index < len(keys)-1 {
			buffer.WriteByte(',')
		}
		buffer.WriteByte('\n')
	}
	writeOpenCodeIndent(buffer, depth)
	buffer.WriteByte('}')
	return nil
}

func orderedOpenCodeJSONKeys(object map[string]any, kind openCodeJSONKind, pruned []string) []string {
	prunedSet := map[string]bool{}
	if kind == openCodeJSONSkillPermission {
		for _, skill := range pruned {
			prunedSet[skill] = true
		}
	}
	keys := make([]string, 0, len(object))
	for key := range object {
		if kind == openCodeJSONPermission && key == "skill" {
			continue
		}
		if prunedSet[key] {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if kind == openCodeJSONPermission {
		if _, exists := object["skill"]; exists {
			keys = append(keys, "skill")
		}
	}
	if kind == openCodeJSONSkillPermission {
		prunedKeys := make([]string, 0, len(prunedSet))
		for skill := range prunedSet {
			if _, exists := object[skill]; exists {
				prunedKeys = append(prunedKeys, skill)
			}
		}
		sort.Strings(prunedKeys)
		keys = append(keys, prunedKeys...)
	}
	return keys
}

func nextOpenCodeJSONKind(parent openCodeJSONKind, key string) openCodeJSONKind {
	switch {
	case parent == openCodeJSONRoot && key == "agent":
		return openCodeJSONAgents
	case (parent == openCodeJSONRoot || parent == openCodeJSONAgent) && key == "permission":
		return openCodeJSONPermission
	case parent == openCodeJSONPermission && key == "skill":
		return openCodeJSONSkillPermission
	case parent == openCodeJSONAgents:
		return openCodeJSONAgent
	default:
		return openCodeJSONNormal
	}
}

func writeOpenCodeIndent(buffer *bytes.Buffer, depth int) {
	buffer.WriteString(string(bytes.Repeat([]byte("  "), depth)))
}
