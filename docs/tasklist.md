# Task list

## Completed: catalog parity

- Extracted the Goshtoso demo frame into the public
  `github.com/araihu/goshtoso-app-shells/catalogshell` module and adopted it
  here. The shell owns the full-width header, search-first grouped sidebar,
  theme and dark controls, responsive drawer, main scroll, and page TOC.
- Kept the sidebar beside content from 720px upward and switched to a drawer
  below that breakpoint; chart rendering remains server-side.
- Added concise Heartbeat and Line component contracts: purpose, primitive,
  `Kind`, configuration dimensions, and accessibility contract.

## Chart primitives

- Add categorical and distribution primitives after their Xisnove projections
  have a concrete consumer need; retain the generic typed-Config, Instance,
  `Kind`, and SSR contract.
