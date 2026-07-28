package interactive

import (
	"fmt"
	"math"
	"time"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// HeatMapCoordinate selects the heatmap coordinate system.
type HeatMapCoordinate string

const (
	// HeatMapCoordinateCartesian maps X and Y indexes to category axes. It is the default.
	HeatMapCoordinateCartesian HeatMapCoordinate = ""
	// HeatMapCoordinateCalendar maps Date values to a calendar range.
	HeatMapCoordinateCalendar HeatMapCoordinate = "calendar"
)

// HeatMapConfig describes an accessible, browser-rendered heatmap.
//
// Values must be application-owned because the browser renderer serializes them.
type HeatMapConfig struct {
	Label         string
	Caption       string
	Coordinate    HeatMapCoordinate
	XAxis         []string
	YAxis         []string
	Calendar      *HeatMapCalendar
	ValueRange    HeatMapValueRange
	Series        []HeatMapSeries
	Width         string
	Height        string
	Options       ChartOptions
	SeriesOptions SeriesOptions
	Style         charttheme.Style
}

// HeatMapCalendar defines an inclusive calendar date range. Options customizes
// calendar presentation; its Range is replaced by Start and End.
type HeatMapCalendar struct {
	Start   time.Time
	End     time.Time
	Options CalendarOptions
}

// HeatMapValueRange defines the inclusive visual-map and accepted data domain.
type HeatMapValueRange struct {
	Min float64
	Max float64
}

// HeatMapSeries describes one named heatmap series.
type HeatMapSeries struct {
	Name    string
	Data    []HeatMapData
	Options SeriesOptions
}

// HeatMapData describes one heatmap cell. Cartesian mode uses X and Y category
// indexes. Calendar mode uses Date. Value must fall within ValueRange.
type HeatMapData struct {
	X     int
	Y     int
	Date  time.Time
	Value float64
}

// HeatMap builds a reusable interactive Cartesian or calendar heatmap component.
func HeatMap(cfg HeatMapConfig) Instance {
	if err := validateHeatMapConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveHeatMap, err)
	}

	chart := charts.NewHeatMap()
	globalOptions := heatMapGlobalOptions(cfg)
	chart.SetGlobalOptions(globalOptions...)
	normalizeHeatMapVisualMap(chart, cfg)
	if cfg.Coordinate == HeatMapCoordinateCalendar {
		calendar := rendererCalendar(cfg.Calendar.Options)
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
			if cfg.Coordinate == HeatMapCoordinateCalendar {
				data[index] = opts.HeatMapData{Value: [2]interface{}{dateString(point.Date), point.Value}}
			} else {
				data[index] = opts.HeatMapData{Value: [3]interface{}{point.X, point.Y, point.Value}}
			}
		}
		options := mergeSeriesOptions(cfg.SeriesOptions, series.Options)
		if cfg.Coordinate == HeatMapCoordinateCalendar {
			// Coordinate is a component invariant, not an escape-hatch option.
			options = append(options, charts.WithCoordinateSystem("calendar"))
		}
		chart.AddSeries(series.Name, data, options...)
	}

	return newInstance(chartcomponents.KindInteractiveHeatMap, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export, ResponsiveWidth: responsiveWidth(cfg.Width),
	})
}

func heatMapGlobalOptions(cfg HeatMapConfig) []charts.GlobalOpts {
	options := []charts.GlobalOpts{
		charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())),
	}
	if cfg.Coordinate == HeatMapCoordinateCartesian {
		options = append(options,
			charts.WithXAxisOpts(opts.XAxis{Type: "category"}),
			charts.WithYAxisOpts(opts.YAxis{Type: "category", Data: cfg.YAxis}),
		)
	}
	options = append(options, chartGlobalOptions(cfg.Options)...)
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

func normalizeHeatMapVisualMap(chart *charts.HeatMap, cfg HeatMapConfig) {
	visualMap := heatMapVisualMap(cfg)
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

func heatMapVisualMap(cfg HeatMapConfig) opts.VisualMap {
	return opts.VisualMap{
		Min: float32(cfg.ValueRange.Min), Max: float32(cfg.ValueRange.Max),
		InRange: &opts.VisualMapInRange{Color: cfg.Style.ResolvedColors()},
	}
}

func validateHeatMapConfig(cfg HeatMapConfig) error {
	if cfg.Label == "" {
		return fmt.Errorf("heatmap chart label is required")
	}
	if cfg.Coordinate != HeatMapCoordinateCartesian && cfg.Coordinate != HeatMapCoordinateCalendar {
		return fmt.Errorf("heatmap chart coordinate %q is not supported", cfg.Coordinate)
	}
	if !finiteHeatMapNumber(cfg.ValueRange.Min) || !finiteHeatMapNumber(cfg.ValueRange.Max) || cfg.ValueRange.Min >= cfg.ValueRange.Max {
		return fmt.Errorf("heatmap chart value range must contain finite min and max with min less than max")
	}
	if math.Abs(cfg.ValueRange.Min) > math.MaxFloat32 || math.Abs(cfg.ValueRange.Max) > math.MaxFloat32 {
		return fmt.Errorf("heatmap chart value range exceeds renderer limits")
	}
	if cfg.Coordinate == HeatMapCoordinateCalendar {
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
			if !finiteHeatMapNumber(point.Value) {
				return fmt.Errorf("heatmap chart series %q data point %d value must be finite", series.Name, dataIndex)
			}
			if point.Value < cfg.ValueRange.Min || point.Value > cfg.ValueRange.Max {
				return fmt.Errorf("heatmap chart series %q data point %d value is outside the configured range", series.Name, dataIndex)
			}
			if cfg.Coordinate == HeatMapCoordinateCalendar {
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

func finiteHeatMapNumber(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func dateString(value time.Time) string { return value.Format("2006-01-02") }
