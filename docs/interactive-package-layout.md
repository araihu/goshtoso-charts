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
chart instance. `components/interactive.Instance` and the Bar and Line child
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

## Current boundary

Bar and Line physical ownership is complete. The supported parent facade
delegates those migrated charts. Remaining chart families continue using their
existing paths until their own bounded migrations land, and the final v1 policy
for the parent compatibility facade remains open. Shared public and private
foundation ownership is fixed by
[ADR 0001](decisions/0001-interactive-chart-package-ownership.md).
