---
name: branch-pr
description: "Create or regenerate pull requests from committed changes while respecting explicit repo PR policy; issue links and labels are contextual, not universal. Trigger: PR create/regenerate/update/sync requests."
license: Apache-2.0
metadata:
  author: manuelfernandez
  version: "1.0"
---

## When to Use

Use this skill when:

- local commits already exist and the user wants PR-ready content or PR synchronization
- the user asks to create, draft, regenerate, update, or sync a PR from committed changes
- the repository may have PR policy, templates, labels, branch rules, or chained-PR conventions that must be respected when they are explicitly declared

Do **not** auto-run this skill after SDD, after commits, or after push. This is an explicit PR finalization step.

## Critical Patterns

1. **This is the preferred final PR skill** for `/pr-create` and `/pr-regenerate` in this overlay.
2. **Two modes only**:
   - `create` -> draft PR content for the current committed branch state and optionally open a new PR
   - `regenerate` -> regenerate PR content from scratch for an existing PR and optionally update it
3. **Committed diff only**. The PR title and body must be based only on committed changes, never on unstaged or uncommitted work.
4. **MANDATORY repo-policy detection**. You MUST explicitly search for and read repository PR conventions before generating content. Order of precedence:
   - `.github/PULL_REQUEST_TEMPLATE.md`
   - `docs/CONVENTIONS.md`
   - `CONTRIBUTING.md`
   - `README.md`
   - companion skills or repo workflow docs when they are explicitly relevant, especially `chained-pr`
   - fallback generic PR template
5. **Repo policy is contextual, not universal**. Preserve and apply issue linkage, approval labels, PR type labels, branch naming rules, contributor checklists, and required checks only when the repository explicitly declares them. Do **not** invent global requirements such as mandatory approved issues, universal `Closes #N`, universal `status:approved`, or universal `type:*` labels.
6. **Base branch resolution order**:
   - explicit user input
   - in `regenerate` mode, resolved PR metadata (`baseRefName`)
   - remote default branch from GitHub metadata or local remote HEAD metadata
   - ask the user only if still ambiguous
7. **Remote resolution order**:
   - explicit user input
   - `origin` when it exists and is usable
   - ask the user to choose from available remotes only if `origin` is unavailable
8. **Safe command construction is mandatory**. Never build shell command strings by concatenating or interpolating remote names, branch names, titles, file paths, labels, or PR numbers. Prefer tool APIs or argument-vector execution. If a shell is unavoidable, use command arrays and validate/quote every argument.
9. **Read-only remote refresh is automatic**. Run `git fetch` for the resolved remote without asking so the skill works from fresh remote-tracking refs.
10. **Push gate comes before content generation**. If the remote head branch is missing or stale, stop and request approval for `git push -u <remote> HEAD` before generating the PR title/body.
11. **Duplicate PR detection is mandatory in `create` mode**. If an open PR already exists for the resolved head branch, or for the same head/base pair when the base is already known, stop and direct the user to `regenerate` mode instead.
12. **Chained PR context is preserved when it exists**. If the user, repo docs, or retained `chained-pr` workflow provide chain context, keep or add a `Chain Context` section without replacing the repository template. If stacked/chained work is detected but the base or chain context is ambiguous, stop and ask for explicit chain context. Do **not** invent tracker-PR automation.
13. **The only explicit approval in the happy path is the generated content**. After the user approves the proposed title/body, create or update the PR automatically. Do not add a second approval checkpoint for `gh pr create`, `gh pr edit`, temp-file writes, or repo-policy label updates.
14. **If `gh` is unavailable or unauthenticated**, return copy-paste-ready PR content and stop without pretending the PR was created or updated.
15. **Use unique temporary artifacts outside the repository**. Create temp files with a safe unique-name mechanism such as `mktemp` under `${TMPDIR:-/tmp}`; never use predictable fixed paths. Report the absolute path used after writing each artifact.
16. **Never invent changes or compliance**. If the diff, repo policy, or PR metadata do not support a claim, warn or ask instead of fabricating it.
17. **Language constraint**. Unless the user explicitly requests otherwise, all generated titles, PR bodies, and comments must be in **ENGLISH**.

## Supported Inputs

Resolve these inputs from the user or caller context, asking only for what is missing:

- `mode`: `create` or `regenerate`
- `remote`: optional
- `head branch`: optional, default to current local branch
- `base branch`: optional, auto-detect if possible
- `PR number`: optional in `regenerate`, preferred if the user already knows it
- `diff output path`: optional; default to a safely-created unique temp file outside the repo
- `body output path`: optional; default to a safely-created unique temp file outside the repo
- `chain context`: optional, but required when stacked/chained work is detected and repo policy does not resolve it

## Read-Only Evidence Gathering

These checks are read-only and may run without approval. Execute them with argument arrays or tool APIs, not concatenated shell strings:

```bash
git remote -v
git rev-parse --abbrev-ref HEAD
git status -sb
git remote get-url <remote>
git ls-remote --heads <remote> <head-branch>
gh --version
gh auth status
```

For base branch resolution, first resolve `owner/repo` from the selected remote when GitHub CLI is used. Then prefer one or both of these read-only checks when available:

```bash
gh repo view <owner/repo> --json defaultBranchRef
git symbolic-ref --quiet --short refs/remotes/<remote>/HEAD
```

In `regenerate` mode, resolve the target PR with one of these read-only commands:

```bash
gh pr view <pr-number> --repo <owner/repo> --json number,title,body,headRefName,baseRefName,url
gh pr view --repo <owner/repo> --json number,title,body,headRefName,baseRefName,url
```

In `create` mode, detect whether an **open** PR already exists for the head branch before attempting creation:

```bash
gh pr list --repo <owner/repo> --head <head-branch> --state open --json number,title,state,headRefName,baseRefName,url
```

## Repo Policy Interpretation

Apply these rules strictly:

- If `.github/PULL_REQUEST_TEMPLATE.md` exists, treat it as the preferred formatting source.
- If no explicit PR template exists, use repository docs only when they define a clear PR structure or policy.
- If the repository explicitly requires issue linkage, approved issue state, PR type labels, or branch naming, preserve those requirements and ask for the minimum missing input needed to comply.
- If the repository does **not** explicitly require issue linkage, do **not** fabricate `Closes #N`, `Fixes #N`, approved-issue state, or issue labels.
- If the repository explicitly requires a PR type label and it cannot be inferred safely from explicit user input or repo policy, stop and ask for the label value.
- If the repository explicitly requires a branch naming format and the current branch violates it, stop and explain the mismatch before generating PR content.
- If chained or stacked workflow guidance exists, preserve its required context, but never replace the repository template with a chain-only template.
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

## Phase 1 - Generate Committed Diff

Before generating the committed diff, run `git fetch` for the resolved remote automatically to refresh remote-tracking refs. If the remote head branch is missing or stale, stop and request approval for `git push -u <remote> HEAD` before continuing. The diff file write does not require approval.

Use safe argument execution:

- Program: `git`; args: `fetch`, `<remote>`
- Program: `git`; args: `diff`, `<remote>/<base-branch>...<head-branch>`; write stdout to `<diff-output-path>`

In `regenerate` mode, use the resolved PR base branch when available.

## Phase 2 - Evidence Gathering

After the diff artifact exists, gather focused evidence with read-only commands:

- Program: `git`; args: `diff`, `--name-status`, `<remote>/<base-branch>...<head-branch>`
- Program: `git`; args: `diff`, `--stat`, `<remote>/<base-branch>...<head-branch>`

If output is truncated, read the diff file from disk and treat it as the source of truth.

## Content Generation Rules

### `create` mode

- Describe the current net effect of the head branch relative to the chosen base branch.
- If an open PR already exists for the same head branch, and for the same base branch when that base is already known, stop and report that the caller should use `regenerate` mode instead of creating a duplicate.
- A PR that is `merged` or `closed` for that head branch does **not** block `create` mode.

### `regenerate` mode

- Regenerate the PR title and body **from scratch**.
- Treat the current committed net diff against the resolved base branch as the only source of truth.
- Do **not** reuse the previous PR title or body as factual input.

## Approval Workflow

Follow this loop exactly:

1. Refresh remote refs with `git fetch` for the resolved remote.
2. Verify the remote head branch with read-only checks such as:

- Program: `git`; args: `ls-remote`, `--heads`, `<remote>`, `<head-branch>`
- Program: `git`; args: `rev-parse`, `--abbrev-ref`, `HEAD`
- Program: `git`; args: `rev-parse`, `--verify`, `HEAD`
- Program: `git`; args: `rev-list`, `--left-right`, `--count`, `<remote>/<head-branch>...HEAD`

3. If the remote branch is missing or stale, stop and request approval for:

Program: `git`; args: `push`, `-u`, `<remote>`, `HEAD`

4. Only after the remote branch is current enough, generate a proposed PR title and PR body.
5. If repo policy requires missing issue metadata, label choice, or chain context, stop and ask for that minimum input before finalizing the proposal.
6. Return the title and body in copy-paste-ready form plus an explicit approval prompt for the content only.
7. If the user rejects the proposal, incorporate feedback without violating diff evidence constraints and repeat.
8. After the user approves the generated content, continue automatically to GitHub CLI actions and any explicit repo-policy label updates; do not ask for a second approval checkpoint.

## PR Creation via CLI (`create` mode)

After the user approves the generated content:

1. Write the body to `<body-output-path>` without requesting approval.
2. Run automatically:

Use safe argument execution with the resolved repo:

- Program: `gh`; args: `pr`, `create`, `--repo`, `<owner/repo>`, `--base`, `<base-branch>`, `--head`, `<head-branch>`, `--title`, `<title>`, `--body-file`, `<body-output-path>`

3. If repo policy requires explicit PR labels and they were resolved during policy detection, apply them automatically with `gh pr edit --repo <owner/repo>` using safe argument execution.

## PR Update via CLI (`regenerate` mode)

After the user approves the regenerated content:

1. Write the body to `<body-output-path>` without requesting approval.
2. Run automatically:

Use safe argument execution with the resolved repo:

- Program: `gh`; args: `pr`, `edit`, `<pr-number>`, `--repo`, `<owner/repo>`, `--title`, `<title>`, `--body-file`, `<body-output-path>`

3. If repo policy requires explicit PR labels and they were resolved during policy detection, apply them automatically with `gh pr edit --repo <owner/repo>` using safe argument execution.

## Output Requirements

When presenting content:

- show the mode clearly: `create` or `regenerate`
- show the resolved remote, head branch, and base branch
- show the convention source and any repo-policy requirements that were applied
- show the PR title clearly
- show the PR body clearly in copy-paste-ready Markdown
- show any `Chain Context` block that was preserved or added
- make approval boundaries explicit for push approval and content approval only

## Safety and Edge Cases

- If the repository has uncommitted changes, warn that PR content excludes them because committed diff is the source of truth.
- If the remote head branch does not exist according to `git ls-remote --heads`, stop and request push approval before creating or updating the PR.
- If repository policy requirements appear to be missing, warn about the gap when evidenced by templates/docs/checks, but do not fabricate compliance.
- If `<remote>/<base-branch>` does not exist, stop and ask the user to confirm the remote/base pair.
- If `regenerate` mode cannot resolve PR metadata and the user did not provide enough information, ask only for the missing minimum input.
- If the current local branch does not match the resolved PR head branch in `regenerate` mode, stop and explain the mismatch.
- If stacked or chained work is detected but the required base branch or chain context is ambiguous, stop and ask for explicit chain context.
- Never claim that a PR was created or updated unless the corresponding CLI command actually succeeded.

## Commands

```bash
/pr-create
/pr-regenerate
```

These wrappers may be exposed differently by each agent surface, but the workflow stays the same: explicit invocation, diff evidence, repo-policy awareness, and explicit approval before state-changing steps that are not already covered by the content approval rule.

## Resources

- Agent-level instruction surface when present (examples: `~/.config/opencode/AGENTS.md`, `~/.claude/CLAUDE.md`)
- Companion workflow skills when present, especially `chained-pr`
- Repo-local PR conventions when they exist (`.github/PULL_REQUEST_TEMPLATE.md`, `docs/CONVENTIONS.md`, `CONTRIBUTING.md`, `README.md`)
