package pages

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/bar"
	"github.com/araihu/goshtoso-charts/components/candlestick"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/araihu/goshtoso-charts/components/funnel"
	"github.com/araihu/goshtoso-charts/components/heatmap"
	"github.com/araihu/goshtoso-charts/components/line"
	"github.com/araihu/goshtoso-charts/components/pie"
	"github.com/araihu/goshtoso-charts/components/radar"
	"github.com/araihu/goshtoso-charts/components/scatter"
	charttable "github.com/araihu/goshtoso-charts/components/table"
	"github.com/araihu/goshtoso-charts/components/violin"
	"github.com/araihu/goshtoso/components/codeblock"
)

const (
	horizontalBarUpstreamPath     = "examples/1-Painter/horizontal_bar_chart-1-basic/main.go"
	horizontalBarUpstreamRevision = "1fe31b06b8a82e00df877ff4417a75858547c1c2"
	horizontalBarUpstreamSHA256   = "735240dd8433bd2494ae019f272840a8ff2fcf5572166b78269e23cbff7111a0"
	doughnutUpstreamPath          = "examples/1-Painter/doughnut_chart-1-basic/main.go"
	doughnutUpstreamRevision      = "1fe31b06b8a82e00df877ff4417a75858547c1c2"
	doughnutUpstreamSHA256        = "b97bca2322e90e2f03ab49aa77f683d0c58e027846b939e5a61100602dad1ebf"
	dualAxisLineUpstreamPath      = "examples/1-Painter/line_chart-8-dual_y_axis/main.go"
	dualAxisLineUpstreamRevision  = "1fe31b06b8a82e00df877ff4417a75858547c1c2"
	dualAxisLineUpstreamSHA256    = "78a3edd9aa356dc798c367b40cc5abecdb765b634795c38767f34bf266b805af"
	areaLineUpstreamPath          = "examples/1-Painter/line_chart-5-area/main.go"
	areaLineUpstreamRevision      = "1fe31b06b8a82e00df877ff4417a75858547c1c2"
	areaLineUpstreamSHA256        = "b2d7b87ff675f437dbc95f2d7a0447c2040e18c5b873256a5808987dfc6131d0"
)

type deterministicLCG struct{ state uint64 }

func (generator *deterministicLCG) next() float64 {
	generator.state = generator.state*6364136223846793005 + 1442695040888963407
	return float64(generator.state>>33) / float64(1<<31)
}

func (generator *deterministicLCG) normal(mean, standardDeviation float64) float64 {
	first, second := generator.next(), generator.next()
	if first < 1e-10 {
		first = 1e-10
	}
	return mean + standardDeviation*math.Sqrt(-2*math.Log(first))*math.Cos(2*math.Pi*second)
}

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
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "https-monitor-latency"},
	}
}

func sampleDualAxisLine() line.Config {
	return line.Config{
		Label:   "Dual Axis Line",
		Caption: "Two series retain independent left and right scales.",
		Title:   line.Title{Text: "Dual Axis Line"},
		Labels:  []string{"A", "B", "C", "D", "E", "F", "G"},
		Series: []line.Series{
			{Name: "Left Series", Values: []float64{120, 132, 101, 134, 90, 230, 210}},
			{Name: "Right Series", Values: []float64{820, 932, 901, 934, 1290, 1330, 1320}, YAxisIndex: 1},
		},
		YAxes:    []line.Axis{{}, {}},
		Width:    600,
		Height:   400,
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "dual-axis-line"},
	}
}

func sampleAreaLine() line.Config {
	minimum := 0.0
	noGap := false
	return line.Config{
		Label:    "Line",
		Caption:  "A filled area emphasizes magnitude across the seven ordered labels.",
		Title:    line.Title{Text: "Line"},
		Labels:   []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
		Series:   []line.Series{{Name: "Email", Values: []float64{120, 132, 101, 134, 90, 230, 210}}},
		Area:     line.AreaOptions{Enabled: true, Opacity: 150.0 / 255.0},
		XAxis:    line.CategoryAxisOptions{BoundaryGap: &noGap},
		Legend:   line.LegendOptions{Padding: line.Padding{Top: 5, Bottom: 10}},
		YAxes:    []line.Axis{{Min: &minimum}},
		Width:    600,
		Height:   400,
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "filled-area-line"},
	}
}

func sampleDualAxisLineOverrides() line.Config {
	cfg := sampleDualAxisLine()
	cfg.Label = "Dual Axis Line caller presentation overrides"
	cfg.Series[0].Color = "#14532d"
	cfg.Series[1].Class = "caller-right-series"
	cfg.YAxes[0].Class = "caller-left-axis"
	cfg.YAxes[1].Color = "#7e22ce"
	cfg.Export = &chartcontrol.ExportOptions{Filename: "dual-axis-line-overrides"}
	return cfg
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
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "deployments-by-environment"},
	}
}

func sampleHorizontalWorldPopulation() bar.Config {
	return bar.Config{
		Label:       "World population by reporting series",
		Caption:     "Population values reported for 2011 and 2012.",
		Title:       "World Population",
		Orientation: bar.OrientationHorizontal,
		Labels:      []string{"UN", "Brazil", "Indonesia", "USA", "India", "China", "World"},
		Series: []bar.Series{
			{Name: "2011", Values: []float64{10, 30, 50, 70, 90, 110, 130}},
			{Name: "2012", Values: []float64{20, 40, 60, 80, 100, 120, 140}},
		},
		Padding:  bar.Padding{Top: 20, Right: 40, Bottom: 20, Left: 20},
		Width:    600,
		Height:   400,
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "world-population"},
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
		Controls: chartcontrol.Options{Fullscreen: true},
		Export: &chartcontrol.ExportOptions{
			Filename: "observation-states", Background: chartcontrol.ExportBackgroundTransparent,
		},
	}
}

func sampleDoughnutChart() pie.Config {
	return pie.Config{
		Label:              "Doughnut Chart",
		Variant:            pie.VariantDoughnut,
		InnerRadiusPercent: 60,
		Title: pie.TitleOptions{
			Text: "Doughnut Chart", Subtitle: "(Fake Data)",
			Placement: pie.PlacementCenter, FontSize: 16, SubtitleFontSize: 10,
		},
		Legend: pie.LegendOptions{
			Orientation: pie.LegendVertical, LeftPercent: 80,
			VerticalPlacement: pie.VerticalPlacementBottom, FontSize: 10,
		},
		Padding: pie.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20},
		Slices: []pie.Slice{
			{Name: "Search Engine", Value: 1048},
			{Name: "Direct", Value: 735},
			{Name: "Email", Value: 580},
			{Name: "Union Ads", Value: 484},
			{Name: "Video Ads", Value: 300},
		},
		Width: 600, Height: 400,
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "doughnut-chart"},
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
		Controls: chartcontrol.Options{Fullscreen: true},
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
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "basic-candlestick-chart"},
	}
}

const (
	staticCandlestickBollingerUpstreamPath     = "examples/1-Painter/candlestick_chart-3-bollinger_bands/main.go"
	staticCandlestickBollingerUpstreamRevision = "1fe31b06b8a82e00df877ff4417a75858547c1c2"
	staticCandlestickBollingerUpstreamSHA256   = "cc3b347d5faea1a15ca22554dcc46a35beed74e49da56701659a1a7d1f000202"
)

func sampleCandlestickBollingerBands() candlestick.Config {
	return candlestick.Config{
		Label:      "Candlestick Chart with Bollinger Bands",
		Caption:    "Close-price Bollinger upper, simple moving average middle, and Bollinger lower lines use a centered five-period window.",
		Title:      "Candlestick Chart with Bollinger Bands",
		SeriesName: "Price",
		Data: []candlestick.Datum{
			{Label: "1", Open: 100, High: 110, Low: 95, Close: 105},
			{Label: "2", Open: 105, High: 115, Low: 100, Close: 112},
			{Label: "3", Open: 112, High: 118, Low: 108, Close: 115},
			{Label: "4", Open: 115, High: 125, Low: 110, Close: 120},
			{Label: "5", Open: 120, High: 130, Low: 115, Close: 125},
			{Label: "6", Open: 125, High: 135, Low: 120, Close: 130},
			{Label: "7", Open: 130, High: 140, Low: 125, Close: 135},
			{Label: "8", Open: 135, High: 145, Low: 130, Close: 140},
			{Label: "9", Open: 140, High: 150, Low: 135, Close: 145},
			{Label: "10", Open: 145, High: 155, Low: 140, Close: 150},
			{Label: "11", Open: 150, High: 160, Low: 145, Close: 148},
			{Label: "12", Open: 148, High: 153, Low: 143, Close: 146},
			{Label: "13", Open: 146, High: 151, Low: 141, Close: 144},
			{Label: "14", Open: 144, High: 149, Low: 139, Close: 142},
			{Label: "15", Open: 142, High: 147, Low: 137, Close: 140},
			{Label: "16", Open: 140, High: 145, Low: 135, Close: 138},
			{Label: "17", Open: 138, High: 143, Low: 133, Close: 136},
			{Label: "18", Open: 136, High: 141, Low: 131, Close: 134},
			{Label: "19", Open: 134, High: 139, Low: 129, Close: 132},
			{Label: "20", Open: 132, High: 137, Low: 127, Close: 130},
		},
		TrendLines: []candlestick.TrendLine{
			{Type: candlestick.TrendTypeBollingerUpper, Period: 5},
			{Type: candlestick.TrendTypeSimpleMovingAverage, Period: 5},
			{Type: candlestick.TrendTypeBollingerLower, Period: 5},
		},
		Options: candlestick.Options{
			TitleFontSize: 18,
			YUnit:         1,
			Padding:       candlestick.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20},
		},
		Width: 800, Height: 600,
		RootAttrs: templ.Attributes{"data-goshtoso-candidate": "candlestick-bollinger-fc218c7fedf84c7a"},
		Controls:  chartcontrol.Options{Fullscreen: true},
		Export:    &chartcontrol.ExportOptions{Filename: "candlestick-bollinger-bands"},
	}
}

func sampleCandlestickPresentationOverrides() candlestick.Config {
	cfg := sampleCandlestickBollingerBands()
	cfg.Label = "Candlestick caller presentation overrides"
	cfg.Options.Increasing.Color = "#14532d"
	cfg.Options.Decreasing.Class = "caller-decreasing-candles"
	cfg.TrendLines[0].Color = "#1d4ed8"
	cfg.TrendLines[1].Class = "caller-middle-band"
	cfg.Export = &chartcontrol.ExportOptions{Filename: "candlestick-presentation-overrides"}
	return cfg
}

func sampleBasicFunnel() funnel.Config {
	return funnel.Config{
		Label:   "Basic funnel",
		Caption: "Seven ordered stages, with every exact value available below the chart.",
		Title:   "Funnel",
		Stages: []funnel.Stage{
			{Label: "Show", Value: 100},
			{Label: "Click", Value: 80},
			{Label: "Visit", Value: 60},
			{Label: "Inquiry", Value: 40},
			{Label: "Order", Value: 20},
			{Label: "Pay", Value: 10},
			{Label: "Cancel", Value: 2},
		},
		Options:  funnel.Options{Legend: funnel.Legend{Padding: funnel.Padding{Left: 100}}},
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "basic-funnel"},
	}
}

func samplePeopleTable() charttable.Config {
	return charttable.Config{
		Label:   "People directory",
		Caption: "Three people, their age, address, tags, and available action.",
		Columns: []charttable.Column{
			{Header: "Name", Span: 2},
			{Header: "Age", Span: 1},
			{Header: "Address", Span: 3},
			{Header: "Tag", Span: 2},
			{Header: "Action", Span: 2},
		},
		Rows: [][]string{
			{"John Brown", "32", "New York No. 1 Lake Park", "nice, developer", "Send Mail"},
			{"Jim Green", "42", "London No. 1 Lake Park", "wow", "Send Mail"},
			{"Joe Black", "32", "Sidney No. 1 Lake Park", "cool, teacher", "Send Mail"},
		},
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "people-directory"},
	}
}

func sampleMarketTable() charttable.Config {
	return charttable.Config{
		Label: "Market changes",
		Columns: []charttable.Column{
			{Header: "Name"},
			{Header: "Price", Align: charttable.AlignEnd},
			{Header: "Change", Align: charttable.AlignEnd},
		},
		Rows: [][]string{
			{"Datadog Inc", "97.32", "-7.49%"},
			{"Hashicorp Inc", "28.66", "-9.25%"},
			{"Gitlab Inc", "51.63", "+4.32%"},
		},
		Colors: charttable.Colors{
			Surface: "#1c1c20", HeaderBackground: "#505050", HeaderText: "#ffffff",
			Text: "#ffffff", RowBackgrounds: []string{"#1c1c20"},
		},
		CellStyle: func(cell charttable.Cell) charttable.CellAppearance {
			if cell.Column != 2 {
				return charttable.CellAppearance{}
			}
			if strings.HasPrefix(cell.Value, "+") {
				return charttable.CellAppearance{BackgroundColor: "#b33514"}
			}
			return charttable.CellAppearance{BackgroundColor: "#217c32"}
		},
		Export: &chartcontrol.ExportOptions{Filename: "market-changes"},
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
		Controls:   chartcontrol.Options{Fullscreen: true},
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

func sampleDistributionShapes() violin.Config {
	const sampleCount = 200
	generator := &deterministicLCG{state: 42}
	series := []violin.Series{
		{Name: "Normal", Samples: make([]float64, sampleCount)},
		{Name: "Right Skewed", Samples: make([]float64, sampleCount)},
		{Name: "Bimodal", Samples: make([]float64, sampleCount)},
		{Name: "Tight", Samples: make([]float64, sampleCount)},
	}
	for index := 0; index < sampleCount; index++ {
		series[0].Samples[index] = generator.normal(50, 10)
		series[1].Samples[index] = 30 + math.Exp(generator.normal(0, 1))*10
		if generator.next() < .45 {
			series[2].Samples[index] = generator.normal(35, 6)
		} else {
			series[2].Samples[index] = generator.normal(65, 6)
		}
		series[3].Samples[index] = generator.normal(50, 3)
	}
	for index := range series {
		series[index].Marks = violin.MarkLines{Mean: true, Median: true}
		series[index].Statistics = violin.Statistics{Quantiles: []float64{.25, .75}}
		series[index].Class = "distribution-" + []string{"normal", "right-skewed", "bimodal", "tight"}[index]
	}
	return violin.Config{
		Label:   "Distribution shapes from deterministic samples",
		Caption: "Four 200-sample distributions; lines mark each mean and median, with quartiles in the adjacent summary.",
		Title:   "Distribution Shapes",
		Series:  series,
		Density: violin.Distribution{Points: 80},
		Padding: violin.Padding{Top: 5, Right: 5, Bottom: 50, Left: 5},
		Width:   1200, Height: 800,
	}
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
		Controls: chartcontrol.Options{Fullscreen: true},
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

func dualAxisLineCode() string {
	return `@line.Line(line.Config{
  Label: "Dual Axis Line",
  Title: line.Title{Text: "Dual Axis Line"},
  Labels: []string{"A", "B", "C", "D", "E", "F", "G"},
  Series: []line.Series{
    {Name: "Left Series", Values: []float64{120, 132, 101, 134, 90, 230, 210}},
    {Name: "Right Series", Values: []float64{820, 932, 901, 934, 1290, 1330, 1320}, YAxisIndex: 1},
  },
  YAxes: []line.Axis{{}, {}},
  Width: 600,
  Height: 400,
})`
}

func areaLineCode() string {
	return `minimum := 0.0
noGap := false

@line.Line(line.Config{
  Label: "Line",
  Title: line.Title{Text: "Line"},
  Labels: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
  Series: []line.Series{{Name: "Email", Values: []float64{120, 132, 101, 134, 90, 230, 210}}},
  Area: line.AreaOptions{Enabled: true, Opacity: 150.0 / 255.0},
  XAxis: line.CategoryAxisOptions{BoundaryGap: &noGap},
  Legend: line.LegendOptions{Padding: line.Padding{Top: 5, Bottom: 10}},
  YAxes: []line.Axis{{Min: &minimum}},
  Width: 600,
  Height: 400,
})`
}

func dualAxisLineOverridesCode() string {
	return `cfg := dualAxisLineConfig()
cfg.Series[0].Color = "#14532d"
cfg.Series[1].Class = "caller-right-series"
cfg.YAxes[0].Class = "caller-left-axis"
cfg.YAxes[1].Color = "#7e22ce"

@line.Line(cfg)`
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

func horizontalBarCode() string {
	return `@bar.Bar(bar.Config{
  Label: "World population by reporting series",
  Title: "World Population",
  Orientation: bar.OrientationHorizontal,
  Labels: []string{"UN", "Brazil", "Indonesia", "USA", "India", "China", "World"},
  Series: []bar.Series{
    {Name: "2011", Values: []float64{10, 30, 50, 70, 90, 110, 130}},
    {Name: "2012", Values: []float64{20, 40, 60, 80, 100, 120, 140}},
  },
  Padding: bar.Padding{Top: 20, Right: 40, Bottom: 20, Left: 20},
  Width: 600, Height: 400,
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

func doughnutCode() string {
	return `@pie.Pie(pie.Config{
  Label: "Doughnut Chart",
  Variant: pie.VariantDoughnut,
  InnerRadiusPercent: 60,
  Title: pie.TitleOptions{
    Text: "Doughnut Chart", Subtitle: "(Fake Data)",
    Placement: pie.PlacementCenter, FontSize: 16, SubtitleFontSize: 10,
  },
  Legend: pie.LegendOptions{
    Orientation: pie.LegendVertical, LeftPercent: 80,
    VerticalPlacement: pie.VerticalPlacementBottom, FontSize: 10,
  },
  Padding: pie.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20},
  Slices: []pie.Slice{
    {Name: "Search Engine", Value: 1048},
    {Name: "Direct", Value: 735},
    {Name: "Email", Value: 580},
    {Name: "Union Ads", Value: 484},
    {Name: "Video Ads", Value: 300},
  },
  Width: 600, Height: 400,
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

func candlestickBollingerCode() string {
	return `@candlestick.Candlestick(candlestick.Config{
  Label: "Candlestick Chart with Bollinger Bands",
  Title: "Candlestick Chart with Bollinger Bands",
  SeriesName: "Price",
  Data: []candlestick.Datum{
    {Label: "1", Open: 100, High: 110, Low: 95, Close: 105},
    {Label: "2", Open: 105, High: 115, Low: 100, Close: 112},
    {Label: "3", Open: 112, High: 118, Low: 108, Close: 115},
    {Label: "4", Open: 115, High: 125, Low: 110, Close: 120},
    {Label: "5", Open: 120, High: 130, Low: 115, Close: 125},
    {Label: "6", Open: 125, High: 135, Low: 120, Close: 130},
    {Label: "7", Open: 130, High: 140, Low: 125, Close: 135},
    {Label: "8", Open: 135, High: 145, Low: 130, Close: 140},
    {Label: "9", Open: 140, High: 150, Low: 135, Close: 145},
    {Label: "10", Open: 145, High: 155, Low: 140, Close: 150},
    {Label: "11", Open: 150, High: 160, Low: 145, Close: 148},
    {Label: "12", Open: 148, High: 153, Low: 143, Close: 146},
    {Label: "13", Open: 146, High: 151, Low: 141, Close: 144},
    {Label: "14", Open: 144, High: 149, Low: 139, Close: 142},
    {Label: "15", Open: 142, High: 147, Low: 137, Close: 140},
    {Label: "16", Open: 140, High: 145, Low: 135, Close: 138},
    {Label: "17", Open: 138, High: 143, Low: 133, Close: 136},
    {Label: "18", Open: 136, High: 141, Low: 131, Close: 134},
    {Label: "19", Open: 134, High: 139, Low: 129, Close: 132},
    {Label: "20", Open: 132, High: 137, Low: 127, Close: 130},
  },
  TrendLines: []candlestick.TrendLine{
    {Type: candlestick.TrendTypeBollingerUpper, Period: 5},
    {Type: candlestick.TrendTypeSimpleMovingAverage, Period: 5},
    {Type: candlestick.TrendTypeBollingerLower, Period: 5},
  },
  Options: candlestick.Options{
    TitleFontSize: 18,
    YUnit: 1,
    Padding: candlestick.Padding{Top: 20, Right: 20, Bottom: 20, Left: 20},
  },
  Width: 800,
  Height: 600,
})`
}

func candlestickPresentationCode() string {
	return `cfg := sampleCandlestickBollingerBands()
cfg.Options.Increasing.Color = "#14532d"
cfg.Options.Decreasing.Class = "caller-decreasing-candles"
cfg.TrendLines[0].Color = "#1d4ed8"
cfg.TrendLines[1].Class = "caller-middle-band"
@candlestick.Candlestick(cfg)`
}

func funnelCode() string {
	return `@funnel.Funnel(funnel.Config{
  Label: "Basic funnel",
  Title: "Funnel",
  Stages: []funnel.Stage{
    {Label: "Show", Value: 100},
    {Label: "Click", Value: 80},
    {Label: "Visit", Value: 60},
    {Label: "Inquiry", Value: 40},
    {Label: "Order", Value: 20},
    {Label: "Pay", Value: 10},
    {Label: "Cancel", Value: 2},
  },
  Options: funnel.Options{
    Legend: funnel.Legend{Padding: funnel.Padding{Left: 100}},
  },
})`
}

func tableCode() string {
	return `@table.Table(table.Config{
  Label: "People directory",
  Columns: []table.Column{
    {Header: "Name", Span: 2},
    {Header: "Age"},
    {Header: "Address", Span: 3},
    {Header: "Tag", Span: 2},
    {Header: "Action", Span: 2},
  },
  Rows: [][]string{
    {"John Brown", "32", "New York No. 1 Lake Park", "nice, developer", "Send Mail"},
    {"Jim Green", "42", "London No. 1 Lake Park", "wow", "Send Mail"},
    {"Joe Black", "32", "Sidney No. 1 Lake Park", "cool, teacher", "Send Mail"},
  },
})`
}

func marketTableCode() string {
	return `cfg := table.Config{
  Label: "Market changes",
  Columns: []table.Column{
    {Header: "Name"},
    {Header: "Price", Align: table.AlignEnd},
    {Header: "Change", Align: table.AlignEnd},
  },
  Rows: [][]string{
    {"Datadog Inc", "97.32", "-7.49%"},
    {"Hashicorp Inc", "28.66", "-9.25%"},
    {"Gitlab Inc", "51.63", "+4.32%"},
  },
  Colors: table.Colors{
    Surface: "#1c1c20", HeaderBackground: "#505050",
    HeaderText: "#ffffff", Text: "#ffffff",
  },
  CellStyle: marketChangeStyle,
}
@table.Table(cfg)`
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

func violinCode() string {
	return `@violin.Violin(violin.Config{
  Label: "Distribution shapes from deterministic samples",
  Caption: "Four 200-sample distributions with exact adjacent statistics.",
  Title: "Distribution Shapes",
  Series: []violin.Series{
    {Name: "Normal", Samples: normalSamples, Marks: violin.MarkLines{Mean: true, Median: true}, Statistics: violin.Statistics{Quantiles: []float64{.25, .75}}},
    {Name: "Right Skewed", Samples: rightSkewedSamples, Marks: violin.MarkLines{Mean: true, Median: true}, Statistics: violin.Statistics{Quantiles: []float64{.25, .75}}},
    {Name: "Bimodal", Samples: bimodalSamples, Marks: violin.MarkLines{Mean: true, Median: true}, Statistics: violin.Statistics{Quantiles: []float64{.25, .75}}},
    {Name: "Tight", Samples: tightSamples, Marks: violin.MarkLines{Mean: true, Median: true}, Statistics: violin.Statistics{Quantiles: []float64{.25, .75}}},
  },
  Density: violin.Distribution{Points: 80},
  Padding: violin.Padding{Top: 5, Right: 5, Bottom: 50, Left: 5},
  Width: 1200, Height: 800,
})`
}
