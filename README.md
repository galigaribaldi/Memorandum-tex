# Memorandum-tex

Plantillas LaTeX en español para documentos académicos, presentaciones, infografías, correspondencia y congresos.
Están pensadas para descargarse **bajo demanda** desde [Creative-Ink](https://github.com/galigaribaldi/Creative-Ink) y
también se pueden usar a mano con cualquier distribución de TeX.

> **Estado:** en construcción. Se está definiendo el contrato v1 y la herramienta de validación y empaquetado (fases
> M0–M2). Las plantillas se migran a partir de la fase M4. El plan completo está en [docs/MAP.md](docs/MAP.md).

## Qué ofrece

- **Una plantilla por necesidad:** el usuario descarga solo lo que va a usar, ya sea una plantilla o un paquete con
  nombre (`academico`, `presentaciones`, `correspondencia`, `congresos`, `todo`).
- **Personalización antes de usar:** cada plantilla declara sus parámetros (título, autor, con o sin portada, paleta de
  color…) en su manifiesto `template.json`. Creative-Ink los pide o los recibe como `clave=valor` y los escribe en
  `ink-params.tex`.
- **Varios motores:** pdfLaTeX por defecto, con XeLaTeX y LuaLaTeX opcionales. Bibliografía natbib/apalike con BibTeX,
  y biblatex con Biber donde la plantilla lo ofrezca.
- **Multiplataforma:** macOS, Linux y Windows. Sin Makefiles ni scripts de shell; la herramienta está escrita en Go.

## Catálogo previsto

| Plantilla | Para qué |
|-----------|----------|
| `ensayo` | Ensayo académico, con o sin portada institucional |
| `articulo` | Artículo académico, con o sin portada y palabras clave |
| `registro-avances` | Registro de avances con diff entre versiones |
| `tesis` | Tesis de posgrado con un kit de seguimiento de cambios (issues, versiones, revisiones) |
| `presentacion-institucional` | Presentación Beamer 16:9 con paletas intercambiables |
| `presentacion-sotm-latam` | Presentación Beamer con la identidad de SotM Latam |
| `infografia` | Infografía TikZ en A4 o A3, vertical u horizontal |
| `carta-institucional` | Carta oficial membretada |
| `carta-apimetro` | Documento oficial de Apimetro |
| `isprs-abstract` | Resumen para congresos ISPRS |

## Cómo se usa

- **Desde Creative-Ink (recomendado):** desde la Terminal CI se consulta el catálogo, se descarga una plantilla o un
  paquete y se crea el proyecto con sus parámetros. Los comandos los define Creative-Ink; el plan está en
  [docs/integracion-creative-ink.md](docs/integracion-creative-ink.md).
- **A mano:** cada versión se publica como un zip en las Releases de este repositorio (`<id>-<versión>.zip`).
  Descomprímelo, edita los valores por defecto de `main.tex` o crea un `ink-params.tex`, y compila con
  `latexmk -pdf main.tex`.

## Estructura

```
templates/<categoria>/<id>/   ← una plantilla: template.json, README, CHANGELOG, preview.png, src/
shared/                       ← componentes comunes (motor, paletas, comandos, instituciones)
packs/                        ← paquetes de plantillas
schema/                       ← JSON Schema del contrato
tools/                        ← herramienta inkt y librería Go que importa Creative-Ink
docs/                         ← mapa de arquitectura, contrato y guías
```

## Contribuir

Agregar una plantilla no requiere tocar nada central: ver [CONTRIBUTING.md](CONTRIBUTING.md).

## Licencia

Software libre, con dos licencias según el contenido (detalle en [NOTICE.md](NOTICE.md)):

- **Plantillas y componentes LaTeX:** [MIT-0](LICENSES/MIT-0.txt). Lo que escribas con una plantilla es tuyo y no
  tiene que conservar ningún aviso de Memorandum-tex.
- **Herramientas, librería Go, esquemas y documentación:** [MIT](LICENSE).

Los escudos y logotipos institucionales pertenecen a sus instituciones y no están cubiertos por estas licencias.
