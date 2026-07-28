# Goshtoso component documentation shell

## Pattern boundary

`componentdocshell` is the reusable documentation frame: full-width brand
header, search-first grouped navigation, appearance controls, responsive drawer,
main scroll region, optional on-page table of contents, and HTMX-enhanced SSR
navigation. It is not a catalog/browse shell.

`componentpage` is a composition helper inside that frame. It renders the
repeated component-reference rhythm: title and description, optional controls or
contract, framed preview, usage code, and variant sections. It owns no routing,
header, navigation, theme policy, or document shell.

`catalogshell` is reserved for a future browse/discovery pattern with its own
accepted consumer projection. No alias links it to `componentdocshell`.

## Configuration

The shell accepts typed nested configuration:

- `Brand` owns name, home link, logo component, and favicon URL.
- `Navigation` owns top-level items, grouped sections, search placeholder, and
  optional search suppression.
- `AppearanceConfig` owns available themes, default theme, initial color
  scheme, persistence, control visibility, and custom theme stylesheets.
- `InteractionConfig` owns progressive enhancement such as HTMX fragments.
- `Page` owns route-specific title, exact optional document title, description,
  canonical URL, active navigation ID, content, head additions, and TOC policy.

Defaults expose canonical Arai Hû plus all built-in Goshtoso themes and select
Arai Hû. Consumers may replace the list, lock a theme while hiding controls, or
disable bundled Arai Hû CSS. Goshtoso Charts enables HTMX and preference
persistence while retaining normal server-rendered links.

## Contracts

- Goshtoso assets remain mounted at `GET /assets/` through `assets.Handler()`;
  shell assets use their own embedded handler.
- `head.Dependencies()` remains the Goshtoso runtime source.
- Full responses render complete SSR documents. Fragment responses render a
  title, main content, and out-of-band active navigation.
- Mobile navigation closes after a successful fragment swap. Focus moves to the
  new page heading, sidebar scroll is retained, and main scroll resets.
- Static/vector charts retain generic typed `Config`, `Instance`, and `Kind`
  contracts. Interactive rendering remains behind renderer-neutral public APIs.

## Visual acceptance

Desktop and 390px checks cover Arai Hû and Goshtoso/Minimal dark modes, sidebar
geometry or drawer behavior, 16px within-example spacing, 40px between variants,
theme inventory, dark/light icon swap, fragment title and active-nav updates,
zero document-level horizontal overflow, and zero console errors.
