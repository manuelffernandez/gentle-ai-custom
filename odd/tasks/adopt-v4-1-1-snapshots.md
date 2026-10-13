# Adopt v4.1.1 Repository Snapshots

## Objective

Sync approved upstream audit snapshots and the maintained release boundary to exact Gentle AI v4.1.1 without changing installed runtimes.

## Problem and scope

The pre-sync canonical audit verified installed binary v4.1.1 and a clean exact `v4.1.1` upstream checkout at `898d3976e8049a7238eaa7de60bff325c30d2b3d`. Six structural checks passed, while 22 managed assets had changed since v4.0.0. The user first authorized only the repository snapshot sync, then separately authorized the bounded owned-source, test-fixture, and maintainer-documentation updates needed to complete repository verification.

- Authorized writes: the canonical sync command's repository snapshot/state outputs; this progress document; the owned OpenCode `AGENTS.md` inherited Engram full block; the two Codex success-fixture versions; current-status maintainer docs; and one consolidated closed-event update-log entry after verification.
- Do not run `gentle-ai sync`, reinstall, or `apply-gentle-ai-custom.sh`; do not change installed runtime files, keep/prune policy, unrelated owned blocks, or historical v4.0.0 evidence.
- Branch: `chore/adopt-v4-1-1-snapshots`, based on `46194d4`.
- Forecast: 22 changed upstream snapshot sources copied by the sync (generated from upstream, excluded from authored-line delivery count), plus approximately 200 authored lines for inherited Engram parity, fixtures, current-status documentation, the log, and this task document. Delivery strategy: `ask-on-risk`.

## Tasks

- [x] **T1 — Adopt and verify repository snapshots.** Route: inline direct for the completed exact canonical sync command, then delegated direct for the approved multi-file parity and documentation update (two non-trivial surfaces: owned inherited instructions and fixture tests). `bash sync-gentle-ai-upstream-assets.sh` succeeded: 22 changed files, 33 unchanged, zero new/deleted, 19 directories, and a v4.1.1 state write. Baseline RED was observed for two Codex success fixtures using version 4.0.0 and the owned OpenCode full Engram parity test. The separately authorized fixes changed only those fixtures, the inherited full Engram block, current-status maintainer documentation, and one consolidated closed-event log entry. The three focused tests and `go test ./...` passed; `bash audit-gentle-ai-upstream.sh` passed with zero drift and all six structural flags `ok`; `git diff --check` passed. No runtime refresh, reinstall, or overlay apply occurred. Outcome and checks observed; work-unit commit: `c3cab57`; native committed-only assessment from `46194d4`: medium/under_budget (not due for review).

## Progress and next step

The separately approved repository sync advanced snapshots and the maintained boundary to exact v4.1.1. The later separately approved source/test/doc corrections brought repository verification to GREEN; the canonical audit found no drift or structural mismatch. A fresh-context read-only consistency pass found no unexpected path or source conflict; it did not rerun tests and issued no native review receipt. Work-unit `c3cab57` assessed medium/under_budget, so no native review started. Next: request separate authorization for any runtime refresh or overlay apply. Runtime sync, reinstall, and overlay apply remain unauthorized.
