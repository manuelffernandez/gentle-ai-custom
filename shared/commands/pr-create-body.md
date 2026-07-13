TASK:
Draft a pull request from committed changes and, if approved, create it with GitHub CLI.

Rules for this command:
- The preferred owner skill is `branch-pr`; keep the command name stable but apply `branch-pr` behavior
- This command is **state-changing**: the only explicit approval checkpoint is the generated PR content; once the user approves title/body, continue automatically unless a real blocker requires human input
- Do not add a second approval prompt for `git fetch`, temp artifact writes, or `gh pr create`
- Use the committed net diff as the only factual source of truth
- Detect and respect explicit repository PR policy first: template structure, issue linkage, labels, branch naming, chained PR context, and required checks only when the repository declares them
- Do **not** impose a universal approved-issue rule, universal `Closes #N`, or universal `type:*` label
- Detect repository PR conventions first; if none exist, fall back to a generic English Markdown PR template
- Resolve the base branch in this order: explicit user input, detected remote default branch, then ask only if still ambiguous
- Never assume branch names like `development`, `main`, or `master`
- If an open PR already exists for the current head branch, or for the same head/base pair when the base branch is already known, stop and tell the user to use regenerate/update mode instead of creating a duplicate
- Resolve `owner/repo` from the selected remote before GitHub CLI PR reads, and use `gh pr list --repo <owner/repo> --head <head-branch> --state open --json number,title,state,headRefName,baseRefName,url` to check for existing PRs
- Do not build shell command strings by interpolating branch names, titles, paths, labels, or PR numbers; use safe argument execution
- A PR that is `merged` or `closed` for that head branch does not block create mode
- If stacked or chained work is detected but base or chain context is ambiguous, stop and ask for explicit chain context before generating content
- If `gh` is unavailable or unauthenticated, stop after returning copy-paste-ready PR content
