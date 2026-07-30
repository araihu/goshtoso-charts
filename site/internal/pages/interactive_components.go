package pages

import (
	"fmt"
	"math/rand"
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

func sampleInteractiveLine() interactive.LineConfig {
	return interactive.LineConfig{
		Label: "Weekly latency trend", Caption: "Interactive line component.",
		XAxis:   []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
		Series:  []interactive.LineSeries{{Name: "Latency (ms)", Data: []interactive.LineData{{Value: 42}, {Value: 47}, {Value: 45}, {Value: 51}, {Value: 44}}}},
		Options: controlledOptions("Latency", "interactive-latency"),
	}
}

var interactiveHeatMapSourceData = [][3]int{
	{0, 0, 5}, {0, 1, 1}, {0, 2, 0}, {0, 3, 0}, {0, 4, 0}, {0, 5, 0},
	{0, 6, 0}, {0, 7, 0}, {0, 8, 0}, {0, 9, 0}, {0, 10, 0}, {0, 11, 2},
	{0, 12, 4}, {0, 13, 1}, {0, 14, 1}, {0, 15, 3}, {0, 16, 4}, {0, 17, 6},
	{0, 18, 4}, {0, 19, 4}, {0, 20, 3}, {0, 21, 3}, {0, 22, 2}, {0, 23, 5},
	{1, 0, 7}, {1, 1, 0}, {1, 2, 0}, {1, 3, 0}, {1, 4, 0}, {1, 5, 0},
	{1, 6, 0}, {1, 7, 0}, {1, 8, 0}, {1, 9, 0}, {1, 10, 5}, {1, 11, 2},
	{1, 12, 2}, {1, 13, 6}, {1, 14, 9}, {1, 15, 11}, {1, 16, 6}, {1, 17, 7},
	{1, 18, 8}, {1, 19, 12}, {1, 20, 5}, {1, 21, 5}, {1, 22, 7}, {1, 23, 2},
	{2, 0, 1}, {2, 1, 1}, {2, 2, 0}, {2, 3, 0}, {2, 4, 0}, {2, 5, 0}, {2, 6, 0},
	{2, 7, 0}, {2, 8, 0}, {2, 9, 0}, {2, 10, 3}, {2, 11, 2}, {2, 12, 1}, {2, 13, 9},
	{2, 14, 8}, {2, 15, 10}, {2, 16, 6}, {2, 17, 5}, {2, 18, 5}, {2, 19, 5},
	{2, 20, 7}, {2, 21, 4}, {2, 22, 2}, {2, 23, 4}, {3, 0, 7}, {3, 1, 3},
	{3, 2, 0}, {3, 3, 0}, {3, 4, 0}, {3, 5, 0}, {3, 6, 0}, {3, 7, 0},
	{3, 8, 1}, {3, 9, 0}, {3, 10, 5}, {3, 11, 4}, {3, 12, 7}, {3, 13, 14},
	{3, 14, 13}, {3, 15, 12}, {3, 16, 9}, {3, 17, 5}, {3, 18, 5}, {3, 19, 10},
	{3, 20, 6}, {3, 21, 4}, {3, 22, 4}, {3, 23, 1}, {4, 0, 1}, {4, 1, 3},
	{4, 2, 0}, {4, 3, 0}, {4, 4, 0}, {4, 5, 1}, {4, 6, 0}, {4, 7, 0},
	{4, 8, 0}, {4, 9, 2}, {4, 10, 4}, {4, 11, 4}, {4, 12, 2}, {4, 13, 4},
	{4, 14, 4}, {4, 15, 14}, {4, 16, 12}, {4, 17, 1}, {4, 18, 8}, {4, 19, 5},
	{4, 20, 3}, {4, 21, 7}, {4, 22, 3}, {4, 23, 0}, {5, 0, 2}, {5, 1, 1},
	{5, 2, 0}, {5, 3, 3}, {5, 4, 0}, {5, 5, 0}, {5, 6, 0}, {5, 7, 0}, {5, 8, 2},
	{5, 9, 0}, {5, 10, 4}, {5, 11, 1}, {5, 12, 5}, {5, 13, 10}, {5, 14, 5},
	{5, 15, 7}, {5, 16, 11}, {5, 17, 6}, {5, 18, 0}, {5, 19, 5}, {5, 20, 3},
	{5, 21, 4}, {5, 22, 2}, {5, 23, 0}, {6, 0, 1}, {6, 1, 0}, {6, 2, 0},
	{6, 3, 0}, {6, 4, 0}, {6, 5, 0}, {6, 6, 0}, {6, 7, 0}, {6, 8, 0},
	{6, 9, 0}, {6, 10, 1}, {6, 11, 0}, {6, 12, 2}, {6, 13, 1}, {6, 14, 3},
	{6, 15, 4}, {6, 16, 0}, {6, 17, 0}, {6, 18, 0}, {6, 19, 0}, {6, 20, 1},
	{6, 21, 2}, {6, 22, 2}, {6, 23, 6},
}

var interactiveHeatMapWeekDays = []string{"Saturday", "Friday", "Thursday", "Wednesday", "Tuesday", "Monday", "Sunday"}

var interactiveHeatMapDayHours = []string{
	"12a", "1a", "2a", "3a", "4a", "5a", "6a", "7a", "8a", "9a", "10a", "11a",
	"12p", "1p", "2p", "3p", "4p", "5p", "6p", "7p", "8p", "9p", "10p", "11p",
}

func sampleInteractiveHeatMap() interactive.HeatMapConfig {
	data := make([]interactive.HeatMapData, len(interactiveHeatMapSourceData))
	for index, source := range interactiveHeatMapSourceData {
		data[index] = interactive.HeatMapData{X: source[1], Y: source[0], Value: float64(source[2]), Missing: source[2] == 0}
	}
	options := controlledOptions("Basic heatmap example", "weekly-activity-heatmap")
	options.Legend = &interactive.LegendOptions{Top: "32", Left: "center"}
	return interactive.HeatMapConfig{
		Label: "Weekly activity by hour", Caption: "Seven days across twenty-four hourly buckets; empty source cells remain no-data cells.",
		XAxis: interactiveHeatMapDayHours, YAxis: interactiveHeatMapWeekDays,
		ValueRange: interactive.HeatMapValueRange{Min: 0, Max: 10, Calculable: interactive.Bool(true)},
		SplitArea:  interactive.Bool(true),
		Series:     []interactive.HeatMapSeries{{Name: "Activity", Data: data}},
		Height:     "34rem",
		Options:    options,
	}
}

func sampleInteractiveCalendarHeatMap() interactive.HeatMapConfig {
	start := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC)
	random := rand.New(rand.NewSource(1))
	data := make([]interactive.HeatMapData, 0, 366)
	for date := start; !date.After(end); date = date.AddDate(0, 0, 1) {
		value := random.Intn(21)
		data = append(data, interactive.HeatMapData{Date: date, Value: float64(value), Missing: value == 0})
	}
	options := controlledOptions("Calendar heatmap example", "calendar-activity-heatmap")
	options.Legend = &interactive.LegendOptions{Top: "32", Left: "center"}
	return interactive.HeatMapConfig{
		Label: "Calendar activity", Caption: "A deterministic 366-day sequence preserves the source range and explicit no-data days.",
		Coordinate: interactive.HeatMapCoordinateCalendar,
		Calendar: &interactive.HeatMapCalendar{
			Start: start, End: end,
			Options: interactive.CalendarOptions{
				Top: "100", Left: "30", Right: "30", CellSize: "auto", Orient: "horizontal",
				CellStyle: &interactive.ItemStyle{BorderWidth: 0.5}, MonthLabel: &interactive.CalendarLabelOptions{FontSize: 7},
			},
		},
		ValueRange: interactive.HeatMapValueRange{Min: 0, Max: 20},
		Series:     []interactive.HeatMapSeries{{Name: "Calendar activity", Data: data}},
		Height:     "34rem",
		Options:    options,
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

func interactiveChartLineCode() string {
	return `@interactive.Line(interactive.LineConfig{
  Label: "Basic line example",
  XAxis: []string{"Apple", "Banana", "Peach", "Lemon", "Pear", "Cherry"},
  Series: []interactive.LineSeries{{
    Name: "Category A",
    Data: []interactive.LineData{{Value: 120}, {Value: 132}, {Value: 101}, {Value: 134}, {Value: 90}, {Value: 230}},
  }},
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

func interactiveChartHeatMapCode() string {
	return `@interactive.HeatMap(interactive.HeatMapConfig{
  Label: "Weekly activity by hour",
  XAxis: []string{
    "12a", "1a", "2a", "3a", "4a", "5a", "6a", "7a", "8a", "9a", "10a", "11a",
    "12p", "1p", "2p", "3p", "4p", "5p", "6p", "7p", "8p", "9p", "10p", "11p",
  },
  YAxis: []string{"Saturday", "Friday", "Thursday", "Wednesday", "Tuesday", "Monday", "Sunday"},
  ValueRange: interactive.HeatMapValueRange{
    Min: 0, Max: 10, Calculable: interactive.Bool(true),
  },
  SplitArea: interactive.Bool(true),
  Series: []interactive.HeatMapSeries{{
    Name: "Activity",
    Data: []interactive.HeatMapData{
      {X: 0, Y: 0, Value: 5},
      {X: 1, Y: 0, Value: 1},
      {X: 2, Y: 0, Missing: true},
    },
  }},
})`
}

func interactiveCalendarHeatMapCode() string {
	return `// Add "time" to the templ file's Go import block.
@interactive.HeatMap(interactive.HeatMapConfig{
  Label: "Calendar activity",
  Coordinate: interactive.HeatMapCoordinateCalendar,
  Calendar: &interactive.HeatMapCalendar{
    Start: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
    End: time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC),
    Options: interactive.CalendarOptions{
      Top: "100", Left: "30", Right: "30",
      CellSize: "auto", Orient: "horizontal",
      CellStyle: &interactive.ItemStyle{BorderWidth: 0.5},
      MonthLabel: &interactive.CalendarLabelOptions{FontSize: 7},
    },
  },
  ValueRange: interactive.HeatMapValueRange{Min: 0, Max: 20},
  Series: []interactive.HeatMapSeries{{
    Name: "Calendar activity",
    Data: []interactive.HeatMapData{
      {Date: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC), Value: 12},
      {Date: time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC), Missing: true},
      {Date: time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC), Value: 7},
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
