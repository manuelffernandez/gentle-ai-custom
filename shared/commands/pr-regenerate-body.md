TASK:
Regenerate or update an existing pull request from the current committed diff, present its title/body for audit visibility, and sync it with GitHub CLI.

Rules for this command:
- The preferred owner skill is `branch-pr`; keep the command name stable but apply `branch-pr` behavior
- This command is **state-changing**: explicit invocation authorizes PR update after evidence gathering and required metadata resolution; present title/body for audit visibility, but do not ask for or wait for content approval
- Preserve the explicit push-approval gate when the remote head branch is missing or stale; do not add approval prompts for `git fetch`, temp artifact writes, or `gh pr edit`
- Resolve the target PR from explicit input first; otherwise detect it from the current branch when possible
- Use the current committed net diff against the resolved base branch as the only factual source of truth
- Detect and respect explicit repository PR policy first: template structure, issue linkage, labels, branch naming, chained PR context, and required checks only when the repository declares them
- Do **not** impose a universal approved-issue rule, universal `Closes #N`, or universal `type:*` label
- Do **not** reuse the previous PR body as factual input; regenerate from scratch
- Resolve the base branch in this order: explicit user input, resolved PR metadata (`baseRefName`), detected remote default branch, then ask only if still ambiguous
- Never assume branch names like `development`, `main`, or `master`
- If the current local branch does not match the resolved PR head branch, stop and explain the mismatch before proceeding
- Detect repository PR conventions first; if none exist, fall back to a generic English Markdown PR template
- If no open PR is found for the resolved head branch, stop and tell the user to use create mode instead
- Resolve `owner/repo` from the selected remote before GitHub CLI PR reads, and scope PR reads/updates with `--repo <owner/repo>`
- Do not build shell command strings by interpolating branch names, titles, paths, labels, or PR numbers; use safe argument execution
- If stacked or chained work is detected but base or chain context is ambiguous, stop and ask for explicit chain context before generating content
- If `gh` is unavailable or unauthenticated, stop after returning copy-paste-ready regenerated content
