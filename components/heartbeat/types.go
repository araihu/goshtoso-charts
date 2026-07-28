// Package heartbeat renders server-side availability history charts.
package heartbeat

import (
	"fmt"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

const maxPoints = 300

// State describes a single observation's availability state.
type State string

const (
	// StateUp marks a successful observation.
	StateUp State = "up"
	// StateDegraded marks a successful but degraded observation.
	StateDegraded State = "degraded"
	// StateDown marks a failed observation.
	StateDown State = "down"
	// StateUnknown marks an observation with insufficient evidence.
	StateUnknown State = "unknown"
)

// Point is one time-ordered monitor observation. Latency is optional; it is
// displayed in the SVG title when positive.
type Point struct {
	At      time.Time
	State   State
	Latency time.Duration
}

// Config configures a compact availability history chart.
//
// Points are intentionally application-neutral. Map monitoring/domain values
// at the boundary instead of importing a product package into this library.
type Config struct {
	// Label is required and becomes the figure's accessible name.
	Label string
	// Caption is visible supporting context, such as "Last 30 checks".
	Caption string
	// Points must be in chronological order. Empty is a supported no-data state.
	Points []Point
	// RootClass appends CSS classes to the figure.
	RootClass string
	// RootAttrs appends HTML attributes to the figure.
	RootAttrs templ.Attributes
	// Style selects state colors and appends chart-specific root classes.
	Style charttheme.Style
}

func (cfg Config) validate() error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("heartbeat label is required")
	}
	if len(cfg.Points) > maxPoints {
		return fmt.Errorf("heartbeat has %d points; maximum is %d", len(cfg.Points), maxPoints)
	}
	for index, point := range cfg.Points {
		if point.At.IsZero() {
			return fmt.Errorf("heartbeat point %d has no timestamp", index+1)
		}
		if point.Latency < 0 {
			return fmt.Errorf("heartbeat point %d has negative latency", index+1)
		}
		if !point.State.valid() {
			return fmt.Errorf("heartbeat point %d has unsupported state %q", index+1, point.State)
		}
		if index > 0 && !cfg.Points[index-1].At.Before(point.At) {
			return fmt.Errorf("heartbeat points must be in chronological order")
		}
	}
	return nil
}

func (state State) valid() bool {
	switch state {
	case StateUp, StateDegraded, StateDown, StateUnknown:
		return true
	default:
		return false
	}
}

func (cfg Config) rootClasses() string {
	base := "block w-full"
	if class := strings.TrimSpace(cfg.RootClass); class != "" {
		base += " " + class
	}
	return cfg.Style.RootClasses(base)
}
