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
)

// Line3DPalette identifies a renderer-neutral visual-range palette.
type Line3DPalette string

const (
	// Line3DPaletteColdToWarm uses the source treatment through chart theme tokens.
	Line3DPaletteColdToWarm Line3DPalette = ""
	// Line3DPaletteCustom uses VisualRange.Colors exactly as supplied.
	Line3DPaletteCustom Line3DPalette = "custom"
)

// Line3DVisualRange maps ordered point values to a continuous color scale.
type Line3DVisualRange struct {
	Min        float64
	Max        float64
	Calculable *bool
	Palette    Line3DPalette
	Colors     []string
}

// Line3DView configures optional automatic camera rotation.
type Line3DView struct {
	AutoRotate      *bool
	AutoRotateSpeed float64
}

// Line3DGrid configures optional three-dimensional box dimensions and view.
type Line3DGrid struct {
	Width  float64
	Height float64
	Depth  float64
	View   *Line3DView
}

// Line3DSeries contains one named ordered path.
type Line3DSeries struct {
	Name    string
	Points  []Point3D
	Options SeriesOptions
	Color   string
	Class   string
}

// Line3DDataSummary describes the exact formula and parameter domain.
type Line3DDataSummary struct {
	Formula      string
	Parameter    string
	ParameterMin float64
	ParameterMax float64
}

// Line3DConfig describes an accessible browser-rendered three-dimensional line.
type Line3DConfig struct {
	Label       string
	Caption     string
	Series      []Line3DSeries
	VisualRange *Line3DVisualRange
	Grid        Line3DGrid
	DataSummary Line3DDataSummary
	Width       string
	Height      string
	Options     ChartOptions
	Style       charttheme.Style
	RootAttrs   templ.Attributes
}

type line3DPaint struct {
	Color string `json:"color,omitempty"`
	Class string `json:"class,omitempty"`
}

// Line3D builds a reusable renderer-neutral three-dimensional line component.
func Line3D(cfg Line3DConfig) Instance {
	if err := validateLine3DConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveLine3D, err)
	}

	style := cfg.Style
	style.Class = strings.TrimSpace("goshtoso-charts-line-3d " + style.Class)
	width, height := cfg.Width, cfg.Height
	if width == "" {
		width = "100%"
	}
	if height == "" {
		height = "38rem"
	}

	chart := charts.NewLine3D()
	global := []charts.GlobalOpts{
		charts.WithInitializationOpts(opts.Initialization{Width: width, Height: height}),
		charts.WithColorsOpts(opts.Colors(style.ResolvedColors())),
	}
	global = append(global, line3DChartOptions(cfg.Options)...)
	grid := opts.Grid3D{BoxWidth: float32(cfg.Grid.Width), BoxHeight: float32(cfg.Grid.Height), BoxDepth: float32(cfg.Grid.Depth)}
	if cfg.Grid.View != nil {
		grid.ViewControl = &opts.ViewControl{
			AutoRotate: rendererBool(cfg.Grid.View.AutoRotate), AutoRotateSpeed: float32(cfg.Grid.View.AutoRotateSpeed),
		}
	}
	global = append(global, charts.WithGrid3DOpts(grid))
	if cfg.VisualRange != nil {
		global = append(global, charts.WithVisualMapOpts(rendererLine3DVisualRange(cfg.VisualRange)))
	}
	chart.SetGlobalOptions(global...)

	paints := make([]line3DPaint, len(cfg.Series))
	for seriesIndex, series := range cfg.Series {
		data := make([]opts.Chart3DData, len(series.Points))
		for pointIndex, point := range series.Points {
			data[pointIndex] = opts.Chart3DData{Value: []interface{}{point.X, point.Y, point.Z}}
		}
		options := chartSeriesOptions(series.Options)
		if series.Color != "" {
			options = append(options, charts.WithLineStyleOpts(opts.LineStyle{Color: series.Color}))
		}
		chart.AddSeries(series.Name, data, options...)
		paints[seriesIndex] = line3DPaint{Color: line3DSeriesColor(series), Class: series.Class}
	}
	paintJSON, _ := json.Marshal(paints)
	autoRotate := cfg.Grid.View != nil && cfg.Grid.View.AutoRotate != nil && *cfg.Grid.View.AutoRotate

	return newInstance(chartcomponents.KindInteractiveLine3D, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: style,
		Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		ResponsiveWidth: responsiveWidth(cfg.Width), RootAttrs: cfg.RootAttrs, Details: line3DExactData(cfg.Label, cfg.Series, cfg.DataSummary, autoRotate),
		ExplicitVisualMapColors: cfg.VisualRange != nil && cfg.VisualRange.Palette == Line3DPaletteCustom,
		Line3DPaints:            string(paintJSON),
		Line3DColdToWarm:        cfg.VisualRange != nil && cfg.VisualRange.Palette == Line3DPaletteColdToWarm,
		Line3DAutoRotate:        autoRotate,
	})
}

func line3DChartOptions(options ChartOptions) []charts.GlobalOpts {
	copy := options
	copy.XAxis = nil
	copy.YAxis = nil
	return chartGlobalOptions(copy)
}

func rendererLine3DVisualRange(value *Line3DVisualRange) opts.VisualMap {
	colors := append([]string(nil), value.Colors...)
	if value.Palette == Line3DPaletteColdToWarm {
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

func validateLine3DConfig(cfg Line3DConfig) error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("line 3D chart label is required")
	}
	if strings.TrimSpace(cfg.DataSummary.Formula) == "" {
		return fmt.Errorf("line 3D chart data formula is required")
	}
	if strings.TrimSpace(cfg.DataSummary.Parameter) == "" {
		return fmt.Errorf("line 3D chart data parameter is required")
	}
	if !finiteNumber(cfg.DataSummary.ParameterMin) || !finiteNumber(cfg.DataSummary.ParameterMax) {
		return fmt.Errorf("line 3D chart parameter domain must be finite")
	}
	if cfg.DataSummary.ParameterMin > cfg.DataSummary.ParameterMax {
		return fmt.Errorf("line 3D chart parameter minimum must not exceed maximum")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("line 3D chart series is required")
	}
	if cfg.Options.XAxis != nil || cfg.Options.YAxis != nil {
		return fmt.Errorf("line 3D chart Cartesian axes are not supported")
	}
	if tooltip := cfg.Options.Tooltip; tooltip != nil && tooltip.Trigger != "" && tooltip.Trigger != "item" {
		return fmt.Errorf("line 3D chart tooltip trigger %q is not supported", tooltip.Trigger)
	}
	for seriesIndex, series := range cfg.Series {
		if strings.TrimSpace(series.Name) == "" {
			return fmt.Errorf("line 3D chart series %d name is required", seriesIndex)
		}
		if len(series.Points) < 2 {
			return fmt.Errorf("line 3D chart series %q requires at least two points", series.Name)
		}
		for pointIndex, point := range series.Points {
			if !finiteNumber(point.X) || !finiteNumber(point.Y) || !finiteNumber(point.Z) {
				return fmt.Errorf("line 3D chart series %q point %d coordinates must be finite", series.Name, pointIndex)
			}
			if point.Name != "" || point.Value != nil || point.Symbol != "" || point.SymbolSize != 0 || point.Color != "" || point.Class != "" {
				return fmt.Errorf("line 3D chart series %q point %d supports coordinates only", series.Name, pointIndex)
			}
		}
		if err := validateLine3DSeriesOptions(series); err != nil {
			return err
		}
	}
	if err := validateLine3DGrid(cfg.Grid); err != nil {
		return err
	}
	if cfg.VisualRange != nil {
		value := cfg.VisualRange
		if !finiteNumber(value.Min) || !finiteNumber(value.Max) {
			return fmt.Errorf("line 3D chart visual range bounds must be finite")
		}
		if value.Min > value.Max {
			return fmt.Errorf("line 3D chart visual range minimum must not exceed maximum")
		}
		if value.Palette != Line3DPaletteColdToWarm && value.Palette != Line3DPaletteCustom {
			return fmt.Errorf("line 3D chart visual range palette %q is not supported", value.Palette)
		}
		if value.Palette == Line3DPaletteColdToWarm && len(value.Colors) != 0 {
			return fmt.Errorf("line 3D chart cold-to-warm palette does not accept custom colors")
		}
		if value.Palette == Line3DPaletteCustom && len(value.Colors) < 2 {
			return fmt.Errorf("line 3D chart custom visual range requires at least two colors")
		}
		for index, color := range value.Colors {
			if strings.TrimSpace(color) == "" {
				return fmt.Errorf("line 3D chart visual range color %d is required", index)
			}
		}
	}
	return validateChartOptions(cfg.Options)
}

func validateLine3DSeriesOptions(series Line3DSeries) error {
	if strings.TrimSpace(series.Color) != "" && strings.TrimSpace(series.Class) != "" {
		return fmt.Errorf("line 3D chart series %q color and class are mutually exclusive", series.Name)
	}
	options := series.Options
	if options.Label != nil || options.ItemStyle != nil || options.AreaStyle != nil || options.Stack != "" ||
		options.Symbol != "" || options.SymbolSize != 0 || options.Smooth != nil || options.ShowSymbol != nil ||
		options.Step != "" || options.BarWidth != "" || options.BarGap != "" {
		return fmt.Errorf("line 3D chart series %q contains unsupported series options", series.Name)
	}
	if options.LineStyle != nil {
		line := options.LineStyle
		if !finiteNumber(line.Width) || line.Width < 0 {
			return fmt.Errorf("line 3D chart series %q line width must be finite and nonnegative", series.Name)
		}
		if line.Opacity != nil && (!finiteNumber(*line.Opacity) || *line.Opacity < 0 || *line.Opacity > 1) {
			return fmt.Errorf("line 3D chart series %q line opacity must be finite and between zero and one", series.Name)
		}
		if line.Type != "" && line.Type != "solid" && line.Type != "dashed" && line.Type != "dotted" {
			return fmt.Errorf("line 3D chart series %q line type %q is not supported", series.Name, line.Type)
		}
		if line.Color != "" && series.Color != "" {
			return fmt.Errorf("line 3D chart series %q color is configured twice", series.Name)
		}
		if line.Color != "" && series.Class != "" {
			return fmt.Errorf("line 3D chart series %q line color and class are mutually exclusive", series.Name)
		}
	}
	return nil
}

func validateLine3DGrid(grid Line3DGrid) error {
	sizes := []float64{grid.Width, grid.Height, grid.Depth}
	set := 0
	for _, size := range sizes {
		if size != 0 {
			set++
		}
	}
	if set != 0 && set != len(sizes) {
		return fmt.Errorf("line 3D chart grid width, height, and depth must be set together")
	}
	if set != 0 {
		for _, size := range sizes {
			if !finiteNumber(size) || size <= 0 {
				return fmt.Errorf("line 3D chart grid sizes must be finite and positive when set")
			}
		}
	}
	if grid.View != nil {
		if !finiteNumber(grid.View.AutoRotateSpeed) || grid.View.AutoRotateSpeed < 0 {
			return fmt.Errorf("line 3D chart auto-rotation speed must be finite and nonnegative")
		}
	}
	return nil
}

func line3DSeriesColor(series Line3DSeries) string {
	if series.Color != "" {
		return series.Color
	}
	if series.Options.LineStyle != nil {
		return series.Options.LineStyle.Color
	}
	return ""
}

func line3DPointCount(series []Line3DSeries) int {
	count := 0
	for _, item := range series {
		count += len(item.Points)
	}
	return count
}

func line3DBounds(series []Line3DSeries) [6]float64 {
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

func line3DCSVURL(series []Line3DSeries) templ.SafeURL {
	var csv strings.Builder
	csv.WriteString("series,index,x,y,z\n")
	for _, item := range series {
		for index, point := range item.Points {
			csv.WriteString(strconv.Quote(item.Name))
			csv.WriteByte(',')
			csv.WriteString(strconv.Itoa(index))
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

func line3DDownloadName(label string) string {
	return chartcontrol.SafeFilename(label)
}
