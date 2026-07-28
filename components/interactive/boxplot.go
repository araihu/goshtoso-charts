package interactive

import (
	"fmt"
	"math"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// BoxPlotConfig describes an accessible, browser-rendered box plot.
//
// Each series contains one five-number summary per category. Values must be
// application-owned because the browser renderer serializes them.
type BoxPlotConfig struct {
	Label         string
	Caption       string
	Categories    []string
	Series        []BoxPlotSeries
	Width         string
	Height        string
	Options       ChartOptions
	SeriesOptions SeriesOptions
	Style         charttheme.Style
}

// BoxPlotSeries describes one named series across all Categories.
type BoxPlotSeries struct {
	Name    string
	Data    []BoxPlotData
	Options SeriesOptions
}

// BoxPlotData is an ordered five-number summary: minimum, first quartile,
// median, third quartile, and maximum.
type BoxPlotData struct {
	Name      string
	Min       float64
	Q1        float64
	Median    float64
	Q3        float64
	Max       float64
	Label     *LabelOptions
	ItemStyle *ItemStyle
	Emphasis  *EmphasisOptions
	Tooltip   *TooltipOptions
}

// BoxPlot builds a reusable interactive box plot component.
func BoxPlot(cfg BoxPlotConfig) Instance {
	if err := validateBoxPlotConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveBoxPlot, err)
	}

	chart := charts.NewBoxPlot()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
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
	chart.SetXAxis(cfg.Categories)
	palette := cfg.Style.ResolvedColors()
	themeSeriesItems := make([]int, 0, len(cfg.Series))
	for seriesIndex, series := range cfg.Series {
		data := make([]opts.BoxPlotData, len(series.Data))
		for index, summary := range series.Data {
			data[index] = opts.BoxPlotData{Name: summary.Name, Value: [5]float64{summary.Min, summary.Q1, summary.Median, summary.Q3, summary.Max}}
			if summary.Label != nil {
				label := rendererLabel(summary.Label)
				data[index].Label = &label
			}
			if summary.ItemStyle != nil {
				style := rendererItemStyle(summary.ItemStyle)
				data[index].ItemStyle = &style
			}
			if summary.Emphasis != nil {
				data[index].Emphasis = rendererEmphasis(summary.Emphasis)
			}
			if summary.Tooltip != nil {
				tooltip := rendererTooltip(summary.Tooltip)
				data[index].Tooltip = &tooltip
			}
		}
		options := make([]charts.SeriesOpts, 0, 1+len(chartSeriesOptions(cfg.SeriesOptions))+len(chartSeriesOptions(series.Options)))
		if cfg.SeriesOptions.ItemStyle == nil && series.Options.ItemStyle == nil {
			color := palette[seriesIndex%len(palette)]
			options = append(options, charts.WithItemStyleOpts(opts.ItemStyle{Color: color, BorderColor: color}))
			themeSeriesItems = append(themeSeriesItems, seriesIndex)
		}
		options = append(options, mergeSeriesOptions(cfg.SeriesOptions, series.Options)...)
		chart.AddSeries(series.Name, data, options...)
	}

	return newInstance(chartcomponents.KindInteractiveBoxPlot, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export, AxisLabelIntervals: axisLabelIntervals(cfg.Options), ThemeSeriesItems: themeSeriesItems,
	})
}

func validateBoxPlotConfig(cfg BoxPlotConfig) error {
	if err := validateChartOptions(cfg.Options); err != nil {
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
