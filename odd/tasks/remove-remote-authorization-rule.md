# Remove the Remote Authorization Rule from Installed Prompts

## Objective

Remove the upstream remote-operation authorization instruction from OpenCode prompts installed or maintained by this overlay, without editing upstream or its audit-only snapshots. Retire the local relaxation patch because its target rule should no longer be installed.

## Scope and constraints

- Authorized repository edits only; do not apply the overlay, sync, edit the upstream checkout or installed runtime, commit, push, or create a PR.
- Keep upstream source snapshots and upstream projection audit intact.
- Preserve unrelated prompts, unmanaged agents, file modes, and fail-closed preflight behavior.
- Route: delegated direct writer. Mapping involved Go apply/preflight, manifest, owned instructions, tests, and maintenance docs (more than four files); writing involves multiple non-trivial files.
- TDD: unknown (no explicit project/session TDD setting established); use ordinary functional checks. Runner: `go test ./internal/overlay/...` then `go test ./...`.
- Delivery strategy: ask-on-risk; no delivery action is authorized.

## Tasks

- [x] RA-1: Replace the relaxation patch with an idempotent, strictly scoped removal of the upstream and previously relaxed sections across maintained prompt carriers. Remove the obsolete patch declaration and file. Acceptance: apply preflight rejects unknown/duplicate/malformed sections and unsafe carriers before writes; removal preserves other content and file modes. Check: focused Go tests passed (`go test ./internal/overlay -run 'Test(RemoteAuthorization|ApplyPreservesUpstreamRuntimeAndReusesBasePromptsForProfiles|ApplyPreflightFailureLeavesRuntimeUnchanged|OpenCodeAgentsOwnedSourceMatchesApprovedInheritedSections|ValidateOwnedAgentsComposition)'`).
- [x] RA-2: Remove the repo-owned global instruction and update manifest expectations, fixtures, and tests. Acceptance: no installed overlay-owned rule remains; upstream snapshot/audit checks remain. Checks: `go test ./internal/overlay/...` and `go test ./...` passed; the upstream snapshot and audit source remained unchanged.
- [x] RA-3: Align live maintainer/human documentation and record one consolidated closed maintenance event. Acceptance: docs accurately distinguish repository-only changes from separately authorized runtime apply. Checks: `git diff --check` passed; final consistency review found no active docs describing the retired exception.

## Progress

- Upstream v3.7.0 canonical audit passed with no managed drift; repository clean at start.
- Mapping found that the old section-replacement format rejects an empty replacement and owned `AGENTS.md` independently installs a local rule.
- All tasks are complete. Observed checks: `go test ./internal/overlay/...` passed; `go test ./...` passed; `git diff --check` passed; `git diff --quiet -- overlay/gentle-ai/assets/upstream/opencode/remote-authorization-contract.md internal/overlay/audit_upstream.go` returned exit 0. The final documentation review confirmed that only the upstream audit-only snapshot and audit invariant retain the source rule. No runtime action or commit was authorized or performed.

## Next step

No remaining task. The repository work remains uncommitted, and the installed runtime still requires a separately authorized overlay apply.
