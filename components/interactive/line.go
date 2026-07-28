package interactive

import (
	"fmt"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// LineConfig describes an accessible, browser-rendered line chart.
//
// Values must be application-owned because the browser renderer serializes them.
type LineConfig struct {
	Label         string
	Caption       string
	XAxis         []string
	Series        []LineSeries
	Width         string
	Height        string
	Options       ChartOptions
	SeriesOptions SeriesOptions
	Style         charttheme.Style
	Live          *LiveData
}

// LineSeries describes one named line series.
type LineSeries struct {
	Name    string
	Data    []LineData
	Options SeriesOptions
}

// LineData describes one finite line value and optional point symbol.
type LineData struct {
	Name       string
	Value      float64
	Symbol     string
	SymbolSize int
}

// Line builds a reusable interactive line component.
func Line(cfg LineConfig) Instance {
	if err := validateLineConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveLine, err)
	}

	chart := charts.NewLine()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, chartGlobalOptions(cfg.Options)...)
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
		data := make([]opts.LineData, len(series.Data))
		for index, point := range series.Data {
			data[index] = opts.LineData{Name: point.Name, Value: point.Value, Symbol: point.Symbol, SymbolSize: point.SymbolSize}
		}
		chart.AddSeries(series.Name, data, mergeSeriesOptions(cfg.SeriesOptions, series.Options)...)
	}

	return newInstance(chartcomponents.KindInteractiveLine, renderConfig{
		Label:              cfg.Label,
		Caption:            cfg.Caption,
		Chart:              chart,
		Style:              cfg.Style,
		Live:               cartesianLiveConfig(cfg.Live),
		Animation:          cfg.Options.Animation,
		Controls:           cfg.Options.Controls,
		Export:             cfg.Options.Export,
		AxisLabelIntervals: axisLabelIntervals(cfg.Options),
	})
}

func validateLineConfig(cfg LineConfig) error {
	if err := validateChartOptions(cfg.Options); err != nil {
		return err
	}
	if err := validateLiveData(cfg.Live); err != nil {
		return err
	}
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
		for dataIndex, point := range series.Data {
			if !finiteNumber(point.Value) {
				return fmt.Errorf("line chart series %q data point %d value must be finite", series.Name, dataIndex)
			}
		}
	}
	return nil
}
