TASK:
Draft a pull request from committed changes and, if approved, create it with GitHub CLI.

Rules for this command:
- This command is **state-changing**: the only explicit approval checkpoint is the generated PR content; once the user approves title/body, continue automatically unless a real blocker requires human input
- Do not add a second approval prompt for `git fetch`, temp artifact writes, or `gh pr create`
- Use the committed net diff as the only factual source of truth
- Treat repo governance (issue linkage, labels, branch naming, merge policy) as external policy handled by the repo, CI, or companion skills such as `branch-pr`
- Detect repository PR conventions first; if none exist, fall back to a generic English Markdown PR template
- Resolve the base branch in this order: explicit user input, detected remote default branch, then ask only if still ambiguous
- Never assume branch names like `development`, `main`, or `master`
- If an open PR already exists for the current head branch, or for the same head/base pair when the base branch is already known, stop and tell the user to use regenerate/update mode instead of creating a duplicate
- Use `gh pr list --head "<head-branch>" --state open --json number,title,state,headRefName,baseRefName,url` to check for existing PRs
- A PR that is `merged` or `closed` for that head branch does not block create mode
- If `gh` is unavailable or unauthenticated, stop after returning copy-paste-ready PR content
