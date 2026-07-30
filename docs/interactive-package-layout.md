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
`interactive.LiveData` remain parent-owned during phase 1 and appear unchanged
through the aliased `Config` and `Series` fields.

Existing `interactive.Bar(interactive.BarConfig{...})` code still compiles and
renders identically. The child declarations are type aliases and a forwarding
constructor, so old and new configurations retain the same type identity,
validation, component kind, and markup semantics. No bulk rewrite is required.

Phase 1 does not complete package ownership. Bar implementation and common
interactive types still live in the parent package; remaining chart families
continue using their existing paths until their own bounded migrations land.
