// Package candlestick provides the canonical interactive OHLC chart API.
//
// Candlestick-specific types and implementation live here. Shared
// renderer-neutral options remain in components/chart, while
// components/interactive preserves its legacy aliases and constructor.
package candlestick

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	sharedchart "github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

const maxCandlestickDetailRows = 96

// Instance is the renderer-neutral chart instance returned by Candlestick.
type Instance = sharedchart.Instance

// Config describes an accessible interactive OHLC chart.
type Config struct {
	Label      string
	Caption    string
	Categories []string
	Series     []Series
	DataZoom   []DataZoom
	Width      string
	Height     string
	Options    sharedchart.ChartOptions
	Style      charttheme.Style
	RootAttrs  templ.Attributes
}

// Candle is one ordered open, close, low, and high observation.
type Candle struct {
	Open  float64
	Close float64
	Low   float64
	High  float64
}

// Series describes one named OHLC series aligned with Categories.
type Series struct {
	Name    string
	Data    []Candle
	Options SeriesOptions
}

// SeriesOptions configures candle geometry, direction styles, and marks.
type SeriesOptions struct {
	Rise        DirectionStyle
	Fall        DirectionStyle
	BorderWidth float64
	BarWidth    string
	BarMinWidth string
	BarMaxWidth string
	Marks       MarkOptions
}

// DirectionStyle configures one price direction.
// Class provides application semantics; Color and BorderColor are optional overrides.
type DirectionStyle struct {
	Color       string
	Class       string
	BorderColor string
}

// MarkOptions configures high and low reference marks.
type MarkOptions struct {
	Highest   bool
	Lowest    bool
	ShowLabel *bool
}

// DataZoomType selects a renderer-neutral zoom interaction.
type DataZoomType string

const (
	DataZoomSlider DataZoomType = ""
	DataZoomInside DataZoomType = "inside"
)

// DataZoomAxis selects which axis a zoom interaction controls.
type DataZoomAxis string

const (
	DataZoomXAxis DataZoomAxis = ""
	DataZoomYAxis DataZoomAxis = "y"
)

// DataZoom configures one bounded zoom window.
type DataZoom struct {
	Type         DataZoomType
	Axis         DataZoomAxis
	StartPercent float64
	EndPercent   float64
}

type candlestickRuntimeStyle struct {
	RiseColor       string  `json:"riseColor,omitempty"`
	RiseClass       string  `json:"riseClass,omitempty"`
	RiseBorderColor string  `json:"riseBorderColor,omitempty"`
	FallColor       string  `json:"fallColor,omitempty"`
	FallClass       string  `json:"fallClass,omitempty"`
	FallBorderColor string  `json:"fallBorderColor,omitempty"`
	BorderWidth     float64 `json:"borderWidth,omitempty"`
}

// Candlestick builds a reusable interactive OHLC component.
func Candlestick(cfg Config) Instance {
	if err := validateConfig(cfg); err != nil {
		return internalinteractive.Invalid(chartcomponents.KindInteractiveCandlestick, err)
	}

	chart := charts.NewKLine()
	width, height := cfg.Width, cfg.Height
	if width == "" {
		width = "100%"
	}
	if height == "" {
		height = "500px"
	}
	gridRight := "4%"
	for _, series := range cfg.Series {
		if series.Options.Marks.Highest || series.Options.Marks.Lowest {
			gridRight = "14%"
			break
		}
	}
	globalOptions := []charts.GlobalOpts{
		charts.WithInitializationOpts(opts.Initialization{Width: width, Height: height}),
		charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())),
		charts.WithGridOpts(opts.Grid{Top: "24%", Right: gridRight, Bottom: "15%", Left: "4%", ContainLabel: opts.Bool(true)}),
	}
	globalOptions = append(globalOptions, internalinteractive.ChartGlobalOptions(cfg.Options)...)
	for _, zoom := range cfg.DataZoom {
		privateZoom := opts.DataZoom{
			Type: string(zoom.Type), Start: float32(zoom.StartPercent), End: float32(zoom.EndPercent),
		}
		if zoom.Axis == DataZoomYAxis {
			privateZoom.YAxisIndex = []int{0}
			privateZoom.Orient = "vertical"
		} else {
			privateZoom.XAxisIndex = []int{0}
		}
		globalOptions = append(globalOptions, charts.WithDataZoomOpts(privateZoom))
	}
	chart.SetGlobalOptions(globalOptions...)
	chart.SetXAxis(cfg.Categories)

	runtimeStyles := make([]candlestickRuntimeStyle, len(cfg.Series))
	for seriesIndex, series := range cfg.Series {
		data := make([]opts.KlineData, len(series.Data))
		for index, candle := range series.Data {
			data[index] = opts.KlineData{Value: [4]float64{candle.Open, candle.Close, candle.Low, candle.High}}
		}
		runtimeStyle := runtimeCandlestickStyle(series.Options)
		runtimeStyles[seriesIndex] = runtimeStyle
		style := renderedCandlestickStyle(runtimeStyle)
		options := []charts.SeriesOpts{
			charts.WithKlineChartOpts(opts.KlineChart{
				BarWidth: series.Options.BarWidth, BarMinWidth: series.Options.BarMinWidth, BarMaxWidth: series.Options.BarMaxWidth,
			}),
			charts.WithItemStyleOpts(opts.ItemStyle{
				Color: style.RiseColor, Color0: style.FallColor,
				BorderColor: style.RiseBorderColor, BorderColor0: style.FallBorderColor,
				BorderWidth: float32(style.BorderWidth),
			}),
		}
		if series.Options.Marks.Highest {
			options = append(options, charts.WithMarkPointNameTypeItemOpts(opts.MarkPointNameTypeItem{Name: "highest value", Type: "max", ValueDim: "highest"}))
		}
		if series.Options.Marks.Lowest {
			options = append(options, charts.WithMarkPointNameTypeItemOpts(opts.MarkPointNameTypeItem{Name: "lowest value", Type: "min", ValueDim: "lowest"}))
		}
		if series.Options.Marks.ShowLabel != nil {
			options = append(options, charts.WithMarkPointStyleOpts(opts.MarkPointStyle{Label: &opts.Label{Show: opts.Bool(*series.Options.Marks.ShowLabel)}}))
		}
		chart.AddSeries(series.Name, data, options...)
	}
	stylesJSON, _ := json.Marshal(runtimeStyles)

	style := cfg.Style
	style.Class = strings.TrimSpace("goshtoso-charts-interactive-candlestick " + style.Class)
	return internalinteractive.New(chartcomponents.KindInteractiveCandlestick, internalinteractive.RenderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: style, ResponsiveWidth: internalinteractive.ResponsiveWidth(cfg.Width),
		Animation: cfg.Options.Animation, RootAttrs: cfg.RootAttrs,
		Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		AxisLabelIntervals: internalinteractive.AxisLabelIntervals(cfg.Options),
		CandlestickStyles:  string(stylesJSON),
		Details:            candlestickExactValues(candlestickDetailRows(cfg, maxCandlestickDetailRows)),
	})
}

func runtimeCandlestickStyle(options SeriesOptions) candlestickRuntimeStyle {
	return candlestickRuntimeStyle{
		RiseColor: options.Rise.Color, RiseClass: options.Rise.Class, RiseBorderColor: options.Rise.BorderColor,
		FallColor: options.Fall.Color, FallClass: options.Fall.Class, FallBorderColor: options.Fall.BorderColor,
		BorderWidth: options.BorderWidth,
	}
}

func renderedCandlestickStyle(style candlestickRuntimeStyle) candlestickRuntimeStyle {
	if style.RiseColor == "" {
		style.RiseColor = "#15803d"
	}
	if style.RiseBorderColor == "" {
		style.RiseBorderColor = style.RiseColor
	}
	if style.FallColor == "" {
		style.FallColor = "#b91c1c"
	}
	if style.FallBorderColor == "" {
		style.FallBorderColor = style.FallColor
	}
	return style
}

func validateConfig(cfg Config) error {
	if err := internalinteractive.ValidateChartOptions(cfg.Options); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("candlestick label is required")
	}
	if len(cfg.Categories) == 0 {
		return fmt.Errorf("candlestick categories are required")
	}
	for index, category := range cfg.Categories {
		if strings.TrimSpace(category) == "" {
			return fmt.Errorf("candlestick category %d is required", index)
		}
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("candlestick series is required")
	}
	names := make(map[string]struct{}, len(cfg.Series))
	for seriesIndex, series := range cfg.Series {
		name := strings.TrimSpace(series.Name)
		if name == "" {
			return fmt.Errorf("candlestick series %d name is required", seriesIndex)
		}
		if _, exists := names[name]; exists {
			return fmt.Errorf("candlestick series name %q must be unique", series.Name)
		}
		names[name] = struct{}{}
		if len(series.Data) == 0 {
			return fmt.Errorf("candlestick series %q data is required", series.Name)
		}
		if len(series.Data) != len(cfg.Categories) {
			return fmt.Errorf("candlestick series %q has %d candles for %d categories", series.Name, len(series.Data), len(cfg.Categories))
		}
		if !internalinteractive.FiniteNumber(series.Options.BorderWidth) || series.Options.BorderWidth < 0 {
			return fmt.Errorf("candlestick series %q border width must be finite and nonnegative", series.Name)
		}
		for _, width := range []struct{ name, value string }{
			{"bar width", series.Options.BarWidth}, {"minimum bar width", series.Options.BarMinWidth}, {"maximum bar width", series.Options.BarMaxWidth},
		} {
			if err := validateCandlestickWidth(width.value); err != nil {
				return fmt.Errorf("candlestick series %q %s %w", series.Name, width.name, err)
			}
		}
		for dataIndex, candle := range series.Data {
			for _, value := range []float64{candle.Open, candle.Close, candle.Low, candle.High} {
				if !internalinteractive.FiniteNumber(value) {
					return fmt.Errorf("candlestick series %q candle %d values must be finite", series.Name, dataIndex)
				}
			}
			if candle.Low > min(candle.Open, candle.Close) {
				return fmt.Errorf("candlestick series %q candle %d low must not exceed open or close", series.Name, dataIndex)
			}
			if candle.High < max(candle.Open, candle.Close) {
				return fmt.Errorf("candlestick series %q candle %d high must not be below open or close", series.Name, dataIndex)
			}
			if candle.Low > candle.High {
				return fmt.Errorf("candlestick series %q candle %d low must not exceed high", series.Name, dataIndex)
			}
		}
	}
	if tooltip := cfg.Options.Tooltip; tooltip != nil && tooltip.Trigger != "" && tooltip.Trigger != "item" && tooltip.Trigger != "axis" {
		return fmt.Errorf("candlestick tooltip trigger must be item or axis")
	}
	if legend := cfg.Options.Legend; legend != nil && legend.Orient != "" && legend.Orient != "horizontal" && legend.Orient != "vertical" {
		return fmt.Errorf("candlestick legend orientation must be horizontal or vertical")
	}
	if axis := cfg.Options.XAxis; axis != nil && axis.Type != "" && axis.Type != "category" {
		return fmt.Errorf("candlestick x axis type must be category")
	}
	if axis := cfg.Options.YAxis; axis != nil && axis.Type != "" && axis.Type != "value" && axis.Type != "log" {
		return fmt.Errorf("candlestick y axis type must be value or log")
	}
	for index, zoom := range cfg.DataZoom {
		if zoom.Type != DataZoomSlider && zoom.Type != DataZoomInside {
			return fmt.Errorf("candlestick data zoom %d type is invalid", index)
		}
		if zoom.Axis != DataZoomXAxis && zoom.Axis != DataZoomYAxis {
			return fmt.Errorf("candlestick data zoom %d axis is invalid", index)
		}
		if !internalinteractive.FiniteNumber(zoom.StartPercent) || !internalinteractive.FiniteNumber(zoom.EndPercent) ||
			zoom.StartPercent < 0 || zoom.StartPercent > 100 || zoom.EndPercent < 0 || zoom.EndPercent > 100 {
			return fmt.Errorf("candlestick data zoom %d bounds must be finite percentages from 0 to 100", index)
		}
		if zoom.StartPercent > zoom.EndPercent {
			return fmt.Errorf("candlestick data zoom %d start must not exceed end", index)
		}
	}
	return nil
}

func validateCandlestickWidth(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	number := value
	if strings.HasSuffix(number, "%") {
		number = strings.TrimSuffix(number, "%")
	} else if strings.HasSuffix(number, "px") {
		number = strings.TrimSuffix(number, "px")
	}
	parsed, err := strconv.ParseFloat(number, 64)
	if err != nil || !internalinteractive.FiniteNumber(parsed) || parsed <= 0 {
		return fmt.Errorf("must be a positive number, px length, or percentage")
	}
	return nil
}

type candlestickValueRow struct {
	Category, Series, Open, Close, Low, High string
	Direction, DirectionClass, SemanticClass string
}

type candlestickValueRows struct {
	Rows    []candlestickValueRow
	Omitted int
}

func candlestickDetailRows(cfg Config, limit int) candlestickValueRows {
	total := len(cfg.Categories) * len(cfg.Series)
	count := min(total, limit)
	rows := make([]candlestickValueRow, 0, count)
	for categoryIndex, category := range cfg.Categories {
		for _, series := range cfg.Series {
			if len(rows) == limit {
				return candlestickValueRows{Rows: rows, Omitted: total - len(rows)}
			}
			candle := series.Data[categoryIndex]
			direction, directionClass, semanticClass := "Rise", "goshtoso-charts-candlestick__direction--increasing", series.Options.Rise.Class
			if semanticClass == "" {
				semanticClass = "rise"
			}
			if candle.Close < candle.Open {
				direction, directionClass, semanticClass = "Fall", "goshtoso-charts-candlestick__direction--decreasing", series.Options.Fall.Class
				if semanticClass == "" {
					semanticClass = "fall"
				}
			}
			rows = append(rows, candlestickValueRow{
				Category: category, Series: series.Name,
				Open: strconv.FormatFloat(candle.Open, 'f', -1, 64), Close: strconv.FormatFloat(candle.Close, 'f', -1, 64),
				Low: strconv.FormatFloat(candle.Low, 'f', -1, 64), High: strconv.FormatFloat(candle.High, 'f', -1, 64),
				Direction: direction, DirectionClass: directionClass, SemanticClass: semanticClass,
			})
		}
	}
	return candlestickValueRows{Rows: rows, Omitted: total - len(rows)}
}
