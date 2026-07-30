// Package funnel provides the canonical interactive funnel API.
//
// Funnel-specific types and implementation live here; shared renderer-neutral
// options remain in components/chart.
package funnel

import (
	"fmt"
	"math"
	"strconv"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// Instance is the renderer-neutral chart instance returned by Funnel.
type Instance = chart.Instance

// Order controls funnel stage ordering.
type Order string

const (
	// OrderDescending renders the largest stage first. It is the default.
	OrderDescending Order = ""
	// OrderAscending renders the smallest stage first.
	OrderAscending Order = "ascending"
	// OrderData preserves caller data order.
	OrderData Order = "none"
)

// Config describes an accessible, browser-rendered funnel chart.
//
// Values must be application-owned because the browser renderer serializes them.
type Config struct {
	Label         string
	Caption       string
	Order         Order
	Series        []Series
	Width         string
	Height        string
	Options       chart.ChartOptions
	SeriesOptions chart.SeriesOptions
	Style         charttheme.Style
}

// Series describes one named funnel series.
type Series struct {
	Name    string
	Data    []Data
	Options chart.SeriesOptions
}

// Data describes one named stage with a finite nonnegative value.
type Data struct {
	Name  string
	Value float64
}

// Funnel builds a reusable interactive funnel component.
func Funnel(cfg Config) Instance {
	if err := validateConfig(cfg); err != nil {
		return internalinteractive.Invalid(chartcomponents.KindInteractiveFunnel, err)
	}

	funnelChart := charts.NewFunnel()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, internalinteractive.ChartGlobalOptions(cfg.Options)...)
	if len(cfg.Style.Colors) > 0 {
		// Explicit component colors remain authoritative over escape-hatch options.
		globalOptions = append(globalOptions, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
	if cfg.Width != "" || cfg.Height != "" {
		globalOptions = append([]charts.GlobalOpts{
			charts.WithInitializationOpts(opts.Initialization{Width: cfg.Width, Height: cfg.Height}),
		}, globalOptions...)
	}
	funnelChart.SetGlobalOptions(globalOptions...)

	for _, series := range cfg.Series {
		data := make([]opts.FunnelData, len(series.Data))
		for index, stage := range series.Data {
			data[index] = opts.FunnelData{Name: stage.Name, Value: stage.Value}
		}
		options := make([]charts.SeriesOpts, 0, 1+len(internalinteractive.ChartSeriesOptions(cfg.SeriesOptions))+len(internalinteractive.ChartSeriesOptions(series.Options)))
		options = append(options, orderOption(cfg.Order))
		options = append(options, internalinteractive.MergeSeriesOptions(cfg.SeriesOptions, series.Options)...)
		funnelChart.AddSeries(series.Name, data, options...)
	}

	return internalinteractive.New(chartcomponents.KindInteractiveFunnel, internalinteractive.RenderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: funnelChart, Style: cfg.Style, Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export, ResponsiveWidth: internalinteractive.ResponsiveWidth(cfg.Width),
		Details: funnelExactValues(cfg.Label, detailRows(cfg.Series)),
	})
}

type valueRow struct {
	Series string
	Stage  string
	Value  string
}

func detailRows(series []Series) []valueRow {
	rows := make([]valueRow, 0)
	for _, current := range series {
		for _, stage := range current.Data {
			rows = append(rows, valueRow{
				Series: current.Name,
				Stage:  stage.Name,
				Value:  strconv.FormatFloat(stage.Value, 'f', -1, 64),
			})
		}
	}
	return rows
}

func orderOption(order Order) charts.SeriesOpts {
	return func(series *charts.SingleSeries) {
		if order == OrderDescending {
			series.Sort = "descending"
			return
		}
		series.Sort = string(order)
	}
}

func validateConfig(cfg Config) error {
	if err := internalinteractive.ValidateChartOptions(cfg.Options); err != nil {
		return err
	}
	if cfg.Label == "" {
		return fmt.Errorf("funnel chart label is required")
	}
	if cfg.Order != OrderDescending && cfg.Order != OrderAscending && cfg.Order != OrderData {
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
