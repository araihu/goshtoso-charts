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
	renderTextAnnotations(painter, cfg)
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
		values = append(values, rendererSeriesValues(series))
		names = append(names, series.Name)
	}
	options := chart.NewLineChartOptionWithData(values)
	options.XAxis.Labels = cfg.Labels
	options.XAxis.BoundaryGap = cfg.XAxis.BoundaryGap
	options.Legend.SeriesNames = names
	if cfg.Legend.Hidden {
		options.Legend.Show = chart.Ptr(false)
	}
	if cfg.Legend.Padding != (Padding{}) {
		padding := cfg.Legend.Padding
		options.Legend.Padding = chart.NewBox(padding.Left, padding.Top, padding.Right, padding.Bottom)
	}
	options.Theme = tokenPalette()
	options.Title.Text = cfg.Title.Text
	options.Title.Subtext = cfg.Title.Subtext
	options.Title.Offset = placementOffset(cfg.Title.Placement)
	if cfg.Title.FontSize > 0 {
		options.Title.FontStyle = chart.NewFontStyleWithSize(cfg.Title.FontSize)
	}
	if cfg.Padding != (Padding{}) {
		options.Padding = rendererPadding(cfg.Padding)
	}
	options.Symbol = rendererSymbol(cfg.Symbol)
	options.LineStrokeWidth = cfg.StrokeWidth
	options.StrokeSmoothingTension = cfg.SmoothingTension
	if cfg.Stacked {
		options.StackSeries = chart.Ptr(true)
	}
	options.XAxis.Unit = cfg.XAxis.Unit
	options.XAxis.LabelCount = cfg.XAxis.LabelCount
	options.XAxis.LabelRotation = cfg.XAxis.LabelRotation * math.Pi / 180
	if cfg.XAxis.LabelFontSize > 0 {
		options.XAxis.LabelFontStyle = chart.NewFontStyleWithSize(cfg.XAxis.LabelFontSize)
	}
	if cfg.Area.Enabled {
		options.FillArea = chart.Ptr(true)
		if cfg.Area.Opacity > 0 {
			options.FillOpacity = uint8(math.Round(cfg.Area.Opacity * 255))
		}
	}
	for index, series := range cfg.Series {
		options.SeriesList[index].YAxisIndex = series.YAxisIndex
		options.SeriesList[index].Name = series.Name
		options.SeriesList[index].Symbol = rendererSymbol(series.Symbol)
		options.SeriesList[index].Label = rendererDataLabels(series)
		applyReferences(&options.SeriesList[index], series.References)
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
			if axis.Hidden {
				options.YAxis[index].Show = chart.Ptr(false)
			}
			options.YAxis[index].LabelCount = axis.LabelCount
			if axis.LabelFontSize > 0 {
				options.YAxis[index].LabelFontStyle = chart.NewFontStyleWithSize(axis.LabelFontSize)
			}
			if axis.TitleFontSize > 0 {
				options.YAxis[index].TitleFontStyle = chart.NewFontStyleWithSize(axis.TitleFontSize)
			}
			if axis.SpineLine {
				options.YAxis[index].SpineLineShow = chart.Ptr(true)
			}
		}
	}
	return options
}

func rendererSeriesValues(series Series) []float64 {
	if series.Points == nil {
		return series.Values
	}
	values := make([]float64, len(series.Points))
	for index, point := range series.Points {
		if point.Missing {
			values[index] = chart.GetNullValue()
		} else {
			values[index] = point.Value
		}
	}
	return values
}

func rendererPadding(padding Padding) chart.Box {
	return chart.Box{Top: padding.Top, Right: padding.Right, Bottom: padding.Bottom, Left: padding.Left}
}

func placementOffset(placement Placement) chart.OffsetStr {
	switch placement {
	case PlacementCenter:
		return chart.OffsetCenter
	case PlacementRight:
		return chart.OffsetRight
	default:
		return chart.OffsetStr{}
	}
}

func rendererSymbol(symbol Symbol) chart.Symbol {
	shape := chart.SymbolShape(symbol.Shape)
	return chart.Symbol{Shape: shape, Size: symbol.Size}
}

func rendererDataLabels(series Series) chart.SeriesLabel {
	labels := series.Labels
	if labels == (DataLabelOptions{}) {
		return chart.SeriesLabel{}
	}
	result := chart.SeriesLabel{Show: chart.Ptr(labels.Show)}
	if labels.FontSize > 0 {
		result.FontStyle = chart.NewFontStyleWithSize(labels.FontSize)
	}
	if labels.ColorScale == LabelColorScaleColdToWarm {
		minimum, maximum := seriesExtent(series)
		result.LabelFormatter = func(_ int, _ string, value float64) (string, *chart.LabelStyle) {
			return formatWithDecimals(value, labels.Format, labels.Decimals, labels.TrailingZeros), &chart.LabelStyle{
				FontStyle: chart.FontStyle{FontColor: gradientLabelSentinel(value, minimum, maximum)},
			}
		}
	} else if labels.Format != ValueFormatDefault || labels.Decimals > 0 {
		result.ValueFormatter = func(value float64) string {
			return formatWithDecimals(value, labels.Format, labels.Decimals, labels.TrailingZeros)
		}
	}
	return result
}

func applyReferences(series *chart.LineSeries, references References) {
	if references.Average {
		series.MarkLine.AddLines(chart.SeriesMarkTypeAverage)
		series.MarkLine.ValueFormatter = referenceFormatter(references, false)
	}
	marks := make([]string, 0, 2)
	if references.Maximum {
		marks = append(marks, chart.SeriesMarkTypeMax)
	}
	if references.Minimum {
		marks = append(marks, chart.SeriesMarkTypeMin)
	}
	if len(marks) > 0 {
		series.MarkPoint.AddPoints(marks...)
		series.MarkPoint.SymbolSize = references.PointSize
		series.MarkPoint.ValueFormatter = referenceFormatter(references, true)
	}
}

func referenceFormatter(references References, point bool) chart.ValueFormatter {
	return func(value float64) string {
		prefix := ""
		if point {
			prefix = references.PointPrefix
		}
		return prefix + formatWithDecimals(value, references.Format, references.Decimals, false)
	}
}

func formatWithDecimals(value float64, format ValueFormat, decimals int, trailingZeros bool) string {
	if format == ValueFormatHumanized {
		return chart.FormatValueHumanizeShort(value, decimals, trailingZeros)
	}
	if decimals > 0 {
		return strconv.FormatFloat(value, 'f', decimals, 64)
	}
	return formatValue(value)
}

func seriesExtent(series Series) (float64, float64) {
	minimum, maximum := math.Inf(1), math.Inf(-1)
	for index, value := range resolvedSeriesValues(series) {
		if series.Points != nil && series.Points[index].Missing {
			continue
		}
		minimum, maximum = math.Min(minimum, value), math.Max(maximum, value)
	}
	return minimum, maximum
}

func gradientLabelSentinel(value, minimum, maximum float64) chart.Color {
	factor := 0.0
	if maximum > minimum {
		factor = (value - minimum) / (maximum - minimum)
	}
	factor = math.Max(0, math.Min(1, factor))
	encoded := uint16(math.Round(factor * 65535))
	return chart.Color{R: 123, G: uint8(encoded >> 8), B: uint8(encoded), A: 255}
}

func gradientLabelCSS(value, minimum, maximum float64) string {
	factor := 0.0
	if maximum > minimum {
		factor = (value - minimum) / (maximum - minimum)
	}
	factor = math.Max(0, math.Min(1, factor))
	if factor <= .5 {
		mid := factor * 200
		return fmt.Sprintf("color-mix(in srgb, var(--color-chart-scale-low) %.6f%%, var(--color-chart-scale-mid) %.6f%%)", 100-mid, mid)
	}
	high := (factor - .5) * 200
	return fmt.Sprintf("color-mix(in srgb, var(--color-chart-scale-mid) %.6f%%, var(--color-chart-scale-high) %.6f%%)", 100-high, high)
}

func renderTextAnnotations(painter *chart.Painter, cfg Config) {
	for index, annotation := range cfg.Annotations {
		fontSize := annotation.FontSize
		if fontSize == 0 {
			fontSize = 12
		}
		painter.Text(annotation.Text, annotation.X, annotation.Y, 0, chart.FontStyle{
			Font: chart.GetDefaultFont(), FontSize: fontSize,
			FontColor: annotationSentinel(index),
		})
	}
}

func annotationSentinel(index int) chart.Color {
	encoded := uint16(index)
	return chart.Color{R: 124, G: uint8(encoded >> 8), B: uint8(encoded), A: 255}
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
			fillSentinel := fmt.Sprintf("rgba(%d,%d,%d,%.3f)", 5+index, 5+index, 5+index, float64(opacityByte)/255)
			fillColor := fmt.Sprintf("color-mix(in srgb, %s %.6f%%, transparent)", html.EscapeString(seriesColor(cfg, index)), float64(opacityByte)/255*100)
			replacements = append(replacements, fillSentinel, fillColor)
		}
		replacements = append(replacements, colorSentinel(5+index), html.EscapeString(seriesColor(cfg, index)))
	}
	for _, series := range cfg.Series {
		if series.Labels.ColorScale != LabelColorScaleColdToWarm {
			continue
		}
		minimum, maximum := seriesExtent(series)
		for pointIndex, value := range resolvedSeriesValues(series) {
			if series.Points != nil && series.Points[pointIndex].Missing {
				continue
			}
			sentinel := gradientLabelSentinel(value, minimum, maximum)
			replacements = append(replacements, colorRGB(sentinel), gradientLabelCSS(value, minimum, maximum))
		}
	}
	for index, annotation := range cfg.Annotations {
		replacements = append(replacements, colorRGB(annotationSentinel(index)), html.EscapeString(annotationColor(cfg, annotation)))
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

func colorRGB(color chart.Color) string {
	r, g, b, _ := color.RGBA()
	return fmt.Sprintf("rgb(%d,%d,%d)", r>>8, g>>8, b>>8)
}

func annotationColor(cfg Config, annotation TextAnnotation) string {
	if strings.TrimSpace(annotation.Color) != "" {
		return annotation.Color
	}
	return seriesColor(cfg, annotation.SeriesIndex)
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
	for index, annotation := range cfg.Annotations {
		svg = addClassToColoredElements(svg, colorRGB(annotationSentinel(index)), annotation.Class)
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

func formatSeriesValue(series Series, index int) string {
	if series.Points != nil {
		if series.Points[index].Missing {
			return "Unavailable"
		}
		return formatValue(series.Points[index].Value)
	}
	return formatValue(series.Values[index])
}

type lineReferenceSummary struct {
	Average      float64
	Minimum      float64
	Maximum      float64
	MinimumIndex int
	MaximumIndex int
}

func summarizeLineReferences(series Series) lineReferenceSummary {
	values := resolvedSeriesValues(series)
	summary := lineReferenceSummary{Minimum: math.Inf(1), Maximum: math.Inf(-1)}
	count := 0
	for index, value := range values {
		if series.Points != nil && series.Points[index].Missing {
			continue
		}
		summary.Average += value
		count++
		if value < summary.Minimum {
			summary.Minimum, summary.MinimumIndex = value, index
		}
		if value > summary.Maximum {
			summary.Maximum, summary.MaximumIndex = value, index
		}
	}
	if count > 0 {
		summary.Average /= float64(count)
	}
	return summary
}

func formatReferenceValue(value float64, references References) string {
	return formatWithDecimals(value, references.Format, references.Decimals, false)
}

func (cfg Config) hasReferences() bool {
	for _, series := range cfg.Series {
		if series.References.Average || series.References.Minimum || series.References.Maximum {
			return true
		}
	}
	return false
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
