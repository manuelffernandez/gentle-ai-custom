# Judgment Retrospective Semantic Memory Contract

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

## Topic keys

- `judgment-retrospective/run/{target-slug}`
- `judgment-retrospective/pattern/{pattern-slug}`
- `judgment-retrospective/intervention/{intervention-slug}`

## Stable slug example

- Target: `API timeout spikes`
- Normalized run key: `judgment-retrospective/run/api-timeout-spikes`

## Record shapes

### Run summary

- target
- terminal state
- round count
- verdict counts
- fix / re-judge result
- recurrence signal
- updated patterns / interventions

### Pattern record

- pattern summary
- evidence count
- first seen
- last seen
- recurrence after mitigation count
- current effectiveness
- related interventions

### Intervention record

- intervention summary
- intended effect
- first applied
- last applied
- observed effectiveness
- recurrence count
- notes on when it stopped working

## Update rules

- Use `mem_save` with `capture_prompt: false` when supported.
- Update the same topic key when the same pattern or intervention recurs.
- If the latest run shows recurrence after a prior mitigation, mention the previous mitigation explicitly.
- Prefer compact facts over detailed quotations.
