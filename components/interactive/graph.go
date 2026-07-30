package interactive

import interactivegraph "github.com/araihu/goshtoso-charts/components/interactive/graph"

// GraphLayout is the compatibility name for graph.Layout.
type GraphLayout = interactivegraph.Layout

const (
	// GraphLayoutForce uses a force-directed layout. It is the zero-value default.
	GraphLayoutForce GraphLayout = interactivegraph.LayoutForce
	// GraphLayoutNone uses the X and Y coordinates supplied by each node.
	GraphLayoutNone GraphLayout = interactivegraph.LayoutNone
	// GraphLayoutCircular places nodes in a circle.
	GraphLayoutCircular GraphLayout = interactivegraph.LayoutCircular
)

// GraphRoam is the compatibility name for graph.Roam.
type GraphRoam = interactivegraph.Roam

const (
	// GraphRoamDisabled disables mouse zooming and translation. It is the default.
	GraphRoamDisabled GraphRoam = interactivegraph.RoamDisabled
	// GraphRoamEnabled enables mouse zooming and translation.
	GraphRoamEnabled GraphRoam = interactivegraph.RoamEnabled
)

// ForceInitialLayout is the compatibility name for graph.ForceInitialLayout.
type ForceInitialLayout = interactivegraph.ForceInitialLayout

const (
	// ForceInitialLayoutNone lets the renderer use node coordinates or generated positions.
	ForceInitialLayoutNone ForceInitialLayout = interactivegraph.ForceInitialLayoutNone
	// ForceInitialLayoutCircular starts nodes in a circle.
	ForceInitialLayoutCircular ForceInitialLayout = interactivegraph.ForceInitialLayoutCircular
)

// ForceOptions is the compatibility name for graph.ForceOptions.
type ForceOptions = interactivegraph.ForceOptions

// GraphConfig is the compatibility name for graph.Config.
type GraphConfig = interactivegraph.Config

// Node is the compatibility name for graph.Node.
type Node = interactivegraph.Node

// Link is the compatibility name for graph.Link.
type Link = interactivegraph.Link

// Category is the compatibility name for graph.Category.
type Category = interactivegraph.Category

// Graph forwards to the canonical graph package.
func Graph(cfg GraphConfig) Instance { return interactivegraph.Graph(cfg) }
