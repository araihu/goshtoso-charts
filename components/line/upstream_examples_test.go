package line

import (
	"bytes"
	"context"
	"math"
	"reflect"
	"strings"
	"testing"

	chart "github.com/go-analyze/charts"
)

func renderMarkup(cfg Config) (string, error) {
	var output bytes.Buffer
	err := Line(cfg).Render(context.Background(), &output)
	return output.String(), err
}

func TestLineMapsBasicExampleGapAndPresentation(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Label:  "Basic line chart",
		Title:  Title{Text: "Line", FontSize: 16},
		Labels: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
		Series: []Series{
			{Name: "Email", Points: []Point{{Value: 120}, {Value: 132}, {Value: 101}, {Missing: true}, {Value: 90}, {Value: 230}, {Value: 210}}},
			{Name: "Union Ads", Values: []float64{220, 182, 191, 234, 290, 330, 310}},
			{Name: "Video Ads", Values: []float64{150, 232, 201, 154, 190, 330, 410}},
			{Name: "Direct", Values: []float64{320, 332, 301, 334, 390, 330, 320}},
			{Name: "Search Engine", Values: []float64{820, 932, 901, 934, 1290, 1330, 1320}},
		},
		Legend:      LegendOptions{Padding: Padding{Left: 100}},
		Symbol:      Symbol{Shape: SymbolCircle},
		StrokeWidth: 1.2,
		Width:       600,
		Height:      400,
	}
	options := lineOptions(cfg)
	if options.Title.Text != "Line" || options.Title.FontStyle.FontSize != 16 || options.LineStrokeWidth != 1.2 {
		t.Fatalf("title/stroke = %#v / %g", options.Title, options.LineStrokeWidth)
	}
	if options.Symbol.Shape != chart.SymbolCircle || options.Legend.Padding.Left != 100 {
		t.Fatalf("symbol/legend = %#v / %#v", options.Symbol, options.Legend)
	}
	if got := options.SeriesList[0].Values; len(got) != 7 || got[3] != chart.GetNullValue() {
		t.Fatalf("gap mapping = %#v", got)
	}
	if err := cfg.validate(); err != nil {
		t.Fatalf("validate() error = %v", err)
	}
}

func TestLineMapsPerSeriesSymbolsAndSmoothTreatment(t *testing.T) {
	t.Parallel()
	labels := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	values := []float64{120, 132, 101, 96, 90, 230, 210}
	cfg := Config{
		Label:  "Series symbols",
		Labels: labels,
		Series: []Series{
			{Name: "Email", Values: values, Symbol: Symbol{Shape: SymbolCircle}},
			{Name: "Union Ads", Values: values, Symbol: Symbol{Shape: SymbolDiamond}},
			{Name: "Video Ads", Values: values, Symbol: Symbol{Shape: SymbolSquare}},
			{Name: "Direct", Values: values, Symbol: Symbol{Shape: SymbolDot}},
		},
		StrokeWidth: 1.2,
	}
	options := lineOptions(cfg)
	want := []chart.SymbolShape{chart.SymbolCircle, chart.SymbolDiamond, chart.SymbolSquare, chart.SymbolDot}
	for index := range want {
		if options.SeriesList[index].Symbol.Shape != want[index] {
			t.Errorf("series %d symbol = %q, want %q", index, options.SeriesList[index].Symbol.Shape, want[index])
		}
	}

	cfg.Legend.Hidden = true
	cfg.Symbol = Symbol{Shape: SymbolNone}
	cfg.StrokeWidth = 4
	cfg.SmoothingTension = .9
	options = lineOptions(cfg)
	if options.Legend.Show == nil || *options.Legend.Show || options.Symbol.Shape != chart.SymbolNone || options.LineStrokeWidth != 4 || options.StrokeSmoothingTension != .9 {
		t.Fatalf("smooth treatment = legend %v symbol %#v width %g tension %g", options.Legend.Show, options.Symbol, options.LineStrokeWidth, options.StrokeSmoothingTension)
	}
}

func TestLineMapsStatisticalReferencesAndAccessibleEvidence(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Label:  "Marked line chart",
		Labels: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
		Series: []Series{{
			Name: "Email", Values: []float64{120, 132, 101, 95, 90, 230, 210},
			References: References{Average: true, Maximum: true, PointSize: 30, PointPrefix: "Max:", Format: ValueFormatHumanized, Decimals: 1},
		}},
		Padding: Padding{Top: 20, Right: 48, Bottom: 20, Left: 20},
	}
	options := lineOptions(cfg)
	series := options.SeriesList[0]
	if len(series.MarkLine.Lines) != 1 || series.MarkLine.Lines[0].Type != chart.SeriesMarkTypeAverage {
		t.Fatalf("mark lines = %#v", series.MarkLine.Lines)
	}
	if len(series.MarkPoint.Points) != 1 || series.MarkPoint.Points[0].Type != chart.SeriesMarkTypeMax || series.MarkPoint.SymbolSize != 30 {
		t.Fatalf("mark points = %#v size %d", series.MarkPoint.Points, series.MarkPoint.SymbolSize)
	}
	if got := series.MarkLine.ValueFormatter(4.64); got != "4.6" || series.MarkPoint.ValueFormatter(4.64) != "Max:4.6" {
		t.Fatalf("reference formatters = %q / %q", got, series.MarkPoint.ValueFormatter(4.64))
	}
	if options.Padding != (chart.Box{Top: 20, Right: 48, Bottom: 20, Left: 20}) {
		t.Fatalf("padding = %#v", options.Padding)
	}
	markup, err := renderMarkup(cfg)
	if err != nil {
		t.Fatalf("renderMarkup() error = %v", err)
	}
	for _, want := range []string{"Average", "139.7", "Maximum", "230", "Sat"} {
		if !strings.Contains(markup, want) {
			t.Errorf("accessible reference evidence missing %q", want)
		}
	}
}

func TestLineMapsStackedSeriesAndDataLabels(t *testing.T) {
	t.Parallel()
	noGap := false
	cfg := Config{
		Label:  "Stacked line chart",
		Labels: []string{"1", "2", "3", "4", "5", "6", "7", "8"},
		Series: []Series{
			{Name: "A", Values: []float64{1.9, 23.2, 25.6, 102.6, 142.2, 32.6, 20, 2.3}, Labels: DataLabelOptions{Show: true, Format: ValueFormatHumanized, Decimals: 1, TrailingZeros: true}, References: References{Maximum: true, PointSize: 30, PointPrefix: "Max:", Format: ValueFormatHumanized}},
			{Name: "B", Values: []float64{12, 26.4, 28.7, 144.6, 122.2, 48.7, 18.8, 13.3}, Labels: DataLabelOptions{Show: true, Format: ValueFormatHumanized, Decimals: 1, TrailingZeros: true}, References: References{Maximum: true, PointSize: 30, PointPrefix: "Max:", Format: ValueFormatHumanized}},
			{Name: "C", Values: []float64{80, 40.4, 28.4, 48.8, 24.4, 24.2, 40.8, 80.8}, Labels: DataLabelOptions{Show: true, Format: ValueFormatHumanized, Decimals: 1, TrailingZeros: true}, References: References{Maximum: true, PointSize: 30, PointPrefix: "Max:", Format: ValueFormatHumanized}},
		},
		Stacked: true,
		XAxis:   CategoryAxisOptions{BoundaryGap: &noGap},
		YAxes:   []Axis{{Title: "A+B+C Sum", TitleFontSize: 12, LabelFontSize: 8}},
		Padding: Padding{Top: 10, Right: 40, Bottom: 10, Left: 10},
	}
	options := lineOptions(cfg)
	if options.StackSeries == nil || !*options.StackSeries || options.XAxis.BoundaryGap == nil || *options.XAxis.BoundaryGap {
		t.Fatalf("stack/gap = %v / %v", options.StackSeries, options.XAxis.BoundaryGap)
	}
	if options.YAxis[0].Title != "A+B+C Sum" || options.SeriesList[0].Label.Show == nil || !*options.SeriesList[0].Label.Show {
		t.Fatalf("axis/labels = %#v / %#v", options.YAxis[0], options.SeriesList[0].Label)
	}
	if label := options.SeriesList[0].Label.ValueFormatter(23.24); label != "23.2" {
		t.Fatalf("data label = %q", label)
	}
	if label := options.SeriesList[0].Label.ValueFormatter(20); label != "20.0" {
		t.Fatalf("trailing-zero data label = %q", label)
	}
	if options.YAxis[0].TitleFontStyle.FontSize != 12 || options.YAxis[0].LabelFontStyle.FontSize != 8 {
		t.Fatalf("axis font sizes = %#v / %#v", options.YAxis[0].TitleFontStyle, options.YAxis[0].LabelFontStyle)
	}
}

func TestLineMapsDenseAxisControlsAndPositionedLabels(t *testing.T) {
	t.Parallel()
	minimum, maximum := 1.4, 8.0
	cfg := Config{
		Label:       "Canon RF Zoom Lenses",
		Title:       Title{Text: "Canon RF Zoom Lenses", Placement: PlacementCenter, FontSize: 16},
		Labels:      []string{"60mm", "61mm", "62mm", "63mm"},
		Series:      []Series{{Name: "70-200mm f/2.8", Points: []Point{{Missing: true}, {Value: 2.8}, {Value: 2.8}, {Missing: true}}}},
		Legend:      LegendOptions{Hidden: true},
		Symbol:      Symbol{Shape: SymbolNone},
		StrokeWidth: 1.5,
		XAxis:       CategoryAxisOptions{BoundaryGap: boolPointer(true), Unit: 40, LabelCount: 10, LabelRotation: 45, LabelFontSize: 6},
		YAxes:       []Axis{{Hidden: true, Min: &minimum, Max: &maximum, LabelCount: 4, SpineLine: true, LabelFontSize: 8}},
		Padding:     Padding{Top: 20, Right: 20, Bottom: 10, Left: 20},
		Annotations: []TextAnnotation{{Text: "70-200mm f/2.8", X: 420, Y: 84, SeriesIndex: 0, FontSize: 12}},
		Width:       600, Height: 400,
	}
	options := lineOptions(cfg)
	if options.Title.Offset != chart.OffsetCenter || options.XAxis.Unit != 40 || options.XAxis.LabelCount != 10 || math.Abs(options.XAxis.LabelRotation-math.Pi/4) > .0001 || options.XAxis.LabelFontStyle.FontSize != 6 {
		t.Fatalf("title/x-axis = %#v / %#v", options.Title, options.XAxis)
	}
	axis := options.YAxis[0]
	if axis.Show == nil || *axis.Show || axis.LabelCount != 4 || axis.SpineLineShow == nil || !*axis.SpineLineShow || axis.LabelFontStyle.FontSize != 8 {
		t.Fatalf("y-axis = %#v", axis)
	}
	svg, err := renderSVG(cfg)
	if err != nil {
		t.Fatalf("renderSVG() error = %v", err)
	}
	for _, want := range []string{"70-200mm f/2.8", "var(--color-chart-series-1)"} {
		if !strings.Contains(svg, want) {
			t.Errorf("positioned label SVG missing %q", want)
		}
	}
}

func TestLineMapsThemeAwareColdToWarmDataLabels(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Label:   "Sales performance",
		Title:   Title{Text: "Sales Performance with Gradient Label Colors", Subtext: "Cold = Low Values, Warm = High Values"},
		Labels:  []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct"},
		Series:  []Series{{Name: "Sales Performance", Values: []float64{20, 15, 35, 40, 10, 55, 25, 45, 30, 50}, Labels: DataLabelOptions{Show: true, ColorScale: LabelColorScaleColdToWarm}}},
		Legend:  LegendOptions{Hidden: true},
		Padding: Padding{Top: 20, Right: 20, Bottom: 20, Left: 20},
		Width:   800, Height: 500,
	}
	options := lineOptions(cfg)
	if options.Title.Subtext != cfg.Title.Subtext || options.SeriesList[0].Label.Show == nil || !*options.SeriesList[0].Label.Show {
		t.Fatalf("title/labels = %#v / %#v", options.Title, options.SeriesList[0].Label)
	}
	svg, err := renderSVG(cfg)
	if err != nil {
		t.Fatalf("renderSVG() error = %v", err)
	}
	for _, want := range []string{"var(--color-chart-scale-low)", "var(--color-chart-scale-mid)", "var(--color-chart-scale-high)"} {
		if !strings.Contains(svg, want) {
			t.Errorf("gradient label SVG missing %q", want)
		}
	}
	for _, unwanted := range []string{"rgb(123,", "go-analyze"} {
		if strings.Contains(svg, unwanted) {
			t.Errorf("gradient label SVG leaks %q", unwanted)
		}
	}
}

func TestLineRejectsInvalidExtendedOptions(t *testing.T) {
	t.Parallel()
	base := Config{Label: "Line", Labels: []string{"A", "B"}, Series: []Series{{Name: "Value", Values: []float64{1, 2}}}}
	tests := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{name: "both values and points", edit: func(cfg *Config) { cfg.Series[0].Points = []Point{{Value: 1}, {Value: 2}} }, want: "cannot use both values and points"},
		{name: "missing data", edit: func(cfg *Config) { cfg.Series[0].Values = nil }, want: "needs values or points"},
		{name: "point length", edit: func(cfg *Config) { cfg.Series[0].Values = nil; cfg.Series[0].Points = []Point{{Value: 1}} }, want: "has 1 points; need 2"},
		{name: "point nan", edit: func(cfg *Config) {
			cfg.Series[0].Values = nil
			cfg.Series[0].Points = []Point{{Value: math.NaN()}, {Value: 2}}
		}, want: "point 0 must be finite"},
		{name: "symbol", edit: func(cfg *Config) { cfg.Symbol.Shape = SymbolShape("triangle") }, want: "symbol shape"},
		{name: "symbol size", edit: func(cfg *Config) { cfg.Symbol.Size = -1 }, want: "symbol size"},
		{name: "stroke width", edit: func(cfg *Config) { cfg.StrokeWidth = -1 }, want: "stroke width"},
		{name: "smoothing", edit: func(cfg *Config) { cfg.SmoothingTension = 1.1 }, want: "smoothing tension"},
		{name: "x rotation", edit: func(cfg *Config) { cfg.XAxis.LabelRotation = 361 }, want: "X axis label rotation"},
		{name: "y label count", edit: func(cfg *Config) { cfg.YAxes = []Axis{{LabelCount: -1}} }, want: "Y axis 0 label count"},
		{name: "y label font size", edit: func(cfg *Config) { cfg.YAxes = []Axis{{LabelFontSize: math.NaN()}} }, want: "Y axis 0 label font size"},
		{name: "y title font size", edit: func(cfg *Config) { cfg.YAxes = []Axis{{TitleFontSize: -1}} }, want: "Y axis 0 title font size"},
		{name: "label scale", edit: func(cfg *Config) { cfg.Series[0].Labels.ColorScale = LabelColorScale("rainbow") }, want: "label color scale"},
		{name: "reference decimals", edit: func(cfg *Config) { cfg.Series[0].References.Decimals = -1 }, want: "reference decimals"},
		{name: "annotation coordinates", edit: func(cfg *Config) { cfg.Annotations = []TextAnnotation{{Text: "Label", X: -1}} }, want: "annotation 1 coordinates"},
		{name: "annotation series", edit: func(cfg *Config) { cfg.Annotations = []TextAnnotation{{Text: "Label", SeriesIndex: 2}} }, want: "annotation 1 series index"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := base
			cfg.Labels = append([]string(nil), base.Labels...)
			cfg.Series = append([]Series(nil), base.Series...)
			test.edit(&cfg)
			if err := cfg.validate(); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestLineExtendedOptionsDoNotMutateInput(t *testing.T) {
	t.Parallel()
	cfg := Config{Label: "Line", Labels: []string{"A", "B"}, Series: []Series{{Name: "Value", Points: []Point{{Value: 1}, {Missing: true}}}}}
	want := append([]Point(nil), cfg.Series[0].Points...)
	_ = lineOptions(cfg)
	if !reflect.DeepEqual(cfg.Series[0].Points, want) {
		t.Fatalf("input points mutated: %#v", cfg.Series[0].Points)
	}
}

func boolPointer(value bool) *bool { return &value }
