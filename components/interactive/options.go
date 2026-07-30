package interactive

import (
	"strconv"

	sharedchart "github.com/araihu/goshtoso-charts/components/chart"
	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
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

func Bool(value bool) *bool        { return sharedchart.Bool(value) }
func Float(value float64) *float64 { return sharedchart.Float(value) }
func Int(value int) *int           { return sharedchart.Int(value) }

func finiteNumber(value float64) bool { return internalinteractive.FiniteNumber(value) }

func percentage(value float64) string { return internalinteractive.Percentage(value) }

func validPercentage(value float64) bool { return internalinteractive.ValidPercentage(value) }

func validateChartOptions(value ChartOptions) error {
	return internalinteractive.ValidateChartOptions(value)
}

func axisLabelIntervalSentinel(value int) string {
	return "__goshtoso_charts_axis_label_interval_" + strconv.Itoa(value) + "__"
}

func axisLabelIntervals(value ChartOptions) []int {
	return internalinteractive.AxisLabelIntervals(value)
}

func chartGlobalOptions(value ChartOptions) []charts.GlobalOpts {
	return internalinteractive.ChartGlobalOptions(value)
}

func chartSeriesOptions(value SeriesOptions) []charts.SeriesOpts {
	return internalinteractive.ChartSeriesOptions(value)
}

func mergeSeriesOptions(base, override SeriesOptions) []charts.SeriesOpts {
	return internalinteractive.MergeSeriesOptions(base, override)
}

func rendererLabel(value *LabelOptions) opts.Label {
	return internalinteractive.RendererLabel(value)
}

func rendererItemStyle(value *ItemStyle) opts.ItemStyle {
	return internalinteractive.RendererItemStyle(value)
}

func rendererLineStyle(value *LineStyle) opts.LineStyle {
	return internalinteractive.RendererLineStyle(value)
}

func rendererAreaStyle(value *AreaStyle) opts.AreaStyle {
	return internalinteractive.RendererAreaStyle(value)
}

func rendererEmphasis(value *EmphasisOptions) *opts.Emphasis {
	return internalinteractive.RendererEmphasis(value)
}

func rendererTooltip(value *TooltipOptions) opts.Tooltip {
	return internalinteractive.RendererTooltip(value)
}

func rendererXAxis(value *AxisOptions) opts.XAxis {
	return internalinteractive.RendererXAxis(value)
}

func rendererYAxis(value *AxisOptions) opts.YAxis {
	return internalinteractive.RendererYAxis(value)
}

func rendererCalendar(value CalendarOptions) opts.Calendar {
	return internalinteractive.RendererCalendar(value)
}

func rendererCalendarLabel(value *CalendarLabelOptions) *opts.CalendarLabel {
	return internalinteractive.RendererCalendarLabel(value)
}
