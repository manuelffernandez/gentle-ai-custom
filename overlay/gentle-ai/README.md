# Gentle AI overlay assets

Esta carpeta agrupa los assets operativos del overlay de Gentle AI.

## Qué leer primero

- guía principal del repo: `../../README.md`
- guía humana de mantenimiento: `maintenance.md`
- skill operativa del maintainer: `../../.agents/skills/gentle-ai-overlay-maintainer/SKILL.md`

## Qué vive acá

- `maintenance.md` — guía humana centralizada para mantenimiento
- `assets/upstream/` — copias upstream aprobadas para review/diff
- `assets/owned/` — pocos assets realmente propios que `apply-gentle-ai-custom` instala, como `AGENTS.md`
- `patches/` — deltas locales puntuales aplicados sobre un runtime upstream compatible
- `policy/` — intención y policy machine-readable del overlay
- `state/` — frontera upstream mantenida
- `logs/` — historial de decisiones cerradas del overlay

## Modelo vigente

- `policy/managed-assets.json` es el mapa canónico entre upstream aprobado y runtime owned.
- `audit-gentle-ai-upstream` cubre audit + recomendación antes de cualquier mutación.
- Antes de analizar o copiar, `audit` y `sync` exigen que el clon upstream esté totalmente limpio y que `HEAD` coincida con el commit del tag exacto `v<V>` para la versión instalada. No usan tags cercanos ni cambian el checkout automáticamente.
- `sync-gentle-ai-upstream-assets` hace el repo sync de `assets/upstream/...` más la frontera aprobada y guarda la versión, el tag y el commit exactos verificados; no refresca el runtime.
- No haber ejecutado todavía `gentle-ai sync` o una reinstalación es el estado normal antes de auditar. `apply` no necesita el clon upstream y conserva su chequeo separado contra la frontera mantenida.
- `apply-gentle-ai-custom` conserva los prompts, contratos, comandos y plugins SDD materializados por Gentle AI; los snapshots de `assets/upstream/...` no son fuentes de instalación.
- Antes de escribir, `apply` valida solo lo que sus propias operaciones necesitan: límites de ownership, prompts base para perfiles configurados, migraciones seguras de referencias viejas y anchors de Judgment Day.
- La configuración local puede asignar o heredar explícitamente los tres actores globales de Judgment Day y los cinco actores globales de review. Los perfiles nombrados solo pueden crear actores JD con sufijo; review siempre queda global.
- Los actores JD con sufijo se clonan desde el actor upstream global correspondiente. `apply` conserva su estructura y cambia solo la asignación y el permiso del orchestrator nombrado.
- `apply` no interpreta ni certifica la semántica de prompts, contratos o plugins de Gentle AI. Un plugin ausente u opaco no bloquea operaciones custom que no lo usan.
- El estado de ownership registra todos los outputs instalados por esta capa. La limpieza protege árboles devueltos a upstream y targets de parches puntuales, pero no los vuelve a declarar como archivos propios.
- El runtime refresh (`gentle-ai sync` o reinstalación completa) es un paso separado y depende de si el cambio adoptado afecta el runtime target/materialized state que este repo mantiene.
- La estrategia vigente es upstream-first: SDD, orchestrator y workflows de PR/chained/work-unit se adquieren desde upstream por defecto, salvo conflicto concreto con una customización local explícita.
- `chained-pr` y `work-unit-commits` siguen retenidas como parte de esa estrategia.
- `issue-creation` queda fuera del runtime OpenCode por defecto: se poda del runtime y no tiene copia repo-owned instalada.
- `shared/skills/chained-pr/` instala una copia repo-owned sanitizada para runtime: preserva la semántica upstream, pero fuerza resolver `owner/repo`, usar `--repo <owner/repo>`, ejecutar `gh` con argumentos seguros, y no crear branches/PRs desde activación automática sin pedido explícito.
- la copia upstream de `branch-pr` sigue podada porque su regla obligatoria de issue-first choca con el flujo local, pero `shared/skills/branch-pr/` la reemplaza con una versión repo-owned sanitizada.
- `hermes-ephemeral-delegation` sigue fuera del runtime OpenCode mantenido y permanece como skill upstream no retenida.
- `shared/skills/` sigue siendo la fuente canónica de las skills repo-owned portables, incluida `judgment-retrospective`.
- `shared/skills/commit-planner/SKILL.md` sigue siendo la capa operativa local para `/commit-plan`, `/commit-apply` y `/commit-fast`; cuando `work-unit-commits` está disponible, usa ese criterio para agrupar commits.
- `shared/skills/branch-pr/SKILL.md` pasa a ser la capa operativa local preferida para `/pr-create` y `/pr-regenerate`; `pr-finalizer` queda como compatibilidad temporal.
- El handoff de `judgment-day` vive en `patches/opencode/judgment-retrospective.json`; se aplica solo si el skill upstream instalado tiene los anchors exactos esperados.
- Las referencias SDD siguen en `assets/upstream/` para auditoría, pero Gentle AI las instala y materializa en runtime.

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
