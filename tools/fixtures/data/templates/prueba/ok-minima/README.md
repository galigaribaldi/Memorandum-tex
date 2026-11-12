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

Modos: `completo`, `rapido`, `borrador`, `capitulo`. Hay dos capítulos (`secciones/introduccion` y
`secciones/desarrollo`); con `\def\SoloCapitulo{secciones/introduccion}`, «Desarrollo» no debe aparecer.

Todos los parámetros se imprimen en el cuerpo, con o sin portada, para que sus escapes lleguen siempre a TeX.

## Comprobación manual (la ejecuta el humano)

Dentro de una copia de `src/` en una carpeta temporal:

| Paso | Comando | Debe verse |
|------|---------|------------|
| 1 | `latexmk -pdf main.tex` | Portada con `\maketitle`; «Paleta elegida: Azul»; Introducción y Desarrollo |
| 2 | `latexmk -C && latexmk -xelatex main.tex` | Lo mismo |
| 3 | `latexmk -C && latexmk -lualatex main.tex` | Lo mismo |
| 4 | Copiar `tools/fixtures/data/params/ok-minima.todos-los-tipos.ink-params.tex` como `ink-params.tex` y `latexmk -C && latexmk -pdf main.tex` | Sin portada; título «Prueba & 100% {lista}»; «Autor: Ana_María ~ #1»; «Fecha de entrega: 2026-11-10»; «Paleta elegida: Verde»; lista Dra. A, Dr. B |
| 5 | `latexmk -C && latexmk -pdf -usepretex='\def\borrador{}' main.tex` | Compila sin errores |
| 6 | `latexmk -C && latexmk -pdf -usepretex='\def\SoloCapitulo{secciones/introduccion}' main.tex` | Introducción sí, Desarrollo no |

Para conservar el registro de cada paso, añade `-jobname=paso<N>` en lugar de `latexmk -C`.
