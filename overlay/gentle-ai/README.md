# Gentle AI Overlay / Control Plane

Este overlay mantiene una política **persistente y reaplicable** para tu stack de Gentle AI, desacoplada del repo upstream (`/home/manuel/Documentos/gentle-ai`).

## Quick path

1. Ejecutá tu sync normal de Gentle AI.
2. Reaplicá la capa custom completa:
   - Linux/macOS: `bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh all`
   - Windows: `~\Documentos\gentle-ai-custom\apply-gentle-ai-custom.ps1 all`
3. Reiniciá OpenCode si el script tocó `opencode.json`.

## Qué contiene

- `policy/gentle-ai-policy.json`  
  Baseline machine-readable de keep/prune, overrides de agentes (`general`, `explore`) y rutas operativas de OpenCode.
- `policy/maintenance-intent.md`  
  Fuente de verdad semántica: explica qué se quiere conservar, qué se quiere depurar y por qué.
- `policy/orchestrator-policy.md`  
  Reglas de limpieza del prompt inline de los orchestrators.
- `state/upstream-state.json`  
  Frontera operativa de la última versión/tag/commit upstream mantenido.
- `runbooks/maintain-upstream-overlay.md`  
  Runbook humano para mantener la capa local frente a cambios del upstream.
- `logs/update-log.md`  
  Log incremental de decisiones y updates aplicados.
- `scripts/apply-gentle-ai-policy.sh` y `.ps1`  
  Helpers internos que podan skills, aplican `agent_overrides`, capturan prompts inline y generan orchestrators derivados bajo `~/.config/opencode/prompts/sdd/orchestrators/`. También recuperan desde snapshot si un `.overlay.md` faltó en disco, detectan drift de topología (orchestrators desconocidos/faltantes) y verifican post-write que `opencode.json` quedó consistente.

## Convenciones

- El source of truth del orchestrator **no** es un archivo estático del repo.
- El helper lee el prompt inline real desde `~/.config/opencode/opencode.json`, lo snapshottea por agente, lo sanitiza y recién después genera el `.overlay.md` operativo.
- Si el `.overlay.md` falta en disco pero el snapshot existe, el helper recupera desde el snapshot. Si no hay snapshot, falla cerrado pidiendo `gentle-ai sync`.
- Si faltan anchors esperados, el sanitizador debe fallar cerrado y NO reescribir prompts automáticamente.
- El repo upstream se trata como **fuente de verdad de entrada**; este overlay como **fuente de verdad de decisiones locales**.
- El maintainer debe leer las cuatro capas en este orden: `maintenance-intent.md` → `gentle-ai-policy.json` → `upstream-state.json` → `update-log.md`.
- `update-log.md` no reemplaza al estado upstream mantenido; solo deja trazabilidad narrativa.
- Cada cambio sobre cualquier asset del overlay agrega una entrada a `update-log.md` (ver `AGENTS.md` regla 4).
