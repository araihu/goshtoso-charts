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
	chart "github.com/go-analyze/charts"
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
		`data-goshtoso-chart-expand`, `-chart-expand-export"`, `<span class="block">Download SVG</span>`, `<span class="block">Download PNG</span>`,
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

func TestHeatMapMapsFinitePresentationAPI(t *testing.T) {
	t.Parallel()
	cfg := basicConfig()
	cfg.TitleOptions = TitleOptions{
		Subtext: "Same pinned matrix", Placement: PlacementEnd,
		FontSize: 18, SubtextFontSize: 11, BorderWidth: 2,
	}
	cfg.XAxis.TitleFontSize = 12
	cfg.XAxis.LabelFontSize = 8
	cfg.XAxis.LabelRotation = 30
	cfg.XAxis.LabelCount = 4
	cfg.XAxis.LabelCountAdjustment = -1
	cfg.YAxis.TitleFontSize = 13
	cfg.YAxis.LabelFontSize = 9
	cfg.YAxis.LabelRotation = -15
	cfg.YAxis.LabelCount = 3
	cfg.YAxis.LabelCountAdjustment = 1
	cfg.Padding = Padding{Top: 18, Right: 24, Bottom: 30, Left: 36}
	cfg.ValueLabels = ValueLabelOptions{
		Show: true, Format: ValueFormatInteger, FontSize: 9,
		Distance: 2, Offset: Offset{Left: 1, Top: -1},
	}

	options := heatMapOptions(cfg, cfg.Rows)
	if options.Title.Text != "Heat Map Chart" || options.Title.Subtext != "Same pinned matrix" ||
		options.Title.Offset != chart.OffsetRight || options.Title.FontStyle.FontSize != 18 ||
		options.Title.SubtextFontStyle.FontSize != 11 || options.Title.BorderWidth != 2 {
		t.Fatalf("title options = %#v", options.Title)
	}
	if options.Padding != chart.NewBox(36, 18, 24, 30) {
		t.Fatalf("padding = %#v", options.Padding)
	}
	if options.XAxis.TitleFontStyle.FontSize != 12 || options.XAxis.LabelFontStyle.FontSize != 8 ||
		math.Abs(options.XAxis.LabelRotation-math.Pi/6) > .0001 || options.XAxis.LabelCount != 4 ||
		options.XAxis.LabelCountAdjustment != -1 {
		t.Fatalf("X axis options = %#v", options.XAxis)
	}
	if options.YAxis.TitleFontStyle.FontSize != 13 || options.YAxis.LabelFontStyle.FontSize != 9 ||
		math.Abs(options.YAxis.LabelRotation+math.Pi/12) > .0001 || options.YAxis.LabelCount != 3 ||
		options.YAxis.LabelCountAdjustment != 1 {
		t.Fatalf("Y axis options = %#v", options.YAxis)
	}
	if options.ValuesLabel.Show == nil || !*options.ValuesLabel.Show || options.ValuesLabel.FontStyle.FontSize != 9 ||
		options.ValuesLabel.Distance != 2 || options.ValuesLabel.Offset != (chart.OffsetInt{Left: 1, Top: -1}) ||
		options.ValuesLabel.ValueFormatter == nil || options.ValuesLabel.ValueFormatter(4.9) != "5" {
		t.Fatalf("value-label options = %#v", options.ValuesLabel)
	}

	markup := render(t, HeatMap(cfg))
	for _, want := range []string{"Same pinned matrix", ">4</text>", ">5</text>", "var(--color-chart-text)"} {
		if !strings.Contains(markup, want) {
			t.Errorf("finite presentation markup missing %q", want)
		}
	}
}

func TestHeatMapMapsEveryTitlePlacement(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		placement Placement
		want      chart.OffsetStr
	}{
		{placement: PlacementStart, want: chart.OffsetLeft},
		{placement: PlacementCenter, want: chart.OffsetCenter},
		{placement: PlacementEnd, want: chart.OffsetRight},
		{placement: PlacementDefault, want: chart.OffsetCenter},
	} {
		if got := titleOffset(test.placement); got != test.want {
			t.Errorf("titleOffset(%q) = %#v, want %#v", test.placement, got, test.want)
		}
	}
}

func TestHeatMapTitleValidationOrderIsDeterministic(t *testing.T) {
	t.Parallel()
	err := (TitleOptions{FontSize: -1, SubtextFontSize: -1, BorderWidth: -1}).validate()
	if err == nil || err.Error() != "heat map title font size must be finite and non-negative" {
		t.Fatalf("TitleOptions.validate() error = %v", err)
	}
}

func TestHeatMapTitleCanBeHiddenWithoutLosingAccessibleName(t *testing.T) {
	t.Parallel()
	cfg := basicConfig()
	cfg.TitleOptions.Hidden = true
	markup := render(t, HeatMap(cfg))
	if strings.Contains(markup, ">Heat Map Chart</text>") {
		t.Error("hidden visible title was rendered")
	}
	if !strings.Contains(markup, `aria-label="Basic heat map"`) {
		t.Error("hidden visible title removed figure accessible name")
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
		{"x title font", func(cfg *Config) { cfg.XAxis.TitleFontSize = math.NaN() }, "x axis title font size"},
		{"y label font", func(cfg *Config) { cfg.YAxis.LabelFontSize = -1 }, "y axis label font size"},
		{"axis rotation", func(cfg *Config) { cfg.XAxis.LabelRotation = 361 }, "x axis label rotation"},
		{"axis label count", func(cfg *Config) { cfg.YAxis.LabelCount = -1 }, "y axis label count"},
		{"title placement", func(cfg *Config) { cfg.TitleOptions.Placement = "middle" }, "title placement"},
		{"title font", func(cfg *Config) { cfg.TitleOptions.FontSize = math.Inf(1) }, "title font size"},
		{"title border", func(cfg *Config) { cfg.TitleOptions.BorderWidth = -1 }, "title border width"},
		{"value label format", func(cfg *Config) { cfg.ValueLabels = ValueLabelOptions{Show: true, Format: "callback"} }, "value label format"},
		{"value label decimals", func(cfg *Config) { cfg.ValueLabels = ValueLabelOptions{Show: true, Decimals: 16} }, "value label decimals"},
		{"value label font", func(cfg *Config) { cfg.ValueLabels = ValueLabelOptions{Show: true, FontSize: math.NaN()} }, "value label font size"},
		{"value label distance", func(cfg *Config) { cfg.ValueLabels = ValueLabelOptions{Show: true, Distance: -1} }, "value label distance"},
		{"value labels hidden", func(cfg *Config) { cfg.ValueLabels = ValueLabelOptions{Format: ValueFormatExact} }, "requires labels to be shown"},
		{"negative padding", func(cfg *Config) { cfg.Padding.Left = -1 }, "padding cannot be negative"},
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
