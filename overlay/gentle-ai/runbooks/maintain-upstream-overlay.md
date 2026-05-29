# Runbook — mantenimiento del overlay de Gentle AI

## Objetivo

Mantener la capa de personalización de `gentle-ai-custom` alineada con el upstream en `/home/manuel/Documentos/gentle-ai` sin volver a empezar desde cero en cada actualización.

## Modelo operativo

Este mantenimiento se apoya en cuatro capas explícitas:

- **Intento** → `overlay/gentle-ai/policy/maintenance-intent.md`
- **Política** → `overlay/gentle-ai/policy/gentle-ai-policy.json`
- **Estado** → `overlay/gentle-ai/state/upstream-state.json`
- **Log** → `overlay/gentle-ai/logs/update-log.md`

No cumplen el mismo rol:

- el **intento** describe el criterio del usuario
- la **política** alimenta los scripts runtime
- el **estado** marca desde dónde hay que auditar el upstream
- el **log** deja historial narrativo de mantenimiento

## Cuándo correr este runbook

- después de actualizar `gentle-ai`
- cuando `gentle-ai sync` vuelva a introducir convenciones de PR/budget
- cuando cambie la estructura del prompt inline de algún orchestrator
- cuando aparezcan nuevas skills upstream que puedan entrar en conflicto con la política local

## Quick path

1. Trabajá desde `gentle-ai-custom`.
2. Leé el intento en `overlay/gentle-ai/policy/maintenance-intent.md`.
3. Leé la política en `overlay/gentle-ai/policy/gentle-ai-policy.json`.
4. Leé el estado en `overlay/gentle-ai/state/upstream-state.json`.
5. Determiná el rango a auditar desde `last_maintained_commit` / `last_maintained_tag` hasta el estado actual del upstream.
6. Revisá snapshots en `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/`.
7. Si cambió la estructura del orchestrator, ajustá el sanitizador en ambos scripts:
   - `overlay/gentle-ai/scripts/apply-gentle-ai-policy.sh`
   - `overlay/gentle-ai/scripts/apply-gentle-ai-policy.ps1`
8. Separá cambios relevantes del overlay de bugfix/chore noise.
9. Si hay cambios relevantes de comportamiento o nuevas convenciones, frená y pedile al usuario una decisión explícita.
10. Actualizá docs, skill, política y estado si cambió el workflow.
11. Registrá la decisión en `overlay/gentle-ai/logs/update-log.md`.

## Gate humana

No todo cambio upstream debe mutar el overlay.

El agente debe pedir aprobación humana cuando aparezcan cambios que puedan afectar:

- la intención keep/prune
- la conveniencia de nuevas skills upstream
- la lógica del sanitizador del orchestrator
- la interpretación de qué cambios sí importan para este repo

## Checklist de mantenimiento

- [ ] `maintenance-intent.md` sigue representando lo que el usuario quiere conservar y depurar.
- [ ] La política keep/prune sigue representando la intención del usuario.
- [ ] `upstream-state.json` apunta a la última versión/commit realmente mantenida.
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
- El log no reemplaza al estado: `update-log.md` cuenta qué se decidió; `upstream-state.json` marca cuál fue la última versión/commit mantenida.
