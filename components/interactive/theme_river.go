package interactive

import interactivethemeriver "github.com/araihu/goshtoso-charts/components/interactive/themeriver"

// ThemeRiverConfig is the compatibility name for themeriver.Config.
type ThemeRiverConfig = interactivethemeriver.Config

// ThemeRiverStream is the compatibility name for themeriver.Stream.
type ThemeRiverStream = interactivethemeriver.Stream

// ThemeRiverPoint is the compatibility name for themeriver.Point.
type ThemeRiverPoint = interactivethemeriver.Point

// ThemeRiverLayout is the compatibility name for themeriver.Layout.
type ThemeRiverLayout = interactivethemeriver.Layout

// ThemeRiverBoundaryGap is the compatibility name for themeriver.BoundaryGap.
type ThemeRiverBoundaryGap = interactivethemeriver.BoundaryGap

// ThemeRiver forwards to the canonical themeriver package.
func ThemeRiver(cfg ThemeRiverConfig) Instance { return interactivethemeriver.ThemeRiver(cfg) }
