package pie

import (
	"context"
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	chart "github.com/go-analyze/charts"
)

// Instance is a renderable server-side pie chart.
type Instance struct{ cfg Config }

// Pie returns a server-side SVG pie chart.
func Pie(cfg Config) Instance { return Instance{cfg: cfg} }

// Kind identifies the component as a pie chart.
func (Instance) Kind() chartcomponents.Kind { return chartcomponents.KindPieChart }

// Render writes an accessible figure and SVG without browser rendering or hydration.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	cfg := instance.cfg
	if cfg.Caption == "" && !cfg.hasData() {
		cfg.Caption = "No data in this period."
	}
	svg, err := renderSVG(cfg)
	if err != nil {
		return err
	}
	return pieTemplate(cfg, templ.Raw(svg)).Render(ctx, writer)
}

func (cfg Config) hasData() bool {
	for _, slice := range cfg.Slices {
		if slice.Value > 0 {
			return true
		}
	}
	return false
}

func renderSVG(cfg Config) (string, error) {
	if err := cfg.validate(); err != nil {
		return "", err
	}
	values := make([]float64, 0, len(cfg.Slices))
	names := make([]string, 0, len(cfg.Slices))
	for _, slice := range cfg.Slices {
		values, names = append(values, slice.Value), append(names, slice.Name)
	}
	options := chart.NewPieChartOptionWithData(values)
	options.Legend.SeriesNames = names
	options.Theme = tokenPalette()
	painter := chart.NewPainter(chart.PainterOptions{
		OutputFormat: chart.ChartOutputSVG,
		Width:        cfg.width(),
		Height:       cfg.height(),
		Theme:        tokenPalette(),
	})
	if err := painter.PieChart(options); err != nil {
		return "", fmt.Errorf("render pie chart: %w", err)
	}
	data, err := painter.Bytes()
	if err != nil {
		return "", fmt.Errorf("encode pie chart SVG: %w", err)
	}
	return tokenizedSVG(string(data), cfg.Style), nil
}

// tokenPalette uses unique placeholder colors replaced with Goshtoso CSS tokens.
// This keeps SVG rendering on the server while following light and dark themes.
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
		WithLegendBorderColor(chart.Color{R: 2, G: 2, B: 2, A: 255}).
		WithSeriesColors([]chart.Color{
			{R: 5, G: 5, B: 5, A: 255}, {R: 6, G: 6, B: 6, A: 255},
			{R: 7, G: 7, B: 7, A: 255}, {R: 8, G: 8, B: 8, A: 255},
			{R: 9, G: 9, B: 9, A: 255}, {R: 10, G: 10, B: 10, A: 255},
			{R: 11, G: 11, B: 11, A: 255}, {R: 12, G: 12, B: 12, A: 255},
		})
}

func tokenizedSVG(svg string, style charttheme.Style) string {
	replacer := strings.NewReplacer(
		"rgb(1,1,1)", "var(--goshtoso-charts-surface)",
		"rgb(2,2,2)", "var(--goshtoso-charts-outline)",
		"rgb(3,3,3)", "var(--goshtoso-charts-grid)",
		"rgb(4,4,4)", "var(--goshtoso-charts-text)",
		"rgb(5,5,5)", html.EscapeString(style.SeriesColor(0)),
		"rgb(6,6,6)", html.EscapeString(style.SeriesColor(1)),
		"rgb(7,7,7)", html.EscapeString(style.SeriesColor(2)),
		"rgb(8,8,8)", html.EscapeString(style.SeriesColor(3)),
		"rgb(9,9,9)", html.EscapeString(style.SeriesColor(4)),
		"rgb(10,10,10)", html.EscapeString(style.SeriesColor(5)),
		"rgb(11,11,11)", html.EscapeString(style.SeriesColor(6)),
		"rgb(12,12,12)", html.EscapeString(style.SeriesColor(7)),
		"'Roboto Medium',sans-serif", "var(--font-paragraph), sans-serif",
	)
	return replacer.Replace(svg)
}

var _ chartcomponents.Component = Instance{}
