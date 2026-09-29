# Remote authorization scope

## Objective

Make explicit requests for remote operations usable with the tool's normally configured authentication without a redundant confirmation, while requiring clarification for ambiguous destination or account and forbidding credential/session discovery or switching. Keep upstream audit snapshots unchanged. The user will run overlay apply separately.

## Scope and constraints

- Authorized: repo-local, narrowly scoped customization of upstream-installed OpenCode remote-authorization prompt projections, associated policy, tests, and documentation.
- Not authorized: modifying the upstream checkout or audit snapshots, executing sync/apply or authenticated remote operations, committing, pushing, or creating PRs.
- Apply must fail closed on unknown upstream anchors before writes, retain unrelated bytes/modes, and be idempotent.
- Route: delegated direct. Mapping required more than three files; implementation spans multiple non-trivial files and must remain one writer.
- TDD: no current active ODD-specific TDD setting found; use ordinary Go checks (`go test ./internal/overlay`, `go test ./...`, `git diff --check`) and report any configuration discovered before implementation.
- Delivery strategy: ask-on-risk. No delivery operation is authorized.
- Initial forecast: likely more than 400 authored changed lines across runtime wiring, tests, and documentation; advisory task size only, not an acceptance gate.

## Tasks

- [x] RA-1: Corrected the exact-section parser and regression fixture for the installed upstream closing marker. The parser now replaces only the marked body, preserves opening/closing markers and surrounding bytes, supports a closing marker at EOF or before a later heading, remains idempotent, and rejects malformed/duplicate boundaries and unexpected pre-heading content. Realistic marked fixtures now cover `sdd-onboard`; malformed-marker preflight is verified write-free. `go test ./internal/overlay`, `go test ./...`, and `git diff --check` passed. Commit: not authorized.
- [x] RA-2: Updated the owned global instructions, maintenance intent, root instructions, Spanish README, overlay README, maintenance guide, maintainer skill, and one closed-event log entry. The approved upstream snapshot remains unchanged and apply remains a separate user action. Commit: not authorized.
- [x] RA-3: Verified the repository diff, required Go tests, and an independent fresh-context consistency pass (PASS). Native review was approved and acknowledged for the frozen implementation candidate. The separately authorized runtime apply remains pending; no commit was authorized.
- [x] RA-4: Verified the marked-section correction with a realistic `sdd-onboard` fixture, required Go suite, parent readback, and independent fresh-context consistency check. The prior RA-3 native receipt remains specific to the earlier frozen candidate; installed runtime behavior is still pending user-run apply. Commit: not authorized.

## Progress and evidence

- Canonical audit: `bash audit-gentle-ai-upstream.sh` passed for v3.7.0 (`6dee8f833aec9e46015759c5065a9035795d9af1`) with no managed drift and remote-authorization invariant `ok`.
- User approved a local effective customization instead of editing the audit-only upstream snapshot.
- The new manifest-declared local section delta leaves `assets/upstream/opencode/remote-authorization-contract.md` unchanged. Tests verify exact section replacement, atomic file writes preserving modes, idempotence, prewrite rejection, and preservation of unrelated prompts/content.
- Observed checks: `go test ./internal/overlay` passed; `go test ./...` passed for `cmd/gentle-ai-overlay` and `internal/overlay`; `git diff --check` passed.
- Fresh-context consistency review: PASS, no blocker. Native high-risk review: approved and acknowledged for target `sha256:38b9141683a319594128fa9d2633fb117b29196c1c33fb0bf082bdaf6d4a4f02` (lineage `review-753cd08983a29b68`). One non-blocking resilience warning remains: a failure during a later prompt write could leave earlier prompt files patched and later ones unchanged; this review did not offer a correction transition. This task-document progress update happened after acknowledgement and is not part of that frozen target.
- No upstream clone, runtime, credential file, or network/authenticated operation was accessed. No sync, apply, commit, push, or PR was run.
- User-run `bash apply-gentle-ai-custom.sh opencode --verbose` returned preflight error for `sdd-onboard-mixed` and `(no changes — all assets already up to date)`. Read-only diagnosis confirmed the installed prompt contains the expected old section followed by a closing marker; the section parser treats the marker as part of the body, unlike the marker-free fixture. A fresh v3.7.0 upstream audit again passed with no drift. The generic `sync`/reinstall advice is not a remedy for this local parser bug. User approved a repo-only fix, not apply/sync.
- Correction verification: added a marked prompt regression first; focused tests failed as expected because the closing marker was included in the compared body. After the parser fix, focused and required Go checks passed. This does not verify installed runtime behavior; no apply or sync was run.
- Independent correction consistency check found no parser/test defect. It ran `go test -count=1 ./internal/overlay`, `go test -count=1 ./...`, and `git diff --check` successfully; its only reported stale task status was resolved by this RA-4 update. The correction has not been covered by the earlier native review target.

## Next step

The user can retry overlay apply separately and inspect its result after the repo checks; runtime behavior is not yet verified. No automatic commit.
