# Surface brief

Primary user and task: Go/templ application developer rendering monitor/status evidence without browser chart runtime; initial consumer is Xisnove.

Usage scene and constraints: SSR page; must remain meaningful before JavaScript and work under restrictive runtime policies.

Product register: product. Archetype: dashboard supporting evidence region.

Information priority: application task and exact data remain primary; chart summarizes change.

Navigation model: document links. Consequential states: empty, invalid data, server-render failure.

Existing identity: Goshtoso design system. Density: standard. Motion: none.

Visual direction: neutral figure, semantic series colors, restrained grid, no elevation or dashboard-card gallery.

Chosen primitives: extension-owned static/vector `line`, `bar`, `pie`, and
categorical `scatter`
figures plus renderer-neutral interactive components. Goshtoso `panel.Panel` is
supplied by the consumer. Availability history is a Bar-chart example, not a
domain-specific component.

Invariant ledger: none; chart rendering is read-only. Consumer must retain a textual/table alternative for exact values and no-data state.
