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
	chart "github.com/go-analyze/charts"
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

func TestBarPropagatesSharedWrapperLifecycle(t *testing.T) {
	t.Parallel()
	base := Config{
		Label: "Deployments", Labels: []string{"Development"},
		Series: []Series{{Name: "Successful", Values: []float64{14}}},
	}
	base.Controls.Mode = chartcontrol.WrapperModeHidden
	var hidden bytes.Buffer
	if err := Bar(base).Render(context.Background(), &hidden); err != nil {
		t.Fatalf("hidden Render() error = %v", err)
	}
	if !strings.Contains(hidden.String(), `data-goshtoso-chart-wrapper-mode="hidden"`) ||
		!strings.Contains(hidden.String(), `hidden inert aria-hidden="true"`) {
		t.Fatalf("static chart did not propagate hidden wrapper mode")
	}

	base.Controls.Mode = chartcontrol.WrapperModeOmitted
	var omitted bytes.Buffer
	if err := Bar(base).Render(context.Background(), &omitted); err != nil {
		t.Fatalf("omitted Render() error = %v", err)
	}
	if strings.Contains(omitted.String(), "data-goshtoso-chart-wrapper") ||
		!strings.Contains(omitted.String(), `class="goshtoso-charts-bar`) {
		t.Fatalf("static chart did not propagate omitted wrapper mode")
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
	cfg.Controls = chartcontrol.Options{Fullscreen: true}
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
		`data-goshtoso-chart-expand`, `-chart-expand-export"`,
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

func TestBarReferencesRenderMarksAndAccessibleEvidence(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Label: "Monthly references", Labels: []string{"Jan", "Feb", "Mar"}, Width: 600, Height: 400,
		Series: []Series{{Name: "Rainfall", Values: []float64{2.4, 8.6, 3.1}, References: References{Average: true, Minimum: true, Maximum: true, Format: ValueFormatHumanized, PointSize: 20, Style: ReferenceStyle{Color: "#14532d", Class: "rainfall-references"}}}},
	}
	options := barOptions(cfg)
	series := options.SeriesList[0]
	if len(series.MarkLine.Lines) != 1 || series.MarkLine.Lines[0].Type != chart.SeriesMarkTypeAverage {
		t.Fatalf("mark lines = %#v, want average", series.MarkLine.Lines)
	}
	if got := series.MarkPoint.Points; len(got) != 2 || got[0].Type != chart.SeriesMarkTypeMax || got[1].Type != chart.SeriesMarkTypeMin || series.MarkPoint.SymbolSize != 20 {
		t.Fatalf("mark points = %#v size %d, want max/min size 20", got, series.MarkPoint.SymbolSize)
	}
	if got := series.MarkLine.ValueFormatter(4.6); got != "5" || series.MarkPoint.ValueFormatter(4.6) != "5" {
		t.Fatalf("humanized formatter = %q, want 5", got)
	}
	var output bytes.Buffer
	if err := Bar(cfg).Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	markup := output.String()
	for _, want := range []string{
		`viewBox="0 0 600 400"`, "Exact values and reference annotations", "Monthly references exact values and reference annotations", "Monthly references computed reference annotations",
		">Jan</th><td class=\"px-3 py-2 tabular-nums\">2</td>", ">Feb</th><td class=\"px-3 py-2 tabular-nums\">9</td>",
		">5</td><td class=\"px-3 py-2 tabular-nums\">9 at Feb</td><td class=\"px-3 py-2 tabular-nums\">2 at Jan</td>", "rainfall-references", "#14532d",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("reference markup missing %q", want)
		}
	}
}

func TestBarMapsRendererNeutralUpstreamGeometryLabelsLegendAxisAndMarks(t *testing.T) {
	t.Parallel()
	gap := 0.02
	cfg := Config{
		Label: "Stacked rainfall and evaporation", Labels: []string{"Jan", "Feb", "Mar"}, Stacked: true,
		Series: []Series{
			{
				Name: "Rainfall", Values: []float64{2.0, 4.9, 7.0},
				Labels:     DataLabelOptions{Show: true, Format: ValueFormatHumanized},
				References: References{MaximumLine: true, Format: ValueFormatHumanized},
			},
			{
				Name: "Evaporation", Values: []float64{2.6, 5.9, 9.0},
				Labels:     DataLabelOptions{Show: true, Format: ValueFormatHumanized},
				References: References{GlobalMaximum: true, PointPrefix: "Sum:", PointSize: 32, Format: ValueFormatHumanized},
			},
		},
		Geometry:      GeometryOptions{ThicknessRatio: 0.15, GapRatio: &gap, RoundedCaps: true},
		LabelPosition: DataLabelPositionStart,
		Legend:        LegendOptions{Placement: LegendPlacementEnd, Overlay: true},
		ValueAxis:     ValueAxisOptions{Hidden: true},
	}

	options := barOptions(cfg)
	if options.BarSize != 0.15 || options.BarMargin == nil || *options.BarMargin != gap || options.RoundedBarCaps == nil || !*options.RoundedBarCaps {
		t.Fatalf("geometry mapping = size %v gap %v rounded %v", options.BarSize, options.BarMargin, options.RoundedBarCaps)
	}
	if options.SeriesLabelPosition != chart.PositionBottom {
		t.Fatalf("vertical start label position = %q, want %q", options.SeriesLabelPosition, chart.PositionBottom)
	}
	if options.Legend.Offset != chart.OffsetRight || options.Legend.OverlayChart == nil || !*options.Legend.OverlayChart {
		t.Fatalf("legend mapping = %#v", options.Legend)
	}
	if len(options.ValueAxis) != 1 || options.ValueAxis[0].Show == nil || *options.ValueAxis[0].Show {
		t.Fatalf("value-axis mapping = %#v", options.ValueAxis)
	}
	for index, series := range options.SeriesList {
		if series.Label.Show == nil || !*series.Label.Show || series.Label.ValueFormatter == nil || series.Label.ValueFormatter(4.6) != "5" {
			t.Errorf("series %d labels = %#v", index, series.Label)
		}
	}
	if got := options.SeriesList[0].MarkLine.Lines; len(got) != 1 || got[0].Type != chart.SeriesMarkTypeMax || got[0].Global {
		t.Fatalf("maximum line = %#v", got)
	}
	if got := options.SeriesList[1].MarkPoint.Points; len(got) != 1 || got[0].Type != chart.SeriesMarkTypeMax || !got[0].Global {
		t.Fatalf("global maximum point = %#v", got)
	}
	if got := options.SeriesList[1].MarkPoint.ValueFormatter(14.9); got != "Sum:15" {
		t.Fatalf("global maximum formatter = %q, want Sum:15", got)
	}
}

func TestStackedGlobalMaximumKeepsExactAccessibleEvidence(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Label: "Stacked rainfall and evaporation", Labels: []string{"Jan", "Feb", "Mar"}, Stacked: true,
		Series: []Series{
			{Name: "Rainfall", Values: []float64{2, 4.9, 7}, References: References{MaximumLine: true, Format: ValueFormatHumanized}},
			{Name: "Evaporation", Values: []float64{2.6, 5.9, 9}, References: References{GlobalMaximum: true, PointPrefix: "Sum:", Format: ValueFormatHumanized}},
		},
	}
	var output bytes.Buffer
	if err := Bar(cfg).Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	markup := output.String()
	for _, want := range []string{
		"Exact values and reference annotations",
		`aria-label="Stacked rainfall and evaporation exact values and reference annotations"`,
		`aria-label="Stacked rainfall and evaporation maximum-line annotations"`,
		`>Rainfall maximum line</th><td class="px-3 py-2 tabular-nums">7 at Mar</td>`,
		`aria-label="Stacked rainfall and evaporation stacked reference annotations"`,
		">Maximum stack total</th>", ">16 at Mar</td>",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("stacked reference evidence missing %q", want)
		}
	}
}

func TestHorizontalBarMapsSemanticStartLabelPosition(t *testing.T) {
	t.Parallel()
	cfg := horizontalWorldPopulationConfig()
	cfg.LabelPosition = DataLabelPositionStart
	cfg.Series[0].Labels.Show = true
	options := barOptions(cfg)
	if options.SeriesLabelPosition != chart.PositionLeft {
		t.Fatalf("horizontal start label position = %q, want %q", options.SeriesLabelPosition, chart.PositionLeft)
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
		{"reference format", func(cfg *Config) { cfg.Series[0].References.Format = ValueFormat("scientific") }, `reference format "scientific" is unsupported`},
		{"reference point size", func(cfg *Config) { cfg.Series[0].References.PointSize = -1 }, "reference point size cannot be negative"},
		{"reference color", func(cfg *Config) { cfg.Series[0].References.Style.Color = "url(https://example.test)" }, "reference color is unsafe"},
		{"reference class", func(cfg *Config) { cfg.Series[0].References.Style.Class = "bad;class" }, "reference class is unsafe"},
		{"geometry thickness finite", func(cfg *Config) { cfg.Geometry.ThicknessRatio = math.NaN() }, "thickness ratio must be finite"},
		{"geometry thickness range", func(cfg *Config) { cfg.Geometry.ThicknessRatio = 1.1 }, "thickness ratio must be between 0 and 1"},
		{"geometry gap finite", func(cfg *Config) { gap := math.Inf(1); cfg.Geometry.GapRatio = &gap }, "gap ratio must be finite"},
		{"geometry gap range", func(cfg *Config) { gap := -0.1; cfg.Geometry.GapRatio = &gap }, "gap ratio must be between 0 and 1"},
		{"label position", func(cfg *Config) { cfg.LabelPosition = DataLabelPosition("middle") }, `label position "middle" is unsupported`},
		{"legend placement", func(cfg *Config) { cfg.Legend.Placement = LegendPlacement("somewhere") }, `legend placement "somewhere" is unsupported`},
		{"data label format", func(cfg *Config) { cfg.Series[0].Labels.Format = ValueFormat("scientific") }, `data-label format "scientific" is unsupported`},
		{"global maximum requires stacking", func(cfg *Config) { cfg.Series[0].References.GlobalMaximum = true }, "global maximum reference requires stacked bars"},
		{"global maximum requires last series", func(cfg *Config) { cfg.Stacked = true; cfg.Series[0].References.GlobalMaximum = true }, "global maximum reference must be on the last series"},
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
