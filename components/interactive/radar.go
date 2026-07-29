package interactive

import (
	"fmt"
	"math"
	"strconv"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// RadarConfig describes an accessible, browser-rendered radar chart.
//
// Values must be application-owned because the browser renderer serializes them.
type RadarConfig struct {
	Label         string
	Caption       string
	Indicators    []RadarIndicator
	Coordinate    RadarCoordinateOptions
	Series        []RadarSeries
	Width         string
	Height        string
	Options       ChartOptions
	SeriesOptions SeriesOptions
	Style         charttheme.Style
}

// RadarShape controls the coordinate boundary geometry.
type RadarShape string

const (
	// RadarShapeDefault preserves the standard polygon boundary.
	RadarShapeDefault RadarShape = ""
	// RadarShapePolygon renders straight-sided concentric boundaries.
	RadarShapePolygon RadarShape = "polygon"
	// RadarShapeCircle renders circular concentric boundaries.
	RadarShapeCircle RadarShape = "circle"
)

// RadarCoordinateOptions configures the shared bounded dimensions without
// exposing renderer-specific coordinate types.
type RadarCoordinateOptions struct {
	Shape       RadarShape
	SplitNumber int
	SplitArea   *bool
	SplitLine   *RadarSplitLineOptions
}

// RadarSplitLineOptions configures concentric coordinate guides.
type RadarSplitLineOptions struct {
	Show  *bool
	Style *LineStyle
}

// RadarIndicator describes one named radar dimension and its positive maximum.
type RadarIndicator struct {
	Name string
	Max  float32
}

// RadarSeries describes one named radar series.
type RadarSeries struct {
	Name    string
	Data    []RadarData
	Options SeriesOptions
}

// RadarData describes one named vector whose values align with Indicators.
type RadarData struct {
	Name   string
	Values []float64
}

// Radar builds a reusable interactive radar component.
func Radar(cfg RadarConfig) Instance {
	if err := validateRadarConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveRadar, err)
	}

	indicators := make([]*opts.Indicator, len(cfg.Indicators))
	for index, indicator := range cfg.Indicators {
		indicators[index] = &opts.Indicator{Name: indicator.Name, Max: indicator.Max}
	}

	chart := charts.NewRadar()
	radarCoordinate := opts.RadarComponent{Indicator: indicators, Shape: string(cfg.Coordinate.Shape), SplitNumber: cfg.Coordinate.SplitNumber}
	if cfg.Coordinate.SplitArea != nil {
		radarCoordinate.SplitArea = &opts.SplitArea{Show: opts.Bool(*cfg.Coordinate.SplitArea)}
	}
	if cfg.Coordinate.SplitLine != nil {
		radarCoordinate.SplitLine = &opts.SplitLine{}
		if cfg.Coordinate.SplitLine.Show != nil {
			radarCoordinate.SplitLine.Show = opts.Bool(*cfg.Coordinate.SplitLine.Show)
		}
		if cfg.Coordinate.SplitLine.Style != nil {
			style := rendererLineStyle(cfg.Coordinate.SplitLine.Style)
			radarCoordinate.SplitLine.LineStyle = &style
		}
	}

	globalOptions := []charts.GlobalOpts{
		charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())),
		charts.WithRadarComponentOpts(radarCoordinate),
	}
	globalOptions = append(globalOptions, chartGlobalOptions(cfg.Options)...)
	// Explicit component colors remain authoritative over escape-hatch options.
	if len(cfg.Style.Colors) > 0 {
		globalOptions = append(globalOptions, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
	if cfg.Width != "" || cfg.Height != "" {
		globalOptions = append([]charts.GlobalOpts{
			charts.WithInitializationOpts(opts.Initialization{Width: cfg.Width, Height: cfg.Height}),
		}, globalOptions...)
	}
	chart.SetGlobalOptions(globalOptions...)
	for _, series := range cfg.Series {
		data := make([]opts.RadarData, len(series.Data))
		for index, point := range series.Data {
			data[index] = opts.RadarData{Name: point.Name, Value: point.Values}
		}
		chart.AddSeries(series.Name, data, mergeSeriesOptions(cfg.SeriesOptions, series.Options)...)
	}

	return newInstance(chartcomponents.KindInteractiveRadar, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export, ResponsiveWidth: responsiveWidth(cfg.Width),
		Details: radarExactValues(cfg.Label, cfg.Indicators, radarDetailRows(cfg.Series)),
	})
}

type radarValueRow struct {
	Series      string
	Observation string
	Values      []string
}

func radarDetailRows(series []RadarSeries) []radarValueRow {
	rows := make([]radarValueRow, 0)
	for _, current := range series {
		for _, observation := range current.Data {
			values := make([]string, len(observation.Values))
			for index, value := range observation.Values {
				values[index] = strconv.FormatFloat(value, 'f', -1, 64)
			}
			rows = append(rows, radarValueRow{Series: current.Name, Observation: observation.Name, Values: values})
		}
	}
	return rows
}

func validateRadarConfig(cfg RadarConfig) error {
	if err := validateChartOptions(cfg.Options); err != nil {
		return err
	}
	if cfg.Label == "" {
		return fmt.Errorf("radar chart label is required")
	}
	if len(cfg.Indicators) == 0 {
		return fmt.Errorf("radar chart indicators are required")
	}
	for index, indicator := range cfg.Indicators {
		if indicator.Name == "" {
			return fmt.Errorf("radar chart indicator %d name is required", index)
		}
		if math.IsNaN(float64(indicator.Max)) || math.IsInf(float64(indicator.Max), 0) || indicator.Max <= 0 {
			return fmt.Errorf("radar chart indicator %q maximum must be positive", indicator.Name)
		}
	}
	if cfg.Coordinate.Shape != RadarShapeDefault && cfg.Coordinate.Shape != RadarShapePolygon && cfg.Coordinate.Shape != RadarShapeCircle {
		return fmt.Errorf("radar chart shape %q is not supported", cfg.Coordinate.Shape)
	}
	if cfg.Coordinate.SplitNumber < 0 {
		return fmt.Errorf("radar chart split number must be nonnegative")
	}
	if splitLine := cfg.Coordinate.SplitLine; splitLine != nil && splitLine.Style != nil {
		style := splitLine.Style
		if style.Width < 0 {
			return fmt.Errorf("radar chart split-line width must be nonnegative")
		}
		if style.Opacity != nil && (!finiteNumber(*style.Opacity) || *style.Opacity < 0 || *style.Opacity > 1) {
			return fmt.Errorf("radar chart split-line opacity must be between 0 and 1")
		}
		if style.Type != "" && style.Type != "solid" && style.Type != "dashed" && style.Type != "dotted" {
			return fmt.Errorf("radar chart split-line type %q is not supported", style.Type)
		}
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("radar chart series is required")
	}
	for seriesIndex, series := range cfg.Series {
		if series.Name == "" {
			return fmt.Errorf("radar chart series %d name is required", seriesIndex)
		}
		if len(series.Data) == 0 {
			return fmt.Errorf("radar chart series %q data is required", series.Name)
		}
		for dataIndex, point := range series.Data {
			if point.Name == "" {
				return fmt.Errorf("radar chart series %q data point %d name is required", series.Name, dataIndex)
			}
			if len(point.Values) == 0 {
				return fmt.Errorf("radar chart series %q data %q values are required", series.Name, point.Name)
			}
			if len(point.Values) != len(cfg.Indicators) {
				return fmt.Errorf("radar chart series %q data %q has %d values for %d indicators", series.Name, point.Name, len(point.Values), len(cfg.Indicators))
			}
			for valueIndex, value := range point.Values {
				if math.IsNaN(value) || math.IsInf(value, 0) {
					return fmt.Errorf("radar chart series %q data %q value %d must be finite", series.Name, point.Name, valueIndex)
				}
			}
		}
	}
	return nil
}
