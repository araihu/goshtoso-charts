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

Use scatter charts for observations across explicit ordered categories. This
example adapts the upstream
`examples/1-Painter/scatter_chart-1-basic/main.go` dataset. Its null Thursday
value is represented by omitting the typed point. Categories are equally spaced
keys; use another component when a proportional numeric X axis is required.

```templ
@scatter.Scatter(scatter.Config{
	Label:      "Scatter series by day",
	Categories: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
	Options:    scatter.Options{Size: 4},
	Series: []scatter.Series{
		{
			Name: "Email",
			Points: []scatter.Point{
				{Category: "Mon", Value: 120},
				{Category: "Tue", Value: 132},
				{Category: "Wed", Value: 101},
				{Category: "Fri", Value: 90},
				{Category: "Sat", Value: 230},
				{Category: "Sun", Value: 210},
			},
		},
		{
			Name: "Union Ads",
			Points: []scatter.Point{
				{Category: "Mon", Value: 220},
				{Category: "Tue", Value: 182},
				{Category: "Wed", Value: 191},
				{Category: "Thu", Value: 234},
				{Category: "Fri", Value: 290},
				{Category: "Sat", Value: 330},
				{Category: "Sun", Value: 310},
			},
		},
	},
})
```

Marker symbol and size are typed options on the same component. Keep exact
points in nearby text or a table when readers need them.

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
