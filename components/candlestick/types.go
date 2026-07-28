// Package candlestick renders accessible server-side SVG candlestick charts.
package candlestick

import (
	"fmt"
	"math"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

// Datum is one open-high-low-close observation at a named category or time.
type Datum struct {
	Label string
	Open  float64
	High  float64
	Low   float64
	Close float64
}

// Axis controls renderer-neutral axis presentation.
type Axis struct{ Title string }

// Config describes an SSR SVG candlestick chart.
type Config struct {
	Label      string
	Caption    string
	Title      string
	SeriesName string
	Data       []Datum
	XAxis      Axis
	YAxis      Axis
	Width      int
	Height     int
	Style      charttheme.Style
	RootAttrs  templ.Attributes
	// Controls configures shared controls; Expand defaults on while fullscreen and collapse default off.
	Controls chartcontrol.Options
	// Export customizes or disables default SVG and PNG export.
	Export *chartcontrol.ExportOptions
}

func (cfg Config) validate() error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("candlestick chart label is required")
	}
	if strings.TrimSpace(cfg.SeriesName) == "" {
		return fmt.Errorf("candlestick chart series name is required")
	}
	if len(cfg.Data) == 0 {
		return fmt.Errorf("candlestick chart needs at least one datum")
	}
	if cfg.Width < 0 {
		return fmt.Errorf("candlestick chart width cannot be negative")
	}
	if cfg.Height < 0 {
		return fmt.Errorf("candlestick chart height cannot be negative")
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("candlestick chart root attribute %q is reserved", attribute)
			}
		}
	}
	labels := make(map[string]struct{}, len(cfg.Data))
	for index, datum := range cfg.Data {
		if strings.TrimSpace(datum.Label) == "" {
			return fmt.Errorf("candlestick chart datum %d needs a label", index+1)
		}
		if _, exists := labels[datum.Label]; exists {
			return fmt.Errorf("candlestick chart datum label %q is duplicated", datum.Label)
		}
		labels[datum.Label] = struct{}{}
		for name, value := range map[string]float64{"open": datum.Open, "high": datum.High, "low": datum.Low, "close": datum.Close} {
			if !finite(value) {
				return fmt.Errorf("candlestick chart datum %q %s must be finite", datum.Label, name)
			}
		}
		if datum.Low > datum.Open || datum.Low > datum.Close {
			return fmt.Errorf("candlestick chart datum %q low must be less than or equal to open and close", datum.Label)
		}
		if datum.High < datum.Open || datum.High < datum.Close {
			return fmt.Errorf("candlestick chart datum %q high must be greater than or equal to open and close", datum.Label)
		}
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
