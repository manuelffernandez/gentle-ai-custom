package overlay

import (
	"fmt"
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
	return runGit(repo, false, "show", fmt.Sprintf("%s:%s", commit, path))
}
