package pages

import (
	interactive "github.com/araihu/goshtoso-charts/components/echarts"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func sampleEChartsBar() interactive.BarConfig {
	return interactive.BarConfig{
		Label: "Weekly deployments by environment", Caption: "Interactive bar component.",
		XAxis: []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
		Series: []interactive.BarSeries{
			{Name: "Staging", Data: []opts.BarData{{Value: 8}, {Value: 12}, {Value: 9}, {Value: 14}, {Value: 11}}},
			{Name: "Production", Data: []opts.BarData{{Value: 3}, {Value: 5}, {Value: 4}, {Value: 6}, {Value: 7}}},
		},
		GlobalOptions: []charts.GlobalOpts{charts.WithTitleOpts(opts.Title{Title: "Deployments"})},
	}
}

func sampleEChartsLine() interactive.LineConfig {
	return interactive.LineConfig{
		Label: "Weekly latency trend", Caption: "Interactive line component.",
		XAxis:         []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
		Series:        []interactive.LineSeries{{Name: "Latency (ms)", Data: []opts.LineData{{Value: 42}, {Value: 47}, {Value: 45}, {Value: 51}, {Value: 44}}}},
		GlobalOptions: []charts.GlobalOpts{charts.WithTitleOpts(opts.Title{Title: "Latency"})},
	}
}

func sampleEChartsScatter() interactive.ScatterConfig {
	return interactive.ScatterConfig{
		Label: "Release impact", Caption: "Ripple highlights the highest-impact releases.",
		Variant: interactive.ScatterVariantEffect,
		XAxis:   []string{"v1.8", "v1.9", "v2.0", "v2.1"},
		Series: []interactive.ScatterSeries{{
			Name: "Impact",
			Data: []opts.ScatterData{{Value: 35}, {Value: 52}, {Value: 91}, {Value: 64}},
		}},
		Ripple:        &opts.RippleEffect{Scale: 6, BrushType: "stroke"},
		GlobalOptions: []charts.GlobalOpts{charts.WithTitleOpts(opts.Title{Title: "Release impact"})},
	}
}

func sampleEChartsPie() interactive.PieConfig {
	return interactive.PieConfig{
		Label: "Incident states", Caption: "Current incidents grouped by state.",
		Series: []interactive.PieSeries{{
			Name: "Incidents", InnerRadius: 32, OuterRadius: 72,
			Data: []interactive.PieData{{Name: "Open", Value: 12}, {Name: "Investigating", Value: 7}, {Name: "Resolved", Value: 28}},
		}},
		GlobalOptions: []charts.GlobalOpts{charts.WithTitleOpts(opts.Title{Title: "Incident states"})},
	}
}

func sampleEChartsRadar() interactive.RadarConfig {
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
		GlobalOptions: []charts.GlobalOpts{charts.WithTitleOpts(opts.Title{Title: "Service health"})},
	}
}

func sampleEChartsHeatMap() interactive.HeatMapConfig {
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
		GlobalOptions: []charts.GlobalOpts{charts.WithTitleOpts(opts.Title{Title: "Deployment activity"})},
	}
}

func sampleEChartsBoxPlot() interactive.BoxPlotConfig {
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
		GlobalOptions: []charts.GlobalOpts{charts.WithTitleOpts(opts.Title{Title: "Request latency"})},
	}
}

func eChartsBarCode() string {
	return `@echarts.Bar(echarts.BarConfig{
  Label: "Weekly deployments",
  XAxis: []string{"Mon", "Tue"},
  Series: []echarts.BarSeries{{Name: "Production", Data: []opts.BarData{{Value: 3}, {Value: 5}}}},
})`
}

func eChartsLineCode() string {
	return `@echarts.Line(echarts.LineConfig{
  Label: "Weekly latency",
  XAxis: []string{"Mon", "Tue"},
  Series: []echarts.LineSeries{{Name: "Latency", Data: []opts.LineData{{Value: 42}, {Value: 47}}}},
})`
}

func eChartsScatterCode() string {
	return `@echarts.Scatter(echarts.ScatterConfig{
  Label: "Release impact",
	  Variant: echarts.ScatterVariantEffect,
  XAxis: []string{"v1.8", "v1.9", "v2.0"},
	  Series: []echarts.ScatterSeries{{
    Name: "Impact",
	    Data: []opts.ScatterData{{Value: 35}, {Value: 52}, {Value: 91}},
  }},
	  Ripple: &opts.RippleEffect{Scale: 6, BrushType: "stroke"},
})`
}

func eChartsPieCode() string {
	return `@echarts.Pie(echarts.PieConfig{
  Label: "Incident states",
  Series: []echarts.PieSeries{{
    Name: "Incidents",
    InnerRadius: 32,
    OuterRadius: 72,
    Data: []echarts.PieData{
      {Name: "Open", Value: 12},
      {Name: "Resolved", Value: 28},
    },
  }},
})`
}

func eChartsRadarCode() string {
	return `@echarts.Radar(echarts.RadarConfig{
  Label: "Service health profile",
  Indicators: []echarts.RadarIndicator{
    {Name: "Availability", Max: 100},
    {Name: "Latency", Max: 500},
    {Name: "Capacity", Max: 200},
  },
  Series: []echarts.RadarSeries{{
    Name: "Profile",
    Data: []echarts.RadarData{{Name: "Current", Values: []float64{99.8, 180, 124}}},
  }},
})`
}

func eChartsHeatMapCode() string {
	return `@echarts.HeatMap(echarts.HeatMapConfig{
  Label: "Deployment activity",
  XAxis: []string{"08:00", "10:00", "12:00"},
  YAxis: []string{"Development", "Production"},
  ValueRange: echarts.HeatMapValueRange{Min: 0, Max: 20},
  Series: []echarts.HeatMapSeries{{
    Name: "Deployments",
    Data: []echarts.HeatMapData{
      {X: 0, Y: 0, Value: 12},
      {X: 2, Y: 1, Value: 4},
    },
  }},
})`
}

func eChartsBoxPlotCode() string {
	return `@echarts.BoxPlot(echarts.BoxPlotConfig{
  Label: "Request latency distribution",
  Categories: []string{"Development", "Production"},
  Series: []echarts.BoxPlotSeries{{
    Name: "Latency (ms)",
    Data: []echarts.BoxPlotData{
      {Min: 18, Q1: 31, Median: 42, Q3: 58, Max: 94},
      {Min: 25, Q1: 44, Median: 62, Q3: 86, Max: 138},
    },
  }},
})`
}
