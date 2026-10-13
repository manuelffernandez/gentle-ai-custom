# gentle-ai-custom

Los assets aprobados de Gentle AI se guardan como copias de auditoría. Gentle AI instala y prepara los archivos reales del runtime; esta capa los conserva y aplica solamente los ajustes locales explícitos definidos en `maintenance-intent.md`.

Una capa de customización mantenida con IA para [Gentle AI](https://github.com/Gentleman-Programming/gentle-ai): extiende la experiencia base con skills, prompts, wrappers y política operativa propios, y hace más mantenible el trabajo de reaplicar y auditar lo que `gentle-ai sync` vuelve a materializar.

## Qué es Gentle AI

[Gentle AI](https://github.com/Gentleman-Programming/gentle-ai) es el proyecto original de Gentleman Programming para mejorar la experiencia de desarrollo con IA mediante agentes, skills, orquestación y herramientas para trabajar con asistentes en el código.

Este repo existe sobre esa base, no en reemplazo. La idea es dar el crédito que corresponde al proyecto upstream y, al mismo tiempo, construir una capa más mantenible para workflows concretos.

## Por qué existe este repo

Gentle AI resuelve gran parte de la experiencia, pero hay decisiones del upstream que no siempre encajan con mi flujo diario. Por eso este repo actúa como una **capa de customización mantenida con IA**:

- conserva lo mejor del upstream
- depura lo que no se adapta a mi forma de trabajar
- agrega skills y wrappers que sí uso todos los días
- convierte una customización profunda de Gentle AI en algo más mantenible a través de automatización, auditoría y agentes

Hoy sigue estando orientado principalmente a mi flujo de trabajo. La idea es seguir iterándolo para que, con el tiempo, resulte más simple de adaptar y usar también en otros contextos.

## Visión

La dirección de este repo es seguir mejorando la experiencia con una capa cada vez más amigable: un instalador TUI más personalizado, mejor ergonomía operativa y más posibilidades de expansión para skills, overlays y flujos de trabajo reales.

## Objetivo

Hoy este repo funciona como una **capa unificada de personalización y mantenimiento** sobre Gentle AI. En lugar de modificar el proyecto original a mano cada vez que se actualiza, este repositorio automatiza la adaptación del entorno para que se ajuste a un flujo de trabajo específico.

Su meta principal es permitir consumir las actualizaciones del proyecto original (`gentle-ai sync` o reinstalaciones) sin perder las configuraciones propias, manteniendo una base segura y auditable. El detalle técnico de las modificaciones exactas que se aplican (qué skills se conservan o podan, y cómo se altera el orquestador) se encuentra documentado en la sección **Política actual**, más abajo en este documento.

## Por qué Go

La automatización principal vive en Go porque permite mantener **un solo lugar de verdad** para la lógica compartida entre los entrypoints `.sh` y `.ps1`. En vez de duplicar comportamiento entre Bash y PowerShell, ambos wrappers delegan en la misma CLI.

Además, Go ya forma parte natural del stack porque es una dependencia directa de [Engram](https://github.com/Gentleman-Programming/engram). Reutilizarlo acá simplifica el ecosistema, reduce drift entre plataformas y hace más sostenible la evolución del overlay.

## Agentes soportados

- `opencode` → `~/.config/opencode`
- `codex` → `~/.codex`

## Uso

Para el uso normal, usá el entrypoint de apply y elegí el target que querés actualizar.

Los entrypoints públicos son estos:

- `apply-gentle-ai-custom.sh`
- `apply-gentle-ai-custom.ps1`

Los targets disponibles son `opencode`, `codex` y `all`. `all` aplica ambos targets. Elegí un solo target si querés mantener el efecto acotado.

Los scripts `audit-gentle-ai-upstream.*` y `sync-gentle-ai-upstream-assets.*` existen, pero no están pensados como punto de entrada normal para una persona. Su uso recomendado es a través del agente en modo mantenimiento; salvo debugging puntual o un caso excepcional, no hace falta correrlos manualmente.

Todos estos wrappers delegan en la CLI Go compartida (`go run ./cmd/gentle-ai-overlay ...`) y no duplican lógica entre shell y PowerShell.

### Uso rápido

### Linux / macOS

```bash
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh opencode
# o solo para Codex:
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh codex
# ambos targets:
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh all
```

Opcional: agregá `--verbose` si querés ver el detalle de archivos tocados.

### Windows (PowerShell 5.1+)

> **Requisito previo — política de ejecución:**
>
> ```powershell
> Set-ExecutionPolicy -Scope Process Bypass
> ```

```powershell
~\Documentos\gentle-ai-custom\apply-gentle-ai-custom.ps1 opencode
```

Opcional: agregá `--verbose` si querés ver el detalle de archivos tocados.

El target `opencode` modifica solo OpenCode; `codex` modifica solo Codex. `all` selecciona ambos y requiere autorización para el apply de los dos.

## Mantenimiento del overlay

El mantenimiento en este repositorio sigue siempre los mismos pasos. Lo único que cambia es decidir si las novedades del proyecto original (upstream) necesitan que actualicemos nuestra capa antes de aplicarlas.

### Pasos recomendados

1. Actualizá el ejecutable de `gentle-ai`.
2. Actualizá los tags en tu carpeta local de `gentle-ai` y dejala en el tag exacto que querés revisar.
3. Desde `gentle-ai-custom`, pedile al agente de mantenimiento que revise qué cambió. Apenas carga sus instrucciones, debe ejecutar la auditoría antes de explorar muchos archivos o delegar lecturas. Por defecto revisa el tag de la versión instalada. Si querés revisar otra versión sin actualizar el ejecutable que usan sesiones abiertas, indicá el tag exacto.
4. Antes de cambiar cualquier archivo acá, el agente te tiene que dar un resumen claro: qué hay de nuevo, qué sugiere `Adquirir`, `Sanitizar` o `Ignorar`, por qué, y qué comando recomienda para actualizar.
5. Confirmá por separado qué autorizás: cambios en este repo, copia de snapshots, actualización del runtime y reaplicación del overlay no son el mismo permiso.
6. Una auditoría con `--release` es solo una vista previa: no adopta esa versión ni autoriza la copia al repo. Si después aceptaste esa versión y autorizaste la copia, el agente puede ejecutar `sync-gentle-ai-upstream-assets.sh --release v<V>` con el mismo tag.
7. El agente te recomienda `gentle-ai sync` o una reinstalación y te muestra los comandos exactos. Por defecto los ejecutás vos; el agente solo puede hacerlo si le pedís de forma explícita esa acción concreta.
8. Volvé a aplicar nuestra capa con `apply-gentle-ai-custom.sh`. El agente también necesita una autorización explícita separada para ejecutarlo.
9. Terminá con una revisión general para asegurar que todo quedó bien y dejá un resumen final de lo que se adquirió, sanitizó o ignoró.
10. Reiniciá OpenCode si el archivo `~/.config/opencode/opencode.json` tuvo algún cambio.

Es normal no haber ejecutado todavía `gentle-ai sync` ni una reinstalación: esos pasos vienen después de la auditoría y de tu aprobación. La opción `--release v<V>` no actualiza el ejecutable ni el runtime; solo elige qué tag exacto se revisa. El tag debe existir, estar limpio y coincidir con `HEAD`. El modo habitual sin esa opción sigue exigiendo que coincida con la versión instalada.

### Estado actual: snapshots v4 adoptados; runtime sin verificar

La frontera de snapshots de este repo ya está en Gentle AI v4.0.0, commit `ff77164d4f56f1665b22fb6fac51c2ccbb769400`. La sincronización aprobada informó 11 archivos nuevos, 13 actualizados, 31 sin cambios y 35 retirados, en 19 carpetas. OpenCode y Codex no se actualizaron: sus versiones de runtime siguen sin verificar.

La auditoría posterior a la sincronización pasó: no detectó diferencias administradas ni no administradas vigiladas, y todos los controles estructurales informaron `ok`. La limpieza afectó solo a los snapshots declarados por el repo; no cambió archivos ni configuración del runtime.

El estado de snapshots v3.7.0 que aparece a continuación es histórico. La versión mantenida actualmente es v4.0.0; cualquier refresh o apply del runtime necesita autorización separada.

#### Estado histórico de snapshots v3.7.0

En una adopción anterior, el repo avanzó sus snapshots a Gentle AI v3.7.0, commit `6dee8f833aec9e46015759c5065a9035795d9af1`. La sincronización informó 5 archivos nuevos, 31 actualizados, 43 sin cambios y ninguno eliminado, en 19 carpetas.

La auditoría posterior pasó para ese commit. Esa evidencia corresponde a la adopción histórica de snapshots; no describe la frontera actual ni verifica el runtime materializado o Codex.

No se verificó la versión del runtime materializado ni la de Codex. La versión 2.9.1 corresponde a evidencia histórica, no a un estado actual confirmado. OpenCode 1.18.32 fue la versión en una verificación anterior; los cinco plugins V2 siguen siendo solo snapshots de auditoría y sincronización, y este repo no declara soporte para instalarlos. No se autorizó ni ejecutó `gentle-ai sync`, reinstalación ni `apply` como parte de la adopción.

#### Commits ODD pendientes de propagación al runtime

El workflow upstream `work-unit-commits` exige cerrar cada tarea ODD con un commit. Los parches locales que anulaban esa regla ya se quitaron del repo para OpenCode y Codex. El runtime no se actualizó ni verificó: todavía hace falta autorizar por separado el refresh correspondiente y el `apply` del target. Esta regla de cierre no autoriza push ni PR.

El snapshot v3.7.0 incorporaba el contrato output-only de `sdd-research`; esa evidencia y el retiro de su parche local son históricos. El handoff local de Judgment Day se mantiene.

#### Target Codex

`apply-gentle-ai-custom.sh codex` instala las siete skills propias reutilizando sus fuentes en `shared/skills/`; reemplaza solo las skills con el mismo nombre que estén en esa lista y no poda otras skills upstream. También agrega reglas locales puntuales después del marcador de Engram en `~/.codex/AGENTS.md`, cambia solo las filas de `odd-explorer` y `odd-worker` de `gpt-6.1-luna` a `gpt-6.1-sol` (ambas en `high`), y aplica el parche exacto de retrospectiva de Judgment Day.

La capa no reemplaza el archivo completo `AGENTS.md` ni modifica las otras filas ODD, Judgment Day o review. Tampoco poda skills upstream ajenas a la lista local, genera comandos ni cambia `config.toml`, Engram o perfiles de modelos. Si falta, cambia o se duplica una fila esperada, se detiene antes de escribir. Este ajuste está solo en el repo: no se ejecutó `apply`, y no se comprobó si `gpt-6.1-sol` está disponible en la cuenta. La auditoría actual no verifica el runtime de Codex; su comportamiento queda sin confirmar hasta que se autorice y revise un apply de Codex.

Si quedó una segunda copia antigua del bloque local entre el final del orquestador y el inicio de CodeGraph, el próximo `apply` autorizado quitará solo esa copia exacta con la regla que prohibía commits automáticos. Conserva el bloque actual junto a Engram y el resto del archivo. Si la copia antigua está incompleta, cambiada, duplicada o fuera de lugar, se detiene antes de escribir en cualquiera de los targets seleccionados. Esta migración también está solo en el repo: no cambia el Codex instalado hasta que se autorice `apply` por separado.

El repo aceptó de forma consciente una conducta de archivo SDD en v3.7.0. Esa decisión es histórica: el baseline actual es ODD sin SDD y esto no prueba ni cambia el estado de la revisión RDD escalada por separado.

#### Retiro de la regla de autorización remota

El repo conserva intacta la copia upstream para auditoría, pero ya no instala esa regla en los prompts OpenCode administrados. Al ejecutar `apply`, se elimina únicamente una sección reconocida: la versión upstream o la versión local relajada que se había aplicado antes. Si encuentra una sección duplicada, incompleta o desconocida, o una ruta insegura, se detiene antes de escribir. El resto del contenido y los permisos de los archivos se conservan.

El `AGENTS.md` propio tampoco incluye esa regla. Esto no modifica la protección separada contra la lectura de archivos de credenciales ni otros permisos del runtime. El cambio afecta solo al repo: el runtime instalado requiere un `apply` autorizado por separado. Esta sesión no ejecutó `gentle-ai sync`, una reinstalación ni `apply`.

#### Evidencia histórica del runtime v2.9.1 y v2.5.0

En el cierre anterior de v2.9.1, la auditoría pasó para `9ec0cf443622f20fe511815ff5a83088c7467bff`, y una verificación de solo lectura confirmó 602 comprobaciones sobre asignaciones, perfiles, prompts, permisos, alcance de tareas y referencias. También confirmó el orchestrator con consentimiento compatible, las 7 skills propias, 5 wrappers, el `AGENTS.md` propio, la poda de skills y los parches locales. Esa evidencia corresponde al runtime v2.9.1 y no verifica un runtime v3.7.0.

#### Evidencia histórica del runtime v2.5.0

En la etapa histórica v2.5.0, la policy y la auditoría incorporaban la fase `sdd-research` y perfiles de 11 fases. La instalación completa del componente SDD y la posterior ejecución autorizada de `bash apply-gentle-ai-custom.sh opencode` terminaron correctamente entonces; esa no es la policy actual.

El `apply` actualizó la configuración, podó 2 skills, cambió y verificó 1 asset propio, no informó warnings de topología y administró 4 perfiles. En los agentes informó 3 creados, 4 actualizados y 49 sin cambios, sin perfiles no administrados. La verificación final pasó la auditoría canónica, `go test ./...`, `git diff --check` y la carga resuelta de los agentes `sdd-research`.

En esa verificación histórica, la copia sanitizada y propia de `branch-pr` estaba instalada; las copias conflictivas upstream de `branch-pr`, `issue-creation` y Hermes no estaban. Las 7 skills propias, los 5 wrappers, el `AGENTS.md` propio, los overrides y los 4 perfiles administrados coincidían con sus fuentes o con la configuración local. `general` usaba `openai/gpt-5.6-terra` / `high` y `explore` usaba `openai/gpt-5.6-sol` / `medium`. Los cuatro perfiles tenían 11 fases en orden y existían los 3 actores nombrados de `sdd-research`. Los actores de review eran globales, sin copias con sufijo, y Judgment Day difería del upstream solo por el parche de retrospectiva. Esto no verifica el runtime actual.

En esa verificación histórica del runtime v2.5.0, la intención y la política efectiva de background estaban en `on`, y el launcher era `~/.gentle-ai/bin/opencode`. Los procesos abiertos desde Desktop o con `opencode` normal seguían en foreground. También se observó que `~/.gentle-ai/state.json` informaba `installed_binary_version: 2.4.0`; no se conoce la causa y no se modificó ese estado externo.

La política del repositorio considera RDD activo y funcional por decisión explícita del propietario, basada en validaciones operativas realizadas en otros repositorios. Esta decisión no cambia el runtime: en esta máquina se observó el modo global en `on` y sin ajuste local para este clon. El usuario excluyó de forma explícita este mantenimiento de v2.9.1 de RDD. No se autorizó ni realizó ningún commit, push o PR para este cierre documental.

### Historial: adaptación del repositorio para v2.9

La auditoría entiende la nueva composición del orchestrator sin depender del número de versión: sigue la cadena real desde la asignación del prompt hasta el asset OpenCode y rechaza asignaciones sueltas o señuelos. También vigila el contrato upstream de autorización remota y comprueba que llegue a los prompts de OpenCode que este repo mantiene.

En los snapshots v3.7.0, el skill y el comando `sdd-research` indicaban que la recolección era output-only y que el orchestrator se ocupaba de la persistencia. Esta evidencia es histórica; no describe el runtime actual.

La adopción de v2.9.1 incorpora desde upstream el consentimiento de revisión compatible y la migración de prompts preservados para OpenCode. Los cambios de autoridad de despacho derivados de transcripciones de Claude Code quedan fuera del alcance OpenCode mantenido por este repo. No fue necesario cambiar la política, los parches, los perfiles ni la topología.

La copia y el avance de `upstream-state.json` a v2.9.1 terminaron con autorización separada. Después, el usuario completó manualmente `gentle-ai sync` y el `apply` del overlay. La verificación independiente de solo lectura pasó; no queda una actualización del runtime pendiente para este cierre.

Antes de explorar el repo o delegar el análisis, la auditoría informa la versión instalada y verifica el release elegido: por defecto `v<versión instalada>`, o el tag indicado con `--release v<V>`. Comprueba que la carpeta original esté totalmente limpia, que exista ese tag exacto y que `HEAD` apunte al commit del tag. También revisa cambios dentro de submódulos. Si algo no coincide, se detiene sin delegar nada y explica cómo actualizar tags o cambiar al tag correcto de forma manual. Nunca descarga tags ni cambia de versión por su cuenta. Una auditoría de otro release es solo una vista previa: no significa que se aceptó ni que se autorizó copiarlo al repo. `last_maintained_commit` sirve para elegir desde dónde comparar cambios; no demuestra esta coincidencia.

Para encontrar el clon upstream durante la auditoría o la copia de snapshots, solo lee `upstream_repo_path` del archivo local. Así puede auditar una nueva fase antes de que actualices las asignaciones de modelos de cada perfil. La validación completa de esos perfiles sigue siendo obligatoria antes de que `apply` escriba en el runtime.

Esto también vale para versiones preliminares: solo se aceptan si existe su tag exacto. No alcanza con estar cerca de un tag, en una rama que lo contiene o en un commit posterior. El mismo control se repite antes de copiar snapshots con `sync-gentle-ai-upstream-assets`; la copia sigue necesitando autorización separada y, para un release alternativo, usa el mismo `--release v<V>`.

```bash
brew upgrade gentle-ai
git -C /path/to/gentle-ai pull

# dejá el upstream en el tag exacto que vas a revisar
# Reemplazá <V> por el número exacto de la versión elegida.
bash ~/Documentos/gentle-ai-custom/audit-gentle-ai-upstream.sh --release 'v<V>'

# solo después de aprobar la adopción y autorizar la copia al repo
bash ~/Documentos/gentle-ai-custom/sync-gentle-ai-upstream-assets.sh --release 'v<V>'

# acción del runtime: la ejecuta la persona, salvo pedido explícito al agente
gentle-ai sync

# autorización separada: actualizar el runtime no autoriza este paso
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh opencode

# o el equivalente para todo (hoy opencode es el único agente soportado)
# bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh all
```

### Qué tiene que decidir la auditoría

Antes de tocar cualquier archivo, la auditoría tiene que responder esto de forma simple y fácil de leer:

| Eje | Valores |
| --- | --- |
| `Scope` | `Managed` / `Unmanaged` |
| `Impact` | `Behavioral` / `Runtime` / `Housekeeping` |
| `Decision` | `Adquirir` / `Sanitizar` / `Ignorar` |

- `Adquirir`: incorporar el cambio al overlay/runtime.
- `Sanitizar`: adaptar el cambio para que siga mandando `maintenance-intent`.
- `Ignorar`: se evaluó, pero no aplica al alcance mantenido.
- El informe debe usar estas columnas: `Upstream change`, `Files`, `Scope`, `Impact`, `Decision`, `Why`, `Follow-up`.
- `Upstream change` tiene que ser un resumen humano corto del delta upstream y de por qué importa; no una lista de rutas de archivos.
- `Follow-up` es opcional; dejalo vacío cuando no haga falta ninguna acción extra.
- Cada fila representa una sola decisión. `Decision` aplica exactamente al cambio descrito en `Upstream change`.
- Una adaptación local distinta va en `Follow-up` o en otra fila. Una sanitización del runtime siempre tiene su propia fila.
- Cada ajuste necesario mencionado en el resumen debe aparecer en una fila.
- No mezclar filas cuando cambien `Scope`, `Impact` o `Decision`.
- Antes de mostrar el informe, hacer coincidir el resumen, las filas y la recomendación final.
- `Runtime` incluye wiring, instalación, configuración o materialización del target mantenido, aunque no cambie el comportamiento de los agentes.
- `Housekeeping` cubre documentación irrelevante, agentes no relacionados o fixes internos sin efecto en el target mantenido.
- No usar `descartar` como etiqueta principal del informe.

Hoy en día, la auditoría busca diferencias usando comandos como `git diff` y las filtra para separar lo que sí importa de lo que no, pero sin dejar de revisar cosas importantes de estructura.

Los scripts `audit-gentle-ai-upstream` y `sync-gentle-ai-upstream-assets` se pueden usar a mano, pero es mejor pedirle al agente de mantenimiento que los use. Así te puede explicar qué cambió y pedirte permiso antes de tocar nada. El comando `sync-gentle-ai-upstream-assets` solo copia cosas aprobadas a este repositorio; no actualiza la configuración de tu computadora.

La actualización del runtime sigue siendo tuya por defecto. Aprobar la auditoría, aceptar cambios upstream, copiar snapshots al repo o permitir ediciones del overlay NO autoriza al agente a ejecutar `gentle-ai sync` ni una reinstalación. También tenés que autorizar por separado cualquier ejecución de `apply-gentle-ai-custom`.

El comando `apply-gentle-ai-custom` sigue siendo independiente de la carpeta original: compara la versión instalada con la versión mantenida por este overlay y no necesita encontrar el clon de upstream.

El `AGENTS.md` propio de OpenCode sigue siendo un archivo completo que instala este repo. Su persona y el protocolo completo de Engram para OpenCode deben coincidir con los snapshots aprobados. Codex mantiene su archivo completo bajo control de upstream y recibe un bloque local delimitado y el cambio exacto de dos filas ODD para `odd-explorer` y `odd-worker`; el bloque incluye alcance de implementación, comunicación clara, estándar de Markdown para agentes, seguridad de credenciales y el ajuste condicional para Gemini. Ninguno de los dos archivos propios incorpora la regla upstream de autorización remota.

La skill `notion-project-knowledge` se instala para ambos targets desde `shared/skills/notion-project-knowledge/SKILL.md`. Consulta las guías actuales de Notion cuando necesita ese conocimiento y ayuda al agente principal a conservar conocimiento útil del proyecto y evidencia de actividad personal en Me, dentro del alcance autorizado. Ambos targets reciben un bloque de activación separado. Agregarlo al repo no actualiza el runtime ni escribe en Notion; hace falta un apply autorizado por separado. El registro nativo no busca en `shared/skills/`, por lo que la nueva skill no aparece allí antes de instalarla.

### Cuándo usar `gentle-ai sync` y cuándo reinstalar

- Es preferible usar `gentle-ai sync` si los cambios que adoptamos sirven y el target mantenido conserva una estructura compatible.
- Conviene hacer una reinstalación completa cuando los cambios del proyecto original afectan la topología o materialización del target mantenido, o cuando un simple `gentle-ai sync` ya no alcanza para prepararlo.
- Si el proyecto original agrega soporte para un nuevo agente o plataforma que no usamos, no hace falta reinstalar.
- Las reglas upstream sobre estrategia de PR, PRs encadenadas, presupuesto de revisión y commit ODD al cierre de tarea se adoptan por defecto. Los parches locales que anulaban el commit se quitaron del repo para ambos targets. El runtime todavía no se actualizó ni verificó; el commit de cierre no autoriza push ni PR.

### Qué hace el comando `apply-gentle-ai-custom`

- instala nuestras skills propias para el target elegido; los comandos personalizados solo se generan para OpenCode.
- borra las skills que no queremos usar solo en OpenCode y únicamente para los targets seleccionados; Codex conserva sus skills upstream.
- bloquea en OpenCode cada skill podada que no tenga un reemplazo propio con el mismo nombre mediante `permission.skill`, incluso si OpenCode encuentra una copia en `~/.agents/skills` o `~/.claude/skills`. No borra ni modifica esas copias externas, y conserva los permisos de skills no podadas ni de reemplazos propios como `branch-pr`.
- aplica nuestras preferencias de modelos de IA si las configuramos.
- comprueba solo los archivos, referencias y anchors que necesita para las operaciones solicitadas o los parches locales declarados.
- aplica el parche puntual de retrospectiva de Judgment Day y, solo en Codex, las dos filas ODD declaradas junto al bloque local de instrucciones. Los parches que anulaban el commit ODD se quitaron para ambos targets; el runtime aún no se actualizó ni verificó.
- conserva los prompts y plugins preparados por Gentle AI; no los reemplaza con plantillas fuente. Los snapshots mantenidos están en v4.0.0, pero el runtime materializado no está verificado. La política acepta RDD como activo y funcional, pero su estado efectivo se consulta por máquina o clon.
- aplica overrides compatibles sin crear perfiles o agentes SDD. Si encuentra opciones SDD heredadas, se detiene antes de escribir y las deja intactas para revisión manual.
- agrega de forma puntual el handoff de `judgment-retrospective` al skill `judgment-day` compatible.
- mantiene los nombres SDD fuera de la lista de skills conservadas y los poda en OpenCode si aún existen. No borra ni migra configuración o archivos locales por esta actualización documental; cualquier runtime apply requiere autorización independiente.
- al podar skills, muestra solo las que eliminó en esta ejecución; si una ya no estaba, no la menciona. La poda y el bloqueo de permisos para las skills SDD siguen activos aunque no haya nada que eliminar.
- registra todos los archivos que instala como propios. Al limpiar registros viejos, conserva los archivos devueltos a Gentle AI y los targets de parches puntuales sin declararlos como propios.

`opencode` y `codex` limitan el apply a un target. `all` selecciona los dos. Codex no altera permisos, configuración, pruning ni comandos de OpenCode.

La configuración para elegir qué modelos de IA vas a usar o dónde está clonado el proyecto original se guarda en un archivo solo tuyo en tu computadora: `~/.config/gentle-ai-custom/opencode-local-config.json`. Más abajo se explica cómo armarlo.

### RDD: revisión y recibos de entrega

Receipt-Driven Development (RDD) revisa un candidato ya terminado y genera un recibo ligado a ese contenido. Los controles de entrega pueden reutilizar el mismo recibo. No significa Test-Driven Development (TDD): los tests siguen siendo parte de la implementación y la verificación.

Para ver o cambiar el modo nativo:

```bash
gentle-ai review mode status --cwd .
gentle-ai review mode disable --cwd .
gentle-ai review mode enable --cwd .
```

- `status` solo informa; no cambia nada.
- `disable` y `enable` actúan de forma global por defecto.
- Con `--scope clone --cwd <ruta>` podés desactivar RDD solo en un clon o quitar esa desactivación local.
- Si el valor global o el del clon está en `off`, RDD queda apagado. Un clon no puede encenderlo por encima de un `off` global y su ajuste no pasa a otros clones.
- Volver a habilitarlo se aplica a candidatos nuevos desde ese momento. No aprueba, repara ni reanuda automáticamente una autoridad vieja inválida o pendiente. El contenido creado mientras RDD estaba apagado no queda varado: después de habilitarlo, el contenido actual puede convertirse en un candidato nuevo y pasar por una revisión nueva.

SDD y RDD son conceptos distintos. En general, SDD organiza la planificación y la implementación, mientras RDD revisa el candidato terminado. Este repo mantiene el flujo ODD de v4 sin SDD; si RDD está apagado, no se deben lanzar revisores y la entrega sigue la política normal sin inventar un recibo.

RDD está aceptado como activo y funcional por decisión explícita del propietario, basada en validaciones operativas realizadas en otros repositorios. El estado efectivo pertenece al usuario y a cada máquina o clon; Gentle AI upstream no lo controla de forma remota. Consúltalo siempre con `gentle-ai review mode status --cwd .`. En esta máquina se observó `on` por la configuración global y ningún ajuste local para este clon.

Como contexto histórico, Gentle AI v2.2.2 tuvo un bug confirmado en la entrega del alcance de una corrección. El código y los tests de v2.5.0 aportaron el plan previo a la edición, y este repositorio evaluó pero aplazó un parche local. Ese análisis se conserva en `overlay/gentle-ai/maintenance.md`. El canary con corrección puede seguir usándose para diagnosticar regresiones, pero ya no bloquea la activación. Siguen siendo obligatorios el interruptor controlado por el usuario, los recibos ligados al candidato exacto, los límites de corrección y la política normal de entrega.

En la verificación histórica de v2.5.0, el runtime pasó la instalación completa del componente SDD y el `apply` autorizado por separado. En ese momento, la intención y la política efectiva de background estaban en `on`, y `~/.gentle-ai/bin/opencode` era el launcher observado para sesiones CLI con background; Desktop y el comando `opencode` normal iniciaban procesos en foreground. Estos datos no describen el runtime, el modo de background ni el launcher actuales, que no están verificados. Un futuro `gentle-ai sync` o una reinstalación requeriría un `apply` autorizado por separado. La aceptación de RDD no autoriza esas acciones ni modifica la configuración.

### Chequeo previo de versiones

Antes de modificar nada, el comando `apply-gentle-ai-custom` revisa la versión mantenida y también valida el runtime instalado.

- si son idénticas -> sigue adelante.
- si son distintas o no se sabe -> te avisa primero.
- si lo estás corriendo a mano, podés confirmar y seguir.
- si se ejecuta de forma automática y hay diferencias -> frena todo por seguridad.
- si faltan los prompts base necesarios para los perfiles configurados, una referencia vieja no se puede migrar con seguridad o no coinciden los anchors de algún parche declarado -> frena sin escribir y pide ejecutar primero `gentle-ai sync` o reinstalar.
- el plugin de review puede faltar o tener cualquier contenido: `apply` no lo interpreta ni lo certifica porque pertenece a Gentle AI.

Dónde busca el proyecto original:

1. En la ruta `upstream_repo_path` configurada en tu `opencode-local-config.json`.
2. En la variable de entorno `GENTLE_AI_CUSTOM_UPSTREAM_REPO`.
3. Por defecto, asume que está en `../gentle-ai`, al lado de esta misma carpeta.
4. Si no lo encuentra en ningún lado, te tira un error claro.

Si falta algo en tu configuración:

- si no ponés `agent_overrides`, no cambia los modelos que vienen por defecto.
- no se generan agentes ni perfiles SDD; las configuraciones SDD locales anteriores se rechazan sin modificarlas.
- si omitís un actor dentro de `judgment_day` o `review`, conserva la asignación que ya tenga el runtime. Para heredar dinámicamente el modelo y variant del orchestrator que delega, usá `{"inherit": true}`.

### Archivos clave del mantenimiento

| Archivo | Para qué sirve |
| --- | --- |
| `overlay/gentle-ai/policy/maintenance-intent.md` | Es el documento donde explicamos claramente nuestra intención: qué cosas queremos conservar, cuáles quitar y cómo queremos que funcione todo. |
| `overlay/gentle-ai/policy/gentle-ai-policy.json` | Son las reglas exactas que leen los scripts cuando se ejecutan. |
| `overlay/gentle-ai/policy/managed-assets.json` | Es el mapa oficial que separa snapshots de auditoría, assets propios instalables y parches locales puntuales. |
| `overlay/gentle-ai/state/upstream-state.json` | Guarda cuál fue la última versión del proyecto original que revisamos. |
| `overlay/gentle-ai/logs/update-log.md` | Es un registro donde anotamos las decisiones importantes sobre qué cosas nuevas aceptamos o rechazamos. |

Para ver el código exacto que cambió, siempre podés mirar el historial de Git. El archivo `update-log.md` lo usamos solo para anotar las decisiones y eventos importantes que cambian cómo nos alineamos con el proyecto original.

### Carpetas y archivos importantes

| Ruta | Para qué sirve |
| --- | --- |
| `overlay/gentle-ai/assets/` | Contiene las copias aprobadas del proyecto original para auditoría y los pocos archivos que este repo realmente posee, como el `AGENTS.md` completo. Sus secciones heredadas se comparan con los snapshots aprobados mediante tests. |
| `overlay/gentle-ai/patches/` | Define ajustes locales pequeños y verificables que se aplican sobre un runtime upstream compatible. |
| `~/.config/gentle-ai-custom/opencode-local-config.json` | Tu configuración personal: dónde está el código original, qué modelos preferís usar, etc. |
| `overlay/gentle-ai/maintenance.md` | Guía y notas técnicas pensadas para que las lea una persona que esté manteniendo esto. |
| `.agents/skills/gentle-ai-overlay-maintainer/SKILL.md` | Instrucciones para que el agente entienda cómo hacer la auditoría y guiarnos en el mantenimiento. |

### Herramientas de mantenimiento incluidas

- **Agente de mantenimiento (Maintainer skill)**: es la forma recomendada de trabajar, pidiéndole al agente que nos asista.
- **Documentación de mantenimiento**: concentra las explicaciones y notas técnicas.
- **Scripts de auditoría**: se pueden usar a mano, pero normalmente no hace falta salvo para buscar algún problema puntual.

Toda la guía para humanos está en `overlay/gentle-ai/maintenance.md`. El comportamiento del agente está en su propio `SKILL.md`.

> **Nota sobre OpenCode:** si el script llega a cambiar tu `~/.config/opencode/opencode.json`, tenés que reiniciar OpenCode porque no toma los cambios al vuelo.

## Política actual

Esta sección resume la policy en lenguaje humano. El detalle más preciso, técnico y normativo vive en `overlay/gentle-ai/policy/maintenance-intent.md`.

La idea general ahora es **upstream-first**: esta capa arranca desde el comportamiento actual de Gentle AI v4, conserva unas pocas personalizaciones locales explícitas y evita seguir cargando parches viejos si ya no hacen falta. El foco está en mantener un overlay chico y entendible.

### Skills del overlay

| Skill                  | Estado      | Qué hace                                                                       | Por qué se queda / se va                                                                                                                     |
| ---------------------- | ----------- | ------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `_shared`              | Se conserva | Expone referencias compartidas para skills upstream.                            | Sostiene las skills conservadas sin imponer proceso extra.                                                                                  |
| `cognitive-doc-design` | Se conserva | Ayuda a escribir documentación más clara y escaneable.                         | Mejora la calidad de los artifacts sin meter gobernanza de repo.                                                                             |
| `comment-writer`       | Se conserva | Ayuda a redactar comentarios de colaboración.                                  | Suma claridad comunicacional sin forzar una estrategia de PR.                                                                                |
| `gentle-ai-bench`      | Snapshot retenido | Documenta cómo crear y verificar journeys del bench upstream.              | Se guarda para auditoría, pero este repo no la instala como propia ni obliga a usarla.                                                        |
| `go-testing`           | Se conserva | Da patrones prácticos para tests en Go.                                        | Aporta calidad técnica directa.                                                                                                              |
| `judgment-day`         | Se conserva | Hace revisión adversarial/dual review y dispara la retrospectiva automática al cerrar desde el hook runtime del overlay. | Refuerza criterio técnico, valida de verdad y deja aprendizaje reutilizable.                                                                |
| `judgment-retrospective` | Se agrega   | Resume el juicio, guarda patrones reutilizables e historial de intervenciones.  | Diseñada para ser invocada por el hook runtime de `judgment-day`; también se puede cargar manualmente. Evita guardar el output crudo y ayuda a ver si una corrección realmente funcionó. |
| `rdd-defect-workflow`  | Snapshot retenido | Describe el trabajo upstream sobre defectos RDD.                           | Se guarda para auditoría y no se instala como skill propia; el modo efectivo de RDD se decide fuera de este snapshot.                         |
| Skills SDD             | Se podan   | Los nombres SDD ya no forman parte de la superficie mantenida para OpenCode.   | El baseline v4 usa ODD; las opciones y archivos SDD de la máquina siguen siendo propiedad del usuario.                                      |
| `skill-creator`        | Se conserva | Ayuda a crear nuevas skills.                                                   | Hace extensible la capa custom.                                                                                                              |
| `skill-improver`       | Se conserva | Ayuda a revisar y mejorar skills existentes.                                   | Permite evolucionar tooling propio con criterio.                                                                                             |
| `skill-registry`       | Se conserva | Indexa las skills disponibles y sus triggers.                                  | Hace mantenible el ecosistema de skills.                                                                                                     |
| `systemic-issue-triage` | Snapshot retenido | Agrupa issues upstream por causa raíz.                                     | Se guarda para auditoría, pero no se fuerza como selección del runtime ni se instala como propia.                                             |
| `branch-pr`            | Se sanitiza | La copia upstream se poda, pero se reinstala como skill repo-owned para crear o regenerar PRs desde cambios ya commiteados. | Conserva el nombre y lo útil del flujo de PR, pero elimina el issue-first universal: issue link, labels y checks quedan como policy contextual del repo, no como regla global. |
| `chained-pr`           | Se retiene + se sanitiza en runtime | El snapshot upstream sigue auditado, pero `apply` instala una copia repo-owned sanitizada para PRs encadenadas o stacked. | Conserva el workflow upstream y agrega protecciones operativas: resolver `owner/repo`, usar `--repo` en comandos `gh`, y permitir planificación automática sin crear branches/PRs hasta que haya pedido explícito. |
| `issue-creation`       | Se poda     | Queda bloqueada para OpenCode por la policy, aunque exista una copia externa detectada por OpenCode; no se reinstala como skill repo-owned. | La política de issues pertenece a cada proyecto objetivo. Esta capa no borra ni modifica copias de otros runtimes y no impone un workflow de issues global. |
| `hermes-ephemeral-delegation` | Se poda     | Queda fuera del boundary mantenido de OpenCode.                                 | No forma parte del comportamiento mantenido local.                                                                                           |
| `work-unit-commits`    | Se adquiere | Trae el workflow upstream para dividir cambios en unidades de trabajo.         | Se conserva como criterio upstream de agrupación: work units revisables, evidencia, rollback boundary y awareness de chained/review-budget. |
| `code-design`          | Se agrega   | Ayuda a pensar estructura, separación de responsabilidades y diseño de código. | Es una mejora concreta para calidad técnica diaria.                                                                                          |
| `commit-planner`       | Se agrega   | Ordena cambios en commits coherentes.                                          | Queda como capa operativa local: cuando la invoco, propone o ejecuta commits explícitos y usa `work-unit-commits` como criterio si existe. |
| `package-security`     | Se agrega   | Revisa riesgos al instalar o actualizar paquetes.                              | Agrega una capa práctica de seguridad de supply chain.                                                                                       |
| `pr-finalizer`         | Compat temporal | Sigue instalada de forma transitoria para referencias viejas mientras migra la capa final de PR a `branch-pr`. | Ya no es la skill preferida: queda como compatibilidad/deprecada hasta terminar la migración de comandos y referencias. |

### Comportamiento del orchestrator upstream

Gentle AI instala el orchestrator ODD ya preparado con su contrato nativo de review. Esta capa no mantiene una copia ejecutable del template fuente ni certifica su semántica: lo conserva y aplica solo las personalizaciones locales declaradas.

Se mantiene localmente:

- `notion-project-knowledge` y su activación para consultar y conservar conocimiento según las guías actuales de Notion
- el hook `judgment-day` -> `judgment-retrospective`
- las skills repo-owned `code-design`, `package-security`, `commit-planner` y la `branch-pr` sanitizada, con `pr-finalizer` como compatibilidad temporal
- la copia repo-owned sanitizada de `chained-pr` para runtime, mientras su snapshot upstream sigue retenido para auditoría
- el parche puntual que conecta `judgment-day` con `judgment-retrospective`
- Agent-Readable Markdown Standard, Gemini anti-sycophancy y alcance de implementación; el detalle global está en `overlay/gentle-ai/assets/owned/opencode/AGENTS.md`. El runtime adopta el commit ODD de work-unit al cierre de tarea; esto no autoriza push ni PR.
- la regla global que evita que el agente principal, los subagentes y las herramientas accedan a archivos de credenciales; su detalle vive en `overlay/gentle-ai/assets/owned/opencode/AGENTS.md`

Se adquiere desde upstream por defecto:

- la orquestación ODD y el ciclo de review/remediation
- el lifecycle moderno de review y remediation
- los conceptos de PR/chained/work-unit/review-budget, salvo conflicto concreto con la capa local

Excepción concreta:

- la copia upstream de `branch-pr` se poda, pero se reemplaza por una `branch-pr` repo-owned sanitizada que no obliga issue aprobado para toda PR

En otras palabras: la fuente semántica ya no es una lista de parches heredados de v1.44, sino `overlay/gentle-ai/policy/maintenance-intent.md`, que define qué queda local y qué vuelve a upstream.

### Seguridad al instalar paquetes

La guía canónica está en `shared/skills/package-security/SKILL.md` y sus plantillas, dentro de `assets/`. Nunca se deben leer ni buscar archivos de credenciales, como `.env` o un `.npmrc` real; utiliza plantillas, documentación o resultados redactados.

- El ejemplo de npm bloquea scripts de dependencias. npm también permite aprobar scripts revisados, pero esa política es solo informativa por defecto: úsala solo si la versión efectiva admite y tiene activado `strict-allow-scripts`. `ignore-scripts=true` bloquea esos scripts aprobados.
- pnpm ofrece `allowBuilds` para permitir o denegar scripts y `strictDepBuilds` para fallar ante paquetes sin revisar. Comprueba que el pnpm efectivo admita cada opción (`allowBuilds` desde 10.26; `strictDepBuilds` desde 10.3).
- Editar el origen no actualiza la copia global de OpenCode. Para propagarlo hace falta ejecutar `bash apply-gentle-ai-custom.sh opencode` con autorización; hasta entonces, no des por actualizado el runtime.

### Overrides de agentes

Estos overrides existen por cómo OpenCode delega trabajo. Verificado contra la documentación oficial de OpenCode: el sistema trae agentes built-in propios (`Build`, `Plan`, `General`, `Explore`, `Scout`) y, si un subagente no tiene modelo explícito, usa el modelo del agente primario que lo invocó. Esta capa permite fijar de forma explícita los modelos de agentes ajenos a SDD.

Tabla operativa:

| Caso                            | Qué agentes intervienen                                                                                                                                         | Cómo se resuelve el modelo                                                                                                                   | Dónde se cambia                                         |
| ------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------- |
| Delegación built-in de OpenCode | `general`, `explore` | Se asignan modelos explícitos mediante `agent_overrides`. | `~/.config/gentle-ai-custom/opencode-local-config.json` |
| Actores globales | Judgment Day y review | Se asignan por `global_actors`; se admite `{ "inherit": true }`. | `~/.config/gentle-ai-custom/opencode-local-config.json` |

El overlay no crea ni administra perfiles SDD. Las opciones SDD de configuraciones anteriores deben revisarse y quitarse manualmente; los modelos SDD no se reasignan automáticamente a otros agentes.

Si querés fijar modelos para `general` o `explore`, declaralos explícitamente en `agent_overrides` dentro del config local canónico.

En la verificación histórica de v2.5.0, `general` usaba `openai/gpt-5.6-terra` / `high` y `explore` usaba `openai/gpt-5.6-sol` / `medium`. Las asignaciones actuales del runtime no fueron verificadas en esta adopción de snapshots.

### Configuración local

```json
{
  "version": 1,
  "upstream_repo_path": "/path/to/gentle-ai",
  "opencode_config_path": "/path/to/opencode.json",
  "agent_overrides": [
    { "key": "general", "model": "openai/gpt-5.4", "variant": "high" },
    { "key": "explore", "model": "google-vertex/gemini-3.1-pro-preview", "variant": "high" }
  ],
  "global_actors": {
    "jd-judge-a": { "model": "openai/gpt-5.4", "variant": "medium" },
    "review-risk": { "inherit": true }
  }
}
```

- **`agent_overrides`:** Acepta agentes no-SDD. No reasigna opciones SDD a roles ODD.
- **`global_actors`:** Acepta los tres actores globales de Judgment Day y los seis de review, incluido `review-validator`. La omisión conserva el valor actual y `{ "inherit": true }` elimina modelo y variante.
- Si existen `default_profile`, `profiles`, overrides para claves `sdd-*` o claves de agente `sdd-*` en OpenCode, apply se detiene antes de escribir. Revise y elimine esos valores manualmente; apply no los borra ni migra.

## Comandos custom disponibles

- `/commit-plan`
- `/commit-apply`
- `/commit-fast`
- `/pr-create`
- `/pr-regenerate`

Todos se instalan desde `shared/` y generan wrappers específicos por agente en tiempo de aplicación.

`/pr-create` y `/pr-regenerate` mantienen el mismo nombre, pero ahora pasan a estar gobernados por `branch-pr`.

Cuando se invocan de forma explícita, muestran el título y el cuerpo para poder auditarlos, pero no esperan una aprobación extra de ese contenido: crean o actualizan la PR en seguida. Siguen frenando cuando falta autorizar un push o resolver datos obligatorios de la política del repo, del remoto, de la rama base, de la PR o de una cadena de PRs.
