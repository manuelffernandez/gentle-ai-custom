# gentle-ai-custom

Los assets aprobados de Gentle AI se guardan como copias de auditoría. Gentle AI instala y prepara los archivos reales del runtime; esta capa los conserva y aplica solamente los ajustes locales explícitos definidos en `maintenance-intent.md`.

Una capa de customización mantenida con IA para [Gentle AI](https://github.com/Gentleman-Programming/gentle-ai): extiende la experiencia base con skills, prompts, wrappers y política operativa propios, y hace más mantenible el trabajo de reaplicar y auditar lo que `gentle-ai sync` vuelve a materializar.

## Qué es Gentle AI

[Gentle AI](https://github.com/Gentleman-Programming/gentle-ai) es el proyecto original de Gentleman Programming para mejorar de forma MUY fuerte la experiencia de desarrollo con IA: agentes, skills, orquestación SDD, perfiles y tooling real para trabajar mejor con asistentes en el código.

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

## Uso

Para el uso normal, pensalo así: **hay un solo comando importante** y sirve para reaplicar esta capa sobre OpenCode.

Los entrypoints públicos son estos:

- `apply-gentle-ai-custom.sh`
- `apply-gentle-ai-custom.ps1`

Ese es el comando que vas a usar normalmente. El target recomendado es `opencode`. `all` hoy es equivalente porque `opencode` es el único agente soportado.

Los scripts `audit-gentle-ai-upstream.*` y `sync-gentle-ai-upstream-assets.*` existen, pero no están pensados como punto de entrada normal para una persona. Su uso recomendado es a través del agente en modo mantenimiento; salvo debugging puntual o un caso excepcional, no hace falta correrlos manualmente.

Todos estos wrappers delegan en la CLI Go compartida (`go run ./cmd/gentle-ai-overlay ...`) y no duplican lógica entre shell y PowerShell.

### Uso rápido

### Linux / macOS

```bash
bash ~/Documentos/gentle-ai-custom/apply-gentle-ai-custom.sh opencode
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

### Estado actual: snapshots del repo v3.7.0; runtime v2.9.1

El repo ya adoptó Gentle AI v3.7.0 para sus snapshots y su frontera mantenida: `6dee8f833aec9e46015759c5065a9035795d9af1`. La sincronización autorizada informó 36 cambios administrados: 5 archivos nuevos, 31 actualizados, 43 sin cambios, ninguno eliminado y 19 carpetas sincronizadas. También actualizó `upstream-state.json`.

La auditoría posterior pasó para ese mismo commit, sin diferencias administradas ni cambios no administrados vigilados; todos los controles estructurales informaron `ok`. Esto verifica la adopción del repo, no el runtime instalado.

El binario y el runtime siguen en Gentle AI 2.9.1; OpenCode instalado es 1.18.32. No se autorizó ni ejecutó `gentle-ai sync`, una reinstalación ni `apply` para esta adopción. Los cinco plugins V2 son solo snapshots para auditoría y sincronización; OpenCode 1.18.32 usa el camino V1 y este repo no declara soporte para instalar plugins V2.

#### Seguridad de commits ODD pendiente en runtime

El snapshot v3.7.0 de `work-unit-commits` exige cerrar cada tarea ODD con un commit. El parche local que convierte esa instrucción en optativa ya está preparado en el repo, pero todavía no está instalado en el runtime. No des por protegida una sesión que use ese skill hasta que una actualización y un `apply` autorizados por separado terminen y se verifique el resultado. El `apply` actual, con binario 2.9.1, es intencionalmente incompatible con el anchor v3.7.0 y debe detenerse antes de escribir. Mientras tanto, no hagas commits sin autorización explícita.

La copia de v3.7.0 incorpora el contrato output-only de `sdd-research`; por eso se retiró su parche local redundante. El handoff local de Judgment Day se mantiene.

El repo también acepta de forma consciente que una solicitud explícita de archivo SDD pueda guardar specs aplicables aunque queden tareas sin marcar. Es una decisión local: no confirma la intención de upstream ni demuestra que la implementación esté completa. Tampoco cambia la revisión RDD escalada por separado. La guía de mantenimiento incluye qué revisar en futuras auditorías.

#### Evidencia histórica del runtime v2.9.1 y v2.5.0

En el cierre anterior de v2.9.1, la auditoría pasó para `9ec0cf443622f20fe511815ff5a83088c7467bff`, y una verificación de solo lectura confirmó 602 comprobaciones sobre asignaciones, perfiles, prompts, permisos, alcance de tareas y referencias. También confirmó el orchestrator con consentimiento compatible, las 7 skills propias, 5 wrappers, el `AGENTS.md` propio, la poda de skills y los parches locales. Esa evidencia corresponde al runtime v2.9.1 y no verifica un runtime v3.7.0.

#### Evidencia histórica del runtime v2.5.0

La policy y la auditoría de este repositorio incorporan la fase real `sdd-research`, ubicada después de `sdd-explore`, y validan perfiles con 11 fases. Históricamente, la instalación completa del componente SDD v2.5.0 realizada manualmente y la posterior ejecución autorizada por separado de `bash apply-gentle-ai-custom.sh opencode` terminaron correctamente.

El `apply` actualizó la configuración, podó 2 skills, cambió y verificó 1 asset propio, no informó warnings de topología y administró 4 perfiles. En los agentes informó 3 creados, 4 actualizados y 49 sin cambios, sin perfiles no administrados. La verificación final pasó la auditoría canónica, `go test ./...`, `git diff --check` y la carga resuelta de los agentes `sdd-research`.

La copia sanitizada y propia de `branch-pr` está instalada. La copia conflictiva upstream de `branch-pr`, `issue-creation` y Hermes no están. Las 7 skills propias, los 5 wrappers, el `AGENTS.md` propio, los overrides y los 4 perfiles administrados coinciden con sus fuentes o con la configuración local. `general` usa `openai/gpt-5.6-terra` / `high`; `explore` usa `openai/gpt-5.6-sol` / `medium`. Los cuatro perfiles tienen 11 fases en orden y existen los 3 actores nombrados de `sdd-research`. Los actores de review siguen siendo globales, sin copias con sufijo, y Judgment Day difiere del upstream solo por el parche de retrospectiva.

En esa verificación histórica del runtime v2.5.0, la intención y la política efectiva de background estaban en `on`, y el launcher era `~/.gentle-ai/bin/opencode`. Los procesos abiertos desde Desktop o con `opencode` normal seguían en foreground. También se observó que `~/.gentle-ai/state.json` informaba `installed_binary_version: 2.4.0`; no se conoce la causa y no se modificó ese estado externo.

La política del repositorio considera RDD activo y funcional por decisión explícita del propietario, basada en validaciones operativas realizadas en otros repositorios. Esta decisión no cambia el runtime: en esta máquina se observó el modo global en `on` y sin ajuste local para este clon. El usuario excluyó de forma explícita este mantenimiento de v2.9.1 de RDD. No se autorizó ni realizó ningún commit, push o PR para este cierre documental.

### Adaptación del repositorio para v2.9

La auditoría entiende la nueva composición del orchestrator sin depender del número de versión: sigue la cadena real desde la asignación del prompt hasta el asset OpenCode y rechaza asignaciones sueltas o señuelos. También vigila el contrato upstream de autorización remota y comprueba que llegue a los prompts de OpenCode que este repo mantiene.

El skill upstream `sdd-research` sigue siendo la fuente del runtime. En los snapshots adoptados de v3.7.0, tanto el skill como el comando de investigación indican que la recolección es output-only y que el orchestrator se ocupa de la persistencia. Por eso se retiró el parche local de precedencia. El runtime instalado sigue en 2.9.1 y no se afirma que ya use ese contrato.

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

El `AGENTS.md` propio sigue siendo un archivo completo que instala este repo. Su persona y el protocolo completo de Engram para OpenCode deben coincidir con los snapshots aprobados. Los tests permiten cinco secciones propias: alcance de implementación, estándar de Markdown para agentes, seguridad de archivos de credenciales, regla de no hacer commits automáticamente y ajuste para Gemini.

### Cuándo usar `gentle-ai sync` y cuándo reinstalar

- Es preferible usar `gentle-ai sync` si los cambios que adoptamos sirven, pero no rompen de forma profunda cómo están configurados los agentes y perfiles.
- Conviene hacer una reinstalación completa cuando los cambios del proyecto original afectan la estructura básica de los perfiles que usamos, o cuando un simple `gentle-ai sync` ya no alcanza para armar bien la configuración.
- Si el proyecto original agrega soporte para un nuevo agente o plataforma que no usamos, no hace falta reinstalar.
- Las reglas del proyecto original sobre estrategia de PR, PRs encadenadas y presupuesto de revisión se adoptan por defecto. Solo se sanitizan si chocan con una personalización local explícita. En v3.7.0, ODD pide cerrar cada tarea con un commit. El snapshot y el parche exacto que lo vuelve optativo ya están en el repo, pero aún no en el runtime. El runtime sigue en 2.9.1; su `apply` es intencionalmente incompatible con el anchor v3.7.0. No se debe afirmar que esa protección está activa hasta una actualización y un `apply` autorizados por separado, seguidos de verificación. Hasta entonces, no hagas commits sin autorización explícita.

### Qué hace el comando `apply-gentle-ai-custom`

- vuelve a instalar nuestras skills y comandos personalizados.
- borra las skills que no queremos usar, pero solo para los agentes donde estamos trabajando.
- bloquea en OpenCode cada skill podada que no tenga un reemplazo propio con el mismo nombre mediante `permission.skill`, incluso si OpenCode encuentra una copia en `~/.agents/skills` o `~/.claude/skills`. No borra ni modifica esas copias externas, y conserva los permisos de skills no podadas ni de reemplazos propios como `branch-pr`.
- aplica nuestras preferencias de modelos de IA si las configuramos.
- comprueba solo los archivos, referencias y anchors que necesita para aplicar perfiles o los parches locales declarados.
- aplica el parche puntual de Judgment Day y, cuando el runtime contenga el anchor v3.7.0 compatible, el parche ODD que deja el commit optativo. Si el anchor no coincide, `apply` se detiene antes de escribir. El parche de `sdd-research` se retiró de los snapshots adoptados por estar cubierto por el contrato upstream v3.7.0.
- conserva los prompts y plugins preparados por Gentle AI; no los reemplaza con plantillas fuente. La frontera de snapshots está en v3.7.0, pero el runtime y el binario siguen en 2.9.1; no se autorizó su actualización ni el `apply`. La política acepta RDD como activo y funcional, pero su estado efectivo se consulta por máquina o clon.
- ajusta modelos y perfiles reutilizando los prompts base ya instalados.
- agrega de forma puntual el handoff de `judgment-retrospective` al skill `judgment-day` compatible.
- conserva el contrato upstream output-only de `sdd-research`; no instala un parche local redundante.
- registra todos los archivos que instala como propios. Al limpiar registros viejos, conserva los archivos devueltos a Gentle AI y los targets de parches puntuales sin declararlos como propios.

Ejecutarlo apuntando a `opencode` reinstala todo para OpenCode. Si usás `all` hace lo mismo, porque hoy es el único entorno soportado.

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

SDD y RDD cumplen funciones distintas. SDD organiza la planificación, implementación, tests y verificación. RDD cruza ese flujo al final: cuando está activo, puede revisar el candidato inmutable después de `apply` y aportar el recibo usado en la entrega. Si RDD está apagado, no debe lanzar revisores; SDD sigue funcionando, incluidos tests y verificación, y el cierre usa la ruta `disabled/unmanaged` bajo la política normal del repositorio sin inventar un recibo.

RDD está aceptado como activo y funcional por decisión explícita del propietario, basada en validaciones operativas realizadas en otros repositorios. El estado efectivo pertenece al usuario y a cada máquina o clon; Gentle AI upstream no lo controla de forma remota. Consúltalo siempre con `gentle-ai review mode status --cwd .`. En esta máquina se observó `on` por la configuración global y ningún ajuste local para este clon.

Como contexto histórico, Gentle AI v2.2.2 tuvo un bug confirmado en la entrega del alcance de una corrección. El código y los tests de v2.5.0 aportaron el plan previo a la edición, y este repositorio evaluó pero aplazó un parche local. Ese análisis se conserva en `overlay/gentle-ai/maintenance.md`. El canary con corrección puede seguir usándose para diagnosticar regresiones, pero ya no bloquea la activación. Siguen siendo obligatorios el interruptor controlado por el usuario, los recibos ligados al candidato exacto, los límites de corrección y la política normal de entrega.

El runtime actual de v2.5.0 ya pasó la instalación completa del componente SDD y el `apply` autorizado por separado. La intención y la política efectiva de background están en `on`; para una sesión CLI con background usa `~/.gentle-ai/bin/opencode`. Desktop y el comando `opencode` normal siguen iniciando procesos en foreground. Un futuro `gentle-ai sync` o una reinstalación volverá a requerir un `apply` autorizado por separado. La aceptación de RDD no autoriza esas acciones ni modifica la configuración.

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

- si no ponés `profiles`, no agrega perfiles especiales.
- si no ponés `agent_overrides`, no cambia los modelos que vienen por defecto.
- si no ponés `default_profile`, no altera la familia base de perfiles SDD.
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

La idea general ahora es **upstream-first**: esta capa arranca desde el comportamiento actual de Gentle AI v2+, conserva unas pocas personalizaciones locales explícitas y evita seguir cargando parches viejos si ya no hacen falta. El foco está en mantener un overlay chico y entendible.

### Skills del overlay

| Skill                  | Estado      | Qué hace                                                                       | Por qué se queda / se va                                                                                                                     |
| ---------------------- | ----------- | ------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `_shared`              | Se conserva | Expone referencias compartidas para skills SDD.                                | Sostiene el resto del flujo sin imponer proceso extra.                                                                                       |
| `cognitive-doc-design` | Se conserva | Ayuda a escribir documentación más clara y escaneable.                         | Mejora la calidad de los artifacts sin meter gobernanza de repo.                                                                             |
| `comment-writer`       | Se conserva | Ayuda a redactar comentarios de colaboración.                                  | Suma claridad comunicacional sin forzar una estrategia de PR.                                                                                |
| `gentle-ai-bench`      | Snapshot retenido | Documenta cómo crear y verificar journeys del bench upstream.              | Se guarda para auditoría, pero este repo no la instala como propia ni obliga a usarla.                                                        |
| `go-testing`           | Se conserva | Da patrones prácticos para tests en Go.                                        | Aporta calidad técnica directa.                                                                                                              |
| `judgment-day`         | Se conserva | Hace revisión adversarial/dual review y dispara la retrospectiva automática al cerrar desde el hook runtime del overlay. | Refuerza criterio técnico, valida de verdad y deja aprendizaje reutilizable.                                                                |
| `judgment-retrospective` | Se agrega   | Resume el juicio, guarda patrones reutilizables e historial de intervenciones.  | Diseñada para ser invocada por el hook runtime de `judgment-day`; también se puede cargar manualmente. Evita guardar el output crudo y ayuda a ver si una corrección realmente funcionó. |
| `rdd-defect-workflow`  | Snapshot retenido | Describe el trabajo upstream sobre defectos RDD.                           | Se guarda para auditoría y no se instala como skill propia; el modo efectivo de RDD se decide fuera de este snapshot.                         |
| `sdd-init`             | Se conserva | Inicializa contexto, testing y capacidades SDD.                                | Es parte del núcleo útil del flujo SDD.                                                                                                      |
| `sdd-explore`          | Se conserva | Explora ideas antes de proponer cambios.                                       | Mejora análisis y reduce implementación impulsiva.                                                                                           |
| `sdd-research`         | Se adquiere | Reúne evidencia externa auditable antes de la propuesta cuando el orchestrator la solicita. | Agrega una fase real y separa investigación con fuentes de las decisiones de producto.                                                       |
| `sdd-propose`          | Se conserva | Convierte exploración en propuesta concreta.                                   | Mantiene trazabilidad y foco de cambio.                                                                                                      |
| `sdd-spec`             | Se conserva | Escribe requisitos y escenarios.                                               | Conserva la disciplina del flujo SDD.                                                                                                        |
| `sdd-design`           | Se conserva | Desarrolla el diseño técnico del cambio.                                       | Aporta arquitectura y decisiones explícitas.                                                                                                 |
| `sdd-tasks`            | Se conserva | Baja diseño/especificación a tareas ejecutables.                               | Hace operativo el trabajo sin perder estructura.                                                                                             |
| `sdd-apply`            | Se conserva | Implementa tareas definidas por SDD.                                           | Conserva la parte más útil del flujo técnico.                                                                                                |
| `sdd-verify`           | Se conserva | Verifica que lo aplicado cumpla con spec y tasks.                              | Mantiene validación técnica al final del ciclo.                                                                                              |
| `sdd-archive`          | Se conserva | Cierra el cambio y deja persistencia final.                                    | Completa el circuito SDD con trazabilidad.                                                                                                   |
| `sdd-onboard`          | Se conserva | Guía un uso completo del flujo SDD.                                            | Facilita adopción sin tocar gobernanza de repo.                                                                                              |
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

Gentle AI instala el orchestrator SDD ya preparado con su contrato nativo de review. Esta capa no mantiene una copia ejecutable del template fuente ni certifica su semántica: lo conserva y solo reutiliza sus valores cuando configura perfiles.

Se mantiene localmente:

- el hook `judgment-day` -> `judgment-retrospective`
- las skills repo-owned `code-design`, `package-security`, `commit-planner` y la `branch-pr` sanitizada, con `pr-finalizer` como compatibilidad temporal
- la copia repo-owned sanitizada de `chained-pr` para runtime, mientras su snapshot upstream sigue retenido para auditoría
- el parche puntual que conecta `judgment-day` con `judgment-retrospective`
- las reglas `no-auto-commit`, Agent-Readable Markdown Standard, Gemini anti-sycophancy y alcance de implementación; el detalle global está en `overlay/gentle-ai/assets/owned/opencode/AGENTS.md`
- la regla global que evita que el agente principal, los subagentes y las herramientas accedan a archivos de credenciales; su detalle vive en `overlay/gentle-ai/assets/owned/opencode/AGENTS.md`

Se adquiere desde upstream por defecto:

- el comportamiento SDD/orchestrator
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

Estos overrides existen por cómo OpenCode delega trabajo. Verificado contra la documentación oficial de OpenCode: el sistema trae agentes built-in propios (`Build`, `Plan`, `General`, `Explore`, `Scout`) y, si un subagente no tiene modelo explícito, usa el modelo del agente primario que lo invocó. Esta capa agrega overrides porque, en este setup, quiero fijar de forma explícita qué modelo usan algunos agentes built-in de OpenCode cuando entran en juego fuera del sistema de perfiles SDD.

Tabla operativa:

| Caso                            | Qué agentes intervienen                                                                                                                                         | Cómo se resuelve el modelo                                                                                                                   | Dónde se cambia                                         |
| ------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------- |
| Delegación built-in de OpenCode | `general`, `explore`                                                                                                                                            | Este overlay les fija modelo/variant explícitos mediante `agent_overrides`, en vez de dejar que hereden el modelo del agente que los invoca. | `~/.config/gentle-ai-custom/opencode-local-config.json` |
| Familia base SDD                | `gentle-orchestrator`, fases SDD, tres actores de Judgment Day y seis actores de review globales                                                                 | Se resuelve por `default_profile`, no por `agent_overrides`. Los actores pueden fijar modelo o usar herencia explícita.                      | `~/.config/gentle-ai-custom/opencode-local-config.json` |
| Perfiles SDD nombrados          | `sdd-orchestrator-<name>` + `sdd-<phase>-<name>` y, solo si se configuran, actores JD con sufijo                                                                 | Se resuelve por `profiles`. Los seis actores de review siempre son globales y no reciben sufijo.                                             | `~/.config/gentle-ai-custom/opencode-local-config.json` |

En otras palabras: estos overrides no duplican la configuración SDD. Cubren una capa distinta: los agentes built-in de OpenCode que el orchestrator puede usar al delegar fuera del sistema de perfiles SDD.

Si querés fijar modelos para `general` o `explore`, declaralos explícitamente en `agent_overrides` dentro del config local canónico.

En la máquina verificada actualmente, `general` usa `openai/gpt-5.6-terra` / `high` y `explore` usa `openai/gpt-5.6-sol` / `medium`.

### Configuración avanzada

Si necesitás ir más allá de los overrides básicos y querés configurar familias enteras de perfiles SDD, este es el esquema completo que soporta el archivo:

```json
{
  "version": 1,
  "upstream_repo_path": "/path/to/gentle-ai",
  "opencode_config_path": "/path/to/opencode.json",
  "agent_overrides": [
    { "key": "general", "model": "openai/gpt-5.4", "variant": "high" },
    {
      "key": "explore",
      "model": "google-vertex/gemini-3.1-pro-preview",
      "variant": "high"
    }
  ],
  "default_profile": {
    "orchestrator": { "model": "openai/gpt-5.4", "variant": "high" },
    "phases": {
      "sdd-init": { "model": "openai/gpt-5.4", "variant": "medium" },
      "sdd-explore": { "model": "openai/gpt-5.3-codex", "variant": "high" },
      "sdd-research": { "model": "openai/gpt-5.4", "variant": "high" },
      "sdd-propose": { "model": "openai/gpt-5.5", "variant": "xhigh" },
      "sdd-spec": { "model": "openai/gpt-5.3-codex", "variant": "high" },
      "sdd-design": { "model": "openai/gpt-5.5", "variant": "high" },
      "sdd-tasks": { "model": "openai/gpt-5.3-codex", "variant": "high" },
      "sdd-apply": { "model": "openai/gpt-5.3-codex", "variant": "high" },
      "sdd-verify": { "model": "openai/gpt-5.4", "variant": "xhigh" },
      "sdd-archive": { "model": "openai/gpt-5.4-mini", "variant": "medium" },
      "sdd-onboard": { "model": "openai/gpt-5.4", "variant": "medium" }
    },
    "judgment_day": {
      "jd-judge-a": { "model": "openai/gpt-5.4", "variant": "medium" },
      "jd-judge-b": { "inherit": true },
      "jd-fix-agent": { "model": "openai/gpt-5.4-mini" }
    },
    "review": {
      "review-risk": { "inherit": true },
      "review-readability": { "inherit": true },
      "review-reliability": { "inherit": true },
      "review-resilience": { "inherit": true },
      "review-refuter": { "inherit": true },
      "review-validator": { "inherit": true }
    }
  },
  "profiles": [
    {
      "name": "cheap",
      "orchestrator": { "model": "openai/gpt-5.4-mini", "variant": "low" },
      "phases": {
        "sdd-init": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-explore": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-research": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-propose": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-spec": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-design": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-tasks": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-apply": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-verify": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-archive": { "model": "openai/gpt-5.4-mini", "variant": "low" },
        "sdd-onboard": { "model": "openai/gpt-5.4-mini", "variant": "low" }
      },
      "judgment_day": {
        "jd-judge-a": { "model": "openai/gpt-5.4-mini", "variant": "low" }
      }
    }
  ]
}
```

- **`agent_overrides`:** Fija modelos para agentes específicos fuera de SDD (como `general` o `explore`).
- **`default_profile`:** Fija los modelos base para toda la familia de SDD por defecto.
- **`default_profile.judgment_day`:** Acepta solo `jd-judge-a`, `jd-judge-b` y `jd-fix-agent`.
- **`default_profile.review`:** Acepta solo los seis actores globales de review, incluido `review-validator`. `{"inherit": true}` elimina su modelo y variant para que OpenCode herede ambos al delegar.
- **`profiles`:** Crea perfiles SDD con nombres personalizados. `judgment_day` es opcional y solo crea los actores JD declarados con sufijo; los omitidos usan el actor JD global. `review` no está permitido acá.

## Comandos custom disponibles

- `/commit-plan`
- `/commit-apply`
- `/commit-fast`
- `/pr-create`
- `/pr-regenerate`

Todos se instalan desde `shared/` y generan wrappers específicos por agente en tiempo de aplicación.

`/pr-create` y `/pr-regenerate` mantienen el mismo nombre, pero ahora pasan a estar gobernados por `branch-pr`.

Cuando se invocan de forma explícita, muestran el título y el cuerpo para poder auditarlos, pero no esperan una aprobación extra de ese contenido: crean o actualizan la PR en seguida. Siguen frenando cuando falta autorizar un push o resolver datos obligatorios de la política del repo, del remoto, de la rama base, de la PR o de una cadena de PRs.
