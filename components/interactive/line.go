package interactive

import (
	"fmt"
	"time"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// LineConfig describes an accessible, browser-rendered line chart.
//
// Values must be application-owned because the browser renderer serializes them.
type LineConfig struct {
	Label   string
	Caption string
	XAxis   []string
	// TimeAxis selects a temporal x axis. It is mutually exclusive with XAxis.
	// LiveData remains categorical because CartesianSnapshot carries categories.
	TimeAxis      *LineTimeAxis
	Series        []LineSeries
	Width         string
	Height        string
	Options       ChartOptions
	SeriesOptions SeriesOptions
	Style         charttheme.Style
	Live          *LiveData
}

// LineTimeAxis defines ordered instants and a required inclusive lower bound.
// Values and Minimum use time.Time so callers never supply renderer values.
type LineTimeAxis struct {
	Values      []time.Time
	Minimum     time.Time
	// SplitNumber recommends readable temporal tick density. Zero uses the
	// responsive default of four segments; the private renderer also hides
	// any remaining overlap while retaining endpoint labels.
	SplitNumber int
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
	if cfg.TimeAxis != nil {
		globalOptions = append(globalOptions, charts.WithXAxisOpts(opts.XAxis{
			Type:        "time",
			Min:         lineTimeValue(cfg.TimeAxis.Minimum),
			SplitNumber: lineTimeSplitNumber(*cfg.TimeAxis),
			AxisLabel: &opts.AxisLabel{
				HideOverlap:   opts.Bool(true),
				ShowMinLabel:  opts.Bool(true),
				ShowMaxLabel:  opts.Bool(true),
				AlignMinLabel: "left",
				AlignMaxLabel: "right",
			},
		}), charts.WithGridOpts(opts.Grid{Top: "18%", Bottom: "15%", ContainLabel: opts.Bool(true)}))
	}
	chart.SetGlobalOptions(globalOptions...)
	if cfg.TimeAxis == nil {
		chart.SetXAxis(cfg.XAxis)
	}
	for _, series := range cfg.Series {
		data := make([]opts.LineData, len(series.Data))
		for index, point := range series.Data {
			value := any(point.Value)
			if cfg.TimeAxis != nil {
				value = []any{lineTimeValue(cfg.TimeAxis.Values[index]), point.Value}
			}
			data[index] = opts.LineData{Name: point.Name, Value: value, Symbol: point.Symbol, SymbolSize: point.SymbolSize}
		}
		chart.AddSeries(series.Name, data, mergeSeriesOptions(cfg.SeriesOptions, series.Options)...)
	}

	return newInstance(chartcomponents.KindInteractiveLine, renderConfig{
		Label:              cfg.Label,
		Caption:            cfg.Caption,
		Chart:              chart,
		ResponsiveWidth:    responsiveWidth(cfg.Width),
		Style:              cfg.Style,
		Live:               cartesianLiveConfig(cfg.Live),
		Details:            lineTimeExactValues(cfg),
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
	if cfg.TimeAxis != nil && len(cfg.XAxis) != 0 {
		return fmt.Errorf("line chart x axis and time axis are mutually exclusive")
	}
	if cfg.TimeAxis == nil && len(cfg.XAxis) == 0 {
		return fmt.Errorf("line chart x axis is required")
	}
	if cfg.TimeAxis != nil {
		if cfg.Live != nil {
			return fmt.Errorf("line chart live data supports categorical x axis only")
		}
		if err := validateLineTimeAxis(*cfg.TimeAxis); err != nil {
			return err
		}
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("line chart series is required")
	}
	for index, series := range cfg.Series {
		if series.Name == "" {
			return fmt.Errorf("line chart series %d name is required", index)
		}
		axisLen := len(cfg.XAxis)
		if cfg.TimeAxis != nil {
			axisLen = len(cfg.TimeAxis.Values)
		}
		if len(series.Data) != axisLen {
			return fmt.Errorf("line chart series %q has %d data points for %d x-axis values", series.Name, len(series.Data), axisLen)
		}
		for dataIndex, point := range series.Data {
			if !finiteNumber(point.Value) {
				return fmt.Errorf("line chart series %q data point %d value must be finite", series.Name, dataIndex)
			}
		}
	}
	return nil
}

func validateLineTimeAxis(axis LineTimeAxis) error {
	if axis.Minimum.IsZero() {
		return fmt.Errorf("line chart time axis minimum is required")
	}
	if len(axis.Values) == 0 {
		return fmt.Errorf("line chart time axis values are required")
	}
	if axis.SplitNumber < 0 {
		return fmt.Errorf("line chart time axis split number must be nonnegative")
	}
	minimum := axis.Minimum.UTC()
	previous := time.Time{}
	for index, value := range axis.Values {
		if value.IsZero() {
			return fmt.Errorf("line chart time axis value %d is required", index)
		}
		value = value.UTC()
		if value.Before(minimum) {
			return fmt.Errorf("line chart time axis value %d precedes minimum", index)
		}
		if index > 0 && !value.After(previous) {
			return fmt.Errorf("line chart time axis values must be strictly chronological")
		}
		previous = value
	}
	return nil
}

func lineTimeSplitNumber(axis LineTimeAxis) int {
	if axis.SplitNumber == 0 {
		return 4
	}
	return axis.SplitNumber
}

func lineTimeValue(value time.Time) string { return value.UTC().Format(time.RFC3339) }
