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
- La estrategia vigente es upstream-first: SDD, orchestrator y workflows de PR/chained/work-unit se adquieren desde upstream por defecto, salvo conflicto concreto con una customización local explícita.
- `chained-pr`, `issue-creation` y `work-unit-commits` siguen retenidas como parte de esa estrategia.
- `shared/skills/chained-pr/` y `shared/skills/issue-creation/` instalan copias repo-owned sanitizadas para runtime: preservan la semántica upstream, pero fuerzan resolver `owner/repo`, usar `--repo <owner/repo>` y ejecutar `gh` con argumentos seguros.
- la copia upstream de `branch-pr` sigue podada porque su regla obligatoria de issue-first choca con el flujo local, pero `shared/skills/branch-pr/` la reemplaza con una versión repo-owned sanitizada.
- `hermes-ephemeral-delegation` sigue fuera del runtime OpenCode mantenido y permanece como skill upstream no retenida.
- `shared/skills/` sigue siendo la fuente canónica de las skills repo-owned portables, incluida `judgment-retrospective`.
- `shared/skills/commit-planner/SKILL.md` sigue siendo la capa operativa local para `/commit-plan`, `/commit-apply` y `/commit-fast`; cuando `work-unit-commits` está disponible, usa ese criterio para agrupar commits.
- `shared/skills/branch-pr/SKILL.md` pasa a ser la capa operativa local preferida para `/pr-create` y `/pr-regenerate`; `pr-finalizer` queda como compatibilidad temporal.
- El hook runtime de `judgment-day` vive en `assets/owned/opencode/skills/judgment-day/` y es el que activa automáticamente la retrospectiva terminal.

## Contrato del informe del maintainer

- `Scope`: `Managed` / `Unmanaged`
- `Impact`: `Behavioral` / `Runtime` / `Housekeeping`
- `Decision`: `Adquirir` / `Sanitizar` / `Ignorar`
- Columnas: `Upstream change`, `Files`, `Scope`, `Impact`, `Decision`, `Why`, `Follow-up`
- `Upstream change` tiene que ser un resumen humano corto del delta upstream y de por qué importa; no una lista de rutas de archivos.
- `Follow-up` es opcional; dejalo vacío cuando no haga falta ninguna acción extra.
- `Runtime` incluye wiring, instalación, configuración o materialización del target mantenido.
- `Housekeeping` cubre docs irrelevantes, agentes no relacionados o fixes internos sin efecto en el target mantenido.
- No usar `descartar` como etiqueta principal.

## Regla simple

- Si querés entender o usar el repo, arrancá por `README.md` en la raíz.
- Si estás manteniendo el overlay frente a cambios del upstream, seguí `maintenance.md`.
