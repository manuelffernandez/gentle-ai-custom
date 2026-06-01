package overlay

import (
	"encoding/json"
	"fmt"
	"os"
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

// validateLocalProfilesConfig reads and strictly validates the per-machine SDD
// profile config file. Returns the validated profiles or a descriptive error.
func validateLocalProfilesConfig(path string, sddPhases []string, sddPhasesSet map[string]bool) ([]validatedProfile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Cannot read local SDD profile config at %s: %v", path, err)
	}
	var data any
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("local SDD profile config at %s is not valid JSON: %v. Fix or remove the file before re-running this script.", path, err)
	}
	top, ok := data.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("local SDD profile config at %s must be a JSON object at the top level", path)
	}
	for key := range top {
		if key != "version" && key != "profiles" {
			return nil, fmt.Errorf("local SDD profile config at %s has unexpected top-level fields [%s]; only \"version\" and \"profiles\" are allowed", path, strings.Join(sortedStringKeys(top, "version", "profiles"), ", "))
		}
	}
	version, ok := top["version"].(float64)
	if !ok || version != 1 {
		return nil, fmt.Errorf("local SDD profile config at %s has unsupported \"version\" %v; expected 1", path, top["version"])
	}
	profilesRaw, ok := top["profiles"].([]any)
	if !ok || len(profilesRaw) == 0 {
		return nil, fmt.Errorf("local SDD profile config at %s must contain a non-empty \"profiles\" array", path)
	}

	seenNames := map[string]bool{}
	validated := make([]validatedProfile, 0, len(profilesRaw))
	namePattern := regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*$`)

	for idx, rawProfile := range profilesRaw {
		prefix := fmt.Sprintf("profiles[%d]", idx)
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
		return profileAssignment{}, fmt.Errorf("%s: must be an object with \"model\" and \"variant\"", label)
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
	variant, ok := assignment["variant"].(string)
	if !ok {
		return profileAssignment{}, fmt.Errorf("%s: field \"variant\" must be a string (use \"\" for no variant)", label)
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
	if !pathExists(s.localProfilesPath) {
		fmt.Printf("  no local SDD profile config at %s - SDD profiles untouched\n", s.localProfilesPath)
		return nil
	}

	profiles, err := validateLocalProfilesConfig(s.localProfilesPath, s.policy.OpenCode.SDDPhases, s.sddPhasesSet)
	if err != nil {
		return err
	}

	for _, profile := range profiles {
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
