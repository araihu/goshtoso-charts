package echartsexamples

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/render"
)

const statisticalExamplesSource = "https://github.com/go-echarts/examples/blob/master/examples/"

// StatisticalExamples is the deterministic port of the upstream boxplot,
// kline, funnel, gauge, pie, and radar examples. Dynamic upstream examples
// render a representative static state so server rendering stays repeatable.
var StatisticalExamples = []Example{
	{Slug: "boxplot-basic", Title: "Basic boxplot", Group: "Statistical", Source: statisticalExamplesSource + "boxplot.go", Build: func() render.Renderer { return boxPlotBase() }},
	{Slug: "boxplot-multi", Title: "Multi-series boxplot", Group: "Statistical", Source: statisticalExamplesSource + "boxplot.go", Build: func() render.Renderer { return boxPlotMulti() }},
	{Slug: "kline-basic", Title: "Basic Kline", Group: "Statistical", Source: statisticalExamplesSource + "kline.go", Build: func() render.Renderer { return klineBase() }},
	{Slug: "kline-datazoom-inside", Title: "Kline inside data zoom", Group: "Statistical", Source: statisticalExamplesSource + "kline.go", Build: func() render.Renderer { return klineDataZoomInside() }},
	{Slug: "kline-datazoom-both", Title: "Kline inside and slider zoom", Group: "Statistical", Source: statisticalExamplesSource + "kline.go", Build: func() render.Renderer { return klineDataZoomBoth() }},
	{Slug: "kline-datazoom-yaxis", Title: "Kline Y-axis data zoom", Group: "Statistical", Source: statisticalExamplesSource + "kline.go", Build: func() render.Renderer { return klineDataZoomYAxis() }},
	{Slug: "kline-style", Title: "Styled Kline", Group: "Statistical", Source: statisticalExamplesSource + "kline.go", Build: func() render.Renderer { return klineStyle() }},
	{Slug: "funnel-basic", Title: "Basic funnel", Group: "Statistical", Source: statisticalExamplesSource + "funnel.go", Build: func() render.Renderer { return funnelBase() }},
	{Slug: "funnel-label", Title: "Funnel labels", Group: "Statistical", Source: statisticalExamplesSource + "funnel.go", Build: func() render.Renderer { return funnelShowLabel() }},
	{Slug: "gauge-basic", Title: "Basic gauge", Group: "Statistical", Source: statisticalExamplesSource + "gauge.go", Build: func() render.Renderer { return gaugeBase() }},
	{Slug: "gauge-timer", Title: "Gauge timer state", Group: "Statistical", Source: statisticalExamplesSource + "gauge.go", Build: func() render.Renderer { return gaugeTimer() }},
	{Slug: "pie-basic", Title: "Basic pie", Group: "Statistical", Source: statisticalExamplesSource + "pie.go", Build: func() render.Renderer { return pieBase() }},
	{Slug: "pie-label", Title: "Pie labels", Group: "Statistical", Source: statisticalExamplesSource + "pie.go", Build: func() render.Renderer { return pieShowLabel() }},
	{Slug: "pie-radius", Title: "Pie radius", Group: "Statistical", Source: statisticalExamplesSource + "pie.go", Build: func() render.Renderer { return pieRadius() }},
	{Slug: "pie-radius-pad-angle", Title: "Pie radius with pad angle", Group: "Statistical", Source: statisticalExamplesSource + "pie.go", Build: func() render.Renderer { return pieRadiusWithPadAngle() }},
	{Slug: "pie-rose-area", Title: "Pie rose area", Group: "Statistical", Source: statisticalExamplesSource + "pie.go", Build: func() render.Renderer { return pieRoseArea() }},
	{Slug: "pie-rose-radius", Title: "Pie rose radius", Group: "Statistical", Source: statisticalExamplesSource + "pie.go", Build: func() render.Renderer { return pieRoseRadius() }},
	{Slug: "pie-rose-area-radius", Title: "Pie rose area and radius", Group: "Statistical", Source: statisticalExamplesSource + "pie.go", Build: func() render.Renderer { return pieRoseAreaRadius() }},
	{Slug: "pie-in-pie", Title: "Pie in pie", Group: "Statistical", Source: statisticalExamplesSource + "pie.go", Build: func() render.Renderer { return pieInPie() }},
	{Slug: "pie-dispatch-action", Title: "Pie dispatch action state", Group: "Statistical", Source: statisticalExamplesSource + "pie.go", Build: func() render.Renderer { return pieWithDispatchAction() }},
	{Slug: "radar-basic", Title: "Basic radar", Group: "Statistical", Source: statisticalExamplesSource + "radar.go", Build: func() render.Renderer { return radarBase() }},
	{Slug: "radar-style", Title: "Styled radar", Group: "Statistical", Source: statisticalExamplesSource + "radar.go", Build: func() render.Renderer { return radarStyle() }},
	{Slug: "radar-legend-multi", Title: "Radar multi legend", Group: "Statistical", Source: statisticalExamplesSource + "radar.go", Build: func() render.Renderer { return radarLegendMulti() }},
	{Slug: "radar-legend-single", Title: "Radar single legend", Group: "Statistical", Source: statisticalExamplesSource + "radar.go", Build: func() render.Renderer { return radarLegendSingle() }},
}

var boxPlotValues = [][]float64{
	{650, 740, 850, 880, 900, 930, 950, 960, 980, 1000, 1070},
	{760, 790, 800, 810, 830, 840, 850, 880, 900, 940, 960},
	{620, 720, 840, 840, 850, 860, 870, 880, 880, 910, 970},
	{720, 740, 760, 780, 800, 810, 820, 850, 860, 890, 920},
	{740, 760, 780, 800, 810, 810, 820, 840, 850, 870, 950},
}

func boxPlotItems(values [][]float64) []opts.BoxPlotData {
	items := make([]opts.BoxPlotData, 0, len(values))
	for _, value := range values {
		items = append(items, opts.BoxPlotData{Value: []float64{value[0], value[2], value[5], value[8], value[len(value)-1]}})
	}
	return items
}

func boxPlotBase() *charts.BoxPlot {
	chart := charts.NewBoxPlot()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "basic boxplot example"}))
	chart.SetXAxis([]string{"expr1", "expr2", "expr3", "expr4", "expr5"}).AddSeries("boxplot", boxPlotItems(boxPlotValues))
	return chart
}

func boxPlotMulti() *charts.BoxPlot {
	chart := charts.NewBoxPlot()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "boxplot with multi-series"}))
	chart.SetXAxis([]string{"expr1", "expr2"}).AddSeries("series1", boxPlotItems(boxPlotValues[:2])).AddSeries("series2", boxPlotItems(boxPlotValues[2:]))
	return chart
}

var klineDates = []string{"2018/1/24", "2018/1/25", "2018/1/28", "2018/1/29", "2018/1/30", "2018/1/31", "2018/2/1", "2018/2/4", "2018/2/5", "2018/2/6", "2018/2/7", "2018/2/8"}
var klineValues = []opts.KlineData{{Value: [4]float32{2320.26, 2320.26, 2287.3, 2362.94}}, {Value: [4]float32{2300, 2291.3, 2288.26, 2308.38}}, {Value: [4]float32{2295.35, 2346.5, 2295.35, 2346.92}}, {Value: [4]float32{2347.22, 2358.98, 2337.35, 2363.8}}, {Value: [4]float32{2360.75, 2382.48, 2347.89, 2383.76}}, {Value: [4]float32{2383.43, 2385.42, 2371.23, 2391.82}}, {Value: [4]float32{2377.41, 2419.02, 2369.57, 2421.15}}, {Value: [4]float32{2425.92, 2428.15, 2417.58, 2440.38}}, {Value: [4]float32{2411, 2433.13, 2403.3, 2437.42}}, {Value: [4]float32{2432.68, 2434.48, 2427.7, 2441.73}}, {Value: [4]float32{2430.69, 2418.53, 2394.22, 2433.89}}, {Value: [4]float32{2416.62, 2432.4, 2414.4, 2443.03}}}

func newKline(title string, zoom opts.DataZoom) *charts.Kline {
	chart := charts.NewKLine()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: title}), charts.WithXAxisOpts(opts.XAxis{SplitNumber: 20}), charts.WithYAxisOpts(opts.YAxis{Scale: opts.Bool(true)}), charts.WithDataZoomOpts(zoom))
	chart.SetXAxis(klineDates).AddSeries("kline", klineValues)
	return chart
}

func klineBase() *charts.Kline {
	return newKline("Kline-example", opts.DataZoom{Start: 50, End: 100, XAxisIndex: []int{0}})
}
func klineDataZoomInside() *charts.Kline {
	return newKline("DataZoom(inside)", opts.DataZoom{Type: "inside", Start: 50, End: 100, XAxisIndex: []int{0}})
}
func klineDataZoomBoth() *charts.Kline {
	chart := newKline("DataZoom(inside&slider)", opts.DataZoom{Type: "inside", Start: 50, End: 100, XAxisIndex: []int{0}})
	chart.SetGlobalOptions(charts.WithDataZoomOpts(opts.DataZoom{Type: "slider", Start: 50, End: 100, XAxisIndex: []int{0}}))
	return chart
}
func klineDataZoomYAxis() *charts.Kline {
	return newKline("DataZoom(yAxis)", opts.DataZoom{Type: "slider", Start: 50, End: 100, YAxisIndex: []int{0}})
}
func klineStyle() *charts.Kline {
	chart := klineBase()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "different style"}))
	chart.SetSeriesOptions(charts.WithMarkPointNameTypeItemOpts(opts.MarkPointNameTypeItem{Name: "highest value", Type: "max", ValueDim: "highest"}, opts.MarkPointNameTypeItem{Name: "lowest value", Type: "min", ValueDim: "lowest"}), charts.WithMarkPointStyleOpts(opts.MarkPointStyle{Label: &opts.Label{Show: opts.Bool(true)}}), charts.WithItemStyleOpts(opts.ItemStyle{Color: "#ec0000", Color0: "#00da3c", BorderColor: "#8A0000", BorderColor0: "#008F28"}))
	return chart
}

func funnelItems() []opts.FunnelData {
	return []opts.FunnelData{{Name: "Visit", Value: 48}, {Name: "Add", Value: 36}, {Name: "Order", Value: 26}, {Name: "Payment", Value: 19}, {Name: "Deal", Value: 12}}
}
func funnelBase() *charts.Funnel {
	chart := charts.NewFunnel()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "basic funnel example"}))
	chart.AddSeries("Analytics", funnelItems())
	return chart
}
func funnelShowLabel() *charts.Funnel {
	chart := funnelBase()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "show label"}))
	chart.SetSeriesOptions(charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Position: "left"}))
	return chart
}

func gaugeBase() *charts.Gauge  { return newGauge("basic Gauge example", "ProjectA", 43) }
func gaugeTimer() *charts.Gauge { return newGauge("javascript timer", "ProjectB", 64) }
func newGauge(title, series string, value int) *charts.Gauge {
	chart := charts.NewGauge()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: title}))
	chart.AddSeries(series, []opts.GaugeData{{Name: "Work progress", Value: value}})
	return chart
}

func pieItems() []opts.PieData {
	return []opts.PieData{{Name: "Spring", Value: 71}, {Name: "Summer", Value: 92}, {Name: "Autumn", Value: 58}, {Name: "Winter", Value: 36}}
}
func newPie(title string) *charts.Pie {
	chart := charts.NewPie()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: title}))
	chart.AddSeries("pie", pieItems())
	return chart
}
func pieBase() *charts.Pie { return newPie("basic pie example") }
func pieShowLabel() *charts.Pie {
	chart := newPie("label options")
	chart.SetSeriesOptions(charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Formatter: "{b}: {c}"}))
	return chart
}
func pieRadius() *charts.Pie {
	chart := pieShowLabel()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Radius style"}))
	chart.SetSeriesOptions(charts.WithPieChartOpts(opts.PieChart{Radius: []string{"40%", "75%"}}))
	return chart
}
func pieRadiusWithPadAngle() *charts.Pie {
	chart := newPie("Radius style with padAngle")
	chart.SetGlobalOptions(charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Formatter: "{b}: {d}%"}), charts.WithLegendOpts(opts.Legend{Orient: "vertical", Right: "20%", Top: "center", Padding: []any{1, 1, 1, 1}}))
	chart.SetSeriesOptions(charts.WithLabelOpts(opts.Label{Show: opts.Bool(false)}), charts.WithPieChartOpts(opts.PieChart{Radius: []string{"40%", "75%"}, Center: []string{"40%", "50%"}, PadAngle: 5}))
	return chart
}
func pieRoseArea() *charts.Pie {
	chart := pieShowLabel()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Rose(Area)"}))
	chart.SetSeriesOptions(charts.WithPieChartOpts(opts.PieChart{Radius: []string{"40%", "75%"}, RoseType: "area"}))
	return chart
}
func pieRoseRadius() *charts.Pie {
	chart := pieShowLabel()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Rose(Radius)"}))
	chart.SetSeriesOptions(charts.WithPieChartOpts(opts.PieChart{Radius: []string{"30%", "75%"}, RoseType: "radius"}))
	return chart
}
func pieRoseAreaRadius() *charts.Pie {
	chart := charts.NewPie()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Rose(Area/Radius)"}))
	chart.AddSeries("area", pieItems(), charts.WithPieChartOpts(opts.PieChart{Radius: []string{"30%", "75%"}, RoseType: "area", Center: []string{"25%", "50%"}})).AddSeries("pie", pieItems(), charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Formatter: "{b}: {c}"}), charts.WithPieChartOpts(opts.PieChart{Radius: []string{"30%", "75%"}, RoseType: "radius", Center: []string{"75%", "50%"}}))
	return chart
}
func pieInPie() *charts.Pie {
	chart := charts.NewPie()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "pie in pie"}))
	chart.AddSeries("area", pieItems(), charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Formatter: "{b}: {c}"}), charts.WithPieChartOpts(opts.PieChart{Radius: []string{"50%", "55%"}, RoseType: "area"})).AddSeries("radius", pieItems(), charts.WithPieChartOpts(opts.PieChart{Radius: []string{"0%", "45%"}, RoseType: "radius"}))
	return chart
}
func pieWithDispatchAction() *charts.Pie {
	chart := newPie("dispatchAction pie")
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "dispatchAction pie", Right: "40%"}), charts.WithTooltipOpts(opts.Tooltip{Trigger: "item", Formatter: "{a} <br/>{b} : {c} ({d}%)"}), charts.WithLegendOpts(opts.Legend{Left: "left", Orient: "vertical"}))
	chart.SetSeriesOptions(charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Formatter: "{b}: {c}"}), charts.WithPieChartOpts(opts.PieChart{Radius: []string{"55%"}, Center: []string{"50%", "60%"}}), charts.WithEmphasisOpts(opts.Emphasis{ItemStyle: &opts.ItemStyle{ShadowBlur: 10, ShadowOffsetX: 0, ShadowColor: "rgba(0, 0, 0, 0.5)"}}))
	return chart
}

var radarIndicators = []*opts.Indicator{{Name: "AQI", Max: 300}, {Name: "PM2.5", Max: 250}, {Name: "PM10", Max: 300}, {Name: "CO", Max: 5}, {Name: "NO2", Max: 200}, {Name: "SO2", Max: 100}}
var radarBeijing = []opts.RadarData{{Value: []float32{55, 9, 56, .46, 18, 6}}, {Value: []float32{82, 58, 90, 1.77, 68, 33}}, {Value: []float32{185, 127, 216, 2.52, 61, 27}}}
var radarGuangzhou = []opts.RadarData{{Value: []float32{26, 37, 27, 1.163, 27, 13}}, {Value: []float32{85, 62, 71, 1.195, 60, 8}}, {Value: []float32{91, 81, 104, 1.041, 56, 40}}}
var radarShanghai = []opts.RadarData{{Value: []float32{91, 45, 125, .82, 34, 23}}, {Value: []float32{109, 81, 121, 1.28, 68, 51}}, {Value: []float32{134, 83, 167, 1.16, 57, 43}}}

func radarBase() *charts.Radar {
	chart := charts.NewRadar()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "basic radar example"}), charts.WithRadarComponentOpts(opts.RadarComponent{Indicator: radarIndicators, SplitArea: &opts.SplitArea{Show: opts.Bool(true)}, SplitLine: &opts.SplitLine{Show: opts.Bool(true)}}))
	chart.AddSeries("Beijing", radarBeijing)
	return chart
}
func styledRadar(title string) *charts.Radar {
	chart := charts.NewRadar()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: title, Right: "center", TitleStyle: &opts.TextStyle{Color: "#eee"}}), charts.WithInitializationOpts(opts.Initialization{BackgroundColor: "#161627"}), charts.WithRadarComponentOpts(opts.RadarComponent{Indicator: radarIndicators, Shape: "circle", SplitNumber: 5, SplitLine: &opts.SplitLine{Show: opts.Bool(true), LineStyle: &opts.LineStyle{Opacity: opts.Float(.1)}}}), charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true), Bottom: "5px", TextStyle: &opts.TextStyle{Color: "#eee"}}))
	return chart
}
func radarStyle() *charts.Radar {
	chart := styledRadar("style options")
	chart.AddSeries("Beijing", radarBeijing).SetSeriesOptions(charts.WithLineStyleOpts(opts.LineStyle{Width: 1, Opacity: opts.Float(.5)}), charts.WithAreaStyleOpts(opts.AreaStyle{Opacity: opts.Float(.1)}), charts.WithItemStyleOpts(opts.ItemStyle{Color: "#F9713C"}))
	return chart
}
func radarLegendMulti() *charts.Radar {
	chart := styledRadar("Legend(Multi)")
	chart.AddSeries("Beijing", radarBeijing, charts.WithItemStyleOpts(opts.ItemStyle{Color: "#F9713C"})).AddSeries("Guangzhou", radarGuangzhou, charts.WithItemStyleOpts(opts.ItemStyle{Color: "#B3E4A1"})).AddSeries("Shanghai", radarShanghai, charts.WithItemStyleOpts(opts.ItemStyle{Color: "rgb(238, 197, 102)"})).SetSeriesOptions(charts.WithLineStyleOpts(opts.LineStyle{Width: 1, Opacity: opts.Float(.5)}), charts.WithAreaStyleOpts(opts.AreaStyle{Opacity: opts.Float(.1)}))
	return chart
}
func radarLegendSingle() *charts.Radar {
	chart := styledRadar("Legend(Single)")
	chart.SetGlobalOptions(charts.WithLegendOpts(opts.Legend{Show: opts.Bool(true), Bottom: "5px", TextStyle: &opts.TextStyle{Color: "#eee"}, SelectedMode: "single"}))
	chart.AddSeries("Beijing", radarBeijing, charts.WithItemStyleOpts(opts.ItemStyle{Color: "#F9713C"})).AddSeries("Guangzhou", radarGuangzhou, charts.WithItemStyleOpts(opts.ItemStyle{Color: "#B3E4A1"})).AddSeries("Shanghai", radarShanghai, charts.WithItemStyleOpts(opts.ItemStyle{Color: "rgb(238, 197, 102)"})).SetSeriesOptions(charts.WithLineStyleOpts(opts.LineStyle{Width: 1, Opacity: opts.Float(.5)}), charts.WithAreaStyleOpts(opts.AreaStyle{Opacity: opts.Float(.5)}))
	return chart
}
