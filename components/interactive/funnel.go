package interactive

import (
	"fmt"
	"math"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// FunnelOrder controls funnel stage ordering.
type FunnelOrder string

const (
	// FunnelOrderDescending renders the largest stage first. It is the default.
	FunnelOrderDescending FunnelOrder = ""
	// FunnelOrderAscending renders the smallest stage first.
	FunnelOrderAscending FunnelOrder = "ascending"
	// FunnelOrderData preserves caller data order.
	FunnelOrderData FunnelOrder = "none"
)

// FunnelConfig describes an accessible, browser-rendered funnel chart.
//
// Values must be application-owned because the browser renderer serializes them.
type FunnelConfig struct {
	Label         string
	Caption       string
	Order         FunnelOrder
	Series        []FunnelSeries
	Width         string
	Height        string
	Options       ChartOptions
	SeriesOptions SeriesOptions
	Style         charttheme.Style
}

// FunnelSeries describes one named funnel series.
type FunnelSeries struct {
	Name    string
	Data    []FunnelData
	Options SeriesOptions
}

// FunnelData describes one named stage with a finite nonnegative value.
type FunnelData struct {
	Name  string
	Value float64
}

// Funnel builds a reusable interactive funnel component.
func Funnel(cfg FunnelConfig) Instance {
	if err := validateFunnelConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveFunnel, err)
	}

	chart := charts.NewFunnel()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, chartGlobalOptions(cfg.Options)...)
	if len(cfg.Style.Colors) > 0 {
		// Explicit component colors remain authoritative over escape-hatch options.
		globalOptions = append(globalOptions, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
	if cfg.Width != "" || cfg.Height != "" {
		globalOptions = append([]charts.GlobalOpts{
			charts.WithInitializationOpts(opts.Initialization{Width: cfg.Width, Height: cfg.Height}),
		}, globalOptions...)
	}
	chart.SetGlobalOptions(globalOptions...)

	for _, series := range cfg.Series {
		data := make([]opts.FunnelData, len(series.Data))
		for index, stage := range series.Data {
			data[index] = opts.FunnelData{Name: stage.Name, Value: stage.Value}
		}
		options := make([]charts.SeriesOpts, 0, 1+len(chartSeriesOptions(cfg.SeriesOptions))+len(chartSeriesOptions(series.Options)))
		options = append(options, funnelOrderOption(cfg.Order))
		options = append(options, mergeSeriesOptions(cfg.SeriesOptions, series.Options)...)
		chart.AddSeries(series.Name, data, options...)
	}

	return newInstance(chartcomponents.KindInteractiveFunnel, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, Animation: cfg.Options.Animation,
	})
}

func funnelOrderOption(order FunnelOrder) charts.SeriesOpts {
	return func(series *charts.SingleSeries) {
		if order == FunnelOrderDescending {
			series.Sort = "descending"
			return
		}
		series.Sort = string(order)
	}
}

func validateFunnelConfig(cfg FunnelConfig) error {
	if cfg.Label == "" {
		return fmt.Errorf("funnel chart label is required")
	}
	if cfg.Order != FunnelOrderDescending && cfg.Order != FunnelOrderAscending && cfg.Order != FunnelOrderData {
		return fmt.Errorf("funnel chart order %q is not supported", cfg.Order)
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("funnel chart series is required")
	}
	for seriesIndex, series := range cfg.Series {
		if series.Name == "" {
			return fmt.Errorf("funnel chart series %d name is required", seriesIndex)
		}
		if len(series.Data) == 0 {
			return fmt.Errorf("funnel chart series %q data is required", series.Name)
		}
		for dataIndex, stage := range series.Data {
			if stage.Name == "" {
				return fmt.Errorf("funnel chart series %q data point %d name is required", series.Name, dataIndex)
			}
			if math.IsNaN(stage.Value) || math.IsInf(stage.Value, 0) || stage.Value < 0 {
				return fmt.Errorf("funnel chart series %q data point %q value must be a finite nonnegative value", series.Name, stage.Name)
			}
		}
	}
	return nil
}
