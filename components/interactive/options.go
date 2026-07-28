package interactive

import (
	"math"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// ChartOptions contains renderer-neutral options shared by interactive charts.
// Component invariants and charttheme.Style take precedence where documented.
type ChartOptions struct {
	Title     *TitleOptions
	Legend    *LegendOptions
	Tooltip   *TooltipOptions
	XAxis     *AxisOptions
	YAxis     *AxisOptions
	Animation *bool
}

// TitleOptions configures chart title text and placement.
type TitleOptions struct {
	Text     string
	Subtitle string
	Left     string
	Right    string
	Top      string
	Bottom   string
}

// LegendOptions configures legend visibility, orientation, and placement.
type LegendOptions struct {
	Show   *bool
	Orient string
	Left   string
	Right  string
	Top    string
	Bottom string
}

// TooltipOptions configures tooltip visibility and formatting.
type TooltipOptions struct {
	Show    *bool
	Trigger string
}

// AxisOptions configures a Cartesian axis without exposing renderer types.
type AxisOptions struct {
	Name          string
	Type          string
	Min           *float64
	Max           *float64
	ShowSplitLine *bool
}

// SeriesOptions contains renderer-neutral presentation options shared by series.
type SeriesOptions struct {
	Label      *LabelOptions
	ItemStyle  *ItemStyle
	LineStyle  *LineStyle
	AreaStyle  *AreaStyle
	Animation  *bool
	Stack      string
	Symbol     string
	SymbolSize int
	Smooth     *bool
	ShowSymbol *bool
	Step       string
	BarWidth   string
	BarGap     string
}

// LabelOptions configures data-label visibility and presentation.
type LabelOptions struct {
	Show     *bool
	Position string
	Color    string
}

// ItemStyle configures renderer-neutral fill and border presentation.
type ItemStyle struct {
	Color       string
	BorderColor string
	BorderWidth float64
	Opacity     *float64
}

// LineStyle configures renderer-neutral line presentation.
type LineStyle struct {
	Color   string
	Width   float64
	Type    string
	Opacity *float64
}

// AreaStyle configures renderer-neutral area fill presentation.
type AreaStyle struct {
	Color   string
	Opacity *float64
}

// EmphasisOptions configures highlighted data presentation.
type EmphasisOptions struct {
	ItemStyle *ItemStyle
	Label     *LabelOptions
}

// RippleOptions configures animated effect-scatter ripples.
type RippleOptions struct {
	Period    float64
	Scale     float64
	BrushType string
}

// CalendarOptions configures calendar heatmap placement and cells.
type CalendarOptions struct {
	Left     string
	Right    string
	Top      string
	Bottom   string
	Width    string
	Height   string
	CellSize string
	Orient   string
}

// Bool returns a pointer for renderer-neutral tri-state boolean options.
func Bool(value bool) *bool { return &value }

// Float returns a pointer for renderer-neutral optional numeric options.
func Float(value float64) *float64 { return &value }

func finiteNumber(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func chartGlobalOptions(options ChartOptions) []charts.GlobalOpts {
	result := make([]charts.GlobalOpts, 0, 6)
	if value := options.Title; value != nil {
		result = append(result, charts.WithTitleOpts(opts.Title{
			Title: value.Text, Subtitle: value.Subtitle, Left: value.Left, Right: value.Right,
			Top: value.Top, Bottom: value.Bottom,
		}))
	}
	if value := options.Legend; value != nil {
		legend := opts.Legend{Orient: value.Orient, Left: value.Left, Right: value.Right, Top: value.Top, Bottom: value.Bottom}
		if value.Show != nil {
			legend.Show = opts.Bool(*value.Show)
		}
		result = append(result, charts.WithLegendOpts(legend))
	}
	if value := options.Tooltip; value != nil {
		result = append(result, charts.WithTooltipOpts(rendererTooltip(value)))
	}
	if value := options.XAxis; value != nil {
		result = append(result, charts.WithXAxisOpts(rendererXAxis(value)))
	}
	if value := options.YAxis; value != nil {
		result = append(result, charts.WithYAxisOpts(rendererYAxis(value)))
	}
	if options.Animation != nil {
		result = append(result, charts.WithAnimation(*options.Animation))
	}
	return result
}

func chartSeriesOptions(options SeriesOptions) []charts.SeriesOpts {
	result := make([]charts.SeriesOpts, 0, 8)
	if options.Label != nil {
		result = append(result, charts.WithLabelOpts(rendererLabel(options.Label)))
	}
	if options.ItemStyle != nil {
		result = append(result, charts.WithItemStyleOpts(rendererItemStyle(options.ItemStyle)))
	}
	if options.LineStyle != nil {
		result = append(result, charts.WithLineStyleOpts(rendererLineStyle(options.LineStyle)))
	}
	if options.AreaStyle != nil {
		result = append(result, charts.WithAreaStyleOpts(rendererAreaStyle(options.AreaStyle)))
	}
	if options.Animation != nil {
		result = append(result, charts.WithAnimationOpts(opts.Animation{Animation: opts.Bool(*options.Animation)}))
	}
	if options.Stack != "" || options.Symbol != "" || options.SymbolSize != 0 || options.Smooth != nil ||
		options.ShowSymbol != nil || options.Step != "" || options.BarWidth != "" || options.BarGap != "" {
		result = append(result, func(series *charts.SingleSeries) {
			series.Stack = options.Stack
			series.Symbol = options.Symbol
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
			series.BarWidth = options.BarWidth
			series.BarGap = options.BarGap
		})
	}
	return result
}

func mergeSeriesOptions(base, override SeriesOptions) []charts.SeriesOpts {
	result := chartSeriesOptions(base)
	return append(result, chartSeriesOptions(override)...)
}

func rendererLabel(value *LabelOptions) opts.Label {
	result := opts.Label{Position: value.Position, Color: value.Color}
	if value.Show != nil {
		result.Show = opts.Bool(*value.Show)
	}
	return result
}

func rendererItemStyle(value *ItemStyle) opts.ItemStyle {
	result := opts.ItemStyle{Color: value.Color, BorderColor: value.BorderColor, BorderWidth: float32(value.BorderWidth)}
	if value.Opacity != nil {
		result.Opacity = opts.Float(float32(*value.Opacity))
	}
	return result
}

func rendererLineStyle(value *LineStyle) opts.LineStyle {
	result := opts.LineStyle{Color: value.Color, Width: float32(value.Width), Type: value.Type}
	if value.Opacity != nil {
		result.Opacity = opts.Float(float32(*value.Opacity))
	}
	return result
}

func rendererAreaStyle(value *AreaStyle) opts.AreaStyle {
	result := opts.AreaStyle{Color: value.Color}
	if value.Opacity != nil {
		result.Opacity = opts.Float(float32(*value.Opacity))
	}
	return result
}

func rendererEmphasis(value *EmphasisOptions) *opts.Emphasis {
	result := &opts.Emphasis{}
	if value.ItemStyle != nil {
		style := rendererItemStyle(value.ItemStyle)
		result.ItemStyle = &style
	}
	if value.Label != nil {
		label := rendererLabel(value.Label)
		result.Label = &label
	}
	return result
}

func rendererTooltip(value *TooltipOptions) opts.Tooltip {
	result := opts.Tooltip{Trigger: value.Trigger}
	if value.Show != nil {
		result.Show = opts.Bool(*value.Show)
	}
	return result
}

func rendererXAxis(value *AxisOptions) opts.XAxis {
	result := opts.XAxis{Name: value.Name, Type: value.Type}
	if value.Min != nil {
		result.Min = *value.Min
	}
	if value.Max != nil {
		result.Max = *value.Max
	}
	if value.ShowSplitLine != nil {
		result.SplitLine = &opts.SplitLine{Show: opts.Bool(*value.ShowSplitLine)}
	}
	return result
}

func rendererYAxis(value *AxisOptions) opts.YAxis {
	result := opts.YAxis{Name: value.Name, Type: value.Type}
	if value.Min != nil {
		result.Min = *value.Min
	}
	if value.Max != nil {
		result.Max = *value.Max
	}
	if value.ShowSplitLine != nil {
		result.SplitLine = &opts.SplitLine{Show: opts.Bool(*value.ShowSplitLine)}
	}
	return result
}

func rendererCalendar(value CalendarOptions) opts.Calendar {
	return opts.Calendar{
		Left: value.Left, Right: value.Right, Top: value.Top, Bottom: value.Bottom,
		Width: value.Width, Height: value.Height, CellSize: value.CellSize, Orient: value.Orient,
	}
}
