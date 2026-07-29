package interactive

import (
	"fmt"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// BarOrientation selects whether categories run left to right or top to bottom.
type BarOrientation string

const (
	// BarOrientationVertical places categories on the horizontal axis.
	BarOrientationVertical BarOrientation = "vertical"
	// BarOrientationHorizontal places categories on the vertical axis.
	BarOrientationHorizontal BarOrientation = "horizontal"
)

// BarZoomMode selects direct gesture exploration or a visible range slider.
type BarZoomMode string

const (
	// BarZoomInside supports wheel, pinch, and drag exploration inside the plot.
	BarZoomInside BarZoomMode = "inside"
	// BarZoomSlider exposes a visible range control.
	BarZoomSlider BarZoomMode = "slider"
)

// BarZoom selects an initial percentage window over ordered categories.
type BarZoom struct {
	Mode         BarZoomMode
	StartPercent float64
	EndPercent   float64
}

// BarStatistic selects a calculated series reference.
type BarStatistic string

const (
	// BarStatisticMinimum selects the smallest series value.
	BarStatisticMinimum BarStatistic = "minimum"
	// BarStatisticMaximum selects the largest series value.
	BarStatisticMaximum BarStatistic = "maximum"
	// BarStatisticAverage selects the arithmetic mean of series values.
	BarStatisticAverage BarStatistic = "average"
)

// BarCoordinate identifies one category and finite value.
type BarCoordinate struct {
	Category string
	Value    float64
}

// BarPointReference places either a calculated or explicit point on a series.
// Exactly one of Statistic and Coordinate must be configured.
type BarPointReference struct {
	Name       string
	Statistic  BarStatistic
	Coordinate *BarCoordinate
	Label      *LabelOptions
}

// BarGuideReference draws a calculated guide across a series.
type BarGuideReference struct {
	Name      string
	Statistic BarStatistic
}

// BarReferences configures calculated points, explicit points, and guides.
type BarReferences struct {
	Points     []BarPointReference
	Lines      []BarGuideReference
	ShowLabels *bool
}

// BarConfig describes an accessible, browser-rendered bar chart.
//
// Values must be application-owned because the browser renderer serializes them.
type BarConfig struct {
	Label         string
	Caption       string
	XAxis         []string
	Series        []BarSeries
	Orientation   BarOrientation
	Zoom          *BarZoom
	Width         string
	Height        string
	Options       ChartOptions
	SeriesOptions SeriesOptions
	Style         charttheme.Style
	Live          *LiveData
}

// BarSeries describes one named bar series.
type BarSeries struct {
	Name       string
	Data       []BarData
	Options    SeriesOptions
	References BarReferences
}

// BarData describes one finite bar value and optional per-point presentation.
type BarData struct {
	Name      string
	Value     float64
	Label     *LabelOptions
	ItemStyle *ItemStyle
	Tooltip   *TooltipOptions
}

// Bar builds a reusable interactive bar component.
func Bar(cfg BarConfig) Instance {
	if err := validateBarConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveBar, err)
	}

	chart := charts.NewBar()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, chartGlobalOptions(cfg.Options)...)
	if cfg.Zoom != nil {
		zoom := opts.DataZoom{Type: string(cfg.Zoom.Mode), Start: float32(cfg.Zoom.StartPercent), End: float32(cfg.Zoom.EndPercent)}
		if cfg.Orientation == BarOrientationHorizontal {
			zoom.YAxisIndex = 0
		} else {
			zoom.XAxisIndex = 0
		}
		globalOptions = append(globalOptions, charts.WithDataZoomOpts(zoom))
	}
	if len(cfg.Style.Colors) > 0 {
		globalOptions = append(globalOptions, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
	if cfg.Width != "" || cfg.Height != "" {
		globalOptions = append([]charts.GlobalOpts{
			charts.WithInitializationOpts(opts.Initialization{
				Width:  cfg.Width,
				Height: cfg.Height,
			}),
		}, globalOptions...)
	}
	chart.SetGlobalOptions(globalOptions...)
	chart.SetXAxis(cfg.XAxis)
	if cfg.Orientation == BarOrientationHorizontal {
		chart.XYReversal()
	}
	for _, series := range cfg.Series {
		data := make([]opts.BarData, len(series.Data))
		for index, point := range series.Data {
			data[index] = opts.BarData{Name: point.Name, Value: point.Value}
			if point.Label != nil {
				label := rendererLabel(point.Label)
				data[index].Label = &label
			}
			if point.ItemStyle != nil {
				style := rendererItemStyle(point.ItemStyle)
				data[index].ItemStyle = &style
			}
			if point.Tooltip != nil {
				tooltip := rendererTooltip(point.Tooltip)
				data[index].Tooltip = &tooltip
			}
		}
		seriesOptions := mergeSeriesOptions(cfg.SeriesOptions, series.Options)
		seriesOptions = append(seriesOptions, rendererBarReferences(series.References)...)
		chart.AddSeries(series.Name, data, seriesOptions...)
	}

	return newInstance(chartcomponents.KindInteractiveBar, renderConfig{
		Label:              cfg.Label,
		Caption:            cfg.Caption,
		Chart:              chart,
		ResponsiveWidth:    responsiveWidth(cfg.Width),
		Style:              cfg.Style,
		Live:               cartesianLiveConfig(cfg.Live),
		Details:            barDetails(cfg),
		Animation:          cfg.Options.Animation,
		Controls:           cfg.Options.Controls,
		Export:             cfg.Options.Export,
		AxisLabelIntervals: axisLabelIntervals(cfg.Options),
	})
}

func validateBarConfig(cfg BarConfig) error {
	if err := validateChartOptions(cfg.Options); err != nil {
		return err
	}
	if err := validateLiveData(cfg.Live); err != nil {
		return err
	}
	if cfg.Orientation != "" && cfg.Orientation != BarOrientationVertical && cfg.Orientation != BarOrientationHorizontal {
		return fmt.Errorf("bar chart orientation %q is not supported", cfg.Orientation)
	}
	if cfg.Zoom != nil {
		if cfg.Zoom.Mode != BarZoomInside && cfg.Zoom.Mode != BarZoomSlider {
			return fmt.Errorf("bar chart zoom mode %q is not supported", cfg.Zoom.Mode)
		}
		if !finiteNumber(cfg.Zoom.StartPercent) || !finiteNumber(cfg.Zoom.EndPercent) || cfg.Zoom.StartPercent < 0 || cfg.Zoom.EndPercent > 100 || cfg.Zoom.StartPercent >= cfg.Zoom.EndPercent {
			return fmt.Errorf("bar chart zoom range must satisfy 0 <= start < end <= 100")
		}
	}
	if len(cfg.XAxis) == 0 {
		return fmt.Errorf("bar chart x axis is required")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("bar chart series is required")
	}
	for index, series := range cfg.Series {
		if series.Name == "" {
			return fmt.Errorf("bar chart series %d name is required", index)
		}
		if len(series.Data) != len(cfg.XAxis) {
			return fmt.Errorf("bar chart series %q has %d values for %d x-axis categories", series.Name, len(series.Data), len(cfg.XAxis))
		}
		for dataIndex, point := range series.Data {
			if !finiteNumber(point.Value) {
				return fmt.Errorf("bar chart series %q data point %d value must be finite", series.Name, dataIndex)
			}
		}
		if err := validateBarReferences(series.References, cfg.XAxis); err != nil {
			return fmt.Errorf("bar chart series %q: %w", series.Name, err)
		}
	}
	return nil
}

func validateBarStatistic(statistic BarStatistic) error {
	switch statistic {
	case BarStatisticMinimum, BarStatisticMaximum, BarStatisticAverage:
		return nil
	default:
		return fmt.Errorf("statistic %q is not supported", statistic)
	}
}

func validateBarReferences(references BarReferences, categories []string) error {
	categorySet := make(map[string]struct{}, len(categories))
	for _, category := range categories {
		categorySet[category] = struct{}{}
	}
	for index, point := range references.Points {
		if point.Name == "" {
			return fmt.Errorf("point reference %d name is required", index)
		}
		modes := 0
		if point.Statistic != "" {
			modes++
			if err := validateBarStatistic(point.Statistic); err != nil {
				return fmt.Errorf("point reference %d %w", index, err)
			}
		}
		if point.Coordinate != nil {
			modes++
			if _, ok := categorySet[point.Coordinate.Category]; !ok {
				return fmt.Errorf("point reference %d category %q is not on the category axis", index, point.Coordinate.Category)
			}
			if !finiteNumber(point.Coordinate.Value) {
				return fmt.Errorf("point reference %d value must be finite", index)
			}
		}
		if modes != 1 {
			return fmt.Errorf("point reference %d requires exactly one reference mode", index)
		}
	}
	for index, line := range references.Lines {
		if line.Name == "" {
			return fmt.Errorf("guide reference %d name is required", index)
		}
		if err := validateBarStatistic(line.Statistic); err != nil {
			return fmt.Errorf("guide reference %d %w", index, err)
		}
	}
	return nil
}

func barRendererStatistic(statistic BarStatistic) string {
	switch statistic {
	case BarStatisticMinimum:
		return "min"
	case BarStatisticMaximum:
		return "max"
	case BarStatisticAverage:
		return "average"
	default:
		return ""
	}
}

func rendererBarReferences(references BarReferences) []charts.SeriesOpts {
	result := make([]charts.SeriesOpts, 0, 4)
	statisticPoints := make([]opts.MarkPointNameTypeItem, 0, len(references.Points))
	coordinatePoints := make([]opts.MarkPointNameCoordItem, 0, len(references.Points))
	for _, point := range references.Points {
		if point.Statistic != "" {
			statisticPoints = append(statisticPoints, opts.MarkPointNameTypeItem{Name: point.Name, Type: barRendererStatistic(point.Statistic)})
			continue
		}
		value := opts.MarkPointNameCoordItem{Name: point.Name, Coordinate: []interface{}{point.Coordinate.Category, point.Coordinate.Value}}
		if point.Label != nil {
			label := rendererLabel(point.Label)
			value.Label = &label
		}
		coordinatePoints = append(coordinatePoints, value)
	}
	if len(statisticPoints) > 0 {
		result = append(result, charts.WithMarkPointNameTypeItemOpts(statisticPoints...))
	}
	if len(coordinatePoints) > 0 {
		result = append(result, charts.WithMarkPointNameCoordItemOpts(coordinatePoints...))
	}
	if references.ShowLabels != nil {
		result = append(result, charts.WithMarkPointStyleOpts(opts.MarkPointStyle{Label: &opts.Label{Show: opts.Bool(*references.ShowLabels)}}))
	}
	lines := make([]opts.MarkLineNameTypeItem, len(references.Lines))
	for index, line := range references.Lines {
		lines[index] = opts.MarkLineNameTypeItem{Name: line.Name, Type: barRendererStatistic(line.Statistic)}
	}
	if len(lines) > 0 {
		result = append(result, charts.WithMarkLineNameTypeItemOpts(lines...))
	}
	return result
}

func barDetails(cfg BarConfig) templ.Component {
	if cfg.Live != nil {
		return nil
	}
	return barExactValues(cfg)
}
