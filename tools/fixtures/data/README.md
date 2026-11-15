# Fixtures del contrato v1

Datos de prueba de `docs/contrato-plantillas.md`. Los usan los tests de `tools/` (M2) y, a través de la librería, los
de Creative-Ink. **No necesitan red.**

```
tools/fixtures/data/
├── templates/prueba/
│   ├── ok-minima/        ← plantilla válida: un parámetro de cada tipo y los cuatro modos (caso base)
│   ├── futura/           ← schemaVersion 99 + campo desconocido → incompatible (viola V02 a propósito)
│   └── con-paquetes/     ← requires.files inexistente → missing-packages
├── packs/prueba.json     ← paquete con ok-minima y con-paquetes@1.0.0
├── params/
│   ├── escape-cases.json                           ← EscapeTeX (compartido con core.EscapeTeX de Creative-Ink)
│   ├── resolve-cases.json                          ← params.Resolve: valores y errores por campo
│   └── ok-minima.<caso>.ink-params.tex             ← salida exacta de RenderInkParams (comparación byte a byte)
├── invalid/
│   ├── manifests.json    ← un caso por regla V01–V20 (JSON Merge Patch sobre ok-minima)
│   └── folders.json      ← un caso por regla V30–V39 (operaciones sobre una copia de ok-minima)
└── bundles/cases.json    ← 16 zips válidos y maliciosos que construye el test (zip slip, enlaces, .latexmkrc en
                            cualquier forma, colisión por mayúsculas, bomba…)
```

## Por qué no hay zips ni casos rotos guardados en git

- **Reproducibles:** los zips se construyen en cada test con fecha fija (`2000-01-01T00:00:00Z`), entradas ordenadas y
  Deflate. El `sha256` es el mismo en macOS, Linux y Windows, y no hay binarios que revisar en los PR.
- **El repositorio cumple sus propias reglas:** un archivo con espacios, un `.latexmkrc` o un enlace simbólico
  guardados aquí harían fallar `inkt validate` sobre el propio repositorio.
- **Fin de línea:** `.gitattributes` convierte a LF al hacer checkout, así que un caso CRLF guardado dejaría de ser
  CRLF. Por eso el CRLF, el BOM y el Latin-1 se escriben en tiempo de test.

## Catálogo de prueba (⏳ M2, decisión 12 del MAP)

No se guarda en git. `tools/pkg/catalogtest` lo construye en una carpeta temporal a partir de `templates/prueba/`,
`packs/` y `bundles/cases.json`, con el mismo código que publica el catálogo real. Lo que construye:
- el índice, los manifiestos publicados, los `versions.json`, los paquetes y los zips;
- `futura`, publicada sin validar;
- `checksum-malo`, con un hash alterado.

Devuelve un `httptest.Server` listo para usar desde los tests de este repositorio y de Creative-Ink.

## Cómo añadir un caso

1. Si es una regla nueva del contrato, primero va en `docs/contrato-plantillas.md` §12 con su código.
2. Añade el caso al JSON correspondiente con ese código y un `name` descriptivo en kebab-case.
3. El test de M2 recorre los casos: un caso sin test que lo consuma es un error del test.
