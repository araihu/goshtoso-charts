package bar

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	chartassets "github.com/araihu/goshtoso-charts/assets"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func TestBarRendersSSRAccessibleSVG(t *testing.T) {
	t.Parallel()
	instance := Bar(Config{
		Label: "Deployments by environment", Caption: "This week.", Labels: []string{"Development", "Production"},
		Series: []Series{{Name: "Successful", Values: []float64{14, 8}}, {Name: "Failed", Values: []float64{1, 2}}}, Stacked: true,
	})
	if instance.Kind() != chartcomponents.KindBarChart {
		t.Fatalf("Kind() = %q, want %q", instance.Kind(), chartcomponents.KindBarChart)
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{"<figure class=\"goshtoso-charts-bar goshtoso-charts-palette goshtoso-charts-palette-auto\" role=\"img\" aria-label=\"Deployments by environment\"", "<svg", "This week.", "var(--color-chart-series-1)", "var(--color-chart-surface)", `[data-theme="araihu"] .goshtoso-charts-palette-auto`} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	if !strings.Contains(markup, `src="`+chartassets.ControlRuntimeURL+`"`) {
		t.Errorf("SSR chart missing shared controls runtime")
	}
	if got := strings.Count(markup, "<script"); got != 1 {
		t.Errorf("SSR chart script count = %d, want shared controls runtime only", got)
	}
}

func TestBarRejectsMisalignedSeries(t *testing.T) {
	t.Parallel()
	_, err := renderSVG(Config{Label: "Deployments", Labels: []string{"Development", "Production"}, Series: []Series{{Name: "Successful", Values: []float64{14}}}})
	if err == nil || !strings.Contains(err.Error(), "has 1 values; need 2") {
		t.Fatalf("renderSVG() error = %v, want alignment error", err)
	}
}

func horizontalWorldPopulationConfig() Config {
	return Config{
		Label:       "World population by reporting series",
		Title:       "World Population",
		Orientation: OrientationHorizontal,
		Labels:      []string{"UN", "Brazil", "Indonesia", "USA", "India", "China", "World"},
		Series: []Series{
			{Name: "2011", Values: []float64{10, 30, 50, 70, 90, 110, 130}},
			{Name: "2012", Values: []float64{20, 40, 60, 80, 100, 120, 140}},
		},
		Padding: Padding{Top: 20, Right: 40, Bottom: 20, Left: 20},
		Width:   600,
		Height:  400,
	}
}

func TestHorizontalBarPreservesPinnedUpstreamDataGeometryAndAccessibleSummary(t *testing.T) {
	t.Parallel()
	cfg := horizontalWorldPopulationConfig()
	cfg.Caption = "Population comparison for 2011 and 2012."
	cfg.Style = charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#112233"}, Class: "caller-chart"}
	cfg.Controls = chartcontrol.Options{Fullscreen: true, Collapsible: true}
	cfg.Export = &chartcontrol.ExportOptions{Filename: "world-population"}

	options := barOptions(cfg)
	if !options.Horizontal || options.Title.Text != "World Population" {
		t.Fatalf("horizontal title options = %#v", options)
	}
	if options.Padding.Top != 20 || options.Padding.Right != 40 || options.Padding.Bottom != 20 || options.Padding.Left != 20 {
		t.Fatalf("padding = %#v, want top 20 right 40 bottom 20 left 20", options.Padding)
	}
	if !reflect.DeepEqual(options.CategoryAxis.Labels, cfg.Labels) || !reflect.DeepEqual(options.Legend.SeriesNames, []string{"2011", "2012"}) {
		t.Fatalf("axis/legend mapping = labels %v names %v", options.CategoryAxis.Labels, options.Legend.SeriesNames)
	}
	for seriesIndex, want := range [][]float64{{10, 30, 50, 70, 90, 110, 130}, {20, 40, 60, 80, 100, 120, 140}} {
		if !reflect.DeepEqual(options.SeriesList[seriesIndex].Values, want) {
			t.Fatalf("series %d data = %v, want %v", seriesIndex, options.SeriesList[seriesIndex].Values, want)
		}
	}

	var output bytes.Buffer
	if err := Bar(cfg).Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`viewBox="0 0 600 400"`, `preserveAspectRatio="xMidYMid meet"`,
		"World Population", "UN", "Brazil", "Indonesia", "USA", "India", "China", "World",
		"2011", "2012", "Exact category values", "caller-chart", "#112233",
		"stroke:var(--color-chart-text);stroke-width:1",
		"Population comparison for 2011 and 2012.", "goshtoso-charts-bar__viewport",
		`aria-label="World population by reporting series exact category values"`,
		`data-goshtoso-chart-expand`, `data-goshtoso-chart-export-menu`,
		`>SVG</button>`, `>PNG</button>`, `-fullscreen-action`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("horizontal markup missing %q", want)
		}
	}
	for _, row := range []string{
		`<th class="px-3 py-2 font-semibold" scope="row">UN</th><td class="px-3 py-2 tabular-nums">10</td><td class="px-3 py-2 tabular-nums">20</td>`,
		`<th class="px-3 py-2 font-semibold" scope="row">World</th><td class="px-3 py-2 tabular-nums">130</td><td class="px-3 py-2 tabular-nums">140</td>`,
	} {
		if !strings.Contains(markup, row) {
			t.Errorf("exact-value table missing row %q", row)
		}
	}
	if got := strings.Count(markup, "<path"); got < 14 {
		t.Errorf("horizontal SVG path count = %d, want at least 14 bars plus axes", got)
	}
}

func TestVerticalDefaultRemainsByteCompatible(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Label: "Compatibility probe", Labels: []string{"A", "B"},
		Series:  []Series{{Name: "One", Values: []float64{1, 2}}, {Name: "Two", Values: []float64{3, 4}}},
		Stacked: true, Width: 640, Height: 360,
	}
	defaultSVG, err := renderSVG(cfg)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Orientation = OrientationVertical
	explicitSVG, err := renderSVG(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if defaultSVG != explicitSVG {
		t.Fatal("explicit vertical orientation changed zero-value vertical output")
	}
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(defaultSVG))); got != "03414abfc933c86901b6e6050a66f1d4fc9fed57a2f410e8765938dfe8c79405" {
		t.Fatalf("vertical SVG SHA-256 = %s, want pre-extension bytes", got)
	}
}

func TestBarValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{"label", func(cfg *Config) { cfg.Label = " " }, "label is required"},
		{"labels", func(cfg *Config) { cfg.Labels = nil }, "at least one label"},
		{"empty category", func(cfg *Config) { cfg.Labels[0] = " " }, "label 1 cannot be empty"},
		{"duplicate category", func(cfg *Config) { cfg.Labels[1] = cfg.Labels[0] }, `label "UN" is duplicated`},
		{"series", func(cfg *Config) { cfg.Series = nil }, "at least one series"},
		{"series name", func(cfg *Config) { cfg.Series[0].Name = " " }, "series 1 needs a name"},
		{"misaligned series", func(cfg *Config) { cfg.Series[0].Values = cfg.Series[0].Values[:1] }, "has 1 values; need 7"},
		{"nonfinite value", func(cfg *Config) { cfg.Series[0].Values[0] = math.NaN() }, `series "2011" value 0 must be finite`},
		{"orientation", func(cfg *Config) { cfg.Orientation = Orientation("diagonal") }, `orientation "diagonal" is unsupported`},
		{"padding", func(cfg *Config) { cfg.Padding.Left = -1 }, "padding cannot be negative"},
		{"width", func(cfg *Config) { cfg.Width = -1 }, "width cannot be negative"},
		{"height", func(cfg *Config) { cfg.Height = -1 }, "height cannot be negative"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			cfg := horizontalWorldPopulationConfig()
			cfg.Labels = append([]string(nil), cfg.Labels...)
			cfg.Series = append([]Series(nil), cfg.Series...)
			for index := range cfg.Series {
				cfg.Series[index].Values = append([]float64(nil), cfg.Series[index].Values...)
			}
			test.edit(&cfg)
			if err := cfg.validate(); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validate() error = %v, want %q", err, test.want)
			}
		})
	}
}
