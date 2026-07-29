package pages

import (
	interactive "github.com/araihu/goshtoso-charts/components/interactive"
)

const (
	interactiveLineUpstreamPath     = "examples/line.go"
	interactiveLineUpstreamRevision = "bda428480a82d6d77ebb9fa939cf8d52528453dd"
	interactiveLineUpstreamSHA256   = "1f36444bd373eafde876af19746d6b0115a776fd7c019e5996bdf2d00ecd7b1c"

	lineCoverageExample     = "example"
	lineCoverageUnsupported = "unsupported"
)

type interactiveLineCoverageEntry struct {
	Name   string
	Status string
	Reason string
}

func interactiveLineUpstreamCoverage() []interactiveLineCoverageEntry {
	return []interactiveLineCoverageEntry{
		{Name: "lineBase", Status: lineCoverageExample},
		{Name: "lineShowLabel", Status: lineCoverageExample},
		{Name: "lineMarkPoint", Status: lineCoverageExample},
		{Name: "lineSplitLine", Status: lineCoverageExample},
		{Name: "lineNumerical", Status: lineCoverageExample},
		{Name: "lineTime", Status: lineCoverageExample},
		{Name: "lineStep", Status: lineCoverageExample},
		{Name: "lineSmooth", Status: lineCoverageExample},
		{Name: "lineArea", Status: lineCoverageExample},
		{Name: "lineSmoothArea", Status: lineCoverageExample},
		{Name: "lineOverlap", Status: lineCoverageUnsupported, Reason: "mixed-series composition requires a renderer-neutral composite chart API"},
		{Name: "lineMulti", Status: lineCoverageExample},
		{Name: "lineDemo", Status: lineCoverageExample},
		{Name: "lineSymbols", Status: lineCoverageExample},
	}
}

func interactiveLineCategories() []string {
	// Upstream includes a trailing space in Peach. Correct the obvious label typo.
	return []string{"Apple", "Banana", "Peach", "Lemon", "Pear", "Cherry"}
}

func fixedInteractiveLineData(values ...float64) []interactive.LineData {
	data := make([]interactive.LineData, len(values))
	for index, value := range values {
		data[index] = interactive.LineData{Value: value}
	}
	return data
}

func controlledInteractiveLineOptions(title, filename string) interactive.ChartOptions {
	options := controlledOptions(title, filename)
	options.Legend = &interactive.LegendOptions{Bottom: "0"}
	return options
}

func sampleInteractiveLineBase() interactive.LineConfig {
	return interactive.LineConfig{
		Label: "Basic line example", Caption: "Six deterministic category values preserve the upstream shape.",
		XAxis:   interactiveLineCategories(),
		Series:  []interactive.LineSeries{{Name: "Category A", Data: fixedInteractiveLineData(120, 132, 101, 134, 90, 230)}},
		Options: controlledInteractiveLineOptions("basic line example", "basic-line-example"),
	}
}

func sampleInteractiveLineLabels() interactive.LineConfig {
	return interactive.LineConfig{
		Label: "Visible point labels", Caption: "Every point exposes its value directly.",
		XAxis:         interactiveLineCategories(),
		Series:        []interactive.LineSeries{{Name: "Category A", Data: fixedInteractiveLineData(150, 232, 201, 154, 190, 130)}},
		SeriesOptions: interactive.SeriesOptions{ShowSymbol: interactive.Bool(true), Label: &interactive.LabelOptions{Show: interactive.Bool(true)}},
		Options:       controlledInteractiveLineOptions("title and label options", "visible-point-labels"),
	}
}

func sampleInteractiveLineSymbols() interactive.LineConfig {
	return interactive.LineConfig{
		Label: "Diamond symbols", Caption: "Two smoothed series use visible diamond points.",
		XAxis: interactiveLineCategories(),
		Series: []interactive.LineSeries{
			{Name: "Category A", Data: fixedInteractiveLineData(120, 182, 191, 234, 290, 330)},
			{Name: "Category B", Data: fixedInteractiveLineData(220, 282, 291, 334, 390, 430)},
		},
		SeriesOptions: interactive.SeriesOptions{Smooth: interactive.Bool(true), ShowSymbol: interactive.Bool(true), Symbol: "diamond", SymbolSize: 15},
		Options:       controlledInteractiveLineOptions("symbol options", "diamond-symbols"),
	}
}

func sampleInteractiveLineMarkPoints() interactive.LineConfig {
	return interactive.LineConfig{
		Label: "Calculated point references", Caption: "Maximum, average, and minimum points are named beside exact values.",
		XAxis: interactiveLineCategories(),
		Series: []interactive.LineSeries{{
			Name: "Category A", Data: fixedInteractiveLineData(120, 132, 101, 134, 90, 230),
			References: interactive.LineReferences{
				Points: []interactive.LinePointReference{
					{Name: "Maximum", Statistic: interactive.LineStatisticMaximum},
					{Name: "Average", Statistic: interactive.LineStatisticAverage},
					{Name: "Minimum", Statistic: interactive.LineStatisticMinimum},
				},
				ShowLabels: interactive.Bool(true),
			},
		}},
		Options: controlledInteractiveLineOptions("mark point options", "calculated-point-references"),
	}
}

func sampleInteractiveLineSplitLines() interactive.LineConfig {
	options := controlledInteractiveLineOptions("split line options", "split-line-options")
	options.YAxis = &interactive.AxisOptions{ShowSplitLine: interactive.Bool(true)}
	return interactive.LineConfig{
		Label: "Visible split lines", Caption: "Horizontal guides support scanning across labeled points.",
		XAxis:   interactiveLineCategories(),
		Series:  []interactive.LineSeries{{Name: "Category A", Data: fixedInteractiveLineData(120, 132, 101, 134, 90, 230), Options: interactive.SeriesOptions{Label: &interactive.LabelOptions{Show: interactive.Bool(true)}}}},
		Options: options,
	}
}

func sampleInteractiveLineNumerical() interactive.LineConfig {
	xValues := make([]float64, 30)
	for index := range xValues {
		xValues[index] = float64(index)
	}
	return interactive.LineConfig{
		Label: "Numerical x axis with guides", Caption: "Thirty ordered coordinates include two scale pieces, a highlighted range, and two guide lines.",
		ValueAxis: &interactive.LineValueAxis{Values: xValues},
		Series: []interactive.LineSeries{{
			Name: "Category A", Data: deterministicLineTimeData(30),
			Options: interactive.SeriesOptions{Symbol: "triangle", SymbolSize: 10, AreaStyle: &interactive.AreaStyle{}},
			References: interactive.LineReferences{
				Areas: []interactive.LineRangeReference{{Name: "Danger zone", StartX: 20, EndX: 25}},
				Lines: []interactive.LineGuideReference{
					{Name: "Danger level", Start: &interactive.LineCoordinate{X: 20, Y: 10}, End: &interactive.LineCoordinate{X: 25, Y: 50}},
					{Name: "Line of no return", X: interactive.Float(28)},
				},
				ShowLabels: interactive.Bool(true), StartSymbol: "square", EndSymbol: "circle", SymbolSize: 10,
			},
		}},
		VisualScale: &interactive.LineVisualScale{Dimension: interactive.LineVisualDimensionX, Pieces: []interactive.LineVisualPiece{
			{GreaterThan: interactive.Float(1), LessThan: interactive.Float(7)},
			{GreaterThan: interactive.Float(10), LessThan: interactive.Float(15)},
		}},
		Options: interactive.ChartOptions{
			Title:    &interactive.TitleOptions{Text: "numerical X axis and accessories", Subtitle: "theme-aware ranges, guides, and scale pieces"},
			Legend:   &interactive.LegendOptions{Bottom: "0"},
			YAxis:    &interactive.AxisOptions{Max: interactive.Float(200)},
			Controls: controlledInteractiveLineOptions("", "numerical-line").Controls,
			Export:   controlledInteractiveLineOptions("", "numerical-line").Export,
		},
	}
}

func sampleInteractiveLineStep() interactive.LineConfig {
	cfg := sampleInteractiveLineBase()
	cfg.Label, cfg.Caption = "Step line", "The line changes after each category boundary."
	cfg.SeriesOptions.Step = "end"
	cfg.Options = controlledInteractiveLineOptions("step style", "step-line")
	return cfg
}

func sampleInteractiveLineSmooth() interactive.LineConfig {
	cfg := sampleInteractiveLineBase()
	cfg.Label, cfg.Caption = "Smooth line", "A curved treatment emphasizes overall direction."
	cfg.SeriesOptions.Smooth = interactive.Bool(true)
	cfg.Options = controlledInteractiveLineOptions("smooth style", "smooth-line")
	return cfg
}

func sampleInteractiveLineArea() interactive.LineConfig {
	cfg := sampleInteractiveLineBase()
	cfg.Label, cfg.Caption = "Area line", "A half-opacity area emphasizes magnitude; the marked interval identifies in-stock categories."
	cfg.SeriesOptions.Label = &interactive.LabelOptions{Show: interactive.Bool(true)}
	cfg.SeriesOptions.AreaStyle = &interactive.AreaStyle{Opacity: interactive.Float(0.5)}
	cfg.Series[0].References = interactive.LineReferences{Areas: []interactive.LineRangeReference{{Name: "In stock", StartX: 2, EndX: 4}}, ShowLabels: interactive.Bool(true)}
	cfg.Options = controlledInteractiveLineOptions("area options", "area-line")
	return cfg
}

func sampleInteractiveLineSmoothArea() interactive.LineConfig {
	cfg := sampleInteractiveLineBase()
	cfg.Label, cfg.Caption = "Smooth area", "A light area fill sits below the smoothed line."
	cfg.SeriesOptions.Label = &interactive.LabelOptions{Show: interactive.Bool(true)}
	cfg.SeriesOptions.AreaStyle = &interactive.AreaStyle{Opacity: interactive.Float(0.2)}
	cfg.SeriesOptions.Smooth = interactive.Bool(true)
	cfg.Options = controlledInteractiveLineOptions("smooth area", "smooth-area-line")
	return cfg
}

func sampleInteractiveLineMulti() interactive.LineConfig {
	return interactive.LineConfig{
		Label: "Four line comparison", Caption: "Four aligned series preserve the upstream multi-line shape.",
		XAxis: interactiveLineCategories(),
		Series: []interactive.LineSeries{
			{Name: "Category A", Data: fixedInteractiveLineData(120, 132, 101, 134, 90, 230)},
			{Name: "Category B", Data: fixedInteractiveLineData(220, 182, 191, 234, 290, 330)},
			{Name: "Category C", Data: fixedInteractiveLineData(150, 232, 201, 154, 190, 130)},
			{Name: "Category D", Data: fixedInteractiveLineData(320, 332, 301, 334, 390, 430)},
		},
		Options: controlledInteractiveLineOptions("multi lines", "multi-lines"),
	}
}

func sampleInteractiveLineDemo() interactive.LineConfig {
	options := controlledInteractiveLineOptions("Search Time: Hash table vs Binary search", "search-time-comparison")
	options.XAxis = &interactive.AxisOptions{Name: "Elements"}
	options.YAxis = &interactive.AxisOptions{Name: "Cost time (ns)", ShowSplitLine: interactive.Bool(true)}
	return interactive.LineConfig{
		Label: "Search time comparison", Caption: "Two labeled series compare cost across increasing element counts.",
		XAxis: []string{"10e1", "10e2", "10e3", "10e4", "10e5", "10e6", "10e7"},
		Series: []interactive.LineSeries{
			{Name: "map", Data: fixedInteractiveLineData(19, 31, 43, 57, 72, 118, 127), Options: interactive.SeriesOptions{Label: &interactive.LabelOptions{Show: interactive.Bool(true), Position: "bottom"}, Smooth: interactive.Bool(true)}, References: interactive.LineReferences{Lines: []interactive.LineGuideReference{{Name: "Average", Statistic: interactive.LineStatisticAverage}}, ShowLabels: interactive.Bool(true)}},
			{Name: "slice", Data: fixedInteractiveLineData(24.9, 34.9, 48.1, 58.3, 69.7, 123, 131), Options: interactive.SeriesOptions{Label: &interactive.LabelOptions{Show: interactive.Bool(true), Position: "top"}, Smooth: interactive.Bool(true)}, References: interactive.LineReferences{Lines: []interactive.LineGuideReference{{Name: "Average", Statistic: interactive.LineStatisticAverage}}, ShowLabels: interactive.Bool(true)}},
		},
		Options: options,
	}
}

func interactiveLineLabelsCode() string {
	return `@interactive.Line(interactive.LineConfig{
  Label: "Visible point labels",
  XAxis: []string{"Apple", "Banana", "Peach"},
  Series: []interactive.LineSeries{{
    Name: "Category A",
    Data: []interactive.LineData{{Value: 150}, {Value: 232}, {Value: 201}},
  }},
  SeriesOptions: interactive.SeriesOptions{
    Label: &interactive.LabelOptions{Show: interactive.Bool(true)},
    ShowSymbol: interactive.Bool(true), Symbol: "diamond", SymbolSize: 15,
  },
})`
}

func interactiveLineReferencesCode() string {
	return `References: interactive.LineReferences{
  Points: []interactive.LinePointReference{
    {Name: "Maximum", Statistic: interactive.LineStatisticMaximum},
    {Name: "Average", Statistic: interactive.LineStatisticAverage},
    {Name: "Minimum", Statistic: interactive.LineStatisticMinimum},
  },
  ShowLabels: interactive.Bool(true),
}`
}

func interactiveLineNumericalCode() string {
	return `@interactive.Line(interactive.LineConfig{
  Label: "Numerical x axis with guides",
  ValueAxis: &interactive.LineValueAxis{Values: []float64{0, 1, 2}},
  Series: []interactive.LineSeries{{
    Name: "Category A",
    Data: []interactive.LineData{{Value: 107}, {Value: 112}, {Value: 118}},
    References: interactive.LineReferences{
      Areas: []interactive.LineRangeReference{{Name: "Danger zone", StartX: 1, EndX: 2}},
    },
  }},
  VisualScale: &interactive.LineVisualScale{
    Dimension: interactive.LineVisualDimensionX,
    Pieces: []interactive.LineVisualPiece{{GreaterThan: interactive.Float(0), LessThan: interactive.Float(2)}},
  },
})`
}

func interactiveLineShapesCode() string {
	return `SeriesOptions: interactive.SeriesOptions{
  Step: "end",
  Smooth: interactive.Bool(true),
  AreaStyle: &interactive.AreaStyle{Opacity: interactive.Float(0.2)},
}`
}

func interactiveLineMultiCode() string {
	return `Series: []interactive.LineSeries{
  {Name: "Category A", Data: categoryA},
  {Name: "Category B", Data: categoryB},
  {Name: "Category C", Data: categoryC},
  {Name: "Category D", Data: categoryD},
}`
}
