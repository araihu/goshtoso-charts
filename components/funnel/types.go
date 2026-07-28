// Package funnel renders accessible server-side SVG funnel charts.
package funnel

import (
	"fmt"
	"math"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

// LabelMode controls the text rendered inside each funnel stage.
type LabelMode string

const (
	// LabelPercent renders the stage label and its percentage of the first stage.
	LabelPercent LabelMode = ""
	// LabelName renders only the stage label.
	LabelName LabelMode = "name"
	// LabelValue renders only the exact stage value.
	LabelValue LabelMode = "value"
	// LabelNameValue renders the stage label and exact value.
	LabelNameValue LabelMode = "name-value"
	// LabelHidden hides labels inside the funnel; the adjacent summary remains.
	LabelHidden LabelMode = "hidden"
)

// LegendOrientation controls legend flow.
type LegendOrientation string

const (
	// LegendHorizontal lays legend entries out in one row.
	LegendHorizontal LegendOrientation = ""
	// LegendVertical stacks legend entries.
	LegendVertical LegendOrientation = "vertical"
)

// LegendPlacement controls horizontal legend alignment.
type LegendPlacement string

const (
	// LegendStart aligns the legend to the start edge.
	LegendStart LegendPlacement = ""
	// LegendCenter centers the legend.
	LegendCenter LegendPlacement = "center"
	// LegendEnd aligns the legend to the end edge.
	LegendEnd LegendPlacement = "end"
)

// Padding controls chart or legend inset in pixels.
type Padding struct{ Top, Right, Bottom, Left int }

// Legend controls renderer-neutral legend presentation.
type Legend struct {
	Hidden      bool
	Orientation LegendOrientation
	Placement   LegendPlacement
	Padding     Padding
}

// Options controls meaningful funnel presentation variants.
type Options struct {
	Labels LabelMode
	Legend Legend
	// Padding controls chart inset. Its zero value keeps the renderer default.
	Padding Padding
}

// Stage is one ordered step in a funnel. Values must not increase as the
// funnel progresses. Color overrides the palette; Class styles rendered marks
// and the matching adjacent-summary row.
type Stage struct {
	Label string
	Value float64
	Color string
	Class string
}

// Config describes one SSR SVG funnel chart.
type Config struct {
	Label     string
	Caption   string
	Title     string
	Stages    []Stage
	Options   Options
	Width     int
	Height    int
	Style     charttheme.Style
	RootAttrs templ.Attributes
	// Controls configures shared controls; Expand defaults on while fullscreen and collapse default off.
	Controls chartcontrol.Options
	// Export customizes or disables default SVG and PNG export.
	Export *chartcontrol.ExportOptions
}

func (cfg Config) validate() error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("funnel chart label is required")
	}
	if len(cfg.Stages) == 0 {
		return fmt.Errorf("funnel chart needs at least one stage")
	}
	if cfg.Width < 0 {
		return fmt.Errorf("funnel chart width cannot be negative")
	}
	if cfg.Height < 0 {
		return fmt.Errorf("funnel chart height cannot be negative")
	}
	if cfg.Options.Labels != LabelPercent && cfg.Options.Labels != LabelName && cfg.Options.Labels != LabelValue && cfg.Options.Labels != LabelNameValue && cfg.Options.Labels != LabelHidden {
		return fmt.Errorf("funnel chart label mode %q is unsupported", cfg.Options.Labels)
	}
	if cfg.Options.Legend.Orientation != LegendHorizontal && cfg.Options.Legend.Orientation != LegendVertical {
		return fmt.Errorf("funnel chart legend orientation %q is unsupported", cfg.Options.Legend.Orientation)
	}
	if cfg.Options.Legend.Placement != LegendStart && cfg.Options.Legend.Placement != LegendCenter && cfg.Options.Legend.Placement != LegendEnd {
		return fmt.Errorf("funnel chart legend placement %q is unsupported", cfg.Options.Legend.Placement)
	}
	if negativePadding(cfg.Options.Padding) {
		return fmt.Errorf("funnel chart padding cannot be negative")
	}
	if negativePadding(cfg.Options.Legend.Padding) {
		return fmt.Errorf("funnel chart legend padding cannot be negative")
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("funnel chart root attribute %q is reserved", attribute)
			}
		}
	}
	labels := make(map[string]struct{}, len(cfg.Stages))
	for index, stage := range cfg.Stages {
		label := strings.TrimSpace(stage.Label)
		if label == "" {
			return fmt.Errorf("funnel chart stage %d needs a label", index+1)
		}
		if _, exists := labels[label]; exists {
			return fmt.Errorf("funnel chart stage label %q is duplicated", label)
		}
		labels[label] = struct{}{}
		if !finite(stage.Value) || stage.Value < 0 {
			return fmt.Errorf("funnel chart stage %q value must be finite and non-negative", label)
		}
		if index > 0 && stage.Value > cfg.Stages[index-1].Value {
			return fmt.Errorf("funnel chart stage %q value cannot exceed previous stage %q", label, cfg.Stages[index-1].Label)
		}
		if unsafeCSS(stage.Color) {
			return fmt.Errorf("funnel chart stage %q color is unsafe", label)
		}
		if strings.ContainsAny(stage.Class, "\"'<>;") {
			return fmt.Errorf("funnel chart stage %q class is unsafe", label)
		}
	}
	return nil
}

func negativePadding(padding Padding) bool {
	return padding.Top < 0 || padding.Right < 0 || padding.Bottom < 0 || padding.Left < 0
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
	return 600
}

func (cfg Config) height() int {
	if cfg.Height > 0 {
		return cfg.Height
	}
	return 400
}
