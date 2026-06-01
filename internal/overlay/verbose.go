package overlay

import (
	"fmt"
	"strings"
)

type verboseRecorder struct {
	enabled bool
	order   []string
	entries map[string][]string
}

func newVerboseRecorder(enabled bool) *verboseRecorder {
	return &verboseRecorder{enabled: enabled, entries: map[string][]string{}}
}

func (r *verboseRecorder) record(path, detail string) {
	if r == nil || !r.enabled || path == "" || detail == "" {
		return
	}
	if _, ok := r.entries[path]; !ok {
		r.order = append(r.order, path)
	}
	r.entries[path] = append(r.entries[path], detail)
}

func (r *verboseRecorder) print() {
	if r == nil || !r.enabled {
		return
	}
	fmt.Println()
	fmt.Println("Verbose changes:")
	if len(r.order) == 0 {
		fmt.Println("  (no changes — all assets already up to date)")
		return
	}
	for _, path := range r.order {
		fmt.Printf("  - %s\n", path)
		for _, detail := range r.entries[path] {
			fmt.Printf("      - %s\n", detail)
		}
	}
}

func (s *applyPolicyState) recordVerbose(path, detail string) {
	if s == nil {
		return
	}
	s.recorder.record(path, detail)
}

func quotedValue(value string) string {
	if value == "" {
		return "<empty>"
	}
	return fmt.Sprintf("%q", value)
}

func summarizePromptValue(prompt string) string {
	trimmed := strings.TrimSpace(prompt)
	if trimmed == "" {
		return "<empty>"
	}
	if strings.HasPrefix(trimmed, "{file:") && strings.HasSuffix(trimmed, "}") {
		return trimmed
	}
	return "inline prompt"
}

func describeWriteStatus(status string) string {
	switch status {
	case "new":
		return "created"
	case "changed":
		return "updated"
	case "unchanged":
		return "rewrote with unchanged content"
	default:
		return status
	}
}

func shouldRecordWriteStatus(status string) bool {
	return status == "new" || status == "changed"
}
