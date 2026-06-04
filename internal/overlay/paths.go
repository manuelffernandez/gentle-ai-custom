package overlay

import (
	"os"
	"path/filepath"
	"strings"
)

func expandUser(pathValue string) string {
	if pathValue == "" || pathValue[0] != '~' {
		return pathValue
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return pathValue
	}
	if pathValue == "~" {
		return home
	}
	if strings.HasPrefix(pathValue, "~/") || strings.HasPrefix(pathValue, "~\\") {
		return filepath.Join(home, pathValue[2:])
	}
	return pathValue
}

func usageCommandName(subcommand string) string {
	if entrypoint := strings.TrimSpace(os.Getenv("GENTLE_AI_CUSTOM_ENTRYPOINT")); entrypoint != "" {
		return entrypoint
	}
	name := filepath.Base(os.Args[0])
	if strings.Contains(name, "gentle-ai-overlay") && subcommand != "" {
		return name + " " + subcommand
	}
	return name
}
