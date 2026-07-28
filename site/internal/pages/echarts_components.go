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
