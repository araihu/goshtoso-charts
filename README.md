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
`--color-chart-series-*` token. Heartbeat keeps semantic status colors by
default; its `Colors` order is up, degraded, down, unknown. Interactive canvas
charts resolve these CSS tokens through the private runtime bridge and refresh
when Goshtoso theme or dark-mode state changes. Explicit `Style.Colors` remain
authoritative.

## Heartbeat

Heartbeat is the first monitoring primitive. It shows ordered availability observations, keeps status text in SVG titles, provides an accessible summary, and has a no-data state. It has no Xisnove dependency.

It accepts up to 300 points. Aggregate or bucket denser monitoring streams at application boundary; never silently discard recent failures in presentation code.

```templ
package monitor

import (
	"time"

	"github.com/araihu/goshtoso-charts/components/heartbeat"
)

templ Availability() {
	@heartbeat.Heartbeat(heartbeat.Config{
		Label:   "API availability over last three checks",
		Caption: "Last three checks",
		Points: []heartbeat.Point{
			{At: time.Date(2026, time.July, 27, 12, 0, 0, 0, time.UTC), State: heartbeat.StateUp, Latency: 42 * time.Millisecond},
			{At: time.Date(2026, time.July, 27, 12, 1, 0, 0, time.UTC), State: heartbeat.StateDegraded},
			{At: time.Date(2026, time.July, 27, 12, 2, 0, 0, time.UTC), State: heartbeat.StateDown},
		},
	})
}
```

For Xisnove, map retained daily uptime or raw probe-result rows into `heartbeat.Point` in Xisnove's view-model layer. Keep application/domain types out of this module.

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

## Interactive components

Interactive components expose only typed Goshtoso Charts configs. Renderer
types and callbacks are implementation details and cannot cross the public API.

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

## Roadmap

Foundation: `heartbeat`, `line`, `bar`/stacked bar, and `pie`. Next generic primitives: area and distribution/histogram. Add each only with a real monitor/status-page use case, stable kind, SSR output, no-data behavior, semantic-token palette, and focused tests.

See [chart-library evaluation](docs/chart-library-evaluation.md), [surface brief](docs/surface-brief.md), and [Xisnove heartbeat brief](docs/xisnove-heartbeat.md).

## Development

```bash
templ generate
go test ./...
git diff --exit-code
```

## Demo site

`site/` is a separate consumer module, shaped after the Goshtoso component demo catalog. It mounts Goshtoso assets at `/assets/` and uses `head.Dependencies()`.

```bash
cd site
GOWORK=off go run ./cmd/server
```

Open `http://localhost:8091`.
