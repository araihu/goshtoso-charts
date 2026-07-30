// Package bar provides the canonical interactive categorical-bar API.
//
// Bar-specific names are concise in this package. Shared interactive options
// remain in the parent interactive package during the additive package-layout
// migration. Existing interactive.Bar configurations keep the same type
// identity and behavior because these declarations are compatibility aliases.
package bar

import interactive "github.com/araihu/goshtoso-charts/components/interactive"

// Config describes an accessible, browser-rendered bar chart.
type Config = interactive.BarConfig

// Series describes one named bar series.
type Series = interactive.BarSeries

// Data describes one finite bar value and optional per-point presentation.
type Data = interactive.BarData

// Orientation selects whether categories run left to right or top to bottom.
type Orientation = interactive.BarOrientation

const (
	// OrientationVertical places categories on the horizontal axis.
	OrientationVertical Orientation = interactive.BarOrientationVertical
	// OrientationHorizontal places categories on the vertical axis.
	OrientationHorizontal Orientation = interactive.BarOrientationHorizontal
)

// ZoomMode selects direct gesture exploration or a visible range slider.
type ZoomMode = interactive.BarZoomMode

const (
	// ZoomInside supports wheel, pinch, and drag exploration inside the plot.
	ZoomInside ZoomMode = interactive.BarZoomInside
	// ZoomSlider exposes a visible range control.
	ZoomSlider ZoomMode = interactive.BarZoomSlider
)

// Zoom selects an initial percentage window over ordered categories.
type Zoom = interactive.BarZoom

// Statistic selects a calculated series reference.
type Statistic = interactive.BarStatistic

const (
	// StatisticMinimum selects the smallest series value.
	StatisticMinimum Statistic = interactive.BarStatisticMinimum
	// StatisticMaximum selects the largest series value.
	StatisticMaximum Statistic = interactive.BarStatisticMaximum
	// StatisticAverage selects the arithmetic mean of series values.
	StatisticAverage Statistic = interactive.BarStatisticAverage
)

// Coordinate identifies one category and finite value.
type Coordinate = interactive.BarCoordinate

// PointReference places either a calculated or explicit point on a series.
type PointReference = interactive.BarPointReference

// GuideReference draws a calculated guide across a series.
type GuideReference = interactive.BarGuideReference

// References configures calculated points, explicit points, and guides.
type References = interactive.BarReferences

// Bar builds a reusable interactive bar component.
func Bar(cfg Config) interactive.Instance { return interactive.Bar(cfg) }
