// Package scatter provides the canonical interactive scatter API.
//
// Standard and effect scatter remain behavior variants of one component.
// Scatter-specific types and implementation live here; shared renderer-neutral
// options remain in components/chart.
package scatter

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

// Instance is the renderer-neutral chart instance returned by Scatter.
type Instance = chart.Instance

// Variant selects the scatter renderer without changing the component contract.
type Variant string

const (
	// VariantStandard renders ordinary scatter points. It is the default.
	VariantStandard Variant = ""
	// VariantEffect adds an animated ripple around points.
	VariantEffect Variant = "effect"
)

// AxisType selects how scatter x values are interpreted.
type AxisType string

const (
	// AxisCategory uses XAxis as an ordered category list. It is the default.
	AxisCategory AxisType = "category"
	// AxisValue reads numeric x/y coordinates from each data value.
	AxisValue AxisType = "value"
)

// Config describes an accessible, browser-rendered scatter chart.
//
// Category axes use XAxis and one scalar value per category. Value axes leave
// XAxis empty and use two-value [x, y] coordinates in Data. Values must be
// application-owned because the browser renderer serializes them.
type Config struct {
	Label         string
	Caption       string
	Variant       Variant
	XAxisType     AxisType
	XAxis         []string
	Series        []Series
	Width         string
	Height        string
	Options       chart.ChartOptions
	SeriesOptions chart.SeriesOptions
	// Ripple configures the effect variant for every series. Per-series Options run after it.
	Ripple *chart.RippleOptions
	Style  charttheme.Style
}

// Series describes one named scatter series.
type Series struct {
	Name    string
	Data    []Data
	Options chart.SeriesOptions
	// Ripple overrides the shared effect treatment for this series.
	Ripple *chart.RippleOptions
}

// Data describes either a category value or a numeric x/y coordinate.
// Category axes use Value. Value axes use X and Y.
type Data struct {
	Name         string
	Value        float64
	X            float64
	Y            float64
	Symbol       string
	SymbolSize   int
	SymbolRotate int
}

// Scatter builds a reusable interactive scatter component.
func Scatter(cfg Config) Instance {
	if err := validateConfig(cfg); err != nil {
		return internalinteractive.Invalid(chartcomponents.KindInteractiveScatter, err)
	}

	globalOptions := []charts.GlobalOpts{
		charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())),
		charts.WithXAxisOpts(opts.XAxis{Type: string(resolvedAxisType(cfg.XAxisType))}),
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
	if cfg.Variant == VariantEffect {
		renderer := charts.NewEffectScatter()
		renderer.SetGlobalOptions(globalOptions...)
		if resolvedAxisType(cfg.XAxisType) == AxisCategory {
			renderer.SetXAxis(cfg.XAxis)
		}
		for _, series := range cfg.Series {
			data := make([]opts.EffectScatterData, len(series.Data))
			for index, point := range series.Data {
				data[index] = opts.EffectScatterData{Name: point.Name, Value: scatterValue(cfg.XAxisType, point)}
			}
			renderer.AddSeries(series.Name, data, seriesOptions(cfg, series)...)
		}
		return internalinteractive.New(chartcomponents.KindInteractiveScatter, internalinteractive.RenderConfig{
			Label: cfg.Label, Caption: cfg.Caption, Chart: renderer, Style: cfg.Style,
			Details: exactValues(cfg), Animation: cfg.Options.Animation,
			Controls: cfg.Options.Controls, Export: cfg.Options.Export,
			ResponsiveWidth:    internalinteractive.ResponsiveWidth(cfg.Width),
			AxisLabelIntervals: internalinteractive.AxisLabelIntervals(cfg.Options),
		})
	}

	renderer := charts.NewScatter()
	renderer.SetGlobalOptions(globalOptions...)
	if resolvedAxisType(cfg.XAxisType) == AxisCategory {
		renderer.SetXAxis(cfg.XAxis)
	}
	for _, series := range cfg.Series {
		data := make([]opts.ScatterData, len(series.Data))
		for index, point := range series.Data {
			data[index] = opts.ScatterData{
				Name: point.Name, Value: scatterValue(cfg.XAxisType, point), Symbol: point.Symbol,
				SymbolSize: point.SymbolSize, SymbolRotate: point.SymbolRotate,
			}
		}
		renderer.AddSeries(series.Name, data, seriesOptions(cfg, series)...)
	}
	return internalinteractive.New(chartcomponents.KindInteractiveScatter, internalinteractive.RenderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: renderer, Style: cfg.Style,
		Details: exactValues(cfg), Animation: cfg.Options.Animation,
		Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		ResponsiveWidth:    internalinteractive.ResponsiveWidth(cfg.Width),
		AxisLabelIntervals: internalinteractive.AxisLabelIntervals(cfg.Options),
	})
}

func seriesOptions(cfg Config, series Series) []charts.SeriesOpts {
	options := internalinteractive.MergeSeriesOptions(cfg.SeriesOptions, chart.SeriesOptions{})
	if cfg.Ripple != nil {
		options = append(options, rendererRipple(*cfg.Ripple))
	}
	if series.Ripple != nil {
		options = append(options, rendererRipple(*series.Ripple))
	}
	return append(options, internalinteractive.ChartSeriesOptions(series.Options)...)
}

func rendererRipple(ripple chart.RippleOptions) charts.SeriesOpts {
	return charts.WithRippleEffectOpts(opts.RippleEffect{
		Period: float32(ripple.Period), Scale: float32(ripple.Scale), BrushType: ripple.BrushType,
	})
}

func scatterValue(axisType AxisType, point Data) any {
	if resolvedAxisType(axisType) == AxisValue {
		return [2]float64{point.X, point.Y}
	}
	return point.Value
}

func validateConfig(cfg Config) error {
	if err := internalinteractive.ValidateChartOptions(cfg.Options); err != nil {
		return err
	}
	if cfg.Variant != VariantStandard && cfg.Variant != VariantEffect {
		return fmt.Errorf("scatter chart variant %q is not supported", cfg.Variant)
	}
	if cfg.Ripple != nil && cfg.Variant != VariantEffect {
		return fmt.Errorf("scatter chart ripple requires the effect variant")
	}
	if cfg.Ripple != nil {
		if err := validateRipple(*cfg.Ripple); err != nil {
			return fmt.Errorf("scatter chart shared ripple: %w", err)
		}
	}
	if err := validateSymbol(cfg.SeriesOptions.Symbol); err != nil {
		return fmt.Errorf("scatter chart shared series options %w", err)
	}
	if cfg.SeriesOptions.SymbolSize < 0 {
		return fmt.Errorf("scatter chart shared series options symbol size must be nonnegative")
	}
	axisType := resolvedAxisType(cfg.XAxisType)
	if axisType != AxisCategory && axisType != AxisValue {
		return fmt.Errorf("scatter chart x axis type %q is not supported", cfg.XAxisType)
	}
	if axisType == AxisCategory && len(cfg.XAxis) == 0 {
		return fmt.Errorf("scatter chart x axis is required for category mode")
	}
	if axisType == AxisValue && len(cfg.XAxis) != 0 {
		return fmt.Errorf("scatter chart x axis categories are not allowed for value mode")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("scatter chart series is required")
	}
	for index, series := range cfg.Series {
		if series.Name == "" {
			return fmt.Errorf("scatter chart series %d name is required", index)
		}
		if len(series.Data) == 0 {
			return fmt.Errorf("scatter chart series %q data is required", series.Name)
		}
		if series.Ripple != nil {
			if cfg.Variant != VariantEffect {
				return fmt.Errorf("scatter chart series %q ripple requires the effect variant", series.Name)
			}
			if err := validateRipple(*series.Ripple); err != nil {
				return fmt.Errorf("scatter chart series %q ripple: %w", series.Name, err)
			}
		}
		if err := validateSymbol(series.Options.Symbol); err != nil {
			return fmt.Errorf("scatter chart series %q %w", series.Name, err)
		}
		if series.Options.SymbolSize < 0 {
			return fmt.Errorf("scatter chart series %q symbol size must be nonnegative", series.Name)
		}
		if axisType == AxisCategory && len(series.Data) != len(cfg.XAxis) {
			return fmt.Errorf("scatter chart series %q has %d data points for %d x-axis categories", series.Name, len(series.Data), len(cfg.XAxis))
		}
		for dataIndex, point := range series.Data {
			if axisType == AxisValue {
				if !internalinteractive.FiniteNumber(point.X) || !internalinteractive.FiniteNumber(point.Y) {
					return fmt.Errorf("scatter chart series %q data point %d must contain a numeric [x, y] coordinate", series.Name, dataIndex)
				}
			} else if !internalinteractive.FiniteNumber(point.Value) {
				return fmt.Errorf("scatter chart series %q data point %d value must be finite", series.Name, dataIndex)
			}
			if err := validatePointPresentation(point); err != nil {
				return fmt.Errorf("scatter chart series %q data point %d %w", series.Name, dataIndex, err)
			}
			if cfg.Variant == VariantEffect && (point.Symbol != "" || point.SymbolSize != 0 || point.SymbolRotate != 0) {
				return fmt.Errorf("scatter chart series %q data point %d per-point symbol presentation is unsupported for the effect variant; use series options", series.Name, dataIndex)
			}
		}
	}
	return nil
}

func validatePointPresentation(point Data) error {
	if err := validateSymbol(point.Symbol); err != nil {
		return err
	}
	if point.SymbolSize < 0 {
		return fmt.Errorf("symbol size must be nonnegative")
	}
	return nil
}

func validateSymbol(symbol string) error {
	switch symbol {
	case "", "circle", "rect", "roundRect", "triangle", "diamond", "pin", "arrow", "none":
		return nil
	default:
		return fmt.Errorf("symbol %q is not supported", symbol)
	}
}

func validateRipple(ripple chart.RippleOptions) error {
	if !internalinteractive.FiniteNumber(ripple.Period) || ripple.Period < 0 {
		return fmt.Errorf("period must be finite and nonnegative")
	}
	if !internalinteractive.FiniteNumber(ripple.Scale) || ripple.Scale < 0 {
		return fmt.Errorf("scale must be finite and nonnegative")
	}
	if ripple.BrushType != "" && ripple.BrushType != "stroke" && ripple.BrushType != "fill" {
		return fmt.Errorf("brush type %q is not supported", ripple.BrushType)
	}
	return nil
}

func resolvedAxisType(axisType AxisType) AxisType {
	if axisType == "" {
		return AxisCategory
	}
	return axisType
}

func exactValues(cfg Config) templ.Component {
	return exactValuesTemplate(cfg, resolvedAxisType(cfg.XAxisType) == AxisValue)
}

func pointName(point Data, fallback string) string {
	if point.Name != "" {
		return point.Name
	}
	return fallback
}

func pointCount(cfg Config) int {
	total := 0
	for _, series := range cfg.Series {
		total += len(series.Data)
	}
	return total
}
