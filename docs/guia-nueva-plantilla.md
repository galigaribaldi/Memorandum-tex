# Guía técnica: escribir una plantilla nueva

Complementa a `CONTRIBUTING.md` (el proceso) con el **cómo** técnico. La referencia completa es
`docs/contrato-plantillas.md`; si esta guía y el contrato discrepan, gana el contrato.

## 1. Esqueleto

```
templates/academico/informe-tecnico/
├── template.json
├── README.md
├── CHANGELOG.md
├── preview.png
└── src/
    ├── main.tex
    ├── Latex/
    │   └── preambulo.tex      ← paquetes y comandos propios de la plantilla
    ├── secciones/
    │   ├── 1-introduccion.tex
    │   └── 2-desarrollo.tex
    ├── img/
    └── referencias.bib
```

Partir de `tools/fixtures/data/templates/prueba/ok-minima/` es lo más rápido: es la plantilla válida más pequeña posible.

## 2. Del bloque "PERSONALIZACIÓN" a parámetros

Lo que antes se editaba a mano en `main.tex` pasa a declararse en `template.json` y a tener un valor por defecto en
`main.tex`.

**Antes** (plantillas de Trabajos-Maestria-Urbanismo):
```tex
\newcommand{\TituloEnsayo}{Título del ensayo}
\newcommand{\Autor}{Nombre del autor}
```

**Ahora**, en `template.json`:
```json
"parameters": [
  { "key": "titulo", "macro": "TituloDoc", "type": "string", "label": "Título", "required": true, "fromProjectName": true },
  { "key": "autor", "macro": "Autor", "type": "string", "label": "Autor", "required": true }
]
```

y en `main.tex`, siempre en este orden:
```tex
\InputIfFileExists{ink-params}{}{}          % 1. lo que eligió el usuario
\providecommand{\TituloDoc}{Título del documento}   % 2. valores por defecto
\providecommand{\Autor}{Nombre del autor}
```

| Lo que había | Tipo de parámetro | Ejemplo |
|--------------|-------------------|---------|
| Variante con o sin portada (dos carpetas) | `boolean` | `portada`, comparado con `\ifthenelse{\equal{\ConPortada}{true}}{…}{…}` |
| `COLOR=Teal` en el Makefile | `enum` | `color` con los nombres de las paletas |
| `TAMANIO=A3Portrait` | `enum` | `tamano` |
| Fecha de entrega | `date` | `AAAA-MM-DD`; `"default": "today"` |
| Miembros del comité, palabras clave | `list` | Se recorre con `\inkItem` |

## 3. Modos

Declara en `modes` solo los que la plantilla respeta:

```tex
\ifdefined\borrador
  \PassOptionsToPackage{draft}{graphicx}   % antes de \usepackage{graphicx}
\fi
\usepackage{graphicx}
\ifdefined\SoloCapitulo
  \includeonly{\SoloCapitulo}              % los capítulos con \include, no \input
\fi
```

## 4. Motor y bibliografía

- No cargues `inputenc`, `fontenc` ni `fontspec` a mano: declara el componente `motor` y usa
  `\input{Latex/ink/ink-motor}` (⏳ M3). Mientras tanto, el patrón de `ok-minima` (con iftex) es el equivalente.
- Por defecto: natbib con `\bibliographystyle{apalike}` y BibTeX (`"bibliography": "bibtex"`). Si quieres ofrecer Biber,
  añade un parámetro `enum` `bibliografia` con `["bibtex", "biber"]`.
- `build.engines` debe incluir `pdflatex` y `xelatex`; añade `lualatex` si la compilaste con él.

## 5. Antes de abrir el PR

1. Compila dentro de `src/` con cada motor: `latexmk -pdf main.tex`, `latexmk -xelatex main.tex`.
2. Compila sin `ink-params.tex` (los valores por defecto deben bastar) y con uno escrito a mano.
3. Si declaras `borrador` o `capitulo`, compila con `latexmk -pdf -usepretex='\def\borrador{}' main.tex`.
4. `go run ./tools/cmd/inkt validate templates/<categoria>/<id>` (⏳ M2).
5. Genera `preview.png` de la primera página (200 KB como máximo).
