// Package radar renders accessible server-side SVG radar charts.
package radar

import (
	"fmt"
	"math"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

// Indicator is one named radar axis with an explicit upper bound.
type Indicator struct {
	Name string
	Max  float64
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
}

// SeriesOptions controls presentation for one series.
type SeriesOptions struct {
	ValueLabels ValueLabels
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
	Width      int
	Height     int
	Style      charttheme.Style
	RootAttrs  templ.Attributes
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
		if !finite(indicator.Max) || indicator.Max <= 0 {
			return fmt.Errorf("radar chart indicator %q max must be a finite positive number", indicator.Name)
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
			if value < 0 {
				return fmt.Errorf("radar chart series %q value %d cannot be negative", series.Name, valueIndex+1)
			}
			indicator := cfg.Indicators[valueIndex]
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
	return options.ValueLabels.validate(prefix)
}

func (options SeriesOptions) validate(prefix string) error {
	return options.ValueLabels.validate(prefix)
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
