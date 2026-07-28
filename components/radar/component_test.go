package radar

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"

	"github.com/a-h/templ"
	chartassets "github.com/araihu/goshtoso-charts/assets"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	chart "github.com/go-analyze/charts"
)

func validConfig() Config {
	return Config{
		Label: "Budget comparison",
		Indicators: []Indicator{
			{Name: "Sales", Max: 6500},
			{Name: "Administration", Max: 16000},
			{Name: "Technology", Max: 30000},
		},
		Series: []Series{
			{Name: "Allocated Budget", Values: []float64{4200, 3000, 20000}},
			{Name: "Actual Spending", Values: []float64{5000, 14000, 28000}},
		},
	}
}

func TestRadarRendersSSRAccessibleSVGAndExactValues(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Caption = "Allocated and actual spending."
	cfg.Options = Options{RadiusPercent: 45, ValueLabels: ValueLabelsShown}
	cfg.Series[1].Options = SeriesOptions{ValueLabels: ValueLabelsHidden}
	cfg.Style = charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "mx-auto"}
	cfg.RootAttrs = templ.Attributes{"id": "budget-radar", "data-chart-purpose": "comparison"}

	instance := Radar(cfg)
	if instance.Kind() != chartcomponents.KindRadarChart {
		t.Fatalf("Kind() = %q, want %q", instance.Kind(), chartcomponents.KindRadarChart)
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`<figure class="goshtoso-charts-radar goshtoso-charts-palette goshtoso-charts-palette-araihu mx-auto" role="img" aria-label="Budget comparison"`,
		`id="budget-radar"`, `data-chart-purpose="comparison"`,
		`class="goshtoso-charts-radar__viewport mx-auto overflow-x-auto"`, "<svg",
		"Allocated and actual spending.", "#123456", "var(--color-chart-surface)",
		"var(--font-paragraph), sans-serif", "Exact values", "Allocated Budget", "Actual Spending",
		"Sales", "6500", "4200", "5000",
		".goshtoso-charts-radar__viewport > svg", "min-width: 36rem",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q:\n%s", want, markup)
		}
	}
	if !strings.Contains(markup, `src="`+chartassets.ControlRuntimeURL+`"`) {
		t.Errorf("SSR chart missing shared controls runtime")
	}
	if got := strings.Count(markup, "<script"); got != 1 {
		t.Errorf("SSR chart script count = %d, want shared controls runtime only", got)
	}
	for _, want := range []string{`data-goshtoso-chart-expand`, `data-goshtoso-chart-export-menu`, `>SVG</button>`, `>PNG</button>`} {
		if !strings.Contains(markup, want) {
			t.Errorf("radar shared wrapper missing %q", want)
		}
	}
	if strings.Contains(markup, "rgba(5,5,5") || strings.Contains(markup, "rgba(6,6,6") {
		t.Errorf("radar area fill retained private placeholder colors: %s", markup)
	}
	if !strings.Contains(markup, "color-mix(in srgb, #123456 10%, transparent)") {
		t.Errorf("explicit series color missing from translucent area fill: %s", markup)
	}
}

func TestRadarMapsIndicatorsSeriesAndTypedOptions(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Options = Options{RadiusPercent: 52, ValueLabels: ValueLabelsShown}
	cfg.Series[1].Options = SeriesOptions{ValueLabels: ValueLabelsHidden}
	options := radarOptions(cfg)

	if options.Radius != "52%" {
		t.Fatalf("Radius = %q, want 52%%", options.Radius)
	}
	if len(options.RadarIndicators) != 3 || options.RadarIndicators[0].Name != "Sales" || options.RadarIndicators[0].Max != 6500 {
		t.Fatalf("RadarIndicators = %#v", options.RadarIndicators)
	}
	if got := options.Legend.SeriesNames; len(got) != 2 || got[0] != "Allocated Budget" || got[1] != "Actual Spending" {
		t.Fatalf("Legend.SeriesNames = %v", got)
	}
	if options.SeriesList[0].Name != "Allocated Budget" || len(options.SeriesList[0].Values) != 3 || options.SeriesList[0].Values[2] != 20000 {
		t.Fatalf("first series = %#v", options.SeriesList[0])
	}
	if options.SeriesList[0].Label.Show == nil || !*options.SeriesList[0].Label.Show {
		t.Fatalf("first series labels = %#v, want shown", options.SeriesList[0].Label.Show)
	}
	if options.SeriesList[1].Label.Show == nil || *options.SeriesList[1].Label.Show {
		t.Fatalf("second series labels = %#v, want hidden override", options.SeriesList[1].Label.Show)
	}
}

func TestRadarUsesUpstreamDimensionsAndThemeDefaults(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	if cfg.width() != 600 || cfg.height() != 400 {
		t.Fatalf("default dimensions = %dx%d, want upstream 600x400", cfg.width(), cfg.height())
	}
	options := radarOptions(cfg)
	if options.Radius != "40%" {
		t.Fatalf("default Radius = %q, want 40%%", options.Radius)
	}
	var output bytes.Buffer
	if err := Radar(cfg).Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	for _, want := range []string{"var(--color-chart-series-1)", "var(--color-chart-series-2)"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("default radar missing theme token %q", want)
		}
	}
}

func TestRadarValidation(t *testing.T) {
	t.Parallel()
	base := validConfig()
	tests := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{name: "label", edit: func(cfg *Config) { cfg.Label = "" }, want: "label is required"},
		{name: "axis count", edit: func(cfg *Config) {
			cfg.Indicators = cfg.Indicators[:2]
			cfg.Series[0].Values = cfg.Series[0].Values[:2]
			cfg.Series[1].Values = cfg.Series[1].Values[:2]
		}, want: "at least three indicators"},
		{name: "indicator name", edit: func(cfg *Config) { cfg.Indicators[0].Name = "" }, want: "indicator 1 needs a name"},
		{name: "duplicate indicator", edit: func(cfg *Config) { cfg.Indicators[1].Name = "Sales" }, want: `indicator "Sales" is duplicated`},
		{name: "indicator max zero", edit: func(cfg *Config) { cfg.Indicators[0].Max = 0 }, want: `indicator "Sales" max must be a finite positive number`},
		{name: "indicator max infinite", edit: func(cfg *Config) { cfg.Indicators[0].Max = math.Inf(1) }, want: `indicator "Sales" max must be a finite positive number`},
		{name: "series", edit: func(cfg *Config) { cfg.Series = nil }, want: "at least one series"},
		{name: "series name", edit: func(cfg *Config) { cfg.Series[0].Name = "" }, want: "series 1 needs a name"},
		{name: "series length", edit: func(cfg *Config) { cfg.Series[0].Values = cfg.Series[0].Values[:2] }, want: `series "Allocated Budget" has 2 values; need 3`},
		{name: "finite value", edit: func(cfg *Config) { cfg.Series[0].Values[0] = math.NaN() }, want: "value 1 must be finite"},
		{name: "negative value", edit: func(cfg *Config) { cfg.Series[0].Values[0] = -1 }, want: "value 1 cannot be negative"},
		{name: "over max", edit: func(cfg *Config) { cfg.Series[0].Values[0] = 6501 }, want: `value 1 exceeds indicator "Sales" max 6500`},
		{name: "radius", edit: func(cfg *Config) { cfg.Options.RadiusPercent = 101 }, want: "radius percent must be zero or between 1 and 100"},
		{name: "value labels", edit: func(cfg *Config) { cfg.Options.ValueLabels = "sometimes" }, want: `unsupported value labels "sometimes"`},
		{name: "width", edit: func(cfg *Config) { cfg.Width = -1 }, want: "width cannot be negative"},
		{name: "height", edit: func(cfg *Config) { cfg.Height = -1 }, want: "height cannot be negative"},
		{name: "root attr", edit: func(cfg *Config) { cfg.RootAttrs = templ.Attributes{"role": "presentation"} }, want: `root attribute "role" is reserved`},
		{name: "root attr case", edit: func(cfg *Config) { cfg.RootAttrs = templ.Attributes{"Aria-Label": "override"} }, want: `root attribute "Aria-Label" is reserved`},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			cfg := base
			cfg.Indicators = append([]Indicator(nil), base.Indicators...)
			cfg.Series = append([]Series(nil), base.Series...)
			for index := range cfg.Series {
				cfg.Series[index].Values = append([]float64(nil), base.Series[index].Values...)
			}
			test.edit(&cfg)
			_, err := renderSVG(cfg)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("renderSVG() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestRadarEscapesProgrammaticSeriesColors(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Style = charttheme.Style{Colors: []string{`red" onload="alert(1)`}}
	var output bytes.Buffer
	if err := Radar(cfg).Render(context.Background(), &output); err != nil {
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

var _ = chart.ChartOutputSVG
