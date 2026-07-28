package interactive

import (
	"fmt"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// GraphLayout selects how graph nodes are positioned.
type GraphLayout string

const (
	// GraphLayoutForce uses a force-directed layout. It is the zero-value default.
	GraphLayoutForce GraphLayout = ""
	// GraphLayoutNone uses the X and Y coordinates supplied by each node.
	GraphLayoutNone GraphLayout = "none"
	// GraphLayoutCircular places nodes in a circle.
	GraphLayoutCircular GraphLayout = "circular"
)

// GraphRoam controls mouse zooming and translation.
type GraphRoam uint8

const (
	// GraphRoamDisabled disables mouse zooming and translation. It is the default.
	GraphRoamDisabled GraphRoam = iota
	// GraphRoamEnabled enables mouse zooming and translation.
	GraphRoamEnabled
)

// ForceInitialLayout selects the starting arrangement for force simulation.
type ForceInitialLayout string

const (
	// ForceInitialLayoutNone lets the renderer use node coordinates or generated positions.
	ForceInitialLayoutNone ForceInitialLayout = ""
	// ForceInitialLayoutCircular starts nodes in a circle.
	ForceInitialLayoutCircular ForceInitialLayout = "circular"
)

// ForceOptions configures force-directed graph positioning.
// Zero numeric values retain renderer defaults.
type ForceOptions struct {
	InitialLayout ForceInitialLayout
	Repulsion     float64
	Gravity       float64
	EdgeLength    float64
}

// GraphConfig describes an accessible, browser-rendered relationship graph.
//
// Node and link values must be application-owned because the browser renderer serializes them.
type GraphConfig struct {
	Label              string
	Caption            string
	Nodes              []Node
	Links              []Link
	Categories         []Category
	Layout             GraphLayout
	Roam               GraphRoam
	Force              *ForceOptions
	Draggable          *bool
	FocusNodeAdjacency *bool
	Width              string
	Height             string
	Options            ChartOptions
	SeriesOptions      SeriesOptions
	Style              charttheme.Style
}

// Node describes one uniquely named graph node.
// Category refers to a Category name. X and Y are used by GraphLayoutNone.
type Node struct {
	Name      string
	Value     float64
	X         *float64
	Y         *float64
	Category  string
	Size      float64
	Fixed     *bool
	ItemStyle *ItemStyle
}

// Link describes a directed relationship between two named nodes.
type Link struct {
	Source    string
	Target    string
	Value     float64
	LineStyle *LineStyle
}

// Category describes one named node classification.
type Category struct {
	Name      string
	ItemStyle *ItemStyle
	Label     *LabelOptions
}

// Graph builds a reusable interactive relationship graph.
func Graph(cfg GraphConfig) Instance {
	if err := validateGraphConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveGraph, err)
	}

	chart := charts.NewGraph()
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

	nodes := make([]opts.GraphNode, len(cfg.Nodes))
	for index, node := range cfg.Nodes {
		nodes[index] = rendererGraphNode(node)
	}
	links := make([]opts.GraphLink, len(cfg.Links))
	for index, link := range cfg.Links {
		links[index] = rendererGraphLink(link)
	}

	seriesOptions := []charts.SeriesOpts{graphChartOptions(cfg)}
	seriesOptions = append(seriesOptions, chartSeriesOptions(cfg.SeriesOptions)...)
	chart.AddSeries(cfg.Label, nodes, links, seriesOptions...)

	return newInstance(chartcomponents.KindInteractiveGraph, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export, ResponsiveWidth: responsiveWidth(cfg.Width),
	})
}

func graphChartOptions(cfg GraphConfig) charts.SeriesOpts {
	var categories []*opts.GraphCategory
	if len(cfg.Categories) > 0 {
		categories = make([]*opts.GraphCategory, len(cfg.Categories))
	}
	for index, category := range cfg.Categories {
		rendered := &opts.GraphCategory{Name: category.Name}
		if category.ItemStyle != nil {
			style := rendererItemStyle(category.ItemStyle)
			rendered.ItemStyle = &style
		}
		if category.Label != nil {
			label := rendererLabel(category.Label)
			rendered.Label = &label
		}
		categories[index] = rendered
	}

	options := opts.GraphChart{Layout: resolvedGraphLayout(cfg.Layout), Categories: categories}
	if cfg.Roam == GraphRoamEnabled {
		options.Roam = opts.Bool(true)
	}
	if cfg.Draggable != nil {
		options.Draggable = opts.Bool(*cfg.Draggable)
	}
	if cfg.FocusNodeAdjacency != nil {
		options.FocusNodeAdjacency = opts.Bool(*cfg.FocusNodeAdjacency)
	}
	if cfg.Force != nil {
		force := &opts.GraphForce{
			InitLayout: string(cfg.Force.InitialLayout),
			Repulsion:  float32(cfg.Force.Repulsion),
			Gravity:    float32(cfg.Force.Gravity),
		}
		if cfg.Force.EdgeLength != 0 {
			force.EdgeLength = float32(cfg.Force.EdgeLength)
		}
		options.Force = force
	}
	return charts.WithGraphChartOpts(options)
}

func rendererGraphNode(node Node) opts.GraphNode {
	result := opts.GraphNode{Name: node.Name, Value: float32(node.Value)}
	if node.X != nil {
		result.X = float32(*node.X)
	}
	if node.Y != nil {
		result.Y = float32(*node.Y)
	}
	if node.Category != "" {
		result.Category = node.Category
	}
	if node.Size != 0 {
		result.SymbolSize = float32(node.Size)
	}
	if node.Fixed != nil {
		result.Fixed = opts.Bool(*node.Fixed)
	}
	if node.ItemStyle != nil {
		style := rendererItemStyle(node.ItemStyle)
		result.ItemStyle = &style
	}
	return result
}

func rendererGraphLink(link Link) opts.GraphLink {
	result := opts.GraphLink{Source: link.Source, Target: link.Target, Value: float32(link.Value)}
	if link.LineStyle != nil {
		style := rendererLineStyle(link.LineStyle)
		result.LineStyle = &style
	}
	return result
}

func resolvedGraphLayout(layout GraphLayout) string {
	if layout == GraphLayoutForce {
		return "force"
	}
	return string(layout)
}

func validateGraphConfig(cfg GraphConfig) error {
	if cfg.Label == "" {
		return fmt.Errorf("graph chart label is required")
	}
	if cfg.Layout != GraphLayoutForce && cfg.Layout != GraphLayoutNone && cfg.Layout != GraphLayoutCircular {
		return fmt.Errorf("graph chart layout %q is not supported", cfg.Layout)
	}
	if cfg.Roam != GraphRoamDisabled && cfg.Roam != GraphRoamEnabled {
		return fmt.Errorf("graph chart roam mode %d is not supported", cfg.Roam)
	}
	if cfg.Force != nil {
		if cfg.Layout != GraphLayoutForce {
			return fmt.Errorf("graph chart force options require force layout")
		}
		if cfg.Force.InitialLayout != ForceInitialLayoutNone && cfg.Force.InitialLayout != ForceInitialLayoutCircular {
			return fmt.Errorf("graph chart force initial layout %q is not supported", cfg.Force.InitialLayout)
		}
		if !finiteNumber(cfg.Force.Repulsion) || cfg.Force.Repulsion < 0 {
			return fmt.Errorf("graph chart force repulsion must be finite and nonnegative")
		}
		if !finiteNumber(cfg.Force.Gravity) || cfg.Force.Gravity < 0 || cfg.Force.Gravity > 1 {
			return fmt.Errorf("graph chart force gravity must be finite and between 0 and 1")
		}
		if !finiteNumber(cfg.Force.EdgeLength) || cfg.Force.EdgeLength < 0 {
			return fmt.Errorf("graph chart force edge length must be finite and nonnegative")
		}
	}

	categoryNames := make(map[string]struct{}, len(cfg.Categories))
	for index, category := range cfg.Categories {
		if category.Name == "" {
			return fmt.Errorf("graph chart category %d name is required", index)
		}
		if _, exists := categoryNames[category.Name]; exists {
			return fmt.Errorf("graph chart category %q is duplicated", category.Name)
		}
		categoryNames[category.Name] = struct{}{}
	}
	if len(cfg.Nodes) == 0 {
		return fmt.Errorf("graph chart nodes are required")
	}
	nodeNames := make(map[string]struct{}, len(cfg.Nodes))
	for index, node := range cfg.Nodes {
		if node.Name == "" {
			return fmt.Errorf("graph chart node %d name is required", index)
		}
		if _, exists := nodeNames[node.Name]; exists {
			return fmt.Errorf("graph chart node %q is duplicated", node.Name)
		}
		if !finiteNumber(node.Value) {
			return fmt.Errorf("graph chart node %q value must be finite", node.Name)
		}
		if node.X != nil && !finiteNumber(*node.X) {
			return fmt.Errorf("graph chart node %q x coordinate must be finite", node.Name)
		}
		if node.Y != nil && !finiteNumber(*node.Y) {
			return fmt.Errorf("graph chart node %q y coordinate must be finite", node.Name)
		}
		if !finiteNumber(node.Size) || node.Size < 0 {
			return fmt.Errorf("graph chart node %q size must be finite and nonnegative", node.Name)
		}
		if node.Category != "" {
			if _, exists := categoryNames[node.Category]; !exists {
				return fmt.Errorf("graph chart node %q category %q is not defined", node.Name, node.Category)
			}
		}
		nodeNames[node.Name] = struct{}{}
	}
	for index, link := range cfg.Links {
		if link.Source == "" {
			return fmt.Errorf("graph chart link %d source is required", index)
		}
		if _, exists := nodeNames[link.Source]; !exists {
			return fmt.Errorf("graph chart link %d source %q is not defined", index, link.Source)
		}
		if link.Target == "" {
			return fmt.Errorf("graph chart link %d target is required", index)
		}
		if _, exists := nodeNames[link.Target]; !exists {
			return fmt.Errorf("graph chart link %d target %q is not defined", index, link.Target)
		}
		if !finiteNumber(link.Value) {
			return fmt.Errorf("graph chart link %d value must be finite", index)
		}
	}
	return nil
}
