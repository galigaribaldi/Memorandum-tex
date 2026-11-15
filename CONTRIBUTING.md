# Cómo contribuir a Memorandum-tex

Gracias por proponer una plantilla o una mejora. Este repositorio está pensado para crecer con plantillas de otras
personas **sin tocar nada central**: una plantilla nueva es una carpeta con su manifiesto, y el catálogo se genera
solo.

> **Estado:** el contrato v1 y la herramienta `inkt` se están construyendo (fases M1 y M2, ver `docs/MAP.md` §4.8).
> Hasta entonces, los pasos marcados con ⏳ todavía no se pueden ejecutar.

## 1. Antes de empezar

- Lee `docs/MAP.md` (estructura, reglas de capas y contrato) y, cuando exista, `docs/contrato-plantillas.md` (⏳ M1).
- Trabaja en una rama `DEV-feat-<Tema>` que nazca de `DEV`. Los cambios llegan a `DEV` con `git merge --ff-only` y a
  `main` cuando una etapa está probada.
- Idioma: documentación, ids de plantilla y macros LaTeX en **español**; claves del manifiesto e identificadores de Go en
  **inglés**.

## 2. Agregar una plantilla

1. **Elige el id y la categoría.** El id va en kebab-case ASCII (`informe-tecnico`, `poster-a0`) y debe ser único.
   Categorías actuales: `academico`, `presentacion`, `difusion`, `correspondencia`, `congresos`. Si ninguna encaja,
   propón una nueva en el PR.
2. **Crea la carpeta** `templates/<categoria>/<id>/` con:
   ```
   template.json     ← manifiesto (contrato v1)
   README.md         ← para qué sirve, parámetros, modos, ejemplo de uso
   CHANGELOG.md      ← empieza en 1.0.0
   preview.png       ← miniatura de la primera página, 200 KB como máximo
   src/              ← lo que recibe el usuario: main.tex, Latex/, secciones/, referencias.bib, img/
   ```
3. **Declara los parámetros** en `template.json`: título, autor, variantes como booleanos (`portada`), opciones como
   `enum` (`color`). En `main.tex`, carga primero `ink-params` y después los valores por defecto:
   ```tex
   \InputIfFileExists{ink-params}{}{}
   \providecommand{\TituloDoc}{Título del documento}
   \providecommand{\ConPortada}{true}
   ```
4. **Usa los componentes compartidos** (`shared/`) en lugar de copiar paletas, logos o el preámbulo de motor. Decláralos
   en `components` (⏳ M3).
5. **Compila con pdfLaTeX** (`latexmk -pdf main.tex` dentro de `src/`) y, si el manifiesto los lista, con XeLaTeX y
   LuaLaTeX.
6. **Valida:** `go run ./tools/cmd/inkt validate templates/<categoria>/<id>` (⏳ M2).
7. **Abre el PR** con la lista de comprobación de §4.

## 3. Reglas que el validador hace cumplir

- Nada fuera de la carpeta de la plantilla: sin `../`, sin rutas absolutas, sin `\graphicspath` hacia arriba.
- Sin `.latexmkrc` (en ninguna combinación de mayúsculas), sin enlaces simbólicos, sin Makefiles ni scripts de shell,
  sin Git LFS.
- Nombres de archivo en ASCII, sin espacios y en kebab-case (`marco-teorico.tex`, no `Marco Teórico.tex`), y sin dos
  archivos que solo difieran en mayúsculas (`intro.tex` e `Intro.tex` son el mismo en macOS y Windows).
- Fin de línea LF y UTF-8 (lo garantizan `.gitattributes` y `.editorconfig`).
- Extensiones permitidas en `src/`: `tex sty cls bst bib bbx cbx lbx cfg def md txt json png jpg jpeg pdf eps svg otf ttf`.
- Como máximo 2000 archivos y 200 MB por plantilla. Sin imágenes pesadas de ejemplo.
- **Sin `--shell-escape`, nunca** (`"shellEscape": false`). Permite que un `.tex` ejecute comandos en el equipo de
  quien lo compila, y v1 del contrato lo prohíbe: Creative-Ink bloquea cualquier plantilla que lo pida. Eso descarta
  la externalización de TikZ (`\usetikzlibrary{external}`), `minted`, `\write18` y similares (regla V39).
- Sin los marcadores `{{TITULO}}` ni `{{TITLE}}`: el título es un parámetro con `"fromProjectName": true`.
- Sin `lipsum`: el texto de muestra lo da el repositorio (⏳ M8).
- Licencia (`NOTICE.md`):
  - **Plantillas:** `"license": "MIT-0"`. Si incluyes material de terceros (por ejemplo, la clase oficial de un
    congreso), añade `AND LicenseRef-<nombre>` y su texto en `LICENSES/LicenseRef-<nombre>.txt`. Indica también su
    origen en el README de la plantilla.
  - **Logos institucionales:** no se licencian; van en `shared/instituciones/<id>/logos/`.

## 4. Lista de comprobación del PR

- [ ] `template.json` válido y `version` actualizada (semver: corrección = parche, parámetro nuevo = menor, cambio que
      rompe proyectos existentes = mayor).
- [ ] `CHANGELOG.md` con la entrada de la versión.
- [ ] README de la plantilla al día con sus parámetros y modos.
- [ ] Compila sin errores con cada motor de `build.engines` (indica cuáles probaste y en qué sistema operativo).
- [ ] `inkt validate` sin errores (⏳ M2).
- [ ] Si cambiaste el contrato, `schema/`, `tools/fixtures/data/` y `docs/integracion-creative-ink.md` cambian en el mismo PR.

Al contribuir aceptas que tu aporte se publique con la licencia que corresponde a su tipo de contenido (MIT-0 para
plantillas y `shared/`, MIT para lo demás); conservas el copyright de lo que escribas.

## 5. Commits

```
<tipo>(<ámbito>): <qué cambia, en español>

<Por qué y cómo, en 1–3 frases.>
```

Tipos: `feat`, `fix`, `docs`, `test`, `build`, `chore`. El ámbito suele ser el id de la plantilla (`feat(ensayo): …`),
`contrato`, `tools` o `shared`. Cada commit deja el repositorio coherente.

## 6. Publicación

Una versión se publica con un tag anotado `<id>-v<versión>` (por ejemplo `ensayo-v1.2.0`). La CI valida, empaqueta, sube
el zip a la Release y actualiza el catálogo en GitHub Pages (⏳ M7). Los tags los crean los mantenedores.
