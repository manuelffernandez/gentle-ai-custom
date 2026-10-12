# Adopt v4 ODD without SDD

## Objective

Adapt this repository's OpenCode and Codex overlay for Gentle AI v4.0.0: use upstream ODD task-close commits, stop creating SDD agents/profiles, and preserve unrelated runtime configuration.

## Problem and rationale

The clean v4.0.0 audit verified installed 4.0.0 and tag `ff77164d4f56f1665b22fb6fac51c2ccbb769400` against the maintained v3.7.0 boundary `6dee8f833aec9e46015759c5065a9035795d9af1`. Before this adaptation, it reported 49 managed changes, 11 unmapped new assets, and fixed-path SDD invariant failures. The overlay also cloned 11 SDD phase agents into configured profiles and patched mandatory ODD commits into opt-in commits. This repository-only adaptation addresses those local assumptions; snapshot adoption remains a separate decision.

## Authorized scope and constraints

- Repository-only adaptation of owned source, policy, exact local patches, tests, and affected documentation. Do not run repo snapshot sync, change `upstream-state.json`, mutate machine-local configuration, run `gentle-ai sync` or install/reinstall, or run overlay apply. Do not commit, push, or create a PR without a separate explicit request.
- Preserve user-owned legacy agent/profile keys instead of deleting them by prefix. Detect and report migration blockers before later apply; avoid silently mapping SDD model assignments onto ODD roles.
- Keep upstream assets audit-only and the existing remote-authorization and RDD ownership boundaries; do not invent a new runtime install source. Historical v3.7.0 snapshots remain until separately approved sync.
- ODD per-task commit requirement is intended future runtime behavior; current repo agent no-auto-commit overlay remains in force during this uncommitted implementation until changed and installed separately.
- ODD TDD mode has not been established by a current project/session setting. Historical OpenSpec-only TDD notes are not evidence for ODD. Use ordinary functional checks (`go test ./...`, `git diff --check`) unless a current ODD configuration proves otherwise.
- Forecast: likely above 400 authored changed lines in aggregate because profile code, patch tests, audit fixtures, and documentation are coupled. The 400-line per-task heuristic is advisory, not an acceptance limit. Delivery strategy `ask-on-risk`; delivery itself is not authorized.

## Tasks

- [x] **T1 — Remove SDD profile generation safely.** Delegated direct: multiple non-trivial Go/config/test files and preparatory reads. SDD phase/profile reconciliation was removed; legacy `default_profile`, `profiles`, SDD overrides and existing SDD keys now refuse apply before writes. Global JD/review assignments remain supported, and Codex's local instruction preflight uses the Engram marker instead of the removed SDD marker. Writer observed `go test ./internal/overlay`, `go test ./...`, and `git diff --check` pass; parent reran `go test ./internal/overlay` successfully and read back the new actor/preflight paths. No commit requested or made.
- [x] **T2 — Adopt upstream ODD commits.** Delegated direct: removed both work-unit opt-out patches and manifest declarations, the repo-owned OpenCode no-auto-commit section, and the Codex local instruction; preserved independent push/PR and safety boundaries. Active tests check both targets preserve the installed upstream work-unit skill. The first focused run failed on a stale patch expectation; after correcting the test, `go test ./internal/overlay`, `go test ./...`, and `git diff --check` passed. Independent review found two live-documentation mistakes and an inactive preservation test; corrected both, reran all checks successfully, and parent reran the focused Go test and diff check. No commit requested or made.
- [x] **T3 — Align v4 audit/manifest and maintenance docs without syncing snapshots.** Delegated direct: classified 11 new upstream-only assets, replaced obsolete SDD checks with v4 ODD binding, routing, remote authorization, and JD/review invariants, and updated docs. Added exact-path SDD snapshot retirement with safety checks for a future separately approved sync; deleted no snapshots while the maintained boundary is v3.7.0. Fresh-context review and a targeted recheck found structural false-pass cases; a first scoped correction addressed discarded projections and disconnected registration, and the separately user-approved second work unit rejected projection overwrite/conditional assignment and unreachable nested installer calls. The maintainer skill's stale profile guidance was corrected. Writer observed `go test ./internal/overlay`, `go test -count=1 ./...`, `git diff --check` pass; parent reran `go test -count=1 ./internal/overlay` and `git diff --check`, and a separate targeted read-only check passed for the corrected cases. Canonical read-only audit exited 1 solely for 41 unsynced managed snapshot changes, zero watched-unmapped changes and every structural flag `ok`. These are bounded source-shape checks, not installed-runtime proof. No commit requested or made.

## Acceptance criteria

- Local overlay code cannot recreate SDD agents in an eventual v4 apply, and never silently removes or migrates user-owned legacy config.
- The local commit-authorization override no longer contradicts upstream ODD task-close commits, for either target.
- Tests cover refusal/safe handling of legacy profiles and the v4 topology; changed docs agree with owned behavior. No snapshot boundary or installed runtime state is claimed updated.

## Progress and evidence

- Pre-write audit: v4.0.0 exact release parity passed; drift and obsolete SDD checks failed as expected. Repo worktree was clean on branch `upstream-first` before this document. Post-adaptation audit still awaits separately approved snapshot sync; it is not a runtime or installation verification.
- Route for T1/T2/T3: delegated direct; multiple non-trivial files and mapping/preparation triggers. No commits authorized, so work units will remain uncommitted; record checks rather than commit identities.
- Engram mirror: full-document copy saved under `odd/adopt-v4-odd-without-sdd/tasks`. Next: request separate approval for repository snapshot sync; only that step retires obsolete SDD snapshots and advances the v3.7.0 boundary. Full runtime reinstall and overlay apply require their own later approvals. Current machine-local SDD profile/config cleanup requires a separate user-owned decision; native RDD review awaits an eligible work-unit commit or PR slice, neither authorized here.

## Separately approved snapshot adoption — v4.0.0

- The user separately authorized `bash sync-gentle-ai-upstream-assets.sh`. It verified the installed binary and clean tag `v4.0.0` at `ff77164d4f56f1665b22fb6fac51c2ccbb769400`, then changed the maintained boundary from v3.7.0 to v4.0.0. Result: 11 new, 13 changed, 31 unchanged, and 35 deleted snapshot files across 19 directories; no runtime action or commit occurred.
- The first post-sync `bash audit-gentle-ai-upstream.sh` passed: no managed/unmapped drift, no base prompt drift, every structural flag `ok`. `git diff --check` passed. The first `go test -count=1 ./...` failed because `TestRunApplyCustomCodexInstallsSharedSkillsWithoutTouchingOtherTargets` and `TestRunApplyCustomAllAppliesBothTargetsWithoutCrossTargetCommands` still used a 3.7.0 simulated installed version; the advisory preflight correctly blocked them against the v4.0.0 maintained boundary. T4 corrected the fixtures and reran the tests successfully, as recorded below.
- This approval covers repository snapshot sync and its necessary tests/live documentation/log closure only. It does not authorize `gentle-ai sync`, reinstall, overlay apply, machine-local SDD cleanup, commit, push, or PR. The prior repository-only scope above remains historical evidence of the first three tasks.
- [x] **T4 — Close repository snapshot adoption.** Delegated direct: updated the stale simulated installed-version fixtures to 4.0.0 without weakening the advisory version guard, retired SDD keep entries in current OpenCode policy, aligned v4 live documentation, and wrote one consolidated adoption entry in `overlay/gentle-ai/logs/update-log.md`. Fresh-context review found two documentary inconsistencies; follow-up readback confirmed 18 manifest-listed exact retired files plus 17 stale files pruned within declared synced directories, and recast v2.5.0 runtime assertions as historical. Final parent verification: `go test -count=1 ./...` and `git diff --check` passed; `bash audit-gentle-ai-upstream.sh` passed with no managed or unmapped drift, no base prompt drift, and all structural flags `ok`. This verifies repository state, not installed OpenCode/Codex. No runtime refresh, reinstall, overlay apply, machine-local SDD cleanup, commit, push, or PR was authorized or run.

## Next step

- Recommend a full OpenCode runtime reinstall for the v4 topology; run it only on a separate explicit request. After that, separately authorize overlay apply. Inspect legacy machine-local SDD settings before either apply target; never delete or remap user-owned values automatically.
