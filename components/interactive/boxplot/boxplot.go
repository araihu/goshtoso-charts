// Package boxplot provides the canonical interactive box plot API.
//
// Box-plot-specific types and implementation live here; shared renderer-neutral
// options remain in components/chart.
package boxplot

import (
	"fmt"
	"math"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// Instance is the renderer-neutral chart instance returned by BoxPlot.
type Instance = chart.Instance

// Config describes an accessible, browser-rendered box plot.
//
// Each series contains one five-number summary per category. Values must be
// application-owned because the browser renderer serializes them.
type Config struct {
	Label         string
	Caption       string
	Categories    []string
	Series        []Series
	Width         string
	Height        string
	Options       chart.ChartOptions
	SeriesOptions chart.SeriesOptions
	Style         charttheme.Style
}

// Series describes one named series across all Categories.
type Series struct {
	Name    string
	Data    []Data
	Options chart.SeriesOptions
}

// Data is an ordered five-number summary: minimum, first quartile, median,
// third quartile, and maximum.
type Data struct {
	Name      string
	Min       float64
	Q1        float64
	Median    float64
	Q3        float64
	Max       float64
	Label     *chart.LabelOptions
	ItemStyle *chart.ItemStyle
	Emphasis  *chart.EmphasisOptions
	Tooltip   *chart.TooltipOptions
}

// BoxPlot builds a reusable interactive box plot component.
func BoxPlot(cfg Config) Instance {
	if err := validateConfig(cfg); err != nil {
		return internalinteractive.Invalid(chartcomponents.KindInteractiveBoxPlot, err)
	}

	boxPlot := charts.NewBoxPlot()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, internalinteractive.ChartGlobalOptions(cfg.Options)...)
	// Explicit component colors remain authoritative over escape-hatch options.
	if len(cfg.Style.Colors) > 0 {
		globalOptions = append(globalOptions, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
	if cfg.Width != "" || cfg.Height != "" {
		globalOptions = append([]charts.GlobalOpts{
			charts.WithInitializationOpts(opts.Initialization{Width: cfg.Width, Height: cfg.Height}),
		}, globalOptions...)
	}
	boxPlot.SetGlobalOptions(globalOptions...)
	boxPlot.SetXAxis(cfg.Categories)
	palette := cfg.Style.ResolvedColors()
	themeSeriesItems := make([]int, 0, len(cfg.Series))
	for seriesIndex, series := range cfg.Series {
		data := make([]opts.BoxPlotData, len(series.Data))
		for index, summary := range series.Data {
			data[index] = opts.BoxPlotData{Name: summary.Name, Value: [5]float64{summary.Min, summary.Q1, summary.Median, summary.Q3, summary.Max}}
			if summary.Label != nil {
				label := internalinteractive.RendererLabel(summary.Label)
				data[index].Label = &label
			}
			if summary.ItemStyle != nil {
				style := internalinteractive.RendererItemStyle(summary.ItemStyle)
				data[index].ItemStyle = &style
			}
			if summary.Emphasis != nil {
				data[index].Emphasis = internalinteractive.RendererEmphasis(summary.Emphasis)
			}
			if summary.Tooltip != nil {
				tooltip := internalinteractive.RendererTooltip(summary.Tooltip)
				data[index].Tooltip = &tooltip
			}
		}
		options := make([]charts.SeriesOpts, 0, 1+len(internalinteractive.ChartSeriesOptions(cfg.SeriesOptions))+len(internalinteractive.ChartSeriesOptions(series.Options)))
		if cfg.SeriesOptions.ItemStyle == nil && series.Options.ItemStyle == nil {
			color := palette[seriesIndex%len(palette)]
			options = append(options, charts.WithItemStyleOpts(opts.ItemStyle{Color: color, BorderColor: color}))
			themeSeriesItems = append(themeSeriesItems, seriesIndex)
		}
		options = append(options, internalinteractive.MergeSeriesOptions(cfg.SeriesOptions, series.Options)...)
		boxPlot.AddSeries(series.Name, data, options...)
	}

	return internalinteractive.New(chartcomponents.KindInteractiveBoxPlot, internalinteractive.RenderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: boxPlot, Style: cfg.Style, Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export, ResponsiveWidth: internalinteractive.ResponsiveWidth(cfg.Width), AxisLabelIntervals: internalinteractive.AxisLabelIntervals(cfg.Options), ThemeSeriesItems: themeSeriesItems,
	})
}

func validateConfig(cfg Config) error {
	if err := internalinteractive.ValidateChartOptions(cfg.Options); err != nil {
		return err
	}
	if cfg.Label == "" {
		return fmt.Errorf("box plot label is required")
	}
	if len(cfg.Categories) == 0 {
		return fmt.Errorf("box plot categories are required")
	}
	for index, category := range cfg.Categories {
		if category == "" {
			return fmt.Errorf("box plot category %d name is required", index)
		}
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("box plot series is required")
	}
	for seriesIndex, series := range cfg.Series {
		if series.Name == "" {
			return fmt.Errorf("box plot series %d name is required", seriesIndex)
		}
		if len(series.Data) == 0 {
			return fmt.Errorf("box plot series %q data is required", series.Name)
		}
		if len(series.Data) != len(cfg.Categories) {
			return fmt.Errorf("box plot series %q has %d summaries for %d categories", series.Name, len(series.Data), len(cfg.Categories))
		}
		for dataIndex, summary := range series.Data {
			values := [...]float64{summary.Min, summary.Q1, summary.Median, summary.Q3, summary.Max}
			for _, value := range values {
				if math.IsNaN(value) || math.IsInf(value, 0) {
					return fmt.Errorf("box plot series %q summary %d values must be finite", series.Name, dataIndex)
				}
			}
			if summary.Min > summary.Q1 || summary.Q1 > summary.Median || summary.Median > summary.Q3 || summary.Q3 > summary.Max {
				return fmt.Errorf("box plot series %q summary %d values must be ordered min, q1, median, q3, max", series.Name, dataIndex)
			}
		}
	}
	return nil
}
