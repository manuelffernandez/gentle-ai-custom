package overlay

import (
	"fmt"
	"regexp"
	"strings"
)

type upstreamReleaseIdentity struct {
	InstalledVersion string
	Version          string
	Tag              string
	Commit           string
}

var releaseVersionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$`)

func verifyUpstreamReleaseParity(upstreamRepo, selectedTag string) (upstreamReleaseIdentity, error) {
	installedVersion, err := detectGentleAIVersion()
	if err != nil {
		return upstreamReleaseIdentity{}, err
	}
	installedVersion = strings.TrimSpace(installedVersion)
	tag := "v" + installedVersion
	if selectedTag != "" {
		if !isValidReleaseTag(selectedTag) {
			return upstreamReleaseIdentity{}, fmt.Errorf("invalid release %q; expected an exact v<V> release tag such as v2.9.1", selectedTag)
		}
		tag = selectedTag
	}

	status, err := runGit(upstreamRepo, false, "status", "--porcelain=v1", "--untracked-files=all", "--ignore-submodules=none")
	if err != nil {
		return upstreamReleaseIdentity{}, fmt.Errorf("cannot inspect upstream checkout cleanliness: %w", err)
	}
	if status != "" {
		return upstreamReleaseIdentity{}, fmt.Errorf("upstream checkout is dirty; commit, stash, or remove all staged, unstaged, untracked, and submodule changes before retrying:\n%s", status)
	}

	commit, err := runGit(upstreamRepo, false, "rev-parse", "--verify", "refs/tags/"+tag+"^{commit}")
	if err != nil {
		return upstreamReleaseIdentity{}, fmt.Errorf("exact release tag refs/tags/%s is missing or stale in %s; run `git -C %s fetch --tags`, verify the tag, and check out %s manually before retrying (nothing was fetched or checked out automatically)", tag, upstreamRepo, upstreamRepo, tag)
	}
	head, err := runGit(upstreamRepo, false, "rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return upstreamReleaseIdentity{}, fmt.Errorf("cannot resolve upstream HEAD commit: %w", err)
	}
	if head != commit {
		return upstreamReleaseIdentity{}, fmt.Errorf("upstream HEAD %s does not equal exact release %s commit %s; fetch tags if needed, then check out %s manually before retrying (audit/sync never uses branch ancestry or nearest-tag matches)", head, tag, commit, tag)
	}

	return upstreamReleaseIdentity{
		InstalledVersion: installedVersion,
		Version:          strings.TrimPrefix(tag, "v"),
		Tag:              tag,
		Commit:           commit,
	}, nil
}

func isValidReleaseTag(tag string) bool {
	return tag == strings.TrimSpace(tag) && strings.HasPrefix(tag, "v") && releaseVersionPattern.MatchString(strings.TrimPrefix(tag, "v"))
}
