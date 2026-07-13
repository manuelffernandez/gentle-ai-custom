TASK:
Regenerate or update an existing pull request from the current committed diff and, if approved, sync it with GitHub CLI.

Rules for this command:
- The preferred owner skill is `branch-pr`; keep the command name stable but apply `branch-pr` behavior
- This command is **state-changing**: the only explicit approval checkpoint is the regenerated PR content; once the user approves title/body, continue automatically unless a real blocker requires human input
- Do not add a second approval prompt for `git fetch`, temp artifact writes, or `gh pr edit`
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
