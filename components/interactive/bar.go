package interactive

import interactivebar "github.com/araihu/goshtoso-charts/components/interactive/bar"

// BarOrientation selects whether categories run left to right or top to bottom.
type BarOrientation = interactivebar.Orientation

const (
	// BarOrientationVertical places categories on the horizontal axis.
	BarOrientationVertical = interactivebar.OrientationVertical
	// BarOrientationHorizontal places categories on the vertical axis.
	BarOrientationHorizontal = interactivebar.OrientationHorizontal
)

// BarZoomMode selects direct gesture exploration or a visible range slider.
type BarZoomMode = interactivebar.ZoomMode

const (
	// BarZoomInside supports wheel, pinch, and drag exploration inside the plot.
	BarZoomInside = interactivebar.ZoomInside
	// BarZoomSlider exposes a visible range control.
	BarZoomSlider = interactivebar.ZoomSlider
)

// BarZoom selects an initial percentage window over ordered categories.
type BarZoom = interactivebar.Zoom

// BarStatistic selects a calculated series reference.
type BarStatistic = interactivebar.Statistic

const (
	// BarStatisticMinimum selects the smallest series value.
	BarStatisticMinimum = interactivebar.StatisticMinimum
	// BarStatisticMaximum selects the largest series value.
	BarStatisticMaximum = interactivebar.StatisticMaximum
	// BarStatisticAverage selects the arithmetic mean of series values.
	BarStatisticAverage = interactivebar.StatisticAverage
)

// BarCoordinate identifies one category and finite value.
type BarCoordinate = interactivebar.Coordinate

// BarPointReference places either a calculated or explicit point on a series.
type BarPointReference = interactivebar.PointReference

// BarGuideReference draws a calculated guide across a series.
type BarGuideReference = interactivebar.GuideReference

// BarReferences configures calculated points, explicit points, and guides.
type BarReferences = interactivebar.References

// BarConfig describes an accessible, browser-rendered bar chart.
type BarConfig = interactivebar.Config

// BarSeries describes one named bar series.
type BarSeries = interactivebar.Series

// BarData describes one finite bar value and optional per-point presentation.
type BarData = interactivebar.Data

// Bar builds a reusable interactive bar component.
func Bar(cfg BarConfig) Instance { return interactivebar.Bar(cfg) }
