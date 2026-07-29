package bar

import (
	"context"
	"fmt"
	"html"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	chart "github.com/go-analyze/charts"
)

// Instance is a renderable server-side bar chart.
type Instance struct{ cfg Config }

// Bar returns a server-side SVG bar chart.
func Bar(cfg Config) Instance { return Instance{cfg: cfg} }

// Kind identifies the component as a bar chart.
func (Instance) Kind() chartcomponents.Kind { return chartcomponents.KindBarChart }

// Render writes an accessible figure and SVG without browser rendering or hydration.
// Horizontal charts also include an adjacent exact-value table.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	svg, err := renderSVG(instance.cfg)
	if err != nil {
		return err
	}
	chart := barTemplate(instance.cfg, templ.Raw(svg))
	return chartcontrol.Wrapper(chartcontrol.WrapperConfig{
		Label: instance.cfg.Label, Controls: instance.cfg.Controls, Export: instance.cfg.Export,
		Capability: chartcontrol.ExportCapabilityStaticSVG,
	}, chart).Render(ctx, writer)
}

func renderSVG(cfg Config) (string, error) {
	if err := cfg.validate(); err != nil {
		return "", err
	}
	options := barOptions(cfg)
	painter := chart.NewPainter(chart.PainterOptions{OutputFormat: chart.ChartOutputSVG, Width: cfg.width(), Height: cfg.height(), Theme: tokenPalette()})
	if err := painter.BarChart(options); err != nil {
		return "", fmt.Errorf("render bar chart: %w", err)
	}
	data, err := painter.Bytes()
	if err != nil {
		return "", fmt.Errorf("encode bar chart SVG: %w", err)
	}
	svg := tokenizedSVG(string(data), cfg)
	if cfg.horizontal() {
		svg = strings.Replace(svg, "<svg ", `<svg preserveAspectRatio="xMidYMid meet" `, 1)
		for index := range cfg.Series {
			color := html.EscapeString(seriesColor(cfg, index))
			svg = strings.ReplaceAll(svg, "fill:"+color, "fill:"+color+";stroke:var(--color-chart-text);stroke-width:1")
		}
	}
	return svg, nil
}

func barOptions(cfg Config) chart.BarChartOption {
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
	options.Horizontal = cfg.horizontal()
	options.Title.Text = cfg.Title
	for index, series := range cfg.Series {
		references := series.References
		if references.Average {
			options.SeriesList[index].MarkLine.AddLines(chart.SeriesMarkTypeAverage)
			options.SeriesList[index].MarkLine.ValueFormatter = referenceValueFormatter(references.Format)
		}
		marks := make([]string, 0, 2)
		if references.Maximum {
			marks = append(marks, chart.SeriesMarkTypeMax)
		}
		if references.Minimum {
			marks = append(marks, chart.SeriesMarkTypeMin)
		}
		if len(marks) > 0 {
			options.SeriesList[index].MarkPoint.AddPoints(marks...)
			options.SeriesList[index].MarkPoint.ValueFormatter = referenceValueFormatter(references.Format)
			options.SeriesList[index].MarkPoint.SymbolSize = references.PointSize
		}
	}
	if cfg.Padding != (Padding{}) {
		options.Padding = chart.Box{
			Top: cfg.Padding.Top, Right: cfg.Padding.Right,
			Bottom: cfg.Padding.Bottom, Left: cfg.Padding.Left,
		}
	}
	return options
}

func tokenPalette() chart.ColorPalette {
	return chart.GetTheme(chart.ThemeLight).
		WithBackgroundColor(chart.Color{R: 1, G: 1, B: 1, A: 255}).
		WithXAxisColor(chart.Color{R: 2, G: 2, B: 2, A: 255}).WithYAxisColor(chart.Color{R: 2, G: 2, B: 2, A: 255}).
		WithAxisSplitLineColor(chart.Color{R: 3, G: 3, B: 3, A: 255}).
		WithTitleTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).WithMarkTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).WithLabelTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).WithLegendTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).WithXAxisTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).WithYAxisTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).
		WithTitleBorderColor(chart.Color{R: 2, G: 2, B: 2, A: 255}).WithLegendBorderColor(chart.Color{R: 2, G: 2, B: 2, A: 255}).
		WithSeriesColors([]chart.Color{{R: 5, G: 5, B: 5, A: 255}, {R: 6, G: 6, B: 6, A: 255}, {R: 7, G: 7, B: 7, A: 255}, {R: 8, G: 8, B: 8, A: 255}, {R: 9, G: 9, B: 9, A: 255}, {R: 10, G: 10, B: 10, A: 255}, {R: 11, G: 11, B: 11, A: 255}, {R: 12, G: 12, B: 12, A: 255}})
}

func seriesColor(cfg Config, index int) string {
	if index < len(cfg.Series) {
		if color := strings.TrimSpace(cfg.Series[index].References.Style.Color); color != "" {
			return color
		}
	}
	return cfg.Style.SeriesColor(index)
}

func tokenizedSVG(svg string, cfg Config) string {
	replacer := strings.NewReplacer("rgb(1,1,1)", "var(--color-chart-surface)", "rgb(2,2,2)", "var(--color-chart-outline)", "rgb(3,3,3)", "var(--color-chart-grid)", "rgb(4,4,4)", "var(--color-chart-text)", "rgb(5,5,5)", html.EscapeString(seriesColor(cfg, 0)), "rgb(6,6,6)", html.EscapeString(seriesColor(cfg, 1)), "rgb(7,7,7)", html.EscapeString(seriesColor(cfg, 2)), "rgb(8,8,8)", html.EscapeString(seriesColor(cfg, 3)), "rgb(9,9,9)", html.EscapeString(seriesColor(cfg, 4)), "rgb(10,10,10)", html.EscapeString(seriesColor(cfg, 5)), "rgb(11,11,11)", html.EscapeString(seriesColor(cfg, 6)), "rgb(12,12,12)", html.EscapeString(seriesColor(cfg, 7)), "'Roboto Medium',sans-serif", "var(--font-paragraph), sans-serif")
	return replacer.Replace(svg)
}

func formatValue(value float64) string { return strconv.FormatFloat(value, 'f', -1, 64) }

func referenceValueFormatter(format ValueFormat) chart.ValueFormatter {
	if format == ValueFormatHumanized {
		return func(value float64) string { return strconv.FormatFloat(math.Round(value), 'f', 0, 64) }
	}
	return nil
}

type referenceSummary struct {
	Average      float64
	Minimum      float64
	Maximum      float64
	MinimumIndex int
	MaximumIndex int
}

func summarizeReferences(values []float64) referenceSummary {
	summary := referenceSummary{Minimum: values[0], Maximum: values[0]}
	for index, value := range values {
		summary.Average += value
		if value < summary.Minimum {
			summary.Minimum, summary.MinimumIndex = value, index
		}
		if value > summary.Maximum {
			summary.Maximum, summary.MaximumIndex = value, index
		}
	}
	summary.Average /= float64(len(values))
	return summary
}

func formatReferenceValue(value float64, format ValueFormat) string {
	if format == ValueFormatHumanized {
		return strconv.FormatFloat(math.Round(value), 'f', 0, 64)
	}
	return formatValue(value)
}

var _ chartcomponents.Component = Instance{}
