// Package heatmap renders accessible server-side SVG heat maps.
package heatmap

import (
	"fmt"
	"math"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

// Axis defines one categorical heat-map axis.
type Axis struct {
	Title  string
	Labels []string
}

// Cell identifies one value by zero-based X and Y category indexes.
type Cell struct {
	X     int
	Y     int
	Value float64
}

// ValueRange defines the inclusive accepted and visualized value domain.
type ValueRange struct {
	Min float64
	Max float64
}

// GradientStop defines one ordered renderer-neutral sequential-color stop.
// At is a normalized position from 0 through 1. Color accepts any CSS color;
// Class is added to cells influenced by the stop for application overrides.
type GradientStop struct {
	At    float64
	Color string
	Class string
}

// Gradient defines a sequential value scale. Its zero value uses the
// theme-aware cold-to-warm scale. Reverse swaps cold and warm direction.
type Gradient struct {
	Stops   []GradientStop
	Reverse bool
}

// Config describes an SSR SVG heat map.
//
// Supply either Rows or Cells. Rows must match the Y-by-X axis dimensions.
// Cells must cover every axis coordinate exactly once. Label is required and
// names the figure; Caption remains visible. RootAttrs cannot replace class,
// role, or aria-label, which remain component-owned.
type Config struct {
	Label      string
	Caption    string
	Title      string
	XAxis      Axis
	YAxis      Axis
	Rows       [][]float64
	Cells      []Cell
	ValueRange ValueRange
	Gradient   Gradient
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
		return fmt.Errorf("heat map label is required")
	}
	if len(cfg.XAxis.Labels) == 0 || len(cfg.YAxis.Labels) == 0 {
		return fmt.Errorf("heat map x and y labels are required")
	}
	if cfg.Width < 0 || cfg.Height < 0 {
		return fmt.Errorf("heat map dimensions cannot be negative")
	}
	if !finite(cfg.ValueRange.Min) || !finite(cfg.ValueRange.Max) || cfg.ValueRange.Min >= cfg.ValueRange.Max {
		return fmt.Errorf("heat map value range needs finite min and max with min less than max")
	}
	if len(cfg.Rows) == 0 && len(cfg.Cells) == 0 {
		return fmt.Errorf("heat map rows or cells are required")
	}
	if len(cfg.Rows) > 0 && len(cfg.Cells) > 0 {
		return fmt.Errorf("heat map rows and cells cannot be combined")
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("heat map root attribute %q is reserved", attribute)
			}
		}
	}
	if err := validateLabels("x", cfg.XAxis.Labels); err != nil {
		return err
	}
	if err := validateLabels("y", cfg.YAxis.Labels); err != nil {
		return err
	}
	if len(cfg.Rows) > 0 {
		if len(cfg.Rows) != len(cfg.YAxis.Labels) {
			return fmt.Errorf("heat map has %d rows; need %d y labels", len(cfg.Rows), len(cfg.YAxis.Labels))
		}
		for y, row := range cfg.Rows {
			if len(row) != len(cfg.XAxis.Labels) {
				return fmt.Errorf("heat map row %d has %d values; need %d x labels", y, len(row), len(cfg.XAxis.Labels))
			}
			for x, value := range row {
				if err := cfg.validateValue(value, x, y); err != nil {
					return err
				}
			}
		}
	} else {
		expected := len(cfg.XAxis.Labels) * len(cfg.YAxis.Labels)
		if len(cfg.Cells) != expected {
			return fmt.Errorf("heat map has %d cells; need %d to cover the axis matrix", len(cfg.Cells), expected)
		}
		seen := make(map[[2]int]struct{}, expected)
		for index, cell := range cfg.Cells {
			if cell.X < 0 || cell.X >= len(cfg.XAxis.Labels) || cell.Y < 0 || cell.Y >= len(cfg.YAxis.Labels) {
				return fmt.Errorf("heat map cell %d indexes are outside the axes", index)
			}
			key := [2]int{cell.X, cell.Y}
			if _, exists := seen[key]; exists {
				return fmt.Errorf("heat map cell %d duplicates indexes (%d, %d)", index, cell.X, cell.Y)
			}
			seen[key] = struct{}{}
			if err := cfg.validateValue(cell.Value, cell.X, cell.Y); err != nil {
				return err
			}
		}
	}
	return cfg.Gradient.validate()
}

func validateLabels(axis string, labels []string) error {
	for index, label := range labels {
		if strings.TrimSpace(label) == "" {
			return fmt.Errorf("heat map %s label %d is empty", axis, index)
		}
	}
	return nil
}

func (cfg Config) validateValue(value float64, x, y int) error {
	if !finite(value) {
		return fmt.Errorf("heat map value at (%d, %d) must be finite", x, y)
	}
	if value < cfg.ValueRange.Min || value > cfg.ValueRange.Max {
		return fmt.Errorf("heat map value at (%d, %d) is outside the configured range", x, y)
	}
	return nil
}

func (gradient Gradient) validate() error {
	if len(gradient.Stops) == 0 {
		return nil
	}
	if len(gradient.Stops) < 2 {
		return fmt.Errorf("heat map gradient needs at least two stops")
	}
	for index, stop := range gradient.Stops {
		if !finite(stop.At) || stop.At < 0 || stop.At > 1 {
			return fmt.Errorf("heat map gradient stop %d position must be finite and between 0 and 1", index)
		}
		if strings.TrimSpace(stop.Color) == "" && strings.TrimSpace(stop.Class) == "" {
			return fmt.Errorf("heat map gradient stop %d needs a color or class", index)
		}
		if color := strings.ToLower(stop.Color); strings.ContainsAny(color, ";{}<>\\\"") || strings.Contains(color, "url(") || strings.Contains(color, "expression(") {
			return fmt.Errorf("heat map gradient stop %d color is unsafe", index)
		}
		if index > 0 && stop.At <= gradient.Stops[index-1].At {
			return fmt.Errorf("heat map gradient stops must be in strictly increasing order")
		}
	}
	if gradient.Stops[0].At != 0 || gradient.Stops[len(gradient.Stops)-1].At != 1 {
		return fmt.Errorf("heat map gradient stops must start at 0 and end at 1")
	}
	return nil
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
