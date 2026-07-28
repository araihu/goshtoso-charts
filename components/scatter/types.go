// Package scatter renders accessible server-side SVG scatter charts.
package scatter

import (
	"fmt"
	"math"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

// Point is one observation assigned to an ordered category.
type Point struct {
	Category string
	Value    float64
}

// Symbol selects the marker drawn for scatter points.
type Symbol string

const (
	// SymbolDefault inherits chart options or draws a dot when used globally.
	SymbolDefault Symbol = ""
	// SymbolDot draws a compact filled dot.
	SymbolDot Symbol = "dot"
	// SymbolCircle draws a hollow circle.
	SymbolCircle Symbol = "circle"
	// SymbolSquare draws a square.
	SymbolSquare Symbol = "square"
	// SymbolDiamond draws a diamond.
	SymbolDiamond Symbol = "diamond"
)

// Options controls scatter-specific marker presentation.
//
// Size is the marker radius in pixels. Zero selects the renderer default.
type Options struct {
	Symbol Symbol
	Size   float64
}

// Series is one named population of categorical points.
//
// Options override Config.Options for this series when their fields are set.
type Series struct {
	Name    string
	Points  []Point
	Options Options
}

// Config describes an SSR SVG scatter chart.
//
// Label is required and becomes the accessible name of the rendered figure.
// Caption remains visible below the chart. Style controls the theme palette and
// caller root class. RootAttrs accepts additional figure attributes except
// class, role, and aria-label, which remain owned by the component.
type Config struct {
	Label      string
	Caption    string
	Categories []string
	Series     []Series
	Options    Options
	Width      int
	Height     int
	Style      charttheme.Style
	RootAttrs  templ.Attributes
}

func (cfg Config) validate() error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("scatter chart label is required")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("scatter chart needs at least one series")
	}
	if len(cfg.Categories) == 0 {
		return fmt.Errorf("scatter chart needs at least one category")
	}
	if cfg.Width < 0 {
		return fmt.Errorf("scatter chart width cannot be negative")
	}
	if cfg.Height < 0 {
		return fmt.Errorf("scatter chart height cannot be negative")
	}
	if err := cfg.Options.validate("scatter chart options"); err != nil {
		return err
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("scatter chart root attribute %q is reserved", attribute)
			}
		}
	}
	categoryIndexes := make(map[string]int, len(cfg.Categories))
	for index, category := range cfg.Categories {
		if strings.TrimSpace(category) == "" {
			return fmt.Errorf("scatter chart category %d cannot be empty", index+1)
		}
		if _, exists := categoryIndexes[category]; exists {
			return fmt.Errorf("scatter chart category %q is duplicated", category)
		}
		categoryIndexes[category] = index
	}
	for index, series := range cfg.Series {
		if strings.TrimSpace(series.Name) == "" {
			return fmt.Errorf("scatter chart series %d needs a name", index+1)
		}
		if len(series.Points) == 0 {
			return fmt.Errorf("scatter chart series %q needs at least one point", series.Name)
		}
		if err := series.Options.validate(fmt.Sprintf("scatter chart series %q options", series.Name)); err != nil {
			return err
		}
		for pointIndex, point := range series.Points {
			if _, ok := categoryIndexes[point.Category]; !ok {
				return fmt.Errorf("scatter chart series %q point %d references unknown category %q", series.Name, pointIndex, point.Category)
			}
			if math.IsNaN(point.Value) || math.IsInf(point.Value, 0) {
				return fmt.Errorf("scatter chart series %q point %d must contain a finite value", series.Name, pointIndex)
			}
		}
	}
	return nil
}

func (options Options) validate(prefix string) error {
	switch options.Symbol {
	case SymbolDefault, SymbolDot, SymbolCircle, SymbolSquare, SymbolDiamond:
	default:
		return fmt.Errorf("%s has unsupported symbol %q", prefix, options.Symbol)
	}
	if math.IsNaN(options.Size) || math.IsInf(options.Size, 0) || options.Size < 0 {
		return fmt.Errorf("%s size must be a finite non-negative number", prefix)
	}
	return nil
}

func (options Options) resolved(fallback Options) Options {
	if options.Symbol == SymbolDefault {
		options.Symbol = fallback.Symbol
	}
	if options.Size == 0 {
		options.Size = fallback.Size
	}
	return options
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
