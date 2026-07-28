package funnel

import (
	"context"
	"fmt"
	"html"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	chart "github.com/go-analyze/charts"
)

// Instance is a renderable server-side funnel chart.
type Instance struct{ cfg Config }

// Funnel returns a server-side SVG funnel chart.
func Funnel(cfg Config) Instance { return Instance{cfg: cfg} }

// Kind identifies the component as a funnel chart.
func (Instance) Kind() chartcomponents.Kind { return chartcomponents.KindFunnelChart }

// Render writes an accessible figure, SSR SVG, and adjacent exact stage summary.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	svg, err := renderSVG(instance.cfg)
	if err != nil {
		return err
	}
	content := funnelTemplate(instance.cfg, templ.Raw(svg))
	return chartcontrol.Wrapper(chartcontrol.WrapperConfig{
		Label: instance.cfg.Label, Controls: instance.cfg.Controls, Export: instance.cfg.Export,
		Capability: chartcontrol.ExportCapabilityStaticSVG,
	}, content).Render(ctx, writer)
}

func renderSVG(cfg Config) (string, error) {
	if err := cfg.validate(); err != nil {
		return "", err
	}
	painter := chart.NewPainter(chart.PainterOptions{
		OutputFormat: chart.ChartOutputSVG, Width: cfg.width(), Height: cfg.height(), Theme: tokenPalette(),
	})
	if err := painter.FunnelChart(funnelOptions(cfg)); err != nil {
		return "", fmt.Errorf("render funnel chart: %w", err)
	}
	data, err := painter.Bytes()
	if err != nil {
		return "", fmt.Errorf("encode funnel chart SVG: %w", err)
	}
	return decorateSVG(tokenizedSVG(string(data)), cfg), nil
}

func funnelOptions(cfg Config) chart.FunnelChartOption {
	values := make([]float64, len(cfg.Stages))
	labels := make([]string, len(cfg.Stages))
	for index, stage := range cfg.Stages {
		values[index] = stage.Value
		labels[index] = stage.Label
	}
	options := chart.NewFunnelChartOptionWithData(values)
	options.Theme = tokenPalette()
	options.Title.Text = cfg.Title
	options.Legend.SeriesNames = labels
	options.Legend.Show = chart.Ptr(!cfg.Options.Legend.Hidden)
	options.Legend.Padding = chart.Box{
		Top: cfg.Options.Legend.Padding.Top, Right: cfg.Options.Legend.Padding.Right,
		Bottom: cfg.Options.Legend.Padding.Bottom, Left: cfg.Options.Legend.Padding.Left,
	}
	if cfg.Options.Legend.Orientation == LegendVertical {
		options.Legend.Vertical = chart.Ptr(true)
	}
	switch cfg.Options.Legend.Placement {
	case LegendCenter:
		options.Legend.Offset.Left = chart.PositionCenter
	case LegendEnd:
		options.Legend.Offset.Left = chart.PositionRight
	}
	if cfg.Options.Padding != (Padding{}) {
		options.Padding = chart.NewBox(cfg.Options.Padding.Left, cfg.Options.Padding.Top, cfg.Options.Padding.Right, cfg.Options.Padding.Bottom)
	}
	for index := range options.SeriesList {
		options.SeriesList[index].Name = labels[index]
		applyLabelMode(&options.SeriesList[index], cfg.Options.Labels)
	}
	return options
}

func applyLabelMode(series *chart.FunnelSeries, mode LabelMode) {
	switch mode {
	case LabelHidden:
		series.Label.Show = chart.Ptr(false)
	case LabelName:
		series.Label.LabelFormatter = func(_ int, name string, _ float64) (string, *chart.LabelStyle) { return name, nil }
	case LabelValue:
		series.Label.ValueFormatter = func(value float64) string { return formatValue(value) }
	case LabelNameValue:
		series.Label.LabelFormatter = func(_ int, name string, value float64) (string, *chart.LabelStyle) {
			return name + " (" + formatValue(value) + ")", nil
		}
	}
}

func tokenPalette() chart.ColorPalette {
	colors := make([]chart.Color, 8)
	for index := range colors {
		colors[index] = chart.Color{R: uint8(10 + index), G: uint8(30 + index), B: uint8(50 + index), A: 255}
	}
	return chart.GetTheme(chart.ThemeLight).
		WithBackgroundColor(chart.Color{R: 1, G: 1, B: 1, A: 255}).
		WithTitleTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).
		WithLabelTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).
		WithLegendTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).
		WithSeriesColors(colors)
}

func tokenizedSVG(svg string) string {
	return strings.NewReplacer(
		"rgb(1,1,1)", "var(--color-chart-surface)",
		"rgb(4,4,4)", "var(--color-chart-text-strong)",
		"'Roboto Medium',sans-serif", "var(--font-paragraph), sans-serif",
	).Replace(svg)
}

var stageShape = regexp.MustCompile(`<path ([^>]*?)style="([^"]*?)fill:(rgb\([0-9]+,[0-9]+,[0-9]+\))([^"]*)"/>`)

func decorateSVG(svg string, cfg Config) string {
	svg = strings.Replace(svg, "<svg ", `<svg preserveAspectRatio="xMidYMid meet" `, 1)
	return stageShape.ReplaceAllStringFunc(svg, func(shape string) string {
		for index, stage := range cfg.Stages {
			sentinel := fmt.Sprintf("rgb(%d,%d,%d)", 10+index%8, 30+index%8, 50+index%8)
			if !strings.Contains(shape, sentinel) {
				continue
			}
			color := cfg.Style.SeriesColor(index)
			if strings.TrimSpace(stage.Color) != "" {
				color = stage.Color
			}
			shape = strings.ReplaceAll(shape, sentinel, html.EscapeString(color))
			if class := strings.TrimSpace(stage.Class); class != "" {
				shape = strings.Replace(shape, "<path ", `<path class="`+html.EscapeString(class)+`" `, 1)
			}
			return shape
		}
		return shape
	})
}

func formatValue(value float64) string { return strconv.FormatFloat(value, 'f', -1, 64) }

func formatPercent(value, total float64) string {
	if total == 0 {
		return "100%"
	}
	return strconv.FormatFloat(value/total*100, 'f', -1, 64) + "%"
}

var _ chartcomponents.Component = Instance{}
