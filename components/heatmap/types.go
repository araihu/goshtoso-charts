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
	Title         string
	Labels        []string
	TitleFontSize float64
	LabelFontSize float64
	// LabelRotation is expressed in degrees.
	LabelRotation        float64
	LabelCount           int
	LabelCountAdjustment int
}

// Placement selects logical horizontal title placement.
type Placement string

const (
	// PlacementDefault preserves the centered title used by the upstream example.
	PlacementDefault Placement = ""
	PlacementStart   Placement = "start"
	PlacementCenter  Placement = "center"
	PlacementEnd     Placement = "end"
)

// Padding controls chart inset in pixels. Its zero value preserves renderer defaults.
type Padding struct{ Top, Right, Bottom, Left int }

// TitleOptions controls finite title presentation without exposing renderer types.
type TitleOptions struct {
	Subtext         string
	Hidden          bool
	Placement       Placement
	FontSize        float64
	SubtextFontSize float64
	BorderWidth     float64
}

// ValueFormat selects safe built-in cell-label formatting.
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

// Offset shifts cell labels from their cell center in pixels.
type Offset struct{ Left, Top int }

// ValueLabelOptions controls labels drawn inside heat-map cells.
//
// Format is intentionally a finite built-in choice. Arbitrary formatter and
// per-cell styling callbacks remain private renderer details.
type ValueLabelOptions struct {
	Show          bool
	Format        ValueFormat
	Decimals      int
	TrailingZeros bool
	FontSize      float64
	Distance      int
	Offset        Offset
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
	Label        string
	Caption      string
	Title        string
	XAxis        Axis
	YAxis        Axis
	Rows         [][]float64
	Cells        []Cell
	ValueRange   ValueRange
	Gradient     Gradient
	TitleOptions TitleOptions
	ValueLabels  ValueLabelOptions
	Padding      Padding
	Width        int
	Height       int
	Style        charttheme.Style
	RootAttrs    templ.Attributes
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
	if err := cfg.XAxis.validate("x"); err != nil {
		return err
	}
	if err := cfg.YAxis.validate("y"); err != nil {
		return err
	}
	if err := cfg.TitleOptions.validate(); err != nil {
		return err
	}
	if err := cfg.ValueLabels.validate(); err != nil {
		return err
	}
	if negativePadding(cfg.Padding) {
		return fmt.Errorf("heat map padding cannot be negative")
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

func (axis Axis) validate(name string) error {
	if !finite(axis.TitleFontSize) || axis.TitleFontSize < 0 {
		return fmt.Errorf("heat map %s axis title font size must be finite and non-negative", name)
	}
	if !finite(axis.LabelFontSize) || axis.LabelFontSize < 0 {
		return fmt.Errorf("heat map %s axis label font size must be finite and non-negative", name)
	}
	if !finite(axis.LabelRotation) || axis.LabelRotation < -360 || axis.LabelRotation > 360 {
		return fmt.Errorf("heat map %s axis label rotation must be finite and between -360 and 360 degrees", name)
	}
	if axis.LabelCount < 0 {
		return fmt.Errorf("heat map %s axis label count cannot be negative", name)
	}
	return nil
}

func (options TitleOptions) validate() error {
	switch options.Placement {
	case PlacementDefault, PlacementStart, PlacementCenter, PlacementEnd:
	default:
		return fmt.Errorf("heat map title placement %q is unsupported", options.Placement)
	}
	for name, value := range map[string]float64{
		"font size": options.FontSize, "subtext font size": options.SubtextFontSize, "border width": options.BorderWidth,
	} {
		if !finite(value) || value < 0 {
			return fmt.Errorf("heat map title %s must be finite and non-negative", name)
		}
	}
	return nil
}

func (options ValueLabelOptions) validate() error {
	switch options.Format {
	case ValueFormatDefault, ValueFormatExact, ValueFormatInteger, ValueFormatHumanized:
	default:
		return fmt.Errorf("heat map value label format %q is unsupported", options.Format)
	}
	if options.Decimals < 0 || options.Decimals > 15 {
		return fmt.Errorf("heat map value label decimals must be between 0 and 15")
	}
	if !finite(options.FontSize) || options.FontSize < 0 {
		return fmt.Errorf("heat map value label font size must be finite and non-negative")
	}
	if options.Distance < 0 {
		return fmt.Errorf("heat map value label distance cannot be negative")
	}
	if !options.Show && options != (ValueLabelOptions{}) {
		return fmt.Errorf("heat map value label presentation requires labels to be shown")
	}
	return nil
}

func negativePadding(padding Padding) bool {
	return padding.Top < 0 || padding.Right < 0 || padding.Bottom < 0 || padding.Left < 0
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
