package interactive

import (
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
	maxSunburstDepth                = 256
	sunburstDisabledNavigationValue = "__goshtoso_charts_sunburst_navigation_disabled__"
	sunburstInputSortValue          = "__goshtoso_charts_sunburst_input_sort__"
	sunburstZeroValueSentinel       = -1
)

// SunburstNavigation controls sector-click navigation.
type SunburstNavigation string

const (
	// SunburstNavigationDrillDown makes a clicked sector the visible root. It is the default.
	// Selecting the center returns to the previous root.
	SunburstNavigationDrillDown SunburstNavigation = ""
	// SunburstNavigationDisabled keeps the full hierarchy visible when sectors are clicked.
	SunburstNavigationDisabled SunburstNavigation = "disabled"
)

// SunburstSort controls sibling-sector ordering.
type SunburstSort string

const (
	// SunburstSortDescending orders sibling sectors from largest to smallest. It is the default.
	SunburstSortDescending SunburstSort = ""
	// SunburstSortAscending orders sibling sectors from smallest to largest.
	SunburstSortAscending SunburstSort = "ascending"
	// SunburstSortInput preserves caller order.
	SunburstSortInput SunburstSort = "input"
)

// SunburstConfig describes an accessible radial hierarchy.
//
// Label is required and names the figure. Caption remains visible. Nodes are
// application-owned because the browser renderer serializes them. Style accepts
// theme palette, explicit colors, and caller root classes. RootAttrs accepts
// additional figure attributes except class, role, and aria-label.
type SunburstConfig struct {
	Label             string
	Caption           string
	Nodes             []*SunburstNode
	Navigation        SunburstNavigation
	Sort              SunburstSort
	ShowLabelsForZero *bool
	LabelOptions      *LabelOptions
	ItemStyle         *ItemStyle
	InnerRadius       float64
	OuterRadius       float64
	Width             string
	Height            string
	Options           ChartOptions
	Style             charttheme.Style
	RootAttrs         templ.Attributes
}

// SunburstNode describes one weighted sector and its descendants.
type SunburstNode struct {
	Name      string
	Value     float64
	Children  []*SunburstNode
	Label     *LabelOptions
	ItemStyle *ItemStyle
}

// Sunburst builds a reusable interactive radial hierarchy.
func Sunburst(cfg SunburstConfig) Instance {
	if err := validateSunburstConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveSunburst, err)
	}

	chart := charts.NewSunburst()
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

	nodes := make([]opts.SunBurstData, len(cfg.Nodes))
	for index, node := range cfg.Nodes {
		nodes[index] = rendererSunburstNode(node)
	}
	sunburstOptions := opts.SunburstChart{
		NodeClick: rendererSunburstNavigation(cfg.Navigation),
		Sort:      rendererSunburstSort(cfg.Sort),
	}
	if cfg.ShowLabelsForZero != nil {
		sunburstOptions.RenderLabelForZeroData = opts.Bool(*cfg.ShowLabelsForZero)
	}
	seriesOptions := []charts.SeriesOpts{charts.WithSunburstOpts(sunburstOptions)}
	if cfg.LabelOptions != nil {
		seriesOptions = append(seriesOptions, charts.WithLabelOpts(rendererLabel(cfg.LabelOptions)))
	}
	if cfg.ItemStyle != nil {
		seriesOptions = append(seriesOptions, charts.WithItemStyleOpts(rendererItemStyle(cfg.ItemStyle)))
	}
	outerRadius := cfg.OuterRadius
	if outerRadius == 0 {
		outerRadius = 75
	}
	seriesOptions = append(seriesOptions, func(series *charts.SingleSeries) {
		series.Radius = []string{percentage(cfg.InnerRadius), percentage(outerRadius)}
	})
	chart.AddSeries(cfg.Label, nodes, seriesOptions...)

	return newInstance(chartcomponents.KindInteractiveSunburst, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, ResponsiveWidth: responsiveWidth(cfg.Width),
		Animation: cfg.Options.Animation, RootAttrs: cfg.RootAttrs,
		Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		Details: sunburstExactValues(flattenSunburstNodes(cfg.Nodes)),
		ScriptReplacements: []scriptReplacement{
			{Old: `"nodeClick":"` + sunburstDisabledNavigationValue + `"`, New: `"nodeClick":false`},
			{Old: `"sort":"` + sunburstInputSortValue + `"`, New: `"sort":null`},
			{Old: `"value":` + strconv.Itoa(sunburstZeroValueSentinel), New: `"value":0`},
		},
	})
}

func rendererSunburstNode(node *SunburstNode) opts.SunBurstData {
	value := node.Value
	if value == 0 {
		value = sunburstZeroValueSentinel
	}
	rendered := opts.SunBurstData{Name: node.Name, Value: value}
	if node.Label != nil {
		label := rendererLabel(node.Label)
		rendered.Label = &label
	}
	if node.ItemStyle != nil {
		style := rendererItemStyle(node.ItemStyle)
		rendered.ItemStyle = &style
	}
	if len(node.Children) > 0 {
		rendered.Children = make([]*opts.SunBurstData, len(node.Children))
		for index, child := range node.Children {
			value := rendererSunburstNode(child)
			rendered.Children[index] = &value
		}
	}
	return rendered
}

func rendererSunburstNavigation(value SunburstNavigation) string {
	if value == SunburstNavigationDisabled {
		return sunburstDisabledNavigationValue
	}
	return "rootToNode"
}

func rendererSunburstSort(value SunburstSort) string {
	switch value {
	case SunburstSortAscending:
		return "asc"
	case SunburstSortInput:
		return sunburstInputSortValue
	default:
		return "desc"
	}
}

func validateSunburstConfig(cfg SunburstConfig) error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("sunburst chart label is required")
	}
	if len(cfg.Nodes) == 0 {
		return fmt.Errorf("sunburst chart nodes are required")
	}
	if cfg.Navigation != SunburstNavigationDrillDown && cfg.Navigation != SunburstNavigationDisabled {
		return fmt.Errorf("sunburst chart navigation %q is not supported", cfg.Navigation)
	}
	if cfg.Sort != SunburstSortDescending && cfg.Sort != SunburstSortAscending && cfg.Sort != SunburstSortInput {
		return fmt.Errorf("sunburst chart sort %q is not supported", cfg.Sort)
	}
	if cfg.LabelOptions != nil && cfg.LabelOptions.FontSize < 0 {
		return fmt.Errorf("sunburst chart label font size must be nonnegative")
	}
	if !validPercentage(cfg.InnerRadius) {
		return fmt.Errorf("sunburst chart inner radius must be between 0 and 100")
	}
	if !validPercentage(cfg.OuterRadius) {
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
	active := make(map[*SunburstNode]bool)
	seen := make(map[*SunburstNode]bool)
	for index, node := range cfg.Nodes {
		if err := validateSunburstNode(node, "root "+strconv.Itoa(index), 0, active, seen); err != nil {
			return err
		}
	}
	return validateChartOptions(cfg.Options)
}

func validateSunburstNode(node *SunburstNode, path string, depth int, active, seen map[*SunburstNode]bool) error {
	if node == nil {
		return fmt.Errorf("sunburst chart %s is nil", path)
	}
	if active[node] {
		return fmt.Errorf("sunburst chart %s contains a cycle", path)
	}
	if seen[node] {
		return fmt.Errorf("sunburst chart %s reuses a node", path)
	}
	if depth > maxSunburstDepth {
		return fmt.Errorf("sunburst chart %s exceeds maximum depth %d", path, maxSunburstDepth)
	}
	if strings.TrimSpace(node.Name) == "" {
		return fmt.Errorf("sunburst chart %s name is required", path)
	}
	if !finiteNumber(node.Value) {
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
		if err := validateSunburstNode(child, childPath, depth+1, active, seen); err != nil {
			return err
		}
	}
	delete(active, node)
	seen[node] = true
	return nil
}

type sunburstValueRow struct {
	Path   string
	Parent string
	Value  string
}

func flattenSunburstNodes(nodes []*SunburstNode) []sunburstValueRow {
	rows := make([]sunburstValueRow, 0)
	var appendNode func(*SunburstNode, []string)
	appendNode = func(node *SunburstNode, parents []string) {
		path := append(append([]string(nil), parents...), node.Name)
		parent := "—"
		if len(parents) > 0 {
			parent = parents[len(parents)-1]
		}
		rows = append(rows, sunburstValueRow{
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
