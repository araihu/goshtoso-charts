package pages

import (
	"fmt"
	"math/rand"

	"github.com/araihu/goshtoso-charts/components/bar"
	"github.com/araihu/goshtoso-charts/components/candlestick"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/araihu/goshtoso-charts/components/heatmap"
	"github.com/araihu/goshtoso-charts/components/line"
	"github.com/araihu/goshtoso-charts/components/pie"
	"github.com/araihu/goshtoso-charts/components/radar"
	"github.com/araihu/goshtoso-charts/components/scatter"
	"github.com/araihu/goshtoso/components/codeblock"
)

func gettingStartedCodeBlock(language, label, code string) codeblock.Config {
	return codeblock.Config{Language: language, Label: label, Code: code}
}

const (
	gettingStartedInstallCode = `go get github.com/araihu/goshtoso-charts`
	gettingStartedAssetsCode  = `import chartassets "github.com/araihu/goshtoso-charts/assets"

mux.Handle("GET "+chartassets.Prefix, chartassets.Handler())`
	gettingStartedDependenciesCode = `import "github.com/araihu/goshtoso-charts/components/dependencies"

templ Layout() {
  <head>
    @dependencies.Dependencies()
  </head>
}`
	gettingStartedStaticCode = `@line.Line(line.Config{
  Label: "Request latency",
  Labels: []string{"Mon", "Tue", "Wed"},
  Series: []line.Series{{Name: "p95 (ms)", Values: []float64{42, 47, 44}}},
})`
	gettingStartedInteractiveCode = `@interactive.Bar(interactive.BarConfig{
  Label: "Deployments",
  XAxis: []string{"Mon", "Tue", "Wed"},
  Series: []interactive.BarSeries{{
    Name: "Production",
    Data: []interactive.BarData{{Value: 3}, {Value: 5}, {Value: 4}},
  }},
})`
)

func sampleLatency() line.Config {
	return line.Config{
		Label:    "HTTPS monitor latency in milliseconds",
		Caption:  "Median latency, last seven checks.",
		Labels:   []string{"08:00", "08:01", "08:02", "08:03", "08:04", "08:05", "08:06"},
		Series:   []line.Series{{Name: "Latency (ms)", Values: []float64{42, 47, 900, 51, 2_000, 44, 46}}},
		Controls: chartcontrol.Options{Fullscreen: true, Collapsible: true},
		Export:   &chartcontrol.ExportOptions{Filename: "https-monitor-latency"},
	}
}

func sampleDeployments() bar.Config {
	return bar.Config{
		Label:   "Deployments by environment",
		Caption: "Successful and failed deployments this week.",
		Labels:  []string{"Development", "Staging", "Production"},
		Series: []bar.Series{
			{Name: "Successful", Values: []float64{18, 12, 9}},
			{Name: "Failed", Values: []float64{1, 2, 1}},
		},
		Stacked:  true,
		Controls: chartcontrol.Options{Fullscreen: true, Collapsible: true},
		Export:   &chartcontrol.ExportOptions{Filename: "deployments-by-environment"},
	}
}

func sampleObservationStates() pie.Config {
	return pie.Config{
		Label:   "Observation states",
		Caption: "Most recent 100 retained monitor observations.",
		Slices: []pie.Slice{
			{Name: "Up", Value: 94},
			{Name: "Degraded", Value: 4},
			{Name: "Down", Value: 2},
		},
		Controls: chartcontrol.Options{Fullscreen: true, Collapsible: true},
		Export: &chartcontrol.ExportOptions{
			Filename: "observation-states", Background: chartcontrol.ExportBackgroundTransparent,
		},
	}
}

func sampleBasicRadar() radar.Config {
	return radar.Config{
		Label:   "Basic radar chart",
		Caption: "Allocated budget and actual spending across six shared dimensions.",
		Indicators: []radar.Indicator{
			{Name: "Sales", Max: 6500},
			{Name: "Administration", Max: 16000},
			{Name: "Information Technology", Max: 30000},
			{Name: "Customer Support", Max: 38000},
			{Name: "Development", Max: 52000},
			{Name: "Marketing", Max: 25000},
		},
		Series: []radar.Series{
			{Name: "Allocated Budget", Values: []float64{4200, 3000, 20000, 35000, 50000, 18000}},
			{Name: "Actual Spending", Values: []float64{5000, 14000, 28000, 26000, 42000, 21000}},
		},
		Controls: chartcontrol.Options{Fullscreen: true, Collapsible: true},
		Export:   &chartcontrol.ExportOptions{Filename: "basic-radar-chart"},
	}
}

func sampleBasicCandlestick() candlestick.Config {
	return candlestick.Config{
		Label:      "Seven-day stock price",
		Caption:    "Daily open, high, low, and close values; Day 4 decreases while the other days increase.",
		Title:      "Candlestick Chart",
		SeriesName: "Stock Price",
		Data: []candlestick.Datum{
			{Label: "Day 1", Open: 100, High: 110, Low: 95, Close: 105},
			{Label: "Day 2", Open: 105, High: 115, Low: 100, Close: 112},
			{Label: "Day 3", Open: 112, High: 118, Low: 108, Close: 115},
			{Label: "Day 4", Open: 115, High: 120, Low: 104, Close: 108},
			{Label: "Day 5", Open: 108, High: 113, Low: 105, Close: 109},
			{Label: "Day 6", Open: 109, High: 116, Low: 106, Close: 114},
			{Label: "Day 7", Open: 114, High: 121, Low: 111, Close: 119},
		},
		Controls: chartcontrol.Options{Fullscreen: true, Collapsible: true},
		Export:   &chartcontrol.ExportOptions{Filename: "basic-candlestick-chart"},
	}
}

func sampleBasicRadarOverride() radar.Config {
	cfg := sampleBasicRadar()
	cfg.Label = "Basic radar chart with caller overrides"
	cfg.Style = charttheme.Style{
		Palette: charttheme.PaletteAraiHu,
		Colors:  []string{"#365314", "#c2410c"},
		Class:   "radar-explicit-override",
	}
	cfg.Options = radar.Options{RadiusPercent: 44, ValueLabels: radar.ValueLabelsShown}
	return cfg
}

func sampleBasicHeatMap() heatmap.Config {
	return heatmap.Config{
		Label: "Basic heat map", Caption: "Values across five X and Y categories.", Title: "Heat Map Chart",
		XAxis: heatmap.Axis{Title: "X-Axis", Labels: []string{"0", "1", "2", "3", "4"}},
		YAxis: heatmap.Axis{Title: "Y-Axis", Labels: []string{"0", "1", "2", "3", "4"}},
		Rows: [][]float64{
			{4.4, 4.9, 7.0, 7.5, 4.3},
			{2.6, 5.9, 9.0, 6.4, 2.3},
			{3.3, 6.4, 7.0, 4.9, 3.2},
			{1.9, 6.0, 9.0, 5.9, 2.6},
			{4.4, 5.9, 7.0, 6.4, 4.6},
		},
		ValueRange: heatmap.ValueRange{Min: 1.9, Max: 9.0},
		Controls:   chartcontrol.Options{Fullscreen: true, Collapsible: true},
		Export:     &chartcontrol.ExportOptions{Filename: "basic-heat-map"},
	}
}

func sampleBasicHeatMapOverride() heatmap.Config {
	cfg := sampleBasicHeatMap()
	cfg.Label = "Basic heat map with reversed custom scale"
	cfg.Gradient = heatmap.Gradient{Reverse: true, Stops: []heatmap.GradientStop{
		{At: 0, Color: "#0e7490", Class: "scale-cold"},
		{At: 0.5, Color: "#fbbf24", Class: "scale-middle"},
		{At: 1, Color: "#e11d48", Class: "scale-warm"},
	}}
	return cfg
}

func sampleDenseScatter() scatter.Config {
	const dataPointCount = 1000
	categories := make([]string, dataPointCount)
	for index := range categories {
		categories[index] = fmt.Sprintf("foo %d", index)
	}
	values := denseScatterValues(rand.New(rand.NewSource(20260728)), 3, dataPointCount, 10)
	zero, maximum := 0.0, 280.0
	noGap := false
	return scatter.Config{
		Label:      "Dense scatter data",
		Caption:    "Three bounded random-walk series across 1,000 categories; sampled axis labels, trend lines, and maximum references summarize the shape. Applications own any adjacent exact-data view.",
		Categories: categories,
		Width:      600, Height: 400,
		Options: scatter.Options{Size: 0.5, Trend: scatter.TrendLine{Kind: scatter.TrendSimpleMovingAverage, Period: 100}, ValueFormat: scatter.ValueFormatHumanized},
		Series: []scatter.Series{
			{Name: "One", Values: values[0], Options: scatter.Options{ReferenceLine: scatter.ReferenceLineMaximum}},
			{Name: "Two", Values: values[1], Options: scatter.Options{ReferenceLine: scatter.ReferenceLineMaximum}},
			{Name: "Three", Values: values[2]},
		},
		Title:    scatter.TitleOptions{Text: "Dense Scatter Chart Demo", Placement: scatter.PlacementCenter},
		Legend:   scatter.LegendOptions{Orientation: scatter.LegendVertical, Placement: scatter.PlacementRight, Alignment: scatter.AlignmentRight, FontSize: 6},
		XAxis:    scatter.CategoryAxisOptions{BoundaryGap: &noGap, LabelCount: 10, LabelFontSize: 6, LabelRotation: 45},
		YAxis:    scatter.ValueAxisOptions{Min: &zero, Max: &maximum, Unit: 10, LabelSkip: 1, LabelFontSize: 6},
		Padding:  scatter.Padding{Top: 16, Right: 32, Bottom: 16, Left: 16},
		Controls: chartcontrol.Options{Fullscreen: true, Collapsible: true},
		Export:   &chartcontrol.ExportOptions{Filename: "dense-scatter-data"},
	}
}

func denseScatterValues(rng *rand.Rand, seriesCount, pointCount int, maxVariationPercentage float64) [][][]float64 {
	data := make([][][]float64, seriesCount)
	for seriesIndex := range data {
		data[seriesIndex] = make([][]float64, pointCount)
		for pointIndex := 0; pointIndex < pointCount; pointIndex++ {
			if pointIndex == 0 {
				data[seriesIndex][pointIndex] = []float64{rng.Float64() * 100}
				continue
			}
			previous := data[seriesIndex][pointIndex-1][0]
			variation := previous * maxVariationPercentage / 100
			minimum, maximum := previous-variation, previous+variation
			values := []float64{minimum + rng.Float64()*(maximum-minimum)}
			if pointIndex%2 == 0 {
				values = append(values, minimum+rng.Float64()*(maximum-minimum))
			}
			if pointIndex%10 == 0 {
				values = append(values, minimum+rng.Float64()*(maximum-minimum))
			}
			data[seriesIndex][pointIndex] = values
		}
	}
	return data
}

func lineCode() string {
	return `@line.Line(line.Config{
  Label: "HTTPS monitor latency in milliseconds",
  Labels: []string{"08:00", "08:01", "08:02"},
  Series: []line.Series{{Name: "Latency (ms)", Values: []float64{42, 47, 51}}},
})`
}

func barCode() string {
	return `@bar.Bar(bar.Config{
  Label: "Deployments by environment",
  Labels: []string{"Development", "Staging", "Production"},
  Series: []bar.Series{
    {Name: "Successful", Values: []float64{18, 12, 9}},
    {Name: "Failed", Values: []float64{1, 2, 1}},
  },
  Stacked: true,
})`
}

func pieCode() string {
	return `@pie.Pie(pie.Config{
  Label: "Observation states",
  Slices: []pie.Slice{
    {Name: "Up", Value: 94},
    {Name: "Degraded", Value: 4},
    {Name: "Down", Value: 2},
  },
})`
}

func scatterCode() string {
	return `@scatter.Scatter(scatter.Config{
	Label: "Dense scatter data",
	Categories: labels,
	Width: 600, Height: 400,
	Options: scatter.Options{Size: 0.5, Trend: scatter.TrendLine{
		Kind: scatter.TrendSimpleMovingAverage, Period: 100,
	}},
	Series: []scatter.Series{
		{Name: "One", Values: values[0], Options: scatter.Options{ReferenceLine: scatter.ReferenceLineMaximum}},
		{Name: "Two", Values: values[1], Options: scatter.Options{ReferenceLine: scatter.ReferenceLineMaximum}},
		{Name: "Three", Values: values[2]},
	},
})`
}

func radarCode() string {
	return `@radar.Radar(radar.Config{
  Label: "Basic radar chart",
  Indicators: []radar.Indicator{
    {Name: "Sales", Max: 6500},
    {Name: "Administration", Max: 16000},
    {Name: "Information Technology", Max: 30000},
    {Name: "Customer Support", Max: 38000},
    {Name: "Development", Max: 52000},
    {Name: "Marketing", Max: 25000},
  },
  Series: []radar.Series{
    {Name: "Allocated Budget", Values: []float64{4200, 3000, 20000, 35000, 50000, 18000}},
    {Name: "Actual Spending", Values: []float64{5000, 14000, 28000, 26000, 42000, 21000}},
  },
})`
}

func radarOverrideCode() string {
	return `cfg := sampleBasicRadar()
cfg.Options = radar.Options{
  RadiusPercent: 44,
  ValueLabels: radar.ValueLabelsShown,
}
cfg.Style = charttheme.Style{
  Palette: charttheme.PaletteAraiHu,
  Colors: []string{"#365314", "#c2410c"},
  Class: "my-radar",
}
@radar.Radar(cfg)`
}

func candlestickCode() string {
	return `@candlestick.Candlestick(candlestick.Config{
  Label: "Seven-day stock price",
  Title: "Candlestick Chart",
  SeriesName: "Stock Price",
  Data: []candlestick.Datum{
    {Label: "Day 1", Open: 100, High: 110, Low: 95, Close: 105},
    {Label: "Day 2", Open: 105, High: 115, Low: 100, Close: 112},
    {Label: "Day 3", Open: 112, High: 118, Low: 108, Close: 115},
    {Label: "Day 4", Open: 115, High: 120, Low: 104, Close: 108},
    {Label: "Day 5", Open: 108, High: 113, Low: 105, Close: 109},
    {Label: "Day 6", Open: 109, High: 116, Low: 106, Close: 114},
    {Label: "Day 7", Open: 114, High: 121, Low: 111, Close: 119},
  },
})`
}

func heatMapCode() string {
	return `@heatmap.HeatMap(heatmap.Config{
  Label: "Basic heat map",
  Title: "Heat Map Chart",
  XAxis: heatmap.Axis{Title: "X-Axis", Labels: []string{"0", "1", "2", "3", "4"}},
  YAxis: heatmap.Axis{Title: "Y-Axis", Labels: []string{"0", "1", "2", "3", "4"}},
  Rows: [][]float64{
    {4.4, 4.9, 7.0, 7.5, 4.3},
    {2.6, 5.9, 9.0, 6.4, 2.3},
    {3.3, 6.4, 7.0, 4.9, 3.2},
    {1.9, 6.0, 9.0, 5.9, 2.6},
    {4.4, 5.9, 7.0, 6.4, 4.6},
  },
  ValueRange: heatmap.ValueRange{Min: 1.9, Max: 9},
})`
}

func heatMapOverrideCode() string {
	return `cfg := sampleBasicHeatMap()
cfg.Gradient = heatmap.Gradient{
  Reverse: true,
  Stops: []heatmap.GradientStop{
    {At: 0, Color: "#0e7490", Class: "scale-cold"},
    {At: 0.5, Color: "#fbbf24", Class: "scale-middle"},
    {At: 1, Color: "#e11d48", Class: "scale-warm"},
  },
}
@heatmap.HeatMap(cfg)`
}
