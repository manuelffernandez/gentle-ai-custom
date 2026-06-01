package overlay

import (
	"fmt"
	"regexp"
	"strings"
)

// sanitizePrompt removes upstream PR/review workflow content from an orchestrator
// prompt, then verifies required markers are present and forbidden markers are absent.
func sanitizePrompt(text string, policy Policy) (string, error) {
	for _, marker := range policy.Sanitizer.RequiredMarkers {
		if !strings.Contains(text, marker) {
			return "", fmt.Errorf("missing required marker before sanitizing: %s", marker)
		}
	}

	var err error
	text, err = replaceOnce(text, "3. **Chained PR strategy**: `auto-forecast`, `ask-always`, `single-pr-default`, or `force-chained`.\n4. **Review budget**: maximum changed lines before stopping for reviewer-burden approval.\n", "", "preflight PR/review choices")
	if err != nil {
		return "", err
	}
	text, err = replaceOnce(text, `Reply with "use recommended" or with codes like: A1, B1, C1, D1.`, `Reply with "use recommended" or with codes like: A1, B1.`, "english preflight codes")
	if err != nil {
		return "", err
	}
	text, err = replaceOnce(text, `Respondé con "usar recomendado" o con códigos como: A1, B1, C1, D1.`, `Respondé con "usar recomendado" o con códigos como: A1, B1.`, "spanish preflight codes")
	if err != nil {
		return "", err
	}
	text, err = removeBlock(text, `(?ms)^C\. PRs\n.*?^   D3 Other: ask for the number afterwards\.\n`, "english PR/review prompt block")
	if err != nil {
		return "", err
	}
	text, err = removeBlock(text, `(?ms)^C\. PRs\n.*?^   D3 Otro: preguntar el número después\.\n`, "spanish PR/review prompt block")
	if err != nil {
		return "", err
	}
	text, err = removeLine(text, `(?m)^- PRs:.*\n`, "PR answer mapping")
	if err != nil {
		return "", err
	}
	text, err = removeLine(text, `(?m)^- Review:.*\n`, "review answer mapping")
	if err != nil {
		return "", err
	}
	text, err = replaceOnce(text, "If the user explicitly provided all four choices in the current conversation, summarize them as the session preflight block and continue.", "If the user explicitly provided both choices in the current conversation, summarize them as the session preflight block and continue.", "all four choices wording")
	if err != nil {
		return "", err
	}
	text, err = removeSection(text, "### Delivery Strategy\n", []string{"### Chain Strategy\n", "### Dependency Graph\n"}, "Delivery Strategy section")
	if err != nil {
		return "", err
	}
	text, err = removeSection(text, "### Chain Strategy\n", []string{"### Dependency Graph\n"}, "Chain Strategy section")
	if err != nil {
		return "", err
	}
	text, err = removeSection(text, "### Review Workload Guard (MANDATORY)\n", []string{"<!-- gentle-ai:sdd-model-assignments -->\n"}, "Review Workload Guard section")
	if err != nil {
		return "", err
	}
	text, err = replaceOnce(text, "3. The user explicitly asked to apply/continue implementation, or the prior SDD planning phase completed and the orchestrator has passed the review workload guard.", "3. The user explicitly asked to apply/continue implementation, or the prior SDD planning phase completed.", "apply routing review-workload clause")
	if err != nil {
		return "", err
	}

	for _, marker := range policy.Sanitizer.RequiredMarkers {
		if !strings.Contains(text, marker) {
			return "", fmt.Errorf("missing required marker after sanitizing: %s", marker)
		}
	}
	for _, marker := range policy.Sanitizer.ForbiddenMarkers {
		if strings.Contains(text, marker) {
			return "", fmt.Errorf("forbidden marker still present after sanitizing: %s", marker)
		}
	}
	return text, nil
}

func replaceOnce(text, old, new, label string) (string, error) {
	idx := strings.Index(text, old)
	if idx < 0 {
		return "", fmt.Errorf("missing expected text: %s", label)
	}
	return text[:idx] + new + text[idx+len(old):], nil
}

func removeBlock(text, pattern, label string) (string, error) {
	re := regexp.MustCompile(pattern)
	loc := re.FindStringIndex(text)
	if loc == nil {
		return "", fmt.Errorf("missing expected block: %s", label)
	}
	return text[:loc[0]] + text[loc[1]:], nil
}

func removeLine(text, pattern, label string) (string, error) {
	re := regexp.MustCompile(pattern)
	loc := re.FindStringIndex(text)
	if loc == nil {
		return "", fmt.Errorf("missing expected line: %s", label)
	}
	return text[:loc[0]] + text[loc[1]:], nil
}

func removeSection(text, startMarker string, endMarkers []string, label string) (string, error) {
	start := strings.Index(text, startMarker)
	if start < 0 {
		return "", fmt.Errorf("missing expected block: %s", label)
	}
	searchFrom := start + len(startMarker)
	end := -1
	for _, marker := range endMarkers {
		idx := strings.Index(text[searchFrom:], marker)
		if idx < 0 {
			continue
		}
		candidate := searchFrom + idx
		if end < 0 || candidate < end {
			end = candidate
		}
	}
	if end < 0 {
		return "", fmt.Errorf("missing expected block: %s", label)
	}
	return text[:start] + text[end:], nil
}
