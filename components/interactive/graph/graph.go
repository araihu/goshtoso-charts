// Package graph provides the canonical interactive relationship-graph API.
//
// Graph-specific types and implementation live here; shared renderer-neutral
// options remain in components/chart.
package graph

import (
	"fmt"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// Instance is the renderer-neutral chart instance returned by Graph.
type Instance = chart.Instance

// Layout selects how graph nodes are positioned.
type Layout string

const (
	// LayoutForce uses a force-directed layout. It is the zero-value default.
	LayoutForce Layout = ""
	// LayoutNone uses the X and Y coordinates supplied by each node.
	LayoutNone Layout = "none"
	// LayoutCircular places nodes in a circle.
	LayoutCircular Layout = "circular"
)

// Roam controls mouse zooming and translation.
type Roam uint8

const (
	// RoamDisabled disables mouse zooming and translation. It is the default.
	RoamDisabled Roam = iota
	// RoamEnabled enables mouse zooming and translation.
	RoamEnabled
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

// Config describes an accessible, browser-rendered relationship graph.
//
// Node and link values must be application-owned because the browser renderer serializes them.
type Config struct {
	Label              string
	Caption            string
	Nodes              []Node
	Links              []Link
	Categories         []Category
	Layout             Layout
	Roam               Roam
	Force              *ForceOptions
	Draggable          *bool
	FocusNodeAdjacency *bool
	Width              string
	Height             string
	Options            chart.ChartOptions
	SeriesOptions      chart.SeriesOptions
	Style              charttheme.Style
}

// Node describes one uniquely named graph node.
// Category refers to a Category name. X and Y are used by LayoutNone.
type Node struct {
	Name      string
	Value     float64
	X         *float64
	Y         *float64
	Category  string
	Size      float64
	Fixed     *bool
	ItemStyle *chart.ItemStyle
}

// Link describes a directed relationship between two named nodes.
type Link struct {
	Source    string
	Target    string
	Value     float64
	LineStyle *chart.LineStyle
}

// Category describes one named node classification.
type Category struct {
	Name      string
	ItemStyle *chart.ItemStyle
	Label     *chart.LabelOptions
}

// Graph builds a reusable interactive relationship graph.
func Graph(cfg Config) Instance {
	if err := validateGraphConfig(cfg); err != nil {
		return internalinteractive.Invalid(chartcomponents.KindInteractiveGraph, err)
	}

	chart := charts.NewGraph()
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

	nodes := make([]opts.GraphNode, len(cfg.Nodes))
	for index, node := range cfg.Nodes {
		nodes[index] = rendererGraphNode(node)
	}
	links := make([]opts.GraphLink, len(cfg.Links))
	for index, link := range cfg.Links {
		links[index] = rendererGraphLink(link)
	}

	seriesOptions := []charts.SeriesOpts{graphChartOptions(cfg)}
	seriesOptions = append(seriesOptions, internalinteractive.ChartSeriesOptions(cfg.SeriesOptions)...)
	chart.AddSeries(cfg.Label, nodes, links, seriesOptions...)

	return internalinteractive.New(chartcomponents.KindInteractiveGraph, internalinteractive.RenderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export, ResponsiveWidth: internalinteractive.ResponsiveWidth(cfg.Width),
	})
}

func graphChartOptions(cfg Config) charts.SeriesOpts {
	var categories []*opts.GraphCategory
	if len(cfg.Categories) > 0 {
		categories = make([]*opts.GraphCategory, len(cfg.Categories))
	}
	for index, category := range cfg.Categories {
		rendered := &opts.GraphCategory{Name: category.Name}
		if category.ItemStyle != nil {
			style := internalinteractive.RendererItemStyle(category.ItemStyle)
			rendered.ItemStyle = &style
		}
		if category.Label != nil {
			label := internalinteractive.RendererLabel(category.Label)
			rendered.Label = &label
		}
		categories[index] = rendered
	}

	options := opts.GraphChart{Layout: resolvedGraphLayout(cfg.Layout), Categories: categories}
	if cfg.Roam == RoamEnabled {
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
		style := internalinteractive.RendererItemStyle(node.ItemStyle)
		result.ItemStyle = &style
	}
	return result
}

func rendererGraphLink(link Link) opts.GraphLink {
	result := opts.GraphLink{Source: link.Source, Target: link.Target, Value: float32(link.Value)}
	if link.LineStyle != nil {
		style := internalinteractive.RendererLineStyle(link.LineStyle)
		result.LineStyle = &style
	}
	return result
}

func resolvedGraphLayout(layout Layout) string {
	if layout == LayoutForce {
		return "force"
	}
	return string(layout)
}

func validateGraphConfig(cfg Config) error {
	if cfg.Label == "" {
		return fmt.Errorf("graph chart label is required")
	}
	if cfg.Layout != LayoutForce && cfg.Layout != LayoutNone && cfg.Layout != LayoutCircular {
		return fmt.Errorf("graph chart layout %q is not supported", cfg.Layout)
	}
	if cfg.Roam != RoamDisabled && cfg.Roam != RoamEnabled {
		return fmt.Errorf("graph chart roam mode %d is not supported", cfg.Roam)
	}
	if cfg.Force != nil {
		if cfg.Layout != LayoutForce {
			return fmt.Errorf("graph chart force options require force layout")
		}
		if cfg.Force.InitialLayout != ForceInitialLayoutNone && cfg.Force.InitialLayout != ForceInitialLayoutCircular {
			return fmt.Errorf("graph chart force initial layout %q is not supported", cfg.Force.InitialLayout)
		}
		if !internalinteractive.FiniteNumber(cfg.Force.Repulsion) || cfg.Force.Repulsion < 0 {
			return fmt.Errorf("graph chart force repulsion must be finite and nonnegative")
		}
		if !internalinteractive.FiniteNumber(cfg.Force.Gravity) || cfg.Force.Gravity < 0 || cfg.Force.Gravity > 1 {
			return fmt.Errorf("graph chart force gravity must be finite and between 0 and 1")
		}
		if !internalinteractive.FiniteNumber(cfg.Force.EdgeLength) || cfg.Force.EdgeLength < 0 {
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
		if !internalinteractive.FiniteNumber(node.Value) {
			return fmt.Errorf("graph chart node %q value must be finite", node.Name)
		}
		if node.X != nil && !internalinteractive.FiniteNumber(*node.X) {
			return fmt.Errorf("graph chart node %q x coordinate must be finite", node.Name)
		}
		if node.Y != nil && !internalinteractive.FiniteNumber(*node.Y) {
			return fmt.Errorf("graph chart node %q y coordinate must be finite", node.Name)
		}
		if !internalinteractive.FiniteNumber(node.Size) || node.Size < 0 {
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
		if !internalinteractive.FiniteNumber(link.Value) {
			return fmt.Errorf("graph chart link %d value must be finite", index)
		}
	}
	return nil
}
