package interactive

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// GaugeVariant selects a visual treatment without changing component identity.
type GaugeVariant string

const (
	// GaugeVariantStandard renders the default dial and pointer. It is the zero-value default.
	GaugeVariantStandard GaugeVariant = ""
	// GaugeVariantProgress renders a rounded progress arc without a pointer.
	GaugeVariantProgress GaugeVariant = "progress"
	// GaugeVariantLiquid renders bounded readings as layered waves.
	GaugeVariantLiquid GaugeVariant = "liquid"
)

// GaugeLiquidShape selects a built-in silhouette for the liquid treatment.
type GaugeLiquidShape string

const (
	GaugeLiquidShapeCircle    GaugeLiquidShape = ""
	GaugeLiquidShapeRect      GaugeLiquidShape = "rect"
	GaugeLiquidShapeRoundRect GaugeLiquidShape = "roundRect"
	GaugeLiquidShapeTriangle  GaugeLiquidShape = "triangle"
	GaugeLiquidShapeDiamond   GaugeLiquidShape = "diamond"
	GaugeLiquidShapePin       GaugeLiquidShape = "pin"
	GaugeLiquidShapeArrow     GaugeLiquidShape = "arrow"
)

// GaugeLiquidDirection selects horizontal wave travel. Right is the zero-value default.
type GaugeLiquidDirection string

const (
	GaugeLiquidDirectionRight GaugeLiquidDirection = ""
	GaugeLiquidDirectionLeft  GaugeLiquidDirection = "left"
)

// GaugeLiquidTreatment configures the liquid treatment without exposing its renderer.
type GaugeLiquidTreatment struct {
	Shape             GaugeLiquidShape
	WaveLengthPercent *float64
	AmplitudePercent  *float64
	PhaseDegrees      *float64
	Animate           *bool
	Direction         GaugeLiquidDirection
	Outline           *GaugeLiquidOutline
	Background        *GaugeLiquidBackground
	Label             *GaugeLiquidLabel
	Style             *GaugeLiquidStyle
}

// GaugeLiquidOutline configures the optional boundary around the liquid shape.
type GaugeLiquidOutline struct {
	Show         *bool
	Width        float64
	Color, Class string
}

// GaugeLiquidBackground configures the unfilled portion of the liquid shape.
type GaugeLiquidBackground struct {
	Color, Class             string
	BorderWidth              float64
	BorderColor, BorderClass string
}

// GaugeLiquidLabel configures the central bounded-reading label.
type GaugeLiquidLabel struct {
	Show         *bool
	FontSize     int
	Color, Class string
}

// GaugeLiquidStyle configures wave paint and opacity.
type GaugeLiquidStyle struct {
	Color, Class string
	Opacity      *float64
}

// GaugeConfig describes an accessible, browser-rendered gauge chart.
//
// Min defaults to 0. Max defaults to 100 when zero. Values must be
// application-owned because the browser renderer serializes them.
type GaugeConfig struct {
	Label         string
	Caption       string
	Variant       GaugeVariant
	Min           int
	Max           int
	Series        []GaugeSeries
	Width         string
	Height        string
	Options       ChartOptions
	SeriesOptions SeriesOptions
	Style         charttheme.Style
	Scale         GaugeScale
	Liquid        GaugeLiquidTreatment
}

// GaugeScaleMode selects scale treatment. Zero value is thermal sequential.
type GaugeScaleMode string

const (
	GaugeScaleThermal     GaugeScaleMode = ""
	GaugeScaleCustom      GaugeScaleMode = "custom"
	GaugeScaleSingleColor GaugeScaleMode = "single-color"
)

// GaugeScale configures the full Min-to-Max arc independently from progress.
type GaugeScale struct {
	Mode    GaugeScaleMode
	Reverse bool
	Stops   []GaugeScaleStop
	Color   string
	Class   string
}

// GaugeScaleStop applies a semantic color from the prior stop through Value.
type GaugeScaleStop struct {
	Value        float64
	Color, Class string
}

// GaugeSeries describes one named dial series. Series options override variant defaults.
type GaugeSeries struct {
	Name        string
	Data        []GaugeData
	Options     SeriesOptions
	Progress    *GaugeProgressOptions
	ShowPointer *bool
}

// GaugeProgressOptions customizes progress-arc rendering after variant defaults.
type GaugeProgressOptions struct {
	Show     *bool
	Width    int
	RoundCap *bool
	Clip     *bool
}

// GaugeData describes one named finite reading.
type GaugeData struct {
	Name  string
	Value float64
}

// Gauge builds a reusable interactive gauge component.
func Gauge(cfg GaugeConfig) Instance {
	if err := validateGaugeConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveGauge, err)
	}

	minimum, maximum := resolvedGaugeRange(cfg)
	if cfg.Variant == GaugeVariantLiquid {
		return liquidGauge(cfg, minimum, maximum)
	}
	chart := charts.NewGauge()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, chartGlobalOptions(cfg.Options)...)
	// Explicit component colors remain authoritative over escape-hatch options.
	if len(cfg.Style.Colors) > 0 {
		globalOptions = append(globalOptions, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
	if cfg.Width != "" || cfg.Height != "" {
		globalOptions = append([]charts.GlobalOpts{
			charts.WithInitializationOpts(opts.Initialization{Width: cfg.Width, Height: cfg.Height}),
		}, globalOptions...)
	}
	chart.SetGlobalOptions(globalOptions...)

	themeSeriesItems := make([]int, 0, len(cfg.Series))
	for seriesIndex, series := range cfg.Series {
		data := make([]opts.GaugeData, len(series.Data))
		for index, point := range series.Data {
			data[index] = opts.GaugeData{Name: point.Name, Value: point.Value}
		}
		options := make([]charts.SeriesOpts, 0, 1+len(chartSeriesOptions(cfg.SeriesOptions))+len(chartSeriesOptions(series.Options)))
		options = append(options, gaugeVariantOptions(cfg.Variant, minimum, maximum))
		if cfg.SeriesOptions.ItemStyle == nil && series.Options.ItemStyle == nil {
			themeSeriesItems = append(themeSeriesItems, seriesIndex)
		}
		options = append(options, mergeSeriesOptions(cfg.SeriesOptions, series.Options)...)
		if series.Progress != nil || series.ShowPointer != nil {
			progress, showPointer := series.Progress, series.ShowPointer
			options = append(options, func(rendered *charts.SingleSeries) {
				if progress != nil {
					if rendered.Progress == nil {
						rendered.Progress = &opts.Progress{}
					}
					if progress.Show != nil {
						rendered.Progress.Show = opts.Bool(*progress.Show)
					}
					rendered.Progress.Width = progress.Width
					if progress.RoundCap != nil {
						rendered.Progress.RoundCap = opts.Bool(*progress.RoundCap)
					}
					if progress.Clip != nil {
						rendered.Progress.Clip = opts.Bool(*progress.Clip)
					}
				}
				if showPointer != nil {
					if rendered.Pointer == nil {
						rendered.Pointer = &opts.Pointer{}
					}
					rendered.Pointer.Show = opts.Bool(*showPointer)
				}
			})
		}
		chart.AddSeries(series.Name, data, options...)
	}

	return newInstance(chartcomponents.KindInteractiveGauge, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export, ThemeSeriesItems: themeSeriesItems, GaugeScale: gaugeScaleJSON(cfg, minimum, maximum),
	})
}

func liquidGauge(cfg GaugeConfig, minimum, maximum int) Instance {
	chart := charts.NewLiquid()
	width, height := cfg.Width, cfg.Height
	if width == "" {
		width = "100%"
	}
	if height == "" {
		height = "500px"
	}
	globalOptions := []charts.GlobalOpts{
		charts.WithInitializationOpts(opts.Initialization{Width: width, Height: height}),
		charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())),
	}
	globalOptions = append(globalOptions, chartGlobalOptions(cfg.Options)...)
	if len(cfg.Style.Colors) > 0 {
		globalOptions = append(globalOptions, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
	chart.SetGlobalOptions(globalOptions...)
	series := cfg.Series[0]
	seriesOptions := mergeSeriesOptions(cfg.SeriesOptions, series.Options)
	seriesOptions = append(seriesOptions, charts.WithLiquidChartOpts(opts.LiquidChart{Shape: string(cfg.Liquid.Shape)}))
	chart.AddSeries(series.Name, normalizeGaugeLiquidData(series.Data, minimum, maximum), seriesOptions...)

	return newInstance(chartcomponents.KindInteractiveGauge, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style,
		Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		Details: gaugeLiquidExactValues(cfg.Label, minimum, maximum, series.Data),
		Liquid:  gaugeLiquidJSON(cfg.Liquid), ScriptReplacements: gaugeLiquidScriptReplacements(cfg.Liquid),
	})
}

func normalizeGaugeLiquidData(data []GaugeData, minimum, maximum int) []opts.LiquidData {
	result := make([]opts.LiquidData, len(data))
	span := float64(maximum - minimum)
	for index, point := range data {
		result[index] = opts.LiquidData{Name: point.Name, Value: (point.Value - float64(minimum)) / span}
	}
	return result
}

func gaugeLiquidScriptReplacements(treatment GaugeLiquidTreatment) []scriptReplacement {
	fields := make([]string, 0, 9)
	if treatment.Shape != "" {
		fields = append(fields, fmt.Sprintf(`"shape":%q`, treatment.Shape))
	}
	if treatment.WaveLengthPercent != nil {
		fields = append(fields, fmt.Sprintf(`"waveLength":%q`, percentage(*treatment.WaveLengthPercent)))
	}
	if treatment.AmplitudePercent != nil {
		fields = append(fields, fmt.Sprintf(`"amplitude":%q`, percentage(*treatment.AmplitudePercent)))
	}
	if treatment.PhaseDegrees != nil {
		fields = append(fields, fmt.Sprintf(`"phase":%g`, *treatment.PhaseDegrees*math.Pi/180))
	}
	if treatment.Animate != nil {
		fields = append(fields, fmt.Sprintf(`"waveAnimation":%t`, *treatment.Animate))
	}
	if treatment.Direction != "" {
		fields = append(fields, fmt.Sprintf(`"direction":%q`, treatment.Direction))
	}
	if outline := treatment.Outline; outline != nil {
		outlineFields := make([]string, 0, 2)
		if outline.Show != nil {
			outlineFields = append(outlineFields, fmt.Sprintf(`"show":%t`, *outline.Show))
		}
		if outline.Width != 0 {
			outlineFields = append(outlineFields, fmt.Sprintf(`"itemStyle":{"borderWidth":%g}`, outline.Width))
		}
		if len(outlineFields) > 0 {
			fields = append(fields, `"outline":{`+strings.Join(outlineFields, ",")+`}`)
		}
	}
	if background := treatment.Background; background != nil && background.BorderWidth != 0 {
		fields = append(fields, fmt.Sprintf(`"backgroundStyle":{"borderWidth":%g}`, background.BorderWidth))
	}
	if label := treatment.Label; label != nil {
		labelFields := make([]string, 0, 2)
		if label.Show != nil {
			labelFields = append(labelFields, fmt.Sprintf(`"show":%t`, *label.Show))
		}
		if label.FontSize != 0 {
			labelFields = append(labelFields, fmt.Sprintf(`"fontSize":%d`, label.FontSize))
		}
		if len(labelFields) > 0 {
			fields = append(fields, `"label":{`+strings.Join(labelFields, ",")+`}`)
		}
	}
	if style := treatment.Style; style != nil && style.Opacity != nil {
		fields = append(fields, fmt.Sprintf(`"itemStyle":{"opacity":%g}`, *style.Opacity))
	}
	if len(fields) == 0 {
		return nil
	}
	return []scriptReplacement{{Old: `"type":"liquidFill"`, New: `"type":"liquidFill",` + strings.Join(fields, ",")}}
}

type gaugeLiquidPayload struct {
	Outline    *gaugeLiquidOutlinePayload    `json:"outline,omitempty"`
	Background *gaugeLiquidBackgroundPayload `json:"background,omitempty"`
	Label      *gaugeLiquidPaintPayload      `json:"label,omitempty"`
	Style      *gaugeLiquidStylePayload      `json:"style,omitempty"`
}

type gaugeLiquidPaintPayload struct {
	Color string `json:"color,omitempty"`
	Class string `json:"class,omitempty"`
}

type gaugeLiquidOutlinePayload struct {
	gaugeLiquidPaintPayload
	Width float64 `json:"width,omitempty"`
}

type gaugeLiquidBackgroundPayload struct {
	gaugeLiquidPaintPayload
	BorderColor string  `json:"borderColor,omitempty"`
	BorderClass string  `json:"borderClass,omitempty"`
	BorderWidth float64 `json:"borderWidth,omitempty"`
}

type gaugeLiquidStylePayload struct {
	gaugeLiquidPaintPayload
	Opacity *float64 `json:"opacity,omitempty"`
}

func gaugeLiquidJSON(treatment GaugeLiquidTreatment) string {
	payload := gaugeLiquidPayload{}
	if value := treatment.Outline; value != nil {
		payload.Outline = &gaugeLiquidOutlinePayload{gaugeLiquidPaintPayload: gaugeLiquidPaintPayload{Color: value.Color, Class: value.Class}, Width: value.Width}
	}
	if value := treatment.Background; value != nil {
		payload.Background = &gaugeLiquidBackgroundPayload{gaugeLiquidPaintPayload: gaugeLiquidPaintPayload{Color: value.Color, Class: value.Class}, BorderColor: value.BorderColor, BorderClass: value.BorderClass, BorderWidth: value.BorderWidth}
	}
	if value := treatment.Label; value != nil {
		payload.Label = &gaugeLiquidPaintPayload{Color: value.Color, Class: value.Class}
	}
	if value := treatment.Style; value != nil {
		payload.Style = &gaugeLiquidStylePayload{gaugeLiquidPaintPayload: gaugeLiquidPaintPayload{Color: value.Color, Class: value.Class}, Opacity: value.Opacity}
	}
	data, _ := json.Marshal(payload)
	return string(data)
}

type gaugeScalePayload struct {
	Mode    string                  `json:"mode"`
	Reverse bool                    `json:"reverse,omitempty"`
	Stops   []gaugeScalePayloadStop `json:"stops,omitempty"`
}
type gaugeScalePayloadStop struct {
	Position float64 `json:"position"`
	Color    string  `json:"color,omitempty"`
	Class    string  `json:"class,omitempty"`
	Token    string  `json:"token,omitempty"`
}

func gaugeScaleJSON(cfg GaugeConfig, minimum, maximum int) string {
	payload := gaugeScalePayload{Mode: string(cfg.Scale.Mode), Reverse: cfg.Scale.Reverse}
	if cfg.Scale.Mode == GaugeScaleCustom {
		for _, stop := range cfg.Scale.Stops {
			payload.Stops = append(payload.Stops, gaugeScalePayloadStop{Position: (stop.Value - float64(minimum)) / float64(maximum-minimum), Color: stop.Color, Class: stop.Class})
		}
	} else if cfg.Scale.Mode == GaugeScaleSingleColor {
		payload.Stops = []gaugeScalePayloadStop{{Position: 1, Color: cfg.Scale.Color, Class: cfg.Scale.Class}}
	} else {
		payload.Stops = []gaugeScalePayloadStop{{Position: .34, Token: "low"}, {Position: .67, Token: "mid"}, {Position: 1, Token: "high"}}
	}
	data, _ := json.Marshal(payload)
	return string(data)
}

func gaugeVariantOptions(variant GaugeVariant, minimum, maximum int) charts.SeriesOpts {
	return func(series *charts.SingleSeries) {
		series.Min = minimum
		series.Max = maximum
		if variant == GaugeVariantProgress {
			series.Progress = &opts.Progress{
				Show: opts.Bool(true), Width: 6, RoundCap: opts.Bool(true), Clip: opts.Bool(true),
			}
			series.Pointer = &opts.Pointer{Show: opts.Bool(false)}
		}
	}
}

func resolvedGaugeRange(cfg GaugeConfig) (int, int) {
	maximum := cfg.Max
	if maximum == 0 {
		maximum = 100
	}
	return cfg.Min, maximum
}

func validateGaugeConfig(cfg GaugeConfig) error {
	if cfg.Label == "" {
		return fmt.Errorf("gauge chart label is required")
	}
	if cfg.Variant != GaugeVariantStandard && cfg.Variant != GaugeVariantProgress && cfg.Variant != GaugeVariantLiquid {
		return fmt.Errorf("gauge chart variant %q is not supported", cfg.Variant)
	}
	minimum, maximum := resolvedGaugeRange(cfg)
	if minimum >= maximum {
		return fmt.Errorf("gauge chart minimum must be less than maximum")
	}
	if cfg.Variant != GaugeVariantLiquid {
		if err := validateGaugeScale(cfg.Scale, minimum, maximum); err != nil {
			return err
		}
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("gauge chart series is required")
	}
	if err := validateGaugeLiquid(cfg); err != nil {
		return err
	}
	for seriesIndex, series := range cfg.Series {
		if series.Name == "" {
			return fmt.Errorf("gauge chart series %d name is required", seriesIndex)
		}
		if len(series.Data) == 0 {
			return fmt.Errorf("gauge chart series %q data is required", series.Name)
		}
		for dataIndex, point := range series.Data {
			if point.Name == "" {
				return fmt.Errorf("gauge chart series %q data point %d name is required", series.Name, dataIndex)
			}
			if math.IsNaN(point.Value) || math.IsInf(point.Value, 0) {
				return fmt.Errorf("gauge chart series %q data point %q value must be finite", series.Name, point.Name)
			}
			if point.Value < float64(minimum) || point.Value > float64(maximum) {
				return fmt.Errorf("gauge chart series %q data point %q value must be between %d and %d", series.Name, point.Name, minimum, maximum)
			}
		}
	}
	return nil
}

func validateGaugeLiquid(cfg GaugeConfig) error {
	if cfg.Variant != GaugeVariantLiquid {
		if cfg.Liquid != (GaugeLiquidTreatment{}) {
			name := "standard"
			if cfg.Variant == GaugeVariantProgress {
				name = "progress"
			}
			return fmt.Errorf("%s gauge treatment does not accept liquid options", name)
		}
		return nil
	}
	if len(cfg.Series) != 1 {
		return fmt.Errorf("liquid gauge treatment requires exactly one series")
	}
	if cfg.Series[0].Progress != nil {
		return fmt.Errorf("liquid gauge treatment does not accept progress options")
	}
	if cfg.Series[0].ShowPointer != nil {
		return fmt.Errorf("liquid gauge treatment does not accept pointer options")
	}
	if cfg.Scale.Mode != GaugeScaleThermal || cfg.Scale.Reverse || len(cfg.Scale.Stops) > 0 || cfg.Scale.Color != "" || cfg.Scale.Class != "" {
		return fmt.Errorf("liquid gauge treatment does not accept dial scale options")
	}
	if cfg.Options.XAxis != nil || cfg.Options.YAxis != nil {
		return fmt.Errorf("liquid gauge treatment does not accept Cartesian axes")
	}
	if cfg.Options.Legend != nil {
		return fmt.Errorf("liquid gauge treatment does not accept a legend")
	}
	if tooltip := cfg.Options.Tooltip; tooltip != nil && tooltip.Trigger != "" && tooltip.Trigger != "item" {
		return fmt.Errorf("liquid gauge tooltip trigger %q is not supported", tooltip.Trigger)
	}
	if err := validateGaugeLiquidSeriesOptions(cfg.SeriesOptions); err != nil {
		return err
	}
	if err := validateGaugeLiquidSeriesOptions(cfg.Series[0].Options); err != nil {
		return err
	}
	validShapes := map[GaugeLiquidShape]bool{GaugeLiquidShapeCircle: true, GaugeLiquidShapeRect: true, GaugeLiquidShapeRoundRect: true, GaugeLiquidShapeTriangle: true, GaugeLiquidShapeDiamond: true, GaugeLiquidShapePin: true, GaugeLiquidShapeArrow: true}
	if !validShapes[cfg.Liquid.Shape] {
		return fmt.Errorf("liquid gauge shape %q is not supported", cfg.Liquid.Shape)
	}
	if value := cfg.Liquid.WaveLengthPercent; value != nil && (!finiteNumber(*value) || *value < 1 || *value > 100) {
		return fmt.Errorf("liquid gauge wave length percentage must be finite and between 1 and 100")
	}
	if value := cfg.Liquid.AmplitudePercent; value != nil && (!finiteNumber(*value) || *value < 0 || *value > 100) {
		return fmt.Errorf("liquid gauge amplitude percentage must be finite and between 0 and 100")
	}
	if value := cfg.Liquid.PhaseDegrees; value != nil && (!finiteNumber(*value) || *value < -360 || *value > 360) {
		return fmt.Errorf("liquid gauge phase must be finite and between -360 and 360 degrees")
	}
	if cfg.Liquid.Direction != GaugeLiquidDirectionRight && cfg.Liquid.Direction != GaugeLiquidDirectionLeft {
		return fmt.Errorf("liquid gauge direction %q is not supported", cfg.Liquid.Direction)
	}
	if value := cfg.Liquid.Outline; value != nil {
		if err := validateGaugeLiquidPaint("outline", value.Color, value.Class); err != nil {
			return err
		}
		if !finiteNumber(value.Width) || value.Width < 0 || value.Width > 64 {
			return fmt.Errorf("liquid gauge outline width must be finite and between 0 and 64 pixels")
		}
	}
	if value := cfg.Liquid.Background; value != nil {
		if err := validateGaugeLiquidPaint("background", value.Color, value.Class); err != nil {
			return err
		}
		if err := validateGaugeLiquidPaint("background border", value.BorderColor, value.BorderClass); err != nil {
			return err
		}
		if !finiteNumber(value.BorderWidth) || value.BorderWidth < 0 || value.BorderWidth > 64 {
			return fmt.Errorf("liquid gauge background border width must be finite and between 0 and 64 pixels")
		}
	}
	if value := cfg.Liquid.Label; value != nil {
		if err := validateGaugeLiquidPaint("label", value.Color, value.Class); err != nil {
			return err
		}
		if value.FontSize < 0 || value.FontSize > 256 {
			return fmt.Errorf("liquid gauge label font size must be between 0 and 256 pixels")
		}
	}
	if value := cfg.Liquid.Style; value != nil {
		if err := validateGaugeLiquidPaint("wave", value.Color, value.Class); err != nil {
			return err
		}
		if value.Opacity != nil && (!finiteNumber(*value.Opacity) || *value.Opacity < 0 || *value.Opacity > 1) {
			return fmt.Errorf("liquid gauge wave opacity must be finite and between 0 and 1")
		}
	}
	return nil
}

func validateGaugeLiquidSeriesOptions(options SeriesOptions) error {
	switch {
	case options.LineStyle != nil:
		return fmt.Errorf("liquid gauge line style is not supported")
	case options.AreaStyle != nil:
		return fmt.Errorf("liquid gauge area style is not supported")
	case options.Stack != "":
		return fmt.Errorf("liquid gauge stacking is not supported")
	case options.Symbol != "" || options.SymbolSize != 0 || options.ShowSymbol != nil:
		return fmt.Errorf("liquid gauge symbols are not supported")
	case options.Smooth != nil || options.Step != "":
		return fmt.Errorf("liquid gauge line interpolation is not supported")
	case options.BarWidth != "":
		return fmt.Errorf("liquid gauge bar width is not supported")
	case options.BarGap != "":
		return fmt.Errorf("liquid gauge bar gap is not supported")
	}
	return nil
}

func validateGaugeLiquidPaint(name, color, class string) error {
	if color != "" && strings.TrimSpace(color) == "" {
		return fmt.Errorf("liquid gauge %s color must not be blank", name)
	}
	if class != "" && strings.TrimSpace(class) == "" {
		return fmt.Errorf("liquid gauge %s class must not be blank", name)
	}
	if color != "" && class != "" {
		return fmt.Errorf("liquid gauge %s requires at most one color or class", name)
	}
	return nil
}

func validateGaugeScale(scale GaugeScale, minimum, maximum int) error {
	if scale.Mode != GaugeScaleThermal && scale.Mode != GaugeScaleCustom && scale.Mode != GaugeScaleSingleColor {
		return fmt.Errorf("gauge chart scale mode %q is not supported", scale.Mode)
	}
	validPaint := func(color, class string) bool {
		return (strings.TrimSpace(color) == "") != (strings.TrimSpace(class) == "")
	}
	if scale.Mode == GaugeScaleThermal {
		if len(scale.Stops) > 0 || scale.Color != "" || scale.Class != "" {
			return fmt.Errorf("gauge chart thermal scale does not accept custom paint")
		}
		return nil
	}
	if scale.Mode == GaugeScaleSingleColor {
		if !validPaint(scale.Color, scale.Class) {
			return fmt.Errorf("gauge chart single-color scale requires exactly one color or class")
		}
		if len(scale.Stops) > 0 {
			return fmt.Errorf("gauge chart single-color scale does not accept stops")
		}
		return nil
	}
	if len(scale.Stops) < 2 {
		return fmt.Errorf("gauge chart custom scale requires at least two stops")
	}
	previous := float64(minimum)
	for index, stop := range scale.Stops {
		if math.IsNaN(stop.Value) || math.IsInf(stop.Value, 0) || stop.Value < float64(minimum) || stop.Value > float64(maximum) {
			return fmt.Errorf("gauge chart scale stop %d value must be within gauge range", index)
		}
		if index > 0 && stop.Value <= previous {
			return fmt.Errorf("gauge chart scale stops must be strictly increasing")
		}
		if !validPaint(stop.Color, stop.Class) {
			return fmt.Errorf("gauge chart scale stop %d requires exactly one color or class", index)
		}
		previous = stop.Value
	}
	if scale.Stops[len(scale.Stops)-1].Value != float64(maximum) {
		return fmt.Errorf("gauge chart final scale stop must equal maximum")
	}
	return nil
}
