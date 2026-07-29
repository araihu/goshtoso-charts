# Upstream example coverage

This is the central source ledger for examples adapted into Goshtoso Charts.
Component pages stay renderer-neutral; backing-library attribution and immutable
source evidence live here and on the Attributions page.

## Static/vector Bar

- Source repository: `github.com/go-analyze/charts`
- Revision: `1fe31b06b8a82e00df877ff4417a75858547c1c2`
- Status: all eleven dedicated Bar-family files at that revision are covered by
  the one renderer-neutral `bar.Bar` component. They contain nine distinct
  visual behaviors; the two option-function examples repeat the corresponding
  basic vertical and horizontal treatments through a second upstream API style.

| Upstream example | SHA-256 | Goshtoso coverage | Adaptation note |
| --- | --- | --- | --- |
| `examples/1-Painter/bar_chart-1-basic/main.go` | `30f03d99e6f1394096d18ebed5edc04210096945c163719ce38fdd4c97368383` | Basic vertical comparison | Preserves twelve monthly rainfall and evaporation values, right-side overlay legend, title, and 600x400 geometry. |
| `examples/1-Painter/bar_chart-2-size_margin/main.go` | `fc0d426db8e09dc2032bd0c8f6bf67dbcdd66668bac0547821e4094006a55c7b` | Vertical thickness and group gap comparison | Preserves the direct automatic, 15%-thickness, and zero-gap comparison over the first six months. |
| `examples/1-Painter/bar_chart-3-label_position-round_caps/main.go` | `b29d5387ab867885d627c565ce11c2284dd046272a43a86dabb83549363696ce` | Rounded caps and start/end value labels | Physical top/bottom positions become orientation-neutral value-end and category-axis-start anchors. |
| `examples/1-Painter/bar_chart-4-mark/main.go` | `544fea22c29db4225c7b10bb6d12137d484a4ca9b6c647dc29730a61ce4ced4c` | Statistical references | Preserves average lines and minimum/maximum points with adjacent exact evidence. |
| `examples/1-Painter/bar_chart-5-stacked/main.go` | `d35e6866b6e4d4071d3e09e26db77165fbd9aecefecf0e6e880bb35752e23119` | Stacked totals and references | Preserves contributions, maximum line, global maximum point with `Sum:` prefix, point size, padding, and right-side overlay legend. |
| `examples/1-Painter/horizontal_bar_chart-1-basic/main.go` | `735240dd8433bd2494ae019f272840a8ff2fcf5572166b78269e23cbff7111a0` | Basic horizontal comparison | Preserves the seven categories, two reporting series, title, padding, and 600x400 geometry. |
| `examples/1-Painter/horizontal_bar_chart-2-size_margin/main.go` | `34d9f682f5168830cfac75a4f35a4bcc2e8216ea25024142be892bc7784597da` | Horizontal thickness and group gap comparison | Preserves automatic, 15%-thickness, and zero-gap treatments. The upstream type-invalid month labels on the numeric value axis are corrected to categorical labels. |
| `examples/1-Painter/horizontal_bar_chart-3-mark/main.go` | `c2bd6eaf3f47d8bce333186aa0212204fe3c6ff67cc66236fc6fe1040d180cb4` | Horizontal maximum reference lines | Preserves maximum lines on both reporting series. |
| `examples/1-Painter/horizontal_bar_chart-4-stacked/main.go` | `82ff73196a355aeda6ddb12b4e6d6cfc2d191dc32dfcd6e1830eb69a321b7b24` | Stacked horizontal labels | Preserves stacking, visible exact segment labels, hidden numeric axis, title, and padding. |
| `examples/2-OptionFunc/bar_chart-1-basic/main.go` | `eeff9689b6279ecbbdf9475a8034e8f57aa582c29e8d74db74f34cf78b385ca8` | Basic vertical and statistical examples | Same dataset and reference semantics as the Painter basic and mark examples; no renderer-specific option-function API enters the public component. |
| `examples/2-OptionFunc/horizontal_bar_chart-1-basic/main.go` | `84268867a32a2a4ff81cf29d2a56a48174d7c1d46c5edc8ac20a82c07e13fea4` | Basic horizontal example | Same dataset and geometry as the Painter horizontal basic example; covered once in the page. |

No dedicated Bar example at this revision requires a raw renderer option or
engine-specific public type. Generic multi-chart and web examples exercise page
composition rather than additional Bar-family behavior and remain outside this
component coverage count.

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

## Interactive Line

- Source repository: `github.com/go-echarts/examples`
- Source file: `examples/line.go`
- Revision: `bda428480a82d6d77ebb9fa939cf8d52528453dd`
- Source SHA-256: `1f36444bd373eafde876af19746d6b0115a776fd7c019e5996bdf2d00ecd7b1c`
- Status: thirteen of fourteen upstream functions are covered by the one
  renderer-neutral `interactive.Line` component. `lineOverlap` remains an
  explicit unsupported case pending a renderer-neutral composite-chart API.
- Deterministic adaptation: random values and ambient local time are replaced by
  fixed category and numerical values plus UTC timestamps. The upstream `Peach `
  label typo is corrected to `Peach`.

| Upstream function | Coverage | Goshtoso Charts treatment |
| --- | --- | --- |
| `lineBase` | Example | Basic categorical line |
| `lineShowLabel` | Example | Visible point labels |
| `lineMarkPoint` | Example | Typed minimum, maximum, and average point references |
| `lineSplitLine` | Example | Visible horizontal split lines |
| `lineNumerical` | Example | Numerical x axis, piecewise visual scale, guide lines, and marked range |
| `lineTime` | Example | Typed UTC time axis and exact-value evidence |
| `lineStep` | Example | Step treatment |
| `lineSmooth` | Example | Smooth treatment |
| `lineArea` | Example | Area treatment and marked range |
| `lineSmoothArea` | Example | Smooth area treatment |
| `lineOverlap` | Unsupported | Mixed line, scatter, and effect-scatter series require a renderer-neutral composite-chart API; Line does not expose backing-engine series types |
| `lineMulti` | Example | Four aligned series |
| `lineDemo` | Example | Two-series comparison with labels and average guides |
| `lineSymbols` | Example | Smoothed multi-series line with diamond symbols |

## Interactive Bar

- Source repository: `github.com/go-echarts/examples`
- Source file: `examples/bar.go`
- Revision: `bda428480a82d6d77ebb9fa939cf8d52528453dd`
- Source SHA-256: `dcda545f978fdd055ecff5a6050b2ad9dc8cf9fe350bd7e4768952e8068fc9f9`
- Status: eighteen of nineteen upstream functions are covered by the one
  renderer-neutral `interactive.Bar` component. `barOverlap` remains an
  explicit unsupported case pending a renderer-neutral composite-chart API.
- Deterministic adaptation: ambient random values are replaced by local fixed
  seeds while preserving seven categories, two series, integer values, and the
  upstream `[0,300)` value domain.

| Upstream function | Coverage | Goshtoso Charts treatment |
| --- | --- | --- |
| `barBasic` | Example | Basic two-series categorical presentation |
| `barTitle` | Example | Title and subtitle on the basic presentation |
| `barTooltip` | Example | Axis-triggered hover details on the basic presentation |
| `barSetToolbox` | Example | Shared Expand and PNG export controls plus an exact-value disclosure table |
| `barShowLabel` | Example | Visible values above each bar |
| `barXYName` | Example | Named category and value axes |
| `barXYFormatter` | Example | Literal unit suffixes through typed axis options; no executable formatter API |
| `barColor` | Example | Explicit caller colors overriding theme series tokens |
| `barSplitLine` | Example | Visible split lines on both axes |
| `barGap` | Example | 150% inter-series gap |
| `barDataZoomInside` | Example | Inside gesture zoom over the 10–50% category window |
| `barDataZoomSlider` | Example | Visible slider over the 10–50% category window |
| `barReverse` | Example | Horizontal category orientation |
| `barStack` | Example | Two series sharing one stack |
| `barMarkPoints` | Example | Named explicit point plus calculated minimum and maximum references |
| `barMarkLines` | Example | Calculated maximum and average guides |
| `barOverlap` | Unsupported | Mixed Bar, Line, and Scatter composition requires a renderer-neutral composite-chart API; Bar does not expose backing-engine series types |
| `barSize` | Example | Upstream 600-pixel height retained; fixed 1200-pixel width adapted to container width because consumers own page layout |
| `barWidth` | Example | Absolute and percentage per-series widths |

Supplementary Bar-adjacent evidence at the same revision:

| Upstream source | SHA-256 | Scope boundary |
| --- | --- | --- |
| `examples/page_center_layout.go` | `106456904719dfacfb13adcc1b9e66df83cf28a5a801539bad4d1958554166c9` | Layout reference; chart remains centered while consumers own page layout |
| `examples/page_flex_layout.go` | `3113b7bdf78a2365ae62502fe86ab001f3ff3034b1d77752c693e95b28a0fd68` | Layout reference; responsive chart width works inside a flex consumer |
| `examples/page_none_layout.go` | `ce38424de2ffeb919661e536c7f44921de098ae14643d4f2975d8e72296c32f8` | Layout reference; no page-layout mode enters the chart API |
| `examples/themes.go` | `843c478c63b9cf3ab13b1e13518ea98912332bb34caf0dae5d48343fabd121a0` | Site themes and chart tokens cover theme switching centrally |
| `examples/renderer.go` | `c4956db261f554c6a161c0d25baa7dbd7c2c179523997d297020cd55916e6a3f` | Private renderer integration, not a public chart option |
| `examples/bar3d.go` | `110b3b85f2528d76eb8271b64f1facd81a974e30ecc0dd77319d5a409ff64275` | Separate existing Bar 3D component, not a two-dimensional Bar variant |
