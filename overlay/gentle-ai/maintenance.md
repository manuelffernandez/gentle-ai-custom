# Maintenance

Human guide to the Gentle AI overlay.

This file describes the current operating model: upstream Gentle AI owns ODD/review runtime materialization; `apply-gentle-ai-custom` preserves upstream artifacts and installs only true local assets or surgical deltas.

Approved upstream runtime assets are adopted replacement-first: let upstream materialize the finished artifact, keep raw source snapshots for audit, and apply only a documented compatible delta from `policy/maintenance-intent.md`.

## Current v4 repository and runtime status

- The installed binary and clean upstream checkout were verified at v4.0.0 / `ff77164d4f56f1665b22fb6fac51c2ccbb769400`.
- The maintained repository snapshot boundary is v4.0.0 / `ff77164d4f56f1665b22fb6fac51c2ccbb769400`, recorded in `state/upstream-state.json`.
- The authorized snapshot sync reported 11 new, 13 changed, 31 unchanged, and 35 deleted snapshots across 19 directories. The post-sync audit passed with no managed or watched-unmanaged drift and every current structural flag `ok`.
- OpenCode and Codex runtime versions remain unverified. No runtime refresh, reinstall, or overlay apply is implied.
- The 11 new OpenCode orchestrator/agent assets are upstream-only snapshots, not custom runtime sources. Of the 35 deleted snapshots, 18 were exact SDD paths declared by `retired_upstream_snapshots`; the other 17 were stale files pruned inside still-declared synced directories (11 commands, 2 plugins, and 4 `_shared` files). Manifest removals target exact files only and reject traversal, overlap with active paths, symlink parents, and directories; directory pruning is bounded to declared synced assets. Neither mechanism changes runtime files.
- Current OpenCode keep/prune policy no longer retains SDD skills and prunes legacy SDD skill names if found. This is repository policy only: no overlay apply ran, and machine-local SDD settings or runtime files were not removed.

## What this file does not define

- agent/runtime behavior -> overlay-owned runtime assets first (especially `overlay/gentle-ai/assets/owned/opencode/AGENTS.md`), then maintainer workflow specifics in `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`
- credential-file safety -> `overlay/gentle-ai/assets/owned/opencode/AGENTS.md`; it governs the primary agent and every delegated subagent or tool, while technical permission barriers remain outside this policy
- Markdown style for agent-consumed docs -> overlay-owned runtime standard in `overlay/gentle-ai/assets/owned/opencode/AGENTS.md` (installed to `~/.config/opencode/AGENTS.md`); repo-local additions live in `AGENTS.md` and only supplement it
- repo ownership and policy -> `AGENTS.md`
- keep/prune intent, upstream-first boundaries, and explicit local customizations -> `policy/maintenance-intent.md`
- machine-readable policy -> `policy/gentle-ai-policy.json`
- maintained upstream boundary -> `state/upstream-state.json`
- closed-event ledger -> `logs/update-log.md`

## Quick path

1. Update the `gentle-ai` binary.
2. Update tags in the resolved upstream clone and manually check out the exact `v<V>` release to inspect.
3. From `gentle-ai-custom`, load the maintainer skill/minimal invocation context and immediately run `bash audit-gentle-ai-upstream.sh` synchronously. By default it checks the installed binary's release; to inspect another release while preserving that binary, pass `--release v<V>`. Do not broadly explore the repo or delegate repository reading/drift analysis first. If selected-release verification fails, STOP and delegate nothing.
4. Convert the audit into a concise decision summary before any mutation:
   - what changed upstream and why it matters
   - `Scope`: `Managed` / `Unmanaged`
   - `Impact`: `Behavioral` / `Runtime` / `Housekeeping`
   - `Decision`: `Adquirir` / `Sanitizar` / `Ignorar`
   - `Why`
   - `Follow-up` (optional)
   - repo sync requirement: whether approval of a new boundary requires `bash sync-gentle-ai-upstream-assets.sh`
   - recommended runtime path: `gentle-ai sync` vs full reinstall
5. STOP for explicit approval before updating this repo, advancing `state/upstream-state.json`, or running `bash sync-gentle-ai-upstream-assets.sh`.
6. If a new upstream boundary and repo sync were separately approved, run `bash sync-gentle-ai-upstream-assets.sh` to refresh `overlay/gentle-ai/assets/upstream/...`.
7. Recommend the upstream refresh path and print exact instructions:
   - `gentle-ai sync` if the maintained runtime target stays effectively compatible
   - full reinstall if adopted changes affect topology, presets, or materialization for the maintained runtime target, or if sync no longer materializes the right state
   Runtime refresh is human-owned by default. Do not execute either path unless the human explicitly requests that specific runtime action.
8. Recommend re-applying the overlay with `bash apply-gentle-ai-custom.sh opencode`. Execute it only with separate explicit authorization; runtime-refresh approval is not apply approval.
9. Read `Summary:`, verify the final on-disk state, run one fresh-context consistency review, and return a closing summary of what was actually `Adquirir`, `Sanitizar`, or `Ignorar`, and why.
10. If `~/.config/opencode/opencode.json` changed, restart OpenCode.

No `gentle-ai sync` or reinstall yet is the normal state before step 3. Asking what was updated helps diagnosis but proves nothing; the audit activation gate always runs first.

## Release parity preflight

Both canonical audit and sync entrypoints default to the installed binary release. `--release v<V>` explicitly selects another release without changing the binary or runtime. Either mode stops before drift analysis, snapshot copy, or state writes unless all conditions pass:

- The selected version is `V` (the installed version by default, or the explicit selector).
- The upstream checkout has no staged, unstaged, untracked, or submodule changes.
- The selected exact `refs/tags/v<V>` exists and peels to a commit.
- Upstream `HEAD` equals that peeled commit.

This accepts a prerelease only at its exact tag. The explicit selector requires the exact `v<V>` syntax; same-named branches are not tags. It does not use `git describe`, nearest tags, version ordering, or branch ancestry. It never fetches tags or changes the checkout automatically.

An audit with `--release` is preview evidence only. It reports installed binary and selected audit release separately and does not approve adoption or repository snapshot sync. If adoption and repo sync are later approved, pass the same selector to `sync-gentle-ai-upstream-assets`; sync remains a separate repository mutation. Runtime refresh and overlay apply remain separately authorized.

Audit and repo sync consume only `upstream_repo_path` while resolving the upstream clone. They intentionally do not validate runtime profile assignments at this stage, because a new phase topology must be auditable before the machine-local profile schema is migrated. Full local runtime config validation remains an apply precondition.

For the maintainer flow, this is an activation hard gate. Run it immediately after the minimal skill context and before broad repository exploration or repository-reading/drift-analysis delegation. On failure, stop immediately and delegate nothing. `last_maintained_commit` defines the later drift range; it is not parity proof.

If the selected tag is missing or stale, run `git -C <upstream-repo> fetch --tags`, verify `v<V>`, and check it out manually. If a pulled branch is ahead of the selected release, check out that exact tag before retrying. The default mode blocks when binary and clone differ; explicit `--release v<V>` allows work on another clean exact release without upgrading the binary. The default still verifies the installed release when neither changed.

## Receipt-Driven Development policy and effective mode

Receipt-Driven Development (RDD) is maintainer-approved as active and functional. This is an explicit owner decision based on operational validation in other repositories; it is not a claim that a specific v2.9 canary or receipt was observed.

Repository policy does not set the effective mode, and upstream does not remotely control it. The setting remains user- and machine-owned. On this machine, the observed state is global `on` with clone-local unset, but maintainers must query current state before work rather than rely on this observation.

Check the effective setting before work. The setting is machine-specific and is not version-controlled state in this repository:

```bash
gentle-ai review mode status --cwd .
gentle-ai review mode disable --cwd .
gentle-ai review mode enable --cwd .
```

`status` is read-only. `disable` and `enable` use global scope by default. Add `--scope clone --cwd <clone>` to disable RDD only in one clone or to clear that clone-local disable. Any `off` source wins: clone scope can opt one clone out, but it cannot require RDD while global scope is off, and no other clone inherits that override.

Enabling RDD is operationally simple, but it governs new candidates from that point forward. It does not automatically approve, repair, or resume an old invalid or pending authority. Content created while RDD was off is not stranded: after re-enabling, the current content can become a new candidate and undergo a new review. When RDD is disabled, new review work and reviewers must not start. Delivery without an exact governing receipt reports `disabled/unmanaged`, follows ordinary repository policy, and must not fabricate a receipt. An existing receipt that exactly governs unchanged content remains authoritative.

Activation does not weaken the review contract. Receipts must remain bound to the exact candidate and lineage. Correction work remains bounded by the admitted findings and authorized correction scope. The user-owned kill switch remains authoritative, and ordinary repository policy still decides delivery.

### RDD and planning workflows

- In general, Spec-Driven Development (SDD) describes a planning and implementation lifecycle. It is not the current v4 workflow maintained by this repository, which follows ODD without SDD.
- Receipt-Driven Development (RDD) is a cross-cutting candidate-review and delivery-receipt lifecycle. When enabled, it can review the immutable post-apply candidate and reuse that receipt at delivery gates.
- RDD does not mean Test-Driven Development (TDD). Tests remain part of implementation and verification regardless of the RDD switch.
- RDD is separate from any planning lifecycle. The v4 workflow maintained here is ODD without SDD; disabling RDD does not change that workflow or its ordinary tests and verification.

### Historical correction-handoff context

The official stable v2.2.2 binary at commit `09c2d8b6369860058aa44704df76b1b33c745675` is affected. The bug remained unfixed at upstream commit `d260cdbc8c4be083333cb8b0517ed125a450a411`.

The v2.5.0 source and tests appear to close the missing pre-edit handoff at the code-contract level: negotiated STATUS now issues a bound correction-plan request, and `review capture-correction-plan` records the positive bounded forecast before candidate mutation. The request is derived from admitted corroborated findings and exact `FixFindingIDs`.

This source and test evidence resolved the earlier code-contract concern for policy purposes when combined with the owner's later cross-repository operational validation. No specific v2.9 canary or receipt is claimed here.

Previously confirmed affected lifecycle (v2.2.2 through the pre-v2.5 evidence boundary):

1. All selected reviewers are admitted.
2. Finalization enters `correction_required`.
3. A positive correction forecast is recorded while the candidate is unchanged.
4. Negotiated STATUS stops with `corrected_candidate_unavailable` but exposes no confirmed finding details or authoritative correction scope.

In that affected lifecycle, the native authority has the missing information: it stores the full findings and `FixFindingIDs`. The OpenCode result-capture plugin replaces each successful reviewer output in the session with an opaque capture manifest. Before candidate mutation, negotiated STATUS does not return the corroborated findings or another authoritative correction handoff. A `TargetedValidationRequest` containing the finding IDs is built only after the candidate changes. The consumer is therefore told to correct the candidate without first receiving a trustworthy statement of what to correct.

Source evidence:

- `internal/reviewtransaction/compact.go` stores canonical `Findings` and `FixFindingIDs`.
- `internal/reviewtransaction/targeted_validation_request.go` requires a changed correction candidate before it emits `FixFindingIDs` in `TargetedValidationRequest`.
- `internal/cli/review_status_contract.go` defines the negotiated STATUS surface but exposes no pre-mutation correction scope.
- `internal/cli/review_next_transition.go` emits `corrected_candidate_unavailable` after a forecast when no corrected candidate exists.
- `internal/assets/opencode/plugins/review-result-artifacts.ts` replaces successful reviewer output with the native capture manifest.

v2.5.0 correction evidence:

- `internal/reviewtransaction/correction_plan_request.go` derives the correction request from the admitted compact review view and exact corroborated finding IDs.
- `internal/cli/review_correction_plan_capture.go` binds and records the positive correction forecast before editing.
- `internal/cli/review_correction_plan_request_test.go` and correction-plan capture tests exercise the pre-edit handoff and binding failures.

A temporary overlay patch was assessed and deferred. No correction-handoff patch is implemented by this repository.

### Optional regression diagnostic

The former correction-required canary is retained only as an optional regression diagnostic. It may compare the source paths above, confirm that authoritative corroborated correction scope is available before mutation, and exercise correction through targeted validation and an exact valid receipt. It is not a prerequisite for activation or a reason to override the owner's accepted policy.

If this diagnostic exposes a regression, record the evidence and reassess the affected release. Do not weaken candidate identity, receipt lineage, correction bounds, or delivery rules as a workaround. Always confirm effective mode with `gentle-ai review mode status --cwd .`.

## Operating model

| Artifact | Role |
| --- | --- |
| `policy/maintenance-intent.md` | Human intent: upstream-first baseline plus the small local layer to keep |
| `policy/gentle-ai-policy.json` | Runtime policy consumed by the Go CLI and wrappers |
| `policy/managed-assets.json` | Canonical map of approved upstream assets and installable owned assets |
| `assets/upstream/` | Approved upstream copies currently on disk for review/diff, including `persona-gentleman.md`, `engram-protocol.md` (mirrors `internal/assets/engram/protocol.md`), `background-subagents.md`, the orchestrator prompt, overlay JSONs/plugins, retained `commands/` snapshots, and selected skill snapshots. The maintained boundary is v4.0.0 at `ff77164d4f56f1665b22fb6fac51c2ccbb769400`. The authorized sync reported 11 new, 13 changed, 31 unchanged, and 35 deleted snapshots across 19 directories; the post-sync audit passed with no drift and all structural flags `ok`. This verifies repository snapshots only, not either materialized runtime. |
| `assets/owned/` | True repo-owned runtime assets installed by `apply`, currently the full-file local OpenCode `AGENTS.md` surface. Its inherited persona and OpenCode full Engram sections track the approved snapshots, with six explicit overlay-only additions. |
| `patches/` | Exact-anchor local deltas applied only to compatible upstream runtime files |
| `shared/skills/` | Portable repo-owned skills installed for OpenCode and Codex, including the sanitized `branch-pr` and `chained-pr`, the temporary `pr-finalizer` compatibility skill, and the Judgment Day retrospective companion skill |
| `shared/commands/` | Source bodies for OpenCode wrappers; Codex does not receive command wrappers |
| `state/upstream-state.json` | Last maintained upstream boundary |

The retained skill snapshots `gentle-ai-bench`, `rdd-defect-workflow`, and `systemic-issue-triage` are audit/sync inputs only. They are not repo-owned install sources and do not force runtime selection or RDD activation.

Package-install guidance lives in the canonical `shared/skills/package-security/SKILL.md` and its local templates. It forbids credential-file inspection, requires human review for script approvals, and distinguishes npm's advisory-by-default policy from pnpm's version-gated build controls. Check the effective CLI's support for each feature. Editing the source does not update installed skills; separately authorize `apply-gentle-ai-custom.sh opencode` and/or `codex` to propagate them.

## What each command does

### `audit-gentle-ai-upstream`

- verifies and reports the installed binary version separately from the selected exact audit release and commit before analysis
- accepts `--release v<V>` for a clean exact upstream release newer or older than the installed binary; default behavior remains installed-binary parity
- marks explicit-release output as preview only; it does not adopt the release or authorize snapshot sync
- uses `last_maintained_commit` from `state/upstream-state.json`
- discovers drift with `git diff --name-status --find-renames <last_maintained_commit>..<verified-release-commit>`
- filters that drift through `policy/managed-assets.json`
- verifies the current v4 source and injection topology in `agentguidance/orchestrator.go`, `agentguidance/inject.go`, `agentguidance/remote_authorization.go`, and `internal/opencode/models.go`
- verifies that OpenCode uses the canonical `opencode/orchestrator.md` asset and sends ODD routing into its always-loaded carrier; unbound or decoy source calls fail
- verifies the remote-authorization helper and its projection into that routing carrier; if v4 adds or changes a projection that cannot be proven, sync fails closed
- keeps the upstream remote-authorization snapshot audit-only and verifies projection into the current OpenCode routing prompt carrier
- validates the current ODD explore/verify/worker agent set and exact JD/review actor topology
- does not require removed SDD profile-generator paths or SDD profile-order checks

### Decision handoff before mutation

Before the maintainer edits repo files or refreshes runtime, turn the audit into an approval gate. Use these report columns:

| Upstream change | Files | Scope | Impact | Decision | Why | Follow-up |
| --- | --- | --- | --- | --- | --- | --- |

- `Upstream change`: concise human summary of the upstream delta and why it matters; do not use a file-path list as the main content.
- `Scope`: `Managed` if the change is inside the overlay/runtime surface this repo maintains; `Unmanaged` if it is outside that surface or not relevant.
- `Impact`: `Behavioral` for agent/orchestrator behavior; `Runtime` for maintained-target wiring, install, config, or materialization; `Housekeeping` for docs, unrelated agents, or internal fixes with no maintained-target effect.
- `Decision`: `Adquirir`, `Sanitizar`, or `Ignorar`.
- `Follow-up`: optional; leave empty when nothing else is needed.
- One row represents one atomic decision, and `Decision` applies exactly to its `Upstream change`.
- Put distinct local remediation in `Follow-up` or a separate row. Runtime sanitization always gets a separate row.
- Every required adjustment named in the summary must map to an explicit row.
- Never merge rows when `Scope`, `Impact`, or `Decision` differ.
- Reconcile the summary, rows, and final recommendation before presenting them.
- Do not use `descartar` as the main report label.

No repo mutation happens before that handoff is approved.

### Runtime refresh recommendation

- Prefer `gentle-ai sync` when the adopted change matters to the runtime target this repo actually maintains, but that target's topology, presets, and materialized state remain effectively compatible.
- Do not recommend reinstall only because upstream added support for some other agent or platform outside the maintained runtime target/materialized state.
- Recommend a full reinstall when the adopted change affects topology, presets, or materialization for the maintained runtime target, or when `gentle-ai sync` cannot materialize the required state.
- Historical v2.9.1 evidence: the separately authorized repository sync advanced the maintained boundary from `be49554794917ae92a6dc9dbfa2eb3db5cf70084` (`v2.9.0`) to `9ec0cf443622f20fe511815ff5a83088c7467bff` (`v2.9.1`). It reported 0 managed file changes, 74 unchanged files, 0 deletions, and 18 synced directories.
- The post-sync canonical audit passed at exact v2.9.1 / `9ec0cf443622f20fe511815ff5a83088c7467bff`, with no managed or watched-unmanaged drift and all profile, injection, remote-authorization, and JD/review topology invariants `ok`.
- Acquire the compatible OpenCode review-consent restoration and preserved-prompt migration upstream-first. Ignore Claude Code transcript-derived dispatch authority because it is outside this repository's maintained OpenCode target. No keep/prune, runtime ownership, local patch, topology, profile, preset, or materialization changes are required.
- After the repository-only closure, the user reports manually running both `gentle-ai sync` and `bash apply-gentle-ai-custom.sh opencode`. Their original command summaries and exit output were not captured, so do not infer or publish new apply counts.
- Independent read-only post-state verification passed 602 selected assertions over runtime assignments, named profiles, prompt parity, task scoping, skill permissions, absence of suffixed review actors, and obsolete prompt references. The installed orchestrator contains the compatible `gentle-ai.review-integration.consent/v3` semantics: `multiple: false`, `custom: false`, and missing custom consent alone is not a compatibility block.
- All 7 repo-owned skill sources/assets, all 5 wrapper bodies, and the owned `AGENTS.md` matched their installed targets. `issue-creation` and `hermes-ephemeral-delegation` were absent, `branch-pr` was the repo-owned replacement, and all 3 installed runtime patch targets matched approved upstream snapshots after reversing only their declared local patches.
- Historical v2.9.0 repository evidence remains valid: its authorized sync advanced the maintained boundary from v2.5.0 and refreshed the approved snapshots for the v2.9 OpenCode composition model.
- Historical v2.5.0 runtime evidence remains valid: the user manually completed the full SDD component installation, and the later, separately authorized `bash apply-gentle-ai-custom.sh opencode` completed successfully and updated OpenCode config.
- Apply reported: skills pruned 2; owned assets changed 1 and verified 1; topology warnings 0; profiles managed 4; profile agents created 3, updated 4, unchanged 49; unmanaged profiles 0.
- Final runtime verification passed the canonical audit, `go test ./...`, `git diff --check`, and resolved `sdd-research` agent loading. Sanitized repo-owned `branch-pr` is present; conflicting upstream `branch-pr`, `issue-creation`, and Hermes are absent. All 7 repo-owned skills, 5 wrappers, owned `AGENTS.md`, overrides, and 4 managed profiles match canonical or local config. `general` is `openai/gpt-5.6-terra` / `high`; `explore` is `openai/gpt-5.6-sol` / `medium`.
- At that historical v2.5.0 verification point, all 4 profiles had the ordered 11-phase topology and the 3 named `sdd-research` actors. Review actors remained global with no suffixed review actors, and Judgment Day differed only by the retrospective patch. Managed background intent and effective policy were `on`; the launcher was `~/.gentle-ai/bin/opencode`, and Desktop/plain OpenCode processes remained foreground.
- Historical non-blocking discrepancy from the v2.5.0 runtime verification: `~/.gentle-ai/state.json` reported `installed_binary_version` as `2.4.0`. Do not assert a root cause or edit external state without separate authorization.
- RDD is maintainer-approved as active and functional by owner decision based on cross-repository operational validation. Effective mode remains machine-owned; this machine was observed global `on` with clone-local unset. The user explicitly excluded this v2.9.1 maintenance candidate from RDD, so no v2.9.1 receipt is claimed.
- Any future `gentle-ai sync` or reinstall restores upstream runtime state and therefore requires a new overlay apply with separate explicit authorization.
- If upstream changes commit/PR governance, compare it against the explicit local customizations first; only sanitize when there is a concrete conflict.

The maintainer may audit, classify, recommend, and print exact runtime-refresh instructions. It must not run `gentle-ai sync` or a reinstall unless the human explicitly requests that specific action. Approval of findings, adoption, repo snapshot sync, or overlay edits is not authorization for runtime refresh.

Runtime refresh, repo snapshot sync, and overlay apply are separate actions. Approval for one never implies approval for another.

`Runtime` includes changes to the maintained target's wiring, install path, config, or materialized files even when the agent behavior itself does not change.

`Housekeeping` covers changes that do not matter for the maintained target, such as irrelevant docs, unrelated agents, or internal fixes with no maintained-target effect.

### `sync-gentle-ai-upstream-assets`

- repeats the same exact release preflight before copy or write; accepts `--release v<V>` for a separately approved release, while default behavior remains installed-binary parity
- copies approved upstream assets into `assets/upstream/...`
- after exact release parity, watched-drift mapping, and all structural gates pass, removes exact files named by `retired_upstream_snapshots`; this manifest lists 18 retired SDD paths and rejects traversal, overlap with active snapshots, symlink parents, and directories
- also prunes stale files from still-declared synced directory assets (the 17 other deletions in the v4 sync); this pruning stays inside those declared asset directories and is separate from manifest retirement
- never removes retired snapshots during audit or repository-only adaptation; this cleanup runs only in an explicitly authorized snapshot sync
- after a snapshot refresh, requires the owned `AGENTS.md` persona and OpenCode full Engram sections to be rebased when the focused semantic parity test reports drift; do not replace full-file ownership with runtime patching or marker merging
- stores the selected release version, exact tag, and release commit in `state/upstream-state.json` when the new upstream was accepted
- does not touch local runtime under `~/.config/opencode/`
- always runs after approval of a newly accepted upstream boundary; it is the repo sync step, not the runtime refresh step

### `apply-gentle-ai-custom`

- keeps its existing installed-binary versus maintained-overlay-baseline preflight and does not require an upstream clone
- validates only custom-operation prerequisites before any write: ownership boundaries, safe obsolete-reference migration, declared runtime-patch anchors, and exact recognized forms of the legacy prompt section being removed
- preflights reachable maintained inline prompts, `{file:...}` references, and existing legacy shared SDD prompt files still declared among owned overlay assets for recognized upstream or previously relaxed remote-authorization sections before the first write. Missing references, duplicate, malformed, or unknown sections, out-of-root/symlink paths, and remote-bearing unmanaged named profiles fail closed; absent optional shared prompt files are skipped.
- removes only the complete known section. Absence is valid and idempotent. Unrelated content and file modes are preserved, and resulting file bytes/modes plus persisted prompt values are verified. A later OpenCode JSON remarshal may change formatting; prompt contents and unrelated parsed JSON values remain protected.
- does not inspect or certify upstream prompt, contract, plugin, or review semantics; a missing or opaque plugin does not block unrelated custom operations
- preserves upstream-owned prompt, contract, and plugin bytes and file modes
- migrates managed obsolete `.overlay.md` references only when an upstream base prompt is available
- prunes only explicitly rejected upstream skills in the selected registered targets
- reports only skills actually removed during pruning; already-absent skills produce no no-op line. The SDD prune and effective permission-deny policy remains protective on later applies even when no SDD skill is present.
- denies policy-pruned skills that lack a repo-owned same-name replacement through global and configured-agent OpenCode `permission.skill` rules. This covers native external-skill discovery from `~/.agents/skills` and `~/.claude/skills` without deleting or changing those externally owned copies; non-pruned rules and replacements such as `branch-pr` are preserved. Because OpenCode uses the last matching rule, apply canonically persists effective exact denies after matching wildcard rules even when parsed values already equal `deny`
- applies `agent_overrides`
- reconciles explicit global Judgment Day/review assignments from `global_actors`
- installs repo-owned skills from `shared/skills/`
- installs the sanitized repo-owned `chained-pr` copy while keeping its upstream snapshot retained for audit/sync, and keeps `issue-creation` pruned from the default OpenCode runtime surface
- does not create SDD agents or profiles. It rejects legacy `default_profile`, `profiles`, SDD `agent_overrides`, and SDD agent keys before writes; user-owned values remain untouched for manual migration
- keeps all six review actors global, including `review-validator`; no suffixed review actors are created
- treats omission as preserve/delegate and `{"inherit": true}` as an explicit request to remove both `model` and `variant`
- starts from the upstream standalone Judgment Day protocol, patches only the declared retrospective handoff, and rejects incompatible anchors
- removes the known upstream or previously relaxed remote-authorization section across maintained inline agents, referenced prompt files, and any existing legacy shared SDD prompt files still declared among owned overlay assets; unmanaged named profiles containing it block apply
- keeps the v3.7.0 `sdd-research` contract as historical snapshot evidence; it is not a current v4 runtime verification
- records every custom-installed runtime output, including single-file assets without a sync mode
- preserves stale-state paths returned to upstream, retained upstream skill trees, and surgical patch targets without recording those files as wholly custom-owned
- renders custom wrappers from `shared/commands/` for OpenCode

### Codex target

`apply-gentle-ai-custom.sh codex` installs the eight allowlisted skills from the canonical `shared/skills/` directory into `~/.codex/skills/`; allowlisted same-name skills are repo-owned replacements. It does not run OpenCode policy/config reconciliation, prune unrelated Codex skills, render command wrappers, or modify `config.toml`, Engram, SDD model profiles, or the full Codex `AGENTS.md`.

- Codex's full `AGENTS.md` remains upstream-owned. Apply inserts the marker-bounded local instruction rules after `<!-- /gentle-ai:engram-protocol -->` and replaces exactly the `odd-explorer` and `odd-worker` table rows from `gpt-6.1-luna` / `high` to `gpt-6.1-sol` / `high` in the same declared target; it does not require the removed SDD marker. Other ODD, JD, and review rows remain unchanged. Actual account model availability is unverified.
- The separate Notion Project Knowledge activation follows the local-instructions closing marker through a second sequential replacement. Replacement zero stays unchanged so recognized legacy-block migration remains compatible. The canonical `shared/skills/notion-project-knowledge/SKILL.md` follows current Notion guides for retrieval and parent-agent recording of useful project knowledge and personal activity evidence. No Notion structure or procedure is copied into the overlay. Repository identity lookup and explicitly requested project creation follow the current global Notion index; detailed metadata and structure rules remain in Notion.
- The local rules cover implementation scope, plain-language responses, agent-readable Markdown, credential-file safety, and the conditional Gemini response rule. They do not override upstream ODD task-close commits or add the retired remote-operation authorization rule.
- Codex patch preflight recognizes an exact stale second local block containing the retired no-commit line only between the orchestrator and CodeGraph markers. It removes that block in memory before preparing the one `AGENTS.md` patch; an absent legacy block and an applied or unapplied current Engram-adjacent block are valid. Duplicate, incomplete, changed, or relocated local blocks fail before any selected-target write, including `all`. No full-file replacement or generic empty replacement is used.
- Judgment Day uses an exact Codex patch definition; the ODD work-unit commit is no longer patched. Declared anchors are preflighted before the first selected-target write; missing or changed anchors fail closed.
- Unrelated upstream instructions, untouched skills, and existing modes are preserved. Only copied repo-owned skill outputs are tracked as owned. Patched upstream targets are protected during owned-state cleanup but are never claimed as whole-file owned assets.
- The canonical upstream audit currently covers OpenCode managed assets only. It does not verify the installed Codex runtime. Codex behavior must remain reported as unverified until the user separately authorizes a Codex apply and checks its result.

Native skill-registry refresh does not scan canonical `shared/skills/`; it cannot prove discovery of the new Notion skill before a separately authorized runtime apply. Repository tests install into temporary homes only, not the machine runtime.

`apply` never copies raw upstream source snapshots into runtime. A missing or incompatible prerequisite for a requested custom operation fails closed before pruning, config changes, patch writes, or owned asset installation.

### Historical v3.7.0 adoption and runtime boundary

The repo snapshot adoption is closed at `6dee8f833aec9e46015759c5065a9035795d9af1`. The authorized `bash sync-gentle-ai-upstream-assets.sh --release v3.7.0` reported 36 managed changes (5 new, 31 changed, 43 unchanged, 0 deleted), 19 synced directories, and an `upstream-state.json` write. The canonical post-sync audit passed with no managed or watched-unmanaged drift and all structural checks `ok`.

- All five `internal/assets/opencode/plugins-v2/` files are mapped as upstream-only source snapshots for audit/sync. They are not custom apply install sources.
- The canonical audit verified the installed Gentle AI binary as v3.7.0 for the OpenCode managed baseline. It does not verify the current materialized runtime or the Codex runtime; their versions are unverified. OpenCode 1.18.32 was the version in earlier historical checks and used the V1 plugin path. This repository does not claim V2 runtime support.
- The v3.7.0 `sdd-research` source skill and OpenCode command enforce output-only collection with orchestrator-owned persistence. Its local precedence patch is retired; do not infer that the materialized runtime has been refreshed or verified against this source contract.
- Upstream ODD calls for a work-unit commit at each task close. The local OpenCode and Codex opt-out patches and OpenCode no-auto-commit instruction have been removed from repo-owned runtime sources. This repository-only change does not verify the materialized runtime; runtime refresh and apply remain separately authorized. Task-close commits do not authorize pushes or PRs, which retain their explicit user-authorization gates.
- The Judgment Day retrospective patch remains because the reviewed v3.7.0 source does not replace that local terminal handoff.

### Historical partial SDD archive decision

The repo consciously accepted v3.7.0's explicit partial-archive behavior as local policy: an explicit archive request could sync a syntactically applicable delta to canonical specs while tasks remained unchecked, even though the normal next recommendation could still be `apply`. This does not establish upstream intent or disprove a bug. This is historical only; v4's maintained target is ODD without SDD, so the old watchpoint is not a current audit requirement.

- These historical semantics do not establish current v4 runtime behavior or authorize removal of user-owned local SDD configuration.

## Update types and impact

| Update path | What changes | Overlay impact |
| --- | --- | --- |
| `brew upgrade gentle-ai` | Only the binary | Usually does not reset the overlay |
| `gentle-ai sync` | Prompts, skills, MCP configs, and upstream runtime assets | Restores upstream runtime state; re-applying the overlay is mandatory |
| TUI reinstall | Full installation, topology, presets, and config | Resets everything and may change agents/presets |

## Highest-signal indicators

| Signal | Meaning | Action |
| --- | --- | --- |
| `base prompt drift: yes` | Upstream `gentle-orchestrator` changed relative to the approved upstream asset | Read `Drift summary:` first |
| A current v4 structural invariant reports `mismatch` | The maintained ODD/routing/authorization/actor source shape changed | Stop and inspect the current source path before recommending `sync` |
| `remote authorization projection invariant: mismatch` | The canonical contract or an OpenCode projection carrier changed structurally | Stop and inspect the contract/helper/projection chain before adoption |
| `JD/review actor topology invariant: mismatch` | Upstream changed the exact JD or global review actor set | Stop and align schema/topology before applying |
| `topology: unknown orchestrator matched by prefix only` | A new upstream orchestrator appeared | Audit it and decide whether policy must include it |
| `topology: expected orchestrator missing from opencode.json` | A known orchestrator disappeared or was renamed | Audit upstream and update policy/intent if needed |
| `owned asset writes - ...` | `apply` installed or left repo-owned assets untouched | Review `--verbose` output if you need file-level detail |

## Post-state verification

After `apply`, confirm this:

- explicitly pruned skills no longer exist in each selected registered target
- every effective external-skill deny resolves to `deny` in global OpenCode skill permissions and each configured agent; repo-owned same-name replacements such as `branch-pr` receive no overlay deny, while unrelated permissions remain unchanged
- persisted OpenCode config matches canonical effective-deny order, so an exact deny cannot remain before a matching wildcard allow
- each effective `agent_override` resolves to the expected `model` / `variant`
- `agent.gentle-orchestrator.prompt` remains the upstream-owned prompt or reference
- `~/.config/opencode/AGENTS.md` matches the repo-owned asset from `assets/owned/opencode/AGENTS.md`
- the installed `AGENTS.md` contains the repo-owned credential-file safety section without replacing it with filename-only rules or technical permission claims
- the repo-owned and installed `AGENTS.md` omit the remote-operation authorization rule
- maintained inline prompts and reachable prompt files contain no recognized upstream or previously relaxed remote-authorization section; unrelated bytes and file modes are preserved
- the upstream remote-authorization snapshot and structural projection audit remain unchanged
- the upstream OpenCode snapshot inputs under `assets/upstream/opencode/`, including `background-subagents.md`, match the maintained boundary recorded in `state/upstream-state.json`
- no SDD agents or profiles were generated by custom apply
- representative upstream-owned prompt, plugin, and contract files are byte- and mode-unchanged after apply
- no managed agent references `.overlay.md`
- runtime files/directories declared in `policy/managed-assets.json` exist on disk
- configured `global_actors` assignments match; inherited actors have neither `model` nor `variant`
- if legacy SDD config or keys exist, apply fails before writes and leaves them untouched
- the retired `sdd-research` precedence patch is absent from the manifest; verify upstream output-only behavior after a separately authorized runtime refresh

For a separately authorized Codex apply, also verify:

- each installed portable Codex skill under `~/.codex/skills/` matches its canonical source in `shared/skills/`
- the upstream Engram block and all unrelated bytes in `~/.codex/AGENTS.md` remain unchanged outside the declared local instruction block and two ODD table-row replacements
- `odd-explorer` and `odd-worker` are each `gpt-6.1-sol` / `high` exactly once, while `odd-verify`, JD, review, and other rows retain their upstream assignments; no model profile or OpenCode assignment changes
- the local block appears exactly once after the upstream Engram marker and contains no remote-operation authorization rule
- the obsolete second local block and its no-commit line are absent from the orchestrator-to-CodeGraph gap; unrelated bytes and file mode remain unchanged
- Judgment Day differs from its pre-apply bytes only by the declared Codex patch delta; the upstream work-unit skill remains unchanged and file modes are preserved
- upstream Codex SDD skills, untouched skills, Engram files, `config.toml`, and model profile files remain byte- and mode-unchanged
- Codex apply neither pruned skills nor rendered command wrappers, and its owned-state record claims only repo-owned skill outputs
- this check verifies only the files inspected for that apply; the OpenCode upstream audit does not verify Codex runtime materialization
- one fresh-context reviewer/subagent pass checked the changed maintainer artifacts and final summary for workflow consistency

## Local overlay config

The canonical per-machine config lives outside the repo at `~/.config/gentle-ai-custom/opencode-local-config.json`.

Operational rules:

- `upstream_repo_path` takes precedence over `GENTLE_AI_CUSTOM_UPSTREAM_REPO`, and both take precedence over `../gentle-ai`
- `opencode_config_path` is optional; when omitted, the default remains `~/.config/opencode/opencode.json`
- `agent_overrides` manages only explicit built-in agent assignments such as `general` or `explore`
- `global_actors` accepts only the three global Judgment Day and six review actors, including `review-validator`
- global actor values are `{ "model": "<provider/model>", "variant": "<optional>" }` or `{ "inherit": true }`; omission preserves runtime state
- `default_profile`, `profiles`, SDD-prefixed `agent_overrides`, and SDD agent keys are migration blockers; remove/reassign them manually because apply never deletes or maps them to ODD roles

## Maintenance checklist

- [ ] `maintenance-intent.md` still reflects what to keep and depure
- [ ] `managed-assets.json` still aligns with `assets/upstream/` and `assets/owned/`
- [ ] the owned OpenCode `AGENTS.md` inherited persona and full Engram sections pass semantic parity against the approved snapshots, while the six overlay-only sections remain present
- [ ] `upstream-state.json` still points to the last upstream boundary that was actually maintained
- [ ] current v4 policy keeps SDD names out of the keep list and records them as pruned, without changing machine-local runtime state
- [ ] `apply-gentle-ai-custom` still preserves upstream prompt refs and fails closed when a requested custom operation lacks its prerequisite
- [ ] the public shell and PowerShell entrypoints remain equivalent
- [ ] Legacy SDD profile assignments remain user-owned and did not leak into versioned policy
- [ ] upstream resolution still respects: local config -> env -> fallback `../gentle-ai`

## References

- `README.md`
- `AGENTS.md`
- `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`
- `policy/maintenance-intent.md`
- `policy/gentle-ai-policy.json`
- `policy/managed-assets.json`
- `state/upstream-state.json`
- `logs/update-log.md`
