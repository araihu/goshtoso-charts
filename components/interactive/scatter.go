package interactive

import (
	"fmt"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// ScatterVariant selects the scatter renderer without changing the component contract.
type ScatterVariant string

const (
	// ScatterVariantStandard renders ordinary scatter points. It is the default.
	ScatterVariantStandard ScatterVariant = ""
	// ScatterVariantEffect adds an animated ripple around points.
	ScatterVariantEffect ScatterVariant = "effect"
)

// ScatterConfig describes an accessible, browser-rendered scatter chart.
//
// Category axes use XAxis and one scalar value per category. Value axes leave
// XAxis empty and use two-value [x, y] coordinates in ScatterData.Value.
// Values must be application-owned because the browser renderer serializes them.
type ScatterConfig struct {
	Label         string
	Caption       string
	Variant       ScatterVariant
	XAxisType     CartesianAxisType
	XAxis         []string
	Series        []ScatterSeries
	Width         string
	Height        string
	Options       ChartOptions
	SeriesOptions SeriesOptions
	// Ripple configures the effect variant for every series. Per-series Options run after it.
	Ripple *RippleOptions
	Style  charttheme.Style
}

// ScatterSeries describes one named scatter series.
type ScatterSeries struct {
	Name    string
	Data    []ScatterData
	Options SeriesOptions
	// Ripple overrides the shared effect treatment for this series.
	Ripple *RippleOptions
}

// ScatterData describes either a category value or a numeric x/y coordinate.
// Category axes use Value. Value axes use X and Y.
type ScatterData struct {
	Name         string
	Value        float64
	X            float64
	Y            float64
	Symbol       string
	SymbolSize   int
	SymbolRotate int
}

// Scatter builds a reusable interactive scatter component.
func Scatter(cfg ScatterConfig) Instance {
	if err := validateScatterConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveScatter, err)
	}

	globalOptions := []charts.GlobalOpts{
		charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())),
		charts.WithXAxisOpts(opts.XAxis{Type: string(resolvedCartesianAxisType(cfg.XAxisType))}),
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
	if cfg.Variant == ScatterVariantEffect {
		chart := charts.NewEffectScatter()
		chart.SetGlobalOptions(globalOptions...)
		if resolvedCartesianAxisType(cfg.XAxisType) == CartesianAxisCategory {
			chart.SetXAxis(cfg.XAxis)
		}
		for _, series := range cfg.Series {
			data := make([]opts.EffectScatterData, len(series.Data))
			for index, point := range series.Data {
				data[index] = opts.EffectScatterData{Name: point.Name, Value: scatterValue(cfg.XAxisType, point)}
			}
			chart.AddSeries(series.Name, data, scatterSeriesOptions(cfg, series)...)
		}
		return newInstance(chartcomponents.KindInteractiveScatter, renderConfig{Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, Details: scatterExactValues(cfg), Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export, ResponsiveWidth: responsiveWidth(cfg.Width), AxisLabelIntervals: axisLabelIntervals(cfg.Options)})
	}

	chart := charts.NewScatter()
	chart.SetGlobalOptions(globalOptions...)
	if resolvedCartesianAxisType(cfg.XAxisType) == CartesianAxisCategory {
		chart.SetXAxis(cfg.XAxis)
	}
	for _, series := range cfg.Series {
		data := make([]opts.ScatterData, len(series.Data))
		for index, point := range series.Data {
			data[index] = opts.ScatterData{
				Name: point.Name, Value: scatterValue(cfg.XAxisType, point), Symbol: point.Symbol,
				SymbolSize: point.SymbolSize, SymbolRotate: point.SymbolRotate,
			}
		}
		chart.AddSeries(series.Name, data, scatterSeriesOptions(cfg, series)...)
	}
	return newInstance(chartcomponents.KindInteractiveScatter, renderConfig{Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, Details: scatterExactValues(cfg), Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export, ResponsiveWidth: responsiveWidth(cfg.Width), AxisLabelIntervals: axisLabelIntervals(cfg.Options)})
}

func scatterSeriesOptions(cfg ScatterConfig, series ScatterSeries) []charts.SeriesOpts {
	options := mergeSeriesOptions(cfg.SeriesOptions, SeriesOptions{})
	if cfg.Ripple != nil {
		options = append(options, rendererScatterRipple(*cfg.Ripple))
	}
	if series.Ripple != nil {
		options = append(options, rendererScatterRipple(*series.Ripple))
	}
	return append(options, chartSeriesOptions(series.Options)...)
}

func rendererScatterRipple(ripple RippleOptions) charts.SeriesOpts {
	return charts.WithRippleEffectOpts(opts.RippleEffect{
		Period: float32(ripple.Period), Scale: float32(ripple.Scale), BrushType: ripple.BrushType,
	})
}

func scatterValue(axisType CartesianAxisType, point ScatterData) any {
	if resolvedCartesianAxisType(axisType) == CartesianAxisValue {
		return [2]float64{point.X, point.Y}
	}
	return point.Value
}

func validateScatterConfig(cfg ScatterConfig) error {
	if err := validateChartOptions(cfg.Options); err != nil {
		return err
	}
	if cfg.Variant != ScatterVariantStandard && cfg.Variant != ScatterVariantEffect {
		return fmt.Errorf("scatter chart variant %q is not supported", cfg.Variant)
	}
	if cfg.Ripple != nil && cfg.Variant != ScatterVariantEffect {
		return fmt.Errorf("scatter chart ripple requires the effect variant")
	}
	if cfg.Ripple != nil {
		if err := validateScatterRipple(*cfg.Ripple); err != nil {
			return fmt.Errorf("scatter chart shared ripple: %w", err)
		}
	}
	if err := validateScatterSymbol(cfg.SeriesOptions.Symbol); err != nil {
		return fmt.Errorf("scatter chart shared series options %w", err)
	}
	if cfg.SeriesOptions.SymbolSize < 0 {
		return fmt.Errorf("scatter chart shared series options symbol size must be nonnegative")
	}
	axisType := resolvedCartesianAxisType(cfg.XAxisType)
	if axisType != CartesianAxisCategory && axisType != CartesianAxisValue {
		return fmt.Errorf("scatter chart x axis type %q is not supported", cfg.XAxisType)
	}
	if axisType == CartesianAxisCategory && len(cfg.XAxis) == 0 {
		return fmt.Errorf("scatter chart x axis is required for category mode")
	}
	if axisType == CartesianAxisValue && len(cfg.XAxis) != 0 {
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
			if cfg.Variant != ScatterVariantEffect {
				return fmt.Errorf("scatter chart series %q ripple requires the effect variant", series.Name)
			}
			if err := validateScatterRipple(*series.Ripple); err != nil {
				return fmt.Errorf("scatter chart series %q ripple: %w", series.Name, err)
			}
		}
		if err := validateScatterSymbol(series.Options.Symbol); err != nil {
			return fmt.Errorf("scatter chart series %q %w", series.Name, err)
		}
		if series.Options.SymbolSize < 0 {
			return fmt.Errorf("scatter chart series %q symbol size must be nonnegative", series.Name)
		}
		if axisType == CartesianAxisCategory && len(series.Data) != len(cfg.XAxis) {
			return fmt.Errorf("scatter chart series %q has %d data points for %d x-axis categories", series.Name, len(series.Data), len(cfg.XAxis))
		}
		if axisType == CartesianAxisValue {
			for dataIndex, point := range series.Data {
				if !finiteNumber(point.X) || !finiteNumber(point.Y) {
					return fmt.Errorf("scatter chart series %q data point %d must contain a numeric [x, y] coordinate", series.Name, dataIndex)
				}
				if err := validateScatterPointPresentation(point); err != nil {
					return fmt.Errorf("scatter chart series %q data point %d %w", series.Name, dataIndex, err)
				}
				if cfg.Variant == ScatterVariantEffect && (point.Symbol != "" || point.SymbolSize != 0 || point.SymbolRotate != 0) {
					return fmt.Errorf("scatter chart series %q data point %d per-point symbol presentation is unsupported for the effect variant; use series options", series.Name, dataIndex)
				}
			}
		} else {
			for dataIndex, point := range series.Data {
				if !finiteNumber(point.Value) {
					return fmt.Errorf("scatter chart series %q data point %d value must be finite", series.Name, dataIndex)
				}
				if err := validateScatterPointPresentation(point); err != nil {
					return fmt.Errorf("scatter chart series %q data point %d %w", series.Name, dataIndex, err)
				}
				if cfg.Variant == ScatterVariantEffect && (point.Symbol != "" || point.SymbolSize != 0 || point.SymbolRotate != 0) {
					return fmt.Errorf("scatter chart series %q data point %d per-point symbol presentation is unsupported for the effect variant; use series options", series.Name, dataIndex)
				}
			}
		}
	}
	return nil
}

func validateScatterPointPresentation(point ScatterData) error {
	if err := validateScatterSymbol(point.Symbol); err != nil {
		return err
	}
	if point.SymbolSize < 0 {
		return fmt.Errorf("symbol size must be nonnegative")
	}
	return nil
}

func validateScatterSymbol(symbol string) error {
	switch symbol {
	case "", "circle", "rect", "roundRect", "triangle", "diamond", "pin", "arrow", "none":
		return nil
	default:
		return fmt.Errorf("symbol %q is not supported", symbol)
	}
}

func validateScatterRipple(ripple RippleOptions) error {
	if !finiteNumber(ripple.Period) || ripple.Period < 0 {
		return fmt.Errorf("period must be finite and nonnegative")
	}
	if !finiteNumber(ripple.Scale) || ripple.Scale < 0 {
		return fmt.Errorf("scale must be finite and nonnegative")
	}
	if ripple.BrushType != "" && ripple.BrushType != "stroke" && ripple.BrushType != "fill" {
		return fmt.Errorf("brush type %q is not supported", ripple.BrushType)
	}
	return nil
}

func scatterExactValues(cfg ScatterConfig) templ.Component {
	return scatterExactValuesTemplate(cfg, resolvedCartesianAxisType(cfg.XAxisType) == CartesianAxisValue)
}

func scatterPointName(point ScatterData, fallback string) string {
	if point.Name != "" {
		return point.Name
	}
	return fallback
}

func scatterPointCount(cfg ScatterConfig) int {
	total := 0
	for _, series := range cfg.Series {
		total += len(series.Data)
	}
	return total
}
