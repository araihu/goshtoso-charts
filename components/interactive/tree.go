package interactive

import interactivetree "github.com/araihu/goshtoso-charts/components/interactive/tree"

// TreeLayout is the compatibility name for tree.Layout.
type TreeLayout = interactivetree.Layout

const (
	// TreeLayoutLayered places each depth on a separate layer. It is the default.
	TreeLayoutLayered TreeLayout = interactivetree.LayoutLayered
	// TreeLayoutRadial places the root at the center and depths on concentric rings.
	TreeLayoutRadial TreeLayout = interactivetree.LayoutRadial
)

// TreeOrientation is the compatibility name for tree.Orientation.
type TreeOrientation = interactivetree.Orientation

const (
	// TreeOrientationLeftToRight grows from left to right. It is the default.
	TreeOrientationLeftToRight TreeOrientation = interactivetree.OrientationLeftToRight
	// TreeOrientationRightToLeft grows from right to left.
	TreeOrientationRightToLeft TreeOrientation = interactivetree.OrientationRightToLeft
	// TreeOrientationTopToBottom grows from top to bottom.
	TreeOrientationTopToBottom TreeOrientation = interactivetree.OrientationTopToBottom
	// TreeOrientationBottomToTop grows from bottom to top.
	TreeOrientationBottomToTop TreeOrientation = interactivetree.OrientationBottomToTop
)

// TreeRoam is the compatibility name for tree.Roam.
type TreeRoam = interactivetree.Roam

const (
	// TreeRoamDisabled disables zooming and translation. It is the default.
	TreeRoamDisabled TreeRoam = interactivetree.RoamDisabled
	// TreeRoamEnabled enables zooming and translation.
	TreeRoamEnabled TreeRoam = interactivetree.RoamEnabled
)

// TreeSymbol is the compatibility name for tree.Symbol.
type TreeSymbol = interactivetree.Symbol

const (
	// TreeSymbolCircle uses circular nodes. It is the default.
	TreeSymbolCircle TreeSymbol = interactivetree.SymbolCircle
	// TreeSymbolRectangle uses rectangular nodes.
	TreeSymbolRectangle TreeSymbol = interactivetree.SymbolRectangle
	// TreeSymbolRoundedRectangle uses rectangles with rounded corners.
	TreeSymbolRoundedRectangle TreeSymbol = interactivetree.SymbolRoundedRectangle
	// TreeSymbolTriangle uses triangular nodes.
	TreeSymbolTriangle TreeSymbol = interactivetree.SymbolTriangle
	// TreeSymbolDiamond uses diamond-shaped nodes.
	TreeSymbolDiamond TreeSymbol = interactivetree.SymbolDiamond
	// TreeSymbolPin uses pin-shaped nodes.
	TreeSymbolPin TreeSymbol = interactivetree.SymbolPin
	// TreeSymbolArrow uses arrow-shaped nodes.
	TreeSymbolArrow TreeSymbol = interactivetree.SymbolArrow
	// TreeSymbolNone hides node symbols while preserving branches and labels.
	TreeSymbolNone TreeSymbol = interactivetree.SymbolNone
)

// TreeInsets is the compatibility name for tree.Insets.
type TreeInsets = interactivetree.Insets

// TreeConfig is the compatibility name for tree.Config.
type TreeConfig = interactivetree.Config

// TreeNode is the compatibility name for tree.Node.
type TreeNode = interactivetree.Node

// Tree forwards to the canonical tree package.
func Tree(cfg TreeConfig) Instance { return interactivetree.Tree(cfg) }
