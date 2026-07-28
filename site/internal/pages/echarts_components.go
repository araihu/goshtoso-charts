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
		Label: "Latency by request volume", Caption: "Each point represents one service node.",
		XAxisType: interactive.CartesianAxisValue,
		Series: []interactive.ScatterSeries{{
			Name: "Nodes",
			Data: []opts.ScatterData{
				{Name: "api-1", Value: [2]float64{120, 42}},
				{Name: "api-2", Value: [2]float64{180, 61}},
				{Name: "worker-1", Value: [2]float64{95, 35}},
				{Name: "worker-2", Value: [2]float64{240, 78}},
			},
		}},
		GlobalOptions: []charts.GlobalOpts{charts.WithTitleOpts(opts.Title{Title: "Load and latency"})},
	}
}

func sampleEChartsEffectScatter() interactive.EffectScatterConfig {
	return interactive.EffectScatterConfig{
		Label: "Release impact", Caption: "Ripple highlights the highest-impact releases.",
		XAxis: []string{"v1.8", "v1.9", "v2.0", "v2.1"},
		Series: []interactive.EffectScatterSeries{{
			Name:    "Impact",
			Data:    []opts.EffectScatterData{{Value: 35}, {Value: 52}, {Value: 91}, {Value: 64}},
			Options: []charts.SeriesOpts{charts.WithRippleEffectOpts(opts.RippleEffect{Scale: 6, BrushType: "stroke"})},
		}},
		GlobalOptions: []charts.GlobalOpts{charts.WithTitleOpts(opts.Title{Title: "Release impact"})},
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
  Label: "Latency by request volume",
  XAxisType: echarts.CartesianAxisValue,
  Series: []echarts.ScatterSeries{{Name: "Nodes", Data: []opts.ScatterData{
    {Name: "api-1", Value: [2]float64{120, 42}},
    {Name: "api-2", Value: [2]float64{180, 61}},
  }}},
})`
}

func eChartsEffectScatterCode() string {
	return `@echarts.EffectScatter(echarts.EffectScatterConfig{
  Label: "Release impact",
  XAxis: []string{"v1.8", "v1.9", "v2.0"},
  Series: []echarts.EffectScatterSeries{{
    Name: "Impact",
    Data: []opts.EffectScatterData{{Value: 35}, {Value: 52}, {Value: 91}},
    Options: []charts.SeriesOpts{charts.WithRippleEffectOpts(opts.RippleEffect{Scale: 6})},
  }},
})`
}
