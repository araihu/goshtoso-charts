package interactive

import interactivescatter "github.com/araihu/goshtoso-charts/components/interactive/scatter"

// ScatterVariant selects the scatter renderer without changing the component contract.
type ScatterVariant = interactivescatter.Variant

const (
	// ScatterVariantStandard renders ordinary scatter points. It is the default.
	ScatterVariantStandard = interactivescatter.VariantStandard
	// ScatterVariantEffect adds an animated ripple around points.
	ScatterVariantEffect = interactivescatter.VariantEffect
)

// CartesianAxisType selects how scatter-family x values are interpreted.
type CartesianAxisType = interactivescatter.AxisType

const (
	// CartesianAxisCategory uses XAxis as an ordered category list. It is the default.
	CartesianAxisCategory = interactivescatter.AxisCategory
	// CartesianAxisValue reads numeric x/y coordinates from each data value.
	CartesianAxisValue = interactivescatter.AxisValue
)

// ScatterConfig describes an accessible, browser-rendered scatter chart.
type ScatterConfig = interactivescatter.Config

// ScatterSeries describes one named scatter series.
type ScatterSeries = interactivescatter.Series

// ScatterData describes either a category value or a numeric x/y coordinate.
type ScatterData = interactivescatter.Data

// Scatter builds a reusable interactive scatter component.
func Scatter(cfg ScatterConfig) Instance { return interactivescatter.Scatter(cfg) }
