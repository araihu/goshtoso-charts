// Package sunburst provides the canonical interactive radial-hierarchy API.
//
// Sunburst-specific types and implementation live here; shared
// renderer-neutral options remain in components/chart.
package sunburst

import (
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

const (
	maxDepth                = 256
	disabledNavigationValue = "__goshtoso_charts_sunburst_navigation_disabled__"
	inputSortValue          = "__goshtoso_charts_sunburst_input_sort__"
	zeroValueSentinel       = -1
)

// Instance is the renderer-neutral chart instance returned by Sunburst.
type Instance = chart.Instance

// Navigation controls sector-click navigation.
type Navigation string

const (
	// NavigationDrillDown makes a clicked sector the visible root. It is the default.
	// Selecting the center returns to the previous root.
	NavigationDrillDown Navigation = ""
	// NavigationDisabled keeps the full hierarchy visible when sectors are clicked.
	NavigationDisabled Navigation = "disabled"
)

// Sort controls sibling-sector ordering.
type Sort string

const (
	// SortDescending orders sibling sectors from largest to smallest. It is the default.
	SortDescending Sort = ""
	// SortAscending orders sibling sectors from smallest to largest.
	SortAscending Sort = "ascending"
	// SortInput preserves caller order.
	SortInput Sort = "input"
)

// Config describes an accessible radial hierarchy.
//
// Label is required and names the figure. Caption remains visible. Nodes are
// application-owned because the browser renderer serializes them. Style accepts
// theme palette, explicit colors, and caller root classes. RootAttrs accepts
// additional figure attributes except class, role, and aria-label.
type Config struct {
	Label             string
	Caption           string
	Nodes             []*Node
	Navigation        Navigation
	Sort              Sort
	ShowLabelsForZero *bool
	LabelOptions      *chart.LabelOptions
	ItemStyle         *chart.ItemStyle
	InnerRadius       float64
	OuterRadius       float64
	Width             string
	Height            string
	Options           chart.ChartOptions
	Style             charttheme.Style
	RootAttrs         templ.Attributes
}

// Node describes one weighted sector and its descendants.
type Node struct {
	Name      string
	Value     float64
	Children  []*Node
	Label     *chart.LabelOptions
	ItemStyle *chart.ItemStyle
}

// Sunburst builds a reusable interactive radial hierarchy.
func Sunburst(cfg Config) Instance {
	if err := validateConfig(cfg); err != nil {
		return internalinteractive.Invalid(chartcomponents.KindInteractiveSunburst, err)
	}

	sunburstChart := charts.NewSunburst()
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
	sunburstChart.SetGlobalOptions(globalOptions...)

	nodes := make([]opts.SunBurstData, len(cfg.Nodes))
	for index, node := range cfg.Nodes {
		nodes[index] = rendererNode(node)
	}
	sunburstOptions := opts.SunburstChart{
		NodeClick: rendererNavigation(cfg.Navigation),
		Sort:      rendererSort(cfg.Sort),
	}
	if cfg.ShowLabelsForZero != nil {
		sunburstOptions.RenderLabelForZeroData = opts.Bool(*cfg.ShowLabelsForZero)
	}
	seriesOptions := []charts.SeriesOpts{charts.WithSunburstOpts(sunburstOptions)}
	if cfg.LabelOptions != nil {
		seriesOptions = append(seriesOptions, charts.WithLabelOpts(internalinteractive.RendererLabel(cfg.LabelOptions)))
	}
	if cfg.ItemStyle != nil {
		seriesOptions = append(seriesOptions, charts.WithItemStyleOpts(internalinteractive.RendererItemStyle(cfg.ItemStyle)))
	}
	outerRadius := cfg.OuterRadius
	if outerRadius == 0 {
		outerRadius = 75
	}
	seriesOptions = append(seriesOptions, func(series *charts.SingleSeries) {
		series.Radius = []string{internalinteractive.Percentage(cfg.InnerRadius), internalinteractive.Percentage(outerRadius)}
	})
	sunburstChart.AddSeries(cfg.Label, nodes, seriesOptions...)

	return internalinteractive.New(chartcomponents.KindInteractiveSunburst, internalinteractive.RenderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: sunburstChart, Style: cfg.Style, ResponsiveWidth: internalinteractive.ResponsiveWidth(cfg.Width),
		Animation: cfg.Options.Animation, RootAttrs: cfg.RootAttrs,
		Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		Details: sunburstExactValues(flattenNodes(cfg.Nodes)),
		ScriptReplacements: []internalinteractive.ScriptReplacement{
			{Old: `"nodeClick":"` + disabledNavigationValue + `"`, New: `"nodeClick":false`},
			{Old: `"sort":"` + inputSortValue + `"`, New: `"sort":null`},
			{Old: `"value":` + strconv.Itoa(zeroValueSentinel), New: `"value":0`},
		},
	})
}

func rendererNode(node *Node) opts.SunBurstData {
	value := node.Value
	if value == 0 {
		value = zeroValueSentinel
	}
	rendered := opts.SunBurstData{Name: node.Name, Value: value}
	if node.Label != nil {
		label := internalinteractive.RendererLabel(node.Label)
		rendered.Label = &label
	}
	if node.ItemStyle != nil {
		style := internalinteractive.RendererItemStyle(node.ItemStyle)
		rendered.ItemStyle = &style
	}
	if len(node.Children) > 0 {
		rendered.Children = make([]*opts.SunBurstData, len(node.Children))
		for index, child := range node.Children {
			value := rendererNode(child)
			rendered.Children[index] = &value
		}
	}
	return rendered
}

func rendererNavigation(value Navigation) string {
	if value == NavigationDisabled {
		return disabledNavigationValue
	}
	return "rootToNode"
}

func rendererSort(value Sort) string {
	switch value {
	case SortAscending:
		return "asc"
	case SortInput:
		return inputSortValue
	default:
		return "desc"
	}
}

func validateConfig(cfg Config) error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("sunburst chart label is required")
	}
	if len(cfg.Nodes) == 0 {
		return fmt.Errorf("sunburst chart nodes are required")
	}
	if cfg.Navigation != NavigationDrillDown && cfg.Navigation != NavigationDisabled {
		return fmt.Errorf("sunburst chart navigation %q is not supported", cfg.Navigation)
	}
	if cfg.Sort != SortDescending && cfg.Sort != SortAscending && cfg.Sort != SortInput {
		return fmt.Errorf("sunburst chart sort %q is not supported", cfg.Sort)
	}
	if cfg.LabelOptions != nil && cfg.LabelOptions.FontSize < 0 {
		return fmt.Errorf("sunburst chart label font size must be nonnegative")
	}
	if !internalinteractive.ValidPercentage(cfg.InnerRadius) {
		return fmt.Errorf("sunburst chart inner radius must be between 0 and 100")
	}
	if !internalinteractive.ValidPercentage(cfg.OuterRadius) {
		return fmt.Errorf("sunburst chart outer radius must be between 0 and 100")
	}
	outerRadius := cfg.OuterRadius
	if outerRadius == 0 {
		outerRadius = 75
	}
	if cfg.InnerRadius >= outerRadius {
		return fmt.Errorf("sunburst chart inner radius must be less than outer radius")
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("sunburst chart root attribute %q is reserved", attribute)
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

func validateNode(node *Node, path string, depth int, active, seen map[*Node]bool) error {
	if node == nil {
		return fmt.Errorf("sunburst chart %s is nil", path)
	}
	if active[node] {
		return fmt.Errorf("sunburst chart %s contains a cycle", path)
	}
	if seen[node] {
		return fmt.Errorf("sunburst chart %s reuses a node", path)
	}
	if depth > maxDepth {
		return fmt.Errorf("sunburst chart %s exceeds maximum depth %d", path, maxDepth)
	}
	if strings.TrimSpace(node.Name) == "" {
		return fmt.Errorf("sunburst chart %s name is required", path)
	}
	if !internalinteractive.FiniteNumber(node.Value) {
		return fmt.Errorf("sunburst chart node %q value must be finite", node.Name)
	}
	if node.Value < 0 {
		return fmt.Errorf("sunburst chart node %q value must be nonnegative", node.Name)
	}
	if node.Label != nil && node.Label.FontSize < 0 {
		return fmt.Errorf("sunburst chart node %q label font size must be nonnegative", node.Name)
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

type valueRow struct {
	Path   string
	Parent string
	Value  string
}

func flattenNodes(nodes []*Node) []valueRow {
	rows := make([]valueRow, 0)
	var appendNode func(*Node, []string)
	appendNode = func(node *Node, parents []string) {
		path := append(append([]string(nil), parents...), node.Name)
		parent := "—"
		if len(parents) > 0 {
			parent = parents[len(parents)-1]
		}
		rows = append(rows, valueRow{
			Path: strings.Join(path, " / "), Parent: parent,
			Value: strconv.FormatFloat(node.Value, 'f', -1, 64),
		})
		for _, child := range node.Children {
			appendNode(child, path)
		}
	}
	for _, node := range nodes {
		appendNode(node, nil)
	}
	return rows
}
