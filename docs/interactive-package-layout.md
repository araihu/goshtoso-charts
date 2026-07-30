# Interactive package migration

Interactive chart families are moving to discoverable chart-specific packages.
Phase 1 is additive: existing code under `components/interactive` remains
supported while canonical child packages are introduced one chart at a time.

## Bar

Use `components/interactive/bar` for new interactive Bar code:

```go
import (
	interactive "github.com/araihu/goshtoso-charts/components/interactive"
	interactivebar "github.com/araihu/goshtoso-charts/components/interactive/bar"
)

chart := interactivebar.Bar(interactivebar.Config{
	Label: "Weekly deployments",
	XAxis: []string{"Mon", "Tue"},
	Series: []interactivebar.Series{{
		Name: "Production",
		Data: []interactivebar.Data{{Value: 3}, {Value: 5}},
	}},
	Options: interactive.ChartOptions{
		Animation: interactive.Bool(false),
	},
})
```

The chart-named `Bar` constructor matches other Goshtoso component packages.
Chart-specific names are concise: `Config`, `Series`, `Data`, `Orientation`,
`Zoom`, `Statistic`, `Coordinate`, and reference types. Shared options such as
`interactive.ChartOptions`, `interactive.SeriesOptions`, and
`interactive.LiveData` remain available through exact parent-package aliases
during phase 1. Their canonical ownership now lives in `components/chart`, and
they appear unchanged through the aliased `Config` and `Series` fields.

Existing `interactive.Bar(interactive.BarConfig{...})` code still compiles and
renders identically. The child declarations are type aliases and a forwarding
constructor, so old and new configurations retain the same type identity,
validation, component kind, and markup semantics. No bulk rewrite is required.

Phase 1 does not complete chart-specific ownership. Bar implementation still
lives in the parent package; shared public types now live in `components/chart`
and private renderer/runtime behavior in `components/internal/interactive`.

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
3. `components/interactive` may depend on the internal implementation and the
   chart foundation, but never on a child package.
4. Child packages may temporarily forward to the parent facade. Future child
   implementation moves must preserve the same inward direction rather than
   making the internal package depend on a facade.

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
	interactive "github.com/araihu/goshtoso-charts/components/interactive"
	interactiveline "github.com/araihu/goshtoso-charts/components/interactive/line"
)

chart := interactiveline.Line(interactiveline.Config{
	Label: "Weekly latency",
	XAxis: []string{"Mon", "Tue"},
	Series: []interactiveline.Series{{
		Name: "p95 (ms)",
		Data: []interactiveline.Data{{Value: 42}, {Value: 47}},
	}},
	Options: interactive.ChartOptions{
		Animation: interactive.Bool(false),
	},
})
```

Line-specific names are concise: `Config`, `Series`, `Data`, `TimeAxis`,
`ValueAxis`, `VisualScale`, `VisualPiece`, `Statistic`, `Coordinate`, and
reference types. Shared options stay available through exact parent aliases
during phase 1 while canonical ownership lives in `components/chart`. Existing
`interactive.Line(interactive.LineConfig{...})` code remains supported and
assignment-compatible with the canonical package. Constructor behavior,
validation, component kind, live data, time and value axes, visual scales,
references, and rendered markup remain identical.

## Phase 1 boundary

Bar and Line are the first canonical child packages. Remaining chart families
continue using their existing paths until their own bounded migrations land.
Physical chart implementation ownership and the final fate of the parent
compatibility facade remain deliberately deferred decisions that must be
settled before v1. Shared public and private foundation ownership is fixed by
[ADR 0001](decisions/0001-interactive-chart-package-ownership.md).
