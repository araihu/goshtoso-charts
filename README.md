# Goshtoso Charts

Static vector and interactive charts for Go applications using [Goshtoso](https://github.com/araihu/goshtoso). Initial product driver: monitor detail and public status pages in Xisnove. Public API remains product-neutral.

`goshtoso-charts` provides Go/templ chart components in two layers: lightweight
static vector primitives and opt-in interactive components. The public API is
renderer-neutral, and no Node process or CDN asset is required.

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

Default `charttheme.PaletteAuto` detects AraiHu in SSR markup and adds a warm,
contrasting palette around its lime accent. Themes without chart CSS receive the
high-contrast Bold fallback. Neutral and Pastel fallbacks are also built in.

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

Application CSS may target `my-chart` and override any
`--color-chart-series-*` token. Interactive canvas
charts resolve these CSS tokens through the private runtime bridge and refresh
when Goshtoso theme or dark-mode state changes. Explicit `Style.Colors` remain
authoritative.

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

As with line charts, keep exact values in nearby text or a table when readers need them.

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
explicit ordered categories. This example adapts upstream
`examples/1-Painter/scatter_chart-3-dense_data/main.go` with a fixed local RNG
seed: three 1,000-category bounded random walks, repeated samples, SMA(100), and
maximum references. Categories remain equally spaced keys.

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
For large datasets, keep exact data in a caller-owned table, download, or drill-down
instead of expanding thousands of values into the chart DOM.

## Shared controls and export

Every current chart accepts the same renderer-neutral control contract. All
charts default to Expand plus capability-derived Export. Fullscreen and collapse
remain independent opt-ins. Controls require the local `assets.Handler()` mount
shown below.

```go
import "github.com/araihu/goshtoso-charts/components/chartcontrol"

staticCfg.Controls = chartcontrol.Options{
	Fullscreen:  true,
	Collapsible: true,
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
rendering instead of creating dead controls. Expand, collapse, and fullscreen
preserve chart DOM, interaction state, and live SSE instances.
See the [capability matrix, layout contract, and pending integration checks](docs/chart-controls.md).

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
@interactive.Bar(interactive.BarConfig{
	Label: "Weekly deployments",
	XAxis: []string{"Mon", "Tue"},
	Series: []interactive.BarSeries{{
		Name: "Production",
		Data: []interactive.BarData{{Value: 3}, {Value: 5}},
	}},
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

The demo's live-availability page uses this primitive with narrow stacked bars.
Availability remains application semantics and is not a separate chart type.

## Roadmap

Foundation: static/vector `line`, `bar`/stacked bar, `pie`, and categorical
`scatter`, plus typed interactive components. Next generic primitives: area and
distribution/histogram. Add each only with a real use case, stable kind,
accessible evidence, semantic-token palette, and focused tests.

See [chart-library evaluation](docs/chart-library-evaluation.md), [surface brief](docs/surface-brief.md), and [Xisnove heartbeat brief](docs/xisnove-heartbeat.md).

## Development

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
