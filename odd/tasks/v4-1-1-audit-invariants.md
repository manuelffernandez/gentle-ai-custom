# v4.1.1 Audit Invariants

## Objective

Make the structural upstream audit recognize the verified v4.1.1 OpenCode rendering and settings-registration paths without weakening fail-closed checks.

## Problem and rationale

The exact clean v4.1.1 release passes parity, but the audit reports four structural mismatches. Source inspection shows that the orchestrator asset read moved into a renderer helper and ODD settings registration now uses a path-aware JSON merge. Routing and remote-authorization checks depend on the orchestrator binding result. The 22 changed managed assets are real drift and must remain separately reported.

## Authorized scope and constraints

- Repair structural audit checks and their deterministic regression tests only; update affected operating documentation if the audit contract changes.
- Do not sync upstream snapshots, advance the maintained boundary, apply the overlay, refresh/reinstall any runtime, or change keep/prune or local ownership policy.
- Preserve fail-closed checks for wrong assets, disconnected injection, lost remote projection, and unreachable ODD installation.
- Current branch: `fix/v4-1-1-audit-invariants`; starting boundary: `99c1cf9` on `upstream-first`.
- Forecast: approximately 250 authored changed lines (additions plus deletions, excluding generated files); delivery strategy: `ask-on-risk`. The completed work is approximately 400 authored lines including tests and documentation; it remains one coherent audit-check unit.

## Tasks

- [x] **T1 — Update v4.1.1 structural recognition.** Route: delegated direct; trigger: two non-trivial Go files (auditor and regression tests), with source reading kept in the writer. Deterministic fixtures observed RED for new renderer and path-aware settings-write shapes, then GREEN after narrow AST changes; negative cases reject discarded asset text, overwritten settings, unreachable calls, and switch-case truncation. Acceptance observed: all six structural flags are `ok` on exact v4.1.1; deliberate breakage fails closed; 22 changed managed snapshots remain reported. Verification: `go test ./internal/overlay -run 'TestV4' -count=1` passed; `go test ./internal/overlay -count=1` passed; `go test ./...` passed after formatting; `git diff --check` and `gofmt -l` passed. `bash audit-gentle-ai-upstream.sh` exits 1 solely for the known pending snapshot drift, not structural failure. Status: complete; assessed tier and review outcome: pending post-commit assessment; commit: pending.

## Progress and next step

The canonical pre-change audit verified exact release v4.1.1 and reported four structural mismatches plus 22 changed managed files. The repaired audit reports all six structural checks `ok` while preserving the 22-file adoption gate. A fresh-context consistency check found three reachable mutation gaps; targeted RED/GREEN regressions closed them before final verification. Next: record work-unit commit and its native review assessment. Repository snapshot sync, runtime refresh, and overlay apply remain separately unauthorized.
