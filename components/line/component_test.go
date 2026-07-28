package line

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func dualAxisConfig() Config {
	return Config{
		Label:  "Dual Axis Line",
		Title:  Title{Text: "Dual Axis Line"},
		Labels: []string{"A", "B", "C", "D", "E", "F", "G"},
		Series: []Series{
			{Name: "Left Series", Values: []float64{120, 132, 101, 134, 90, 230, 210}},
			{Name: "Right Series", Values: []float64{820, 932, 901, 934, 1290, 1330, 1320}, YAxisIndex: 1},
		},
		YAxes:  []Axis{{}, {}},
		Width:  600,
		Height: 400,
	}
}

func TestLineDefaultSVGCompatibilityHash(t *testing.T) {
	t.Parallel()
	svg, err := renderSVG(Config{
		Label:  "Compatibility",
		Labels: []string{"Mon", "Tue", "Wed"},
		Series: []Series{{Name: "Value", Values: []float64{12, 18, 15}}},
	})
	if err != nil {
		t.Fatalf("renderSVG() error = %v", err)
	}
	got := fmt.Sprintf("%x", sha256.Sum256([]byte(svg)))
	const want = "da1c17f69cfc37e474c3d678f756d67782c2942e3901f8203c59c6e5d6e39063"
	if got != want {
		t.Fatalf("default SVG SHA-256 = %s, want %s", got, want)
	}
}

func TestLineSupportsSharedControlsAndExport(t *testing.T) {
	t.Parallel()
	instance := Line(Config{
		Label:    "Latency",
		Labels:   []string{"Mon"},
		Series:   []Series{{Name: "p95", Values: []float64{12}}},
		Controls: chartcontrol.Options{Fullscreen: true, Collapsible: true},
		Export:   &chartcontrol.ExportOptions{Filename: "latency"},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`data-goshtoso-chart-control="fullscreen"`,
		`data-goshtoso-chart-control="collapse"`,
		`data-goshtoso-chart-expand`, `data-goshtoso-chart-export-menu`,
		`>SVG</button>`, `>PNG</button>`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("markup missing %q", want)
		}
	}
}

func TestLineRendersSSRAccessibleSVG(t *testing.T) {
	t.Parallel()
	instance := Line(Config{
		Label:   "Weekly signups",
		Caption: "Seven-day trend",
		Labels:  []string{"Mon", "Tue", "Wed"},
		Series:  []Series{{Name: "Signups", Values: []float64{12, 18, 15}}},
		Style:   charttheme.Style{Palette: charttheme.PalettePastel, Colors: []string{"#123456"}, Class: "ring-2"},
	})
	if instance.Kind() != chartcomponents.KindLineChart {
		t.Fatalf("Kind() = %q, want %q", instance.Kind(), chartcomponents.KindLineChart)
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{"<figure class=\"goshtoso-charts-line goshtoso-charts-palette goshtoso-charts-palette-pastel ring-2\" role=\"img\" aria-label=\"Weekly signups\"", "<svg", "Seven-day trend", "#123456", "var(--color-chart-surface)", "var(--font-paragraph), sans-serif"} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q:\n%s", want, markup)
		}
	}
	if strings.Contains(markup, "echarts.init") {
		t.Errorf("SSR chart unexpectedly contains interactive renderer initialization: %s", markup)
	}
}

func TestLineEscapesProgrammaticSeriesColors(t *testing.T) {
	instance := Line(Config{
		Label:  "Safe chart",
		Labels: []string{"one"},
		Series: []Series{{Name: "value", Values: []float64{1}}},
		Style:  charttheme.Style{Colors: []string{`red" onload="alert(1)`}},
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

func TestLineRejectsMisalignedSeries(t *testing.T) {
	t.Parallel()
	_, err := renderSVG(Config{
		Label:  "Weekly signups",
		Labels: []string{"Mon", "Tue"},
		Series: []Series{{Name: "Signups", Values: []float64{12}}},
	})
	if err == nil || !strings.Contains(err.Error(), "has 1 values; need 2") {
		t.Fatalf("renderSVG() error = %v, want value alignment error", err)
	}
}

func TestLineMechanicallyMapsPinnedDualAxisExample(t *testing.T) {
	t.Parallel()
	cfg := dualAxisConfig()
	if cfg.width() != 600 || cfg.height() != 400 {
		t.Fatalf("dimensions = %dx%d, want 600x400", cfg.width(), cfg.height())
	}
	options := lineOptions(cfg)
	if options.Title.Text != "Dual Axis Line" {
		t.Fatalf("title = %q", options.Title.Text)
	}
	if !reflect.DeepEqual(options.XAxis.Labels, []string{"A", "B", "C", "D", "E", "F", "G"}) {
		t.Fatalf("labels = %#v", options.XAxis.Labels)
	}
	if !reflect.DeepEqual(options.Legend.SeriesNames, []string{"Left Series", "Right Series"}) {
		t.Fatalf("legend names = %#v", options.Legend.SeriesNames)
	}
	wantValues := [][]float64{
		{120, 132, 101, 134, 90, 230, 210},
		{820, 932, 901, 934, 1290, 1330, 1320},
	}
	if len(options.SeriesList) != 2 || len(options.YAxis) != 2 {
		t.Fatalf("series/axis count = %d/%d, want 2/2", len(options.SeriesList), len(options.YAxis))
	}
	for index := range options.SeriesList {
		if options.SeriesList[index].Name != options.Legend.SeriesNames[index] ||
			options.SeriesList[index].YAxisIndex != index ||
			!reflect.DeepEqual(options.SeriesList[index].Values, wantValues[index]) {
			t.Fatalf("series %d = %#v", index, options.SeriesList[index])
		}
	}
}

func TestLineDualAxisSVGHash(t *testing.T) {
	t.Parallel()
	svg, err := renderSVG(dualAxisConfig())
	if err != nil {
		t.Fatalf("renderSVG() error = %v", err)
	}
	got := fmt.Sprintf("%x", sha256.Sum256([]byte(svg)))
	const want = "191a64ba43b2267ade4622cd0bf79f5167e2452448ef0f21287d941c283461ed"
	if got != want {
		t.Fatalf("dual-axis SVG SHA-256 = %s, want %s", got, want)
	}
}

func TestLineDualAxisRendersThemeMatchedAxesAndExactMapping(t *testing.T) {
	t.Parallel()
	cfg := dualAxisConfig()
	cfg.Caption = "Two scales."
	cfg.Controls = chartcontrol.Options{Fullscreen: true, Collapsible: true}
	cfg.Export = &chartcontrol.ExportOptions{Filename: "dual-axis-line"}
	var output bytes.Buffer
	if err := Line(cfg).Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`aria-label="Dual Axis Line"`, `viewBox="0 0 600 400"`, "Dual Axis Line",
		"Left Series", "Right Series", "Left Y axis", "Right Y axis",
		"var(--color-chart-series-1)", "var(--color-chart-series-2)",
		`aria-label="Dual Axis Line exact series values and Y axis mapping"`,
		"Exact series values", "Two scales.",
		`data-goshtoso-chart-expand`, `data-goshtoso-chart-export-menu`,
		`>SVG</button>`, `>PNG</button>`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	for _, unwanted := range []string{"rgb(21,21,21)", "rgb(22,22,22)", "go-analyze"} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("rendered markup contains %q", unwanted)
		}
	}
	for _, row := range []string{
		"<th class=\"px-3 py-2 font-semibold\" scope=\"row\">A</th><td class=\"px-3 py-2\">Left Series</td><td class=\"px-3 py-2 \">Left Y axis</td><td class=\"px-3 py-2 tabular-nums\">120</td>",
		"<th class=\"px-3 py-2 font-semibold\" scope=\"row\">G</th><td class=\"px-3 py-2\">Right Series</td><td class=\"px-3 py-2 \">Right Y axis</td><td class=\"px-3 py-2 tabular-nums\">1320</td>",
	} {
		if !strings.Contains(markup, row) {
			t.Errorf("exact mapping row missing %q", row)
		}
	}
}

func TestLineCallerSeriesAndAxisPresentationOverrides(t *testing.T) {
	t.Parallel()
	cfg := dualAxisConfig()
	cfg.Series[0].Color = "#14532d"
	cfg.Series[1].Class = "caller-right-series"
	cfg.YAxes[0].Class = "caller-left-axis"
	cfg.YAxes[1].Color = "#7e22ce"
	var output bytes.Buffer
	if err := Line(cfg).Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{"#14532d", "#7e22ce", "caller-right-series", "caller-left-axis"} {
		if !strings.Contains(markup, want) {
			t.Errorf("caller override missing %q", want)
		}
	}
	if strings.Contains(markup, "rgb(21,21,21)") || strings.Contains(markup, "rgb(22,22,22)") {
		t.Fatal("axis placeholder leaked")
	}
}

func TestLineValidation(t *testing.T) {
	t.Parallel()
	minimum, maximum := 0.0, 1500.0
	tests := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{name: "label", edit: func(c *Config) { c.Label = "" }, want: "label is required"},
		{name: "labels", edit: func(c *Config) { c.Labels = nil }, want: "at least one label"},
		{name: "empty label", edit: func(c *Config) { c.Labels[0] = "" }, want: "label 1 cannot be empty"},
		{name: "duplicate label", edit: func(c *Config) { c.Labels[1] = "A" }, want: `label "A" is duplicated`},
		{name: "series", edit: func(c *Config) { c.Series = nil }, want: "at least one series"},
		{name: "series name", edit: func(c *Config) { c.Series[0].Name = "" }, want: "series 1 needs a name"},
		{name: "duplicate series", edit: func(c *Config) { c.Series[1].Name = "Left Series" }, want: `series name "Left Series" is duplicated`},
		{name: "length", edit: func(c *Config) { c.Series[0].Values = c.Series[0].Values[:6] }, want: "has 6 values; need 7"},
		{name: "nan", edit: func(c *Config) { c.Series[0].Values[0] = math.NaN() }, want: "value 0 must be finite"},
		{name: "infinite", edit: func(c *Config) { c.Series[0].Values[0] = math.Inf(1) }, want: "value 0 must be finite"},
		{name: "width", edit: func(c *Config) { c.Width = -1 }, want: "width cannot be negative"},
		{name: "height", edit: func(c *Config) { c.Height = -1 }, want: "height cannot be negative"},
		{name: "axis count", edit: func(c *Config) { c.YAxes = []Axis{{}, {}, {}} }, want: "at most two Y axes"},
		{name: "axis index negative", edit: func(c *Config) { c.Series[0].YAxisIndex = -1 }, want: "Y axis index -1 is out of bounds"},
		{name: "axis index high", edit: func(c *Config) { c.Series[1].YAxisIndex = 2 }, want: "Y axis index 2 is out of bounds"},
		{name: "unused axis", edit: func(c *Config) { c.Series[1].YAxisIndex = 0 }, want: "Y axis 1 has no assigned series"},
		{name: "unit nan", edit: func(c *Config) { c.YAxes[0].Unit = math.NaN() }, want: "unit must be finite and non-negative"},
		{name: "unit negative", edit: func(c *Config) { c.YAxes[0].Unit = -1 }, want: "unit must be finite and non-negative"},
		{name: "min nan", edit: func(c *Config) { value := math.NaN(); c.YAxes[0].Min = &value }, want: "minimum must be finite"},
		{name: "max infinite", edit: func(c *Config) { value := math.Inf(1); c.YAxes[0].Max = &value }, want: "maximum must be finite"},
		{name: "range", edit: func(c *Config) { c.YAxes[0].Min, c.YAxes[0].Max = &maximum, &minimum }, want: "minimum must be less than maximum"},
		{name: "below bound", edit: func(c *Config) { value := 121.0; c.YAxes[0].Min = &value }, want: "below Y axis 0 minimum"},
		{name: "above bound", edit: func(c *Config) { value := 1319.0; c.YAxes[1].Max = &value }, want: "above Y axis 1 maximum"},
		{name: "series conflict", edit: func(c *Config) { c.Series[0].Color, c.Series[0].Class = "red", "custom" }, want: "cannot set both color and class"},
		{name: "axis conflict", edit: func(c *Config) { c.YAxes[0].Color, c.YAxes[0].Class = "red", "custom" }, want: "cannot set both color and class"},
		{name: "series unsafe color", edit: func(c *Config) { c.Series[0].Color = "red;stroke:black" }, want: "series \"Left Series\" color is unsafe"},
		{name: "axis unsafe color", edit: func(c *Config) { c.YAxes[0].Color = "url(javascript:bad)" }, want: "Y axis 0 color is unsafe"},
		{name: "series unsafe class", edit: func(c *Config) { c.Series[0].Class = `x" onclick="bad` }, want: "series \"Left Series\" class is unsafe"},
		{name: "axis unsafe class", edit: func(c *Config) { c.YAxes[0].Class = "x;y" }, want: "Y axis 0 class is unsafe"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			cfg := dualAxisConfig()
			cfg.Labels = append([]string(nil), cfg.Labels...)
			cfg.Series = append([]Series(nil), cfg.Series...)
			for index := range cfg.Series {
				cfg.Series[index].Values = append([]float64(nil), cfg.Series[index].Values...)
			}
			cfg.YAxes = append([]Axis(nil), cfg.YAxes...)
			test.edit(&cfg)
			err := cfg.validate()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validate() error = %v, want %q", err, test.want)
			}
		})
	}
}
