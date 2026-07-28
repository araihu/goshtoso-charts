package interactive

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func TestBar3DRendersCategoricalSourceSemantics(t *testing.T) {
	t.Parallel()
	cfg := validBar3DConfig()
	cfg.Label = "basic bar3d example"
	cfg.Caption = "Exact categorical values."
	cfg.VisualRange = &Bar3DVisualRange{Min: 0, Max: 30, Calculable: Bool(true), Palette: Bar3DPaletteColdToWarm}
	cfg.Grid = Bar3DGridSize{Width: 200, Depth: 80}
	cfg.Options = ChartOptions{Title: &TitleOptions{Text: "basic bar3d example"}}
	cfg.Style = charttheme.Style{Class: "caller-class"}
	cfg.RootAttrs = templ.Attributes{"id": "weekly-hours"}

	instance := Bar3D(cfg)
	if instance.Kind() != chartcomponents.KindInteractiveBar3D {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	markup := renderBar3D(t, instance)
	for _, want := range []string{
		`goshtoso-charts-bar-3d caller-class`, `role="img"`, `aria-label="basic bar3d example"`, `id="weekly-hours"`,
		`style="width:100%;height:38rem;"`, `"type":"bar3D"`, `"coordinateSystem":"cartesian3D"`,
		`"xAxis3D":{"name":"Hour","type":"category","data":["12a","1a"]}`, `"yAxis3D":{"name":"Day","type":"category","data":["Saturday","Friday"]}`,
		`"calculable":true`, `"max":30`, `"range":[0,30]`,
		`"color":["#313695","#4575b4","#74add1","#abd9e9","#e0f3f8","#fee090","#fdae61","#f46d43","#d73027","#a50026"]`,
		`"grid3D":{"boxWidth":200,"boxDepth":80}`, `{"value":[0,0,5]}`, `{"value":[1,0,1]}`,
		`data-goshtoso-charts-bar3d-cold-to-warm="true"`, `Exact 3D bar values`, `2 cells across 1 series.`,
		`>12a</th>`, `>Saturday</td>`, `data-goshtoso-chart-expand`, `data-goshtoso-chart-export="png"`,
		`var resizeObserver = window.ResizeObserver ? new ResizeObserver`, `series.type === "bar3D"`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	for _, unwanted := range []string{"go-echarts", "Apache ECharts", "echarts-gl", "echarts.dispose"} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("rendered markup contains %q", unwanted)
		}
	}
}

func TestBar3DRendersRotationShadingAndCallerPaint(t *testing.T) {
	t.Parallel()
	cfg := validBar3DConfig()
	cfg.Grid = Bar3DGridSize{Width: 160, Depth: 80}
	cfg.View = &Bar3DView{AutoRotate: Bool(true), AutoRotateSpeed: 200}
	cfg.Series[0].Options = Bar3DSeriesOptions{Shading: Bar3DShadingLambert, Class: "series-class"}
	cfg.Series[0].Cells[0].Class = "cold-cell"
	cfg.Series[0].Cells[1].Color = "#a50026"

	markup := renderBar3D(t, Bar3D(cfg))
	for _, want := range []string{
		`"grid3D":{"boxWidth":160,"boxDepth":80,"viewControl":{"autoRotate":true,"autoRotateSpeed":200}}`,
		`"shading":"lambert"`, `"className":"cold-cell"`, `"sourceColor":"#a50026"`,
		`data-goshtoso-charts-bar3d-paints="[{&#34;class&#34;:&#34;series-class&#34;}]"`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestBar3DSerializesCallerColorAndClassPaths(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		mutate func(*Bar3DConfig)
		want   string
	}{
		"series color": {func(cfg *Bar3DConfig) { cfg.Series[0].Options.Color = "#4575b4" }, `data-goshtoso-charts-bar3d-paints="[{&#34;color&#34;:&#34;#4575b4&#34;}]"`},
		"series class": {func(cfg *Bar3DConfig) { cfg.Series[0].Options.Class = "cold-series" }, `data-goshtoso-charts-bar3d-paints="[{&#34;class&#34;:&#34;cold-series&#34;}]"`},
		"cell color":   {func(cfg *Bar3DConfig) { cfg.Series[0].Cells[0].Color = "#d73027" }, `"sourceColor":"#d73027"`},
		"cell class":   {func(cfg *Bar3DConfig) { cfg.Series[0].Cells[0].Class = "warm-cell" }, `"className":"warm-cell"`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := validBar3DConfig()
			test.mutate(&cfg)
			if markup := renderBar3D(t, Bar3D(cfg)); !strings.Contains(markup, test.want) {
				t.Errorf("rendered markup missing %q", test.want)
			}
		})
	}
}

func TestBar3DRejectsInvalidDataAndConflicts(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		mutate func(*Bar3DConfig)
		want   string
	}{
		"label":        {func(cfg *Bar3DConfig) { cfg.Label = " " }, "bar 3D chart label is required"},
		"axis name":    {func(cfg *Bar3DConfig) { cfg.Axes.X.Name = "" }, "bar 3D chart x axis name is required"},
		"x categories": {func(cfg *Bar3DConfig) { cfg.Axes.X.Categories = nil }, "bar 3D chart x axis categories are required"},
		"category":     {func(cfg *Bar3DConfig) { cfg.Axes.X.Categories[0] = "" }, "bar 3D chart x axis category 0 is required"},
		"duplicate category": {func(cfg *Bar3DConfig) {
			cfg.Axes.X.Categories[1] = cfg.Axes.X.Categories[0]
		}, `bar 3D chart x axis category "12a" is duplicated`},
		"series":      {func(cfg *Bar3DConfig) { cfg.Series = nil }, "bar 3D chart series is required"},
		"series name": {func(cfg *Bar3DConfig) { cfg.Series[0].Name = "" }, "bar 3D chart series 0 name is required"},
		"cells":       {func(cfg *Bar3DConfig) { cfg.Series[0].Cells = nil }, `bar 3D chart series "bar3d" cells are required`},
		"x index":     {func(cfg *Bar3DConfig) { cfg.Series[0].Cells[0].XIndex = 2 }, `bar 3D chart series "bar3d" cell 0 x index 2 is out of range`},
		"y index":     {func(cfg *Bar3DConfig) { cfg.Series[0].Cells[0].YIndex = -1 }, `bar 3D chart series "bar3d" cell 0 y index -1 is out of range`},
		"duplicate coordinate": {func(cfg *Bar3DConfig) {
			cfg.Series[0].Cells[1].XIndex, cfg.Series[0].Cells[1].YIndex = 0, 0
		}, `bar 3D chart series "bar3d" coordinate [0,0] is duplicated`},
		"value": {func(cfg *Bar3DConfig) {
			cfg.Series[0].Cells[0].Value = math.NaN()
		}, `bar 3D chart series "bar3d" cell [0,0] value must be finite`},
		"cell paint": {func(cfg *Bar3DConfig) {
			cfg.Series[0].Cells[0].Color, cfg.Series[0].Cells[0].Class = "red", "danger"
		}, `bar 3D chart series "bar3d" cell [0,0] color and class are mutually exclusive`},
		"series paint": {func(cfg *Bar3DConfig) {
			cfg.Series[0].Options.Color, cfg.Series[0].Options.Class = "red", "danger"
		}, `bar 3D chart series "bar3d" color and class are mutually exclusive`},
		"shading": {func(cfg *Bar3DConfig) {
			cfg.Series[0].Options.Shading = "glossy"
		}, `bar 3D chart series "bar3d" shading "glossy" is not supported`},
		"grid partial": {func(cfg *Bar3DConfig) {
			cfg.Grid = Bar3DGridSize{Width: 160}
		}, "bar 3D chart grid width and depth must be set together"},
		"grid negative": {func(cfg *Bar3DConfig) {
			cfg.Grid = Bar3DGridSize{Width: -1, Depth: 80}
		}, "bar 3D chart grid sizes must be finite and positive when set"},
		"speed": {func(cfg *Bar3DConfig) {
			cfg.View = &Bar3DView{AutoRotate: Bool(true), AutoRotateSpeed: -1}
		}, "bar 3D chart auto-rotate speed must be finite and positive when set"},
		"speed rotation": {func(cfg *Bar3DConfig) {
			cfg.View = &Bar3DView{AutoRotate: Bool(false), AutoRotateSpeed: 200}
		}, "bar 3D chart auto-rotate speed requires auto rotation"},
		"range": {func(cfg *Bar3DConfig) {
			cfg.VisualRange = &Bar3DVisualRange{Min: 2, Max: 1, Colors: []string{"a", "b"}}
		}, "bar 3D chart visual range minimum must not exceed maximum"},
		"cold custom": {func(cfg *Bar3DConfig) {
			cfg.VisualRange = &Bar3DVisualRange{Palette: Bar3DPaletteColdToWarm, Colors: []string{"a", "b"}}
		}, "bar 3D chart cold-to-warm palette conflicts with custom colors"},
		"outside range": {func(cfg *Bar3DConfig) {
			cfg.VisualRange = &Bar3DVisualRange{Min: 0, Max: 4, Colors: []string{"a", "b"}}
		}, `bar 3D chart series "bar3d" cell [0,0] value is outside visual range`},
		"range cell paint": {func(cfg *Bar3DConfig) {
			cfg.Series[0].Cells[0].Color = "red"
			cfg.VisualRange = &Bar3DVisualRange{Min: 0, Max: 30, Colors: []string{"a", "b"}}
		}, "bar 3D chart visual range conflicts with cell paint"},
		"legend":    {func(cfg *Bar3DConfig) { cfg.Options.Legend = &LegendOptions{} }, "bar 3D chart legend is not supported"},
		"2D axis":   {func(cfg *Bar3DConfig) { cfg.Options.XAxis = &AxisOptions{} }, "bar 3D chart Cartesian axes must use Axes"},
		"root attr": {func(cfg *Bar3DConfig) { cfg.RootAttrs = templ.Attributes{"role": "group"} }, `bar 3D chart root attribute "role" is reserved`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := validBar3DConfig()
			test.mutate(&cfg)
			var output bytes.Buffer
			err := Bar3D(cfg).Render(context.Background(), &output)
			if err == nil || err.Error() != test.want {
				t.Fatalf("Render() error = %v, want %q", err, test.want)
			}
		})
	}
}

func validBar3DConfig() Bar3DConfig {
	return Bar3DConfig{
		Label: "bar3d",
		Axes: Bar3DAxes{
			X: Bar3DCategoricalAxis{Name: "Hour", Categories: []string{"12a", "1a"}},
			Y: Bar3DCategoricalAxis{Name: "Day", Categories: []string{"Saturday", "Friday"}},
		},
		Series: []Bar3DSeries{{Name: "bar3d", Cells: []Bar3DCell{{XIndex: 0, YIndex: 0, Value: 5}, {XIndex: 1, YIndex: 0, Value: 1}}}},
	}
}

func renderBar3D(t *testing.T, instance Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	return output.String()
}
