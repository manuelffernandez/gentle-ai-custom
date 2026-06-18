---
name: judgment-retrospective
description: "Trigger: invoked automatically by the judgment-day terminal hook after APPROVED or ESCALATED. Persist compact JD run summaries, reusable patterns, and intervention history. Can also be loaded manually."
license: Apache-2.0
metadata:
  author: manuelfernandez
  version: "1.0"
---

## Activation Contract

Use this skill when:

- The terminal Judgment Day hook invokes it after `JUDGMENT: APPROVED` or `JUDGMENT: ESCALATED`
- The parent flow needs its automatic retrospective step
- The user explicitly loads it to run a standalone retrospective outside the automated hook flow

Installing this skill alone does not wire the automatic behavior; the runtime `judgment-day` hook is what activates it automatically. Manual loading is supported but the caller is responsible for providing the terminal package.

Do **not** use it for live judging, and do **not** use it to replace the judge/fix loop.

## Hard Rules

- Consume only the compact terminal package from Judgment Day.
- Never store raw full judge output, judge transcripts, or unfiltered verdict dumps.
- Persist selective semantic knowledge only: compact run summary, aggregated reusable patterns, and intervention history.
- Track recurrence after mitigation. If a pattern reappears after a prior fix, update the existing pattern/intervention records with the recurrence signal.
- Keep run summaries compact and overwrite them with the exact topic key `judgment-retrospective/run/{target-slug}`.
- Use Engram upserts, not ad-hoc notes.
- Set `capture_prompt: false` when the schema supports it; this is automated retrospective work.

## Decision Gates

| Condition | Action |
|---|---|
| Terminal state not reached | Skip the retrospective and report parent-facing status `skipped`. |
| Existing pattern/intervention topic found | Update that topic instead of creating a duplicate family. |
| Prior mitigation exists and recurrence is detected | Mark the mitigation as only partially effective or ineffective, and note the recurrence count. |
| Engram unavailable | Return parent-facing status `failed` and do not fabricate persistence. |

## Execution Steps

1. Normalize the terminal Judgment Day package into compact semantic facts.
2. Derive stable slugs for the target, pattern family, and intervention family.
3. Search Engram for existing run-summary, pattern, and intervention observations.
4. Save or update the compact run-summary observation at `judgment-retrospective/run/{target-slug}`.
5. Save or update one or more pattern observations with aggregate evidence and recurrence data.
6. Save or update intervention history with effectiveness, recurrence after mitigation, and the last observed terminal state.
7. Return a concise retrospective report with observation IDs/keys and effectiveness notes.

## Output Contract

Return `## Judgment Retrospective — {target}` with:

- retrospective status: `executed`, `failed`, or `skipped`
- persistence detail: `saved`, `updated`, or `unavailable`
- terminal state
- compact run summary
- pattern updates
- intervention history updates
- recurrence/effectiveness note
- persisted observation IDs or topic keys

Status mapping:

- `executed` when the retrospective completed and Engram persistence was saved or updated
- `failed` when the retrospective could not complete persistence
- `skipped` when the terminal state was not reached

## References

- [assets/semantic-memory-contract.md](assets/semantic-memory-contract.md) — topic key scheme, record shapes, and minimal persistence examples.
