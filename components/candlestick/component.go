package candlestick

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	chart "github.com/go-analyze/charts"
)

// Instance is a renderable server-side candlestick chart.
type Instance struct{ cfg Config }

// Candlestick returns a server-side SVG candlestick chart.
func Candlestick(cfg Config) Instance { return Instance{cfg: cfg} }

// Kind identifies the component as a candlestick chart.
func (Instance) Kind() chartcomponents.Kind { return chartcomponents.KindCandlestickChart }

// Render writes an accessible figure, SSR SVG, and adjacent exact-value table.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	svg, err := renderSVG(instance.cfg)
	if err != nil {
		return err
	}
	chart := candlestickTemplate(instance.cfg, templ.Raw(svg))
	return chartcontrol.Wrapper(chartcontrol.WrapperConfig{
		Label: instance.cfg.Label, Controls: instance.cfg.Controls, Export: instance.cfg.Export,
		Capability: chartcontrol.ExportCapabilityStaticSVG,
	}, chart).Render(ctx, writer)
}

func renderSVG(cfg Config) (string, error) {
	if err := cfg.validate(); err != nil {
		return "", err
	}
	painter := chart.NewPainter(chart.PainterOptions{OutputFormat: chart.ChartOutputSVG, Width: cfg.width(), Height: cfg.height(), Theme: tokenPalette()})
	if err := painter.CandlestickChart(candlestickOptions(cfg)); err != nil {
		return "", fmt.Errorf("render candlestick chart: %w", err)
	}
	data, err := painter.Bytes()
	if err != nil {
		return "", fmt.Errorf("encode candlestick chart SVG: %w", err)
	}
	return tokenizedSVG(string(data)), nil
}

func candlestickOptions(cfg Config) chart.CandlestickChartOption {
	data := make([]chart.OHLCData, len(cfg.Data))
	labels := make([]string, len(cfg.Data))
	for index, datum := range cfg.Data {
		labels[index] = datum.Label
		data[index] = chart.OHLCData{Open: datum.Open, High: datum.High, Low: datum.Low, Close: datum.Close}
	}
	options := chart.NewCandlestickOptionWithData(data)
	options.Theme = tokenPalette()
	options.Title.Text = cfg.Title
	options.XAxis.Labels = labels
	options.XAxis.Title = cfg.XAxis.Title
	options.YAxis[0].Title = cfg.YAxis.Title
	options.Legend.SeriesNames = []string{cfg.SeriesName}
	options.Legend.Show = chart.Ptr(true)
	options.SeriesList[0].Name = cfg.SeriesName
	return options
}

func tokenPalette() chart.ColorPalette {
	return chart.GetTheme(chart.ThemeLight).
		WithBackgroundColor(chart.Color{R: 1, G: 1, B: 1, A: 255}).
		WithXAxisColor(chart.Color{R: 2, G: 2, B: 2, A: 255}).
		WithYAxisColor(chart.Color{R: 2, G: 2, B: 2, A: 255}).
		WithAxisSplitLineColor(chart.Color{R: 3, G: 3, B: 3, A: 255}).
		WithTitleTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).
		WithMarkTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).
		WithLabelTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).
		WithLegendTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).
		WithXAxisTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).
		WithYAxisTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).
		WithTitleBorderColor(chart.Color{R: 2, G: 2, B: 2, A: 255}).
		WithLegendBorderColor(chart.Color{R: 2, G: 2, B: 2, A: 255})
}

func tokenizedSVG(svg string) string {
	return strings.NewReplacer(
		"rgb(1,1,1)", "var(--color-chart-surface)",
		"rgb(2,2,2)", "var(--color-chart-outline)",
		"rgb(3,3,3)", "var(--color-chart-grid)",
		"rgb(4,4,4)", "var(--color-chart-text)",
		"rgb(145,204,117)", "var(--color-chart-increasing)",
		"rgb(238,102,102)", "var(--color-chart-decreasing)",
		"'Roboto Medium',sans-serif", "var(--font-paragraph), sans-serif",
	).Replace(svg)
}

func formatValue(value float64) string { return strconv.FormatFloat(value, 'f', -1, 64) }
func direction(datum Datum) string {
	if datum.Close >= datum.Open {
		return "Increase"
	}
	return "Decrease"
}
func directionClass(datum Datum) string {
	if datum.Close >= datum.Open {
		return "goshtoso-charts-candlestick__direction--increasing"
	}
	return "goshtoso-charts-candlestick__direction--decreasing"
}

var _ chartcomponents.Component = Instance{}
