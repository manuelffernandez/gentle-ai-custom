package overlay

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type profileAssignment struct {
	Model   string
	Variant string
}

type validatedProfile struct {
	Name         string
	Orchestrator profileAssignment
	Phases       map[string]profileAssignment
}

func validateProfilesValue(value any, label string, sddPhases []string, sddPhasesSet map[string]bool) ([]validatedProfile, error) {
	profilesRaw, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%s: must be an array", label)
	}

	seenNames := map[string]bool{}
	validated := make([]validatedProfile, 0, len(profilesRaw))
	namePattern := regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

	for idx, rawProfile := range profilesRaw {
		prefix := fmt.Sprintf("%s[%d]", label, idx)
		profile, ok := rawProfile.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s: must be an object", prefix)
		}
		for key := range profile {
			if key != "name" && key != "orchestrator" && key != "phases" {
				return nil, fmt.Errorf("%s: unexpected fields [%s]; only \"name\", \"orchestrator\", \"phases\" are allowed", prefix, strings.Join(sortedStringKeys(profile, "name", "orchestrator", "phases"), ", "))
			}
		}
		name, ok := profile["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf("%s: \"name\" must be a non-empty string", prefix)
		}
		if !namePattern.MatchString(name) {
			return nil, fmt.Errorf("%s: \"name\" %q must match ^[a-z0-9][a-z0-9._-]*$ to be safe as an agent-key suffix", prefix, name)
		}
		if seenNames[name] {
			return nil, fmt.Errorf("%s: duplicate profile name %q", prefix, name)
		}
		seenNames[name] = true

		orchestrator, err := validateAssignment(profile["orchestrator"], prefix+".orchestrator")
		if err != nil {
			return nil, err
		}
		phasesObj, ok := profile["phases"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s.phases: must be an object keyed by SDD phase name", prefix)
		}
		phaseKeys := map[string]bool{}
		for phaseName := range phasesObj {
			phaseKeys[phaseName] = true
		}
		var missing []string
		for _, phaseName := range sddPhases {
			if !phaseKeys[phaseName] {
				missing = append(missing, phaseName)
			}
		}
		if len(missing) > 0 {
			return nil, fmt.Errorf("%s.phases: missing required phases %v (no defaults are inherited)", prefix, missing)
		}
		var unknown []string
		for phaseName := range phaseKeys {
			if !sddPhasesSet[phaseName] {
				unknown = append(unknown, phaseName)
			}
		}
		sort.Strings(unknown)
		if len(unknown) > 0 {
			return nil, fmt.Errorf("%s.phases: unknown phases %v; allowed: %v", prefix, unknown, sddPhases)
		}
		validatedPhases := map[string]profileAssignment{}
		for _, phaseName := range sddPhases {
			assignment, err := validateAssignment(phasesObj[phaseName], prefix+".phases."+phaseName)
			if err != nil {
				return nil, err
			}
			validatedPhases[phaseName] = assignment
		}
		validated = append(validated, validatedProfile{Name: name, Orchestrator: orchestrator, Phases: validatedPhases})
	}

	return validated, nil
}

func validateAssignment(value any, label string) (profileAssignment, error) {
	assignment, ok := value.(map[string]any)
	if !ok {
		return profileAssignment{}, fmt.Errorf("%s: must be an object with \"model\" and optional \"variant\"", label)
	}
	for key := range assignment {
		if key != "model" && key != "variant" {
			return profileAssignment{}, fmt.Errorf("%s: unexpected fields [%s]; only \"model\" and \"variant\" are allowed", label, strings.Join(sortedStringKeys(assignment, "model", "variant"), ", "))
		}
	}
	model, ok := assignment["model"].(string)
	if !ok || model == "" {
		return profileAssignment{}, fmt.Errorf("%s: field \"model\" must be a non-empty string", label)
	}
	var variant string
	if rawVariant, exists := assignment["variant"]; exists {
		v, ok := rawVariant.(string)
		if !ok {
			return profileAssignment{}, fmt.Errorf("%s: field \"variant\" must be a string (use \"\" for no variant)", label)
		}
		variant = v
	}
	return profileAssignment{Model: model, Variant: variant}, nil
}

// sortedStringKeys returns the keys in values that are not in the allowed set,
// sorted alphabetically. Used to build descriptive "unexpected fields" error messages.
func sortedStringKeys(values map[string]any, allowed ...string) []string {
	allowedSet := map[string]bool{}
	for _, key := range allowed {
		allowedSet[key] = true
	}
	var extras []string
	for key := range values {
		if !allowedSet[key] {
			extras = append(extras, key)
		}
	}
	sort.Strings(extras)
	return extras
}

// reconcileProfiles applies the local SDD profile config to the agents map,
// creating or updating orchestrator and phase agents for each managed profile.
func (s *applyPolicyState) reconcileProfiles() error {
	if s.resolvedDefaultProfile != nil {
		if err := s.reconcileBaseProfile(*s.resolvedDefaultProfile); err != nil {
			return err
		}
	}
	if !s.profilesDefined {
		fmt.Printf("  no named SDD profiles declared in local config - named SDD profiles untouched\n")
		return nil
	}
	if len(s.resolvedProfiles) == 0 {
		fmt.Printf("  local profile config at %s declares no managed SDD profiles\n", s.profileConfigSourcePath)
		return nil
	}

	for _, profile := range s.resolvedProfiles {
		s.managedProfiles[profile.Name] = true
		s.profilesManagedCount++

		orchKey := s.policy.OpenCode.ProfileOrchestratorPrefix + profile.Name
		if err := s.reconcileProfileAgent(profile.Name, orchKey, profile.Orchestrator, true); err != nil {
			return err
		}
		for _, phase := range s.policy.OpenCode.SDDPhases {
			phaseKey := phase + "-" + profile.Name
			if err := s.reconcileProfileAgent(profile.Name, phaseKey, profile.Phases[phase], false); err != nil {
				return err
			}
		}
	}

	discovered := map[string]bool{}
	for key := range s.originalAgentKeys {
		if s.isProfileOrchestrator(key) {
			name := strings.TrimPrefix(key, s.policy.OpenCode.ProfileOrchestratorPrefix)
			if name != "" {
				discovered[name] = true
			}
		}
	}
	var unmanaged []string
	for name := range discovered {
		if !s.managedProfiles[name] {
			unmanaged = append(unmanaged, name)
		}
	}
	sort.Strings(unmanaged)
	for _, name := range unmanaged {
		s.unmanagedProfiles = append(s.unmanagedProfiles, name)
		fmt.Printf("  unmanaged SDD profile present in opencode.json (left untouched): %s\n", name)
	}
	return nil
}

// reconcileBaseProfile applies the default_profile assignment to the base
// orchestrator and unsuffixed SDD phase agents (e.g. gentle-orchestrator,
// sdd-init, sdd-apply, ...).
// It increments profilesManagedCount but does NOT register in s.managedProfiles,
// because:
//   - s.managedProfiles tracks only named profiles (sdd-orchestrator-<name> family)
//   - the base profile agents use unmodified keys (gentle-orchestrator, sdd-<phase>)
//   - verifyPersistedState verifies the base separately via s.resolvedDefaultProfile
func (s *applyPolicyState) reconcileBaseProfile(profile validatedProfile) error {
	s.profilesManagedCount++
	if err := s.reconcileProfileAgent("default", s.policy.OpenCode.BaseOrchestratorKey, profile.Orchestrator, true); err != nil {
		return err
	}
	for _, phase := range s.policy.OpenCode.SDDPhases {
		if err := s.reconcileProfileAgent("default", phase, profile.Phases[phase], false); err != nil {
			return err
		}
	}
	return nil
}

func (s *applyPolicyState) reconcileProfileAgent(profileName, key string, assignment profileAssignment, orchestrator bool) error {
	existing, ok := jsonObject(s.agents[key])
	if !ok {
		agentObj := map[string]any{"model": assignment.Model}
		if assignment.Variant != "" {
			agentObj["variant"] = assignment.Variant
		}
		s.agents[key] = agentObj
		s.profileAgentsCreated++
		s.configChanged = true
		detail := fmt.Sprintf("profile %s: created %s with model %s", profileName, key, quotedValue(assignment.Model))
		if assignment.Variant != "" {
			detail += fmt.Sprintf(" and variant %s", quotedValue(assignment.Variant))
		}
		s.recordVerbose(s.configPath, detail)
		if orchestrator {
			fmt.Printf("  profile %s: created orchestrator agent %s (no prompt; run `gentle-ai sync` to materialize)\n", profileName, key)
		} else {
			suffix := ""
			if assignment.Variant != "" {
				suffix = fmt.Sprintf(" (%s)", assignment.Variant)
			}
			fmt.Printf("  profile %s: created phase agent %s -> %s%s\n", profileName, key, assignment.Model, suffix)
		}
		return nil
	}
	changed := false
	oldModel := jsonString(existing["model"])
	if oldModel != assignment.Model {
		existing["model"] = assignment.Model
		changed = true
		s.recordVerbose(s.configPath, fmt.Sprintf("profile %s: %s.model: %s -> %s", profileName, key, quotedValue(oldModel), quotedValue(assignment.Model)))
	}
	if assignment.Variant != "" {
		oldVariant := jsonString(existing["variant"])
		if oldVariant != assignment.Variant {
			existing["variant"] = assignment.Variant
			changed = true
			s.recordVerbose(s.configPath, fmt.Sprintf("profile %s: %s.variant: %s -> %s", profileName, key, quotedValue(oldVariant), quotedValue(assignment.Variant)))
		}
	} else if _, hasVariant := existing["variant"]; hasVariant {
		oldVariant := jsonString(existing["variant"])
		delete(existing, "variant")
		changed = true
		s.recordVerbose(s.configPath, fmt.Sprintf("profile %s: %s.variant removed %s", profileName, key, quotedValue(oldVariant)))
	}
	if changed {
		s.profileAgentsUpdated++
		s.configChanged = true
		suffix := ""
		if assignment.Variant != "" {
			suffix = fmt.Sprintf(" (%s)", assignment.Variant)
		}
		if orchestrator {
			fmt.Printf("  profile %s: updated orchestrator agent %s -> %s%s\n", profileName, key, assignment.Model, suffix)
		} else {
			fmt.Printf("  profile %s: updated phase agent %s -> %s%s\n", profileName, key, assignment.Model, suffix)
		}
	} else {
		s.profileAgentsSame++
	}
	return nil
}
