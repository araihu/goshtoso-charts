// Package radar provides the canonical interactive radar API.
//
// Default polygon, explicit polygon, and circle coordinates remain variants of
// one component. Radar-specific types and implementation live here; shared
// renderer-neutral options remain in components/chart.
package radar

import (
	"fmt"
	"strconv"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// Instance is the renderer-neutral chart instance returned by Radar.
type Instance = chart.Instance

// Config describes an accessible, browser-rendered radar chart.
//
// Values must be application-owned because the browser renderer serializes them.
type Config struct {
	Label         string
	Caption       string
	Indicators    []Indicator
	Coordinate    CoordinateOptions
	Series        []Series
	Width         string
	Height        string
	Options       chart.ChartOptions
	SeriesOptions chart.SeriesOptions
	Style         charttheme.Style
}

// Shape controls the coordinate boundary geometry.
type Shape string

const (
	// ShapeDefault preserves the standard polygon boundary.
	ShapeDefault Shape = ""
	// ShapePolygon renders straight-sided concentric boundaries.
	ShapePolygon Shape = "polygon"
	// ShapeCircle renders circular concentric boundaries.
	ShapeCircle Shape = "circle"
)

// CoordinateOptions configures the shared bounded dimensions without
// exposing renderer-specific coordinate types.
type CoordinateOptions struct {
	Shape       Shape
	SplitNumber int
	SplitArea   *bool
	SplitLine   *SplitLineOptions
}

// SplitLineOptions configures concentric coordinate guides.
type SplitLineOptions struct {
	Show  *bool
	Style *chart.LineStyle
}

// Indicator describes one named radar dimension and its positive maximum.
type Indicator struct {
	Name string
	Max  float32
}

// Series describes one named radar series.
type Series struct {
	Name    string
	Data    []Data
	Options chart.SeriesOptions
}

// Data describes one named vector whose values align with Indicators.
type Data struct {
	Name   string
	Values []float64
}

// Radar builds a reusable interactive radar component.
func Radar(cfg Config) Instance {
	if err := validateConfig(cfg); err != nil {
		return internalinteractive.Invalid(chartcomponents.KindInteractiveRadar, err)
	}

	indicators := make([]*opts.Indicator, len(cfg.Indicators))
	for index, indicator := range cfg.Indicators {
		indicators[index] = &opts.Indicator{Name: indicator.Name, Max: indicator.Max}
	}

	radarChart := charts.NewRadar()
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
			style := internalinteractive.RendererLineStyle(cfg.Coordinate.SplitLine.Style)
			radarCoordinate.SplitLine.LineStyle = &style
		}
	}

	globalOptions := []charts.GlobalOpts{
		charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())),
		charts.WithRadarComponentOpts(radarCoordinate),
	}
	globalOptions = append(globalOptions, internalinteractive.ChartGlobalOptions(cfg.Options)...)
	// Explicit component colors remain authoritative over escape-hatch options.
	if len(cfg.Style.Colors) > 0 {
		globalOptions = append(globalOptions, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
	if cfg.Width != "" || cfg.Height != "" {
		globalOptions = append([]charts.GlobalOpts{
			charts.WithInitializationOpts(opts.Initialization{Width: cfg.Width, Height: cfg.Height}),
		}, globalOptions...)
	}
	radarChart.SetGlobalOptions(globalOptions...)
	for _, series := range cfg.Series {
		data := make([]opts.RadarData, len(series.Data))
		for index, point := range series.Data {
			data[index] = opts.RadarData{Name: point.Name, Value: point.Values}
		}
		radarChart.AddSeries(series.Name, data, internalinteractive.MergeSeriesOptions(cfg.SeriesOptions, series.Options)...)
	}

	return internalinteractive.New(chartcomponents.KindInteractiveRadar, internalinteractive.RenderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: radarChart, Style: cfg.Style, Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export, ResponsiveWidth: internalinteractive.ResponsiveWidth(cfg.Width),
		Details: exactValues(cfg.Label, cfg.Indicators, detailRows(cfg.Series)),
	})
}

type valueRow struct {
	Series      string
	Observation string
	Values      []string
}

func detailRows(series []Series) []valueRow {
	rows := make([]valueRow, 0)
	for _, current := range series {
		for _, observation := range current.Data {
			values := make([]string, len(observation.Values))
			for index, value := range observation.Values {
				values[index] = strconv.FormatFloat(value, 'f', -1, 64)
			}
			rows = append(rows, valueRow{Series: current.Name, Observation: observation.Name, Values: values})
		}
	}
	return rows
}

func validateConfig(cfg Config) error {
	if err := internalinteractive.ValidateChartOptions(cfg.Options); err != nil {
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
		if !internalinteractive.FiniteNumber(float64(indicator.Max)) || indicator.Max <= 0 {
			return fmt.Errorf("radar chart indicator %q maximum must be positive", indicator.Name)
		}
	}
	if cfg.Coordinate.Shape != ShapeDefault && cfg.Coordinate.Shape != ShapePolygon && cfg.Coordinate.Shape != ShapeCircle {
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
		if style.Opacity != nil && (!internalinteractive.FiniteNumber(*style.Opacity) || *style.Opacity < 0 || *style.Opacity > 1) {
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
				if !internalinteractive.FiniteNumber(value) {
					return fmt.Errorf("radar chart series %q data %q value %d must be finite", series.Name, point.Name, valueIndex)
				}
			}
		}
	}
	return nil
}
