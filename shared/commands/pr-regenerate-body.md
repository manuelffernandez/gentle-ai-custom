TASK:
Regenerate or update an existing pull request from the current committed diff and, if approved, sync it with GitHub CLI.

Rules for this command:
- This command is **state-changing**: never assume approval for `git fetch`, `git push`, temp artifact writes, `gh pr edit`, or `gh pr create`
- Resolve the target PR from explicit input first; otherwise detect it from the current branch when possible
- Use the current committed net diff against the resolved base branch as the only factual source of truth
- Do **not** reuse the previous PR body as factual input; regenerate from scratch
- Resolve the base branch in this order: explicit user input, resolved PR metadata (`baseRefName`), detected remote default branch, then ask only if still ambiguous
- Never assume branch names like `development`, `main`, or `master`
- If the current local branch does not match the resolved PR head branch, stop and explain the mismatch before proceeding
- Detect repository PR conventions first; if none exist, fall back to a generic English Markdown PR template
- If no open PR is found for the resolved head branch, stop and tell the user to use create mode instead
- If `gh` is unavailable or unauthenticated, stop after returning copy-paste-ready regenerated content
