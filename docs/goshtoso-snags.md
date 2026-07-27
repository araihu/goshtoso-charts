# Goshtoso integration snags

## 2026-07-27: templ runtime import

`line.templ` initially declared `import "github.com/a-h/templ"`. `templ generate`
already injects that import when the template refers to `templ.Component`, so Go
failed with `templ redeclared in this block`. Consumer and extension templates
must use templ helpers without declaring that runtime import themselves.

## 2026-07-27: consumer utility coverage

The released embedded Goshtoso stylesheet does not include every Tailwind
utility a catalog layout might use (for example, `lg:flex-col`). Keep the
component primitives on Goshtoso utilities and tokens, but put demo/application
layout rules in the consumer's own stylesheet rather than assuming an arbitrary
Tailwind class is emitted.

## 2026-07-27: catalog sidebar follow-through

The searchable grouped demo sidebar used only Goshtoso's bundled Alpine
dependency and application-owned CSS. `templ generate` regenerated both page
artifacts cleanly; no new Goshtoso component, helper, or generated-template
friction appeared beyond the existing consumer utility-coverage constraint.
