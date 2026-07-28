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
)

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
	if cfg.Variant != GaugeVariantStandard && cfg.Variant != GaugeVariantProgress {
		return fmt.Errorf("gauge chart variant %q is not supported", cfg.Variant)
	}
	minimum, maximum := resolvedGaugeRange(cfg)
	if minimum >= maximum {
		return fmt.Errorf("gauge chart minimum must be less than maximum")
	}
	if err := validateGaugeScale(cfg.Scale, minimum, maximum); err != nil {
		return err
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("gauge chart series is required")
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
