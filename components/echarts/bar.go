package echarts

import (
	"fmt"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// BarConfig describes an accessible, browser-rendered bar chart.
//
// Values must be application-owned. go-echarts serializes chart values into
// executable JavaScript.
type BarConfig struct {
	Label         string
	Caption       string
	XAxis         []string
	Series        []BarSeries
	Width         string
	Height        string
	GlobalOptions []charts.GlobalOpts
	SeriesOptions []charts.SeriesOpts
	Style         charttheme.Style
}

// BarSeries describes one named bar series.
type BarSeries struct {
	Name    string
	Data    []opts.BarData
	Options []charts.SeriesOpts
}

// Bar builds a reusable go-echarts bar component.
func Bar(cfg BarConfig) Instance {
	if err := validateBarConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindEChartsBar, err)
	}

	chart := charts.NewBar()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, cfg.GlobalOptions...)
	if len(cfg.Style.Colors) > 0 {
		globalOptions = append(globalOptions, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
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

	return newInstance(chartcomponents.KindEChartsBar, Config{
		Label:   cfg.Label,
		Caption: cfg.Caption,
		Chart:   chart,
		Style:   cfg.Style,
	})
}

func validateBarConfig(cfg BarConfig) error {
	if len(cfg.XAxis) == 0 {
		return fmt.Errorf("bar chart x axis is required")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("bar chart series is required")
	}
	for index, series := range cfg.Series {
		if series.Name == "" {
			return fmt.Errorf("bar chart series %d name is required", index)
		}
		if len(series.Data) != len(cfg.XAxis) {
			return fmt.Errorf("bar chart series %q has %d values for %d x-axis categories", series.Name, len(series.Data), len(cfg.XAxis))
		}
	}
	return nil
}
