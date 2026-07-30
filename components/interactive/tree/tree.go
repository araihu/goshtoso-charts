// Package tree provides the canonical interactive hierarchy-chart API.
//
// Tree-specific types and implementation live here; shared renderer-neutral
// options remain in components/chart.
package tree

import (
	"fmt"
	"strconv"
	"strings"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// Instance is the renderer-neutral chart instance returned by Tree.
type Instance = chart.Instance

const (
	maxTreeDepth                  = 256
	treeInitialDepthZeroSentinel  = -2147483648
	treeInitialDepthSentinelField = `"initialTreeDepth":-2147483648`
)

// Layout selects the hierarchy's overall geometry.
type Layout string

const (
	// LayoutLayered places each depth on a separate layer. It is the default.
	LayoutLayered Layout = ""
	// LayoutRadial places the root at the center and depths on concentric rings.
	LayoutRadial Layout = "radial"
)

// Orientation selects the direction in which a layered tree grows.
// Radial trees ignore orientation.
type Orientation string

const (
	// OrientationLeftToRight grows from left to right. It is the default.
	OrientationLeftToRight Orientation = ""
	// OrientationRightToLeft grows from right to left.
	OrientationRightToLeft Orientation = "right-to-left"
	// OrientationTopToBottom grows from top to bottom.
	OrientationTopToBottom Orientation = "top-to-bottom"
	// OrientationBottomToTop grows from bottom to top.
	OrientationBottomToTop Orientation = "bottom-to-top"
)

// Roam controls mouse and touch zooming and translation.
type Roam uint8

const (
	// RoamDisabled disables zooming and translation. It is the default.
	RoamDisabled Roam = iota
	// RoamEnabled enables zooming and translation.
	RoamEnabled
)

// Symbol selects a built-in node shape.
type Symbol string

const (
	// SymbolCircle uses circular nodes. It is the default.
	SymbolCircle Symbol = ""
	// SymbolRectangle uses rectangular nodes.
	SymbolRectangle Symbol = "rectangle"
	// SymbolRoundedRectangle uses rectangles with rounded corners.
	SymbolRoundedRectangle Symbol = "rounded-rectangle"
	// SymbolTriangle uses triangular nodes.
	SymbolTriangle Symbol = "triangle"
	// SymbolDiamond uses diamond-shaped nodes.
	SymbolDiamond Symbol = "diamond"
	// SymbolPin uses pin-shaped nodes.
	SymbolPin Symbol = "pin"
	// SymbolArrow uses arrow-shaped nodes.
	SymbolArrow Symbol = "arrow"
	// SymbolNone hides node symbols while preserving branches and labels.
	SymbolNone Symbol = "none"
)

// Insets controls spacing between the hierarchy and its chart container.
// Values accept CSS lengths or percentages. Empty values retain renderer defaults.
type Insets struct {
	Left   string
	Right  string
	Top    string
	Bottom string
}

// Config describes an accessible, browser-rendered hierarchy.
//
// Nodes must be application-owned because the browser renderer serializes them.
type Config struct {
	Label             string
	Caption           string
	Roots             []*Node
	Layout            Layout
	Orientation       Orientation
	Roam              Roam
	ExpandAndCollapse *bool
	// InitialDepth controls initial expansion. Nil retains the renderer default,
	// -1 expands every level, and zero displays only roots.
	InitialDepth *int
	NodeLabel    *chart.LabelOptions
	LeafLabel    *chart.LabelOptions
	Symbol       Symbol
	SymbolSize   int
	Insets       Insets
	Width        string
	Height       string
	Options      chart.ChartOptions
	Style        charttheme.Style
}

// Node describes one node and its descendants.
type Node struct {
	Name       string
	Value      float64
	Children   []*Node
	Collapsed  *bool
	Symbol     Symbol
	SymbolSize int
	ItemStyle  *chart.ItemStyle
	LineStyle  *chart.LineStyle
}

// Tree builds a reusable interactive hierarchy.
func Tree(cfg Config) Instance {
	if err := validateConfig(cfg); err != nil {
		return internalinteractive.Invalid(chartcomponents.KindInteractiveTree, err)
	}

	chart := charts.NewTree()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, internalinteractive.ChartGlobalOptions(cfg.Options)...)
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
		roots[index] = rendererNode(root)
	}
	seriesOptions := []charts.SeriesOpts{treeChartOptions(cfg)}
	seriesOptions = append(seriesOptions, func(series *charts.SingleSeries) {
		// Tree relayout must settle before wrapper resize/theme transitions. Keep
		// this private: callers configure whether animation exists, not engine timing.
		series.AnimationDuration = 150
		series.AnimationDurationUpdate = 100
	})
	if cfg.Symbol != SymbolCircle || cfg.SymbolSize != 0 {
		seriesOptions = append(seriesOptions, func(series *charts.SingleSeries) {
			series.Symbol = rendererSymbol(cfg.Symbol)
			if cfg.SymbolSize != 0 {
				series.SymbolSize = cfg.SymbolSize
			}
		})
	}
	chart.AddSeries(cfg.Label, roots, seriesOptions...)

	render := internalinteractive.RenderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, Animation: cfg.Options.Animation, ResponsiveWidth: internalinteractive.ResponsiveWidth(cfg.Width),
		Controls: cfg.Options.Controls, Export: cfg.Options.Export,
	}
	if cfg.InitialDepth != nil && *cfg.InitialDepth == 0 {
		render.ScriptReplacements = append(render.ScriptReplacements, internalinteractive.ScriptReplacement{
			Old: treeInitialDepthSentinelField,
			New: `"initialTreeDepth":0`,
		})
	}
	return internalinteractive.New(chartcomponents.KindInteractiveTree, render)
}

func treeChartOptions(cfg Config) charts.SeriesOpts {
	initialDepth := 0
	if cfg.InitialDepth != nil {
		initialDepth = *cfg.InitialDepth
		if initialDepth == 0 {
			initialDepth = treeInitialDepthZeroSentinel
		}
	}
	insets := resolvedInsets(cfg.Insets)
	tree := opts.TreeChart{
		Layout:           rendererLayout(cfg.Layout),
		Orient:           rendererOrientation(cfg.Orientation),
		InitialTreeDepth: initialDepth,
		Left:             insets.Left,
		Right:            insets.Right,
		Top:              insets.Top,
		Bottom:           insets.Bottom,
	}
	if cfg.Roam == RoamEnabled {
		tree.Roam = opts.Bool(true)
	}
	if cfg.ExpandAndCollapse != nil {
		tree.ExpandAndCollapse = opts.Bool(*cfg.ExpandAndCollapse)
	}
	if cfg.NodeLabel != nil {
		label := internalinteractive.RendererLabel(cfg.NodeLabel)
		tree.Label = &label
	}
	if cfg.LeafLabel != nil {
		label := internalinteractive.RendererLabel(cfg.LeafLabel)
		tree.Leaves = &opts.TreeLeaves{Label: &label}
	}
	return charts.WithTreeOpts(tree)
}

func resolvedInsets(insets Insets) Insets {
	if insets.Left == "" {
		insets.Left = "14%"
	}
	if insets.Right == "" {
		insets.Right = "14%"
	}
	if insets.Top == "" {
		insets.Top = "12%"
	}
	if insets.Bottom == "" {
		insets.Bottom = "12%"
	}
	return insets
}

func rendererNode(node *Node) opts.TreeData {
	rendered := opts.TreeData{
		Name:      node.Name,
		Value:     node.Value,
		Collapsed: node.Collapsed,
		Symbol:    rendererSymbol(node.Symbol),
	}
	if node.SymbolSize != 0 {
		rendered.SymbolSize = node.SymbolSize
	}
	if node.ItemStyle != nil {
		style := internalinteractive.RendererItemStyle(node.ItemStyle)
		rendered.ItemStyle = &style
	}
	if node.LineStyle != nil {
		style := internalinteractive.RendererLineStyle(node.LineStyle)
		rendered.LineStyle = &style
	}
	if len(node.Children) > 0 {
		rendered.Children = make([]*opts.TreeData, len(node.Children))
		for index, child := range node.Children {
			value := rendererNode(child)
			rendered.Children[index] = &value
		}
	}
	return rendered
}

func rendererLayout(layout Layout) string {
	if layout == LayoutLayered {
		return "orthogonal"
	}
	return string(layout)
}

func rendererOrientation(orientation Orientation) string {
	switch orientation {
	case OrientationRightToLeft:
		return "RL"
	case OrientationTopToBottom:
		return "TB"
	case OrientationBottomToTop:
		return "BT"
	default:
		return "LR"
	}
}

func rendererSymbol(symbol Symbol) string {
	switch symbol {
	case SymbolRectangle:
		return "rect"
	case SymbolRoundedRectangle:
		return "roundRect"
	default:
		return string(symbol)
	}
}

func validateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("tree chart label is required")
	}
	if cfg.Layout != LayoutLayered && cfg.Layout != LayoutRadial {
		return fmt.Errorf("tree chart layout %q is not supported", cfg.Layout)
	}
	if cfg.Orientation != OrientationLeftToRight && cfg.Orientation != OrientationRightToLeft &&
		cfg.Orientation != OrientationTopToBottom && cfg.Orientation != OrientationBottomToTop {
		return fmt.Errorf("tree chart orientation %q is not supported", cfg.Orientation)
	}
	if cfg.Layout == LayoutRadial && cfg.Orientation != OrientationLeftToRight {
		return fmt.Errorf("tree chart orientation requires layered layout")
	}
	if cfg.Roam != RoamDisabled && cfg.Roam != RoamEnabled {
		return fmt.Errorf("tree chart roam mode %d is not supported", cfg.Roam)
	}
	if cfg.InitialDepth != nil && *cfg.InitialDepth < -1 {
		return fmt.Errorf("tree chart initial depth must be -1 or nonnegative")
	}
	if err := validateSymbol("tree chart", cfg.Symbol, cfg.SymbolSize); err != nil {
		return err
	}
	if len(cfg.Roots) == 0 {
		return fmt.Errorf("tree chart roots are required")
	}
	active := make(map[*Node]bool)
	seen := make(map[*Node]bool)
	for index, root := range cfg.Roots {
		if err := validateNode(root, "root "+strconv.Itoa(index), 0, active, seen); err != nil {
			return err
		}
	}
	return internalinteractive.ValidateChartOptions(cfg.Options)
}

func validateNode(node *Node, path string, depth int, active, seen map[*Node]bool) error {
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
	if !internalinteractive.FiniteNumber(node.Value) {
		return fmt.Errorf("tree chart node %q value must be finite", node.Name)
	}
	if err := validateSymbol("tree chart node "+strconv.Quote(node.Name), node.Symbol, node.SymbolSize); err != nil {
		return err
	}
	active[node] = true
	for index, child := range node.Children {
		childPath := "node " + strconv.Quote(node.Name) + " child " + strconv.Itoa(index)
		if err := validateNode(child, childPath, depth+1, active, seen); err != nil {
			return err
		}
	}
	delete(active, node)
	seen[node] = true
	return nil
}

func validateSymbol(owner string, symbol Symbol, size int) error {
	switch symbol {
	case SymbolCircle, SymbolRectangle, SymbolRoundedRectangle, SymbolTriangle,
		SymbolDiamond, SymbolPin, SymbolArrow, SymbolNone:
	default:
		return fmt.Errorf("%s symbol %q is not supported", owner, symbol)
	}
	if size < 0 {
		return fmt.Errorf("%s symbol size must be nonnegative", owner)
	}
	return nil
}
