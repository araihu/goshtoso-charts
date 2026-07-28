package interactive

import (
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

// Surface3DPalette identifies a renderer-neutral visual-range palette.
type Surface3DPalette string

const (
	// Surface3DPaletteCustom uses VisualRange.Colors.
	Surface3DPaletteCustom Surface3DPalette = ""
	// Surface3DPaletteColdToWarm preserves the pinned source's ten-color scale.
	Surface3DPaletteColdToWarm Surface3DPalette = "cold-to-warm"
)

// Surface3DShading identifies a supported surface treatment.
type Surface3DShading string

const (
	// Surface3DShadingDefault preserves the renderer default.
	Surface3DShadingDefault Surface3DShading = ""
	// Surface3DShadingColor uses unlit flat color.
	Surface3DShadingColor Surface3DShading = "color"
	// Surface3DShadingLambert uses diffuse light shading.
	Surface3DShadingLambert Surface3DShading = "lambert"
	// Surface3DShadingRealistic uses physically based surface shading.
	Surface3DShadingRealistic Surface3DShading = "realistic"
)

// Surface3DAxis configures one named numeric axis.
type Surface3DAxis struct {
	Name string
	Show *bool
	Min  *float64
	Max  *float64
}

// Surface3DAxes configures all three axes as one validated unit.
type Surface3DAxes struct {
	X Surface3DAxis
	Y Surface3DAxis
	Z Surface3DAxis
}

// Surface3DSeriesStyle configures renderer-neutral surface shading and paint.
type Surface3DSeriesStyle struct {
	Shading Surface3DShading
	Color   string
	Class   string
}

// Surface3DSeries contains one named ordered surface grid.
type Surface3DSeries struct {
	Name   string
	Points []Point3D
	Style  Surface3DSeriesStyle
}

// Surface3DVisualRange maps Z values to a continuous color scale.
type Surface3DVisualRange struct {
	Min        float64
	Max        float64
	Calculable *bool
	Palette    Surface3DPalette
	Colors     []string
}

// Surface3DGrid controls optional three-dimensional box dimensions and view.
type Surface3DGrid struct {
	Width  float64
	Height float64
	Depth  float64
	View   *Surface3DView
}

// Surface3DView configures optional automatic view rotation.
type Surface3DView struct {
	AutoRotate      *bool
	AutoRotateSpeed float64
}

// Surface3DDataSummary describes the exact formula represented by ordered points.
type Surface3DDataSummary struct {
	Formula string
}

// Surface3DConfig describes an accessible browser-rendered three-dimensional surface.
type Surface3DConfig struct {
	Label       string
	Caption     string
	Series      []Surface3DSeries
	Axes        *Surface3DAxes
	VisualRange *Surface3DVisualRange
	Grid        Surface3DGrid
	DataSummary Surface3DDataSummary
	Width       string
	Height      string
	Options     ChartOptions
	Style       charttheme.Style
	RootAttrs   templ.Attributes
}

type rendererSurface3DPoint struct {
	Value       []float64       `json:"value"`
	ClassName   string          `json:"className,omitempty"`
	SourceColor string          `json:"sourceColor,omitempty"`
	ItemStyle   *opts.ItemStyle `json:"itemStyle,omitempty"`
}

type surface3DPaint struct {
	Color string `json:"color,omitempty"`
	Class string `json:"class,omitempty"`
}

// Surface3D builds a reusable renderer-neutral three-dimensional surface component.
func Surface3D(cfg Surface3DConfig) Instance {
	if err := validateSurface3DConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveSurface3D, err)
	}

	style := cfg.Style
	style.Class = strings.TrimSpace("goshtoso-charts-surface-3d " + style.Class)
	width, height := cfg.Width, cfg.Height
	if width == "" {
		width = "100%"
	}
	if height == "" {
		height = "38rem"
	}

	chart := charts.NewSurface3D()
	global := []charts.GlobalOpts{
		charts.WithInitializationOpts(opts.Initialization{Width: width, Height: height}),
		charts.WithColorsOpts(opts.Colors(style.ResolvedColors())),
	}
	global = append(global, surface3DChartOptions(cfg.Options)...)
	if cfg.Axes != nil {
		global = append(global,
			charts.WithXAxis3DOpts(rendererSurface3DAxisX(cfg.Axes.X)),
			charts.WithYAxis3DOpts(rendererSurface3DAxisY(cfg.Axes.Y)),
			charts.WithZAxis3DOpts(rendererSurface3DAxisZ(cfg.Axes.Z)),
		)
	}
	grid := opts.Grid3D{BoxWidth: float32(cfg.Grid.Width), BoxHeight: float32(cfg.Grid.Height), BoxDepth: float32(cfg.Grid.Depth)}
	if cfg.Grid.View != nil {
		grid.ViewControl = &opts.ViewControl{
			AutoRotate: rendererBool(cfg.Grid.View.AutoRotate), AutoRotateSpeed: float32(cfg.Grid.View.AutoRotateSpeed),
		}
	}
	global = append(global, charts.WithGrid3DOpts(grid))
	if cfg.VisualRange != nil {
		global = append(global, charts.WithVisualMapOpts(rendererSurface3DVisualRange(cfg.VisualRange)))
	}
	chart.SetGlobalOptions(global...)

	paints := make([]surface3DPaint, len(cfg.Series))
	for seriesIndex, series := range cfg.Series {
		data := make([]opts.Chart3DData, len(series.Points))
		options := make([]charts.SeriesOpts, 0, 2)
		if series.Style.Shading != Surface3DShadingDefault {
			options = append(options, charts.WithBar3DChartOpts(opts.Bar3DChart{Shading: string(series.Style.Shading)}))
		}
		if series.Style.Color != "" {
			options = append(options, charts.WithItemStyleOpts(opts.ItemStyle{Color: series.Style.Color}))
		}
		chart.AddSeries(series.Name, data, options...)
		// go-echarts v2.7.2 serializes Surface3D series as scatter3D. Keep this
		// renderer repair private so no backing-engine type crosses the API.
		chart.MultiSeries[seriesIndex].Type = types.ChartSurface3D
		rendered := make([]rendererSurface3DPoint, len(series.Points))
		for pointIndex, point := range series.Points {
			rendered[pointIndex] = rendererSurface3DPoint{
				Value: []float64{point.X, point.Y, point.Z}, ClassName: point.Class, SourceColor: point.Color,
			}
			if point.Color != "" {
				rendered[pointIndex].ItemStyle = &opts.ItemStyle{Color: point.Color}
			}
		}
		chart.MultiSeries[seriesIndex].Data = rendered
		paints[seriesIndex] = surface3DPaint{Color: series.Style.Color, Class: series.Style.Class}
	}
	paintJSON, _ := json.Marshal(paints)

	return newInstance(chartcomponents.KindInteractiveSurface3D, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: style,
		Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		ResponsiveWidth: responsiveWidth(cfg.Width), RootAttrs: cfg.RootAttrs, Details: surface3DExactData(cfg.Label, cfg.Series, cfg.DataSummary),
		ExplicitVisualMapColors: cfg.VisualRange != nil && cfg.VisualRange.Palette != Surface3DPaletteColdToWarm,
		Surface3DPaints:         string(paintJSON), Surface3DColdToWarm: cfg.VisualRange != nil && cfg.VisualRange.Palette == Surface3DPaletteColdToWarm,
	})
}

func surface3DChartOptions(options ChartOptions) []charts.GlobalOpts {
	copy := options
	copy.Legend = nil
	copy.XAxis = nil
	copy.YAxis = nil
	return chartGlobalOptions(copy)
}

func rendererSurface3DAxisX(axis Surface3DAxis) opts.XAxis3D {
	return opts.XAxis3D{Name: axis.Name, Show: rendererBool(axis.Show), Min: rendererBound(axis.Min), Max: rendererBound(axis.Max)}
}

func rendererSurface3DAxisY(axis Surface3DAxis) opts.YAxis3D {
	return opts.YAxis3D{Name: axis.Name, Show: rendererBool(axis.Show), Min: rendererBound(axis.Min), Max: rendererBound(axis.Max)}
}

func rendererSurface3DAxisZ(axis Surface3DAxis) opts.ZAxis3D {
	return opts.ZAxis3D{Name: axis.Name, Show: rendererBool(axis.Show), Min: rendererBound(axis.Min), Max: rendererBound(axis.Max)}
}

func rendererSurface3DVisualRange(value *Surface3DVisualRange) opts.VisualMap {
	colors := append([]string(nil), value.Colors...)
	if value.Palette == Surface3DPaletteColdToWarm {
		colors = append([]string(nil), scatter3DColdToWarmColors...)
	}
	result := opts.VisualMap{
		Min: float32(value.Min), Max: float32(value.Max), Range: []float32{float32(value.Min), float32(value.Max)},
		InRange: &opts.VisualMapInRange{Color: colors},
	}
	if value.Calculable != nil {
		result.Calculable = opts.Bool(*value.Calculable)
	}
	return result
}

func validateSurface3DConfig(cfg Surface3DConfig) error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("surface 3D chart label is required")
	}
	if strings.TrimSpace(cfg.DataSummary.Formula) == "" {
		return fmt.Errorf("surface 3D chart data formula is required")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("surface 3D chart series is required")
	}
	if cfg.Options.Legend != nil {
		return fmt.Errorf("surface 3D chart legend is not supported")
	}
	if cfg.Options.XAxis != nil || cfg.Options.YAxis != nil {
		return fmt.Errorf("surface 3D chart Cartesian axes must use Axes")
	}
	if tooltip := cfg.Options.Tooltip; tooltip != nil && tooltip.Trigger != "" && tooltip.Trigger != "item" {
		return fmt.Errorf("surface 3D chart tooltip trigger %q is not supported", tooltip.Trigger)
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("surface 3D chart root attribute %q is reserved", attribute)
			}
		}
	}
	if cfg.Axes != nil {
		for name, axis := range map[string]Surface3DAxis{"x": cfg.Axes.X, "y": cfg.Axes.Y, "z": cfg.Axes.Z} {
			if strings.TrimSpace(axis.Name) == "" {
				return fmt.Errorf("surface 3D chart %s axis name is required", name)
			}
			if axis.Min != nil && !finiteNumber(*axis.Min) {
				return fmt.Errorf("surface 3D chart %s axis minimum must be finite", name)
			}
			if axis.Max != nil && !finiteNumber(*axis.Max) {
				return fmt.Errorf("surface 3D chart %s axis maximum must be finite", name)
			}
			if axis.Min != nil && axis.Max != nil && *axis.Min > *axis.Max {
				return fmt.Errorf("surface 3D chart %s axis minimum must not exceed maximum", name)
			}
		}
	}
	gridValues := []float64{cfg.Grid.Width, cfg.Grid.Height, cfg.Grid.Depth}
	setGridValues := 0
	for _, value := range gridValues {
		if !finiteNumber(value) || value < 0 {
			return fmt.Errorf("surface 3D chart grid sizes must be finite and positive when set")
		}
		if value > 0 {
			setGridValues++
		}
	}
	if setGridValues != 0 && setGridValues != len(gridValues) {
		return fmt.Errorf("surface 3D chart grid width, height, and depth must be set together")
	}
	if cfg.Grid.View != nil {
		if !finiteNumber(cfg.Grid.View.AutoRotateSpeed) || cfg.Grid.View.AutoRotateSpeed < 0 {
			return fmt.Errorf("surface 3D chart auto-rotate speed must be finite and positive when set")
		}
		if cfg.Grid.View.AutoRotateSpeed > 0 && (cfg.Grid.View.AutoRotate == nil || !*cfg.Grid.View.AutoRotate) {
			return fmt.Errorf("surface 3D chart auto-rotate speed requires auto rotation")
		}
	}
	seriesNames := make(map[string]bool, len(cfg.Series))
	for seriesIndex, series := range cfg.Series {
		name := strings.TrimSpace(series.Name)
		if name == "" {
			return fmt.Errorf("surface 3D chart series %d name is required", seriesIndex)
		}
		if seriesNames[name] {
			return fmt.Errorf("surface 3D chart series %q is duplicated", name)
		}
		seriesNames[name] = true
		if len(series.Points) == 0 {
			return fmt.Errorf("surface 3D chart series %q points are required", name)
		}
		if err := validateSurface3DPaint("series "+strconv.Quote(name), series.Style.Color, series.Style.Class); err != nil {
			return err
		}
		switch series.Style.Shading {
		case Surface3DShadingDefault, Surface3DShadingColor, Surface3DShadingLambert, Surface3DShadingRealistic:
		default:
			return fmt.Errorf("surface 3D chart series %q shading %q is not supported", name, series.Style.Shading)
		}
		coordinates := make(map[[2]uint64]bool, len(series.Points))
		for pointIndex, point := range series.Points {
			if !finiteNumber(point.X) || !finiteNumber(point.Y) || !finiteNumber(point.Z) {
				return fmt.Errorf("surface 3D chart series %q point %d coordinates must be finite", name, pointIndex)
			}
			key := [2]uint64{math.Float64bits(point.X), math.Float64bits(point.Y)}
			if coordinates[key] {
				return fmt.Errorf("surface 3D chart series %q coordinate [%g,%g] is duplicated", name, point.X, point.Y)
			}
			coordinates[key] = true
			if point.Value != nil {
				return fmt.Errorf("surface 3D chart series %q point %d separate visual value is not supported", name, pointIndex)
			}
			if point.Symbol != "" || point.SymbolSize != 0 {
				return fmt.Errorf("surface 3D chart series %q point %d symbol options are not supported", name, pointIndex)
			}
			if err := validateSurface3DPaint(fmt.Sprintf("series %q point %d", name, pointIndex), point.Color, point.Class); err != nil {
				return err
			}
		}
	}
	if rangeOptions := cfg.VisualRange; rangeOptions != nil {
		if !finiteNumber(rangeOptions.Min) || !finiteNumber(rangeOptions.Max) {
			return fmt.Errorf("surface 3D chart visual range bounds must be finite")
		}
		if rangeOptions.Min > rangeOptions.Max {
			return fmt.Errorf("surface 3D chart visual range minimum must not exceed maximum")
		}
		if rangeOptions.Palette != Surface3DPaletteCustom && rangeOptions.Palette != Surface3DPaletteColdToWarm {
			return fmt.Errorf("surface 3D chart visual range palette %q is not supported", rangeOptions.Palette)
		}
		if rangeOptions.Palette == Surface3DPaletteColdToWarm && len(rangeOptions.Colors) != 0 {
			return fmt.Errorf("surface 3D chart cold-to-warm palette conflicts with custom colors")
		}
		if rangeOptions.Palette == Surface3DPaletteCustom && len(rangeOptions.Colors) < 2 {
			return fmt.Errorf("surface 3D chart custom visual range requires at least two colors")
		}
		for index, color := range rangeOptions.Colors {
			if strings.TrimSpace(color) == "" {
				return fmt.Errorf("surface 3D chart visual range color %d is required", index)
			}
		}
		for _, series := range cfg.Series {
			if series.Style.Color != "" || series.Style.Class != "" {
				return fmt.Errorf("surface 3D chart visual range conflicts with series paint")
			}
			for _, point := range series.Points {
				if point.Color != "" || point.Class != "" {
					return fmt.Errorf("surface 3D chart visual range conflicts with point paint")
				}
			}
		}
	}
	return validateChartOptions(cfg.Options)
}

func validateSurface3DPaint(subject, color, class string) error {
	if strings.TrimSpace(color) != "" && strings.TrimSpace(class) != "" {
		return fmt.Errorf("surface 3D chart %s color and class are mutually exclusive", subject)
	}
	return nil
}

func surface3DPointCount(series []Surface3DSeries) int {
	count := 0
	for _, item := range series {
		count += len(item.Points)
	}
	return count
}

func surface3DBounds(series []Surface3DSeries) [6]float64 {
	bounds := [6]float64{math.Inf(1), math.Inf(-1), math.Inf(1), math.Inf(-1), math.Inf(1), math.Inf(-1)}
	for _, item := range series {
		for _, point := range item.Points {
			bounds[0] = math.Min(bounds[0], point.X)
			bounds[1] = math.Max(bounds[1], point.X)
			bounds[2] = math.Min(bounds[2], point.Y)
			bounds[3] = math.Max(bounds[3], point.Y)
			bounds[4] = math.Min(bounds[4], point.Z)
			bounds[5] = math.Max(bounds[5], point.Z)
		}
	}
	return bounds
}

func surface3DCSVURL(series []Surface3DSeries) templ.SafeURL {
	var csv strings.Builder
	csv.WriteString("series,x,y,z\n")
	for _, item := range series {
		for _, point := range item.Points {
			csv.WriteString(strconv.Quote(item.Name))
			csv.WriteByte(',')
			csv.WriteString(strconv.FormatFloat(point.X, 'g', -1, 64))
			csv.WriteByte(',')
			csv.WriteString(strconv.FormatFloat(point.Y, 'g', -1, 64))
			csv.WriteByte(',')
			csv.WriteString(strconv.FormatFloat(point.Z, 'g', -1, 64))
			csv.WriteByte('\n')
		}
	}
	return templ.SafeURL("data:text/csv;charset=utf-8," + url.PathEscape(csv.String()))
}

func surface3DDownloadName(label string) string {
	return chartcontrol.SafeFilename(label)
}
