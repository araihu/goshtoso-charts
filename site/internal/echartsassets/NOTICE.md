# ECharts runtime notice

The private interactive renderer uses ECharts 4 through the go-echarts adapter.
ECharts is vendored from the go-echarts asset distribution and licensed under
Apache-2.0. Pinned file SHA-256:

- `echarts@4.min.js`: `97000c70420ce0b6c7d9e450d7c9919f97f034fa8ec046ac96719c08a2bbf324`

The Go adapter is `github.com/go-echarts/go-echarts/v2`, licensed under MIT.
No interactive chart runtime is loaded from a third-party origin.
