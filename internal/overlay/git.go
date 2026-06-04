package overlay

import (
	"fmt"
	"os/exec"
	"strings"
)

func runGit(repo string, allowFailure bool, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	output, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(output))
	if err != nil {
		if allowFailure {
			return "", nil
		}
		if text == "" {
			text = err.Error()
		}
		return "", fmt.Errorf("%s: %s", repo, text)
	}
	return text, nil
}
