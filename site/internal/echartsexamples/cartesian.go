package echartsexamples

import (
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/render"
)

const cartesianSource = "https://github.com/go-echarts/examples/blob/master/examples/"

var (
	weeks  = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	fruits = []string{"Apple", "Banana", "Peach", "Lemon", "Pear", "Cherry"}
	sports = []string{"Swimming", "Surfing", "Shooting", "Skating", "Wrestling", "Diving"}
	player = []string{"Kobe", "Jordan", "Iverson", "LeBron", "Wade", "McGrady"}
)

// CartesianExamples is the deterministic port of bar, line, scatter,
// effectscatter, and heatmap upstream examples.
var CartesianExamples = []Example{
	example("bar-basic", "Basic bar", "bar", "bar.go", func() render.Renderer { return barBasic() }), example("bar-title", "Title and legend", "bar", "bar.go", func() render.Renderer { return barTitle() }), example("bar-tooltip", "Tooltip", "bar", "bar.go", func() render.Renderer { return barTooltip() }), example("bar-toolbox", "Toolbox", "bar", "bar.go", func() render.Renderer { return barSetToolbox() }), example("bar-label", "Labels", "bar", "bar.go", func() render.Renderer { return barShowLabel() }), example("bar-axis-names", "Axis names", "bar", "bar.go", func() render.Renderer { return barXYName() }), example("bar-axis-formatter", "Axis formatter", "bar", "bar.go", func() render.Renderer { return barXYFormatter() }), example("bar-color", "Colors", "bar", "bar.go", func() render.Renderer { return barColor() }), example("bar-split-line", "Split lines", "bar", "bar.go", func() render.Renderer { return barSplitLine() }), example("bar-gap", "Bar gap", "bar", "bar.go", func() render.Renderer { return barGap() }), example("bar-datazoom-inside", "Inside data zoom", "bar", "bar.go", func() render.Renderer { return barDataZoomInside() }), example("bar-datazoom-slider", "Slider data zoom", "bar", "bar.go", func() render.Renderer { return barDataZoomSlider() }), example("bar-reverse", "Reversed axes", "bar", "bar.go", func() render.Renderer { return barReverse() }), example("bar-stack", "Stacked bars", "bar", "bar.go", func() render.Renderer { return barStack() }), example("bar-mark-points", "Mark points", "bar", "bar.go", func() render.Renderer { return barMarkPoints() }), example("bar-mark-lines", "Mark lines", "bar", "bar.go", func() render.Renderer { return barMarkLines() }), example("bar-overlap", "Overlapped charts", "bar", "bar.go", func() render.Renderer { return barOverlap() }), example("bar-size", "Canvas size", "bar", "bar.go", func() render.Renderer { return barSize() }), example("bar-width", "Bar widths", "bar", "bar.go", func() render.Renderer { return barWidth() }),
	example("line-basic", "Basic line", "line", "line.go", func() render.Renderer { return lineBase() }), example("line-label", "Line labels", "line", "line.go", func() render.Renderer { return lineShowLabel() }), example("line-mark-point", "Line mark points", "line", "line.go", func() render.Renderer { return lineMarkPoint() }), example("line-split-line", "Line split lines", "line", "line.go", func() render.Renderer { return lineSplitLine() }), example("line-numerical", "Numerical X axis", "line", "line.go", func() render.Renderer { return lineNumerical() }), example("line-time", "Temporal X axis", "line", "line.go", func() render.Renderer { return lineTime() }), example("line-step", "Step line", "line", "line.go", func() render.Renderer { return lineStep() }), example("line-smooth", "Smooth line", "line", "line.go", func() render.Renderer { return lineSmooth() }), example("line-area", "Area line", "line", "line.go", func() render.Renderer { return lineArea() }), example("line-smooth-area", "Smooth area", "line", "line.go", func() render.Renderer { return lineSmoothArea() }), example("line-overlap", "Overlapped charts", "line", "line.go", func() render.Renderer { return lineOverlap() }), example("line-multi", "Multiple lines", "line", "line.go", func() render.Renderer { return lineMulti() }), example("line-demo", "Search benchmark", "line", "line.go", func() render.Renderer { return lineDemo() }), example("line-symbols", "Line symbols", "line", "line.go", func() render.Renderer { return lineSymbols() }),
	example("scatter-basic", "Basic scatter", "scatter", "scatter.go", func() render.Renderer { return scatterBase() }), example("scatter-label", "Scatter labels", "scatter", "scatter.go", func() render.Renderer { return scatterShowLabel() }), example("scatter-split-line", "Scatter split lines", "scatter", "scatter.go", func() render.Renderer { return scatterSplitLine() }),
	example("effectscatter-basic", "Basic effect scatter", "effectscatter", "effectscatter.go", func() render.Renderer { return esBase() }), example("effectscatter-wave", "Wave style", "effectscatter", "effectscatter.go", func() render.Renderer { return esEffectStyle() }),
	example("heatmap-basic", "Basic heatmap", "heatmap", "heatmap.go", func() render.Renderer { return heatMapBase() }), example("heatmap-calendar", "Calendar heatmap", "heatmap", "heatmap.go", func() render.Renderer { return heatMapCalendar() }),
}

func example(slug, title, group, source string, build func() render.Renderer) Example {
	return Example{Slug: slug, Title: title, Group: group, Source: cartesianSource + source, Build: build}
}

func barData(offset int) []opts.BarData {
	return []opts.BarData{{Value: 42 + offset}, {Value: 91 + offset}, {Value: 67 + offset}, {Value: 128 + offset}, {Value: 83 + offset}, {Value: 156 + offset}, {Value: 104 + offset}}
}
func lineData(offset int) []opts.LineData {
	return []opts.LineData{{Value: 31 + offset}, {Value: 82 + offset}, {Value: 58 + offset}, {Value: 121 + offset}, {Value: 76 + offset}, {Value: 143 + offset}}
}
func scatterData(offset int) []opts.ScatterData {
	return []opts.ScatterData{{Value: 18 + offset, Symbol: "roundRect", SymbolSize: 20, SymbolRotate: 10}, {Value: 67 + offset, Symbol: "roundRect", SymbolSize: 20, SymbolRotate: 10}, {Value: 44 + offset, Symbol: "roundRect", SymbolSize: 20, SymbolRotate: 10}, {Value: 82 + offset, Symbol: "roundRect", SymbolSize: 20, SymbolRotate: 10}, {Value: 53 + offset, Symbol: "roundRect", SymbolSize: 20, SymbolRotate: 10}, {Value: 91 + offset, Symbol: "roundRect", SymbolSize: 20, SymbolRotate: 10}}
}
func effectScatterData(offset int) []opts.EffectScatterData {
	return []opts.EffectScatterData{{Value: 88 + offset}, {Value: 96 + offset}, {Value: 72 + offset}, {Value: 94 + offset}, {Value: 79 + offset}, {Value: 85 + offset}}
}

func titledBar(title string) *charts.Bar {
	b := charts.NewBar()
	b.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: title}))
	b.SetXAxis(weeks).AddSeries("Category A", barData(0)).AddSeries("Category B", barData(18))
	return b
}
func barBasic() *charts.Bar {
	b := titledBar("basic bar example")
	b.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "basic bar example", Subtitle: "This is the subtitle."}))
	return b
}
func barTitle() *charts.Bar {
	b := titledBar("title and legend options")
	b.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "title and legend options", Subtitle: "go-echarts is an awesome chart library written in Golang", Link: "https://github.com/go-echarts/go-echarts", Right: "40%"}), charts.WithLegendOpts(opts.Legend{Right: "80%"}))
	return b
}
func barTooltip() *charts.Bar {
	b := titledBar("tooltip options")
	b.SetGlobalOptions(charts.WithLegendOpts(opts.Legend{Right: "80px"}))
	return b
}
func barSetToolbox() *charts.Bar {
	b := titledBar("toolbox options")
	b.SetGlobalOptions(charts.WithToolboxOpts(opts.Toolbox{Right: "20%", Feature: &opts.ToolBoxFeature{SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{Type: "jpg", Title: "Anything you want"}, DataView: &opts.ToolBoxFeatureDataView{Title: "DataView", Lang: []string{"data view", "turn off", "refresh"}}}}))
	return b
}
func barShowLabel() *charts.Bar {
	b := titledBar("label options")
	b.SetSeriesOptions(charts.WithLabelOpts(opts.Label{Position: "top"}))
	return b
}
func barXYName() *charts.Bar {
	b := titledBar("display the axes name")
	b.SetGlobalOptions(charts.WithXAxisOpts(opts.XAxis{Name: "XAxisName"}), charts.WithYAxisOpts(opts.YAxis{Name: "YAxisName"}))
	return b
}
func barXYFormatter() *charts.Bar {
	b := titledBar("customized the xaxis/yaxis formatter")
	b.SetGlobalOptions(charts.WithXAxisOpts(opts.XAxis{AxisLabel: &opts.AxisLabel{Formatter: "{value} x-unit"}}), charts.WithYAxisOpts(opts.YAxis{AxisLabel: &opts.AxisLabel{Formatter: "{value} y-unit"}}))
	return b
}
func barColor() *charts.Bar {
	b := titledBar("set user-defined colors")
	b.SetGlobalOptions(charts.WithColorsOpts(opts.Colors{"blue", "pink"}))
	return b
}
func barSplitLine() *charts.Bar {
	b := titledBar("splitline options")
	b.SetGlobalOptions(charts.WithXAxisOpts(opts.XAxis{Name: "XAxisName", SplitLine: &opts.SplitLine{Show: opts.Bool(true)}}), charts.WithYAxisOpts(opts.YAxis{Name: "YAxisName", SplitLine: &opts.SplitLine{Show: opts.Bool(true)}}))
	return b
}
func barGap() *charts.Bar {
	b := titledBar("set the gap of each bar")
	b.SetSeriesOptions(charts.WithBarChartOpts(opts.BarChart{BarGap: "150%"}))
	return b
}
func barDataZoomInside() *charts.Bar {
	b := titledBar("datazoom options(inside)")
	b.SetGlobalOptions(charts.WithDataZoomOpts(opts.DataZoom{Type: "inside", Start: 10, End: 50}))
	return b
}
func barDataZoomSlider() *charts.Bar {
	b := titledBar("datazoom options(slider)")
	b.SetGlobalOptions(charts.WithDataZoomOpts(opts.DataZoom{Type: "slider", Start: 10, End: 50}))
	return b
}
func barReverse() *charts.Bar { b := titledBar("reverse xaxis and yaxis"); b.XYReversal(); return b }
func barStack() *charts.Bar {
	b := titledBar("stack style")
	b.SetSeriesOptions(charts.WithBarChartOpts(opts.BarChart{Stack: "stackA"}))
	return b
}
func barMarkPoints() *charts.Bar {
	b := titledBar("markpoint options")
	b.SetSeriesOptions(charts.WithMarkPointNameTypeItemOpts(opts.MarkPointNameTypeItem{Name: "Maximum", Type: "max"}, opts.MarkPointNameTypeItem{Name: "Minimum", Type: "min"}))
	return b
}
func barMarkLines() *charts.Bar {
	b := titledBar("markline options")
	b.SetSeriesOptions(charts.WithMarkLineNameTypeItemOpts(opts.MarkLineNameTypeItem{Name: "Maximum", Type: "max"}, opts.MarkLineNameTypeItem{Name: "Avg", Type: "average"}))
	return b
}
func barOverlap() *charts.Bar {
	b := titledBar("overlap rect-charts")
	b.Overlap(lineBase())
	b.Overlap(scatterBase())
	return b
}
func barSize() *charts.Bar {
	b := titledBar("adjust canvas size")
	b.SetGlobalOptions(charts.WithInitializationOpts(opts.Initialization{Width: "1200px", Height: "600px"}))
	return b
}
func barWidth() *charts.Bar {
	b := titledBar("adjust width of each bar")
	b.SetSeriesOptions(charts.WithBarChartOpts(opts.BarChart{BarWidth: "35"}))
	return b
}

func titledLine(title string) *charts.Line {
	l := charts.NewLine()
	l.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: title}))
	l.SetXAxis(fruits).AddSeries("Category A", lineData(0))
	return l
}
func lineBase() *charts.Line {
	l := titledLine("basic line example")
	l.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "basic line example", Subtitle: "This is the subtitle."}))
	return l
}
func lineShowLabel() *charts.Line {
	l := titledLine("title and label options")
	l.SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{ShowSymbol: opts.Bool(true)}), charts.WithLabelOpts(opts.Label{Show: opts.Bool(true)}))
	return l
}
func lineMarkPoint() *charts.Line {
	l := titledLine("markpoint options")
	l.SetSeriesOptions(charts.WithMarkPointNameTypeItemOpts(opts.MarkPointNameTypeItem{Name: "Maximum", Type: "max"}, opts.MarkPointNameTypeItem{Name: "Average", Type: "average"}, opts.MarkPointNameTypeItem{Name: "Minimum", Type: "min"}))
	return l
}
func lineSplitLine() *charts.Line {
	l := titledLine("splitline options")
	l.SetGlobalOptions(charts.WithYAxisOpts(opts.YAxis{SplitLine: &opts.SplitLine{Show: opts.Bool(true)}}))
	l.SetSeriesOptions(charts.WithLabelOpts(opts.Label{Show: opts.Bool(true)}))
	return l
}
func lineNumerical() *charts.Line {
	l := charts.NewLine()
	l.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "numerical X axis & accessories", Subtitle: "styled mark areas, mark lines and visual maps"}), charts.WithYAxisOpts(opts.YAxis{Max: 200}))
	points := make([]opts.LineData, 30)
	for i := range points {
		points[i] = opts.LineData{Value: []interface{}{i, 101 + (i*7)%19}}
	}
	l.AddSeries("Category A", points, charts.WithLineChartOpts(opts.LineChart{Symbol: "triangle", SymbolSize: 10}), charts.WithAreaStyleOpts(opts.AreaStyle{}))
	return l
}
func lineTime() *charts.Line {
	l := charts.NewLine()
	start := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	l.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "temporal X axis", Subtitle: "time.Date as X axis values"}), charts.WithYAxisOpts(opts.YAxis{Min: 0, Max: 200}), charts.WithXAxisOpts(opts.XAxis{Type: "time", Min: start}), charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}))
	values := make([]opts.LineData, 50)
	for i := range values {
		values[i] = opts.LineData{Value: []interface{}{start.AddDate(0, 1, i), 102 + (i*9)%20}}
	}
	l.AddSeries("Category A", values)
	return l
}
func lineStep() *charts.Line {
	l := titledLine("step style")
	l.SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Step: true}))
	return l
}
func lineSmooth() *charts.Line {
	l := titledLine("smooth style")
	l.SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}))
	return l
}
func lineArea() *charts.Line {
	l := titledLine("area options")
	l.SetSeriesOptions(charts.WithLabelOpts(opts.Label{Show: opts.Bool(true)}), charts.WithAreaStyleOpts(opts.AreaStyle{Opacity: opts.Float(0.5)}))
	return l
}
func lineSmoothArea() *charts.Line {
	l := titledLine("smooth area")
	l.SetSeriesOptions(charts.WithLabelOpts(opts.Label{Show: opts.Bool(true)}), charts.WithAreaStyleOpts(opts.AreaStyle{Opacity: opts.Float(0.2)}), charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}))
	return l
}
func lineOverlap() *charts.Line {
	l := titledLine("overlap rect-charts")
	l.Overlap(esEffectStyle())
	l.Overlap(scatterBase())
	return l
}
func lineMulti() *charts.Line {
	l := titledLine("multi lines")
	l.SetGlobalOptions(charts.WithInitializationOpts(opts.Initialization{Theme: "shine"}))
	l.AddSeries("Category B", lineData(13)).AddSeries("Category C", lineData(27)).AddSeries("Category D", lineData(41))
	return l
}
func lineDemo() *charts.Line {
	l := charts.NewLine()
	l.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Search Time: Hash table vs Binary search"}), charts.WithYAxisOpts(opts.YAxis{Name: "Cost time(ns)", SplitLine: &opts.SplitLine{Show: opts.Bool(true)}}), charts.WithXAxisOpts(opts.XAxis{Name: "Elements"}))
	l.SetXAxis([]string{"10e1", "10e2", "10e3", "10e4", "10e5", "10e6", "10e7"}).AddSeries("map", []opts.LineData{{Value: 31}, {Value: 47}, {Value: 63}, {Value: 81}, {Value: 107}, {Value: 139}, {Value: 166}}).AddSeries("slice", []opts.LineData{{Value: 24.9}, {Value: 34.9}, {Value: 48.1}, {Value: 58.3}, {Value: 69.7}, {Value: 123}, {Value: 131}})
	return l
}
func lineSymbols() *charts.Line {
	l := titledLine("symbol options")
	l.AddSeries("Category B", lineData(22))
	l.SetGlobalOptions(charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}))
	l.SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true), ShowSymbol: opts.Bool(true), SymbolSize: 15, Symbol: "diamond"}))
	return l
}

func titledScatter(title string) *charts.Scatter {
	s := charts.NewScatter()
	s.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: title}))
	s.SetXAxis(sports).AddSeries("Category A", scatterData(0)).AddSeries("Category B", scatterData(8))
	return s
}
func scatterBase() *charts.Scatter { return titledScatter("basic scatter example") }
func scatterShowLabel() *charts.Scatter {
	s := titledScatter("label options")
	s.SetSeriesOptions(charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Position: "right"}))
	return s
}
func scatterSplitLine() *charts.Scatter {
	s := titledScatter("splitline options")
	s.SetGlobalOptions(charts.WithXAxisOpts(opts.XAxis{Name: "Sports", SplitLine: &opts.SplitLine{Show: opts.Bool(true)}}), charts.WithYAxisOpts(opts.YAxis{Name: "Score", SplitLine: &opts.SplitLine{Show: opts.Bool(true)}}))
	return s
}

func esBase() *charts.EffectScatter {
	e := charts.NewEffectScatter()
	e.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "basic EffectScatter example"}))
	e.SetXAxis(player).AddSeries("Dunk", effectScatterData(0))
	return e
}
func esEffectStyle() *charts.EffectScatter {
	e := charts.NewEffectScatter()
	e.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "wave style"}))
	e.SetXAxis(player).AddSeries("Dunk", effectScatterData(0), charts.WithRippleEffectOpts(opts.RippleEffect{Period: 4, Scale: 10, BrushType: "stroke"})).AddSeries("Shoot", effectScatterData(-12), charts.WithRippleEffectOpts(opts.RippleEffect{Period: 3, Scale: 6, BrushType: "fill"}))
	return e
}

func heatMapData() []opts.HeatMapData {
	data := make([]opts.HeatMapData, 0, 168)
	for day := 0; day < 7; day++ {
		for hour := 0; hour < 24; hour++ {
			value := (day*5 + hour*3) % 15
			var point interface{} = value
			if value == 0 {
				point = "-"
			}
			data = append(data, opts.HeatMapData{Value: [3]interface{}{hour, day, point}})
		}
	}
	return data
}
func heatMapBase() *charts.HeatMap {
	h := charts.NewHeatMap()
	days := []string{"Saturday", "Friday", "Thursday", "Wednesday", "Tuesday", "Monday", "Sunday"}
	hours := []string{"12a", "1a", "2a", "3a", "4a", "5a", "6a", "7a", "8a", "9a", "10a", "11a", "12p", "1p", "2p", "3p", "4p", "5p", "6p", "7p", "8p", "9p", "10p", "11p"}
	h.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "basic heatmap example"}), charts.WithXAxisOpts(opts.XAxis{Type: "category", SplitArea: &opts.SplitArea{Show: opts.Bool(true)}}), charts.WithYAxisOpts(opts.YAxis{Type: "category", Data: days, SplitArea: &opts.SplitArea{Show: opts.Bool(true)}}), charts.WithVisualMapOpts(opts.VisualMap{Calculable: opts.Bool(true), Min: 0, Max: 10, InRange: &opts.VisualMapInRange{Color: []string{"#50a3ba", "#eac736", "#d94e5d"}}}))
	h.SetXAxis(hours).AddSeries("heatmap", heatMapData())
	return h
}
func heatMapCalendar() *charts.HeatMap {
	h := charts.NewHeatMap()
	start := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(1, 0, 0)
	data := make([]opts.HeatMapData, 0, 365)
	for date := start; date.Before(end); date = date.AddDate(0, 0, 1) {
		value := (date.YearDay()*7 + 3) % 21
		var point interface{} = value
		if value == 0 {
			point = "-"
		}
		data = append(data, opts.HeatMapData{Value: [2]interface{}{date.Format("2006-01-02"), point}})
	}
	h.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "calendar heatmap example"}), charts.WithVisualMapOpts(opts.VisualMap{Min: 0, Max: 20, InRange: &opts.VisualMapInRange{Color: []string{"#50a3ba", "#eac736", "#d94e5d"}}}))
	h.AddCalendar(&opts.Calendar{Top: "80", Left: "30", Right: "30", CellSize: "20", ItemStyle: &opts.ItemStyle{BorderWidth: 0.5}, Orient: "horizontal", Range: []string{start.Format("2006-01-02"), end.Format("2006-01-02")}}).AddSeries("heatmap calendar", data, charts.WithCoordinateSystem("calendar"))
	return h
}
