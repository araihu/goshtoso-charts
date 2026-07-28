package echartsexamples

import (
	"math"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/render"
)

const webGLExamplesSource = "https://github.com/go-echarts/examples/blob/master/examples/"

// WebGLExamples ports every Bar3D, Line3D, Scatter3D, and Surface3D builder
// from the upstream examples. Values are static or deterministically derived,
// so a rendered page never changes between requests.
var WebGLExamples = []Example{
	{Slug: "bar3d-basic", Title: "Basic Bar3D", Group: "WebGL", Source: webGLExamplesSource + "bar3d.go", Build: func() render.Renderer { return bar3DBase() }},
	{Slug: "bar3d-auto-rotate", Title: "Auto-rotating Bar3D", Group: "WebGL", Source: webGLExamplesSource + "bar3d.go", Build: func() render.Renderer { return bar3DAutoRotate() }},
	{Slug: "bar3d-rotate-speed", Title: "Fast rotating Bar3D", Group: "WebGL", Source: webGLExamplesSource + "bar3d.go", Build: func() render.Renderer { return bar3DRotateSpeed() }},
	{Slug: "bar3d-shading", Title: "Lambert Bar3D", Group: "WebGL", Source: webGLExamplesSource + "bar3d.go", Build: func() render.Renderer { return bar3DShading() }},
	{Slug: "line3d-basic", Title: "Basic Line3D", Group: "WebGL", Source: webGLExamplesSource + "line3d.go", Build: func() render.Renderer { return line3DBase() }},
	{Slug: "line3d-auto-rotate", Title: "Auto-rotating Line3D", Group: "WebGL", Source: webGLExamplesSource + "line3d.go", Build: func() render.Renderer { return line3DAutoRotate() }},
	{Slug: "scatter3d-basic", Title: "Basic Scatter3D", Group: "WebGL", Source: webGLExamplesSource + "scatter3d.go", Build: func() render.Renderer { return scatter3DBase() }},
	{Slug: "scatter3d-item-style", Title: "Styled Scatter3D", Group: "WebGL", Source: webGLExamplesSource + "scatter3d.go", Build: func() render.Renderer { return scatter3DDataItem() }},
	{Slug: "surface3d-basic", Title: "Basic Surface3D", Group: "WebGL", Source: webGLExamplesSource + "surface3d.go", Build: func() render.Renderer { return surface3DBase() }},
	{Slug: "surface3d-rose", Title: "Rose Surface3D", Group: "WebGL", Source: webGLExamplesSource + "surface3d.go", Build: func() render.Renderer { return surface3DRose() }},
}

var (
	webGLRangeColor = []string{"#313695", "#4575b4", "#74add1", "#abd9e9", "#e0f3f8", "#fee090", "#fdae61", "#f46d43", "#d73027", "#a50026"}
	bar3DHrs        = []string{"12a", "1a", "2a", "3a", "4a", "5a", "6a", "7a", "8a", "9a", "10a", "11a", "12p", "1p", "2p", "3p", "4p", "5p", "6p", "7p", "8p", "9p", "10p", "11p"}
	bar3DDays       = []string{"Saturday", "Friday", "Thursday", "Wednesday", "Tuesday", "Monday", "Sunday"}
)

func bar3DData() []opts.Chart3DData {
	data := make([]opts.Chart3DData, 0, len(bar3DHrs)*len(bar3DDays))
	for day := range bar3DDays {
		for hour := range bar3DHrs {
			// Deterministic hourly activity profile, preserving upstream axes.
			value := (day*7 + hour*hour + 3*hour) % 15
			data = append(data, opts.Chart3DData{Value: []interface{}{hour, day, value}})
		}
	}
	return data
}

func bar3D(title string, grid opts.Grid3D) *charts.Bar3D {
	chart := charts.NewBar3D()
	chart.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: title}),
		charts.WithVisualMapOpts(opts.VisualMap{Calculable: opts.Bool(true), Max: 30, Range: []float32{0, 30}, InRange: &opts.VisualMapInRange{Color: webGLRangeColor}}),
		charts.WithGrid3DOpts(grid),
		charts.WithXAxis3DOpts(opts.XAxis3D{Data: bar3DHrs}),
		charts.WithYAxis3DOpts(opts.YAxis3D{Data: bar3DDays}),
	)
	chart.AddSeries("bar3d", bar3DData())
	return chart
}

func bar3DBase() *charts.Bar3D {
	return bar3D("basic bar3d example", opts.Grid3D{BoxWidth: 200, BoxDepth: 80})
}
func bar3DAutoRotate() *charts.Bar3D {
	return bar3D("auto rotating", opts.Grid3D{BoxWidth: 160, BoxDepth: 80, ViewControl: &opts.ViewControl{AutoRotate: opts.Bool(true)}})
}
func bar3DRotateSpeed() *charts.Bar3D {
	return bar3D("rotating faster", opts.Grid3D{BoxWidth: 160, BoxDepth: 80, ViewControl: &opts.ViewControl{AutoRotate: opts.Bool(true), AutoRotateSpeed: 200}})
}
func bar3DShading() *charts.Bar3D {
	chart := bar3D("Bar3D-shading(lambert)", opts.Grid3D{BoxWidth: 200, BoxDepth: 80})
	chart.SetSeriesOptions(charts.WithBar3DChartOpts(opts.Bar3DChart{Shading: "lambert"}))
	return chart
}

func line3DData() []opts.Chart3DData {
	data := make([]opts.Chart3DData, 0, 25000)
	for i := 0; i < 25000; i++ {
		t := float64(i) / 1000
		data = append(data, opts.Chart3DData{Value: []interface{}{
			(1 + 0.25*math.Cos(75*t)) * math.Cos(t),
			(1 + 0.25*math.Cos(75*t)) * math.Sin(t),
			t + 2*math.Sin(75*t),
		}})
	}
	return data
}

func line3DBase() *charts.Line3D {
	chart := charts.NewLine3D()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "basic line3d example"}), charts.WithVisualMapOpts(opts.VisualMap{Calculable: opts.Bool(true), Max: 30, InRange: &opts.VisualMapInRange{Color: webGLRangeColor}}))
	chart.AddSeries("line3D", line3DData())
	return chart
}
func line3DAutoRotate() *charts.Line3D {
	chart := charts.NewLine3D()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "auto rotating"}), charts.WithVisualMapOpts(opts.VisualMap{Calculable: opts.Bool(true), Max: 30, InRange: &opts.VisualMapInRange{Color: webGLRangeColor}}), charts.WithGrid3DOpts(opts.Grid3D{ViewControl: &opts.ViewControl{AutoRotate: opts.Bool(true)}}))
	chart.AddSeries("line3D", line3DData())
	return chart
}

func scatter3DData() []opts.Chart3DData {
	data := make([]opts.Chart3DData, 0, 80)
	for i := 0; i < 80; i++ {
		data = append(data, opts.Chart3DData{Value: []interface{}{(i * 37) % 100, (i * 61) % 100, (i * 83) % 100}})
	}
	return data
}
func scatter3DBase() *charts.Scatter3D {
	chart := charts.NewScatter3D()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "basic Scatter3D example"}), charts.WithVisualMapOpts(opts.VisualMap{Calculable: opts.Bool(true), Max: 100, InRange: &opts.VisualMapInRange{Color: webGLRangeColor}}))
	chart.AddSeries("scatter3d", scatter3DData())
	return chart
}
func scatter3DDataItem() *charts.Scatter3D {
	chart := charts.NewScatter3D()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "user-defined item style"}), charts.WithXAxis3DOpts(opts.XAxis3D{Name: "MY-X-AXIS", Show: opts.Bool(true)}), charts.WithYAxis3DOpts(opts.YAxis3D{Name: "MY-Y-AXIS"}), charts.WithZAxis3DOpts(opts.ZAxis3D{Name: "MY-Z-AXIS"}))
	chart.AddSeries("scatter3d", []opts.Chart3DData{{Name: "point1", Value: []interface{}{10, 10, 10}, ItemStyle: &opts.ItemStyle{Color: "green"}}, {Name: "point2", Value: []interface{}{15, 15, 15}, ItemStyle: &opts.ItemStyle{Color: "blue"}}, {Name: "point3", Value: []interface{}{20, 20, 20}, ItemStyle: &opts.ItemStyle{Color: "red"}}})
	return chart
}

func surface3DData(scaled bool) []opts.Chart3DData {
	limit, divisor := 60, 60.0
	if scaled {
		limit, divisor = 30, 10
	}
	data := make([]opts.Chart3DData, 0, 4*limit*limit)
	for i := -limit; i < limit; i++ {
		y := float64(i) / divisor
		for j := -limit; j < limit; j++ {
			x := float64(j) / divisor
			z := math.Sin(x*math.Pi) * math.Sin(y*math.Pi)
			if scaled {
				z = math.Sin(x*x+y*y) * x / math.Pi
			}
			data = append(data, opts.Chart3DData{Value: []interface{}{x, y, z}})
		}
	}
	return data
}
func surface3D(title string, rose bool) *charts.Surface3D {
	chart := charts.NewSurface3D()
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: title}), charts.WithVisualMapOpts(opts.VisualMap{Calculable: opts.Bool(true), InRange: &opts.VisualMapInRange{Color: webGLRangeColor}, Max: 3, Min: -3}))
	chart.AddSeries("surface3d", surface3DData(rose))
	return chart
}
func surface3DBase() *charts.Surface3D { return surface3D("basic surface3D example", false) }
func surface3DRose() *charts.Surface3D { return surface3D("Rose style", true) }
