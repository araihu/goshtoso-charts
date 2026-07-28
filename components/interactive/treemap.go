package interactive

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

const (
	maxTreemapDepth       = 256
	maxTreemapDetailRows  = 64
	treemapDisabledClick  = "__goshtoso_charts_treemap_navigation_disabled__"
	treemapSeriesTypeJSON = `"type":"treemap"`
)

// TreemapNavigation controls rectangle-click navigation.
type TreemapNavigation string

const (
	// TreemapNavigationDrillDown focuses a clicked branch in the same chart.
	// The breadcrumb returns to an ancestor. It is the default.
	TreemapNavigationDrillDown TreemapNavigation = ""
	// TreemapNavigationDisabled keeps clicks from changing the visible root.
	TreemapNavigationDisabled TreemapNavigation = "disabled"
)

// TreemapRoam controls mouse and touch zooming and translation.
type TreemapRoam uint8

const (
	// TreemapRoamDisabled disables zooming and translation. It is the default.
	TreemapRoamDisabled TreemapRoam = iota
	// TreemapRoamEnabled enables zooming and translation.
	TreemapRoamEnabled
)

// TreemapBreadcrumb configures same-instance ancestor navigation.
type TreemapBreadcrumb struct {
	Show    *bool
	Height  float64
	ItemGap float64
}

// TreemapNodeStyle configures borders and gaps between rectangles.
type TreemapNodeStyle struct {
	BorderColor string
	BorderWidth float64
	GapWidth    float64
}

// TreemapColorRange constrains level color saturation.
type TreemapColorRange struct {
	Min float64
	Max float64
}

// TreemapLevel configures one hierarchy depth.
type TreemapLevel struct {
	UpperLabel      *LabelOptions
	NodeStyle       TreemapNodeStyle
	ColorSaturation *TreemapColorRange
}

// TreemapConfig describes an accessible space-filling hierarchy.
//
// Label is required and names the figure. Caption remains visible. Parent
// values must remain zero because child values determine parent area.
type TreemapConfig struct {
	Label        string
	Caption      string
	Nodes        []*TreemapNode
	Navigation   TreemapNavigation
	Roam         TreemapRoam
	LabelOptions *LabelOptions
	UpperLabel   *LabelOptions
	Breadcrumb   *TreemapBreadcrumb
	NodeStyle    TreemapNodeStyle
	LeafDepth    *int
	Levels       []TreemapLevel
	Width        string
	Height       string
	Options      ChartOptions
	Style        charttheme.Style
	RootAttrs    templ.Attributes
}

// TreemapNode describes one rectangle and its descendants.
//
// Class is a renderer-neutral semantic classification retained in chart data
// and the adjacent hierarchy table. Color optionally sets this node's fill.
type TreemapNode struct {
	Name     string
	Value    float64
	Children []*TreemapNode
	Class    string
	Color    string
}

// Treemap builds a reusable interactive space-filling hierarchy.
func Treemap(cfg TreemapConfig) Instance {
	if err := validateTreemapConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveTreemap, err)
	}

	chart := charts.NewTreeMap()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, chartGlobalOptions(cfg.Options)...)
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
		seriesOptions = append(seriesOptions, charts.WithLabelOpts(rendererLabel(cfg.LabelOptions)))
	}
	if style := rendererTreemapItemStyle(cfg.NodeStyle, ""); style != nil {
		seriesOptions = append(seriesOptions, charts.WithItemStyleOpts(*style))
	}
	seriesOptions = append(seriesOptions, charts.WithSeriesOpts(func(series *charts.SingleSeries) {
		series.NodeClick = rendererTreemapNavigation(cfg.Navigation)
	}))
	chart.AddSeries(cfg.Label, placeholder, seriesOptions...)
	nodes := make([]rendererTreemapNode, len(cfg.Nodes))
	for index, node := range cfg.Nodes {
		nodes[index] = renderTreemapNode(node)
	}
	chart.MultiSeries[len(chart.MultiSeries)-1].Data = nodes

	replacements := []scriptReplacement{{
		Old: `"nodeClick":"` + treemapDisabledClick + `"`,
		New: `"nodeClick":false`,
	}}
	if cfg.Breadcrumb != nil {
		breadcrumb, _ := json.Marshal(rendererTreemapBreadcrumb(cfg.Breadcrumb))
		replacements = append(replacements, scriptReplacement{
			Old: treemapSeriesTypeJSON,
			New: treemapSeriesTypeJSON + `,"breadcrumb":` + string(breadcrumb),
		})
	}

	return newInstance(chartcomponents.KindInteractiveTreemap, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style,
		Animation: cfg.Options.Animation, RootAttrs: cfg.RootAttrs,
		Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		Details:            treemapExactValues(flattenTreemapNodes(cfg.Nodes, maxTreemapDetailRows)),
		ScriptReplacements: replacements,
	})
}

type rendererTreemapNode struct {
	Name      string                `json:"name"`
	Value     *float64              `json:"value,omitempty"`
	ClassName string                `json:"className,omitempty"`
	ItemStyle *opts.ItemStyle       `json:"itemStyle,omitempty"`
	Children  []rendererTreemapNode `json:"children,omitempty"`
}

type rendererTreemapBreadcrumbOptions struct {
	Show    *bool   `json:"show,omitempty"`
	Height  float64 `json:"height,omitempty"`
	ItemGap float64 `json:"itemGap,omitempty"`
}

func renderTreemapNode(node *TreemapNode) rendererTreemapNode {
	rendered := rendererTreemapNode{
		Name: node.Name, ClassName: node.Class,
		ItemStyle: rendererTreemapItemStyle(TreemapNodeStyle{}, node.Color),
	}
	if len(node.Children) == 0 {
		rendered.Value = &node.Value
		return rendered
	}
	rendered.Children = make([]rendererTreemapNode, len(node.Children))
	for index, child := range node.Children {
		rendered.Children[index] = renderTreemapNode(child)
	}
	return rendered
}

func rendererTreemapBreadcrumb(value *TreemapBreadcrumb) rendererTreemapBreadcrumbOptions {
	return rendererTreemapBreadcrumbOptions{
		Show: value.Show, Height: value.Height, ItemGap: value.ItemGap,
	}
}

func treemapChartOptions(cfg TreemapConfig) charts.SeriesOpts {
	chartOptions := opts.TreeMapChart{Roam: opts.Bool(cfg.Roam == TreemapRoamEnabled)}
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

func rendererTreemapUpperLabel(value *LabelOptions) opts.UpperLabel {
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

func rendererTreemapItemStyle(value TreemapNodeStyle, color string) *opts.ItemStyle {
	if color == "" && value.BorderColor == "" && value.BorderWidth == 0 && value.GapWidth == 0 {
		return nil
	}
	return &opts.ItemStyle{
		Color: color, BorderColor: value.BorderColor,
		BorderWidth: float32(value.BorderWidth), GapWidth: float32(value.GapWidth),
	}
}

func rendererTreemapNavigation(value TreemapNavigation) string {
	if value == TreemapNavigationDisabled {
		return treemapDisabledClick
	}
	return "zoomToNode"
}

func validateTreemapConfig(cfg TreemapConfig) error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("treemap chart label is required")
	}
	if len(cfg.Nodes) == 0 {
		return fmt.Errorf("treemap chart nodes are required")
	}
	if cfg.Navigation != TreemapNavigationDrillDown && cfg.Navigation != TreemapNavigationDisabled {
		return fmt.Errorf("treemap chart navigation %q is not supported", cfg.Navigation)
	}
	if cfg.Roam != TreemapRoamDisabled && cfg.Roam != TreemapRoamEnabled {
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
		if !finiteNumber(cfg.Breadcrumb.Height) || cfg.Breadcrumb.Height < 0 {
			return fmt.Errorf("treemap chart breadcrumb height must be nonnegative")
		}
		if !finiteNumber(cfg.Breadcrumb.ItemGap) || cfg.Breadcrumb.ItemGap < 0 {
			return fmt.Errorf("treemap chart breadcrumb item gap must be nonnegative")
		}
	}
	if err := validateTreemapNodeStyle("treemap chart", cfg.NodeStyle); err != nil {
		return err
	}
	for index, level := range cfg.Levels {
		if level.UpperLabel != nil && level.UpperLabel.FontSize < 0 {
			return fmt.Errorf("treemap chart level %d upper label font size must be nonnegative", index)
		}
		if err := validateTreemapNodeStyle("treemap chart level "+strconv.Itoa(index), level.NodeStyle); err != nil {
			return err
		}
		if value := level.ColorSaturation; value != nil {
			if !finiteNumber(value.Min) || !finiteNumber(value.Max) ||
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
	active := make(map[*TreemapNode]bool)
	seen := make(map[*TreemapNode]bool)
	for index, node := range cfg.Nodes {
		if err := validateTreemapNode(node, "root "+strconv.Itoa(index), 0, active, seen); err != nil {
			return err
		}
	}
	return validateChartOptions(cfg.Options)
}

func validateTreemapNodeStyle(owner string, value TreemapNodeStyle) error {
	if !finiteNumber(value.BorderWidth) || value.BorderWidth < 0 {
		return fmt.Errorf("%s border width must be nonnegative", owner)
	}
	if !finiteNumber(value.GapWidth) || value.GapWidth < 0 {
		return fmt.Errorf("%s gap width must be nonnegative", owner)
	}
	return nil
}

func validateTreemapNode(node *TreemapNode, path string, depth int, active, seen map[*TreemapNode]bool) error {
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
	if !finiteNumber(node.Value) {
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
		if err := validateTreemapNode(child, childPath, depth+1, active, seen); err != nil {
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

func flattenTreemapNodes(nodes []*TreemapNode, limit int) treemapValueRows {
	result := treemapValueRows{Rows: make([]treemapValueRow, 0, min(limit, len(nodes)))}
	var visit func(*TreemapNode, []string) float64
	visit = func(node *TreemapNode, parents []string) float64 {
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
