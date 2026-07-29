// Package line renders accessible server-side SVG line charts.
package line

import (
	"fmt"
	"math"
	"strings"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

// Title controls the visible chart title.
type Title struct {
	Text string
}

// AreaOptions controls optional fill below every line series.
// Opacity uses the renderer-neutral range 0..1. Zero selects the renderer's
// default opacity when Enabled is true.
type AreaOptions struct {
	Enabled bool
	Opacity float64
}

// CategoryAxisOptions controls the horizontal category axis.
// A nil BoundaryGap preserves the Line component's legacy behavior.
type CategoryAxisOptions struct {
	BoundaryGap *bool
}

// LegendOptions controls legend layout.
type LegendOptions struct {
	Padding Padding
}

// Padding is spacing in pixels.
type Padding struct{ Top, Right, Bottom, Left int }

// Axis controls one numeric Y axis. Unit is a suggested positive tick step.
// Min and Max optionally bound the axis and its assigned values. Color and
// Class are mutually exclusive presentation overrides.
type Axis struct {
	Title string
	Unit  float64
	Min   *float64
	Max   *float64
	Color string
	Class string
}

// Series is one labeled sequence of values. Values must align with Config.Labels.
type Series struct {
	Name       string
	Values     []float64
	YAxisIndex int
	Color      string
	Class      string
}

// Config describes an SSR SVG line chart.
//
// Label is required and becomes the accessible name of the rendered figure.
// Caption remains visible below the chart. Keep the data table near the chart
// when users need exact values; charts are summary views, not the only data UI.
type Config struct {
	Label   string
	Caption string
	Title   Title
	Labels  []string
	Series  []Series
	Area    AreaOptions
	XAxis   CategoryAxisOptions
	Legend  LegendOptions
	// YAxes optionally configures one or two numeric axes. Its zero value keeps
	// the original single-axis Line rendering byte-for-byte.
	YAxes  []Axis
	Width  int
	Height int
	Style  charttheme.Style
	// Controls configures shared controls; Expand defaults on while fullscreen defaults off.
	Controls chartcontrol.Options
	// Export customizes or disables default SVG and PNG export.
	Export *chartcontrol.ExportOptions
}

func (cfg Config) validate() error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("line chart label is required")
	}
	if len(cfg.Labels) == 0 {
		return fmt.Errorf("line chart needs at least one label")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("line chart needs at least one series")
	}
	if cfg.Width < 0 {
		return fmt.Errorf("line chart width cannot be negative")
	}
	if cfg.Height < 0 {
		return fmt.Errorf("line chart height cannot be negative")
	}
	if !finite(cfg.Area.Opacity) || cfg.Area.Opacity < 0 || cfg.Area.Opacity > 1 {
		return fmt.Errorf("line chart area opacity must be finite and between 0 and 1")
	}
	if !cfg.Area.Enabled && cfg.Area.Opacity != 0 {
		return fmt.Errorf("line chart area opacity requires area fill to be enabled")
	}
	if err := validatePadding("legend", cfg.Legend.Padding); err != nil {
		return err
	}
	if len(cfg.YAxes) > 2 {
		return fmt.Errorf("line chart supports at most two Y axes")
	}
	axisCount := len(cfg.YAxes)
	if axisCount == 0 {
		axisCount = 1
	}
	labels := make(map[string]struct{}, len(cfg.Labels))
	for index, label := range cfg.Labels {
		label = strings.TrimSpace(label)
		if label == "" {
			return fmt.Errorf("line chart label %d cannot be empty", index+1)
		}
		if _, exists := labels[label]; exists {
			return fmt.Errorf("line chart label %q is duplicated", label)
		}
		labels[label] = struct{}{}
	}
	seriesNames := make(map[string]struct{}, len(cfg.Series))
	usedAxes := make([]bool, axisCount)
	for index, axis := range cfg.YAxes {
		if !finite(axis.Unit) || axis.Unit < 0 {
			return fmt.Errorf("line chart Y axis %d unit must be finite and non-negative", index)
		}
		if axis.Min != nil && !finite(*axis.Min) {
			return fmt.Errorf("line chart Y axis %d minimum must be finite", index)
		}
		if axis.Max != nil && !finite(*axis.Max) {
			return fmt.Errorf("line chart Y axis %d maximum must be finite", index)
		}
		if axis.Min != nil && axis.Max != nil && *axis.Min >= *axis.Max {
			return fmt.Errorf("line chart Y axis %d minimum must be less than maximum", index)
		}
		if strings.TrimSpace(axis.Color) != "" && strings.TrimSpace(axis.Class) != "" {
			return fmt.Errorf("line chart Y axis %d cannot set both color and class", index)
		}
		if unsafeCSS(axis.Color) {
			return fmt.Errorf("line chart Y axis %d color is unsafe", index)
		}
		if unsafeClass(axis.Class) {
			return fmt.Errorf("line chart Y axis %d class is unsafe", index)
		}
	}
	for index, series := range cfg.Series {
		name := strings.TrimSpace(series.Name)
		if name == "" {
			return fmt.Errorf("line chart series %d needs a name", index+1)
		}
		if _, exists := seriesNames[name]; exists {
			return fmt.Errorf("line chart series name %q is duplicated", name)
		}
		seriesNames[name] = struct{}{}
		if len(series.Values) != len(cfg.Labels) {
			return fmt.Errorf("line chart series %q has %d values; need %d", series.Name, len(series.Values), len(cfg.Labels))
		}
		if series.YAxisIndex < 0 || series.YAxisIndex >= axisCount {
			return fmt.Errorf("line chart series %q Y axis index %d is out of bounds for %d axes", series.Name, series.YAxisIndex, axisCount)
		}
		usedAxes[series.YAxisIndex] = true
		if strings.TrimSpace(series.Color) != "" && strings.TrimSpace(series.Class) != "" {
			return fmt.Errorf("line chart series %q cannot set both color and class", series.Name)
		}
		if unsafeCSS(series.Color) {
			return fmt.Errorf("line chart series %q color is unsafe", series.Name)
		}
		if unsafeClass(series.Class) {
			return fmt.Errorf("line chart series %q class is unsafe", series.Name)
		}
		for valueIndex, value := range series.Values {
			if !finite(value) {
				return fmt.Errorf("line chart series %q value %d must be finite", series.Name, valueIndex)
			}
			if len(cfg.YAxes) > 0 {
				axis := cfg.YAxes[series.YAxisIndex]
				if axis.Min != nil && value < *axis.Min {
					return fmt.Errorf("line chart series %q value %d is below Y axis %d minimum", series.Name, valueIndex, series.YAxisIndex)
				}
				if axis.Max != nil && value > *axis.Max {
					return fmt.Errorf("line chart series %q value %d is above Y axis %d maximum", series.Name, valueIndex, series.YAxisIndex)
				}
			}
		}
	}
	for index := range cfg.YAxes {
		if !usedAxes[index] {
			return fmt.Errorf("line chart Y axis %d has no assigned series", index)
		}
	}
	return nil
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func unsafeCSS(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.ContainsAny(value, ";{}<>\\\"") || strings.Contains(value, "url(") || strings.Contains(value, "expression(")
}

func unsafeClass(value string) bool { return strings.ContainsAny(value, "\"'<>;") }

func validatePadding(name string, padding Padding) error {
	if padding.Top < 0 || padding.Right < 0 || padding.Bottom < 0 || padding.Left < 0 {
		return fmt.Errorf("line chart %s padding cannot be negative", name)
	}
	return nil
}

func (cfg Config) width() int {
	if cfg.Width > 0 {
		return cfg.Width
	}
	return 720
}

func (cfg Config) height() int {
	if cfg.Height > 0 {
		return cfg.Height
	}
	return 320
}
