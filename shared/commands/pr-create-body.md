TASK:
Draft a pull request from committed changes and, if approved, create it with GitHub CLI.

Rules for this command:
- This command is **state-changing**: never assume approval for `git fetch`, `git push`, temp artifact writes, or `gh pr create`
- Use the committed net diff as the only factual source of truth
- Detect repository PR conventions first; if none exist, fall back to a generic English Markdown PR template
- Resolve the base branch in this order: explicit user input, detected remote default branch, then ask only if still ambiguous
- Never assume branch names like `development`, `main`, or `master`
- If an open PR already exists for the current head branch, stop and tell the user to use regenerate/update mode instead of creating a duplicate
- If `gh` is unavailable or unauthenticated, stop after returning copy-paste-ready PR content
