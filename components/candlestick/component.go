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

func candlestickOptions(cfg Config) chart.CandlestickChartOption {
	data := make([]chart.OHLCData, len(cfg.Data))
	labels := make([]string, len(cfg.Data))
	for index, datum := range cfg.Data {
		labels[index] = datum.Label
		data[index] = chart.OHLCData{Open: datum.Open, High: datum.High, Low: datum.Low, Close: datum.Close}
	}
	options := chart.NewCandlestickOptionWithData(data)
	options.Theme = tokenPalette()
	options.Title.Text = cfg.Title
	if cfg.Options.TitleFontSize > 0 {
		options.Title.FontStyle = chart.NewFontStyleWithSize(cfg.Options.TitleFontSize)
	}
	options.XAxis.Labels = labels
	options.XAxis.Title = cfg.XAxis.Title
	options.YAxis[0].Title = cfg.YAxis.Title
	options.YAxis[0].Unit = cfg.Options.YUnit
	options.Legend.SeriesNames = []string{cfg.SeriesName}
	options.Legend.Show = chart.Ptr(!cfg.Options.LegendHidden)
	options.SeriesList[0].Name = cfg.SeriesName
	options.SeriesList[0].CloseTrendLine = chartTrendLines(cfg.TrendLines)
	if cfg.Options.Padding != (Padding{}) {
		options.Padding = chart.NewBox(cfg.Options.Padding.Left, cfg.Options.Padding.Top, cfg.Options.Padding.Right, cfg.Options.Padding.Bottom)
	}
	return options
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
	closeValues := make([]float64, len(cfg.Data))
	for index, datum := range cfg.Data {
		closeValues[index] = datum.Close
	}
	rows := make([][]trendValue, len(cfg.Data))
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

var _ chartcomponents.Component = Instance{}
