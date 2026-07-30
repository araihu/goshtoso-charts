package interactive

import interactiveheatmap "github.com/araihu/goshtoso-charts/components/interactive/heatmap"

// HeatMapCoordinate selects the heatmap coordinate system.
type HeatMapCoordinate = interactiveheatmap.Coordinate

const (
	// HeatMapCoordinateCartesian maps X and Y indexes to category axes. It is the default.
	HeatMapCoordinateCartesian = interactiveheatmap.CoordinateCartesian
	// HeatMapCoordinateCalendar maps Date values to a calendar range.
	HeatMapCoordinateCalendar = interactiveheatmap.CoordinateCalendar
)

// HeatMapConfig describes an accessible, browser-rendered heatmap.
//
// Values must be application-owned because the browser renderer serializes them.
type HeatMapConfig = interactiveheatmap.Config

// HeatMapCalendar defines an inclusive calendar date range. Options customizes
// calendar presentation; its Range is replaced by Start and End.
type HeatMapCalendar = interactiveheatmap.Calendar

// HeatMapValueRange defines the visual-map domain. Values outside the range are
// preserved and rendered with the nearest endpoint color.
type HeatMapValueRange = interactiveheatmap.ValueRange

// HeatMapSeries describes one named heatmap series.
type HeatMapSeries = interactiveheatmap.Series

// HeatMapData describes one heatmap cell. Cartesian mode uses X and Y category
// indexes. Calendar mode uses Date.
type HeatMapData = interactiveheatmap.Data

// HeatMap builds a reusable interactive Cartesian or calendar heatmap component.
func HeatMap(cfg HeatMapConfig) Instance { return interactiveheatmap.HeatMap(cfg) }
