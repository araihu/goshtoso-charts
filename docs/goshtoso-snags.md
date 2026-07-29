# Goshtoso integration snags

## 2026-07-29: ActionGroup primary does not render stacked items

Goshtoso `v0.0.14-0.20260729011747-809b903c1296` removes the chart-owned
responsive toolbar: ActionGroup now measures the local container, renders small
buttons, and flattens grouped children into one overflow Dropdown. Its
`Action.Items` contract is rendered only for secondary actions, however; a
grouped primary is rendered as a plain button.

The chart consumer keeps a direct primary Expand action for constrained widths
and an equivalent first-priority stacked Expand/Fullscreen secondary action for
wide widths. Consumer CSS swaps their presentation when ActionGroup moves that
stacked action into overflow. Goshtoso remains unchanged. Collapse was removed
from the chart public API and runtime rather than carried into the new group.

## 2026-07-28: responsive action overflow is not a generic primitive

Goshtoso v0.0.13 `Toolbar` wraps search, filter, and action regions but exposes
no action priority, fit measurement, or overflow-menu contract. `Dropdown`
provides the accessible icon-only trigger and menu behavior, but labeled
triggers cannot also accept `TriggerIcon`; chart Expand therefore uses the
Dropdown's visible label plus a chart-owned token-colored icon treatment.

Goshtoso Charts composes only its known actions: Expand stays primary,
Fullscreen joins its menu, and constrained Collapse/export actions flatten into
one Dropdown without nested menus. Generic responsive action-toolbar overflow
belongs in base Goshtoso; chart-specific action priority and capability-derived
format choices belong here. No cross-repository change was made.

## 2026-07-28: dual-axis SVG classes and browser color spaces need private resolution

The static renderer accepts per-axis colors but has no semantic-class hook that
can cover the axis spine, ticks, labels, and title together. The Line adapter
therefore assigns private unique placeholder colors, decorates only matching
SVG elements with caller classes, then replaces placeholders with
renderer-neutral Goshtoso tokens or validated caller colors. None of those
renderer details escape the public API.

Modern browser computed styles may preserve theme colors as `oklch(...)`
instead of returning `rgb(...)`. Browser contrast checks cannot parse those
strings with an RGB-only helper. Resolve the computed color by painting one
pixel to a canvas and reading its RGB channels before calculating luminance.

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

## 2026-07-28: templ is pinned but not declared as a Go tool

The modules require `github.com/a-h/templ` v0.3.1020, but neither module has a
Go `tool` directive. `go tool templ generate` therefore fails with
`go: no such tool "templ"` even when the matching binary exists in `GOPATH/bin`.
Use the pinned `templ` binary directly for root and nested-site generation.

## 2026-07-28: modal geometry settles after its scale transition

Goshtoso Modal reports its full computed width and height immediately after
opening while its bounding rectangle remains at the transition's intermediate
scale. Browser geometry gates must wait for the 200 ms transition plus its
100 ms delay before asserting large-panel ratios. Renderer instance and
ResizeObserver convergence checks remain independent of that visual delay.

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

## 2026-07-28: semantic classes may not define a color

The interactive theme runtime originally treated any caller semantic class as
a color utility. A class used only for meaning inherits the figure text color,
which made Candlestick rise and fall treatments identical. The shared class
probe now compares the classed computed color with its unclassed inherited
color and retains the typed color or theme-token fallback when the class does
not actually set color. This keeps `Class` useful for application semantics
without silently overriding readable light/dark defaults.

## 2026-07-28: Geo point data omits paint metadata

go-echarts v2.7.2 `opts.GeoData` exposes only `Name` and `Value`, so it cannot
carry point color or renderer-neutral semantic-class metadata. The Geo
component validates typed longitude, latitude, value, Color, and Class fields,
then replaces only the private serialized series data with an internal shape
that retains those optional paint hints. Series and geometry paint metadata
stay private data attributes consumed by the shared theme runtime. No raw map,
callback, or renderer type crosses the public API. Component tests cover every
exported geometry, series, and point Color/Class field; browser tests prove the
resolved paints affect both scatter variants and follow light/dark theme changes.

## 2026-07-28: theme series fills may need a contrasting boundary

AraiHu light mode resolves its second chart-series token to a bright orange
against a warm light chart surface. The fill stays visually distinct but does
not reach 3:1 by itself. Horizontal Bar preserves that shared theme token and
adds a one-pixel chart-text-token boundary to series marks. Browser checks
require 4.5:1 chart-text contrast, 3:1 mark-boundary contrast, distinct series
fills, and the adjacent exact-value table across Goshtoso and AraiHu light/dark
modes. This avoids a component-local palette fork and keeps meaning independent
of color.

## 2026-07-28: Surface3D adapter serializes the wrong series type

go-echarts v2.7.2 `charts.Surface3D.AddSeries` passes the Scatter3D series type
to its shared 3D-series builder. The resulting option registers `scatter3D`
instead of the required `surface`, even though `Surface3D.Type()` reports the
surface chart type. The Surface3D component repairs only the private serialized
series discriminator after using the typed adapter. Public configuration stays
renderer-neutral, and component plus browser tests assert the final series type.

Dense mathematical grids expose a second integration constraint: 18,000 exact
points cannot become initial-DOM table rows. Surface3D keeps the formula, point
count, and computed X/Y/Z domains visible, then provides an adjacent,
user-triggered deterministic CSV download containing every point in plot order.
The browser gate verifies CSV counts and endpoints alongside independent Go and
browser serialization hashes.

## 2026-07-28: fixed Heatmap reserves must account for responsive hosts

The visual-scale overlap fix initially used grid left `64`, right `5%`, and
visual-scale left `12`. After integration with shared responsive sizing, the
390px documentation viewport produced a 292px chart host and only a
134.694921875px plot. Scale and Y labels stayed separated, but the categorical
plot became too narrow. Grid left `52`, right `0`, and visual-scale left `8`
preserve at least 12px measured scale-to-label separation while retaining at
least 160px of plot width across the 390px matrix. The same-instance modal,
caller colors, exact data and domain, and PNG dimensions remain unchanged.

## 2026-07-28: map iteration cannot define Surface3D validation order

Surface3D validated X, Y, and Z axes by ranging over a Go map. An entirely
empty typed axis set could therefore report the Z-axis error before the
contract's expected X-axis error, making the root test gate nondeterministic.
Validation now visits X, Y, then Z through an ordered slice. Rendering and the
renderer-neutral public API are unchanged; the focused invalid-config test
passes repeatedly.

## 2026-07-28: geographic simplification needs visual review and explicit legend control

An initially evaluated municipality-derived Brazil outline passed identity and
hash checks but produced visible sliver artifacts after dissolve and
simplification. Browser screenshots exposed the defect, so Map and Geo now use
IBGE's official 2025 state and municipality Shapefiles directly, with pinned
input and output hashes, CC BY 4.0 attribution, keep-shapes simplification, and
deterministic bounds checks. Further visual review found two private adapter
defaults: an unused legend collided with the title at 320px and 390px, and map
legend symbols appeared as one dot at every state centroid even with that legend
hidden. Both components suppress the legend, Map suppresses its symbols in the
private theme bridge, and rendered-text checks prove the labels variant still
draws all 27 official UF codes without long-name collisions. The adjacent exact
table retains every Portuguese state name and UF pairing. The public API remains one renderer-neutral
Map and one renderer-neutral Geo component.

## 2026-07-28: controls v3 integration exposed stale browser selectors

Final pair integration initially passed the focused controls suite but failed
57 of 176 full browser tests. Eight older component-specific files still
expected four adjacent buttons or clicked the modal component's intentionally
hidden Expand button. Controls v3 instead presents one primary Expand action,
places Fullscreen in that primary menu, and moves Collapse and export actions
into responsive overflow.

The affected tests now exercise the public control composition: open the
primary action and choose Expand or Fullscreen when present, use overflow for a
hidden Collapse or export action, and expect two visible buttons at narrow
widths versus three at wide widths. Renderer, data, theme, and geometry
assertions were unchanged. The complete random-port browser gate then passed
176 of 176 tests at concurrency one.
