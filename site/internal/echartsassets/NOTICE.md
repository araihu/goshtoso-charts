# ECharts runtime notice

The interactive catalog uses ECharts 4 because go-echarts requires it for its
WebGL, liquid-fill, and word-cloud examples. ECharts and its extensions are
vendored from the go-echarts asset distribution and licensed under Apache-2.0.
Pinned file SHA-256 values:

- `echarts@4.min.js`: `97000c70420ce0b6c7d9e450d7c9919f97f034fa8ec046ac96719c08a2bbf324`
- `echarts-gl.min.js`: `397b92806949702de3460f81c6083a0974dcea4a692f8020a53bc53970e0ff9f`
- `echarts-liquidfill.min.js`: `ba223092fb0e0d082fdcca48c919c816aaf5acbd25c8b51ff930f5ad06dd77e3`
- `echarts-wordcloud.min.js`: `9f1cf57eb566f10b18059a22e16882b59e500aef92b5686a7c388ded2260ee87`
- `maps/china.js`: `146a69f110aca347228447319216ad665fbf6a57d81c73ddc911c1167aa39249`
- `maps/guangdong.js`: `ca870acf1f735d4b8fda33bb41c0a2804c320ee0885a772e428cfdc4d66f4757`

The Go adapter is `github.com/go-echarts/go-echarts/v2`, licensed under MIT.
No interactive chart runtime is loaded from a third-party origin.
