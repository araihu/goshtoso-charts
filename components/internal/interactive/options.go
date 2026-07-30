package interactive

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	sharedchart "github.com/araihu/goshtoso-charts/components/chart"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

type ChartOptions = sharedchart.ChartOptions
type TitleOptions = sharedchart.TitleOptions
type LegendOptions = sharedchart.LegendOptions
type LegendSelectionMode = sharedchart.LegendSelectionMode
type EdgeInsets = sharedchart.EdgeInsets
type TooltipOptions = sharedchart.TooltipOptions
type AxisOptions = sharedchart.AxisOptions
type SeriesOptions = sharedchart.SeriesOptions
type LabelOptions = sharedchart.LabelOptions
type ItemStyle = sharedchart.ItemStyle
type LineStyle = sharedchart.LineStyle
type AreaStyle = sharedchart.AreaStyle
type EmphasisOptions = sharedchart.EmphasisOptions
type RippleOptions = sharedchart.RippleOptions
type CalendarOptions = sharedchart.CalendarOptions
type CalendarLabelOptions = sharedchart.CalendarLabelOptions

const (
	LegendSelectionDefault  = sharedchart.LegendSelectionDefault
	LegendSelectionMultiple = sharedchart.LegendSelectionMultiple
	LegendSelectionSingle   = sharedchart.LegendSelectionSingle
)

func FiniteNumber(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

// Percentage formats a finite percentage value for private renderer options.
func Percentage(value float64) string { return fmt.Sprintf("%g%%", value) }

// ValidPercentage reports whether value is within the inclusive 0..100 range.
func ValidPercentage(value float64) bool {
	return FiniteNumber(value) && value >= 0 && value <= 100
}

func ValidateChartOptions(options ChartOptions) error {
	if legend := options.Legend; legend != nil {
		if legend.SelectionMode != LegendSelectionDefault && legend.SelectionMode != LegendSelectionMultiple && legend.SelectionMode != LegendSelectionSingle {
			return fmt.Errorf("legend selection mode %q is not supported", legend.SelectionMode)
		}
		if legend.Padding != nil {
			padding := legend.Padding
			if padding.Top < 0 || padding.Right < 0 || padding.Bottom < 0 || padding.Left < 0 {
				return fmt.Errorf("legend padding must be nonnegative")
			}
		}
	}
	for name, axis := range map[string]*AxisOptions{"x": options.XAxis, "y": options.YAxis} {
		if axis == nil {
			continue
		}
		if axis.LabelInterval != nil && *axis.LabelInterval < 0 {
			return fmt.Errorf("%s axis label interval must be nonnegative", name)
		}
		if strings.ContainsAny(axis.LabelPrefix, "{}") || strings.ContainsAny(axis.LabelSuffix, "{}") {
			return fmt.Errorf("%s axis label prefix and suffix must be literal text without braces", name)
		}
		if axis.SplitNumber < 0 {
			return fmt.Errorf("%s axis split number must be nonnegative", name)
		}
		if axis.Min != nil && !FiniteNumber(*axis.Min) {
			return fmt.Errorf("%s axis minimum must be finite", name)
		}
		if axis.Max != nil && !FiniteNumber(*axis.Max) {
			return fmt.Errorf("%s axis maximum must be finite", name)
		}
		if axis.Min != nil && axis.Max != nil && *axis.Min > *axis.Max {
			return fmt.Errorf("%s axis minimum must not exceed maximum", name)
		}
	}
	return nil
}

func axisLabelIntervalSentinel(value int) string {
	return "__goshtoso_charts_axis_label_interval_" + strconv.Itoa(value) + "__"
}

func AxisLabelIntervals(options ChartOptions) []int {
	result := make([]int, 0, 2)
	for _, axis := range []*AxisOptions{options.XAxis, options.YAxis} {
		if axis != nil && axis.LabelInterval != nil {
			result = append(result, *axis.LabelInterval)
		}
	}
	return result
}

func ChartGlobalOptions(options ChartOptions) []charts.GlobalOpts {
	result := make([]charts.GlobalOpts, 0, 6)
	if value := options.Title; value != nil {
		result = append(result, charts.WithTitleOpts(opts.Title{
			Title: value.Text, Subtitle: value.Subtitle, Left: value.Left, Right: value.Right,
			Top: value.Top, Bottom: value.Bottom,
		}))
	}
	if value := options.Legend; value != nil {
		legend := opts.Legend{Orient: value.Orient, Left: value.Left, Right: value.Right, Top: value.Top, Bottom: value.Bottom, SelectedMode: string(value.SelectionMode)}
		if value.Show != nil {
			legend.Show = opts.Bool(*value.Show)
		}
		if value.Padding != nil {
			legend.Padding = []int{value.Padding.Top, value.Padding.Right, value.Padding.Bottom, value.Padding.Left}
		}
		result = append(result, charts.WithLegendOpts(legend))
	}
	if value := options.Tooltip; value != nil {
		result = append(result, charts.WithTooltipOpts(RendererTooltip(value)))
	}
	if value := options.XAxis; value != nil {
		result = append(result, charts.WithXAxisOpts(RendererXAxis(value)))
	}
	if value := options.YAxis; value != nil {
		result = append(result, charts.WithYAxisOpts(RendererYAxis(value)))
	}
	if options.Animation != nil {
		result = append(result, charts.WithAnimation(*options.Animation))
	}
	return result
}

func ChartSeriesOptions(options SeriesOptions) []charts.SeriesOpts {
	result := make([]charts.SeriesOpts, 0, 8)
	if options.Label != nil {
		result = append(result, charts.WithLabelOpts(RendererLabel(options.Label)))
	}
	if options.ItemStyle != nil {
		result = append(result, charts.WithItemStyleOpts(RendererItemStyle(options.ItemStyle)))
	}
	if options.LineStyle != nil {
		result = append(result, charts.WithLineStyleOpts(RendererLineStyle(options.LineStyle)))
	}
	if options.AreaStyle != nil {
		result = append(result, charts.WithAreaStyleOpts(RendererAreaStyle(options.AreaStyle)))
	}
	if options.Emphasis != nil {
		result = append(result, func(series *charts.SingleSeries) {
			series.Emphasis = RendererEmphasis(options.Emphasis)
		})
	}
	if options.Animation != nil {
		result = append(result, charts.WithAnimationOpts(opts.Animation{Animation: opts.Bool(*options.Animation)}))
	}
	if options.Stack != "" || options.Symbol != "" || options.SymbolSize != 0 || options.Smooth != nil ||
		options.ShowSymbol != nil || options.Step != "" || options.BarWidth != "" || options.BarGap != "" {
		result = append(result, func(series *charts.SingleSeries) {
			if options.Stack != "" {
				series.Stack = options.Stack
			}
			if options.Symbol != "" {
				series.Symbol = options.Symbol
			}
			if options.SymbolSize != 0 {
				series.SymbolSize = options.SymbolSize
			}
			if options.Smooth != nil {
				series.Smooth = opts.Bool(*options.Smooth)
			}
			if options.ShowSymbol != nil {
				series.ShowSymbol = opts.Bool(*options.ShowSymbol)
			}
			if options.Step != "" {
				series.Step = options.Step
			}
			if options.BarWidth != "" {
				series.BarWidth = options.BarWidth
			}
			if options.BarGap != "" {
				series.BarGap = options.BarGap
			}
		})
	}
	return result
}

func MergeSeriesOptions(base, override SeriesOptions) []charts.SeriesOpts {
	result := ChartSeriesOptions(base)
	return append(result, ChartSeriesOptions(override)...)
}

func RendererLabel(value *LabelOptions) opts.Label {
	result := opts.Label{Position: value.Position, Color: value.Color, FontSize: float32(value.FontSize)}
	if value.Show != nil {
		result.Show = opts.Bool(*value.Show)
	}
	return result
}

func RendererItemStyle(value *ItemStyle) opts.ItemStyle {
	result := opts.ItemStyle{
		Color: value.Color, BorderColor: value.BorderColor, BorderWidth: float32(value.BorderWidth),
		ShadowBlur: value.ShadowBlur, ShadowColor: value.ShadowColor,
		ShadowOffsetX: value.ShadowOffsetX, ShadowOffsetY: value.ShadowOffsetY,
	}
	if value.Opacity != nil {
		result.Opacity = opts.Float(float32(*value.Opacity))
	}
	return result
}

func RendererLineStyle(value *LineStyle) opts.LineStyle {
	result := opts.LineStyle{Color: value.Color, Width: float32(value.Width), Type: value.Type}
	if value.Opacity != nil {
		result.Opacity = opts.Float(float32(*value.Opacity))
	}
	return result
}

func RendererAreaStyle(value *AreaStyle) opts.AreaStyle {
	result := opts.AreaStyle{Color: value.Color}
	if value.Opacity != nil {
		result.Opacity = opts.Float(float32(*value.Opacity))
	}
	return result
}

func RendererEmphasis(value *EmphasisOptions) *opts.Emphasis {
	result := &opts.Emphasis{}
	if value.ItemStyle != nil {
		style := RendererItemStyle(value.ItemStyle)
		result.ItemStyle = &style
	}
	if value.Label != nil {
		label := RendererLabel(value.Label)
		result.Label = &label
	}
	return result
}

func RendererTooltip(value *TooltipOptions) opts.Tooltip {
	result := opts.Tooltip{Trigger: value.Trigger}
	if value.Show != nil {
		result.Show = opts.Bool(*value.Show)
	}
	return result
}

func RendererXAxis(value *AxisOptions) opts.XAxis {
	result := opts.XAxis{Name: value.Name, Type: value.Type, SplitNumber: value.SplitNumber}
	if value.Scale != nil {
		result.Scale = opts.Bool(*value.Scale)
	}
	if value.Show != nil {
		result.Show = opts.Bool(*value.Show)
	}
	if value.Min != nil {
		result.Min = *value.Min
	}
	if value.Max != nil {
		result.Max = *value.Max
	}
	if value.ShowSplitLine != nil {
		result.SplitLine = &opts.SplitLine{Show: opts.Bool(*value.ShowSplitLine)}
	}
	if value.LabelInterval != nil {
		result.AxisLabel = &opts.AxisLabel{Interval: axisLabelIntervalSentinel(*value.LabelInterval)}
	}
	if value.ShowFirstLabel != nil || value.ShowLastLabel != nil {
		if result.AxisLabel == nil {
			result.AxisLabel = &opts.AxisLabel{}
		}
		if value.ShowFirstLabel != nil {
			result.AxisLabel.ShowMinLabel = opts.Bool(*value.ShowFirstLabel)
		}
		if value.ShowLastLabel != nil {
			result.AxisLabel.ShowMaxLabel = opts.Bool(*value.ShowLastLabel)
		}
	}
	if value.LabelPrefix != "" || value.LabelSuffix != "" {
		if result.AxisLabel == nil {
			result.AxisLabel = &opts.AxisLabel{}
		}
		result.AxisLabel.Formatter = types.FuncStr(value.LabelPrefix + "{value}" + value.LabelSuffix)
	}
	return result
}

func RendererYAxis(value *AxisOptions) opts.YAxis {
	result := opts.YAxis{Name: value.Name, Type: value.Type, SplitNumber: value.SplitNumber}
	if value.Scale != nil {
		result.Scale = opts.Bool(*value.Scale)
	}
	if value.Show != nil {
		result.Show = opts.Bool(*value.Show)
	}
	if value.Min != nil {
		result.Min = *value.Min
	}
	if value.Max != nil {
		result.Max = *value.Max
	}
	if value.ShowSplitLine != nil {
		result.SplitLine = &opts.SplitLine{Show: opts.Bool(*value.ShowSplitLine)}
	}
	if value.LabelInterval != nil {
		result.AxisLabel = &opts.AxisLabel{Interval: axisLabelIntervalSentinel(*value.LabelInterval)}
	}
	if value.ShowFirstLabel != nil || value.ShowLastLabel != nil {
		if result.AxisLabel == nil {
			result.AxisLabel = &opts.AxisLabel{}
		}
		if value.ShowFirstLabel != nil {
			result.AxisLabel.ShowMinLabel = opts.Bool(*value.ShowFirstLabel)
		}
		if value.ShowLastLabel != nil {
			result.AxisLabel.ShowMaxLabel = opts.Bool(*value.ShowLastLabel)
		}
	}
	if value.LabelPrefix != "" || value.LabelSuffix != "" {
		if result.AxisLabel == nil {
			result.AxisLabel = &opts.AxisLabel{}
		}
		result.AxisLabel.Formatter = types.FuncStr(value.LabelPrefix + "{value}" + value.LabelSuffix)
	}
	return result
}

func RendererCalendar(value CalendarOptions) opts.Calendar {
	result := opts.Calendar{
		Left: value.Left, Right: value.Right, Top: value.Top, Bottom: value.Bottom,
		Width: value.Width, Height: value.Height, CellSize: value.CellSize, Orient: value.Orient,
	}
	if value.CellStyle != nil {
		style := RendererItemStyle(value.CellStyle)
		result.ItemStyle = &style
	}
	result.DayLabel = RendererCalendarLabel(value.DayLabel)
	result.MonthLabel = RendererCalendarLabel(value.MonthLabel)
	result.YearLabel = RendererCalendarLabel(value.YearLabel)
	return result
}

func RendererCalendarLabel(value *CalendarLabelOptions) *opts.CalendarLabel {
	if value == nil {
		return nil
	}
	result := &opts.CalendarLabel{Margin: value.Margin, Position: value.Position, FontSize: value.FontSize}
	if value.Show != nil {
		result.Show = opts.Bool(*value.Show)
	}
	return result
}
