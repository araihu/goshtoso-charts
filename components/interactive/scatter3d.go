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

// Scatter3DPalette identifies a renderer-neutral visual-range palette.
type Scatter3DPalette string

const (
	// Scatter3DPaletteCustom uses VisualRange.Colors.
	Scatter3DPaletteCustom Scatter3DPalette = ""
	// Scatter3DPaletteColdToWarm preserves the pinned source's ten-color diverging intent.
	Scatter3DPaletteColdToWarm Scatter3DPalette = "cold-to-warm"
)

var scatter3DColdToWarmColors = []string{
	"#313695", "#4575b4", "#74add1", "#abd9e9", "#e0f3f8",
	"#fee090", "#fdae61", "#f46d43", "#d73027", "#a50026",
}

// Point3D is one finite point in three-dimensional Cartesian space.
// Value optionally supplies a separate finite scalar for visual-range mapping.
type Point3D struct {
	Name       string
	X          float64
	Y          float64
	Z          float64
	Value      *float64
	Symbol     string
	SymbolSize int
	Color      string
	Class      string
}

// Scatter3DAxis configures one named numeric axis.
type Scatter3DAxis struct {
	Name string
	Show *bool
	Min  *float64
	Max  *float64
}

// Scatter3DAxes configures all three axes as one validated unit.
type Scatter3DAxes struct {
	X Scatter3DAxis
	Y Scatter3DAxis
	Z Scatter3DAxis
}

// Scatter3DVisualRange maps point values to a continuous color scale.
type Scatter3DVisualRange struct {
	Min        float64
	Max        float64
	Calculable *bool
	Palette    Scatter3DPalette
	Colors     []string
}

// Scatter3DSeriesOptions configures renderer-neutral symbol and paint choices.
type Scatter3DSeriesOptions struct {
	Symbol     string
	SymbolSize int
	Color      string
	Class      string
}

// Scatter3DSeries contains one named point collection.
type Scatter3DSeries struct {
	Name    string
	Points  []Point3D
	Options Scatter3DSeriesOptions
}

// Scatter3DConfig describes an accessible browser-rendered 3D scatter chart.
type Scatter3DConfig struct {
	Label       string
	Caption     string
	Series      []Scatter3DSeries
	Axes        *Scatter3DAxes
	VisualRange *Scatter3DVisualRange
	Width       string
	Height      string
	Options     ChartOptions
	Style       charttheme.Style
	RootAttrs   templ.Attributes
}

type rendererPoint3D struct {
	Name        string          `json:"name"`
	Value       []float64       `json:"value"`
	Symbol      string          `json:"symbol,omitempty"`
	SymbolSize  int             `json:"symbolSize,omitempty"`
	ClassName   string          `json:"className,omitempty"`
	SourceColor string          `json:"sourceColor,omitempty"`
	ItemStyle   *opts.ItemStyle `json:"itemStyle,omitempty"`
}

type scatter3DPaint struct {
	Color string `json:"color,omitempty"`
	Class string `json:"class,omitempty"`
}

// Scatter3D builds a reusable renderer-neutral three-dimensional scatter component.
func Scatter3D(cfg Scatter3DConfig) Instance {
	if err := validateScatter3DConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveScatter3D, err)
	}

	style := cfg.Style
	style.Class = strings.TrimSpace("goshtoso-charts-scatter-3d " + style.Class)
	width, height := cfg.Width, cfg.Height
	if width == "" {
		width = "100%"
	}
	if height == "" {
		height = "38rem"
	}

	chart := charts.NewScatter3D()
	global := []charts.GlobalOpts{
		charts.WithInitializationOpts(opts.Initialization{Width: width, Height: height}),
		charts.WithColorsOpts(opts.Colors(style.ResolvedColors())),
	}
	global = append(global, scatter3DChartOptions(cfg.Options)...)
	if cfg.Axes != nil {
		global = append(global,
			charts.WithXAxis3DOpts(rendererScatter3DAxisX(cfg.Axes.X)),
			charts.WithYAxis3DOpts(rendererScatter3DAxisY(cfg.Axes.Y)),
			charts.WithZAxis3DOpts(rendererScatter3DAxisZ(cfg.Axes.Z)),
		)
	}
	if cfg.VisualRange != nil {
		global = append(global, charts.WithVisualMapOpts(rendererScatter3DVisualRange(cfg.VisualRange)))
	}
	chart.SetGlobalOptions(global...)

	paints := make([]scatter3DPaint, len(cfg.Series))
	for seriesIndex, series := range cfg.Series {
		data := make([]opts.Chart3DData, len(series.Points))
		chart.AddSeries(series.Name, data, scatter3DSeriesRendererOptions(series.Options)...)
		rendered := make([]rendererPoint3D, len(series.Points))
		for pointIndex, point := range series.Points {
			values := []float64{point.X, point.Y, point.Z}
			if point.Value != nil {
				values = append(values, *point.Value)
			}
			rendered[pointIndex] = rendererPoint3D{
				Name: point.Name, Value: values, Symbol: point.Symbol, SymbolSize: point.SymbolSize,
				ClassName: point.Class, SourceColor: point.Color,
			}
			if point.Color != "" {
				rendered[pointIndex].ItemStyle = &opts.ItemStyle{Color: point.Color}
			}
		}
		chart.MultiSeries[seriesIndex].Data = rendered
		paints[seriesIndex] = scatter3DPaint{Color: series.Options.Color, Class: series.Options.Class}
	}
	paintJSON, _ := json.Marshal(paints)

	return newInstance(chartcomponents.KindInteractiveScatter3D, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: style, ResponsiveWidth: responsiveWidth(cfg.Width),
		Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		RootAttrs: cfg.RootAttrs, Details: scatter3DExactPoints(cfg.Series),
		ExplicitVisualMapColors: cfg.VisualRange != nil && cfg.VisualRange.Palette != Scatter3DPaletteColdToWarm,
		Scatter3DPaints:         string(paintJSON), Scatter3DColdToWarm: cfg.VisualRange != nil && cfg.VisualRange.Palette == Scatter3DPaletteColdToWarm,
	})
}

func scatter3DChartOptions(options ChartOptions) []charts.GlobalOpts {
	copy := options
	copy.Legend = nil
	copy.XAxis = nil
	copy.YAxis = nil
	return chartGlobalOptions(copy)
}

func scatter3DSeriesRendererOptions(options Scatter3DSeriesOptions) []charts.SeriesOpts {
	result := chartSeriesOptions(SeriesOptions{Symbol: options.Symbol, SymbolSize: options.SymbolSize})
	if options.Color != "" {
		result = append(result, charts.WithItemStyleOpts(opts.ItemStyle{Color: options.Color}))
	}
	return result
}

func rendererScatter3DAxisX(axis Scatter3DAxis) opts.XAxis3D {
	return opts.XAxis3D{Name: axis.Name, Show: rendererBool(axis.Show), Min: rendererBound(axis.Min), Max: rendererBound(axis.Max)}
}

func rendererScatter3DAxisY(axis Scatter3DAxis) opts.YAxis3D {
	return opts.YAxis3D{Name: axis.Name, Show: rendererBool(axis.Show), Min: rendererBound(axis.Min), Max: rendererBound(axis.Max)}
}

func rendererScatter3DAxisZ(axis Scatter3DAxis) opts.ZAxis3D {
	return opts.ZAxis3D{Name: axis.Name, Show: rendererBool(axis.Show), Min: rendererBound(axis.Min), Max: rendererBound(axis.Max)}
}

func rendererBool(value *bool) types.Bool {
	if value == nil {
		return nil
	}
	return opts.Bool(*value)
}

func rendererBound(value *float64) any {
	if value == nil {
		return nil
	}
	return *value
}

func rendererScatter3DVisualRange(value *Scatter3DVisualRange) opts.VisualMap {
	colors := append([]string(nil), value.Colors...)
	if value.Palette == Scatter3DPaletteColdToWarm {
		colors = append([]string(nil), scatter3DColdToWarmColors...)
	}
	result := opts.VisualMap{Min: float32(value.Min), Max: float32(value.Max), InRange: &opts.VisualMapInRange{Color: colors}}
	if value.Calculable != nil {
		result.Calculable = opts.Bool(*value.Calculable)
	}
	return result
}

func validateScatter3DConfig(cfg Scatter3DConfig) error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("scatter 3D chart label is required")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("scatter 3D chart series is required")
	}
	if cfg.Options.Legend != nil {
		return fmt.Errorf("scatter 3D chart legend is not supported")
	}
	if cfg.Options.XAxis != nil || cfg.Options.YAxis != nil {
		return fmt.Errorf("scatter 3D chart Cartesian axes must use Axes")
	}
	if tooltip := cfg.Options.Tooltip; tooltip != nil && tooltip.Trigger != "" && tooltip.Trigger != "item" {
		return fmt.Errorf("scatter 3D chart tooltip trigger %q is not supported", tooltip.Trigger)
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("scatter 3D chart root attribute %q is reserved", attribute)
			}
		}
	}
	if cfg.Axes != nil {
		for name, axis := range map[string]Scatter3DAxis{"x": cfg.Axes.X, "y": cfg.Axes.Y, "z": cfg.Axes.Z} {
			if strings.TrimSpace(axis.Name) == "" {
				return fmt.Errorf("scatter 3D chart %s axis name is required", name)
			}
			if axis.Min != nil && !finiteNumber(*axis.Min) {
				return fmt.Errorf("scatter 3D chart %s axis minimum must be finite", name)
			}
			if axis.Max != nil && !finiteNumber(*axis.Max) {
				return fmt.Errorf("scatter 3D chart %s axis maximum must be finite", name)
			}
			if axis.Min != nil && axis.Max != nil && *axis.Min > *axis.Max {
				return fmt.Errorf("scatter 3D chart %s axis minimum must not exceed maximum", name)
			}
		}
	}
	seriesNames := make(map[string]bool, len(cfg.Series))
	for seriesIndex, series := range cfg.Series {
		name := strings.TrimSpace(series.Name)
		if name == "" {
			return fmt.Errorf("scatter 3D chart series %d name is required", seriesIndex)
		}
		if seriesNames[name] {
			return fmt.Errorf("scatter 3D chart series %q is duplicated", name)
		}
		seriesNames[name] = true
		if len(series.Points) == 0 {
			return fmt.Errorf("scatter 3D chart series %q points are required", name)
		}
		if err := validateScatter3DPaint("series "+fmt.Sprintf("%q", name), series.Options.Color, series.Options.Class); err != nil {
			return err
		}
		if series.Options.SymbolSize < 0 {
			return fmt.Errorf("scatter 3D chart series %q symbol size must not be negative", name)
		}
		pointNames := make(map[string]bool, len(series.Points))
		for pointIndex, point := range series.Points {
			pointName := strings.TrimSpace(point.Name)
			if pointName == "" {
				return fmt.Errorf("scatter 3D chart series %q point %d name is required", name, pointIndex)
			}
			if pointNames[pointName] {
				return fmt.Errorf("scatter 3D chart series %q point %q is duplicated", name, pointName)
			}
			pointNames[pointName] = true
			if !finiteNumber(point.X) || !finiteNumber(point.Y) || !finiteNumber(point.Z) {
				return fmt.Errorf("scatter 3D chart series %q point %q coordinates must be finite", name, pointName)
			}
			if point.Value != nil && !finiteNumber(*point.Value) {
				return fmt.Errorf("scatter 3D chart series %q point %q value must be finite", name, pointName)
			}
			if point.SymbolSize < 0 {
				return fmt.Errorf("scatter 3D chart series %q point %q symbol size must not be negative", name, pointName)
			}
			if err := validateScatter3DPaint("series "+fmt.Sprintf("%q point %q", name, pointName), point.Color, point.Class); err != nil {
				return err
			}
		}
	}
	if rangeOptions := cfg.VisualRange; rangeOptions != nil {
		if !finiteNumber(rangeOptions.Min) || !finiteNumber(rangeOptions.Max) {
			return fmt.Errorf("scatter 3D chart visual range bounds must be finite")
		}
		if rangeOptions.Min > rangeOptions.Max {
			return fmt.Errorf("scatter 3D chart visual range minimum must not exceed maximum")
		}
		if rangeOptions.Palette != Scatter3DPaletteCustom && rangeOptions.Palette != Scatter3DPaletteColdToWarm {
			return fmt.Errorf("scatter 3D chart visual range palette %q is not supported", rangeOptions.Palette)
		}
		if rangeOptions.Palette == Scatter3DPaletteColdToWarm && len(rangeOptions.Colors) != 0 {
			return fmt.Errorf("scatter 3D chart cold-to-warm palette conflicts with custom colors")
		}
		if rangeOptions.Palette == Scatter3DPaletteCustom && len(rangeOptions.Colors) < 2 {
			return fmt.Errorf("scatter 3D chart custom visual range requires at least two colors")
		}
		for index, color := range rangeOptions.Colors {
			if strings.TrimSpace(color) == "" {
				return fmt.Errorf("scatter 3D chart visual range color %d is required", index)
			}
		}
		for _, series := range cfg.Series {
			if series.Options.Color != "" || series.Options.Class != "" {
				return fmt.Errorf("scatter 3D chart visual range conflicts with series paint")
			}
			for _, point := range series.Points {
				if point.Color != "" || point.Class != "" {
					return fmt.Errorf("scatter 3D chart visual range conflicts with point paint")
				}
			}
		}
	}
	return validateChartOptions(cfg.Options)
}

func validateScatter3DPaint(subject, color, class string) error {
	if strings.TrimSpace(color) != "" && strings.TrimSpace(class) != "" {
		return fmt.Errorf("scatter 3D chart %s color and class are mutually exclusive", subject)
	}
	return nil
}

func scatter3DPointCount(series []Scatter3DSeries) int {
	count := 0
	for _, item := range series {
		count += len(item.Points)
	}
	return count
}

func scatter3DPointValue(point Point3D) string {
	if point.Value == nil {
		return ""
	}
	return fmt.Sprintf("%g", *point.Value)
}
