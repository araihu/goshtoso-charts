# Goshtoso Charts

Static/vector and opt-in interactive chart components for Go/templ applications
using [Goshtoso](https://github.com/araihu/goshtoso). The public contract stays
renderer-neutral: typed configs, instances, and options do not expose a backing
renderer. Initial product driver was Xisnove monitoring, but the library is
product-neutral.

Charts use Goshtoso theme tokens, including Tailwind-compatible chart tokens,
and responsive Goshtoso ActionGroup controls. Static/vector and interactive
families deliberately differ where their proven export capabilities differ.
Interactive runtime assets are vendored locally by default; CDN delivery is an
explicit opt-in. Components provide accessible labels and adjacent exact data
where a visual alone would be insufficient; interactive charts can also receive
live Cartesian data through a renderer-neutral SSE snapshot contract.

## Install

Requires Go 1.26.5+ and a Goshtoso-integrated templ application.

```bash
go get github.com/araihu/goshtoso-charts
```

## Component contract

Every chart follows Goshtoso component shape: a typed `Config`, concrete `Instance`, stable `Kind()`, and `templ.Component` rendering. Extension-owned kinds live in `github.com/araihu/goshtoso-charts/components`; they deliberately do not modify Goshtoso's core kind registry.

Chart surfaces use Goshtoso tokens. Categorical series use Tailwind-compatible
chart tokens (`--color-chart-series-1` through `--color-chart-series-8`) so a
theme's small semantic palette does not make unrelated series look equivalent.
Surface, outline, grid, and text equivalents use the same `--color-chart-*`
namespace. Tailwind arbitrary-variable utilities such as
`text-(--color-chart-series-1)` and `bg-(--color-chart-surface)` can reuse them.

## Chart palettes

Default `charttheme.PaletteAuto` follows dedicated chart tokens for all 16
built-in Goshtoso themes. Neutral, Pastel, Bold, Status, and Arai Hû palettes
remain available when an application needs an explicit semantic treatment.

```go
import "github.com/araihu/goshtoso-charts/components/charttheme"

style := charttheme.Style{
	Palette: charttheme.PalettePastel,
	Class:   "rounded-lg ring-1", // Tailwind utilities or an application class
}
```

Pass `Style: style` to SSR or typed interactive configs. Explicit ordered colors
have highest priority:

```go
Style: charttheme.Style{
	Palette: charttheme.PaletteAraiHu,
	Colors:  []string{"#2563eb", "oklch(70% 0.19 25)"},
	Class:   "my-chart",
},
```

Application CSS may target `my-chart` and override any `--color-chart-*` token;
the built-in theme values are defaults, not a closed styling boundary.
Interactive canvas
charts resolve these CSS tokens through the private runtime bridge and refresh
when Goshtoso theme or dark-mode state changes. Explicit `Style.Colors` remain
authoritative.

## Source policy

Examples preserve a concrete upstream example's dataset, hierarchy, and chart
semantics before being adapted to typed Goshtoso APIs, accessibility, responsive
layout, and theme tokens. Static/vector examples originate in
[`go-analyze/charts`](https://github.com/go-analyze/charts/tree/main/examples);
interactive examples originate in
[`go-echarts/examples`](https://github.com/go-echarts/examples/tree/master/examples).
The demo's [Attributions page](site/internal/pages/attributions.go) centrally
records exact example paths, revisions, runtime, and geography sources.

## Line chart

Use the chart as supporting evidence inside a Goshtoso `panel.Panel`; keep exact data in nearby text or a table when users need it.

```templ
package dashboard

import (
	"github.com/araihu/goshtoso/components/panel"
	"github.com/araihu/goshtoso-charts/components/line"
)

templ SignupTrend() {
	@panel.Panel(panel.Config{
		Header: templ.Raw("<h2>Signups</h2>"),
		Body: line.Line(line.Config{
			Label:   "Weekly signups",
			Caption: "Seven-day trend",
			Labels:  []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
			Series: []line.Series{{Name: "Signups", Values: []float64{12, 18, 15, 21, 24, 19, 28}}},
		}),
	})
}
```

Line charts render SVG geometry on the server, then resolve surfaces from
Goshtoso tokens and series from chart tokens at display time. They follow theme
and `.dark` surface changes without browser rendering or a fresh response.

## Bar chart

Use bar charts for categorical comparisons such as deployment outcomes, monitor-result counts, or status breakdowns. `Stacked` keeps multiple named series within each category; leave it false for side-by-side comparisons.

```templ
@bar.Bar(bar.Config{
	Label: "Deployments by environment",
	Labels: []string{"Development", "Staging", "Production"},
	Series: []bar.Series{
		{Name: "Successful", Values: []float64{18, 12, 9}},
		{Name: "Failed", Values: []float64{1, 2, 1}},
	},
	Stacked: true,
})
```

Vertical remains the zero-value orientation. Set `OrientationHorizontal` for
category labels beside the bars; horizontal charts include an adjacent exact
value table.

```templ
@bar.Bar(bar.Config{
	Label:       "World population by reporting series",
	Title:       "World Population",
	Orientation: bar.OrientationHorizontal,
	Labels:      []string{"UN", "Brazil", "Indonesia", "USA", "India", "China", "World"},
	Series: []bar.Series{
		{Name: "2011", Values: []float64{10, 30, 50, 70, 90, 110, 130}},
		{Name: "2012", Values: []float64{20, 40, 60, 80, 100, 120, 140}},
	},
	Padding: bar.Padding{Top: 20, Right: 40, Bottom: 20, Left: 20},
	Width: 600, Height: 400,
})
```

## Pie chart

Use pie charts for a small categorical distribution that is meaningful as parts of a total, such as retained monitor observation states. Avoid them for ordered time-series data or many categories.

```templ
@pie.Pie(pie.Config{
	Label: "Observation states",
	Slices: []pie.Slice{
		{Name: "Up", Value: 94},
		{Name: "Degraded", Value: 4},
		{Name: "Down", Value: 2},
	},
})
```

Keep exact values in nearby text or a table when readers need them.

## Scatter chart

Use scatter charts for sparse observations or dense aligned samples across
explicit ordered categories. The component covers all five dedicated upstream
Scatter-family files at revision `1fe31b06b8a82e00df877ff4417a75858547c1c2`:
basic aligned data and a missing observation, per-series symbols, dense repeated
samples and statistical guides, top-N labels, and the option-function source's
distinct circle and integer-format treatment. This example preserves the dense
source with a fixed local RNG seed: three 1,000-category bounded random walks,
repeated samples, SMA(100), and maximum references. Categories remain equally
spaced keys.

```templ
@scatter.Scatter(scatter.Config{
	Label:      "Dense scatter data",
	Categories: labels,
	Width: 600, Height: 400,
	Options: scatter.Options{Size: 0.5, Trend: scatter.TrendLine{
		Kind: scatter.TrendSimpleMovingAverage, Period: 100,
	}},
	Series: []scatter.Series{
		{Name: "One", Values: values[0], Options: scatter.Options{ReferenceLine: scatter.ReferenceLineMaximum}},
		{Name: "Two", Values: values[1], Options: scatter.Options{ReferenceLine: scatter.ReferenceLineMaximum}},
		{Name: "Three", Values: values[2]},
	},
})
```

`Values` aligns zero, one, or many samples to each category without repeating
category strings. Sparse `Points` remains supported; one series cannot mix both.
`TitleOptions`, `LegendOptions`, `CategoryAxisOptions`, and `ValueAxisOptions`
retain renderer-neutral typography, placement, padding, label sampling,
rotation, bounds, units, and theme-aware presentation. `ValueFormatInteger`
provides whole-number labels without exposing a formatting callback.
For large datasets, keep exact data in a caller-owned table, download, or drill-down
instead of expanding thousands of values into the chart DOM.

## Shared controls and export

Every current chart accepts the same renderer-neutral control contract. All
charts default to Expand plus capability-derived Export. Fullscreen remains an
independent opt-in. Controls require the local `assets.Handler()` mount shown
below; Goshtoso's head dependencies provide ActionGroup measurement.

```go
import "github.com/araihu/goshtoso-charts/components/chartcontrol"

staticCfg.Controls = chartcontrol.Options{
	Fullscreen: true,
}
staticCfg.Export = &chartcontrol.ExportOptions{
	Filename: "weekly-signups",
}

interactiveCfg.Options.Controls = staticCfg.Controls
interactiveCfg.Options.Export = staticCfg.Export

// Explicit opt-outs remain renderer-neutral.
staticCfg.Controls.Expand = chartcontrol.Bool(false)
staticCfg.Export = &chartcontrol.ExportOptions{Disabled: true}
```

Static Line, Bar, Pie, Scatter, and Radar expose one Goshtoso Export dropdown
with SVG and PNG. Current interactive components, including Sunburst, expose one
direct PNG Export button.
Zero proven formats render no Export control; one renders a direct button; more
than one renders a Goshtoso Dropdown. Unsupported explicit formats fail
rendering instead of creating dead controls. Expand and fullscreen
preserve chart DOM, interaction state, and live SSE instances.
When Expand and Fullscreen are both enabled, one stacked Expand Dropdown offers
both choices at wide widths. Goshtoso ActionGroup keeps Expand primary and
flattens stacked Fullscreen and export actions into one icon-only overflow
Dropdown as space contracts; menus never nest.
The site route `/docs/chart-controls` is the canonical wrapper guide; it covers
defaults, lifecycle states, client events, responsive actions, export, HTMX,
accessibility, and failure behavior. `/docs/chart-modes` compares static/vector
and interactive delivery. Repository maintainers can use the
[chart-control implementation evidence](docs/chart-controls.md) to locate the
corresponding source and verification gates without maintaining a second copy
of the consumer contract.

## Candlestick

`candlestick.Candlestick` renders typed OHLC data as accessible SSR SVG. Each
datum must satisfy `Low <= Open/Close <= High`; an adjacent table exposes exact
values and increase/decrease text without relying on color.

```templ
@candlestick.Candlestick(candlestick.Config{
	Label:      "Seven-day stock price",
	Title:      "Candlestick Chart",
	SeriesName: "Stock Price",
	Data: []candlestick.Datum{
		{Label: "Day 1", Open: 100, High: 110, Low: 95, Close: 105},
		{Label: "Day 2", Open: 105, High: 115, Low: 100, Close: 112},
	},
})
```

## Interactive components

Interactive components expose only typed Goshtoso Charts configs. Renderer
types and callbacks are implementation details and cannot cross the public API.

Interactive charts require one browser runtime tag. Local, vendored delivery is
the default:

```go
import chartassets "github.com/araihu/goshtoso-charts/assets"

mux.Handle("GET "+chartassets.Prefix, chartassets.Handler())
```

```templ
import "github.com/araihu/goshtoso-charts/components/dependencies"

templ Layout() {
	<head>
		@dependencies.Dependencies()
	</head>
}
```

Mount `assets.Handler()` directly; it already removes `/charts/assets/`.
Applications choosing third-party delivery must opt in explicitly with
`dependencies.WithCDN()`. See [interactive chart dependencies](docs/dependencies.md)
for custom paths, SRI, CSP, and load-order details.

```templ
import interactivebar "github.com/araihu/goshtoso-charts/components/interactive/bar"

@interactivebar.Bar(interactivebar.Config{
	Label: "Weekly deployments",
	XAxis: []string{"Mon", "Tue"},
	Series: []interactivebar.Series{{
		Name: "Production",
		Data: []interactivebar.Data{{Value: 3}, {Value: 5}},
	}},
})
```

Interactive Bar now uses its canonical chart-specific package. Existing
`interactive.Bar` code remains compatible during the additive migration.

```templ
import interactiveline "github.com/araihu/goshtoso-charts/components/interactive/line"

@interactiveline.Line(interactiveline.Config{
	Label: "Weekly latency",
	XAxis: []string{"Mon", "Tue"},
	Series: []interactiveline.Series{{
		Name: "p95 (ms)",
		Data: []interactiveline.Data{{Value: 42}, {Value: 47}},
	}},
})
```

Interactive Line uses its canonical chart-specific package too. Existing
`interactive.Line` code remains compatible. See
[interactive package migration](docs/interactive-package-layout.md) for the
shared-options boundary and phase-1 compatibility policy.

Word clouds keep weighted words, silhouettes, sizing, rotation, layout, and
semantic classes in a typed renderer-neutral config. Exact values remain in a
bounded adjacent table.

```templ
@interactive.WordCloud(interactive.WordCloudConfig{
	Label: "Weighted search terms",
	Series: interactive.WordCloudSeries{
		Name: "terms",
		Words: []interactive.Word{
			{Name: "Sam S Club", Value: 10000, Class: "retail"},
			{Name: "Macys", Value: 6181, Color: "#ff8a3d"},
		},
		Options: interactive.WordCloudSeriesOptions{
			Shape:     interactive.WordCloudShapeStar,
			SizeRange: &interactive.WordCloudSizeRange{Min: 14, Max: 80},
		},
	},
})
```

Use application-owned values only. Shared `ChartOptions`, `SeriesOptions`, and
component-specific typed variants provide controlled customization without
exposing the backing renderer.

Parallel coordinates use ordered typed dimensions and aligned observations.
Numeric and categorical values cannot be mixed accidentally:

```go
chart := interactive.Parallel(interactive.ParallelConfig{
	Label: "Air quality profiles",
	Dimensions: []interactive.ParallelDimension{
		{Name: "AQI", Range: &interactive.ParallelRange{Max: interactive.Float(300)}},
		{Name: "Level", Categories: []string{"Good", "Moderate", "Heavily"}},
	},
	Series: []interactive.ParallelSeries{{
		Name: "Beijing",
		Observations: []interactive.ParallelObservation{{
			Name: "Day 1",
			Values: []interactive.ParallelValue{
				interactive.ParallelNumber(55),
				interactive.ParallelCategory("Moderate"),
			},
		}},
	}},
})
```

### Live Cartesian data

Interactive Bar and Line components can opt into an SSE source. Each named
event carries a complete renderer-neutral `interactive.CartesianSnapshot`;
full replacement makes reconnects deterministic and avoids exposing renderer
update structures.

```go
config.Live = &interactive.LiveData{
	URL:   "/metrics/events",
	Event: "chart",
}
```

Maps use one component for Brazil's 26 states and Federal District. Region
names, UF codes, and values remain typed; caller colors or semantic classes
stay programmatic; exact values remain available beside the canvas.

```templ
@interactive.Map(interactive.MapConfig{
	Label:    "Brazil states",
	Geometry: interactive.MapGeometryBrazil,
	Variant:  interactive.MapVariantScale,
	Series: interactive.MapSeries{
		Name: "Brazil state values",
		Regions: brazilStateValues, // all 26 states plus DF
	},
})
```

Geo stays distinct from Map: it plots typed longitude/latitude/value points
over pinned Brazil-state or São Paulo-municipality geometry. Scatter and effect-scatter
behavior remain series kinds on one component; visual ranges use theme
cold-to-warm tokens unless callers provide explicit colors.

```templ
@interactive.Geo(interactive.GeoConfig{
	Label:    "São Paulo cities",
	Geometry: interactive.GeoGeometrySaoPaulo,
	VisualRange: &interactive.GeoVisualRange{
		Min: 0, Max: 100, Calculable: interactive.Bool(true),
	},
	Series: []interactive.GeoSeries{{
		Name: "geo",
		Kind: interactive.GeoSeriesScatter,
		Points: []interactive.GeoPoint{
			{Name: "São Paulo", Longitude: -46.63, Latitude: -23.55, Value: 12},
			{Name: "Campinas", Longitude: -47.06, Latitude: -22.91, Value: 76},
			{Name: "Ribeirão Preto", Longitude: -47.81, Latitude: -21.18, Value: 41},
		},
	}},
})
```

The demo's live-availability page uses this primitive with narrow stacked bars.
Availability remains application semantics and is not a separate chart type.

## Roadmap

Current library covers a broad static/vector family (including line, bar, pie,
scatter, radar, candlestick, funnel, heatmap, table, and violin) and an opt-in
interactive family with Cartesian, hierarchy, distribution, map/geo, and 3D
variants. Brazil-state and São Paulo-municipality geometry are bundled for the
typed map and geo contracts; live availability remains an application use of the
shared live Cartesian contract, not a separate chart type.

Future primitives need a concrete use case and upstream source, stable
renderer-neutral kind, accessible exact evidence, semantic-token palette,
responsive behavior, capability-derived controls/export, and focused tests.

See [chart-library evaluation](docs/chart-library-evaluation.md), [surface brief](docs/surface-brief.md), and [Xisnove heartbeat brief](docs/xisnove-heartbeat.md).

## Development

Delivery trade-offs and deferred broad checks live in the actionable
[technical-debt ledger](docs/technical-debt.md). Review it after every two
merged delivery lanes and before a release.

```bash
templ generate
go test ./...
git diff --exit-code
```

## Demo site

`site/` is a separate consumer module, shaped after the Goshtoso component demo catalog. It mounts Goshtoso assets at `/assets/`, Goshtoso Charts assets at `/charts/assets/`, and uses both dependency components.

```bash
cd site
GOWORK=off go run ./cmd/server
```

Open `http://localhost:8091`.
