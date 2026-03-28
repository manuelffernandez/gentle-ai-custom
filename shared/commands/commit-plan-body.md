TASK:
Inspect the current git working tree and propose a local commit plan for the current changes.

Rules for this command:
- This command is **read-only**: do not stage, commit, push, or create/update PRs
- Detect explicit repository commit conventions first; if none exist, fall back to Conventional Commits
- If a clean plan would require splitting hunks from the same file across different commits, report that blocker instead of guessing
- End by asking whether the proposed plan should be approved for execution
