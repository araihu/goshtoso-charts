# Interactive runtime notice

Goshtoso Charts embeds Apache ECharts 5.4.3 for interactive charts. The local
file comes from the `github.com/go-echarts/go-echarts/v2` v2.7.2 asset
distribution and is licensed under Apache-2.0.

- Local path: `js/runtime/echarts/5.4.3/echarts.min.js`
- Local SHA-256: `987554a0014ad7be585eccc91c4329d050b40c2c0ebd2e8ec84adca82c0eb843`
- Local SHA-384 SRI: `sha384-PpgrXpnquO1pL7h9CGCAR6K7B62r8ZFzdHKzC/FtQ/0Pm4MLQNTcS2yt12JX+rkb`
- Upstream: `https://github.com/apache/echarts/tree/5.4.3`
- License: `https://github.com/apache/echarts/blob/5.4.3/LICENSE`

The explicit CDN option uses the pinned jsDelivr npm URL declared in
`assets.RuntimeCDNURL`. Its bytes differ from the embedded file only by CRLF
line endings, so it has a distinct SRI value declared in
`assets.RuntimeCDNIntegrity`.

## Word-cloud extension

Goshtoso Charts also embeds `echarts-wordcloud` 2.1.0 for interactive word
clouds. Version 2.x supports the bundled ECharts 5.x runtime. The npm package
declares the ISC license; its bundled `wordcloud2.js` code carries an MIT
license notice.

- Local path: `js/runtime/word-cloud/2.1.0/runtime.min.js`
- Local SHA-256: `4bda7da093a269a48f3d5541ebe0a2843cfed56a284f3039caa551d854f3068b`
- Local SHA-384 SRI: `sha384-LlxaHZfP53fZT+lrIPNI4Mpi7tiscM7orbp47yM6l26/RQMcaB9HbAkoldIpQ6Ws`
- CDN SHA-256: `7b6f0d55971d9de5913120c7ce6342f3551efd00b4a1df8a50f08385bb25f155`
- CDN SHA-384 SRI: `sha384-U1KEY0DDCF4Dq6Yx1J+EZ5Hnj8X5bMn52OAcJB8C4OiAWeU4iJhJ/Tv5KhTqu8zZ`
- Release source: `https://github.com/ecomfe/echarts-wordcloud/tree/2.1.0`
- Published package: `https://www.npmjs.com/package/echarts-wordcloud/v/2.1.0`
- Bundled MIT notice: `https://github.com/ecomfe/echarts-wordcloud/blob/2.1.0/dist/echarts-wordcloud.min.js.LICENSE.txt`

## Liquid-gauge extension

Goshtoso Charts embeds `echarts-liquidfill` 3.1.0 for the liquid Gauge
treatment. Version 3 declares compatibility with ECharts 5; its peer range is
`^5.0.1`, which includes the bundled 5.4.3 runtime. Package metadata declares
MIT while the distributed `license.md` contains BSD-3-Clause terms. The
embedded distribution is therefore attributed under BSD-3-Clause.

- Local path: `js/runtime/liquid/3.1.0/runtime.min.js`
- Bundled license: `js/runtime/liquid/3.1.0/LICENSE.md`
- License SHA-256: `7fdf029806a89db319a8f2a68b3cac9f36fc798b3207037152a38a2031fa4d05`
- Local/CDN SHA-256: `7925141a342a2e92fb6223d94a47d8d39534cb40d70f1b28410e9933a8ac840b`
- Local/CDN SHA-384 SRI: `sha384-+LS91q88WjMob2zpAaAPTyASiqV4HPo9zJHsEwjcukMZevj//sFrxBXdAHe1t2CL`
- Release source: `https://github.com/ecomfe/echarts-liquidfill/tree/v3.1.0`
- Published package: `https://www.npmjs.com/package/echarts-liquidfill/v/3.1.0`
- Distributed license: `https://github.com/ecomfe/echarts-liquidfill/blob/v3.1.0/license.md`

## Geographic resources

Goshtoso Charts embeds Brazil-state and São Paulo-municipality geometry
registration scripts derived mechanically from the official IBGE Malha
Municipal Digital 2025 Shapefiles. IBGE Nota Metodológica 01/2026 makes the
publication available under CC BY 4.0. Processing used mapshaper 0.6.113 to
clean, simplify while keeping shapes, remove national detached fragments below
0.1 km², reduce coordinate precision to four decimals, and convert to GeoJSON;
the runtime assets retain official IBGE names, UF codes, and geography identifiers.

- Local paths: `js/maps/ibge-mmd-2025/brazil.js`, `js/maps/ibge-mmd-2025/sao-paulo.js`
- Bundled reuse notice: `js/maps/ibge-mmd-2025/LICENSE.md`
- Revision: IBGE Malha Municipal Digital 2025; Nota Metodológica 01/2026
- Publication page: `https://www.ibge.gov.br/geociencias/organizacao-do-territorio/estrutura-territorial/15761-areas-dos-municipios.html?lang=pt-BR`
- Legal note: `https://biblioteca.ibge.gov.br/visualizacao/livros/liv102268.pdf`
- Legal-note SHA-256: `bc88c44624c8852b4796b72b525f9092cd770415403c1a3486a8d48bcefc9e89`
- National source: `https://geoftp.ibge.gov.br/organizacao_do_territorio/malhas_territoriais/malhas_municipais/municipio_2025/Brasil/BR_UF_2025.zip`
- National source SHA-256: `cdbbf05f79153802cbfa74d0c29814cd76a9c0b925aea910c9f04dffc28e6167`
- São Paulo source: `https://geoftp.ibge.gov.br/organizacao_do_territorio/malhas_territoriais/malhas_municipais/municipio_2025/UFs/SP/SP_Municipios_2025.zip`
- São Paulo source SHA-256: `3ea8041f69e10e68045ff1275616f01673d94019d8105a00212354b80a067c3c`
- Derived SHA-256 (`brazil.js`): `1b3719c82f6e2278a3e6ea8b7fc2e195460ee6a7de1546d0a8e05e6d0174bb3d`
- Derived SHA-256 (`sao-paulo.js`): `657dee960c4c4d991f5b0e6d59681152d5e2b9c48091e5094085a666c97ff317`
- License: CC-BY-4.0, `https://creativecommons.org/licenses/by/4.0/`
- CDN policy: upstream publishes GeoJSON, not executable registration scripts;
  these two pinned local assets remain local even when `WithCDN` moves supported
  runtime packages to their pinned CDN distributions.

## Three-dimensional chart extension

Goshtoso Charts embeds the official `echarts-gl` 2.0.9 npm distribution for
three-dimensional interactive charts. Its package declares peer compatibility
with `echarts ^5.1.2`, which includes the bundled ECharts 5.4.3 runtime.

- Local path: `js/runtime/three-d/2.0.9/runtime.min.js`
- Bundled license: `js/runtime/three-d/2.0.9/LICENSE`
- Local/CDN SHA-256: `bfba1b87b8c3c06e5c7ed7741002586c747b00e4efdaa92077d15c2dc721bda0`
- Local/CDN SHA-384 SRI: `sha384-f4gAUkb5Y6LE9n50CbiH1hCBCw7021OeJu0ZrgRpgW6G1CZjPR8cu33e8rCFLqCl`
- npm tarball SHA-256: `319e0520d0b3fbebcb43a4cff1c19cab38806d1d24d42c8dc2e4349afec9953e`
- npm tarball integrity: `sha512-oKeMdkkkpJGWOzjgZUsF41DOh6cMsyrGGXimbjK2l6Xeq/dBQu4ShG2w2Dzrs/1bD27b2pLTGSaUzouY191gzA==`
- Published package: `https://www.npmjs.com/package/echarts-gl/v/2.0.9`
- Package git head: `a3cb1c6bf0f64bed9c8ca144096689defb8e1ce3`
- License: BSD-3-Clause
