# ok-minima (fixture)

Plantilla mínima **válida** del contrato v1. La usan los tests de `tools/` y de Creative-Ink como caso base: tiene un
parámetro de cada tipo y declara los cuatro modos.

| Clave | Macro | Tipo | Por defecto |
|-------|-------|------|-------------|
| `titulo` | `\TituloDoc` | string (máx. 120, nombre del proyecto) | — |
| `autor` | `\Autor` | string, obligatorio | — |
| `nota` | `\Nota` | string, opcional | — |
| `portada` | `\ConPortada` | boolean | `true` |
| `color` | `\ColorTema` | enum `Azul`, `Verde`, `Gris` | `Azul` |
| `fecha` | `\FechaEntrega` | date | `today` |
| `comite` | `\ComiteTutor` | list (máx. 5) | `[]` |

Modos: `completo`, `rapido`, `borrador`, `capitulo` (`secciones/introduccion`).
