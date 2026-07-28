# Chart-library evaluation

Decision: `github.com/go-analyze/charts` v0.6.0 for SSR primitives, plus a
separate opt-in `github.com/go-echarts/go-echarts/v2` v2.7.2 internal renderer
for typed interactive components.

Required contract: Go request produces inline SVG with no browser chart runtime, Node process, hydration, or external fetch. Output must fit inside Goshtoso surfaces, expose stable component kinds, and use Goshtoso semantic CSS tokens where primitive SVG permits it.

| Library | Result | Reason |
| --- | --- | --- |
| `github.com/go-analyze/charts` v0.6.0 | Chosen | Active Go-native renderer; SVG output; line, bar, pie, doughnut, scatter, heat-map and other future quantitative types; explicit color palette API. |
| `github.com/go-echarts/go-echarts/v2` v2.7.2 | Chosen as the private interactive renderer | `RenderSnippet` emits an element plus browser-side initialization; `components/interactive` exposes only renderer-neutral types. The demo serves a pinned local runtime, never the default CDN. |
| `gonum.org/v1/plot` v0.17.0 | Rejected for v1 | Mature scientific plotting and SVG backend, but adds plotting/font configuration complexity that is poorly matched to application-dashboard defaults. |
| `github.com/wcharczuk/go-chart/v2` v2.1.2 | Rejected | SVG-capable, but upstream is archived; do not start a new public dependency on it. |

Availability history is an application example built from the generic Bar
component rather than a chart type. The interactive example uses the public
SSE snapshot contract; monitoring states remain outside the component taxonomy.

The interactive boundary accepts only trusted, Go-owned option data. Renderer
callbacks and raw JavaScript-bearing option fields are not exposed through the
public API.

Version checks performed 2026-07-27. Re-evaluate before any major renderer expansion.
