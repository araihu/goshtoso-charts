package heatmap

import (
	"context"
	"fmt"
	"html"
	"io"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	chart "github.com/go-analyze/charts"
)

// Instance is a renderable server-side heat map.
type Instance struct{ cfg Config }

// HeatMap returns a server-side SVG heat map.
func HeatMap(cfg Config) Instance { return Instance{cfg: cfg} }

// Kind identifies the component as a heat map.
func (Instance) Kind() chartcomponents.Kind { return chartcomponents.KindHeatMapChart }

// Render writes an accessible figure, SSR SVG, scale key, and bounded exact-value table.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	svg, err := renderSVG(instance.cfg)
	if err != nil {
		return err
	}
	chart := heatMapTemplate(instance.cfg, templ.Raw(svg))
	return chartcontrol.Wrapper(chartcontrol.WrapperConfig{
		Label: instance.cfg.Label, Controls: instance.cfg.Controls, Export: instance.cfg.Export,
		Capability: chartcontrol.ExportCapabilityStaticSVG,
	}, chart).Render(ctx, writer)
}

func renderSVG(cfg Config) (string, error) {
	if err := cfg.validate(); err != nil {
		return "", err
	}
	rows := cfg.matrix()
	painter := chart.NewPainter(chart.PainterOptions{
		OutputFormat: chart.ChartOutputSVG,
		Width:        cfg.width(), Height: cfg.height(), Theme: tokenPalette(),
	})
	if err := painter.HeatMapChart(heatMapOptions(cfg, rows)); err != nil {
		return "", fmt.Errorf("render heat map: %w", err)
	}
	data, err := painter.Bytes()
	if err != nil {
		return "", fmt.Errorf("encode heat map SVG: %w", err)
	}
	svg, err := applyGradient(string(data), cfg, rows)
	if err != nil {
		return "", err
	}
	return tokenizedSVG(svg), nil
}

func heatMapOptions(cfg Config, rows [][]float64) chart.HeatMapOption {
	options := chart.NewHeatMapOptionWithData(rows)
	options.Theme = tokenPalette()
	options.Title.Text = cfg.Title
	options.Title.Subtext = cfg.TitleOptions.Subtext
	options.Title.Offset = titleOffset(cfg.TitleOptions.Placement)
	if cfg.TitleOptions.Hidden {
		options.Title.Show = chart.Ptr(false)
	}
	if cfg.TitleOptions.FontSize > 0 {
		options.Title.FontStyle = chart.NewFontStyleWithSize(cfg.TitleOptions.FontSize)
	}
	if cfg.TitleOptions.SubtextFontSize > 0 {
		options.Title.SubtextFontStyle = chart.NewFontStyleWithSize(cfg.TitleOptions.SubtextFontSize)
	}
	options.Title.BorderWidth = cfg.TitleOptions.BorderWidth
	if cfg.Padding != (Padding{}) {
		options.Padding = chart.NewBox(cfg.Padding.Left, cfg.Padding.Top, cfg.Padding.Right, cfg.Padding.Bottom)
	}
	options.XAxis.Title = cfg.XAxis.Title
	options.XAxis.Labels = append([]string(nil), cfg.XAxis.Labels...)
	options.YAxis.Title = cfg.YAxis.Title
	options.YAxis.Labels = append([]string(nil), cfg.YAxis.Labels...)
	applyAxisOptions(&options.XAxis, cfg.XAxis)
	applyAxisOptions(&options.YAxis, cfg.YAxis)
	minimum, maximum := cfg.ValueRange.Min, cfg.ValueRange.Max
	options.ScaleMinValue = &minimum
	options.ScaleMaxValue = &maximum
	if cfg.ValueLabels.Show {
		options.ValuesLabel.Show = chart.Ptr(true)
		if cfg.ValueLabels.FontSize > 0 {
			options.ValuesLabel.FontStyle = chart.NewFontStyleWithSize(cfg.ValueLabels.FontSize)
		}
		options.ValuesLabel.Distance = cfg.ValueLabels.Distance
		options.ValuesLabel.Offset = chart.OffsetInt{Left: cfg.ValueLabels.Offset.Left, Top: cfg.ValueLabels.Offset.Top}
		options.ValuesLabel.ValueFormatter = valueLabelFormatter(cfg.ValueLabels)
	}
	return options
}

func titleOffset(placement Placement) chart.OffsetStr {
	switch placement {
	case PlacementStart:
		return chart.OffsetLeft
	case PlacementEnd:
		return chart.OffsetRight
	default:
		return chart.OffsetCenter
	}
}

func applyAxisOptions(target *chart.HeatMapAxis, source Axis) {
	if source.TitleFontSize > 0 {
		target.TitleFontStyle = chart.NewFontStyleWithSize(source.TitleFontSize)
	}
	if source.LabelFontSize > 0 {
		target.LabelFontStyle = chart.NewFontStyleWithSize(source.LabelFontSize)
	}
	target.LabelRotation = source.LabelRotation * math.Pi / 180
	target.LabelCount = source.LabelCount
	target.LabelCountAdjustment = source.LabelCountAdjustment
}

func valueLabelFormatter(options ValueLabelOptions) chart.ValueFormatter {
	if options.Format == ValueFormatDefault && options.Decimals == 0 {
		return nil
	}
	return func(value float64) string {
		switch options.Format {
		case ValueFormatHumanized:
			return chart.FormatValueHumanizeShort(value, options.Decimals, options.TrailingZeros)
		case ValueFormatInteger:
			return strconv.FormatFloat(value, 'f', 0, 64)
		default:
			if options.Decimals > 0 {
				return strconv.FormatFloat(value, 'f', options.Decimals, 64)
			}
			return formatValue(value)
		}
	}
}

func (cfg Config) matrix() [][]float64 {
	if len(cfg.Rows) > 0 {
		rows := make([][]float64, len(cfg.Rows))
		for y := range cfg.Rows {
			rows[y] = append([]float64(nil), cfg.Rows[y]...)
		}
		return rows
	}
	rows := make([][]float64, len(cfg.YAxis.Labels))
	for y := range rows {
		rows[y] = make([]float64, len(cfg.XAxis.Labels))
	}
	for _, cell := range cfg.Cells {
		rows[cell.Y][cell.X] = cell.Value
	}
	return rows
}

func (cfg Config) exactCells() []Cell {
	if len(cfg.Cells) > 0 {
		cells := append([]Cell(nil), cfg.Cells...)
		sort.Slice(cells, func(i, j int) bool {
			if cells[i].Y == cells[j].Y {
				return cells[i].X < cells[j].X
			}
			return cells[i].Y < cells[j].Y
		})
		return cells
	}
	cells := make([]Cell, 0, len(cfg.XAxis.Labels)*len(cfg.YAxis.Labels))
	for y, row := range cfg.Rows {
		for x, value := range row {
			cells = append(cells, Cell{X: x, Y: y, Value: value})
		}
	}
	return cells
}

func (gradient Gradient) resolvedStops() []GradientStop {
	if len(gradient.Stops) > 0 {
		stops := append([]GradientStop(nil), gradient.Stops...)
		for index := range stops {
			if strings.TrimSpace(stops[index].Color) == "" {
				stops[index].Color = defaultColorAt(stops[index].At)
			}
		}
		return stops
	}
	return []GradientStop{
		{At: 0, Color: "var(--color-chart-scale-low)", Class: "goshtoso-charts-heatmap-stop-low"},
		{At: 0.5, Color: "var(--color-chart-scale-mid)", Class: "goshtoso-charts-heatmap-stop-mid"},
		{At: 1, Color: "var(--color-chart-scale-high)", Class: "goshtoso-charts-heatmap-stop-high"},
	}
}

func defaultColorAt(at float64) string {
	if at <= 0.5 {
		return fmt.Sprintf("color-mix(in srgb, var(--color-chart-scale-low) %g%%, var(--color-chart-scale-mid) %g%%)", (1-at*2)*100, at*2*100)
	}
	return fmt.Sprintf("color-mix(in srgb, var(--color-chart-scale-mid) %g%%, var(--color-chart-scale-high) %g%%)", (2-at*2)*100, (at*2-1)*100)
}

func (cfg Config) cellPresentation(value float64) (string, string) {
	ratio := (value - cfg.ValueRange.Min) / (cfg.ValueRange.Max - cfg.ValueRange.Min)
	if cfg.Gradient.Reverse {
		ratio = 1 - ratio
	}
	stops := cfg.Gradient.resolvedStops()
	if ratio <= stops[0].At {
		return stops[0].Color, stops[0].Class
	}
	for index := 1; index < len(stops); index++ {
		right := stops[index]
		if ratio > right.At {
			continue
		}
		left := stops[index-1]
		amount := (ratio - left.At) / (right.At - left.At)
		if amount == 0 {
			return left.Color, left.Class
		}
		if amount == 1 {
			return right.Color, right.Class
		}
		classes := strings.TrimSpace(strings.Join([]string{left.Class, right.Class}, " "))
		color := fmt.Sprintf("color-mix(in srgb, %s %g%%, %s %g%%)", left.Color, (1-amount)*100, right.Color, amount*100)
		return color, classes
	}
	last := stops[len(stops)-1]
	return last.Color, last.Class
}

func (cfg Config) scaleBackground() templ.SafeCSSProperty {
	stops := cfg.Gradient.resolvedStops()
	parts := make([]string, len(stops))
	for index := range stops {
		stop := stops[index]
		position := stop.At
		if cfg.Gradient.Reverse {
			stop = stops[len(stops)-1-index]
			position = 1 - stop.At
		}
		parts[index] = fmt.Sprintf("%s %g%%", stop.Color, position*100)
	}
	return templ.SafeCSSProperty("linear-gradient(90deg, " + strings.Join(parts, ", ") + ")")
}

var filledPath = regexp.MustCompile(`(?s)<path d="([^"]+)" style="stroke:none;fill:[^"]+"/>`)

func applyGradient(svg string, cfg Config, rows [][]float64) (string, error) {
	matches := filledPath.FindAllStringSubmatchIndex(svg, -1)
	cellCount := len(cfg.XAxis.Labels) * len(cfg.YAxis.Labels)
	if len(matches) < cellCount {
		return "", fmt.Errorf("render heat map: found %d filled paths for %d cells", len(matches), cellCount)
	}
	firstCell := len(matches) - cellCount
	var output strings.Builder
	position := 0
	for matchIndex, match := range matches {
		output.WriteString(svg[position:match[0]])
		if matchIndex < firstCell {
			output.WriteString(svg[match[0]:match[1]])
		} else {
			cellIndex := matchIndex - firstCell
			y, x := cellIndex/len(cfg.XAxis.Labels), cellIndex%len(cfg.XAxis.Labels)
			color, class := cfg.cellPresentation(rows[y][x])
			output.WriteString(`<path d="`)
			output.WriteString(svg[match[2]:match[3]])
			output.WriteString(`"`)
			if class != "" {
				output.WriteString(` class="` + html.EscapeString(class) + `"`)
			}
			output.WriteString(` style="stroke:none;fill:` + html.EscapeString(color) + `"/>`)
		}
		position = match[1]
	}
	output.WriteString(svg[position:])
	return output.String(), nil
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
		WithXAxisTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).
		WithYAxisTextColor(chart.Color{R: 4, G: 4, B: 4, A: 255}).
		WithSeriesColors([]chart.Color{{R: 5, G: 5, B: 5, A: 255}})
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

func formatValue(value float64) string { return strconv.FormatFloat(value, 'f', -1, 64) }
func formatInt(value int) string       { return strconv.Itoa(value) }
func formatBool(value bool) string     { return strconv.FormatBool(value) }

func scaleLowLabel(gradient Gradient) string {
	if gradient.Reverse {
		return "warm"
	}
	return "cold"
}

func scaleHighLabel(gradient Gradient) string {
	if gradient.Reverse {
		return "cold"
	}
	return "warm"
}

const exactValueLimit = 100

var _ chartcomponents.Component = Instance{}
