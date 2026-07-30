package interactive

import internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"

// Keep same-package runtime contract tests attached to private runtime bytes
// while runtime ownership moves behind the compatibility facade.
const liveRuntimeMarkup = internalinteractive.LiveRuntimeMarkup
const themeRuntimeMarkup = internalinteractive.ThemeRuntimeMarkup
