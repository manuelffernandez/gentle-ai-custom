---
name: code-modularization
description: >
  Apply consistent modularization judgment when writing or refactoring code, across any language or paradigm.
  Trigger: When writing new code from scratch, or when asked to refactor, restructure, or split existing code.
license: Apache-2.0
metadata:
  author: manuelfernandez
  version: "1.0"
---

## When to Use

Use this skill:

- When writing **new code from scratch** — think about file/module structure before writing the first line
- When **refactoring or splitting** existing code the user explicitly wants restructured
- When a file, class, or module is described as "too big" or "hard to follow"

Do **not** apply this skill to refactor code the user did not ask to change. Never propose unsolicited splits.

## The One Rule That Governs All Others

**Split by reason to change, not by line count.**

A unit is a splitting candidate only when two independent external forces could each cause it to change. Size is a symptom; mixed responsibilities are the actual problem.

## Three Questions Before Every Split

Before proposing or making a split, answer these:

1. **What would make this change?** — list the forces. If more than one exists independently, you have a candidate.
2. **What lives and dies together?** — code that always changes in the same edit belongs in the same unit.
3. **Can you name it?** — if you can't give a proposed split a clear, single-purpose name, the boundary is wrong.

## Extraction Priority Order

When splitting a large unit, apply this sequence:

1. **Pure units first.** Functions or classes with zero dependency on mutable state or I/O. These are always safe to extract and provide the most immediate clarity gain.
2. **Domain clusters second.** Coherent groups of types and their operations that share a single responsibility (e.g., validation logic, profile reconciliation, snapshot management). Name the cluster; if you can't, it is not a cluster.
3. **Orchestration stays last.** The main pipeline or coordinator stays in place until everything extractable is gone. What remains after extraction is the true core — it may still be large, and that is fine.

## Paradigm Adaptation

The reasoning is the same across paradigms. Only the encapsulation unit differs:

| Paradigm | Encapsulation unit | Typical split trigger |
|---|---|---|
| Functional | Function / module / file | Mixes pure and effectful logic; module handles unrelated domains |
| OOP | Class / interface / package | Class has multiple responsibilities; package mixes unrelated hierarchies |
| Mixed | Either, per context | Apply FP rules to pure functions, OOP rules to stateful objects in the same codebase |

In a mixed codebase, identify the paradigm in use for each unit and apply the matching rules.

## When Writing New Code

Before writing the first file, ask:

- How many independent reasons to change does this feature have?
- Which parts are pure computation vs. I/O vs. orchestration?
- What is the smallest set of files that maps cleanly to those boundaries?

Start with the fewest files that honestly separate the concerns. Do not pre-split speculatively.

## Hard Stops

- **Never split by line count alone.** 400 cohesive lines in one file beats four files of 100 lines with artificial boundaries.
- **Name the god object.** If a struct, class, or module accumulates state or behavior for multiple domains, say so explicitly. Splitting files around it without addressing the root type is cosmetic — the coupling survives.
- **Stop when what remains is genuinely cohesive.** If the residual core is the real orchestrator or pipeline, it belongs together even if it is still large.
- **No catch-all files.** Do not create `utils`, `helpers`, `common`, or `misc` units. These are where cohesion goes to die.

## Replace Repetition with Data

When adding a new item of the same kind requires touching more than one place,
the variation is likely **data**, not behavior.

**N-touch test**: if adding one conceptual unit (a config entry, a handler, a validator) forces
edits in N > 1 locations, extract the varying part to a declarative structure
and iterate over it once.

Apply when: items have the same shape and the only difference between them is
identity (name, path, key).

Do not apply when: items have genuinely different behavior — use polymorphism
or strategy instead. Or when only two instances exist and abstraction cost
exceeds the gain.

**Catch it early**: the highest-value moment is when the second or third item
appears. Recognizing it then costs far less than unwinding five duplicated
blocks later.

## Anti-Patterns

- Extracting single-use helpers only to reduce line count.
- Splitting a long but linear algorithm into pieces — length is not mixed responsibility.
- Hiding a god object by distributing its fields across files. The coupling remains; only the visibility changes.
- Creating more than one level of nesting just to organize code that has only one reason to change.
- Proposing a split when you cannot clearly name both resulting units.
- Adding a new case by copying an imperative block when the only variation is data.
