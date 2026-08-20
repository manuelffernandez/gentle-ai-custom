# Gentle AI Overlay Update Log

This file records only closed upstream-maintenance / overlay-maintenance events and maintenance-contract changes that matter for alignment with Gentle AI upstream.

It is not a mirror of every repo change. Git history carries implementation-level edits, intermediate iterations, and doc wording churn. `overlay/gentle-ai/state/upstream-state.json` remains the source of truth for the last maintained upstream boundary.

## 2026-08-20 | Adopted the Gentle AI v2.4.0 repository boundary

- **Date**: `2026-08-20`
- **Title**: Adopted the Gentle AI v2.4.0 repository boundary
- **Type**: `adoption`
- **Upstream scope/range**: `d299eb8a4009f770a605174410b0cbacc5c7524f` (`v2.3.0`) -> `301fb2ad7f3f3bda71f516d6e2848ef3fa6fe9bb` (`v2.4.0`)
- **Decision**: `Adquirir` the relevant managed OpenCode changes, including conditional background-subagent prompt composition, the new review transport and task-result artifact flow, the sixth global `review-validator` actor, and the audit/profile support required by the v2.4.0 topology and materialization model. `Ignorar` unrelated platform and housekeeping changes outside the maintained target. No `Sanitizar` was required; keep/prune policy, runtime ownership, local patch behavior, and RDD activation state remain unchanged.
- **Why it mattered**: v2.4.0 changes maintained OpenCode review transport, actor topology, prompt composition, and runtime materialization. Advancing the repository boundary and snapshots recorded the accepted upstream behavior; the separately authorized runtime sync materialized it in OpenCode, and the later separately authorized overlay apply restored the intended local layer without activating RDD.
- **Affected artifacts**: v2.4.0 audit/profile support and tests under `internal/overlay/`; `overlay/gentle-ai/policy/{managed-assets.json,maintenance-intent.md}`; refreshed snapshots under `overlay/gentle-ai/assets/upstream/opencode/`, including `background-subagents.md`, review/task-result plugins, orchestrator/overlay assets, shared contracts, and affected skills; `overlay/gentle-ai/state/upstream-state.json`; `AGENTS.md`; `README.md`; `overlay/gentle-ai/{README.md,maintenance.md,logs/update-log.md}`
- **Verification**: the authorized `bash sync-gentle-ai-upstream-assets.sh` completed successfully with 17 managed range changes, 3 new files, 13 changed files, 50 unchanged files, 1 deleted file, 18 directories synced, and upstream state changed to verified release `2.4.0`, exact tag `v2.4.0`, commit `301fb2ad7f3f3bda71f516d6e2848ef3fa6fe9bb`. The separately authorized `gentle-ai sync --agent opencode --opencode-background-subagents=auto` then completed successfully with 25 files reported changed. The later separately authorized `bash apply-gentle-ai-custom.sh opencode` updated OpenCode config and reported: skills pruned 2; owned asset writes new 0, changed 1, unchanged 0, deleted 0; prompt references updated 0, unchanged 0; owned runtime targets verified 1; topology warnings 0; profiles managed 4; profile agents created 0, updated 5, unchanged 47; unmanaged profiles 0. Final runtime verification passed: rejected skills are absent; sanitized repo-owned `branch-pr` and `chained-pr`, all 7 repo-owned skills, all 5 wrappers, and owned `AGENTS.md` match canonical outputs; `general`/`explore` overrides and all four profiles are correct; six review actors including `review-validator` remain global with no suffixed review actors; named orchestrators reuse the current upstream base; upstream plugins, contracts, and phase prompts are preserved; and Judgment Day differs only by the retrospective patch. Background intent is `auto`, effective policy is `off`, activation is ready, no launcher is present, and no background restart is required. The final canonical audit passed with no managed drift or structural invariant mismatch.
- **Follow-up**: restart OpenCode once because `opencode.json` changed. Optionally authorize and run the RDD correction-required canary later; RDD remains off/default and blocked until it passes. Any future `gentle-ai sync` or reinstall requires a new, separately authorized overlay apply.

## 2026-08-11 | Adopted the Gentle AI v2.3.0 repository boundary

- **Date**: `2026-08-11`
- **Title**: Adopted the Gentle AI v2.3.0 repository boundary
- **Type**: `adoption`
- **Upstream scope/range**: `fdecaf570a20b042da36d0b0b6e0a21ee5fed182` (`v2.2.4`) -> `d299eb8a4009f770a605174410b0cbacc5c7524f` (`v2.3.0`)
- **Decision**: `Adquirir` the v2.3.0 SDD orchestration and phase behavior, edit-authority consent, delegation visibility, review contracts/plugin transport, post-verify review routing, typed phase failures, and mechanical archive copy/verification. `Sanitizar` only Judgment Day by retaining its standalone receipt-independent upstream protocol and updating the exact-anchor patch to add the existing terminal `judgment-retrospective` handoff. Retain `gentle-ai-bench`, `rdd-defect-workflow`, and `systemic-issue-triage` as upstream audit/sync snapshots only; they are not repo-owned install sources or forced runtime selections. Keep/prune policy, runtime ownership, profile prompt inheritance, commit/PR approval boundaries, and no-auto-commit behavior remain unchanged.
- **Why it mattered**: the repository baseline now captures the accepted v2.3.0 SDD/review behavior without forking upstream protocols, changing maintained topology, enabling RDD, or implying a runtime refresh.
- **Affected artifacts**: `overlay/gentle-ai/policy/managed-assets.json`; `overlay/gentle-ai/patches/opencode/judgment-retrospective.json`; refreshed v2.3.0 snapshots under `overlay/gentle-ai/assets/upstream/opencode/{commands,plugins,prompts,skills}`; `overlay/gentle-ai/state/upstream-state.json`; `internal/overlay/managed_assets_test.go`; `AGENTS.md`; `README.md`; `overlay/gentle-ai/README.md`; `overlay/gentle-ai/maintenance.md`; `overlay/gentle-ai/logs/update-log.md`
- **Verification**: the authorized one-time `bash sync-gentle-ai-upstream-assets.sh` reported 21 managed files changed, 3 new snapshots, 18 changed snapshots, 43 unchanged files, no deletions, and advanced the exact boundary to `v2.3.0` / `d299eb8a4009f770a605174410b0cbacc5c7524f`. The post-sync canonical audit reported no managed or unmapped drift, no base prompt drift, and all profile/topology invariants `ok`. Focused runtime-patch/ownership tests and `go test ./internal/overlay -count=1` passed; the temporary-runtime verifier proved Judgment Day differs from the adopted snapshot only by declared replacements.
- **Follow-up**: the recommended future runtime refresh is `gentle-ai sync`, followed by separately authorized overlay apply. Runtime sync, reinstall, overlay apply, global configuration mutation, commit, push, and PR creation were not executed. RDD remains disabled until source/tests and the correction-required canary pass.

## 2026-08-03 | Restored the Gentle AI v2.2.4 snapshot adoption

- **Type**: `adoption`
- **Upstream scope/range**: `09c2d8b6369860058aa44704df76b1b33c745675` (`v2.2.2`) -> `fdecaf570a20b042da36d0b0b6e0a21ee5fed182` (`v2.2.4`)
- **Decision**: `Adquirir` provider-defect handoff and attempt routing, passive SDD learning closure, safe reviewer recapture, provider-owned accepted correction findings, and native reviewer inspection changes. `Sanitizar` the updated `issue-creation` audit snapshot while preserving the local prune policy. `Ignorar` unmanaged binary, model, MCP, CI, and documentation changes outside the maintained scope. Refreshed six managed snapshots and advanced `overlay/gentle-ai/state/upstream-state.json`; keep/prune, runtime ownership, topology, and local Judgment Day patch behavior remain unchanged. This closes the repo-only adoption; runtime reinstall and overlay apply remain separately authorized human actions.
- **Why it mattered**: the maintained audit baseline now captures the accepted v2.2.4 review and SDD behavior without expanding the local runtime surface or implying that the installed runtime was refreshed.
- **Affected artifacts**: `overlay/gentle-ai/assets/upstream/opencode/plugins/review-result-artifacts.ts`; `overlay/gentle-ai/assets/upstream/opencode/prompts/orchestrators/gentle-orchestrator.md`; `overlay/gentle-ai/assets/upstream/opencode/skills/_shared/review-ledger-contract.md`; `overlay/gentle-ai/assets/upstream/opencode/skills/_shared/sdd-phase-common.md`; `overlay/gentle-ai/assets/upstream/opencode/skills/_shared/sdd-status-contract.md`; `overlay/gentle-ai/assets/upstream/opencode/skills/issue-creation/SKILL.md`; `overlay/gentle-ai/state/upstream-state.json`; `overlay/gentle-ai/logs/update-log.md`
- **Verification**: exact binary/tag/HEAD parity at `v2.2.4`; all six snapshots matched upstream byte-for-byte with aligned modes; the canonical audit reported no drift and all structural invariants `ok`; `git diff --check` passed; fresh-context review found the combined repo state consistent. These were repo-only checks; no runtime reinstall or overlay apply was performed.
- **Follow-up**: runtime reinstall and overlay apply remain pending and require separate human authorization. Receipt-Driven Development remains disabled until the required correction canary passes.

## 2026-08-03 | Made release parity immediate and runtime refresh human-owned

- **Type**: `policy-change`
- **Upstream scope/range**: maintenance workflow contract at verified installed `gentle-ai 2.2.4`, exact tag `v2.2.4`, and release commit `fdecaf570a20b042da36d0b0b6e0a21ee5fed182`; no upstream boundary advance
- **Decision**: made the canonical audit a synchronous activation hard gate immediately after minimal maintainer skill/context load and before broad repository exploration or repository-reading/drift-analysis delegation. A parity failure now requires an immediate stop with no delegation; `last_maintained_commit` remains only the drift-range boundary. Reserved `gentle-ai sync` and reinstall execution for a specific human request, while keeping audit/adoption approval, repo snapshot sync, runtime refresh, overlay edits, and overlay apply as separately authorized actions.
- **Why it mattered**: the guard algorithm already rejected binary/tag/HEAD mismatch before drift analysis, but the maintainer orchestration delayed invoking it until after broad reads and could execute runtime refresh after general adoption approval. The workflow now fails closed at activation and does not treat repo approval as permission to mutate the installed runtime.
- **Affected artifacts**: `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`; `.opencode/commands/maintain.md`; `overlay/gentle-ai/policy/maintenance-intent.md`; `AGENTS.md`; `README.md`; `overlay/gentle-ai/maintenance.md`; `overlay/gentle-ai/README.md`; `overlay/gentle-ai/logs/update-log.md`
- **Verification**: canonical read-only audit passed exact `v2.2.4` parity and reported no managed drift or structural mismatch; final contradiction search and `git diff --check` passed. No install, runtime refresh, repo snapshot sync, overlay apply, commit, push, or PR was performed.

## 2026-07-30 | Enforced exact release parity before upstream audit and sync

- **Type**: `tooling-change`
- **Upstream scope/range**: maintenance tooling contract; installed `gentle-ai 2.2.2`, exact `v2.2.2` commit `09c2d8b6369860058aa44704df76b1b33c745675`, and local upstream HEAD `d260cdbc8c4be083333cb8b0517ed125a450a411`; no upstream boundary advance
- **Decision**: made audit and snapshot sync fail before analysis, copy, or writes unless the resolved upstream checkout is fully clean and its `HEAD` equals the commit peeled from exact `refs/tags/v<V>` for installed `gentle-ai --version`. Prereleases require their exact tag. The tools never use `git describe`, nearest tags, version ordering, or ancestry as parity proof and never fetch or check out automatically. Apply keeps its separate maintained-baseline version guard and does not require an upstream clone.
- **Why it mattered**: the previous audit inspected live upstream `HEAD`, and sync copied the live worktree while persisting describe-derived identity. A branch ahead of a published release could therefore audit or sync unreleased commits even though the installed binary represented the stable release.
- **Affected artifacts**: shared release guard, audit/sync orchestration, and focused tests under `internal/overlay/`; `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`; `.opencode/commands/maintain.md`; `overlay/gentle-ai/policy/maintenance-intent.md`; `overlay/gentle-ai/maintenance.md`; `overlay/gentle-ai/owned-assets-refactor.md`; `README.md`; `AGENTS.md`; `overlay/gentle-ai/README.md`; `overlay/gentle-ai/logs/update-log.md`
- **Verification**: `gofmt` on changed Go files; `go test ./internal/overlay -count=1`; `go test ./... -count=1`; `git diff --check`. Focused tests cover stable and prerelease exact matches, ahead/newer/older/missing-tag refusal, staged/unstaged/untracked/submodule dirt, no sync mutation on refusal, exact release identity persistence on successful sync, and apply's clone-independent maintained-baseline guard. A read-only real audit canary stopped before drift analysis because upstream HEAD `d260cdbc8c4be083333cb8b0517ed125a450a411` did not equal exact `v2.2.2` commit `09c2d8b6369860058aa44704df76b1b33c745675`.

## 2026-07-30 | Adopted stable Gentle AI v2.2.2 review lifecycle

- **Type**: `adoption`
- **Upstream scope/range**: `ee83e83d56f0d149c52f93fd13b3296858f5147f` (`v2.2.0`) -> `09c2d8b6369860058aa44704df76b1b33c745675` (`v2.2.2`)
- **Decision**: `Adquirir` the stable negotiated review and frozen-context lifecycle plus the archive disabled-review fix. The initially approved prerelease `d260cdbc8c4be083333cb8b0517ed125a450a411` was replaced by the exact Brew stable boundary for repo, binary, and runtime reproducibility. `Ignorar` for now the post-v2.2.2 prerelease-only attempt/retry, Key Learnings, and issue-workflow changes. Keep/prune, runtime ownership, and local patch behavior remain unchanged.
- **Why it mattered**: the stable boundary preserves the accepted review improvements and fixes archive deadlock when review is disabled, while excluding unpublished behavior and keeping the maintained snapshots, installed binary, and runtime exactly aligned.
- **Affected artifacts**: approved snapshots `overlay/gentle-ai/assets/upstream/opencode/commands/sdd-apply.md`, `overlay/gentle-ai/assets/upstream/opencode/plugins/review-result-artifacts.ts`, `overlay/gentle-ai/assets/upstream/opencode/skills/_shared/review-ledger-contract.md`, and `overlay/gentle-ai/assets/upstream/opencode/skills/sdd-archive/SKILL.md`; `overlay/gentle-ai/state/upstream-state.json`; installed OpenCode runtime under `~/.config/opencode/`.
- **Verification**: post-sync audit, `go test ./internal/overlay`, `go test ./... -count=1`, `git diff --check`, and `opencode debug config --pure` passed. Runtime ownership, hashes, modes, prompt/profile invariants, pruned skills, and the Judgment Day retrospective-only delta matched expectations. `gentle-ai sync` updated 23 runtime files; overlay apply pruned 2 skills, changed 1 owned asset, updated 0 prompt refs, verified 1 owned target, managed 4 profiles, updated 6 agents, left 46 unchanged, and reported 0 topology warnings and 0 unmanaged profiles.
- **Follow-up**: restart OpenCode because `opencode.json` changed.

## 2026-07-28 | Added local Judgment Day and review assignment reconciliation

- **Type**: `tooling-change`
- **Upstream scope/range**: maintained Gentle AI `v2.2.0` runtime topology at `ee83e83d56f0d149c52f93fd13b3296858f5147f`; no upstream boundary advance
- **Decision**: extended schema version 1 so `default_profile` can explicitly assign or inherit the three global Judgment Day actors and five global review actors. Named profiles may explicitly clone only Judgment Day actors with profile suffixes; omitted JD slots delegate globally and review actors always remain unsuffixed. Added fail-closed source-actor checks and an audit invariant for the exact upstream actor key sets.
- **Why it mattered**: omission already means preserve upstream/runtime assignment, so dynamic inheritance needed an unambiguous `{"inherit": true}` value. The overlay also needed to follow upstream's optional named-JD topology without fabricating review actors or copying prompt/tool semantics into custom sources.
- **Affected artifacts**: local config parsing, profile reconciliation, runtime preflight, verification, and tests under `internal/overlay/`; `overlay/gentle-ai/policy/managed-assets.json`; maintainer policy/docs/skill; `~/.config/gentle-ai-custom/opencode-local-config.json`; `~/.config/opencode/opencode.json`; `overlay/gentle-ai/logs/update-log.md`.
- **Verification**: focused table-driven actor/schema/reconciliation tests passed; `go test ./internal/overlay -count=1` and `go test ./... -count=1` passed; `bash audit-gentle-ai-upstream.sh` reported no drift and all structural invariants `ok`, including JD/review topology. The first `bash apply-gentle-ai-custom.sh opencode` reported config updated, 1 skill pruned, 1 owned asset unchanged, 1 owned target verified, 4 profiles managed, 13 profile agents updated, 39 unchanged, 0 created, 0 topology warnings, and 0 unmanaged profiles. Readback confirmed exact global JD assignments, all five review actors without `model` or `variant`, three intact named profiles, no suffixed JD/review actors, global JD/review delegation permissions, and unchanged `build`/`plan` sentinels. A convergence apply reported config unchanged, 0 profile agents updated, 52 unchanged, 0 topology warnings, and 0 unmanaged profiles. `gentle-ai skill-registry refresh --force` regenerated `.atl/skill-registry.md` with 23 skills; `git diff --check` passed.
- **Follow-up**: restart OpenCode so the active process loads the updated global configuration.

## 2026-07-27 | Adopted Gentle AI v2.2.0 after full OpenCode reinstall

- **Type**: `adoption`
- **Upstream scope/range**: `4907129ec9dc0887c6a48548e9b3ef206f25bbed` (`v2.1.11-94-g4907129e`) -> `ee83e83d56f0d149c52f93fd13b3296858f5147f` (`v2.2.0`)
- **Decision**: acquired the v2.2.0 OpenCode runtime and approved snapshots after adapting the structural auditor to accept the canonical adapter-local `agent` render argument without weakening its exact assignment checks. Advanced the maintained boundary, completed a full TUI reinstall, and reapplied the local overlay with the existing keep/prune, ownership, profile, Judgment Day, and plan-only chained-PR boundaries unchanged.
- **Why it mattered**: v2.2.0 changed maintained OpenCode materialization and the canonical SDD orchestrator render call. The old audit predicate rejected the valid v2.2.0 source shape, while a sync-only refresh could not prove the newly installed topology and runtime files matched the accepted release.
- **Affected artifacts**: `internal/overlay/audit_upstream.go`; `internal/overlay/managed_assets_test.go`; 11 approved snapshots under `overlay/gentle-ai/assets/upstream/opencode/`; `overlay/gentle-ai/state/upstream-state.json`; installed OpenCode config and custom runtime under `~/.config/opencode/`; `overlay/gentle-ai/logs/update-log.md`.
- **Verification**: upstream and binary resolved to exact clean v2.2.0 at `ee83e83d56f0d149c52f93fd13b3296858f5147f`. Pre-apply topology inspection found all base agents, all three declared named profiles, all 30 suffixed phase agents, exact v2.2.0 Judgment Day patch anchors, and no unmanaged profile or obsolete managed prompt reference. The canonical apply reported config updated, 2 skills pruned, 1 owned asset changed, 1 owned runtime target verified, 4 profiles managed, 34 profile agents updated, 10 unchanged, 0 created, 0 topology warnings, and 0 unmanaged profiles. Post-state checks confirmed local assignments and prompt reuse, scoped pruning, owned target presence, unchanged representative upstream prompt/plugin/contract bytes and modes, the exact declared retrospective-only Judgment Day delta, and the canonical plan-only `chained-pr` runtime. A convergence apply left config and all 44 profile agents unchanged with 0 topology warnings. `bash audit-gentle-ai-upstream.sh`, `go test ./...`, `opencode debug config --pure`, and `git diff --check` passed.
- **Follow-up**: restart OpenCode so the active process loads the updated global configuration; the parent fresh-context final review remains separate.

## 2026-07-22 | Closed raw-source runtime overwrite incident

- **Type**: `incident`
- **Upstream scope/range**: installed release runtime `v2.1.11` versus approved source snapshots through `4907129ec9dc0887c6a48548e9b3ef206f25bbed` (`v2.1.11-94-g4907129e`)
- **Decision**: stopped installing raw upstream source snapshots as runtime artifacts. Upstream Gentle AI exclusively owns and validates the native orchestrator, phase prompts, contracts, commands, skills, review plugin, and review semantics. Custom apply validates only its own prerequisites, preserves consumed upstream files byte-for-byte with their modes, accepts the upstream internal `permission.task.__replace__` input shape when reconciling named profiles, persists schema-valid direct task permission maps with upstream suffixing semantics, and applies only the exact Judgment Day retrospective patch. Owned-state migration records every custom-installed output while protecting upstream-returned trees and patch targets without claiming whole-file ownership. Plugin absence or opaque content does not block unrelated custom operations.
- **Why it mattered**: custom apply copied raw `sdd-orchestrator.md` over the upstream-rendered prompt, removing `GENTLE_AI_REVIEW_BINDING` and breaking plugin capture/finalize/receipt flow while pre-commit correctly rejected the missing review receipt. The old model also mixed accepted post-release source snapshots with a release runtime.
- **Affected artifacts**: `internal/overlay/{apply_custom,apply_policy,managed_assets,owned_assets,profiles,runtime_owned_state,runtime_preflight}.go` and tests; deletion of `internal/overlay/typescript_structure.go` and its tests; `overlay/gentle-ai/policy/managed-assets.json`; `overlay/gentle-ai/patches/opencode/judgment-retrospective.json`; removal of the raw installable SDD/Judgment Day fork under `overlay/gentle-ai/assets/owned/opencode/`; maintainer docs and policy; repaired derived runtime config at `~/.config/opencode/opencode.json`.
- **Verification**: temporary-home tests used arbitrary opaque orchestrator/plugin/contract bytes, proved plugin absence succeeds, verified upstream files remain byte- and mode-unchanged, proved custom prerequisite failures occur before writes, and confirmed successful apply is idempotent. Owned-state migration tests proved patched Judgment Day files, upstream SDD/shared files, and freshly installed `AGENTS.md` survive same-run reconciliation; the new state claims `AGENTS.md` but not patch targets; obsolete custom-only files are deleted; prefix collisions and outside-root paths are rejected; and repeat apply is idempotent. Profile creation/reconciliation tests prove direct named-orchestrator task maps retain wildcard deny, global/review/Judgment Day allows, suffixed phase allows, no leaked `__replace__`, and byte-idempotency. `go test ./internal/overlay -count=1` and `go test ./... -count=1` passed. The three derived runtime profile blocks were unwrapped directly, retained mode `0644`, passed focused `jq` assertions, and loaded successfully through `opencode debug config --pure` on OpenCode 1.18.2. No full apply was run.
- **Follow-up**: restart OpenCode so the repaired global configuration is loaded by the active process.

## 2026-07-22 | Adopted native bounded review orchestration

- **Type**: `adoption`
- **Upstream scope/range**: `b47f8af5d1cc9f6368bff17ac9e1c5694713d2c5` (`v2.0.2-24-gb47f8af`) -> `4907129ec9dc0887c6a48548e9b3ef206f25bbed` (`v2.1.11-94-g4907129e`)
- **Decision**: adopted the approved managed review/runtime assets replacement-first: each owned baseline is copied verbatim from its approved upstream snapshot, then only the documented Judgment Day retrospective terminal handoff is reapplied. The obsolete owned `native-review-contract.md` duplicate was deleted. Repo adoption is closed; binary and runtime refresh remain intentionally deferred to the user.
- **Why it mattered**: native result capture, immutable candidate binding, canonical validation, bounded correction, receipt validation, and authority-first terminal handling replace the obsolete model-managed review lifecycle.
- **Affected artifacts**: upstream snapshots `overlay/gentle-ai/assets/upstream/opencode/skills/_shared/{review-ledger-contract,sdd-status-contract}.md`, `overlay/gentle-ai/assets/upstream/opencode/prompts/orchestrators/gentle-orchestrator.md`, `overlay/gentle-ai/assets/upstream/opencode/skills/judgment-day/{SKILL.md,references/prompts-and-formats.md}`, and `overlay/gentle-ai/assets/upstream/opencode/plugins/review-result-artifacts.ts`; owned runtime `overlay/gentle-ai/assets/owned/opencode/prompts/orchestrators/gentle-orchestrator.md`, `overlay/gentle-ai/assets/owned/opencode/skills/_shared/{review-ledger-contract,sdd-status-contract}.md`, `overlay/gentle-ai/assets/owned/opencode/skills/judgment-day/{SKILL.md,references/prompts-and-formats.md}`, and deleted `overlay/gentle-ai/assets/owned/opencode/skills/_shared/native-review-contract.md`; maintainer docs, `internal/overlay/managed_assets_test.go`, and `overlay/gentle-ai/state/upstream-state.json`.
- **Verification**: `bash sync-gentle-ai-upstream-assets.sh`; `go test ./internal/overlay -run TestOwnedReviewRuntimeUsesNativeAuthorityContracts`; `go test ./internal/overlay`; `bash audit-gentle-ai-upstream.sh`; and `git diff --check`. All repo-only checks; no runtime mutation.
- **Follow-up**: later update binary, run `gentle-ai sync`, then run `bash apply-gentle-ai-custom.sh opencode` and restart OpenCode if config changes.

## 2026-07-14 | Closed managed SDD reference coverage gap for sdd-init

- **Type**: `tooling-change`
- **Upstream scope/range**: current approved v2 boundary `b47f8af5d1cc9f6368bff17ac9e1c5694713d2c5` (`v2.0.2-24-gb47f8af`), post-sync maintenance fix
- **Decision**: added `sdd-init/references/init-details.md` to `managed-assets.json` as an explicit required SDD shared file, materialized matching upstream and owned copies inside the repo, and added behavior-level overlay tests that fail when required SDD phase skills link to unmanaged or missing `references/*.md` files.
- **Why it mattered**: the repo had accepted the v2 snapshot boundary, but `sdd-init/SKILL.md` in both trees linked to a missing local reference file that was neither tracked by the manifest nor covered by tests. That gap would let future upstream/shared-reference drift bypass audit confidence and repo snapshot consistency.
- **Affected artifacts**: `overlay/gentle-ai/policy/managed-assets.json`, `overlay/gentle-ai/assets/upstream/opencode/skills/sdd-init/references/init-details.md`, `overlay/gentle-ai/assets/owned/opencode/skills/sdd-init/references/init-details.md`, `internal/overlay/managed_assets_test.go`, `overlay/gentle-ai/logs/update-log.md`
- **Verification**: `gofmt -w internal/overlay/managed_assets_test.go`; `go test ./internal/overlay`; `git diff --check`; SHA-256 comparison against upstream `internal/assets/skills/sdd-init/references/init-details.md` confirmed the new repo upstream snapshot file matches upstream HEAD exactly.
- **Follow-up**: no runtime refresh or overlay re-apply was performed in this maintenance fix.

## 2026-07-14 | Synced upstream snapshots to the current v2 boundary

- **Type**: `adoption`
- **Upstream scope/range**: `b6737129a295980962ec5e8daf41c971fac700b9` (`v1.44.3`) -> `b47f8af5d1cc9f6368bff17ac9e1c5694713d2c5` (`v2.0.2-24-gb47f8af`)
- **Decision**: advanced the approved upstream snapshot boundary and refreshed only the repo's upstream snapshot assets. No `gentle-ai sync`, reinstall, overlay re-apply, or owned-runtime mutation was performed.
- **Why it mattered**: the owned runtime had already been prepared for the upstream-first v2 behavior, but the approved upstream snapshot tree still mixed `v1.44.3`, newer v2 content, and files that matched neither boundary. Refreshing the snapshots restores the audit baseline without changing the installed runtime layer.
- **Affected artifacts**: `overlay/gentle-ai/assets/upstream/opencode/commands/sdd-apply.md`, `overlay/gentle-ai/assets/upstream/opencode/commands/sdd-archive.md`, `overlay/gentle-ai/assets/upstream/opencode/commands/sdd-verify.md`, `overlay/gentle-ai/assets/upstream/opencode/engram-protocol.md`, `overlay/gentle-ai/assets/upstream/opencode/prompts/orchestrators/gentle-orchestrator.md`, `overlay/gentle-ai/assets/upstream/opencode/sdd-overlay-single.json`, `overlay/gentle-ai/assets/upstream/opencode/sdd-overlay-multi.json`, `overlay/gentle-ai/assets/upstream/opencode/skills/_shared/*.md`, `overlay/gentle-ai/assets/upstream/opencode/skills/judgment-day/`, `overlay/gentle-ai/assets/upstream/opencode/skills/sdd-*/SKILL.md`, `overlay/gentle-ai/assets/upstream/opencode/skills/work-unit-commits/SKILL.md`, `overlay/gentle-ai/state/upstream-state.json`
- **Verification**: read-only git diff confirmed the only newly added managed upstream asset in the range was `internal/assets/skills/sdd-design/references/threat-matrix.md`, already tracked by `sdd-design-threat-matrix` and already present in the local snapshot; `bash sync-gentle-ai-upstream-assets.sh` reported 26 managed files changed, 24 files changed, 35 unchanged, 15 directories synced, and upstream state changed; fresh-context reliability review found no blockers. `managed-assets.json` did not change.
- **Follow-up**: runtime refresh remains intentionally pending; do not run `gentle-ai sync` or `apply-gentle-ai-custom.sh` until the user explicitly requests that later phase.

## 2026-07-13 | Prepared the owned runtime for the upstream-first SDD v2 baseline

- **Type**: `policy-change`
- **Upstream scope/range**: owned-runtime preparation for `b6737129a295980962ec5e8daf41c971fac700b9` (`v1.44.3`) -> `b47f8af5d1cc9f6368bff17ac9e1c5694713d2c5` (`v2.0.2-24-gb47f8af`) without boundary advance
- **Decision**: refreshed the repo-owned OpenCode orchestrator, shared SDD contracts, SDD phase skills, and SDD commands so the owned runtime defaults to the upstream v2 behavior while preserving only the confirmed local workflow layer (`branch-pr` pruned/replaced, `issue-creation` pruned, sanitized runtime `chained-pr`, local shared skills/hooks). Added the new upstream SDD support references for the design threat matrix and verify report format to the managed-assets/runtime surface. Relaxed maintainer audit invariants so they verify deny-by-default task scoping and the canonical OpenCode orchestrator asset render path instead of failing on obsolete string literals.
- **Why it mattered**: the repo-owned SDD/runtime surface was still carrying v1.44-era overlay behavior. Reacquiring upstream v2 by default shrinks local drift, preserves only real custom behavior, and prepares a later approved sync/apply cycle without reintroducing issue-first PR gating or unsafe automatic branch/PR mutation.
- **Affected artifacts**: `overlay/gentle-ai/assets/owned/opencode/prompts/orchestrators/gentle-orchestrator.md`, `overlay/gentle-ai/assets/owned/opencode/skills/_shared/*.md`, `overlay/gentle-ai/assets/owned/opencode/skills/sdd-*/SKILL.md`, `overlay/gentle-ai/assets/owned/opencode/skills/sdd-design/references/threat-matrix.md`, `overlay/gentle-ai/assets/owned/opencode/skills/sdd-verify/references/report-format.md`, `overlay/gentle-ai/assets/owned/opencode/commands/sdd-*.md`, `overlay/gentle-ai/assets/upstream/opencode/skills/sdd-design/references/threat-matrix.md`, `overlay/gentle-ai/assets/upstream/opencode/skills/sdd-verify/references/report-format.md`, `overlay/gentle-ai/policy/managed-assets.json`, `internal/overlay/audit_upstream.go`, `internal/overlay/managed_assets_test.go`, `README.md`, `overlay/gentle-ai/README.md`, `overlay/gentle-ai/maintenance.md`, `AGENTS.md`, `overlay/gentle-ai/logs/update-log.md`
- **Verification**: `git diff --check`; `go test ./internal/overlay`; repo searches for stale `issue-creation` runtime claims, upstream branch-pr dependencies inside owned SDD assets, old local orchestrator sanitization wording, and missing managed reference files; `git status --short --untracked-files=all`. This event did not advance `overlay/gentle-ai/state/upstream-state.json` and did not run upstream snapshot sync, runtime refresh, or overlay apply.
- **Follow-up**: approved upstream snapshot sync, runtime refresh, and `upstream-state.json` advancement remain intentionally pending.

## 2026-07-13 | Pruned issue-creation and tightened chained-pr runtime safety

- **Type**: `policy-change`
- **Upstream scope/range**: maintenance contract / OpenCode runtime surface, not a new upstream boundary
- **Decision**: kept `issue-creation` auditable in upstream snapshots, but moved it to the pruned runtime boundary and removed the repo-owned shared install surface so `apply-gentle-ai-custom` no longer installs or preserves an issue workflow by default. Kept sanitized `chained-pr` installable and tightened it so automatic SDD/review-budget activation may plan only; branch/PR mutation requires explicit user request or approved PR workflow handoff.
- **Why it mattered**: the current local overlay is not meant to enforce an upstream-collaboration-oriented issue workflow when the user is not actively collaborating with that upstream repository, and retained chained PR behavior must not bypass the local no-self-initiated PR boundary.
- **Affected artifacts**: `overlay/gentle-ai/policy/gentle-ai-policy.json`, `overlay/gentle-ai/policy/managed-assets.json`, `overlay/gentle-ai/policy/maintenance-intent.md`, `shared/skills/branch-pr/SKILL.md`, `shared/skills/chained-pr/SKILL.md`, `shared/skills/chained-pr/assets/chaining-details.md`, `shared/skills/issue-creation/SKILL.md`, `README.md`, `overlay/gentle-ai/README.md`, `overlay/gentle-ai/maintenance.md`, `AGENTS.md`, `overlay/gentle-ai/logs/update-log.md`
- **Verification**: `git diff --check`; repo search for stale `issue-creation` install/runtime claims; fresh-context risk/readability review found chained-pr execution wording and examples to tighten, then those findings were fixed inline

## 2026-07-09 | Raised exploration and fix delegation thresholds

- **Type**: `policy-change`
- **Upstream scope/range**: maintenance contract / OpenCode runtime surface, not a new upstream boundary
- **Decision**: changed the repo-owned orchestrator thresholds so exploration delegates by default only after a metadata-only size preflight for 5+ files, and fixes stay inline by default up to 3 files when safe/manageable; 4+ non-trivial writes still delegate by default.
- **Why it mattered**: small multi-file explorations and 2-3 file fixes were paying unnecessary delegation cost, losing precision through an extra agent layer even when direct inline handling was cheaper and clearer.
- **Affected artifacts**: `overlay/gentle-ai/assets/owned/opencode/prompts/orchestrators/gentle-orchestrator.md`, `overlay/gentle-ai/policy/maintenance-intent.md`, `overlay/gentle-ai/maintenance.md`, `overlay/gentle-ai/README.md`, `AGENTS.md`, `README.md`
- **Verification**: `git diff --check`; `bash apply-gentle-ai-custom.sh opencode` wrote 1 changed owned asset with 0 topology warnings; runtime prompt grep confirmed the 5+ exploration, metadata-size preflight, 4+ write, and long-session threshold changes were installed; fresh-context reliability review found stale pending-verification wording in this entry, then the finding was fixed inline

## 2026-07-09 | Added review/fix convergence guard

- **Type**: `policy-change`
- **Upstream scope/range**: maintenance contract / OpenCode runtime surface, not a new upstream boundary
- **Decision**: kept fresh review and safety gates intact, but capped automatic delegated review/fix loops. The repo-owned orchestrator now defaults to one delegated fix round plus one scoped re-review iteration, explicitly allows multiple review lenses in that iteration including full 4R, fixes small understood residuals inline when safe, and stops to ask or escalate when the same pattern persists.
- **Why it mattered**: unbounded review -> fix -> review delegation cycles waste time and tokens without improving safety when the residual findings are local, already understood, or repetitive.
- **Affected artifacts**: `overlay/gentle-ai/assets/owned/opencode/prompts/orchestrators/gentle-orchestrator.md`, `overlay/gentle-ai/policy/maintenance-intent.md`, `overlay/gentle-ai/maintenance.md`, `overlay/gentle-ai/README.md`, `AGENTS.md`, `README.md`
- **Verification**: `git diff --check`; fresh-context readability review found stale pending-verification wording in this entry, then the finding was fixed inline; `bash apply-gentle-ai-custom.sh opencode` wrote 1 changed owned asset with 0 topology warnings; runtime prompt grep confirmed the convergence guard was installed

## 2026-07-09 | Expanded scoped inline overrides to multi-file operational delegation

- **Type**: `policy-change`
- **Upstream scope/range**: maintenance contract / OpenCode runtime surface, not a new upstream boundary
- **Decision**: kept delegation as the default for context/cost/coordination-heavy work, but extended the repo-owned orchestrator contract so an explicit scoped user request may keep a specific multi-file write inline, alongside the existing 4+ file exploration override. Preserved the non-bypassable safety, permission, data-loss, security, commit/push/PR, review, and incident gates.
- **Why it mattered**: the previous contract still treated multi-file writes as an absolute delegation requirement, which caused avoidable resistance even when the human intentionally wanted a narrowly scoped inline change and the task remained safe and manageable.
- **Affected artifacts**: `overlay/gentle-ai/assets/owned/opencode/prompts/orchestrators/gentle-orchestrator.md`, `overlay/gentle-ai/policy/maintenance-intent.md`, `overlay/gentle-ai/maintenance.md`, `overlay/gentle-ai/README.md`, `AGENTS.md`, `README.md`
- **Verification**: final diff review plus repo search for stale contradictory wording around the old multi-file hard gate; upstream prompt snapshots intentionally left untouched

## 2026-07-08 | Adopted upstream v1.44.3 review ledger contract

- **Type**: `adoption`
- **Upstream scope/range**: `42cf98501413610d3dd8e94cd390e8663fc014f2` (`v1.43.4-2-g42cf985`) -> `b6737129a295980962ec5e8daf41c971fac700b9` (`v1.44.3`)
- **Decision**: adopted the review ledger contract (review convergence with persisted findings, exhaustive sweep loop, scoped re-review) into the owned orchestrator, judgment-day, and `_shared`, then converted the runtime prompts to reference the shared contract instead of repeating it. Sanitized the engram protocol source path in `managed-assets.json` after upstream consolidated and relocated the file.
- **Why it mattered**: the ledger mechanism improves review and JD convergence quality without reintroducing PR governance, chained-pr, or review-budget gates. The engram path repoint was mandatory because upstream relocated the file to `internal/assets/engram/protocol.md`, and the previous boundary needed to remain visible for audit traceability.
- **Affected artifacts**: `overlay/gentle-ai/assets/owned/opencode/prompts/orchestrators/gentle-orchestrator.md`, `overlay/gentle-ai/assets/owned/opencode/commands/sdd-status.md`, `overlay/gentle-ai/assets/owned/opencode/skills/judgment-day/SKILL.md`, `overlay/gentle-ai/assets/owned/opencode/skills/judgment-day/references/prompts-and-formats.md`, `overlay/gentle-ai/assets/owned/opencode/skills/_shared/review-ledger-contract.md` (new), `overlay/gentle-ai/assets/owned/opencode/skills/_shared/sdd-status-contract.md`, `overlay/gentle-ai/policy/managed-assets.json`, `overlay/gentle-ai/state/upstream-state.json`, `overlay/gentle-ai/maintenance.md`, `AGENTS.md`, `overlay/gentle-ai/assets/upstream/opencode/README.md`, `overlay/gentle-ai/assets/upstream/opencode/engram-protocol.md`, `shared/skills/judgment-retrospective/SKILL.md`
- **Verification**: `bash sync-gentle-ai-upstream-assets.sh`; `go test ./internal/overlay -run 'TestCategorizeGitDiffTreatsOnlyOpenCodeAndEngramSourcesAsManaged|TestSyncManagedTargetCopiesOpenCodeAgentsSourceFilesFromUpstream'`; `gentle-ai sync`; `bash apply-gentle-ai-custom.sh opencode`; post-state on-disk checks for pruned skills, prompt refs, owned assets, and upstream-state consistency; fresh-context R3 reliability review of the owned/upstream judgment-day skill and prompt mirrors plus the shared review-ledger contract block (6 files total) passed 7/7 checks, with one SUGGESTION-level gap fixed (review-ledger topic key added to orchestrator table)

## 2026-07-07 | Adopted upstream persona guardrails and sanitized orchestrator review lenses

- **Type**: `adoption`
- **Upstream scope/range**: `8a673472a402f1949f1080f72e72d760a61d28bf` (`v1.42.0-1-g8a67347`) -> `42cf98501413610d3dd8e94cd390e8663fc014f2` (`v1.43.4-2-g42cf985`)
- **Decision**: accepted the upstream reply-language guardrails into the repo-owned OpenCode AGENTS runtime asset and sanitized the owned orchestrator prompt to keep the concrete review-lens routing while preserving the local depuration that rejects PR strategy selection, review budgets, chained/stacked PR flow, `size:exception`, and review-workload branching.
- **Why it mattered**: the upstream prompt drift improved direct-reply language consistency and made the review path more concrete, but the repo must keep its owned orchestrator free of repository-governance controls that do not belong in the local workflow.
- **Affected artifacts**: `overlay/gentle-ai/assets/owned/opencode/AGENTS.md`, `overlay/gentle-ai/assets/owned/opencode/prompts/orchestrators/gentle-orchestrator.md`, `overlay/gentle-ai/assets/upstream/opencode/persona-gentleman.md`, `overlay/gentle-ai/assets/upstream/opencode/prompts/orchestrators/gentle-orchestrator.md`, `overlay/gentle-ai/state/upstream-state.json`
- **Verification**: `bash sync-gentle-ai-upstream-assets.sh`; `gentle-ai sync`; post-edit `bash apply-gentle-ai-custom.sh opencode` initially stopped on the noninteractive version preflight, then passed when rerun in a pseudo-tty; on-disk runtime checks for prompt refs, pruned skills, and upstream-state consistency

## 2026-06-30 | Softened the 4-file exploration gate with a scoped inline override

- **Type**: `policy-change`
- **Upstream scope/range**: maintenance contract / OpenCode runtime surface, not a new upstream boundary
- **Decision**: kept delegation as the default for 4+ file exploration, but allowed a user-requested inline exception for that specific exploration only; recorded the override in the owned orchestrator prompt, maintenance intent, and maintainer docs so the other hard gates stay intact.
- **Why it mattered**: the old wording treated 4-file exploration as an unskippable gate even when the human explicitly asked to keep that read inline, which created unnecessary friction without improving the other delegation safeguards.
- **Affected artifacts**: `overlay/gentle-ai/assets/owned/opencode/prompts/orchestrators/gentle-orchestrator.md`, `overlay/gentle-ai/policy/maintenance-intent.md`, `AGENTS.md`, `README.md`, `overlay/gentle-ai/maintenance.md`, `overlay/gentle-ai/README.md`
- **Verification**: manual diff review for consistency across prompt and docs; upstream assets intentionally left untouched

## 2026-06-25 | Adopted upstream v1.42.0-1-g8a67347 and classified Hermes-only drift as outside the maintained OpenCode boundary

- **Type**: `adoption`
- **Upstream scope/range**: `660917927b4821f5e540dc8fa501d6bee723222c` (`v1.40.2`) -> `8a673472a402f1949f1080f72e72d760a61d28bf` (`v1.42.0-1-g8a67347`)
- **Decision**: accepted the new upstream OpenCode baseline, added `hermes-ephemeral-delegation` to the pruned upstream-skill boundary so the maintained overlay stays OpenCode-only, relaxed the profile task-scoping audit check to tolerate upstream formatting changes in the `task` tool entry, synced the approved upstream snapshots, and re-applied the overlay.
- **Why it mattered**: upstream added the Automatic Mode Gatekeeper plus new OpenCode commands/skills, while Hermes-only additions remained outside the maintained target. The audit also had a brittle whitespace-sensitive invariant that needed to track the real behavior, not a formatting accident.
- **Affected artifacts**: `overlay/gentle-ai/policy/managed-assets.json`, `internal/overlay/audit_upstream.go`, `overlay/gentle-ai/policy/maintenance-intent.md`, `AGENTS.md`, `overlay/gentle-ai/assets/upstream/opencode/commands/sdd-continue.md`, `overlay/gentle-ai/assets/upstream/opencode/commands/sdd-status.md`, `overlay/gentle-ai/assets/upstream/opencode/commands/skill-creator.md`, `overlay/gentle-ai/assets/upstream/opencode/commands/skill-registry.md`, `overlay/gentle-ai/assets/upstream/opencode/prompts/orchestrators/gentle-orchestrator.md`, `overlay/gentle-ai/assets/upstream/opencode/sdd-overlay-single.json`, `overlay/gentle-ai/assets/upstream/opencode/sdd-overlay-multi.json`, `overlay/gentle-ai/assets/upstream/opencode/skills/_shared/sdd-status-contract.md`, `overlay/gentle-ai/assets/upstream/opencode/skills/hermes-ephemeral-delegation/`, `overlay/gentle-ai/assets/upstream/opencode/skills/skill-creator/`, `overlay/gentle-ai/assets/upstream/opencode/skills/skill-improver/`, `overlay/gentle-ai/state/upstream-state.json`
- **Verification**: `bash audit-gentle-ai-upstream.sh` (initially failed on unmapped Hermes drift and a whitespace-sensitive task-scoping check); `bash sync-gentle-ai-upstream-assets.sh`; `go test ./internal/overlay`; `bash apply-gentle-ai-custom.sh opencode`; on-disk post-state verification of runtime prompt refs, profile assignments, pruned skills, and upstream-state consistency
- **Follow-up**: `gentle-ai sync` still failed in this environment because the rollback logic refuses the symlinked `~/.gemini/antigravity-cli/mcp_config.json`; investigate that host-level config if a future runtime refresh must complete without manual intervention.

## 2026-06-17 | Wired the owned Judgment Day retrospective runtime hook

- **Type**: `tooling-change`
- **Upstream scope/range**: maintenance contract / OpenCode runtime surface, not a new upstream boundary
- **Decision**: kept the approved upstream `judgment-day` snapshot as audit-only, moved the automatic retrospective handoff into the owned runtime `judgment-day` asset under `overlay/gentle-ai/assets/owned/opencode/skills/judgment-day/`, kept `judgment-retrospective` as a portable shared skill, and updated the maintainer docs/manifest boundaries so `apply-gentle-ai-custom` installs the runtime hook explicitly. Applied a targeted `merge`/`prune` filter in `collectOwnedRuntimePaths` so only assets that declare a sync mode enter the stale-deletion surface.
- **Why it mattered**: the new learning loop only becomes real when the runtime hook is owned and installable; snapshot-only edits would not materialize at apply time, and the owned/shared split needed to be explicit for future audit/apply work.
- **Affected artifacts**: `overlay/gentle-ai/policy/managed-assets.json`, `overlay/gentle-ai/assets/owned/opencode/skills/judgment-day/SKILL.md`, `overlay/gentle-ai/assets/owned/opencode/skills/judgment-day/references/prompts-and-formats.md`, `internal/overlay/runtime_owned_state.go`, `internal/overlay/runtime_owned_state_test.go`, `shared/skills/judgment-retrospective/SKILL.md`, `shared/skills/judgment-retrospective/assets/semantic-memory-contract.md`, `AGENTS.md`, `README.md`, `overlay/gentle-ai/assets/README.md`, `overlay/gentle-ai/assets/owned/opencode/README.md`, `overlay/gentle-ai/README.md`, `overlay/gentle-ai/maintenance.md`
- **Verification**: `go test ./internal/overlay`; `git diff --check`

## 2026-06-14 | Accepted upstream v1.40.2 for OpenCode/runtime assets and ignored Claude-only drift

- **Type**: `adoption`
- **Upstream scope/range**: `03457e9e3406ee5695da6dca5cd16c1f49a50dad` (`v1.37.0`) -> `660917927b4821f5e540dc8fa501d6bee723222c` (`v1.40.2`)
- **Decision**: accepted the `v1.40.2` boundary for the OpenCode/runtime-impact surface; kept the OpenCode overlay JSONs/plugins plus the approved `internal/assets/claude/engram-protocol.md` materialization in scope; and intentionally ignored Claude agent/command/output-style drift so it would not be promoted into the managed asset tree.
- **Why it mattered**: the upstream range widened Claude-side materialization beyond the maintained OpenCode surface. Narrowing the manifest keeps audit/sync/apply aligned with the actual runtime target and preserves the approved keep/prune baseline.
- **Affected artifacts**: `overlay/gentle-ai/policy/managed-assets.json`, `overlay/gentle-ai/assets/upstream/opencode/` (including new `sdd-overlay-single.json`, `sdd-overlay-multi.json`, `plugins/`), `internal/overlay/audit_upstream.go`, `internal/overlay/managed_assets_test.go`, `overlay/gentle-ai/state/upstream-state.json`, `AGENTS.md`, `README.md`, `overlay/gentle-ai/maintenance.md`, `overlay/gentle-ai/policy/maintenance-intent.md`, `overlay/gentle-ai/assets/owned/opencode/AGENTS.md`, `overlay/gentle-ai/assets/owned/opencode/prompts/orchestrators/gentle-orchestrator.md`, `overlay/gentle-ai/logs/update-log.md`
- **Verification**: `go test ./internal/overlay`; `git diff --check`

## 2026-06-13 | Standardized maintainer report terminology and table shape

- **Type**: `policy-change`
- **Upstream scope/range**: maintenance workflow contract, not a new upstream boundary
- **Decision**: standardized maintainer audit reports around `Scope` (`Managed` / `Unmanaged`), `Impact` (`Behavioral` / `Runtime` / `Housekeeping`), and `Decision` (`Adquirir`, `Sanitizar`, `Ignorar`); defined the canonical report columns (`Upstream change`, `Files`, `Scope`, `Impact`, `Decision`, `Why`, `Follow-up`); removed `descartar` as the primary report label.
- **Why it mattered**: future maintainer runs need one readable decision contract instead of mixed adopt/discard wording, and the report needs to separate maintained-target changes from irrelevant or housekeeping-only drift without changing the existing approval gates.
- **Affected artifacts**: `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`, `AGENTS.md`, `README.md`, `overlay/gentle-ai/README.md`, `overlay/gentle-ai/maintenance.md`, `overlay/gentle-ai/logs/update-log.md`
- **Verification**: doc consistency review across the changed maintainer artifacts

## 2026-06-11 | Promoted OpenCode AGENTS.md to a managed asset

- **Type**: `tooling-change`
- **Upstream scope/range**: maintenance contract / OpenCode runtime surface, not a new upstream boundary
- **Decision**: made `~/.config/opencode/AGENTS.md` a first-class managed asset with an approved upstream-equivalent materialization in `overlay/gentle-ai/assets/upstream/opencode/AGENTS.md` (built from the upstream OpenCode persona + engram sources), a repo-owned runtime copy in `overlay/gentle-ai/assets/owned/opencode/AGENTS.md`, manifest coverage in `policy/managed-assets.json`, and apply/install handling through the normal owned-asset pipeline; removed the post-apply `injectCustomRules()` mutation from the OpenCode agent path.
- **Why it mattered**: the AGENTS instruction surface was previously mutated ad hoc after apply, which made the runtime state harder to audit and decouple from the maintained upstream materialization inputs. Treating it as a managed asset keeps audit/sync/apply behavior consistent with the rest of the overlay.
- **Affected artifacts**: `internal/overlay/opencode_agent.go`, `internal/overlay/managed_assets.go`, `internal/overlay/markdown_materialization.go`, `internal/overlay/sync_upstream_assets.go`, `internal/overlay/audit_upstream.go`, `internal/overlay/managed_assets_test.go`, `overlay/gentle-ai/policy/managed-assets.json`, `overlay/gentle-ai/assets/upstream/opencode/AGENTS.md`, `overlay/gentle-ai/assets/owned/opencode/AGENTS.md`, `overlay/gentle-ai/policy/maintenance-intent.md`, `overlay/gentle-ai/maintenance.md`, `overlay/gentle-ai/assets/README.md`, `overlay/gentle-ai/assets/upstream/opencode/README.md`, `overlay/gentle-ai/assets/owned/opencode/README.md`, `overlay/gentle-ai/owned-assets-refactor.md`, `AGENTS.md`
- **Verification**: `go test ./...`; `git diff --check`

## 2026-06-09 | Adopted upstream v1.37.0 boundary — sanitized chained-pr orchestrator binding

- **Type**: `audit`
- **Upstream scope/range**: `122b35816d3fbc1627359fe0613c6541604980bc` (`v1.36.8`) → `03457e9e3406ee5695da6dca5cd16c1f49a50dad` (`v1.37.0`)
- **Decision**: accepted the new upstream boundary v1.37.0; sanitized the 2-line addition that binds the `chained-pr` skill (`gentle-ai-chained-pr`) as a required skill match in the orchestrator chain strategy — this directly violates the `maintenance-intent.md` hard rule against chained/stacked PR governance in the repo-owned orchestrator; noted Hermes agent support as an upstream addition outside the maintained OpenCode runtime target (no overlay action required); kept `gentle-ai sync` as the upstream refresh path because no topology, preset, or materialization drift was detected for the maintained runtime target.
- **Why it mattered**: upstream v1.37.0 included a new feature (`feat(sdd): bind chained-pr skill into orchestrator chain strategy #792`) that embeds `chained-pr` skill enforcement into core orchestrator delivery planning. The owned orchestrator prompt already depurates the entire delivery-strategy/chained-PR block, so the new binding line was already excluded by the existing depuration. Accepting the boundary without adopting the content preserves the repo's depuration invariant.
- **Affected artifacts**: `overlay/gentle-ai/state/upstream-state.json`, `overlay/gentle-ai/assets/upstream/opencode/prompts/orchestrators/gentle-orchestrator.md`
- **Verification**: fresh-context consistency review — CONSISTENT (6/6 assertions passed); owned orchestrator prompt confirmed absent of `gentle-ai-chained-pr`; runtime pruning confirmed (4 skills removed); `bash apply-gentle-ai-custom.sh opencode` completed with 0 topology warnings and 0 unmanaged profiles.

## 2026-06-09 | Separated repo sync from runtime refresh recommendations

- **Type**: `policy-change`
- **Upstream scope/range**: maintenance workflow contract, not a new upstream boundary
- **Decision**: refined the maintenance contract so audit/approval, repo sync (`bash sync-gentle-ai-upstream-assets.sh`), runtime refresh (`gentle-ai sync` vs full reinstall), and overlay re-apply are distinct phases; made the runtime-refresh recommendation depend on the maintained runtime target/materialized state instead of any upstream topology expansion; and kept rejecting upstream attempts to reintroduce `chained-pr` governance into the repo-owned orchestrator behavior.
- **Why it mattered**: the previous wording still conflated approved-upstream repo sync with later runtime refresh and was too coarse about reinstall recommendations. The maintainer contract now says an approved boundary always triggers repo sync, while reinstall is reserved for adopted changes that actually affect the maintained runtime target or when sync cannot materialize the required state.
- **Affected artifacts**: `AGENTS.md`, `README.md`, `overlay/gentle-ai/README.md`, `overlay/gentle-ai/maintenance.md`, `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`, `overlay/gentle-ai/logs/update-log.md`
- **Verification**: fresh consistency review across the changed maintainer artifacts plus `git diff --check`

## 2026-06-08 | Made maintainer workflow decision-oriented and approval-gated

- **Type**: `policy-change`
- **Upstream scope/range**: maintenance workflow contract, not a new upstream boundary
- **Decision**: changed the maintainer contract so every upstream audit must produce a pre-mutation decision summary (`what is new upstream`, `Scope`, `Impact`, `Decision` with `Adquirir` / `Sanitizar` / `Ignorar`, `why`, and refresh recommendation), STOP for explicit user approval before any repo/runtime mutation, then finish with a closing summary in the same terminology plus a fresh-context consistency review.
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
