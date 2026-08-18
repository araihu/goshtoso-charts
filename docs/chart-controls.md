# Chart-control implementation evidence

The canonical consumer documentation is the site guide served at
`/docs/chart-controls`. It owns defaults, lifecycle modes, client events,
responsive actions, export capabilities, accessibility, HTMX behavior, and
failure guidance. The companion `/docs/chart-modes` guide owns the choice
between static/vector and interactive delivery. Component pages link both
guides through one shared footer helper instead of repeating this contract.

This repository file records implementation and verification boundaries only.
Do not add a second prose version of the public behavior here.

## Ownership

- `components/chartcontrol` defines the renderer-neutral wrapper, lifecycle,
  action, and export types.
- `assets/js/controls/6/controls.js` owns wrapper transitions, modal and
  fullscreen behavior, export requests, resize settlement, and idempotent DOM
  preparation. Immutable v1 through v4 asset paths remain compatibility
  outputs.
- Goshtoso `components/actiongroup` owns primary action, stacked action, and
  constrained overflow presentation.
- Interactive renderer export remains behind its private runtime; the common
  control runtime requests capture through a renderer-neutral event.
- The site guide and component footers are documentation concerns. They do not
  change component runtime or public Go API.

## Verification anchors

- `components/chartcontrol/control_test.go` covers lifecycle rendering,
  omission, action configuration, and validation.
- `assets/assets_test.go` covers stable runtime events, HTMX hooks, mutation
  observation, and immutable asset delivery.
- `site/internal/server/guides_test.go` covers the canonical guide contract.
- `site/internal/server/documentation_test.go` proves that every chart route
  uses one shared guide/API footer.
- `site/browser/chart_controls.test.cjs` covers state transitions, focus,
  responsive actions, fullscreen, modal expansion, export, HTMX swaps, theme,
  and renderer identity.
- `site/browser/guides.test.cjs` and
  `site/browser/documentation_footer.test.cjs` cover narrow/wide, light/dark
  guide and footer presentation.

## Source attribution boundary

Backing-library evidence and pinned upstream example paths remain centralized
in the attributions and upstream-coverage documents. Public component pages and
the two shared guides stay renderer-neutral.
