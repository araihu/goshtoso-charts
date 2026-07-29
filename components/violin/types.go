// Package violin renders accessible server-side SVG violin charts.
package violin

import (
	"fmt"
	"math"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

// Normalization controls how density widths compare across series.
type Normalization string

const (
	// NormalizePerSeries scales each distribution to its own maximum density.
	NormalizePerSeries Normalization = "series"
	// NormalizeGlobal scales every distribution to one shared maximum density.
	NormalizeGlobal Normalization = "global"
)

// Distribution controls deterministic Gaussian kernel-density estimation.
type Distribution struct {
	// Points is the number of density bands. Zero uses 80.
	Points int
	// Bandwidth overrides the automatic bandwidth when greater than zero.
	Bandwidth float64
	// Normalization defaults to NormalizePerSeries.
	Normalization Normalization
}

// MarkLines controls statistical lines drawn through each distribution.
type MarkLines struct {
	Mean   bool
	Median bool
}

// Statistics controls additional exact summaries adjacent to the chart.
// Quantiles use values strictly between zero and one, such as 0.25 and 0.75.
type Statistics struct {
	Quantiles []float64
}

// Series is one named population of exact samples.
type Series struct {
	Name    string
	Samples []float64
	// Color overrides the theme color. Class is applied to rendered density marks
	// and the adjacent summary row for semantic caller styling.
	Color      string
	Class      string
	Marks      MarkLines
	Statistics Statistics
}

// Axis controls the symmetric density axis presentation.
type Axis struct {
	Title      string
	Limit      float64
	LabelCount int
}

// Padding controls chart inset in pixels. Its zero value uses the renderer default.
type Padding struct{ Top, Right, Bottom, Left int }

// Config describes one SSR SVG violin chart.
type Config struct {
	Label     string
	Caption   string
	Title     string
	Series    []Series
	Density   Distribution
	Axis      Axis
	Padding   Padding
	Width     int
	Height    int
	Style     charttheme.Style
	RootAttrs templ.Attributes
	// Controls configures shared controls; Expand defaults on while fullscreen defaults off.
	Controls chartcontrol.Options
	// Export customizes or disables default SVG and PNG export.
	Export *chartcontrol.ExportOptions
}

func (cfg Config) validate() error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("violin chart label is required")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("violin chart needs at least one series")
	}
	if cfg.Width < 0 || cfg.Height < 0 {
		return fmt.Errorf("violin chart dimensions cannot be negative")
	}
	if cfg.Density.Points < 0 || cfg.Density.Points == 1 {
		return fmt.Errorf("violin chart density points must be zero or at least 2")
	}
	if !finite(cfg.Density.Bandwidth) || cfg.Density.Bandwidth < 0 {
		return fmt.Errorf("violin chart bandwidth must be finite and non-negative")
	}
	if cfg.Density.Normalization != "" && cfg.Density.Normalization != NormalizePerSeries && cfg.Density.Normalization != NormalizeGlobal {
		return fmt.Errorf("violin chart normalization %q is unsupported", cfg.Density.Normalization)
	}
	if !finite(cfg.Axis.Limit) || cfg.Axis.Limit < 0 {
		return fmt.Errorf("violin chart axis limit must be finite and non-negative")
	}
	if cfg.Axis.LabelCount < 0 || cfg.Axis.LabelCount == 1 {
		return fmt.Errorf("violin chart axis label count must be zero or at least 2")
	}
	for _, value := range []int{cfg.Padding.Top, cfg.Padding.Right, cfg.Padding.Bottom, cfg.Padding.Left} {
		if value < 0 {
			return fmt.Errorf("violin chart padding cannot be negative")
		}
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("violin chart root attribute %q is reserved", attribute)
			}
		}
	}
	names := make(map[string]struct{}, len(cfg.Series))
	for index, series := range cfg.Series {
		name := strings.TrimSpace(series.Name)
		if name == "" {
			return fmt.Errorf("violin chart series %d needs a name", index+1)
		}
		if _, exists := names[name]; exists {
			return fmt.Errorf("violin chart series name %q is duplicated", name)
		}
		names[name] = struct{}{}
		if len(series.Samples) < 2 {
			return fmt.Errorf("violin chart series %q needs at least 2 samples", name)
		}
		minimum, maximum := series.Samples[0], series.Samples[0]
		for sampleIndex, sample := range series.Samples {
			if !finite(sample) {
				return fmt.Errorf("violin chart series %q sample %d must be finite", name, sampleIndex)
			}
			minimum, maximum = math.Min(minimum, sample), math.Max(maximum, sample)
		}
		if minimum == maximum {
			return fmt.Errorf("violin chart series %q needs varying samples", name)
		}
		if unsafeCSS(series.Color) {
			return fmt.Errorf("violin chart series %q color is unsafe", name)
		}
		if strings.ContainsAny(series.Class, "\"'<>;") {
			return fmt.Errorf("violin chart series %q class is unsafe", name)
		}
		seenQuantiles := map[float64]bool{}
		for _, quantile := range series.Statistics.Quantiles {
			if !finite(quantile) || quantile <= 0 || quantile >= 1 {
				return fmt.Errorf("violin chart series %q quantiles must be finite and between 0 and 1", name)
			}
			if seenQuantiles[quantile] {
				return fmt.Errorf("violin chart series %q quantile %g is duplicated", name, quantile)
			}
			seenQuantiles[quantile] = true
		}
	}
	return nil
}

func unsafeCSS(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.ContainsAny(value, ";{}<>\\\"") || strings.Contains(value, "url(") || strings.Contains(value, "expression(")
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
func (cfg Config) width() int {
	if cfg.Width > 0 {
		return cfg.Width
	}
	return 1200
}
func (cfg Config) height() int {
	if cfg.Height > 0 {
		return cfg.Height
	}
	return 800
}
func (density Distribution) points() int {
	if density.Points > 0 {
		return density.Points
	}
	return 80
}
