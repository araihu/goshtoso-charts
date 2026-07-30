package line

import interactive "github.com/araihu/goshtoso-charts/components/interactive"

// Config describes an accessible, browser-rendered line chart.
type Config = interactive.LineConfig

// TimeAxis defines ordered instants and a required inclusive lower bound.
type TimeAxis = interactive.LineTimeAxis

// ValueAxis defines ordered finite numerical x coordinates.
type ValueAxis = interactive.LineValueAxis

// VisualDimension selects the coordinate used by a piecewise visual scale.
type VisualDimension = interactive.LineVisualDimension

const (
	// VisualDimensionX maps x coordinates.
	VisualDimensionX = interactive.LineVisualDimensionX
	// VisualDimensionY maps y values.
	VisualDimensionY = interactive.LineVisualDimensionY
)

// VisualScale defines theme-aware numeric pieces.
type VisualScale = interactive.LineVisualScale

// VisualPiece defines one open numeric interval.
type VisualPiece = interactive.LineVisualPiece

// Statistic selects a calculated series reference.
type Statistic = interactive.LineStatistic

const (
	// StatisticMinimum selects the smallest series value.
	StatisticMinimum = interactive.LineStatisticMinimum
	// StatisticMaximum selects the largest series value.
	StatisticMaximum = interactive.LineStatisticMaximum
	// StatisticAverage selects the arithmetic mean of series values.
	StatisticAverage = interactive.LineStatisticAverage
)

// Coordinate identifies one finite x/y point.
type Coordinate = interactive.LineCoordinate

// PointReference places a calculated point on a series.
type PointReference = interactive.LinePointReference

// GuideReference draws a calculated, vertical, or two-coordinate guide.
type GuideReference = interactive.LineGuideReference

// RangeReference highlights an inclusive x-axis interval.
type RangeReference = interactive.LineRangeReference

// References configures theme-aware points, guides, and ranges.
type References = interactive.LineReferences

// Series describes one named line series.
type Series = interactive.LineSeries

// Data describes one finite line value and optional point symbol.
type Data = interactive.LineData

// Line builds a reusable interactive line component.
func Line(cfg Config) interactive.Instance { return interactive.Line(cfg) }
