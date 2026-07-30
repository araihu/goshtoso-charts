package chart

import "github.com/araihu/goshtoso-charts/components/chartcontrol"

// ChartOptions contains renderer-neutral options shared by interactive charts.
// Component invariants and charttheme.Style take precedence where documented.
type ChartOptions struct {
	Title     *TitleOptions
	Legend    *LegendOptions
	Tooltip   *TooltipOptions
	XAxis     *AxisOptions
	YAxis     *AxisOptions
	Animation *bool
	// Controls configures shared controls; Expand defaults on while fullscreen defaults off.
	Controls chartcontrol.Options
	// Export customizes or disables default PNG export.
	Export *chartcontrol.ExportOptions
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
	Show          *bool
	Orient        string
	Left          string
	Right         string
	Top           string
	Bottom        string
	SelectionMode LegendSelectionMode
	// Padding adds top, right, bottom, and left space around legend content.
	Padding *EdgeInsets
}

// LegendSelectionMode controls whether a legend may expose multiple series or
// only one selected series at a time.
type LegendSelectionMode string

const (
	// LegendSelectionDefault preserves the chart renderer's standard behavior.
	LegendSelectionDefault LegendSelectionMode = ""
	// LegendSelectionMultiple lets readers toggle multiple series independently.
	LegendSelectionMultiple LegendSelectionMode = "multiple"
	// LegendSelectionSingle keeps one legend series selected at a time.
	LegendSelectionSingle LegendSelectionMode = "single"
)

// EdgeInsets describes nonnegative top, right, bottom, and left spacing.
type EdgeInsets struct {
	Top    int
	Right  int
	Bottom int
	Left   int
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
	Show          *bool
	ShowSplitLine *bool
	// SplitNumber recommends a number of numeric-axis segments. Zero preserves the renderer default.
	SplitNumber int
	// Scale excludes zero when useful for tightly bounded numeric data.
	Scale *bool
	// LabelInterval fixes category-label cadence. Zero shows every label; N shows one label after N skipped labels.
	LabelInterval *int
	// ShowFirstLabel and ShowLastLabel keep category-axis endpoint labels visible.
	ShowFirstLabel *bool
	ShowLastLabel  *bool
	// LabelPrefix and LabelSuffix add literal text around every axis value.
	// They do not accept renderer templates or executable formatter code.
	LabelPrefix string
	LabelSuffix string
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
	Emphasis   *EmphasisOptions
}

// LabelOptions configures data-label visibility and presentation.
type LabelOptions struct {
	Show     *bool
	Position string
	Color    string
	FontSize int
}

// ItemStyle configures renderer-neutral fill and border presentation.
type ItemStyle struct {
	Color         string
	BorderColor   string
	BorderWidth   float64
	Opacity       *float64
	ShadowBlur    int
	ShadowColor   string
	ShadowOffsetX int
	ShadowOffsetY int
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
	// CellStyle controls calendar-cell borders and fill without exposing renderer types.
	CellStyle  *ItemStyle
	DayLabel   *CalendarLabelOptions
	MonthLabel *CalendarLabelOptions
	YearLabel  *CalendarLabelOptions
}

// CalendarLabelOptions configures day, month, or year labels around a calendar.
type CalendarLabelOptions struct {
	Show     *bool
	Margin   float64
	Position string
	FontSize int
}

// Bool returns a pointer for renderer-neutral tri-state boolean options.
func Bool(value bool) *bool { return &value }

// Float returns a pointer for renderer-neutral optional numeric options.
func Float(value float64) *float64 { return &value }

// Int returns a pointer for renderer-neutral optional integer options.
func Int(value int) *int { return &value }
