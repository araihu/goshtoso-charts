package interactive

import (
	"fmt"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// SankeyOrientation controls the direction in which flow progresses.
type SankeyOrientation string

const (
	// SankeyOrientationHorizontal renders flow from left to right. It is the default.
	SankeyOrientationHorizontal SankeyOrientation = ""
	// SankeyOrientationVertical renders flow from top to bottom.
	SankeyOrientationVertical SankeyOrientation = "vertical"
)

// SankeyAlignment controls how nodes without outgoing links are placed.
type SankeyAlignment string

const (
	// SankeyAlignmentJustify places terminal nodes at the far edge. It is the default.
	SankeyAlignmentJustify SankeyAlignment = ""
	// SankeyAlignmentLeft aligns nodes to the start of the flow.
	SankeyAlignmentLeft SankeyAlignment = "left"
	// SankeyAlignmentRight aligns nodes to the end of the flow.
	SankeyAlignmentRight SankeyAlignment = "right"
)

// SankeyLayout controls renderer-neutral node placement and size.
// NodeWidth and NodeGap use CSS pixels. Zero keeps renderer defaults.
type SankeyLayout struct {
	Orientation SankeyOrientation
	Alignment   SankeyAlignment
	NodeWidth   int
	NodeGap     int
}

// SankeyConfig describes an accessible, browser-rendered Sankey chart.
//
// Nodes and links must be application-owned because the browser renderer
// serializes them.
type SankeyConfig struct {
	Label         string
	Caption       string
	Series        []SankeySeries
	Layout        SankeyLayout
	Width         string
	Height        string
	Options       ChartOptions
	SeriesOptions SeriesOptions
	Style         charttheme.Style
}

// SankeySeries describes one named flow network.
type SankeySeries struct {
	Name    string
	Nodes   []SankeyNode
	Links   []SankeyLink
	Options SeriesOptions
}

// SankeyNode describes one uniquely named point in a flow network.
// Depth optionally pins the zero-based layer; nil lets the renderer place it.
type SankeyNode struct {
	Name      string
	Depth     *int
	ItemStyle *ItemStyle
}

// SankeyLink describes one weighted connection between named nodes.
type SankeyLink struct {
	Source string
	Target string
	Value  float64
}

// Sankey builds a reusable interactive Sankey component.
func Sankey(cfg SankeyConfig) Instance {
	if err := validateSankeyConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveSankey, err)
	}

	chart := charts.NewSankey()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, chartGlobalOptions(cfg.Options)...)
	// Explicit component colors remain authoritative over escape-hatch options.
	if len(cfg.Style.Colors) > 0 {
		globalOptions = append(globalOptions, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
	if cfg.Width != "" || cfg.Height != "" {
		globalOptions = append([]charts.GlobalOpts{
			charts.WithInitializationOpts(opts.Initialization{Width: cfg.Width, Height: cfg.Height}),
		}, globalOptions...)
	}
	chart.SetGlobalOptions(globalOptions...)

	for _, series := range cfg.Series {
		nodes := make([]opts.SankeyNode, len(series.Nodes))
		for index, node := range series.Nodes {
			nodes[index] = opts.SankeyNode{Name: node.Name, Depth: node.Depth}
			if node.ItemStyle != nil {
				style := rendererItemStyle(node.ItemStyle)
				nodes[index].ItemStyle = &style
			}
		}
		links := make([]opts.SankeyLink, len(series.Links))
		for index, link := range series.Links {
			links[index] = opts.SankeyLink{Source: link.Source, Target: link.Target, Value: float32(link.Value)}
		}

		options := make([]charts.SeriesOpts, 0, 1+len(chartSeriesOptions(cfg.SeriesOptions))+len(chartSeriesOptions(series.Options)))
		options = append(options, sankeyLayoutOption(cfg.Layout))
		options = append(options, mergeSeriesOptions(cfg.SeriesOptions, series.Options)...)
		chart.AddSeries(series.Name, nodes, links, options...)
	}

	return newInstance(chartcomponents.KindInteractiveSankey, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, Animation: cfg.Options.Animation,
	})
}

func sankeyLayoutOption(layout SankeyLayout) charts.SeriesOpts {
	return func(series *charts.SingleSeries) {
		series.Orient = "horizontal"
		if layout.Orientation == SankeyOrientationVertical {
			series.Orient = string(layout.Orientation)
		}
		series.NodeAlign = "justify"
		if layout.Alignment != SankeyAlignmentJustify {
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

func validateSankeyConfig(cfg SankeyConfig) error {
	if cfg.Label == "" {
		return fmt.Errorf("sankey chart label is required")
	}
	if cfg.Layout.Orientation != SankeyOrientationHorizontal && cfg.Layout.Orientation != SankeyOrientationVertical {
		return fmt.Errorf("sankey chart orientation %q is not supported", cfg.Layout.Orientation)
	}
	if cfg.Layout.Alignment != SankeyAlignmentJustify && cfg.Layout.Alignment != SankeyAlignmentLeft && cfg.Layout.Alignment != SankeyAlignmentRight {
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
		if err := validateSankeySeries(seriesIndex, series); err != nil {
			return err
		}
	}
	return nil
}

func validateSankeySeries(seriesIndex int, series SankeySeries) error {
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
		if !finiteNumber(link.Value) || link.Value < 0 {
			return fmt.Errorf("sankey chart series %q link %d value must be a finite nonnegative value", series.Name, linkIndex)
		}
	}
	return nil
}
