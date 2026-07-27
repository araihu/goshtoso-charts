# Goshtoso Catalog Shell Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create `github.com/araihu/goshtoso-app-shells/catalogshell`, extract the Goshtoso demo shell into a flexible public module, and adopt it in Goshtoso and Goshtoso Charts.

**Architecture:** `catalogshell` renders complete SSR documents and HTMX fragments from typed `Config` and `Page` values. Embedded deterministic shell assets use Goshtoso semantic tokens; consumer repositories retain routes, content, data, SEO values, and policy.

**Tech Stack:** Go 1.26.5, templ v0.3.1020, Goshtoso v0.0.13+, Alpine.js and HTMX supplied by Goshtoso, embedded `net/http` assets.

## Global Constraints

- Public module path is `github.com/araihu/goshtoso-app-shells`.
- First exported package is `catalogshell`; no speculative additional shells.
- Goshtoso demo is the visual and behavioral source of truth.
- Full-page SSR works without JavaScript; Alpine and HTMX are progressive enhancement only.
- Shell has no application-specific imports or policy.
- Persistent left sidebar begins at 720px; below 720px use accessible off-canvas navigation.
- Required visual viewports: 1440px, 736px, and 390px.
- Required themes: Goshtoso light, Goshtoso dark, and Minimal dark.
- No document-level horizontal overflow, console errors, CDN-only runtime, or generated templ drift.
- Dependency-sensitive consumer checks run with `GOWORK=off`.

---

## File structure

- New repository `araihu/goshtoso-app-shells`
  - `go.mod`: public module and released Goshtoso dependency.
  - `catalogshell/config.go`: public shell/page types and defaults.
  - `catalogshell/validation.go`: deterministic validation.
  - `catalogshell/layout.templ`: full SSR document and shell structure.
  - `catalogshell/fragment.templ`: HTMX main/sidebar fragment contract.
  - `catalogshell/render.go`: exported render constructors.
  - `catalogshell/assets/assets.go`: embedded asset handler and URLs.
  - `catalogshell/assets/shell.css`: deterministic responsive layout.
  - `catalogshell/assets/shell.js`: focus, TOC, and HTMX enhancement.
  - `catalogshell/catalogshell_test.go`: API, validation, and semantic HTML tests.
  - `catalogshell/assets/assets_test.go`: asset handler tests.
  - `example/internal/pages/pages.templ`: standalone consumer pages.
  - `example/internal/server/server.go`: reference mount and routes.
  - `example/internal/server/server_test.go`: full/fragment route tests.
  - `example/cmd/server/main.go`: visual-smoke server.
  - `README.md`: installation, mount, and usage contract.
- Goshtoso repository
  - `site/internal/pages/demo/layout.templ`: compose `catalogshell` with existing policy and catalog data.
  - `site/internal/pages/demo/fragment.templ`: use public fragment contract.
  - `site/internal/server/server.go`: mount catalog shell assets.
  - existing site tests: preserve route, SEO, navigation, and fragment behavior.
- Goshtoso Charts repository
  - `site/go.mod`, `site/go.sum`: consume released/pushed shell module.
  - `site/internal/pages/layout.templ`: replace copied shell with `catalogshell` config.
  - `site/internal/siteassets/site.css`: remove shell-owned rules.
  - `site/internal/server/server.go`: mount catalog shell assets.
  - `site/internal/server/server_test.go`: verify mount and shell semantics.

### Task 1: Create public repository and module skeleton

**Files:**
- Create: repository `araihu/goshtoso-app-shells`
- Create: `go.mod`
- Create: `README.md`
- Create: `.gitignore`

**Interfaces:**
- Consumes: public GitHub organization `araihu` and Go 1.26.5.
- Produces: cloneable public repository with module path `github.com/araihu/goshtoso-app-shells`.

- [ ] **Step 1: Confirm repository does not already exist**

Run: `gh repo view araihu/goshtoso-app-shells`

Expected: not found. If it exists, inspect it and preserve all existing work.

- [ ] **Step 2: Create public repository and clone it**

Run: `gh repo create araihu/goshtoso-app-shells --public --clone --description "Reusable application shell patterns for Goshtoso"`

Expected: repository exists locally and on GitHub.

- [ ] **Step 3: Create module metadata**

```go
module github.com/araihu/goshtoso-app-shells

go 1.26.5

require (
	github.com/a-h/templ v0.3.1020
	github.com/araihu/goshtoso v0.0.13
)
```

- [ ] **Step 4: Verify empty module**

Run: `go mod tidy && go test ./...`

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum README.md .gitignore
git commit -m "chore: bootstrap Goshtoso app shells"
git push -u origin main
```

### Task 2: Define typed configuration and validation

**Files:**
- Create: `catalogshell/config.go`
- Create: `catalogshell/validation.go`
- Create: `catalogshell/config_test.go`

**Interfaces:**
- Consumes: `templ.Component`, `sidebar.Item`, `sidebar.Section`, `select.Option`.
- Produces: `Config`, `Brand`, `Page`, `Layout(Config, Page)`, and precise validation errors.

- [ ] **Step 1: Write failing validation tests**

```go
func TestValidateRejectsDuplicateNavigationIDs(t *testing.T) {
	cfg := validConfig()
	cfg.Navigation.Sections = []sidebar.Section{{Title: "Components", Items: []sidebar.Item{
		{ID: "line", Label: "Line", Href: "/line"},
		{ID: "line", Label: "Again", Href: "/again"},
	}}}
	err := validate(cfg, Page{Title: "Line", Active: "line", Content: templ.NopComponent})
	if err == nil || !strings.Contains(err.Error(), `duplicate navigation ID "line"`) {
		t.Fatalf("validate error = %v", err)
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run: `go test ./catalogshell -run TestValidate -count=1`

Expected: FAIL because public types and validation do not exist.

- [ ] **Step 3: Implement public types**

```go
type Config struct {
	Brand       Brand
	Navigation Navigation
	Themes      []selectfield.Option
	HeaderActions templ.Component
	Footer      templ.Component
	RepositoryURL string
	AssetPrefix string
	EnableHTMX bool
	PersistPreferences bool
}

type Page struct {
	Title string
	Description string
	CanonicalURL string
	Active string
	Content templ.Component
	Head templ.Component
	EnableTOC bool
}
```

- [ ] **Step 4: Implement recursive navigation validation**

Validate required brand/page fields, nonempty unique IDs, active-key presence,
absolute asset prefix, and fragment compatibility. Return field-specific errors.

- [ ] **Step 5: Run focused and package tests**

Run: `go test ./catalogshell -count=1`

Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add catalogshell/config.go catalogshell/validation.go catalogshell/config_test.go
git commit -m "feat(catalogshell): define shell configuration"
```

### Task 3: Ship deterministic embedded assets

**Files:**
- Create: `catalogshell/assets/assets.go`
- Create: `catalogshell/assets/assets_test.go`
- Create: `catalogshell/assets/shell.css`
- Create: `catalogshell/assets/shell.js`

**Interfaces:**
- Consumes: Goshtoso semantic CSS tokens and stable shell DOM IDs.
- Produces: `assets.Handler() http.Handler`, `assets.StylesheetURL(prefix string) string`, and `assets.ScriptURL(prefix string) string`.

- [ ] **Step 1: Write failing asset tests**

Assert `GET /catalogshell/assets/shell.css` and `shell.js` return 200 with correct
content types, while traversal and unknown paths return 404.

- [ ] **Step 2: Run tests to verify failure**

Run: `go test ./catalogshell/assets -count=1`

Expected: FAIL because handler does not exist.

- [ ] **Step 3: Implement embedded handler**

Use `//go:embed shell.css shell.js`, exact path dispatch, `nosniff`, and cacheable
responses. The handler receives full `/catalogshell/assets/` paths without a
consumer-side `StripPrefix`.

- [ ] **Step 4: Extract shell CSS from Goshtoso demo structure**

Implement full-height frame, 18rem sidebar, 4rem header, fixed/sticky regions,
720px drawer breakpoint, wide TOC breakpoint, overflow containment, focus,
reduced motion, and Goshtoso tokens. Do not depend on unproven emitted utilities.

- [ ] **Step 5: Extract small enhancement script**

Implement TOC construction, fragment focus, sidebar scroll preservation, main
scroll reset, and HTMX event integration. No domain data or policy.

- [ ] **Step 6: Run tests**

Run: `go test ./catalogshell/assets -count=1`

Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add catalogshell/assets
git commit -m "feat(catalogshell): embed shell assets"
```

### Task 4: Implement full document and fragment renderers

**Files:**
- Create: `catalogshell/layout.templ`
- Create: `catalogshell/fragment.templ`
- Create: `catalogshell/render.go`
- Create: `catalogshell/catalogshell_test.go`
- Generate: `catalogshell/layout_templ.go`
- Generate: `catalogshell/fragment_templ.go`

**Interfaces:**
- Consumes: Task 2 `Config`/`Page`; Task 3 asset URLs; Goshtoso `head`, `select`, and `sidebar` components.
- Produces: complete `Layout`, composable `Head`, and HTMX `Fragment` components.

- [ ] **Step 1: Write failing semantic markup tests**

Render a valid page and assert doctype, skip link, full-width header, labeled
theme control, dark-mode control, desktop sidebar, mobile trigger, grouped
navigation, `aria-current="page"`, main landmark, TOC rail, and shell asset URLs.

- [ ] **Step 2: Run tests to verify failure**

Run: `go test ./catalogshell -run 'TestLayout|TestFragment' -count=1`

Expected: FAIL because renderers do not exist.

- [ ] **Step 3: Port Goshtoso demo shell into typed template**

Use Goshtoso primitives for theme selector and sidebar. Preserve normal link
fallback. Render consumer slots only when non-nil. Apply active state without
mutating caller-owned slices.

- [ ] **Step 4: Implement fragment contract**

Render `#main-content` plus out-of-band `#catalogshell-sidebar-content` with the
same active state. Include no complete document tags.

- [ ] **Step 5: Generate and verify**

Run: `templ generate && go test ./catalogshell -count=1 && git diff --check`

Expected: PASS and generated files match templates.

- [ ] **Step 6: Commit**

```bash
git add catalogshell
git commit -m "feat(catalogshell): render reusable catalog frame"
```

### Task 5: Add standalone consumer example and documentation

**Files:**
- Create: `example/internal/pages/pages.templ`
- Create: `example/internal/server/server.go`
- Create: `example/internal/server/server_test.go`
- Create: `example/cmd/server/main.go`
- Modify: `README.md`

**Interfaces:**
- Consumes: Tasks 2-4 public APIs.
- Produces: runnable reference at `go run ./example/cmd/server` and copyable integration docs.

- [ ] **Step 1: Write failing server tests**

Assert `/`, `/components/button`, `/assets/styles.css`, and
`/catalogshell/assets/shell.css` return 200. Assert HTMX request returns a
fragment without `<html>` and with OOB active navigation.

- [ ] **Step 2: Run test to verify failure**

Run: `go test ./example/internal/server -count=1`

Expected: FAIL because example server does not exist.

- [ ] **Step 3: Implement minimal example**

Mount `assets.Handler()` at `GET /assets/`, catalog shell assets at
`GET /catalogshell/assets/`, and two generic SSR pages using shared config.

- [ ] **Step 4: Document consumer contract**

README includes install, asset mounts, config, full-page rendering, HTMX
fragment detection, theme persistence choice, and accessibility responsibilities.

- [ ] **Step 5: Run repository gates**

Run: `templ generate && go test ./... && go vet ./... && go build ./...`

Expected: PASS.

- [ ] **Step 6: Commit and push**

```bash
git add README.md example
git commit -m "docs: add catalog shell consumer example"
git push
```

### Task 6: Dogfood in Goshtoso demo

**Files:**
- Modify: `site/go.mod`, `site/go.sum`
- Modify: `site/internal/pages/demo/layout.templ`
- Modify: `site/internal/pages/demo/fragment.templ`
- Modify: `site/internal/server/server.go`
- Modify: existing focused site tests

**Interfaces:**
- Consumes: pushed `goshtoso-app-shells` commit and existing Goshtoso catalog/policy helpers.
- Produces: Goshtoso demo rendered through `catalogshell` with preserved behavior.

- [ ] **Step 1: Create isolated Goshtoso worktree**

Create branch `codex/adopt-catalogshell` from current `origin/main`; stop if
another task has overlapping uncommitted shell changes.

- [ ] **Step 2: Write/adjust parity tests before migration**

Lock header controls, sidebar catalog, search modal, fragment OOB response, SEO,
storage consent, and existing route behavior.

- [ ] **Step 3: Update dependency to pushed module commit**

Run: `GOWORK=off go get github.com/araihu/goshtoso-app-shells@main` in `site/`,
then record the resolved pseudo-version from `go list -m -json`.

Expected: `go.mod` resolves remote module, not local workspace.

- [ ] **Step 4: Replace local frame with catalog shell composition**

Retain Goshtoso-specific catalog data, SEO metadata, theme inventory, storage
consent, and footer as consumer configuration/slots. Remove only shell-owned
markup, styles, and scripts.

- [ ] **Step 5: Run Goshtoso site gates**

Run: `GOWORK=off templ generate && GOWORK=off go test ./... && GOWORK=off go vet ./... && GOWORK=off go build ./cmd/server`

Expected: PASS.

- [ ] **Step 6: Run required visual parity matrix**

Compare 1440px, 736px, and 390px in Goshtoso light/dark and Minimal dark. Check
sidebar/drawer, theme, search, TOC, fragment navigation, overflow, and console.

- [ ] **Step 7: Record snags, commit, and push**

Commit focused Goshtoso adoption on `codex/adopt-catalogshell` and push branch.

### Task 7: Adopt in Goshtoso Charts

**Files:**
- Modify: `site/go.mod`, `site/go.sum`
- Modify: `site/internal/pages/layout.templ`
- Modify: `site/internal/siteassets/site.css`
- Modify: `site/internal/server/server.go`
- Modify: `site/internal/server/server_test.go`
- Modify: `docs/tasklist.md`, `docs/goshtoso-snags.md`

**Interfaces:**
- Consumes: pushed `goshtoso-app-shells` commit and existing chart content pages.
- Produces: Goshtoso Charts with no copied shell implementation.

- [ ] **Step 1: Strengthen consumer tests**

Assert public shell assets, brand/navigation configuration, active item, and
Heartbeat/Line content contracts. Retain `/assets/` and `head.Dependencies()`
coverage through rendered shell output.

- [ ] **Step 2: Update remote dependency**

Run: `cd site && GOWORK=off go get github.com/araihu/goshtoso-app-shells@main`,
then record the resolved pseudo-version from `go list -m -json`.

- [ ] **Step 3: Replace copied layout**

Build shared `catalogshell.Config`; render current page content through
`catalogshell.Layout`. Mount shell assets. Delete only shell-owned CSS and
helpers; keep chart-preview styling and generic chart components untouched.

- [ ] **Step 4: Regenerate and run acceptance gates**

Run root: `go test ./... && go vet ./... && go build ./...`

Run site: `GOWORK=off templ generate && GOWORK=off go test ./... && GOWORK=off go vet ./... && GOWORK=off go build ./cmd/server`

Expected: PASS.

- [ ] **Step 5: Run consumer visual matrix**

Verify 1440px, 736px, and 390px; Goshtoso light/dark and Minimal dark; search,
sidebar/drawer, charts, no overflow, and no console errors.

- [ ] **Step 6: Record snag checkpoint, commit, and push**

Commit focused adoption on `codex/goshtoso-charts-control-plane` and push.

### Task 8: Release readiness and handoff

**Files:**
- Modify: `README.md` in new module if integration reveals contract gaps.
- Modify: consumer snag/task documentation only when evidence requires it.

**Interfaces:**
- Consumes: green module and consumer branches.
- Produces: precise release recommendation and pushed SHAs.

- [ ] **Step 1: Verify clean remote-backed state**

For every repository, verify clean status, branch, HEAD SHA, pushed SHA, and
remote module resolution with `GOWORK=off go list -m -json`.

- [ ] **Step 2: Run cross-repository final gates**

Re-run module, Goshtoso site, and Goshtoso Charts root/site commands from clean
worktrees. Confirm generated drift is absent.

- [ ] **Step 3: Decide release status**

Recommend a first tag only if both consumer gates and visual matrix pass.
Otherwise report exact blocker and keep consumers pinned to pushed commit.

- [ ] **Step 4: Send completion evidence**

Report repository URLs, branch and SHA mappings, test commands/results, visual
evidence, recorded snags, blockers, and any deliberately deferred assurance.
