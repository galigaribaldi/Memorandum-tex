# Memorandum-tex — Mapa de arquitectura

> Fuente única de verdad de la arquitectura. Describe lo que **existe**; lo que todavía no existe se marca con la fase
> en la que nace (⏳ M1, ⏳ M2…). Cuando algo cambie, el párrafo viejo se reescribe o se tacha con fecha
> (`~~…~~ **Resuelto el AAAA-MM-DD:** …`), nunca se deja contradiciendo al repositorio.

## 1. Mapa de directorios

```
Memorandum-tex/
├── CLAUDE.md                      ← reglas, stack, gotchas y estado (se carga en cada sesión)
├── README.md                      ← qué es, cómo se usa, estado
├── CONTRIBUTING.md                ← cómo proponer una plantilla nueva o un cambio
├── WORKFLOW_BLUEPRINT.md          ← método de trabajo con IA (Tríada, contrato vivo, deuda, verificación)
├── LICENSE                        ← MIT (lo detecta GitHub): herramientas, librería, esquemas, documentación
├── LICENSES/                      ← textos exactos por identificador SPDX (REUSE): MIT.txt, MIT-0.txt (+ LicenseRef-*)
├── NOTICE.md                      ← qué licencia cubre qué; exclusiones (logos, material de terceros)
├── .gitattributes                 ← LF en todo; binarios sin conversión; sin LFS
├── .gitignore                     ← auxiliares de LaTeX, dist/, go.work, claves de firma (*.key), ajustes personales
├── .editorconfig                  ← UTF-8, LF, 2 espacios (tabs en Go)
├── .claude/
│   ├── settings.json              ← permisos: sin borrar, sin commit/push/tags; leer y escribir con permiso
│   └── skills/
│       ├── latex-template-author/SKILL.md ← crear o migrar plantillas y componentes, con lista de verificación
│       └── template-contract/SKILL.md     ← manifiesto, índice, paquetes, librería y catálogo
├── docs/
│   ├── MAP.md                     ← este archivo
│   ├── integracion-creative-ink.md ← plan y requerimientos para el agente de Creative-Ink (su §4 quedó reemplazada)
│   ├── contrato-plantillas.md     ← contrato v1: manifiesto, catálogo, paquetes, bundle, ink-params, reglas V01–V52
│   └── guia-nueva-plantilla.md    ← cómo escribir una plantilla (del bloque PERSONALIZACIÓN a parámetros, modos, motor)
├── schema/                        ← JSON Schema draft-07 (ayuda al editor; la validación que manda es inkt)
│   ├── template.schema.json       ← manifiesto (parámetros con reglas por tipo)
│   ├── index.schema.json          ← índice del catálogo (define "ref": ruta relativa o https)
│   ├── versions.schema.json       ← versiones publicadas de una plantilla
│   └── pack.schema.json           ← paquete ("todo" reservado)
├── packs/                         ← academico, presentaciones, correspondencia, congresos (todo se genera)
├── shared/                        ← ⏳ M3: componentes que se resuelven dentro de cada bundle
│   ├── motor/ink-motor.tex        ← iftex: inputenc/fontenc (pdfLaTeX) o fontspec (Xe/LuaLaTeX)
│   ├── base/ink-base.tex          ← carga de ink-params, \inkIfTrue, \inkItem
│   ├── paletas/                   ← Institucional, Teal, VerdeOlivo, Purpura, Rojo, VerdeEsmeralda
│   ├── comandos/                  ← base, academico, revision, codigo
│   └── instituciones/unam-fes-acatlan/ ← textos institucionales y logos
├── templates/                     ← ⏳ M4–M6: templates/<categoria>/<id>/{template.json,README.md,CHANGELOG.md,preview.png,src/}
├── tools/                         ← módulo Go github.com/galigaribaldi/Memorandum-tex/tools (go 1.22, solo biblioteca estándar)
│   ├── go.mod
│   ├── cmd/inkt/                  ← ⏳ M2 paso 8: CLI validate, bundle, index, sign (después: new, check, format, diff, clean)
│   ├── fixtures/                  ← paquete Go que incrusta los datos de prueba (embed.FS): viajan con el módulo
│   │   └── data/                  ← datos de prueba del contrato (ver su README)
│   │       ├── templates/prueba/  ← ok-minima (válida), futura (schemaVersion 99), con-paquetes (missing-packages)
│   │       ├── packs/prueba.json
│   │       ├── params/            ← EscapeTeX, Resolve y salida exacta de ink-params.tex
│   │       ├── invalid/           ← manifests.json (V01–V20) y folders.json (V30–V39), generados al vuelo en los tests
│   │       └── bundles/cases.json ← 16 zips válidos y maliciosos, construidos en los tests (sha256 reproducible)
│   ├── internal/fixturetest/      ← ayudas de test: carga de casos, JSON Merge Patch, copia de carpetas
│   ├── internal/validate/         ← ✅ validación de una carpeta de plantilla completa (manifiesto + V30–V39), para inkt y la CI
│   └── pkg/
│       ├── contract/              ← ✅ tipos, errores compartidos, validación del manifiesto (V01–V20) y paquetes (V50–V52),
│       │                             lista única de archivos permitidos (files.go: la usan validate y bundle)
│       ├── params/                ← ✅ ParseAssignments, Resolve, RenderInkParams, EscapeTeX
│       ├── compat/                ← ✅ Evaluate, EvaluateJSON, UpdateAvailable (ready … blocked)
│       ├── bundle/                ← ✅ Collect/Build/WriteZip (reproducible), SHA256, VerifySHA256, SafeExtract (B1–B11)
│       ├── catalogtest/           ← ⏳ paso 6: catálogo de prueba firmado (clave efímera) servido con httptest
│       └── client/                ← ⏳ paso 7: índice, manifiestos, firmas, caché ETag, Get y GetPack
└── .github/workflows/             ← ⏳ M7: validate.yml (3 SO) y release.yml (Pages + Releases)
```

## 2. Reglas de capas

```
 templates/<cat>/<id>/src   ──declara──►  template.json (components, parameters, build)
          │                                     │
          │  solo referencia lo suyo y          │ inkt bundle (M2) resuelve los componentes
          │  los componentes declarados         ▼
 shared/*  ─────────copiados a──────────►  <id>-<versión>.zip  (autocontenido, Latex/ink/<componente>/…)
                                                 │
                               inkt index + inkt sign ▼   (clave privada: solo en un secreto de GitHub Actions)
                  catalog/v1/index.json + manifiestos + versions.json + paquetes, cada uno con su .sig (Pages)
                                                 │
                                                 ▼   client verifica la firma (clave pública fija) y el sha256
                   Creative-Ink (importa tools/pkg/*; decide rutas, comandos y UI)
```

- Una plantilla **nunca** referencia archivos fuera de su carpeta: lo compartido entra solo a través de `components`.
- `shared/` no conoce a ninguna plantilla (sin condicionales por id).
- Los paquetes de `tools/pkg/*` importan **solo la biblioteca estándar**. `contract`, `params`, `compat` y `bundle` no
  usan la red ni ejecutan procesos; solo `client` usa la red.
- Creative-Ink consume el catálogo y la librería; nunca lee este repositorio con `git clone` ni con la API de GitHub.
- Nada de la cadena ejecuta código de una plantilla: sin `.latexmkrc`, sin `--shell-escape` y sin scripts (decisión 15).

## 3. Contrato (tabla viva)

La especificación completa es `docs/contrato-plantillas.md` (v1, 2026-11-10). Esta tabla es el resumen; si discrepan,
gana el contrato.

| Artefacto | Productor → consumidor | Forma y semántica | Contrato | Estado |
|-----------|------------------------|-------------------|----------|--------|
| `template.json` | Autor → `inkt`, Creative-Ink | `id` kebab-case = carpeta; `version` semver; `build` (`engines` con `pdflatex` y `xelatex`, `defaultEngine: pdflatex`, `bibliography: bibtex`, `shellEscape` siempre `false` en v1); `requires`; `components`; `parameters` (`string`, `boolean`, `enum`, `date`, `list`); `modes` con `completo`; `license` SPDX con `MIT-0`. Campos desconocidos: aviso al validar, ignorados por los clientes | §3–§5 | ✅ validado por `pkg/contract` (48 casos) |
| Manifiesto publicado | CI → Creative-Ink | El del repositorio sin `$schema` ni `components`, más `dist {url, sha256, size, publishedAt}` | §6 | ⏳ M2 (generador), M7 (real) |
| `versions.json` | CI → Creative-Ink | Versiones publicadas, más nueva primero; rutas relativas al archivo | §6 | ⏳ M2 |
| `index.json` | CI → Creative-Ink | Última versión de cada plantilla y lista de paquetes; ≤ 1 MB; rutas relativas a la URL del índice | §7 | ⏳ M2 (fixture), M7 (real) |
| `<archivo>.sig` | `inkt sign` (CI) → `client` | `{keyId, algorithm: ed25519, signature}` sobre los bytes exactos de cada JSON del catálogo; obligatoria en el oficial | §7.1 | ⏳ M2 pasos 6–8; clave real en M7 |
| `packs/<id>.json` | Repositorio → Creative-Ink | 1–100 `{id, version?}`; sin zip propio; `todo` reservado y generado | §8 | ✅ 4 paquetes, validados por `pkg/contract` |
| Bundle `<id>-<versión>.zip` | `inkt bundle` → librería | Reglas B1–B11: una raíz = id; segmentos `^[A-Za-z0-9._-]+$`; sin enlaces ni colisiones por mayúsculas; lista de extensiones + `LICENSE` y `.creative-ink-meta`; rc de latexmk rechazados en cualquier forma; ≤ 2000 archivos, 200 MB y 100:1 (desde 1 MB); nunca más de lo declarado; sin `shellEscape`; componentes en `Latex/ink/<componente>/`; reproducible con la misma versión de Go | §9 | ✅ `pkg/bundle`, 16 casos + 7 tests |
| `.creative-ink-meta` del bundle | `inkt bundle` → Creative-Ink | Exactamente `{"version": 1, "mainFile": "<entry>"}` (`core.ProjectMeta` de Creative-Ink) | §9 B7 | Especificado |
| `ink-params.tex` | Librería → plantilla | Dos líneas de cabecera y `\def\<macro>{valor}` en el orden del manifiesto; `EscapeTeX` igual a `core.EscapeTeX` | §10 | ✅ `pkg/params`, 2 salidas exactas |
| Errores de parámetros | Librería → Creative-Ink | Códigos `unknown`, `required`, `type`, `enum`, `length`, `items`, por campo | §4.3 | ✅ `pkg/params`, 10 casos |
| Estado de compatibilidad | Librería → Creative-Ink | `ready`, `missing-packages`, `missing-engine`, `blocked` (`--shell-escape`), `incompatible`; gana el más grave | integración §7 | ✅ `pkg/compat` |
| Macros de modo | Creative-Ink → plantilla | `\borrador` y `\SoloCapitulo` (nombres fijos, los usa `ci compile`) | §5 | Especificado; probado en `ok-minima`; ⏳ M3 en componentes |
| Librería Go | Este repo → Creative-Ink | `contract`, `params`, `compat`, `bundle`, `catalogtest`, `client`; errores tipados de §13 (con `ErrSignature`) | §13 | ✅ 3 de 6 paquetes; ⏳ M2 pasos 4–8 |

## 4. Subsistemas

### 4.1 Plantillas (⏳ M4)
Anatomía: `templates/<categoria>/<id>/` con `template.json`, `README.md` (uso, parámetros, modos), `CHANGELOG.md`
(semver propio), `preview.png` (≤ 200 KB) y `src/` (lo que se copia al proyecto: `main.tex`, `Latex/`, `secciones/`,
`referencias.bib`, `img/`). El archivo principal es `main.tex` salvo que `entry` diga otra cosa.

### 4.2 Componentes compartidos (⏳ M3)
Fuente única de paletas, motor, comandos e instituciones. En el repositorio viven una sola vez; `inkt bundle` copia cada
componente declarado de `shared/<componente>/` a `Latex/ink/<componente>/` del bundle (implementado en `pkg/bundle`,
2026-11-15); `src/` no puede usar `Latex/ink/`. Una institución nueva es una carpeta en `shared/instituciones/` con su
`institucion.tex` (textos) y `logos/`; ninguna plantilla escribe a mano "UNAM" o "FES Acatlán".

### 4.3 Parámetros y modos (✅ especificado en M1, implementado en `pkg/params` el 2026-11-11)
Variantes como booleanos (`portada`), opciones como `enum` (`color`, `tamano`), datos como `string`/`date`/`list`.
Orden obligatorio en `main.tex`: `\InputIfFileExists{ink-params}{}{}` → `\providecommand` con los valores por
defecto → resto del preámbulo. Modos: `completo`, `rapido`, `borrador`, `capitulo` (solo los que la plantilla declara).
Flujo: `params.ParseAssignments` (palabras `clave=valor`) → `params.Resolve` (valor del usuario, nombre del proyecto,
valor por defecto; errores por campo) → `params.RenderInkParams` (el archivo exacto). Comprobado en LaTeX real con
`ok-minima` en los tres motores y los modos borrador y capítulo (2026-11-11).

### 4.4 Motores y bibliografía (⏳ M3)
`ink-motor.tex` decide con iftex. Por defecto pdfLaTeX + natbib/apalike + BibTeX; XeLaTeX y LuaLaTeX donde
`build.engines` los liste; biblatex + Biber solo como parámetro `bibliografia` de las plantillas que lo ofrezcan.

### 4.5 Herramienta `inkt` y librería (M2 en curso: pasos 1–3 de 8)
Sustituye a los dos Makefiles y a los scripts de Python de los repositorios de origen.
- **Librería** (`tools/pkg/`, la importa Creative-Ink): ✅ `contract`, ✅ `params`, ✅ `compat`; ⏳ `bundle`,
  `catalogtest`, `client`. Organización, estado y API en `docs/integracion-creative-ink.md` §5.
- **CLI `inkt`** (solo para este repositorio y su CI): `validate`, `bundle`, `index` y `sign` en M2; `new`,
  `check bib|labels|refs`, `format`, `diff` y `clean` en M6.
- **Reglas de implementación:** solo biblioteca estándar; errores compartidos en `contract`; ninguna función
  ejecuta procesos y solo `client` usa la red; cada regla del contrato tiene casos en `tools/fixtures/data/` y un test
  que comprueba que ningún caso queda sin probar.

### 4.6 Publicación y catálogo (⏳ M7)
Tag `<id>-v<versión>` (lo crea el humano) → CI valida, empaqueta, sube el zip a la Release, regenera `catalog/v1/` y
**firma cada JSON** con la clave privada del secreto de GitHub Actions (`inkt sign`) antes de publicarlo en GitHub
Pages. Escalabilidad estimada en `docs/integracion-creative-ink.md` §3.2.
- **Antes de la primera publicación:** el humano genera el par de claves Ed25519 y guarda la privada en el secreto.
  La pública entra en `client.Official` y se publica una versión de la librería que ya la conozca.
- **Rotación:** contrato §7.1.

### 4.7 Kit de seguimiento de la tesis (⏳ M6)
Extraído de TesisUrbanismo: plantillas de issues (revisión de capítulo, figura o tabla, bibliografía, error de LaTeX),
`CHANGELOG.md` + `Nota-Version.tex`, `docs/contexto/CONTEXTO_CAP.md` (retomar aquí, decisiones D-n fechadas, pendientes,
bitácora), `revisiones/<revisor>/<ronda>/` con categorías, `docs/VERSIONADO.md`. Opcionales: CI de release por tag y
`CLAUDE.md` + `.claude/settings.json` para la tesis.

## 4.8 Diseño acordado

### Decisiones del usuario (2026-11-10)
1. Distribución en GitHub: catálogo en **Pages**, bundles en **Releases** del repositorio `galigaribaldi/Memorandum-tex`.
2. **Descarga bajo demanda**, individual o por paquete; nada se instala de golpe. Dónde se guarda en
   `My-Creative-Ink-Repository` lo decide Creative-Ink.
3. Herramientas en **Go** (librería incrustada en la Terminal CI de Creative-Ink); Python solo para extensiones futuras.
4. Manifiesto en **JSON** con JSON Schema.
5. Variantes como **parámetros booleanos**; **semver por plantilla**; parámetros en **`ink-params.tex`**.
6. Motor por defecto **pdfLaTeX** (Xe/LuaLaTeX opcionales). Bibliografía por defecto **natbib/apalike + BibTeX** (la
   que predomina en Trabajos-Maestria-Urbanismo); biblatex + Biber opcional.
7. Se incluyen los logos UNAM/FES.
8. ~~Software libre; licencia (probablemente MIT) e identidad (texto de muestra con el nombre y la descripción del
   repositorio) se deciden al final, en M8.~~ **Resuelto el 2026-11-11:** la licencia es MIT + MIT-0 (decisión 14);
   en M8 queda solo la identidad.
9. Solo español por ahora.
10. Kit de la tesis: issues, CHANGELOG, CONTEXTO y revisiones por defecto; CI y `CLAUDE.md` opcionales.
11. En Creative-Ink: catálogo en el **núcleo** y **catálogos de terceros** marcados como tales.

### Decisiones del usuario (2026-11-11)
12. **El catálogo de prueba es un paquete Go** (`tools/pkg/catalogtest`): construye en una carpeta temporal el índice,
    los manifiestos publicados, los paquetes y los 14 zips (16 desde la revisión de seguridad del 2026-11-11, y firmado
    con una clave efímera) de `tools/fixtures/data/`, y devuelve un `httptest.Server`. No se
    guardan zips ni catálogos generados en git; Creative-Ink lo llama desde sus tests.
13. **La licencia se decide antes de la primera publicación (M7)**, porque B6 exige `LICENSE` en cada bundle. En M8
    queda solo la identidad (texto de muestra).
    *Prerrequisito descubierto el 2026-11-11:* al importar el módulo `tools` con `go get`, Creative-Ink solo recibe lo
    que está dentro de `tools/`, y `go:embed` no admite `../`. Por eso los fixtures se movieron de `tests/fixtures/` a
    `tools/fixtures/data/` y el paquete `tools/fixtures` los incrusta con `embed.FS`.
14. **Licencias: MIT + MIT-0** (convención REUSE, `NOTICE.md`):
    - **MIT-0:** plantillas y `shared/`, para que los documentos de los usuarios no tengan que conservar avisos.
    - **MIT:** herramientas, librería, esquemas y documentación; es compatible con cualquier licencia que elija
      Creative-Ink.
    - **Exclusiones:** los logos institucionales y el material de terceros (la clase ISPRS no declara licencia: se
      resuelve en M5).
    - **Uso en el contrato:** `license` es una expresión SPDX y `inkt bundle` genera el `LICENSE` del bundle desde
      `LICENSES/`.
15. **Sin `--shell-escape` en v1** (revisión de seguridad del agente de Creative-Ink):
    - **En este repositorio:** `build.shellEscape: true` está prohibido (V10).
    - **En la carpeta:** V39 rechaza externalización de TikZ, `minted`, `\write18` y los marcadores `{{TITULO}}`.
    - **En los clientes:** bloquean la plantilla (`compat` → `blocked`).
    - **Por qué no basta con "requerir confianza":** en Creative-Ink una carpeta sin `.latexmkrc` ya es de confianza.
16. **Catálogo oficial firmado con Ed25519** (contrato §7.1):
    - **Qué se firma:** cada JSON del catálogo, con su `.sig`.
    - **Claves:** la pública va fija en la librería; la privada, en un secreto de GitHub Actions que custodia el
      humano.
    - **Terceros:** con clave declarada, firma obligatoria; sin ella, se muestran como "sin firma".
17. **Endurecimiento de rutas:** se rechazan las entradas que colisionan por mayúsculas (B10) y los rc de latexmk sin
    distinguir mayúsculas (B4).

### Plan por fases
- [x] **M0 — Gobierno** (`DEV-feat-Contrato-Plantillas`, hecho el 2026-11-10, commits `f83dcf8`…`20852d6`): CLAUDE.md,
  MAP.md, CONTRIBUTING.md, README.md, `.gitattributes`, `.gitignore`, `.editorconfig`, skills `latex-template-author` y
  `template-contract`.
- [x] **M1 — Contrato v1** (hecho el 2026-11-11, commits `dba68f5`…`6513b66`): `docs/contrato-plantillas.md`, `schema/*.json` (4), `packs/*.json` (4), `tools/fixtures/data/`
  (fuentes y casos), `docs/guia-nueva-plantilla.md`. *Ajuste del 2026-11-10:* los zips y el catálogo de prueba no se
  escriben a mano; los genera M2 con el mismo código que publica el catálogo real (reproducibles).
- [ ] **M2 — Librería Go:** `tools/pkg/{contract,params,compat,bundle,client}` + `inkt validate|index|bundle`, con tests,
  `-race` y `go vet` para darwin, linux y windows. Incluye `pkg/catalogtest` (decisión 12) y los tests que recorren
  todos los casos de `tools/fixtures/data/`.
  - [x] Paso 1 (2026-11-11): `go.mod` (`go 1.22`), `fixtures` (embed), `internal/fixturetest`, `pkg/contract` (tipos,
    errores compartidos, V01–V20 con línea de cada hallazgo, V50–V52). 47/47 casos de manifiesto.
  - [x] Paso 2 (2026-11-11): `pkg/params` (`ParseAssignments`, `Resolve`, `RenderInkParams`, `EscapeTeX`). 10/10 casos
    de resolución, 2 salidas exactas de `ink-params.tex` y 9 de escapado.
  - [x] Paso 3 (2026-11-11): `pkg/compat` (`Evaluate`, `EvaluateJSON`, `UpdateAvailable`). Severidad: incompatible >
    blocked > missing-engine > missing-packages > ready. `--shell-escape` siempre da `blocked`. 7 tests.
  - [x] Paso 4 (2026-11-15): `internal/validate.Folder` (manifiesto + V30–V39) y la lista única de archivos permitidos
    en `contract/files.go`. 26 de 27 casos en macOS; `colision-mayusculas` se omite en sistemas de archivos que no
    distinguen mayúsculas (se prueba en Linux y, para bundles, B10 en todos).
  - [x] Paso 5 (2026-11-15): `pkg/bundle` (`Collect`, `Build`, `WriteZip` reproducible, `LICENSE` desde `LICENSES/`,
    `SHA256`, `VerifySHA256`, `SafeExtract` B1–B11 en carpeta temporal y `rename`). 16/16 casos, más tamaño falso,
    cancelación, destino ocupado, id o versión inesperados y `shellEscape`.
  - [ ] Paso 6: `pkg/catalogtest` + generación del índice + firma Ed25519 de cada JSON con clave efímera.
  - [ ] Paso 7: `pkg/client` (firma obligatoria en el oficial, `sha256`, caché ETag, sin conexión, `Get`, `GetPack`).
  - [ ] Paso 8: `cmd/inkt` (`validate`, `bundle`, `index`, `sign`, `keygen`) y binario de comprobación.
- [ ] **M3 — Componentes compartidos:** motor, base, paletas, comandos, institución UNAM-FES con logos.
- [ ] **M4 — Piloto `ensayo`** de punta a punta (el humano lo compila con los tres motores).
- [ ] **M5 — Migración del resto** (§5).
- [ ] **M6 — Tesis** + kit de seguimiento + `inkt check|format|diff|new`.
- [x] **Licencia** (decisiones 13 y 14, hecho el 2026-11-11): `LICENSE`, `LICENSES/{MIT,MIT-0}.txt`, `NOTICE.md`,
  `license` como expresión SPDX en el contrato y en los fixtures. *Queda para M5:* el `LicenseRef` de la clase ISPRS.
- [ ] **M7 — CI y publicación:** `validate.yml` (3 SO) y `release.yml` (Pages + Releases).
- [ ] **M8 — Identidad:** texto de muestra con el nombre y la descripción del repositorio en cada plantilla.

## 4.9 Deuda técnica heredada (detectada el 2026-11-10 al analizar los repositorios de origen)

Se resuelve al migrar cada plantilla; se tacha con fecha, nunca se borra.

1. **El `COLOR` por defecto sobrescribe la configuración del documento** (Makefile de Trabajos-Maestría, `_compile`
   inyecta `\def\ColorTema{Institucional}`). Impacto: usuario. Decisión: parámetros en `ink-params.tex` (§4.3). Estado: pendiente.
2. **Imágenes resueltas por profundidad de carpeta** (`\graphicspath{../../../Img/}`) y logos fuera de la plantilla.
   Impacto: una plantilla descargada sola no compila. Decisión: componentes resueltos dentro del bundle. Estado: pendiente.
3. **Logos PDF en Git LFS** (`Img/UNAM_crest_*.pdf`). Impacto: zips con punteros. Decisión: sin LFS. Estado: pendiente.
4. **Paletas copiadas 7 veces y temas divergentes** (`BeamerTheme.tex` con 3 variantes). Decisión: `shared/`. Estado: pendiente.
5. **Construcción solo POSIX/macOS** (`sed -i ''`, `rm`, `find -delete`, `/Library/TeX/texbin`, `python3.10`).
   Impacto: Windows y Linux. Decisión: `inkt` en Go. Estado: pendiente.
6. **La plantilla Apimetro no tiene paleta `Institucional`** y falla con `make Document`. Decisión: `enum` propio en su
   manifiesto. Estado: pendiente.
7. **`\FirmaBloqueDoble` de CartaUNAM mezcla sus argumentos** (imprime `#2 — Visto Bueno` y `#3`). Estado: pendiente.
8. **El tipo de documento se adivina con `grep` de comentarios** (`compilar-limpio`). Decisión: el manifiesto lo
   declara. Estado: pendiente.
9. **`bibtex main || true` oculta errores.** Decisión: latexmk/Creative-Ink reportan el fallo como resultado. Estado: pendiente.
10. **ISPRS usa `subfigure`, `epsfig`, `times` y una figura EPS** (requiere epstopdf). Decisión: modernizar o declarar
    el requisito. Estado: pendiente.
    *Además (2026-11-11):* `isprs.cls` e `isprs.bst` son © Carsten Steger e ISPRS y **no declaran licencia**: no
    pueden ir como MIT-0. Al migrar, decidir entre redistribuirlos como `LicenseRef-isprs` con aviso o pedir que se
    descarguen de ISPRS.
11. **`lipsum` en las plantillas e infografía con `\vspace` fijos** para librar la cabecera. Decisión: texto de muestra
    propio (M8) y medidas derivadas de la cabecera. Estado: pendiente.
12. **Documentación que contradice al código** (instrucciones de color, README con carpetas inexistentes, CartaUNAM
    que se "compila" con el objetivo de scaffolding). Decisión: README por plantilla validado contra el manifiesto. Estado: pendiente.
13. **Tesis: BibTeX/natbib frente a Biber/biblatex según el archivo** (`CLAUDE.md`, README y CI de TesisUrbanismo).
    Decisión: decisión 6 de §4.8. Estado: pendiente.
14. **Tesis: externalización de TikZ con `--shell-escape` obligatorio.** ~~Decisión: opcional y apagada
    (`shellEscape: false`).~~ **Decisión del 2026-11-11 (decisión 15):** sin externalización en v1, porque
    `--shell-escape` está bloqueado. La tesis compila sus diagramas TikZ en cada pasada; si el tiempo de compilación
    lo exige, se estudiará más adelante una alternativa sin shell (por ejemplo, PDF de los diagramas generados aparte).
    Estado: pendiente (se aplica en M6).
15. **Tesis: `make solo` renombra `tesis.pdf` y borra el PDF completo.** Decisión: compatible con el `-jobname` de
    Creative-Ink. Estado: pendiente.
16. **Nombres con espacios, acentos o erratas** (`Instructions Plantilla *.md`, `Comands.tex`, `SociologíaUrbana.md`).
    Decisión: ASCII en kebab-case. Estado: pendiente.

## 5. Catálogo inicial a migrar

| Origen | id | Categoría | Parámetros clave | Paquetes | Fase |
|--------|----|-----------|------------------|----------|------|
| `DocumentosPlantilla/Ensayos/{ConPortada,SinPortada}` | `ensayo` | `academico` | `portada` (bool), `color` | academico | M4 |
| `DocumentosPlantilla/Articulos/{ConPortada,SinPortada}` | `articulo` | `academico` | `portada` (bool), `palabrasClave` (list), `color` | academico | M5 |
| `DocumentosPlantilla/Articulos/RegistroDiff` | `registro-avances` | `academico` | repositorio, versiones de referencia | academico | M5 |
| `TesisUrbanismo` | `tesis` | `academico` | grado, programa, tutor, comité (list), `bibliografia` (sin externalización de TikZ: decisión 15) | academico | M6 |
| `PresentacionPlantilla` | `presentacion-institucional` | `presentacion` | `color`, profesor, semestre | presentaciones | M5 |
| `DocumentosPlantilla/SotmLatam` | `presentacion-sotm-latam` | `presentacion` | afiliación | presentaciones | M5 |
| `InfografiaPlantilla` | `infografia` | `difusion` | `tamano` (A4Portrait, A3Portrait, A4Landscape), `color` | presentaciones | M5 |
| `DocumentosPlantilla/CartaUNAM` | `carta-institucional` | `correspondencia` | folio, firmante, `firmaDoble` (bool) | correspondencia | M5 |
| `DocumentosPlantilla/Apimetro` | `carta-apimetro` | `correspondencia` | folio, firmante, `color` (Apimetro, Negro) | correspondencia | M5 |
| `DocumentosPlantilla/ISPRS` | `isprs-abstract` | `congresos` | — | congresos | M5 |

Paquete `todo`: se genera al publicar con todas las plantillas.

## 6. Conocimiento requerido por área

| Área | Qué hay que dominar |
|------|---------------------|
| LaTeX | iftex, inputenc/fontenc frente a fontspec, babel `spanish`, natbib/apalike, biblatex/biber, beamer, tcolorbox, TikZ, `\includeonly`, `\InputIfFileExists`, `\providecommand` |
| Go | `encoding/json` (y `json.Decoder.Token` para ubicar líneas), `archive/zip`, `crypto/sha256`, `crypto/ed25519`, `embed`, `io/fs`, `net/http` (timeouts, ETag, proxy), `context`, `path/filepath` frente a `path`, `httptest`, `-race`, `GOOS` |
| Seguridad | Zip slip, bombas zip, colisiones por mayúsculas en macOS/Windows, `.latexmkrc` como código Perl, `--shell-escape`, firma de catálogos y custodia de claves |
| Publicación | GitHub Releases (assets), GitHub Pages con Actions, tags anotados |
| Creative-Ink | Su `docs/MAP.md` §1, §4.3.6 (`ci`) y §4.6 (configuración y confianza), y `internal/core/templates.go` |

## 7. Comandos que ejecuta el humano

| Acción | Comando |
|--------|---------|
| Ramas | `git switch -c DEV-feat-<Tema>` desde `DEV`; integrar con `git switch DEV && git merge --ff-only DEV-feat-<Tema>` |
| Commits | Los que la IA propone, con `git add <archivos>` explícitos |
| Compilar una plantilla en local | `latexmk -pdf main.tex` (o `-xelatex`, `-lualatex`) dentro de `src/` |
| Paquetes de TeX | `tlmgr install <paquetes>` |
| Verificar la librería | `cd tools && gofmt -l . && go vet ./... && go test -race ./...` (también los puede ejecutar la IA, con permiso) |
| Clave de firma del catálogo (⏳ M7) | Generarla con `inkt keygen` (paso 8), guardar la privada **solo** en un secreto de GitHub Actions y borrarla del disco |
| Publicar una versión (⏳ M7) | `git tag -a <id>-v<versión> -m "…"` y `git push origin <id>-v<versión>` |
| Activar GitHub Pages (⏳ M7) | Settings → Pages → Source: GitHub Actions |
