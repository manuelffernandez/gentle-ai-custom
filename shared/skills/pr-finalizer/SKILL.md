---
name: pr-finalizer
description: >
  Generate or regenerate pull request content from committed changes only, detect repository PR conventions and a sensible base branch,
  and optionally create or update the PR after explicit approval.
  Trigger: When the user asks to create, update, regenerate, or sync a pull request from committed changes.
license: Apache-2.0
metadata:
  author: manuelfernandez
  version: "1.0"
---

## When to Use

Use this skill when:

- local commits already exist and the user wants PR-ready content or PR synchronization
- the user asks things like `creame la PR`, `generá la PR`, `regenerá la PR`, `actualizá la PR`, or `prepará el body de la PR`
- the workflow must stay explicit and approval-driven after implementation or commit planning is already done

Do **not** auto-run this skill after SDD, after commits, or after push. This is a **post-commit finalization** step that must be explicitly requested.

## Critical Patterns

1. **Committed diff only**. The PR title and body must be based only on committed changes, never on unstaged or uncommitted work.
2. **Two modes only**:
   - `create` → draft PR content for the current committed branch state and optionally open a new PR
   - `regenerate` → regenerate PR content from scratch for an existing PR and optionally update it
3. **Convention detection order**:
   - `.github/PULL_REQUEST_TEMPLATE.md`
   - `docs/CONVENTIONS.md`
   - `CONTRIBUTING.md`
   - `README.md`
   - fallback generic PR template
4. **Base branch resolution order**:
   - explicit user input
   - in `regenerate` mode, resolved PR metadata (`baseRefName`)
   - remote default branch from GitHub metadata or local remote HEAD metadata
   - ask the user only if still ambiguous
5. **Never assume branch naming conventions** such as `development`, `main`, or `master`.
6. **Remote resolution order**:
   - explicit user input
   - `origin` when it exists and is usable
   - if `origin` is unavailable, ask the user to choose from the available remotes
7. **State-changing commands require explicit approval**. This includes `git fetch`, `git push`, temporary file writes, `gh pr create`, and `gh pr edit`.
8. **If remote branch state is stale**, stop and request push approval before PR creation or update.
9. **If `gh` is unavailable or unauthenticated**, return copy-paste-ready PR content and stop without pretending the PR was created or updated.
10. **Do not mention branch names in the PR description body** unless the repository template explicitly requires them.
11. **Do not propose commit messages** as part of this skill.
12. **Use temporary artifacts outside the repository by default**, for example `${TMPDIR:-/tmp}/gentle-ai-pr-diff.txt` and `${TMPDIR:-/tmp}/gentle-ai-pr-body.md`.

## Supported Inputs

Resolve these inputs from the user or caller context, asking only for what is missing:

- `mode`: `create` or `regenerate`
- `remote`: optional
- `head branch`: optional, default to current local branch
- `base branch`: optional, auto-detect if possible
- `PR number`: optional in `regenerate`, preferred if the user already knows it
- `diff output path`: optional, default to `${TMPDIR:-/tmp}/gentle-ai-pr-diff.txt`
- `body output path`: optional, default to `${TMPDIR:-/tmp}/gentle-ai-pr-body.md`

## Read-Only Evidence Gathering

These checks are read-only and may run without approval:

```bash
git remote -v
git rev-parse --abbrev-ref HEAD
git status -sb
git remote get-url <remote>
gh --version
gh auth status
```

For base branch resolution, prefer one or both of these read-only checks when available:

```bash
gh repo view --json defaultBranchRef
git symbolic-ref --quiet --short refs/remotes/<remote>/HEAD
```

In `regenerate` mode, resolve the target PR with one of these read-only commands:

```bash
gh pr view <pr-number> --json number,title,body,headRefName,baseRefName,url
gh pr view --json number,title,body,headRefName,baseRefName,url
```

In `create` mode, optionally detect whether a PR already exists for the head branch before attempting creation:

```bash
gh pr view --json number,title,headRefName,baseRefName,url
```

If repository-specific PR convention files exist, read them before generating content.

## Convention Rules

Apply these rules strictly:

- If `.github/PULL_REQUEST_TEMPLATE.md` exists, treat it as the preferred formatting source.
- If no explicit PR template exists, use repository docs only when they define a clear PR structure.
- If no clear repository convention exists, write in **English Markdown** using this fallback structure:

```md
## Summary

- ...

## Context

- ...

## Changes

- ...

## Technical Details

- ...

## Breaking Changes

- None.
```

- Do not invent changes not supported by the committed diff.
- Do not describe branch-only history or iterative drafting as if it were product behavior.

## Phase 1 — Generate Committed Diff

Before generating the committed diff, request one approval checkpoint for this command block because it updates remote-tracking refs and writes a temporary diff artifact:

```bash
git fetch --all --prune
git diff <remote>/<base-branch>...<head-branch> > <diff-output-path>
```

In `regenerate` mode, use the resolved PR base branch when available.

## Phase 2 — Evidence Gathering

After the diff artifact exists, gather focused evidence with read-only commands:

```bash
git diff --name-status <remote>/<base-branch>...<head-branch>
git diff --stat <remote>/<base-branch>...<head-branch>
```

If output is truncated, read the diff file from disk and treat it as the source of truth.

## Content Generation Rules

### `create` mode

- Describe the current net effect of the head branch relative to the chosen base branch.
- If an open PR already exists for the same head branch, stop and report that the caller should use `regenerate` mode instead of creating a duplicate.

### `regenerate` mode

- Regenerate the PR title and body **from scratch**.
- Treat the current committed net diff against the resolved base branch as the only source of truth.
- Do **not** reuse the previous PR title or body as factual input.

## Approval Workflow

Follow this loop exactly:

1. Generate a proposed PR title and PR body.
2. Return both in copy-paste-ready form plus an explicit approval prompt.
3. If the user rejects the proposal, incorporate feedback without violating diff evidence constraints and repeat.
4. After approval, verify whether the remote head branch reflects local `HEAD` with read-only checks such as:

```bash
git rev-parse --abbrev-ref HEAD
git rev-parse --verify HEAD
git rev-parse --verify <remote>/<head-branch>
git rev-list --left-right --count <remote>/<head-branch>...HEAD
```

5. If the remote branch is missing or stale, stop and request approval for:

```bash
git push -u <remote> HEAD
```

6. Only continue to GitHub CLI actions after the remote branch state is current enough for the requested PR action.

## PR Creation via CLI (`create` mode)

After the user approves the generated content:

1. Request separate approvals in this order:
   - writing the temporary PR body artifact
   - running `gh pr create`
2. Write the body to `<body-output-path>`.
3. Run:

```bash
gh pr create --base <base-branch> --head <head-branch> --title "<title>" --body-file <body-output-path>
```

4. If remotes map to a different repository than the current default, resolve `owner/repo` from the remote URL and use `--repo owner/repo`.

## PR Update via CLI (`regenerate` mode)

After the user approves the regenerated content:

1. Request separate approvals in this order:
   - writing the temporary PR body artifact
   - running `gh pr edit`
2. Write the body to `<body-output-path>`.
3. Run:

```bash
gh pr edit <pr-number> --title "<title>" --body-file <body-output-path>
```

4. If remotes map to a different repository than the current default, resolve `owner/repo` from the remote URL and use `--repo owner/repo`.

## Output Requirements

When presenting content:

- show the PR title clearly
- show the PR body clearly in copy-paste-ready Markdown
- say whether the output is for `create` or `regenerate`
- make approval boundaries explicit for temp-file writes, pushes, and GitHub CLI commands

## Safety and Edge Cases

- If the repository has uncommitted changes, warn that PR content excludes them because committed diff is the source of truth.
- If `<remote>/<base-branch>` does not exist, stop and ask the user to confirm the remote/base branch pair.
- If `regenerate` mode cannot resolve PR metadata and the user did not provide enough information, ask only for the missing minimum input.
- If the current local branch does not match the resolved PR head branch in `regenerate` mode, stop and explain the mismatch.
- Never claim that a PR was created or updated unless the corresponding CLI command actually succeeded.

## Commands

```bash
/pr-create
/pr-regenerate
```

These wrappers may be exposed differently by each agent surface, but the workflow stays the same: explicit invocation, diff evidence, and explicit approval before state changes.

## Resources

- Agent-level instruction surface when present (examples: `~/.config/opencode/AGENTS.md`, `~/.claude/CLAUDE.md`)
- Repo-local PR conventions when they exist (`.github/PULL_REQUEST_TEMPLATE.md`, `docs/CONVENTIONS.md`, `CONTRIBUTING.md`, `README.md`)
