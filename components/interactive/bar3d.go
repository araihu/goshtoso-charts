package interactive

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// Bar3DPalette identifies a renderer-neutral visual-range palette.
type Bar3DPalette string

const (
	// Bar3DPaletteCustom uses VisualRange.Colors.
	Bar3DPaletteCustom Bar3DPalette = ""
	// Bar3DPaletteColdToWarm preserves the pinned source's ten-step diverging scale.
	Bar3DPaletteColdToWarm Bar3DPalette = "cold-to-warm"
)

// Bar3DShading identifies a supported three-dimensional surface treatment.
type Bar3DShading string

const (
	// Bar3DShadingDefault preserves the renderer default.
	Bar3DShadingDefault Bar3DShading = ""
	// Bar3DShadingColor uses unlit flat color.
	Bar3DShadingColor Bar3DShading = "color"
	// Bar3DShadingLambert uses diffuse light shading.
	Bar3DShadingLambert Bar3DShading = "lambert"
	// Bar3DShadingRealistic uses physically based surface shading.
	Bar3DShadingRealistic Bar3DShading = "realistic"
)

// Bar3DCategoricalAxis configures one ordered categorical axis.
type Bar3DCategoricalAxis struct {
	Name       string
	Categories []string
	Show       *bool
}

// Bar3DAxes configures the two categorical base axes.
type Bar3DAxes struct {
	X Bar3DCategoricalAxis
	Y Bar3DCategoricalAxis
}

// Bar3DCell is one finite value at categorical X and Y indexes.
type Bar3DCell struct {
	XIndex int
	YIndex int
	Value  float64
	Color  string
	Class  string
}

// Bar3DSeriesOptions configures renderer-neutral surface shading and paint.
type Bar3DSeriesOptions struct {
	Shading Bar3DShading
	Color   string
	Class   string
}

// Bar3DSeries contains one named categorical cell collection.
type Bar3DSeries struct {
	Name    string
	Cells   []Bar3DCell
	Options Bar3DSeriesOptions
}

// Bar3DVisualRange maps cell values to a continuous color scale.
type Bar3DVisualRange struct {
	Min        float64
	Max        float64
	Calculable *bool
	Palette    Bar3DPalette
	Colors     []string
}

// Bar3DGridSize controls the categorical base-box dimensions.
type Bar3DGridSize struct {
	Width float64
	Depth float64
}

// Bar3DView configures optional automatic view rotation.
type Bar3DView struct {
	AutoRotate      *bool
	AutoRotateSpeed float64
}

// Bar3DConfig describes an accessible browser-rendered categorical 3D bar chart.
type Bar3DConfig struct {
	Label       string
	Caption     string
	Axes        Bar3DAxes
	Series      []Bar3DSeries
	VisualRange *Bar3DVisualRange
	Grid        Bar3DGridSize
	View        *Bar3DView
	Width       string
	Height      string
	Options     ChartOptions
	Style       charttheme.Style
	RootAttrs   templ.Attributes
}

type rendererBar3DCell struct {
	Value       []any           `json:"value"`
	ClassName   string          `json:"className,omitempty"`
	SourceColor string          `json:"sourceColor,omitempty"`
	ItemStyle   *opts.ItemStyle `json:"itemStyle,omitempty"`
}

type bar3DPaint struct {
	Color string `json:"color,omitempty"`
	Class string `json:"class,omitempty"`
}

// Bar3D builds a reusable renderer-neutral categorical three-dimensional bar component.
func Bar3D(cfg Bar3DConfig) Instance {
	if err := validateBar3DConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveBar3D, err)
	}

	style := cfg.Style
	style.Class = strings.TrimSpace("goshtoso-charts-bar-3d " + style.Class)
	width, height := cfg.Width, cfg.Height
	if width == "" {
		width = "100%"
	}
	if height == "" {
		height = "38rem"
	}

	chart := charts.NewBar3D()
	global := []charts.GlobalOpts{
		charts.WithInitializationOpts(opts.Initialization{Width: width, Height: height}),
		charts.WithColorsOpts(opts.Colors(style.ResolvedColors())),
		charts.WithXAxis3DOpts(opts.XAxis3D{Name: cfg.Axes.X.Name, Type: "category", Data: append([]string(nil), cfg.Axes.X.Categories...), Show: rendererBool(cfg.Axes.X.Show)}),
		charts.WithYAxis3DOpts(opts.YAxis3D{Name: cfg.Axes.Y.Name, Type: "category", Data: append([]string(nil), cfg.Axes.Y.Categories...), Show: rendererBool(cfg.Axes.Y.Show)}),
	}
	global = append(global, bar3DChartOptions(cfg.Options)...)
	grid := opts.Grid3D{BoxWidth: float32(cfg.Grid.Width), BoxDepth: float32(cfg.Grid.Depth)}
	if cfg.View != nil {
		grid.ViewControl = &opts.ViewControl{AutoRotate: rendererBool(cfg.View.AutoRotate), AutoRotateSpeed: float32(cfg.View.AutoRotateSpeed)}
	}
	global = append(global, charts.WithGrid3DOpts(grid))
	if cfg.VisualRange != nil {
		global = append(global, charts.WithVisualMapOpts(rendererBar3DVisualRange(cfg.VisualRange)))
	}
	chart.SetGlobalOptions(global...)

	paints := make([]bar3DPaint, len(cfg.Series))
	for seriesIndex, series := range cfg.Series {
		data := make([]opts.Chart3DData, len(series.Cells))
		options := make([]charts.SeriesOpts, 0, 2)
		if series.Options.Shading != Bar3DShadingDefault {
			options = append(options, charts.WithBar3DChartOpts(opts.Bar3DChart{Shading: string(series.Options.Shading)}))
		}
		if series.Options.Color != "" {
			options = append(options, charts.WithItemStyleOpts(opts.ItemStyle{Color: series.Options.Color}))
		}
		chart.AddSeries(series.Name, data, options...)
		rendered := make([]rendererBar3DCell, len(series.Cells))
		for cellIndex, cell := range series.Cells {
			rendered[cellIndex] = rendererBar3DCell{
				Value: []any{cell.XIndex, cell.YIndex, cell.Value}, ClassName: cell.Class, SourceColor: cell.Color,
			}
			if cell.Color != "" {
				rendered[cellIndex].ItemStyle = &opts.ItemStyle{Color: cell.Color}
			}
		}
		chart.MultiSeries[seriesIndex].Data = rendered
		paints[seriesIndex] = bar3DPaint{Color: series.Options.Color, Class: series.Options.Class}
	}
	paintJSON, _ := json.Marshal(paints)

	return newInstance(chartcomponents.KindInteractiveBar3D, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: style, ResponsiveWidth: responsiveWidth(cfg.Width),
		Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		RootAttrs: cfg.RootAttrs, Details: bar3DExactCells(cfg.Axes, cfg.Series),
		ExplicitVisualMapColors: cfg.VisualRange != nil && cfg.VisualRange.Palette != Bar3DPaletteColdToWarm,
		Bar3DPaints:             string(paintJSON),
		Bar3DColdToWarm:         cfg.VisualRange != nil && cfg.VisualRange.Palette == Bar3DPaletteColdToWarm,
	})
}

func bar3DChartOptions(options ChartOptions) []charts.GlobalOpts {
	copy := options
	copy.Legend = nil
	copy.XAxis = nil
	copy.YAxis = nil
	return chartGlobalOptions(copy)
}

func rendererBar3DVisualRange(value *Bar3DVisualRange) opts.VisualMap {
	colors := append([]string(nil), value.Colors...)
	if value.Palette == Bar3DPaletteColdToWarm {
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

func validateBar3DConfig(cfg Bar3DConfig) error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("bar 3D chart label is required")
	}
	for name, axis := range map[string]Bar3DCategoricalAxis{"x": cfg.Axes.X, "y": cfg.Axes.Y} {
		if strings.TrimSpace(axis.Name) == "" {
			return fmt.Errorf("bar 3D chart %s axis name is required", name)
		}
		if len(axis.Categories) == 0 {
			return fmt.Errorf("bar 3D chart %s axis categories are required", name)
		}
		seen := make(map[string]bool, len(axis.Categories))
		for index, category := range axis.Categories {
			category = strings.TrimSpace(category)
			if category == "" {
				return fmt.Errorf("bar 3D chart %s axis category %d is required", name, index)
			}
			if seen[category] {
				return fmt.Errorf("bar 3D chart %s axis category %q is duplicated", name, category)
			}
			seen[category] = true
		}
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("bar 3D chart series is required")
	}
	if cfg.Grid.Width < 0 || cfg.Grid.Depth < 0 || !bar3DRendererNumber(cfg.Grid.Width) || !bar3DRendererNumber(cfg.Grid.Depth) {
		return fmt.Errorf("bar 3D chart grid sizes must be finite and positive when set")
	}
	if cfg.Grid.Width == 0 && cfg.Grid.Depth != 0 || cfg.Grid.Width != 0 && cfg.Grid.Depth == 0 {
		return fmt.Errorf("bar 3D chart grid width and depth must be set together")
	}
	if cfg.View != nil {
		if !bar3DRendererNumber(cfg.View.AutoRotateSpeed) || cfg.View.AutoRotateSpeed < 0 {
			return fmt.Errorf("bar 3D chart auto-rotate speed must be finite and positive when set")
		}
		if cfg.View.AutoRotateSpeed > 0 && (cfg.View.AutoRotate == nil || !*cfg.View.AutoRotate) {
			return fmt.Errorf("bar 3D chart auto-rotate speed requires auto rotation")
		}
	}
	if cfg.Options.Legend != nil {
		return fmt.Errorf("bar 3D chart legend is not supported")
	}
	if cfg.Options.XAxis != nil || cfg.Options.YAxis != nil {
		return fmt.Errorf("bar 3D chart Cartesian axes must use Axes")
	}
	if tooltip := cfg.Options.Tooltip; tooltip != nil && tooltip.Trigger != "" && tooltip.Trigger != "item" {
		return fmt.Errorf("bar 3D chart tooltip trigger %q is not supported", tooltip.Trigger)
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("bar 3D chart root attribute %q is reserved", attribute)
			}
		}
	}
	seriesNames := make(map[string]bool, len(cfg.Series))
	for seriesIndex, series := range cfg.Series {
		name := strings.TrimSpace(series.Name)
		if name == "" {
			return fmt.Errorf("bar 3D chart series %d name is required", seriesIndex)
		}
		if seriesNames[name] {
			return fmt.Errorf("bar 3D chart series %q is duplicated", name)
		}
		seriesNames[name] = true
		if len(series.Cells) == 0 {
			return fmt.Errorf("bar 3D chart series %q cells are required", name)
		}
		if err := validateBar3DPaint("series "+fmt.Sprintf("%q", name), series.Options.Color, series.Options.Class); err != nil {
			return err
		}
		switch series.Options.Shading {
		case Bar3DShadingDefault, Bar3DShadingColor, Bar3DShadingLambert, Bar3DShadingRealistic:
		default:
			return fmt.Errorf("bar 3D chart series %q shading %q is not supported", name, series.Options.Shading)
		}
		coordinates := make(map[[2]int]bool, len(series.Cells))
		for cellIndex, cell := range series.Cells {
			if cell.XIndex < 0 || cell.XIndex >= len(cfg.Axes.X.Categories) {
				return fmt.Errorf("bar 3D chart series %q cell %d x index %d is out of range", name, cellIndex, cell.XIndex)
			}
			if cell.YIndex < 0 || cell.YIndex >= len(cfg.Axes.Y.Categories) {
				return fmt.Errorf("bar 3D chart series %q cell %d y index %d is out of range", name, cellIndex, cell.YIndex)
			}
			coordinate := [2]int{cell.XIndex, cell.YIndex}
			if coordinates[coordinate] {
				return fmt.Errorf("bar 3D chart series %q coordinate [%d,%d] is duplicated", name, cell.XIndex, cell.YIndex)
			}
			coordinates[coordinate] = true
			if !finiteNumber(cell.Value) {
				return fmt.Errorf("bar 3D chart series %q cell [%d,%d] value must be finite", name, cell.XIndex, cell.YIndex)
			}
			if err := validateBar3DPaint("series "+fmt.Sprintf("%q cell [%d,%d]", name, cell.XIndex, cell.YIndex), cell.Color, cell.Class); err != nil {
				return err
			}
		}
	}
	if value := cfg.VisualRange; value != nil {
		if !bar3DRendererNumber(value.Min) || !bar3DRendererNumber(value.Max) {
			return fmt.Errorf("bar 3D chart visual range bounds must be finite")
		}
		if value.Min > value.Max {
			return fmt.Errorf("bar 3D chart visual range minimum must not exceed maximum")
		}
		if value.Palette != Bar3DPaletteCustom && value.Palette != Bar3DPaletteColdToWarm {
			return fmt.Errorf("bar 3D chart visual range palette %q is not supported", value.Palette)
		}
		if value.Palette == Bar3DPaletteColdToWarm && len(value.Colors) != 0 {
			return fmt.Errorf("bar 3D chart cold-to-warm palette conflicts with custom colors")
		}
		if value.Palette == Bar3DPaletteCustom && len(value.Colors) < 2 {
			return fmt.Errorf("bar 3D chart custom visual range requires at least two colors")
		}
		for index, color := range value.Colors {
			if strings.TrimSpace(color) == "" {
				return fmt.Errorf("bar 3D chart visual range color %d is required", index)
			}
		}
		for _, series := range cfg.Series {
			if series.Options.Color != "" || series.Options.Class != "" {
				return fmt.Errorf("bar 3D chart visual range conflicts with series paint")
			}
			for _, cell := range series.Cells {
				if cell.Value < value.Min || cell.Value > value.Max {
					return fmt.Errorf("bar 3D chart series %q cell [%d,%d] value is outside visual range", series.Name, cell.XIndex, cell.YIndex)
				}
				if cell.Color != "" || cell.Class != "" {
					return fmt.Errorf("bar 3D chart visual range conflicts with cell paint")
				}
			}
		}
	}
	return validateChartOptions(cfg.Options)
}

func bar3DRendererNumber(value float64) bool {
	return finiteNumber(value) && math.Abs(value) <= math.MaxFloat32
}

func validateBar3DPaint(subject, color, class string) error {
	if strings.TrimSpace(color) != "" && strings.TrimSpace(class) != "" {
		return fmt.Errorf("bar 3D chart %s color and class are mutually exclusive", subject)
	}
	return nil
}

func bar3DCellCount(series []Bar3DSeries) int {
	count := 0
	for _, item := range series {
		count += len(item.Cells)
	}
	return count
}
