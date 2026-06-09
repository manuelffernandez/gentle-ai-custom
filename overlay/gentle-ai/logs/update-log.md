# Gentle AI Overlay Update Log

This file records only closed upstream-maintenance / overlay-maintenance events and maintenance-contract changes that matter for alignment with Gentle AI upstream.

It is not a mirror of every repo change. Git history carries implementation-level edits, intermediate iterations, and doc wording churn. `overlay/gentle-ai/state/upstream-state.json` remains the source of truth for the last maintained upstream boundary.

## 2026-06-08 | Made maintainer workflow decision-oriented and approval-gated

- **Type**: `policy-change`
- **Upstream scope/range**: maintenance workflow contract, not a new upstream boundary
- **Decision**: changed the maintainer contract so every upstream audit must produce a pre-mutation decision summary (`what is new upstream`, `recommend adopt`, `recommend discard`, `why`, and refresh recommendation), STOP for explicit user approval before any repo/runtime mutation, then finish with a closing adopted-vs-discarded summary plus a fresh-context consistency review.
- **Why it mattered**: the previous workflow required audit-first behavior but did not force the maintainer to separate recommendation from execution. The new contract makes upstream adoption decisions explicit before mutation and adds a second review pass after the approved maintenance work lands.
- **Affected artifacts**: `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`, `AGENTS.md`, `README.md`, `overlay/gentle-ai/maintenance.md`, `overlay/gentle-ai/logs/update-log.md`
- **Verification**: targeted doc/skill consistency review plus `git diff --check`

## 2026-06-08 | Adopted upstream v1.36.8 native SDD status baseline

- **Type**: `adoption`
- **Upstream scope/range**: `3883470b175dc6b95904594135c34cc5f6ad2413` (`v1.34.1`) -> `122b35816d3fbc1627359fe0613c6541604980bc` (`v1.36.8`)
- **Decision**: adopted the upstream `v1.36.8` SDD dispatcher/status baseline, added the approved upstream `sdd-status` command and shared status contract assets, preserved the local overlay depuration that removes PR/review-budget governance from repo-owned runtime prompts, and kept `gentle-ai sync` as the recommended upstream refresh path because topology and profile invariants stayed stable.
- **Why it mattered**: upstream introduced native structured SDD status and dispatcher routing that affect orchestrator and SDD phase behavior. The overlay needed to accept those capabilities without reintroducing the repository-governance policy that this repo intentionally prunes.
- **Affected artifacts**: `overlay/gentle-ai/assets/upstream/opencode/commands/sdd-apply.md`, `overlay/gentle-ai/assets/upstream/opencode/commands/sdd-archive.md`, `overlay/gentle-ai/assets/upstream/opencode/commands/sdd-continue.md`, `overlay/gentle-ai/assets/upstream/opencode/commands/sdd-status.md`, `overlay/gentle-ai/assets/upstream/opencode/commands/sdd-verify.md`, `overlay/gentle-ai/assets/upstream/opencode/prompts/orchestrators/gentle-orchestrator.md`, `overlay/gentle-ai/assets/upstream/opencode/skills/_shared/openspec-convention.md`, `overlay/gentle-ai/assets/upstream/opencode/skills/_shared/sdd-status-contract.md`, `overlay/gentle-ai/assets/upstream/opencode/skills/sdd-apply/SKILL.md`, `overlay/gentle-ai/assets/upstream/opencode/skills/sdd-archive/SKILL.md`, `overlay/gentle-ai/assets/upstream/opencode/skills/sdd-spec/SKILL.md`, `overlay/gentle-ai/assets/upstream/opencode/skills/sdd-verify/SKILL.md`, `overlay/gentle-ai/assets/owned/opencode/commands/sdd-apply.md`, `overlay/gentle-ai/assets/owned/opencode/commands/sdd-archive.md`, `overlay/gentle-ai/assets/owned/opencode/commands/sdd-continue.md`, `overlay/gentle-ai/assets/owned/opencode/commands/sdd-explore.md`, `overlay/gentle-ai/assets/owned/opencode/commands/sdd-ff.md`, `overlay/gentle-ai/assets/owned/opencode/commands/sdd-init.md`, `overlay/gentle-ai/assets/owned/opencode/commands/sdd-new.md`, `overlay/gentle-ai/assets/owned/opencode/commands/sdd-onboard.md`, `overlay/gentle-ai/assets/owned/opencode/commands/sdd-status.md`, `overlay/gentle-ai/assets/owned/opencode/commands/sdd-verify.md`, `overlay/gentle-ai/assets/owned/opencode/prompts/orchestrators/gentle-orchestrator.md`, `overlay/gentle-ai/assets/owned/opencode/skills/_shared/openspec-convention.md`, `overlay/gentle-ai/assets/owned/opencode/skills/_shared/sdd-phase-common.md`, `overlay/gentle-ai/assets/owned/opencode/skills/_shared/sdd-status-contract.md`, `overlay/gentle-ai/assets/owned/opencode/skills/sdd-apply/SKILL.md`, `overlay/gentle-ai/assets/owned/opencode/skills/sdd-archive/SKILL.md`, `overlay/gentle-ai/assets/owned/opencode/skills/sdd-spec/SKILL.md`, `overlay/gentle-ai/assets/owned/opencode/skills/sdd-tasks/SKILL.md`, `overlay/gentle-ai/assets/owned/opencode/skills/sdd-verify/SKILL.md`, `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.md`, `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.meta.yaml`, `overlay/gentle-ai/state/upstream-state.json`, `overlay/gentle-ai/logs/update-log.md`
- **Verification**: `bash audit-gentle-ai-upstream.sh` with `managed assets drift: ok`, `base prompt drift: no`, and all profile/base-asset invariants `ok`
- **Follow-up**: user-side runtime adoption still requires `gentle-ai sync` and then `apply-gentle-ai-custom.sh opencode`, but those steps were intentionally not run in this session.

## 2026-06-05 | Completed the owned-assets control-plane cutover

- **Type**: `tooling-change`
- **Upstream scope/range**: maintenance tooling/runtime contract, not a new upstream boundary
- **Decision**: completed the repo-owned managed-assets control plane by making `overlay/gentle-ai/policy/managed-assets.json` the canonical map for audit/sync/apply, switching `apply-gentle-ai-custom` to install runtime SDD/orchestrator assets from `overlay/gentle-ai/assets/owned/...`, preserving repo-owned portable skills from `shared/skills/` and custom wrappers from `shared/commands/`, rewriting prompt refs directly to owned runtime files, and keeping `sync-gentle-ai-upstream-assets` as the mechanical refresh step for approved upstream snapshots and the audited `gentle-orchestrator` baseline.
- **Why it mattered**: the old runtime relied on sanitizing and capturing upstream inline prompts, plus local operational snapshot behavior that no longer matched the desired ownership model. The repo now has one explicit control plane: owned runtime assets for apply, approved upstream snapshots for review, and git+manifest discovery for maintainer audit.
- **Affected artifacts**: `internal/overlay/audit_upstream.go`, `internal/overlay/git_diff.go`, `internal/overlay/managed_assets.go`, `internal/overlay/sync_upstream_assets.go`, `cmd/gentle-ai-overlay/main.go`, `audit-gentle-ai-upstream.sh`, `audit-gentle-ai-upstream.ps1`, `sync-gentle-ai-upstream-assets.sh`, `sync-gentle-ai-upstream-assets.ps1`, `overlay/gentle-ai/policy/managed-assets.json`, `overlay/gentle-ai/owned-assets-refactor.md`, `overlay/gentle-ai/maintenance.md`, `overlay/gentle-ai/assets/upstream/opencode/README.md`, `README.md`, `AGENTS.md`, `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`
- **Verification**: `go test ./...`; `go run ./cmd/gentle-ai-overlay audit-upstream`; `go run ./cmd/gentle-ai-overlay sync-upstream-assets --help`; `bash audit-gentle-ai-upstream.sh --help`; `bash sync-gentle-ai-upstream-assets.sh --help`

## 2026-06-04 | Made upstream/runtime config resolution portable

- **Type**: `tooling-change`
- **Upstream scope/range**: maintenance runtime portability, not a new upstream boundary
- **Decision**: removed the versioned absolute upstream repo path from shared policy behavior, introduced the canonical local config `~/.config/gentle-ai-custom/opencode-local-config.json`, separated local `agent_overrides` from `profiles`, added optional `opencode_config_path`, implemented upstream resolution precedence (`local config -> $GENTLE_AI_CUSTOM_UPSTREAM_REPO -> ../gentle-ai`), and kept the legacy `opencode-sdd-profiles.json` fallback when the new config omits `profiles`.
- **Why it mattered**: the overlay was still anchored to one machine's upstream checkout path and split local runtime choices across multiple files; portability required one canonical local config plus deterministic fallback rules that keep existing profile setups working during migration.
- **Affected artifacts**: `overlay/gentle-ai/policy/gentle-ai-policy.json`, `internal/overlay/policy.go`, `internal/overlay/local_config.go`, `internal/overlay/local_config_test.go`, `internal/overlay/apply_policy.go`, `internal/overlay/profiles.go`, `internal/overlay/audit_upstream.go`, `internal/overlay/summary.go`, `README.md`, `AGENTS.md`, `overlay/gentle-ai/README.md`, `overlay/gentle-ai/maintenance.md`, `overlay/gentle-ai/policy/maintenance-intent.md`, `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`
- **Verification**: `gofmt -w internal/overlay/*.go`; `go test ./...`

## 2026-06-04 | Adopted upstream v1.34.1 interactive SDD baseline

- **Type**: `audit`
- **Upstream scope/range**: `55a5bfe43594d6409307c4bcdf3a1d22a8c42560` (`v1.34.0`) -> `3883470b175dc6b95904594135c34cc5f6ad2413` (`v1.34.1`)
- **Decision**: adopted the new `gentle-orchestrator` baseline, kept both upstream interactive SDD additions (phase-scoped approval and the proposal question round before `sdd-propose`), and advanced the maintained boundary to `v1.34.1` without changing policy or sanitizer behavior.
- **Why it mattered**: upstream changed interactive orchestration behavior in ways that affect the coordinator UX, but the overlay’s topology and profile invariants stayed stable, so the right response was to accept the new baseline and preserve the existing pruning/sanitization contract.
- **Affected artifacts**: `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.md`, `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.meta.yaml`, `overlay/gentle-ai/state/upstream-state.json`, `overlay/gentle-ai/logs/update-log.md`
- **Verification**: audit/update pass against upstream `v1.34.1`; post-edit `bash audit-gentle-ai-upstream.sh`

## 2026-06-03 | Closed v1.34.0 prompt-language maintenance

- **Type**: `audit`
- **Upstream scope/range**: `0fa9f2d1d2d3a8ebd822cdd5c82fcb4bff60f0fc` (`v1.33.2`) -> `55a5bfe43594d6409307c4bcdf3a1d22a8c42560` (`v1.34.0`)
- **Decision**: adopted the new `gentle-orchestrator` baseline, kept `gentle-ai sync` as the upstream refresh path, patched the sanitizer for the neutral-Spanish preflight wording shift, and added a human-readable `Drift summary:` to the maintainer audit.
- **Why it mattered**: upstream introduced a real prompt-language contract change (`Language Domain Contract`) without topology or profile-invariant drift; the overlay needed both a compatible sanitizer and a clearer audit handoff before the next `sync` + `apply` cycle.
- **Affected artifacts**: `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.md`, `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.meta.yaml`, `overlay/gentle-ai/state/upstream-state.json`, `internal/overlay/sanitize.go`, `internal/overlay/audit_upstream.go`, `AGENTS.md`, `README.md`, `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`, `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`, `overlay/gentle-ai/README.md`
- **Verification**: `gofmt -w internal/overlay/audit_upstream.go`; `go test ./...`; `bash audit-gentle-ai-upstream.sh`
- **Follow-up**: user-side runtime materialization still requires `gentle-ai sync` and then `apply-gentle-ai-custom`.

## 2026-06-03 | Moved orchestrator sanitization intent into `maintenance-intent.md`

- **Type**: `policy-change`
- **Upstream scope/range**: `n/a`
- **Decision**: treated `overlay/gentle-ai/policy/maintenance-intent.md` as the semantic source of truth for both keep/prune intent and orchestrator sanitization goals.
- **Why it mattered**: maintainer decisions about upstream drift must read one semantic contract, not reconstruct intent across multiple partially overlapping policy files.
- **Affected artifacts**: `overlay/gentle-ai/policy/maintenance-intent.md`, `AGENTS.md`, `README.md`, `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`
- **Verification**: manual cross-check of file roles and maintainer references after the consolidation

## 2026-06-01 | Closed the post-tag v1.33.2 docs-only upstream audit

- **Type**: `audit`
- **Upstream scope/range**: `0fa9f2d1d2d3a8ebd822cdd5c82fcb4bff60f0fc` (`v1.33.2`) -> `21634526`
- **Decision**: closed the audit with `no overlay changes required`; only advanced `upstream-state.json` to the new reviewed head.
- **Why it mattered**: the upstream delta was documentation/comments only, so the repo needed an explicit closed-audit record without implying that policy, scripts, or snapshots changed.
- **Affected artifacts**: `overlay/gentle-ai/state/upstream-state.json`
- **Verification**: `bash audit-gentle-ai-upstream.sh` with `base prompt drift: no` and all profile/base-asset invariants `ok`

## 2026-06-01 | Unified pre-sync audit, post-sync apply, and maintainer runtime reporting

- **Type**: `tooling-change`
- **Upstream scope/range**: maintenance runtime, not a new upstream boundary
- **Decision**: separated upstream auditing from overlay application, moved the maintainer runtime behind the shared Go CLI, added fail-closed baseline verification to apply, exposed file-level `--verbose` reporting, and documented `opencode` as the minimum re-apply target with `all` as optional broader refresh.
- **Why it mattered**: before `gentle-ai sync`, local `opencode.json` does not reliably expose the new upstream inline prompt; the maintainer needed a real pre-sync auditor, a single shared runtime implementation, and better visibility into what apply actually changed on disk.
- **Affected artifacts**: `cmd/gentle-ai-overlay/main.go`, `internal/overlay/audit_upstream.go`, `internal/overlay/apply_custom.go`, `internal/overlay/apply_policy.go`, `internal/overlay/overlays.go`, `internal/overlay/profiles.go`, `internal/overlay/snapshots.go`, `internal/overlay/summary.go`, `internal/overlay/util.go`, `internal/overlay/verbose.go`, `apply-gentle-ai-custom.sh`, `apply-gentle-ai-custom.ps1`, `audit-gentle-ai-upstream.sh`, `audit-gentle-ai-upstream.ps1`, `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh`, `overlay/gentle-ai/scripts/apply-gentle-ai-policy.ps1`, `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.meta.yaml`, `AGENTS.md`, `README.md`, `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`, `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`, `overlay/gentle-ai/README.md`
- **Verification**: `go test ./...`; `bash audit-gentle-ai-upstream.sh`; `bash apply-gentle-ai-custom.sh all`; `go run ./cmd/gentle-ai-overlay --help`; `go run ./cmd/gentle-ai-overlay apply-policy --help`; `go run ./cmd/gentle-ai-overlay apply-custom --help`

## 2026-05-30 | Localized SDD profile-managed state and tightened its contract

- **Type**: `tooling-change`
- **Upstream scope/range**: local SDD profile management contract
- **Decision**: moved named SDD profile assignments and per-profile snapshots out of the versioned repo, introduced strict local profile reconciliation, fixed Bash/PowerShell validation parity, and kept only the portable `gentle-orchestrator` baseline in git.
- **Why it mattered**: per-machine profile choices were leaking into shared policy and snapshots, and Windows accepted states that Bash rejected; the overlay needed a portable repo boundary plus strict cross-platform validation.
- **Affected artifacts**: `overlay/gentle-ai/policy/gentle-ai-policy.json`, `overlay/gentle-ai/policy/maintenance-intent.md`, `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh`, `overlay/gentle-ai/scripts/apply-gentle-ai-policy.ps1`, `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/`, `AGENTS.md`, `README.md`, `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`, `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`, `overlay/gentle-ai/README.md`
- **Verification**: idempotent runs with and without local config; negative validation tests; positive update/create tests; Bash/PowerShell parity inspection; confirmation that the repo snapshot tree keeps only `gentle-orchestrator.last.md`

## 2026-05-30 | Closed the upstream v1.33.2 audit and made adoption guidance explicit

- **Type**: `audit`
- **Upstream scope/range**: `412eed3d39defb2f955a63e21ca13cef4df358c9` (`v1.32.0`) -> `0fa9f2d1d2d3a8ebd822cdd5c82fcb4bff60f0fc` (`v1.33.2`)
- **Decision**: closed the audit without changing keep/prune or sanitizer behavior, advanced the maintained boundary to `v1.33.2`, and made `sync` vs reinstall an explicit maintainer recommendation in the workflow contract.
- **Why it mattered**: the upstream range introduced JD agents and sub-agent launch deduplication, but did not break the overlay; the maintainer still needed an explicit rule for when topology drift requires reinstall instead of `gentle-ai sync`.
- **Affected artifacts**: `overlay/gentle-ai/state/upstream-state.json`, `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`, `README.md`, `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`, `overlay/gentle-ai/README.md`
- **Verification**: upstream range review (`git log`, `git diff --name-only`, `git show --stat`); direct prompt-anchor inspection; consistency review across maintainer docs and skill

## 2026-05-29 | Hardened the apply pipeline and maintainer recovery workflow

- **Type**: `incident`
- **Upstream scope/range**: overlay apply/recovery path
- **Decision**: consolidated the hardening work around `gentle-ai sync` resets into a stronger apply pipeline with update-type triage, mandatory re-apply rules, snapshot recovery, topology warnings, post-write verification, safer sanitizer behavior, explicit `ERROR:` contracts, execute-bit-safe invocation, and clearer steady-state reporting.
- **Why it mattered**: `gentle-ai sync` resets orchestrator prompts and reinstalls skills; the overlay needed reliable recovery/verification behavior and cross-platform parity when local state or upstream anchors drift.
- **Affected artifacts**: `apply-gentle-ai-custom.sh`, `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh`, `overlay/gentle-ai/scripts/apply-gentle-ai-policy.ps1`, `overlay/gentle-ai/policy/gentle-ai-policy.json`, `AGENTS.md`, `README.md`, `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`, `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`
- **Verification**: idempotent `bash apply-gentle-ai-custom.sh all`; manual snapshot-recovery test; fresh-context adversarial review findings resolved; parity inspection across Bash/PowerShell behavior

## 2026-05-29 | Established the maintenance control plane and dynamic overlay model

- **Type**: `policy-change`
- **Upstream scope/range**: repository maintenance model
- **Decision**: formalized the split between intent, policy, state, and log; documented the update types (`brew upgrade`, `gentle-ai sync`, reinstall); made the maintainer workflow explicitly version-aware; and switched the overlay to canonical custom entrypoints with dynamic orchestrator generation from inline upstream prompts.
- **Why it mattered**: the repo needed both a durable maintenance contract and a runtime model tied to real upstream materialization instead of ad hoc prompts or a repo-owned static orchestrator derivative.
- **Affected artifacts**: `overlay/gentle-ai/policy/maintenance-intent.md`, `overlay/gentle-ai/state/upstream-state.json`, `overlay/gentle-ai/runbooks/maintain-upstream-overlay.md`, `overlay/gentle-ai/logs/update-log.md`, `AGENTS.md`, `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`, `apply-gentle-ai-custom.sh`, `apply-gentle-ai-custom.ps1`, `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh`, `overlay/gentle-ai/scripts/apply-gentle-ai-policy.ps1`
- **Verification**: documentation derived from direct upstream code inspection (`internal/cli/sync.go`, `internal/components/sdd/inject.go`, `internal/components/sdd/profiles.go`); manual apply-flow review; entrypoint and maintainer-workflow alignment review

## 2026-05-28 | Bootstrapped the overlay baseline and cross-platform apply contract

- **Type**: `tooling-change`
- **Upstream scope/range**: initial overlay baseline
- **Decision**: created the initial overlay structure, established keep/prune defaults, added the first apply helpers, fixed early Bash/PowerShell parity issues, and introduced built-in OpenCode agent overrides.
- **Why it mattered**: this was the initial operational baseline that turned the repo into a real overlay/control plane rather than a loose collection of scripts.
- **Affected artifacts**: `overlay/gentle-ai/**`, `AGENTS.md`, `overlay/gentle-ai/README.md`, the initial apply helpers, and the first snapshot/policy assets
- **Verification**: parity review across Bash/PowerShell helpers and manual validation of the initial keep/prune plus override behavior
