TASK:
Execute a post-SDD local commit flow for the current changes.

Rules for this command:
- This command is **state-changing**: never assume approval for `git add` or `git commit`
- If there is already an explicitly approved commit plan in this conversation, execute that exact plan
- If there is no approved plan yet, generate one first and **stop** for approval
- Use repository commit conventions first; if none are defined, fall back to Conventional Commits
- Never push or create/update PRs as part of this command
