---
name: judgment-retrospective
description: "Trigger: invoked automatically by the judgment-day terminal hook after APPROVED or ESCALATED. Persist compact JD run summaries, reusable patterns, and intervention history. Can also be loaded manually."
license: Apache-2.0
metadata:
  author: manuelfernandez
  version: "1.1"
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
- Never store raw judge output, full transcripts, or verdict dumps.
- Persist only selective semantic knowledge: compact run summary, aggregated patterns, and intervention history.
- Keep run summaries on the exact topic key `judgment-retrospective/run/{target-slug}`.
- Preserve the existing topic-key families unchanged:
  - `judgment-retrospective/run/{target-slug}`
  - `judgment-retrospective/pattern/{pattern-slug}`
  - `judgment-retrospective/intervention/{intervention-slug}`
- Use the v2 semantic-memory contract in [assets/semantic-memory-contract.md](assets/semantic-memory-contract.md) for new topic-key families and when hydrating legacy observations.
- Use `mem_save` / `mem_update` only. No ad-hoc notes.
- Set `capture_prompt: false` when the schema supports it.
- Track recurrence after mitigation and update the same family when the pattern or intervention reappears.

## Compatibility Rules

- Before creating a new pattern or intervention, search current topic-key candidates and legacy title/text shapes such as `JD pattern:`, `JD retrospective:`, `Pattern:`, `Retrospective:`, and `Issue:` plus domain terms from the current target.
- If a v2 record is found, expand lookup using its `aliases` and `retrieval_terms` to surface related legacy observations. If no v2 record is found, derive lookup terms from the current compact terminal package: target name, normalized slug, and key domain terms.
- If initial lookup is sparse, broaden to nearby title/text shapes before concluding no match exists.
- If a relevant legacy record exists, hydrate it in place by adding missing v2 fields. If the legacy content includes raw judge output, transcripts, prompt dumps, or unfiltered verdict dumps, replace that material with a compact summary before appending v2 fields. See the asset for the compact summary definition.
- If partial matches exist, compare them semantically and update the most relevant record.
- Do not fork near-duplicate pattern or intervention families.
- Migration is opportunistic and lazy, not a required bulk backfill.

## Decision Gates

| Condition | Action |
|---|---|
| Terminal state not reached | Skip the retrospective; report parent-facing status `skipped`. |
| Existing pattern/intervention topic found | Update that topic instead of creating a duplicate family. |
| Prior mitigation exists and recurrence detected | Mark the mitigation as only partially effective or ineffective; note the recurrence count. |
| Engram unavailable | Return parent-facing status `failed`; do not fabricate persistence. |

## Execution Steps

1. Normalize the terminal Judgment Day package into compact semantic facts.
2. Derive stable slugs for the target, pattern family, and intervention family.
3. Search Engram for existing run-summary, pattern, and intervention observations, including legacy title/text shapes.
4. Always upsert the compact run-summary observation to the exact topic key `judgment-retrospective/run/{target-slug}`.
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
