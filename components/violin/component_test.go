package violin

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func validConfig() Config {
	return Config{
		Label: "Distribution shapes", Caption: "Four deterministic sample distributions.", Title: "Distribution Shapes",
		Series: []Series{
			{Name: "Normal", Samples: []float64{35, 42, 48, 50, 51, 57, 65}, Color: "#365314", Class: "distribution-normal", Marks: MarkLines{Mean: true, Median: true}, Statistics: Statistics{Quantiles: []float64{.25, .75}}},
			{Name: "Tight", Samples: []float64{46, 48, 49, 50, 51, 52, 54}, Marks: MarkLines{Mean: true}},
		},
		Density: Distribution{Points: 40}, Width: 640, Height: 420,
	}
}

func render(t *testing.T, cfg Config) string {
	t.Helper()
	var output bytes.Buffer
	if err := Violin(cfg).Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return output.String()
}

func TestRenderIsDeterministicAccessibleAndThemeAware(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	first, second := render(t, cfg), render(t, cfg)
	if first != second {
		t.Fatal("identical violin config produced different output")
	}
	for _, want := range []string{
		`role="img"`, `aria-label="Distribution shapes"`, "Distribution Shapes", "Normal", "Tight",
		"Exact sample statistics", "Minimum", "Q1", "Median", "Mean", "Q3", "Maximum",
		"25% = 45.00", "75% = 54.00", "distribution-normal", "#365314",
		`preserveAspectRatio="xMidYMid meet"`,
		"var(--color-chart-surface)", "var(--color-chart-text)", "var(--font-paragraph)",
		`data-goshtoso-chart-expand`, `data-goshtoso-chart-export-menu`, ">SVG</button>", ">PNG</button>",
	} {
		if !strings.Contains(first, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	for _, unwanted := range []string{"<script src=\"http", "<image", "go-analyze", "onclick="} {
		if strings.Contains(first, unwanted) {
			t.Errorf("rendered markup contains %q", unwanted)
		}
	}
}

func TestSummaryUsesLinearQuantilesAndExactPopulationMean(t *testing.T) {
	t.Parallel()
	got := summarize([]float64{9, 1, 3, 7, 5})
	want := summary{Count: 5, Minimum: 1, Q1: 3, Median: 5, Mean: 5, Q3: 7, Maximum: 9}
	if got != want {
		t.Fatalf("summarize() = %#v, want %#v", got, want)
	}
	if got := quantile([]float64{0, 10}, .25); got != 2.5 {
		t.Fatalf("quantile() = %g, want 2.5", got)
	}
}

func TestValidationRejectsInvalidContracts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*Config)
		want   string
	}{
		{"label", func(cfg *Config) { cfg.Label = " " }, "label is required"},
		{"series", func(cfg *Config) { cfg.Series = nil }, "at least one series"},
		{"duplicate", func(cfg *Config) { cfg.Series[1].Name = cfg.Series[0].Name }, "duplicated"},
		{"samples", func(cfg *Config) { cfg.Series[0].Samples = []float64{1} }, "at least 2 samples"},
		{"finite", func(cfg *Config) { cfg.Series[0].Samples[0] = math.NaN() }, "must be finite"},
		{"variance", func(cfg *Config) { cfg.Series[0].Samples = []float64{1, 1} }, "varying samples"},
		{"points", func(cfg *Config) { cfg.Density.Points = 1 }, "density points"},
		{"bandwidth", func(cfg *Config) { cfg.Density.Bandwidth = math.Inf(1) }, "bandwidth"},
		{"normalization", func(cfg *Config) { cfg.Density.Normalization = "area" }, "normalization"},
		{"limit", func(cfg *Config) { cfg.Axis.Limit = -1 }, "axis limit"},
		{"padding", func(cfg *Config) { cfg.Padding.Left = -1 }, "padding"},
		{"quantile", func(cfg *Config) { cfg.Series[0].Statistics.Quantiles = []float64{1} }, "quantiles"},
		{"color", func(cfg *Config) { cfg.Series[0].Color = "url(https://bad.example)" }, "color is unsafe"},
		{"class", func(cfg *Config) { cfg.Series[0].Class = `bad" class="escape` }, "class is unsafe"},
		{"attrs", func(cfg *Config) { cfg.RootAttrs = templ.Attributes{"role": "group"} }, "reserved"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := validConfig()
			test.mutate(&cfg)
			var output bytes.Buffer
			err := Violin(cfg).Render(context.Background(), &output)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Render() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestControlsRemainExplicitOptIns(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Controls = chartcontrol.Options{Fullscreen: true, Collapsible: true, Expand: chartcontrol.Bool(false)}
	cfg.Export = &chartcontrol.ExportOptions{Disabled: true}
	markup := render(t, cfg)
	for _, want := range []string{`data-goshtoso-chart-control="fullscreen"`, `data-goshtoso-chart-control="collapse"`} {
		if !strings.Contains(markup, want) {
			t.Errorf("opt-in markup missing %q", want)
		}
	}
	for _, unwanted := range []string{"data-goshtoso-chart-expand", "data-goshtoso-chart-export-menu"} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("opted-out markup contains %q", unwanted)
		}
	}
}

func TestCallerStyleAndSeriesColorWin(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Style = charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#112233"}, Class: "caller-root"}
	cfg.Series[0].Color = "#abcdef"
	markup := render(t, cfg)
	for _, want := range []string{"goshtoso-charts-palette-araihu", "caller-root", "#abcdef"} {
		if !strings.Contains(markup, want) {
			t.Errorf("styled markup missing %q", want)
		}
	}
}
