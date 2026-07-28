package bar

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	chart "github.com/go-analyze/charts"
)

// Instance is a renderable server-side bar chart.
type Instance struct{ cfg Config }

// Bar returns a server-side SVG bar chart.
func Bar(cfg Config) Instance { return Instance{cfg: cfg} }

// Kind identifies the component as a bar chart.
func (Instance) Kind() chartcomponents.Kind { return chartcomponents.KindBarChart }

// Render writes accessible figure and SVG without browser rendering or hydration.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	svg, err := renderSVG(instance.cfg)
	if err != nil {
		return err
	}
	return barTemplate(instance.cfg, templ.Raw(svg)).Render(ctx, writer)
}

func renderSVG(cfg Config) (string, error) {
	if err := cfg.validate(); err != nil {
		return "", err
	}
	values := make([][]float64, 0, len(cfg.Series))
	names := make([]string, 0, len(cfg.Series))
	for _, series := range cfg.Series {
		values, names = append(values, series.Values), append(names, series.Name)
	}
	options := chart.NewBarChartOptionWithData(values)
	options.CategoryAxis.Labels = cfg.Labels
	options.Legend.SeriesNames = names
	options.StackSeries = chart.Ptr(cfg.Stacked)
	options.Theme = tokenPalette()
	painter := chart.NewPainter(chart.PainterOptions{OutputFormat: chart.ChartOutputSVG, Width: cfg.width(), Height: cfg.height(), Theme: tokenPalette()})
	if err := painter.BarChart(options); err != nil {
		return "", fmt.Errorf("render bar chart: %w", err)
	}
	data, err := painter.Bytes()
	if err != nil {
		return "", fmt.Errorf("encode bar chart SVG: %w", err)
	}
	return tokenizedSVG(string(data)), nil
}

func tokenPalette() chart.ColorPalette {
	return chart.GetTheme(chart.ThemeLight).
		WithBackgroundColor(chart.Color{R: 1, G: 1, B: 1, A: 255}).
		WithXAxisColor(chart.Color{R: 2, G: 2, B: 2, A: 255}).WithYAxisColor(chart.Color{R: 2, G: 2, B: 2, A: 255}).
		WithAxisSplitLineColor(chart.Color{R: 3, G: 3, B: 3, A: 255}).
		WithTitleTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).WithMarkTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).WithLabelTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).WithLegendTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).WithXAxisTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).WithYAxisTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).
		WithTitleBorderColor(chart.Color{R: 2, G: 2, B: 2, A: 255}).WithLegendBorderColor(chart.Color{R: 2, G: 2, B: 2, A: 255}).
		WithSeriesColors([]chart.Color{{R: 5, G: 5, B: 5, A: 255}, {R: 6, G: 6, B: 6, A: 255}, {R: 7, G: 7, B: 7, A: 255}, {R: 8, G: 8, B: 8, A: 255}, {R: 9, G: 9, B: 9, A: 255}, {R: 10, G: 10, B: 10, A: 255}})
}

func tokenizedSVG(svg string) string {
	replacer := strings.NewReplacer("rgb(1,1,1)", "var(--goshtoso-charts-surface)", "rgb(2,2,2)", "var(--goshtoso-charts-outline)", "rgb(3,3,3)", "var(--goshtoso-charts-grid)", "rgb(4,4,4)", "var(--goshtoso-charts-text)", "rgb(5,5,5)", "var(--goshtoso-charts-primary)", "rgb(6,6,6)", "var(--goshtoso-charts-success)", "rgb(7,7,7)", "var(--goshtoso-charts-secondary)", "rgb(8,8,8)", "var(--goshtoso-charts-info)", "rgb(9,9,9)", "var(--goshtoso-charts-warning)", "rgb(10,10,10)", "var(--goshtoso-charts-danger)", "'Roboto Medium',sans-serif", "var(--font-paragraph), sans-serif")
	return replacer.Replace(svg)
}

var _ chartcomponents.Component = Instance{}
