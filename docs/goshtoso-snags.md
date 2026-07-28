# Goshtoso integration snags

## 2026-07-27: templ control flow must remain structurally separate

The first `bar.templ` draft placed `@chart` and `if cfg.Caption != ""` on one
line inside `<figure>`. `templ generate` rejected it with an unterminated `if`.
Keep templ control flow on its own indented lines around nested components, then
regenerate the checked-in `*_templ.go` artifact.

## 2026-07-27: templ runtime import

`line.templ` initially declared `import "github.com/a-h/templ"`. `templ generate`
already injects that import when the template refers to `templ.Component`, so Go
failed with `templ redeclared in this block`. Consumer and extension templates
must use templ helpers without declaring that runtime import themselves.

## 2026-07-27: consumer utility coverage

The released embedded Goshtoso stylesheet does not include every Tailwind
utility an application shell might use (for example, `lg:flex-col`). Keep the
component primitives on Goshtoso utilities and tokens, but put demo/application
layout rules in the consumer's own stylesheet rather than assuming an arbitrary
Tailwind class is emitted.

## 2026-07-27: component docs sidebar follow-through

The searchable grouped demo sidebar used only Goshtoso's bundled Alpine
dependency and application-owned CSS. `templ generate` regenerated both page
artifacts cleanly; no new Goshtoso component, helper, or generated-template
friction appeared beyond the existing consumer utility-coverage constraint.

## 2026-07-27: shell boundary correction

The first extraction treated a component documentation shell as a generic
catalog shell. Those are different navigation and content patterns. The public
API was renamed to `componentdocshell` before release, repeated component-page
composition moved to `componentpage`, and `catalogshell` remains unclaimed.

## 2026-07-28: component preview overlay revision

The Charts site had pinned `goshtoso-app-shells` before the merged rounded
preview overlay fix. Production review requires the shell-owned preview
outline, rather than an application-local border workaround, so consumer sites
must update to the revision containing `3ca565d` and verify the transparent
layout border plus `::after` outline in light and dark themes.

## 2026-07-28: v11 consumer-brand boundary

`componentdocshell.Brand` accepts a custom logo component and favicon URL, but
does not serve consumer-owned identity assets. The Charts site therefore mounts
the approved v11 SVGs itself at `/brand/` and supplies the logo surface, ink,
and signal variables for Goshtoso's existing `.dark` state. This keeps the
shell reusable and preserves the upstream SVG geometry unchanged.

## 2026-07-28: heatmap category-axis validation gap

In go-echarts v2.7.2, `charts.HeatMap.Validate` does not call
`RectChart.Validate`, so category data supplied through `SetXAxis` is omitted
from the rendered option. The HeatMap component writes the validated X and Y
category data to `XAxisList[0].Data` and `YAxisList[0].Data` directly. The
renderer-specific workaround is covered by `TestHeatMapRendersCartesianChart`.

## 2026-07-28: funnel ordering lacks a dedicated option helper

go-echarts v2.7.2 exposes no dedicated funnel-series option for ordering. The
Funnel component keeps ordering typed as `FunnelOrder` and applies the validated
value through the exported `charts.SingleSeries.Sort` field. Rendering tests
cover descending, ascending, and caller-data order.

## 2026-07-28: tree zero depth and responsive width need explicit mapping

go-echarts v2.7.2 omits `initialTreeDepth: 0` because its integer field uses
`omitempty`, even though zero means roots-only and differs from an unspecified
renderer default. The Tree component uses a private serialization sentinel and
replaces it before emitting browser code. Its renderer also defaults the chart
host to 900px, wider than the documentation preview at common desktop widths;
the demo therefore sets the public, renderer-neutral `Width` to `100%`. This
matches the official `go-echarts/examples` Tree sample, which uses a 100% chart
inside the default centered page layout. Its flex layout targets wrapping
multiple charts, while none deliberately supplies no page layout; neither
belongs in this single-component wrapper.
