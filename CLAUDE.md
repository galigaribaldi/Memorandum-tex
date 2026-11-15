# Memorandum-tex — Contexto para Claude
Repositorio de plantillas LaTeX en español, colaborativo y para varios motores. Creative-Ink las descarga bajo demanda
(macOS, Linux y Windows).
Mapa de arquitectura, contrato, decisiones y deuda técnica: @docs/MAP.md
Integración con Creative-Ink (plan y requerimientos para su agente): `docs/integracion-creative-ink.md`.
Método de trabajo: `WORKFLOW_BLUEPRINT.md` (Tríada, contrato vivo, deuda como artefacto, verificación objetiva).

## Stack
- **LaTeX:** motor por defecto **pdfLaTeX**; XeLaTeX y LuaLaTeX opcionales por plantilla (vía `ink-motor.tex` con iftex).
  Bibliografía por defecto **natbib/apalike + BibTeX**; biblatex + Biber solo como parámetro opcional.
- **Herramientas:** **Go, solo biblioteca estándar** (`tools/`, módulo `github.com/galigaribaldi/Memorandum-tex/tools`).
  Python solo para extensiones concretas futuras, nunca en el camino principal.
- **Manifiestos:** JSON con JSON Schema (`template.json`, `index.json`, paquetes).
- **Publicación:** catálogo en GitHub Pages (`catalog/v1/`, **cada JSON firmado con Ed25519**), bundles zip en GitHub
  Releases (tag `<id>-v<versión>`). La clave privada vive solo en un secreto de GitHub Actions.
- **Prohibido:** Makefiles o scripts de shell como herramienta de construcción; Git LFS; dependencias Go externas;
  `.latexmkrc` dentro de una plantilla; rutas `../` que salgan de la plantilla; nombres de archivo con espacios o acentos.

## Ubicación del código
| Carpeta | Qué contiene |
|---------|--------------|
| `templates/<categoria>/<id>/` | Una plantilla: `template.json`, `README.md`, `CHANGELOG.md`, `preview.png`, `src/` |
| `shared/` | Componentes reutilizables (motor, paletas, comandos, instituciones). Se resuelven dentro de cada bundle |
| `packs/` | Definiciones de paquetes (listas de ids) |
| `schema/` | JSON Schema del contrato |
| `tools/` | Herramienta `inkt` y la librería que importa Creative-Ink (`pkg/contract`, `params`, `compat`, `bundle`, `client`) |
| `tools/fixtures/` | Datos de prueba del contrato en `data/` (incluye casos maliciosos), incrustados con `embed.FS` para que viajen con el módulo |
| `docs/` | MAP, contrato, integración con Creative-Ink, guías |

## Visión
- **Bajo demanda:** el usuario descarga una plantilla o un paquete; nada se instala de golpe. Dónde se guarda dentro de
  `My-Creative-Ink-Repository` lo decide Creative-Ink.
- **Personalizable antes de usar:** el manifiesto declara parámetros (variantes como booleanos), motores y modos; la
  librería valida los valores y escribe `ink-params.tex`.
- **Extensible sin tocar nada central:** una plantilla nueva es una carpeta con su manifiesto; el índice se genera.
- **Licencias (2026-11-11, `NOTICE.md`):** MIT-0 para plantillas y `shared/`, MIT para el resto. Los logos
  institucionales y el material de terceros quedan excluidos.
- Decisiones tomadas: MAP §4.8. **Pendiente (no la asumas, espera instrucciones):** identidad, es decir, el texto de
  muestra con el nombre y la descripción del repositorio (fase M8).

## Skills del proyecto (`.claude/skills/`)
- **`latex-template-author`:** crear, migrar o editar una plantilla o un componente de `shared/`. Incluye la lista de
  verificación de solo lectura.
- **`template-contract`:** cualquier cambio en `template.json`, `schema/`, `packs/`, `tools/pkg/*`, el catálogo o
  `docs/integracion-creative-ink.md`.

## Reglas de trabajo (OBLIGATORIAS)
Las hace cumplir `.claude/settings.json`; respétalas aunque un permiso falte.
1. **Archivos:** no borrar archivos ni carpetas. Leer, listar y escribir siempre con permiso del usuario.
2. **Git:** no `git add`, `git commit`, `git push`, `git tag`, borrar ramas ni reescribir historia. Se proponen los
   commits y los comandos exactos; el usuario los ejecuta. Para ver tags usa `git for-each-ref refs/tags` (`git tag`
   está bloqueado incluso para listar).
3. **El humano ejecuta:** compiladores LaTeX (`latexmk`, `pdflatex`, `xelatex`, `bibtex`, `biber`), `tlmgr`,
   `go get`/`go mod tidy`/`go install`, `gh release`, configuración de GitHub Pages y cualquier publicación.
4. **Verificación permitida (con permiso):** `gofmt -l`, `go vet ./...`, `GOOS=<os> go vet ./...`, `go test -race ./...`,
   `go build -o /tmp/…`, y `go run ./tools/cmd/inkt validate` cuando exista.
5. **Idioma:** comunicación, documentación, ids de plantilla y macros LaTeX en **español**; identificadores de Go y
   claves del manifiesto en **inglés**.
6. **Commits propuestos:** `<tipo>(<ámbito>): <qué cambia, en español>`; tipos `feat`, `fix`, `docs`, `test`, `build`,
   `chore`; sin atribución a la IA (regla del blueprint, §2 y §8.4). Cada commit deja el repositorio coherente.
7. **Ramas:** `main` (estable) ← `DEV` (integración) ← `DEV-feat-<Tema>` (trabajo, nace de `DEV`). Integración con
   `git merge --ff-only`. Antes de proponer commits, confirma la rama con `git branch --show-current`.
8. **Cada cambio arrastra su documentación** en el mismo commit: MAP (mapa, contrato, decisiones, deuda), este archivo
   (estado, gotchas) y la skill del dominio. Un cambio de contrato actualiza también el esquema, los fixtures y
   `docs/integracion-creative-ink.md`.

## Gotchas técnicos
- **Creative-Ink compila hoy siempre con `latexmk -xelatex`** (`compile.engine` está pendiente allí): toda plantilla debe
  compilar también con XeLaTeX. Por eso el motor se elige en `ink-motor.tex` con iftex, nunca con `inputenc` suelto.
- **Macros de modo con nombre fijo:** `\borrador` y `\SoloCapitulo` son las que inyecta `ci compile borrador|solo` de
  Creative-Ink. No las renombres.
- **`ink-params.tex` se carga antes de los valores por defecto** (`\InputIfFileExists{ink-params}{}{}` y después
  `\providecommand`). Al revés, el parámetro nunca gana. Este es el origen del bug del `COLOR` en el Makefile antiguo.
- **Creative-Ink confía en el `.latexmkrc` de las plantillas propias:** una plantilla descargada nunca debe traer uno
  (el validador lo rechaza, también como `.LATEXMKRC`).
- **`--shell-escape` está prohibido en v1** (MAP decisión 15). "Requiere confianza" no protege: en Creative-Ink una
  carpeta sin `.latexmkrc` ya es de confianza. Nada de externalización de TikZ, `minted` ni `\write18`.
- **Nombres que solo difieren en mayúsculas** (`a.tex` y `A.tex`) se sobrescriben en macOS y Windows: se rechazan.
  En macOS (APFS no distingue mayúsculas) el test de carpeta de ese caso se omite; B10 de los bundles sí se prueba.
- **La lista de archivos permitidos vive solo en `contract/files.go`:** validación y extracción la comparten. No la
  copies en otro paquete.
- **Límites de copia de Creative-Ink:** 2000 archivos y 200 MB por plantilla. Sin imágenes pesadas en `src/`.
- **Los `.log` de TeX no son UTF-8** (pdfTeX escribe ISO-8859): revisa con `grep -a`, o un conteo de errores da 0
  en falso. En Go, léelos como bytes, nunca con validación UTF-8.
- **babel `spanish`** activa `<`, `>` y `~`; dentro de TikZ usa `\usetikzlibrary{babel}`, y en diffs, `\nobreakspace`.
- **Logos:** archivos normales de git (sin LFS); un zip de GitHub con LFS trae punteros, no imágenes.
- **Caracteres invisibles:** nunca escribas un BOM o un espacio de ancho cero literal (ni en fixtures): usa su escape
  (`\ufeff` en JSON). El hook de marcas lo bloquea; corrígelo por bytes con `perl -CSD -i -pe 's/\x{FEFF}/\\ufeff/g'`.

## Estado actual
| Fase | Qué | Estado |
|------|-----|--------|
| M0 | Gobierno: Tríada, CONTRIBUTING, README, `.gitattributes`, `.gitignore`, `.editorconfig`, skills | Hecho (2026-11-10, `20852d6`) |
| M1 | Contrato v1 (`docs/contrato-plantillas.md`), 4 esquemas, 4 paquetes, fixtures, guía técnica, licencias | Hecho (2026-11-11, `6513b66`) |
| M2 | Librería Go y `inkt`: ✅ `contract`, ✅ `params`, ✅ `compat`, ✅ validación de carpeta, ✅ `bundle`; siguen `catalogtest` (con firma), `client` (verifica firma), `cmd/inkt` | En curso (`DEV-feat-Contrato-Plantillas`, pasos 1–5 de 8) |
| M3–M8 | Componentes, piloto `ensayo`, migración, tesis, publicación, identidad | Pendiente (MAP §4.8) |
