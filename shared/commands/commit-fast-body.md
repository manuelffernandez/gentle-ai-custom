TASK:
Generate a local commit plan for the current changes and execute it immediately without waiting for approval.

Rules for this command:
- **Single-invocation scope — HARD RULE**: this command applies ONLY to the changes present at the moment it is called. Once executed, this command is DONE. Do NOT commit again after any subsequent implementation work in the same session unless the user explicitly invokes this command again. Finishing an implementation task is never an implicit trigger for another commit.
- This command is **state-changing**: it will run `git add` and `git commit` without pausing for plan approval
- Generate the plan using the same rules as `/commit-plan` (convention detection, coherent grouping)
- Display the full plan **before** executing — for audit visibility, not for approval
- Execute all commits sequentially after displaying the plan
- **Still stop** if any blocker is found:
  - same file would be split across multiple commits
  - suspected secret or credential file detected
  - unrelated changes the plan cannot cleanly separate
  - any `git commit` fails mid-execution
- Use repository commit conventions first; if none are defined, fall back to Conventional Commits
- Never push or create/update PRs as part of this command
