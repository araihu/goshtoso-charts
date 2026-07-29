package radar

import (
	"context"
	"fmt"
	"html"
	"io"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	chart "github.com/go-analyze/charts"
)

// Instance is a renderable server-side radar chart.
type Instance struct{ cfg Config }

// Radar returns a server-side SVG radar chart.
func Radar(cfg Config) Instance { return Instance{cfg: cfg} }

// Kind identifies the component as a radar chart.
func (Instance) Kind() chartcomponents.Kind { return chartcomponents.KindRadarChart }

// Render writes an accessible figure, SSR SVG, and adjacent exact-value table.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	svg, err := renderSVG(instance.cfg)
	if err != nil {
		return err
	}
	chart := radarTemplate(instance.cfg, templ.Raw(svg))
	return chartcontrol.Wrapper(chartcontrol.WrapperConfig{
		Label: instance.cfg.Label, Controls: instance.cfg.Controls, Export: instance.cfg.Export,
		Capability: chartcontrol.ExportCapabilityStaticSVG,
	}, chart).Render(ctx, writer)
}

func renderSVG(cfg Config) (string, error) {
	if err := cfg.validate(); err != nil {
		return "", err
	}
	painter := chart.NewPainter(chart.PainterOptions{
		OutputFormat: chart.ChartOutputSVG,
		Width:        cfg.width(),
		Height:       cfg.height(),
		Theme:        tokenPalette(),
	})
	if err := painter.RadarChart(radarOptions(cfg)); err != nil {
		return "", fmt.Errorf("render radar chart: %w", err)
	}
	data, err := painter.Bytes()
	if err != nil {
		return "", fmt.Errorf("encode radar chart SVG: %w", err)
	}
	return tokenizedSVG(string(data), cfg.Style), nil
}

func radarOptions(cfg Config) chart.RadarChartOption {
	values := make([][]float64, len(cfg.Series))
	names := make([]string, len(cfg.Series))
	indicators := make([]chart.RadarIndicator, len(cfg.Indicators))
	for index, indicator := range cfg.Indicators {
		indicators[index] = chart.RadarIndicator{Name: indicator.Name, Min: indicator.Min, Max: indicator.Max}
		if indicator.Label.FontSize > 0 {
			indicators[index].FontStyle = chart.NewFontStyleWithSize(indicator.Label.FontSize)
		}
	}
	for index, series := range cfg.Series {
		values[index] = append([]float64(nil), series.Values...)
		names[index] = series.Name
	}
	options := chart.NewRadarChartOptionWithData(values, nil, nil)
	options.RadarIndicators = indicators
	options.Theme = tokenPalette()
	options.Legend.SeriesNames = names
	options.Title.Text = cfg.Title.Text
	options.Title.Subtext = cfg.Title.Subtext
	options.Title.Show = chart.Ptr(!cfg.Title.Hidden)
	options.Title.Offset = rendererOffset(cfg.Title.Horizontal, cfg.Title.Vertical)
	options.Title.BorderWidth = cfg.Title.BorderWidth
	if cfg.Title.FontSize > 0 {
		options.Title.FontStyle = chart.NewFontStyleWithSize(cfg.Title.FontSize)
	}
	if cfg.Title.SubtextFontSize > 0 {
		options.Title.SubtextFontStyle = chart.NewFontStyleWithSize(cfg.Title.SubtextFontSize)
	}
	options.Legend.Show = chart.Ptr(!cfg.Legend.Hidden)
	options.Legend.Offset = rendererOffset(cfg.Legend.Horizontal, cfg.Legend.Vertical)
	options.Legend.Align = rendererAlignment(cfg.Legend.Alignment)
	options.Legend.Vertical = chart.Ptr(cfg.Legend.Orientation == LegendVertical)
	options.Legend.OverlayChart = chart.Ptr(cfg.Legend.Overlay)
	options.Legend.BorderWidth = cfg.Legend.BorderWidth
	if cfg.Legend.FontSize > 0 {
		options.Legend.FontStyle = chart.NewFontStyleWithSize(cfg.Legend.FontSize)
	}
	if cfg.Legend.Padding != (Padding{}) {
		options.Legend.Padding = rendererPadding(cfg.Legend.Padding)
	}
	if cfg.Padding != (Padding{}) {
		options.Padding = rendererPadding(cfg.Padding)
	}
	radius := cfg.Options.RadiusPercent
	if radius == 0 {
		radius = 40
	}
	options.Radius = strconv.FormatFloat(radius, 'f', -1, 64) + "%"
	options.ValueFormatter = rendererValueFormatter(cfg.Options.ValueFormat)
	for index, series := range cfg.Series {
		options.SeriesList[index].Name = series.Name
		show := series.Options.ValueLabels.resolve(cfg.Options.ValueLabels) == ValueLabelsShown
		options.SeriesList[index].Label.Show = chart.Ptr(show)
		if series.Options.ValueFormat != ValueFormatDefault {
			options.SeriesList[index].Label.ValueFormatter = rendererValueFormatter(series.Options.ValueFormat)
		}
		if series.Options.LabelFontSize > 0 {
			options.SeriesList[index].Label.FontStyle = chart.NewFontStyleWithSize(series.Options.LabelFontSize)
		}
	}
	return options
}

func rendererPadding(padding Padding) chart.Box {
	return chart.NewBox(padding.Left, padding.Top, padding.Right, padding.Bottom)
}

func rendererOffset(horizontal, vertical Placement) chart.OffsetStr {
	return chart.OffsetStr{Left: rendererHorizontalPlacement(horizontal), Top: rendererVerticalPlacement(vertical)}
}

func rendererHorizontalPlacement(placement Placement) string {
	switch placement {
	case PlacementStart:
		return chart.PositionLeft
	case PlacementCenter:
		return chart.PositionCenter
	case PlacementEnd:
		return chart.PositionRight
	default:
		return ""
	}
}

func rendererVerticalPlacement(placement Placement) string {
	switch placement {
	case PlacementStart:
		return chart.PositionTop
	case PlacementCenter:
		return "50%"
	case PlacementEnd:
		return chart.PositionBottom
	default:
		return ""
	}
}

func rendererAlignment(alignment Alignment) string {
	switch alignment {
	case AlignmentStart:
		return chart.AlignLeft
	case AlignmentCenter:
		return chart.AlignCenter
	case AlignmentEnd:
		return chart.AlignRight
	default:
		return ""
	}
}

func rendererValueFormatter(format ValueFormat) chart.ValueFormatter {
	switch format {
	case ValueFormatExact:
		return formatValue
	case ValueFormatInteger:
		return func(value float64) string { return strconv.FormatFloat(value, 'f', 0, 64) }
	case ValueFormatHumanized:
		return func(value float64) string { return chart.FormatValueHumanizeShort(value, 2, false) }
	default:
		return nil
	}
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
		"rgba(5,5,5,0.1)", "color-mix(in srgb, "+html.EscapeString(style.SeriesColor(0))+" 10%, transparent)",
		"rgba(6,6,6,0.1)", "color-mix(in srgb, "+html.EscapeString(style.SeriesColor(1))+" 10%, transparent)",
		"rgba(7,7,7,0.1)", "color-mix(in srgb, "+html.EscapeString(style.SeriesColor(2))+" 10%, transparent)",
		"rgba(8,8,8,0.1)", "color-mix(in srgb, "+html.EscapeString(style.SeriesColor(3))+" 10%, transparent)",
		"rgba(9,9,9,0.1)", "color-mix(in srgb, "+html.EscapeString(style.SeriesColor(4))+" 10%, transparent)",
		"rgba(10,10,10,0.1)", "color-mix(in srgb, "+html.EscapeString(style.SeriesColor(5))+" 10%, transparent)",
		"rgba(11,11,11,0.1)", "color-mix(in srgb, "+html.EscapeString(style.SeriesColor(6))+" 10%, transparent)",
		"rgba(12,12,12,0.1)", "color-mix(in srgb, "+html.EscapeString(style.SeriesColor(7))+" 10%, transparent)",
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

func formatValue(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

var _ chartcomponents.Component = Instance{}
