// Package treemap provides the canonical interactive space-filling hierarchy API.
//
// Treemap-specific types and implementation live here; shared renderer-neutral
// options remain in components/chart.
package treemap

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// Instance is the renderer-neutral chart instance returned by Treemap.
type Instance = chart.Instance

const (
	maxTreemapDepth       = 256
	maxTreemapDetailRows  = 64
	treemapDisabledClick  = "__goshtoso_charts_treemap_navigation_disabled__"
	treemapSeriesTypeJSON = `"type":"treemap"`
)

// Navigation controls rectangle-click navigation.
type Navigation string

const (
	// NavigationDrillDown focuses a clicked branch in the same chart.
	// The breadcrumb returns to an ancestor. It is the default.
	NavigationDrillDown Navigation = ""
	// NavigationDisabled keeps clicks from changing the visible root.
	NavigationDisabled Navigation = "disabled"
)

// Roam controls mouse and touch zooming and translation.
type Roam uint8

const (
	// RoamDisabled disables zooming and translation. It is the default.
	RoamDisabled Roam = iota
	// RoamEnabled enables zooming and translation.
	RoamEnabled
)

// Breadcrumb configures same-instance ancestor navigation.
type Breadcrumb struct {
	Show    *bool
	Height  float64
	ItemGap float64
}

// NodeStyle configures borders and gaps between rectangles.
type NodeStyle struct {
	BorderColor string
	BorderWidth float64
	GapWidth    float64
}

// ColorRange constrains level color saturation.
type ColorRange struct {
	Min float64
	Max float64
}

// Level configures one hierarchy depth.
type Level struct {
	UpperLabel      *chart.LabelOptions
	NodeStyle       NodeStyle
	ColorSaturation *ColorRange
}

// Config describes an accessible space-filling hierarchy.
//
// Label is required and names the figure. Caption remains visible. Parent
// values must remain zero because child values determine parent area.
type Config struct {
	Label        string
	Caption      string
	Nodes        []*Node
	Navigation   Navigation
	Roam         Roam
	LabelOptions *chart.LabelOptions
	UpperLabel   *chart.LabelOptions
	Breadcrumb   *Breadcrumb
	NodeStyle    NodeStyle
	LeafDepth    *int
	Levels       []Level
	Width        string
	Height       string
	Options      chart.ChartOptions
	Style        charttheme.Style
	RootAttrs    templ.Attributes
}

// Node describes one rectangle and its descendants.
//
// Class is a renderer-neutral semantic classification retained in chart data
// and the adjacent hierarchy table. Color optionally sets this node's fill.
type Node struct {
	Name     string
	Value    float64
	Children []*Node
	Class    string
	Color    string
}

// Treemap builds a reusable interactive space-filling hierarchy.
func Treemap(cfg Config) Instance {
	if err := validateConfig(cfg); err != nil {
		return internalinteractive.Invalid(chartcomponents.KindInteractiveTreemap, err)
	}

	chart := charts.NewTreeMap()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, internalinteractive.ChartGlobalOptions(cfg.Options)...)
	if len(cfg.Style.Colors) > 0 {
		globalOptions = append(globalOptions, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
	if cfg.Width != "" || cfg.Height != "" {
		globalOptions = append([]charts.GlobalOpts{
			charts.WithInitializationOpts(opts.Initialization{Width: cfg.Width, Height: cfg.Height}),
		}, globalOptions...)
	}
	chart.SetGlobalOptions(globalOptions...)

	placeholder := make([]opts.TreeMapNode, len(cfg.Nodes))
	seriesOptions := []charts.SeriesOpts{treemapChartOptions(cfg)}
	if cfg.LabelOptions != nil {
		seriesOptions = append(seriesOptions, charts.WithLabelOpts(internalinteractive.RendererLabel(cfg.LabelOptions)))
	}
	if style := rendererTreemapItemStyle(cfg.NodeStyle, ""); style != nil {
		seriesOptions = append(seriesOptions, charts.WithItemStyleOpts(*style))
	}
	seriesOptions = append(seriesOptions, charts.WithSeriesOpts(func(series *charts.SingleSeries) {
		series.NodeClick = rendererNavigation(cfg.Navigation)
	}))
	chart.AddSeries(cfg.Label, placeholder, seriesOptions...)
	nodes := make([]rendererNode, len(cfg.Nodes))
	for index, node := range cfg.Nodes {
		nodes[index] = renderNode(node)
	}
	chart.MultiSeries[len(chart.MultiSeries)-1].Data = nodes

	replacements := []internalinteractive.ScriptReplacement{{
		Old: `"nodeClick":"` + treemapDisabledClick + `"`,
		New: `"nodeClick":false`,
	}}
	if cfg.Breadcrumb != nil {
		breadcrumb, _ := json.Marshal(rendererBreadcrumb(cfg.Breadcrumb))
		replacements = append(replacements, internalinteractive.ScriptReplacement{
			Old: treemapSeriesTypeJSON,
			New: treemapSeriesTypeJSON + `,"breadcrumb":` + string(breadcrumb),
		})
	}

	return internalinteractive.New(chartcomponents.KindInteractiveTreemap, internalinteractive.RenderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, ResponsiveWidth: internalinteractive.ResponsiveWidth(cfg.Width),
		Animation: cfg.Options.Animation, RootAttrs: cfg.RootAttrs,
		Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		Details:            treemapExactValues(flattenNodes(cfg.Nodes, maxTreemapDetailRows)),
		ScriptReplacements: replacements,
	})
}

type rendererNode struct {
	Name      string          `json:"name"`
	Value     *float64        `json:"value,omitempty"`
	ClassName string          `json:"className,omitempty"`
	ItemStyle *opts.ItemStyle `json:"itemStyle,omitempty"`
	Children  []rendererNode  `json:"children,omitempty"`
}

type rendererBreadcrumbOptions struct {
	Show    *bool   `json:"show,omitempty"`
	Height  float64 `json:"height,omitempty"`
	ItemGap float64 `json:"itemGap,omitempty"`
}

func renderNode(node *Node) rendererNode {
	rendered := rendererNode{
		Name: node.Name, ClassName: node.Class,
		ItemStyle: rendererTreemapItemStyle(NodeStyle{}, node.Color),
	}
	if len(node.Children) == 0 {
		rendered.Value = &node.Value
		return rendered
	}
	rendered.Children = make([]rendererNode, len(node.Children))
	for index, child := range node.Children {
		rendered.Children[index] = renderNode(child)
	}
	return rendered
}

func rendererBreadcrumb(value *Breadcrumb) rendererBreadcrumbOptions {
	return rendererBreadcrumbOptions{
		Show: value.Show, Height: value.Height, ItemGap: value.ItemGap,
	}
}

func treemapChartOptions(cfg Config) charts.SeriesOpts {
	chartOptions := opts.TreeMapChart{Roam: opts.Bool(cfg.Roam == RoamEnabled)}
	if cfg.LeafDepth != nil {
		chartOptions.LeafDepth = *cfg.LeafDepth
	}
	if cfg.UpperLabel != nil {
		label := rendererTreemapUpperLabel(cfg.UpperLabel)
		chartOptions.UpperLabel = &label
	}
	if len(cfg.Levels) > 0 {
		levels := make([]opts.TreeMapLevel, len(cfg.Levels))
		for index, level := range cfg.Levels {
			if level.UpperLabel != nil {
				label := rendererTreemapUpperLabel(level.UpperLabel)
				levels[index].UpperLabel = &label
			}
			levels[index].ItemStyle = rendererTreemapItemStyle(level.NodeStyle, "")
			if level.ColorSaturation != nil {
				levels[index].ColorSaturation = []float32{
					float32(level.ColorSaturation.Min),
					float32(level.ColorSaturation.Max),
				}
			}
		}
		chartOptions.Levels = &levels
	}
	return charts.WithTreeMapOpts(chartOptions)
}

func rendererTreemapUpperLabel(value *chart.LabelOptions) opts.UpperLabel {
	result := opts.UpperLabel{
		Position: value.Position,
		Color:    value.Color,
		FontSize: float32(value.FontSize),
	}
	if value.Show != nil {
		result.Show = opts.Bool(*value.Show)
	}
	return result
}

func rendererTreemapItemStyle(value NodeStyle, color string) *opts.ItemStyle {
	if color == "" && value.BorderColor == "" && value.BorderWidth == 0 && value.GapWidth == 0 {
		return nil
	}
	return &opts.ItemStyle{
		Color: color, BorderColor: value.BorderColor,
		BorderWidth: float32(value.BorderWidth), GapWidth: float32(value.GapWidth),
	}
}

func rendererNavigation(value Navigation) string {
	if value == NavigationDisabled {
		return treemapDisabledClick
	}
	return "zoomToNode"
}

func validateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("treemap chart label is required")
	}
	if len(cfg.Nodes) == 0 {
		return fmt.Errorf("treemap chart nodes are required")
	}
	if cfg.Navigation != NavigationDrillDown && cfg.Navigation != NavigationDisabled {
		return fmt.Errorf("treemap chart navigation %q is not supported", cfg.Navigation)
	}
	if cfg.Roam != RoamDisabled && cfg.Roam != RoamEnabled {
		return fmt.Errorf("treemap chart roam mode %d is not supported", cfg.Roam)
	}
	if cfg.LeafDepth != nil && *cfg.LeafDepth <= 0 {
		return fmt.Errorf("treemap chart leaf depth must be positive")
	}
	if cfg.LabelOptions != nil && cfg.LabelOptions.FontSize < 0 {
		return fmt.Errorf("treemap chart label font size must be nonnegative")
	}
	if cfg.UpperLabel != nil && cfg.UpperLabel.FontSize < 0 {
		return fmt.Errorf("treemap chart upper label font size must be nonnegative")
	}
	if cfg.Breadcrumb != nil {
		if !internalinteractive.FiniteNumber(cfg.Breadcrumb.Height) || cfg.Breadcrumb.Height < 0 {
			return fmt.Errorf("treemap chart breadcrumb height must be nonnegative")
		}
		if !internalinteractive.FiniteNumber(cfg.Breadcrumb.ItemGap) || cfg.Breadcrumb.ItemGap < 0 {
			return fmt.Errorf("treemap chart breadcrumb item gap must be nonnegative")
		}
	}
	if err := validateNodeStyle("treemap chart", cfg.NodeStyle); err != nil {
		return err
	}
	for index, level := range cfg.Levels {
		if level.UpperLabel != nil && level.UpperLabel.FontSize < 0 {
			return fmt.Errorf("treemap chart level %d upper label font size must be nonnegative", index)
		}
		if err := validateNodeStyle("treemap chart level "+strconv.Itoa(index), level.NodeStyle); err != nil {
			return err
		}
		if value := level.ColorSaturation; value != nil {
			if !internalinteractive.FiniteNumber(value.Min) || !internalinteractive.FiniteNumber(value.Max) ||
				value.Min < 0 || value.Max > 1 || value.Min > value.Max {
				return fmt.Errorf("treemap chart level %d color saturation range must be between 0 and 1 with min not greater than max", index)
			}
		}
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("treemap chart root attribute %q is reserved", attribute)
			}
		}
	}
	active := make(map[*Node]bool)
	seen := make(map[*Node]bool)
	for index, node := range cfg.Nodes {
		if err := validateNode(node, "root "+strconv.Itoa(index), 0, active, seen); err != nil {
			return err
		}
	}
	return internalinteractive.ValidateChartOptions(cfg.Options)
}

func validateNodeStyle(owner string, value NodeStyle) error {
	if !internalinteractive.FiniteNumber(value.BorderWidth) || value.BorderWidth < 0 {
		return fmt.Errorf("%s border width must be nonnegative", owner)
	}
	if !internalinteractive.FiniteNumber(value.GapWidth) || value.GapWidth < 0 {
		return fmt.Errorf("%s gap width must be nonnegative", owner)
	}
	return nil
}

func validateNode(node *Node, path string, depth int, active, seen map[*Node]bool) error {
	if node == nil {
		return fmt.Errorf("treemap chart %s is nil", path)
	}
	if active[node] {
		return fmt.Errorf("treemap chart %s contains a cycle", path)
	}
	if seen[node] {
		return fmt.Errorf("treemap chart %s reuses a node", path)
	}
	if depth > maxTreemapDepth {
		return fmt.Errorf("treemap chart %s exceeds maximum depth %d", path, maxTreemapDepth)
	}
	if strings.TrimSpace(node.Name) == "" {
		return fmt.Errorf("treemap chart %s name is required", path)
	}
	if !internalinteractive.FiniteNumber(node.Value) {
		return fmt.Errorf("treemap chart node %q value must be finite", node.Name)
	}
	if node.Value < 0 {
		return fmt.Errorf("treemap chart node %q value must be nonnegative", node.Name)
	}
	if len(node.Children) > 0 && node.Value != 0 {
		return fmt.Errorf("treemap chart parent node %q value must be zero; child values determine parent area", node.Name)
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

type treemapValueRow struct {
	Path   string
	Parent string
	Value  string
	Class  string
}

type treemapValueRows struct {
	Rows    []treemapValueRow
	Omitted int
}

func flattenNodes(nodes []*Node, limit int) treemapValueRows {
	result := treemapValueRows{Rows: make([]treemapValueRow, 0, min(limit, len(nodes)))}
	var visit func(*Node, []string) float64
	visit = func(node *Node, parents []string) float64 {
		total := node.Value
		if len(node.Children) > 0 {
			total = 0
			for _, child := range node.Children {
				total += visit(child, append(parents, node.Name))
			}
		}
		if len(result.Rows) >= limit {
			result.Omitted++
			return total
		}
		parent := "—"
		if len(parents) > 0 {
			parent = parents[len(parents)-1]
		}
		path := append(append([]string(nil), parents...), node.Name)
		result.Rows = append(result.Rows, treemapValueRow{
			Path: strings.Join(path, " / "), Parent: parent,
			Value: strconv.FormatFloat(total, 'f', -1, 64), Class: node.Class,
		})
		return total
	}
	for _, node := range nodes {
		visit(node, nil)
	}
	return result
}
