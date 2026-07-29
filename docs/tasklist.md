# Task list

## Completed: component documentation parity

- Extracted the Goshtoso demo frame into the public
  `github.com/araihu/goshtoso-app-shells/componentdocshell` module and adopted it
  here. The shell owns the full-width header, search-first grouped sidebar,
  theme and dark controls, responsive drawer, main scroll, and page TOC.
- Kept the sidebar beside content from 720px upward and switched to a drawer
  below that breakpoint; chart rendering remains server-side.
- Added concise component contracts covering purpose, primitive, `Kind`,
  configuration dimensions, and accessibility expectations. Availability is
  now an application example built from interactive Bar plus renderer-neutral
  SSE snapshots, not a public component taxonomy entry.
- Kept catalog/browse shells as a separate pattern. Component reference page
  composition lives in `componentpage`; it is not another shell.

## Chart primitives

- Added `bar` with grouped and stacked named series for categorical operations
  data, plus `pie` for bounded part-to-whole status distributions. Both retain
  typed config, concrete instance, stable kind, SSR SVG, semantic-token palette,
  accessible figure label/caption, and focused validation/render tests.
- Next: area and distribution/histogram, only when a concrete consumer need
  exists; retain the same public and SSR contracts.
