package heatmap

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"

	"github.com/a-h/templ"
	chartassets "github.com/araihu/goshtoso-charts/assets"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func basicConfig() Config {
	return Config{
		Label: "Basic heat map", Title: "Heat Map Chart",
		XAxis: Axis{Title: "X-Axis", Labels: []string{"0", "1", "2", "3", "4"}},
		YAxis: Axis{Title: "Y-Axis", Labels: []string{"0", "1", "2", "3", "4"}},
		Rows: [][]float64{
			{4.4, 4.9, 7.0, 7.5, 4.3},
			{2.6, 5.9, 9.0, 6.4, 2.3},
			{3.3, 6.4, 7.0, 4.9, 3.2},
			{1.9, 6.0, 9.0, 5.9, 2.6},
			{4.4, 5.9, 7.0, 6.4, 4.6},
		},
		ValueRange: ValueRange{Min: 1.9, Max: 9},
	}
}

func TestHeatMapRendersUpstreamGeometryAccessibleDataAndWrapper(t *testing.T) {
	t.Parallel()
	cfg := basicConfig()
	cfg.Caption = "Values across the X and Y categories."
	cfg.Style = charttheme.Style{Palette: charttheme.PaletteAraiHu, Class: "mx-auto"}
	cfg.RootAttrs = templ.Attributes{"id": "basic-heat-map", "data-purpose": "comparison"}
	cfg.Controls = chartcontrol.Options{Fullscreen: true}
	cfg.Export = &chartcontrol.ExportOptions{Filename: "basic-heat-map"}

	instance := HeatMap(cfg)
	if instance.Kind() != chartcomponents.KindHeatMapChart {
		t.Fatalf("Kind() = %q, want %q", instance.Kind(), chartcomponents.KindHeatMapChart)
	}
	markup := render(t, instance)
	for _, want := range []string{
		`viewBox="0 0 600 400"`, "Heat Map Chart", "X-Axis", "Y-Axis",
		`aria-label="Basic heat map"`, `id="basic-heat-map"`, `data-purpose="comparison"`,
		"Values across the X and Y categories.", "Exact values", "4.4", "1.9", "9",
		"var(--color-chart-scale-low)", "var(--color-chart-scale-mid)", "var(--color-chart-scale-high)",
		"var(--color-chart-surface)", "var(--color-chart-text)", "var(--font-paragraph), sans-serif",
		"goshtoso-charts-heatmap__viewport", "min-width: 36rem", "goshtoso-charts-palette-araihu mx-auto",
		"goshtoso-charts-expanded .goshtoso-charts-heatmap__viewport", "aspect-ratio: 3 / 2",
		`data-goshtoso-chart-expand`, `-chart-expand-export"`, `>SVG</button>`, `>PNG</button>`,
		`-fullscreen-action`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	if !strings.Contains(markup, `src="`+chartassets.ControlRuntimeURL+`"`) {
		t.Error("heat map missing shared controls runtime")
	}
	if got := strings.Count(markup, `<path d=`); got < 26 {
		t.Errorf("SVG path count = %d, want background plus 25 cells and axes", got)
	}
}

func TestHeatMapSupportsIndexedCellsCustomStopsAndReverse(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Label: "Custom heat map",
		XAxis: Axis{Labels: []string{"A", "B"}}, YAxis: Axis{Labels: []string{"C", "D"}},
		Cells:      []Cell{{X: 1, Y: 1, Value: 10}, {X: 0, Y: 0, Value: 0}, {X: 1, Y: 0, Value: 5}, {X: 0, Y: 1, Value: 7.5}},
		ValueRange: ValueRange{Min: 0, Max: 10},
		Gradient: Gradient{Reverse: true, Stops: []GradientStop{
			{At: 0, Color: "#001122", Class: "cold-stop"},
			{At: 0.5, Color: "#778899", Class: "middle-stop"},
			{At: 1, Color: "#ffeecc", Class: "warm-stop"},
		}},
	}
	markup := render(t, HeatMap(cfg))
	for _, want := range []string{
		`data-heatmap-reverse="true"`, "#001122", "#778899", "#ffeecc",
		"cold-stop", "middle-stop", "warm-stop", "0 · warm", "10 · cold",
		">C</th>", ">A</td>", ">0</td>", ">D</th>", ">B</td>", ">10</td>",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("custom markup missing %q", want)
		}
	}
}

func TestHeatMapValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{"label", func(cfg *Config) { cfg.Label = "" }, "label is required"},
		{"x labels", func(cfg *Config) { cfg.XAxis.Labels = nil }, "x and y labels are required"},
		{"empty label", func(cfg *Config) { cfg.XAxis.Labels[0] = "" }, "x label 0 is empty"},
		{"negative width", func(cfg *Config) { cfg.Width = -1 }, "dimensions cannot be negative"},
		{"range order", func(cfg *Config) { cfg.ValueRange.Max = cfg.ValueRange.Min }, "min less than max"},
		{"range finite", func(cfg *Config) { cfg.ValueRange.Max = math.Inf(1) }, "finite min and max"},
		{"missing data", func(cfg *Config) { cfg.Rows = nil }, "rows or cells are required"},
		{"mixed data", func(cfg *Config) { cfg.Cells = []Cell{{}} }, "rows and cells cannot be combined"},
		{"row count", func(cfg *Config) { cfg.Rows = cfg.Rows[:4] }, "has 4 rows; need 5 y labels"},
		{"column count", func(cfg *Config) { cfg.Rows[0] = cfg.Rows[0][:4] }, "row 0 has 4 values; need 5 x labels"},
		{"value finite", func(cfg *Config) { cfg.Rows[0][0] = math.NaN() }, "value at (0, 0) must be finite"},
		{"value range", func(cfg *Config) { cfg.Rows[0][0] = 10 }, "value at (0, 0) is outside"},
		{"root attr", func(cfg *Config) { cfg.RootAttrs = templ.Attributes{"role": "presentation"} }, `root attribute "role" is reserved`},
		{"one stop", func(cfg *Config) { cfg.Gradient.Stops = []GradientStop{{At: 0, Color: "red"}} }, "at least two stops"},
		{"stop position", func(cfg *Config) { cfg.Gradient.Stops = []GradientStop{{At: -1, Color: "red"}, {At: 1, Color: "blue"}} }, "position must be finite and between"},
		{"empty stop", func(cfg *Config) { cfg.Gradient.Stops = []GradientStop{{At: 0}, {At: 1, Color: "blue"}} }, "needs a color or class"},
		{"unsafe color", func(cfg *Config) {
			cfg.Gradient.Stops = []GradientStop{{At: 0, Color: "red;display:none"}, {At: 1, Color: "blue"}}
		}, "color is unsafe"},
		{"stop order", func(cfg *Config) { cfg.Gradient.Stops = []GradientStop{{At: 0, Color: "red"}, {At: 0, Color: "blue"}} }, "strictly increasing order"},
		{"stop bounds", func(cfg *Config) { cfg.Gradient.Stops = []GradientStop{{At: .1, Color: "red"}, {At: 1, Color: "blue"}} }, "start at 0 and end at 1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			cfg := basicConfig()
			test.edit(&cfg)
			var output bytes.Buffer
			err := HeatMap(cfg).Render(context.Background(), &output)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Render() error = %v, want substring %q", err, test.want)
			}
			if output.Len() != 0 {
				t.Fatalf("invalid render wrote %d bytes", output.Len())
			}
		})
	}
}

func TestHeatMapIndexedCellValidation(t *testing.T) {
	t.Parallel()
	base := Config{
		Label: "Indexed", XAxis: Axis{Labels: []string{"A", "B"}}, YAxis: Axis{Labels: []string{"C"}},
		Cells: []Cell{{X: 0, Y: 0, Value: 1}, {X: 1, Y: 0, Value: 2}}, ValueRange: ValueRange{Min: 0, Max: 2},
	}
	for name, test := range map[string]struct {
		edit func(*Config)
		want string
	}{
		"coverage":  {func(cfg *Config) { cfg.Cells = cfg.Cells[:1] }, "need 2 to cover"},
		"index":     {func(cfg *Config) { cfg.Cells[1].X = 2 }, "indexes are outside"},
		"duplicate": {func(cfg *Config) { cfg.Cells[1].X = 0 }, "duplicates indexes"},
	} {
		t.Run(name, func(t *testing.T) {
			cfg := base
			cfg.Cells = append([]Cell(nil), base.Cells...)
			test.edit(&cfg)
			var output bytes.Buffer
			err := HeatMap(cfg).Render(context.Background(), &output)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Render() error = %v, want %q", err, test.want)
			}
		})
	}
}

func render(t *testing.T, instance Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return output.String()
}
