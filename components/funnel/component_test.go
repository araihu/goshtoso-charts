package funnel

import (
	"bytes"
	"context"
	"encoding/xml"
	"math"
	"strings"
	"testing"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func upstreamConfig() Config {
	return Config{
		Label: "Basic funnel",
		Title: "Funnel",
		Stages: []Stage{
			{Label: "Show", Value: 100}, {Label: "Click", Value: 80}, {Label: "Visit", Value: 60},
			{Label: "Inquiry", Value: 40}, {Label: "Order", Value: 20}, {Label: "Pay", Value: 10},
			{Label: "Cancel", Value: 2},
		},
		Options: Options{Legend: Legend{Padding: Padding{Left: 100}}},
	}
}

func TestFunnelRendersAccessibleUpstreamShapeAndSharedControls(t *testing.T) {
	t.Parallel()
	cfg := upstreamConfig()
	cfg.Caption = "Seven ordered stages from Show to Cancel."
	cfg.Style = charttheme.Style{Palette: charttheme.PaletteAraiHu, Class: "mx-auto"}
	cfg.Stages[0].Color = "#365314"
	cfg.Stages[0].Class = "stage-show"
	cfg.RootAttrs = templ.Attributes{"id": "basic-funnel", "data-chart-purpose": "conversion"}
	cfg.Controls = chartcontrol.Options{Fullscreen: true}
	cfg.Export = &chartcontrol.ExportOptions{Filename: "basic-funnel"}

	instance := Funnel(cfg)
	if instance.Kind() != chartcomponents.KindFunnelChart {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`<figure class="goshtoso-charts-funnel goshtoso-charts-palette goshtoso-charts-palette-araihu mx-auto" role="img" aria-label="Basic funnel"`,
		`id="basic-funnel"`, `data-chart-purpose="conversion"`, `<svg`, `viewBox="0 0 600 400"`,
		"Funnel", "Show", "Click", "Visit", "Inquiry", "Order", "Pay", "Cancel",
		"100", "80", "60", "40", "20", "10", "2", "Seven ordered stages from Show to Cancel.",
		"Exact stage values", "Share of first stage", `aria-label="Basic funnel exact stage values"`, "stage-show", "#365314",
		"var(--color-chart-series-2)", "var(--color-chart-text-strong)", "var(--font-paragraph), sans-serif",
		`data-goshtoso-chart-expand`, `-chart-expand-export"`, `>SVG</button>`, `>PNG</button>`,
		`-fullscreen-action`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	for _, unwanted := range []string{"rgb(1,1,1)", "rgb(4,4,4)", "rgb(10,30,50)", "go-analyze"} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("rendered markup contains %q", unwanted)
		}
	}
}

func TestFunnelMechanicallyMapsPinnedUpstreamDataAndGeometry(t *testing.T) {
	t.Parallel()
	cfg := upstreamConfig()
	wantLabels := []string{"Show", "Click", "Visit", "Inquiry", "Order", "Pay", "Cancel"}
	wantValues := []float64{100, 80, 60, 40, 20, 10, 2}
	if cfg.width() != 600 || cfg.height() != 400 {
		t.Fatalf("dimensions = %dx%d, want upstream 600x400", cfg.width(), cfg.height())
	}
	options := funnelOptions(cfg)
	if options.Title.Text != "Funnel" || options.Legend.Padding.Left != 100 {
		t.Fatalf("upstream title/legend geometry = %#v", options)
	}
	if len(options.SeriesList) != len(wantLabels) || len(options.Legend.SeriesNames) != len(wantLabels) {
		t.Fatalf("series/legend counts = %d/%d", len(options.SeriesList), len(options.Legend.SeriesNames))
	}
	for index := range wantLabels {
		if options.SeriesList[index].Name != wantLabels[index] || options.Legend.SeriesNames[index] != wantLabels[index] || options.SeriesList[index].Value != wantValues[index] {
			t.Fatalf("stage %d = (%q, %q, %g), want (%q, %q, %g)", index, options.SeriesList[index].Name, options.Legend.SeriesNames[index], options.SeriesList[index].Value, wantLabels[index], wantLabels[index], wantValues[index])
		}
	}
}

func TestFunnelMapsTypedPresentationOptions(t *testing.T) {
	t.Parallel()
	cfg := upstreamConfig()
	cfg.Options = Options{
		Labels:  LabelNameValue,
		Legend:  Legend{Orientation: LegendVertical, Placement: LegendEnd, Padding: Padding{Top: 4, Right: 5, Bottom: 6, Left: 7}},
		Padding: Padding{Top: 8, Right: 9, Bottom: 10, Left: 11},
	}
	options := funnelOptions(cfg)
	if options.Legend.Vertical == nil || !*options.Legend.Vertical || options.Legend.Offset.Left != "right" {
		t.Fatalf("legend options = %#v", options.Legend)
	}
	if options.Legend.Padding.Top != 4 || options.Legend.Padding.Right != 5 || options.Legend.Padding.Bottom != 6 || options.Legend.Padding.Left != 7 {
		t.Fatalf("legend padding = %#v", options.Legend.Padding)
	}
	if options.Padding.Top != 8 || options.Padding.Right != 9 || options.Padding.Bottom != 10 || options.Padding.Left != 11 {
		t.Fatalf("chart padding = %#v", options.Padding)
	}
	text, _ := options.SeriesList[1].Label.LabelFormatter(1, "Click", 80)
	if text != "Click (80)" {
		t.Fatalf("label text = %q", text)
	}
	cfg.Options.Labels = LabelHidden
	hidden := funnelOptions(cfg)
	if hidden.SeriesList[0].Label.Show == nil || *hidden.SeriesList[0].Label.Show {
		t.Fatalf("hidden labels = %#v", hidden.SeriesList[0].Label.Show)
	}
}

func TestFunnelSVGIsDeterministicAndParseable(t *testing.T) {
	t.Parallel()
	first, err := renderSVG(upstreamConfig())
	if err != nil {
		t.Fatalf("first render: %v", err)
	}
	second, err := renderSVG(upstreamConfig())
	if err != nil {
		t.Fatalf("second render: %v", err)
	}
	if first != second {
		t.Fatal("identical funnel config produced different SVG bytes")
	}
	decoder := xml.NewDecoder(strings.NewReader(first))
	for {
		if _, err := decoder.Token(); err != nil {
			if err.Error() == "EOF" {
				break
			}
			t.Fatalf("parse SVG: %v", err)
		}
	}
}

func TestFunnelValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{name: "label", edit: func(c *Config) { c.Label = "" }, want: "label is required"},
		{name: "stages", edit: func(c *Config) { c.Stages = nil }, want: "at least one stage"},
		{name: "stage label", edit: func(c *Config) { c.Stages[0].Label = "" }, want: "stage 1 needs a label"},
		{name: "duplicate", edit: func(c *Config) { c.Stages[1].Label = "Show" }, want: `stage label "Show" is duplicated`},
		{name: "nan", edit: func(c *Config) { c.Stages[0].Value = math.NaN() }, want: `stage "Show" value must be finite and non-negative`},
		{name: "infinite", edit: func(c *Config) { c.Stages[0].Value = math.Inf(1) }, want: `stage "Show" value must be finite and non-negative`},
		{name: "negative", edit: func(c *Config) { c.Stages[1].Value = -1 }, want: `stage "Click" value must be finite and non-negative`},
		{name: "increasing", edit: func(c *Config) { c.Stages[1].Value = 101 }, want: `stage "Click" value cannot exceed previous stage "Show"`},
		{name: "width", edit: func(c *Config) { c.Width = -1 }, want: "width cannot be negative"},
		{name: "height", edit: func(c *Config) { c.Height = -1 }, want: "height cannot be negative"},
		{name: "label mode", edit: func(c *Config) { c.Options.Labels = "both-ish" }, want: `label mode "both-ish" is unsupported`},
		{name: "legend orientation", edit: func(c *Config) { c.Options.Legend.Orientation = "diagonal" }, want: `legend orientation "diagonal" is unsupported`},
		{name: "legend placement", edit: func(c *Config) { c.Options.Legend.Placement = "middle-ish" }, want: `legend placement "middle-ish" is unsupported`},
		{name: "padding", edit: func(c *Config) { c.Options.Padding.Left = -1 }, want: "padding cannot be negative"},
		{name: "legend padding", edit: func(c *Config) { c.Options.Legend.Padding.Top = -1 }, want: "legend padding cannot be negative"},
		{name: "unsafe color", edit: func(c *Config) { c.Stages[0].Color = "red;stroke:black" }, want: `stage "Show" color is unsafe`},
		{name: "unsafe class", edit: func(c *Config) { c.Stages[0].Class = `stage" onload="alert(1)` }, want: `stage "Show" class is unsafe`},
		{name: "root attr", edit: func(c *Config) { c.RootAttrs = templ.Attributes{"Aria-Label": "override"} }, want: `root attribute "Aria-Label" is reserved`},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			cfg := upstreamConfig()
			cfg.Stages = append([]Stage(nil), cfg.Stages...)
			test.edit(&cfg)
			err := cfg.validate()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validate() error = %v, want %q", err, test.want)
			}
		})
	}
}
