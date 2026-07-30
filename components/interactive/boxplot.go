package interactive

import interactiveboxplot "github.com/araihu/goshtoso-charts/components/interactive/boxplot"

// BoxPlotConfig is the compatibility name for boxplot.Config.
type BoxPlotConfig = interactiveboxplot.Config

// BoxPlotSeries is the compatibility name for boxplot.Series.
type BoxPlotSeries = interactiveboxplot.Series

// BoxPlotData is the compatibility name for boxplot.Data.
type BoxPlotData = interactiveboxplot.Data

// BoxPlot forwards to the canonical boxplot package.
func BoxPlot(cfg BoxPlotConfig) Instance { return interactiveboxplot.BoxPlot(cfg) }
