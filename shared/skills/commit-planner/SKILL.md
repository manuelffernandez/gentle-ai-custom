---
name: commit-planner
description: >
  Propose and optionally execute a post-SDD commit plan by grouping changed files into coherent local commits,
  detecting repository commit conventions first and falling back to Conventional Commits.
  Trigger: When the user asks for a commit plan, commit grouping, post-SDD finalization, or to commit current changes.
license: Apache-2.0
metadata:
  author: manuelfernandez
  version: "1.0"
---

## When to Use

Use this skill when:

- SDD or normal implementation work is already finished and the user wants to decide the next finalization step
- The user says things like `proponeme un plan de commits`, `generame un plan de commits`, `commiteá estos cambios`, or `cómo agruparías estos cambios`
- You need to decide whether the current working tree should become one commit or several meaningful local commits

Do **not** use this skill as part of the default SDD flow. This is a **post-SDD**, user-invoked finalization step.

## Critical Patterns

1. **Never auto-run after SDD**. Only activate this flow when the user explicitly asks.
2. **Three modes**:
   - `plan` → inspect and propose commit grouping, no state changes
   - `apply` → execute an already approved plan, or generate a plan first and stop for approval
   - `auto` → generate and execute in one shot without pausing for plan approval; still stops on blockers
3. **MANDATORY Convention Detection**: You MUST explicitly search for and read repository conventions BEFORE generating any commit message. Do NOT guess or skip this step.
   - 1st: Check for explicit repository guidance (`docs/CONVENTIONS.md`, `CONTRIBUTING.md`, `README.md`, and relevant repo-local workflow docs).
   - 2nd: Inspect recent `git log` subjects to understand the actual style used in the repository.
   - 3rd: Only if the above yield no clear pattern, fall back to **Conventional Commits**.
4. **Prefer file-level grouping**. Never assume hunk splitting is safe.
5. If a clean plan would require splitting the **same file** across different commits, **stop and report the blocker**.
6. Never stage or commit likely secrets (`.env`, credential files, tokens, generated secret dumps).
7. Never push, open PRs, or edit PRs as part of this skill. That belongs to a separate release/PR flow.
8. State-changing steps (`git add`, `git commit`) require explicit user approval.
9. **Language constraint**: Unless the user explicitly requests otherwise, all generated commit messages, plan bodies, and comments must be in **ENGLISH**.

## Read-Only Evidence Gathering

These checks are allowed before proposing a plan:

```bash
git status -sb
git diff --name-status
git diff --staged --name-status
git diff --stat
git diff --staged --stat
git log --format=%s -n 15
```

If a convention file exists, read it before proposing commit messages.

## Divergence Detection (passive)

After running `git status -sb`, inspect the tracking branch status line:

- If the output shows `behind N` (e.g. `## main...origin/main [behind 3]`): **stop and warn the user** before proposing or executing any commits. The remote has commits that are not yet in the local branch. Continuing may create unnecessary merge commits or conflicts on push. Let the user decide whether to pull first.
- If the output shows `ahead N`: safe to proceed — local has unpushed commits, which is the expected state.
- If the output shows `ahead N, behind M` (diverged): **stop and warn**. This is a diverged state; committing is allowed but pushing will require a merge or rebase. Surface the information clearly.
- If there is no tracking branch configured: proceed normally without warning.

Do NOT run `git fetch` at any point. Detection is passive — based solely on the cached remote-tracking ref already present locally.

## Planning Rules

When building the plan:

- Prefer **one commit per coherent intention**, not per folder by default
- Keep infrastructure, docs, tests, and implementation changes separate only when that improves history clarity
- If all changed files belong to one single behavioral change, prefer a single commit
- If there are unrelated changes in the working tree, call them out explicitly and either:
  - exclude them from the plan, or
  - stop and ask the user to confirm inclusion
- If the working tree is clean, say so and stop

### Message generation

- Use the repository convention when explicitly documented
- If repo docs and recent history conflict, prefer docs and mention the mismatch
- If no repo convention exists, use Conventional Commits
- Messages should explain the **why** of the change, not just restate filenames

## Mode: `plan`

Return a proposal with this structure:

- `Working Tree Summary`
- `Convention Source`
- `Proposed Commits`
- `Excluded or Blocked Files`
- `Approval Prompt`

Inside `Proposed Commits`, include for each commit:

- commit number
- intent / rationale
- files to stage
- suggested commit message

`plan` mode is strictly read-only.

## Mode: `apply`

Execution rules:

1. If there is **no explicitly approved plan** in the conversation, generate the plan first and stop.
2. If the plan is approved, execute commits **sequentially**.
3. For each commit group:
   - stage files by path
   - create the commit with the approved message
   - verify status before continuing to the next group
4. If any commit fails, stop immediately and report the exact failure.
5. Never invent or silently modify an approved plan. If execution reality differs from the plan, stop and explain why.

Return this structure after execution:

- `Approved Plan`
- `Actions Taken`
- `Remaining Changes`
- `Blockers`

## Mode: `auto`

Triggered by `/commit-fast` or natural-language cues like `commiteá directamente`, `aplicá sin preguntar`, `commit rápido`.

**Single-invocation rule**: `auto` mode applies to the changes present at the moment it is invoked. It does NOT carry over to subsequent changes in the same session. Each new set of changes requires an explicit new invocation — never assume `auto` mode is the default going forward.

1. Generate the plan using the same rules as `plan` mode.
2. Display the full plan in the output **before** executing — for audit visibility, not for approval.
3. Execute all commits immediately without waiting for user confirmation.
4. **Still stop** if any of these conditions apply:
   - a blocker exists (same file across multiple commits, suspected secret, ambiguous file)
   - the working tree contains unrelated changes the plan cannot cleanly separate
   - any `git commit` fails mid-execution

Return the same structure as `apply` mode after execution.

## Commands

```bash
/commit-plan
/commit-apply
/commit-fast
```

These wrappers MAY be exposed by different agent surfaces, but the workflow stays the same:

- `plan` is read-only
- `apply` is state-changing and requires explicit user approval before staging or committing
- `fast` is state-changing and executes without approval pause; blockers still require human decision

Natural-language triggers:

- `proponeme un plan de commits`
- `generame un plan de commits`
- `commiteá estos cambios`
- `commiteá directamente`
- `aplicá sin preguntar`
- `commit rápido`

## Resources

- **Agent-level instruction surface** when present (examples: `~/.config/opencode/AGENTS.md`, `~/.claude/CLAUDE.md`)
- **Examples of release / PR conventions** in repo-local docs or companion skills such as `branch-pr` when available
