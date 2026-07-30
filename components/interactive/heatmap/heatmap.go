// Package heatmap provides the canonical interactive heatmap API.
//
// Cartesian and calendar coordinates remain behavior variants of one component.
// Heatmap-specific types and implementation live here; shared renderer-neutral
// options remain in components/chart.
package heatmap

import (
	"fmt"
	"math"
	"strconv"
	"time"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// Instance is the renderer-neutral chart instance returned by HeatMap.
type Instance = chart.Instance

// Coordinate selects the heatmap coordinate system.
type Coordinate string

const (
	// CoordinateCartesian maps X and Y indexes to category axes. It is the default.
	CoordinateCartesian Coordinate = ""
	// CoordinateCalendar maps Date values to a calendar range.
	CoordinateCalendar Coordinate = "calendar"
)

// Config describes an accessible, browser-rendered heatmap.
//
// Values must be application-owned because the browser renderer serializes them.
type Config struct {
	Label      string
	Caption    string
	Coordinate Coordinate
	XAxis      []string
	YAxis      []string
	Calendar   *Calendar
	ValueRange ValueRange
	// SplitArea shows alternating category-cell regions on both Cartesian axes.
	SplitArea     *bool
	Series        []Series
	Width         string
	Height        string
	Options       chart.ChartOptions
	SeriesOptions chart.SeriesOptions
	Style         charttheme.Style
}

// Calendar defines an inclusive calendar date range. Options customizes
// calendar presentation; its Range is replaced by Start and End.
type Calendar struct {
	Start   time.Time
	End     time.Time
	Options chart.CalendarOptions
}

// ValueRange defines the visual-map domain. Values outside the range are
// preserved and rendered with the nearest endpoint color.
type ValueRange struct {
	Min        float64
	Max        float64
	Calculable *bool
}

// Series describes one named heatmap series.
type Series struct {
	Name    string
	Data    []Data
	Options chart.SeriesOptions
}

// Data describes one heatmap cell. Cartesian mode uses X and Y category
// indexes. Calendar mode uses Date.
type Data struct {
	X     int
	Y     int
	Date  time.Time
	Value float64
	// Missing preserves an explicit no-data cell. Value is ignored when true.
	Missing bool
}

// HeatMap builds a reusable interactive Cartesian or calendar heatmap component.
func HeatMap(cfg Config) Instance {
	if err := validateConfig(cfg); err != nil {
		return internalinteractive.Invalid(chartcomponents.KindInteractiveHeatMap, err)
	}

	chart := charts.NewHeatMap()
	globalOptions := globalOptions(cfg)
	chart.SetGlobalOptions(globalOptions...)
	normalizeVisualMap(chart, cfg)
	if cfg.Coordinate == CoordinateCalendar {
		calendar := internalinteractive.RendererCalendar(cfg.Calendar.Options)
		calendar.Range = []string{dateString(cfg.Calendar.Start), dateString(cfg.Calendar.End)}
		chart.AddCalendar(&calendar)
	} else {
		// HeatMap.Validate does not call RectChart.Validate, so SetXAxis alone
		// does not reach rendered output in go-echarts v2.7.2.
		chart.XAxisList[0].Data = cfg.XAxis
		chart.YAxisList[0].Data = cfg.YAxis
	}

	for _, series := range cfg.Series {
		data := make([]opts.HeatMapData, len(series.Data))
		for index, point := range series.Data {
			value := any(point.Value)
			if point.Missing {
				value = "-"
			}
			if cfg.Coordinate == CoordinateCalendar {
				data[index] = opts.HeatMapData{Value: [2]interface{}{dateString(point.Date), value}}
			} else {
				data[index] = opts.HeatMapData{Value: [3]interface{}{point.X, point.Y, value}}
			}
		}
		options := internalinteractive.MergeSeriesOptions(cfg.SeriesOptions, series.Options)
		if cfg.Coordinate == CoordinateCalendar {
			// Coordinate is a component invariant, not an escape-hatch option.
			options = append(options, charts.WithCoordinateSystem("calendar"))
		}
		chart.AddSeries(series.Name, data, options...)
	}

	return internalinteractive.New(chartcomponents.KindInteractiveHeatMap, internalinteractive.RenderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		ResponsiveWidth: internalinteractive.ResponsiveWidth(cfg.Width), ExplicitVisualMapColors: len(cfg.Style.Colors) > 0,
		Details: exactValues(cfg.Label, cfg.Coordinate, detailRows(cfg)),
	})
}

func globalOptions(cfg Config) []charts.GlobalOpts {
	options := []charts.GlobalOpts{
		charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())),
	}
	if cfg.Coordinate == CoordinateCartesian {
		xAxis := opts.XAxis{Type: "category"}
		yAxis := opts.YAxis{Type: "category"}
		if cfg.SplitArea != nil {
			xAxis.SplitArea = &opts.SplitArea{Show: opts.Bool(*cfg.SplitArea)}
			yAxis.SplitArea = &opts.SplitArea{Show: opts.Bool(*cfg.SplitArea)}
		}
		options = append(options,
			charts.WithGridOpts(opts.Grid{Left: "52", Right: "0", Bottom: "56", ContainLabel: opts.Bool(true)}),
			charts.WithXAxisOpts(xAxis),
			charts.WithYAxisOpts(yAxis),
		)
	}
	options = append(options, internalinteractive.ChartGlobalOptions(cfg.Options)...)
	if len(cfg.Style.Colors) > 0 {
		// Explicit component colors remain authoritative over escape-hatch options.
		options = append(options, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
	if cfg.Width != "" || cfg.Height != "" {
		options = append([]charts.GlobalOpts{
			charts.WithInitializationOpts(opts.Initialization{Width: cfg.Width, Height: cfg.Height}),
		}, options...)
	}
	return options
}

func normalizeVisualMap(chart *charts.HeatMap, cfg Config) {
	visualMap := visualMap(cfg)
	if len(chart.VisualMapList) > 0 {
		// GlobalOptions use last-write-wins semantics for this single visual map.
		visualMap = chart.VisualMapList[len(chart.VisualMapList)-1]
		visualMap.Min = float32(cfg.ValueRange.Min)
		visualMap.Max = float32(cfg.ValueRange.Max)
		if len(cfg.Style.Colors) > 0 || visualMap.InRange == nil || len(visualMap.InRange.Color) == 0 {
			visualMap.InRange = &opts.VisualMapInRange{Color: cfg.Style.ResolvedColors()}
		}
	}
	chart.VisualMapList = []opts.VisualMap{visualMap}
}

func visualMap(cfg Config) opts.VisualMap {
	result := opts.VisualMap{
		Min: float32(cfg.ValueRange.Min), Max: float32(cfg.ValueRange.Max),
		Left: "8", Bottom: "24",
		InRange: &opts.VisualMapInRange{Color: cfg.Style.ResolvedColors()},
	}
	if cfg.Coordinate == CoordinateCalendar {
		// Calendar weekday labels occupy the left edge. Keep the continuous
		// scale on the opposite side so narrow layouts preserve every label.
		result.Left = ""
		result.Right = "0"
	}
	if cfg.ValueRange.Calculable != nil {
		result.Calculable = opts.Bool(*cfg.ValueRange.Calculable)
	}
	return result
}

type valueRow struct {
	Series  string
	X       string
	Y       string
	Date    string
	Value   string
	Missing bool
}

func detailRows(cfg Config) []valueRow {
	rows := make([]valueRow, 0)
	for _, series := range cfg.Series {
		for _, point := range series.Data {
			row := valueRow{Series: series.Name, Missing: point.Missing}
			if cfg.Coordinate == CoordinateCalendar {
				row.Date = dateString(point.Date)
			} else {
				row.X = cfg.XAxis[point.X]
				row.Y = cfg.YAxis[point.Y]
			}
			if point.Missing {
				row.Value = "No data"
			} else {
				row.Value = strconv.FormatFloat(point.Value, 'f', -1, 64)
			}
			rows = append(rows, row)
		}
	}
	return rows
}

func validateConfig(cfg Config) error {
	if err := internalinteractive.ValidateChartOptions(cfg.Options); err != nil {
		return err
	}
	if cfg.Label == "" {
		return fmt.Errorf("heatmap chart label is required")
	}
	if cfg.Coordinate != CoordinateCartesian && cfg.Coordinate != CoordinateCalendar {
		return fmt.Errorf("heatmap chart coordinate %q is not supported", cfg.Coordinate)
	}
	if !internalinteractive.FiniteNumber(cfg.ValueRange.Min) || !internalinteractive.FiniteNumber(cfg.ValueRange.Max) || cfg.ValueRange.Min >= cfg.ValueRange.Max {
		return fmt.Errorf("heatmap chart value range must contain finite min and max with min less than max")
	}
	if math.Abs(cfg.ValueRange.Min) > math.MaxFloat32 || math.Abs(cfg.ValueRange.Max) > math.MaxFloat32 {
		return fmt.Errorf("heatmap chart value range exceeds renderer limits")
	}
	if cfg.Coordinate == CoordinateCalendar {
		if cfg.SplitArea != nil {
			return fmt.Errorf("heatmap chart split area is not allowed for calendar coordinates")
		}
		if len(cfg.XAxis) != 0 || len(cfg.YAxis) != 0 {
			return fmt.Errorf("heatmap chart category axes are not allowed for calendar coordinates")
		}
		if cfg.Calendar == nil {
			return fmt.Errorf("heatmap chart calendar is required for calendar coordinates")
		}
		if cfg.Calendar.Start.IsZero() || cfg.Calendar.End.IsZero() {
			return fmt.Errorf("heatmap chart calendar start and end dates are required")
		}
		if dateString(cfg.Calendar.Start) > dateString(cfg.Calendar.End) {
			return fmt.Errorf("heatmap chart calendar start must not follow end")
		}
		if style := cfg.Calendar.Options.CellStyle; style != nil {
			if !internalinteractive.FiniteNumber(style.BorderWidth) || style.BorderWidth < 0 {
				return fmt.Errorf("heatmap chart calendar cell border width must be finite and nonnegative")
			}
			if style.BorderWidth > math.MaxFloat32 {
				return fmt.Errorf("heatmap chart calendar cell border width exceeds renderer limits")
			}
			if style.Opacity != nil && (!internalinteractive.FiniteNumber(*style.Opacity) || *style.Opacity < 0 || *style.Opacity > 1) {
				return fmt.Errorf("heatmap chart calendar cell opacity must be between 0 and 1")
			}
		}
		for _, candidate := range []struct {
			name  string
			label *chart.CalendarLabelOptions
		}{
			{name: "day", label: cfg.Calendar.Options.DayLabel},
			{name: "month", label: cfg.Calendar.Options.MonthLabel},
			{name: "year", label: cfg.Calendar.Options.YearLabel},
		} {
			name, label := candidate.name, candidate.label
			if label == nil {
				continue
			}
			if !internalinteractive.FiniteNumber(label.Margin) || label.Margin < 0 || label.FontSize < 0 {
				return fmt.Errorf("heatmap chart calendar %s label margin and font size must be finite and nonnegative", name)
			}
			if label.Position != "" && label.Position != "left" && label.Position != "right" && label.Position != "top" && label.Position != "bottom" {
				return fmt.Errorf("heatmap chart calendar %s label position %q is not supported", name, label.Position)
			}
		}
	} else {
		if cfg.Calendar != nil {
			return fmt.Errorf("heatmap chart calendar is not allowed for Cartesian coordinates")
		}
		if len(cfg.XAxis) == 0 || len(cfg.YAxis) == 0 {
			return fmt.Errorf("heatmap chart x and y axes are required for Cartesian coordinates")
		}
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("heatmap chart series is required")
	}
	for seriesIndex, series := range cfg.Series {
		if series.Name == "" {
			return fmt.Errorf("heatmap chart series %d name is required", seriesIndex)
		}
		if len(series.Data) == 0 {
			return fmt.Errorf("heatmap chart series %q data is required", series.Name)
		}
		for dataIndex, point := range series.Data {
			if !point.Missing && !internalinteractive.FiniteNumber(point.Value) {
				return fmt.Errorf("heatmap chart series %q data point %d value must be finite", series.Name, dataIndex)
			}
			if cfg.Coordinate == CoordinateCalendar {
				if point.Date.IsZero() {
					return fmt.Errorf("heatmap chart series %q data point %d date is required", series.Name, dataIndex)
				}
				date := dateString(point.Date)
				if date < dateString(cfg.Calendar.Start) || date > dateString(cfg.Calendar.End) {
					return fmt.Errorf("heatmap chart series %q data point %d date is outside the calendar range", series.Name, dataIndex)
				}
			} else if point.X < 0 || point.X >= len(cfg.XAxis) || point.Y < 0 || point.Y >= len(cfg.YAxis) {
				return fmt.Errorf("heatmap chart series %q data point %d category indexes are outside the axes", series.Name, dataIndex)
			}
		}
	}
	return nil
}

func dateString(value time.Time) string { return value.Format("2006-01-02") }

func boolString(value bool) string { return strconv.FormatBool(value) }
