package interactive

import interactiveline "github.com/araihu/goshtoso-charts/components/interactive/line"

// LineConfig is the compatibility name for line.Config.
type LineConfig = interactiveline.Config

// LineTimeAxis is the compatibility name for line.TimeAxis.
type LineTimeAxis = interactiveline.TimeAxis

// LineValueAxis is the compatibility name for line.ValueAxis.
type LineValueAxis = interactiveline.ValueAxis

// LineVisualDimension is the compatibility name for line.VisualDimension.
type LineVisualDimension = interactiveline.VisualDimension

const (
	// LineVisualDimensionX maps x coordinates.
	LineVisualDimensionX LineVisualDimension = interactiveline.VisualDimensionX
	// LineVisualDimensionY maps y values.
	LineVisualDimensionY LineVisualDimension = interactiveline.VisualDimensionY
)

// LineVisualScale is the compatibility name for line.VisualScale.
type LineVisualScale = interactiveline.VisualScale

// LineVisualPiece is the compatibility name for line.VisualPiece.
type LineVisualPiece = interactiveline.VisualPiece

// LineStatistic is the compatibility name for line.Statistic.
type LineStatistic = interactiveline.Statistic

const (
	// LineStatisticMinimum selects the smallest series value.
	LineStatisticMinimum LineStatistic = interactiveline.StatisticMinimum
	// LineStatisticMaximum selects the largest series value.
	LineStatisticMaximum LineStatistic = interactiveline.StatisticMaximum
	// LineStatisticAverage selects the arithmetic mean of series values.
	LineStatisticAverage LineStatistic = interactiveline.StatisticAverage
)

// LineCoordinate is the compatibility name for line.Coordinate.
type LineCoordinate = interactiveline.Coordinate

// LinePointReference is the compatibility name for line.PointReference.
type LinePointReference = interactiveline.PointReference

// LineGuideReference is the compatibility name for line.GuideReference.
type LineGuideReference = interactiveline.GuideReference

// LineRangeReference is the compatibility name for line.RangeReference.
type LineRangeReference = interactiveline.RangeReference

// LineReferences is the compatibility name for line.References.
type LineReferences = interactiveline.References

// LineSeries is the compatibility name for line.Series.
type LineSeries = interactiveline.Series

// LineData is the compatibility name for line.Data.
type LineData = interactiveline.Data

// Line forwards to the canonical line package.
func Line(cfg LineConfig) Instance { return interactiveline.Line(cfg) }
