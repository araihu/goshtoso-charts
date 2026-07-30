# ADR 0001: Interactive chart package ownership

- Status: Accepted
- Date: 2026-07-30
- Scope: interactive chart package migration before v1

## Context

`components/interactive` historically combined shared public values, renderer
conversion, browser runtime markup, chart-specific implementations, and legacy
import paths. The first Bar and Line child packages began as compatibility
facades. Moving their implementations required an acyclic shared home because
the parent must continue to re-export child APIs.

Static chart packages use different rendering, sizing, export, and lifecycle
semantics. A common `Instance` boundary does not make their option models or
implementations identical.

## Decision

Public `components/chart` owns renderer-neutral `Instance` and only shared
types whose current field, zero-value, and behavioral semantics are identical.
This foundation moves interactive shared options and live snapshot values there.
Static configuration types remain in their chart-family packages.

`chart.NewInstance(components.Component)` is an intentional extension point.
It permits custom and external components without exposing an engine or private
adapter. Nil and typed-nil components produce the safe zero instance. Wrapping
an existing `chart.Instance` is idempotent. Delegate kind, rendering, and errors
pass through unchanged.

Private `components/internal/interactive` owns renderer conversion, render
configuration, the generated templ shell, theme runtime, and live runtime. No
type from this package is re-exported by a public API. Renderer dependencies stay
behind this boundary.

Parent `components/interactive` remains the compatibility surface. Its
`Instance` and shared types are exact aliases to `components/chart`; functions
retain their existing signatures. Bar and Line chart-specific types, validation,
rendering, generated templates, and exact-value details now live in their
canonical child packages. The parent exposes exact aliases and forwarding
constructors for those migrated families while retaining the remaining chart
implementations.

The dependency direction is:

```text
components
  <- components/chart
  <- components/internal/interactive
  <- components/interactive/bar and components/interactive/line
  <- components/interactive (compatibility facade and remaining implementations)
```

Bar and Line children depend only on `components/chart`,
`components/internal/interactive`, and other inward dependencies; they never
import the parent. The parent may import exactly those migrated children and
must alias or forward their public APIs. Later chart children follow the same
bounded move. Neither shared package may import the parent or a chart child.

## Compatibility

Existing constructors, field literals, assignments, and imports continue to
compile. Exact aliases preserve type identity between parent shared names and
`components/chart` names and between the parent Bar/Line names and their child
declarations. Rendered markup, runtime bytes, local asset paths, validation
behavior, and all 24 constructor signatures remain unchanged.

One pre-v1 reflection limit is explicit: shared named types now report
`github.com/araihu/goshtoso-charts/components/chart` as their `reflect.Type`
package path and appear there in generated API documentation. Source and
assignment compatibility do not hide that ownership correction.

The zero `interactive.Instance` now has exact `chart.Instance` identity. Its
`Kind` remains empty and its `Render` error remains `interactive chart label is
required`.

## Phased sequence

1. Land shared `chart.Instance`, common public types, private renderer adapters,
   compatibility aliases, and dependency/API contracts.
2. Move one chart-family implementation at a time into its canonical child;
   keep parent aliases and constructor forwarders.
3. Integrate completed chart migrations in bounded pairs, preserving separate
   static and interactive ownership.
4. Freeze v1 import paths only after every child owns its chart-specific API and
   implementation and the parent facade has a documented long-term policy.

GSC-TD-004 remains open until that sequence completes.

## Consequences

- Child migrations gain an acyclic shared home and one private renderer adapter.
- Bar and Line now physically own their chart-specific implementations while
  the parent facade remains supported.
- External chart components can return the same concrete `chart.Instance`.
- Parent imports remain valid while canonical ownership moves incrementally.
- Static and interactive options are not forced into a false common model.
- Reflection-based consumers must accept the documented pre-v1 package-path
  correction for shared types.
