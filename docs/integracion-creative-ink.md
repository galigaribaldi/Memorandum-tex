# Integración Memorandum-tex ↔ Creative-Ink: plan y requerimientos

> **Estado:** v1, revisado de punta a punta el 2026-11-12, con M0 y M1 hechos y M2 en curso (pasos 1–3 de 8). Lo
> escribe el lado de Memorandum-tex para que el agente de IA de Creative-Ink diseñe la descarga de plantillas.
> **Precedencia:**
> - **Formato del catálogo, manifiestos, bundles y firma:** manda **`docs/contrato-plantillas.md` (contrato v1)**.
>   Este documento lo resume y explica cómo usarlo desde Creative-Ink.
> - **Lo que pasa dentro de Creative-Ink** (rutas, IPC, UI, comandos `ci`, seguridad del editor): mandan su
>   `CLAUDE.md` y su `docs/MAP.md`. Aquí solo hay **requerimientos y propuestas** para ese lado.
> **Cómo leerlo:** §5 dice qué está implementado y qué falta en la librería; §8 son las reglas de seguridad; §12, las
> decisiones del programador; §14, la bitácora de cambios.
> **Reglas:** este documento es para analizar. No implementes nada en Creative-Ink sin la aprobación de su
> programador y sigue su ciclo por secciones (análisis → decisiones → implementación → verificación → Tríada).

---

## 1. Filosofía: descarga bajo demanda

- **Nada se instala de golpe.** Creative-Ink no trae plantillas de Memorandum-tex incluidas ni las descarga
  todas al arrancar. El usuario **consulta el catálogo** y **descarga solo lo que necesita**.
- **Dos formas de descarga:**
  1. **Individual:** una plantilla (la última versión o una concreta).
  2. **Por paquete:** un conjunto con nombre (p. ej. `academico`), que la librería resuelve en descargas
     individuales.
- **Personalizar antes de usar:** el manifiesto de cada plantilla (un JSON pequeño) se consulta **sin descargar el
  bundle**. Declara qué se puede cambiar (parámetros, motores, modos) y con qué valores válidos. Así Creative-Ink
  puede preguntar o recibir `clave=valor` antes de descargar, y la librería aplica esos valores al extraer.
- **La terminal integrada (Terminal CI) es la interfaz principal.** Una galería en la UI es opcional y posterior.
- **Dónde se guarda lo descargado lo decide Creative-Ink**, dentro de su repositorio central
  `My-Creative-Ink-Repository` (sus §1 y §4.6). Memorandum-tex no impone rutas: la librería recibe siempre un
  destino absoluto.

## 2. Reparto de responsabilidades

| Memorandum-tex (este repositorio) | Creative-Ink |
|-----------------------------------|--------------|
| Contrato v1: `template.json`, `index.json`, paquetes, bundle, `ink-params.tex` | Dónde y cómo se guarda cada plantilla dentro de `My-Creative-Ink-Repository` |
| Las plantillas, sus manifiestos, README, CHANGELOG y licencia | Comandos `ci` (registro), ayuda, guía y Tab |
| La **librería Go** (`tools/pkg/…`): leer el catálogo, validar, resolver parámetros, evaluar compatibilidad, descargar, verificar y extraer de forma segura | Detectar motores y paquetes de TeX (`system.FindBinary`, `kpsewhich`) y pasárselos a la librería |
| Publicar: GitHub Pages (catálogo **firmado**) y Releases (bundles) | Política de confianza (confiar solo en las plantillas propias del usuario; nunca en algo descargado) y claves de configuración |
| Fixtures de prueba sin red (`tools/fixtures`) y un catálogo de prueba servible (`pkg/catalogtest`) | IPC, eventos y, si se decide, la galería en la UI |
| Validador del contrato en CI; custodia de la clave de firma (el humano, en un secreto de GitHub Actions) | Importar el módulo: `go.work` local mientras se desarrolla y `go get` al publicar (lo ejecuta el programador) |

## 3. Distribución y escalabilidad

### 3.1 Dónde vive cada cosa

```
Catálogo (GitHub Pages, estático y con CDN)
https://galigaribaldi.github.io/Memorandum-tex/
├── index.html                                   ← galería para humanos (más adelante)
└── catalog/v1/                                  ← "v1" = schemaVersion del contrato
    ├── index.json        + index.json.sig       ← resumen ligero: plantillas + paquetes
    ├── templates/<id>/versions.json      + .sig ← versiones publicadas de esa plantilla
    ├── templates/<id>/<version>/template.json + .sig ← manifiesto publicado (incluye "dist")
    ├── templates/<id>/<version>/preview.png     ← miniatura (≤ 200 KB, sin firma: solo se muestra)
    └── packs/<pack-id>.json              + .sig ← definición de un paquete

Cada `<archivo>.sig` es la firma Ed25519 de los bytes exactos del JSON (contrato §7.1).

Bundles (GitHub Releases, un tag por plantilla y versión)
https://github.com/galigaribaldi/Memorandum-tex/releases/download/<id>-v<version>/<id>-<version>.zip
```

- Las rutas dentro de `index.json` son **relativas a la URL del índice**. La librería las resuelve. Un catálogo
  de terceros puede alojarse en cualquier servidor estático con la misma forma.
- **El cliente nunca usa `api.github.com`:** la API REST sin autenticación limita a 60 peticiones por hora. Pages y
  las descargas de Releases son archivos estáticos sin ese límite.

### 3.2 Cómo escala (estimaciones que se ajustan con datos)

| Recurso | Límite conocido | Uso estimado | Margen |
|---------|-----------------|--------------|--------|
| `index.json` | La librería acepta 1 MB como máximo | ~600 B por plantilla → 1000 plantillas ≈ 600 KB (Pages lo sirve comprimido, ~100 KB) | Amplio. Si se acerca, se parte por categoría (`index-<categoria>.json`) sin romper v1 |
| Sitio de Pages | 1 GB publicado, ~100 GB/mes de tráfico (límite blando) | Solo JSON y miniaturas: < 50 MB con cientos de plantillas | Amplio, porque los zips no pasan por Pages |
| Assets de Releases | 2 GB por archivo | Un bundle típico < 5 MB; tesis < 50 MB (sin imágenes pesadas) | Amplio. La librería limita a 200 MB, como `copyTemplate` |
| Peticiones por consulta | — | `list` = 2 (índice y su `.sig`; con ETag, 304 si no cambió); `info` = 2; descarga = 2 + 1 zip | Mínimo |

## 4. Contrato v1 (borrador)

> **Reemplazado el 2026-11-10 por `docs/contrato-plantillas.md` (M1), que es ahora la fuente de verdad. No
> implementes a partir de esta sección:** se conserva solo como contexto del análisis y sus ejemplos tienen valores
> viejos (`"license": "MIT"`, `shellEscape` como opción). Las diferencias están en §14 del contrato. Las que más
> afectan a Creative-Ink:
> - **Bundle:** `.creative-ink-meta` lo genera el bundle con solo `version` y `mainFile`.
> - **Versiones:** existe `versions.json`.
> - **Errores de parámetros:** tienen códigos estables (`unknown`, `required`, `type`, `enum`, `length`, `items`).
> - **Rutas del zip:** se rechazan `C:`, `\` y las colisiones por mayúsculas en todos los sistemas.
> - **Licencia:** las plantillas son `MIT-0`.
> - **`shellEscape: true`:** está **bloqueado**.
> - **Firma:** cada JSON del catálogo va firmado (§7.1).

### 4.1 Manifiesto `template.json` (en el repositorio, por plantilla)

```json
{
  "$schema": "../../../schema/template.schema.json",
  "schemaVersion": 1,
  "id": "ensayo",
  "version": "1.0.0",
  "name": "Ensayo académico",
  "description": "Ensayo con portada institucional opcional, índice y bibliografía APA.",
  "category": "academico",
  "tags": ["ensayo", "apa", "unam"],
  "language": "es-MX",
  "license": "MIT",
  "authors": [{ "name": "…", "github": "…" }],
  "entry": "main.tex",
  "build": {
    "engines": ["pdflatex", "xelatex", "lualatex"],
    "defaultEngine": "pdflatex",
    "bibliography": "bibtex",
    "shellEscape": false
  },
  "requires": {
    "creativeInk": ">=0.1.0",
    "texLive": 2023,
    "packages": ["babel-spanish", "natbib", "tcolorbox"],
    "files": ["natbib.sty", "tcolorbox.sty", "spanish.ldf"]
  },
  "parameters": [
    { "key": "titulo",  "macro": "TituloDoc",  "type": "string",  "required": true, "fromProjectName": true, "label": "Título" },
    { "key": "autor",   "macro": "Autor",      "type": "string",  "required": true, "label": "Autor" },
    { "key": "portada", "macro": "ConPortada", "type": "boolean", "default": true, "label": "Con portada" },
    { "key": "color",   "macro": "ColorTema",  "type": "enum",    "default": "Institucional",
      "values": ["Institucional", "Teal", "VerdeOlivo", "Purpura", "Rojo", "VerdeEsmeralda"] }
  ],
  "modes": ["completo", "rapido", "borrador"],
  "preview": "preview.png"
}
```

- **`id`:** kebab-case ASCII (`^[a-z0-9]+(-[a-z0-9]+)*$`), único en el catálogo. Es también el nombre de la
  carpeta dentro del bundle. Compatible con `core.Slugify`.
- **`version`:** semver propio de cada plantilla. Tag de publicación: `<id>-v<version>`.
- **`build.defaultEngine`:** **`pdflatex`** en todas las plantillas (decisión del 2026-11-10). XeLaTeX y LuaLaTeX
  son opcionales donde `engines` los liste. **`bibliography`:** `bibtex` (natbib/apalike) por defecto; `biber`
  solo si la plantilla lo ofrece como parámetro.
- **`requires.files`:** nombres de archivo para `kpsewhich`, que es lo que Creative-Ink puede comprobar.
  `packages` son los nombres de `tlmgr` que se muestran al usuario si falta algo.
- **`parameters[].type`:**
  - `string`: una línea, 300 caracteres como máximo.
  - `boolean`: `true` o `false`.
  - `enum`: uno de `values`.
  - `date`: `AAAA-MM-DD`.
  - `list`: lista de `string`, por ejemplo el comité de una tesis.
- **`parameters[].macro`:** solo letras (`^[A-Za-z]+$`). El validador rechaza las que chocan con macros del núcleo
  de LaTeX (`title`, `author`, `date`…).
- **`fromProjectName: true`:** si el usuario no da valor, se usa el nombre del proyecto. Equivale al
  `{{TITULO}}` que Creative-Ink ya sustituye.
- **Campos desconocidos:** se ignoran (compatibilidad hacia adelante). Una `schemaVersion` desconocida hace que la
  plantilla sea `incompatible`, nunca un error del catálogo entero.

### 4.2 Manifiesto publicado (`catalog/v1/templates/<id>/<version>/template.json`)

Es el mismo de §4.1 con un bloque `dist` que añade la publicación (el repositorio no lo contiene, porque el hash se
calcula al empaquetar):

```json
"dist": {
  "url": "https://github.com/galigaribaldi/Memorandum-tex/releases/download/ensayo-v1.0.0/ensayo-1.0.0.zip",
  "sha256": "…",
  "size": 482113,
  "publishedAt": "2026-11-20T18:00:00Z"
}
```

### 4.3 Índice `catalog/v1/index.json`

```json
{
  "schemaVersion": 1,
  "generatedAt": "2026-11-20T18:00:00Z",
  "source": { "id": "memorandum-tex", "name": "Memorandum-tex", "homepage": "https://github.com/galigaribaldi/Memorandum-tex" },
  "templates": [
    {
      "id": "ensayo", "name": "Ensayo académico", "category": "academico",
      "description": "…", "tags": ["ensayo", "apa"], "license": "MIT", "language": "es-MX",
      "latest": "1.0.0",
      "engines": ["pdflatex", "xelatex", "lualatex"], "defaultEngine": "pdflatex", "bibliography": "bibtex",
      "shellEscape": false,
      "requires": { "creativeInk": ">=0.1.0" },
      "manifest": "templates/ensayo/1.0.0/template.json",
      "versions": "templates/ensayo/versions.json",
      "preview": "templates/ensayo/1.0.0/preview.png",
      "dist": { "url": "…", "sha256": "…", "size": 482113 }
    }
  ],
  "packs": [
    { "id": "academico", "name": "Paquete académico", "description": "…", "manifest": "packs/academico.json" }
  ]
}
```

### 4.4 Paquetes (`packs/<pack-id>.json`)

```json
{
  "schemaVersion": 1,
  "id": "academico",
  "name": "Paquete académico",
  "description": "Ensayo, artículo, registro de avances y tesis.",
  "templates": [
    { "id": "ensayo" },
    { "id": "articulo" },
    { "id": "registro-avances" },
    { "id": "tesis", "version": "1.0.0" }
  ]
}
```

- Sin `version` se usa la última (`latest`). Un paquete **no tiene un zip propio**: la librería descarga cada
  plantilla por separado (sin duplicados y con concurrencia acotada). Así el paquete y la descarga individual son
  el mismo mecanismo y no hay dos copias del mismo contenido.
- El paquete `todo` (todas las plantillas) se genera automáticamente al publicar.
- Paquetes iniciales previstos: `academico`, `presentaciones`, `correspondencia`, `congresos`, `todo`.

### 4.5 Bundle (`<id>-<version>.zip`)

```
ensayo-1.0.0.zip
└── ensayo/                    ← una sola carpeta raíz = id (la librería la quita al extraer)
    ├── template.json          ← manifiesto de §4.1 (sin "dist")
    ├── LICENSE
    ├── README.md              ← uso, parámetros, modos
    ├── .creative-ink-meta     ← {"version":1,"mainFile":"main.tex"}: lo que ya entiende CreateProjectFromTemplate
    ├── main.tex
    ├── Latex/…                ← preámbulo de la plantilla
    ├── Latex/ink/…            ← componentes compartidos ya resueltos (motor, paletas, institución, logos)
    ├── secciones/…
    └── referencias.bib
```

- **Autocontenido:** no hay rutas `../` hacia fuera ni dependencias de profundidad de carpeta, y todos los logos
  van dentro. No hay punteros de Git LFS.
- **Prohibido en el bundle** (lo rechazan el validador al publicar y la librería al extraer):
  - `.latexmkrc`/`latexmkrc`, enlaces simbólicos, rutas absolutas o con `..`;
  - archivos ocultos salvo `.creative-ink-meta`;
  - extensiones fuera de la lista permitida: `tex sty cls bst bib bbx cbx lbx cfg def md txt json png jpg jpeg pdf
    eps svg otf ttf`.
- **Límites:** 2000 archivos y 200 MB descomprimidos, los mismos que `copyTemplate`.

### 4.6 `ink-params.tex` (lo genera la librería, lo puede editar el usuario)

```tex
% ink-params.tex — generado a partir de ensayo 1.0.0 (template.json). Puedes editarlo a mano.
\def\TituloDoc{Humedales urbanos y movilidad}
\def\Autor{Hernán Galileo Cabrera Garibaldi}
\def\ConPortada{false}
\def\ColorTema{Teal}
\def\ComiteTutor{\inkItem{Dra. A}\inkItem{Dr. B}}   % tipo list
```

- Los valores `string` se escapan para TeX (`\ { } $ & # ^ _ % ~`) con `params.EscapeTeX` de la librería. Debe
  coincidir con `core.EscapeTeX` de Creative-Ink; se comprobará con un test compartido de casos.
- Cada plantilla hace `\InputIfFileExists{ink-params}{}{}` **antes** de sus `\providecommand` por defecto. Por eso
  **compila igual sin el archivo**, ya sea a mano o con otro editor.
- Un booleano se compara con `\inkIfTrue{\ConPortada}{…}{…}` (definido en `Latex/ink/base/ink-base.tex`).

### 4.7 Convenciones LaTeX que Creative-Ink puede dar por hechas

- **Modos:** `\borrador` (graphicx en modo draft) y `\SoloCapitulo{<ruta>}` (`\includeonly`) con **los mismos
  nombres** que ya usa `ci compile borrador|solo`. Las plantillas los respetan si se definen por la línea de
  órdenes o en `ink-params.tex`.
- **Motor:** `Latex/ink/motor/ink-motor.tex` (con iftex) carga `inputenc`/`fontenc` en pdfLaTeX y `fontspec` en
  XeLaTeX o LuaLaTeX. Por eso las plantillas **también compilan con el `latexmk -xelatex` fijo** que Creative-Ink
  tiene hoy, mientras llega `compile.engine`.
- **Sin `--shell-escape`, nunca, en v1** (corregido el 2026-11-11 tras la revisión de seguridad de Creative-Ink).
  - **Qué decía antes:** que `shellEscape: true` marcaba la plantilla como "requiere confianza".
  - **Por qué no protege:** en Creative-Ink una carpeta sin `.latexmkrc` ya cuenta como de confianza, así que esa
    regla habría permitido ejecutar comandos desde un `.tex` descargado.
  - **Regla actual:** los clientes **bloquean** cualquier plantilla con `shellEscape: true` (estado `blocked`) y el
    validador de este repositorio la rechaza. La externalización de TikZ de la tesis sale del plan.
- Nombres de archivo en ASCII, kebab-case, sin espacios; fin de línea LF; UTF-8.

### 4.8 Versionado del contrato

- **Cambios compatibles:** campos nuevos opcionales, tipos de parámetro nuevos que los clientes viejos tratan como
  `string`. No suben `schemaVersion`.
- **Cambios incompatibles:** suben `schemaVersion` y se publican en `catalog/v2/` **sin borrar `v1`**, que sigue
  sirviendo a las versiones antiguas de Creative-Ink.

## 5. La librería Go (fase M2 de Memorandum-tex)

**Módulo:** `github.com/galigaribaldi/Memorandum-tex/tools`.
- **Solo biblioteca estándar**, para que Creative-Ink no herede dependencias.
- **Versión de Go:** `go` ≤ 1.26, la que usa Creative-Ink.
- **Versionado:** tags `tools/vX.Y.Z`, que crea el programador.
- **Desarrollo local:** un **`go.work` local, sin commit** (corregido el 2026-11-11; un `replace` subido al `go.mod`
  rompería la compilación en otras máquinas).
  - **Comando:** desde `Creative-Ink/creative-ink/`, el programador ejecuta
    `go work init . ../../Memorandum-tex/tools`.
  - **Git:** `go.work` y `go.work.sum` van al `.gitignore` de Creative-Ink.
  - **Al publicar:** en el `go.mod` queda solo `require github.com/galigaribaldi/Memorandum-tex/tools vX.Y.Z`.

### 5.1 Organización del repositorio (lo que le importa a Creative-Ink)

```
Memorandum-tex/
├── docs/contrato-plantillas.md   ← EL contrato v1 (fuente de verdad del formato)
├── docs/integracion-creative-ink.md ← este documento
├── schema/*.json                 ← JSON Schema (solo ayuda al editor; la validación real es la de pkg/contract)
├── packs/*.json                  ← paquetes del catálogo oficial
├── templates/<cat>/<id>/         ← ⏳ M4–M6: las plantillas
├── shared/                       ← ⏳ M3: componentes que se copian dentro de cada bundle
├── LICENSE, LICENSES/, NOTICE.md ← MIT (librería) y MIT-0 (plantillas)
└── tools/                        ← EL MÓDULO GO QUE IMPORTA CREATIVE-INK
    ├── go.mod                    ← github.com/galigaribaldi/Memorandum-tex/tools, go 1.22, sin dependencias
    ├── fixtures/                 ← ✅ datos de prueba incrustados (embed.FS): fixtures.FS(), fixtures.Dir()
    │   └── data/                 ← plantillas de prueba, casos de parámetros, casos inválidos, casos de bundles
    ├── internal/fixturetest/     ← ✅ ayudas de test internas (Creative-Ink no puede ni necesita importarlas)
    ├── internal/validate/        ← ✅ validación de carpetas de este repositorio (V30–V39); Creative-Ink no la usa
    ├── pkg/contract/             ← ✅ tipos, errores compartidos (contract.Err*), validación V01–V20 y V50–V52,
    │                                lista única de archivos permitidos (AllowedBundleFile, IsLatexmkrc)
    ├── pkg/params/               ← ✅ clave=valor → valores → ink-params.tex
    ├── pkg/compat/               ← ✅ ¿se puede usar en este equipo? (ready … blocked)
    ├── pkg/bundle/               ← ✅ empaquetar (reproducible), verificar sha256, extraer de forma segura
    ├── pkg/catalogtest/          ← ⏳ paso 6: catálogo de prueba firmado servido con httptest
    ├── pkg/client/               ← ⏳ paso 7: índice, manifiestos, firmas, caché, descargas individuales y paquetes
    └── cmd/inkt/                 ← ⏳ paso 8: CLI de este repositorio (validate, bundle, index, sign); Creative-Ink no la usa
```

### 5.2 Paquetes y estado

| Paquete | Responsabilidad | Estado | Procesos | Red |
|---------|-----------------|--------|----------|-----|
| `fixtures` | `FS()` con los datos de prueba incrustados; `Dir()` con su ruta en disco cuando existe | ✅ 2026-11-11 | No | No |
| `pkg/contract` | Tipos del contrato; errores compartidos; `ValidateManifest(data, Options) → (*Manifest, Report)`; `ValidatePack`; utilidades (`ValidID`, `ValidRelPath`, `ParseSemver`, `ParseLicense`…) | ✅ 2026-11-11 | No | No |
| `pkg/params` | `ParseAssignments`, `Resolve(m, Input) → (*Resolved, error)`, `RenderInkParams`, `EscapeTeX` | ✅ 2026-11-11 | No | No |
| `pkg/compat` | `Evaluate(m, Env) → Report`, `EvaluateJSON`, `UpdateAvailable` | ✅ 2026-11-11 | No: Creative-Ink le pasa `Env` | No |
| `pkg/bundle` | `VerifySHA256`, `SafeExtract`/`SafeExtractFile` con las reglas B1–B11; `Collect`/`Build`/`WriteZip` (los usa este repositorio) | ✅ 2026-11-15 | No | No |
| `pkg/catalogtest` | Solo para tests: construye y **firma** (con una clave efímera) el catálogo de prueba y lo sirve con `httptest.Server` | ⏳ paso 6 | No | Solo local (`127.0.0.1`) |
| `pkg/client` | `Index`, `Manifest`, `Versions`, `Pack`, `Get`, `GetPack`; **verificación de firmas**; caché con ETag; resultado sin conexión | ⏳ paso 7 | No | Sí (HTTPS) |

**API.** Las partes de `params`, `compat` y `bundle` son las **firmas reales**. La parte de `client` es una
**propuesta**: se cierra en el paso 7 y se actualizará aquí. Normalmente Creative-Ink no llamará a `bundle`
directamente, porque `client.Get` lo hará por él; queda disponible para extraer un zip que el usuario ya tenga.

```go
// IMPLEMENTADO (2026-11-15), firma real:
f, _ := os.Open(zipPath)
err := bundle.VerifySHA256(f, dist.SHA256)            // ANTES de extraer; contract.ErrChecksum si no coincide
res, err := bundle.SafeExtractFile(ctx, zipPath, dest, bundle.ExtractOptions{
    ExpectID: "ensayo", ExpectVersion: "1.0.0",         // lo que anunció el índice firmado
})                                                    // Limits cero = DefaultLimits (2000 archivos, 200 MB, 100:1 desde 1 MB)
// res = *bundle.Result{ID, Version, Manifest, Dir, MainFile, Files}
// Sobre el destino:
// - debe no existir o estar vacío; si no, contract.ErrDestExists y no se toca nada;
// - se extrae en una carpeta temporal hermana y se renombra al final: un fallo o un ctx cancelado no dejan nada;
// - ya trae .creative-ink-meta {version, mainFile}, LICENSE (MIT-0), README.md y template.json (sin $schema).
// Errores: contract.ErrUnsafePath | ErrForbiddenFile | ErrBundleLayout | ErrTooLarge | ErrDestExists | ctx.Err()
```

```go
// PROPUESTA (se cierra en el paso 7):
c := client.New(client.Options{
    Sources:   []client.Source{
        client.Official,  // fijo en la librería: URL del catálogo oficial + sus claves públicas (firma obligatoria)
        {ID: "otro", IndexURL: "https://…/catalog/v1/index.json", PublicKeys: nil}, // tercero sin clave = "sin firma"
    },
    HTTP:      httpClient,      // Creative-Ink: timeouts y proxy (http.ProxyFromEnvironment)
    CacheDir:  cacheDir,        // Creative-Ink decide (p. ej. <UserConfigDir>/creative-ink/cache/catalog)
    UserAgent: "creative-ink/<versión>",
})

idx, err := c.Index(ctx, client.IndexOptions{Refresh: false})
// idx.Offline == true si se usó la caché por falta de red; idx.FetchedAt dice de cuándo es.

m, err := c.Manifest(ctx, "ensayo", "")                 // "" = latest; solo JSON, sin bundle
// IMPLEMENTADO (2026-11-11), firma real:
values, err := params.ParseAssignments([]string{"portada=false", "color=Teal"}) // las palabras de la Terminal CI
vals, err := params.Resolve(m, params.Input{Values: values, ProjectName: "Humedales", Now: time.Now()})
// vals es *params.Resolved (en orden del manifiesto; vals.Map() da clave → string|bool|[]string)
// err es *params.Error{Fields: []FieldError{Key, Code, Message}}; Code ∈ unknown|required|type|enum|length|items
tex := params.RenderInkParams(vals) // []byte: el ink-params.tex exacto; params.FileName = "ink-params.tex"

// IMPLEMENTADO (2026-11-11), firma real:
rep := compat.Evaluate(m, compat.Env{
    Engines:     map[string]bool{"pdflatex": true, "xelatex": true}, // nil = no comprobar
    HasFiles:    func(names []string) map[string]bool { /* Creative-Ink: una llamada a kpsewhich */ },
    InkVersion:  "0.3.0",                                            // "" = no comprobar
})
// rep = compat.Report{Status, Reasons, MissingEngines, MissingFiles, Hints}
// Status ∈ incompatible | blocked | missing-engine | missing-packages | ready (gana el más grave)
// Hints son comandos que ejecuta el USUARIO ("tlmgr install …"); nunca se ejecutan solos
rep, m := compat.EvaluateJSON(publishedManifestJSON, env) // valida (OriginPublished) y evalúa; inválido = incompatible
compat.UpdateAvailable("1.0.0", "1.1.0")                   // true

// PROPUESTA (se cierra en el paso 7):
res, err := c.Get(ctx, client.GetRequest{
    ID: "ensayo", Version: "",          // individual
    Dest: absDir,                       // Creative-Ink decide: debe NO existir o estar vacía
    Params: vals,                       // nil = plantilla tal cual (sin ink-params.tex)
    Progress: func(done, total int64) {},
})
// res = {ID, Version, Dir, MainFile, DefaultEngine, Manifest, Files}

results, err := c.GetPack(ctx, client.PackRequest{ID: "academico", DestRoot: absRoot, Progress: …})
// results = []PackItem{ID, Version, Dir, Err}: el fallo de una plantilla es un resultado, no un error global
```

**Garantías de la librería** (cada una tendrá su test):
- **Descargas e instalación:**
  - Todo acepta `context.Context`: cancelar corta la descarga y no deja nada a medias. Se extrae en una carpeta
    temporal hermana de `Dest` y se hace `os.Rename` al final.
  - Nunca escribe fuera de `Dest`/`DestRoot` ni de `CacheDir`, y nunca sobrescribe: `Dest` existente y no vacío da
    `ErrDestExists`.
  - Verifica la **firma** de cada JSON de un catálogo firmado y el `sha256` del zip **antes** de extraer, y respeta
    límites: índice ≤ 1 MB, manifiesto ≤ 256 KB, bundle ≤ 200 MB.
- **Seguridad:** no ejecuta procesos ni sigue rutas peligrosas. Solo acepta HTTPS (HTTP solo con un flag explícito
  para tests con `httptest`). Nunca usa la red por su cuenta: solo cuando Creative-Ink llama a una función.
- **Errores tipados** (todos en `contract`, comparables con `errors.Is`): `ErrOffline`, `ErrNotFound`,
  `ErrChecksum`, `ErrSignature`, `ErrUnsafePath`, `ErrForbiddenFile`, `ErrBundleLayout`, `ErrTooLarge`,
  `ErrUnsupportedSchema`, `ErrDestExists`, además de `*params.Error` (contrato §13).
- **Concurrencia:** es segura (`go test -race`); `GetPack` usa como mucho 3 descargas a la vez.
- **Plataformas:** se verifica con `GOOS=windows`, `GOOS=linux` y `GOOS=darwin` (`go vet`).

## 6. Requerimientos para Creative-Ink

### 6.1 Funcionales

| # | Requerimiento | Notas |
|---|---------------|-------|
| RQ-1 | Listar el catálogo **sin descargar bundles** | 1 petición; filtros por categoría y búsqueda por texto |
| RQ-2 | Ver el detalle de una plantilla **antes de descargarla** | Manifiesto: parámetros, motores, bibliografía, licencia, requisitos y compatibilidad |
| RQ-3 | **Descargar una plantilla individual** (última o versión concreta) | Destino dentro de `My-Creative-Ink-Repository`, a decidir por Creative-Ink |
| RQ-4 | **Personalizar antes de usar** con `clave=valor` | La librería valida (`params.Resolve`) y escribe `ink-params.tex`; los errores se muestran por campo |
| RQ-5 | **Descargar un paquete entero** | Resultado por plantilla; un fallo no cancela las demás |
| RQ-6 | Mostrar **qué puede usar este equipo** | Estados de §7, con lo que falta y el comando exacto (que ejecuta el usuario) |
| RQ-7 | Ver y aplicar actualizaciones de lo descargado | Nunca modifica proyectos ya creados |
| RQ-8 | Funcionar sin conexión | `list`/`info` con la caché y su fecha; `get` falla con `ErrOffline` como resultado claro |
| RQ-9 | Varias fuentes de catálogo | Clave `templates.sources` (lista de `{url, publicKey?}`). La oficial viene fija en el código y siempre firmada; las demás se marcan como "terceros" y, sin clave, como "sin firma" (decisiones del 2026-11-10 y 2026-11-11). **Solo en la capa del usuario:** si se pudiera escribir en la del proyecto, un proyecto clonado podría añadir un catálogo malicioso |
| RQ-10 | Motor del proyecto según el manifiesto | `compile.engine` en la capa del proyecto (`.creative-ink-meta`) con `defaultEngine` (**pdflatex**). Prerrequisito: la clave `compile.engine` y `-pdf/-xelatex/-lualatex` en el orquestador y en `ci compile` |
| RQ-11 | Recordar el origen de cada proyecto | `{id, version, source}` de la plantilla en `.creative-ink-meta`, para avisar de actualizaciones |
| RQ-12 | Mostrar en la pantalla de inicio las plantillas descargadas con su nombre y descripción | Si la carpeta trae `template.json`, `TemplateInfo` puede usar `name`, `description` y `version` en lugar del nombre de carpeta |

### 6.2 No funcionales

| # | Requerimiento |
|---|---------------|
| RNF-1 | macOS, Linux y Windows por igual: sin shell, sin `make`, sin Python (todo en Go, como el resto de `ci`) |
| RNF-2 | Seguridad (ver §8): HTTPS, **firma Ed25519 de cada JSON del catálogo** (contrato §7.1), `sha256` de los zips, extracción segura, límites, sin `.latexmkrc`, sin confianza heredada, **sin `--shell-escape`**, nada se ejecuta al descargar ni al crear el proyecto |
| RNF-3 | Nunca usar `api.github.com` |
| RNF-4 | No bloquear la UI: descargas cancelables (Ctrl+C / "Detener"), con progreso; reutilizar el streaming y el `OutputSpool` de la Terminal CI |
| RNF-5 | La librería se importa sin dependencias transitivas (solo biblioteca estándar) |
| RNF-6 | Tests sin red: `fixtures.FS()` y `catalogtest` (catálogo firmado servido con `httptest.Server`) de Memorandum-tex |
| RNF-7 | Compatibilidad hacia adelante: campos desconocidos se ignoran; `schemaVersion` desconocida = plantilla `incompatible` |

## 7. Estados de compatibilidad (`compat.Status`)

| Estado | Condición | Qué ve el usuario |
|--------|-----------|-------------------|
| `ready` | Todo disponible | ✅ lista |
| `missing-packages` | Algún `requires.files` no aparece en `kpsewhich` | ⚠ faltan paquetes + `tlmgr install <packages>` para que lo ejecute él |
| `missing-engine` | Ningún motor de `build.engines` está instalado | ⚠ falta el motor (amplía el health check, que hoy no detecta pdfLaTeX, LuaLaTeX ni BibTeX) |
| `blocked` | `build.shellEscape: true` (sustituye a `needs-trust` desde el 2026-11-11) | ⛔ bloqueada: pide ejecutar comandos del sistema (`--shell-escape`), que v1 no permite. No se puede descargar ni usar, ni siquiera confiando |
| `incompatible` | `schemaVersion` desconocida o `requires.creativeInk` mayor que la versión instalada | ⛔ actualiza Creative-Ink |

Además, si la versión del índice es mayor que la descargada, se marca `update-available`. Es independiente del
estado.

## 8. Seguridad: lo que Creative-Ink debe garantizar

1. **Es la primera función de red del editor.** `net/http` con tiempos de espera (conexión, cabeceras y total),
   `http.ProxyFromEnvironment`, solo HTTPS. Las redirecciones de Releases van a
   `objects.githubusercontent.com`: se aceptan porque la integridad la dan la firma y el `sha256`.
2. **Nada descargado se confía, y hay que corregir un hueco que ya existe hoy** (hallazgo del agente de Creative-Ink,
   2026-11-11, gravedad alta).
   - **El hueco:** `CreateProjectFromTemplate` (`registry_app.go:110-120`) confía en el `.latexmkrc` de **cualquier**
     plantilla, sin mirar su origen. En un repositorio compartido, una cuenta puede dejar código Perl que otra
     ejecutaría como de confianza.
   - **La corrección:** confiar automáticamente solo en las plantillas propias del usuario
     (`<UserConfigDir>/creative-ink/templates/`).
   - **Las descargadas:** nunca confían ni copian un `.latexmkrc`. El contrato ya lo prohíbe (B4, sin distinguir
     mayúsculas) y `bundle.SafeExtract` lo rechaza, pero Creative-Ink no debe depender solo de eso.
   - **Prioridad:** independiente del catálogo; conviene hacerlo primero (respuesta 7 de §12.2).
3. **Sin `--shell-escape`:** las plantillas con `shellEscape: true` se **bloquean** (`compat` devuelve `blocked`).
   La regla de confianza de Creative-Ink no sirve aquí, porque una carpeta sin `.latexmkrc` ya es de confianza.
4. **Nada se ejecuta al descargar ni al crear el proyecto.** Un proyecto creado desde el catálogo **no se compila
   automáticamente al crearse**, aunque `compile.auto` esté activado. La primera compilación la pide el usuario y
   pasa por el Scheduler y las reglas de confianza.
5. **Extracción con `bundle.SafeExtract`:** nunca con un unzip propio ni con la shell. La librería también rechaza
   las entradas que solo difieren en mayúsculas (B10), que en macOS y Windows se sobrescribirían.
6. **Firma (contrato §7.1):** `client` verifica la firma Ed25519 de cada JSON del catálogo oficial con la clave fija
   en la librería, y la de un catálogo de terceros si su fuente declara clave. Un fallo es `ErrSignature` y la
   plantilla no se usa.
7. **Los textos del catálogo son datos externos:** nombres, descripciones y parámetros nunca se renderizan como
   HTML (`{@html}` prohibido) ni se usan para construir comandos de shell. Los `clave=valor` que escribe el
   usuario solo llegan a `params.ParseAssignments` y `params.Resolve`.
8. **El nombre del proyecto se pasa sin escapar** a `params.Input.ProjectName`: la librería lo escapa una sola vez.
   Las plantillas del catálogo no usan `{{TITULO}}` (V39), así que la sustitución de `copyTemplate` no debe
   aplicarse a ellas.
9. **La red solo se usa bajo demanda** (respuesta 2 de §12.2): nada de consultas en segundo plano ni avisos
   automáticos de actualizaciones. Solo cuando el usuario ejecuta `ci templates …`.
10. **`templates.sources` solo en la capa del usuario** (RQ-9). Los catálogos de terceros se muestran marcados y sus
    fallos no afectan al oficial.

## 9. Comandos `ci` propuestos (sugerencia; los decide Creative-Ink)

Siguiendo el registro (`internal/ci/registry.go`), un solo comando `templates`:

| Forma (en la Terminal CI se escribe sin `ci`) | Qué hace | Requerimientos |
|-----------------------------------------------|----------|----------------|
| `templates list [--categoria <c>]` | Catálogo con estado (§7) y versión descargada | RQ-1, RQ-6 |
| `templates search <texto>` | Busca en nombre, descripción y etiquetas | RQ-1 |
| `templates info <id>[@versión]` | Manifiesto legible: parámetros (clave, tipo, valores, por defecto), motores, licencia, requisitos | RQ-2, RQ-6 |
| `templates get <id>[@versión] [clave=valor …]` | Descarga individual, con parámetros opcionales | RQ-3, RQ-4 |
| `templates pack <pack-id>` | Descarga un paquete entero | RQ-5 |
| `templates packs` | Lista los paquetes | RQ-5 |
| `templates update [<id>\|--todas]` | Actualiza lo descargado | RQ-7 |
| `templates remove <id>` | Quita una plantilla descargada (nunca proyectos) | — |
| `templates refresh` | Vuelve a descargar el índice | RQ-8 |
| `templates new <id> <nombre> [clave=valor …]` | Crea un proyecto desde una plantilla descargada | RQ-4, RQ-10, RQ-11 |
| `templates sources [add <url> [--clave <pública>]\|remove <url>]` | Gestiona `templates.sources` (solo en la capa del usuario); sin clave, la fuente queda como "sin firma" | RQ-9 |

## 10. Pruebas y fixtures (los entrega Memorandum-tex en M1)

**Entregado en M1 (2026-11-10)**, descrito en `tools/fixtures/data/README.md`:

```
tools/fixtures/data/
├── templates/prueba/{ok-minima, futura, con-paquetes}/  ← fuentes: lista, incompatible, missing-packages
├── packs/prueba.json
├── params/{escape-cases.json, resolve-cases.json, ok-minima.*.ink-params.tex}
├── invalid/{manifests.json, folders.json}               ← un caso por regla V01–V39
└── bundles/cases.json                                   ← 16 zips: válido, checksum, zip slip, C:, \, enlace,
                                                            .latexmkrc (también en mayúsculas), colisión por
                                                            mayúsculas, script, dos raíces, sin meta, sin LICENSE,
                                                            raíz distinta del id, bomba
```

- **Los zips no se guardan en git:** se construyen en los tests con fecha fija, así que el `sha256` es reproducible
  en los tres sistemas.
- **Catálogo servible (⏳ M2, decidido el 2026-11-11):** no habrá una carpeta generada en git.
  - **Qué es:** el paquete `tools/pkg/catalogtest` construye el catálogo completo en una carpeta temporal (índice,
    manifiestos publicados, `versions.json`, paquetes, los 16 zips y sus `.sig`) y devuelve un `httptest.Server`.
  - **Firma de prueba:** firma el catálogo con una **clave efímera** generada en cada test y expone su clave pública,
    para que Creative-Ink pruebe también los casos de firma ausente o alterada.
  - **Para qué sirve:** con él, Creative-Ink prueba `ci templates …` de punta a punta sin red, llamándolo desde sus
    tests. La forma prevista es `srv := catalogtest.NewServer(t)`, `srv.IndexURL()` y `srv.PublicKey()`; la API
    definitiva se fija en el paso 6 de M2.
- **Ya utilizable:** `params/escape-cases.json` sirve desde hoy para un test de `core.EscapeTeX` en Creative-Ink. Sus
  casos coinciden con su implementación actual (`repository.go`).

## 11. Plan coordinado

`M` = Memorandum-tex, `C` = Creative-Ink. Una flecha indica que hay que esperar a esa fase.

| Fase | Lado | Qué | Desbloquea |
|------|------|-----|------------|
| M0 | M | Gobierno: `CLAUDE.md`, `docs/MAP.md`, `CONTRIBUTING.md`, skills, `.gitattributes`, ramas | — |
| M1 | M | Contrato v1 (`docs/contrato-plantillas.md`, esquemas JSON) + fixtures de §10 | C1 |
| M2 | M | Librería Go (§5): ✅ `contract`, `params`, `compat`, validación de carpeta y `bundle` (pasos 1–5); ⏳ `catalogtest` firmado, `client` con verificación de firmas e `inkt` (pasos 6–8). Tests, `-race` y `go vet` en 3 sistemas | C2 |
| C0 | C | Prerrequisitos:<br>• **corregir la confianza por origen de la plantilla** (§8.2, hueco actual: va primero);<br>• tipos de clave nuevos en el servicio de configuración (lista para `templates.sources`, enum para `compile.engine`), con capa permitida por clave;<br>• `compile.engine`;<br>• detección de pdfLaTeX, LuaLaTeX, BibTeX y `kpsewhich` | C3 |
| C1 | C | Análisis y diseño de la sección (rutas en `My-Creative-Ink-Repository`, comandos, IPC, decisiones) | C2 |
| C2 | C | Catálogo de solo lectura: `list`, `search`, `info`, `packs`, `refresh` contra los fixtures | C3 |
| M3 | M | Componentes compartidos (motor, paletas, institución UNAM-FES con logos) | M4 |
| M4 | M | Plantilla piloto `ensayo` de punta a punta (manifiesto, bundle, README, vista previa) | M5 |
| C3 | C | `get` individual con parámetros, `new` y origen en `.creative-ink-meta` | C4 |
| M5 | M | Resto de plantillas: artículo, presentaciones, infografía, cartas, ISPRS, registro de avances | M7 |
| C4 | C | `pack`, `update`, `remove`, `sources` | C5 |
| M6 | M | Tesis + kit de seguimiento (issues, CHANGELOG, CONTEXTO, revisiones) | M7 |
| M7 | M | Publicación: GitHub Pages (catálogo **firmado**) + Releases (bundles) desde CI. Antes, el humano genera el par de claves Ed25519, guarda la privada en un secreto de GitHub Actions y la pública entra en la librería (`client.Official`) | C5 |
| C5 | C | Integración con el catálogo real | — |
| — | M | Licencias **hechas el 2026-11-11**: MIT-0 para plantillas, MIT para la librería (compatible con cualquier licencia de Creative-Ink) | — |
| M8 | M | Identidad (texto de muestra con el nombre y la descripción del repositorio) | — |
| C6 | C | (Opcional) galería en la UI | — |

Casillas de seguimiento (se marcan con fecha al cerrar):

- [x] M0 (2026-11-10) · [x] M1 (2026-11-11) · [ ] M2 (en curso: pasos 1–5 de 8) · [ ] M3 · [ ] M4 · [ ] M5 · [ ] M6 · [ ] M7 · [ ] M8
- [x] Licencias (2026-11-11) · [x] Revisión de seguridad incorporada (2026-11-11)
- [ ] C0 · [ ] C1 · [ ] C2 · [ ] C3 · [ ] C4 · [ ] C5 · [ ] C6

## 12. Decisiones

### 12.1 Tomadas (programador, 2026-11-10)

1. Descarga **bajo demanda**: individual o por paquete. Nada se instala de golpe.
2. Dónde se guarda lo descargado dentro de `My-Creative-Ink-Repository` lo decide **Creative-Ink**.
3. Catálogo en **GitHub Pages** y bundles en **GitHub Releases** del repositorio
   `galigaribaldi/Memorandum-tex`.
4. Herramientas en **Go**. Python, solo para extensiones concretas más adelante.
5. Manifiesto en **JSON** con JSON Schema.
6. Variantes como **parámetros booleanos**; **semver por plantilla**; parámetros en **`ink-params.tex`**.
7. Motor por defecto **pdfLaTeX** (XeLaTeX y LuaLaTeX opcionales). Bibliografía por defecto **natbib/apalike +
   BibTeX** (biblatex + Biber opcional).
8. El catálogo va en el **núcleo** de Creative-Ink (no como plugin) y admite **catálogos de terceros** marcados.
9. Se incluyen los logos UNAM/FES. ~~Software libre; la licencia (probablemente MIT) se decide al final (M8).~~
   **Resuelto el 2026-11-11:** MIT para la librería y MIT-0 para las plantillas (`NOTICE.md`).
10. Plantillas solo en español por ahora.

### 12.1 bis. Tomadas (programador, 2026-11-11)

11. **Licencias:** MIT para la librería, esquemas y documentación; **MIT-0** para las plantillas (los documentos de
    los usuarios no arrastran avisos). Los logos y el material de terceros quedan excluidos (`NOTICE.md`).
12. **El catálogo de prueba es un paquete Go** (`pkg/catalogtest`); no hay zips ni catálogos generados en git.
13. **`--shell-escape` prohibido en v1:** los clientes bloquean `shellEscape: true` (`blocked`).
14. **Catálogo oficial firmado con Ed25519:**
    - **Qué se firma:** cada JSON (contrato §7.1).
    - **Terceros:** con clave declarada, firma obligatoria; sin ella, "sin firma".
15. **Endurecimiento de rutas:** colisiones por mayúsculas y `.latexmkrc` en cualquier forma, rechazados.

### 12.2 Respuestas del programador a las preguntas del agente de Creative-Ink (2026-11-11)

1. **Dónde se guardan las descargadas:** en un directorio **del usuario**, por ejemplo
   `<UserConfigDir>/creative-ink/catalog/<fuente>/<id>/`. Nunca en `.creative-ink/templates/` del repositorio: en uno
   compartido, una cuenta inyectaría archivos a las demás. La ruta final la decide Creative-Ink.
2. **La red solo bajo demanda:** prohibidas las consultas autónomas. La red se usa solo cuando el usuario ejecuta
   `ci templates …` ("nada se instala de golpe").
3. **`shellEscape: true` bloqueado en v1:** sí. El contrato lo prohíbe y `compat` devuelve `blocked`.
4. **Catálogos de terceros desde el principio:** sí, con `templates.sources` en la capa del usuario.
   - **Cómo se muestran:** marcados como "de terceros" y, sin clave pública, como "sin firma".
   - **El oficial:** va fijo en el código y siempre firmado.
5. **`compile.engine` antes que el catálogo:** es un prerrequisito (C0).
   - **Mientras tanto:** las plantillas compilan con el XeLaTeX fijo gracias a `ink-motor.tex`.
6. **Versiones:** `versions.json` guarda el historial publicado.
   - **Al actualizar:** `templates update` descarga la versión nueva al catálogo local.
   - **Proyectos ya creados:** **nunca** se modifican.
   - **Si conviven varias versiones localmente:** lo decide Creative-Ink. La librería permite pedir una versión
     concreta.
7. **La confianza automática se limita a las plantillas propias del usuario:** sí, y **es lo primero** (§8.2).
8. **Galería en la interfaz:** queda para la fase opcional C6; la primera versión es solo la Terminal CI.

### 12.3 Pendientes del lado de Creative-Ink

1. **Filas IPC y eventos:** un evento propio de progreso de descarga (el `terminal:progress` de compilación no
   sirve), `templates:changed`, y si `CreateProjectFromTemplate` pasa a aceptar parámetros (cambio de contrato).
2. **Biber:** `ci compile completo` usa siempre BibTeX. Las plantillas con `build.bibliography: "biber"` (o el
   parámetro `bibliografia=biber`) fallarían hasta que lea ese campo. Las de este repositorio usan BibTeX por defecto.
3. **Servicio de configuración:** tipos lista y enum, y la capa permitida por clave.

### 12.4 Rediseño de UI/UX de Creative-Ink (`DEV-feat-UX-Editor`), decidido por el programador el 2026-11-11

No afecta al catálogo. Se registra aquí porque las respuestas llegaron en el mismo mensaje:

- **Bloque A, layout central:** se consolida el CSS Grid principal (`ActionBar | Sidebar | Workspace`) y el
  explorador se queda en su cajón de 240 px.
- **Bloque B, división del editor:** divisor arrastrable de 4 px (`col-resize`) entre los paneles principal y
  secundario. Un archivo solo puede estar en un panel a la vez, para que los autoguardados no choquen.
- **Bloque C, vista previa aislada:** PDF.js con `OffscreenCanvas` y Web Workers locales, para que un PDF grande
  nunca congele CodeMirror.
- **Bloque D, terminal híbrida:** pestañas separadas. "Terminal CI" para el registro de comandos seguros y "Shell"
  para la ejecución directa en el sistema operativo.

## 13. Qué no hacer

- No incluir plantillas de Memorandum-tex dentro del binario de Creative-Ink.
- No reimplementar en Creative-Ink la validación del manifiesto, la resolución de parámetros ni la extracción: se
  importan de la librería (una sola implementación del contrato).
- No usar `git clone` ni la API de GitHub para descargar.
- No ejecutar `tlmgr`, compiladores ni nada que sugiera `compat.Hints`: solo se muestra el comando.
- No confiar automáticamente en nada descargado, ni en plantillas que no sean propias del usuario.
- No compilar con `--shell-escape` nada que venga del catálogo.
- No compilar automáticamente un proyecto recién creado desde el catálogo.
- No consultar la red en segundo plano.
- No subir un `replace` en el `go.mod`: se usa un `go.work` local.

## 14. Bitácora del documento

- **2026-11-10:** borrador v0, tras analizar Creative-Ink (`internal/core/templates.go`, `internal/ci/registry.go`,
  `docs/MAP.md` §1, §4.3.6 y §4.6) y las decisiones del programador.
- **2026-11-10 (M1):**
  - **Contrato:** contrato v1 en `docs/contrato-plantillas.md`, que reemplaza a §4.
  - **Esquemas, paquetes y fixtures:** cuatro JSON Schema en `schema/`, cuatro paquetes en `packs/` y los fixtures
    de §10.
  - **Comprobado en Creative-Ink:** `.creative-ink-meta` (`ProjectMeta {version, mainFile, config}`) y
    `core.EscapeTeX` (mismas 10 sustituciones, una sola pasada).
  - **Errores:** se añaden `ErrForbiddenFile` y `ErrBundleLayout`.
- **2026-11-11:**
  - **Catálogo de prueba:** pasa a ser el paquete `pkg/catalogtest` (sin zips en git).
  - **Licencia:** decidida: **MIT** para la librería que importará Creative-Ink y **MIT-0** para las plantillas
    (`NOTICE.md`). Al importar la librería, Creative-Ink debe conservar el aviso MIT. Las plantillas descargadas no
    imponen nada a los documentos del usuario.
  - **Fixture `ok-minima`:** compilado sin errores ni desbordes con pdfLaTeX, XeLaTeX y LuaLaTeX, con `ink-params.tex`
    y en los modos borrador y capítulo.
- **2026-11-11 (M2, pasos 1 y 2), importante para Creative-Ink:**
  - **Fixtures movidos:** de `tests/fixtures/` a **`tools/fixtures/data/`**. Así viajan con el módulo `tools` cuando
    Creative-Ink lo importe, porque `go:embed` no admite `../`. El paquete
    `github.com/galigaribaldi/Memorandum-tex/tools/fixtures` los expone con `fixtures.FS()` (un `fs.FS` raíz en
    `data/`). Por ejemplo, `escape-cases.json` está en `fixtures.FS()` bajo `params/escape-cases.json`.
  - **`pkg/contract` implementado:**
    - **Tipos y errores:** los tipos del contrato y los errores compartidos `contract.Err*` (todos los paquetes
      devuelven estos).
    - **Validación:** `ValidateManifest(data, Options{Origin, RepoPolicy, Context})` devuelve un `Report` con
      hallazgos `{Code, Level, Field, Message, Line}`, y `ValidatePack` hace lo mismo con los paquetes.
    - **Para Creative-Ink:** valida los manifiestos publicados con `Origin: contract.OriginPublished` y **sin**
      `RepoPolicy`, que es solo para este repositorio.
  - **`pkg/params` implementado:** `ParseAssignments`, `Resolve`, `RenderInkParams` y `EscapeTeX` (firmas en §5).
  - **Verificación:** 47 casos de manifiesto, 10 de resolución, 2 salidas exactas de `ink-params.tex` y 9 de
    escapado; `go test -race` y `go vet` para darwin, linux y windows en verde.
- **2026-11-11 (revisión de seguridad del agente de Creative-Ink, incorporada antes de seguir con M2):**
  - **`shellEscape`:** pasa de "requiere confianza" a **bloqueado** (estado `blocked`, V10 y la regla nueva V39).
  - **Firma:** firma Ed25519 de cada JSON del catálogo (contrato §7.1, `ErrSignature`).
  - **Rutas:** rechazo de colisiones por mayúsculas (B10) y de rc de latexmk sin distinguir mayúsculas.
  - **Marcadores:** prohibidos `{{TITULO}}`/`{{TITLE}}`.
  - **Desarrollo local:** `go.work` en lugar de `replace`.
  - **Requerimientos para Creative-Ink:** confianza por origen de la plantilla, no compilar al crear, red solo bajo
    demanda y `templates.sources` solo en la capa del usuario.
  - **Respuestas:** a las 8 preguntas del agente (§12.2) y decisiones de UI/UX (§12.4).
  - **`pkg/compat` implementado:** con el estado `blocked` (§5.2).
- **2026-11-12 (revisión integral antes de pasar el documento a Creative-Ink):**
  - **Cabecera:** actualizada; el contrato es la fuente de verdad.
  - **§5:** nueva §5.1 con la organización del repositorio y del módulo `tools`, y §5.2 con el estado de cada paquete
    (implementado o pendiente). La API de `client` queda marcada como propuesta.
  - **Firma en todo el documento:** `.sig` en el árbol del catálogo, peticiones por consulta y `catalogtest` con clave
    efímera.
  - **Datos al día:** 16 zips de prueba, casillas de M0 y M1 marcadas, decisiones del 11 de noviembre en §12.1 bis, y
    `go.work` en el reparto de responsabilidades.
- **2026-11-15 (M2, pasos 4 y 5):**
  - **`pkg/bundle` implementado:** firma real en §5.2. Lo más relevante para Creative-Ink:
    - la extracción valida **todo** el zip antes de escribir;
    - nunca escribe más de lo que declara la cabecera;
    - trabaja en una carpeta temporal y la renombra al final;
    - rechaza un destino con archivos;
    - rechaza también `shellEscape` (B11).
  - **Componentes:** dentro del bundle van en `Latex/ink/<componente>/`.
  - **Relación de compresión:** 100:1 solo para archivos de 1 MB o más.
  - **`sha256`:** reproducible con la misma versión de Go; por eso el cliente **verifica** el hash del catálogo firmado
    y nunca lo recalcula a partir de las fuentes.
  - **Validación de carpeta:** implementada en `internal/validate`, solo para este repositorio.
