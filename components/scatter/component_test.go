package scatter

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	chart "github.com/go-analyze/charts"
)

func TestScatterSupportsSharedControlsAndExport(t *testing.T) {
	t.Parallel()
	instance := Scatter(Config{
		Label: "Samples", Categories: []string{"A"},
		Series:   []Series{{Name: "Values", Points: []Point{{Category: "A", Value: 1}}}},
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "samples"},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	for _, want := range []string{
		`-fullscreen-action`,
		`data-goshtoso-chart-expand`, `-chart-expand-export"`,
		`>SVG</button>`, `>PNG</button>`,
	} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("markup missing %q", want)
		}
	}
}

func TestScatterRendersSSRAccessibleSVG(t *testing.T) {
	t.Parallel()
	instance := Scatter(Config{
		Label:      "Latency by request rate",
		Caption:    "Each point is one service sample.",
		Categories: []string{"10 req/s", "20 req/s"},
		Series: []Series{
			{Name: "API", Points: []Point{{Category: "10 req/s", Value: 42}, {Category: "20 req/s", Value: 58}}, Options: Options{Symbol: SymbolDiamond, Size: 5}},
			{Name: "Worker", Points: []Point{{Category: "10 req/s", Value: 35}, {Category: "20 req/s", Value: 47}}},
		},
		Style:     charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "mx-auto"},
		RootAttrs: templ.Attributes{"id": "latency-scatter", "data-chart-purpose": "correlation"},
	})
	if instance.Kind() != chartcomponents.KindScatterChart {
		t.Fatalf("Kind() = %q, want %q", instance.Kind(), chartcomponents.KindScatterChart)
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`<figure class="goshtoso-charts-scatter goshtoso-charts-palette goshtoso-charts-palette-araihu mx-auto" role="img" aria-label="Latency by request rate"`,
		`id="latency-scatter"`, `data-chart-purpose="correlation"`, `class="goshtoso-charts-scatter__viewport"`, "<svg",
		"Each point is one service sample.", "#123456", "var(--color-chart-surface)",
		"var(--font-paragraph), sans-serif",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q:\n%s", want, markup)
		}
	}
	if strings.Contains(markup, "echarts.init") {
		t.Errorf("SSR chart unexpectedly contains interactive renderer initialization: %s", markup)
	}
}

func TestScatterMapsCategoricalPointsAndTypedOptions(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Label:      "Samples",
		Categories: []string{"10", "20"},
		Options:    Options{Symbol: SymbolCircle, Size: 3},
		Series: []Series{
			{Name: "First", Points: []Point{{Category: "20", Value: 2}, {Category: "10", Value: 1}}},
			{Name: "Second", Points: []Point{{Category: "10", Value: 3}, {Category: "10", Value: 4}}, Options: Options{Symbol: SymbolSquare, Size: 6}},
		},
	}
	options := scatterOptions(cfg)
	if got := strings.Join(options.XAxis.Labels, ","); got != "10,20" {
		t.Fatalf("XAxis.Labels = %q, want declared category order", got)
	}
	if got := options.Legend.SeriesNames; len(got) != 2 || got[0] != "First" || got[1] != "Second" {
		t.Fatalf("Legend.SeriesNames = %v", got)
	}
	if options.Symbol.Shape != chart.SymbolCircle || options.Symbol.Size != 3 {
		t.Fatalf("global symbol = %#v", options.Symbol)
	}
	if options.SeriesList[0].Symbol.Shape != chart.SymbolCircle || options.SeriesList[0].Symbol.Size != 3 {
		t.Fatalf("inherited series symbol = %#v", options.SeriesList[0].Symbol)
	}
	if options.SeriesList[1].Symbol.Shape != chart.SymbolSquare || options.SeriesList[1].Symbol.Size != 6 {
		t.Fatalf("overridden series symbol = %#v", options.SeriesList[1].Symbol)
	}
	if got := options.SeriesList[1].Values[0]; len(got) != 2 || got[0] != 3 || got[1] != 4 {
		t.Fatalf("repeated category values = %v, want both samples", got)
	}
}

func TestScatterMapsDenseAlignedDataAndRendererNeutralOptions(t *testing.T) {
	t.Parallel()
	minimum, maximum, gap := 0.0, 280.0, false
	cfg := Config{
		Label: "Dense", Categories: []string{"foo 0", "foo 1", "foo 2"},
		Series: []Series{
			{Name: "One", Values: [][]float64{{10}, {}, {11, 12, 13}}, Options: Options{ReferenceLine: ReferenceLineMaximum}},
			{Name: "Two", Values: [][]float64{{20}, {21}, {22}}},
		},
		Options: Options{Size: .5, Trend: TrendLine{Kind: TrendSimpleMovingAverage, Period: 2}, ValueFormat: ValueFormatHumanized},
		Title:   TitleOptions{Text: "Dense Scatter Chart Demo", Subtext: "samples", Placement: PlacementCenter},
		Legend:  LegendOptions{Hidden: true, Orientation: LegendVertical, Placement: PlacementRight, Alignment: AlignmentRight, FontSize: 6},
		XAxis:   CategoryAxisOptions{BoundaryGap: &gap, LabelCount: 10, LabelFontSize: 6, LabelRotation: 45},
		YAxis:   ValueAxisOptions{Min: &minimum, Max: &maximum, Unit: 10, LabelSkip: 1, LabelFontSize: 6},
		Padding: Padding{Top: 16, Right: 32, Bottom: 16, Left: 16},
	}
	options := scatterOptions(cfg)
	if got := options.SeriesList[0].Values[2]; len(got) != 3 || got[2] != 13 {
		t.Fatalf("aligned repeated samples = %v", got)
	}
	if options.Symbol.Size != .5 || len(options.SeriesList[0].TrendLine) != 1 || options.SeriesList[0].TrendLine[0].Period != 2 {
		t.Fatalf("dense marker/trend options missing: %#v", options.SeriesList[0])
	}
	if len(options.SeriesList[0].MarkLine.Lines) != 1 || options.SeriesList[0].MarkLine.Lines[0].Type != chart.SeriesMarkTypeMax {
		t.Fatalf("maximum reference missing: %#v", options.SeriesList[0].MarkLine)
	}
	if options.XAxis.LabelCount != 10 || math.Abs(options.XAxis.LabelRotation-math.Pi/4) > .0001 || options.XAxis.BoundaryGap == nil || *options.XAxis.BoundaryGap {
		t.Fatalf("x axis = %#v", options.XAxis)
	}
	if options.YAxis[0].Min == nil || *options.YAxis[0].Min != 0 || options.YAxis[0].Max == nil || *options.YAxis[0].Max != 280 || options.YAxis[0].Unit != 10 || options.YAxis[0].LabelSkipCount != 1 {
		t.Fatalf("y axis = %#v", options.YAxis[0])
	}
	if options.Title.Text != "Dense Scatter Chart Demo" || options.Title.Subtext != "samples" || options.Title.Offset.Left != chart.PositionCenter || options.Legend.Show == nil || *options.Legend.Show || options.Legend.Offset.Left != chart.PositionRight || options.Legend.Vertical == nil || !*options.Legend.Vertical {
		t.Fatalf("title/legend options missing: %#v %#v", options.Title, options.Legend)
	}
	if options.Padding.Left != 16 || options.Padding.Right != 32 {
		t.Fatalf("padding = %#v", options.Padding)
	}
}

func TestScatterTopNLabelsSelectExactlyNWithStableTiesForDenseAndSparseData(t *testing.T) {
	t.Parallel()
	for _, series := range []Series{
		{Name: "Dense", Values: [][]float64{{4, 9}, {9}, {3}}},
		{Name: "Sparse", Points: []Point{{Category: "B", Value: 9}, {Category: "A", Value: 4}, {Category: "A", Value: 9}, {Category: "C", Value: 3}}},
	} {
		cfg := Config{Label: "Top values", Categories: []string{"A", "B", "C"}, Series: []Series{series}, Options: Options{TopNLabels: TopNLabels{Count: 2}}}
		options := scatterOptions(cfg)
		formatter := options.SeriesList[0].Label.LabelFormatter
		if formatter == nil || options.SeriesList[0].Label.Show == nil || !*options.SeriesList[0].Label.Show {
			t.Fatalf("top N labels were not enabled: %#v", options.SeriesList[0].Label)
		}
		got := make([]string, 0)
		for _, value := range seriesValues(cfg.Categories, series) {
			text, _ := formatter(0, "", value.value)
			if text != "" {
				got = append(got, value.category+":"+text)
			}
		}
		if want := []string{"A:9", "B:9"}; !reflect.DeepEqual(got, want) {
			t.Errorf("%s selected = %v, want %v", series.Name, got, want)
		}
	}
}

func TestScatterTopNLabelsDisabledAndBoundedByAvailableSamples(t *testing.T) {
	t.Parallel()
	base := Config{Label: "Top values", Categories: []string{"A", "B"}, Series: []Series{{Name: "Values", Values: [][]float64{{2}, {1}}}}}
	if options := scatterOptions(base); options.SeriesList[0].Label.Show != nil {
		t.Fatalf("zero top N enabled labels: %#v", options.SeriesList[0].Label)
	}
	base.Options.TopNLabels.Count = 9
	rows := base.topNLabelRows()
	if len(rows) != 2 || !rows[0].Selected || !rows[1].Selected {
		t.Fatalf("top N above samples = %#v", rows)
	}
}

func TestScatterDefaultsToFilledDotAndThemeToken(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Label:      "Default markers",
		Categories: []string{"A"},
		Series:     []Series{{Name: "Samples", Points: []Point{{Category: "A", Value: 1}}}},
	}
	options := scatterOptions(cfg)
	if options.Symbol.Shape != chart.SymbolDot || options.SeriesList[0].Symbol.Shape != chart.SymbolDot {
		t.Fatalf("default symbols = chart %q, series %q; want filled dots", options.Symbol.Shape, options.SeriesList[0].Symbol.Shape)
	}
	var output bytes.Buffer
	if err := Scatter(cfg).Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(output.String(), "var(--color-chart-series-1)") {
		t.Fatalf("default marker does not use theme series token: %s", output.String())
	}
}

func TestScatterKeepsNumericLookingKeysCategorical(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Label:      "Categorical samples",
		Categories: []string{"1", "2", "100"},
		Series: []Series{{
			Name: "Samples",
			Points: []Point{
				{Category: "1", Value: 10},
				{Category: "2", Value: 20},
				{Category: "100", Value: 30},
			},
		}},
	}
	options := scatterOptions(cfg)
	if got := strings.Join(options.XAxis.Labels, ","); got != "1,2,100" {
		t.Fatalf("XAxis.Labels = %q, want explicit categorical keys", got)
	}
	if len(options.SeriesList[0].Values) != 3 {
		t.Fatalf("series positions = %d, want one equally spaced slot per category", len(options.SeriesList[0].Values))
	}
}

func TestScatterValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{name: "label", cfg: Config{}, want: "label is required"},
		{name: "series", cfg: Config{Label: "Points"}, want: "at least one series"},
		{name: "categories", cfg: Config{Label: "Points", Series: []Series{{Name: "A"}}}, want: "at least one category"},
		{name: "duplicate category", cfg: Config{Label: "Points", Categories: []string{"A", "A"}, Series: []Series{{Name: "A", Points: []Point{{Category: "A", Value: 2}}}}}, want: `category "A" is duplicated`},
		{name: "series name", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Points: []Point{{Category: "A", Value: 2}}}}}, want: "series 1 needs a name"},
		{name: "point", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Name: "A"}}}, want: `series "A" needs points or aligned values`},
		{name: "mixed data", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Name: "A", Points: []Point{{Category: "A", Value: 1}}, Values: [][]float64{{2}}}}}, want: `cannot use both points and aligned values`},
		{name: "dense length", cfg: Config{Label: "Points", Categories: []string{"A", "B"}, Series: []Series{{Name: "A", Values: [][]float64{{1}}}}}, want: `aligned values length 1 must match 2 categories`},
		{name: "dense value", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Name: "A", Values: [][]float64{{math.Inf(1)}}}}}, want: `sample 0 must contain a finite value`},
		{name: "empty dense", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Name: "A", Values: [][]float64{{}}}}}, want: `needs at least one aligned value`},
		{name: "trend period", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Name: "A", Values: [][]float64{{1}}}}, Options: Options{Trend: TrendLine{Kind: TrendSimpleMovingAverage, Period: 2}}}, want: `trend period cannot exceed category count`},
		{name: "axis range", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Name: "A", Values: [][]float64{{1}}}}, YAxis: ValueAxisOptions{Min: float64Pointer(2), Max: float64Pointer(1)}}, want: `minimum must be less than maximum`},
		{name: "padding", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Name: "A", Values: [][]float64{{1}}}}, Padding: Padding{Right: -1}}, want: `padding cannot be negative`},
		{name: "category", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Name: "A", Points: []Point{{Category: "B", Value: 2}}}}}, want: `references unknown category "B"`},
		{name: "value", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Name: "A", Points: []Point{{Category: "A", Value: math.NaN()}}}}}, want: "must contain a finite value"},
		{name: "symbol", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Name: "A", Points: []Point{{Category: "A", Value: 2}}}}, Options: Options{Symbol: "star"}}, want: `unsupported symbol "star"`},
		{name: "top N", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Name: "A", Points: []Point{{Category: "A", Value: 2}}}}, Options: Options{TopNLabels: TopNLabels{Count: -1}}}, want: "top N label count cannot be negative"},
		{name: "top N color class", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Name: "A", Points: []Point{{Category: "A", Value: 2}}}}, Options: Options{TopNLabels: TopNLabels{Color: "red", Class: "label"}}}, want: "top N labels cannot set both color and class"},
		{name: "series color class", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Name: "A", Points: []Point{{Category: "A", Value: 2}}, Color: "red", Class: "series"}}}, want: "cannot set both color and class"},
		{name: "size", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Name: "A", Points: []Point{{Category: "A", Value: 2}}}}, Options: Options{Size: -1}}, want: "size must be a finite non-negative number"},
		{name: "width", cfg: Config{Label: "Points", Width: -1, Categories: []string{"A"}, Series: []Series{{Name: "A", Points: []Point{{Category: "A", Value: 2}}}}}, want: "width cannot be negative"},
		{name: "root attr", cfg: Config{Label: "Points", RootAttrs: templ.Attributes{"role": "presentation"}, Categories: []string{"A"}, Series: []Series{{Name: "A", Points: []Point{{Category: "A", Value: 2}}}}}, want: `root attribute "role" is reserved`},
		{name: "root attr case", cfg: Config{Label: "Points", RootAttrs: templ.Attributes{"Aria-Label": "override"}, Categories: []string{"A"}, Series: []Series{{Name: "A", Points: []Point{{Category: "A", Value: 2}}}}}, want: `root attribute "Aria-Label" is reserved`},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := renderSVG(test.cfg)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("renderSVG() error = %v, want %q", err, test.want)
			}
		})
	}
}

func float64Pointer(value float64) *float64 { return &value }

func TestScatterEscapesProgrammaticSeriesColors(t *testing.T) {
	t.Parallel()
	instance := Scatter(Config{
		Label:      "Safe chart",
		Categories: []string{"A"},
		Series:     []Series{{Name: "value", Points: []Point{{Category: "A", Value: 2}}}},
		Style:      charttheme.Style{Colors: []string{`red" onload="alert(1)`}},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	if strings.Contains(markup, `fill="red" onload=`) {
		t.Fatalf("programmatic color escaped its SVG attribute: %s", markup)
	}
	if !strings.Contains(markup, `red&#34; onload=&#34;alert(1)`) {
		t.Fatalf("escaped programmatic color missing from SVG: %s", markup)
	}
}

func TestScatterDenseRenderStaysBounded(t *testing.T) {
	t.Parallel()
	categories := make([]string, 1000)
	values := make([][]float64, 1000)
	for index := range categories {
		categories[index] = fmt.Sprintf("foo %d", index)
		values[index] = []float64{float64(index % 280)}
		if index%2 == 0 {
			values[index] = append(values[index], float64((index+1)%280))
		}
	}
	start := time.Now()
	var output bytes.Buffer
	if err := Scatter(Config{Label: "Dense", Categories: categories, Series: []Series{{Name: "One", Values: values}}, Width: 600, Height: 400}).Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Fatalf("dense render took %s", elapsed)
	}
	if output.Len() < 100_000 {
		t.Fatalf("dense render unexpectedly small: %d bytes", output.Len())
	}
	if strings.Count(output.String(), "<figcaption") > 1 {
		t.Fatal("dense accessibility output exploded into per-point captions")
	}
}
