package echarts

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
// application-owned because go-echarts serializes them into executable JavaScript.
type BoxPlotConfig struct {
	Label         string
	Caption       string
	Categories    []string
	Series        []BoxPlotSeries
	Width         string
	Height        string
	GlobalOptions []charts.GlobalOpts
	SeriesOptions []charts.SeriesOpts
	Style         charttheme.Style
}

// BoxPlotSeries describes one named series across all Categories.
type BoxPlotSeries struct {
	Name    string
	Data    []BoxPlotData
	Options []charts.SeriesOpts
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
	Label     *opts.Label
	ItemStyle *opts.ItemStyle
	Emphasis  *opts.Emphasis
	Tooltip   *opts.Tooltip
}

// BoxPlot builds a reusable go-echarts box plot component.
func BoxPlot(cfg BoxPlotConfig) Instance {
	if err := validateBoxPlotConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindEChartsBoxPlot, err)
	}

	chart := charts.NewBoxPlot()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
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
	chart.SetXAxis(cfg.Categories)
	for _, series := range cfg.Series {
		data := make([]opts.BoxPlotData, len(series.Data))
		for index, summary := range series.Data {
			data[index] = opts.BoxPlotData{
				Name: summary.Name, Value: [5]float64{summary.Min, summary.Q1, summary.Median, summary.Q3, summary.Max},
				Label: summary.Label, ItemStyle: summary.ItemStyle, Emphasis: summary.Emphasis, Tooltip: summary.Tooltip,
			}
		}
		options := make([]charts.SeriesOpts, 0, len(cfg.SeriesOptions)+len(series.Options))
		options = append(options, cfg.SeriesOptions...)
		options = append(options, series.Options...)
		chart.AddSeries(series.Name, data, options...)
	}

	return newInstance(chartcomponents.KindEChartsBoxPlot, Config{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style,
	})
}

func validateBoxPlotConfig(cfg BoxPlotConfig) error {
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
