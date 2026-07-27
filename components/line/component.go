package line

import (
	"context"
	"fmt"
	"io"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	chart "github.com/go-analyze/charts"
)

// Instance is a renderable server-side line chart.
type Instance struct {
	cfg Config
}

// Line returns a server-side SVG line chart. Configuration errors are returned
// by Render so the component keeps Goshtoso's normal templ.Component contract.
func Line(cfg Config) Instance {
	return Instance{cfg: cfg}
}

// Kind identifies the component as a line chart.
func (Instance) Kind() chartcomponents.Kind {
	return chartcomponents.KindLineChart
}

// Render writes the accessible figure and SVG. It performs no client-side
// rendering, runtime fetch, or hydration.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	svg, err := renderSVG(instance.cfg)
	if err != nil {
		return err
	}
	return lineTemplate(instance.cfg, templ.Raw(svg)).Render(ctx, writer)
}

func renderSVG(cfg Config) (string, error) {
	if err := cfg.validate(); err != nil {
		return "", err
	}

	values := make([][]float64, 0, len(cfg.Series))
	names := make([]string, 0, len(cfg.Series))
	for _, series := range cfg.Series {
		values = append(values, series.Values)
		names = append(names, series.Name)
	}

	options := chart.NewLineChartOptionWithData(values)
	options.XAxis.Labels = cfg.Labels
	options.Legend.SeriesNames = names
	painter := chart.NewPainter(chart.PainterOptions{
		OutputFormat: chart.ChartOutputSVG,
		Width:        cfg.width(),
		Height:       cfg.height(),
		Theme:        palette(cfg.Theme),
	})
	if err := painter.LineChart(options); err != nil {
		return "", fmt.Errorf("render line chart: %w", err)
	}
	data, err := painter.Bytes()
	if err != nil {
		return "", fmt.Errorf("encode line chart SVG: %w", err)
	}
	return string(data), nil
}

func palette(theme Theme) chart.ColorPalette {
	if theme == ThemeGoshtosoDark {
		return chart.GetTheme(chart.ThemeDark).WithSeriesColors([]chart.Color{
			{R: 167, G: 139, B: 250, A: 255}, // primary-dark
			{R: 52, G: 211, B: 153, A: 255},  // success action
			{R: 96, G: 165, B: 250, A: 255},  // info action
			{R: 251, G: 191, B: 36, A: 255},  // warning action
			{R: 248, G: 113, B: 113, A: 255}, // danger action
		})
	}
	return chart.GetTheme(chart.ThemeLight).WithSeriesColors([]chart.Color{
		{R: 124, G: 58, B: 237, A: 255}, // primary
		{R: 5, G: 150, B: 105, A: 255},  // success
		{R: 37, G: 99, B: 235, A: 255},  // info
		{R: 217, G: 119, B: 6, A: 255},  // warning
		{R: 220, G: 38, B: 38, A: 255},  // danger
	})
}

var _ chartcomponents.Component = Instance{}
