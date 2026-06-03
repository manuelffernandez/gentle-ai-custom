# Gentle AI overlay assets

Esta carpeta agrupa los assets operativos del overlay de Gentle AI.

## Qué leer primero

- Guía principal del repo: `../../README.md`
- Procedimiento detallado de mantenimiento: `runbooks/maintain-upstream-overlay.md`
- Skill operativa del maintainer: `../../.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`

## Qué vive acá

- `policy/` — intención y policy machine-readable del overlay
- `state/` — frontera upstream mantenida
- `logs/` — historial de decisiones del overlay
- `runbooks/` — procedimiento humano detallado
- `scripts/` — wrappers internos finos hacia la CLI Go compartida
- `snapshots/` — baseline versionado del `gentle-orchestrator`

## Regla simple

- Si querés entender o usar el repo, arrancá por `README.md` en la raíz.
- Si estás manteniendo el overlay frente a cambios del upstream, seguí el runbook.
