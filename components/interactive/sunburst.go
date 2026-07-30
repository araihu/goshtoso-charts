package interactive

import interactivesunburst "github.com/araihu/goshtoso-charts/components/interactive/sunburst"

// SunburstNavigation is retained as an alias for compatibility.
type SunburstNavigation = interactivesunburst.Navigation

const (
	// SunburstNavigationDrillDown makes a clicked sector the visible root. It is the default.
	// Selecting the center returns to the previous root.
	SunburstNavigationDrillDown SunburstNavigation = interactivesunburst.NavigationDrillDown
	// SunburstNavigationDisabled keeps the full hierarchy visible when sectors are clicked.
	SunburstNavigationDisabled SunburstNavigation = interactivesunburst.NavigationDisabled
)

// SunburstSort is retained as an alias for compatibility.
type SunburstSort = interactivesunburst.Sort

const (
	// SunburstSortDescending orders sibling sectors from largest to smallest. It is the default.
	SunburstSortDescending SunburstSort = interactivesunburst.SortDescending
	// SunburstSortAscending orders sibling sectors from smallest to largest.
	SunburstSortAscending SunburstSort = interactivesunburst.SortAscending
	// SunburstSortInput preserves caller order.
	SunburstSortInput SunburstSort = interactivesunburst.SortInput
)

// SunburstConfig is retained as an alias for compatibility.
type SunburstConfig = interactivesunburst.Config

// SunburstNode is retained as an alias for compatibility.
type SunburstNode = interactivesunburst.Node

// Sunburst forwards to the canonical child-package implementation.
func Sunburst(cfg SunburstConfig) Instance { return interactivesunburst.Sunburst(cfg) }
