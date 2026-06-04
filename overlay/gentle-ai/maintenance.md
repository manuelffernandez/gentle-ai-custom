# Maintenance

Guía humana de mantenimiento del overlay de Gentle AI.

Este archivo centraliza el flujo operativo, los puntos de decisión, las señales de recovery y las notas técnicas útiles durante el mantenimiento contra upstream.

Lo que este archivo **no** define:

- reglas de comportamiento del agente/runtime → `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`
- reglas de ownership y policy del repo → `AGENTS.md`
- intención keep/prune → `overlay/gentle-ai/policy/maintenance-intent.md`
- policy machine-readable → `overlay/gentle-ai/policy/gentle-ai-policy.json`
- última frontera upstream mantenida → `overlay/gentle-ai/state/upstream-state.json`
- bitácora de eventos cerrados → `overlay/gentle-ai/logs/update-log.md`

## Quick path

1. Actualizá el binario de `gentle-ai`.
2. Corré `git pull` en `/home/manuel/Documentos/gentle-ai`.
3. Desde `gentle-ai-custom`, corré `bash audit-gentle-ai-upstream.sh`.
4. Si la auditoría muestra drift relevante para el overlay, adaptá primero este repo.
5. Ejecutá el refresh upstream recomendado por la auditoría:
   - `gentle-ai sync` si la topología no cambió
   - reinstalación completa si cambió la topología o sync no puede materializar la nueva forma upstream
6. Reaplicá el overlay:
   - `bash apply-gentle-ai-custom.sh opencode`
   - o `bash apply-gentle-ai-custom.sh all`
7. Leé `Summary:` y cualquier `Drift summary:` o warning `topology:`.
8. Si aceptaste una nueva frontera upstream, actualizá docs, state, snapshots y log cuando corresponda.

## Modelo operativo

Estos artefactos cumplen roles distintos:

| Artefacto | Rol |
| --- | --- |
| `policy/maintenance-intent.md` | Intención humana: qué conservar, depurar y proteger |
| `policy/gentle-ai-policy.json` | Policy operativa consumida por la CLI Go y los wrappers |
| `state/upstream-state.json` | Última frontera upstream mantenida |
| `logs/update-log.md` | Bitácora de alto valor de eventos cerrados de mantenimiento |

## Cuándo usar esta guía

- después de actualizar `gentle-ai`
- después de `gentle-ai sync`
- después de una reinstalación por TUI
- cuando el drift upstream puede afectar policy, sanitizador, snapshots o topología de agentes
- cuando `apply-gentle-ai-custom` reporta warnings de topología, recovery de broken state o mismatches de baseline

## Tipos de actualización e impacto en el overlay

| Vía de actualización | Qué cambia | Impacto en el overlay |
| --- | --- | --- |
| `brew upgrade gentle-ai` | Solo el binario | Normalmente no resetea el overlay |
| `gentle-ai sync` | Prompts, skills, MCP configs, assets SDD | Resetea prompts de orchestrators a contenido inline upstream y restaura skills podadas |
| Reinstalación por TUI | Instalación completa, topología, presets y config | Resetea todo y puede cambiar agentes/presets |

### Invariante

Después de cualquier refresh upstream que **no** sea solo `brew upgrade`, hay que reaplicar la capa custom.

Comando mínimo:

```bash
bash apply-gentle-ai-custom.sh opencode
```

Usá este si además querés refrescar skills y wrappers custom en todos los targets soportados:

```bash
bash apply-gentle-ai-custom.sh all
```

`opencode` alcanza para restaurar orchestrators de OpenCode, overrides y materialización de la policy del overlay.

## Flujo completo de mantenimiento

1. Leé la intención en `overlay/gentle-ai/policy/maintenance-intent.md`.
2. Leé la policy operativa en `overlay/gentle-ai/policy/gentle-ai-policy.json`.
3. Leé la frontera upstream en `overlay/gentle-ai/state/upstream-state.json`.
4. Corré `bash audit-gentle-ai-upstream.sh` desde este repo.
5. Revisá:
   - `Summary:`
   - `Drift summary:` cuando exista
   - warnings de invariantes de perfiles o topología
6. Si la auditoría muestra drift relevante para el overlay, actualizá primero este repo:
   - docs
   - policy / intent
   - lógica del sanitizador
   - snapshots / metadata
   - state
   - log, cuando corresponda a un evento cerrado elegible
7. Ejecutá la vía de refresh upstream recomendada:
   - `gentle-ai sync`
   - o reinstalación completa
8. Reaplicá el overlay con `apply-gentle-ai-custom.sh`.
9. Leé el `Summary:` resultante y actuá sobre warnings/errores.
10. Verificá el estado final en disco.
11. Si cambió la frontera upstream aceptada, actualizá el boundary mantenido y commiteá.

## Decision gates

| Situación | Acción |
| --- | --- |
| Upstream cambió pero la topología de agentes no | Preferir `gentle-ai sync` y luego reaplicar |
| Se agregaron, quitaron o renombraron agentes/presets, o cambió la forma upstream | Preferir reinstalación completa y luego reaplicar |
| La auditoría reporta solo drift del prompt base | Revisar el drift antes de adoptar un nuevo baseline |
| Se movieron markers o bloques esperados del sanitizador | Arreglar el sanitizador compartido antes de aplicar |
| Aparecen nuevas skills o comportamiento de workflow upstream | Decidir si pertenecen a la intención keep/prune local antes de cambiar policy |
| Se reportan perfiles SDD locales unmanaged | Decidir si agregarlos al config local o borrarlos manualmente; no se eliminan solos |

## Señales de mayor valor

Estas son las señales que conviene interpretar primero.

| Señal | Significado | Acción |
| --- | --- | --- |
| `base prompt drift: yes` | El `gentle-orchestrator` upstream cambió respecto del baseline auditado | Leer primero `Drift summary:` y recién después inspeccionar el diff completo |
| `profile ... mismatch` / `base asset injection invariant: mismatch` | Upstream cambió la mecánica de generación de perfiles | Frenar y auditar antes de recomendar `sync` |
| `topology: unknown orchestrator matched by prefix only` | Apareció un orchestrator upstream nuevo | Auditarlo y decidir si la policy debe incluirlo explícitamente |
| `topology: expected orchestrator missing from opencode.json` | Un orchestrator conocido desapareció o fue renombrado | Auditar upstream y actualizar policy/intent si hace falta |
| `WARNING - unmanaged SDD profiles left untouched` | Hay agentes gestionables por config local que no están declarados en `~/.config/gentle-ai-custom/opencode-sdd-profiles.json` | Decidir si gestionarlos en el config local o removerlos manualmente |
| `repo snapshots - changed: N > 0` | Cambió el baseline versionado de `gentle-orchestrator` | Revisar `git diff overlay/gentle-ai/snapshots/` |
| `orchestrators recovered from snapshot: N > 0` | Se reconstruyeron prompts faltantes desde snapshots | Investigar por qué faltaban y anotar el recovery |
| `ERROR: audited snapshot metadata mismatch` | Los archivos baseline del repo quedaron inconsistentes entre sí | Reparar baseline/state/metadata antes de continuar |
| `ERROR: audited baseline mismatch for orchestrator 'gentle-orchestrator'` | El prompt materializado ya no coincide con el último baseline auditado | Correr primero la auditoría; adoptar solo después de actualizar snapshot + metadata + state en forma consistente |
| `ERROR: broken state for orchestrator` | `opencode.json` apunta a un overlay file inexistente y no hay snapshot válido para recovery | Resetear con `gentle-ai sync` y volver a correr el apply |

## Verificación post-state

Después del apply, confirmá todo esto:

- las skills podadas ya no existen en cada target configurado
- `agent.general` sigue resolviendo a `openai/gpt-5.4` / `high`
- `agent.explore` sigue resolviendo a `google-vertex/gemini-3.1-pro-preview` / `high`
- cada orchestrator listado por la policy apunta a un prompt `{file:...}` existente
- `overlay/gentle-ai/snapshots/upstream/opencode/orchestrators/` mantiene solo el baseline versionado de `gentle-orchestrator` más su metadata
- `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/` contiene el snapshot operativo de `gentle-orchestrator` y cualquier `sdd-orchestrator-<profile>` gestionado
- si existe `~/.config/gentle-ai-custom/opencode-sdd-profiles.json`, cada perfil declarado mantiene `model` y `variant` correctos en `sdd-orchestrator-<name>` y en los 10 agentes `sdd-<phase>-<name>`

## Config local de perfiles SDD

Los assignments por máquina viven fuera del repo en:

```text
~/.config/gentle-ai-custom/opencode-sdd-profiles.json
```

Reglas operativas:

- si el archivo no existe, el helper deja intactos los perfiles SDD de `opencode.json`
- si el archivo existe, el helper valida estricto y falla cerrado antes de escribir ante JSON/schema inválido
- los perfiles declarados se crean o actualizan en `opencode.json`
- los perfiles existentes no declarados quedan intactos y se reportan como unmanaged
- el repo versionado no debe volver a cargar assignments per-perfil de `model` / `variant`

## Notas técnicas

### Por qué `gentle-ai sync` resetea los prompts de orchestrators

En este setup, `~/.config/opencode/profiles/` está vacío. Entonces la resolución de perfiles upstream **no** preserva las referencias `{file:...}` ya existentes y sync las vuelve a escribir con contenido inline upstream. Reaplicar el overlay restaura la materialización sanitizada basada en archivos.

### Opción de hardening: `external-single-active`

Crear cualquier `*.json` directamente bajo `~/.config/opencode/profiles/` puede hacer que upstream preserve la referencia `{file:...}` actual durante `gentle-ai sync`.

Tradeoffs:

- pro: la restauración del prompt sobrevive a `gentle-ai sync`
- contra: el sistema puede seguir ejecutando indefinidamente una versión sanitizada vieja
- contra: el drift upstream se vuelve más difícil de detectar porque los snapshots dejan de refrescarse naturalmente
- contra: el drift de anchors del sanitizador puede quedar oculto durante mucho tiempo

No lo actives por inercia. El comportamiento default mete más fricción, pero mantiene visible el drift upstream y fuerza a que cada apply sanitice contra el contenido upstream actual.

## Checklist de mantenimiento

- [ ] `maintenance-intent.md` sigue reflejando qué conservar y qué depurar
- [ ] la keep/prune policy sigue alineada con esa intención
- [ ] `upstream-state.json` sigue apuntando a la última frontera upstream realmente mantenida
- [ ] los scripts siguen generando prompt files bajo `~/.config/opencode/prompts/sdd/orchestrators/`
- [ ] los snapshots versionados siguen dejando solo el baseline de `gentle-orchestrator` más metadata
- [ ] los snapshots operativos locales siguen existiendo bajo `~/.config/gentle-ai-custom/opencode-orchestrator-snapshots/`
- [ ] el sanitizador sigue removiendo PR/budget/chained-PR/review-workload sin romper `## Model Assignments`
- [ ] los entrypoints públicos en shell y PowerShell siguen siendo equivalentes
- [ ] los assignments de perfiles SDD siguen siendo locales y no reaparecieron en la policy versionada

## Referencias

- `README.md`
- `AGENTS.md`
- `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`
- `overlay/gentle-ai/policy/maintenance-intent.md`
- `overlay/gentle-ai/policy/gentle-ai-policy.json`
- `overlay/gentle-ai/state/upstream-state.json`
- `overlay/gentle-ai/logs/update-log.md`
