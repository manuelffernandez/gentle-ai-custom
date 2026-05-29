# Runbook — mantenimiento del overlay de Gentle AI

## Objetivo

Mantener la capa de personalización de `gentle-ai-custom` alineada con el upstream en `/home/manuel/Documentos/gentle-ai` sin volver a empezar desde cero en cada actualización.

## Cuándo correr este runbook

- después de actualizar `gentle-ai`
- cuando `gentle-ai sync` vuelva a introducir convenciones de PR/budget
- cuando cambie la estructura del prompt inline de algún orchestrator
- cuando aparezcan nuevas skills upstream que puedan entrar en conflicto con la política local

## Quick path

1. Trabajá desde `gentle-ai-custom`.
2. Leé la política en `overlay/gentle-ai/policy/gentle-ai-policy.json`.
3. Compará el upstream real en `/home/manuel/Documentos/gentle-ai`.
4. Revisá snapshots en `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/`.
5. Si cambió la estructura del orchestrator, ajustá el sanitizador en ambos scripts:
   - `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh`
   - `overlay/gentle-ai/scripts/apply-gentle-ai-policy.ps1`
6. Actualizá docs y skill si cambió el workflow.
7. Registrá la decisión en `overlay/gentle-ai/logs/update-log.md`.

## Checklist de mantenimiento

- [ ] La política keep/prune sigue representando la intención del usuario.
- [ ] Los scripts siguen generando prompts derivados bajo `~/.config/opencode/prompts/sdd/orchestrators/`.
- [ ] Los snapshots por orchestrator se siguen escribiendo en `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/`.
- [ ] El sanitizador todavía remueve PR/budget/chained-PR/review-workload sin romper `## Model Assignments`.
- [ ] `agent.general` sigue en `openai/gpt-5.4` / `high`.
- [ ] `agent.explore` sigue en `google-vertex/gemini-3.1-pro-preview` / `high`.
- [ ] `apply-gentle-ai-custom.sh` y `.ps1` siguen siendo el único par de entrypoints públicos y mantienen paridad funcional.

## Notas operativas

- El source of truth del orchestrator **no** es un archivo estático del repo: el script captura el prompt inline real desde `opencode.json`, lo sanitiza y recién después genera el `.overlay.md` operativo.
- Si el sanitizador no encuentra anchors esperados, debe fallar cerrado y no reescribir prompts automáticamente.
- El script principal de uso humano es `apply-gentle-ai-custom.*`; `apply-gentle-ai-policy.*` se mantiene como helper interno invocado por el entrypoint cuando aplica (`opencode` o `claude`).
