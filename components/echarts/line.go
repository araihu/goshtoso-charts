package echarts

import (
	"fmt"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// LineConfig describes an accessible, browser-rendered line chart.
//
// Values must be application-owned. go-echarts serializes chart values into
// executable JavaScript.
type LineConfig struct {
	Label         string
	Caption       string
	XAxis         []string
	Series        []LineSeries
	Width         string
	Height        string
	GlobalOptions []charts.GlobalOpts
	SeriesOptions []charts.SeriesOpts
}

// LineSeries describes one named line series.
type LineSeries struct {
	Name    string
	Data    []opts.LineData
	Options []charts.SeriesOpts
}

// Line builds a reusable go-echarts line component.
func Line(cfg LineConfig) Instance {
	if err := validateLineConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindEChartsLine, err)
	}

	chart := charts.NewLine()
	globalOptions := cfg.GlobalOptions
	if cfg.Width != "" || cfg.Height != "" {
		globalOptions = append([]charts.GlobalOpts{
			charts.WithInitializationOpts(opts.Initialization{
				Width:  cfg.Width,
				Height: cfg.Height,
			}),
		}, globalOptions...)
	}
	chart.SetGlobalOptions(globalOptions...)
	chart.SetXAxis(cfg.XAxis)
	for _, series := range cfg.Series {
		options := make([]charts.SeriesOpts, 0, len(cfg.SeriesOptions)+len(series.Options))
		options = append(options, cfg.SeriesOptions...)
		options = append(options, series.Options...)
		chart.AddSeries(series.Name, series.Data, options...)
	}

	return newInstance(chartcomponents.KindEChartsLine, Config{
		Label:   cfg.Label,
		Caption: cfg.Caption,
		Chart:   chart,
	})
}

func validateLineConfig(cfg LineConfig) error {
	if len(cfg.XAxis) == 0 {
		return fmt.Errorf("line chart x axis is required")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("line chart series is required")
	}
	for index, series := range cfg.Series {
		if series.Name == "" {
			return fmt.Errorf("line chart series %d name is required", index)
		}
		if len(series.Data) != len(cfg.XAxis) {
			return fmt.Errorf("line chart series %q has %d data points for %d x-axis values", series.Name, len(series.Data), len(cfg.XAxis))
		}
	}
	return nil
}
