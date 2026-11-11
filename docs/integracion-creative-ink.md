# Integración Memorandum-tex ↔ Creative-Ink: plan y requerimientos

> **Estado:** borrador v0 (2026-11-10). Lo escribió el lado de Memorandum-tex para que el agente de IA de
> Creative-Ink lo analice **antes** de diseñar la descarga de plantillas.
> **Precedencia:** para el formato del catálogo, los manifiestos y los bundles, este documento manda hasta que exista
> `docs/contrato-plantillas.md` (fase M1), que lo reemplaza. Para todo lo que pasa dentro de Creative-Ink (rutas,
> IPC, UI, comandos `ci`, seguridad del editor) mandan su `CLAUDE.md` y su `docs/MAP.md`. Aquí solo hay
> **requerimientos y propuestas** para ese lado.
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
| Publicar: GitHub Pages (catálogo) y Releases (bundles) | Política de confianza (nunca confiar en algo descargado) y claves de configuración |
| Fixtures de prueba sin red | IPC, eventos y, si se decide, la galería en la UI |
| Validador del contrato en CI | `go get` / `replace` del módulo (lo ejecuta el programador) |

## 3. Distribución y escalabilidad

### 3.1 Dónde vive cada cosa

```
Catálogo (GitHub Pages, estático y con CDN)
https://galigaribaldi.github.io/Memorandum-tex/
├── index.html                                   ← galería para humanos (más adelante)
└── catalog/v1/                                  ← "v1" = schemaVersion del contrato
    ├── index.json                               ← resumen ligero: plantillas + paquetes
    ├── templates/<id>/versions.json             ← versiones publicadas de esa plantilla
    ├── templates/<id>/<version>/template.json   ← manifiesto publicado (incluye "dist")
    ├── templates/<id>/<version>/preview.png     ← miniatura (≤ 200 KB)
    └── packs/<pack-id>.json                     ← definición de un paquete

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
| Peticiones por consulta | — | `list` = 1 petición (con ETag: 304 si no cambió); `info` = 1; descarga = 1 + 1 zip | Mínimo |

## 4. Contrato v1 (borrador)

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
- Un booleano se compara con `\inkIfTrue{\ConPortada}{…}{…}` (definido en `Latex/ink/ink-base.tex`).

### 4.7 Convenciones LaTeX que Creative-Ink puede dar por hechas

- **Modos:** `\borrador` (graphicx en modo draft) y `\SoloCapitulo{<ruta>}` (`\includeonly`) con **los mismos
  nombres** que ya usa `ci compile borrador|solo`. Las plantillas los respetan si se definen por la línea de
  órdenes o en `ink-params.tex`.
- **Motor:** `Latex/ink/ink-motor.tex` (con iftex) carga `inputenc`/`fontenc` en pdfLaTeX y `fontspec` en
  XeLaTeX o LuaLaTeX. Por eso las plantillas **también compilan con el `latexmk -xelatex` fijo** que Creative-Ink
  tiene hoy, mientras llega `compile.engine`.
- **Sin `--shell-escape`** salvo que `build.shellEscape` sea `true`, que lo marca como "requiere confianza". En v1
  ninguna plantilla lo pide; la externalización de TikZ de la tesis queda apagada por defecto.
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
- **Desarrollo local:** `replace github.com/galigaribaldi/Memorandum-tex/tools => ../../Memorandum-tex/tools` en
  el `go.mod` de Creative-Ink. Los dos repositorios están en `Personal_Documents/`.

| Paquete | Responsabilidad | Ejecuta procesos | Usa la red |
|---------|-----------------|------------------|------------|
| `pkg/contract` | Tipos (`Manifest`, `Index`, `Pack`, `Parameter`, `Dist`), `Validate`, `SchemaVersion` | No | No |
| `pkg/params` | `Resolve(manifest, values) → Values` (valida tipos, obligatorios, `enum`; errores por campo), `RenderInkParams`, `EscapeTeX` | No | No |
| `pkg/compat` | `Evaluate(manifest, Env) → Report{Status, Missing, Hints}` | No: Creative-Ink le pasa `Env` | No |
| `pkg/bundle` | `VerifySHA256`, `SafeExtract(zip, dst, Limits)` con la lista permitida de §4.5 | No | No |
| `pkg/client` | `Index`, `Manifest`, `Versions`, `Pack`, `Get` (individual), `GetPack`; caché con ETag; resultado offline | No | Sí (HTTPS) |

**API propuesta** (puede ajustarse en M2; la versión final se documenta en `docs/contrato-plantillas.md`):

```go
c := client.New(client.Options{
    Sources:   []client.Source{{ID: "memorandum-tex", IndexURL: "https://galigaribaldi.github.io/Memorandum-tex/catalog/v1/index.json"}},
    HTTP:      httpClient,      // Creative-Ink: timeouts y proxy (http.ProxyFromEnvironment)
    CacheDir:  cacheDir,        // Creative-Ink decide (p. ej. <UserConfigDir>/creative-ink/cache/catalog)
    UserAgent: "creative-ink/<versión>",
})

idx, err := c.Index(ctx, client.IndexOptions{Refresh: false})
// idx.Offline == true si se usó la caché por falta de red; idx.FetchedAt dice de cuándo es.

m, err := c.Manifest(ctx, "ensayo", "")                 // "" = latest; solo JSON, sin bundle
vals, err := params.Resolve(m, map[string]string{"portada": "false", "color": "Teal"})
// err es *params.Error{Fields: []FieldError{Key, Reason}} si algo no es válido

rep := compat.Evaluate(m, compat.Env{
    Engines:     map[string]bool{"pdflatex": true, "xelatex": true},
    HasFiles:    func(names []string) map[string]bool { /* Creative-Ink: una llamada a kpsewhich */ },
    InkVersion:  "0.3.0",
})

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
  - Verifica el `sha256` **antes** de extraer y respeta límites: índice ≤ 1 MB, manifiesto ≤ 256 KB, bundle ≤ 200 MB.
- **Seguridad:** no ejecuta procesos ni sigue rutas peligrosas. Solo acepta HTTPS (HTTP solo con un flag explícito
  para tests con `httptest`).
- **Errores tipados:** `ErrOffline`, `ErrNotFound`, `ErrChecksum`, `ErrUnsafePath`, `ErrTooLarge`,
  `ErrUnsupportedSchema`, `ErrDestExists`, `*params.Error`.
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
| RQ-9 | Varias fuentes de catálogo | Clave `templates.sources` (lista de URLs de índice). La oficial viene por defecto; las demás se marcan como "terceros" (decisión del 2026-11-10) |
| RQ-10 | Motor del proyecto según el manifiesto | `compile.engine` en la capa del proyecto (`.creative-ink-meta`) con `defaultEngine` (**pdflatex**). Prerrequisito: la clave `compile.engine` y `-pdf/-xelatex/-lualatex` en el orquestador y en `ci compile` |
| RQ-11 | Recordar el origen de cada proyecto | `{id, version, source}` de la plantilla en `.creative-ink-meta`, para avisar de actualizaciones |
| RQ-12 | Mostrar en la pantalla de inicio las plantillas descargadas con su nombre y descripción | Si la carpeta trae `template.json`, `TemplateInfo` puede usar `name`, `description` y `version` en lugar del nombre de carpeta |

### 6.2 No funcionales

| # | Requerimiento |
|---|---------------|
| RNF-1 | macOS, Linux y Windows por igual: sin shell, sin `make`, sin Python (todo en Go, como el resto de `ci`) |
| RNF-2 | Seguridad (ver §8): HTTPS, `sha256`, extracción segura, límites, sin `.latexmkrc`, sin confianza heredada, nada se ejecuta al descargar |
| RNF-3 | Nunca usar `api.github.com` |
| RNF-4 | No bloquear la UI: descargas cancelables (Ctrl+C / "Detener"), con progreso; reutilizar el streaming y el `OutputSpool` de la Terminal CI |
| RNF-5 | La librería se importa sin dependencias transitivas (solo biblioteca estándar) |
| RNF-6 | Tests sin red: fixtures de Memorandum-tex + `httptest.Server` |
| RNF-7 | Compatibilidad hacia adelante: campos desconocidos se ignoran; `schemaVersion` desconocida = plantilla `incompatible` |

## 7. Estados de compatibilidad (`compat.Status`)

| Estado | Condición | Qué ve el usuario |
|--------|-----------|-------------------|
| `ready` | Todo disponible | ✅ lista |
| `missing-packages` | Algún `requires.files` no aparece en `kpsewhich` | ⚠ faltan paquetes + `tlmgr install <packages>` para que lo ejecute él |
| `missing-engine` | Ningún motor de `build.engines` está instalado | ⚠ falta el motor (amplía el health check, que hoy no detecta pdfLaTeX, LuaLaTeX ni BibTeX) |
| `needs-trust` | `build.shellEscape: true` | 🔒 requiere confianza |
| `incompatible` | `schemaVersion` desconocida o `requires.creativeInk` mayor que la versión instalada | ⛔ actualiza Creative-Ink |

Además, si la versión del índice es mayor que la descargada, se marca `update-available`. Es independiente del
estado.

## 8. Seguridad: lo que Creative-Ink debe garantizar

1. **Es la primera función de red del editor.** `net/http` con tiempos de espera (conexión, cabeceras y total),
   `http.ProxyFromEnvironment`, solo HTTPS. Las redirecciones de Releases van a
   `objects.githubusercontent.com`: se aceptan porque la integridad la da el `sha256`.
2. **Nada descargado se confía.** La decisión 7 de Creative-Ink (".latexmkrc de una plantilla propia = de
   confianza") **no aplica** a plantillas descargadas. Aunque el contrato prohíbe `.latexmkrc`, si apareciera uno
   el proyecto debe nacer `untrusted`.
3. **Nada se ejecuta al descargar ni al crear el proyecto.** La primera compilación pasa por el Scheduler y las
   reglas de confianza que ya existen.
4. **Extracción con `bundle.SafeExtract`:** nunca con un unzip propio ni con la shell.
5. **Los textos del catálogo son datos externos:** nombres, descripciones y parámetros nunca se renderizan como
   HTML (`{@html}` prohibido) ni se usan para construir comandos de shell. Los `clave=valor` que escribe el
   usuario solo llegan a `params.Resolve`.
6. **Los catálogos de terceros** se muestran marcados y sus fallos no afectan al oficial.

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
| `templates sources [add\|remove <url>]` | Gestiona `templates.sources` | RQ-9 |

## 10. Pruebas y fixtures (los entrega Memorandum-tex en M1)

```
tests/fixtures/catalog/v1/
├── index.json                         ← 3 plantillas + 1 paquete
├── templates/ok-minima/1.0.0/…        ← compatible, con parámetros de todos los tipos
├── templates/futura/1.0.0/…           ← schemaVersion 99 → incompatible
├── templates/con-paquetes/1.0.0/…     ← requires.files inexistente → missing-packages
├── packs/prueba.json
└── bundles/
    ├── ok-minima-1.0.0.zip            ← válido
    ├── zip-slip.zip                   ← entrada "../../fuera.tex" → ErrUnsafePath
    ├── con-latexmkrc.zip              ← → rechazado
    └── checksum-malo.zip              ← sha256 distinto del índice → ErrChecksum
```

Con eso, Creative-Ink puede probar `ci templates …` de punta a punta con un `httptest.Server` que sirva esa
carpeta, sin red y antes de que exista el catálogo real.

## 11. Plan coordinado

`M` = Memorandum-tex, `C` = Creative-Ink. Una flecha indica que hay que esperar a esa fase.

| Fase | Lado | Qué | Desbloquea |
|------|------|-----|------------|
| M0 | M | Gobierno: `CLAUDE.md`, `docs/MAP.md`, `CONTRIBUTING.md`, skills, `.gitattributes`, ramas | — |
| M1 | M | Contrato v1 (`docs/contrato-plantillas.md`, esquemas JSON) + fixtures de §10 | C1 |
| M2 | M | Librería Go (§5) con tests, `-race` y `go vet` en 3 sistemas | C2 |
| C0 | C | Prerrequisito: clave `compile.engine`, detección de pdfLaTeX, LuaLaTeX, BibTeX y `kpsewhich` | C3 |
| C1 | C | Análisis y diseño de la sección (rutas en `My-Creative-Ink-Repository`, comandos, IPC, decisiones) | C2 |
| C2 | C | Catálogo de solo lectura: `list`, `search`, `info`, `packs`, `refresh` contra los fixtures | C3 |
| M3 | M | Componentes compartidos (motor, paletas, institución UNAM-FES con logos) | M4 |
| M4 | M | Plantilla piloto `ensayo` de punta a punta (manifiesto, bundle, README, vista previa) | M5 |
| C3 | C | `get` individual con parámetros, `new` y origen en `.creative-ink-meta` | C4 |
| M5 | M | Resto de plantillas: artículo, presentaciones, infografía, cartas, ISPRS, registro de avances | M7 |
| C4 | C | `pack`, `update`, `remove`, `sources` | C5 |
| M6 | M | Tesis + kit de seguimiento (issues, CHANGELOG, CONTEXTO, revisiones) | M7 |
| M7 | M | Publicación: GitHub Pages (catálogo) + Releases (bundles) desde CI | C5 |
| C5 | C | Integración con el catálogo real | — |
| M8 | M | Identidad y licencia (texto de muestra con el nombre y la descripción del repositorio, `LICENSE` en cada bundle) | — |
| C6 | C | (Opcional) galería en la UI | — |

Casillas de seguimiento (se marcan con fecha al cerrar):

- [ ] M0 · [ ] M1 · [ ] M2 · [ ] M3 · [ ] M4 · [ ] M5 · [ ] M6 · [ ] M7 · [ ] M8
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
9. Se incluyen los logos UNAM/FES. Software libre; la licencia (probablemente MIT) se decide al final (M8).
10. Plantillas solo en español por ahora.

### 12.2 Pendientes del lado de Creative-Ink (para su fase C1)

1. Estructura dentro de `My-Creative-Ink-Repository`: ¿lo descargado va a `.creative-ink/templates/<id>/` (ya lo
   lee `ListTemplates`) o a otra carpeta? ¿Se permite descargar directamente como proyecto?
2. Dónde va la caché del catálogo.
3. Filas IPC y eventos (por ejemplo `templates:progress`, `templates:changed`), y si
   `CreateProjectFromTemplate` pasa a aceptar parámetros (cambio de contrato).
4. Cómo se muestra el progreso de una descarga en la Terminal CI (`terminal:progress` u otro evento).
5. Si la galería en la UI entra en el MVP.

## 13. Qué no hacer

- No incluir plantillas de Memorandum-tex dentro del binario de Creative-Ink.
- No reimplementar en Creative-Ink la validación del manifiesto, la resolución de parámetros ni la extracción: se
  importan de la librería (una sola implementación del contrato).
- No usar `git clone` ni la API de GitHub para descargar.
- No ejecutar `tlmgr`, compiladores ni nada que sugiera `compat.Hints`: solo se muestra el comando.
- No confiar automáticamente en nada descargado.

## 14. Bitácora del documento

- **2026-11-10:** borrador v0, tras analizar Creative-Ink (`internal/core/templates.go`, `internal/ci/registry.go`,
  `docs/MAP.md` §1, §4.3.6 y §4.6) y las decisiones del programador.
