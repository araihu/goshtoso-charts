package pages

import (
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
				{X: 0, Y: 0, Value: 12}, {X: 1, Y: 0, Value: 8}, {X: 2, Y: 0, Value: 15},
				{X: 1, Y: 1, Value: 6}, {X: 2, Y: 1, Value: 10}, {X: 3, Y: 1, Value: 7},
				{X: 2, Y: 2, Value: 3}, {X: 3, Y: 2, Value: 5}, {X: 4, Y: 2, Value: 4},
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
