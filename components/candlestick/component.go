package candlestick

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

// Instance is a renderable server-side candlestick chart.
type Instance struct{ cfg Config }

// Candlestick returns a server-side SVG candlestick chart.
func Candlestick(cfg Config) Instance { return Instance{cfg: cfg} }

// Kind identifies the component as a candlestick chart.
func (Instance) Kind() chartcomponents.Kind { return chartcomponents.KindCandlestickChart }

// Render writes an accessible figure, SSR SVG, and adjacent exact-value table.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	svg, err := renderSVG(instance.cfg)
	if err != nil {
		return err
	}
	chart := candlestickTemplate(instance.cfg, templ.Raw(svg))
	return chartcontrol.Wrapper(chartcontrol.WrapperConfig{
		Label: instance.cfg.Label, Controls: instance.cfg.Controls, Export: instance.cfg.Export,
		Capability: chartcontrol.ExportCapabilityStaticSVG,
	}, chart).Render(ctx, writer)
}

func renderSVG(cfg Config) (string, error) {
	if err := cfg.validate(); err != nil {
		return "", err
	}
	if cfg.Aggregation.WindowSize > 0 {
		return renderAggregationSVG(cfg)
	}
	painter := chart.NewPainter(chart.PainterOptions{OutputFormat: chart.ChartOutputSVG, Width: cfg.width(), Height: cfg.height(), Theme: tokenPalette()})
	if err := painter.CandlestickChart(candlestickOptions(cfg)); err != nil {
		return "", fmt.Errorf("render candlestick chart: %w", err)
	}
	data, err := painter.Bytes()
	if err != nil {
		return "", fmt.Errorf("encode candlestick chart SVG: %w", err)
	}
	return decorateSVG(tokenizedSVG(string(data)), cfg), nil
}

func renderAggregationSVG(cfg Config) (string, error) {
	width, height := cfg.width(), cfg.height()
	painter := chart.NewPainter(chart.PainterOptions{OutputFormat: chart.ChartOutputSVG, Width: width, Height: height, Theme: tokenPalette()})
	surface := chart.Color{R: 1, G: 1, B: 1, A: 255}
	painter.FilledRect(0, 0, width, height, surface, surface, 0)
	middle := height / 2
	top := painter.Child(chart.PainterBoxOption(chart.NewBox(0, 0, width, middle)))
	bottom := painter.Child(chart.PainterBoxOption(chart.NewBox(0, middle, width, height)))
	if err := top.CandlestickChart(candlestickOptions(cfg)); err != nil {
		return "", fmt.Errorf("render source candlestick chart: %w", err)
	}
	if err := bottom.CandlestickChart(candlestickOptions(aggregatedConfig(cfg))); err != nil {
		return "", fmt.Errorf("render aggregated candlestick chart: %w", err)
	}
	data, err := painter.Bytes()
	if err != nil {
		return "", fmt.Errorf("encode aggregated candlestick chart SVG: %w", err)
	}
	return decorateSVG(tokenizedSVG(string(data)), cfg), nil
}

func aggregateData(data []Datum, windowSize int) []Datum {
	if windowSize <= 1 {
		return append([]Datum(nil), data...)
	}
	result := make([]Datum, 0, (len(data)+windowSize-1)/windowSize)
	for start := 0; start < len(data); start += windowSize {
		end := min(start+windowSize, len(data))
		first, last := data[start], data[end-1]
		aggregated := Datum{Label: first.Label, Open: first.Open, High: first.High, Low: first.Low, Close: last.Close}
		if first.Label != last.Label {
			aggregated.Label = first.Label + "-" + last.Label
		}
		for _, datum := range data[start:end] {
			aggregated.High = max(aggregated.High, datum.High)
			aggregated.Low = min(aggregated.Low, datum.Low)
		}
		result = append(result, aggregated)
	}
	return result
}

func aggregatedConfig(cfg Config) Config {
	cfg.Data = aggregateData(cfg.Data, cfg.Aggregation.WindowSize)
	if title := strings.TrimSpace(cfg.Aggregation.Title); title != "" {
		cfg.Title = title
	} else {
		cfg.Title = "Aggregated " + cfg.Title
	}
	if seriesName := strings.TrimSpace(cfg.Aggregation.SeriesName); seriesName != "" {
		cfg.SeriesName = seriesName
	} else {
		cfg.SeriesName += " aggregated"
	}
	cfg.TrendLines = nil
	cfg.Patterns = PatternOptions{}
	cfg.Aggregation = AggregationOptions{}
	return cfg
}

func candlestickOptions(cfg Config) chart.CandlestickChartOption {
	resolved := cfg.resolvedSeries()
	series := make(chart.CandlestickSeriesList, len(resolved))
	labels := make([]string, len(resolved[0].Data))
	for seriesIndex, candidate := range resolved {
		data := make([]chart.OHLCData, len(candidate.Data))
		for index, datum := range candidate.Data {
			if seriesIndex == 0 {
				labels[index] = datum.Label
			}
			data[index] = chart.OHLCData{Open: datum.Open, High: datum.High, Low: datum.Low, Close: datum.Close}
		}
		series[seriesIndex] = chart.CandlestickSeries{
			Name: candidate.Name, Data: data, CandleStyle: chartBodyStyle(candidate.BodyStyle), ShowWicks: candidate.ShowWicks,
		}
	}
	options := chart.NewCandlestickOptionWithSeries(series...)
	options.Theme = tokenPalette()
	options.Title.Text = cfg.Title
	if cfg.Options.TitleFontSize > 0 {
		options.Title.FontStyle = chart.NewFontStyleWithSize(cfg.Options.TitleFontSize)
	}
	options.XAxis.Labels = labels
	options.XAxis.Title = cfg.XAxis.Title
	options.YAxis[0].Title = cfg.YAxis.Title
	options.YAxis[0].Unit = cfg.Options.YUnit
	options.Legend.SeriesNames = make([]string, len(resolved))
	for index := range resolved {
		options.Legend.SeriesNames[index] = resolved[index].Name
	}
	options.Legend.Show = chart.Ptr(!cfg.Options.LegendHidden)
	if len(resolved) == 1 {
		options.SeriesList[0].CloseTrendLine = chartTrendLines(cfg.TrendLines)
		options.SeriesList[0].PatternConfig = chartPatternConfig(cfg.Patterns)
		options.SeriesList[0].CloseMarkLine = chartCloseReferences(cfg.Patterns.References)
	}
	geometry := cfg.Options.Geometry
	if geometry.CandleWidth > 0 {
		options.CandleWidth = geometry.CandleWidth
	}
	if geometry.WickWidth > 0 {
		options.WickWidth = geometry.WickWidth
	}
	options.CandleMargin = geometry.SeriesGap
	options.ShowWicks = geometry.ShowWicks
	if cfg.Options.Padding != (Padding{}) {
		options.Padding = chart.NewBox(cfg.Options.Padding.Left, cfg.Options.Padding.Top, cfg.Options.Padding.Right, cfg.Options.Padding.Bottom)
	}
	return options
}

func chartBodyStyle(style BodyStyle) string {
	switch style {
	case BodyStyleTraditional:
		return chart.CandleStyleTraditional
	case BodyStyleOutline:
		return chart.CandleStyleOutline
	default:
		return chart.CandleStyleFilled
	}
}

func chartPatternConfig(options PatternOptions) *chart.CandlestickPatternConfig {
	if options.Selection == "" && len(options.Enabled) == 0 {
		return nil
	}
	config := &chart.CandlestickPatternConfig{
		PreferPatternLabels: options.PreferLabels,
		EnabledPatterns:     chartPatternTypes(options),
		DojiThreshold:       options.DojiThreshold,
		ShadowTolerance:     options.ShadowTolerance,
		ShadowRatio:         options.ShadowRatio,
		EngulfingMinSize:    options.EngulfingMinSize,
	}
	config.PatternFormatter = func(patterns []chart.PatternDetectionResult, _ string, _ float64) (string, *chart.LabelStyle) {
		if len(patterns) == 0 {
			return "", nil
		}
		names := make([]string, len(patterns))
		for index, pattern := range patterns {
			names[index] = pattern.PatternName
		}
		text := strings.Join(names, "\n")
		if options.Label.Text == PatternLabelTextNameWithCount && len(names) > 1 {
			text = names[0] + " +" + strconv.Itoa(len(names)-1)
		}
		style := chart.LabelStyle{FontStyle: chart.FontStyle{FontColor: chart.Color{R: 8, G: 8, B: 8, A: 255}, FontSize: 10}, BackgroundColor: chart.Color{R: 9, G: 9, B: 9, A: 255}, BorderColor: chart.Color{R: 10, G: 10, B: 10, A: 255}, CornerRadius: 4}
		if options.Label.FontSize > 0 {
			style.FontStyle.FontSize = options.Label.FontSize
		}
		if options.Label.CornerRadius > 0 {
			style.CornerRadius = options.Label.CornerRadius
		}
		return text, &style
	}
	return config
}

func chartPatternTypes(options PatternOptions) []string {
	types := options.Enabled
	if options.Selection != "" {
		types = patternSelections[options.Selection]
	}
	result := make([]string, len(types))
	for index, kind := range types {
		result[index] = strings.ReplaceAll(string(kind), "-", "_")
	}
	return result
}

func chartCloseReferences(references []CloseReferenceType) chart.SeriesMarkLine {
	result := chart.SeriesMarkLine{}
	for _, reference := range references {
		typeValue := chart.SeriesMarkTypeAverage
		if reference == CloseReferenceMinimum {
			typeValue = chart.SeriesMarkTypeMin
		}
		result.Lines = append(result.Lines, chart.SeriesMark{Type: typeValue})
	}
	return result
}

func chartTrendLines(trends []TrendLine) []chart.SeriesTrendLine {
	result := make([]chart.SeriesTrendLine, len(trends))
	for index, trend := range trends {
		result[index] = chart.SeriesTrendLine{
			Type:      chartTrendType(trend.Type),
			Period:    trend.Period,
			LineColor: trendSentinelColor(trend.Type),
		}
	}
	return result
}

func chartTrendType(trendType TrendType) chart.SeriesTrendType {
	switch trendType {
	case TrendTypeBollingerUpper:
		return chart.SeriesTrendTypeBollingerUpper
	case TrendTypeSimpleMovingAverage:
		return chart.SeriesTrendTypeSMA
	case TrendTypeBollingerLower:
		return chart.SeriesTrendTypeBollingerLower
	default:
		return chart.SeriesTrendType(trendType)
	}
}

func trendSentinelColor(trendType TrendType) chart.Color {
	value := uint8(5)
	switch trendType {
	case TrendTypeSimpleMovingAverage:
		value = 6
	case TrendTypeBollingerLower:
		value = 7
	}
	return chart.Color{R: value, G: value, B: value, A: 255}
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
		WithLegendBorderColor(chart.Color{R: 2, G: 2, B: 2, A: 255})
}

func tokenizedSVG(svg string) string {
	return strings.NewReplacer(
		"rgb(1,1,1)", "var(--color-chart-surface)",
		"rgb(2,2,2)", "var(--color-chart-outline)",
		"rgb(3,3,3)", "var(--color-chart-grid)",
		"rgb(4,4,4)", "var(--color-chart-text)",
		"rgb(8,8,8)", "var(--color-chart-pattern-text)",
		"rgb(9,9,9)", "var(--color-chart-pattern-surface)",
		"rgb(10,10,10)", "var(--color-chart-pattern-outline)",
		"rgb(145,204,117)", "var(--color-chart-increasing)",
		"rgb(238,102,102)", "var(--color-chart-decreasing)",
		"'Roboto Medium',sans-serif", "var(--font-paragraph), sans-serif",
	).Replace(svg)
}

var styledPath = regexp.MustCompile(`<path ([^>]*?)style="([^"]*)"/>`)

func decorateSVG(svg string, cfg Config) string {
	for _, trend := range cfg.TrendLines {
		sentinel := trendSentinelCSS(trend.Type)
		color := trendThemeColor(trend.Type)
		if strings.TrimSpace(trend.Color) != "" {
			color = trend.Color
		}
		svg = replaceStyledPaths(svg, sentinel, color, trend.Class)
	}
	svg = applyCandleStyle(svg, "var(--color-chart-increasing)", cfg.Options.Increasing)
	svg = applyCandleStyle(svg, "var(--color-chart-decreasing)", cfg.Options.Decreasing)
	svg = applyPatternLabelStyle(svg, cfg.Patterns.Label)
	return svg
}

func applyPatternLabelStyle(svg string, style PatternLabelStyle) string {
	textColor := "var(--color-chart-pattern-text)"
	if strings.TrimSpace(style.Color) != "" {
		textColor = style.Color
	}
	backgroundColor := "var(--color-chart-pattern-surface)"
	if strings.TrimSpace(style.BackgroundColor) != "" {
		backgroundColor = style.BackgroundColor
	}
	svg = strings.ReplaceAll(svg, "var(--color-chart-pattern-text)", html.EscapeString(textColor))
	svg = strings.ReplaceAll(svg, "var(--color-chart-pattern-surface)", html.EscapeString(backgroundColor))
	if class := strings.TrimSpace(style.Class); class != "" {
		svg = strings.ReplaceAll(svg, `fill="`+html.EscapeString(textColor)+`"`, `class="`+html.EscapeString(class)+`" fill="`+html.EscapeString(textColor)+`"`)
	}
	return svg
}

func replaceStyledPaths(svg, sentinel, color, class string) string {
	return styledPath.ReplaceAllStringFunc(svg, func(path string) string {
		if !strings.Contains(path, sentinel) {
			return path
		}
		path = strings.ReplaceAll(path, sentinel, html.EscapeString(color))
		if class = strings.TrimSpace(class); class != "" {
			path = strings.Replace(path, "<path ", `<path class="`+html.EscapeString(class)+`" `, 1)
		}
		return path
	})
}

func applyCandleStyle(svg, token string, style CandleStyle) string {
	color := token
	if strings.TrimSpace(style.Color) != "" {
		color = style.Color
	}
	return replaceStyledPaths(svg, token, color, style.Class)
}

func trendSentinelCSS(trendType TrendType) string {
	value := 5
	switch trendType {
	case TrendTypeSimpleMovingAverage:
		value = 6
	case TrendTypeBollingerLower:
		value = 7
	}
	return fmt.Sprintf("rgb(%d,%d,%d)", value, value, value)
}

func trendThemeColor(trendType TrendType) string {
	switch trendType {
	case TrendTypeBollingerUpper:
		return "var(--color-chart-bollinger-upper)"
	case TrendTypeSimpleMovingAverage:
		return "var(--color-chart-bollinger-middle)"
	case TrendTypeBollingerLower:
		return "var(--color-chart-bollinger-lower)"
	default:
		return "var(--color-chart-series-1)"
	}
}

type trendValue struct {
	Type  TrendType
	Value float64
}

func computedTrendValues(cfg Config) [][]trendValue {
	series := cfg.resolvedSeries()
	if len(series) == 0 {
		return nil
	}
	closeValues := make([]float64, len(series[0].Data))
	for index, datum := range series[0].Data {
		closeValues[index] = datum.Close
	}
	rows := make([][]trendValue, len(series[0].Data))
	for _, trend := range cfg.TrendLines {
		values := centeredTrend(closeValues, trend)
		for index, value := range values {
			rows[index] = append(rows[index], trendValue{Type: trend.Type, Value: value})
		}
	}
	return rows
}

func centeredTrend(values []float64, trend TrendLine) []float64 {
	result := make([]float64, len(values))
	halfWindow := trend.Period / 2
	for index := range values {
		start := max(0, index-halfWindow)
		end := min(len(values)-1, index+halfWindow)
		count := end - start + 1
		var sum float64
		for sample := start; sample <= end; sample++ {
			sum += values[sample]
		}
		mean := sum / float64(count)
		if trend.Type == TrendTypeSimpleMovingAverage {
			result[index] = mean
			continue
		}
		var variance float64
		for sample := start; sample <= end; sample++ {
			delta := values[sample] - mean
			variance += delta * delta
		}
		offset := 2 * math.Sqrt(variance/float64(count))
		if trend.Type == TrendTypeBollingerLower {
			offset = -offset
		}
		result[index] = mean + offset
	}
	return result
}

func formatValue(value float64) string { return strconv.FormatFloat(value, 'f', -1, 64) }
func exactValuesSummary(cfg Config) string {
	if cfg.Aggregation.WindowSize > 0 {
		return "Exact source and aggregated OHLC values"
	}
	if len(cfg.resolvedSeries()) > 1 {
		return "Exact OHLC values for all series"
	}
	return "Exact OHLC values"
}

func bodyStyleLabel(style BodyStyle) string {
	switch style {
	case BodyStyleTraditional:
		return "Traditional"
	case BodyStyleOutline:
		return "Outline"
	default:
		return "Filled"
	}
}
func formatTrendValue(value float64) string {
	return strconv.FormatFloat(value, 'f', 6, 64)
}
func trendLabel(trendType TrendType) string {
	switch trendType {
	case TrendTypeBollingerUpper:
		return "Bollinger upper"
	case TrendTypeSimpleMovingAverage:
		return "SMA middle"
	case TrendTypeBollingerLower:
		return "Bollinger lower"
	default:
		return string(trendType)
	}
}
func direction(datum Datum) string {
	if datum.Close >= datum.Open {
		return "Increase"
	}
	return "Decrease"
}
func directionClass(datum Datum) string {
	if datum.Close >= datum.Open {
		return "goshtoso-charts-candlestick__direction--increasing"
	}
	return "goshtoso-charts-candlestick__direction--decreasing"
}

func rowDirectionClass(cfg Config, datum Datum) string {
	class := directionClass(datum)
	if datum.Close >= datum.Open {
		if cfg.Options.Increasing.Color == "" {
			class += " " + strings.TrimSpace(cfg.Options.Increasing.Class)
		}
		return strings.TrimSpace(class)
	}
	if cfg.Options.Decreasing.Color == "" {
		class += " " + strings.TrimSpace(cfg.Options.Decreasing.Class)
	}
	return strings.TrimSpace(class)
}

func patternResultsByIndex(cfg Config) [][]PatternResult {
	rows := make([][]PatternResult, len(cfg.Data))
	results, err := DetectPatterns(cfg.Data, cfg.Patterns)
	if err != nil {
		return rows
	}
	for _, result := range results {
		rows[result.Index] = append(rows[result.Index], result)
	}
	return rows
}

func patternNamesAt(cfg Config, index int) string {
	patterns := patternResultsByIndex(cfg)[index]
	names := make([]string, len(patterns))
	for index, pattern := range patterns {
		names[index] = pattern.Name
	}
	return strings.Join(names, ", ")
}

var _ chartcomponents.Component = Instance{}
