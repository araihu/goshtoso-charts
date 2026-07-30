// Package gauge provides the canonical interactive gauge API.
//
// Standard, progress, and liquid treatments remain variants of one component.
// Gauge-specific types and implementation live here; shared renderer-neutral
// options remain in components/chart.
package gauge

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// Instance is the renderer-neutral chart instance returned by Gauge.
type Instance = chart.Instance

// Variant selects a visual treatment without changing component identity.
type Variant string

const (
	// VariantStandard renders the default dial and pointer. It is the zero-value default.
	VariantStandard Variant = ""
	// VariantProgress renders a rounded progress arc without a pointer.
	VariantProgress Variant = "progress"
	// VariantLiquid renders bounded readings as layered waves.
	VariantLiquid Variant = "liquid"
)

// LiquidShape selects a built-in silhouette for the liquid treatment.
type LiquidShape string

const (
	LiquidShapeCircle    LiquidShape = ""
	LiquidShapeRect      LiquidShape = "rect"
	LiquidShapeRoundRect LiquidShape = "roundRect"
	LiquidShapeTriangle  LiquidShape = "triangle"
	LiquidShapeDiamond   LiquidShape = "diamond"
	LiquidShapePin       LiquidShape = "pin"
	LiquidShapeArrow     LiquidShape = "arrow"
)

// LiquidDirection selects horizontal wave travel. Right is the zero-value default.
type LiquidDirection string

const (
	LiquidDirectionRight LiquidDirection = ""
	LiquidDirectionLeft  LiquidDirection = "left"
)

// LiquidTreatment configures the liquid treatment without exposing its renderer.
type LiquidTreatment struct {
	Shape             LiquidShape
	WaveLengthPercent *float64
	AmplitudePercent  *float64
	PhaseDegrees      *float64
	Animate           *bool
	Direction         LiquidDirection
	Outline           *LiquidOutline
	Background        *LiquidBackground
	Label             *LiquidLabel
	Style             *LiquidStyle
}

// LiquidOutline configures the optional boundary around the liquid shape.
type LiquidOutline struct {
	Show         *bool
	Width        float64
	Color, Class string
}

// LiquidBackground configures the unfilled portion of the liquid shape.
type LiquidBackground struct {
	Color, Class             string
	BorderWidth              float64
	BorderColor, BorderClass string
}

// LiquidLabel configures the central bounded-reading label.
type LiquidLabel struct {
	Show         *bool
	FontSize     int
	Color, Class string
}

// LiquidStyle configures wave paint and opacity.
type LiquidStyle struct {
	Color, Class string
	Opacity      *float64
}

// Config describes an accessible, browser-rendered gauge chart.
//
// Min defaults to 0. Max defaults to 100 when zero. Values must be
// application-owned because the browser renderer serializes them.
type Config struct {
	Label         string
	Caption       string
	Variant       Variant
	Min           int
	Max           int
	Series        []Series
	Width         string
	Height        string
	Options       chart.ChartOptions
	SeriesOptions chart.SeriesOptions
	Style         charttheme.Style
	Scale         Scale
	Liquid        LiquidTreatment
}

// ScaleMode selects scale treatment. Zero value is thermal sequential.
type ScaleMode string

const (
	ScaleThermal     ScaleMode = ""
	ScaleCustom      ScaleMode = "custom"
	ScaleSingleColor ScaleMode = "single-color"
)

// Scale configures the full Min-to-Max arc independently from progress.
type Scale struct {
	Mode    ScaleMode
	Reverse bool
	Stops   []ScaleStop
	Color   string
	Class   string
}

// ScaleStop applies a semantic color from the prior stop through Value.
type ScaleStop struct {
	Value        float64
	Color, Class string
}

// Series describes one named dial series. Series options override variant defaults.
type Series struct {
	Name        string
	Data        []Data
	Options     chart.SeriesOptions
	Progress    *ProgressOptions
	ShowPointer *bool
}

// ProgressOptions customizes progress-arc rendering after variant defaults.
type ProgressOptions struct {
	Show     *bool
	Width    int
	RoundCap *bool
	Clip     *bool
}

// Data describes one named finite reading.
type Data struct {
	Name  string
	Value float64
}

// Gauge builds a reusable interactive gauge component.
func Gauge(cfg Config) Instance {
	if err := validateConfig(cfg); err != nil {
		return internalinteractive.Invalid(chartcomponents.KindInteractiveGauge, err)
	}

	minimum, maximum := resolvedRange(cfg)
	if cfg.Variant == VariantLiquid {
		return liquidGauge(cfg, minimum, maximum)
	}
	chart := charts.NewGauge()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, internalinteractive.ChartGlobalOptions(cfg.Options)...)
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
		options := make([]charts.SeriesOpts, 0, 1+len(internalinteractive.ChartSeriesOptions(cfg.SeriesOptions))+len(internalinteractive.ChartSeriesOptions(series.Options)))
		options = append(options, variantOptions(cfg.Variant, minimum, maximum))
		if cfg.SeriesOptions.ItemStyle == nil && series.Options.ItemStyle == nil {
			themeSeriesItems = append(themeSeriesItems, seriesIndex)
		}
		options = append(options, internalinteractive.MergeSeriesOptions(cfg.SeriesOptions, series.Options)...)
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

	return internalinteractive.New(chartcomponents.KindInteractiveGauge, internalinteractive.RenderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export, ResponsiveWidth: internalinteractive.ResponsiveWidth(cfg.Width), ThemeSeriesItems: themeSeriesItems, GaugeScale: scaleJSON(cfg, minimum, maximum),
	})
}

func liquidGauge(cfg Config, minimum, maximum int) Instance {
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
	globalOptions = append(globalOptions, internalinteractive.ChartGlobalOptions(cfg.Options)...)
	if len(cfg.Style.Colors) > 0 {
		globalOptions = append(globalOptions, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
	chart.SetGlobalOptions(globalOptions...)
	series := cfg.Series[0]
	seriesOptions := internalinteractive.MergeSeriesOptions(cfg.SeriesOptions, series.Options)
	seriesOptions = append(seriesOptions, charts.WithLiquidChartOpts(opts.LiquidChart{Shape: string(cfg.Liquid.Shape)}))
	chart.AddSeries(series.Name, normalizeLiquidData(series.Data, minimum, maximum), seriesOptions...)

	return internalinteractive.New(chartcomponents.KindInteractiveGauge, internalinteractive.RenderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, ResponsiveWidth: internalinteractive.ResponsiveWidth(cfg.Width),
		Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		Details: liquidExactValues(cfg.Label, minimum, maximum, series.Data),
		Liquid:  liquidJSON(cfg.Liquid), ScriptReplacements: liquidScriptReplacements(cfg.Liquid),
	})
}

func normalizeLiquidData(data []Data, minimum, maximum int) []opts.LiquidData {
	result := make([]opts.LiquidData, len(data))
	span := float64(maximum - minimum)
	for index, point := range data {
		result[index] = opts.LiquidData{Name: point.Name, Value: (point.Value - float64(minimum)) / span}
	}
	return result
}

func liquidScriptReplacements(treatment LiquidTreatment) []internalinteractive.ScriptReplacement {
	fields := make([]string, 0, 9)
	if treatment.Shape != "" {
		fields = append(fields, fmt.Sprintf(`"shape":%q`, treatment.Shape))
	}
	if treatment.WaveLengthPercent != nil {
		fields = append(fields, fmt.Sprintf(`"waveLength":%q`, internalinteractive.Percentage(*treatment.WaveLengthPercent)))
	}
	if treatment.AmplitudePercent != nil {
		fields = append(fields, fmt.Sprintf(`"amplitude":%q`, internalinteractive.Percentage(*treatment.AmplitudePercent)))
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
	return []internalinteractive.ScriptReplacement{{Old: `"type":"liquidFill"`, New: `"type":"liquidFill",` + strings.Join(fields, ",")}}
}

type liquidPayload struct {
	Outline    *liquidOutlinePayload    `json:"outline,omitempty"`
	Background *liquidBackgroundPayload `json:"background,omitempty"`
	Label      *liquidPaintPayload      `json:"label,omitempty"`
	Style      *liquidStylePayload      `json:"style,omitempty"`
}

type liquidPaintPayload struct {
	Color string `json:"color,omitempty"`
	Class string `json:"class,omitempty"`
}

type liquidOutlinePayload struct {
	liquidPaintPayload
	Width float64 `json:"width,omitempty"`
}

type liquidBackgroundPayload struct {
	liquidPaintPayload
	BorderColor string  `json:"borderColor,omitempty"`
	BorderClass string  `json:"borderClass,omitempty"`
	BorderWidth float64 `json:"borderWidth,omitempty"`
}

type liquidStylePayload struct {
	liquidPaintPayload
	Opacity *float64 `json:"opacity,omitempty"`
}

func liquidJSON(treatment LiquidTreatment) string {
	payload := liquidPayload{}
	if value := treatment.Outline; value != nil {
		payload.Outline = &liquidOutlinePayload{liquidPaintPayload: liquidPaintPayload{Color: value.Color, Class: value.Class}, Width: value.Width}
	}
	if value := treatment.Background; value != nil {
		payload.Background = &liquidBackgroundPayload{liquidPaintPayload: liquidPaintPayload{Color: value.Color, Class: value.Class}, BorderColor: value.BorderColor, BorderClass: value.BorderClass, BorderWidth: value.BorderWidth}
	}
	if value := treatment.Label; value != nil {
		payload.Label = &liquidPaintPayload{Color: value.Color, Class: value.Class}
	}
	if value := treatment.Style; value != nil {
		payload.Style = &liquidStylePayload{liquidPaintPayload: liquidPaintPayload{Color: value.Color, Class: value.Class}, Opacity: value.Opacity}
	}
	data, _ := json.Marshal(payload)
	return string(data)
}

type scalePayload struct {
	Mode    string             `json:"mode"`
	Reverse bool               `json:"reverse,omitempty"`
	Stops   []scalePayloadStop `json:"stops,omitempty"`
}
type scalePayloadStop struct {
	Position float64 `json:"position"`
	Color    string  `json:"color,omitempty"`
	Class    string  `json:"class,omitempty"`
	Token    string  `json:"token,omitempty"`
}

func scaleJSON(cfg Config, minimum, maximum int) string {
	payload := scalePayload{Mode: string(cfg.Scale.Mode), Reverse: cfg.Scale.Reverse}
	if cfg.Scale.Mode == ScaleCustom {
		for _, stop := range cfg.Scale.Stops {
			payload.Stops = append(payload.Stops, scalePayloadStop{Position: (stop.Value - float64(minimum)) / float64(maximum-minimum), Color: stop.Color, Class: stop.Class})
		}
	} else if cfg.Scale.Mode == ScaleSingleColor {
		payload.Stops = []scalePayloadStop{{Position: 1, Color: cfg.Scale.Color, Class: cfg.Scale.Class}}
	} else {
		payload.Stops = []scalePayloadStop{{Position: .34, Token: "low"}, {Position: .67, Token: "mid"}, {Position: 1, Token: "high"}}
	}
	data, _ := json.Marshal(payload)
	return string(data)
}

func variantOptions(variant Variant, minimum, maximum int) charts.SeriesOpts {
	return func(series *charts.SingleSeries) {
		series.Min = minimum
		series.Max = maximum
		if variant == VariantProgress {
			series.Progress = &opts.Progress{
				Show: opts.Bool(true), Width: 6, RoundCap: opts.Bool(true), Clip: opts.Bool(true),
			}
			series.Pointer = &opts.Pointer{Show: opts.Bool(false)}
		}
	}
}

func resolvedRange(cfg Config) (int, int) {
	maximum := cfg.Max
	if maximum == 0 {
		maximum = 100
	}
	return cfg.Min, maximum
}

func validateConfig(cfg Config) error {
	if cfg.Label == "" {
		return fmt.Errorf("gauge chart label is required")
	}
	if cfg.Variant != VariantStandard && cfg.Variant != VariantProgress && cfg.Variant != VariantLiquid {
		return fmt.Errorf("gauge chart variant %q is not supported", cfg.Variant)
	}
	minimum, maximum := resolvedRange(cfg)
	if minimum >= maximum {
		return fmt.Errorf("gauge chart minimum must be less than maximum")
	}
	if cfg.Variant != VariantLiquid {
		if err := validateScale(cfg.Scale, minimum, maximum); err != nil {
			return err
		}
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("gauge chart series is required")
	}
	if err := validateLiquid(cfg); err != nil {
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

func validateLiquid(cfg Config) error {
	if cfg.Variant != VariantLiquid {
		if cfg.Liquid != (LiquidTreatment{}) {
			name := "standard"
			if cfg.Variant == VariantProgress {
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
	if cfg.Scale.Mode != ScaleThermal || cfg.Scale.Reverse || len(cfg.Scale.Stops) > 0 || cfg.Scale.Color != "" || cfg.Scale.Class != "" {
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
	if err := validateLiquidSeriesOptions(cfg.SeriesOptions); err != nil {
		return err
	}
	if err := validateLiquidSeriesOptions(cfg.Series[0].Options); err != nil {
		return err
	}
	validShapes := map[LiquidShape]bool{LiquidShapeCircle: true, LiquidShapeRect: true, LiquidShapeRoundRect: true, LiquidShapeTriangle: true, LiquidShapeDiamond: true, LiquidShapePin: true, LiquidShapeArrow: true}
	if !validShapes[cfg.Liquid.Shape] {
		return fmt.Errorf("liquid gauge shape %q is not supported", cfg.Liquid.Shape)
	}
	if value := cfg.Liquid.WaveLengthPercent; value != nil && (!internalinteractive.FiniteNumber(*value) || *value < 1 || *value > 100) {
		return fmt.Errorf("liquid gauge wave length percentage must be finite and between 1 and 100")
	}
	if value := cfg.Liquid.AmplitudePercent; value != nil && (!internalinteractive.FiniteNumber(*value) || *value < 0 || *value > 100) {
		return fmt.Errorf("liquid gauge amplitude percentage must be finite and between 0 and 100")
	}
	if value := cfg.Liquid.PhaseDegrees; value != nil && (!internalinteractive.FiniteNumber(*value) || *value < -360 || *value > 360) {
		return fmt.Errorf("liquid gauge phase must be finite and between -360 and 360 degrees")
	}
	if cfg.Liquid.Direction != LiquidDirectionRight && cfg.Liquid.Direction != LiquidDirectionLeft {
		return fmt.Errorf("liquid gauge direction %q is not supported", cfg.Liquid.Direction)
	}
	if value := cfg.Liquid.Outline; value != nil {
		if err := validateLiquidPaint("outline", value.Color, value.Class); err != nil {
			return err
		}
		if !internalinteractive.FiniteNumber(value.Width) || value.Width < 0 || value.Width > 64 {
			return fmt.Errorf("liquid gauge outline width must be finite and between 0 and 64 pixels")
		}
	}
	if value := cfg.Liquid.Background; value != nil {
		if err := validateLiquidPaint("background", value.Color, value.Class); err != nil {
			return err
		}
		if err := validateLiquidPaint("background border", value.BorderColor, value.BorderClass); err != nil {
			return err
		}
		if !internalinteractive.FiniteNumber(value.BorderWidth) || value.BorderWidth < 0 || value.BorderWidth > 64 {
			return fmt.Errorf("liquid gauge background border width must be finite and between 0 and 64 pixels")
		}
	}
	if value := cfg.Liquid.Label; value != nil {
		if err := validateLiquidPaint("label", value.Color, value.Class); err != nil {
			return err
		}
		if value.FontSize < 0 || value.FontSize > 256 {
			return fmt.Errorf("liquid gauge label font size must be between 0 and 256 pixels")
		}
	}
	if value := cfg.Liquid.Style; value != nil {
		if err := validateLiquidPaint("wave", value.Color, value.Class); err != nil {
			return err
		}
		if value.Opacity != nil && (!internalinteractive.FiniteNumber(*value.Opacity) || *value.Opacity < 0 || *value.Opacity > 1) {
			return fmt.Errorf("liquid gauge wave opacity must be finite and between 0 and 1")
		}
	}
	return nil
}

func validateLiquidSeriesOptions(options chart.SeriesOptions) error {
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

func validateLiquidPaint(name, color, class string) error {
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

func validateScale(scale Scale, minimum, maximum int) error {
	if scale.Mode != ScaleThermal && scale.Mode != ScaleCustom && scale.Mode != ScaleSingleColor {
		return fmt.Errorf("gauge chart scale mode %q is not supported", scale.Mode)
	}
	validPaint := func(color, class string) bool {
		return (strings.TrimSpace(color) == "") != (strings.TrimSpace(class) == "")
	}
	if scale.Mode == ScaleThermal {
		if len(scale.Stops) > 0 || scale.Color != "" || scale.Class != "" {
			return fmt.Errorf("gauge chart thermal scale does not accept custom paint")
		}
		return nil
	}
	if scale.Mode == ScaleSingleColor {
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
