---
name: chained-pr
description: "Trigger: PRs over 400 lines, stacked PRs, review slices. Split oversized changes into chained PRs that protect review focus."
license: Apache-2.0
metadata:
  author: gentleman-programming
  version: "1.0"
---

## Activation Contract

Load this skill when a planned PR may exceed **400 changed lines**, SDD forecasts `400-line budget risk: High` or `Chained PRs recommended: Yes`, or the user asks for chained/stacked PRs, review slices, or reviewer-load control.

## Hard Rules

- Split PRs over **400 changed lines** unless a maintainer explicitly accepts `size:exception`.
- Keep each PR reviewable in about **≤60 minutes**.
- Use one deliverable work unit per PR; keep tests/docs with the unit they verify.
- State start, end, prior dependencies, follow-up work, and out-of-scope items in every chained PR.
- Every child PR must include a dependency diagram marking the current PR with `📍`.
- In Feature Branch Chain, create a draft/no-merge tracker PR; child PR #1 targets the tracker branch, later children target the immediate parent branch.
- Treat polluted diffs as base bugs: retarget or rebase until only the current work unit appears.
- Do not mix chain strategies after the user chooses one.
- Before any GitHub CLI usage, resolve the selected remote and the target `owner/repo` from that remote.
- Pass `--repo <owner/repo>` to every `gh` command so chained-PR reads and writes never rely on the current directory's implicit repo.
- Never build shell command strings by concatenating or interpolating repo names, branch names, PR titles, body-file paths, or labels. Use safe argument execution or tool APIs.
- Automatic activation may plan and report a chain, but it must not create branches, push, or open/update PRs. Branch or PR mutation requires an explicit user request for chained PR execution or an approved handoff to the repo PR workflow.

## Decision Gates

| Condition | Action |
|---|---|
| PR ≤400 changed lines and focused | Keep single PR. |
| PR >400, each slice can land independently | Use Stacked PRs to main. |
| PR >400, feature must integrate before main | Use Feature Branch Chain with tracker. |
| Generated/vendor/migration diff cannot split cleanly | Ask maintainer for `size:exception`. |
| SDD provides `delivery_strategy` | Follow it before apply/PR creation. |
| Chain is detected automatically by SDD or review-budget signals | Plan the chain only; ask before branch or PR mutation. |

## Execution Steps

1. Estimate changed lines and identify independent work units.
2. Ask for a chain strategy when none is cached and the budget is exceeded.
3. Resolve the selected remote plus `owner/repo` before any GitHub CLI reads or writes.
4. If the user did not explicitly request chained PR execution, stop with the chain plan and required approval boundary.
5. If execution is approved, create branches/PRs using the chosen strategy only.
6. For approved execution, add Chain Context to each PR without replacing the repo PR template.
7. For approved execution, verify each PR independently: CI/tests/docs/manual checks, rollback scope, and clean diff.
8. For approved Feature Branch Chain execution, keep tracker PR draft/no-merge until all child PRs are reviewed and integrated.

## Output Contract

Return the chosen strategy, resolved remote/repo scope, PR order, current PR boundary, dependency diagram, review budget (`additions + deletions`), verification plan, and any `size:exception` rationale.

## References

- [assets/chaining-details.md](assets/chaining-details.md) — strategy diagrams, repo-scoped GH CLI guidance, PR body section, branch commands, and reviewer guidance.
