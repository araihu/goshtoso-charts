// Package bar renders accessible server-side SVG bar charts.
package bar

import "fmt"

// Series is one named sequence of values aligned with Config.Labels.
type Series struct {
	Name   string
	Values []float64
}

// Config describes an SSR SVG categorical bar chart.
type Config struct {
	Label   string
	Caption string
	Labels  []string
	Series  []Series
	Stacked bool
	Width   int
	Height  int
}

func (cfg Config) validate() error {
	if cfg.Label == "" {
		return fmt.Errorf("bar chart label is required")
	}
	if len(cfg.Labels) == 0 {
		return fmt.Errorf("bar chart needs at least one label")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("bar chart needs at least one series")
	}
	for index, series := range cfg.Series {
		if series.Name == "" {
			return fmt.Errorf("bar chart series %d needs a name", index+1)
		}
		if len(series.Values) != len(cfg.Labels) {
			return fmt.Errorf("bar chart series %q has %d values; need %d", series.Name, len(series.Values), len(cfg.Labels))
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
