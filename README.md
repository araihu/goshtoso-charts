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

## Roadmap

Foundation: `heartbeat` and `line`. Next generic primitives: bar/stacked bar, area, distribution/histogram, and categorical status breakdown. Add each only with a real monitor/status-page use case, stable kind, SSR output, no-data behavior, semantic-token palette, and focused tests.

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
