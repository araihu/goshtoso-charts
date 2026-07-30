# Interactive package migration

Interactive chart families are moving to discoverable chart-specific packages.
Phase 1 is additive: existing code under `components/interactive` remains
supported while canonical child packages are introduced one chart at a time.

## Bar

Use `components/interactive/bar` for new interactive Bar code:

```go
import (
	"github.com/araihu/goshtoso-charts/components/chart"
	interactivebar "github.com/araihu/goshtoso-charts/components/interactive/bar"
)

chart := interactivebar.Bar(interactivebar.Config{
	Label: "Weekly deployments",
	XAxis: []string{"Mon", "Tue"},
	Series: []interactivebar.Series{{
		Name: "Production",
		Data: []interactivebar.Data{{Value: 3}, {Value: 5}},
	}},
	Options: chart.ChartOptions{
		Animation: chart.Bool(false),
	},
})
```

The chart-named `Bar` constructor matches other Goshtoso component packages.
Chart-specific names are concise: `Config`, `Series`, `Data`, `Orientation`,
`Zoom`, `Statistic`, `Coordinate`, and reference types. Shared options such as
`chart.ChartOptions`, `chart.SeriesOptions`, and `chart.LiveData` live in
`components/chart`. Their parent-package names remain exact compatibility
aliases.

Existing `interactive.Bar(interactive.BarConfig{...})` code still compiles and
renders identically. The child owns the concrete chart-specific declarations,
validation, rendering, and template; the parent exposes exact aliases and a
forwarding constructor. Old and new configurations retain the same type
identity, validation, component kind, and markup semantics. No bulk rewrite is
required.

## Shared chart instance foundation

`components/chart.Instance` is the canonical public identity for a renderable
chart instance. `components/interactive.Instance` and all ten migrated child
packages' `Instance` names are exact aliases, so old and new values remain
assignment-compatible and constructor function signatures remain
interchangeable. Consumers can mix the facade and chart-specific package paths
without importing a rendering engine. `chart.NewInstance(components.Component)`
is the renderer-neutral extension point for custom components.

Foundation imports point inward:

1. `components/chart` depends only on public component contracts.
2. `components/internal/interactive` may depend on that foundation, but never
   on the parent facade or a chart-specific child package.
3. Migrated child packages may depend on the chart foundation and private
   interactive implementation, but never on the parent facade.
4. `components/interactive` may depend on exactly the migrated children while
   retaining implementations for unmigrated families. Future moves preserve
   that inward direction.

The ownership move has one intentional pre-v1 compatibility limit. Reflection
reports the canonical package path as `components/chart`, including for values
spelled as `interactive.Instance` or a child-package `Instance`; generated API
documentation also owns the declaration there. Source assignment, methods,
rendering, validation errors, and zero-value behavior remain compatible, but
consumers that assert the old reflected package path must update that assertion.

## Line

Use `components/interactive/line` for new interactive Line code:

```go
import (
	"github.com/araihu/goshtoso-charts/components/chart"
	interactiveline "github.com/araihu/goshtoso-charts/components/interactive/line"
)

chart := interactiveline.Line(interactiveline.Config{
	Label: "Weekly latency",
	XAxis: []string{"Mon", "Tue"},
	Series: []interactiveline.Series{{
		Name: "p95 (ms)",
		Data: []interactiveline.Data{{Value: 42}, {Value: 47}},
	}},
	Options: chart.ChartOptions{
		Animation: chart.Bool(false),
	},
})
```

Line-specific names are concise: `Config`, `Series`, `Data`, `TimeAxis`,
`ValueAxis`, `VisualScale`, `VisualPiece`, `Statistic`, `Coordinate`, and
reference types. Shared options live in `components/chart` and stay available
through exact parent aliases. Existing
`interactive.Line(interactive.LineConfig{...})` code remains supported and
assignment-compatible with the canonical package. The child owns the concrete
Line implementation and template; the parent is an alias-and-forwarder facade.
Constructor behavior, validation, component kind, live data, time and value
axes, visual scales, references, and rendered markup remain identical.

## Scatter

Use `components/interactive/scatter` for new interactive Scatter code:

```go
import (
	"github.com/araihu/goshtoso-charts/components/chart"
	interactivescatter "github.com/araihu/goshtoso-charts/components/interactive/scatter"
)

chart := interactivescatter.Scatter(interactivescatter.Config{
	Label:   "Throughput",
	Variant: interactivescatter.VariantEffect,
	XAxis:   []string{"Mon", "Tue"},
	Series: []interactivescatter.Series{{
		Name: "Requests",
		Data: []interactivescatter.Data{{Value: 42}, {Value: 47}},
		Ripple: &chart.RippleOptions{Period: 4, Scale: 6},
	}},
})
```

Standard and effect rendering remain variants of one `Scatter` component and
one component kind. Scatter-specific names are concise: `Config`, `Series`,
`Data`, `Variant`, and `AxisType`; shared options remain in `components/chart`.
Existing `interactive.Scatter(interactive.ScatterConfig{...})` code is an exact
alias-and-forwarder compatibility path with identical validation, kind, options,
and rendered runtime bytes.

## Candlestick

Use `components/interactive/candlestick` for new interactive Candlestick code:

```go
import interactivecandlestick "github.com/araihu/goshtoso-charts/components/interactive/candlestick"

chart := interactivecandlestick.Candlestick(interactivecandlestick.Config{
	Label:      "Daily prices",
	Categories: []string{"Mon"},
	Series: []interactivecandlestick.Series{{
		Name: "OHLC",
		Data: []interactivecandlestick.Candle{{
			Open: 10, Close: 11, Low: 9, High: 12,
		}},
	}},
})
```

`Candle` preserves the typed open, close, low, and high ordering. Other
chart-specific names are concise: `Config`, `Series`, `SeriesOptions`,
`DirectionStyle`, `MarkOptions`, and `DataZoom`. Shared options remain in
`components/chart`. The legacy parent names are exact aliases and its
`Candlestick` constructor is a single forwarder, preserving validation, kind,
options, OHLC semantics, and rendered runtime bytes.

## HeatMap

Use `components/interactive/heatmap` for new interactive HeatMap code:

```go
import (
	"github.com/araihu/goshtoso-charts/components/chart"
	interactiveheatmap "github.com/araihu/goshtoso-charts/components/interactive/heatmap"
)

chart := interactiveheatmap.HeatMap(interactiveheatmap.Config{
	Label:      "Deployment activity",
	XAxis:      []string{"Mon"},
	YAxis:      []string{"Morning"},
	ValueRange: interactiveheatmap.ValueRange{Min: 0, Max: 10},
	Series: []interactiveheatmap.Series{{
		Name: "Deployments",
		Data: []interactiveheatmap.Data{{Value: 3}},
	}},
	Options: chart.ChartOptions{Animation: chart.Bool(false)},
})
```

Cartesian and calendar coordinates remain variants of one `HeatMap` component
and one component kind. HeatMap-specific names are concise: `Config`,
`Coordinate`, `Calendar`, `ValueRange`, `Series`, and `Data`. Shared chart,
series, calendar, palette, and accessibility options retain their existing
behavior through `components/chart`. The parent keeps exact aliases and one
forwarding constructor.

## Pie

Use `components/interactive/pie` for new interactive Pie code:

```go
import interactivepie "github.com/araihu/goshtoso-charts/components/interactive/pie"

chart := interactivepie.Pie(interactivepie.Config{
	Label: "Deployment outcomes",
	Series: []interactivepie.Series{{
		Name:        "Outcome",
		InnerRadius: 40,
		OuterRadius: 75,
		Data: []interactivepie.Data{
			{Name: "Passed", Value: 8},
			{Name: "Failed", Value: 2},
		},
	}},
})
```

Standard, donut, and rose treatments remain variants of one `Pie` component
and one component kind. Pie-specific names are concise: `Config`, `Series`,
`Data`, `RoseMode`, `LabelContent`, `TooltipContent`, and `Center`. Percentage
formatting and validation live only in the private interactive adapter; thin
parent wrappers preserve Sunburst, ThemeRiver, and WordCloud behavior.
The public parent Pie surface remains exact aliases, constants, and one
constructor forwarder.

## Radar

Use `components/interactive/radar` for new interactive Radar code:

```go
import (
	"github.com/araihu/goshtoso-charts/components/chart"
	interactiveradar "github.com/araihu/goshtoso-charts/components/interactive/radar"
)

profile := interactiveradar.Radar(interactiveradar.Config{
	Label:      "Service profile",
	Indicators: []interactiveradar.Indicator{{Name: "Availability", Max: 100}},
	Series: []interactiveradar.Series{{
		Name: "Current",
		Data: []interactiveradar.Data{{Name: "Today", Values: []float64{99.9}}},
	}},
	Coordinate: interactiveradar.CoordinateOptions{
		Shape:     interactiveradar.ShapeCircle,
		SplitArea: chart.Bool(true),
	},
})
```

Default polygon, explicit polygon, and circular coordinates remain variants of
one `Radar` component and one component kind. Radar-specific names are concise:
`Config`, `Shape`, `CoordinateOptions`, `SplitLineOptions`, `Indicator`,
`Series`, and `Data`. Shared chart and series options remain in
`components/chart`. Parent Radar names and shape constants are exact aliases;
its constructor is one forwarder with identical validation and rendered markup.

## BoxPlot

Use `components/interactive/boxplot` for new interactive BoxPlot code:

```go
import (
	"github.com/araihu/goshtoso-charts/components/chart"
	interactiveboxplot "github.com/araihu/goshtoso-charts/components/interactive/boxplot"
)

distribution := interactiveboxplot.BoxPlot(interactiveboxplot.Config{
	Label:      "Request latency",
	Categories: []string{"Production"},
	Series: []interactiveboxplot.Series{{
		Name: "Milliseconds",
		Data: []interactiveboxplot.Data{{Min: 18, Q1: 31, Median: 42, Q3: 58, Max: 94}},
	}},
	Options: chart.ChartOptions{Animation: chart.Bool(false)},
})
```

Single-series, multiple-series, and point-level label, item-style, emphasis,
and tooltip overrides remain one `BoxPlot` component. BoxPlot-specific names
are `Config`, `Series`, and `Data`; shared options stay in `components/chart`.
The parent preserves three exact type aliases and one forwarding constructor,
including five-number ordering validation, component kind, and markup.

## Gauge

Use `components/interactive/gauge` for new interactive Gauge code:

```go
import interactivegauge "github.com/araihu/goshtoso-charts/components/interactive/gauge"

completion := interactivegauge.Gauge(interactivegauge.Config{
	Label:   "Deployment completion",
	Variant: interactivegauge.VariantProgress,
	Series: []interactivegauge.Series{{
		Name: "Rollout",
		Data: []interactivegauge.Data{{Name: "Complete", Value: 73}},
	}},
})
```

Standard, progress, and liquid treatments remain variants of one `Gauge`
component. Gauge-specific scale, progress, liquid-shape, paint, outline,
background, label, series, and data types live in the child package; shared
chart and series options remain in `components/chart`. The parent preserves
exact aliases, constants, and one forwarding constructor. Remaining parent
percentage helpers serve only unmigrated parent implementations; Gauge calls
the private adapter directly.

## Funnel

Use `components/interactive/funnel` for new interactive Funnel code:

```go
import interactivefunnel "github.com/araihu/goshtoso-charts/components/interactive/funnel"

pipeline := interactivefunnel.Funnel(interactivefunnel.Config{
	Label: "Deployment pipeline",
	Order: interactivefunnel.OrderData,
	Series: []interactivefunnel.Series{{
		Name: "Deployments",
		Data: []interactivefunnel.Data{
			{Name: "Started", Value: 10},
			{Name: "Completed", Value: 8},
		},
	}},
})
```

Descending, ascending, and caller-data ordering remain options on one `Funnel`
component. Funnel-specific `Config`, `Order`, `Series`, and `Data` live in the
child package; shared options remain in `components/chart`. The parent keeps
exact aliases, order constants, and one forwarding constructor.

## Current boundary

Bar, Line, Scatter, Candlestick, HeatMap, Pie, Radar, BoxPlot, Gauge, and Funnel
physical ownership is complete. The supported parent facade delegates those ten
migrated charts. Remaining chart families continue using their existing paths
until their own bounded migrations land, and the final v1 policy for the parent
compatibility facade remains open. Shared public and private foundation
ownership is fixed by
[ADR 0001](decisions/0001-interactive-chart-package-ownership.md).
