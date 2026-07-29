package pie

import (
	"context"
	"fmt"
	"html"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	chart "github.com/go-analyze/charts"
)

// Instance is a renderable server-side pie chart.
type Instance struct{ cfg Config }

// Pie returns a server-side SVG pie chart.
func Pie(cfg Config) Instance { return Instance{cfg: cfg} }

// Kind identifies the component as a pie chart.
func (Instance) Kind() chartcomponents.Kind { return chartcomponents.KindPieChart }

// Render writes an accessible figure, SSR SVG, and adjacent exact slice summary.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	cfg := instance.cfg
	if cfg.Caption == "" && !cfg.hasData() {
		cfg.Caption = "No data in this period."
	}
	svg, err := renderSVG(cfg)
	if err != nil {
		return err
	}
	chart := pieTemplate(cfg, templ.Raw(svg))
	return chartcontrol.Wrapper(chartcontrol.WrapperConfig{
		Label: cfg.Label, Controls: cfg.Controls, Export: cfg.Export,
		Capability: chartcontrol.ExportCapabilityStaticSVG,
	}, chart).Render(ctx, writer)
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
	if cfg.Variant == VariantDoughnut {
		return renderDoughnutSVG(cfg)
	}
	return renderPieSVG(cfg)
}

// renderPieSVG intentionally retains the pre-extension render path. Its
// serialized output is guarded by TestPieDefaultSVGCompatibilityHash.
func renderPieSVG(cfg Config) (string, error) {
	options := pieOptions(cfg)
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
	return decorateSVG(tokenizedSVG(string(data), cfg), cfg), nil
}

func pieOptions(cfg Config) chart.PieChartOption {
	values, names := sliceData(cfg)
	options := chart.NewPieChartOptionWithData(values)
	options.Theme = tokenPalette()
	applyTitle(&options.Title, cfg.Title)
	applyLegend(&options.Legend, names, cfg.Legend)
	applyPadding(&options.Padding, cfg.Padding)
	options.SegmentGap = cfg.SegmentGap
	label := seriesLabel(cfg.Labels)
	for index := range options.SeriesList {
		options.SeriesList[index].Label = label
	}
	if cfg.Radius.OuterPixels > 0 {
		outer := strconv.FormatFloat(cfg.Radius.OuterPixels, 'f', -1, 64)
		if cfg.Radius.Scale == RadiusScaleArea {
			maximum := options.SeriesList.MaxValue()
			if maximum > 0 {
				for index := range options.SeriesList {
					radius := cfg.Radius.OuterPixels * math.Sqrt(options.SeriesList[index].Value/maximum)
					options.SeriesList[index].Radius = strconv.FormatFloat(math.Ceil(radius), 'f', -1, 64)
				}
			}
		} else {
			options.Radius = outer
		}
	}
	return options
}

func renderDoughnutSVG(cfg Config) (string, error) {
	options := doughnutOptions(cfg)
	painter := chart.NewPainter(chart.PainterOptions{
		OutputFormat: chart.ChartOutputSVG,
		Width:        cfg.width(),
		Height:       cfg.height(),
		Theme:        tokenPalette(),
	})
	if err := painter.DoughnutChart(options); err != nil {
		return "", fmt.Errorf("render doughnut chart: %w", err)
	}
	data, err := painter.Bytes()
	if err != nil {
		return "", fmt.Errorf("encode doughnut chart SVG: %w", err)
	}
	return decorateSVG(tokenizedSVG(string(data), cfg), cfg), nil
}

func doughnutOptions(cfg Config) chart.DoughnutChartOption {
	values, names := sliceData(cfg)
	options := chart.NewDoughnutChartOptionWithData(values)
	options.Theme = tokenPalette()
	applyTitle(&options.Title, cfg.Title)
	applyLegend(&options.Legend, names, cfg.Legend)
	applyPadding(&options.Padding, cfg.Padding)
	options.SegmentGap = cfg.SegmentGap
	label := seriesLabel(cfg.Labels)
	for index := range options.SeriesList {
		options.SeriesList[index].Label = label
	}
	if cfg.Radius.OuterPixels > 0 {
		options.RadiusRing = strconv.FormatFloat(cfg.Radius.OuterPixels, 'f', -1, 64)
	}
	if cfg.Labels.Placement == LabelPlacementInside {
		options.CenterValues = "labels"
	} else if cfg.Center.Content == CenterContentTotal {
		options.CenterValues = "sum"
		if cfg.Center.FontSize > 0 {
			options.CenterValuesFontStyle = chart.NewFontStyleWithSize(cfg.Center.FontSize)
		}
		options.ValueFormatter = centerValueFormatter(cfg.Center)
	}
	if cfg.InnerRadiusPercent > 0 {
		if cfg.Radius.OuterPixels > 0 {
			options.RadiusCenter = strconv.FormatFloat(cfg.Radius.OuterPixels*cfg.InnerRadiusPercent/100, 'f', -1, 64)
		} else {
			// Upstream expresses radii against chart diameter. Goshtoso exposes
			// the hole relative to the active default outer ring instead.
			outerFactor := 0.4
			if cfg.Labels.Placement == LabelPlacementInside {
				outerFactor = 0.5
			}
			options.RadiusCenter = strconv.FormatFloat(cfg.InnerRadiusPercent*outerFactor, 'f', -1, 64) + "%"
		}
	}
	return options
}

func applyTitle(options *chart.TitleOption, cfg TitleOptions) {
	options.Text = cfg.Text
	options.Subtext = cfg.Subtitle
	if cfg.Placement == PlacementCenter {
		options.Offset = chart.OffsetCenter
	}
	if cfg.FontSize > 0 {
		options.FontStyle = chart.NewFontStyleWithSize(cfg.FontSize)
	}
	if cfg.SubtitleFontSize > 0 {
		options.SubtextFontStyle = chart.NewFontStyleWithSize(cfg.SubtitleFontSize)
	}
}

func applyLegend(options *chart.LegendOption, names []string, cfg LegendOptions) {
	options.SeriesNames = names
	if cfg.Hidden {
		options.Show = chart.Ptr(false)
	}
	if cfg.Orientation == LegendVertical {
		options.Vertical = chart.Ptr(true)
	}
	if cfg.LeftPercent > 0 {
		options.Offset.Left = strconv.FormatFloat(cfg.LeftPercent, 'f', -1, 64) + "%"
	}
	switch cfg.VerticalPlacement {
	case VerticalPlacementTop:
		options.Offset.Top = chart.PositionTop
	case VerticalPlacementMiddle:
		options.Offset.Top = chart.PositionCenter
	case VerticalPlacementBottom:
		options.Offset.Top = chart.PositionBottom
	}
	if cfg.FontSize > 0 {
		options.FontStyle = chart.NewFontStyleWithSize(cfg.FontSize)
	}
	if cfg.Overlay {
		options.OverlayChart = chart.Ptr(true)
	}
}

func applyPadding(options *chart.Box, cfg Padding) {
	if cfg != (Padding{}) {
		*options = chart.NewBox(cfg.Left, cfg.Top, cfg.Right, cfg.Bottom)
	}
}

func seriesLabel(cfg LabelOptions) chart.SeriesLabel {
	var result chart.SeriesLabel
	if cfg.Hidden {
		result.Show = chart.Ptr(false)
	}
	if cfg.FontSize > 0 {
		result.FontStyle = chart.NewFontStyleWithSize(cfg.FontSize)
	}
	return result
}

func centerValueFormatter(cfg CenterOptions) chart.ValueFormatter {
	return func(value float64) string {
		formatted := strconv.FormatFloat(value, 'f', -1, 64)
		if cfg.Format == ValueFormatHumanized {
			formatted = chart.FormatValueHumanizeShort(value, cfg.Decimals, false)
		} else if cfg.Decimals > 0 {
			formatted = strconv.FormatFloat(value, 'f', cfg.Decimals, 64)
		}
		return cfg.Prefix + formatted
	}
}

func sliceData(cfg Config) ([]float64, []string) {
	values := make([]float64, len(cfg.Slices))
	names := make([]string, len(cfg.Slices))
	for index, slice := range cfg.Slices {
		values[index] = slice.Value
		names[index] = slice.Name
	}
	return values, names
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

func tokenizedSVG(svg string, cfg Config) string {
	seriesColor := func(index int) string {
		if index < len(cfg.Slices) && strings.TrimSpace(cfg.Slices[index].Color) != "" {
			return cfg.Slices[index].Color
		}
		return cfg.Style.SeriesColor(index)
	}
	replacer := strings.NewReplacer(
		"rgb(1,1,1)", "var(--color-chart-surface)",
		"rgb(2,2,2)", "var(--color-chart-outline)",
		"rgb(3,3,3)", "var(--color-chart-grid)",
		"rgb(4,4,4)", "var(--color-chart-text)",
		"rgb(5,5,5)", html.EscapeString(seriesColor(0)),
		"rgb(6,6,6)", html.EscapeString(seriesColor(1)),
		"rgb(7,7,7)", html.EscapeString(seriesColor(2)),
		"rgb(8,8,8)", html.EscapeString(seriesColor(3)),
		"rgb(9,9,9)", html.EscapeString(seriesColor(4)),
		"rgb(10,10,10)", html.EscapeString(seriesColor(5)),
		"rgb(11,11,11)", html.EscapeString(seriesColor(6)),
		"rgb(12,12,12)", html.EscapeString(seriesColor(7)),
		"'Roboto Medium',sans-serif", "var(--font-paragraph), sans-serif",
	)
	return replacer.Replace(svg)
}

var sliceShape = regexp.MustCompile(`<path ([^>]*?)style="([^"]*?)fill:(var\(--color-chart-series-[0-9]+\)|[^;"]+)([^"]*)"/>`)

func decorateSVG(svg string, cfg Config) string {
	if cfg.Variant == VariantDoughnut {
		svg = strings.Replace(svg, "<svg ", `<svg preserveAspectRatio="xMidYMid meet" `, 1)
	}
	if !hasSliceClasses(cfg) {
		return svg
	}
	return sliceShape.ReplaceAllStringFunc(svg, func(shape string) string {
		for index, slice := range cfg.Slices {
			class := strings.TrimSpace(slice.Class)
			if class == "" {
				continue
			}
			color := slice.Color
			if strings.TrimSpace(color) == "" {
				color = cfg.Style.SeriesColor(index)
			}
			if strings.Contains(shape, "fill:"+html.EscapeString(color)) {
				return strings.Replace(shape, "<path ", `<path class="`+html.EscapeString(class)+`" `, 1)
			}
		}
		return shape
	})
}

func hasSliceClasses(cfg Config) bool {
	for _, slice := range cfg.Slices {
		if strings.TrimSpace(slice.Class) != "" {
			return true
		}
	}
	return false
}

func formatValue(value float64) string { return strconv.FormatFloat(value, 'f', -1, 64) }

func formatShare(value, total float64) string {
	if total == 0 {
		return "0%"
	}
	return strconv.FormatFloat(value/total*100, 'f', -1, 64) + "%"
}

func totalValue(slices []Slice) float64 {
	var total float64
	for _, slice := range slices {
		total += slice.Value
	}
	return total
}

var _ chartcomponents.Component = Instance{}
