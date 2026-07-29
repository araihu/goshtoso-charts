// Package bar renders accessible server-side SVG bar charts.
package bar

import (
	"fmt"
	"math"
	"strings"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

// Orientation selects the categorical bar direction.
type Orientation string

const (
	// OrientationVertical keeps categories on the horizontal axis.
	OrientationVertical Orientation = ""
	// OrientationHorizontal keeps categories on the vertical axis.
	OrientationHorizontal Orientation = "horizontal"
)

// Padding controls chart inset in pixels. Its zero value keeps renderer defaults.
type Padding struct{ Top, Right, Bottom, Left int }

// ValueFormat selects renderer-neutral formatting for reference annotation values.
type ValueFormat string

const (
	// ValueFormatDefault keeps the renderer's default numeric formatting.
	ValueFormatDefault ValueFormat = ""
	// ValueFormatHumanized rounds values to zero decimal places for compact labels.
	ValueFormatHumanized ValueFormat = "humanized"
)

// DataLabelPosition selects which end of a bar anchors visible value labels.
// End is the default: above vertical bars and after horizontal bars. Start
// anchors labels at the category-axis side without exposing physical renderer positions.
type DataLabelPosition string

const (
	// DataLabelPositionEnd anchors labels at the value end.
	DataLabelPositionEnd DataLabelPosition = ""
	// DataLabelPositionStart anchors labels at the category-axis side.
	DataLabelPositionStart DataLabelPosition = "start"
)

// DataLabelOptions controls exact labels rendered at individual bars.
type DataLabelOptions struct {
	Show   bool
	Format ValueFormat
}

// GeometryOptions controls bar thickness, group spacing, and value-end caps.
// Ratios use the renderer-neutral range 0..1; zero thickness selects automatic sizing.
// A nil gap selects automatic spacing while a pointer to zero removes the gap.
type GeometryOptions struct {
	ThicknessRatio float64
	GapRatio       *float64
	RoundedCaps    bool
}

// LegendPlacement selects horizontal legend placement.
type LegendPlacement string

const (
	// LegendPlacementDefault preserves the renderer-neutral automatic placement.
	LegendPlacementDefault LegendPlacement = ""
	// LegendPlacementStart aligns the legend to the start edge.
	LegendPlacementStart LegendPlacement = "start"
	// LegendPlacementCenter centers the legend.
	LegendPlacementCenter LegendPlacement = "center"
	// LegendPlacementEnd aligns the legend to the end edge.
	LegendPlacementEnd LegendPlacement = "end"
)

// LegendOptions controls renderer-neutral legend visibility and placement.
type LegendOptions struct {
	Hidden    bool
	Placement LegendPlacement
	Overlay   bool
}

// ValueAxisOptions controls the numeric axis shared by the bar series.
type ValueAxisOptions struct {
	Hidden bool
}

// ReferenceStyle controls semantic presentation for one series' reference annotations.
// Color overrides the series token used by the rendered marks. Class is applied to
// the adjacent evidence row, allowing caller CSS without exposing renderer details.
type ReferenceStyle struct {
	Color string
	Class string
}

// References selects statistical annotations for one series. Average renders a
// reference line; Minimum and Maximum render reference points. Duplicate or
// disabled annotations are resolved deterministically from these booleans.
type References struct {
	Average       bool
	Minimum       bool
	Maximum       bool
	MaximumLine   bool
	GlobalMaximum bool
	PointPrefix   string
	Format        ValueFormat
	PointSize     int
	Style         ReferenceStyle
}

// Series is one named sequence of values aligned with Config.Labels.
type Series struct {
	Name       string
	Values     []float64
	Labels     DataLabelOptions
	References References
}

// Config describes an SSR SVG categorical bar chart. Vertical is the default
// orientation. Horizontal charts include an adjacent exact-value table.
type Config struct {
	Label         string
	Caption       string
	Title         string
	Labels        []string
	Series        []Series
	Stacked       bool
	Orientation   Orientation
	Geometry      GeometryOptions
	LabelPosition DataLabelPosition
	Legend        LegendOptions
	ValueAxis     ValueAxisOptions
	Padding       Padding
	Width         int
	Height        int
	Style         charttheme.Style
	// Controls configures shared controls; Expand defaults on while fullscreen defaults off.
	Controls chartcontrol.Options
	// Export customizes or disables default SVG and PNG export.
	Export *chartcontrol.ExportOptions
}

func (cfg Config) validate() error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("bar chart label is required")
	}
	if len(cfg.Labels) == 0 {
		return fmt.Errorf("bar chart needs at least one label")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("bar chart needs at least one series")
	}
	if cfg.Orientation != OrientationVertical && cfg.Orientation != OrientationHorizontal {
		return fmt.Errorf("bar chart orientation %q is unsupported", cfg.Orientation)
	}
	if math.IsNaN(cfg.Geometry.ThicknessRatio) || math.IsInf(cfg.Geometry.ThicknessRatio, 0) {
		return fmt.Errorf("bar chart thickness ratio must be finite")
	}
	if cfg.Geometry.ThicknessRatio < 0 || cfg.Geometry.ThicknessRatio > 1 {
		return fmt.Errorf("bar chart thickness ratio must be between 0 and 1")
	}
	if cfg.Geometry.GapRatio != nil {
		if math.IsNaN(*cfg.Geometry.GapRatio) || math.IsInf(*cfg.Geometry.GapRatio, 0) {
			return fmt.Errorf("bar chart gap ratio must be finite")
		}
		if *cfg.Geometry.GapRatio < 0 || *cfg.Geometry.GapRatio > 1 {
			return fmt.Errorf("bar chart gap ratio must be between 0 and 1")
		}
	}
	if cfg.LabelPosition != DataLabelPositionEnd && cfg.LabelPosition != DataLabelPositionStart {
		return fmt.Errorf("bar chart label position %q is unsupported", cfg.LabelPosition)
	}
	if cfg.Legend.Placement != LegendPlacementDefault && cfg.Legend.Placement != LegendPlacementStart && cfg.Legend.Placement != LegendPlacementCenter && cfg.Legend.Placement != LegendPlacementEnd {
		return fmt.Errorf("bar chart legend placement %q is unsupported", cfg.Legend.Placement)
	}
	if cfg.Padding.Top < 0 || cfg.Padding.Right < 0 || cfg.Padding.Bottom < 0 || cfg.Padding.Left < 0 {
		return fmt.Errorf("bar chart padding cannot be negative")
	}
	if cfg.Width < 0 {
		return fmt.Errorf("bar chart width cannot be negative")
	}
	if cfg.Height < 0 {
		return fmt.Errorf("bar chart height cannot be negative")
	}
	labels := make(map[string]struct{}, len(cfg.Labels))
	for index, label := range cfg.Labels {
		if strings.TrimSpace(label) == "" {
			return fmt.Errorf("bar chart label %d cannot be empty", index+1)
		}
		if _, exists := labels[label]; exists {
			return fmt.Errorf("bar chart label %q is duplicated", label)
		}
		labels[label] = struct{}{}
	}
	for index, series := range cfg.Series {
		if strings.TrimSpace(series.Name) == "" {
			return fmt.Errorf("bar chart series %d needs a name", index+1)
		}
		if len(series.Values) != len(cfg.Labels) {
			return fmt.Errorf("bar chart series %q has %d values; need %d", series.Name, len(series.Values), len(cfg.Labels))
		}
		for valueIndex, value := range series.Values {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return fmt.Errorf("bar chart series %q value %d must be finite", series.Name, valueIndex)
			}
		}
		if series.References.Format != ValueFormatDefault && series.References.Format != ValueFormatHumanized {
			return fmt.Errorf("bar chart series %q reference format %q is unsupported", series.Name, series.References.Format)
		}
		if series.Labels.Format != ValueFormatDefault && series.Labels.Format != ValueFormatHumanized {
			return fmt.Errorf("bar chart series %q data-label format %q is unsupported", series.Name, series.Labels.Format)
		}
		if series.References.GlobalMaximum && !cfg.Stacked {
			return fmt.Errorf("bar chart series %q global maximum reference requires stacked bars", series.Name)
		}
		if series.References.GlobalMaximum && index != len(cfg.Series)-1 {
			return fmt.Errorf("bar chart series %q global maximum reference must be on the last series", series.Name)
		}
		if series.References.PointSize < 0 {
			return fmt.Errorf("bar chart series %q reference point size cannot be negative", series.Name)
		}
		if unsafeCSS(series.References.Style.Color) {
			return fmt.Errorf("bar chart series %q reference color is unsafe", series.Name)
		}
		if strings.ContainsAny(series.References.Style.Class, "\"'<>;") {
			return fmt.Errorf("bar chart series %q reference class is unsafe", series.Name)
		}
	}
	return nil
}

func (cfg Config) horizontal() bool { return cfg.Orientation == OrientationHorizontal }

func (cfg Config) hasReferences() bool {
	for _, series := range cfg.Series {
		if series.References.Average || series.References.Minimum || series.References.Maximum || series.References.MaximumLine || series.References.GlobalMaximum {
			return true
		}
	}
	return false
}

func (cfg Config) hasGlobalMaximum() bool {
	for _, series := range cfg.Series {
		if series.References.GlobalMaximum {
			return true
		}
	}
	return false
}

func (cfg Config) hasSummaryReferences() bool {
	for _, series := range cfg.Series {
		if series.References.Average || series.References.Minimum || series.References.Maximum {
			return true
		}
	}
	return false
}

func (cfg Config) hasMaximumLine() bool {
	for _, series := range cfg.Series {
		if series.References.MaximumLine {
			return true
		}
	}
	return false
}

func unsafeCSS(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.ContainsAny(value, ";{}<>\\\"") || strings.Contains(value, "url(") || strings.Contains(value, "expression(")
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
