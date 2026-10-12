package overlay

import (
	"fmt"
	"sort"
	"strings"
)

type actorAssignment struct {
	Model   string
	Variant string
	Inherit bool
}

var judgmentDayActorKeys = []string{"jd-judge-a", "jd-judge-b", "jd-fix-agent"}

var reviewActorKeys = []string{"review-risk", "review-readability", "review-reliability", "review-resilience", "review-refuter", "review-validator"}

func validateActorAssignment(value any, label string, allowInherit bool) (actorAssignment, error) {
	assignment, ok := value.(map[string]any)
	if !ok {
		return actorAssignment{}, fmt.Errorf("%s: must be an object with either \"model\" and optional \"variant\", or explicit \"inherit\": true", label)
	}
	for key := range assignment {
		if key != "model" && key != "variant" && key != "inherit" {
			return actorAssignment{}, fmt.Errorf("%s: unexpected fields [%s]; only \"model\", \"variant\", and \"inherit\" are allowed", label, strings.Join(sortedStringKeys(assignment, "model", "variant", "inherit"), ", "))
		}
	}
	if rawInherit, exists := assignment["inherit"]; exists {
		inherit, ok := rawInherit.(bool)
		if !ok || !inherit {
			return actorAssignment{}, fmt.Errorf("%s: field \"inherit\" must be exactly true", label)
		}
		if !allowInherit {
			return actorAssignment{}, fmt.Errorf("%s: explicit inheritance is not allowed", label)
		}
		if _, hasModel := assignment["model"]; hasModel {
			return actorAssignment{}, fmt.Errorf("%s: \"inherit\" cannot be combined with \"model\"", label)
		}
		if _, hasVariant := assignment["variant"]; hasVariant {
			return actorAssignment{}, fmt.Errorf("%s: \"inherit\" cannot be combined with \"variant\"", label)
		}
		return actorAssignment{Inherit: true}, nil
	}
	model, ok := assignment["model"].(string)
	if !ok || strings.TrimSpace(model) == "" {
		return actorAssignment{}, fmt.Errorf("%s: field \"model\" must be a non-empty string", label)
	}
	var variant string
	if rawVariant, exists := assignment["variant"]; exists {
		value, ok := rawVariant.(string)
		if !ok {
			return actorAssignment{}, fmt.Errorf("%s: field \"variant\" must be a string (use \"\" for no variant)", label)
		}
		variant = value
	}
	return actorAssignment{Model: strings.TrimSpace(model), Variant: variant}, nil
}

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

func (s *applyPolicyState) reconcileGlobalActors() {
	for key, assignment := range s.resolvedGlobalActors {
		agent, _ := jsonObject(s.agents[key])
		changed := false
		if assignment.Inherit {
			for _, field := range []string{"model", "variant"} {
				if _, exists := agent[field]; exists {
					delete(agent, field)
					changed = true
					s.recordVerbose(s.configPath, fmt.Sprintf("agent.%s.%s removed for dynamic inheritance", key, field))
				}
			}
		} else {
			if jsonString(agent["model"]) != assignment.Model {
				agent["model"] = assignment.Model
				changed = true
			}
			if assignment.Variant != "" {
				if jsonString(agent["variant"]) != assignment.Variant {
					agent["variant"] = assignment.Variant
					changed = true
				}
			} else if _, exists := agent["variant"]; exists {
				delete(agent, "variant")
				changed = true
			}
		}
		if changed {
			s.configChanged = true
		}
	}
}

func verifyActorAssignment(agents map[string]any, key string, assignment actorAssignment) error {
	actual, ok := jsonObject(agents[key])
	if !ok {
		return fmt.Errorf("agent missing from OpenCode config after write")
	}
	if assignment.Inherit {
		if _, exists := actual["model"]; exists {
			return fmt.Errorf("model remains after inheritance reconciliation")
		}
		if _, exists := actual["variant"]; exists {
			return fmt.Errorf("variant remains after inheritance reconciliation")
		}
		return nil
	}
	if jsonString(actual["model"]) != assignment.Model {
		return fmt.Errorf("model is %q after write, expected %q", jsonString(actual["model"]), assignment.Model)
	}
	if jsonString(actual["variant"]) != assignment.Variant {
		return fmt.Errorf("variant is %q after write, expected %q", jsonString(actual["variant"]), assignment.Variant)
	}
	return nil
}
