# Changelog

## [v0.0.5] - 2026-10-30

### Compatibility

- Upgrade the templ runtime and generated chart components to v0.3.1100.
- Upgrade the Goshtoso dependency to v0.4.6 and use App Shells v0.1.9 in the
  documentation site so its generated components use the same templ runtime.
- Require Go 1.27.3 and use Dagger 0.21.11 for the portable CI pipeline.

### Maintenance

- Refresh the checked-in Arai Hû brand fallbacks to Assets v0.2.6 and support
  its catalog v2 format.
- Remove the retired central demo deployment integration. This module release
  does not publish a new hosted demo.

### Upgrading

- Regenerate application-owned `.templ` files with templ v0.3.1100 when
  updating the runtime. Serve assets from the selected Goshtoso module version.
- Public chart configuration APIs and browser chart-runtime versions remain
  unchanged.
