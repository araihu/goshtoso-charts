package scatter

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	chart "github.com/go-analyze/charts"
)

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
	if strings.Contains(markup, "<script") {
		t.Errorf("SSR chart unexpectedly contains script: %s", markup)
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
		{name: "point", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Name: "A"}}}, want: `series "A" needs at least one point`},
		{name: "category", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Name: "A", Points: []Point{{Category: "B", Value: 2}}}}}, want: `references unknown category "B"`},
		{name: "value", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Name: "A", Points: []Point{{Category: "A", Value: math.NaN()}}}}}, want: "must contain a finite value"},
		{name: "symbol", cfg: Config{Label: "Points", Categories: []string{"A"}, Series: []Series{{Name: "A", Points: []Point{{Category: "A", Value: 2}}}}, Options: Options{Symbol: "star"}}, want: `unsupported symbol "star"`},
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
