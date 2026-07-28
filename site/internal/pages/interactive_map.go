package pages

import (
	"strings"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/interactive"
)

const (
	interactiveMapUpstreamPath     = "examples/map.go"
	interactiveMapUpstreamRevision = "bda428480a82d6d77ebb9fa939cf8d52528453dd"
	interactiveMapUpstreamSHA256   = "3b59b5cb7ed392f3fa436d51fd420704ab2e82e439c95b226d35d12b913cf9da"
)

type interactiveMapVariant struct {
	variant interactive.MapVariant
	chart   interactive.Instance
}

var interactiveMapBaseRegions = []interactive.MapRegion{
	{Name: "北京", Value: 101},
	{Name: "上海", Value: 72},
	{Name: "广东", Value: 134},
	{Name: "辽宁", Value: 53},
	{Name: "山东", Value: 96},
	{Name: "山西", Value: 42},
	{Name: "陕西", Value: 68},
	{Name: "新疆", Value: 29},
	{Name: "内蒙古", Value: 81},
}

var interactiveMapGuangdongRegions = []interactive.MapRegion{
	{Name: "深圳市", Value: 128},
	{Name: "广州市", Value: 117},
	{Name: "湛江市", Value: 43},
	{Name: "汕头市", Value: 76},
	{Name: "东莞市", Value: 109},
	{Name: "佛山市", Value: 92},
	{Name: "云浮市", Value: 31},
	{Name: "肇庆市", Value: 55},
	{Name: "梅州市", Value: 48},
}

func sampleInteractiveMaps() []interactiveMapVariant {
	return []interactiveMapVariant{
		{interactive.MapVariantBasic, sampleInteractiveMap("basic map example", interactive.MapVariantBasic)},
		{interactive.MapVariantLabels, sampleInteractiveMap("show label", interactive.MapVariantLabels)},
		{interactive.MapVariantScale, sampleInteractiveMap("VisualMap", interactive.MapVariantScale)},
		{interactive.MapVariantRegional, sampleInteractiveMap("Guangdong province", interactive.MapVariantRegional)},
		{interactive.MapVariantTheme, sampleInteractiveMap("Map-theme", interactive.MapVariantTheme)},
	}
}

func sampleInteractiveMap(title string, variant interactive.MapVariant) interactive.Instance {
	regions := append([]interactive.MapRegion(nil), interactiveMapBaseRegions...)
	geometry := interactive.MapGeometryChina
	if variant == interactive.MapVariantRegional {
		regions = append([]interactive.MapRegion(nil), interactiveMapGuangdongRegions...)
		geometry = interactive.MapGeometryGuangdong
	}
	return interactive.Map(interactive.MapConfig{
		Label: title, Caption: "Exact region values appear below the responsive map.",
		Geometry: geometry, Variant: variant,
		Series: interactive.MapSeries{Name: "map", Regions: regions},
		Options: interactive.ChartOptions{
			Title:   &interactive.TitleOptions{Text: title},
			Tooltip: &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "item"},
			Export:  &chartcontrol.ExportOptions{Filename: strings.ReplaceAll(strings.ToLower(title), " ", "-")},
		},
	})
}

func interactiveMapCode() string {
	return `@interactive.Map(interactive.MapConfig{
  Label: "Guangdong province",
  Geometry: interactive.MapGeometryGuangdong,
  Variant: interactive.MapVariantRegional,
  Series: interactive.MapSeries{
    Name: "map",
    Regions: []interactive.MapRegion{
      {Name: "深圳市", Value: 128, Class: "major-city"},
      {Name: "广州市", Value: 117, Color: "#ff8a3d"},
      {Name: "湛江市", Value: 43},
    },
  },
  Scale: &interactive.MapScale{
    Min: 0, Max: 150, Calculable: interactive.Bool(true),
  },
  Width: "100%", Height: "500px",
})`
}
