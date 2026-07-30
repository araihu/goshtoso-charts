// Package sankey provides the canonical interactive flow-network chart API.
//
// Sankey-specific types and implementation live here; shared renderer-neutral
// options remain in components/chart.
package sankey

import (
	"fmt"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// Instance is the renderer-neutral chart instance returned by Sankey.
type Instance = chart.Instance

// Orientation controls the direction in which flow progresses.
type Orientation string

const (
	// OrientationHorizontal renders flow from left to right. It is the default.
	OrientationHorizontal Orientation = ""
	// OrientationVertical renders flow from top to bottom.
	OrientationVertical Orientation = "vertical"
)

// Alignment controls how nodes without outgoing links are placed.
type Alignment string

const (
	// AlignmentJustify places terminal nodes at the far edge. It is the default.
	AlignmentJustify Alignment = ""
	// AlignmentLeft aligns nodes to the start of the flow.
	AlignmentLeft Alignment = "left"
	// AlignmentRight aligns nodes to the end of the flow.
	AlignmentRight Alignment = "right"
)

// Layout controls renderer-neutral node placement and size.
// NodeWidth and NodeGap use CSS pixels. Zero keeps renderer defaults.
type Layout struct {
	Orientation Orientation
	Alignment   Alignment
	NodeWidth   int
	NodeGap     int
}

// Config describes an accessible, browser-rendered Sankey chart.
//
// Nodes and links must be application-owned because the browser renderer
// serializes them.
type Config struct {
	Label         string
	Caption       string
	Series        []Series
	Layout        Layout
	Width         string
	Height        string
	Options       chart.ChartOptions
	SeriesOptions chart.SeriesOptions
	Style         charttheme.Style
}

// Series describes one named flow network.
type Series struct {
	Name    string
	Nodes   []Node
	Links   []Link
	Options chart.SeriesOptions
}

// Node describes one uniquely named point in a flow network.
// Depth optionally pins the zero-based layer; nil lets the renderer place it.
type Node struct {
	Name      string
	Depth     *int
	ItemStyle *chart.ItemStyle
}

// Link describes one weighted connection between named nodes.
type Link struct {
	Source string
	Target string
	Value  float64
}

// Sankey builds a reusable interactive Sankey component.
func Sankey(cfg Config) Instance {
	if err := validateConfig(cfg); err != nil {
		return internalinteractive.Invalid(chartcomponents.KindInteractiveSankey, err)
	}

	sankeyChart := charts.NewSankey()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, internalinteractive.ChartGlobalOptions(cfg.Options)...)
	// Explicit component colors remain authoritative over escape-hatch options.
	if len(cfg.Style.Colors) > 0 {
		globalOptions = append(globalOptions, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
	if cfg.Width != "" || cfg.Height != "" {
		globalOptions = append([]charts.GlobalOpts{
			charts.WithInitializationOpts(opts.Initialization{Width: cfg.Width, Height: cfg.Height}),
		}, globalOptions...)
	}
	sankeyChart.SetGlobalOptions(globalOptions...)

	for _, series := range cfg.Series {
		nodes := make([]opts.SankeyNode, len(series.Nodes))
		for index, node := range series.Nodes {
			nodes[index] = opts.SankeyNode{Name: node.Name, Depth: node.Depth}
			if node.ItemStyle != nil {
				style := internalinteractive.RendererItemStyle(node.ItemStyle)
				nodes[index].ItemStyle = &style
			}
		}
		links := make([]opts.SankeyLink, len(series.Links))
		for index, link := range series.Links {
			links[index] = opts.SankeyLink{Source: link.Source, Target: link.Target, Value: float32(link.Value)}
		}

		options := make([]charts.SeriesOpts, 0, 1+len(internalinteractive.ChartSeriesOptions(cfg.SeriesOptions))+len(internalinteractive.ChartSeriesOptions(series.Options)))
		options = append(options, layoutOption(cfg.Layout))
		options = append(options, internalinteractive.MergeSeriesOptions(cfg.SeriesOptions, series.Options)...)
		sankeyChart.AddSeries(series.Name, nodes, links, options...)
	}

	return internalinteractive.New(chartcomponents.KindInteractiveSankey, internalinteractive.RenderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: sankeyChart, Style: cfg.Style, Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export, ResponsiveWidth: internalinteractive.ResponsiveWidth(cfg.Width),
	})
}

func layoutOption(layout Layout) charts.SeriesOpts {
	return func(series *charts.SingleSeries) {
		series.Orient = "horizontal"
		if layout.Orientation == OrientationVertical {
			series.Orient = string(layout.Orientation)
		}
		series.NodeAlign = "justify"
		if layout.Alignment != AlignmentJustify {
			series.NodeAlign = string(layout.Alignment)
		}
		if layout.NodeWidth != 0 {
			series.NodeWidth = opts.Int(layout.NodeWidth)
		}
		if layout.NodeGap != 0 {
			series.NodeGap = opts.Int(layout.NodeGap)
		}
	}
}

func validateConfig(cfg Config) error {
	if cfg.Label == "" {
		return fmt.Errorf("sankey chart label is required")
	}
	if cfg.Layout.Orientation != OrientationHorizontal && cfg.Layout.Orientation != OrientationVertical {
		return fmt.Errorf("sankey chart orientation %q is not supported", cfg.Layout.Orientation)
	}
	if cfg.Layout.Alignment != AlignmentJustify && cfg.Layout.Alignment != AlignmentLeft && cfg.Layout.Alignment != AlignmentRight {
		return fmt.Errorf("sankey chart alignment %q is not supported", cfg.Layout.Alignment)
	}
	if cfg.Layout.NodeWidth < 0 {
		return fmt.Errorf("sankey chart node width must be nonnegative")
	}
	if cfg.Layout.NodeGap < 0 {
		return fmt.Errorf("sankey chart node gap must be nonnegative")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("sankey chart series is required")
	}
	for seriesIndex, series := range cfg.Series {
		if err := validateSeries(seriesIndex, series); err != nil {
			return err
		}
	}
	return nil
}

func validateSeries(seriesIndex int, series Series) error {
	if series.Name == "" {
		return fmt.Errorf("sankey chart series %d name is required", seriesIndex)
	}
	if len(series.Nodes) == 0 {
		return fmt.Errorf("sankey chart series %q nodes are required", series.Name)
	}
	nodes := make(map[string]struct{}, len(series.Nodes))
	for nodeIndex, node := range series.Nodes {
		if node.Name == "" {
			return fmt.Errorf("sankey chart series %q node %d name is required", series.Name, nodeIndex)
		}
		if _, exists := nodes[node.Name]; exists {
			return fmt.Errorf("sankey chart series %q node %q is duplicated", series.Name, node.Name)
		}
		if node.Depth != nil && *node.Depth < 0 {
			return fmt.Errorf("sankey chart series %q node %q depth must be nonnegative", series.Name, node.Name)
		}
		nodes[node.Name] = struct{}{}
	}
	if len(series.Links) == 0 {
		return fmt.Errorf("sankey chart series %q links are required", series.Name)
	}
	for linkIndex, link := range series.Links {
		if link.Source == "" {
			return fmt.Errorf("sankey chart series %q link %d source is required", series.Name, linkIndex)
		}
		if link.Target == "" {
			return fmt.Errorf("sankey chart series %q link %d target is required", series.Name, linkIndex)
		}
		if _, exists := nodes[link.Source]; !exists {
			return fmt.Errorf("sankey chart series %q link %d source %q does not name a node", series.Name, linkIndex, link.Source)
		}
		if _, exists := nodes[link.Target]; !exists {
			return fmt.Errorf("sankey chart series %q link %d target %q does not name a node", series.Name, linkIndex, link.Target)
		}
		if !internalinteractive.FiniteNumber(link.Value) || link.Value < 0 {
			return fmt.Errorf("sankey chart series %q link %d value must be a finite nonnegative value", series.Name, linkIndex)
		}
	}
	return nil
}
