package pages

import (
	"math/rand"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	interactive "github.com/araihu/goshtoso-charts/components/interactive"
)

const (
	interactiveBarUpstreamPath     = "examples/bar.go"
	interactiveBarUpstreamRevision = "bda428480a82d6d77ebb9fa939cf8d52528453dd"
	interactiveBarUpstreamSHA256   = "dcda545f978fdd055ecff5a6050b2ad9dc8cf9fe350bd7e4768952e8068fc9f9"

	barCoverageExample     = "example"
	barCoverageUnsupported = "unsupported"
)

type interactiveBarCoverageEntry struct {
	Name      string
	Status    string
	Treatment string
	Reason    string
}

func interactiveBarUpstreamCoverage() []interactiveBarCoverageEntry {
	return []interactiveBarCoverageEntry{
		{Name: "barBasic", Status: barCoverageExample, Treatment: "basic presentation"},
		{Name: "barTitle", Status: barCoverageExample, Treatment: "basic presentation"},
		{Name: "barTooltip", Status: barCoverageExample, Treatment: "basic presentation"},
		{Name: "barSetToolbox", Status: barCoverageExample, Treatment: "shared controls, PNG export, and exact-value disclosure"},
		{Name: "barShowLabel", Status: barCoverageExample, Treatment: "visible value labels"},
		{Name: "barXYName", Status: barCoverageExample, Treatment: "axis names and units"},
		{Name: "barXYFormatter", Status: barCoverageExample, Treatment: "literal axis units"},
		{Name: "barColor", Status: barCoverageExample, Treatment: "explicit color override"},
		{Name: "barSplitLine", Status: barCoverageExample, Treatment: "axis names and units"},
		{Name: "barGap", Status: barCoverageExample, Treatment: "bar width and gap"},
		{Name: "barDataZoomInside", Status: barCoverageExample, Treatment: "inside category zoom"},
		{Name: "barDataZoomSlider", Status: barCoverageExample, Treatment: "slider category zoom"},
		{Name: "barReverse", Status: barCoverageExample, Treatment: "horizontal orientation"},
		{Name: "barStack", Status: barCoverageExample, Treatment: "stacked series"},
		{Name: "barMarkPoints", Status: barCoverageExample, Treatment: "point references"},
		{Name: "barMarkLines", Status: barCoverageExample, Treatment: "guide references"},
		{Name: "barOverlap", Status: barCoverageUnsupported, Reason: "mixed Bar, Line, and Scatter composition requires a renderer-neutral composite chart API"},
		{Name: "barSize", Status: barCoverageExample, Treatment: "large responsive canvas"},
		{Name: "barWidth", Status: barCoverageExample, Treatment: "bar width and gap"},
	}
}

type interactiveBarSource struct {
	Path   string
	SHA256 string
	Scope  string
}

func interactiveBarSupplementarySources() []interactiveBarSource {
	return []interactiveBarSource{
		{Path: "examples/page_center_layout.go", SHA256: "106456904719dfacfb13adcc1b9e66df83cf28a5a801539bad4d1958554166c9", Scope: "page layout reference"},
		{Path: "examples/page_flex_layout.go", SHA256: "3113b7bdf78a2365ae62502fe86ab001f3ff3034b1d77752c693e95b28a0fd68", Scope: "page layout reference"},
		{Path: "examples/page_none_layout.go", SHA256: "ce38424de2ffeb919661e536c7f44921de098ae14643d4f2975d8e72296c32f8", Scope: "page layout reference"},
		{Path: "examples/themes.go", SHA256: "843c478c63b9cf3ab13b1e13518ea98912332bb34caf0dae5d48343fabd121a0", Scope: "site theme and chart-token reference"},
		{Path: "examples/renderer.go", SHA256: "c4956db261f554c6a161c0d25baa7dbd7c2c179523997d297020cd55916e6a3f", Scope: "private renderer integration; not a chart option"},
		{Path: "examples/bar3d.go", SHA256: "110b3b85f2528d76eb8271b64f1facd81a974e30ecc0dd77319d5a409ff64275", Scope: "separate existing Bar 3D component"},
	}
}

func interactiveBarCategories() []string {
	return []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
}

func fixedInteractiveBarData(seed int64) []interactive.BarData {
	source := rand.New(rand.NewSource(seed))
	data := make([]interactive.BarData, len(interactiveBarCategories()))
	for index := range data {
		data[index].Value = float64(source.Intn(300))
	}
	return data
}

func interactiveBarSeries(name string, seed int64) interactive.BarSeries {
	return interactive.BarSeries{Name: name, Data: fixedInteractiveBarData(seed)}
}

func controlledInteractiveBarOptions(title, filename string) interactive.ChartOptions {
	return interactive.ChartOptions{
		Title:    &interactive.TitleOptions{Text: title},
		Legend:   &interactive.LegendOptions{Bottom: "0"},
		Tooltip:  &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "axis"},
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: filename},
	}
}

func sampleInteractiveBar() interactive.BarConfig {
	options := controlledInteractiveBarOptions("basic bar example", "basic-bar-example")
	options.Title.Subtitle = "This is the subtitle."
	return interactive.BarConfig{
		Label: "Basic bar example", Caption: "Two deterministic seven-day series preserve the upstream categorical shape.",
		XAxis: interactiveBarCategories(),
		Series: []interactive.BarSeries{
			interactiveBarSeries("Category A", 11),
			interactiveBarSeries("Category B", 12),
		},
		Options: options,
	}
}

func sampleInteractiveBarLabels() interactive.BarConfig {
	return interactive.BarConfig{
		Label: "Visible value labels", Caption: "Every bar exposes its exact value above the mark and in the adjacent table.",
		XAxis: interactiveBarCategories(),
		Series: []interactive.BarSeries{
			interactiveBarSeries("Category A", 21),
			interactiveBarSeries("Category B", 22),
		},
		SeriesOptions: interactive.SeriesOptions{Label: &interactive.LabelOptions{Show: interactive.Bool(true), Position: "top"}},
		Options:       controlledInteractiveBarOptions("label options", "visible-bar-labels"),
	}
}

func sampleInteractiveBarAxes() interactive.BarConfig {
	options := controlledInteractiveBarOptions("axis names, units, and split lines", "bar-axis-options")
	options.XAxis = &interactive.AxisOptions{Name: "XAxisName", LabelSuffix: " x-unit", ShowSplitLine: interactive.Bool(true)}
	options.YAxis = &interactive.AxisOptions{Name: "YAxisName", LabelSuffix: " y-unit", ShowSplitLine: interactive.Bool(true)}
	return interactive.BarConfig{
		Label: "Named axes with literal units", Caption: "Axis names, unit suffixes, and split lines clarify how categories and values are read.",
		XAxis:   interactiveBarCategories(),
		Series:  []interactive.BarSeries{interactiveBarSeries("Category A", 31), interactiveBarSeries("Category B", 32)},
		Options: options,
	}
}

func sampleInteractiveBarColors() interactive.BarConfig {
	return interactive.BarConfig{
		Label: "Explicit series colors", Caption: "Caller-selected colors override the theme palette for both series.",
		XAxis:   interactiveBarCategories(),
		Series:  []interactive.BarSeries{interactiveBarSeries("Category A", 41), interactiveBarSeries("Category B", 42)},
		Style:   charttheme.Style{Colors: []string{"#2563eb", "#db2777"}},
		Options: controlledInteractiveBarOptions("user-defined colors", "explicit-bar-colors"),
	}
}

func sampleInteractiveBarWidthsAndGap() interactive.BarConfig {
	return interactive.BarConfig{
		Label: "Bar widths and gap", Caption: "One absolute width, one percentage width, and a 150% inter-series gap preserve the upstream size treatments.",
		XAxis: interactiveBarCategories(),
		Series: []interactive.BarSeries{
			{Name: "Category A", Data: fixedInteractiveBarData(51), Options: interactive.SeriesOptions{BarWidth: "35"}},
			{Name: "Category B", Data: fixedInteractiveBarData(52), Options: interactive.SeriesOptions{BarWidth: "15%"}},
		},
		SeriesOptions: interactive.SeriesOptions{BarGap: "150%"},
		Options:       controlledInteractiveBarOptions("bar width and gap", "bar-width-and-gap"),
	}
}

func sampleInteractiveBarHorizontal() interactive.BarConfig {
	return interactive.BarConfig{
		Label: "Horizontal bar orientation", Caption: "Categories move to the vertical axis for easier comparison of long labels.",
		XAxis: interactiveBarCategories(), Orientation: interactive.BarOrientationHorizontal,
		Series:  []interactive.BarSeries{interactiveBarSeries("Category A", 61), interactiveBarSeries("Category B", 62)},
		Options: controlledInteractiveBarOptions("reverse category and value axes", "horizontal-bar"),
	}
}

func sampleInteractiveBarStacked() interactive.BarConfig {
	return interactive.BarConfig{
		Label: "Stacked bar series", Caption: "Both series share one stack so each category shows a combined total.",
		XAxis:         interactiveBarCategories(),
		Series:        []interactive.BarSeries{interactiveBarSeries("Category A", 71), interactiveBarSeries("Category B", 72)},
		SeriesOptions: interactive.SeriesOptions{Stack: "stackA"},
		Options:       controlledInteractiveBarOptions("stack style", "stacked-bar"),
	}
}

func sampleInteractiveBarZoom(mode interactive.BarZoomMode) interactive.BarConfig {
	label, title, filename := "Inside category zoom", "category zoom (inside)", "inside-bar-zoom"
	if mode == interactive.BarZoomSlider {
		label, title, filename = "Slider category zoom", "category zoom (slider)", "slider-bar-zoom"
	}
	return interactive.BarConfig{
		Label: label, Caption: "The initial window shows 10% through 50% of the seven ordered categories.",
		XAxis:  interactiveBarCategories(),
		Series: []interactive.BarSeries{interactiveBarSeries("Category A", 81), interactiveBarSeries("Category B", 82)},
		Zoom:   &interactive.BarZoom{Mode: mode, StartPercent: 10, EndPercent: 50},
		Height: "460px", Options: controlledInteractiveBarOptions(title, filename),
	}
}

func sampleInteractiveBarMarkPoints() interactive.BarConfig {
	categoryA := fixedInteractiveBarData(91)
	categoryA[0].Value = 100
	calculated := []interactive.BarPointReference{
		{Name: "Maximum", Statistic: interactive.BarStatisticMaximum},
		{Name: "Minimum", Statistic: interactive.BarStatisticMinimum},
	}
	return interactive.BarConfig{
		Label: "Bar point references", Caption: "A named Monday point sits beside calculated minimum and maximum markers.",
		XAxis: interactiveBarCategories(),
		Series: []interactive.BarSeries{
			{Name: "Category A", Data: categoryA, References: interactive.BarReferences{Points: append([]interactive.BarPointReference{{Name: "special mark", Coordinate: &interactive.BarCoordinate{Category: "Mon", Value: 100}, Label: &interactive.LabelOptions{Show: interactive.Bool(true), Position: "inside"}}}, calculated...), ShowLabels: interactive.Bool(true)}},
			{Name: "Category B", Data: fixedInteractiveBarData(92), References: interactive.BarReferences{Points: calculated, ShowLabels: interactive.Bool(true)}},
		},
		Options: controlledInteractiveBarOptions("mark point options", "bar-point-references"),
	}
}

func sampleInteractiveBarMarkLines() interactive.BarConfig {
	guides := []interactive.BarGuideReference{
		{Name: "Maximum", Statistic: interactive.BarStatisticMaximum},
		{Name: "Average", Statistic: interactive.BarStatisticAverage},
	}
	return interactive.BarConfig{
		Label: "Bar guide references", Caption: "Maximum and average guides summarize both seven-value series.",
		XAxis: interactiveBarCategories(),
		Series: []interactive.BarSeries{
			{Name: "Category A", Data: fixedInteractiveBarData(101), References: interactive.BarReferences{Lines: guides}},
			{Name: "Category B", Data: fixedInteractiveBarData(102), References: interactive.BarReferences{Lines: guides}},
		},
		Options: controlledInteractiveBarOptions("mark line options", "bar-guide-references"),
	}
}

func sampleInteractiveBarLargeCanvas() interactive.BarConfig {
	return interactive.BarConfig{
		Label: "Large bar canvas", Caption: "The upstream 1200 by 600 canvas becomes container-wide while preserving the 600-pixel height.",
		XAxis:  interactiveBarCategories(),
		Series: []interactive.BarSeries{interactiveBarSeries("Category A", 111), interactiveBarSeries("Category B", 112)},
		Width:  "100%", Height: "600px",
		Options: controlledInteractiveBarOptions("large canvas size", "large-bar-canvas"),
	}
}

func interactiveChartBarCode() string {
	return `@interactive.Bar(interactive.BarConfig{
  Label: "Basic bar example",
  XAxis: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
  Series: []interactive.BarSeries{
    {Name: "Category A", Data: categoryA},
    {Name: "Category B", Data: categoryB},
  },
  Options: interactive.ChartOptions{
    Title: &interactive.TitleOptions{Text: "basic bar example", Subtitle: "This is the subtitle."},
    Tooltip: &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "axis"},
  },
})`
}

func interactiveBarAxesCode() string {
	return `Options: interactive.ChartOptions{
  XAxis: &interactive.AxisOptions{Name: "XAxisName", LabelSuffix: " x-unit", ShowSplitLine: interactive.Bool(true)},
  YAxis: &interactive.AxisOptions{Name: "YAxisName", LabelSuffix: " y-unit", ShowSplitLine: interactive.Bool(true)},
}`
}

func interactiveBarLayoutCode() string {
	return `Orientation: interactive.BarOrientationHorizontal,
SeriesOptions: interactive.SeriesOptions{Stack: "stackA", BarGap: "150%"},
Series: []interactive.BarSeries{
  {Name: "Category A", Data: categoryA, Options: interactive.SeriesOptions{BarWidth: "35"}},
  {Name: "Category B", Data: categoryB, Options: interactive.SeriesOptions{BarWidth: "15%"}},
}`
}

func interactiveBarZoomCode() string {
	return `Zoom: &interactive.BarZoom{
  Mode: interactive.BarZoomSlider,
  StartPercent: 10,
  EndPercent: 50,
}`
}

func interactiveBarReferencesCode() string {
	return `References: interactive.BarReferences{
  Points: []interactive.BarPointReference{
    {Name: "Maximum", Statistic: interactive.BarStatisticMaximum},
    {Name: "special mark", Coordinate: &interactive.BarCoordinate{Category: "Mon", Value: 100}},
  },
  Lines: []interactive.BarGuideReference{{Name: "Average", Statistic: interactive.BarStatisticAverage}},
  ShowLabels: interactive.Bool(true),
}`
}
