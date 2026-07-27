# Goshtoso Charts

SSR SVG charts for Go applications using [Goshtoso](https://github.com/araihu/goshtoso).

`goshtoso-charts` renders chart markup during the Go request. No Node process, browser chart runtime, hydration, CDN asset, or runtime fetch is required.

## Install

Requires Go 1.26.5+ and a Goshtoso-integrated templ application.

```bash
go get github.com/araihu/goshtoso-charts
```

## First chart

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

Render `line.ThemeGoshtosoDark` when the server selects dark mode. The initial component does not observe client-side theme changes, by design: its SVG is fully server-rendered.

`components/line` is first v1 surface. See [chart-library evaluation](docs/chart-library-evaluation.md) and [surface brief](docs/surface-brief.md).

## Development

```bash
templ generate
go test ./...
git diff --exit-code
```
