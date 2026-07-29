package pie

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
	return decorateSVG(tokenizedSVG(string(data), cfg), cfg), nil
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
	options.Title.Text = cfg.Title.Text
	options.Title.Subtext = cfg.Title.Subtitle
	if cfg.Title.Placement == PlacementCenter {
		options.Title.Offset = chart.OffsetCenter
	}
	if cfg.Title.FontSize > 0 {
		options.Title.FontStyle = chart.NewFontStyleWithSize(cfg.Title.FontSize)
	}
	if cfg.Title.SubtitleFontSize > 0 {
		options.Title.SubtextFontStyle = chart.NewFontStyleWithSize(cfg.Title.SubtitleFontSize)
	}
	options.Legend.SeriesNames = names
	if cfg.Legend.Orientation == LegendVertical {
		options.Legend.Vertical = chart.Ptr(true)
	}
	if cfg.Legend.LeftPercent > 0 {
		options.Legend.Offset.Left = strconv.FormatFloat(cfg.Legend.LeftPercent, 'f', -1, 64) + "%"
	}
	switch cfg.Legend.VerticalPlacement {
	case VerticalPlacementTop:
		options.Legend.Offset.Top = chart.PositionTop
	case VerticalPlacementMiddle:
		options.Legend.Offset.Top = chart.PositionCenter
	case VerticalPlacementBottom:
		options.Legend.Offset.Top = chart.PositionBottom
	}
	if cfg.Legend.FontSize > 0 {
		options.Legend.FontStyle = chart.NewFontStyleWithSize(cfg.Legend.FontSize)
	}
	if cfg.Padding != (Padding{}) {
		options.Padding = chart.NewBox(cfg.Padding.Left, cfg.Padding.Top, cfg.Padding.Right, cfg.Padding.Bottom)
	}
	if cfg.InnerRadiusPercent > 0 {
		// Upstream expresses radii against chart diameter. Goshtoso exposes the
		// hole relative to the stable default outer ring instead.
		options.RadiusCenter = strconv.FormatFloat(cfg.InnerRadiusPercent*0.4, 'f', -1, 64) + "%"
	}
	return options
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
