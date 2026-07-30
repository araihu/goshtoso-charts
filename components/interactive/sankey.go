package interactive

import interactivesankey "github.com/araihu/goshtoso-charts/components/interactive/sankey"

// SankeyOrientation is retained as an alias for compatibility.
type SankeyOrientation = interactivesankey.Orientation

const (
	// SankeyOrientationHorizontal renders flow from left to right. It is the default.
	SankeyOrientationHorizontal = interactivesankey.OrientationHorizontal
	// SankeyOrientationVertical renders flow from top to bottom.
	SankeyOrientationVertical = interactivesankey.OrientationVertical
)

// SankeyAlignment is retained as an alias for compatibility.
type SankeyAlignment = interactivesankey.Alignment

const (
	// SankeyAlignmentJustify places terminal nodes at the far edge. It is the default.
	SankeyAlignmentJustify = interactivesankey.AlignmentJustify
	// SankeyAlignmentLeft aligns nodes to the start of the flow.
	SankeyAlignmentLeft = interactivesankey.AlignmentLeft
	// SankeyAlignmentRight aligns nodes to the end of the flow.
	SankeyAlignmentRight = interactivesankey.AlignmentRight
)

// SankeyLayout is retained as an alias for compatibility.
type SankeyLayout = interactivesankey.Layout

// SankeyConfig is retained as an alias for compatibility.
type SankeyConfig = interactivesankey.Config

// SankeySeries is retained as an alias for compatibility.
type SankeySeries = interactivesankey.Series

// SankeyNode is retained as an alias for compatibility.
type SankeyNode = interactivesankey.Node

// SankeyLink is retained as an alias for compatibility.
type SankeyLink = interactivesankey.Link

// Sankey forwards to the canonical child-package implementation.
func Sankey(cfg SankeyConfig) Instance { return interactivesankey.Sankey(cfg) }
