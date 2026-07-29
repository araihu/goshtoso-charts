package pages

import (
	"fmt"
	"strings"
	"time"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	interactive "github.com/araihu/goshtoso-charts/components/interactive"
)

func controlledOptions(title, filename string) interactive.ChartOptions {
	return interactive.ChartOptions{
		Title:    &interactive.TitleOptions{Text: title},
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: filename},
	}
}

func sampleInteractiveBar() interactive.BarConfig {
	return interactive.BarConfig{
		Label: "Weekly deployments by environment", Caption: "Interactive bar component.",
		XAxis: []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
		Series: []interactive.BarSeries{
			{Name: "Staging", Data: []interactive.BarData{{Value: 8}, {Value: 12}, {Value: 9}, {Value: 14}, {Value: 11}}},
			{Name: "Production", Data: []interactive.BarData{{Value: 3}, {Value: 5}, {Value: 4}, {Value: 6}, {Value: 7}}},
		},
		Options: controlledOptions("Deployments", "interactive-deployments"),
	}
}

func sampleInteractiveLine() interactive.LineConfig {
	return interactive.LineConfig{
		Label: "Weekly latency trend", Caption: "Interactive line component.",
		XAxis:   []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
		Series:  []interactive.LineSeries{{Name: "Latency (ms)", Data: []interactive.LineData{{Value: 42}, {Value: 47}, {Value: 45}, {Value: 51}, {Value: 44}}}},
		Options: controlledOptions("Latency", "interactive-latency"),
	}
}

func sampleInteractiveScatter() interactive.ScatterConfig {
	return interactive.ScatterConfig{
		Label: "Release impact", Caption: "Ripple highlights the highest-impact releases.",
		Variant: interactive.ScatterVariantEffect,
		XAxis:   []string{"v1.8", "v1.9", "v2.0", "v2.1"},
		Series: []interactive.ScatterSeries{{
			Name: "Impact",
			Data: []interactive.ScatterData{{Value: 35}, {Value: 52}, {Value: 91}, {Value: 64}},
		}},
		Ripple:  &interactive.RippleOptions{Scale: 6, BrushType: "stroke"},
		Options: controlledOptions("Release impact", "release-impact"),
	}
}

func sampleInteractiveRadar() interactive.RadarConfig {
	return interactive.RadarConfig{
		Label: "Service health profile", Caption: "Current measurements compared with the target.",
		Indicators: []interactive.RadarIndicator{
			{Name: "Availability", Max: 100},
			{Name: "Latency", Max: 500},
			{Name: "Capacity", Max: 200},
			{Name: "Recovery", Max: 60},
		},
		Series: []interactive.RadarSeries{{
			Name: "Profile",
			Data: []interactive.RadarData{
				{Name: "Current", Values: []float64{99.8, 180, 124, 34}},
				{Name: "Target", Values: []float64{100, 100, 165, 20}},
			},
		}},
		Options: controlledOptions("Service health", "service-health"),
	}
}

func sampleInteractiveHeatMap() interactive.HeatMapConfig {
	return interactive.HeatMapConfig{
		Label: "Deployment activity", Caption: "Deployments by environment and time window.",
		XAxis:      []string{"08:00", "10:00", "12:00", "14:00", "16:00"},
		YAxis:      []string{"Development", "Staging", "Production"},
		ValueRange: interactive.HeatMapValueRange{Min: 0, Max: 20},
		Series: []interactive.HeatMapSeries{{
			Name: "Deployments",
			Data: []interactive.HeatMapData{
				{X: 0, Y: 0, Value: 2}, {X: 1, Y: 0, Value: 10}, {X: 2, Y: 0, Value: 20},
				{X: 1, Y: 1, Value: 5}, {X: 2, Y: 1, Value: 12}, {X: 3, Y: 1, Value: 18},
				{X: 2, Y: 2, Value: 0}, {X: 3, Y: 2, Value: 8}, {X: 4, Y: 2, Value: 16},
			},
		}},
		Options: controlledOptions("Deployment activity", "deployment-activity"),
	}
}

func sampleInteractiveBoxPlot() interactive.BoxPlotConfig {
	return interactive.BoxPlotConfig{
		Label: "Request latency distribution", Caption: "Five-number latency summaries by environment.",
		Categories: []string{"Development", "Staging", "Production"},
		Series: []interactive.BoxPlotSeries{{
			Name: "Latency (ms)",
			Data: []interactive.BoxPlotData{
				{Min: 18, Q1: 31, Median: 42, Q3: 58, Max: 94},
				{Min: 22, Q1: 38, Median: 51, Q3: 73, Max: 116},
				{Min: 25, Q1: 44, Median: 62, Q3: 86, Max: 138},
			},
		}},
		Options: controlledOptions("Request latency", "request-latency-distribution"),
	}
}

func sampleInteractiveGauge() interactive.GaugeConfig {
	return interactive.GaugeConfig{
		Label: "Deployment completion", Caption: "Current rollout completion percentage.",
		Variant: interactive.GaugeVariantProgress,
		Series: []interactive.GaugeSeries{{
			Name: "Rollout", Data: []interactive.GaugeData{{Name: "Complete", Value: 73}},
		}},
		Options: controlledOptions("Deployment completion", "deployment-completion"),
	}
}

func sampleInteractiveLiquidGauge(title string, shape interactive.GaugeLiquidShape, showLabel, showOutline bool, animate *bool) interactive.GaugeConfig {
	liquid := interactive.GaugeLiquidTreatment{Shape: shape, Animate: animate}
	if showLabel {
		liquid.Label = &interactive.GaugeLiquidLabel{Show: interactive.Bool(true)}
	}
	if showOutline {
		liquid.Outline = &interactive.GaugeLiquidOutline{Show: interactive.Bool(true)}
	}
	return interactive.GaugeConfig{
		Label: title, Caption: "Readings: 0.3, 0.4, and 0.5 in range 0 to 1.",
		Variant: interactive.GaugeVariantLiquid, Max: 1, Height: "320px", Liquid: liquid,
		Series:  []interactive.GaugeSeries{{Name: "liquid", Data: sampleLiquidGaugeData()}},
		Options: controlledOptions(title, title),
	}
}

func sampleLiquidGaugeData() []interactive.GaugeData {
	return []interactive.GaugeData{
		{Name: "Wave 1", Value: .3},
		{Name: "Wave 2", Value: .4},
		{Name: "Wave 3", Value: .5},
	}
}

func sampleInteractiveFunnel() interactive.FunnelConfig {
	return interactive.FunnelConfig{
		Label: "Release pipeline", Caption: "Builds progressing through release stages.",
		Series: []interactive.FunnelSeries{{
			Name: "Pipeline",
			Data: []interactive.FunnelData{
				{Name: "Built", Value: 120},
				{Name: "Tested", Value: 94},
				{Name: "Approved", Value: 61},
				{Name: "Deployed", Value: 48},
			},
		}},
		Options: controlledOptions("Release pipeline", "release-pipeline"),
	}
}

func sampleInteractiveGraph() interactive.Instance {
	return interactive.Graph(interactive.GraphConfig{
		Label: "Service dependency graph", Caption: "Dependencies between edge, API, data, and notification services.",
		Categories: []interactive.Category{{Name: "Entry"}, {Name: "Service"}, {Name: "Data"}},
		Nodes: []interactive.Node{
			{Name: "Edge", Category: "Entry", Value: 8, Size: 44},
			{Name: "API", Category: "Service", Value: 12, Size: 52},
			{Name: "Worker", Category: "Service", Value: 7, Size: 38},
			{Name: "Database", Category: "Data", Value: 10, Size: 46},
			{Name: "Notifications", Category: "Service", Value: 5, Size: 34},
		},
		Links: []interactive.Link{
			{Source: "Edge", Target: "API", Value: 12},
			{Source: "API", Target: "Database", Value: 10},
			{Source: "API", Target: "Worker", Value: 7},
			{Source: "Worker", Target: "Notifications", Value: 5},
		},
		Roam:    interactive.GraphRoamEnabled,
		Force:   &interactive.ForceOptions{Repulsion: 180, Gravity: 0.08, EdgeLength: 110},
		Options: controlledOptions("Service dependencies", "service-dependencies"),
	})
}

func sampleInteractiveSankey() interactive.Instance {
	return interactive.Sankey(interactive.SankeyConfig{
		Label: "Request flow", Caption: "Requests flowing from ingress through services to outcomes.",
		Series: []interactive.SankeySeries{{
			Name: "Requests",
			Nodes: []interactive.SankeyNode{
				{Name: "Ingress"}, {Name: "API"}, {Name: "Cache"}, {Name: "Database"}, {Name: "Success"}, {Name: "Error"},
			},
			Links: []interactive.SankeyLink{
				{Source: "Ingress", Target: "API", Value: 100},
				{Source: "API", Target: "Cache", Value: 48},
				{Source: "API", Target: "Database", Value: 45},
				{Source: "API", Target: "Error", Value: 7},
				{Source: "Cache", Target: "Success", Value: 48},
				{Source: "Database", Target: "Success", Value: 42},
				{Source: "Database", Target: "Error", Value: 3},
			},
		}},
		Layout:  interactive.SankeyLayout{NodeWidth: 18, NodeGap: 14},
		Options: controlledOptions("Request flow", "request-flow"),
	})
}

func sampleInteractiveTree() interactive.Instance {
	return interactive.Tree(interactive.TreeConfig{
		Label:   "Basic tree example",
		Caption: "One root with three branches. Node3 starts collapsed; expand it to reveal its children.",
		Roots: []*interactive.TreeNode{{
			Name: "Root",
			Children: []*interactive.TreeNode{
				{
					Name:     "Node1",
					Children: []*interactive.TreeNode{{Name: "Child1"}},
				},
				{
					Name: "Node2",
					Children: []*interactive.TreeNode{
						{Name: "Child1"},
						{Name: "Child2"},
						{Name: "Child3"},
					},
				},
				{
					Name: "Node3", Collapsed: interactive.Bool(true),
					Children: []*interactive.TreeNode{
						{Name: "Child1"},
						{Name: "Child2"},
						{Name: "Child3"},
					},
				},
			},
		}},
		Orientation:  interactive.TreeOrientationLeftToRight,
		InitialDepth: interactive.Int(-1),
		NodeLabel:    &interactive.LabelOptions{Show: interactive.Bool(true), Position: "top"},
		LeafLabel:    &interactive.LabelOptions{Show: interactive.Bool(true), Position: "right"},
		Insets:       interactive.TreeInsets{Left: "14%", Right: "14%", Top: "12%", Bottom: "12%"},
		Width:        "100%",
		Height:       "440px",
		Options:      controlledOptions("Basic tree example", "basic-tree-example"),
	})
}

func sampleInteractiveSunburst() interactive.Instance {
	return interactive.Sunburst(interactive.SunburstConfig{
		Label:   "Basic sunburst example",
		Caption: "Seven parent nodes, each paired with one child. Select a sector to focus the hierarchy; select the center to return.",
		Nodes: []*interactive.SunburstNode{
			{Name: "parent-0", Value: 0.81, Children: []*interactive.SunburstNode{{Name: "child-0", Value: 0.34}}},
			{Name: "parent-1", Value: 0.62, Children: []*interactive.SunburstNode{{Name: "child-1", Value: 0.57}}},
			{Name: "parent-2", Value: 0.45, Children: []*interactive.SunburstNode{{Name: "child-2", Value: 0.73}}},
			{Name: "parent-3", Value: 0.93, Children: []*interactive.SunburstNode{{Name: "child-3", Value: 0.28}}},
			{Name: "parent-4", Value: 0.38, Children: []*interactive.SunburstNode{{Name: "child-4", Value: 0.66}}},
			{Name: "parent-5", Value: 0.71, Children: []*interactive.SunburstNode{{Name: "child-5", Value: 0.49}}},
			{Name: "parent-6", Value: 0.54, Children: []*interactive.SunburstNode{{Name: "child-6", Value: 0.87}}},
		},
		LabelOptions: &interactive.LabelOptions{Show: interactive.Bool(true), Position: "inside", FontSize: 10},
		InnerRadius:  16,
		OuterRadius:  88,
		Width:        "100%",
		Height:       "32rem",
		Options: interactive.ChartOptions{
			Title:    &interactive.TitleOptions{Text: "Basic sunburst example"},
			Legend:   &interactive.LegendOptions{Show: interactive.Bool(false)},
			Tooltip:  &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "item"},
			Controls: chartcontrol.Options{Fullscreen: true},
			Export:   &chartcontrol.ExportOptions{Filename: "basic-sunburst-example"},
		},
	})
}

func sampleInteractiveTreemap() interactive.Instance {
	d3Children := make([]*interactive.TreemapNode, 40)
	for index := range d3Children {
		d3Children[index] = &interactive.TreemapNode{
			Name:  fmt.Sprintf("f%d", index),
			Value: float64(5 + (index*7+3)%10),
			Class: "file",
		}
	}
	return interactive.Treemap(interactive.TreemapConfig{
		Label:   "Basic treemap example",
		Caption: "File system usage in KB. Select a directory to focus it in the same chart; use the breadcrumb to return.",
		Nodes: []*interactive.TreemapNode{
			{
				Name: "d1", Class: "directory",
				Children: []*interactive.TreemapNode{{Name: "f1", Value: 1000, Class: "file"}},
			},
			{
				Name: "d2", Class: "directory",
				Children: []*interactive.TreemapNode{
					{Name: "f1", Value: 100, Class: "file"},
					{Name: "f2", Value: 300, Class: "file"},
					{Name: "f3", Value: 200, Class: "file"},
				},
			},
			{Name: "d3", Class: "directory", Children: d3Children},
			{Name: "f1", Value: 450, Class: "file"},
		},
		Navigation: interactive.TreemapNavigationDrillDown,
		Roam:       interactive.TreemapRoamEnabled,
		LabelOptions: &interactive.LabelOptions{
			Show: interactive.Bool(true), Position: "inside", FontSize: 11,
		},
		UpperLabel: &interactive.LabelOptions{Show: interactive.Bool(true), FontSize: 12},
		Breadcrumb: &interactive.TreemapBreadcrumb{Show: interactive.Bool(true), Height: 24, ItemGap: 8},
		NodeStyle:  interactive.TreemapNodeStyle{BorderWidth: 1, GapWidth: 1},
		LeafDepth:  interactive.Int(1),
		Levels: []interactive.TreemapLevel{
			{
				UpperLabel: &interactive.LabelOptions{Show: interactive.Bool(true)},
				NodeStyle:  interactive.TreemapNodeStyle{BorderWidth: 1, GapWidth: 1},
			},
			{NodeStyle: interactive.TreemapNodeStyle{BorderWidth: 2, GapWidth: 1}},
			{
				NodeStyle:       interactive.TreemapNodeStyle{GapWidth: 1},
				ColorSaturation: &interactive.TreemapColorRange{Min: 0.35, Max: 0.5},
			},
		},
		Width:  "100%",
		Height: "500px",
		Options: interactive.ChartOptions{
			Title:    &interactive.TitleOptions{Text: "Basic treemap example", Subtitle: "File system usage", Left: "center"},
			Legend:   &interactive.LegendOptions{Show: interactive.Bool(false)},
			Tooltip:  &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "item"},
			Controls: chartcontrol.Options{Fullscreen: true},
			Export:   &chartcontrol.ExportOptions{Filename: "basic-treemap-example"},
		},
	})
}

type parallelSampleRow struct {
	values [7]float64
	level  string
}

func sampleInteractiveParallel() interactive.Instance {
	return interactive.Parallel(interactive.ParallelConfig{
		Label:   "Multi Series parallel coordinates",
		Caption: "Daily air-quality measurements for Beijing, Guangzhou, and Shanghai across seven numeric dimensions and one categorical level.",
		Dimensions: []interactive.ParallelDimension{
			{Name: "Date", Range: &interactive.ParallelRange{Max: interactive.Float(31)}, Inverse: true, NameLocation: interactive.ParallelNameStart},
			{Name: "AQI"},
			{Name: "PM2.5"},
			{Name: "PM10"},
			{Name: "CO"},
			{Name: "NO2"},
			{Name: "SO2"},
			{Name: "Level", Categories: []string{"Good", "Moderate", "Lightly", "Moderately", "Heavily", "Severely"}},
		},
		Series: []interactive.ParallelSeries{
			{Name: "Beijing", Observations: parallelSampleObservations([]parallelSampleRow{
				{[7]float64{1, 55, 9, 56, 0.46, 18, 6}, "Moderate"},
				{[7]float64{2, 25, 11, 21, 0.65, 34, 9}, "Good"},
				{[7]float64{3, 56, 7, 63, 0.3, 14, 5}, "Moderate"},
				{[7]float64{4, 33, 7, 29, 0.33, 16, 6}, "Good"},
				{[7]float64{5, 42, 24, 44, 0.76, 40, 16}, "Good"},
				{[7]float64{6, 82, 58, 90, 1.77, 68, 33}, "Moderate"},
				{[7]float64{7, 74, 49, 77, 1.46, 48, 27}, "Moderate"},
				{[7]float64{8, 78, 55, 80, 1.29, 59, 29}, "Moderate"},
				{[7]float64{9, 267, 216, 280, 4.8, 108, 64}, "Heavily"},
				{[7]float64{10, 185, 127, 216, 2.52, 61, 27}, "Moderately"},
				{[7]float64{11, 39, 19, 38, 0.57, 31, 15}, "Good"},
				{[7]float64{12, 41, 11, 40, 0.43, 21, 7}, "Good"},
				{[7]float64{13, 64, 38, 74, 1.04, 46, 22}, "Moderate"},
				{[7]float64{14, 108, 79, 120, 1.7, 75, 41}, "Lightly"},
				{[7]float64{15, 108, 63, 116, 1.48, 44, 26}, "Lightly"},
				{[7]float64{16, 33, 6, 29, 0.34, 13, 5}, "Good"},
				{[7]float64{17, 94, 66, 110, 1.54, 62, 31}, "Moderate"},
				{[7]float64{18, 186, 142, 192, 3.88, 93, 79}, "Moderately"},
				{[7]float64{19, 57, 31, 54, 0.96, 32, 14}, "Moderate"},
				{[7]float64{20, 22, 8, 17, 0.48, 23, 10}, "Good"},
				{[7]float64{21, 39, 15, 36, 0.61, 29, 13}, "Good"},
			})},
			{Name: "Guangzhou", Observations: parallelSampleObservations([]parallelSampleRow{
				{[7]float64{1, 26, 37, 27, 1.163, 27, 13}, "Good"},
				{[7]float64{2, 85, 62, 71, 1.195, 60, 8}, "Moderate"},
				{[7]float64{3, 78, 38, 74, 1.363, 37, 7}, "Moderate"},
				{[7]float64{4, 21, 21, 36, 0.634, 40, 9}, "Good"},
				{[7]float64{5, 41, 42, 46, 0.915, 81, 13}, "Good"},
				{[7]float64{6, 56, 52, 69, 1.067, 92, 16}, "Moderate"},
				{[7]float64{7, 64, 30, 28, 0.924, 51, 2}, "Moderate"},
				{[7]float64{8, 55, 48, 74, 1.236, 75, 26}, "Moderate"},
				{[7]float64{9, 76, 85, 113, 1.237, 114, 27}, "Moderate"},
				{[7]float64{10, 91, 81, 104, 1.041, 56, 40}, "Moderate"},
				{[7]float64{11, 84, 39, 60, 0.964, 25, 11}, "Moderate"},
				{[7]float64{12, 64, 51, 101, 0.862, 58, 23}, "Moderate"},
				{[7]float64{13, 70, 69, 120, 1.198, 65, 36}, "Moderate"},
				{[7]float64{14, 77, 105, 178, 2.549, 64, 16}, "Moderate"},
				{[7]float64{15, 109, 68, 87, 0.996, 74, 29}, "Lightly"},
				{[7]float64{16, 73, 68, 97, 0.905, 51, 34}, "Moderate"},
				{[7]float64{17, 54, 27, 47, 0.592, 53, 12}, "Moderate"},
				{[7]float64{18, 51, 61, 97, 0.811, 65, 19}, "Moderate"},
				{[7]float64{19, 91, 71, 121, 1.374, 43, 18}, "Moderate"},
				{[7]float64{20, 73, 102, 182, 2.787, 44, 19}, "Moderate"},
				{[7]float64{21, 73, 50, 76, 0.717, 31, 20}, "Moderate"},
			})},
			{Name: "Shanghai", Observations: parallelSampleObservations([]parallelSampleRow{
				{[7]float64{1, 91, 45, 125, 0.82, 34, 23}, "Moderate"},
				{[7]float64{2, 65, 27, 78, 0.86, 45, 29}, "Moderate"},
				{[7]float64{3, 83, 60, 84, 1.09, 73, 27}, "Moderate"},
				{[7]float64{4, 109, 81, 121, 1.28, 68, 51}, "Lightly"},
				{[7]float64{5, 106, 77, 114, 1.07, 55, 51}, "Lightly"},
				{[7]float64{6, 109, 81, 121, 1.28, 68, 51}, "Lightly"},
				{[7]float64{7, 106, 77, 114, 1.07, 55, 51}, "Lightly"},
				{[7]float64{8, 89, 65, 78, 0.86, 51, 26}, "Moderate"},
				{[7]float64{9, 53, 33, 47, 0.64, 50, 17}, "Moderate"},
				{[7]float64{10, 80, 55, 80, 1.01, 75, 24}, "Moderate"},
				{[7]float64{11, 117, 81, 124, 1.03, 45, 24}, "Lightly"},
				{[7]float64{12, 99, 71, 142, 1.1, 62, 42}, "Moderate"},
				{[7]float64{13, 95, 69, 130, 1.28, 74, 50}, "Moderate"},
				{[7]float64{14, 116, 87, 131, 1.47, 84, 40}, "Lightly"},
				{[7]float64{15, 108, 80, 121, 1.3, 85, 37}, "Lightly"},
				{[7]float64{16, 134, 83, 167, 1.16, 57, 43}, "Lightly"},
				{[7]float64{17, 79, 43, 107, 1.05, 59, 37}, "Moderate"},
				{[7]float64{18, 71, 46, 89, 0.86, 64, 25}, "Moderate"},
				{[7]float64{19, 97, 71, 113, 1.17, 88, 31}, "Moderate"},
				{[7]float64{20, 84, 57, 91, 0.85, 55, 31}, "Moderate"},
				{[7]float64{21, 87, 63, 101, 0.9, 56, 41}, "Moderate"},
			})},
		},
		Width: "900px", Height: "500px",
		Options: interactive.ChartOptions{
			Title:    &interactive.TitleOptions{Text: "Multi Series"},
			Controls: chartcontrol.Options{Fullscreen: true},
			Export:   &chartcontrol.ExportOptions{Filename: "multi-series-parallel-coordinates"},
		},
		Style: charttheme.Style{Class: "max-w-full"},
	})
}

func parallelSampleObservations(rows []parallelSampleRow) []interactive.ParallelObservation {
	result := make([]interactive.ParallelObservation, len(rows))
	for index, row := range rows {
		values := make([]interactive.ParallelValue, 0, 8)
		for _, value := range row.values {
			values = append(values, interactive.ParallelNumber(value))
		}
		values = append(values, interactive.ParallelCategory(row.level))
		result[index] = interactive.ParallelObservation{Name: fmt.Sprintf("Day %d", index+1), Values: values}
	}
	return result
}

func sampleInteractiveThemeRiver() interactive.Instance {
	return interactive.ThemeRiver(interactive.ThemeRiverConfig{
		Label: "ThemeRiver-SingleAxis-Time", Caption: "Six named streams across aligned daily values from 8–28 November 2015.",
		Streams: sampleThemeRiverStreams(), Layout: interactive.ThemeRiverLayout{BottomPercent: interactive.Float(10)},
		Width: "100%", Height: "500px",
		Options: interactive.ChartOptions{Title: &interactive.TitleOptions{Text: "ThemeRiver-SingleAxis-Time"}, Tooltip: &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "axis"}, Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "theme-river-single-axis-time"}},
	})
}

func sampleThemeRiverStreams() []interactive.ThemeRiverStream {
	dates := make([]time.Time, 21)
	for index := range dates {
		dates[index] = time.Date(2015, time.November, 8+index, 0, 0, 0, 0, time.UTC)
	}
	values := [][]float64{
		{10, 15, 35, 38, 22, 16, 7, 2, 17, 33, 40, 32, 26, 35, 40, 32, 26, 22, 16, 22, 10},
		{35, 36, 37, 22, 24, 26, 34, 21, 18, 45, 32, 35, 30, 28, 27, 26, 15, 30, 35, 42, 42},
		{21, 25, 27, 23, 24, 21, 35, 39, 40, 36, 33, 43, 40, 34, 28, 26, 37, 41, 46, 47, 41},
		{10, 15, 35, 38, 22, 16, 7, 2, 17, 33, 40, 32, 26, 35, 40, 32, 26, 22, 16, 22, 10},
		{10, 15, 35, 38, 22, 16, 7, 2, 17, 33, 40, 32, 26, 35, 4, 32, 26, 22, 16, 22, 10},
		{10, 15, 35, 38, 22, 16, 7, 2, 17, 33, 4, 32, 26, 35, 40, 32, 26, 22, 16, 22, 10},
	}
	names := []string{"DQ", "TY", "SS", "QG", "SY", "DD"}
	streams := make([]interactive.ThemeRiverStream, len(names))
	for streamIndex, name := range names {
		points := make([]interactive.ThemeRiverPoint, len(dates))
		for pointIndex, date := range dates {
			points[pointIndex] = interactive.ThemeRiverPoint{Time: date, Value: values[streamIndex][pointIndex]}
		}
		streams[streamIndex] = interactive.ThemeRiverStream{Name: name, Class: "stream-" + strings.ToLower(name), Points: points}
	}
	return streams
}

func sampleInteractiveWordCloud(title string, shape interactive.WordCloudShape) interactive.Instance {
	return interactive.WordCloud(interactive.WordCloudConfig{
		Label: title, Caption: "Twenty weighted terms with exact values available below the canvas.",
		Series: interactive.WordCloudSeries{
			Name: "wordcloud", Words: sampleWordCloudWords(),
			Options: interactive.WordCloudSeriesOptions{
				Shape: shape, SizeRange: &interactive.WordCloudSizeRange{Min: 14, Max: 80},
			},
		},
		Options: interactive.ChartOptions{
			Title:    &interactive.TitleOptions{Text: title},
			Tooltip:  &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "item"},
			Controls: chartcontrol.Options{Fullscreen: true},
			Export:   &chartcontrol.ExportOptions{Filename: strings.ReplaceAll(strings.ToLower(title), " ", "-")},
		},
	})
}

func sampleWordCloudWords() []interactive.Word {
	return []interactive.Word{
		{Name: "Sam S Club", Value: 10000, Class: "retail"},
		{Name: "Macys", Value: 6181, Class: "retail", Color: "#ff8a3d"},
		{Name: "Amy Schumer", Value: 4386},
		{Name: "Jurassic World", Value: 4055},
		{Name: "Charter Communications", Value: 2467},
		{Name: "Chick Fil A", Value: 2244},
		{Name: "Planet Fitness", Value: 1898},
		{Name: "Pitch Perfect", Value: 1484},
		{Name: "Express", Value: 1689},
		{Name: "Home", Value: 1112},
		{Name: "Johnny Depp", Value: 985},
		{Name: "Lena Dunham", Value: 847},
		{Name: "Lewis Hamilton", Value: 582},
		{Name: "KXAN", Value: 555},
		{Name: "Mary Ellen Mark", Value: 550},
		{Name: "Farrah Abraham", Value: 462},
		{Name: "Rita Ora", Value: 366},
		{Name: "Serena Williams", Value: 282},
		{Name: "NCAA baseball tournament", Value: 273},
		{Name: "Point Break", Value: 265},
	}
}

func liveAvailabilityBar(label, caption string, states []int) interactive.BarConfig {
	categories := availabilityCategories(len(states))
	series := []interactive.BarSeries{
		{Name: "Healthy", Data: make([]interactive.BarData, len(states))},
		{Name: "Degraded", Data: make([]interactive.BarData, len(states))},
		{Name: "Down", Data: make([]interactive.BarData, len(states))},
	}
	for index, state := range states {
		if state >= 0 && state < len(series) {
			series[state].Data[index].Value = 1
		}
	}
	config := interactive.BarConfig{
		Label: label, Caption: caption, XAxis: categories, Series: series,
		Height: "240px",
		Style:  charttheme.Style{Palette: charttheme.PaletteStatus},
		Options: interactive.ChartOptions{
			Legend: &interactive.LegendOptions{Show: interactive.Bool(true)},
			XAxis: &interactive.AxisOptions{
				LabelInterval: interactive.Int(5), ShowFirstLabel: interactive.Bool(true), ShowLastLabel: interactive.Bool(true),
			},
			YAxis:     &interactive.AxisOptions{Min: interactive.Float(0), Max: interactive.Float(1), Show: interactive.Bool(false)},
			Animation: interactive.Bool(false),
			Controls:  chartcontrol.Options{Fullscreen: true},
			Export:    &chartcontrol.ExportOptions{Filename: "live-availability"},
		},
		SeriesOptions: interactive.SeriesOptions{Stack: "availability", BarWidth: "70%", BarGap: "0%"},
	}
	config.Live = &interactive.LiveData{URL: "/examples/live-availability/events", Event: "chart"}
	return config
}

func availabilityCategories(count int) []string {
	categories := make([]string, count)
	for index := range categories {
		categories[index] = fmt.Sprintf("-%ds", (count-1-index)*2)
	}
	return categories
}

func availabilityStates(step int) []int {
	states := make([]int, 36)
	for index := range states {
		switch phase := (step + index) % 24; {
		case phase >= 8 && phase <= 10:
			states[index] = 1
		case phase >= 17 && phase <= 19:
			states[index] = 2
		default:
			states[index] = 0
		}
	}
	return states
}

func sampleLiveAvailability() interactive.BarConfig {
	return liveAvailabilityBar(
		"Live availability from server-sent events",
		"Full snapshots stream through the renderer-neutral live-data contract.",
		availabilityStates(0),
	)
}

func interactiveChartBarCode() string {
	return `@interactive.Bar(interactive.BarConfig{
  Label: "Weekly deployments",
  XAxis: []string{"Mon", "Tue"},
  Series: []interactive.BarSeries{{Name: "Production", Data: []interactive.BarData{{Value: 3}, {Value: 5}}}},
})`
}

func interactiveChartLineCode() string {
	return `@interactive.Line(interactive.LineConfig{
  Label: "Weekly latency",
  XAxis: []string{"Mon", "Tue"},
  Series: []interactive.LineSeries{{Name: "Latency", Data: []interactive.LineData{{Value: 42}, {Value: 47}}}},
})`
}

func interactiveChartScatterCode() string {
	return `@interactive.Scatter(interactive.ScatterConfig{
  Label: "Release impact",
	  Variant: interactive.ScatterVariantEffect,
  XAxis: []string{"v1.8", "v1.9", "v2.0"},
	  Series: []interactive.ScatterSeries{{
    Name: "Impact",
	    Data: []interactive.ScatterData{{Value: 35}, {Value: 52}, {Value: 91}},
  }},
	  Ripple: &interactive.RippleOptions{Scale: 6, BrushType: "stroke"},
})`
}

func interactiveChartPieCode() string {
	return `@interactive.Pie(interactive.PieConfig{
  Label: "Rose area",
  Series: []interactive.PieSeries{{
    Name: "Seasons", InnerRadius: 40, OuterRadius: 75,
    RoseMode: interactive.PieRoseArea,
    LabelContent: interactive.PieLabelNameAndValue,
    Data: seasonalData,
  }},
})`
}

func interactiveChartRadarCode() string {
	return `@interactive.Radar(interactive.RadarConfig{
  Label: "Service health profile",
  Indicators: []interactive.RadarIndicator{
    {Name: "Availability", Max: 100},
    {Name: "Latency", Max: 500},
    {Name: "Capacity", Max: 200},
  },
  Series: []interactive.RadarSeries{{
    Name: "Profile",
    Data: []interactive.RadarData{{Name: "Current", Values: []float64{99.8, 180, 124}}},
  }},
})`
}

func interactiveChartHeatMapCode() string {
	return `@interactive.HeatMap(interactive.HeatMapConfig{
  Label: "Deployment activity",
  XAxis: []string{"08:00", "10:00", "12:00"},
  YAxis: []string{"Development", "Production"},
  ValueRange: interactive.HeatMapValueRange{Min: 0, Max: 20},
  Series: []interactive.HeatMapSeries{{
    Name: "Deployments",
    Data: []interactive.HeatMapData{
      {X: 0, Y: 0, Value: 12},
      {X: 2, Y: 1, Value: 4},
    },
  }},
})`
}

func interactiveChartBoxPlotCode() string {
	return `@interactive.BoxPlot(interactive.BoxPlotConfig{
  Label: "Request latency distribution",
  Categories: []string{"Development", "Production"},
  Series: []interactive.BoxPlotSeries{{
    Name: "Latency (ms)",
    Data: []interactive.BoxPlotData{
      {Min: 18, Q1: 31, Median: 42, Q3: 58, Max: 94},
      {Min: 25, Q1: 44, Median: 62, Q3: 86, Max: 138},
    },
  }},
})`
}

func interactiveChartGaugeCode() string {
	return `@interactive.Gauge(interactive.GaugeConfig{
	  Label: "Deployment completion",
	  Variant: interactive.GaugeVariantProgress,
	  Scale: interactive.GaugeScale{}, // Theme-aware cold-to-warm default.
	  Series: []interactive.GaugeSeries{{
    Name: "Rollout",
    Data: []interactive.GaugeData{{Name: "Complete", Value: 73}},
  }},
})`
}

func interactiveGaugeLiquidCode() string {
	return `@interactive.Gauge(interactive.GaugeConfig{
  Label: "Bounded fill",
  Variant: interactive.GaugeVariantLiquid,
  Min: 0,
  Max: 1,
  Series: []interactive.GaugeSeries{{
    Name: "liquid",
    Data: []interactive.GaugeData{
      {Name: "Wave 1", Value: .3},
      {Name: "Wave 2", Value: .4},
      {Name: "Wave 3", Value: .5},
    },
  }},
  Liquid: interactive.GaugeLiquidTreatment{
    Shape: interactive.GaugeLiquidShapeDiamond,
    Animate: interactive.Bool(true),
    Label: &interactive.GaugeLiquidLabel{Show: interactive.Bool(true)},
  },
})`
}

func interactiveChartFunnelCode() string {
	return `@interactive.Funnel(interactive.FunnelConfig{
  Label: "Release pipeline",
  Series: []interactive.FunnelSeries{{
    Name: "Pipeline",
    Data: []interactive.FunnelData{
      {Name: "Built", Value: 120},
      {Name: "Tested", Value: 94},
      {Name: "Deployed", Value: 48},
    },
  }},
})`
}

func liveAvailabilityCode() string {
	return `@interactive.Bar(interactive.BarConfig{
  Label: "Live availability",
  XAxis: rollingCategories, // 36 evenly spaced buckets.
  Series: oneHotStateSeries, // Healthy, Degraded, Down; equal lengths.
  SeriesOptions: interactive.SeriesOptions{
    Stack: "availability", BarWidth: "70%", BarGap: "0%",
  },
	Options: interactive.ChartOptions{Animation: interactive.Bool(false)},
  Live: &interactive.LiveData{URL: "/examples/live-availability/events", Event: "chart"},
})`
}

func interactiveGraphCode() string {
	return `@interactive.Graph(interactive.GraphConfig{
  Label: "Service dependencies",
  Nodes: []interactive.Node{{Name: "API"}, {Name: "Database"}},
  Links: []interactive.Link{{Source: "API", Target: "Database", Value: 10}},
  Layout: interactive.GraphLayoutForce,
  Roam: interactive.GraphRoamEnabled,
})`
}

func interactiveSankeyCode() string {
	return `@interactive.Sankey(interactive.SankeyConfig{
  Label: "Request flow",
  Series: []interactive.SankeySeries{{
    Name: "Requests",
    Nodes: []interactive.SankeyNode{{Name: "Ingress"}, {Name: "Success"}},
    Links: []interactive.SankeyLink{{Source: "Ingress", Target: "Success", Value: 90}},
  }},
})`
}

func interactiveTreeCode() string {
	return `@interactive.Tree(interactive.TreeConfig{
  Label: "Basic tree example",
  Caption: "One root with three branches; Node3 starts collapsed.",
  Roots: []*interactive.TreeNode{{
    Name: "Root",
    Children: []*interactive.TreeNode{
      {Name: "Node1", Children: []*interactive.TreeNode{{Name: "Child1"}}},
      {Name: "Node2", Children: []*interactive.TreeNode{
        {Name: "Child1"}, {Name: "Child2"}, {Name: "Child3"},
      }},
      {Name: "Node3", Collapsed: interactive.Bool(true), Children: []*interactive.TreeNode{
        {Name: "Child1"}, {Name: "Child2"}, {Name: "Child3"},
      }},
    },
  }},
  Orientation: interactive.TreeOrientationLeftToRight,
  InitialDepth: interactive.Int(-1),
  NodeLabel: &interactive.LabelOptions{Show: interactive.Bool(true), Position: "top"},
  LeafLabel: &interactive.LabelOptions{Show: interactive.Bool(true), Position: "right"},
	Insets: interactive.TreeInsets{Left: "14%", Right: "14%", Top: "12%", Bottom: "12%"},
  Width: "100%",
  Height: "440px",
})`
}

func interactiveSunburstCode() string {
	return `@interactive.Sunburst(interactive.SunburstConfig{
  Label: "Basic sunburst example",
  Caption: "Seven parent nodes, each paired with one child.",
  Nodes: []*interactive.SunburstNode{
    {Name: "parent-0", Value: 0.81, Children: []*interactive.SunburstNode{
      {Name: "child-0", Value: 0.34},
    }},
    // parent-1 through parent-5 retain the same one-child shape.
    {Name: "parent-6", Value: 0.54, Children: []*interactive.SunburstNode{
      {Name: "child-6", Value: 0.87},
    }},
  },
  Navigation: interactive.SunburstNavigationDrillDown,
  Sort: interactive.SunburstSortDescending,
  LabelOptions: &interactive.LabelOptions{Show: interactive.Bool(true), Position: "inside", FontSize: 10},
  InnerRadius: 16,
  OuterRadius: 88,
  Width: "100%",
  Height: "32rem",
  Style: charttheme.Style{Class: "max-w-full"},
  RootAttrs: templ.Attributes{"data-chart-purpose": "hierarchy"},
})`
}

func interactiveTreemapCode() string {
	return `@interactive.Treemap(interactive.TreemapConfig{
  Label: "Basic treemap example",
  Caption: "File system usage in KB. Select a directory; use the breadcrumb to return.",
  Nodes: []*interactive.TreemapNode{
    {Name: "d1", Class: "directory", Children: []*interactive.TreemapNode{
      {Name: "f1", Value: 1000, Class: "file"},
    }},
    {Name: "d2", Class: "directory", Children: []*interactive.TreemapNode{
      {Name: "f1", Value: 100}, {Name: "f2", Value: 300}, {Name: "f3", Value: 200},
    }},
    {Name: "d3", Class: "directory", Children: deterministicFiles},
    {Name: "f1", Value: 450, Class: "file"},
  },
  Navigation: interactive.TreemapNavigationDrillDown,
  Roam: interactive.TreemapRoamEnabled,
  LabelOptions: &interactive.LabelOptions{Show: interactive.Bool(true), Position: "inside"},
  UpperLabel: &interactive.LabelOptions{Show: interactive.Bool(true)},
  Breadcrumb: &interactive.TreemapBreadcrumb{Show: interactive.Bool(true)},
  NodeStyle: interactive.TreemapNodeStyle{BorderWidth: 1, GapWidth: 1},
  LeafDepth: interactive.Int(1),
  Width: "100%",
  Height: "500px",
})`
}

func interactiveParallelCode() string {
	return `@interactive.Parallel(interactive.ParallelConfig{
  Label: "Multi Series parallel coordinates",
  Dimensions: []interactive.ParallelDimension{
    {Name: "Date", Range: &interactive.ParallelRange{Max: interactive.Float(31)},
      Inverse: true, NameLocation: interactive.ParallelNameStart},
    {Name: "AQI"}, {Name: "PM2.5"}, {Name: "PM10"}, {Name: "CO"},
    {Name: "NO2"}, {Name: "SO2"},
    {Name: "Level", Categories: []string{
      "Good", "Moderate", "Lightly", "Moderately", "Heavily", "Severely",
    }},
  },
  Series: []interactive.ParallelSeries{
    {Name: "Beijing", Observations: beijingObservations},
    {Name: "Guangzhou", Observations: guangzhouObservations},
    {Name: "Shanghai", Observations: shanghaiObservations},
  },
  Width: "900px", Height: "500px",
  Style: charttheme.Style{Class: "max-w-full"},
})`
}

func interactiveThemeRiverCode() string {
	return `@interactive.ThemeRiver(interactive.ThemeRiverConfig{
  Label: "ThemeRiver-SingleAxis-Time",
  Caption: "Six named streams across aligned daily values.",
  Streams: []interactive.ThemeRiverStream{
    {Name: "DQ", Class: "stream-dq", Points: []interactive.ThemeRiverPoint{
      {Time: time.Date(2015, time.November, 8, 0, 0, 0, 0, time.UTC), Value: 10},
      {Time: time.Date(2015, time.November, 9, 0, 0, 0, 0, time.UTC), Value: 15},
      // Remaining aligned dates continue through 28 November.
    }},
    // TY, SS, QG, SY, and DD use the same aligned dates.
  },
  Layout: interactive.ThemeRiverLayout{BottomPercent: interactive.Float(10)},
  Width: "100%", Height: "500px",
  Options: interactive.ChartOptions{
    Title: &interactive.TitleOptions{Text: "ThemeRiver-SingleAxis-Time"},
    Tooltip: &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "axis"},
  },
})`
}

func interactiveWordCloudCode() string {
	return `@interactive.WordCloud(interactive.WordCloudConfig{
  Label: "star shape",
  Series: interactive.WordCloudSeries{
    Name: "wordcloud",
    Words: []interactive.Word{
      {Name: "Sam S Club", Value: 10000, Class: "retail"},
      {Name: "Macys", Value: 6181, Color: "#ff8a3d"},
    },
    Options: interactive.WordCloudSeriesOptions{
      Shape: interactive.WordCloudShapeStar,
      SizeRange: &interactive.WordCloudSizeRange{Min: 14, Max: 80},
    },
  },
})`
}
