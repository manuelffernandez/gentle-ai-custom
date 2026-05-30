# Gentle AI Overlay / Control Plane

Este overlay mantiene una política **persistente y reaplicable** para tu stack de Gentle AI, desacoplada del repo upstream (`/home/manuel/Documentos/gentle-ai`).

## Quick path

1. Ejecutá tu sync normal de Gentle AI.
2. Reaplicá la capa custom completa:
   - Linux/macOS: `bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh all`
   - Windows: `~\Documentos\gentle-ai-custom\apply-gentle-ai-custom.ps1 all`
3. Reiniciá OpenCode si el script tocó `opencode.json`.

Cuando audités cambios del upstream con la skill maintainer, esa skill te tiene que decir además si ese delta se resuelve con `gentle-ai sync` o si requiere reinstalación completa. Si hubo cambios de topología, `sync` no alcanza.

## Qué contiene

- `policy/gentle-ai-policy.json`  
  Baseline machine-readable de keep/prune, overrides de agentes (`general`, `explore`), rutas operativas de OpenCode, y la ruta al config local de perfiles SDD (`opencode.sdd_profiles_local_config_path`).
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
  Helpers internos que podan skills, aplican `agent_overrides`, reconcilian perfiles SDD desde el config local (`~/.config/gentle-ai-custom/opencode-sdd-profiles.json`), capturan prompts inline y generan orchestrators derivados bajo `~/.config/opencode/prompts/sdd/orchestrators/`. Mantienen el snapshot versionado de `gentle-orchestrator` en el repo, mantienen snapshots operativos locales bajo `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/`, recuperan desde snapshot si un `.overlay.md` faltó en disco, detectan drift de topología (orchestrators desconocidos/faltantes), validan estrictamente el config local de perfiles antes de cualquier escritura (fail-closed) y verifican post-write que `opencode.json` quedó consistente.

### Config externo gestionado fuera del repo

- `~/.config/gentle-ai-custom/opencode-sdd-profiles.json` (NO está en el repo)  
  Config per-máquina de perfiles SDD. Schema V1 strict, sin defaults ni herencia. Si no existe, el helper no toca ningún perfil SDD en `opencode.json`. Detalle del schema y reglas duras en el `README.md` raíz, sección "Perfiles SDD locales".
- `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/` (NO está en el repo)  
  Directorio operativo per-máquina para snapshots de orchestrators. Guarda siempre `gentle-orchestrator.last.md` como copia operativa local y guarda los `sdd-orchestrator-<profile>.last.md` solo acá.

## Convenciones

- El source of truth del orchestrator **no** es un archivo estático del repo.
- El helper lee el prompt inline real desde `~/.config/opencode/opencode.json`, lo snapshottea por agente, lo sanitiza y recién después genera el `.overlay.md` operativo.
- `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/` conserva solo el baseline versionado de `gentle-orchestrator`.
- `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/` es la fuente operativa local para recuperación: `gentle-orchestrator` se recupera desde ahí y, si falta, puede caer al snapshot versionado del repo; `sdd-orchestrator-<profile>` se recupera solo desde el directorio local.
- Si el `.overlay.md` falta en disco pero el snapshot recuperable existe, el helper recupera desde ese snapshot. Si no existe, falla cerrado pidiendo `gentle-ai sync`.
- Si faltan anchors esperados, el sanitizador debe fallar cerrado y NO reescribir prompts automáticamente.
- El repo upstream se trata como **fuente de verdad de entrada**; este overlay como **fuente de verdad de decisiones locales**.
- El maintainer debe leer las cuatro capas en este orden: `maintenance-intent.md` → `gentle-ai-policy.json` → `upstream-state.json` → `update-log.md`.
- `update-log.md` no reemplaza al estado upstream mantenido; solo deja trazabilidad narrativa.
- Cada cambio sobre cualquier asset del overlay agrega una entrada a `update-log.md` (ver `AGENTS.md` regla 4).
