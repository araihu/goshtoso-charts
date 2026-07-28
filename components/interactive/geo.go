package interactive

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

const maxGeoDetailRows = 100

// GeoGeometry selects a locally supplied geometry over which coordinate
// series are drawn.
type GeoGeometry string

const (
	// GeoGeometryChina contains national and provincial boundaries. It is the default.
	GeoGeometryChina GeoGeometry = ""
	// GeoGeometryGuangdong contains Guangdong prefecture-level boundaries.
	GeoGeometryGuangdong GeoGeometry = "guangdong"
)

// GeoSeriesKind selects coordinate-point behavior without changing component identity.
type GeoSeriesKind string

const (
	// GeoSeriesScatter renders ordinary coordinate points. It is the default.
	GeoSeriesScatter GeoSeriesKind = ""
	// GeoSeriesEffectScatter renders coordinate points with a ripple effect.
	GeoSeriesEffectScatter GeoSeriesKind = "effect-scatter"
)

// GeoPaint selects either an explicit color or a semantic CSS class. Color and
// Class are mutually exclusive.
type GeoPaint struct {
	Color string
	Class string
}

// GeoVisualRange maps point values to an inclusive continuous color scale.
// Empty Colors use the cold-to-warm chart-theme scale tokens.
type GeoVisualRange struct {
	Min        float64
	Max        float64
	Calculable *bool
	Colors     []string
}

// GeoPoint is one named longitude, latitude, and value tuple. Color becomes a
// private point item paint; Class is resolved to a color by the shared theme
// runtime. They are mutually exclusive point-level overrides.
type GeoPoint struct {
	Name      string
	Longitude float64
	Latitude  float64
	Value     float64
	Color     string
	Class     string
}

// GeoSeries is one named coordinate series over the selected Geometry. Color
// becomes a private series item paint; Class is resolved to a color by the
// shared theme runtime. They are mutually exclusive series-level overrides.
type GeoSeries struct {
	Name    string
	Kind    GeoSeriesKind
	Points  []GeoPoint
	Color   string
	Class   string
	Ripple  *RippleOptions
	Options SeriesOptions
}

// GeoConfig describes an accessible coordinate-series chart over registered
// geometry. Geometry resources come from dependencies.Dependencies.
type GeoConfig struct {
	Label         string
	Caption       string
	Geometry      GeoGeometry
	GeometryPaint GeoPaint
	VisualRange   *GeoVisualRange
	Series        []GeoSeries
	Width         string
	Height        string
	Options       ChartOptions
	SeriesOptions SeriesOptions
	Style         charttheme.Style
	RootAttrs     templ.Attributes
}

type rendererGeoPoint struct {
	Name        string          `json:"name"`
	Value       [3]float64      `json:"value"`
	ClassName   string          `json:"className,omitempty"`
	SourceColor string          `json:"sourceColor,omitempty"`
	ItemStyle   *opts.ItemStyle `json:"itemStyle,omitempty"`
}

type geoRuntimePaint struct {
	Color string `json:"color,omitempty"`
	Class string `json:"class,omitempty"`
}

// Geo builds a reusable coordinate-series component over national or regional geometry.
func Geo(cfg GeoConfig) Instance {
	if err := validateGeoConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveGeo, err)
	}

	style := cfg.Style
	style.Class = strings.TrimSpace("goshtoso-charts-geo " + style.Class)
	width, height := cfg.Width, cfg.Height
	if width == "" {
		width = "100%"
	}
	if height == "" {
		height = "500px"
	}

	chart := charts.NewGeo()
	global := []charts.GlobalOpts{
		charts.WithInitializationOpts(opts.Initialization{Width: width, Height: height}),
		charts.WithColorsOpts(opts.Colors(style.ResolvedColors())),
		charts.WithGeoComponentOpts(opts.GeoComponent{
			Map:       rendererGeoName(cfg.Geometry),
			ItemStyle: rendererGeoGeometryItemStyle(cfg.GeometryPaint),
		}),
	}
	global = append(global, chartGlobalOptions(cfg.Options)...)
	if cfg.VisualRange != nil {
		global = append(global, charts.WithVisualMapOpts(rendererGeoVisualRange(cfg.VisualRange)))
	}
	chart.SetGlobalOptions(global...)

	runtimeSeries := make([]geoRuntimePaint, len(cfg.Series))
	for seriesIndex, series := range cfg.Series {
		options := append([]charts.SeriesOpts(nil), chartSeriesOptions(cfg.SeriesOptions)...)
		if series.Ripple != nil {
			options = append(options, charts.WithRippleEffectOpts(opts.RippleEffect{
				Period: float32(series.Ripple.Period), Scale: float32(series.Ripple.Scale), BrushType: series.Ripple.BrushType,
			}))
		}
		options = append(options, chartSeriesOptions(series.Options)...)
		chart.AddSeries(series.Name, rendererGeoSeriesKind(series.Kind), make([]opts.GeoData, len(series.Points)), options...)
		rendered := make([]rendererGeoPoint, len(series.Points))
		for pointIndex, point := range series.Points {
			rendered[pointIndex] = rendererGeoPoint{
				Name: point.Name, Value: [3]float64{point.Longitude, point.Latitude, point.Value},
				ClassName: point.Class, SourceColor: point.Color,
			}
			if point.Color != "" {
				rendered[pointIndex].ItemStyle = &opts.ItemStyle{Color: point.Color}
			}
		}
		chart.MultiSeries[seriesIndex].Data = rendered
		runtimeSeries[seriesIndex] = geoRuntimePaint{Color: series.Color, Class: series.Class}
	}
	geometryJSON, _ := json.Marshal(geoRuntimePaint{Color: cfg.GeometryPaint.Color, Class: cfg.GeometryPaint.Class})
	seriesJSON, _ := json.Marshal(runtimeSeries)

	return newInstance(chartcomponents.KindInteractiveGeo, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: style, ResponsiveWidth: responsiveWidth(cfg.Width),
		Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		RootAttrs: cfg.RootAttrs, Details: geoExactValues(geoDetailRows(cfg.Series, maxGeoDetailRows)),
		ExplicitVisualMapColors: geoVisualRangeHasColors(cfg.VisualRange),
		GeoGeometryPaint:        string(geometryJSON), GeoSeriesPaints: string(seriesJSON),
	})
}

func rendererGeoName(geometry GeoGeometry) string {
	if geometry == GeoGeometryGuangdong {
		return "广东"
	}
	return "china"
}

func rendererGeoSeriesKind(kind GeoSeriesKind) string {
	if kind == GeoSeriesEffectScatter {
		return types.ChartEffectScatter
	}
	return types.ChartScatter
}

func rendererGeoGeometryItemStyle(paint GeoPaint) *opts.ItemStyle {
	if paint.Color == "" {
		return nil
	}
	return &opts.ItemStyle{AreaColor: paint.Color}
}

func rendererGeoVisualRange(value *GeoVisualRange) opts.VisualMap {
	result := opts.VisualMap{Min: float32(value.Min), Max: float32(value.Max)}
	if value.Calculable != nil {
		result.Calculable = opts.Bool(*value.Calculable)
	}
	if len(value.Colors) > 0 {
		result.InRange = &opts.VisualMapInRange{Color: append([]string(nil), value.Colors...)}
	}
	return result
}

func geoVisualRangeHasColors(value *GeoVisualRange) bool {
	return value != nil && len(value.Colors) > 0
}

func validateGeoConfig(cfg GeoConfig) error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("geo chart label is required")
	}
	if cfg.Geometry != GeoGeometryChina && cfg.Geometry != GeoGeometryGuangdong {
		return fmt.Errorf("geo chart geometry %q is not supported", cfg.Geometry)
	}
	if cfg.Options.Legend != nil {
		return fmt.Errorf("geo chart legend is not supported")
	}
	if cfg.Options.XAxis != nil || cfg.Options.YAxis != nil {
		return fmt.Errorf("geo chart Cartesian axes are not supported")
	}
	if tooltip := cfg.Options.Tooltip; tooltip != nil && tooltip.Trigger != "" && tooltip.Trigger != "item" {
		return fmt.Errorf("geo chart tooltip trigger %q is not supported", tooltip.Trigger)
	}
	if err := validateGeoPaint("geometry", cfg.GeometryPaint.Color, cfg.GeometryPaint.Class); err != nil {
		return err
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("geo chart root attribute %q is reserved", attribute)
			}
		}
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("geo chart series are required")
	}
	for seriesIndex, series := range cfg.Series {
		name := strings.TrimSpace(series.Name)
		if name == "" {
			return fmt.Errorf("geo chart series %d name is required", seriesIndex)
		}
		if series.Kind != GeoSeriesScatter && series.Kind != GeoSeriesEffectScatter {
			return fmt.Errorf("geo chart series %q kind %q is not supported", name, series.Kind)
		}
		if len(series.Points) == 0 {
			return fmt.Errorf("geo chart series %q points are required", name)
		}
		if err := validateGeoPaint("series "+fmt.Sprintf("%q", name), series.Color, series.Class); err != nil {
			return err
		}
		if series.Options.ItemStyle != nil && strings.TrimSpace(series.Options.ItemStyle.Color) != "" &&
			(strings.TrimSpace(series.Color) != "" || strings.TrimSpace(series.Class) != "") {
			return fmt.Errorf("geo chart series %q paint and item-style color are mutually exclusive", name)
		}
		if series.Ripple != nil {
			if series.Kind != GeoSeriesEffectScatter {
				return fmt.Errorf("geo chart series %q ripple requires effect scatter", name)
			}
			if !finiteNumber(series.Ripple.Period) || series.Ripple.Period <= 0 {
				return fmt.Errorf("geo chart series %q ripple period must be positive and finite", name)
			}
			if !finiteNumber(series.Ripple.Scale) || series.Ripple.Scale <= 0 {
				return fmt.Errorf("geo chart series %q ripple scale must be positive and finite", name)
			}
			if series.Ripple.BrushType != "" && series.Ripple.BrushType != "fill" && series.Ripple.BrushType != "stroke" {
				return fmt.Errorf("geo chart series %q ripple brush type %q is not supported", name, series.Ripple.BrushType)
			}
		}
		pointNames := make(map[string]bool, len(series.Points))
		for pointIndex, point := range series.Points {
			pointName := strings.TrimSpace(point.Name)
			if pointName == "" {
				return fmt.Errorf("geo chart series %q point %d name is required", name, pointIndex)
			}
			if pointNames[pointName] {
				return fmt.Errorf("geo chart series %q point %q is duplicated", name, pointName)
			}
			pointNames[pointName] = true
			if !finiteNumber(point.Longitude) || point.Longitude < -180 || point.Longitude > 180 {
				return fmt.Errorf("geo chart point %q longitude must be finite and within [-180, 180]", pointName)
			}
			if !finiteNumber(point.Latitude) || point.Latitude < -90 || point.Latitude > 90 {
				return fmt.Errorf("geo chart point %q latitude must be finite and within [-90, 90]", pointName)
			}
			if !finiteNumber(point.Value) {
				return fmt.Errorf("geo chart point %q value must be finite", pointName)
			}
			if err := validateGeoPaint("point "+fmt.Sprintf("%q", pointName), point.Color, point.Class); err != nil {
				return err
			}
		}
	}
	if value := cfg.VisualRange; value != nil {
		if !finiteNumber(value.Min) || !finiteNumber(value.Max) {
			return fmt.Errorf("geo chart visual range bounds must be finite")
		}
		if value.Min >= value.Max {
			return fmt.Errorf("geo chart visual range minimum must be less than maximum")
		}
		if len(value.Colors) == 1 {
			return fmt.Errorf("geo chart visual range colors require at least two values")
		}
		for index, color := range value.Colors {
			if strings.TrimSpace(color) == "" {
				return fmt.Errorf("geo chart visual range color %d is required", index)
			}
		}
	}
	return validateChartOptions(cfg.Options)
}

func validateGeoPaint(name, color, class string) error {
	if strings.TrimSpace(color) != "" && strings.TrimSpace(class) != "" {
		return fmt.Errorf("geo chart %s color and class are mutually exclusive", name)
	}
	return nil
}

type geoValueRow struct {
	Series, Kind, Point, Longitude, Latitude, Value, Class string
}
type geoValueRows struct {
	Rows    []geoValueRow
	Omitted int
}

func geoDetailRows(series []GeoSeries, limit int) geoValueRows {
	total := 0
	rows := make([]geoValueRow, 0, limit)
	for _, current := range series {
		total += len(current.Points)
		for _, point := range current.Points {
			if len(rows) == limit {
				continue
			}
			rows = append(rows, geoValueRow{
				Series: current.Name, Kind: string(current.Kind), Point: point.Name,
				Longitude: fmt.Sprintf("%.2f", point.Longitude), Latitude: fmt.Sprintf("%.2f", point.Latitude),
				Value: fmt.Sprintf("%g", point.Value), Class: point.Class,
			})
		}
	}
	return geoValueRows{Rows: rows, Omitted: total - len(rows)}
}
