// Package bar provides the canonical interactive categorical-bar API.
//
// Bar-specific types and implementation live here. Shared renderer-neutral
// options remain in components/chart, while components/interactive preserves
// its legacy Bar-prefixed aliases and constructor.
package bar

import (
	"fmt"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// Instance is the renderer-neutral chart instance returned by Bar.
type Instance = chart.Instance

// Orientation selects whether categories run left to right or top to bottom.
type Orientation string

const (
	// OrientationVertical places categories on the horizontal axis.
	OrientationVertical Orientation = "vertical"
	// OrientationHorizontal places categories on the vertical axis.
	OrientationHorizontal Orientation = "horizontal"
)

// ZoomMode selects direct gesture exploration or a visible range slider.
type ZoomMode string

const (
	// ZoomInside supports wheel, pinch, and drag exploration inside the plot.
	ZoomInside ZoomMode = "inside"
	// ZoomSlider exposes a visible range control.
	ZoomSlider ZoomMode = "slider"
)

// Zoom selects an initial percentage window over ordered categories.
type Zoom struct {
	Mode         ZoomMode
	StartPercent float64
	EndPercent   float64
}

// Statistic selects a calculated series reference.
type Statistic string

const (
	// StatisticMinimum selects the smallest series value.
	StatisticMinimum Statistic = "minimum"
	// StatisticMaximum selects the largest series value.
	StatisticMaximum Statistic = "maximum"
	// StatisticAverage selects the arithmetic mean of series values.
	StatisticAverage Statistic = "average"
)

// Coordinate identifies one category and finite value.
type Coordinate struct {
	Category string
	Value    float64
}

// PointReference places either a calculated or explicit point on a series.
// Exactly one of Statistic and Coordinate must be configured.
type PointReference struct {
	Name       string
	Statistic  Statistic
	Coordinate *Coordinate
	Label      *chart.LabelOptions
}

// GuideReference draws a calculated guide across a series.
type GuideReference struct {
	Name      string
	Statistic Statistic
}

// References configures calculated points, explicit points, and guides.
type References struct {
	Points     []PointReference
	Lines      []GuideReference
	ShowLabels *bool
}

// Config describes an accessible, browser-rendered bar chart.
//
// Values must be application-owned because the browser renderer serializes them.
type Config struct {
	Label         string
	Caption       string
	XAxis         []string
	Series        []Series
	Orientation   Orientation
	Zoom          *Zoom
	Width         string
	Height        string
	Options       chart.ChartOptions
	SeriesOptions chart.SeriesOptions
	Style         charttheme.Style
	Live          *chart.LiveData
}

// Series describes one named bar series.
type Series struct {
	Name       string
	Data       []Data
	Options    chart.SeriesOptions
	References References
}

// Data describes one finite bar value and optional per-point presentation.
type Data struct {
	Name      string
	Value     float64
	Label     *chart.LabelOptions
	ItemStyle *chart.ItemStyle
	Tooltip   *chart.TooltipOptions
}

// Bar builds a reusable interactive bar component.
func Bar(cfg Config) Instance {
	if err := validateConfig(cfg); err != nil {
		return internalinteractive.Invalid(chartcomponents.KindInteractiveBar, err)
	}

	renderer := charts.NewBar()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, internalinteractive.ChartGlobalOptions(cfg.Options)...)
	if cfg.Zoom != nil {
		zoom := opts.DataZoom{Type: string(cfg.Zoom.Mode), Start: float32(cfg.Zoom.StartPercent), End: float32(cfg.Zoom.EndPercent)}
		if cfg.Orientation == OrientationHorizontal {
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
	renderer.SetGlobalOptions(globalOptions...)
	renderer.SetXAxis(cfg.XAxis)
	if cfg.Orientation == OrientationHorizontal {
		renderer.XYReversal()
	}
	for _, series := range cfg.Series {
		data := make([]opts.BarData, len(series.Data))
		for index, point := range series.Data {
			data[index] = opts.BarData{Name: point.Name, Value: point.Value}
			if point.Label != nil {
				label := internalinteractive.RendererLabel(point.Label)
				data[index].Label = &label
			}
			if point.ItemStyle != nil {
				style := internalinteractive.RendererItemStyle(point.ItemStyle)
				data[index].ItemStyle = &style
			}
			if point.Tooltip != nil {
				tooltip := internalinteractive.RendererTooltip(point.Tooltip)
				data[index].Tooltip = &tooltip
			}
		}
		seriesOptions := internalinteractive.MergeSeriesOptions(cfg.SeriesOptions, series.Options)
		seriesOptions = append(seriesOptions, rendererReferences(series.References)...)
		renderer.AddSeries(series.Name, data, seriesOptions...)
	}

	return internalinteractive.New(chartcomponents.KindInteractiveBar, internalinteractive.RenderConfig{
		Label:              cfg.Label,
		Caption:            cfg.Caption,
		Chart:              renderer,
		ResponsiveWidth:    internalinteractive.ResponsiveWidth(cfg.Width),
		Style:              cfg.Style,
		Live:               internalinteractive.CartesianLiveConfig(cfg.Live),
		Details:            details(cfg),
		Animation:          cfg.Options.Animation,
		Controls:           cfg.Options.Controls,
		Export:             cfg.Options.Export,
		AxisLabelIntervals: internalinteractive.AxisLabelIntervals(cfg.Options),
	})
}

func validateConfig(cfg Config) error {
	if err := internalinteractive.ValidateChartOptions(cfg.Options); err != nil {
		return err
	}
	if err := internalinteractive.ValidateLiveData(cfg.Live); err != nil {
		return err
	}
	if cfg.Orientation != "" && cfg.Orientation != OrientationVertical && cfg.Orientation != OrientationHorizontal {
		return fmt.Errorf("bar chart orientation %q is not supported", cfg.Orientation)
	}
	if cfg.Zoom != nil {
		if cfg.Zoom.Mode != ZoomInside && cfg.Zoom.Mode != ZoomSlider {
			return fmt.Errorf("bar chart zoom mode %q is not supported", cfg.Zoom.Mode)
		}
		if !internalinteractive.FiniteNumber(cfg.Zoom.StartPercent) || !internalinteractive.FiniteNumber(cfg.Zoom.EndPercent) || cfg.Zoom.StartPercent < 0 || cfg.Zoom.EndPercent > 100 || cfg.Zoom.StartPercent >= cfg.Zoom.EndPercent {
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
			if !internalinteractive.FiniteNumber(point.Value) {
				return fmt.Errorf("bar chart series %q data point %d value must be finite", series.Name, dataIndex)
			}
		}
		if err := validateReferences(series.References, cfg.XAxis); err != nil {
			return fmt.Errorf("bar chart series %q: %w", series.Name, err)
		}
	}
	return nil
}

func validateStatistic(statistic Statistic) error {
	switch statistic {
	case StatisticMinimum, StatisticMaximum, StatisticAverage:
		return nil
	default:
		return fmt.Errorf("statistic %q is not supported", statistic)
	}
}

func validateReferences(references References, categories []string) error {
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
			if err := validateStatistic(point.Statistic); err != nil {
				return fmt.Errorf("point reference %d %w", index, err)
			}
		}
		if point.Coordinate != nil {
			modes++
			if _, ok := categorySet[point.Coordinate.Category]; !ok {
				return fmt.Errorf("point reference %d category %q is not on the category axis", index, point.Coordinate.Category)
			}
			if !internalinteractive.FiniteNumber(point.Coordinate.Value) {
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
		if err := validateStatistic(line.Statistic); err != nil {
			return fmt.Errorf("guide reference %d %w", index, err)
		}
	}
	return nil
}

func rendererStatistic(statistic Statistic) string {
	switch statistic {
	case StatisticMinimum:
		return "min"
	case StatisticMaximum:
		return "max"
	case StatisticAverage:
		return "average"
	default:
		return ""
	}
}

func rendererReferences(references References) []charts.SeriesOpts {
	result := make([]charts.SeriesOpts, 0, 4)
	statisticPoints := make([]opts.MarkPointNameTypeItem, 0, len(references.Points))
	coordinatePoints := make([]opts.MarkPointNameCoordItem, 0, len(references.Points))
	for _, point := range references.Points {
		if point.Statistic != "" {
			statisticPoints = append(statisticPoints, opts.MarkPointNameTypeItem{Name: point.Name, Type: rendererStatistic(point.Statistic)})
			continue
		}
		value := opts.MarkPointNameCoordItem{Name: point.Name, Coordinate: []interface{}{point.Coordinate.Category, point.Coordinate.Value}}
		if point.Label != nil {
			label := internalinteractive.RendererLabel(point.Label)
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
		lines[index] = opts.MarkLineNameTypeItem{Name: line.Name, Type: rendererStatistic(line.Statistic)}
	}
	if len(lines) > 0 {
		result = append(result, charts.WithMarkLineNameTypeItemOpts(lines...))
		if references.ShowLabels != nil {
			result = append(result, charts.WithMarkLineStyleOpts(opts.MarkLineStyle{Label: &opts.Label{Show: opts.Bool(*references.ShowLabels)}}))
		}
	}
	return result
}

func details(cfg Config) templ.Component {
	if cfg.Live != nil {
		return nil
	}
	return exactValues(cfg)
}
