# Maintenance

Guía humana del overlay de Gentle AI.

Este archivo describe el modelo operativo vigente: `apply-gentle-ai-custom` reinstala assets repo-owned desde `overlay/gentle-ai/assets/owned/...`; `audit-gentle-ai-upstream` y `sync-gentle-ai-upstream-assets` mantienen la relación con upstream y el baseline aprobado.

## Qué no define este archivo

- reglas de comportamiento del agente/runtime -> `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`
- ownership y policy del repo -> `AGENTS.md`
- intención keep/prune y comportamiento deseado del orchestrator -> `policy/maintenance-intent.md`
- policy machine-readable -> `policy/gentle-ai-policy.json`
- frontera upstream mantenida -> `state/upstream-state.json`
- bitácora de eventos cerrados -> `logs/update-log.md`

## Quick path

1. Actualizá el binario de `gentle-ai`.
2. Corré `git pull` en el clone upstream resuelto de `gentle-ai`.
3. Desde `gentle-ai-custom`, corré `bash audit-gentle-ai-upstream.sh`.
4. Si la auditoría muestra drift relevante, actualizá primero este repo.
5. Si aprobaste una nueva frontera upstream, corré `bash sync-gentle-ai-upstream-assets.sh` para refrescar `overlay/gentle-ai/assets/upstream/...` y el baseline auditado.
6. Ejecutá el refresh upstream recomendado por la auditoría:
   - `gentle-ai sync` si la topología no cambió
   - reinstalación completa si cambió la topología o sync ya no materializa la forma correcta
7. Reaplicá el overlay con `bash apply-gentle-ai-custom.sh opencode`.
8. Leé `Summary:` y verificá el estado final en disco.
9. Si cambió `~/.config/opencode/opencode.json`, reiniciá OpenCode.

## Modelo operativo

| Artefacto | Rol |
| --- | --- |
| `policy/maintenance-intent.md` | Intención humana: qué conservar, depurar y proteger |
| `policy/gentle-ai-policy.json` | Policy operativa consumida por la CLI Go y los wrappers |
| `policy/managed-assets.json` | Mapa canónico de assets upstream aprobados y assets owned instalables |
| `assets/upstream/` | Copias upstream aprobadas para review/diff |
| `assets/owned/` | Assets repo-owned que `apply` instala en runtime |
| `shared/skills/` | Skills portables repo-owned instaladas globalmente por `apply` |
| `shared/commands/` | Cuerpos fuente para wrappers custom renderizados por `apply` |
| `state/upstream-state.json` | Última frontera upstream mantenida |
| `snapshots/upstream/opencode/orchestrators/` | Baseline auditado de `gentle-orchestrator` usado por audit/sync |

## Qué hace cada comando

### `audit-gentle-ai-upstream`

- usa `last_maintained_commit` de `state/upstream-state.json`
- descubre drift con `git diff --name-status --find-renames <last_maintained_commit>..HEAD`
- filtra ese drift con `policy/managed-assets.json`
- sigue verificando invariantes estructurales de upstream (`profiles.go`, `inject.go`, etc.)
- valida que el baseline auditado (`snapshots/.../gentle-orchestrator.last.md` + metadata) siga consistente con upstream/state

### `sync-gentle-ai-upstream-assets`

- copia assets upstream aprobados hacia `assets/upstream/...`
- actualiza el baseline auditado de `gentle-orchestrator`
- avanza `state/upstream-state.json` cuando el nuevo upstream fue aceptado
- no toca runtime local bajo `~/.config/opencode/`

### `apply-gentle-ai-custom`

- instala assets SDD/runtime repo-owned desde `assets/owned/...`
- reescribe `opencode.json` para que base y perfiles SDD usen esos prompt files
- poda las skills upstream rechazadas solo en los targets registrados seleccionados
- aplica `agent_overrides`
- reconcilia `default_profile` y `profiles`
- instala skills repo-owned desde `shared/skills/`
- renderiza wrappers custom desde `shared/commands/`

`apply` ya NO depende de sanitización, captura de prompts inline, snapshots operativos locales ni recovery desde snapshots.

## Tipos de actualización e impacto

| Vía de actualización | Qué cambia | Impacto en el overlay |
| --- | --- | --- |
| `brew upgrade gentle-ai` | Solo el binario | Normalmente no resetea el overlay |
| `gentle-ai sync` | Prompts, skills, MCP configs, assets SDD | Restaura estado upstream en runtime; reaplicar overlay es obligatorio |
| Reinstalación por TUI | Instalación completa, topología, presets y config | Resetea todo y puede cambiar agentes/presets |

## Señales de mayor valor

| Señal | Significado | Acción |
| --- | --- | --- |
| `base prompt drift: yes` | Cambió `gentle-orchestrator` upstream respecto del baseline auditado | Leer primero `Drift summary:` |
| `profile ... mismatch` / `base asset injection invariant: mismatch` | Cambió la mecánica upstream de perfiles SDD | Frenar y auditar antes de recomendar `sync` |
| `topology: unknown orchestrator matched by prefix only` | Apareció un orchestrator upstream nuevo | Auditarlo y decidir si la policy debe incluirlo |
| `topology: expected orchestrator missing from opencode.json` | Un orchestrator conocido desapareció o fue renombrado | Auditar upstream y actualizar policy/intent si hace falta |
| `WARNING - unmanaged SDD profiles left untouched` | Hay perfiles presentes en `opencode.json` que no están en la fuente activa `profiles` | Decidir si gestionarlos en config local o removerlos manualmente |
| `owned asset writes - ...` | `apply` instaló o dejó intactos los assets repo-owned | Revisar verbose si necesitás detalle por archivo |

## Verificación post-state

Después del apply, confirmá esto:

- las skills podadas ya no existen en cada target registrado seleccionado
- cada `agent_override` efectivo resuelve al `model` / `variant` esperado
- `agent.gentle-orchestrator.prompt` apunta a `~/.config/opencode/prompts/sdd/orchestrators/gentle-orchestrator.overlay.md`
- cada `sdd-orchestrator-<name>` gestionado apunta a ese mismo prompt file owned
- cada fase `sdd-<phase>` y cada fase gestionada `sdd-<phase>-<name>` apunta a su prompt file owned bajo `~/.config/opencode/prompts/sdd/`
- los files/directorios runtime declarados en `policy/managed-assets.json` existen en disco
- si `default_profile` existe, la familia base mantiene `model` y `variant` correctos
- si `profiles` existe, cada perfil declarado mantiene `model` y `variant` correctos
- `snapshots/upstream/opencode/orchestrators/gentle-orchestrator.last.md` y `.meta.yaml` siguen consistentes con `state/upstream-state.json`

## Config local del overlay

El config por máquina canónico vive fuera del repo en `~/.config/gentle-ai-custom/opencode-local-config.json`.

Reglas operativas:

- `upstream_repo_path` tiene precedencia sobre `GENTLE_AI_CUSTOM_UPSTREAM_REPO`, y ambos sobre `../gentle-ai`
- `opencode_config_path` es opcional; si se omite, el default sigue siendo `~/.config/opencode/opencode.json`
- `agent_overrides` maneja solo asignaciones explícitas para built-in agents como `general` o `explore`
- `default_profile` maneja solo la familia base `gentle-orchestrator` + fases SDD sin sufijo
- `profiles` maneja solo familias SDD nombradas (`sdd-orchestrator-<name>` + fases)
- perfiles existentes no declarados quedan intactos y se reportan como unmanaged

## Checklist de mantenimiento

- [ ] `maintenance-intent.md` sigue reflejando qué conservar y qué depurar
- [ ] `managed-assets.json` sigue alineado con `assets/upstream/` y `assets/owned/`
- [ ] `upstream-state.json` sigue apuntando a la última frontera upstream realmente mantenida
- [ ] `apply-gentle-ai-custom` sigue reinstalando prompt refs desde assets owned
- [ ] el baseline auditado de `gentle-orchestrator` sigue consistente con su metadata
- [ ] los entrypoints públicos en shell y PowerShell siguen siendo equivalentes
- [ ] los assignments de perfiles SDD siguen siendo locales y no reaparecieron en la policy versionada
- [ ] la resolución upstream sigue respetando: config local -> env -> fallback `../gentle-ai`

## Referencias

- `README.md`
- `AGENTS.md`
- `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`
- `policy/maintenance-intent.md`
- `policy/gentle-ai-policy.json`
- `policy/managed-assets.json`
- `state/upstream-state.json`
- `logs/update-log.md`
