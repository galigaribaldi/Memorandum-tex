# Licencias de Memorandum-tex

Este repositorio usa dos licencias libres, según el tipo de contenido. Los textos completos están en `LICENSES/`
(convención [REUSE](https://reuse.software)).

| Contenido | Licencia | Qué significa para ti |
|-----------|----------|-----------------------|
| Herramientas y librería Go (`tools/`), esquemas (`schema/`), documentación (`docs/`, `README.md`, `CONTRIBUTING.md`…), fixtures que no son plantillas | **MIT** (`LICENSE`, `LICENSES/MIT.txt`) | Puedes usarlos, modificarlos y redistribuirlos, incluso en software privativo, conservando el aviso de copyright y de licencia |
| Plantillas (`templates/**/src/`), componentes compartidos (`shared/`, salvo los logos) y plantillas de prueba (`tools/fixtures/data/templates/**/src/`) | **MIT-0** (`LICENSES/MIT-0.txt`) | Igual que MIT, pero **sin obligación de conservar ningún aviso**: el documento que escribas con una plantilla (tesis, artículo, carta) es tuyo y no tiene que mencionar Memorandum-tex |

## Excluido de estas licencias

- **Escudos y logotipos institucionales** (`shared/instituciones/*/logos/`, por ejemplo los de la UNAM y la FES Acatlán)
  son propiedad de sus instituciones. Se incluyen solo para identificar documentos de esas instituciones, conforme a
  sus normas de uso. Ni MIT ni MIT-0 conceden derechos sobre ellos.
- **Material de terceros dentro de una plantilla** (por ejemplo, la clase oficial de un congreso) conserva la
  licencia y el copyright de sus autores. Cada plantilla que lo incluya lo declara en su `README.md`, en el campo
  `license` de su manifiesto (expresión SPDX, p. ej. `MIT-0 AND LicenseRef-isprs`) y en el `LICENSE` de su bundle.

## Contribuciones

Al contribuir aceptas que tu aporte se publique bajo la licencia que corresponde a su tipo de contenido según la tabla
anterior. Conservas el copyright de lo que escribas.
