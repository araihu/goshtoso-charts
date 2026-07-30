package interactive

import interactiveparallel "github.com/araihu/goshtoso-charts/components/interactive/parallel"

// ParallelScale is the compatibility name for parallel.Scale.
type ParallelScale = interactiveparallel.Scale

const (
	// ParallelScaleLinear uses a continuous linear scale. It is the default.
	ParallelScaleLinear ParallelScale = interactiveparallel.ScaleLinear
	// ParallelScaleLog uses a positive logarithmic scale.
	ParallelScaleLog ParallelScale = interactiveparallel.ScaleLog
)

// ParallelNameLocation is the compatibility name for parallel.NameLocation.
type ParallelNameLocation = interactiveparallel.NameLocation

const (
	ParallelNameEnd    ParallelNameLocation = interactiveparallel.NameEnd
	ParallelNameStart  ParallelNameLocation = interactiveparallel.NameStart
	ParallelNameMiddle ParallelNameLocation = interactiveparallel.NameMiddle
)

// ParallelRange is the compatibility name for parallel.Range.
type ParallelRange = interactiveparallel.Range

// ParallelAxisLabel is the compatibility name for parallel.AxisLabel.
type ParallelAxisLabel = interactiveparallel.AxisLabel

// ParallelAxisLine is the compatibility name for parallel.AxisLine.
type ParallelAxisLine = interactiveparallel.AxisLine

// ParallelDimension is the compatibility name for parallel.Dimension.
type ParallelDimension = interactiveparallel.Dimension

// ParallelLayout is the compatibility name for parallel.Layout.
type ParallelLayout = interactiveparallel.Layout

// ParallelLineOptions is the compatibility name for parallel.LineOptions.
type ParallelLineOptions = interactiveparallel.LineOptions

// ParallelSeriesOptions is the compatibility name for parallel.SeriesOptions.
type ParallelSeriesOptions = interactiveparallel.SeriesOptions

// ParallelValue is the compatibility name for parallel.Value.
type ParallelValue = interactiveparallel.Value

// ParallelNumber forwards to the canonical parallel package.
func ParallelNumber(value float64) ParallelValue { return interactiveparallel.Number(value) }

// ParallelCategory forwards to the canonical parallel package.
func ParallelCategory(value string) ParallelValue { return interactiveparallel.Category(value) }

// ParallelObservation is the compatibility name for parallel.Observation.
type ParallelObservation = interactiveparallel.Observation

// ParallelSeries is the compatibility name for parallel.Series.
type ParallelSeries = interactiveparallel.Series

// ParallelConfig is the compatibility name for parallel.Config.
type ParallelConfig = interactiveparallel.Config

// Parallel forwards to the canonical parallel package.
func Parallel(cfg ParallelConfig) Instance { return interactiveparallel.Parallel(cfg) }
