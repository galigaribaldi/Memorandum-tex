---
name: template-contract
description: Contrato entre Memorandum-tex y Creative-Ink. Úsala para cualquier cambio en template.json, schema/*.json, packs/*.json, el índice del catálogo, el formato de los bundles, ink-params.tex, la librería Go (tools/pkg/contract, params, compat, bundle, client), los fixtures de tests/fixtures/ o docs/integracion-creative-ink.md.
---
# Contrato de plantillas

Lee primero: `CLAUDE.md`, `docs/MAP.md` §2–§3 y la especificación vigente: `docs/contrato-plantillas.md` (desde M1)
o, mientras no exista, `docs/integracion-creative-ink.md` §4–§5. Si discrepan con esta skill, ganan ellos.

## Contexto del proyecto (verifícalo antes de fiarte: puede haber cambiado)
- Esquemas: `schema/template.schema.json`, `index.schema.json`, `pack.schema.json` (⏳ M1).
- Librería: módulo `github.com/galigaribaldi/Memorandum-tex/tools`, paquetes en `tools/pkg/` (⏳ M2). Solo biblioteca
  estándar; `go` ≤ la versión del `go.mod` de Creative-Ink (hoy 1.26).
- Consumidor: Creative-Ink (`/Users/hcabrera/Documents/Personal_Documents/Creative-Ink`). Relevante allí:
  `creative-ink/internal/core/templates.go` (límites 2000 archivos y 200 MB, `{{TITULO}}`, `.creative-ink-meta`),
  `creative-ink/internal/ci/registry.go` (comandos `ci`), `docs/MAP.md` §4.6 (configuración por capas y confianza).

## Principios
1. **Una sola implementación del contrato:** la validación, los parámetros y la extracción viven en `tools/pkg/*` y
   Creative-Ink los importa. No dupliques reglas en otro lado.
2. **Cada cambio de contrato es un solo cambio:** especificación + esquema + tipos Go + fixtures + tests +
   `docs/integracion-creative-ink.md` + MAP §3, en el mismo commit.
3. **Compatibilidad hacia adelante:** los campos nuevos son opcionales y los clientes ignoran los desconocidos. Un
   cambio incompatible sube `schemaVersion` y se publica en `catalog/v<n>/` **sin retirar** el anterior.
4. **Resultados frente a errores:** que una plantilla sea incompatible es un resultado (`compat.Status`). Un zip
   corrupto, un checksum distinto o una ruta insegura son errores tipados.
5. **Seguridad por defecto:** nada descargado se ejecuta ni se confía; la extracción rechaza lo que el validador
   rechaza al publicar (defensa en profundidad).

## Pasos para añadir o cambiar un campo
1. Decide si es compatible (§Principios 3). Si no lo es, pide decisión al humano antes de seguir.
2. Actualiza la especificación y el ejemplo.
3. Actualiza el JSON Schema y los tipos de `pkg/contract` (con etiquetas `json` en camelCase iguales al esquema).
4. Añade o ajusta el fixture y su test (caso válido y caso inválido).
5. Actualiza `docs/integracion-creative-ink.md` (requerimiento afectado y bitácora con fecha) y la tabla del MAP §3.

## Verificación (sin efectos externos; pide permiso para ejecutarla)
```bash
cd tools && gofmt -l . && go vet ./... && go test -race ./...
cd tools && GOOS=windows go vet ./... && GOOS=linux go vet ./... && GOOS=darwin go vet ./...
cd tools && go run ./cmd/inkt validate ../templates                     # desde M2
```
Los fixtures **no** se validan con `inkt validate`: `futura` viola V02 a propósito y los casos inválidos se construyen en
tiempo de test (`tests/fixtures/README.md`). Los recorre `go test`.
`gofmt -l` debe salir vacío. Si algo no se pudo verificar (por ejemplo, la descarga real desde GitHub), dilo en el
informe.

## Gotchas
- **`path` frente a `path/filepath`:** dentro del zip y en el manifiesto las rutas usan siempre `/` (`path`); en disco,
  `filepath`. Mezclarlas rompe en Windows.
- **Zip slip:** valida cada entrada con `filepath.IsLocal` después de quitar la carpeta raíz `<id>/`; rechaza enlaces
  simbólicos por el modo del `FileHeader`.
- **`sha256` antes de extraer**, sobre el archivo descargado completo, nunca mientras se extrae.
- **`EscapeTeX` de `pkg/params` debe coincidir con `core.EscapeTeX` de Creative-Ink:** mantén un test con casos
  compartidos (`\ { } $ & # ^ _ % ~`).
- **Nunca `api.github.com`:** 60 peticiones por hora sin autenticación. Solo Pages y descargas de Releases.
- **ETag:** guarda el ETag junto al índice en caché; un `304` reutiliza la caché y no es un error.
- **Los esquemas usan lookahead (`(?!…)`), que es ECMA-262 y no existe en el RE2 de Go:** las reglas de Go se
  escriben en código (§12 del contrato), nunca traduciendo el `pattern` del esquema tal cual.
- **No escribas caracteres invisibles literales en los fixtures** (BOM, espacios de ancho cero): usa el escape JSON
  (`\ufeff`). El hook de marcas los rechaza y un editor puede borrarlos sin avisar.
