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

// Series is one named sequence of values aligned with Config.Labels.
type Series struct {
	Name   string
	Values []float64
}

// Config describes an SSR SVG categorical bar chart. Vertical is the default
// orientation. Horizontal charts include an adjacent exact-value table.
type Config struct {
	Label       string
	Caption     string
	Title       string
	Labels      []string
	Series      []Series
	Stacked     bool
	Orientation Orientation
	Padding     Padding
	Width       int
	Height      int
	Style       charttheme.Style
	// Controls configures shared controls; Expand defaults on while fullscreen and collapse default off.
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
	}
	return nil
}

func (cfg Config) horizontal() bool { return cfg.Orientation == OrientationHorizontal }

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
