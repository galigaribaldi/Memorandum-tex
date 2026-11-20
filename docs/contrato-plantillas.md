# Contrato de plantillas v1

> **Estado:** v1, borrador cerrado en M1 (2026-11-10). Reemplaza a `docs/integracion-creative-ink.md` §4.
> **Alcance:** formato de los manifiestos (`template.json`), del catálogo (`index.json`, `versions.json`, paquetes),
> de los bundles (`<id>-<versión>.zip`) y de `ink-params.tex`, además de las reglas que una plantilla LaTeX debe
> cumplir para que todo eso funcione.
> **Implementación única:** `tools/pkg/*` (M2). Creative-Ink la importa; nadie reimplementa estas reglas.
> **Esquemas:** `schema/template.schema.json`, `schema/index.schema.json`, `schema/versions.schema.json`,
> `schema/pack.schema.json`. El esquema ayuda al editor; **la validación que manda es la de `inkt validate`**, que
> además aplica las reglas de §11 que un JSON Schema no puede expresar.

---

## 1. Versionado del contrato

- `schemaVersion` es un entero. Esta especificación es la **1**.
- **Cambio compatible** (no sube `schemaVersion`): un campo opcional nuevo, un valor nuevo en una lista abierta
  (categoría, etiqueta) o un modo nuevo que los clientes viejos ignoran.
- **Cambio incompatible** (sube `schemaVersion`): quitar o renombrar un campo, cambiar su tipo o su semántica, añadir
  un campo obligatorio o un tipo de parámetro nuevo. Se publica en `catalog/v<n>/` **sin retirar** `catalog/v<n-1>/`.
- **Los clientes ignoran los campos desconocidos.** Una plantilla con `schemaVersion` desconocida es `incompatible`
  (un resultado), nunca un error que tumbe el catálogo entero.

## 2. Convenciones generales

| Concepto | Regla |
|----------|-------|
| `id` de plantilla o paquete | `^[a-z0-9]+(-[a-z0-9]+)*$`, de 2 a 40 caracteres. Único en su catálogo. Es el nombre de su carpeta |
| Versión | Semver sin prerelease ni metadatos: `^(0\|[1-9][0-9]*)\.(0\|[1-9][0-9]*)\.(0\|[1-9][0-9]*)$` |
| Tag de publicación | `<id>-v<versión>` (p. ej. `ensayo-v1.2.0`) |
| Rutas dentro de manifiestos y zips | Relativas, con `/`, sin `.` ni `..` como segmento, sin `/` inicial |
| Rutas relativas en el catálogo | Se resuelven **contra la URL del JSON que las contiene** (RFC 3986, `url.ResolveReference`) |
| URLs absolutas | Solo `https://` |
| Texto | UTF-8 sin BOM, fin de línea LF |
| Fechas | `AAAA-MM-DD` en parámetros; RFC 3339 en UTC (`2026-11-10T18:00:00Z`) en el catálogo |
| Idioma de las claves | Inglés (`defaultEngine`); valores visibles para el usuario en español |

## 3. Manifiesto `template.json`

Vive en `templates/<categoria>/<id>/template.json`.

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
  "license": "MIT-0",
  "authors": [{ "name": "Hernán Galileo Cabrera Garibaldi", "github": "galigaribaldi" }],
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
    "files": ["spanish.ldf", "natbib.sty", "tcolorbox.sty"]
  },
  "components": ["motor", "base", "paletas", "comandos/academico", "instituciones/unam-fes-acatlan"],
  "parameters": [
    { "key": "titulo",  "macro": "TituloDoc",  "type": "string",  "label": "Título", "required": true, "fromProjectName": true },
    { "key": "autor",   "macro": "Autor",      "type": "string",  "label": "Autor", "required": true },
    { "key": "portada", "macro": "ConPortada", "type": "boolean", "label": "Con portada", "default": true },
    { "key": "color",   "macro": "ColorTema",  "type": "enum",    "label": "Paleta de color", "default": "Institucional",
      "values": ["Institucional", "Teal", "VerdeOlivo", "Purpura", "Rojo", "VerdeEsmeralda"] },
    { "key": "fecha",   "macro": "FechaEntrega", "type": "date",  "label": "Fecha de entrega", "default": "today" }
  ],
  "modes": ["completo", "rapido", "borrador"],
  "preview": "preview.png"
}
```

| Campo | Tipo | Obligatorio | Regla |
|-------|------|-------------|-------|
| `$schema` | string | No | Solo para el editor. Se ignora |
| `schemaVersion` | entero | Sí | `1` |
| `id` | string | Sí | §2. Igual al nombre de la carpeta |
| `version` | string | Sí | §2 (semver) |
| `name` | string | Sí | 1–80 caracteres |
| `description` | string | Sí | 1–300 caracteres, una línea |
| `category` | string | Sí | `^[a-z]+(-[a-z]+)*$`, igual a la carpeta padre. Recomendadas: `academico`, `presentacion`, `difusion`, `correspondencia`, `congresos` (una nueva da aviso, no error) |
| `tags` | string[] | No | Hasta 10, `^[a-z0-9]+(-[a-z0-9]+)*$`, sin repetir |
| `language` | string | Sí | `^[a-z]{2}(-[A-Z]{2})?$`. En este repositorio, `es` o `es-XX` |
| `license` | string | Sí | Expresión SPDX con `AND` (`MIT-0`, `MIT-0 AND LicenseRef-isprs`). Cada identificador tiene su texto en `LICENSES/<identificador>.txt` del repositorio de origen. **Política de este repositorio** (decisión del 2026-11-11, `NOTICE.md`): incluye `MIT-0`; el material de terceros añade un `LicenseRef-<nombre>` |
| `authors` | objeto[] | Sí | Al menos uno: `{ name (obligatorio), github?, email? }` |
| `entry` | string | Sí | Ruta relativa a `src/`, termina en `.tex`, debe existir |
| `build` | objeto | Sí | §3.1 |
| `requires` | objeto | No | §3.2 |
| `components` | string[] | No | Componentes de `shared/` que `inkt bundle` copia a `src/Latex/ink/`. Solo en el repositorio de origen (informativo dentro del bundle) |
| `parameters` | objeto[] | No | §4. Hasta 40 |
| `modes` | string[] | Sí | §5. Incluye `completo` |
| `preview` | string | No | PNG relativo a la carpeta de la plantilla, 200 KB como máximo |
| `dist` | objeto | No | **Prohibido en el repositorio.** Lo añade la publicación (§6) |
| `previewPdf` | objeto | No | **Prohibido en el repositorio.** Lo añade la publicación: la vista previa en PDF (§6.1) |

### 3.1 `build`

| Campo | Tipo | Regla |
|-------|------|-------|
| `engines` | string[] | No vacío, sin repetir, de `pdflatex`, `xelatex`, `lualatex`. **Política de este repositorio:** incluye `pdflatex` y `xelatex` (Creative-Ink compila hoy siempre con XeLaTeX) |
| `defaultEngine` | string | Pertenece a `engines`. **Política de este repositorio:** `pdflatex` |
| `bibliography` | string | `none`, `bibtex` o `biber`. Por defecto, `bibtex` (natbib/apalike). Si la plantilla ofrece Biber, lo hace con un parámetro `enum` `bibliografia` y aquí declara el valor por defecto |
| `shellEscape` | boolean | Por defecto `false`. **En v1, `true` está prohibido en este repositorio (V10) y los clientes bloquean la plantilla** (estado `blocked`): `--shell-escape` permite que un `.tex` ejecute comandos del sistema. No basta con "requerir confianza", porque en Creative-Ink una carpeta sin `.latexmkrc` ya cuenta como de confianza (revisión de seguridad del 2026-11-11). El campo se conserva para una versión futura del contrato |

### 3.2 `requires`

| Campo | Tipo | Regla |
|-------|------|-------|
| `creativeInk` | string | Solo la forma `>=X.Y.Z` en v1 |
| `texLive` | entero | Año mínimo de TeX Live (2020–2100). Informativo |
| `packages` | string[] | Nombres de `tlmgr` (`^[a-z0-9-]+$`), para el mensaje "instala con `tlmgr install …`" |
| `files` | string[] | Nombres de archivo que se comprueban con `kpsewhich` (`^[A-Za-z0-9._-]+$`): `natbib.sty`, `spanish.ldf` |

## 4. Parámetros

### 4.1 Campos comunes

| Campo | Regla |
|-------|-------|
| `key` | `^[a-z][a-zA-Z0-9]{0,39}$` (camelCase: `palabrasClave`). Único. Es lo que el usuario escribe en `clave=valor` |
| `macro` | `^[A-Za-z]{1,40}$`. Única. No puede ser reservada (§4.4) ni empezar por `ink` o `Ink` |
| `type` | `string`, `boolean`, `enum`, `date` o `list` |
| `label` | 1–60 caracteres. Lo que se muestra al usuario |
| `description` | Opcional, 300 caracteres como máximo |
| `group` | Opcional (`Portada`, `Documento`…). Solo para agrupar en la interfaz |
| `required` | Opcional, por defecto `false` |
| `default` | Opcional según el tipo (§4.2). Debe ser válido para el tipo |

### 4.2 Por tipo

| Tipo | Campos propios | `default` | Valor en `clave=valor` | Valor en `ink-params.tex` |
|------|----------------|-----------|------------------------|---------------------------|
| `string` | `maxLength` (1–1000, por defecto 300), `fromProjectName` (bool) | Opcional | Texto de una línea | `EscapeTeX(valor)` |
| `boolean` | — | **Obligatorio** | `true`/`false`; también `si`, `sí`, `no` (sin distinguir mayúsculas) | `true` o `false` |
| `enum` | `values`: 1–50 valores únicos, `^[A-Za-z0-9-]{1,40}$` | **Obligatorio**, uno de `values` | Uno de `values` (exacto) | El valor tal cual |
| `date` | — | Opcional: `AAAA-MM-DD` o `"today"` | `AAAA-MM-DD` (fecha real) o `today` | `AAAA-MM-DD` |
| `list` | `maxItems` (1–50, por defecto 20), `maxLength` por elemento (por defecto 300) | Opcional, `string[]` | Elementos separados por `;`, recortados; los vacíos se descartan | `\inkItem{e1}\inkItem{e2}` (cada uno con `EscapeTeX`) |

- Los valores de `enum` son seguros para nombres de archivo porque las plantillas los usan en `\input`
  (`\input{Latex/ink/paletas/\ColorTema}`).
- `fromProjectName: true` solo en un parámetro `string` por plantilla: si el usuario no da valor, se usa el nombre del
  proyecto que da Creative-Ink. Equivale al `{{TITULO}}` de sus plantillas propias.
- `"today"` se resuelve con la fecha local de quien crea el proyecto (la librería recibe `Now` para que sea testeable).

### 4.3 Resolución (`params.Resolve`)

Entrada: el manifiesto, un mapa `clave → texto`, el nombre del proyecto y la fecha actual. Salida: un valor por
parámetro o una lista de errores **por campo**. Códigos de error (estables; los usan los tests de los dos repositorios):

| Código | Cuándo |
|--------|--------|
| `unknown` | La clave no existe en el manifiesto (evita erratas silenciosas) |
| `required` | Obligatorio, sin valor, sin `default` y sin `fromProjectName` (o sin nombre de proyecto) |
| `type` | El texto no se puede interpretar como el tipo (booleano o fecha mal escritos) |
| `enum` | No es uno de `values` |
| `length` | Supera `maxLength` (en caracteres, no en bytes) o tiene saltos de línea |
| `items` | Una lista supera `maxItems` o un elemento supera su `maxLength` |

Orden de precedencia de un valor: lo que dio el usuario → `fromProjectName` → `default` → sin valor (solo si es
opcional; entonces no se escribe en `ink-params.tex`).

Orden de los errores: primero los de los parámetros, en el orden del manifiesto (uno por parámetro, el primero que
falle); después las claves desconocidas, en orden por bytes (`Color` antes que `portda`). Casos en
`tools/fixtures/data/params/resolve-cases.json`.

### 4.4 Macros reservadas

`borrador`, `SoloCapitulo`, cualquiera que empiece por `ink`/`Ink`, y las del núcleo de LaTeX que una plantilla podría
pisar por accidente: `title`, `author`, `date`, `thanks`, `today`, `maketitle`, `documentclass`, `usepackage`,
`begin`, `end`, `input`, `include`, `includeonly`, `part`, `chapter`, `section`, `subsection`, `paragraph`, `label`,
`ref`, `pageref`, `cite`, `item`, `caption`, `footnote`, `newcommand`, `renewcommand`, `providecommand`, `def`, `let`,
`relax`, `par`, `color`, `textcolor`, `LaTeX`, `TeX`.

## 5. Modos

| Modo | Macro que lo activa | Qué debe hacer la plantilla |
|------|---------------------|-----------------------------|
| `completo` | — | Compilación normal (todas las pasadas que pida latexmk) |
| `rapido` | — | Nada especial; lo decide el compilador (una pasada) |
| `borrador` | `\borrador` (definida, vacía) | `\PassOptionsToPackage{draft}{graphicx}` antes de cargar graphicx |
| `capitulo` | `\SoloCapitulo` = ruta relativa sin `.tex` | `\includeonly{\SoloCapitulo}`; omitir anexos y preliminares pesados |

Los nombres de macro son **fijos**: los inyecta `ci compile borrador|solo` de Creative-Ink por la línea de órdenes
(`"\def\borrador{}\input{main}"`). Los modos no se escriben en `ink-params.tex`.

## 6. Manifiesto publicado y `versions.json`

Al publicar, `inkt index` escribe `catalog/v1/templates/<id>/<versión>/template.json`: el manifiesto del repositorio
(sin `$schema` ni `components`) más:

```json
"dist": {
  "url": "https://github.com/galigaribaldi/Memorandum-tex/releases/download/ensayo-v1.0.0/ensayo-1.0.0.zip",
  "sha256": "9f2c…(64 hexadecimales en minúscula)",
  "size": 482113,
  "publishedAt": "2026-11-20T18:00:00Z"
}
```

### 6.1 Vista previa en PDF (`previewPdf`, decisión del 2026-11-15)

Creative-Ink muestra un PDF de muestra antes de descargar la plantilla. Lo genera la CI de publicación y se describe
en el manifiesto publicado (y en el índice, §7):

```json
"previewPdf": {
  "url": "https://github.com/galigaribaldi/Memorandum-tex/releases/download/ensayo-v1.0.0/ensayo-1.0.0-preview.pdf",
  "sha256": "…(64 hexadecimales en minúscula)",
  "size": 81234,
  "pages": 3
}
```

- **Cómo se genera:** la CI compila el `entry` con el `defaultEngine`, en modo `completo` y **sin `ink-params.tex`**
  (valores por defecto). El número de páginas sale del `.log` ("Output written on … (N pages)").
- **Dónde se aloja:** como asset de la **misma Release** que el zip (`<id>-<versión>-preview.pdf`). Releases no tiene
  límite total y cada versión es inmutable; Pages se reserva para los JSON y las miniaturas.
- **Límites (V17):** `size` de 1 byte a **5 MB**; `pages` de 1 a 10 000; `url` https; `sha256` de 64 hexadecimales.
  La CI rechaza la publicación si el PDF generado supera 5 MB.
- **Integridad:** su `sha256` va dentro del JSON firmado (§7.1), así que el cliente verifica el PDF **antes de
  mostrarlo**, igual que un zip. Se descarga por separado del bundle.
- **Opcional:** un catálogo de terceros puede no traerlo. `preview.png` se mantiene para las miniaturas.

`catalog/v1/templates/<id>/versions.json` (más nueva primero; rutas relativas a este archivo):

```json
{
  "schemaVersion": 1,
  "id": "ensayo",
  "versions": [
    { "version": "1.1.0", "manifest": "1.1.0/template.json", "publishedAt": "2026-12-01T10:00:00Z" },
    { "version": "1.0.0", "manifest": "1.0.0/template.json", "publishedAt": "2026-11-20T18:00:00Z" }
  ]
}
```

## 7. Índice `catalog/v1/index.json`

```json
{
  "schemaVersion": 1,
  "generatedAt": "2026-11-20T18:00:00Z",
  "source": {
    "id": "memorandum-tex",
    "name": "Memorandum-tex",
    "homepage": "https://github.com/galigaribaldi/Memorandum-tex"
  },
  "templates": [
    {
      "id": "ensayo",
      "name": "Ensayo académico",
      "description": "Ensayo con portada institucional opcional, índice y bibliografía APA.",
      "category": "academico",
      "tags": ["ensayo", "apa", "unam"],
      "language": "es-MX",
      "license": "MIT-0",
      "latest": "1.0.0",
      "engines": ["pdflatex", "xelatex", "lualatex"],
      "defaultEngine": "pdflatex",
      "bibliography": "bibtex",
      "shellEscape": false,
      "requires": { "creativeInk": ">=0.1.0" },
      "manifest": "templates/ensayo/1.0.0/template.json",
      "versions": "templates/ensayo/versions.json",
      "preview": "templates/ensayo/1.0.0/preview.png",
      "previewPdf": { "url": "https://github.com/…/ensayo-1.0.0-preview.pdf", "sha256": "…", "size": 81234, "pages": 3 },
      "dist": { "url": "https://github.com/…/ensayo-1.0.0.zip", "sha256": "…", "size": 482113 }
    }
  ],
  "packs": [
    { "id": "academico", "name": "Paquete académico", "description": "…", "manifest": "packs/academico.json" },
    { "id": "todo", "name": "Todas las plantillas", "description": "…", "manifest": "packs/todo.json" }
  ]
}
```

- Cada entrada de `templates` describe **la última versión** (`latest`); las anteriores se encuentran en `versions`.
- `source.id` sigue la regla de §2 e identifica el catálogo (los de terceros usan el suyo).
- Tamaño máximo que acepta la librería: 1 MB.

### 7.1 Firma del catálogo (decisión del 2026-11-11)

El `sha256` de `dist` protege contra un zip corrupto, pero no contra un catálogo comprometido: quien controle el sitio
podría publicar un zip malicioso con su hash correcto. Por eso **cada JSON del catálogo se firma**: `index.json`,
cada `versions.json`, cada manifiesto publicado y cada paquete. La firma va en un archivo hermano `<archivo>.sig`:

```json
{ "keyId": "memorandum-tex-2026", "algorithm": "ed25519", "signature": "<64 bytes en base64 estándar>" }
```

- **Qué se firma:** los bytes exactos del archivo, sin normalizar. Ed25519 está en la biblioteca estándar de Go
  (`crypto/ed25519`).
- **Por qué todos los JSON y no solo el índice:** el índice describe solo la última versión. Firmar también
  `versions.json` y los manifiestos antiguos impide servir una versión vieja con un zip falso. Los zips quedan cubiertos
  porque su `sha256` va dentro de un JSON firmado, y lo mismo vale para `previewPdf` (§6.1). Las miniaturas
  (`preview.png`) no se firman: solo se muestran como imagen, y la librería limita su tamaño.
- **Catálogo oficial:**
  - **Siempre firmado.** La librería lleva fija la clave pública (o las claves, por `keyId`) y **rechaza** cualquier
    JSON oficial sin firma válida (`ErrSignature`).
  - **Custodia de la clave privada:** vive solo en un secreto de GitHub Actions. La genera y la custodia el humano;
    nunca entra al repositorio.
- **Catálogos de terceros:** si su fuente declara una clave pública, la firma es obligatoria igual que en el oficial.
  Si no, se aceptan como **"de terceros, sin firma"** y los clientes lo muestran así.
- **Rotación:**
  1. Se publica una versión de la librería que ya conoce la clave nueva.
  2. Después se empieza a firmar con ella.
  3. La clave vieja se retira cuando ya no haya clientes que la necesiten.

## 8. Paquetes

En el repositorio: `packs/<id>.json`. Publicado: `catalog/v1/packs/<id>.json` (mismo contenido).

```json
{
  "schemaVersion": 1,
  "id": "academico",
  "name": "Paquete académico",
  "description": "Ensayo, artículo, registro de avances y tesis.",
  "templates": [{ "id": "ensayo" }, { "id": "articulo" }, { "id": "registro-avances" }, { "id": "tesis" }]
}
```

- `templates`: de 1 a 100 entradas `{ id, version? }`, sin ids repetidos. Sin `version`, se usa `latest` del índice.
- Un paquete **no tiene zip propio**: se descarga cada plantilla por separado.
- **`todo` está reservado:** lo genera `inkt index` con todas las plantillas publicadas.
- `inkt validate` avisa si un paquete nombra una plantilla que todavía no existe; `inkt index` lo trata como error.

## 9. Bundle `<id>-<versión>.zip`

```
ensayo-1.0.0.zip
└── ensayo/                     ← una sola carpeta raíz, igual al id
    ├── template.json           ← manifiesto del repositorio sin $schema (sin dist)
    ├── LICENSE
    ├── README.md
    ├── .creative-ink-meta      ← {"version": 1, "mainFile": "<entry>"}: lo genera inkt bundle
    ├── main.tex                ← todo lo que había en src/
    ├── Latex/ink/<componente>/… ← componentes declarados, copiados de shared/<componente>/ (p. ej.
    │                              Latex/ink/motor/ink-motor.tex); src/ no puede usar Latex/ink/
    └── …
```

| # | Regla |
|---|-------|
| B1 | Una sola carpeta raíz, igual al `id` del manifiesto que contiene |
| B2 | Cada entrada, sin la carpeta raíz, es una ruta local con `/` como separador: sin `/` inicial, sin `.` ni `..` como segmento, y **cada segmento cumple `^[A-Za-z0-9._-]+$`** (como V31). Así `C:` y `\` se rechazan en cualquier sistema, no solo en Windows. Además se comprueba `filepath.IsLocal` sobre la ruta convertida al sistema |
| B3 | Sin enlaces simbólicos ni entradas que no sean archivos o carpetas normales |
| B4 | Archivos permitidos: extensiones `tex sty cls bst bib bbx cbx lbx cfg def md txt json png jpg jpeg pdf eps svg otf ttf` (en minúsculas) y los nombres exactos `LICENSE` y `.creative-ink-meta`. Cualquier otro (incluidos `.latexmkrc`, `latexmkrc`, `Makefile` y `*.sh`) se rechaza. Los nombres de rc de latexmk se rechazan **sin distinguir mayúsculas** (`.LATEXMKRC` es el mismo archivo en macOS y Windows) |
| B5 | Como máximo 2000 archivos y 200 MB descomprimidos; ningún archivo **de 1 MB o más** con una relación de compresión mayor que 100:1 (bomba zip; un archivo pequeño de espacios en blanco supera esa relación sin ser un ataque). Al extraer, ningún archivo puede ocupar más de lo que declara el zip: se lee como mucho el tamaño declarado |
| B6 | Contiene `template.json`, `LICENSE`, `README.md`, `.creative-ink-meta` y el `entry`. `template.json` coincide en `id` y `version` con el índice. `LICENSE` lo genera `inkt bundle`: los textos de `LICENSES/<identificador>.txt` de cada identificador de `license`, en el orden de la expresión, separados por una línea en blanco |
| B7 | `.creative-ink-meta` tiene exactamente `version: 1` y `mainFile` igual al `entry`. La configuración del proyecto (motor, etc.) la escribe Creative-Ink, no el bundle |
| B8 | Reproducible: entradas ordenadas por ruta, fecha fija `2000-01-01T00:00:00Z`, método Deflate. Mismo contenido → mismo `sha256` en cualquier sistema **con la misma versión de Go** (la salida de Deflate puede cambiar entre versiones; por eso el hash se calcula al publicar y viaja en el catálogo firmado, nunca se recalcula en el cliente) |
| B9 | El `sha256` se verifica sobre el zip completo **antes** de extraer |
| B10 | Ningún par de entradas difiere solo en mayúsculas (`a.tex` y `A.tex`): en macOS y Windows una sobrescribiría a la otra al extraer. Se rechaza como ruta insegura |
| B11 | Defensa en profundidad: `SafeExtract` rechaza un bundle cuyo `template.json` pida `shellEscape: true` (`ErrForbiddenFile`), aunque `compat` ya lo haya bloqueado antes de descargar |

## 10. `ink-params.tex`

Lo escribe `params.RenderInkParams` en la raíz del proyecto (junto al `entry`). Formato exacto (los tests comparan
byte a byte):

```tex
% ink-params.tex — parámetros de ensayo 1.0.0.
% Generado a partir de template.json; puedes editar los valores.
\def\TituloDoc{Humedales urbanos \& movilidad}
\def\Autor{Hernán Galileo Cabrera Garibaldi}
\def\ConPortada{false}
\def\ColorTema{Teal}
\def\FechaEntrega{2026-11-10}
```

- Dos líneas de cabecera, luego una línea `\def\<macro>{<valor>}` por parámetro con valor, **en el orden del
  manifiesto**, cada una terminada en LF. UTF-8 sin BOM.
- **`EscapeTeX`** (idéntico a `core.EscapeTeX` de Creative-Ink, una sola pasada): `\` → `\textbackslash{}`, `{` → `\{`,
  `}` → `\}`, `$` → `\$`, `&` → `\&`, `#` → `\#`, `%` → `\%`, `_` → `\_`, `~` → `\textasciitilde{}`,
  `^` → `\textasciicircum{}`. Casos compartidos en `tools/fixtures/data/params/escape-cases.json`.
- Se usa `\def` (no `\newcommand`) para que cargar el archivo dos veces no falle y para que los `\providecommand`
  posteriores no lo pisen.

## 11. Requisitos de la plantilla LaTeX

1. `main.tex` (el `entry`) carga los parámetros **antes** de cualquier valor por defecto:
   ```tex
   \InputIfFileExists{ink-params}{}{}
   \providecommand{\TituloDoc}{Título del documento}
   \providecommand{\ConPortada}{true}
   ```
2. Cada `macro` del manifiesto tiene su `\providecommand` en `src/`, así la plantilla **compila sin `ink-params.tex`**.
3. Compila con cada motor de `build.engines` (el motor se elige en `Latex/ink/motor/ink-motor.tex` con iftex).
4. Respeta los modos que declara (§5).
5. Ninguna ruta sale de la carpeta (`\input`, `\include`, `\includegraphics`, `\graphicspath`, `\bibliography`).
6. Sin `--shell-escape` ni nada que lo necesite (externalización de TikZ, `minted`, `\write18`): en v1 está bloqueado.
7. Sin los marcadores `{{TITULO}}` ni `{{TITLE}}`: Creative-Ink los sustituye al copiar sus plantillas propias y se
   mezclaría con los parámetros (el título se escaparía dos veces). El título es un parámetro con `fromProjectName`.

## 12. Reglas de validación de `inkt validate`

Los códigos son estables: los usan los fixtures de `tools/fixtures/data/invalid/` (un archivo por regla, nombrado con su
código) y los mensajes de error.

Cada error se reporta **una sola vez, en su regla**: un `id` que ya viola V03 no se compara además con la carpeta
(V04), y una `schemaVersion` distinta de 1 detiene la validación en V02 (las demás reglas son de otro contrato).

**Manifiesto**

| Código | Regla |
|--------|-------|
| V01 | JSON válido, un objeto, UTF-8 sin BOM, 256 KB como máximo |
| V02 | `schemaVersion` = 1 |
| V03 | `id` cumple §2 |
| V04 | `id` igual al nombre de la carpeta y `category` igual a la carpeta padre |
| V05 | `version` es semver (§2) |
| V06 | `name`, `description`, `category` y `tags` cumplen §3 |
| V07 | `language` cumple §3 (y es `es`/`es-XX` en este repositorio) |
| V08 | `license` es una expresión SPDX con `AND`, cada identificador existe en `LICENSES/` y, en este repositorio, incluye `MIT-0`; `authors` con al menos un `name` |
| V09 | `entry` relativo, `.tex` y existente en `src/` |
| V10 | `build` cumple §3.1, incluida la política de motores de este repositorio y `shellEscape` distinto de `true` |
| V11 | `requires` cumple §3.2 |
| V12 | Parámetros: `key` y `macro` válidas y únicas; macro no reservada (§4.4) |
| V13 | Parámetros: tipo conocido, `label`, `description` y `group` dentro de su longitud, y campos propios del tipo válidos (§4.2); como máximo 40 parámetros |
| V14 | Parámetros: `default` válido para su tipo; `boolean` y `enum` lo llevan siempre |
| V15 | `fromProjectName` solo en un parámetro, y de tipo `string` |
| V16 | `modes` sin repetir, de §5, e incluye `completo` |
| V17 | `preview`, si existe: PNG, 200 KB como máximo. `previewPdf` (solo publicado): `url` https, `sha256` válido, `size` de 1 byte a 5 MB y `pages` de 1 a 10 000 (§6.1) |
| V18 | `components` existen en `shared/` |
| V19 | Sin `dist` ni `previewPdf` en el repositorio (los añade la publicación); en un manifiesto publicado, `dist` obligatorio y válido |
| V20 | Campo desconocido en cualquier nivel (raíz, `build`, `requires`, `dist`, `authors[]`, `parameters[]`): **aviso** (posible errata como `defualt`) |

**Contenido de la carpeta**

| Código | Regla |
|--------|-------|
| V30 | Existen `README.md`, `CHANGELOG.md` (con un encabezado `## <version>`) y `src/` |
| V31 | Nombres de archivo y carpeta en `src/`: `^[A-Za-z0-9._-]+$` por segmento, y sin dos rutas que difieran solo en mayúsculas (B10) |
| V32 | Archivos de `src/` permitidos por B4 (sin `LICENSE` ni `.creative-ink-meta`: los añade `inkt bundle`) |
| V33 | Sin enlaces simbólicos |
| V34 | Límites de B5 (contando los componentes resueltos) |
| V35 | Sin rutas que salgan de la plantilla en `\input`, `\include`, `\includegraphics`, `\graphicspath`, `\bibliography`, fuera de comentarios (`% !TeX root = ../main.tex` está permitido) |
| V36 | `\InputIfFileExists{ink-params}` aparece en el `entry` antes del primer `\providecommand` de una macro de parámetro |
| V37 | Cada macro de parámetro tiene un `\providecommand{\<macro>}` en `src/` |
| V38 | Archivos de texto en UTF-8 válido, sin BOM y con LF |
| V39 | Sin `--shell-escape` ni sus usos (`\usetikzlibrary{external}`, `\tikzexternalize`, `minted`, `\write18`, `\ShellEscape`) y sin los marcadores `{{TITULO}}`/`{{TITLE}}` en `src/` (§11, reglas 6 y 7) |

**Paquetes**

| Código | Regla |
|--------|-------|
| V50 | `id` cumple §2, igual al nombre del archivo y distinto de `todo` |
| V51 | `templates` cumple §8 |
| V52 | Cada id existe: aviso en `validate`, error en `index` |

## 13. Errores de la librería

Tipados, comparables con `errors.Is` o `errors.As`:

| Error | Cuándo |
|-------|--------|
| `ErrOffline` | No hay red y no hay caché (o se pidió algo que no está en caché) |
| `ErrNotFound` | El id, la versión o el paquete no existen en el catálogo (404 o ausente del índice) |
| `ErrUnsupportedSchema` | El índice o un manifiesto tienen una `schemaVersion` que la librería no conoce |
| `ErrChecksum` | El `sha256` del zip descargado no coincide |
| `ErrTooLarge` | Se supera un límite (índice, manifiesto, bundle, archivos, relación de compresión) |
| `ErrUnsafePath` | Una entrada del zip viola B2 o B3 |
| `ErrForbiddenFile` | Una entrada del zip viola B4 |
| `ErrBundleLayout` | Se viola B1, B6 o B7 |
| `ErrDestExists` | El destino existe y no está vacío |
| `ErrSignature` | Un JSON de un catálogo firmado no tiene `.sig`, la firma no es válida o su `keyId` es desconocido (§7.1) |
| `*params.Error` | Errores por campo de §4.3 |

## 14. Cambios respecto al borrador v0 (`docs/integracion-creative-ink.md` §4)

- `.creative-ink-meta` **lo genera `inkt bundle`** a partir de `entry` (no se escribe en `src/`) y solo lleva
  `version` y `mainFile`.
- Se añaden `versions.json` (§6), la regla de resolución de URLs relativas (§2) y los nombres de archivo exactos
  permitidos (`LICENSE`, `.creative-ink-meta`).
- Tipos de parámetro precisados (§4.2), códigos de error de resolución (§4.3) y macros reservadas (§4.4).
- `ink-params.tex` con formato exacto y cabecera fija (§10).
- Bundle reproducible (B8) y protección contra bombas zip (B5).
- Errores nuevos: `ErrForbiddenFile` y `ErrBundleLayout`.
- `category` abierta (patrón + lista recomendada) para que añadir una categoría no requiera tocar el contrato.
- Rutas del zip con el mismo juego de caracteres que V31 (B2), para rechazar `C:` y `\` en todos los sistemas.
- (2026-11-11) Licencias:
  - `license` es una expresión SPDX con `AND`, resuelta contra `LICENSES/`.
  - El `LICENSE` del bundle lo genera `inkt bundle` (B6).
  - En este repositorio las plantillas son `MIT-0` y el resto es `MIT` (`NOTICE.md`).
- (2026-11-11) Revisión de seguridad del agente de Creative-Ink:
  - **`shellEscape: true`:** bloqueado en v1 (antes "requiere confianza", que en Creative-Ink no protege), con la
    nueva regla V39.
  - **Rutas:** se rechazan las que colisionan por mayúsculas (B10) y los rc de latexmk sin distinguir mayúsculas (B4).
  - **Marcadores:** prohibidos `{{TITULO}}` y `{{TITLE}}`.
  - **Firma:** firma Ed25519 de cada JSON del catálogo (§7.1) y el error nuevo `ErrSignature`.
- (2026-11-15) Al implementar `pkg/bundle` y la validación de carpeta:
  - **Componentes:** van a `Latex/ink/<componente>/`, para que dos componentes nunca choquen.
  - **B5:** la relación 100:1 aplica a archivos de 1 MB o más, y la extracción nunca escribe más de lo declarado.
  - **B8:** el `sha256` es reproducible con la misma versión de Go.
  - **B11:** `SafeExtract` también rechaza `shellEscape`.
  - **Extensiones:** se comparan exactas y en minúsculas (`figura.PNG` no vale; se renombra).
- (2026-11-15) Solicitud de Creative-Ink: **vista previa en PDF** (`previewPdf`, §6.1).
  - **Dónde va:** en el manifiesto publicado y en el índice, como bloque opcional y compatible (no sube
    `schemaVersion`).
  - **Contenido:** `url`, `sha256`, `size` (≤ 5 MB) y `pages`.
  - **Alojamiento:** Releases.
  - **Validación:** prohibido en el repositorio (V19) y validado en V17.
