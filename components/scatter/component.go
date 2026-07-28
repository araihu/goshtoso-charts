package scatter

import (
	"context"
	"fmt"
	"html"
	"io"
	"math"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	chart "github.com/go-analyze/charts"
)

// Instance is a renderable server-side scatter chart.
type Instance struct{ cfg Config }

// Scatter returns a server-side SVG scatter chart.
func Scatter(cfg Config) Instance { return Instance{cfg: cfg} }

// Kind identifies the component as a scatter chart.
func (Instance) Kind() chartcomponents.Kind { return chartcomponents.KindScatterChart }

// Render writes an accessible figure and SVG without browser rendering or hydration.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	svg, err := renderSVG(instance.cfg)
	if err != nil {
		return err
	}
	chart := scatterTemplate(instance.cfg, templ.Raw(svg))
	return chartcontrol.Wrapper(chartcontrol.WrapperConfig{
		Label: instance.cfg.Label, Controls: instance.cfg.Controls, Export: instance.cfg.Export,
		Capability: chartcontrol.ExportCapabilityStaticSVG,
	}, chart).Render(ctx, writer)
}

func renderSVG(cfg Config) (string, error) {
	if err := cfg.validate(); err != nil {
		return "", err
	}
	options := scatterOptions(cfg)
	painter := chart.NewPainter(chart.PainterOptions{
		OutputFormat: chart.ChartOutputSVG,
		Width:        cfg.width(),
		Height:       cfg.height(),
		Theme:        tokenPalette(),
	})
	if err := painter.ScatterChart(options); err != nil {
		return "", fmt.Errorf("render scatter chart: %w", err)
	}
	data, err := painter.Bytes()
	if err != nil {
		return "", fmt.Errorf("encode scatter chart SVG: %w", err)
	}
	return tokenizedSVG(string(data), cfg.Style), nil
}

func scatterOptions(cfg Config) chart.ScatterChartOption {
	categoryIndexes := make(map[string]int, len(cfg.Categories))
	for index, category := range cfg.Categories {
		categoryIndexes[category] = index
	}

	values := make([][][]float64, len(cfg.Series))
	names := make([]string, len(cfg.Series))
	for seriesIndex, series := range cfg.Series {
		if series.Values != nil {
			values[seriesIndex] = make([][]float64, len(series.Values))
			for categoryIndex := range series.Values {
				values[seriesIndex][categoryIndex] = append([]float64(nil), series.Values[categoryIndex]...)
			}
		} else {
			values[seriesIndex] = make([][]float64, len(cfg.Categories))
			for _, point := range series.Points {
				categoryIndex := categoryIndexes[point.Category]
				values[seriesIndex][categoryIndex] = append(values[seriesIndex][categoryIndex], point.Value)
			}
		}
		names[seriesIndex] = series.Name
	}

	options := chart.NewScatterChartOptionWithSeries(chart.NewSeriesListScatterMultiValue(values))
	options.XAxis.Labels = append([]string(nil), cfg.Categories...)
	options.Legend.SeriesNames = names
	options.Theme = tokenPalette()
	options.Symbol = chartSymbol(cfg.Options)
	if cfg.Padding != (Padding{}) {
		options.Padding = chart.NewBox(cfg.Padding.Left, cfg.Padding.Top, cfg.Padding.Right, cfg.Padding.Bottom)
	}
	options.Title.Text = cfg.Title.Text
	if cfg.Title.Placement == PlacementCenter {
		options.Title.Offset = chart.OffsetCenter
	}
	if cfg.Legend.Orientation == LegendVertical {
		options.Legend.Vertical = chart.Ptr(true)
	}
	switch cfg.Legend.Placement {
	case PlacementCenter:
		options.Legend.Offset = chart.OffsetCenter
	case PlacementRight:
		options.Legend.Offset = chart.OffsetRight
	}
	switch cfg.Legend.Alignment {
	case AlignmentRight:
		options.Legend.Align = chart.AlignRight
	case AlignmentCenter:
		options.Legend.Align = chart.AlignCenter
	}
	if cfg.Legend.FontSize > 0 {
		options.Legend.FontStyle = chart.NewFontStyleWithSize(cfg.Legend.FontSize)
	}
	options.XAxis.BoundaryGap = cfg.XAxis.BoundaryGap
	options.XAxis.LabelCount = cfg.XAxis.LabelCount
	options.XAxis.LabelRotation = cfg.XAxis.LabelRotation * math.Pi / 180
	if cfg.XAxis.LabelFontSize > 0 {
		options.XAxis.LabelFontStyle = chart.NewFontStyleWithSize(cfg.XAxis.LabelFontSize)
	}
	if len(options.YAxis) == 0 {
		options.YAxis = []chart.YAxisOption{{}}
	}
	options.YAxis[0].Min, options.YAxis[0].Max = cfg.YAxis.Min, cfg.YAxis.Max
	options.YAxis[0].Unit, options.YAxis[0].LabelSkipCount = cfg.YAxis.Unit, cfg.YAxis.LabelSkip
	if cfg.YAxis.LabelFontSize > 0 {
		options.YAxis[0].LabelFontStyle = chart.NewFontStyleWithSize(cfg.YAxis.LabelFontSize)
	}
	for index, series := range cfg.Series {
		resolved := series.Options.resolved(cfg.Options)
		options.SeriesList[index].Name = series.Name
		options.SeriesList[index].Symbol = chartSymbol(resolved)
		if resolved.Trend.Kind == TrendSimpleMovingAverage {
			options.SeriesList[index].TrendLine = []chart.SeriesTrendLine{{Type: chart.SeriesTrendTypeSMA, Period: resolved.Trend.Period}}
		}
		if resolved.ReferenceLine == ReferenceLineMaximum {
			options.SeriesList[index].MarkLine.AddLines(chart.SeriesMarkTypeMax)
		}
		if resolved.ValueFormat == ValueFormatHumanized {
			formatter := func(value float64) string { return chart.FormatValueHumanizeShort(value, 0, false) }
			options.SeriesList[index].Label.ValueFormatter = formatter
			options.SeriesList[index].MarkLine.ValueFormatter = formatter
			options.ValueFormatter = formatter
		}
	}
	return options
}

func chartSymbol(options Options) chart.Symbol {
	var shape chart.SymbolShape
	switch options.Symbol {
	case SymbolCircle:
		shape = chart.SymbolCircle
	case SymbolSquare:
		shape = chart.SymbolSquare
	case SymbolDiamond:
		shape = chart.SymbolDiamond
	default:
		shape = chart.SymbolDot
	}
	return chart.Symbol{Shape: shape, Size: options.Size}
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
		"rgb(1,1,1)", "var(--color-chart-surface)",
		"rgb(2,2,2)", "var(--color-chart-outline)",
		"rgb(3,3,3)", "var(--color-chart-grid)",
		"rgb(4,4,4)", "var(--color-chart-text)",
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
