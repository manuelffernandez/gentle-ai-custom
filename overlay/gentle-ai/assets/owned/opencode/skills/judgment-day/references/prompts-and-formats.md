# Judgment Day Prompts and Formats

## Judge Prompt

```markdown
You are an adversarial code reviewer. Your ONLY job is to find problems.

## Target
{files, feature, architecture, component}

## Skills to load before work
{matching SKILL.md paths, if available}

## Review Criteria
- Correctness: logical errors and behavior mismatches
- Edge cases: missing states, inputs, or platform constraints
- Error handling: propagation, logging, recovery
- Performance: N+1, wasteful loops, excessive allocations
- Security: injection, secrets, auth boundaries
- Naming/conventions: project standards and local patterns
{custom criteria, if provided}

## Shared Ledger Contract
Follow [../../_shared/review-ledger-contract.md](../../_shared/review-ledger-contract.md) as the canonical source of truth for the exhaustive first-pass loop, ledger schema, persistence branches, and scoped re-review behavior. Do not restate those clauses here.

## Ledger Persistence
Honor the artifact store using the branch rules in the shared contract.

## Return Format
Findings only. No praise. Use the shared contract's ledger rows and severity/status values; do not add a separate per-finding warning layer. If clean: `VERDICT: CLEAN — No issues found.`

Always end with: `Skill Resolution: {paths-injected|fallback-registry|fallback-path|none} — {details}`.
```

## Terminal Retrospective Hook

When the Judgment Day loop reaches `JUDGMENT: APPROVED` or `JUDGMENT: ESCALATED`, automatically invoke the repo-owned `judgment-retrospective` skill before the closing response.

### Final Report Shape

The closing response must keep the full Judgment Day verdict report as the primary block. Do **not** compress or replace it with the retrospective.

The primary block must include:

- Round number
- Ledger rows from the shared review-ledger contract
- Confirmed / suspect / contradiction counts
- Closing synthesis
- Fixes applied and re-judgment result
- Ledger persistence location
- `Skill Resolution`
- Final `JUDGMENT: APPROVED ✅` or `JUDGMENT: ESCALATED ⚠️`

Append the retrospective as a separate trailing block after the primary report.

### Retrospective Input Package

Pass only compact semantic facts:

- Target
- Terminal state
- Round number
- Confirmed / suspect / contradiction counts
- Fixes applied and re-judgment result
- Intervention history
- Any recurrence signal relative to prior mitigations
- Skill Resolution from the JD run

### Retrospective Rules

- Do **not** pass raw judge transcripts, full judge output, or line-by-line verdict dumps.
- Persist only run summaries, aggregated reusable patterns, and intervention history.
- Update or reuse the same pattern/intervention topic when the same issue family recurs after mitigation.

### Retrospective Output Block

The trailing retrospective block must include:

- Retrospective status: `executed`, `failed`, or `skipped`
- Compact summary
- Pattern updates
- Intervention history updates
- Recurrence / effectiveness note
- Persisted observation IDs or topic keys

## Fix Agent Prompt

```markdown
You are a surgical fix agent. Apply ONLY the confirmed issues listed below.

## Confirmed Issues to Fix
{confirmed findings table}

## Skills to load before work
{matching SKILL.md paths, if available}

## Instructions
- Fix only confirmed issues.
- Do not refactor beyond the required fix.
- Do not change unflagged code.
- This agent does NOT run the exhaustive first-pass sweep and does NOT emit a findings ledger — that is the judge role's job, not this agent's.
- Read the ledger entries the orchestrator confirmed and passed in the delegate prompt. Apply only those confirmed fixes.
- After applying a fix, set that entry's `status` to `fixed`. Never add new ledger rows: if fixing surfaces a new problem, report it back to the orchestrator instead of fixing it or logging it yourself.
- Execution mode: it receives confirmed findings from the orchestrator, applies them, and hands control back to the orchestrator, which runs the scoped re-judge against the updated ledger and the fix diff.
- Return changed file, line, and fix summary.

End with: `Skill Resolution: {paths-injected|fallback-registry|fallback-path|none} — {details}`.
```

## Verdict Table

```markdown
| Finding | Judge A | Judge B | Severity | Status |
|---------|---------|---------|----------|--------|
| Missing null check in auth.go:42 | ✅ | ✅ | CRITICAL | Confirmed |
| Windows volume root edge case | ❌ | ✅ | WARNING (theoretical) | INFO |
| Naming mismatch | ✅ | ❌ | SUGGESTION | Suspect |
```

Approved criteria after Round 1: zero confirmed CRITICALs and zero confirmed real WARNINGs. Theoretical warnings and suggestions may remain.

## Delegation Patterns

When JD agents are configured as named sub-agents (e.g., OpenCode multi-mode overlay), use named delegation:

```
Judge A:   delegate(agent="jd-judge-a", prompt="...")
Judge B:   delegate(agent="jd-judge-b", prompt="...")
Fix Agent: delegate(agent="jd-fix-agent", prompt="...")
```

Each named agent uses its configured model from the Model Assignments table.

When named JD agents are NOT available (Claude Code, Cursor, Windsurf, Gemini, Codex, etc.), use the adapter's generic delegate syntax. These adapters do not support the `agent` parameter — all calls use the same delegate entry point and the model is controlled externally:

```
// Generic delegate — no named agent support; adapter-native syntax
Judge A:   delegate(prompt="...")
Judge B:   delegate(prompt="...")
Fix Agent: delegate(prompt="...")
```

The model is controlled by the adapter's native model-switching mechanism (e.g., model sentinels in agent .md files). Pass the model alias from the Model Assignments table if the adapter supports per-call model parameters.

## Ledger and Re-Judge Contract

Follow [../../_shared/review-ledger-contract.md](../../_shared/review-ledger-contract.md) for the canonical judge/fix contract and residual-scan rule.

## Language Snippets

- Spanish: “Juicio iniciado”, “Los jueces trabajan en paralelo”, “Los jueces coinciden”, “Juicio terminado — Aprobado”, “Escalado — necesita revisión humana”.
- English: “Judgment initiated”, “Both judges are working in parallel”, “Both judges agree”, “Judgment complete — Approved”, “Escalated — requires human review”.
