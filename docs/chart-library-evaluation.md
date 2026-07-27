# Chart-library evaluation

Decision: `github.com/go-analyze/charts` v0.6.0.

Required contract: Go request produces an inline SVG with no browser chart runtime, Node process, hydration, or external fetch. Output must fit inside Goshtoso surfaces and use a constrained light/dark semantic palette.

| Library | Result | Reason |
| --- | --- | --- |
| `github.com/go-analyze/charts` v0.6.0 | Chosen | Active Go-native renderer; SVG output; line, bar, pie, doughnut, scatter, heat-map and other future types; explicit color palette API. |
| `github.com/go-echarts/go-echarts/v2` v2.7.2 | Rejected for v1 | Excellent interactive chart library, but `Render` emits HTML, scripts, ECharts assets and browser-side `echarts.init`. It is not SSR SVG without a separate browser renderer. Reconsider for an explicitly interactive extension. |
| `gonum.org/v1/plot` v0.17.0 | Rejected for v1 | Mature scientific plotting and SVG backend, but adds plotting/font configuration complexity that is poorly matched to application-dashboard defaults. |
| `github.com/wcharczuk/go-chart/v2` v2.1.2 | Rejected | SVG-capable, but upstream is archived; do not start a new public dependency on it. |

Version checks performed 2026-07-27. Re-evaluate before any major renderer expansion.
