// Package components defines stable identities shared by Goshtoso Charts components.
package components

import (
	"slices"

	"github.com/a-h/templ"
)

// Kind identifies a Goshtoso Charts component type.
type Kind string

// Component is a renderable chart component with stable identity.
type Component interface {
	templ.Component
	Kind() Kind
}

const (
	// KindHeartbeat identifies an availability/history heartbeat chart.
	KindHeartbeat Kind = "heartbeat"
	// KindLineChart identifies a quantitative time-series line chart.
	KindLineChart Kind = "line-chart"
)

var allKinds = []Kind{KindHeartbeat, KindLineChart}

// AllKinds returns supported component kinds in stable order.
func AllKinds() []Kind {
	return slices.Clone(allKinds)
}
