---
name: code-design
description: >
  Apply structural code design judgment across any language and paradigm: when to split responsibilities, when to extract repetition into data, and how to comment intent over implementation.
  Trigger: When the user asks to refactor, restructure, split code, write new code from scratch, or when code is described as "too big" or needs a code review. (Also triggers on Spanish requests like "refactorizar", "dividir código", "nuevo código", "implementa", "revisión de código", "comenza con la implementacion").
license: Apache-2.0
metadata:
  author: manuelfernandez
  version: "1.2"
---

## When to Use

Use this skill:

- When writing **new code from scratch** — think about structure before the first line
- When **refactoring or splitting** existing code the user explicitly wants restructured
- When a file, class, or module is described as "too big" or "hard to follow"
- When deciding **whether and how to comment** code — what earns a comment vs. what is noise
- When adding a new item of an existing kind and it requires touching multiple places

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

## Language-Agnostic Application

Identify the primary encapsulation unit for the language in use — file, class, package, module, struct, trait, or otherwise. All rules in this skill apply to that unit.

This skill does not enumerate language idioms. That knowledge lives in the language and in language-specific skills, not here.

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
- **No catch-all units.** Do not create `utils`, `helpers`, `common`, or `misc` units at any granularity level. These are where cohesion goes to die.

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

## Tame Conditional Complexity

The problem is not `if` itself. The problem starts when nested conditionals
mix validation, business rules, and behavior selection in one flow.

Name it directly when you see it:

- **Nested conditionals** or **arrow code** when indentation keeps drifting right
- **Conditional complexity** when understanding behavior requires tracking many branches at once

Use this decision order:

1. **Guard clauses first.** If some cases are invalid, exceptional, or terminal, exit early.
   Keep the main path flat and visible.
2. **Separate rule evaluation from action.** If the code is mostly deciding *whether*
   something is allowed, move rules into clearly named predicates, specifications, or
   a decision table.
3. **Use polymorphism/strategy/state for behavior variation.** If branches exist because
   different actors, types, or states behave differently, model the variation as
   interchangeable behavior instead of a growing conditional tree.
4. **Use data when the variation is declarative.** If branches differ only by values,
   keys, or configuration — see **Replace Repetition with Data**.

### When to Stop Adding Branches

- Do **not** keep adding branches to one unit when each new case introduces a new reason to change.
- Do **not** hide conditional complexity by extracting poorly named helpers; the branching model must become clearer, not merely displaced.
- Do **not** introduce patterns prematurely. A small, local conditional is often clearer than an abstraction.

### The Test

After the change, a reader should be able to answer:

- What is the normal path?
- Which cases exit early?
- Which parts are rules?
- Which parts are true behavior variation?

If those answers are still buried across nested branches, the design is not solved yet.

## Keep One Abstraction Level Per Unit

The problem is not length. The problem is when a single unit forces the reader
to mentally shift zoom level — from orchestration to inline detail and back —
within the same flow.

Name it directly when you see it:

- **Mixed abstraction levels** when high-level policy decisions sit next to
  low-level computation, formatting, or I/O in the same unit
- **Leaking detail** when a named operation also contains the mechanics of how
  it works instead of delegating them

The signal: if you read the unit top to bottom and some lines sound like a
design document ("validate the order", "apply discounts") while others read
like an implementation manual (`sum(i["price"] * i["qty"] for i in items)`),
the levels are mixed.

Use this decision order:

1. **Name the levels first.** Before extracting anything, identify which lines
   are orchestration (what happens, in what order) and which are detail (how a
   step is done). Do not extract until the boundary is clear.
2. **Extract detail into units one level below.** The orchestrating unit should
   read as a sequence of named operations. Each extracted unit handles one
   mechanical concern at a lower level — no policy, no branching on business
   rules.
3. **Let the orchestrator read like a table of contents.** If you can replace
   any line with a meaningful name and the reader loses nothing, that line
   belongs in its own unit.

### When to Stop Extracting

- Do **not** extract just to reduce line count. The extracted unit must genuinely
  operate at a single, lower level — not be a one-line wrapper with a verbose name.
- Do **not** confuse a long unit with a mixed-level unit. A long sequence of
  operations at the same level of abstraction belongs together.
- Do **not** create a chain of delegation where each level only calls the next
  one. The level difference must be meaningful enough to name clearly.

### The Test

After the change, you should be able to read the orchestrating unit without
knowing how any step is implemented. If understanding what the unit does still
requires reading through inline computation, the levels are still mixed.

## Comment the Why, Not the What

A comment earns its place only when it explains something the code cannot express by itself.

**Keep** when:

- A decision looks wrong but is intentional — document the constraint or tradeoff
- An implicit contract exists that the signature alone doesn't reveal
- A no-op or unconditional call would otherwise invite "cleanup" that breaks behavior

**Delete** when the comment restates what the code already says clearly.

**The test**: would removing this comment force a reader to infer something non-obvious?
If yes, keep it. If no, it's noise.

## Anti-Patterns

- Extracting single-use helpers only to reduce line count.
- Splitting a long but linear algorithm into pieces — length is not mixed responsibility.
- Hiding a god object by distributing its members across units. The coupling remains; only the visibility changes.
- Creating more than one level of nesting just to organize code that has only one reason to change.
- Proposing a split when you cannot clearly name both resulting units.
- Adding a new case by copying an imperative block when the only variation is data.
- Writing comments that describe what the code does instead of why it exists.
