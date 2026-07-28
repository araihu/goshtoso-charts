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
