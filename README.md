# Goshtoso Charts

SSR SVG charts for Go applications using [Goshtoso](https://github.com/araihu/goshtoso). Initial product driver: monitor detail and public status pages in Xisnove. Public API remains product-neutral.

`goshtoso-charts` renders chart markup during the Go request. No Node process, browser chart runtime, hydration, CDN asset, or runtime fetch is required.

## Install

Requires Go 1.26.5+ and a Goshtoso-integrated templ application.

```bash
go get github.com/araihu/goshtoso-charts
```

## Component contract

Every chart follows Goshtoso component shape: a typed `Config`, concrete `Instance`, stable `Kind()`, and `templ.Component` rendering. Extension-owned kinds live in `github.com/araihu/goshtoso-charts/components`; they deliberately do not modify Goshtoso's core kind registry.

Charts use Goshtoso semantic CSS tokens (`--color-success`, `--color-warning`, `--color-danger`, and neutral outline) so all Goshtoso themes remain in control.

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

Line charts render their SVG geometry on the server but resolve their surface, outline, text, and series colors from Goshtoso semantic CSS tokens at display time. They therefore follow every built-in Goshtoso theme and `.dark` mode without a browser chart renderer or a fresh response.

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
