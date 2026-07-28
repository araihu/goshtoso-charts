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
