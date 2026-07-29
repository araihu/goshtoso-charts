// Package pie renders accessible server-side SVG pie charts.
package pie

import (
	"fmt"
	"math"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

// Slice is one named proportional value in a pie chart.
type Slice struct {
	Name  string
	Value float64
	// Color overrides the corresponding chart palette color.
	Color string
	// Class styles the rendered slice and matching exact-value row.
	Class string
}

// Variant selects a proportional chart treatment.
type Variant string

const (
	// VariantPie renders the zero-value full pie treatment.
	VariantPie Variant = ""
	// VariantDoughnut opens the center while preserving pie semantics.
	VariantDoughnut Variant = "doughnut"
)

// HorizontalPlacement controls renderer-neutral horizontal placement.
type HorizontalPlacement string

const (
	PlacementDefault HorizontalPlacement = ""
	PlacementCenter  HorizontalPlacement = "center"
)

// VerticalPlacement controls renderer-neutral vertical placement.
type VerticalPlacement string

const (
	VerticalPlacementDefault VerticalPlacement = ""
	VerticalPlacementTop     VerticalPlacement = "top"
	VerticalPlacementMiddle  VerticalPlacement = "middle"
	VerticalPlacementBottom  VerticalPlacement = "bottom"
)

// LegendOrientation controls legend flow.
type LegendOrientation string

const (
	LegendHorizontal LegendOrientation = ""
	LegendVertical   LegendOrientation = "vertical"
)

// TitleOptions controls visible title and subtitle presentation.
type TitleOptions struct {
	Text             string
	Subtitle         string
	Placement        HorizontalPlacement
	FontSize         float64
	SubtitleFontSize float64
}

// LegendOptions controls renderer-neutral legend presentation.
type LegendOptions struct {
	Orientation       LegendOrientation
	LeftPercent       float64
	VerticalPlacement VerticalPlacement
	FontSize          float64
}

// Padding is chart inset in pixels.
type Padding struct{ Top, Right, Bottom, Left int }

// Config describes an SSR SVG pie chart.
//
// Label is required and becomes the accessible name of the rendered figure.
// Caption remains visible below the chart. VariantPie is the zero-value
// treatment. InnerRadiusPercent is the doughnut hole as a percentage of the
// outer ring; zero keeps the renderer-neutral 60 percent default.
type Config struct {
	Label              string
	Caption            string
	Slices             []Slice
	Variant            Variant
	InnerRadiusPercent float64
	Title              TitleOptions
	Legend             LegendOptions
	Padding            Padding
	Width              int
	Height             int
	Style              charttheme.Style
	RootAttrs          templ.Attributes
	// Controls configures shared controls; Expand defaults on while fullscreen defaults off.
	Controls chartcontrol.Options
	// Export customizes or disables default SVG and PNG export.
	Export *chartcontrol.ExportOptions
}

func (cfg Config) validate() error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("pie chart label is required")
	}
	switch cfg.Variant {
	case VariantPie, VariantDoughnut:
	default:
		return fmt.Errorf("pie chart variant %q is unsupported", cfg.Variant)
	}
	if cfg.Width < 0 {
		return fmt.Errorf("pie chart width cannot be negative")
	}
	if cfg.Height < 0 {
		return fmt.Errorf("pie chart height cannot be negative")
	}
	if !finite(cfg.InnerRadiusPercent) || cfg.InnerRadiusPercent < 0 || cfg.InnerRadiusPercent >= 100 {
		return fmt.Errorf("pie chart inner radius percent must be zero or a finite value below 100")
	}
	if cfg.Variant != VariantDoughnut && cfg.InnerRadiusPercent != 0 {
		return fmt.Errorf("pie chart inner radius percent requires doughnut variant")
	}
	if err := cfg.Title.validate(); err != nil {
		return err
	}
	if err := cfg.Legend.validate(); err != nil {
		return err
	}
	if err := cfg.Padding.validate(); err != nil {
		return err
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("pie chart root attribute %q is reserved", attribute)
			}
		}
	}
	names := make(map[string]struct{}, len(cfg.Slices))
	for index, slice := range cfg.Slices {
		if strings.TrimSpace(slice.Name) == "" {
			return fmt.Errorf("pie chart slice %d needs a name", index+1)
		}
		if _, exists := names[slice.Name]; exists {
			return fmt.Errorf("pie chart slice %q is duplicated", slice.Name)
		}
		names[slice.Name] = struct{}{}
		if !finite(slice.Value) {
			return fmt.Errorf("pie chart slice %q needs a finite value", slice.Name)
		}
		if slice.Value < 0 {
			return fmt.Errorf("pie chart slice %q has a negative value", slice.Name)
		}
	}
	return nil
}

func (options TitleOptions) validate() error {
	switch options.Placement {
	case PlacementDefault, PlacementCenter:
	default:
		return fmt.Errorf("pie chart title placement %q is unsupported", options.Placement)
	}
	if !finite(options.FontSize) || options.FontSize < 0 {
		return fmt.Errorf("pie chart title font size must be finite and non-negative")
	}
	if !finite(options.SubtitleFontSize) || options.SubtitleFontSize < 0 {
		return fmt.Errorf("pie chart subtitle font size must be finite and non-negative")
	}
	return nil
}

func (options LegendOptions) validate() error {
	switch options.Orientation {
	case LegendHorizontal, LegendVertical:
	default:
		return fmt.Errorf("pie chart legend orientation %q is unsupported", options.Orientation)
	}
	switch options.VerticalPlacement {
	case VerticalPlacementDefault, VerticalPlacementTop, VerticalPlacementMiddle, VerticalPlacementBottom:
	default:
		return fmt.Errorf("pie chart legend vertical placement %q is unsupported", options.VerticalPlacement)
	}
	if !finite(options.LeftPercent) || options.LeftPercent < 0 || options.LeftPercent > 100 {
		return fmt.Errorf("pie chart legend left percent must be finite and between 0 and 100")
	}
	if !finite(options.FontSize) || options.FontSize < 0 {
		return fmt.Errorf("pie chart legend font size must be finite and non-negative")
	}
	return nil
}

func (padding Padding) validate() error {
	if padding.Top < 0 || padding.Right < 0 || padding.Bottom < 0 || padding.Left < 0 {
		return fmt.Errorf("pie chart padding cannot be negative")
	}
	return nil
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

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
