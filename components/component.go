// Package components defines stable identities shared by Goshtoso Charts components.
package components

import (
	"slices"

	"github.com/a-h/templ"
)

// Kind identifies a Goshtoso Charts component type.
type Kind string

// Component is a renderable chart component with stable identity.
type Component interface {
	templ.Component
	Kind() Kind
}

const (
	// KindLineChart identifies a quantitative time-series line chart.
	KindLineChart Kind = "line-chart"
	// KindBarChart identifies a quantitative categorical bar chart.
	KindBarChart Kind = "bar-chart"
	// KindPieChart identifies a categorical proportional breakdown chart.
	KindPieChart Kind = "pie-chart"
	// KindScatterChart identifies a server-rendered categorical scatter chart.
	KindScatterChart Kind = "scatter-chart"
	// KindRadarChart identifies a server-rendered multivariate radar chart.
	KindRadarChart Kind = "radar-chart"
	// KindCandlestickChart identifies a server-rendered OHLC candlestick chart.
	KindCandlestickChart Kind = "candlestick-chart"
	// KindHeatMapChart identifies a server-rendered categorical heat map.
	KindHeatMapChart Kind = "heatmap-chart"
	// KindFunnelChart identifies a server-rendered ordered funnel chart.
	KindFunnelChart Kind = "funnel-chart"
	// KindTable identifies a server-rendered data table.
	KindTable Kind = "table"
	// KindViolinChart identifies a server-rendered distribution violin chart.
	KindViolinChart Kind = "violin-chart"
	// KindInteractiveBar identifies a browser-rendered bar chart.
	KindInteractiveBar Kind = "interactive-bar"
	// KindInteractiveLine identifies a browser-rendered line chart.
	KindInteractiveLine Kind = "interactive-line"
	// KindInteractiveScatter identifies a browser-rendered scatter chart.
	KindInteractiveScatter Kind = "interactive-scatter"
	// KindInteractivePie identifies a browser-rendered pie chart.
	KindInteractivePie Kind = "interactive-pie"
	// KindInteractiveRadar identifies a browser-rendered radar chart.
	KindInteractiveRadar Kind = "interactive-radar"
	// KindInteractiveHeatMap identifies a browser-rendered heat map chart.
	KindInteractiveHeatMap Kind = "interactive-heatmap"
	// KindInteractiveBoxPlot identifies a browser-rendered box plot chart.
	KindInteractiveBoxPlot Kind = "interactive-boxplot"
	// KindInteractiveGauge identifies a browser-rendered gauge chart.
	KindInteractiveGauge Kind = "interactive-gauge"
	// KindInteractiveFunnel identifies a browser-rendered funnel chart.
	KindInteractiveFunnel Kind = "interactive-funnel"
	// KindDependencies identifies the browser runtime dependency set.
	KindDependencies Kind = "dependencies"
	// KindInteractiveGraph identifies a browser-rendered relationship graph.
	KindInteractiveGraph Kind = "interactive-graph"
	// KindInteractiveSankey identifies a browser-rendered weighted flow chart.
	KindInteractiveSankey Kind = "interactive-sankey"
	// KindInteractiveTree identifies a browser-rendered hierarchical tree.
	KindInteractiveTree Kind = "interactive-tree"
	// KindInteractiveSunburst identifies a browser-rendered radial hierarchy.
	KindInteractiveSunburst Kind = "interactive-sunburst"
	// KindInteractiveTreemap identifies a browser-rendered space-filling hierarchy.
	KindInteractiveTreemap Kind = "interactive-treemap"
	// KindInteractiveParallel identifies a browser-rendered parallel-coordinates chart.
	KindInteractiveParallel Kind = "interactive-parallel"
	// KindInteractiveThemeRiver identifies a browser-rendered temporal stream graph.
	KindInteractiveThemeRiver Kind = "interactive-theme-river"
	// KindInteractiveWordCloud identifies a browser-rendered weighted word cloud.
	KindInteractiveWordCloud Kind = "interactive-word-cloud"
)

var allKinds = []Kind{KindLineChart, KindBarChart, KindPieChart, KindScatterChart, KindRadarChart, KindCandlestickChart, KindHeatMapChart, KindFunnelChart, KindTable, KindViolinChart, KindInteractiveBar, KindInteractiveLine, KindInteractiveScatter, KindInteractivePie, KindInteractiveRadar, KindInteractiveHeatMap, KindInteractiveBoxPlot, KindInteractiveGauge, KindInteractiveFunnel, KindInteractiveGraph, KindInteractiveSankey, KindInteractiveTree, KindInteractiveSunburst, KindInteractiveTreemap, KindInteractiveParallel, KindInteractiveThemeRiver, KindInteractiveWordCloud}

// AllKinds returns supported chart kinds in stable order. Document-level
// dependency components are intentionally excluded.
func AllKinds() []Kind {
	return slices.Clone(allKinds)
}
