// Package pie renders accessible server-side SVG pie charts.
package pie

import (
	"fmt"

	"github.com/araihu/goshtoso-charts/components/charttheme"
)

// Slice is one named proportional value in a pie chart.
type Slice struct {
	Name  string
	Value float64
}

// Config describes an SSR SVG pie chart.
//
// Label is required and becomes the accessible name of the rendered figure.
// Caption remains visible below the chart.
type Config struct {
	Label   string
	Caption string
	Slices  []Slice
	Width   int
	Height  int
	Style   charttheme.Style
}

func (cfg Config) validate() error {
	if cfg.Label == "" {
		return fmt.Errorf("pie chart label is required")
	}
	for index, slice := range cfg.Slices {
		if slice.Name == "" {
			return fmt.Errorf("pie chart slice %d needs a name", index+1)
		}
		if slice.Value < 0 {
			return fmt.Errorf("pie chart slice %q has a negative value", slice.Name)
		}
	}
	return nil
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
