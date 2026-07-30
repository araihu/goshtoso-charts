# Upstream example coverage

This is the central source ledger for examples adapted into Goshtoso Charts.
Component pages stay renderer-neutral; backing-library attribution and immutable
source evidence live here and on the Attributions page.

## Interactive Parallel and ThemeRiver ownership

- Source repository: `github.com/go-echarts/examples`
- Revision: `bda428480a82d6d77ebb9fa939cf8d52528453dd`
- Upstream sources: `examples/parallel.go` and `examples/themeriver.go`
- Canonical APIs: `components/interactive/parallel.Parallel` and
  `components/interactive/themeriver.ThemeRiver`
- Status: the package-ownership migration preserves the existing Parallel
  three-city multivariate dataset and ThemeRiver six-stream aligned-date
  dataset, data shape, chart semantics, and renderer-neutral visual intent.
  It changes only canonical Go ownership and the site imports, snippets, and API
  links; the parent package remains an exact alias-and-forwarder compatibility
  facade.

## Interactive Graph and Sankey ownership

- Source repository: `github.com/go-echarts/examples`
- Revision: `bda428480a82d6d77ebb9fa939cf8d52528453dd`
- Upstream sources: `examples/graph.go` and `examples/sankey.go`
- Canonical APIs: `components/interactive/graph.Graph` and
  `components/interactive/sankey.Sankey`
- Status: the package-ownership migration preserves the existing Graph and
  Sankey datasets, hierarchy and weighted relationships, chart semantics, and
  renderer-neutral visual intent. It changes only canonical Go ownership and
  the site imports, snippets, and API links; the parent package remains an
  exact alias-and-forwarder compatibility facade.

## Interactive Tree and Sunburst ownership

- Source repository: `github.com/go-echarts/examples`
- Revision: `bda428480a82d6d77ebb9fa939cf8d52528453dd`
- Upstream sources: `examples/tree.go` and `examples/sunburst.go`
- Canonical APIs: `components/interactive/tree.Tree` and
  `components/interactive/sunburst.Sunburst`
- Status: Tree preserves the upstream one-root, three-branch hierarchy,
  left-to-right layered layout, fully expanded initial depth, visible node and
  leaf labels, and collapsed third branch while correcting the repeated
  `Chield` typo to `Child`. Sunburst preserves seven parent/child pairs and
  fractional-value semantics with fixed values replacing ambient randomness.
  The ownership migration changes only canonical Go ownership and site imports,
  snippets, and API links; the parent package remains an exact
  alias-and-forwarder compatibility facade.

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

## Static/vector Pie

- Source repository: `github.com/go-analyze/charts`
- Revision: `1fe31b06b8a82e00df877ff4417a75858547c1c2`
- Status: all seven dedicated Pie-family files at that revision are covered by
  the one renderer-neutral `pie.Pie` component. They contain seven distinct visual treatments:
  basic Pie, area-scaled slice radii, Pie segment spacing, basic doughnut,
  outside doughnut labels, inside doughnut labels, and a doughnut center total
  with an overlay legend. The two option-function files duplicate the
  corresponding basic treatments through another upstream API style.

| Upstream example | SHA-256 | Goshtoso coverage | Adaptation note |
| --- | --- | --- | --- |
| `examples/1-Painter/doughnut_chart-1-basic/main.go` | `b97bca2322e90e2f03ab49aa77f683d0c58e027846b939e5a61100602dad1ebf` | Basic doughnut presentation | Preserves values, names, title and subtitle, 20-pixel padding, vertical bottom legend at 80%, 600x400 geometry, and a 60%-of-ring center opening. |
| `examples/1-Painter/doughnut_chart-2-styles/main.go` | `5816db5dd035c8607b2929779353c32d2bca78ed5f6244b3fc04e65292ac3610` | Outside labels, inside labels, and center total | Preserves the separate style dataset and all three outputs: 24-pixel gaps with outside labels, labels inside an enlarged 80%-of-ring opening, and hidden slice labels plus `Total Response: 3.15k`, 8-pixel gaps, and a bottom overlay legend. The fixed dark-gray center text follows theme text tokens instead. |
| `examples/1-Painter/pie_chart-1-basic/main.go` | `06183e92e75445d89917af5dfd318c8b45f624c4efa6565b626a6aff6b3b128f` | Basic pie presentation | Preserves five channel values, title and subtitle, padding, legend placement, and 600x400 geometry. |
| `examples/1-Painter/pie_chart-2-series_radius/main.go` | `54d85c6420a5e8f4fca7691c4969be80cc6bc52f8d4f10cbe5e499715875cbf6` | Area-scaled slice radii | Preserves the 120-pixel maximum and square-root scaling, yielding radii 120, 101, 90, 82, and 65 while retaining proportional sector angles. |
| `examples/1-Painter/pie_chart-3-gap/main.go` | `2392d1fd1a7644158626a261344e79b18bef2c3d802fa1cea8c3add413b980f6` | Segment gap and hidden legend | Preserves the five values, centered title, hidden legend, 16-pixel separation, and 600x400 geometry. |
| `examples/2-OptionFunc/doughnut_chart-1-basic/main.go` | `1936ff4508d6ef3967185e4076804bf53dc0bf8c64a254a569081fb1d399b453` | Duplicate basic doughnut | Same data and presentation as the Painter basic doughnut; covered once without exposing upstream option functions. |
| `examples/2-OptionFunc/pie_chart-1-basic/main.go` | `d09222d5febf104f07a81e05a4235d96004b61e5c032dd3513a501a840bbe9b7` | Duplicate basic pie | Same data and presentation as the Painter basic Pie; covered once without exposing upstream option functions. |

Unsupported dedicated Pie-family behaviors: none. Every behavior in the seven
dedicated files maps to typed renderer-neutral configuration. No dedicated Pie-family source defines statistical references, so Pie does not invent
cross-family reference lines or points.

Supplementary Pie-adjacent evidence at the same revision:

| Upstream source | SHA-256 | Scope boundary |
| --- | --- | --- |
| `examples/2-OptionFunc/web-1/main.go` | `96f110afd2d34cb3b823f3b36ecc4a48692bad91991f4010eb09322b747d20a1` | Generic web catalog and cross-family composition; its Pie snippets add no dedicated Pie behavior or renderer-neutral API requirement. |
| `examples/demo/themes/main.go` | `10e7fc1b80f2b8d151a957f2cb1b4273f7822788282e6bbb104d137fda2d621b` | Multi-family theme demonstration; shared Goshtoso light/dark theme tokens provide the equivalent behavior centrally. |

Raw renderer options, callbacks, and cross-family composition remain outside
the public Pie API. Supporting such inputs would leak engine-specific types or
turn one chart component into a composite-chart API.

## Static/vector Scatter

- Source repository: `github.com/go-analyze/charts`
- Revision: `1fe31b06b8a82e00df877ff4417a75858547c1c2`
- Status: all five dedicated Scatter-family files at that revision are covered
  by the one renderer-neutral `scatter.Scatter` component. They define basic
  aligned observations with a missing value, per-series symbols, dense repeated
  samples with statistical guides, top-N labels, and a second basic treatment
  with hollow circles and integer formatting. The option-function source reuses
  the Painter basic dataset but retains its distinct visual choices.

| Upstream example | SHA-256 | Goshtoso coverage | Adaptation note |
| --- | --- | --- | --- |
| `examples/1-Painter/scatter_chart-1-basic/main.go` | `6bd838c49fc38d6b50be1b2c26e1845348de6a5bce3a4a7e637497b78ad61818` | Basic categorical scatter | Preserves five named Monday-through-Sunday series, the missing Email value on Thursday, title and 16-pixel title text, 100-pixel left legend padding, dot size 4, and 600x400 geometry. |
| `examples/1-Painter/scatter_chart-2-symbols/main.go` | `2667f6f260c63d56dcc22cb036b6b0408ea9da0f943757909d1436be7b9ad515` | Per-series symbols | Preserves four series and their circle, diamond, square, and filled-dot markers at size 4. |
| `examples/1-Painter/scatter_chart-3-dense_data/main.go` | `0a50b43ccad6a96b3248d3e45e83add46e33b8b6ff98133e1f2597bdd46f49bb` | Dense repeated samples and statistical guides | Preserves three 1,000-category bounded random walks, additional samples every second and tenth category, compact markers, SMA(100), maximum lines for the first two series, humanized values, title, right-side vertical legend, axis counts, rotation, fonts, bounds, units, padding, and 600x400 geometry. Ambient global randomness becomes a fixed local seed while retaining the generation algorithm and value domain. The fixed vivid-light paint becomes Goshtoso theme tokens for both site themes and light/dark modes. |
| `examples/1-Painter/scatter_chart-4-top_n_labels/main.go` | `cf92798819fbc010f44eaa406acabd337f16b52eec00793a10679e9c3b7cda81` | Top-five labels | Preserves all thirty daily visitor values, exactly five highest labels with stable input-order tie handling, red-semantic emphasis, hidden legend, 0–50 axis, 20-pixel padding, title and subtitle, and 800x500 geometry. An adjacent disclosure retains every exact value and selection state. |
| `examples/2-OptionFunc/scatter_chart-1-basic/main.go` | `a4528b8943edac99ab99f1632d328a34c64013e7551e3c13f61d1aa45844afd1` | Hollow-circle and integer-format basic treatment | Preserves the Painter basic dataset, missing observation, title, legend padding, global circle symbol, and whole-number formatting through typed values instead of renderer callbacks or option functions. |

Unsupported dedicated Scatter-family behaviors: none. Every visual behavior in
the five dedicated files maps to typed renderer-neutral configuration. Generic
`multiple_charts-1` composes Scatter, Bar, and Line into one output surface; that
is consumer layout and cross-family composition, not another Scatter behavior.
It remains outside the Scatter component instead of introducing a raw painter or
composite renderer type.

## Static/vector Radar

- Source repository: `github.com/go-analyze/charts`
- Revision: `1fe31b06b8a82e00df877ff4417a75858547c1c2`
- Status: both dedicated Radar-family files at that revision are covered by
  the one renderer-neutral `radar.Radar` component. They contain one distinct visual
  treatment: the option-function source repeats the Painter budget
  comparison through a second upstream API style.

| Upstream example | SHA-256 | Goshtoso coverage | Adaptation note |
| --- | --- | --- | --- |
| `examples/1-Painter/radar_chart-1-basic/main.go` | `0cf8dbdd72f6a398b7c560b544a0d800570d17a620549753b726171f996254d4` | Basic bounded budget comparison | Preserves all six ordered dimensions and maxima, both six-value profiles, `Basic Radar Chart` title at 16 points, right-aligned legend, and 600x400 geometry. |
| `examples/2-OptionFunc/radar_chart-1-basic/main.go` | `39a9427d6bb3bcff7d7627943210e2b19b934785c03896e9beffb1a35462c78e` | Duplicate basic treatment | Preserves the same data and presentation once, without exposing upstream option functions. |

Relevant Radar implementation and shared presentation API evidence at the same
revision is pinned separately from the two dedicated example files:

| Upstream API source | SHA-256 | Renderer-neutral coverage |
| --- | --- | --- |
| `radar_chart.go` | `50f9e29787665a03ab744be3081b9582cbb2d4064025245818665923849e79d7` | Indicator names, minima, maxima and label fonts; radius; chart-level value formatting. |
| `series.go` | `953f4e5d555701348ebcb8eb0bfe1753a6df56eb4f94a86403c6dc6cecf79217` | Named aligned vectors. |
| `series_label.go` | `d7b176bea3679542e878c4c5703db3711d1e258b64efb7d19614a16fc5722611` | Per-series label visibility, font size, and value-format overrides. |
| `title.go` | `e85f6a0fe2e8fd7c253ac226d780164beb0cc7214e94979241d1e9fbce824b26` | Title, subtitle, logical horizontal or vertical placement, font sizes, visibility and border width. |
| `legend.go` | `eaad1144ff1c5af84049a2968e952caba2dac3970558d542c18c6a98ba6d515b` | Visibility, flow, logical placement, alignment, font size, padding, overlay and border width. |
| `chart_option.go` | `0b298fcd45fab6bbe476514d90e5107d65d32bd8ba985f2411e79ec88fd2b858` | Option-function parity remains behind the same typed component; chart padding and output dimensions are public renderer-neutral values. |

Unsupported dedicated Radar-family behaviors: none. Every behavior in both
dedicated files maps to typed renderer-neutral configuration. The public API
also covers the finite Radar-specific presentation fields above. Theme objects,
arbitrary formatter callbacks, painter hierarchy, output encoders, and upstream
option functions remain private; typed value formats, Goshtoso theme tokens,
shared controls, and SVG or PNG export provide the supported equivalents.

## Static/vector Candlestick

- Source repository: `github.com/go-analyze/charts`
- Revision: `1fe31b06b8a82e00df877ff4417a75858547c1c2`
- Status: all six dedicated Candlestick-family files at that revision are
  covered by the one renderer-neutral `candlestick.Candlestick` component.
  Their eight source functions define basic seven- and eight-period charts,
  three aligned series with filled, traditional, and outline bodies, period-five
  Bollinger bands, four pattern treatments, and fixed-window aggregation.

| Upstream example | SHA-256 | Goshtoso coverage | Adaptation note |
| --- | --- | --- | --- |
| `examples/1-Painter/candlestick_chart-1-basic/main.go` | `44c216955ae850b824a0e3f3ee2bbaf67a23ca185d8faea77335d048cd19c26b` | Basic seven-period chart | Preserves all OHLC values, labels, title, legend, and 600x400 geometry. |
| `examples/1-Painter/candlestick_chart-2-multiple_series/main.go` | `f132f40ac3e920a891782c5ab6f80e681f8e7ee87a0a05405bad75ee161e964f` | Three aligned series and body styles | Preserves Stock A, B, and C, five periods, filled/traditional/outline bodies, unit, padding, legend, and 1000x700 geometry. Each upstream Day 4 low is above its close and therefore invalid OHLC; the lows become 104, 154, and 204, matching the valid bearish range used by the basic example. |
| `examples/1-Painter/candlestick_chart-3-bollinger_bands/main.go` | `cc3b347d5faea1a15ca22554dcc46a35beed74e49da56701659a1a7d1f000202` | Period-five Bollinger bands | Preserves all twenty observations, upper/SMA/lower close trends, title, dimensions, unit, legend, and padding. |
| `examples/1-Painter/candlestick_chart-4-patterns/main.go` | `ab5891e744bc8ec40fbead6b16af5642ea94c738369469b392ac7acf1e0055ec` | All, core, formatted-label, and bullish pattern treatments | Preserves the ten-observation pattern sequence, automatic detection vocabulary, average/minimum close guides, four geometries, legend choices, and safe first-name-plus-count formatting without exposing a callback. Typed bearish, reversal, trend, explicit-subset, and threshold controls cover the remaining finite pattern API. |
| `examples/1-Painter/candlestick_chart-5-aggregation/main.go` | `ba7d1d31fef54f792e53840d969c4a3d791309a6059b2c5997dd2e509e1cbde1` | Source and five-period aggregate comparison | Preserves fifteen one-minute candles, three five-minute windows, first-open/highest-high/lowest-low/final-close semantics, two vertically composed charts, and 1200x800 geometry. |
| `examples/2-OptionFunc/candlestick_chart-1-basic/main.go` | `aad7ab0297061baac358b63b19b15e6dca48734d2be607d11679bd284263423c` | Extended eight-period basic chart | Preserves Day 8, 18-point title, unit, padding, legend, and 800x600 geometry. Construction-style wording is omitted from the visible title because public pages remain backing-engine neutral. |

Every function in the six pinned files is inventoried below. Function hashes
cover exact bytes from each `func` declaration through its closing brace.

| Source function | SHA-256 | Role |
| --- | --- | --- |
| Painter basic `main` | `4f576ae2cde41f4d76dcd23babee4f7a14440a6c8686c892552d17375a44868b` | Seven-period dataset and basic presentation |
| Painter multiple-series `main` | `7de600489e87d2dcc88a0e56c4aa2417e97bb512d0e000bd615b407063b80c35` | Three aligned series and body styles |
| Painter Bollinger `main` | `2b56840d7ef57840f42670f2b4619cc899503822a3a2bbef51fc13aff38928fe` | Twenty-period close bands |
| Patterns `main` | `2df1c6d6e980195d50db54e3d1410f977b21b60cb9b8316bd17d3222bb9e53d9` | Ten-observation pattern dataset |
| Patterns `createPatternExamples` | `191c4a25a756552c34775b854b902820934dd0c8f6e227e916f489041b168f4d` | Four pattern configurations |
| Patterns `generateExampleCharts` | `7209718f4bfd6a1f7650f049251534fcd17c296dab039b82e88f21406d9fd9f8` | Shared title, axis, padding, legend, and output presentation |
| Aggregation `main` | `d79715bb8b1da24b8b0caa9cd9e3a69f089b30cb47d6a15e16462e4cab0d6ed0` | Fifteen-to-three window aggregation and vertical composition |
| Option-function basic `main` | `2fb069e158b4c28cd1a4ace193380557aadaab443338ca7f21be76f68aaab128` | Eight-period dataset and explicit 800x600 presentation |

Relevant Candlestick-specific implementation and finite presentation API
evidence at the same revision is pinned separately:

| Upstream API source | SHA-256 | Renderer-neutral coverage |
| --- | --- | --- |
| `candlestick_chart.go` | `d70ee4b46e6d95de928e83d48ffcc43d3ea5516bbfc431131d1fd5bc61ab667d` | Body-width ratio, wick visibility and thickness, inter-series gap, aligned series, axes, title, legend, and padding. |
| `candlestick_patterns.go` | `c9579221d8b97477d878baa898d22ae7c60d94e4458218565c762dc5a75d8092` | All finite pattern groups, fourteen typed patterns, explicit subsets, label precedence, doji/shadow/engulfing thresholds, and safe built-in label formatting. |
| `series.go` | `953f4e5d555701348ebcb8eb0bfe1753a6df56eb4f94a86403c6dc6cecf79217` | Named OHLC populations, filled/traditional/outline bodies, per-series wick override, close trends and references used by dedicated examples, and fixed-window aggregation semantics. |
| `painter.go` | `f4ac102e9b21623765e2fdfe4c0910a03265bc751b9f5d019ae41e80611be959` | SVG rendering, vertical child regions for aggregation, dimensions, and output encoding remain private implementation details. |

Unsupported dedicated Candlestick-family behaviors: none. Every behavior in
the six dedicated files maps to typed renderer-neutral configuration. Arbitrary
value or pattern callbacks, raw theme or painter objects, generic option
functions, and mark/trend fields not exercised by these sources remain private;
typed formats, theme tokens, safe pattern labels, shared controls, exact tables,
and SVG or PNG export provide the supported equivalents.

## Static/vector Funnel

- Source repository: `github.com/go-analyze/charts`
- Revision: `1fe31b06b8a82e00df877ff4417a75858547c1c2`
- Status: both dedicated Funnel-family files at that revision are covered by
  the one renderer-neutral `funnel.Funnel` component. They contain one distinct
  visual treatment: the option-function source repeats the Painter seven-stage
  example through a second construction style.

| Upstream example | SHA-256 | Goshtoso coverage | Adaptation note |
| --- | --- | --- | --- |
| `examples/1-Painter/funnel_chart-1-basic/main.go` | `a54875c89cd0be43fa7c0614520a11c489562bbc72f8c577a314eb3f24f75a6d` | Basic seven-stage funnel | Preserves values 100, 80, 60, 40, 20, 10, and 2; Show-through-Cancel order; title; 100-pixel left legend padding; and 600x400 geometry. |
| `examples/2-OptionFunc/funnel_chart-1-basic/main.go` | `332ffeb340e1236170f41a3a93a46499897af75df261ed60f7ac01dd35ae4893` | Duplicate basic treatment | Preserves the same data and presentation once without exposing upstream option functions. |

Every function in the two dedicated files and the exact supplementary web
source span is inventoried below. Hashes cover exact source bytes at the pinned
revision. Filesystem-only helpers are recorded but do not count as chart
behaviors.

| Source span | Lines | SHA-256 | Role |
| --- | --- | --- | --- |
| Painter basic `writeFile` | 14-22 | `9a3e255a47b40c36123b225677519f09a62c7f0f38cd1341515d30ce687f1fc5` | Filesystem output helper; no chart behavior. |
| Painter basic `main` | 24-46 | `82bd67be63d88d6ef2312890964ed595ff9a12990ee89bc4edc85e4bc93f9892` | Seven-stage dataset, title, legend padding, and 600x400 presentation. |
| Option-function basic `writeFile` | 14-22 | `9a3e255a47b40c36123b225677519f09a62c7f0f38cd1341515d30ce687f1fc5` | Identical filesystem output helper; no chart behavior. |
| Option-function basic `main` | 24-44 | `1289983e007306c2d3b1d44136aaa35ec56691a5d05f502a3c718d6d356b69a3` | Duplicate seven-stage presentation through option functions. |
| Web `indexHandler` in `examples/2-OptionFunc/web-1/main.go` | 145-547 | `c9970d3ac7aab849623793deee832f1c1fec5569c4964e3b84b61857424b8068` | Mixed-family page composition containing the supplementary Funnel literal. |
| Funnel literal in `examples/2-OptionFunc/web-1/main.go` | 450-487 | `f2a990a30d9a89a7d982bc79522c1fb11b2611173ac54c31a19382a543b1ba45` | Five-stage Show-through-Order dataset rendered as the compact page example. |

The supplementary file has whole-file SHA-256
`96f110afd2d34cb3b823f3b36ecc4a48692bad91991f4010eb09322b747d20a1`.
Its five-stage Go chart literal is covered as the second example on the page but remains
outside the two-file dedicated coverage denominator. The later raw browser-chart
JSON comparison is not a static/vector Funnel API example and remains outside
this component.

Relevant Funnel-specific implementation and finite presentation API evidence
at the same revision is pinned separately:

| Upstream API source | SHA-256 | Renderer-neutral coverage |
| --- | --- | --- |
| `funnel_chart.go` | `d8c1d8e5ee84ae0f534da3046bd7d43973f7afbf9298e8b8d86c4200f0db6cfb` | Ordered nonnegative stages, proportional widths, title, legend, padding, and default percent labels. |
| `series.go` | `953f4e5d555701348ebcb8eb0bfe1753a6df56eb4f94a86403c6dc6cecf79217` | Named stage values and shared label options. |
| `series_label.go` | `d7b176bea3679542e878c4c5703db3711d1e258b64efb7d19614a16fc5722611` | Safe built-in name, value, name-and-value, percent, and hidden label treatments replace arbitrary formatter callbacks. |
| `title.go` | `e85f6a0fe2e8fd7c253ac226d780164beb0cc7214e94979241d1e9fbce824b26` | Visible title text exercised by both dedicated examples. |
| `legend.go` | `eaad1144ff1c5af84049a2968e952caba2dac3970558d542c18c6a98ba6d515b` | Visibility, flow, logical horizontal placement, and four-side padding use typed options. |
| `chart_option.go` | `0b298fcd45fab6bbe476514d90e5107d65d32bd8ba985f2411e79ec88fd2b858` | Option-function parity remains behind the same typed component. |

Unsupported dedicated Funnel-family behaviors: none. Every behavior in both
dedicated files maps to typed renderer-neutral configuration. Raw theme,
painter, output-encoder, generic-series, and option-function types remain
private. Theme tokens, caller stage colors or classes, shared lifecycle
controls, adjacent exact-value evidence, and SVG or PNG export provide the
supported equivalents.

## Static/vector Heat map

- Source repository: `github.com/go-analyze/charts`
- Revision: `1fe31b06b8a82e00df877ff4417a75858547c1c2`
- Status: exactly one dedicated HeatMap-family file exists at this revision.
  Its one distinct upstream treatment is covered by the renderer-neutral
  `heatmap.HeatMap` component: the exact five-by-five matrix, centered title,
  named axes, sequential value scale, and 600x400 presentation.

| Upstream example | SHA-256 | Goshtoso coverage | Adaptation note |
| --- | --- | --- | --- |
| `examples/1-Painter/heat_map-1-basic/main.go` | `c39a3d85a0df126da5d099a60e1491ae424d0768260ee738b7932288f1bf687f` | Basic five-by-five matrix | Preserves all twenty-five values, row order, column order, title, X/Y titles, and 600x400 geometry. Theme-aware cold-to-warm tokens replace the fixed single-hue rendering while retaining the upstream sequential-scale intent. |

Every function, helper, and chart-relevant subspan in the pinned file is
inventoried below. Hashes cover exact source bytes at the pinned revision.
The filesystem helper is recorded but does not count as chart behavior.

| Source span | Lines | SHA-256 | Role |
| --- | --- | --- | --- |
| `writeFile` | 14-22 | `4582ad5f11c031d2b70604df82f08d30ff21d977b6affb87bf0ade5eb7ae88ce` | Filesystem output helper; no chart behavior. |
| `main` | 24-51 | `227e252486af0ec1e89a2c6920253c42bc26ff2f0366907409ca9af320319fd9` | Five-by-five data, centered title, named axes, and 600x400 presentation. |
| Matrix literal | 25-31 | `85c0b5b518e8d8d8689b6579e1377224dba3f87f844e69a685203740d59e4705` | Exact twenty-five-value matrix. |
| Chart options | 33-37 | `44f2aeb070244f692526b134d79f35d364f4cf588dc30cb044d2d8f10f635182` | Data binding, centered title, and X/Y axis titles. |
| Painter and output | 39-50 | `a593149de21fcd6622c896759f082626f4557d2c3fb7d158a350d8bd15b4bea7` | 600x400 PNG rendering and filesystem delivery. |

The second documentation preview deliberately reuses the same pinned matrix as
a caller-style override of title, axis, padding, value-label, and gradient
options. It is not counted as a second upstream example or behavior.

Relevant HeatMap-specific finite presentation API evidence at the same revision
is pinned separately:

| Upstream API source | SHA-256 | Renderer-neutral coverage |
| --- | --- | --- |
| `heat_map.go` | `dd9b80660b9e0c0b11e5ec9ef00f5f10005a7edc1061da08f928cea15324b23c` | Complete matrices, sequential scale bounds, theme-aware gradient, padding, title, categorical axes, and value-label options. |
| `series_label.go` | `d7b176bea3679542e878c4c5703db3711d1e258b64efb7d19614a16fc5722611` | Visible labels, safe exact, integer, and humanized value formats, font size, distance, and offset. |
| `title.go` | `e85f6a0fe2e8fd7c253ac226d780164beb0cc7214e94979241d1e9fbce824b26` | Visibility, subtext, logical placement, font sizes, and border width. |
| `axis.go` | `72d8ed4c1253122a3ac49e76fbf00ff905a75aff26bcc9e4b6aea48325372a07` | Title and label font sizes, degree-based rotation, label count, and count adjustment. |
| `chart_option.go` | `0b298fcd45fab6bbe476514d90e5107d65d32bd8ba985f2411e79ec88fd2b858` | Generic option-function construction remains private because no dedicated HeatMap option-function example exists. |
| `painter.go` | `f4ac102e9b21623765e2fdfe4c0910a03265bc751b9f5d019ae41e80611be959` | SVG rendering, dimensions, PNG encoding, and filesystem delivery remain private implementation details. |

Unsupported dedicated HeatMap-family behaviors: none. Every behavior in the
one dedicated file maps to typed renderer-neutral configuration. Ragged and
empty matrices are not dedicated example behaviors and remain outside the
component contract because its adjacent exact-value table requires a complete
matrix. Arbitrary label callbacks, raw theme, painter, output-encoder, and
generic option-function types remain private. Safe built-in formats, chart
tokens, caller gradient stops, shared lifecycle controls, exact tables, and SVG
or PNG export provide supported equivalents.

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

## Interactive Scatter

- Source repository: `github.com/go-echarts/examples`
- Revision: `bda428480a82d6d77ebb9fa939cf8d52528453dd`
- Standard source: `examples/scatter.go`
- Standard source SHA-256: `a77ddbf7580210a842a3e1d3966ab62c3f229fdb1a33df8f319ef029bd4188b5`
- Effect source: `examples/effectscatter.go`
- Effect source SHA-256: `1bf49dc5fb02b248ff6794aa549836b4c8fa02ddb89be6adc0c4574327673f1a`
- Status: all five upstream behavior functions are covered by the one
  renderer-neutral `interactive.Scatter` component. Effect scatter remains a
  typed variant with shared or per-series ripple options; it is not another
  component, kind, page, or route.
- Deterministic adaptation: ambient random values are replaced by the recorded
  local seed-1 sequence in original behavior-function and series-call order.
  Six categories or players, series names, integer values in `[0,100)`, and
  standard point geometry remain unchanged. The upstream `Shooting` label has
  one trailing space; Goshtoso corrects it to `Shooting`.

| Upstream behavior function | Coverage | Goshtoso Charts treatment |
| --- | --- | --- |
| `scatterBase` | Example | Two categorical series with round-rectangle symbols, size 20, and 10-degree rotation |
| `scatterShowLabel` | Example | Two categorical series with visible labels positioned right |
| `scatterSplitLine` | Example | Player A and Player B with Sports and Score axis names and visible split lines |
| `esBase` | Example | Dunk series using the existing Scatter effect variant and default ripple treatment |
| `esEffectStyle` | Example | Dunk stroke ripple at period 4 and scale 10; Shoot fill ripple at period 3 and scale 6 |

Every function and method in both pinned source files is inventoried below.
Function SHA-256 values cover exact source text from each `func` declaration
through its closing brace at the pinned revision.

| Source function or method | SHA-256 | Role |
| --- | --- | --- |
| `generateScatterItems` | `2cfe0abcb152c7020f5da65f8e22e616e665263a8e31918edc636499e08d4bb6` | Deterministic standard-point generator adaptation |
| `scatterBase` | `08faff249e4c7eaa65b602662f01896f4d745fda2c131526b3ac01b5354723b5` | Behavior example |
| `scatterShowLabel` | `99ead467aac7ae752e29e2912a3a7f3657c62fc05e0e0685074e1f2db5bc3623` | Behavior example |
| `scatterSplitLine` | `55d2f4d9d6e87a356894fcd71320c94a2ccc40abe4ba616587a759aa9f53acb7` | Behavior example |
| `ScatterExamples.Examples` | `39e747d19ef16f9ac2d20191242e80bf4bcb278d3db3c8f08f3e8891ad231ac9` | Page composition only |
| `generateEffectScatterItems` | `0f295b3eef4924158ea1b39bc0fc60ecc3b86afd61cf694943e9ebca4899a399` | Deterministic effect-point generator adaptation |
| `esBase` | `c2cfce12547c08c27942e5aa51df2fd6a332a9ec9e9db092bdbbe60a6ba3bf69` | Behavior example |
| `esEffectStyle` | `0fa720ed3f610334a6bf0cb8c4bcf8f33c79167f7f9033b2ced0996751316888` | Behavior example |
| `EffectscatterExamples.Examples` | `844e71f0cb14680df92458b561662a41e4cb2ed3cbb83fdbdca9d3c9c376c19d` | Page composition only |

Dedicated Scatter and effect-scatter source behaviors requiring a raw renderer
option or backing-engine public type: none. These files do not exercise visual
maps or pieces, dataset transforms, data zoom, statistical references, or
mixed-family composition, so this source-family slice does not invent those
treatments. Shared typed title, tooltip, axis, label, theme, controls, PNG,
resize, and wrapper lifecycle behavior remains available.

The three page-layout sources and immutable hashes in the shared supplementary
evidence table above also bound Scatter layout: center, flex, and unmanaged
page composition stay consumer-owned and do not become chart API modes.

## Interactive Pie

- Source repository: `github.com/go-echarts/examples`
- Source file: `examples/pie.go`
- Revision: `bda428480a82d6d77ebb9fa939cf8d52528453dd`
- Source SHA-256: `a59bb6f11818d4175d033f025f00a58e6a191eff5acf30f0e0cd5f98cd493ada`
- Status: all nine upstream Pie behavior functions are covered by the one
  renderer-neutral `interactive/pie.Pie` component. No Pie function requires a raw
  renderer option or backing-engine public type.
- Deterministic adaptation: the upstream ambient random sequence is replaced by
  its recorded seed-1 value groups in original call order. The upstream
  `Autumn ` label typo is corrected to `Autumn`.
- Runtime compatibility: `pieRadiusWithPadAngle` requires Apache ECharts 5.6 or
  newer. The vendored and pinned CDN core is therefore 5.6.0; the compatible
  runtime is an implementation dependency, not part of the chart API.

| Upstream function | Coverage | Goshtoso Charts treatment |
| --- | --- | --- |
| `pieBase` | Example | Basic four-season distribution |
| `pieShowLabel` | Example | Visible sector names and exact values |
| `pieRadius` | Example | 40–75% donut radii with labels |
| `pieRadiusWithPadAngle` | Example | Five-degree sector padding, 40/50% center, vertical legend with four-side padding, hidden labels, and name/share item tooltip |
| `pieRoseArea` | Example | Equal-angle area rose with 40–75% radii |
| `pieRoseRadius` | Example | Proportional-angle radius rose with 30–75% radii |
| `pieRoseAreaRadius` | Example | Area and radius roses at independent 25/50% and 75/50% centers |
| `pieInPie` | Example | Thin outer area ring and inner radius rose at one center |
| `pieWithDispatchAction` | Example | Typed one-second rotating emphasis and item tooltip, implemented in the private runtime and disabled for reduced-motion users |

The three page-layout sources and immutable hashes listed in the shared
supplementary evidence table above also bound Pie layout: centered, flex, and
unmanaged pages do not become public chart modes. The Pie page adds one local
selectable-sector example to document the neutral selected-state API; it is not
counted as a tenth upstream function.

## Interactive Radar

- Source repository: `github.com/go-echarts/examples`
- Source file: `examples/radar.go`
- Revision: `bda428480a82d6d77ebb9fa939cf8d52528453dd`
- Source SHA-256: `f6b8e26399826e7f979717fbb4a30b48a8c8d10e8f496da60c430aaadc0e8ffb`
- Status: all four upstream behavior functions are covered by the one
  renderer-neutral `interactiveradar.Radar` component in
  `components/interactive/radar`.
- Data correction: upstream rows contain six pollutant dimensions followed by
  a seventh value used as the day index. Goshtoso retains the first six values
  as the aligned radar vector and names each observation from that day index.
- Theme adaptation: the upstream fixed dark surface and fixed series colors
  are replaced by Goshtoso theme tokens so light, dark, and named themes retain
  contrast without exposing renderer configuration.

| Upstream behavior function | Coverage | Goshtoso Charts treatment |
| --- | --- | --- |
| `radarBase` | Example | Twenty-one Beijing observations on the default polygon coordinate with split areas and lines |
| `radarStyle` | Example | Circular coordinate, five splits, subtle split lines, and translucent lines and areas |
| `radarLegendMulti` | Example | Beijing, Guangzhou, and Shanghai with independent multiple legend selection |
| `radarLegendSingle` | Example | The same three cities with exclusive single legend selection and stronger areas |

Every function and method in the pinned source file is inventoried below.
Function SHA-256 values cover exact source text from each `func` declaration
through its closing brace at the pinned revision.

| Source function or method | SHA-256 | Role |
| --- | --- | --- |
| `generateRadarItems` | `f906e8292d6830bb7983f954d67de496fd21b78dc709dbc86633f1f38d6435ae` | Data adaptation |
| `radarBase` | `e897284229b2e8a01ac1a57a65a4e779c9375083bb6fad2899ab785ce7e808d7` | Example |
| `radarStyle` | `45738725345bf456020df60ea819afbb8931d079cfba95153dd9cfb77b529eaa` | Example |
| `radarLegendMulti` | `692e14a0b753d77b4ea4bf47aa95192d69d42ebc82319e2e47378406b507cb59` | Example |
| `radarLegendSingle` | `503fc8e155fc5a43597ed9fa8a015be1f743ffb865e94fb705f4eb9b2d48a528` | Example |
| `RadarExamples.Examples` | `e5c8ddab877b5227eec0975bcdf4b36531b5212b20d058e4d785d6f99b5e91d8` | Page composition only |

Unsupported dedicated Radar-family behaviors: none. Shape, split count,
split-area visibility, split-line styling, line and area opacity, multiple or
single legend selection, responsive sizing, exact values, theme colors,
controls, PNG export, and wrapper lifecycle are all typed and renderer-neutral.

## Interactive Candlestick

- Source repository: `github.com/go-echarts/examples`
- Source file: `examples/kline.go`
- Revision: `bda428480a82d6d77ebb9fa939cf8d52528453dd`
- Source SHA-256: `712b738662e87ceaab96fe9a3b39cc2591184db4de519a34f628ccee067f0489`
- Status: all five dedicated behavior functions are covered by the one
  renderer-neutral `interactive.Candlestick` component; all 88 ordered OHLC observations are preserved exactly.
- Theme adaptation: the style example's fixed red and green fills and borders
  become Goshtoso direction-token classes. Callers can still supply explicit
  colors and borders through typed direction styles.

| Exact source span | Lines | SHA-256 | Role and coverage |
| --- | --- | --- | --- |
| `klineData` | 12–15 | `844d53233a5d826fdde0d8286bc328d24c1a89066507a30d9505e124f9dc67bc` | Data helper type; adapted to typed `Candle` values ordered open, close, low, high. |
| `kd` | 17–106 | `94baedf445f705b38028f2b9f91997d924be7ed7eac046a25cba1d1bb20e4143` | Exact 88-row dataset variable from 24 January through 13 June 2018. |
| `klineBase` | 108–137 | `83380beaca22d81cce6a0d38facb27ae206d96ccc715b48c647460ad4ac026da` | Category split count 20, scaled value axis, and a visible X-axis slider over the 50–100% window. |
| `klineDataZoomInside` | 139–169 | `d8197311ea2164c43384921c37b0a60a4e7860c1c81489f2c6b16b8a3808325d` | Inside-only X-axis zoom over the 50–100% window. |
| `klineDataZoomBoth` | 171–207 | `a8291af22b113f59db07c8d09b81fe4ba3e8b5f5914e0769d3fb2a3efecf5fab` | Inside and slider X-axis zoom controls over the same window. |
| `klineDataZoomYAxis` | 209–239 | `c60b43ac2b542c04cb041bba90c7cc980e484c06bb6e297dcc0663174f160db3` | Visible Y-axis slider over the 50–100% value window. |
| `klineStyle` | 241–293 | `671aa66b391c119d3b8c3e8bd5da4f064bdb0341c1e4882ab57988b2138574c9` | Direction-token classes, typed border overrides, and labeled highest/lowest range marks. |
| `KlineExamples.Examples` | 295–313 | `d32f020d851c51f971b46e4cd9fa04cb1c76b12e392b2f18df18d28191d3255a` | Page composition; preserves the five-example order without entering the component API. |

Unsupported dedicated Candlestick behaviors: none. OHLC data, scaled axes,
split count, slider and inside zoom on either axis, combined zoom controls,
direction classes or explicit paint, border styling, and highest or lowest
range marks are typed. The upstream page renderer remains outside the public
chart API because consumers own page layout.

## Interactive Funnel

- Source repository: `github.com/go-echarts/examples`
- Source file: `examples/funnel.go`
- Revision: `bda428480a82d6d77ebb9fa939cf8d52528453dd`
- Source SHA-256: `c532e6490bad284b4b6a5dec20825359abc795a8ee9f3bb5febbcfb4e0cd2d55`
- Status: both dedicated behavior functions are covered by the one
  renderer-neutral `interactivefunnel.Funnel` component in
  `components/interactive/funnel`.
- Deterministic adaptation: the ambient random helper becomes a local seed-1
  sequence in original function and helper-call order. `Visit`, `Add`, `Order`,
  `Payment`, and `Deal` retain their source sequence in typed data and the exact
  table, while chart geometry retains the upstream default descending-by-value
  order. Every generated value remains in the upstream `[0,50)` domain.
  Concrete fixture values are reproducible documentation data, not claimed
  upstream constants.

| Upstream behavior function | Coverage | Goshtoso Charts treatment |
| --- | --- | --- |
| `funnelBase` | Example | Basic five-stage Analytics series with the upstream title |
| `funnelShowLabel` | Example | A second five-stage value group with visible labels positioned left |

Every declared data source, function, and method in the pinned file is
inventoried below. Span hashes cover exact source text at the pinned revision.

| Exact source span | Lines | SHA-256 | Role and coverage |
| --- | --- | --- | --- |
| `dimensions` | 13 | `bd5b4e6a9c429f461f802686cfe0539b660b85c12fc382656d0097d07d1f7e83` | Ordered stage labels preserved exactly. |
| `genFunnelKvItems` | 15–21 | `9bf2059b03ae41f499ad99ac7ece0ac9deec70fe0fc3df8052310b1f8a64ac1c` | Ambient random helper adapted to a local fixed seed while preserving call order and value domain. |
| `funnelBase` | 22–30 | `88c9efbc1bdda11af5c7cbad673b7cd568941ed9248a6428ec5ea5c45184fd45` | Basic example. |
| `funnelShowLabel` | 33–47 | `ed198502f5b56653897070b7ba9b7fb862a8ceb4255ac4242cd817d0be0e23d7` | Visible left-label example. |
| `FunnelExamples.Examples` | 51–63 | `6c55c7a63e033a5b0f6de045c16c31d16eee67a134000a1e240da88ffbb1ca97` | Page composition only; preserves the two-example order without entering the component API. |

Unsupported dedicated Funnel behaviors: none. Named finite nonnegative values,
caller or value ordering, title, labels and left placement, responsive sizing,
theme colors, exact-value disclosure, controls, PNG export, resize, and wrapper
lifecycle are typed and renderer-neutral. The upstream page renderer remains
outside the public chart API because consumers own page layout.

## Interactive HeatMap

- Source repository: `github.com/go-echarts/examples`
- Source file: `examples/heatmap.go`
- Revision: `bda428480a82d6d77ebb9fa939cf8d52528453dd`
- Source SHA-256: `c08b194eafa5e02e941ad91f7ff8402448bc77b407cc97903b19063d06dd6f14`
- Status: both dedicated behavior functions are covered by the one
  renderer-neutral `interactive/heatmap.HeatMap` component.
- Deterministic adaptation: the Cartesian grid preserves all 168 source cells,
  seven weekday labels, twenty-four hour labels, ordering, and explicit missing
  cells. Calendar generation uses a local seed-1 sequence over the same 366-day
  span and `[0,21)` value domain; zero outcomes remain explicit missing cells.
  Fixed source colors become theme-aware cold, middle, and warm scale tokens.
- Responsive adaptation: the source's fixed 20-pixel calendar cells become
  automatic cells in the documentation example so the complete year remains
  contained at narrow and wide consumer widths. The public calendar option
  retains typed fixed cell sizing when a consumer provides enough space.

| Upstream behavior function | Coverage | Goshtoso Charts treatment |
| --- | --- | --- |
| `heatMapBase` | Example | Exact 24-by-7 category grid, split areas, calculable 0–10 scale, and explicit no-data cells |
| `heatMapCalendar` | Example | Deterministic 366-day calendar, horizontal layout, positioned bounds, cell borders, 0–20 scale, and explicit no-data days |

Every declared data source, helper, function, type, and method in the pinned file
is inventoried below. Span hashes cover exact source text at the pinned revision.

| Exact source span | Lines | SHA-256 | Role and coverage |
| --- | --- | --- | --- |
| `hmData` | 15–44 | `4e16c58885e24812b235c9adb709b6357b4dfa62765a448b306b94f1b50f6774` | Exact 168-cell Cartesian dataset; zero source values preserve the source helper's explicit no-data treatment. |
| `weekDays` | 46 | `f2105456e595925b61ad3e74deea94a9ca7532f0f70a6bf6a9ff72360e3620c3` | Exact seven labels and order. |
| `dayHrs` | 48–51 | `abe6227a0d8c6fd7abb77b023bdd385d6ca46f8ea0a0c4623bbfff9db4895f8d` | Exact twenty-four labels and order. |
| `genHeatMapData` | 54–64 | `c2e58d4394f9f3e352048d7f6e8863b71419ef26b97ac8dfc223d75328219bcb` | Cartesian coordinate swap and zero-to-missing transformation. |
| `heatMapBase` | 66–93 | `56319619a561ed145f00ae74b66c0d9ad7403f5f5d508c4c1def0742f97f9218` | Basic category-grid example. |
| `genHeatMapCalendarData` | 95–107 | `9f4cb439c4297c6cb550fa7680e6497390a024b47d8e8404dac75d1bf3ed96a8` | Ambient random calendar helper adapted to a fixed local seed. |
| `heatMapCalendar` | 109–142 | `9ee96ab15abff25ed188f25eb885a215bcba6ac69a074c5e21a0280ef32db596` | Calendar-coordinate example. |
| `HeatmapExamples` | 144 | `bdb7c3f6f5387d8c8ec512b899e7853b7572ff96c04d9ddcbe640d42b2c7b702` | Page-group marker only. |
| `HeatmapExamples.Examples` | 146–158 | `aebfa38bdcaaeb0b9e27457f83e925c7b42d65e3fd6352f99fe318e1cae7d286` | Page composition only; preserves the two-example order without entering the component API. |

Unsupported dedicated HeatMap behaviors: none. Category and calendar
coordinates, missing observations distinct from zero, split areas, calculable
continuous scales, theme-aware cold-to-warm colors, calendar placement,
orientation, cell size and borders, responsive sizing, exact-value disclosure,
controls, PNG export, resize, and wrapper lifecycle are typed and
renderer-neutral. The upstream page renderer remains outside the public chart
API because consumers own page layout.
