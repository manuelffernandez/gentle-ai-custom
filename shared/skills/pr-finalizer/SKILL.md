---
name: pr-finalizer
description: >
  Complement repository PR workflows by generating or regenerating pull request content from committed changes only,
  detecting repository PR conventions and a sensible base branch, and optionally creating or updating the PR after
  explicit approval.
  Trigger: When the user asks to draft, regenerate, update, or sync a pull request from committed changes.
license: Apache-2.0
metadata:
  author: manuelfernandez
  version: "1.3"
---

## When to Use

Use this skill when:

- local commits already exist and the user wants PR-ready content or PR synchronization
- the user asks things like `creame la PR`, `generá la PR`, `regenerá la PR`, `actualizá la PR`, or `prepará el body de la PR`
- the workflow must stay explicit and approval-driven after implementation or commit planning is already done
- repository governance is handled elsewhere, typically by repo docs, CI, or companion skills such as `branch-pr`

Do **not** auto-run this skill after SDD, after commits, or after push. This is a **post-commit finalization** step that must be explicitly requested.

## Non-Goals

This skill does **not** define or enforce repository governance. In particular, it does **not**:

- validate issue-first workflows or approval labels
- require or apply `type:*` labels
- validate branch naming conventions
- determine merge readiness
- replace contributor checklists, CI checks, or repo-specific review policy
- fork or duplicate the policy owned by upstream skills such as `branch-pr`

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
4. **Policy awareness without duplication**. If companion repo-policy skills exist (for example `branch-pr`), treat them as contextual guidance only. Do not copy, fork, or re-enforce their rules here.
5. **Base branch resolution order**:
   - explicit user input
   - in `regenerate` mode, resolved PR metadata (`baseRefName`)
   - remote default branch from GitHub metadata or local remote HEAD metadata
   - ask the user only if still ambiguous
6. **Never assume branch naming conventions** such as `development`, `main`, or `master`.
7. **Remote resolution order**:
   - explicit user input
   - `origin` when it exists and is usable
   - if `origin` is unavailable, ask the user to choose from the available remotes
8. **Read-only remote refresh is automatic**. Run `git fetch <remote>` without asking for approval so the skill can work from fresh remote-tracking refs.
9. **Sync gate comes before content generation**. If the remote head branch is missing or stale, stop and request approval for `git push -u <remote> HEAD` before generating the PR title/body.
10. **The only explicit approval in the happy path is the generated content**. After the user approves the proposed title and body, create or update the PR automatically; do not add a second approval checkpoint for `gh pr create` or `gh pr edit`.
11. **If `gh` is unavailable or unauthenticated**, return copy-paste-ready PR content and stop without pretending the PR was created or updated.
12. **Do not mention branch names in the PR description body** unless the repository template explicitly requires them.
13. **Do not propose commit messages** as part of this skill.
14. **Use temporary artifacts outside the repository by default**, for example `${TMPDIR:-/tmp}/pr-diff.txt` and `${TMPDIR:-/tmp}/pr-body.md`. After writing any temporary file, report the absolute path used — do not ask for approval.

## Supported Inputs

Resolve these inputs from the user or caller context, asking only for what is missing:

- `mode`: `create` or `regenerate`
- `remote`: optional
- `head branch`: optional, default to current local branch
- `base branch`: optional, auto-detect if possible
- `PR number`: optional in `regenerate`, preferred if the user already knows it
- `diff output path`: optional, default to `${TMPDIR:-/tmp}/pr-diff.txt`
- `body output path`: optional, default to `${TMPDIR:-/tmp}/pr-body.md`

## Read-Only Evidence Gathering

These checks are read-only and may run without approval:

```bash
git remote -v
git rev-parse --abbrev-ref HEAD
git status -sb
git remote get-url <remote>
git ls-remote --heads <remote> <head-branch>
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
- If companion policy skills or docs describe required issue links, labels, or checklists, preserve those requirements in the generated content when they are explicit, but do not invent or enforce them beyond what the repository already defines.
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

Before generating the committed diff, run `git fetch <remote>` automatically to refresh remote-tracking refs. If the remote head branch is missing or stale, stop and request approval for `git push -u <remote> HEAD` before continuing. The diff file write does not require approval — write it silently and report the absolute path:

```bash
git fetch <remote>
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

1. Refresh remote refs with `git fetch <remote>`.
2. Verify the remote head branch with read-only checks such as:

```bash
git ls-remote --heads <remote> <head-branch>
git rev-parse --abbrev-ref HEAD
git rev-parse --verify HEAD
git rev-list --left-right --count <remote>/<head-branch>...HEAD
```

3. If the remote branch is missing or stale, stop and request approval for:

```bash
git push -u <remote> HEAD
```

4. Only after the remote branch is current enough, generate a proposed PR title and PR body.
5. Return both in copy-paste-ready form plus an explicit approval prompt for the content only.
6. If the user rejects the proposal, incorporate feedback without violating diff evidence constraints and repeat.
7. After the user approves the generated content, continue automatically to GitHub CLI actions; do not ask for a second approval checkpoint.

## PR Creation via CLI (`create` mode)

After the user approves the generated content:

1. Write the body to `<body-output-path>` without requesting approval. Report the absolute path used.
2. Run automatically:

```bash
gh pr create --base <base-branch> --head <head-branch> --title "<title>" --body-file <body-output-path>
```

3. If remotes map to a different repository than the current default, resolve `owner/repo` from the remote URL and use `--repo owner/repo`.

## PR Update via CLI (`regenerate` mode)

After the user approves the regenerated content:

1. Write the body to `<body-output-path>` without requesting approval. Report the absolute path used.
2. Run automatically:

```bash
gh pr edit <pr-number> --title "<title>" --body-file <body-output-path>
```

3. If remotes map to a different repository than the current default, resolve `owner/repo` from the remote URL and use `--repo owner/repo`.

## Output Requirements

When presenting content:

- show the PR title clearly
- show the PR body clearly in copy-paste-ready Markdown
- say whether the output is for `create` or `regenerate`
 - make approval boundaries explicit for pushes and the content approval prompt only; temp-file writes and GitHub CLI execution happen automatically after content approval

## Safety and Edge Cases

- If the repository has uncommitted changes, warn that PR content excludes them because committed diff is the source of truth.
- If the remote head branch does not exist according to `git ls-remote --heads`, stop and request push approval before creating or updating the PR.
- If repository policy requirements appear to be missing (for example issue linkage or labels), warn about the gap when evidenced by templates/docs, but do not fabricate compliance.
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
- Companion repo-policy skills when present (examples: `branch-pr`) as contextual guidance, not as duplicated local policy
- Repo-local PR conventions when they exist (`.github/PULL_REQUEST_TEMPLATE.md`, `docs/CONVENTIONS.md`, `CONTRIBUTING.md`, `README.md`)
