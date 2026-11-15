---
name: latex-template-author
description: Crear, migrar o editar una plantilla LaTeX de Memorandum-tex (templates/<categoria>/<id>/) o un componente compartido (shared/). Úsala al migrar plantillas desde Trabajos-Maestria-Urbanismo o TesisUrbanismo, al añadir parámetros o modos a main.tex, al tocar paletas, motor, comandos o instituciones, y antes de dar por terminada una plantilla.
---
# Autoría de plantillas LaTeX

Lee primero: `CLAUDE.md` y `docs/MAP.md`. Si discrepan con esta skill, ganan ellos. El formato del manifiesto lo
define la skill `template-contract` (y `docs/contrato-plantillas.md` cuando exista).

## Contexto del proyecto (verifícalo antes de fiarte: puede haber cambiado)
- Plantillas: `templates/<categoria>/<id>/{template.json, README.md, CHANGELOG.md, preview.png, src/}`.
- Componentes: `shared/motor/ink-motor.tex`, `shared/base/ink-base.tex`, `shared/paletas/`, `shared/comandos/`,
  `shared/instituciones/<id>/`. Dentro del bundle quedan en `src/Latex/ink/`.
- Origen de las plantillas (solo lectura):
  - `/Users/hcabrera/Documents/Personal_Documents/Trabajos-Maestria-Urbanismo/` (`DocumentosPlantilla/`,
    `PresentacionPlantilla/`, `InfografiaPlantilla/`, `Makefile`, `Img/`).
  - `/Users/hcabrera/Documents/Personal_Documents/TesisUrbanismo/` (tesis y su flujo de seguimiento).
- Tabla de migración: `docs/MAP.md` §5. Deuda heredada que se corrige al migrar: `docs/MAP.md` §4.9.

## Principios
1. **Autocontenida:** nada fuera de la carpeta de la plantilla. Lo compartido entra solo por `components`.
2. **El manifiesto decide:** motor, bibliografía, parámetros y modos se declaran en `template.json`, nunca en un
   Makefile, en un `.latexmkrc` ni en comentarios que alguien tenga que leer con `grep`.
3. **Compila sin Creative-Ink:** sin `ink-params.tex`, los `\providecommand` dan valores por defecto razonables.
4. **Motor neutro:** pdfLaTeX por defecto, y también compila con XeLaTeX (Creative-Ink usa hoy `latexmk -xelatex`).
5. **Sin institución escrita a mano:** "UNAM", "FES Acatlán" y los logos vienen de `shared/instituciones/`.
6. **Nunca ejecuta código:** sin `--shell-escape` ni lo que lo necesita (externalización de TikZ, `minted`,
   `\write18`), sin `.latexmkrc` y sin `{{TITULO}}` (V39, MAP decisión 15). Si la plantilla de origen lo usaba, se
   rediseña sin ello; no se "marca como de confianza".

## Pasos para migrar una plantilla
1. Lee la plantilla de origen completa (`main.tex`, `Latex/*`, `secciones/*`) y su entrada en `MAP §5`.
2. Une las variantes que solo difieren en portada o cabecera en un parámetro booleano (`portada`).
3. Convierte el bloque "PERSONALIZACIÓN" en parámetros del manifiesto. En `main.tex`:
   ```tex
   \InputIfFileExists{ink-params}{}{}   % siempre primero
   \providecommand{\TituloDoc}{…}        % después, los valores por defecto
   ```
4. Sustituye `inputenc`/`fontenc`/`lmodern` por `\input{Latex/ink/ink-motor}`, las paletas copiadas por el componente
   `paletas` y los logos por la institución. Quita `\graphicspath` con `../`.
5. Respeta los modos: `\ifdefined\borrador` → `\PassOptionsToPackage{draft}{graphicx}`; `\ifdefined\SoloCapitulo` →
   `\includeonly{\SoloCapitulo}`. Los nombres son fijos (los inyecta `ci compile` de Creative-Ink).
6. Corrige la deuda de §4.9 que aplique y táchala en el MAP con fecha.
7. Renombra a ASCII en kebab-case (`Comands.tex` → `comandos.tex`, `secciones/1-Introduccion.tex` →
   `secciones/1-introduccion.tex`).
8. Escribe el README (parámetros y modos coinciden con el manifiesto) y el `CHANGELOG.md`.
9. Pasa la verificación y entrega al humano el comando exacto para compilarla.

## Verificación (solo lectura; no instala ni compila nada)
Mientras no exista `inkt validate` (M2), desde la raíz del repositorio y sustituyendo `<ruta>` por la carpeta de la
plantilla:
```bash
grep -rnE '^[^%]*\\(input|include|includegraphics|bibliography|graphicspath)[^%]*\.\./' <ruta>/src   # V35, sin comentarios
grep -rnE '\\usepackage(\[[^]]*\])?\{(inputenc|fontenc|lipsum|subfigure|epsfig)\}' <ruta>/src
grep -rn 'UNAM\|FES Acatl' <ruta>/src --include=*.tex                 # textos institucionales a mano
find <ruta> -name '* *' -o -iname '.latexmkrc' -o -iname 'latexmkrc' -o -type l -o -name '*.sh' -o -name 'Makefile'
find <ruta> -name '*[!A-Za-z0-9._/-]*'                                 # nombres no ASCII
find <ruta>/src | tr 'A-Z' 'a-z' | sort | uniq -d                       # V31/B10: colisiones por mayúsculas (vacío = bien)
grep -rnE 'tikzexternalize|usetikzlibrary\{[^}]*external|\{minted\}|write18|ShellEscape|\{\{TIT(ULO|LE)\}\}' <ruta>/src   # V39
grep -n 'InputIfFileExists{ink-params}' <ruta>/src/main.tex           # debe aparecer antes de los \providecommand
```
Después, **el humano** compila: `latexmk -pdf main.tex` dentro de `src/`, y con `-xelatex` (y `-lualatex` si el
manifiesto lo lista). Si no se pudo compilar, dilo en el informe.

Al revisar sus resultados, cuenta errores y desbordes con `grep -a` (el `.log` no es UTF-8):
```bash
grep -acE '^!|Warning|Overfull|Underfull' <job>.log      # una plantilla terminada debe dar 0
grep -anE -A3 '^!|Warning|Overfull|Underfull' <job>.log  # el detalle
```
Para conservar el registro de cada motor o modo, compila con `-jobname=<paso>` en lugar de `latexmk -C` entre pasos.

## Gotchas (verificados en el código de los repositorios de origen)
- `_compile` del Makefile antiguo pasaba `\def\ColorTema{Institucional}` en cada compilación, por eso
  `\providecommand` en `color-config.tex` nunca ganaba. Con `ink-params` el orden lo arregla.
- `babel` con `spanish` activa `<`, `>` y `~`: carga `\usetikzlibrary{babel}` con TikZ y evita `~` en `\label`.
- `\FirmaBloqueDoble` de CartaUNAM tiene los argumentos cruzados (`#2 — Visto Bueno` y luego `#3`).
- La plantilla Apimetro no tiene `Institucional`: su `enum` de color es `Apimetro | Negro`.
- ISPRS trae `isprs.cls` oficial (licencia de terceros): no lo reescribas; documenta su origen en el README.
- La infografía depende de `\vspace*{4.3cm}` para librar la cabecera dibujada con TikZ; si cambias la cabecera,
  recalcula ese espacio.
- `% !TeX root = ../main.tex` en los archivos de `secciones/` es un comentario mágico para editores, no una ruta que
  cargue LaTeX: está permitido. V35 solo mira comandos fuera de comentarios.
- **El `.log` de TeX no es UTF-8** (pdfTeX escribe ISO-8859): un `grep` sin `-a` lo trata como binario y cuenta 0
  aunque haya errores. Ya pasó el 2026-11-11 y escondió un `Overfull \hbox` de 140 pt.
- **`\verb` no se parte:** un comando largo dentro de un párrafo desborda la línea. Ponlo en su propia línea o en un
  entorno `verbatim`.
