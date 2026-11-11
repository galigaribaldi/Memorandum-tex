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
├── .gitattributes                 ← LF en todo; binarios sin conversión; sin LFS
├── .gitignore                     ← auxiliares de LaTeX, dist/, ajustes personales
├── .editorconfig                  ← UTF-8, LF, 2 espacios (tabs en Go)
├── .claude/
│   ├── settings.json              ← permisos: sin borrar, sin commit/push/tags; leer y escribir con permiso
│   └── skills/
│       ├── latex-template-author/SKILL.md ← crear o migrar plantillas y componentes, con lista de verificación
│       └── template-contract/SKILL.md     ← manifiesto, índice, paquetes, librería y catálogo
├── docs/
│   ├── MAP.md                     ← este archivo
│   ├── integracion-creative-ink.md ← plan y requerimientos para el agente de Creative-Ink (contrato v0)
│   ├── contrato-plantillas.md     ← ⏳ M1: especificación v1 (reemplaza a §4 de la integración)
│   └── guia-nueva-plantilla.md    ← ⏳ M1: paso a paso para colaboradores
├── schema/                        ← ⏳ M1
│   ├── template.schema.json       ← autocompletado y validación del manifiesto en el editor
│   ├── index.schema.json
│   └── pack.schema.json
├── packs/                         ← ⏳ M1: <pack-id>.json (listas de ids)
├── shared/                        ← ⏳ M3: componentes que se resuelven dentro de cada bundle
│   ├── motor/ink-motor.tex        ← iftex: inputenc/fontenc (pdfLaTeX) o fontspec (Xe/LuaLaTeX)
│   ├── base/ink-base.tex          ← carga de ink-params, \inkIfTrue, \inkItem
│   ├── paletas/                   ← Institucional, Teal, VerdeOlivo, Purpura, Rojo, VerdeEsmeralda
│   ├── comandos/                  ← base, academico, revision, codigo
│   └── instituciones/unam-fes-acatlan/ ← textos institucionales y logos
├── templates/                     ← ⏳ M4–M6: templates/<categoria>/<id>/{template.json,README.md,CHANGELOG.md,preview.png,src/}
├── tools/                         ← ⏳ M2: módulo Go (solo biblioteca estándar)
│   ├── go.mod
│   ├── cmd/inkt/                  ← CLI: validate, index, bundle, new, check, format, diff, clean
│   └── pkg/{contract,params,compat,bundle,client}/ ← librería que importa Creative-Ink
├── tests/fixtures/catalog/v1/     ← ⏳ M1: catálogo y bundles de prueba (incluye casos maliciosos)
└── .github/workflows/             ← ⏳ M7: validate.yml (3 SO) y release.yml (Pages + Releases)
```

## 2. Reglas de capas

```
 templates/<cat>/<id>/src   ──declara──►  template.json (components, parameters, build)
          │                                     │
          │  solo referencia lo suyo y          │ inkt bundle (M2) resuelve los componentes
          │  los componentes declarados         ▼
 shared/*  ─────────copiados a──────────►  <id>-<versión>.zip  (autocontenido, Latex/ink/…)
                                                 │
                                    inkt index   ▼
                            catalog/v1/index.json + manifiestos publicados (Pages)
                                                 │
                                                 ▼
                   Creative-Ink (importa tools/pkg/*; decide rutas, comandos y UI)
```

- Una plantilla **nunca** referencia archivos fuera de su carpeta: lo compartido entra solo a través de `components`.
- `shared/` no conoce a ninguna plantilla (sin condicionales por id).
- Los paquetes de `tools/pkg/*` importan **solo la biblioteca estándar**. `contract`, `params`, `compat` y `bundle` no
  usan la red ni ejecutan procesos; solo `client` usa la red.
- Creative-Ink consume el catálogo y la librería; nunca lee este repositorio con `git clone` ni con la API de GitHub.

## 3. Contrato (tabla viva)

Mientras no exista `docs/contrato-plantillas.md` (⏳ M1), el detalle está en `docs/integracion-creative-ink.md` §4.

| Artefacto | Productor → consumidor | Forma y semántica | Fase |
|-----------|------------------------|-------------------|------|
| `template.json` | Autor de la plantilla → `inkt`, Creative-Ink | Manifiesto v1: `id` kebab-case, `version` semver, `build` (`engines`, `defaultEngine: pdflatex`, `bibliography: bibtex`, `shellEscape`), `requires`, `parameters` (`string`, `boolean`, `enum`, `date`, `list`), `modes`. Campos desconocidos se ignoran | M1 |
| Manifiesto publicado | CI → Creative-Ink | El anterior + `dist {url, sha256, size, publishedAt}` en `catalog/v1/templates/<id>/<versión>/template.json` | M7 |
| `index.json` | CI → Creative-Ink | Resumen ligero (≤ 1 MB) de plantillas y paquetes; rutas relativas a la URL del índice | M1 (fixture), M7 (real) |
| `packs/<id>.json` | Repositorio → Creative-Ink | Lista de `{id, version?}`; sin zip propio. `todo` se genera al publicar | M1 |
| Bundle `<id>-<versión>.zip` | `inkt bundle` → librería | Una carpeta raíz `<id>/`; autocontenido; sin `.latexmkrc`, enlaces, `..`, ocultos (salvo `.creative-ink-meta`) ni extensiones fuera de la lista; ≤ 2000 archivos y 200 MB | M2 |
| `ink-params.tex` | Librería → plantilla | Un `\def\<macro>{valor}` por parámetro, escapado para TeX; se carga antes de los `\providecommand` | M1 |
| Macros de modo | Creative-Ink → plantilla | `\borrador` y `\SoloCapitulo{<ruta>}` (nombres fijos, los usa `ci compile`) | M3 |
| Librería Go | Este repo → Creative-Ink | `client.Index/Manifest/Get/GetPack`, `params.Resolve/RenderInkParams/EscapeTeX`, `compat.Evaluate`, `bundle.SafeExtract/VerifySHA256`; errores tipados | M2 |

## 4. Subsistemas

### 4.1 Plantillas (⏳ M4)
Anatomía: `templates/<categoria>/<id>/` con `template.json`, `README.md` (uso, parámetros, modos), `CHANGELOG.md`
(semver propio), `preview.png` (≤ 200 KB) y `src/` (lo que se copia al proyecto: `main.tex`, `Latex/`, `secciones/`,
`referencias.bib`, `img/`). El archivo principal es `main.tex` salvo que `entry` diga otra cosa.

### 4.2 Componentes compartidos (⏳ M3)
Fuente única de paletas, motor, comandos e instituciones. En el repositorio viven una sola vez; `inkt bundle` copia los
declarados a `src/Latex/ink/` del bundle. Una institución nueva es una carpeta en `shared/instituciones/` con su
`institucion.tex` (textos) y `logos/`; ninguna plantilla escribe a mano "UNAM" o "FES Acatlán".

### 4.3 Parámetros y modos (⏳ M1)
Variantes como booleanos (`portada`), opciones como `enum` (`color`, `tamano`), datos como `string`/`date`/`list`.
Orden obligatorio en `main.tex`: `\InputIfFileExists{ink-params}{}{}` → `\providecommand` con los valores por
defecto → resto del preámbulo. Modos: `completo`, `rapido`, `borrador`, `capitulo` (solo los que la plantilla declara).

### 4.4 Motores y bibliografía (⏳ M3)
`ink-motor.tex` decide con iftex. Por defecto pdfLaTeX + natbib/apalike + BibTeX; XeLaTeX y LuaLaTeX donde
`build.engines` los liste; biblatex + Biber solo como parámetro `bibliografia` de las plantillas que lo ofrezcan.

### 4.5 Herramienta `inkt` y librería (⏳ M2)
Sustituye a los dos Makefiles y a los scripts de Python de los repositorios de origen. Comandos previstos: `validate`,
`index`, `bundle`, `new`, `check bib|labels|refs`, `format`, `diff`, `clean`. La librería se diseña en
`docs/integracion-creative-ink.md` §5.

### 4.6 Publicación y catálogo (⏳ M7)
Tag `<id>-v<versión>` (lo crea el humano) → CI valida, empaqueta, sube el zip a la Release y regenera
`catalog/v1/` en GitHub Pages. Escalabilidad estimada en `docs/integracion-creative-ink.md` §3.2.

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
8. Software libre; licencia (probablemente MIT) e identidad (texto de muestra con el nombre y la descripción del
   repositorio) se deciden al final, en M8.
9. Solo español por ahora.
10. Kit de la tesis: issues, CHANGELOG, CONTEXTO y revisiones por defecto; CI y `CLAUDE.md` opcionales.
11. En Creative-Ink: catálogo en el **núcleo** y **catálogos de terceros** marcados como tales.

### Plan por fases
- [ ] **M0 — Gobierno** (`DEV-feat-Contrato-Plantillas`): CLAUDE.md, MAP.md, CONTRIBUTING.md, README.md, `.gitattributes`,
  `.gitignore`, `.editorconfig`, skills `latex-template-author` y `template-contract`.
- [ ] **M1 — Contrato v1:** `docs/contrato-plantillas.md`, `schema/*.json`, `packs/*.json`, `tests/fixtures/catalog/v1/`,
  `docs/guia-nueva-plantilla.md`.
- [ ] **M2 — Librería Go:** `tools/pkg/{contract,params,compat,bundle,client}` + `inkt validate|index|bundle`, con tests,
  `-race` y `go vet` para darwin, linux y windows.
- [ ] **M3 — Componentes compartidos:** motor, base, paletas, comandos, institución UNAM-FES con logos.
- [ ] **M4 — Piloto `ensayo`** de punta a punta (el humano lo compila con los tres motores).
- [ ] **M5 — Migración del resto** (§5).
- [ ] **M6 — Tesis** + kit de seguimiento + `inkt check|format|diff|new`.
- [ ] **M7 — CI y publicación:** `validate.yml` (3 SO) y `release.yml` (Pages + Releases).
- [ ] **M8 — Identidad y licencia.**

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
11. **`lipsum` en las plantillas e infografía con `\vspace` fijos** para librar la cabecera. Decisión: texto de muestra
    propio (M8) y medidas derivadas de la cabecera. Estado: pendiente.
12. **Documentación que contradice al código** (instrucciones de color, README con carpetas inexistentes, CartaUNAM
    que se "compila" con el objetivo de scaffolding). Decisión: README por plantilla validado contra el manifiesto. Estado: pendiente.
13. **Tesis: BibTeX/natbib frente a Biber/biblatex según el archivo** (`CLAUDE.md`, README y CI de TesisUrbanismo).
    Decisión: decisión 6 de §4.8. Estado: pendiente.
14. **Tesis: externalización de TikZ con `--shell-escape` obligatorio.** Decisión: opcional y apagada (`shellEscape:
    false`). Estado: pendiente.
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
| `TesisUrbanismo` | `tesis` | `academico` | grado, programa, tutor, comité (list), `bibliografia`, `tikzExterno` (bool) | academico | M6 |
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
| Go | `encoding/json`, `archive/zip`, `crypto/sha256`, `net/http` (timeouts, ETag, proxy), `context`, `path/filepath` frente a `path`, `httptest`, `-race`, `GOOS` |
| Publicación | GitHub Releases (assets), GitHub Pages con Actions, tags anotados |
| Creative-Ink | Su `docs/MAP.md` §1, §4.3.6 (`ci`) y §4.6 (configuración y confianza), y `internal/core/templates.go` |

## 7. Comandos que ejecuta el humano

| Acción | Comando |
|--------|---------|
| Ramas | `git switch -c DEV-feat-<Tema>` desde `DEV`; integrar con `git switch DEV && git merge --ff-only DEV-feat-<Tema>` |
| Commits | Los que la IA propone, con `git add <archivos>` explícitos |
| Compilar una plantilla en local | `latexmk -pdf main.tex` (o `-xelatex`, `-lualatex`) dentro de `src/` |
| Paquetes de TeX | `tlmgr install <paquetes>` |
| Publicar una versión (⏳ M7) | `git tag -a <id>-v<versión> -m "…"` y `git push origin <id>-v<versión>` |
| Activar GitHub Pages (⏳ M7) | Settings → Pages → Source: GitHub Actions |
