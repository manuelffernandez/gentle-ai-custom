package overlay

import (
	"fmt"
	"strings"
)

type upstreamReleaseIdentity struct {
	Version string
	Tag     string
	Commit  string
}

func verifyUpstreamReleaseParity(upstreamRepo string) (upstreamReleaseIdentity, error) {
	version, err := detectGentleAIVersion()
	if err != nil {
		return upstreamReleaseIdentity{}, err
	}
	tag := "v" + version

	status, err := runGit(upstreamRepo, false, "status", "--porcelain=v1", "--untracked-files=all", "--ignore-submodules=none")
	if err != nil {
		return upstreamReleaseIdentity{}, fmt.Errorf("cannot inspect upstream checkout cleanliness: %w", err)
	}
	if status != "" {
		return upstreamReleaseIdentity{}, fmt.Errorf("upstream checkout is dirty; commit, stash, or remove all staged, unstaged, untracked, and submodule changes before retrying:\n%s", status)
	}

	commit, err := runGit(upstreamRepo, false, "rev-parse", "--verify", "refs/tags/"+tag+"^{commit}")
	if err != nil {
		return upstreamReleaseIdentity{}, fmt.Errorf("exact release tag refs/tags/%s for installed gentle-ai %s is missing or stale in %s; run `git -C %s fetch --tags`, verify the tag, and check out %s manually before retrying (nothing was fetched or checked out automatically)", tag, version, upstreamRepo, upstreamRepo, tag)
	}
	head, err := runGit(upstreamRepo, false, "rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return upstreamReleaseIdentity{}, fmt.Errorf("cannot resolve upstream HEAD commit: %w", err)
	}
	if head != commit {
		return upstreamReleaseIdentity{}, fmt.Errorf("upstream HEAD %s does not equal exact release %s commit %s for installed gentle-ai %s; fetch tags if needed, then check out %s manually before retrying (audit/sync never uses branch ancestry or nearest-tag matches)", head, tag, commit, version, tag)
	}

	return upstreamReleaseIdentity{
		Version: strings.TrimSpace(version),
		Tag:     tag,
		Commit:  commit,
	}, nil
}
