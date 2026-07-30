package interactive

import interactivetreemap "github.com/araihu/goshtoso-charts/components/interactive/treemap"

// TreemapNavigation is the compatibility name for treemap.Navigation.
type TreemapNavigation = interactivetreemap.Navigation

const (
	// TreemapNavigationDrillDown focuses a clicked branch in the same chart.
	// The breadcrumb returns to an ancestor. It is the default.
	TreemapNavigationDrillDown TreemapNavigation = interactivetreemap.NavigationDrillDown
	// TreemapNavigationDisabled keeps clicks from changing the visible root.
	TreemapNavigationDisabled TreemapNavigation = interactivetreemap.NavigationDisabled
)

// TreemapRoam is the compatibility name for treemap.Roam.
type TreemapRoam = interactivetreemap.Roam

const (
	// TreemapRoamDisabled disables zooming and translation. It is the default.
	TreemapRoamDisabled TreemapRoam = interactivetreemap.RoamDisabled
	// TreemapRoamEnabled enables zooming and translation.
	TreemapRoamEnabled TreemapRoam = interactivetreemap.RoamEnabled
)

// TreemapBreadcrumb is the compatibility name for treemap.Breadcrumb.
type TreemapBreadcrumb = interactivetreemap.Breadcrumb

// TreemapNodeStyle is the compatibility name for treemap.NodeStyle.
type TreemapNodeStyle = interactivetreemap.NodeStyle

// TreemapColorRange is the compatibility name for treemap.ColorRange.
type TreemapColorRange = interactivetreemap.ColorRange

// TreemapLevel is the compatibility name for treemap.Level.
type TreemapLevel = interactivetreemap.Level

// TreemapConfig is the compatibility name for treemap.Config.
type TreemapConfig = interactivetreemap.Config

// TreemapNode is the compatibility name for treemap.Node.
type TreemapNode = interactivetreemap.Node

// Treemap forwards to the canonical treemap package.
func Treemap(cfg TreemapConfig) Instance { return interactivetreemap.Treemap(cfg) }
