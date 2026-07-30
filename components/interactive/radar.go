package interactive

import interactiveradar "github.com/araihu/goshtoso-charts/components/interactive/radar"

// RadarConfig describes an accessible, browser-rendered radar chart.
//
// Values must be application-owned because the browser renderer serializes them.
type RadarConfig = interactiveradar.Config

// RadarShape controls the coordinate boundary geometry.
type RadarShape = interactiveradar.Shape

const (
	// RadarShapeDefault preserves the standard polygon boundary.
	RadarShapeDefault = interactiveradar.ShapeDefault
	// RadarShapePolygon renders straight-sided concentric boundaries.
	RadarShapePolygon = interactiveradar.ShapePolygon
	// RadarShapeCircle renders circular concentric boundaries.
	RadarShapeCircle = interactiveradar.ShapeCircle
)

// RadarCoordinateOptions configures the shared bounded dimensions without
// exposing renderer-specific coordinate types.
type RadarCoordinateOptions = interactiveradar.CoordinateOptions

// RadarSplitLineOptions configures concentric coordinate guides.
type RadarSplitLineOptions = interactiveradar.SplitLineOptions

// RadarIndicator describes one named radar dimension and its positive maximum.
type RadarIndicator = interactiveradar.Indicator

// RadarSeries describes one named radar series.
type RadarSeries = interactiveradar.Series

// RadarData describes one named vector whose values align with Indicators.
type RadarData = interactiveradar.Data

// Radar builds a reusable interactive radar component.
func Radar(cfg RadarConfig) Instance {
	return interactiveradar.Radar(cfg)
}
