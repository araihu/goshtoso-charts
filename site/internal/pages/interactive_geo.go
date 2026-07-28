package pages

import (
	"strings"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/interactive"
)

const (
	interactiveGeoUpstreamPath       = "examples/geo.go"
	interactiveGeoUpstreamRevision   = "bda428480a82d6d77ebb9fa939cf8d52528453dd"
	interactiveGeoUpstreamSHA256     = "3a6dbe86c34e5ea478b1dea5430c10cac9f7c4905264e12fc37654f0f5d4550a"
	interactiveGeoSemanticPaintClass = "text-on-surface-strong dark:text-on-surface-dark-strong"
)

type interactiveGeoVariant struct {
	name  string
	chart interactive.Instance
}

// Upstream uses global rand.Intn(100). Fixed literal values preserve its
// [0,100) value shape without process-global nondeterminism.
var interactiveGeoNationalPoints = []interactive.GeoPoint{
	{Name: "北京", Longitude: 116.40, Latitude: 39.90, Value: 81},
	{Name: "上海", Longitude: 121.47, Latitude: 31.23, Value: 27},
	{Name: "重庆", Longitude: 106.55, Latitude: 29.56, Value: 47},
	{Name: "武汉", Longitude: 114.31, Latitude: 30.52, Value: 59},
	{Name: "台湾", Longitude: 121.30, Latitude: 25.03, Value: 18},
	{Name: "香港", Longitude: 114.17, Latitude: 22.28, Value: 63},
}

var interactiveGeoGuangdongPoints = []interactive.GeoPoint{
	{Name: "汕头", Longitude: 116.69, Latitude: 23.39, Value: 12},
	{Name: "深圳", Longitude: 114.07, Latitude: 22.62, Value: 76},
	{Name: "广州", Longitude: 113.23, Latitude: 23.16, Value: 41},
}

func sampleInteractiveGeos() []interactiveGeoVariant {
	return []interactiveGeoVariant{
		{name: "effect-scatter", chart: sampleInteractiveGeoNational()},
		{name: "scatter", chart: sampleInteractiveGeoGuangdong()},
	}
}

func sampleInteractiveGeoNational() interactive.Instance {
	points := append([]interactive.GeoPoint(nil), interactiveGeoNationalPoints...)
	points[0].Class = interactiveGeoSemanticPaintClass
	points[1].Color = "#dc2626"
	return interactive.Geo(interactive.GeoConfig{
		Label: "basic geo example", Caption: "Exact national coordinate values appear below the responsive chart.",
		Geometry:      interactive.GeoGeometryChina,
		GeometryPaint: interactive.GeoPaint{Class: interactiveGeoSemanticPaintClass},
		Series: []interactive.GeoSeries{{
			Name: "geo", Kind: interactive.GeoSeriesEffectScatter, Color: "#7c3aed",
			Points: points,
			Ripple: &interactive.RippleOptions{Period: 4, Scale: 6, BrushType: "stroke"},
		}},
		Options: interactive.ChartOptions{
			Title:   &interactive.TitleOptions{Text: "basic geo example"},
			Tooltip: &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "item"},
			Export:  &chartcontrol.ExportOptions{Filename: "basic-geo-example"},
		},
	})
}

func sampleInteractiveGeoGuangdong() interactive.Instance {
	points := append([]interactive.GeoPoint(nil), interactiveGeoGuangdongPoints...)
	points[1].Color = "#2563eb"
	points[2].Class = interactiveGeoSemanticPaintClass
	return interactive.Geo(interactive.GeoConfig{
		Label: "Guangdong province", Caption: "Exact Guangdong coordinate values appear below the responsive chart.",
		Geometry:      interactive.GeoGeometryGuangdong,
		GeometryPaint: interactive.GeoPaint{Color: "#e2e8f0"},
		VisualRange: &interactive.GeoVisualRange{
			Min: 0, Max: 100, Calculable: interactive.Bool(true),
		},
		Series: []interactive.GeoSeries{{
			Name: "geo", Kind: interactive.GeoSeriesScatter, Class: interactiveGeoSemanticPaintClass,
			Points: points,
		}},
		Options: interactive.ChartOptions{
			Title:   &interactive.TitleOptions{Text: "Guangdong province"},
			Tooltip: &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "item"},
			Export:  &chartcontrol.ExportOptions{Filename: "guangdong-province"},
		},
	})
}

func interactiveGeoCode() string {
	return `@interactive.Geo(interactive.GeoConfig{
  Label: "Guangdong province",
  Geometry: interactive.GeoGeometryGuangdong,
  VisualRange: &interactive.GeoVisualRange{
    Min: 0, Max: 100, Calculable: interactive.Bool(true),
  },
  Series: []interactive.GeoSeries{{
    Name: "geo",
    Kind: interactive.GeoSeriesScatter,
    Points: []interactive.GeoPoint{
      {Name: "汕头", Longitude: 116.69, Latitude: 23.39, Value: 12},
      {Name: "深圳", Longitude: 114.07, Latitude: 22.62, Value: 76},
      {Name: "广州", Longitude: 113.23, Latitude: 23.16, Value: 41},
    },
    Options: interactive.SeriesOptions{
      Label: &interactive.LabelOptions{Show: interactive.Bool(true)},
    },
  }},
  Width: "100%", Height: "500px",
  Options: interactive.ChartOptions{
    Export: &chartcontrol.ExportOptions{
      Filename: "` + strings.ReplaceAll(strings.ToLower("Guangdong province"), " ", "-") + `",
    },
  },
})`
}
