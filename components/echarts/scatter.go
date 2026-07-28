package echarts

import (
	"fmt"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// ScatterConfig describes an accessible, browser-rendered scatter chart.
//
// Category axes use XAxis and one scalar value per category. Value axes leave
// XAxis empty and use two-value [x, y] coordinates in ScatterData.Value.
// Values must be application-owned because go-echarts serializes them into
// executable JavaScript.
type ScatterConfig struct {
	Label         string
	Caption       string
	XAxisType     CartesianAxisType
	XAxis         []string
	Series        []ScatterSeries
	Width         string
	Height        string
	GlobalOptions []charts.GlobalOpts
	SeriesOptions []charts.SeriesOpts
	Style         charttheme.Style
}

// ScatterSeries describes one named scatter series.
type ScatterSeries struct {
	Name    string
	Data    []opts.ScatterData
	Options []charts.SeriesOpts
}

// Scatter builds a reusable go-echarts scatter component.
func Scatter(cfg ScatterConfig) Instance {
	if err := validateScatterConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindEChartsScatter, err)
	}

	chart := charts.NewScatter()
	globalOptions := []charts.GlobalOpts{
		charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())),
		charts.WithXAxisOpts(opts.XAxis{Type: string(resolvedCartesianAxisType(cfg.XAxisType))}),
	}
	globalOptions = append(globalOptions, cfg.GlobalOptions...)
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
	if resolvedCartesianAxisType(cfg.XAxisType) == CartesianAxisCategory {
		chart.SetXAxis(cfg.XAxis)
	}
	for _, series := range cfg.Series {
		options := make([]charts.SeriesOpts, 0, len(cfg.SeriesOptions)+len(series.Options))
		options = append(options, cfg.SeriesOptions...)
		options = append(options, series.Options...)
		chart.AddSeries(series.Name, series.Data, options...)
	}

	return newInstance(chartcomponents.KindEChartsScatter, Config{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style,
	})
}

func validateScatterConfig(cfg ScatterConfig) error {
	axisType := resolvedCartesianAxisType(cfg.XAxisType)
	if axisType != CartesianAxisCategory && axisType != CartesianAxisValue {
		return fmt.Errorf("scatter chart x axis type %q is not supported", cfg.XAxisType)
	}
	if axisType == CartesianAxisCategory && len(cfg.XAxis) == 0 {
		return fmt.Errorf("scatter chart x axis is required for category mode")
	}
	if axisType == CartesianAxisValue && len(cfg.XAxis) != 0 {
		return fmt.Errorf("scatter chart x axis categories are not allowed for value mode")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("scatter chart series is required")
	}
	for index, series := range cfg.Series {
		if series.Name == "" {
			return fmt.Errorf("scatter chart series %d name is required", index)
		}
		if len(series.Data) == 0 {
			return fmt.Errorf("scatter chart series %q data is required", series.Name)
		}
		if axisType == CartesianAxisCategory && len(series.Data) != len(cfg.XAxis) {
			return fmt.Errorf("scatter chart series %q has %d data points for %d x-axis categories", series.Name, len(series.Data), len(cfg.XAxis))
		}
		if axisType == CartesianAxisValue {
			for dataIndex, point := range series.Data {
				if !isCartesianCoordinate(point.Value) {
					return fmt.Errorf("scatter chart series %q data point %d must contain a numeric [x, y] coordinate", series.Name, dataIndex)
				}
			}
		}
	}
	return nil
}
