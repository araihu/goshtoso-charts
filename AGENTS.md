# Repository Policy

## Example Sources

- Use [go-echarts examples](https://github.com/go-echarts/examples/tree/master/examples) as the authoritative source for interactive chart examples.
- Use [go-analyze/charts examples](https://github.com/go-analyze/charts/tree/main/examples) as the authoritative source for static and vector chart examples.
- Before implementing any new component or example, identify a concrete upstream example file or path. Record that path centrally in tests or docs; do not repeat backing-engine branding on every component page.
- Preserve the upstream dataset, hierarchy or shape, chart semantics, variants, and visual intent. Correct obvious upstream typos or defects instead of reproducing them.
- Adapt only the renderer-neutral typed Goshtoso API, theme tokens, accessibility, responsive or layout shell, and idiomatic Go and templ structure.
- Do not invent infrastructure, operations, or domain framing when the upstream example is generic. Keep examples simple; do not build a catalog clone.
- Follow the Goshtoso component model: behavior variants that share semantics remain options on one component. Do not split catalog entries by engine or example variant.
- Keep public APIs and component pages backing-engine neutral. Centralize engine and project attribution.
- Use go-echarts `center`, `flex`, and `none` page-layout examples as layout references. Consumers own page layout; charts own renderer sizing.
- Keep static and interactive tasks under separate ownership when worked in pairs.

## Control Plane

- Keep two implementation tasks active at a time.
- Merge each completed pair through a fresh integration task.
- Archive bounded tasks after completion.
- The control-plane task coordinates work and does not implement source changes.
