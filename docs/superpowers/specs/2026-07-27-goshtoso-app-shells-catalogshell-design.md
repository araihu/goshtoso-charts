# Goshtoso App Shells: Catalog Shell Design

Date: 2026-07-27

## Summary

Create the public Go module `github.com/araihu/goshtoso-app-shells` as a home
for reusable application-shell patterns built from Goshtoso primitives. Its
first exported package, `catalogshell`, extracts the proven shell from the
Goshtoso demo site for component catalogs, API references, design systems, and
product documentation.

The first release must preserve the Goshtoso demo site's visual identity and
interaction model closely enough that Goshtoso itself can adopt the module
without a material visual regression. Goshtoso Charts becomes the second
consumer. The module provides presentation and progressive enhancement; each
application keeps ownership of routes, content, domain data, and policy.

## Goals

- Provide one reusable implementation of the Araihu catalog/documentation
  shell instead of copying templates and application CSS between repositories.
- Match the Goshtoso demo header, navigation, responsive behavior, content
  frame, optional table-of-contents rail, theme controls, and interaction
  rhythm near-identically.
- Keep full-page SSR functional without JavaScript.
- Use Alpine and HTMX only for progressive enhancement.
- Ship deterministic shell CSS and JavaScript using Goshtoso semantic tokens.
- Support future shell patterns without committing the repository to the
  `catalogshell` package name forever.

## Non-goals

- Moving Goshtoso components into this module.
- Owning consumer HTTP routes, content models, authorization, storage-consent
  policy, analytics, or domain state.
- Providing a client-side documentation framework or browser renderer.
- Generalizing hypothetical dashboard, settings, or marketing shells before a
  concrete consumer exists.
- Requiring HTMX fragment navigation for basic use.

## Repository and package structure

The repository is named broadly so later proven patterns can coexist without
renaming the module.

```text
github.com/araihu/goshtoso-app-shells
├── catalogshell/
│   ├── config.go
│   ├── layout.templ
│   ├── fragment.templ
│   ├── validation.go
│   └── assets/
├── example/
├── docs/
└── go.mod
```

Only `catalogshell` is part of the first release. Future package names such as
`dashboardshell` require an accepted consumer projection and their own design.

## Public API

`catalogshell.Config` describes shell-wide presentation:

- brand name, home URL, and logo component;
- Goshtoso theme options and initial theme behavior;
- top-level sidebar items and grouped sidebar sections;
- optional header actions and repository link;
- optional footer component;
- optional HTMX navigation enhancement;
- asset URL prefix, with a stable default.

Navigation uses Goshtoso sidebar types where their contract already fits. The
package introduces no parallel component system.

`catalogshell.Page` describes one response:

- title and description metadata;
- active navigation key;
- rendered `templ.Component` content;
- optional head content;
- optional table-of-contents behavior.

The primary renderers are:

```go
catalogshell.Layout(config, page) templ.Component
catalogshell.Fragment(config, page) templ.Component
catalogshell.Head(config) templ.Component
```

`Layout` renders a complete HTML document. `Fragment` renders the main content
and any required out-of-band navigation state for HTMX. `Head` emits Goshtoso
dependencies and shell assets for consumers that need to compose the document
head themselves.

Embedded assets are exposed through:

```go
catalogassets.Handler() http.Handler
```

Consumers mount that handler at the documented stable prefix. The example app
demonstrates the exact mount contract.

## Ownership boundary

The shell owns:

- full-width branded header;
- theme selector and dark-mode control;
- persistent desktop sidebar and mobile drawer;
- grouped navigation search and empty state;
- main scroll region and content frame;
- optional right-side table-of-contents rail;
- active-navigation presentation;
- focus management, reduced-motion behavior, and shell landmarks;
- HTMX fragment-navigation wiring, sidebar scroll restoration, and main-scroll
  reset when enhancement is enabled;
- shell CSS and small shell-specific JavaScript.

The consumer owns:

- HTTP routes and response selection;
- page and navigation data;
- application content and domain state;
- metadata values and canonical URLs;
- branding values and optional action components;
- storage consent, analytics, authentication, and other policy;
- application-specific CSS outside the shell boundary.

## Rendering and interaction flow

For an ordinary request, the consumer maps its route and domain data into a
`catalogshell.Page` and renders `Layout`. All navigation remains usable through
normal links.

When HTMX enhancement is enabled, shell navigation targets the stable main
content element. The consumer renders `Fragment` for an HTMX request. The
fragment updates page content and active navigation state, restores sidebar
scroll position, resets the main content scroll position, rebuilds the optional
TOC, and moves focus to the main content heading. A failed fragment request
leaves ordinary link navigation available.

Sidebar search filters only navigation. It does not fetch or render page
content. Theme and dark-mode state update Goshtoso's existing document-level
contracts; persistence remains configurable so consumers with consent policies
can disable storage.

## Responsive contract

- At widths of 720px and above, the left sidebar remains visible beside main
  content. This includes the 736px Codex in-app browser viewport.
- Below 720px, navigation becomes an off-canvas drawer with a labeled trigger,
  backdrop, Escape handling, focus return, and scroll containment.
- The right TOC rail appears only at a wide breakpoint and only when page
  headings justify it.
- Header controls remain usable without document-level horizontal overflow.
- The content frame owns horizontal scrolling for code and wide examples;
  neither chart previews nor code blocks may widen the document.

## Visual contract

The Goshtoso demo is the reference implementation. The catalog shell must
preserve its structure and visual rhythm, including:

- header height, full-width alignment, brand treatment, and control placement;
- sidebar width, borders, section labels, active state, and search placement;
- page padding, maximum content width, typography, and vertical rhythm;
- surface, outline, primary, muted, and dark-mode semantic tokens;
- mobile drawer and overlay behavior;
- restrained transitions with reduced-motion support.

Differences require an explicit accessibility, API, or responsive rationale;
they must not result from missing utility classes or copied partial CSS.

## Validation and failure behavior

Rendering validates required configuration. Errors identify the exact invalid
field or navigation entry. Validation covers:

- missing brand name or home URL;
- missing page content or title;
- empty or duplicate navigation IDs;
- active navigation keys absent from configured navigation;
- invalid asset prefixes;
- incompatible fragment configuration.

Optional slots render nothing when absent. Asset requests use normal HTTP
status behavior. JavaScript enhancement failures must not prevent full-page
navigation, theme-readable content, or access to main content.

## Accessibility contract

- Provide a skip link and semantic header, navigation, main, complementary,
  and footer landmarks.
- Mark the active navigation item with `aria-current="page"`.
- Label theme, dark-mode, search, drawer, and close controls.
- Keep drawer focus inside while open, restore trigger focus on close, and
  support Escape.
- Move focus to meaningful main content after fragment navigation.
- Preserve visible focus indicators and reduced-motion preferences.
- Search results and empty state remain understandable without color.

## Verification

The module must pass:

- `templ generate` with no generated drift;
- `go test ./...`;
- `go vet ./...`;
- `go build ./...`;
- focused configuration-validation and semantic-markup tests;
- example-app HTTP tests for full-page and fragment responses;
- asset mount tests using the documented prefix;
- keyboard checks for skip link, search, theme controls, and mobile drawer;
- visual smoke at 1440px, 736px, and 390px in Goshtoso light, Goshtoso dark,
  and Minimal dark;
- no document-level horizontal overflow or console errors;
- no CDN runtime dependency.

Goshtoso demo adoption is the parity gate. Goshtoso Charts adoption is the
external-consumer gate. The first release is not complete until both consumers
use the module successfully.

## Adoption sequence

1. Create public repository `araihu/goshtoso-app-shells` and its Go module.
2. Implement `catalogshell` and a standalone example application.
3. Replace Goshtoso demo's local shell with `catalogshell` while preserving its
   catalog data, SEO, storage consent, and routes.
4. Compare Goshtoso demo before and after at required viewports and themes.
5. Replace Goshtoso Charts' copied shell with the released module.
6. Record integration snags in each consumer and refine only generic issues in
   the shell module.
7. Tag the first module release after both consumer gates pass.

## Release and compatibility policy

The module follows semantic versioning. Goshtoso dependency updates land and
release upstream first; `goshtoso-app-shells` then updates against the released
version. Downstream consumers use a pushed commit or tag and run checks with
`GOWORK=off` so local workspaces cannot hide dependency drift.

The first release stabilizes only `catalogshell`. New shell packages and new
cross-cutting behavior require concrete consumers. Visual changes to the
catalog shell require parity evidence against Goshtoso demo plus external
consumer verification in Goshtoso Charts.
