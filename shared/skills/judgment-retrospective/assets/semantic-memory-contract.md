# Judgment Retrospective Semantic Memory Contract

## Scope

- Persist compact semantic records only.
- Keep v1 observations readable.
- Hydrate legacy records opportunistically instead of bulk-migrating everything.

## What to persist

- Compact run summaries for terminal Judgment Day loops
- Aggregated reusable patterns
- Intervention history and effectiveness
- Recurrence after mitigation

## What to avoid

- Raw judge output
- Full transcripts
- Line-by-line verdict dumps
- Unfiltered copies of the terminal report
- Prompt dumps

## Persistence rules

- Use `mem_save` / `mem_update` only.
- Set `capture_prompt: false` when the tool supports it.
- Preserve original semantic content when hydrating a legacy record, but do not carry forward raw dumps.
- Do not create a duplicate family when a relevant existing record can be updated.

## Topic keys

- `judgment-retrospective/run/{target-slug}`
- `judgment-retrospective/pattern/{pattern-slug}`
- `judgment-retrospective/intervention/{intervention-slug}`

## Stable slug example

- Target: `API timeout spikes`
- Normalized run key: `judgment-retrospective/run/api-timeout-spikes`

## v2 record contract

The common fields below apply to all three sub-schemas (run summary, pattern, intervention). Each sub-schema section lists its own type-specific fields. Fields inherited from common (`related_files`, `related_skills`, `evidence`, `recurrence`) are repeated in some sub-schemas for completeness — they carry the same semantics and are not redefined.

There is no common cross-link field. Each sub-schema defines its own type-specific link field:
- Run summaries use `updated_patterns` and `updated_interventions`
- Patterns use `linked_interventions`
- Interventions use `linked_patterns`

Do not add a generic `linked_records` field to v2 records.

Common fields:

- `schema_marker`: `judgment-retrospective.v2`
- `canonical_name`: short human-readable name for the record
- `aliases`: legacy titles, abbreviations, and normalized name variants
- `retrieval_terms`: search terms that should recall the record
- `recall_when`: the trigger that should bring this record back to mind
- `related_files`: exact paths that matter for the retrospective
- `related_skills`: exact skill names or paths involved in the behavior
- `evidence`: compact facts, counts, and representative observations
- `recurrence`: recurrence count or recurrence state after mitigation

### Run summary fields

Type-specific link fields:
- `updated_patterns`: topic keys or IDs of pattern records updated in this run, each optionally annotated with a short label (e.g. `judgment-retrospective/pattern/missing-retry-cap — incremented recurrence`)
- `updated_interventions`: topic keys or IDs of intervention records updated in this run, each optionally annotated with a short label

Other run-summary fields:
- `target`
- `terminal_state`
- `round_count`
- `verdict_counts`
- `fix_or_rejudge_result`
- `recurrence_signal`
- `related_files` *(inherited from common — repeated here for completeness)*
- `related_skills` *(inherited from common — repeated here for completeness)*

### Pattern record fields

Type-specific link field:
- `linked_interventions`: topic keys or IDs of related intervention records, each optionally annotated with a short label

Other pattern fields:
- `canonical_name`
- `aliases`
- `retrieval_terms`
- `recall_when`
- `pattern_summary`
- `evidence_count`: integer count of observed occurrences — a companion to the common `evidence` field (which holds the compact facts); not a replacement for it
- `first_seen`
- `last_seen`
- `recurrence_after_mitigation_count`
- `current_effectiveness`
- `related_files` *(inherited from common — repeated here for completeness)*
- `related_skills` *(inherited from common — repeated here for completeness)*

### Intervention record fields

Type-specific link field:
- `linked_patterns`: topic keys or IDs of related pattern records, each optionally annotated with a short label

Other intervention fields:
- `canonical_name`
- `aliases`
- `retrieval_terms`
- `recall_when`
- `intervention_summary`
- `intended_effect`
- `first_applied`
- `last_applied`
- `observed_effectiveness`
- `recurrence_count`
- `related_files` *(inherited from common — repeated here for completeness)*
- `related_skills` *(inherited from common — repeated here for completeness)*

## Compact summary definition

A compact summary is 1–3 sentences of semantic outcome only. It must not contain quoted judge text, verdict wording, prompt fragments, or line-by-line findings. State what was confirmed, what was fixed, and whether it worked — nothing more.

Example: "API timeout handler was missing a retry cap. The fix added a bounded retry with exponential backoff. No recurrence observed in the follow-up run."

## Legacy lookup and hydration

**Lookup order:**

1. Search current topic-key candidates (`judgment-retrospective/run/`, `judgment-retrospective/pattern/`, `judgment-retrospective/intervention/`).
2. If a v2 record is found, expand lookup using its `aliases` and `retrieval_terms` to surface any related legacy observations.
3. If no v2 record is found, derive lookup terms from the current compact terminal package: target name, normalized slug, and key domain terms (file paths, skill names, error categories).
4. Also search legacy title/text shapes: `JD pattern:`, `JD retrospective:`, `Pattern:`, `Retrospective:`, and `Issue:` plus domain terms from the current target.
5. If initial lookup is sparse, broaden to nearby title/text shapes before concluding no match exists.

**Hydration rules:**

- If a relevant legacy observation exists, update that same observation in place.
- Preserve the original semantic facts and append missing v2 fields.
- If the legacy record contains raw judge output, transcript text, prompt dumps, or unfiltered verdict dumps, replace that material with a compact summary (see definition above) before appending v2 fields.
- If multiple partial matches exist, compare them semantically and hydrate the best match only.
- Do not fork near-duplicate pattern or intervention families.
- Treat migration as opportunistic and lazy, not a required bulk job.

## Update rules

- Update the same topic key when the same pattern or intervention recurs.
- If the latest run shows recurrence after a prior mitigation, mention the previous mitigation explicitly.
- Prefer compact facts over detailed quotations.
- Keep the lookup footprint broad enough to find legacy records, but narrow enough to avoid family forks.
