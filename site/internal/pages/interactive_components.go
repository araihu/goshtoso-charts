package pages

import (
	"fmt"

	"github.com/araihu/goshtoso-charts/components/charttheme"
	interactive "github.com/araihu/goshtoso-charts/components/interactive"
)

func sampleInteractiveBar() interactive.BarConfig {
	return interactive.BarConfig{
		Label: "Weekly deployments by environment", Caption: "Interactive bar component.",
		XAxis: []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
		Series: []interactive.BarSeries{
			{Name: "Staging", Data: []interactive.BarData{{Value: 8}, {Value: 12}, {Value: 9}, {Value: 14}, {Value: 11}}},
			{Name: "Production", Data: []interactive.BarData{{Value: 3}, {Value: 5}, {Value: 4}, {Value: 6}, {Value: 7}}},
		},
		Options: interactive.ChartOptions{Title: &interactive.TitleOptions{Text: "Deployments"}},
	}
}

func sampleInteractiveLine() interactive.LineConfig {
	return interactive.LineConfig{
		Label: "Weekly latency trend", Caption: "Interactive line component.",
		XAxis:   []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
		Series:  []interactive.LineSeries{{Name: "Latency (ms)", Data: []interactive.LineData{{Value: 42}, {Value: 47}, {Value: 45}, {Value: 51}, {Value: 44}}}},
		Options: interactive.ChartOptions{Title: &interactive.TitleOptions{Text: "Latency"}},
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
		Options: interactive.ChartOptions{Title: &interactive.TitleOptions{Text: "Release impact"}},
	}
}

func sampleInteractivePie() interactive.PieConfig {
	return interactive.PieConfig{
		Label: "Incident states", Caption: "Current incidents grouped by state.",
		Series: []interactive.PieSeries{{
			Name: "Incidents", InnerRadius: 32, OuterRadius: 72,
			Data: []interactive.PieData{{Name: "Open", Value: 12}, {Name: "Investigating", Value: 7}, {Name: "Resolved", Value: 28}},
		}},
		Options: interactive.ChartOptions{Title: &interactive.TitleOptions{Text: "Incident states"}},
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
		Options: interactive.ChartOptions{Title: &interactive.TitleOptions{Text: "Service health"}},
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
		Options: interactive.ChartOptions{Title: &interactive.TitleOptions{Text: "Deployment activity"}},
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
		Options: interactive.ChartOptions{Title: &interactive.TitleOptions{Text: "Request latency"}},
	}
}

func sampleInteractiveGauge() interactive.GaugeConfig {
	return interactive.GaugeConfig{
		Label: "Deployment completion", Caption: "Current rollout completion percentage.",
		Variant: interactive.GaugeVariantProgress,
		Series: []interactive.GaugeSeries{{
			Name: "Rollout", Data: []interactive.GaugeData{{Name: "Complete", Value: 73}},
		}},
		Options: interactive.ChartOptions{Title: &interactive.TitleOptions{Text: "Deployment completion"}},
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
		Options: interactive.ChartOptions{Title: &interactive.TitleOptions{Text: "Release pipeline"}},
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
		Options: interactive.ChartOptions{Title: &interactive.TitleOptions{Text: "Service dependencies"}},
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
		Options: interactive.ChartOptions{Title: &interactive.TitleOptions{Text: "Request flow"}},
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
		Insets:       interactive.TreeInsets{Left: "8%", Right: "40%", Top: "12%", Bottom: "12%"},
		Width:        "100%",
		Height:       "440px",
		Options: interactive.ChartOptions{
			Title:  &interactive.TitleOptions{Text: "Basic tree example"},
			Legend: &interactive.LegendOptions{Show: interactive.Bool(false)},
		},
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
			Title:   &interactive.TitleOptions{Text: "Basic sunburst example"},
			Legend:  &interactive.LegendOptions{Show: interactive.Bool(false)},
			Tooltip: &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "item"},
		},
	})
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
  Label: "Incident states",
  Series: []interactive.PieSeries{{
    Name: "Incidents",
    InnerRadius: 32,
    OuterRadius: 72,
    Data: []interactive.PieData{
      {Name: "Open", Value: 12},
      {Name: "Resolved", Value: 28},
    },
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
  Series: []interactive.GaugeSeries{{
    Name: "Rollout",
    Data: []interactive.GaugeData{{Name: "Complete", Value: 73}},
  }},
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
  Insets: interactive.TreeInsets{Left: "8%", Right: "40%", Top: "12%", Bottom: "12%"},
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
