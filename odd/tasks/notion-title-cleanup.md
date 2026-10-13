# Notion Title Cleanup

## Objective

Remove inherited two-digit title prefixes from the global index, canonical project guides and numbered EcoSport items. Preserve stable identities and current topology.

## Authorized scope

- Global index `3dca58c3fefc815fb60ded077b5127d1`: `Leer primero — Índice de proyectos`.
- Guides: EcoSport `3e9a58c3fefc81db80bcefc42e64ef4f`, E-commerce `3dca58c3fefc8161a750e8c54d99c0db`, Me `3f7a58c3fefc811b91bdf180d43c3736`, Teleférico `3dca58c3fefc81f89064cc55729b7ec5`.
- Each guide title becomes `Guía del proyecto — <project name>`.
- EcoSport pages `3e9a58c3fefc814bbbcfff240214c2e1` and `3e9a58c3fefc815e917fd9439700cbad`: `Vehículo` and `Referencias técnicas`.
- History database `53a95293fd0c402a9bf446322f198c85` and sole data source `cc0207a1-3068-47c9-9355-2a79a32b06db`: `Historial del vehículo`.
- Correct three literal numbered labels in the EcoSport guide and add a short positive naming convention to the global creation procedure.
- One exact display-label correction on the existing EcoSport history child tag, preserving its attributes and placement.

## Constraints

- Titles/content follow the existing Spanish language; this tracking document is English.
- No pages created, moved or deleted; no row/property/schema/view changes.
- Preserve icons, covers, URLs, child IDs/order, canonical routes and non-target content.
- Native mentions and child labels are verified after renames; avoid rewriting child blocks.
- The data-source tool guarantees only source naming. Verify database-container title separately and report any unsupported remainder.
- Keep existing skill, runtime, configuration, upstream-maintenance log and prior pending review unchanged.

## Task

- [ ] **T1 — Rename verified targets and reconcile references.**
  - Route: delegated direct; bounded worker handles multiple live objects and verification.
  - Perform seven page-title updates and one supported title-only source/database naming operation.
  - Perform only the two necessary targeted content edits.
  - Read back every affected title/reference and preserve before-state structure.
  - Record partial/unknown operations honestly and close with one local evidence commit.
  - Status: partial; all actual titles and guide labels corrected, one exported root label remains.

## Checks

- No meaningful deterministic RED or software runner for copy-only live title changes.
- Native fetch/readback for seven pages, database and collection, plus affected project roots.
- Verify exact titles, canonical IDs, icons, child topology, database schema/views and corrected literal labels.
- Native title/collection result uncertainty is not success; use no blind retry or schema-change workaround.

## Delivery

- Branch: `codex/notion-title-cleanup`, starting at `91649a0`.
- Strategy: ask-on-risk; forecast approximately 60 authored tracking lines, cloud edits excluded.
- No push, pull request, runtime apply, sync or source-skill change.
- Prior Notion-module native review remains pending independently.

## Progress

- Seven page titles, history database and sole source renamed and fetched successfully.
- Three EcoSport guide labels corrected with unchanged URLs; global creation rules now use purpose-based titles and `Guía del proyecto — <project name>` without numerical prefixes.
- Native references followed page renames. Worker readbacks preserved page icons/covers, ancestry, child IDs/order, non-target text, history schema and all six views; 31 history rows before/after. Database icon/cover metadata was unavailable, not independently verified.
- Parent spotcheck confirmed global naming rule, clean database/source titles and stable EcoSport child identities.
- Remaining: EcoSport root still exports `02 — Historial del vehículo` in its native child database tag. One exact same-URL targeted correction returned HTTP 400 `validation_error`, `No matches found`, nonretryable. Fresh readbacks showed no mutation; no retry or structural workaround.
- Software RED/tests and runtime harness: N/A for cloud title/copy edits and passive tracking only. No source-skill change or runtime installation.
- Rollback boundary: this task document only in Git; live title changes would require separate explicit Notion reversal, not a Git revert.
- Local evidence commit: pending.

## Next step

Record the local evidence commit and report the remaining exported EcoSport label honestly. Resolve that label only through a verified supported title-only operation; preserve the existing child database.
