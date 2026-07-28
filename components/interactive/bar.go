package interactive

import (
	"fmt"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// BarConfig describes an accessible, browser-rendered bar chart.
//
// Values must be application-owned because the browser renderer serializes them.
type BarConfig struct {
	Label         string
	Caption       string
	XAxis         []string
	Series        []BarSeries
	Width         string
	Height        string
	Options       ChartOptions
	SeriesOptions SeriesOptions
	Style         charttheme.Style
	Live          *LiveData
}

// BarSeries describes one named bar series.
type BarSeries struct {
	Name    string
	Data    []BarData
	Options SeriesOptions
}

// BarData describes one finite bar value and optional per-point presentation.
type BarData struct {
	Name      string
	Value     float64
	Label     *LabelOptions
	ItemStyle *ItemStyle
	Tooltip   *TooltipOptions
}

// Bar builds a reusable interactive bar component.
func Bar(cfg BarConfig) Instance {
	if err := validateBarConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveBar, err)
	}

	chart := charts.NewBar()
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
		data := make([]opts.BarData, len(series.Data))
		for index, point := range series.Data {
			data[index] = opts.BarData{Name: point.Name, Value: point.Value}
			if point.Label != nil {
				label := rendererLabel(point.Label)
				data[index].Label = &label
			}
			if point.ItemStyle != nil {
				style := rendererItemStyle(point.ItemStyle)
				data[index].ItemStyle = &style
			}
			if point.Tooltip != nil {
				tooltip := rendererTooltip(point.Tooltip)
				data[index].Tooltip = &tooltip
			}
		}
		chart.AddSeries(series.Name, data, mergeSeriesOptions(cfg.SeriesOptions, series.Options)...)
	}

	return newInstance(chartcomponents.KindInteractiveBar, renderConfig{
		Label:              cfg.Label,
		Caption:            cfg.Caption,
		Chart:              chart,
		Style:              cfg.Style,
		Live:               cartesianLiveConfig(cfg.Live),
		Animation:          cfg.Options.Animation,
		AxisLabelIntervals: axisLabelIntervals(cfg.Options),
	})
}

func validateBarConfig(cfg BarConfig) error {
	if err := validateChartOptions(cfg.Options); err != nil {
		return err
	}
	if err := validateLiveData(cfg.Live); err != nil {
		return err
	}
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
		for dataIndex, point := range series.Data {
			if !finiteNumber(point.Value) {
				return fmt.Errorf("bar chart series %q data point %d value must be finite", series.Name, dataIndex)
			}
		}
	}
	return nil
}
