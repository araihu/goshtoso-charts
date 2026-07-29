# Shared chart controls

`components/chartcontrol` supplies one renderer-neutral wrapper used by every
current static/vector and interactive chart. Expand and capability-derived
Export default on. `Fullscreen` remains an independent opt-in; Collapse is not
part of the public API or rendered controls. When Expand and Fullscreen are both
enabled, one visible stacked Goshtoso Dropdown named Expand offers Expand and
Fullscreen choices instead of adjacent peer controls. Expand retains its
enlargement icon and remains the primary verb.
Expand relocates the existing chart content into a large Goshtoso Modal and
restores it on close; it never clones or rerenders chart DOM. Fullscreen uses
the browser Fullscreen API and falls back to a fixed overlay where needed.
Settled layout resizes preserve live renderer identity across every transition.

Export presentation follows capability count: zero formats render no control,
one format renders one direct accessible Export button, and multiple formats
render one Goshtoso Dropdown with only the proven items. Current static charts
therefore show SVG and PNG in a dropdown; current interactive charts show one
direct PNG button. `ExportOptions` retains filename, background, and pixel-ratio
customization; `Disabled` is the explicit opt-out.

Goshtoso `ActionGroup` measures the local container. It keeps the primary
Expand action available, moves lower-priority actions into one icon-only,
end-aligned overflow Dropdown, and flattens stacked Expand/Fullscreen and
multi-format Export children into labeled sections. The same zero/one/many
export policy remains intact at the wide boundary, without nested menus.

Goshtoso `v0.0.14-0.20260729011747-809b903c1296` provides the public
`components/actiongroup` API and `/assets/js/action-group.js` dependency used
here. Chart controls no longer measure width or own responsive overflow.
Controls runtime v3 owns chart-specific modal, fullscreen, resize, and export
behavior only; immutable v1 and v2 paths stay served for compatibility.

## Verified export matrix

| Component family | Current components | SVG | PNG | Evidence |
|---|---|---:|---:|---|
| Static/vector | Line, Bar, Pie, Scatter, Radar, Candlestick, Funnel, Heat map, Table, Violin | Yes | Yes | Each component configures `go-analyze/charts` with `ChartOutputSVG`, then reads `Painter.Bytes`. Browser PNG export rasterizes that resolved SVG at its intrinsic dimensions. |
| Interactive | Bar, Line, Scatter, HeatMap, Pie, Radar, BoxPlot, Candlestick, Gauge, Funnel, Graph, Sankey, Tree, Sunburst, Treemap, Parallel coordinates, Theme river, Word cloud, Map, Geo | No | Yes | Every component uses the current browser chart instance. Apache ECharts `getDataURL` documents PNG as the default and branches on the active canvas/SVG painter. Goshtoso Charts currently initializes go-echarts v2.7.2 with its default canvas renderer, so only PNG is exposed. |

Static transparent SVG/PNG backgrounds are supported. Opaque remains default
and resolves the active Goshtoso surface color. Interactive transparent export
is rejected because the live theme runtime deliberately paints its chart
surface; temporarily mutating that live option would risk visible state.

Primary source checks:

- [Goshtoso ActionGroup](https://goshtoso.araihu.com/components/action-group)
  provides primary/secondary priority, container measurement, stacked actions,
  and flat responsive overflow.
- [Goshtoso Dropdown](https://goshtoso.araihu.com/components/dropdown) documents
  icon-only triggers, item icons, action handlers, end alignment, and keyboard
  behavior used by stacked and overflow actions.
- [Goshtoso component model](https://goshtoso.araihu.com/docs/component-model)
  keeps component-specific config types and concrete renderable primitives;
  the chart wrapper adds no generic library API.
- [Apache ECharts `getDataURL` and `getConnectedDataURL`](https://github.com/apache/echarts/blob/e647e7746279397170a1e654e5567cac73b79f9b/src/core/echarts.ts#L911-L1052)
  list `png`, `jpeg`, and `svg`, use SVG output only for an SVG painter, and
  otherwise rasterize through canvas.
- [Apache ECharts renderer guidance](https://echarts.apache.org/handbook/en/best-practices/canvas-vs-svg/)
  confirms canvas and SVG are initialization-time renderer choices.
- [`go-analyze/charts` v0.6.0 painter](https://github.com/go-analyze/charts/blob/v0.6.0/painter.go#L108-L161)
  selects `chartdraw.SVG` for `ChartOutputSVG` and returns encoded bytes.
- Completed example commit `a616b17` supplied reusable filename, resolved-color,
  MIME, PNG-signature/dimension, and light/dark artifact test patterns. Shared
  runtime code was adapted; its dedicated page and private runtime were not
  retained because every component page now provides the same concrete export
  controls through one public API and one shared runtime, without duplicating
  implementation or exposing a backing engine.

## Consumer-owned layout

The wrapper owns controls and local horizontal overflow only. It applies no
page grid, catalog, or documentation layout. Fixed-width charts remain fixed
and scroll inside the wrapper on narrow screens. Omitted width and
`Width: "100%"` use the full available local container width without replacing
the renderer host. Charts remain centered within their local content area.

This uses the official go-echarts examples only as layout references:
[center](https://github.com/go-echarts/examples/blob/bda428480a82d6d77ebb9fa939cf8d52528453dd/examples/page_center_layout.go)
keeps the default page behavior, while [flex](https://github.com/go-echarts/examples/blob/bda428480a82d6d77ebb9fa939cf8d52528453dd/examples/page_flex_layout.go)
and [none](https://github.com/go-echarts/examples/blob/bda428480a82d6d77ebb9fa939cf8d52528453dd/examples/page_none_layout.go)
are explicit consumer page choices. Shared controls do not select any of them.
Word-cloud catalog variants use a consumer-owned flex wrapper; shared
`ResizeObserver` handling watches each actual chart host and its immediate
container. Host, modal, fullscreen, explicit control, and window changes settle
over consecutive frames before the private runtime resizes the existing
instance with zero-duration resize animation. Removed figures unobserve both
targets; repeated registration is idempotent.
Map and Geo reuse pinned local Brazil-state and São Paulo-municipality geometry delivery.
Map owns named region values; Geo owns longitude/latitude coordinate series.

## Pair-integration coverage

Static Scatter and Radar use `ExportCapabilityStaticSVG`, expose SVG and PNG,
and run through the same artifact, theme, responsive-layout, fullscreen, and
modal checks as other static components. Interactive Tree and
Sunburst carry `ChartOptions` into the shared wrapper and expose PNG only. Tree
disclosure and Sunburst drill-down/back state remain on the same browser chart
instance across theme, resize, modal, and fullscreen transitions. All routes
participate in the complete capability matrix and public API leakage gates.

Heartbeat and status remain redirect-only legacy routes; controls add no
component or page semantics for them.
