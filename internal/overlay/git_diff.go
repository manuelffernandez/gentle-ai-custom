package overlay

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type GitDiffEntry struct {
	Status  string
	OldPath string
	Path    string
}

func runGitDiff(repo, baseCommit string) ([]GitDiffEntry, error) {
	if strings.TrimSpace(baseCommit) == "" {
		return nil, fmt.Errorf("last_maintained_commit is empty")
	}
	text, err := runGit(repo, false, "diff", "--name-status", "--find-renames", baseCommit+"..HEAD")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(text) == "" {
		return nil, nil
	}
	lines := strings.Split(strings.TrimSpace(text), "\n")
	entries := make([]GitDiffEntry, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			return nil, fmt.Errorf("unexpected git diff output line: %q", line)
		}
		status := strings.TrimSpace(parts[0])
		code := status
		if len(status) > 0 {
			code = status[:1]
		}
		entry := GitDiffEntry{Status: code}
		switch code {
		case "R":
			if len(parts) != 3 {
				return nil, fmt.Errorf("unexpected git rename output line: %q", line)
			}
			entry.OldPath = strings.TrimSpace(parts[1])
			entry.Path = strings.TrimSpace(parts[2])
		default:
			entry.Path = strings.TrimSpace(parts[1])
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func runGitShow(repo, commit, path string) (string, error) {
	if strings.TrimSpace(commit) == "" {
		commit = "HEAD"
	}
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("git show path is empty")
	}
	entry, err := rawGitBlobCommand(repo, "ls-tree", "-z", "--full-tree", commit, "--", path)
	if err != nil {
		return "", err
	}
	records := bytes.Split(entry, []byte{0})
	if len(records) != 2 || len(records[1]) != 0 {
		return "", fmt.Errorf("git show path is not a blob: %s", path)
	}
	fields := bytes.SplitN(records[0], []byte{'\t'}, 2)
	if len(fields) != 2 || string(fields[1]) != path {
		return "", fmt.Errorf("git show path is not a blob: %s", path)
	}
	metadata := strings.Fields(string(fields[0]))
	if len(metadata) != 3 || metadata[1] != "blob" {
		return "", fmt.Errorf("git show path is not a blob: %s", path)
	}
	if metadata[0] != "100644" && metadata[0] != "100755" {
		return "", fmt.Errorf("git show path is not a regular blob: %s", path)
	}
	output, err := rawGitBlobCommand(repo, "show", fmt.Sprintf("%s:%s", commit, path))
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func rawGitBlobCommand(repo string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
	output, err := cmd.Output()
	if err != nil {
		message := err.Error()
		if exit, ok := err.(*exec.ExitError); ok && len(exit.Stderr) != 0 {
			message = strings.TrimSpace(string(exit.Stderr))
		}
		return nil, fmt.Errorf("%s: %s", repo, message)
	}
	return output, nil
}
