package interactive

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

const maxMapDetailRows = 100

// MapGeometry selects a locally supplied geographic resource.
type MapGeometry string

const (
	// MapGeometryBrazil contains all 26 states and the Federal District. It is the default.
	MapGeometryBrazil MapGeometry = ""
)

// MapVariant selects a documented map behavior while retaining one component identity.
type MapVariant string

const (
	MapVariantBasic    MapVariant = ""
	MapVariantLabels   MapVariant = "labels"
	MapVariantScale    MapVariant = "scale"
	MapVariantRegional MapVariant = "regional"
	MapVariantTheme    MapVariant = "theme"
)

// MapScale maps region values to an inclusive continuous color scale.
// Empty Colors use chart theme scale tokens. Calculable defaults false except
// for Scale, Regional, and Theme variants.
type MapScale struct {
	Min        float64
	Max        float64
	Calculable *bool
	Colors     []string
}

// MapRegion is one named geographic value. Class provides color-independent
// semantics; Color optionally overrides its theme or scale color.
type MapRegion struct {
	Name  string
	Code  string
	Value float64
	Class string
	Color string
}

// MapSeries is one named set of values aligned to Geometry region names.
type MapSeries struct {
	Name    string
	Regions []MapRegion
}

// MapConfig describes an accessible interactive geographic map. Geometry
// resources are supplied by dependencies.Dependencies; local vendored delivery
// is the default and CDN delivery remains explicit.
type MapConfig struct {
	Label      string
	Caption    string
	Geometry   MapGeometry
	Variant    MapVariant
	Series     MapSeries
	ShowLabels *bool
	Scale      *MapScale
	Width      string
	Height     string
	Options    ChartOptions
	Style      charttheme.Style
	RootAttrs  templ.Attributes
}

type rendererMapRegion struct {
	Name        string          `json:"name"`
	Code        string          `json:"code"`
	Value       float64         `json:"value"`
	ClassName   string          `json:"className,omitempty"`
	SourceColor string          `json:"sourceColor,omitempty"`
	ItemStyle   *opts.ItemStyle `json:"itemStyle,omitempty"`
}

// Map builds one reusable component for national, regional, label, scale, and themed variants.
func Map(cfg MapConfig) Instance {
	if err := validateMapConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveMap, err)
	}

	style := cfg.Style
	style.Class = strings.TrimSpace("goshtoso-charts-map " + style.Class)
	width, height := cfg.Width, cfg.Height
	if width == "" {
		width = "100%"
	}
	if height == "" {
		height = "500px"
	}

	chart := charts.NewMap()
	chart.RegisterMapType(rendererMapName(cfg.Geometry))
	global := []charts.GlobalOpts{
		charts.WithInitializationOpts(opts.Initialization{Width: width, Height: height}),
		charts.WithColorsOpts(opts.Colors(style.ResolvedColors())),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
	}
	global = append(global, chartGlobalOptions(cfg.Options)...)
	if scale := resolvedMapScale(cfg); scale != nil {
		global = append(global, charts.WithVisualMapOpts(rendererMapScale(scale)))
	}
	chart.SetGlobalOptions(global...)

	showLabels := cfg.ShowLabels
	if showLabels == nil && cfg.Variant == MapVariantLabels {
		showLabels = Bool(true)
	}
	seriesOptions := make([]charts.SeriesOpts, 0, 1)
	if showLabels != nil {
		seriesOptions = append(seriesOptions, charts.WithLabelOpts(opts.Label{Show: opts.Bool(*showLabels), Formatter: opts.FuncOpts("function(params){return params.data.code;}"), FontSize: 11}))
	}
	chart.AddSeries(cfg.Series.Name, make([]opts.MapData, len(cfg.Series.Regions)), seriesOptions...)
	rendered := make([]rendererMapRegion, len(cfg.Series.Regions))
	for index, region := range cfg.Series.Regions {
		rendered[index] = rendererMapRegion{Name: region.Name, Code: region.Code, Value: region.Value, ClassName: region.Class, SourceColor: region.Color}
		if region.Color != "" {
			rendered[index].ItemStyle = &opts.ItemStyle{Color: region.Color}
		}
	}
	chart.MultiSeries[0].Data = rendered

	return newInstance(chartcomponents.KindInteractiveMap, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: style, ResponsiveWidth: responsiveWidth(cfg.Width),
		Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		RootAttrs: cfg.RootAttrs, Details: mapExactValues(mapDetailRows(cfg.Series.Regions, maxMapDetailRows)),
		ExplicitVisualMapColors: mapScaleHasColors(resolvedMapScale(cfg)),
		ScriptReplacements: geographicLayoutReplacements(
			rendererMapName(cfg.Geometry), hasChartTitle(cfg.Options), resolvedMapScale(cfg) != nil,
		),
	})
}

func rendererMapName(geometry MapGeometry) string {
	return "brazil"
}

func resolvedMapScale(cfg MapConfig) *MapScale {
	if cfg.Scale != nil {
		copy := *cfg.Scale
		copy.Colors = append([]string(nil), cfg.Scale.Colors...)
		return &copy
	}
	switch cfg.Variant {
	case MapVariantScale:
		return &MapScale{Calculable: Bool(true)}
	case MapVariantRegional:
		return &MapScale{Calculable: Bool(true), Colors: []string{"#50a3ba", "#eac736", "#d94e5d"}}
	case MapVariantTheme:
		return &MapScale{Max: 150, Calculable: Bool(true)}
	default:
		return nil
	}
}

func rendererMapScale(scale *MapScale) opts.VisualMap {
	result := opts.VisualMap{
		Min: float32(scale.Min), Max: float32(scale.Max),
		Orient: "horizontal", Left: "center", Bottom: "8",
	}
	if scale.Calculable != nil {
		result.Calculable = opts.Bool(*scale.Calculable)
	}
	if len(scale.Colors) > 0 {
		result.InRange = &opts.VisualMapInRange{Color: append([]string(nil), scale.Colors...)}
	}
	return result
}

func mapScaleHasColors(scale *MapScale) bool { return scale != nil && len(scale.Colors) > 0 }

func validateMapConfig(cfg MapConfig) error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("map chart label is required")
	}
	if strings.TrimSpace(cfg.Series.Name) == "" {
		return fmt.Errorf("map chart series name is required")
	}
	if len(cfg.Series.Regions) == 0 {
		return fmt.Errorf("map chart regions are required")
	}
	if cfg.Geometry != MapGeometryBrazil {
		return fmt.Errorf("map chart geometry %q is not supported", cfg.Geometry)
	}
	validVariants := map[MapVariant]bool{MapVariantBasic: true, MapVariantLabels: true, MapVariantScale: true, MapVariantRegional: true, MapVariantTheme: true}
	if !validVariants[cfg.Variant] {
		return fmt.Errorf("map chart variant %q is not supported", cfg.Variant)
	}
	if cfg.Options.Legend != nil {
		return fmt.Errorf("map chart legend is not supported")
	}
	if cfg.Options.XAxis != nil || cfg.Options.YAxis != nil {
		return fmt.Errorf("map chart Cartesian axes are not supported")
	}
	if tooltip := cfg.Options.Tooltip; tooltip != nil && tooltip.Trigger != "" && tooltip.Trigger != "item" {
		return fmt.Errorf("map chart tooltip trigger %q is not supported", tooltip.Trigger)
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("map chart root attribute %q is reserved", attribute)
			}
		}
	}
	names := make(map[string]bool, len(cfg.Series.Regions))
	for index, region := range cfg.Series.Regions {
		name := strings.TrimSpace(region.Name)
		if name == "" {
			return fmt.Errorf("map chart region %d name is required", index)
		}
		if names[name] {
			return fmt.Errorf("map chart region %q is duplicated", name)
		}
		names[name] = true
		wantCode, known := brazilStateCodes[name]
		if !known {
			return fmt.Errorf("map chart region %q is not in Brazil geometry", name)
		}
		if region.Code != wantCode {
			return fmt.Errorf("map chart region %q code is %q, want %q", name, region.Code, wantCode)
		}
		if !finiteNumber(region.Value) {
			return fmt.Errorf("map chart region %q value must be finite", name)
		}
	}
	if len(names) != len(brazilStateCodes) {
		return fmt.Errorf("map chart Brazil geometry requires all 26 states and the Federal District; got %d of %d regions", len(names), len(brazilStateCodes))
	}
	if scale := cfg.Scale; scale != nil {
		if !finiteNumber(scale.Min) || !finiteNumber(scale.Max) {
			return fmt.Errorf("map chart scale bounds must be finite")
		}
		if scale.Min > scale.Max {
			return fmt.Errorf("map chart scale minimum must not exceed maximum")
		}
		if len(scale.Colors) == 1 {
			return fmt.Errorf("map chart scale colors require at least two values")
		}
		for index, color := range scale.Colors {
			if strings.TrimSpace(color) == "" {
				return fmt.Errorf("map chart scale color %d is required", index)
			}
		}
	}
	return validateChartOptions(cfg.Options)
}

type mapValueRow struct{ Region, Code, Value, Class string }
type mapValueRows struct {
	Rows    []mapValueRow
	Omitted int
}

func mapDetailRows(regions []MapRegion, limit int) mapValueRows {
	rows := make([]mapValueRow, 0, min(len(regions), limit))
	for _, region := range regions {
		if len(rows) == limit {
			return mapValueRows{Rows: rows, Omitted: len(regions) - len(rows)}
		}
		rows = append(rows, mapValueRow{Region: region.Name, Code: region.Code, Value: fmt.Sprintf("%g", region.Value), Class: region.Class})
	}
	return mapValueRows{Rows: rows}
}

var brazilStateCodes = map[string]string{
	"Rondônia": "RO", "Acre": "AC", "Amazonas": "AM", "Roraima": "RR", "Pará": "PA", "Amapá": "AP", "Tocantins": "TO",
	"Maranhão": "MA", "Piauí": "PI", "Ceará": "CE", "Rio Grande do Norte": "RN", "Paraíba": "PB", "Pernambuco": "PE", "Alagoas": "AL", "Sergipe": "SE", "Bahia": "BA",
	"Minas Gerais": "MG", "Espírito Santo": "ES", "Rio de Janeiro": "RJ", "São Paulo": "SP", "Paraná": "PR", "Santa Catarina": "SC", "Rio Grande do Sul": "RS",
	"Mato Grosso do Sul": "MS", "Mato Grosso": "MT", "Goiás": "GO", "Distrito Federal": "DF",
}
