# IBGE Malha Municipal Digital 2025 reuse notice

The bundled Brazil-state and São Paulo-municipality geometry is a mechanically
adapted form of the Instituto Brasileiro de Geografia e Estatística (IBGE)
Malha Municipal Digital 2025.

IBGE's Nota Metodológica 01/2026 states that the publication is available under
the Creative Commons Attribution 4.0 International license (CC BY 4.0). The
license permits sharing and adaptation, including commercial use, provided
appropriate attribution is given, a link to the license is supplied, and
changes are identified.

- Author: Instituto Brasileiro de Geografia e Estatística (IBGE)
- Work: Malha Municipal Digital 2025
- Publication: Nota Metodológica 01/2026, catalog ID 2102268
- Legal note: https://biblioteca.ibge.gov.br/visualizacao/livros/liv102268.pdf
- License: https://creativecommons.org/licenses/by/4.0/
- Changes: converted from Shapefile to GeoJSON, cleaned, coordinate precision
  reduced to four decimals, and simplified with mapshaper 0.6.113 while keeping
  shapes; detached national fragments below 0.1 km² were removed and source
  identity properties were reduced to the fields used at runtime.

This notice is not a substitute for the license or the IBGE legal note.
