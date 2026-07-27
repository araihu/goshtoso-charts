// Package line renders accessible server-side SVG line charts.
package line

import "fmt"

// Theme selects a Goshtoso-aligned static palette. Render the chart again when
// the server-selected color mode changes.
type Theme string

const (
	// ThemeGoshtoso uses the default light Goshtoso surface and semantic colors.
	ThemeGoshtoso Theme = "goshtoso"
	// ThemeGoshtosoDark uses colors that retain contrast on Goshtoso dark surfaces.
	ThemeGoshtosoDark Theme = "goshtoso-dark"
)

// Series is one labeled sequence of values. Values must align with Config.Labels.
type Series struct {
	Name   string
	Values []float64
}

// Config describes an SSR SVG line chart.
//
// Label is required and becomes the accessible name of the rendered figure.
// Caption remains visible below the chart. Keep the data table near the chart
// when users need exact values; charts are summary views, not the only data UI.
type Config struct {
	Label   string
	Caption string
	Labels  []string
	Series  []Series
	Width   int
	Height  int
	Theme   Theme
}

func (cfg Config) validate() error {
	if cfg.Label == "" {
		return fmt.Errorf("line chart label is required")
	}
	if len(cfg.Labels) == 0 {
		return fmt.Errorf("line chart needs at least one label")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("line chart needs at least one series")
	}
	for index, series := range cfg.Series {
		if series.Name == "" {
			return fmt.Errorf("line chart series %d needs a name", index+1)
		}
		if len(series.Values) != len(cfg.Labels) {
			return fmt.Errorf("line chart series %q has %d values; need %d", series.Name, len(series.Values), len(cfg.Labels))
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
