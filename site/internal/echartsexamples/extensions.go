package echartsexamples

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/render"
	"github.com/go-echarts/go-echarts/v2/types"
)

const extensionExamplesSource = "https://github.com/go-echarts/examples/blob/master/examples/"

// ExtensionExamples ports liquid, map, geo, and wordcloud upstream examples.
// Map and geo renderers expect their matching ECharts extension assets to be
// registered by the demo host.
var ExtensionExamples = []Example{
	{Slug: "liquid-basic", Title: "Basic liquid", Group: "Extensions", Source: extensionExamplesSource + "liquid.go", Build: func() render.Renderer { return liquidBase() }},
	{Slug: "liquid-label", Title: "Liquid labels", Group: "Extensions", Source: extensionExamplesSource + "liquid.go", Build: func() render.Renderer { return liquidShowLabel() }},
	{Slug: "liquid-outline", Title: "Liquid outline", Group: "Extensions", Source: extensionExamplesSource + "liquid.go", Build: func() render.Renderer { return liquidOutline() }},
	{Slug: "liquid-wave-animation", Title: "Liquid wave animation", Group: "Extensions", Source: extensionExamplesSource + "liquid.go", Build: func() render.Renderer { return liquidWaveAnimation() }},
	{Slug: "liquid-diamond", Title: "Liquid diamond", Group: "Extensions", Source: extensionExamplesSource + "liquid.go", Build: func() render.Renderer { return liquidDiamond() }},
	{Slug: "liquid-pin", Title: "Liquid pin", Group: "Extensions", Source: extensionExamplesSource + "liquid.go", Build: func() render.Renderer { return liquidPin() }},
	{Slug: "liquid-arrow", Title: "Liquid arrow", Group: "Extensions", Source: extensionExamplesSource + "liquid.go", Build: func() render.Renderer { return liquidArrow() }},
	{Slug: "liquid-triangle", Title: "Liquid triangle", Group: "Extensions", Source: extensionExamplesSource + "liquid.go", Build: func() render.Renderer { return liquidTriangle() }},
	{Slug: "map-basic", Title: "Basic map", Group: "Extensions", Source: extensionExamplesSource + "map.go", Build: func() render.Renderer { return mapBase() }},
	{Slug: "map-label", Title: "Map labels", Group: "Extensions", Source: extensionExamplesSource + "map.go", Build: func() render.Renderer { return mapShowLabel() }},
	{Slug: "map-visual-map", Title: "Map visual map", Group: "Extensions", Source: extensionExamplesSource + "map.go", Build: func() render.Renderer { return mapVisualMap() }},
	{Slug: "map-guangdong", Title: "Guangdong map", Group: "Extensions", Source: extensionExamplesSource + "map.go", Build: func() render.Renderer { return mapRegion() }},
	{Slug: "map-theme", Title: "Map theme", Group: "Extensions", Source: extensionExamplesSource + "map.go", Build: func() render.Renderer { return mapTheme() }},
	{Slug: "geo-basic", Title: "Basic geo", Group: "Extensions", Source: extensionExamplesSource + "geo.go", Build: func() render.Renderer { return geoBase() }},
	{Slug: "geo-guangdong", Title: "Guangdong geo", Group: "Extensions", Source: extensionExamplesSource + "geo.go", Build: func() render.Renderer { return geoGuangdong() }},
	{Slug: "wordcloud-basic", Title: "Basic word cloud", Group: "Extensions", Source: extensionExamplesSource + "wordcloud.go", Build: func() render.Renderer { return wcBase() }},
	{Slug: "wordcloud-cardioid", Title: "Cardioid word cloud", Group: "Extensions", Source: extensionExamplesSource + "wordcloud.go", Build: func() render.Renderer { return wcCardioid() }},
	{Slug: "wordcloud-star", Title: "Star word cloud", Group: "Extensions", Source: extensionExamplesSource + "wordcloud.go", Build: func() render.Renderer { return wcStar() }},
}

func liquidData() []opts.LiquidData {
	return []opts.LiquidData{{Value: 0.3}, {Value: 0.4}, {Value: 0.5}}
}

func titledLiquid(title string) *charts.Liquid {
	liquid := charts.NewLiquid()
	liquid.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: title}))
	liquid.AddSeries("liquid", liquidData())
	return liquid
}

func liquidBase() *charts.Liquid {
	liquid := titledLiquid("basic liquid example")
	liquid.SetSeriesOptions(charts.WithLiquidChartOpts(opts.LiquidChart{IsWaveAnimation: opts.Bool(true)}))
	return liquid
}

func liquidShowLabel() *charts.Liquid {
	liquid := titledLiquid("show label")
	liquid.SetSeriesOptions(charts.WithLabelOpts(opts.Label{Show: opts.Bool(true)}), charts.WithLiquidChartOpts(opts.LiquidChart{IsWaveAnimation: opts.Bool(true)}))
	return liquid
}

func liquidOutline() *charts.Liquid {
	liquid := titledLiquid("show outline")
	liquid.SetSeriesOptions(charts.WithLabelOpts(opts.Label{Show: opts.Bool(true)}), charts.WithLiquidChartOpts(opts.LiquidChart{IsWaveAnimation: opts.Bool(true), IsShowOutline: opts.Bool(true)}))
	return liquid
}

func liquidWaveAnimation() *charts.Liquid {
	liquid := titledLiquid("disable wave animation")
	liquid.SetSeriesOptions(charts.WithLabelOpts(opts.Label{Show: opts.Bool(true)}), charts.WithLiquidChartOpts(opts.LiquidChart{IsWaveAnimation: opts.Bool(true), IsShowOutline: opts.Bool(true)}))
	return liquid
}

func liquidWithShape(title, shape string) *charts.Liquid {
	liquid := titledLiquid(title)
	liquid.SetSeriesOptions(charts.WithLiquidChartOpts(opts.LiquidChart{IsWaveAnimation: opts.Bool(true), Shape: shape}))
	return liquid
}

func liquidDiamond() *charts.Liquid  { return liquidWithShape("shape(Diamond)", "diamond") }
func liquidPin() *charts.Liquid      { return liquidWithShape("shape(Pin)", "pin") }
func liquidArrow() *charts.Liquid    { return liquidWithShape("shape(Arrow)", "arrow") }
func liquidTriangle() *charts.Liquid { return liquidWithShape("shape(Triangle)", "triangle") }

var chinaMapData = []opts.MapData{
	{Name: "北京", Value: 121}, {Name: "上海", Value: 137}, {Name: "广东", Value: 98},
	{Name: "辽宁", Value: 74}, {Name: "山东", Value: 113}, {Name: "山西", Value: 62},
	{Name: "陕西", Value: 86}, {Name: "新疆", Value: 47}, {Name: "内蒙古", Value: 54},
}

var guangdongMapData = []opts.MapData{
	{Name: "深圳市", Value: 142}, {Name: "广州市", Value: 136}, {Name: "湛江市", Value: 71},
	{Name: "汕头市", Value: 88}, {Name: "东莞市", Value: 119}, {Name: "佛山市", Value: 105},
	{Name: "云浮市", Value: 42}, {Name: "肇庆市", Value: 65}, {Name: "梅州市", Value: 57},
}

func titledMap(title, mapType string) *charts.Map {
	chart := charts.NewMap()
	chart.RegisterMapType(mapType)
	chart.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: title}))
	return chart
}

func mapBase() *charts.Map {
	chart := titledMap("basic map example", "china")
	chart.AddSeries("map", chinaMapData)
	return chart
}

func mapShowLabel() *charts.Map {
	chart := titledMap("show label", "china")
	chart.AddSeries("map", chinaMapData).SetSeriesOptions(charts.WithLabelOpts(opts.Label{Show: opts.Bool(true)}))
	return chart
}

func mapVisualMap() *charts.Map {
	chart := titledMap("VisualMap", "china")
	chart.SetGlobalOptions(charts.WithVisualMapOpts(opts.VisualMap{Calculable: opts.Bool(true)}))
	chart.AddSeries("map", chinaMapData)
	return chart
}

func mapRegion() *charts.Map {
	chart := titledMap("Guangdong province", "广东")
	chart.SetGlobalOptions(charts.WithVisualMapOpts(opts.VisualMap{Calculable: opts.Bool(true), InRange: &opts.VisualMapInRange{Color: []string{"#50a3ba", "#eac736", "#d94e5d"}}}))
	chart.AddSeries("map", guangdongMapData)
	return chart
}

func mapTheme() *charts.Map {
	chart := titledMap("Map-theme", "china")
	chart.SetGlobalOptions(charts.WithInitializationOpts(opts.Initialization{Theme: "macarons"}), charts.WithVisualMapOpts(opts.VisualMap{Calculable: opts.Bool(true), Max: 150}))
	chart.AddSeries("map", chinaMapData)
	return chart
}

var chinaGeoData = []opts.GeoData{
	{Name: "北京", Value: []float64{116.40, 39.90, 82}}, {Name: "上海", Value: []float64{121.47, 31.23, 94}},
	{Name: "重庆", Value: []float64{106.55, 29.56, 67}}, {Name: "武汉", Value: []float64{114.31, 30.52, 76}},
	{Name: "台湾", Value: []float64{121.30, 25.03, 58}}, {Name: "香港", Value: []float64{114.17, 22.28, 89}},
}

var guangdongGeoData = []opts.GeoData{
	{Name: "汕头", Value: []float64{116.69, 23.39, 63}}, {Name: "深圳", Value: []float64{114.07, 22.62, 91}},
	{Name: "广州", Value: []float64{113.23, 23.16, 87}},
}

func geoBase() *charts.Geo {
	geo := charts.NewGeo()
	geo.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "basic geo example"}), charts.WithGeoComponentOpts(opts.GeoComponent{Map: "china", ItemStyle: &opts.ItemStyle{Color: "#006666"}}))
	geo.AddSeries("geo", types.ChartEffectScatter, chinaGeoData, charts.WithRippleEffectOpts(opts.RippleEffect{Period: 4, Scale: 6, BrushType: "stroke"}))
	return geo
}

func geoGuangdong() *charts.Geo {
	geo := charts.NewGeo()
	geo.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Guangdong province"}), charts.WithGeoComponentOpts(opts.GeoComponent{Map: "广东"}), charts.WithVisualMapOpts(opts.VisualMap{Calculable: opts.Bool(true), InRange: &opts.VisualMapInRange{Color: []string{"#50a3ba", "#eac736", "#d94e5d"}}}))
	geo.AddSeries("geo", types.ChartScatter, guangdongGeoData)
	return geo
}

var wordCloudData = []opts.WordCloudData{
	{Name: "Sam S Club", Value: 10000}, {Name: "Macys", Value: 6181}, {Name: "Amy Schumer", Value: 4386},
	{Name: "Jurassic World", Value: 4055}, {Name: "Charter Communications", Value: 2467}, {Name: "Chick Fil A", Value: 2244},
	{Name: "Planet Fitness", Value: 1898}, {Name: "Pitch Perfect", Value: 1484}, {Name: "Express", Value: 1689},
	{Name: "Home", Value: 1112}, {Name: "Johnny Depp", Value: 985}, {Name: "Lena Dunham", Value: 847},
	{Name: "Lewis Hamilton", Value: 582}, {Name: "KXAN", Value: 555}, {Name: "Mary Ellen Mark", Value: 550},
	{Name: "Farrah Abraham", Value: 462}, {Name: "Rita Ora", Value: 366}, {Name: "Serena Williams", Value: 282},
	{Name: "NCAA baseball tournament", Value: 273}, {Name: "Point Break", Value: 265},
}

func wordCloudWithShape(title, shape string) *charts.WordCloud {
	wordCloud := charts.NewWordCloud()
	wordCloud.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: title}))
	wordCloud.AddSeries("wordcloud", wordCloudData).SetSeriesOptions(charts.WithWorldCloudChartOpts(opts.WordCloudChart{SizeRange: []float32{14, 80}, Shape: shape}))
	return wordCloud
}

func wcBase() *charts.WordCloud     { return wordCloudWithShape("basic WordCloud example", "") }
func wcCardioid() *charts.WordCloud { return wordCloudWithShape("cardioid shape", "cardioid") }
func wcStar() *charts.WordCloud     { return wordCloudWithShape("star shape", "star") }
