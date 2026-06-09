# Gentle AI overlay assets

Esta carpeta agrupa los assets operativos del overlay de Gentle AI.

## Qué leer primero

- guía principal del repo: `../../README.md`
- guía humana de mantenimiento: `maintenance.md`
- skill operativa del maintainer: `../../.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`

## Qué vive acá

- `maintenance.md` — guía humana centralizada para mantenimiento
- `assets/upstream/` — copias upstream aprobadas para review/diff
- `assets/owned/` — assets repo-owned que `apply-gentle-ai-custom` instala en runtime
- `policy/` — intención y policy machine-readable del overlay
- `state/` — frontera upstream mantenida
- `logs/` — historial de decisiones cerradas del overlay

## Modelo vigente

- `policy/managed-assets.json` es el mapa canónico entre upstream aprobado y runtime owned.
- `audit-gentle-ai-upstream` cubre audit + recomendación antes de cualquier mutación.
- `sync-gentle-ai-upstream-assets` hace el repo sync de `assets/upstream/...` más la frontera aprobada; no refresca el runtime.
- `apply-gentle-ai-custom` usa `assets/owned/...` como source of truth para los prompts/skills/commands SDD del runtime.
- El runtime refresh (`gentle-ai sync` o reinstalación completa) es un paso separado y depende de si el cambio adoptado afecta el runtime target/materialized state que este repo mantiene.
- `shared/skills/` sigue siendo la fuente canónica de las skills repo-owned portables; no se mueve a este árbol.

## Regla simple

- Si querés entender o usar el repo, arrancá por `README.md` en la raíz.
- Si estás manteniendo el overlay frente a cambios del upstream, seguí `maintenance.md`.
