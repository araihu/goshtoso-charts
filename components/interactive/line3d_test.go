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

func TestLine3DRendersTypedPathAndExactDataAccess(t *testing.T) {
	t.Parallel()
	cfg := validLine3DConfig()
	cfg.Style = charttheme.Style{Class: "caller-class"}
	cfg.RootAttrs = templ.Attributes{"id": "parametric-line"}
	instance := Line3D(cfg)
	if instance.Kind() != chartcomponents.KindInteractiveLine3D {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	markup := renderLine3D(t, instance)
	for _, want := range []string{
		`goshtoso-charts-line-3d caller-class`, `role="img"`, `aria-label="basic line3d example"`,
		`id="parametric-line"`, `style="width:100%;height:38rem;"`, `"type":"line3D"`,
		`"name":"line3D"`, `"value":[1.25,0,0]`, `"value":[1,0,1]`,
		`"calculable":true`, `"max":30`,
		`"color":["#313695","#4575b4","#74add1","#abd9e9","#e0f3f8","#fee090","#fdae61","#f46d43","#d73027","#a50026"]`,
		`data-goshtoso-charts-line3d-cold-to-warm="true"`, `Exact line data`, `2 ordered points.`,
		`Formula:`, `t = i / 1000`, `t domain [0, 0.001]`, `Download all exact points as CSV`,
		`download="basic-line3d-example-exact-data.csv"`, `series%2Cindex%2Cx%2Cy%2Cz`,
		`data-goshtoso-chart-expand`, `exportFromMenu($el, &#34;png&#34;)`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	if strings.Contains(markup, "<tbody") || strings.Count(markup, "<tr") != 0 {
		t.Error("line dense data injected table rows into initial DOM")
	}
	for _, unwanted := range []string{"go-echarts", "Apache ECharts", "echarts-gl", "raw option"} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("rendered markup contains %q", unwanted)
		}
	}
}

func TestLine3DSupportsViewSeriesAndPaletteOverrides(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		mutate func(*Line3DConfig)
		want   string
	}{
		"auto rotate": {func(cfg *Line3DConfig) {
			cfg.Grid.View = &Line3DView{AutoRotate: Bool(true), AutoRotateSpeed: 18}
		}, `"autoRotate":true`},
		"custom palette": {func(cfg *Line3DConfig) {
			cfg.VisualRange.Palette = Line3DPaletteCustom
			cfg.VisualRange.Colors = []string{"#111111", "#eeeeee"}
		}, `"color":["#111111","#eeeeee"]`},
		"series color": {func(cfg *Line3DConfig) {
			cfg.VisualRange = nil
			cfg.Series[0].Color = "#4575b4"
		}, `data-goshtoso-charts-line3d-paints="[{&#34;color&#34;:&#34;#4575b4&#34;}]"`},
		"series class": {func(cfg *Line3DConfig) {
			cfg.VisualRange = nil
			cfg.Series[0].Class = "semantic-path"
		}, `data-goshtoso-charts-line3d-paints="[{&#34;class&#34;:&#34;semantic-path&#34;}]"`},
		"series options": {func(cfg *Line3DConfig) {
			cfg.Series[0].Options = SeriesOptions{LineStyle: &LineStyle{Width: 3, Type: "dashed", Opacity: Float(.6)}, Animation: Bool(false)}
		}, `"lineStyle":{"width":3,"type":"dashed","opacity":0.6}`},
		"size": {func(cfg *Line3DConfig) {
			cfg.Width, cfg.Height = "42rem", "28rem"
		}, `style="width:42rem;height:28rem;"`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := validLine3DConfig()
			test.mutate(&cfg)
			if markup := renderLine3D(t, Line3D(cfg)); !strings.Contains(markup, test.want) {
				t.Errorf("rendered markup missing %q", test.want)
			}
		})
	}
}

func TestLine3DRejectsInvalidDataAndOptions(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		mutate func(*Line3DConfig)
		want   string
	}{
		"label":         {func(cfg *Line3DConfig) { cfg.Label = " " }, "line 3D chart label is required"},
		"formula":       {func(cfg *Line3DConfig) { cfg.DataSummary.Formula = "" }, "line 3D chart data formula is required"},
		"parameter":     {func(cfg *Line3DConfig) { cfg.DataSummary.Parameter = "" }, "line 3D chart data parameter is required"},
		"domain finite": {func(cfg *Line3DConfig) { cfg.DataSummary.ParameterMax = math.NaN() }, "line 3D chart parameter domain must be finite"},
		"domain order": {func(cfg *Line3DConfig) {
			cfg.DataSummary.ParameterMin, cfg.DataSummary.ParameterMax = 2, 1
		}, "line 3D chart parameter minimum must not exceed maximum"},
		"series":      {func(cfg *Line3DConfig) { cfg.Series = nil }, "line 3D chart series is required"},
		"series name": {func(cfg *Line3DConfig) { cfg.Series[0].Name = "" }, "line 3D chart series 0 name is required"},
		"point count": {func(cfg *Line3DConfig) { cfg.Series[0].Points = cfg.Series[0].Points[:1] }, `series "line3D" requires at least two points`},
		"finite":      {func(cfg *Line3DConfig) { cfg.Series[0].Points[0].Z = math.Inf(1) }, `point 0 coordinates must be finite`},
		"point option": {func(cfg *Line3DConfig) {
			cfg.Series[0].Points[0].Color = "red"
		}, `point 0 supports coordinates only`},
		"paint": {func(cfg *Line3DConfig) {
			cfg.Series[0].Color, cfg.Series[0].Class = "red", "semantic"
		}, `color and class are mutually exclusive`},
		"unsupported series": {func(cfg *Line3DConfig) {
			cfg.Series[0].Options.Symbol = "circle"
		}, `contains unsupported series options`},
		"line width": {func(cfg *Line3DConfig) {
			cfg.Series[0].Options.LineStyle = &LineStyle{Width: math.NaN()}
		}, `line width must be finite and nonnegative`},
		"line opacity": {func(cfg *Line3DConfig) {
			cfg.Series[0].Options.LineStyle = &LineStyle{Opacity: Float(2)}
		}, `line opacity must be finite and between zero and one`},
		"line type": {func(cfg *Line3DConfig) {
			cfg.Series[0].Options.LineStyle = &LineStyle{Type: "wave"}
		}, `line type "wave" is not supported`},
		"grid partial": {func(cfg *Line3DConfig) {
			cfg.Grid = Line3DGrid{Width: 100}
		}, "grid width, height, and depth must be set together"},
		"grid finite": {func(cfg *Line3DConfig) {
			cfg.Grid = Line3DGrid{Width: math.Inf(1), Height: 100, Depth: 100}
		}, "grid sizes must be finite and positive"},
		"rotation": {func(cfg *Line3DConfig) {
			cfg.Grid.View = &Line3DView{AutoRotateSpeed: math.NaN()}
		}, "auto-rotation speed must be finite and nonnegative"},
		"range": {func(cfg *Line3DConfig) {
			cfg.VisualRange.Min, cfg.VisualRange.Max = 31, 30
		}, "visual range minimum must not exceed maximum"},
		"palette": {func(cfg *Line3DConfig) {
			cfg.VisualRange.Palette = "rainbow"
		}, `visual range palette "rainbow" is not supported`},
		"custom colors": {func(cfg *Line3DConfig) {
			cfg.VisualRange.Palette, cfg.VisualRange.Colors = Line3DPaletteCustom, []string{"red"}
		}, "custom visual range requires at least two colors"},
		"axes": {func(cfg *Line3DConfig) {
			cfg.Options.XAxis = &AxisOptions{}
		}, "Cartesian axes are not supported"},
		"tooltip": {func(cfg *Line3DConfig) {
			cfg.Options.Tooltip = &TooltipOptions{Trigger: "axis"}
		}, `tooltip trigger "axis" is not supported`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := validLine3DConfig()
			test.mutate(&cfg)
			var output bytes.Buffer
			err := Line3D(cfg).Render(context.Background(), &output)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Render() error = %v, want %q", err, test.want)
			}
		})
	}
}

func validLine3DConfig() Line3DConfig {
	return Line3DConfig{
		Label: "basic line3d example",
		Series: []Line3DSeries{{
			Name: "line3D",
			Points: []Point3D{
				{X: 1.25, Y: 0, Z: 0},
				{X: 1, Y: 0, Z: 1},
			},
		}},
		VisualRange: &Line3DVisualRange{Min: 0, Max: 30, Calculable: Bool(true)},
		DataSummary: Line3DDataSummary{
			Formula: "t = i / 1000", Parameter: "t", ParameterMin: 0, ParameterMax: .001,
		},
		Options: ChartOptions{Title: &TitleOptions{Text: "basic line3d example"}},
	}
}

func renderLine3D(t *testing.T, instance Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	return output.String()
}
