package interactive

import (
	"fmt"
	"math"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// RadarConfig describes an accessible, browser-rendered radar chart.
//
// Values must be application-owned because the browser renderer serializes them.
type RadarConfig struct {
	Label         string
	Caption       string
	Indicators    []RadarIndicator
	Series        []RadarSeries
	Width         string
	Height        string
	Options       ChartOptions
	SeriesOptions SeriesOptions
	Style         charttheme.Style
}

// RadarIndicator describes one named radar dimension and its positive maximum.
type RadarIndicator struct {
	Name string
	Max  float32
}

// RadarSeries describes one named radar series.
type RadarSeries struct {
	Name    string
	Data    []RadarData
	Options SeriesOptions
}

// RadarData describes one named vector whose values align with Indicators.
type RadarData struct {
	Name   string
	Values []float64
}

// Radar builds a reusable interactive radar component.
func Radar(cfg RadarConfig) Instance {
	if err := validateRadarConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveRadar, err)
	}

	indicators := make([]*opts.Indicator, len(cfg.Indicators))
	for index, indicator := range cfg.Indicators {
		indicators[index] = &opts.Indicator{Name: indicator.Name, Max: indicator.Max}
	}

	chart := charts.NewRadar()
	globalOptions := []charts.GlobalOpts{
		charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())),
		charts.WithRadarComponentOpts(opts.RadarComponent{Indicator: indicators}),
	}
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
	for _, series := range cfg.Series {
		data := make([]opts.RadarData, len(series.Data))
		for index, point := range series.Data {
			data[index] = opts.RadarData{Name: point.Name, Value: point.Values}
		}
		chart.AddSeries(series.Name, data, mergeSeriesOptions(cfg.SeriesOptions, series.Options)...)
	}

	return newInstance(chartcomponents.KindInteractiveRadar, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export,
	})
}

func validateRadarConfig(cfg RadarConfig) error {
	if cfg.Label == "" {
		return fmt.Errorf("radar chart label is required")
	}
	if len(cfg.Indicators) == 0 {
		return fmt.Errorf("radar chart indicators are required")
	}
	for index, indicator := range cfg.Indicators {
		if indicator.Name == "" {
			return fmt.Errorf("radar chart indicator %d name is required", index)
		}
		if math.IsNaN(float64(indicator.Max)) || math.IsInf(float64(indicator.Max), 0) || indicator.Max <= 0 {
			return fmt.Errorf("radar chart indicator %q maximum must be positive", indicator.Name)
		}
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("radar chart series is required")
	}
	for seriesIndex, series := range cfg.Series {
		if series.Name == "" {
			return fmt.Errorf("radar chart series %d name is required", seriesIndex)
		}
		if len(series.Data) == 0 {
			return fmt.Errorf("radar chart series %q data is required", series.Name)
		}
		for dataIndex, point := range series.Data {
			if point.Name == "" {
				return fmt.Errorf("radar chart series %q data point %d name is required", series.Name, dataIndex)
			}
			if len(point.Values) == 0 {
				return fmt.Errorf("radar chart series %q data %q values are required", series.Name, point.Name)
			}
			if len(point.Values) != len(cfg.Indicators) {
				return fmt.Errorf("radar chart series %q data %q has %d values for %d indicators", series.Name, point.Name, len(point.Values), len(cfg.Indicators))
			}
			for valueIndex, value := range point.Values {
				if math.IsNaN(value) || math.IsInf(value, 0) {
					return fmt.Errorf("radar chart series %q data %q value %d must be finite", series.Name, point.Name, valueIndex)
				}
			}
		}
	}
	return nil
}
