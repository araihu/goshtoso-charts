package interactive

import interactivegauge "github.com/araihu/goshtoso-charts/components/interactive/gauge"

// GaugeVariant selects a visual treatment without changing component identity.
type GaugeVariant = interactivegauge.Variant

const (
	// GaugeVariantStandard renders the default dial and pointer. It is the zero-value default.
	GaugeVariantStandard = interactivegauge.VariantStandard
	// GaugeVariantProgress renders a rounded progress arc without a pointer.
	GaugeVariantProgress = interactivegauge.VariantProgress
	// GaugeVariantLiquid renders bounded readings as layered waves.
	GaugeVariantLiquid = interactivegauge.VariantLiquid
)

// GaugeLiquidShape selects a built-in silhouette for the liquid treatment.
type GaugeLiquidShape = interactivegauge.LiquidShape

const (
	GaugeLiquidShapeCircle    = interactivegauge.LiquidShapeCircle
	GaugeLiquidShapeRect      = interactivegauge.LiquidShapeRect
	GaugeLiquidShapeRoundRect = interactivegauge.LiquidShapeRoundRect
	GaugeLiquidShapeTriangle  = interactivegauge.LiquidShapeTriangle
	GaugeLiquidShapeDiamond   = interactivegauge.LiquidShapeDiamond
	GaugeLiquidShapePin       = interactivegauge.LiquidShapePin
	GaugeLiquidShapeArrow     = interactivegauge.LiquidShapeArrow
)

// GaugeLiquidDirection selects horizontal wave travel. Right is the zero-value default.
type GaugeLiquidDirection = interactivegauge.LiquidDirection

const (
	GaugeLiquidDirectionRight = interactivegauge.LiquidDirectionRight
	GaugeLiquidDirectionLeft  = interactivegauge.LiquidDirectionLeft
)

// GaugeLiquidTreatment configures the liquid treatment without exposing its renderer.
type GaugeLiquidTreatment = interactivegauge.LiquidTreatment

// GaugeLiquidOutline configures the optional boundary around the liquid shape.
type GaugeLiquidOutline = interactivegauge.LiquidOutline

// GaugeLiquidBackground configures the unfilled portion of the liquid shape.
type GaugeLiquidBackground = interactivegauge.LiquidBackground

// GaugeLiquidLabel configures the central bounded-reading label.
type GaugeLiquidLabel = interactivegauge.LiquidLabel

// GaugeLiquidStyle configures wave paint and opacity.
type GaugeLiquidStyle = interactivegauge.LiquidStyle

// GaugeConfig describes an accessible, browser-rendered gauge chart.
//
// Min defaults to 0. Max defaults to 100 when zero. Values must be
// application-owned because the browser renderer serializes them.
type GaugeConfig = interactivegauge.Config

// GaugeScaleMode selects scale treatment. Zero value is thermal sequential.
type GaugeScaleMode = interactivegauge.ScaleMode

const (
	GaugeScaleThermal     = interactivegauge.ScaleThermal
	GaugeScaleCustom      = interactivegauge.ScaleCustom
	GaugeScaleSingleColor = interactivegauge.ScaleSingleColor
)

// GaugeScale configures the full Min-to-Max arc independently from progress.
type GaugeScale = interactivegauge.Scale

// GaugeScaleStop applies a semantic color from the prior stop through Value.
type GaugeScaleStop = interactivegauge.ScaleStop

// GaugeSeries describes one named dial series. Series options override variant defaults.
type GaugeSeries = interactivegauge.Series

// GaugeProgressOptions customizes progress-arc rendering after variant defaults.
type GaugeProgressOptions = interactivegauge.ProgressOptions

// GaugeData describes one named finite reading.
type GaugeData = interactivegauge.Data

// Gauge builds a reusable interactive gauge component.
func Gauge(cfg GaugeConfig) Instance { return interactivegauge.Gauge(cfg) }
