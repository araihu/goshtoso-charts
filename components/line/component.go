package line

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
	chart := lineTemplate(instance.cfg, templ.Raw(svg))
	return chartcontrol.Wrapper(chartcontrol.WrapperConfig{
		Label: instance.cfg.Label, Controls: instance.cfg.Controls, Export: instance.cfg.Export,
		Capability: chartcontrol.ExportCapabilityStaticSVG,
	}, chart).Render(ctx, writer)
}

func renderSVG(cfg Config) (string, error) {
	if err := cfg.validate(); err != nil {
		return "", err
	}

	options := lineOptions(cfg)
	painter := chart.NewPainter(chart.PainterOptions{
		OutputFormat: chart.ChartOutputSVG,
		Width:        cfg.width(),
		Height:       cfg.height(),
		Theme:        tokenPalette(),
	})
	if err := painter.LineChart(options); err != nil {
		return "", fmt.Errorf("render line chart: %w", err)
	}
	data, err := painter.Bytes()
	if err != nil {
		return "", fmt.Errorf("encode line chart SVG: %w", err)
	}
	return tokenizedSVG(decorateSVG(string(data), cfg), cfg), nil
}

func lineOptions(cfg Config) chart.LineChartOption {
	values := make([][]float64, 0, len(cfg.Series))
	names := make([]string, 0, len(cfg.Series))
	for _, series := range cfg.Series {
		values = append(values, series.Values)
		names = append(names, series.Name)
	}
	options := chart.NewLineChartOptionWithData(values)
	options.XAxis.Labels = cfg.Labels
	options.XAxis.BoundaryGap = cfg.XAxis.BoundaryGap
	options.Legend.SeriesNames = names
	if cfg.Legend.Padding != (Padding{}) {
		padding := cfg.Legend.Padding
		options.Legend.Padding = chart.NewBox(padding.Left, padding.Top, padding.Right, padding.Bottom)
	}
	options.Theme = tokenPalette()
	options.Title.Text = cfg.Title.Text
	if cfg.Area.Enabled {
		options.FillArea = chart.Ptr(true)
		if cfg.Area.Opacity > 0 {
			options.FillOpacity = uint8(math.Round(cfg.Area.Opacity * 255))
		}
	}
	for index, series := range cfg.Series {
		options.SeriesList[index].YAxisIndex = series.YAxisIndex
		options.SeriesList[index].Name = series.Name
	}
	if len(cfg.YAxes) > 0 {
		options.YAxis = make([]chart.YAxisOption, len(cfg.YAxes))
		for index, axis := range cfg.YAxes {
			options.YAxis[index] = chart.YAxisOption{
				Title: axis.Title,
				Unit:  axis.Unit,
				Min:   axis.Min,
				Max:   axis.Max,
				Theme: axisTokenPalette(index),
			}
		}
	}
	return options
}

// tokenPalette uses unique placeholder colors that become Goshtoso CSS tokens
// in tokenizedSVG. This preserves Go-native SSR while allowing the rendered
// SVG to follow every Goshtoso theme and its dark mode without hydration.
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
			{R: 5, G: 5, B: 5, A: 255},
			{R: 6, G: 6, B: 6, A: 255},
			{R: 7, G: 7, B: 7, A: 255},
			{R: 8, G: 8, B: 8, A: 255},
			{R: 9, G: 9, B: 9, A: 255},
			{R: 10, G: 10, B: 10, A: 255},
			{R: 11, G: 11, B: 11, A: 255},
			{R: 12, G: 12, B: 12, A: 255},
		})
}

func axisTokenPalette(index int) chart.ColorPalette {
	color := chart.Color{R: uint8(21 + index), G: uint8(21 + index), B: uint8(21 + index), A: 255}
	return tokenPalette().WithYAxisColor(color).WithYAxisTextColor(color)
}

func tokenizedSVG(svg string, cfg Config) string {
	replacements := []string{
		"rgb(1,1,1)", "var(--color-chart-surface)",
		"rgb(2,2,2)", "var(--color-chart-outline)",
		"rgb(3,3,3)", "var(--color-chart-grid)",
		"rgb(4,4,4)", "var(--color-chart-text)",
	}
	for index := 0; index < 8; index++ {
		if cfg.Area.Enabled {
			opacityByte := uint8(200)
			if cfg.Area.Opacity > 0 {
				opacityByte = uint8(math.Round(cfg.Area.Opacity * 255))
			}
			fillSentinel := fmt.Sprintf("rgba(%d,%d,%d,%.1f)", 5+index, 5+index, 5+index, float64(opacityByte)/255)
			fillColor := fmt.Sprintf("color-mix(in srgb, %s %.6f%%, transparent)", html.EscapeString(seriesColor(cfg, index)), float64(opacityByte)/255*100)
			replacements = append(replacements, fillSentinel, fillColor)
		}
		replacements = append(replacements, colorSentinel(5+index), html.EscapeString(seriesColor(cfg, index)))
	}
	for index := range cfg.YAxes {
		replacements = append(replacements, colorSentinel(21+index), html.EscapeString(axisColor(cfg, index)))
	}
	replacements = append(replacements, "'Roboto Medium',sans-serif", "var(--font-paragraph), sans-serif")
	return strings.NewReplacer(replacements...).Replace(svg)
}

func seriesColor(cfg Config, index int) string {
	if index < len(cfg.Series) && strings.TrimSpace(cfg.Series[index].Color) != "" {
		return cfg.Series[index].Color
	}
	return cfg.Style.SeriesColor(index)
}

func axisColor(cfg Config, axisIndex int) string {
	axis := cfg.YAxes[axisIndex]
	if strings.TrimSpace(axis.Color) != "" {
		return axis.Color
	}
	for seriesIndex, series := range cfg.Series {
		if series.YAxisIndex == axisIndex {
			return seriesColor(cfg, seriesIndex)
		}
	}
	return cfg.Style.SeriesColor(axisIndex)
}

func colorSentinel(value int) string {
	return fmt.Sprintf("rgb(%d,%d,%d)", value, value, value)
}

var coloredSVGElement = regexp.MustCompile(`<(?:path|circle|rect|line|polyline|polygon|text)\b[^>]*>`)

func decorateSVG(svg string, cfg Config) string {
	for index, series := range cfg.Series {
		svg = addClassToColoredElements(svg, colorSentinel(5+index%8), series.Class)
		svg = addClassToColoredElements(svg, rgbaSentinelPrefix(5+index%8), series.Class)
	}
	for index, axis := range cfg.YAxes {
		svg = addClassToColoredElements(svg, colorSentinel(21+index), axis.Class)
	}
	return svg
}

func rgbaSentinelPrefix(value int) string {
	return fmt.Sprintf("rgba(%d,%d,%d,", value, value, value)
}

func addClassToColoredElements(svg, color, class string) string {
	class = strings.TrimSpace(class)
	if class == "" {
		return svg
	}
	escapedClass := html.EscapeString(class)
	return coloredSVGElement.ReplaceAllStringFunc(svg, func(element string) string {
		if !strings.Contains(element, color) {
			return element
		}
		if strings.Contains(element, ` class="`) {
			return strings.Replace(element, ` class="`, ` class="`+escapedClass+` `, 1)
		}
		return strings.Replace(element, " ", ` class="`+escapedClass+`" `, 1)
	})
}

func formatValue(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func axisName(cfg Config, index int) string {
	if index < len(cfg.YAxes) && strings.TrimSpace(cfg.YAxes[index].Title) != "" {
		return cfg.YAxes[index].Title
	}
	if len(cfg.YAxes) == 2 {
		if index == 0 {
			return "Left Y axis"
		}
		return "Right Y axis"
	}
	return "Y axis"
}

func axisClass(cfg Config, index int) string {
	if index >= 0 && index < len(cfg.YAxes) {
		return cfg.YAxes[index].Class
	}
	return ""
}

var _ chartcomponents.Component = Instance{}
