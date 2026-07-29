# Upstream example coverage

This is the central source ledger for examples adapted into Goshtoso Charts.
Component pages stay renderer-neutral; backing-library attribution and immutable
source evidence live here and on the Attributions page.

## Static/vector Line

- Source repository: `github.com/go-analyze/charts`
- Revision: `1fe31b06b8a82e00df877ff4417a75858547c1c2`
- Status: all ten Line examples present at that revision are covered by the one
  renderer-neutral `line.Line` component.

| Upstream example | SHA-256 | Goshtoso coverage | Adaptation note |
| --- | --- | --- | --- |
| `examples/1-Painter/line_chart-1-basic/main.go` | `b34cb5bbd51901dfc5d679bb8e87d9722b9be68ea664cb8a20587e0048027f51` | Basic presentation and missing observations | Preserves the five series and explicit Email gap. |
| `examples/1-Painter/line_chart-2-symbols/main.go` | `c294a10629249a274b740fc32bd6d34225822dcda093277f5cb211ca1c5d895c` | Per-series symbols | Corrects the upstream legend mismatch by listing the four rendered series. |
| `examples/1-Painter/line_chart-3-smooth/main.go` | `3e0c8469d7240fc321118827c39814ac1ba6f86cca6307a381d573f2fc6a68da` | Smoothed strokes | Preserves the four series, hidden legend and symbols, wide stroke, and smoothing tension. |
| `examples/1-Painter/line_chart-4-mark/main.go` | `40ed4e702b1f95767702d6601d3ec3c5576099cb53e77d1bb33b0894927f5904` | Statistical references | Average lines and maximum points gain adjacent exact reference evidence. |
| `examples/1-Painter/line_chart-5-area/main.go` | `b2d7b87ff675f437dbc95f2d7a0447c2040e18c5b873256a5808987dfc6131d0` | Filled area treatment | Existing renderer-neutral area option retained. |
| `examples/1-Painter/line_chart-6-stacked/main.go` | `2c021ff3780483d05d48b9e6423bf4cb9cd08c50fcbc537ace40c2c4d286e09e` | Stacked contributions | Preserves A/B/C data, fixed-width decimal labels, maximum points, and cumulative-axis typography. The upstream categorical Y-tick labels are omitted as a type-invalid defect. |
| `examples/1-Painter/line_chart-7-boundary_gap/main.go` | `d6b3b033dd8cf7edbb822c0ec40bc791d90e2b4b8213fe2cb3019c60ebdf86ed` | Boundary-gap comparison | Renders the same dataset twice with explicit gap on and off. |
| `examples/1-Painter/line_chart-8-dual_y_axis/main.go` | `78a3edd9aa356dc798c367b40cc5abecdb765b634795c38767f34bf266b805af` | Dual Y-axis treatment | Existing two-axis example retained. |
| `examples/1-Painter/line_chart-9-custom/main.go` | `67a9f05aba96e9708f8102e57086c6089bfeadf081df2ec4c4d895dec9cb0ef4` | Dense axes, gaps, and positioned labels | Preserves the 60–510 mm lens domain, four sorted lens series, gaps, sampled axes, and positioned labels through typed options. |
| `examples/1-Painter/line_chart-10-gradient_labels/main.go` | `21d84540f36ecdfb77acd0feec7101606d3f3a8f3517d3a4e3a710b55b090808` | Theme-aware value-label scale | Preserves the values and cold-to-warm intent using chart-theme tokens instead of fixed colors. |

No Line example at this revision requires a raw renderer option or engine-specific
public type. All ten are represented by typed renderer-neutral configuration.
