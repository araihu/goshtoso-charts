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
	// KindBarChart identifies a quantitative categorical bar chart.
	KindBarChart Kind = "bar-chart"
	// KindPieChart identifies a categorical proportional breakdown chart.
	KindPieChart Kind = "pie-chart"
	// KindInteractiveECharts identifies a browser-rendered go-echarts chart.
	KindInteractiveECharts Kind = "interactive-echarts"
)

var allKinds = []Kind{KindHeartbeat, KindLineChart, KindBarChart, KindPieChart, KindInteractiveECharts}

// AllKinds returns supported component kinds in stable order.
func AllKinds() []Kind {
	return slices.Clone(allKinds)
}
