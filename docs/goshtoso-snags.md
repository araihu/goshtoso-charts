# Goshtoso integration snags

## 2026-07-28: ThemeRiver boundary-gap adapter gap

The pinned private interactive-chart adapter can serialize ThemeRiver data and
single-axis placement, but its series type omits the renderer's two-sided
`boundaryGap` option. The renderer-neutral `ThemeRiverBoundaryGap` API therefore
uses the shared deterministic snippet-replacement boundary instead of exporting
raw maps, callbacks, or backing-renderer types. A future adapter upgrade should
remove that repair once its typed series model covers the option.

## 2026-07-28: Parallel adapter omits supported axis and inactive styling fields

The pinned private Go adapter exposes parallel-axis maximum, inverse, type, and
categories, but omits minimum, axis label/line options, and per-series inactive
opacity even though the bundled browser runtime supports them. The typed public
API retains those renderer-neutral controls. Its private serialization bridge
replaces only the validated parallel-axis array and named-series fields before
initialization; no raw option map or backing type crosses the component API.

Parallel theme and responsive behavior required no component runtime. The
existing shared runtime already observes the actual browser chart host, resolves
its current instance, and calls `resize()` after consumer flex, modal,
fullscreen, or collapse changes without disposing or reinitializing it.

- Interactive charts require one library-owned ResizeObserver so canvases follow consumer layout changes after initialization. Future chart components, including Sunburst, must register through the shared interactive runtime instead of creating component-specific observers.

## 2026-07-27: templ control flow must remain structurally separate

The first `bar.templ` draft placed `@chart` and `if cfg.Caption != ""` on one
line inside `<figure>`. `templ generate` rejected it with an unterminated `if`.
Keep templ control flow on its own indented lines around nested components, then
regenerate the checked-in `*_templ.go` artifact.

## 2026-07-27: templ runtime import

`line.templ` initially declared `import "github.com/a-h/templ"`. `templ generate`
already injects that import when the template refers to `templ.Component`, so Go
failed with `templ redeclared in this block`. Consumer and extension templates
must use templ helpers without declaring that runtime import themselves.

## 2026-07-27: consumer utility coverage

The released embedded Goshtoso stylesheet does not include every Tailwind
utility an application shell might use (for example, `lg:flex-col`). Keep the
component primitives on Goshtoso utilities and tokens, but put demo/application
layout rules in the consumer's own stylesheet rather than assuming an arbitrary
Tailwind class is emitted.

## 2026-07-27: component docs sidebar follow-through

The searchable grouped demo sidebar used only Goshtoso's bundled Alpine
dependency and application-owned CSS. `templ generate` regenerated both page
artifacts cleanly; no new Goshtoso component, helper, or generated-template
friction appeared beyond the existing consumer utility-coverage constraint.

## 2026-07-27: shell boundary correction

The first extraction treated a component documentation shell as a generic
catalog shell. Those are different navigation and content patterns. The public
API was renamed to `componentdocshell` before release, repeated component-page
composition moved to `componentpage`, and `catalogshell` remains unclaimed.

## 2026-07-28: component preview overlay revision

The Charts site had pinned `goshtoso-app-shells` before the merged rounded
preview overlay fix. Production review requires the shell-owned preview
outline, rather than an application-local border workaround, so consumer sites
must update to the revision containing `3ca565d` and verify the transparent
layout border plus `::after` outline in light and dark themes.

## 2026-07-28: v11 consumer-brand boundary

`componentdocshell.Brand` accepts a custom logo component and favicon URL, but
does not serve consumer-owned identity assets. The Charts site therefore mounts
the approved v11 SVGs itself at `/brand/` and supplies the logo surface, ink,
and signal variables for Goshtoso's existing `.dark` state. This keeps the
shell reusable and preserves the upstream SVG geometry unchanged.

## 2026-07-28: heatmap category-axis validation gap

In go-echarts v2.7.2, `charts.HeatMap.Validate` does not call
`RectChart.Validate`, so category data supplied through `SetXAxis` is omitted
from the rendered option. The HeatMap component writes the validated X and Y
category data to `XAxisList[0].Data` and `YAxisList[0].Data` directly. The
renderer-specific workaround is covered by `TestHeatMapRendersCartesianChart`.

## 2026-07-28: funnel ordering lacks a dedicated option helper

go-echarts v2.7.2 exposes no dedicated funnel-series option for ordering. The
Funnel component keeps ordering typed as `FunnelOrder` and applies the validated
value through the exported `charts.SingleSeries.Sort` field. Rendering tests
cover descending, ascending, and caller-data order.

## 2026-07-28: tree zero depth and responsive width need explicit mapping

go-echarts v2.7.2 omits `initialTreeDepth: 0` because its integer field uses
`omitempty`, even though zero means roots-only and differs from an unspecified
renderer default. The Tree component uses a private serialization sentinel and
replaces it before emitting browser code. Its renderer also defaults the chart
host to 900px, wider than the documentation preview at common desktop widths;
the demo therefore sets the public, renderer-neutral `Width` to `100%`. This
matches the official `go-echarts/examples` Tree sample, which uses a 100% chart
inside the default centered page layout. Its flex layout targets wrapping
multiple charts, while none deliberately supplies no page layout; neither
belongs in this single-component wrapper.

## 2026-07-28: Sunburst sample determinism and zero-value serialization

The official `go-echarts/examples` `examples/sunburst.go` sample generates two
unseeded fractional values for each of seven `parent-N`/`child-N` pairs. The
documentation keeps that exact hierarchy, naming, one-child shape, and
fractional-value semantics, but uses fixed values so visual and regression
checks are reproducible. In go-echarts v2.7.2, `opts.SunBurstData.Value` also
uses `omitempty`, so an exact zero disappears from generated options. The typed
Sunburst adapter uses a private negative sentinel after rejecting caller-owned
negative values, then restores JSON zero before emitting the browser script.

`charts.WithSunburstOpts` copies every field into the series. Applying it twice
silently clears click navigation and sort values from the first call when the
second call only configures zero-label behavior. Build one complete private
option value before applying the helper.

## 2026-07-28: interactive canvas must follow its consumer host

At desktop shell width, the renderer initialized an 847px canvas before the
consumer layout settled to a 607px chart host. `width: 100%` prevented page
overflow but did not resize the already-created canvas. The existing private
interactive theme runtime now registers one shared `ResizeObserver`, calls
`chart.resize()` immediately, and observes the figure for later consumer or
wrapper layout changes. Browser regression checks measured exact host/canvas
parity at 607px desktop and 277px mobile while preserving the same renderer
instance across layout and theme changes. Wrapper integration may carry
equivalent observer logic; reconcile that single private seam rather than
adding component-specific resize runtimes.

## 2026-07-28: chart controls generated templ import

- `control.templ` initially imported `github.com/a-h/templ` because it used
  `templ.Attributes` and `templ.SafeURL`. `templ generate` already injects that
  import for generated templates, producing `templ redeclared in this block`.
  Templ expressions may use the generated import without declaring it in the
  source import block.

## 2026-07-28: templ source paths depend on generation root

Running `templ generate` from the repository root writes nested-site source-map
filenames with a `site/` prefix, while running it from the nested `site` module
writes the same filenames relative to that module. A whole-root check and a
nested-site check therefore cannot both accept one generated path spelling.
The clean gate checks each root component template with `templ generate -check
-f`, then runs `GOWORK=off templ generate -check` from `site`.

## 2026-07-28: Modal has no component-content slot

Goshtoso v0.0.13 `modal.Config` accepts `Body` text but no `templ.Component`
body or slot. Chart Expand still uses `modal.Modal` for its labeled dialog,
Escape handling, focus trap, scroll lock, backdrop, and focus restoration; the
private controls runtime relocates the same chart content node into the Modal
body while open, then restores it without recreating renderer state.

## 2026-07-28: Treemap hierarchy data exceeds the adapter's node shape

go-echarts v2.7.2 exposes `opts.TreeMapNode.Value` as `int` and omits per-node
color or semantic metadata, while its series model accepts arbitrary private
data. The Treemap component validates renderer-neutral `float64` values and
recursive ownership, then replaces the private series data with an internal
JSON shape carrying leaf values, optional `className`, and optional item color.
Parent values stay omitted and must be zero in the public contract because
child values determine parent area.

The same adapter exposes no breadcrumb field even though the bundled browser
runtime supports it. Treemap serializes a validated typed breadcrumb into the
private series option; no raw maps or renderer types cross the public API.
Native `leafDepth: 1` is required to turn the upstream two-level file-system
sample's top-level directories into focusable roots. Browser checks confirmed
root-to-directory and breadcrumb-back transitions retain the same chart
instance. The focused depth intentionally presents the selected root while the
bounded adjacent table remains the exact descendant-value source.

## 2026-07-28: WordCloud adapter targets an older browser extension

go-echarts v2.7.2 registers an ECharts 4 compatibility runtime and an older
word-cloud extension, while Goshtoso Charts embeds ECharts 5.4.3. Its typed
`WordCloudChart` also exposes only shape, size range, and rotation range; the
supported rotation step, grid, draw-out-of-bound, layout animation, and layout
fields are absent. The component therefore vendors the ECharts-5-compatible
word-cloud extension v2.1.0 and restores validated renderer-neutral fields only
inside the private serialized series.

The adapter also injects a `Math.random` color callback. WordCloud replaces it
with deterministic data order and the shared theme palette. Explicit caller
colors take precedence over styled semantic classes, which take precedence
over palette colors. Shared host observation continues to call `resize()` on
the existing instance; flex, modal, theme, collapse, and fullscreen changes do
not add a component-specific runtime or dispose/reinitialize the chart.

## 2026-07-28: isolated browser gates use canonical installed dependencies

Codex worktrees do not inherit the ignored `site/node_modules` directory from
the canonical checkout. The site pins Playwright and Sharp exactly in
`package.json` but intentionally has no `package-lock.json`, so `npm ci` is not
a valid gate. Run browser tests with `NODE_PATH` pointed at the canonical
checkout's installed `site/node_modules`, or use an ignored symlink. Do not
generate or commit a lockfile as part of a component lane.
