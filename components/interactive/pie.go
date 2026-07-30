package interactive

import interactivepie "github.com/araihu/goshtoso-charts/components/interactive/pie"

// PieRoseMode is the compatibility name for pie.RoseMode.
type PieRoseMode = interactivepie.RoseMode

const (
	PieRoseNone   PieRoseMode = interactivepie.RoseNone
	PieRoseRadius PieRoseMode = interactivepie.RoseRadius
	PieRoseArea   PieRoseMode = interactivepie.RoseArea
)

// PieLabelContent is the compatibility name for pie.LabelContent.
type PieLabelContent = interactivepie.LabelContent

const (
	PieLabelDefault      PieLabelContent = interactivepie.LabelDefault
	PieLabelNameAndValue PieLabelContent = interactivepie.LabelNameAndValue
)

// PieTooltipContent is the compatibility name for pie.TooltipContent.
type PieTooltipContent = interactivepie.TooltipContent

const (
	PieTooltipDefault      PieTooltipContent = interactivepie.TooltipDefault
	PieTooltipNameAndShare PieTooltipContent = interactivepie.TooltipNameAndShare
)

// PieAutoEmphasisOptions is the compatibility name for pie.AutoEmphasisOptions.
type PieAutoEmphasisOptions = interactivepie.AutoEmphasisOptions

// PieCenter is the compatibility name for pie.Center.
type PieCenter = interactivepie.Center

// PieConfig is the compatibility name for pie.Config.
type PieConfig = interactivepie.Config

// PieSeries is the compatibility name for pie.Series.
type PieSeries = interactivepie.Series

// PieData is the compatibility name for pie.Data.
type PieData = interactivepie.Data

// Pie forwards to the canonical pie package.
func Pie(cfg PieConfig) Instance { return interactivepie.Pie(cfg) }
