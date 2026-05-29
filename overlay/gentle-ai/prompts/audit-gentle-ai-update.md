# Audit Prompt — Gentle AI Upstream Update

You are auditing a new upstream Gentle AI release for local overlay compatibility.

## Inputs

1. Upstream source repo: `/home/manuel/Documentos/gentle-ai`
2. Overlay repo: `/home/manuel/Documentos/gentle-ai-custom`
3. Local policy file: `overlay/gentle-ai/policy/gentle-ai-policy.json`
4. Orchestrator policy: `overlay/gentle-ai/policy/orchestrator-policy.md`
5. Current derived prompt: `overlay/gentle-ai/derived/opencode/gentle-orchestrator.md`
6. Last upstream snapshot: `overlay/gentle-ai/snapshots/upstream/opencode/gentle-orchestrator.last.md`

## Tasks

1. Compare upstream skill inventory with `skills.keep` and `skills.prune`.
2. Detect renamed, removed, or newly introduced skills that affect keep/prune behavior.
3. Compare upstream OpenCode orchestrator prompt with current derived prompt.
4. Re-derive the prompt under this strict rule: remove all PR/budget/chained-PR/review-workload workflow content, keep the rest of useful SDD orchestration behavior.
5. Identify whether apply scripts (`apply-gentle-ai-policy.sh` and `.ps1`) require updates for parity.

## Output format

Produce:

### 1) Compatibility verdict
- `safe-no-change` | `update-required`

### 2) Findings
- Skill inventory drift
- Prompt drift
- Script drift/parity risk

### 3) Proposed patch plan
- Files to edit
- Why each edit is needed
- Any backward-compatibility caveat

### 4) Human review checklist
- [ ] Keep/prune lists still reflect intent
- [ ] Derived orchestrator still excludes PR/budget workflow content
- [ ] Bash/PowerShell scripts remain behaviorally equivalent
- [ ] Snapshot and update log were refreshed

Keep the report concise and actionable.
