// Package radar renders accessible server-side SVG radar charts.
package radar

import (
	"fmt"
	"math"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

// Indicator is one named radar axis with an explicit upper bound.
type Indicator struct {
	Name  string
	Min   float64
	Max   float64
	Label IndicatorLabelOptions
}

// IndicatorLabelOptions controls one radar-axis label.
type IndicatorLabelOptions struct {
	FontSize float64
}

// ValueLabels controls labels drawn at radar vertices.
type ValueLabels string

const (
	// ValueLabelsDefault inherits the chart-level choice or hides labels globally.
	ValueLabelsDefault ValueLabels = ""
	// ValueLabelsHidden hides labels at radar vertices.
	ValueLabelsHidden ValueLabels = "hidden"
	// ValueLabelsShown draws exact values at radar vertices.
	ValueLabelsShown ValueLabels = "shown"
)

// Options controls chart-wide radar presentation.
//
// RadiusPercent accepts 1 through 100. Zero uses the renderer's 40 percent
// default. ValueLabels defaults to hidden.
type Options struct {
	RadiusPercent float64
	ValueLabels   ValueLabels
	ValueFormat   ValueFormat
}

// SeriesOptions controls presentation for one series.
type SeriesOptions struct {
	ValueLabels   ValueLabels
	ValueFormat   ValueFormat
	LabelFontSize float64
}

// ValueFormat selects renderer-neutral formatting for labels at radar vertices.
type ValueFormat string

const (
	// ValueFormatDefault keeps the renderer's compact default.
	ValueFormatDefault ValueFormat = ""
	// ValueFormatExact uses the shortest decimal representation without scaling.
	ValueFormatExact ValueFormat = "exact"
	// ValueFormatInteger rounds rendered labels to whole numbers.
	ValueFormatInteger ValueFormat = "integer"
	// ValueFormatHumanized uses compact suffixes such as k and M.
	ValueFormatHumanized ValueFormat = "humanized"
)

// Placement selects a logical edge or center. Start and end become left and
// right for horizontal placement, and top and bottom for vertical placement.
type Placement string

const (
	PlacementDefault Placement = ""
	PlacementStart   Placement = "start"
	PlacementCenter  Placement = "center"
	PlacementEnd     Placement = "end"
)

// Alignment controls legend marker-and-text alignment.
type Alignment string

const (
	AlignmentDefault Alignment = ""
	AlignmentStart   Alignment = "start"
	AlignmentCenter  Alignment = "center"
	AlignmentEnd     Alignment = "end"
)

// LegendOrientation controls legend flow.
type LegendOrientation string

const (
	LegendHorizontal LegendOrientation = ""
	LegendVertical   LegendOrientation = "vertical"
)

// Padding controls chart or legend inset in pixels. Zero preserves renderer defaults.
type Padding struct{ Top, Right, Bottom, Left int }

// TitleOptions controls visible chart title and subtitle presentation.
type TitleOptions struct {
	Text            string
	Subtext         string
	Hidden          bool
	Horizontal      Placement
	Vertical        Placement
	FontSize        float64
	SubtextFontSize float64
	BorderWidth     float64
}

// LegendOptions controls renderer-neutral legend layout.
type LegendOptions struct {
	Hidden      bool
	Orientation LegendOrientation
	Horizontal  Placement
	Vertical    Placement
	Alignment   Alignment
	FontSize    float64
	Padding     Padding
	Overlay     bool
	BorderWidth float64
}

// Series is one named vector aligned with Config.Indicators.
type Series struct {
	Name    string
	Values  []float64
	Options SeriesOptions
}

// Config describes an SSR SVG radar chart.
//
// Label is required and becomes the accessible figure name. Caption remains
// visible. Style selects the theme palette and accepts caller color and root
// class overrides. RootAttrs accepts additional figure attributes except
// class, role, and aria-label, which remain owned by the component.
type Config struct {
	Label      string
	Caption    string
	Indicators []Indicator
	Series     []Series
	Options    Options
	Title      TitleOptions
	Legend     LegendOptions
	Padding    Padding
	Width      int
	Height     int
	Style      charttheme.Style
	RootAttrs  templ.Attributes
	// Controls configures shared controls; Expand defaults on while fullscreen defaults off.
	Controls chartcontrol.Options
	// Export customizes or disables default SVG and PNG export.
	Export *chartcontrol.ExportOptions
}

func (cfg Config) validate() error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("radar chart label is required")
	}
	if len(cfg.Indicators) < 3 {
		return fmt.Errorf("radar chart needs at least three indicators")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("radar chart needs at least one series")
	}
	if cfg.Width < 0 {
		return fmt.Errorf("radar chart width cannot be negative")
	}
	if cfg.Height < 0 {
		return fmt.Errorf("radar chart height cannot be negative")
	}
	if err := cfg.Options.validate("radar chart options"); err != nil {
		return err
	}
	if err := cfg.Title.validate("radar chart title"); err != nil {
		return err
	}
	if err := cfg.Legend.validate("radar chart legend"); err != nil {
		return err
	}
	if err := cfg.Padding.validate("radar chart padding"); err != nil {
		return err
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("radar chart root attribute %q is reserved", attribute)
			}
		}
	}
	indicatorNames := make(map[string]struct{}, len(cfg.Indicators))
	for index, indicator := range cfg.Indicators {
		if strings.TrimSpace(indicator.Name) == "" {
			return fmt.Errorf("radar chart indicator %d needs a name", index+1)
		}
		if _, exists := indicatorNames[indicator.Name]; exists {
			return fmt.Errorf("radar chart indicator %q is duplicated", indicator.Name)
		}
		indicatorNames[indicator.Name] = struct{}{}
		if !finite(indicator.Min) {
			return fmt.Errorf("radar chart indicator %q min must be finite", indicator.Name)
		}
		if !finite(indicator.Max) || indicator.Max <= 0 {
			return fmt.Errorf("radar chart indicator %q max must be a finite positive number", indicator.Name)
		}
		if indicator.Max <= indicator.Min {
			return fmt.Errorf("radar chart indicator %q max must be greater than min", indicator.Name)
		}
		if !finite(indicator.Label.FontSize) || indicator.Label.FontSize < 0 {
			return fmt.Errorf("radar chart indicator %q label font size must be a finite non-negative number", indicator.Name)
		}
	}
	for index, series := range cfg.Series {
		if strings.TrimSpace(series.Name) == "" {
			return fmt.Errorf("radar chart series %d needs a name", index+1)
		}
		if len(series.Values) != len(cfg.Indicators) {
			return fmt.Errorf("radar chart series %q has %d values; need %d", series.Name, len(series.Values), len(cfg.Indicators))
		}
		if err := series.Options.validate(fmt.Sprintf("radar chart series %q options", series.Name)); err != nil {
			return err
		}
		for valueIndex, value := range series.Values {
			if !finite(value) {
				return fmt.Errorf("radar chart series %q value %d must be finite", series.Name, valueIndex+1)
			}
			indicator := cfg.Indicators[valueIndex]
			if value < indicator.Min && indicator.Min == 0 {
				return fmt.Errorf("radar chart series %q value %d cannot be negative", series.Name, valueIndex+1)
			}
			if value < indicator.Min {
				return fmt.Errorf("radar chart series %q value %d is below indicator %q min %v", series.Name, valueIndex+1, indicator.Name, indicator.Min)
			}
			if value > indicator.Max {
				return fmt.Errorf("radar chart series %q value %d exceeds indicator %q max %v", series.Name, valueIndex+1, indicator.Name, indicator.Max)
			}
		}
	}
	return nil
}

func (options Options) validate(prefix string) error {
	if !finite(options.RadiusPercent) || options.RadiusPercent < 0 || options.RadiusPercent > 100 || (options.RadiusPercent > 0 && options.RadiusPercent < 1) {
		return fmt.Errorf("%s radius percent must be zero or between 1 and 100", prefix)
	}
	if err := options.ValueLabels.validate(prefix); err != nil {
		return err
	}
	return options.ValueFormat.validate(prefix)
}

func (options SeriesOptions) validate(prefix string) error {
	if err := options.ValueLabels.validate(prefix); err != nil {
		return err
	}
	if err := options.ValueFormat.validate(prefix); err != nil {
		return err
	}
	if !finite(options.LabelFontSize) || options.LabelFontSize < 0 {
		return fmt.Errorf("%s label font size must be a finite non-negative number", prefix)
	}
	return nil
}

func (format ValueFormat) validate(prefix string) error {
	switch format {
	case ValueFormatDefault, ValueFormatExact, ValueFormatInteger, ValueFormatHumanized:
		return nil
	default:
		return fmt.Errorf("%s has unsupported value format %q", prefix, format)
	}
}

func (title TitleOptions) validate(prefix string) error {
	if err := title.Horizontal.validate(prefix + " horizontal"); err != nil {
		return err
	}
	if err := title.Vertical.validate(prefix + " vertical"); err != nil {
		return err
	}
	if !finite(title.FontSize) || title.FontSize < 0 {
		return fmt.Errorf("%s font size must be a finite non-negative number", prefix)
	}
	if !finite(title.SubtextFontSize) || title.SubtextFontSize < 0 {
		return fmt.Errorf("%s subtext font size must be a finite non-negative number", prefix)
	}
	if !finite(title.BorderWidth) || title.BorderWidth < 0 {
		return fmt.Errorf("%s border width must be a finite non-negative number", prefix)
	}
	return nil
}

func (legend LegendOptions) validate(prefix string) error {
	if legend.Orientation != LegendHorizontal && legend.Orientation != LegendVertical {
		return fmt.Errorf("%s orientation %q is unsupported", prefix, legend.Orientation)
	}
	if err := legend.Horizontal.validate(prefix + " horizontal"); err != nil {
		return err
	}
	if err := legend.Vertical.validate(prefix + " vertical"); err != nil {
		return err
	}
	switch legend.Alignment {
	case AlignmentDefault, AlignmentStart, AlignmentCenter, AlignmentEnd:
	default:
		return fmt.Errorf("%s alignment %q is unsupported", prefix, legend.Alignment)
	}
	if !finite(legend.FontSize) || legend.FontSize < 0 {
		return fmt.Errorf("%s font size must be a finite non-negative number", prefix)
	}
	if !finite(legend.BorderWidth) || legend.BorderWidth < 0 {
		return fmt.Errorf("%s border width must be a finite non-negative number", prefix)
	}
	return legend.Padding.validate(prefix + " padding")
}

func (placement Placement) validate(prefix string) error {
	switch placement {
	case PlacementDefault, PlacementStart, PlacementCenter, PlacementEnd:
		return nil
	default:
		return fmt.Errorf("%s placement %q is unsupported", prefix, placement)
	}
}

func (padding Padding) validate(prefix string) error {
	if padding.Top < 0 || padding.Right < 0 || padding.Bottom < 0 || padding.Left < 0 {
		return fmt.Errorf("%s cannot be negative", prefix)
	}
	return nil
}

func (labels ValueLabels) validate(prefix string) error {
	switch labels {
	case ValueLabelsDefault, ValueLabelsHidden, ValueLabelsShown:
		return nil
	default:
		return fmt.Errorf("%s has unsupported value labels %q", prefix, labels)
	}
}

func (labels ValueLabels) resolve(fallback ValueLabels) ValueLabels {
	if labels != ValueLabelsDefault {
		return labels
	}
	if fallback != ValueLabelsDefault {
		return fallback
	}
	return ValueLabelsHidden
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func (cfg Config) width() int {
	if cfg.Width > 0 {
		return cfg.Width
	}
	return 600
}

func (cfg Config) height() int {
	if cfg.Height > 0 {
		return cfg.Height
	}
	return 400
}
