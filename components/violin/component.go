package violin

import (
	"context"
	"fmt"
	"html"
	"io"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	chart "github.com/go-analyze/charts"
)

// Instance is a renderable server-side violin chart.
type Instance struct{ cfg Config }

// Violin returns a server-side SVG violin chart.
func Violin(cfg Config) Instance { return Instance{cfg: cfg} }

// Kind identifies the component as a violin chart.
func (Instance) Kind() chartcomponents.Kind { return chartcomponents.KindViolinChart }

// Render writes an accessible figure, SSR SVG, and adjacent exact statistical summary.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	svg, err := renderSVG(instance.cfg)
	if err != nil {
		return err
	}
	content := violinTemplate(instance.cfg, templ.Raw(svg))
	return chartcontrol.Wrapper(chartcontrol.WrapperConfig{
		Label: instance.cfg.Label, Controls: instance.cfg.Controls, Export: instance.cfg.Export,
		Capability: chartcontrol.ExportCapabilityStaticSVG,
	}, content).Render(ctx, writer)
}

func renderSVG(cfg Config) (string, error) {
	if err := cfg.validate(); err != nil {
		return "", err
	}
	options, err := violinOptions(cfg)
	if err != nil {
		return "", err
	}
	painter := chart.NewPainter(chart.PainterOptions{
		OutputFormat: chart.ChartOutputSVG, Width: cfg.width(), Height: cfg.height(), Theme: tokenPalette(),
	})
	if err := painter.ViolinChart(options); err != nil {
		return "", fmt.Errorf("render violin chart: %w", err)
	}
	data, err := painter.Bytes()
	if err != nil {
		return "", fmt.Errorf("encode violin chart SVG: %w", err)
	}
	return decorateSVG(tokenizedSVG(string(data)), cfg), nil
}

func violinOptions(cfg Config) (chart.ViolinChartOption, error) {
	samples := make([][]float64, len(cfg.Series))
	for index := range cfg.Series {
		samples[index] = append([]float64(nil), cfg.Series[index].Samples...)
	}
	arguments := []string{string(cfg.Density.Normalization)}
	if cfg.Density.Bandwidth > 0 {
		arguments = append(arguments, "bandwidth="+formatValue(cfg.Density.Bandwidth))
	}
	options, err := chart.NewViolinChartOptionWithSamples(samples, cfg.Density.points(), arguments...)
	if err != nil {
		return chart.ViolinChartOption{}, fmt.Errorf("estimate violin chart density: %w", err)
	}
	options.Theme = tokenPalette()
	options.Title.Text = cfg.Title
	options.Title.FontStyle = chart.NewFontStyleWithSize(14)
	options.Legend.SeriesNames = make([]string, len(cfg.Series))
	options.Legend.Offset = chart.OffsetRight
	options.Legend.Show = chart.Ptr(true)
	options.ValueAxis.Title = cfg.Axis.Title
	if cfg.Axis.Limit > 0 {
		options.ValueAxis.Limit = chart.Ptr(cfg.Axis.Limit)
	}
	options.ValueAxis.LabelCount = cfg.Axis.LabelCount
	if cfg.Padding != (Padding{}) {
		options.Padding = chart.NewBox(cfg.Padding.Top, cfg.Padding.Right, cfg.Padding.Bottom, cfg.Padding.Left)
	}
	for index, series := range cfg.Series {
		options.Legend.SeriesNames[index] = series.Name
		options.SeriesList[index].Name = series.Name
		marks := make([]string, 0, 2)
		if series.Marks.Mean {
			marks = append(marks, chart.SeriesMarkTypeAverage)
		}
		if series.Marks.Median {
			marks = append(marks, chart.SeriesMarkTypeMedian)
		}
		if len(marks) > 0 {
			options.SeriesList[index].MarkLine = chart.NewMarkLine(marks...)
		}
	}
	return options, nil
}

func tokenPalette() chart.ColorPalette {
	colors := make([]chart.Color, 8)
	for index := range colors {
		colors[index] = chart.Color{R: uint8(10 + index), G: uint8(30 + index), B: uint8(50 + index), A: 255}
	}
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
		WithSeriesColors(colors)
}

func tokenizedSVG(svg string) string {
	return strings.NewReplacer(
		"rgb(1,1,1)", "var(--color-chart-surface)",
		"rgb(2,2,2)", "var(--color-chart-outline)",
		"rgb(3,3,3)", "var(--color-chart-grid)",
		"rgb(4,4,4)", "var(--color-chart-text)",
		"'Roboto Medium',sans-serif", "var(--font-paragraph), sans-serif",
	).Replace(svg)
}

var classableShape = regexp.MustCompile(`<path ([^>]*?)style="([^"]*?)fill:(rgb\([0-9]+,[0-9]+,[0-9]+\))([^"]*)"/>`)

func decorateSVG(svg string, cfg Config) string {
	svg = strings.Replace(svg, "<svg ", `<svg preserveAspectRatio="xMidYMid meet" `, 1)
	return classableShape.ReplaceAllStringFunc(svg, func(shape string) string {
		for index, series := range cfg.Series {
			sentinel := fmt.Sprintf("rgb(%d,%d,%d)", 10+index%8, 30+index%8, 50+index%8)
			if !strings.Contains(shape, sentinel) {
				continue
			}
			color := cfg.Style.SeriesColor(index)
			if strings.TrimSpace(series.Color) != "" {
				color = series.Color
			}
			shape = strings.ReplaceAll(shape, sentinel, html.EscapeString(color))
			if class := strings.TrimSpace(series.Class); class != "" {
				shape = strings.Replace(shape, "<path ", `<path class="`+html.EscapeString(class)+`" `, 1)
			}
			return shape
		}
		return shape
	})
}

type summary struct {
	Count                                  int
	Minimum, Q1, Median, Mean, Q3, Maximum float64
}

func summarize(samples []float64) summary {
	values := append([]float64(nil), samples...)
	slices.Sort(values)
	var total float64
	for _, value := range values {
		total += value
	}
	return summary{len(values), values[0], quantile(values, .25), quantile(values, .5), total / float64(len(values)), quantile(values, .75), values[len(values)-1]}
}

func quantile(sorted []float64, at float64) float64 {
	position := at * float64(len(sorted)-1)
	left := int(math.Floor(position))
	right := int(math.Ceil(position))
	if left == right {
		return sorted[left]
	}
	return sorted[left] + (sorted[right]-sorted[left])*(position-float64(left))
}

func sortedSamples(samples []float64) []float64 {
	values := append([]float64(nil), samples...)
	slices.Sort(values)
	return values
}

func requestedQuantiles(series Series) []float64 {
	values := append([]float64(nil), series.Statistics.Quantiles...)
	slices.Sort(values)
	return values
}

func formatValue(value float64) string    { return strconv.FormatFloat(value, 'f', 2, 64) }
func formatQuantile(value float64) string { return strconv.FormatFloat(value*100, 'f', -1, 64) + "%" }
func formatInt(value int) string          { return strconv.Itoa(value) }

var _ chartcomponents.Component = Instance{}
