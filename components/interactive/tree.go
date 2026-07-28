package interactive

import (
	"fmt"
	"strconv"
	"strings"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

const (
	maxTreeDepth                  = 256
	treeInitialDepthZeroSentinel  = -2147483648
	treeInitialDepthSentinelField = `"initialTreeDepth":-2147483648`
)

// TreeLayout selects the hierarchy's overall geometry.
type TreeLayout string

const (
	// TreeLayoutLayered places each depth on a separate layer. It is the default.
	TreeLayoutLayered TreeLayout = ""
	// TreeLayoutRadial places the root at the center and depths on concentric rings.
	TreeLayoutRadial TreeLayout = "radial"
)

// TreeOrientation selects the direction in which a layered tree grows.
// Radial trees ignore orientation.
type TreeOrientation string

const (
	// TreeOrientationLeftToRight grows from left to right. It is the default.
	TreeOrientationLeftToRight TreeOrientation = ""
	// TreeOrientationRightToLeft grows from right to left.
	TreeOrientationRightToLeft TreeOrientation = "right-to-left"
	// TreeOrientationTopToBottom grows from top to bottom.
	TreeOrientationTopToBottom TreeOrientation = "top-to-bottom"
	// TreeOrientationBottomToTop grows from bottom to top.
	TreeOrientationBottomToTop TreeOrientation = "bottom-to-top"
)

// TreeRoam controls mouse and touch zooming and translation.
type TreeRoam uint8

const (
	// TreeRoamDisabled disables zooming and translation. It is the default.
	TreeRoamDisabled TreeRoam = iota
	// TreeRoamEnabled enables zooming and translation.
	TreeRoamEnabled
)

// TreeSymbol selects a built-in node shape.
type TreeSymbol string

const (
	// TreeSymbolCircle uses circular nodes. It is the default.
	TreeSymbolCircle TreeSymbol = ""
	// TreeSymbolRectangle uses rectangular nodes.
	TreeSymbolRectangle TreeSymbol = "rectangle"
	// TreeSymbolRoundedRectangle uses rectangles with rounded corners.
	TreeSymbolRoundedRectangle TreeSymbol = "rounded-rectangle"
	// TreeSymbolTriangle uses triangular nodes.
	TreeSymbolTriangle TreeSymbol = "triangle"
	// TreeSymbolDiamond uses diamond-shaped nodes.
	TreeSymbolDiamond TreeSymbol = "diamond"
	// TreeSymbolPin uses pin-shaped nodes.
	TreeSymbolPin TreeSymbol = "pin"
	// TreeSymbolArrow uses arrow-shaped nodes.
	TreeSymbolArrow TreeSymbol = "arrow"
	// TreeSymbolNone hides node symbols while preserving branches and labels.
	TreeSymbolNone TreeSymbol = "none"
)

// TreeInsets controls spacing between the hierarchy and its chart container.
// Values accept CSS lengths or percentages. Empty values retain renderer defaults.
type TreeInsets struct {
	Left   string
	Right  string
	Top    string
	Bottom string
}

// TreeConfig describes an accessible, browser-rendered hierarchy.
//
// Nodes must be application-owned because the browser renderer serializes them.
type TreeConfig struct {
	Label             string
	Caption           string
	Roots             []*TreeNode
	Layout            TreeLayout
	Orientation       TreeOrientation
	Roam              TreeRoam
	ExpandAndCollapse *bool
	// InitialDepth controls initial expansion. Nil retains the renderer default,
	// -1 expands every level, and zero displays only roots.
	InitialDepth *int
	NodeLabel    *LabelOptions
	LeafLabel    *LabelOptions
	Symbol       TreeSymbol
	SymbolSize   int
	Insets       TreeInsets
	Width        string
	Height       string
	Options      ChartOptions
	Style        charttheme.Style
}

// TreeNode describes one node and its descendants.
type TreeNode struct {
	Name       string
	Value      float64
	Children   []*TreeNode
	Collapsed  *bool
	Symbol     TreeSymbol
	SymbolSize int
	ItemStyle  *ItemStyle
	LineStyle  *LineStyle
}

// Tree builds a reusable interactive hierarchy.
func Tree(cfg TreeConfig) Instance {
	if err := validateTreeConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveTree, err)
	}

	chart := charts.NewTree()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, chartGlobalOptions(cfg.Options)...)
	if len(cfg.Style.Colors) > 0 {
		// Explicit component colors remain authoritative over escape-hatch options.
		globalOptions = append(globalOptions, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
	if cfg.Width != "" || cfg.Height != "" {
		globalOptions = append([]charts.GlobalOpts{
			charts.WithInitializationOpts(opts.Initialization{Width: cfg.Width, Height: cfg.Height}),
		}, globalOptions...)
	}
	chart.SetGlobalOptions(globalOptions...)

	roots := make([]opts.TreeData, len(cfg.Roots))
	for index, root := range cfg.Roots {
		roots[index] = rendererTreeNode(root)
	}
	seriesOptions := []charts.SeriesOpts{treeChartOptions(cfg)}
	if cfg.Symbol != TreeSymbolCircle || cfg.SymbolSize != 0 {
		seriesOptions = append(seriesOptions, func(series *charts.SingleSeries) {
			series.Symbol = rendererTreeSymbol(cfg.Symbol)
			if cfg.SymbolSize != 0 {
				series.SymbolSize = cfg.SymbolSize
			}
		})
	}
	chart.AddSeries(cfg.Label, roots, seriesOptions...)

	render := renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, Animation: cfg.Options.Animation,
	}
	if cfg.InitialDepth != nil && *cfg.InitialDepth == 0 {
		render.ScriptReplacements = append(render.ScriptReplacements, scriptReplacement{
			Old: treeInitialDepthSentinelField,
			New: `"initialTreeDepth":0`,
		})
	}
	return newInstance(chartcomponents.KindInteractiveTree, render)
}

func treeChartOptions(cfg TreeConfig) charts.SeriesOpts {
	initialDepth := 0
	if cfg.InitialDepth != nil {
		initialDepth = *cfg.InitialDepth
		if initialDepth == 0 {
			initialDepth = treeInitialDepthZeroSentinel
		}
	}
	tree := opts.TreeChart{
		Layout:           rendererTreeLayout(cfg.Layout),
		Orient:           rendererTreeOrientation(cfg.Orientation),
		InitialTreeDepth: initialDepth,
		Left:             cfg.Insets.Left,
		Right:            cfg.Insets.Right,
		Top:              cfg.Insets.Top,
		Bottom:           cfg.Insets.Bottom,
	}
	if cfg.Roam == TreeRoamEnabled {
		tree.Roam = opts.Bool(true)
	}
	if cfg.ExpandAndCollapse != nil {
		tree.ExpandAndCollapse = opts.Bool(*cfg.ExpandAndCollapse)
	}
	if cfg.NodeLabel != nil {
		label := rendererLabel(cfg.NodeLabel)
		tree.Label = &label
	}
	if cfg.LeafLabel != nil {
		label := rendererLabel(cfg.LeafLabel)
		tree.Leaves = &opts.TreeLeaves{Label: &label}
	}
	return charts.WithTreeOpts(tree)
}

func rendererTreeNode(node *TreeNode) opts.TreeData {
	rendered := opts.TreeData{
		Name:      node.Name,
		Value:     node.Value,
		Collapsed: node.Collapsed,
		Symbol:    rendererTreeSymbol(node.Symbol),
	}
	if node.SymbolSize != 0 {
		rendered.SymbolSize = node.SymbolSize
	}
	if node.ItemStyle != nil {
		style := rendererItemStyle(node.ItemStyle)
		rendered.ItemStyle = &style
	}
	if node.LineStyle != nil {
		style := rendererLineStyle(node.LineStyle)
		rendered.LineStyle = &style
	}
	if len(node.Children) > 0 {
		rendered.Children = make([]*opts.TreeData, len(node.Children))
		for index, child := range node.Children {
			value := rendererTreeNode(child)
			rendered.Children[index] = &value
		}
	}
	return rendered
}

func rendererTreeLayout(layout TreeLayout) string {
	if layout == TreeLayoutLayered {
		return "orthogonal"
	}
	return string(layout)
}

func rendererTreeOrientation(orientation TreeOrientation) string {
	switch orientation {
	case TreeOrientationRightToLeft:
		return "RL"
	case TreeOrientationTopToBottom:
		return "TB"
	case TreeOrientationBottomToTop:
		return "BT"
	default:
		return "LR"
	}
}

func rendererTreeSymbol(symbol TreeSymbol) string {
	switch symbol {
	case TreeSymbolRectangle:
		return "rect"
	case TreeSymbolRoundedRectangle:
		return "roundRect"
	default:
		return string(symbol)
	}
}

func validateTreeConfig(cfg TreeConfig) error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("tree chart label is required")
	}
	if cfg.Layout != TreeLayoutLayered && cfg.Layout != TreeLayoutRadial {
		return fmt.Errorf("tree chart layout %q is not supported", cfg.Layout)
	}
	if cfg.Orientation != TreeOrientationLeftToRight && cfg.Orientation != TreeOrientationRightToLeft &&
		cfg.Orientation != TreeOrientationTopToBottom && cfg.Orientation != TreeOrientationBottomToTop {
		return fmt.Errorf("tree chart orientation %q is not supported", cfg.Orientation)
	}
	if cfg.Layout == TreeLayoutRadial && cfg.Orientation != TreeOrientationLeftToRight {
		return fmt.Errorf("tree chart orientation requires layered layout")
	}
	if cfg.Roam != TreeRoamDisabled && cfg.Roam != TreeRoamEnabled {
		return fmt.Errorf("tree chart roam mode %d is not supported", cfg.Roam)
	}
	if cfg.InitialDepth != nil && *cfg.InitialDepth < -1 {
		return fmt.Errorf("tree chart initial depth must be -1 or nonnegative")
	}
	if err := validateTreeSymbol("tree chart", cfg.Symbol, cfg.SymbolSize); err != nil {
		return err
	}
	if len(cfg.Roots) == 0 {
		return fmt.Errorf("tree chart roots are required")
	}
	active := make(map[*TreeNode]bool)
	seen := make(map[*TreeNode]bool)
	for index, root := range cfg.Roots {
		if err := validateTreeNode(root, "root "+strconv.Itoa(index), 0, active, seen); err != nil {
			return err
		}
	}
	return validateChartOptions(cfg.Options)
}

func validateTreeNode(node *TreeNode, path string, depth int, active, seen map[*TreeNode]bool) error {
	if node == nil {
		return fmt.Errorf("tree chart %s is nil", path)
	}
	if active[node] {
		return fmt.Errorf("tree chart %s contains a cycle", path)
	}
	if seen[node] {
		return fmt.Errorf("tree chart %s reuses a node", path)
	}
	if depth > maxTreeDepth {
		return fmt.Errorf("tree chart %s exceeds maximum depth %d", path, maxTreeDepth)
	}
	if strings.TrimSpace(node.Name) == "" {
		return fmt.Errorf("tree chart %s name is required", path)
	}
	if !finiteNumber(node.Value) {
		return fmt.Errorf("tree chart node %q value must be finite", node.Name)
	}
	if err := validateTreeSymbol("tree chart node "+strconv.Quote(node.Name), node.Symbol, node.SymbolSize); err != nil {
		return err
	}
	active[node] = true
	for index, child := range node.Children {
		childPath := "node " + strconv.Quote(node.Name) + " child " + strconv.Itoa(index)
		if err := validateTreeNode(child, childPath, depth+1, active, seen); err != nil {
			return err
		}
	}
	delete(active, node)
	seen[node] = true
	return nil
}

func validateTreeSymbol(owner string, symbol TreeSymbol, size int) error {
	switch symbol {
	case TreeSymbolCircle, TreeSymbolRectangle, TreeSymbolRoundedRectangle, TreeSymbolTriangle,
		TreeSymbolDiamond, TreeSymbolPin, TreeSymbolArrow, TreeSymbolNone:
	default:
		return fmt.Errorf("%s symbol %q is not supported", owner, symbol)
	}
	if size < 0 {
		return fmt.Errorf("%s symbol size must be nonnegative", owner)
	}
	return nil
}
