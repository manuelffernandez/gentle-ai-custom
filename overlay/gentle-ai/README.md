# Gentle AI overlay assets

Esta carpeta agrupa los assets operativos del overlay de Gentle AI.

## Qué leer primero

- Guía principal del repo: `../../README.md`
- Guía humana de mantenimiento: `maintenance.md`
- Skill operativa del maintainer: `../../.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`

## Qué vive acá

- `maintenance.md` — guía humana centralizada para mantenimiento, señales y notas técnicas
- `policy/` — intención y policy machine-readable del overlay
- `state/` — frontera upstream mantenida
- `logs/` — historial de decisiones del overlay
- `scripts/` — wrappers internos finos hacia la CLI Go compartida
- `snapshots/` — baseline versionado del `gentle-orchestrator`

## Regla simple

- Si querés entender o usar el repo, arrancá por `README.md` en la raíz.
- Si estás manteniendo el overlay frente a cambios del upstream, seguí `maintenance.md`.
