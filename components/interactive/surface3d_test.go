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

func TestSurface3DRendersTypedSurfaceAndDenseDataAccess(t *testing.T) {
	t.Parallel()
	cfg := validSurface3DConfig()
	cfg.Style = charttheme.Style{Class: "caller-class"}
	cfg.RootAttrs = templ.Attributes{"id": "math-surface"}
	instance := Surface3D(cfg)
	if instance.Kind() != chartcomponents.KindInteractiveSurface3D {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	markup := renderSurface3D(t, instance)
	for _, want := range []string{
		`goshtoso-charts-surface-3d caller-class`, `role="img"`, `aria-label="basic surface3D example"`, `id="math-surface"`,
		`style="width:100%;height:38rem;"`, `"type":"surface"`, `{"value":[-1,-1,0]}`, `{"value":[0,-1,0]}`,
		`"calculable":true`, `"min":-3`, `"max":3`, `"range":[-3,3]`,
		`"color":["#313695","#4575b4","#74add1","#abd9e9","#e0f3f8","#fee090","#fdae61","#f46d43","#d73027","#a50026"]`,
		`data-goshtoso-charts-surface3d-cold-to-warm="true"`, `Exact surface data`, `2 ordered points.`,
		`Formula:`, `z = sin(x`, `Download all exact points as CSV`, `download="basic-surface3d-example-exact-data.csv"`,
		`data:text/csv;charset=utf-8,`, `series%2Cx%2Cy%2Cz`, `data-goshtoso-chart-expand`, `exportFromMenu($el, &#34;png&#34;)`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	if strings.Contains(markup, "<tbody") || strings.Count(markup, "<tr") != 0 {
		t.Error("surface dense data injected table rows into initial DOM")
	}
	for _, unwanted := range []string{"go-echarts", "Apache ECharts", "echarts-gl", "echarts.dispose"} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("rendered markup contains %q", unwanted)
		}
	}
}

func TestSurface3DSerializesCallerColorAndClassPaths(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		mutate func(*Surface3DConfig)
		want   string
	}{
		"series color": {func(cfg *Surface3DConfig) { cfg.VisualRange = nil; cfg.Series[0].Style.Color = "#4575b4" }, `data-goshtoso-charts-surface3d-paints="[{&#34;color&#34;:&#34;#4575b4&#34;}]"`},
		"series class": {func(cfg *Surface3DConfig) { cfg.VisualRange = nil; cfg.Series[0].Style.Class = "cold-series" }, `data-goshtoso-charts-surface3d-paints="[{&#34;class&#34;:&#34;cold-series&#34;}]"`},
		"point color":  {func(cfg *Surface3DConfig) { cfg.VisualRange = nil; cfg.Series[0].Points[0].Color = "#d73027" }, `"sourceColor":"#d73027"`},
		"point class":  {func(cfg *Surface3DConfig) { cfg.VisualRange = nil; cfg.Series[0].Points[0].Class = "warm-point" }, `"className":"warm-point"`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := validSurface3DConfig()
			test.mutate(&cfg)
			if markup := renderSurface3D(t, Surface3D(cfg)); !strings.Contains(markup, test.want) {
				t.Errorf("rendered markup missing %q", test.want)
			}
		})
	}
}

func TestSurface3DRejectsInvalidDataAndConflicts(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		mutate func(*Surface3DConfig)
		want   string
	}{
		"label":       {func(cfg *Surface3DConfig) { cfg.Label = " " }, "surface 3D chart label is required"},
		"formula":     {func(cfg *Surface3DConfig) { cfg.DataSummary.Formula = "" }, "surface 3D chart data formula is required"},
		"series":      {func(cfg *Surface3DConfig) { cfg.Series = nil }, "surface 3D chart series is required"},
		"series name": {func(cfg *Surface3DConfig) { cfg.Series[0].Name = "" }, "surface 3D chart series 0 name is required"},
		"points":      {func(cfg *Surface3DConfig) { cfg.Series[0].Points = nil }, `surface 3D chart series "surface3d" points are required`},
		"finite":      {func(cfg *Surface3DConfig) { cfg.Series[0].Points[0].Z = math.NaN() }, `surface 3D chart series "surface3d" point 0 coordinates must be finite`},
		"duplicate": {func(cfg *Surface3DConfig) {
			cfg.Series[0].Points[1].X, cfg.Series[0].Points[1].Y = -1, -1
		}, `surface 3D chart series "surface3d" coordinate [-1,-1] is duplicated`},
		"value": {func(cfg *Surface3DConfig) {
			value := 1.0
			cfg.Series[0].Points[0].Value = &value
		}, `surface 3D chart series "surface3d" point 0 separate visual value is not supported`},
		"symbol": {func(cfg *Surface3DConfig) {
			cfg.Series[0].Points[0].Symbol = "circle"
		}, `surface 3D chart series "surface3d" point 0 symbol options are not supported`},
		"point paint": {func(cfg *Surface3DConfig) {
			cfg.VisualRange = nil
			cfg.Series[0].Points[0].Color, cfg.Series[0].Points[0].Class = "red", "warm"
		}, `surface 3D chart series "surface3d" point 0 color and class are mutually exclusive`},
		"series paint": {func(cfg *Surface3DConfig) {
			cfg.VisualRange = nil
			cfg.Series[0].Style.Color, cfg.Series[0].Style.Class = "red", "warm"
		}, `surface 3D chart series "surface3d" color and class are mutually exclusive`},
		"shading": {func(cfg *Surface3DConfig) {
			cfg.Series[0].Style.Shading = "glossy"
		}, `surface 3D chart series "surface3d" shading "glossy" is not supported`},
		"grid partial": {func(cfg *Surface3DConfig) {
			cfg.Grid = Surface3DGrid{Width: 100}
		}, "surface 3D chart grid width, height, and depth must be set together"},
		"grid finite": {func(cfg *Surface3DConfig) {
			cfg.Grid = Surface3DGrid{Width: math.Inf(1), Height: 100, Depth: 100}
		}, "surface 3D chart grid sizes must be finite and positive when set"},
		"axis": {func(cfg *Surface3DConfig) {
			cfg.Axes = &Surface3DAxes{}
		}, "surface 3D chart x axis name is required"},
		"range": {func(cfg *Surface3DConfig) {
			cfg.VisualRange.Min, cfg.VisualRange.Max = 4, 3
		}, "surface 3D chart visual range minimum must not exceed maximum"},
		"palette": {func(cfg *Surface3DConfig) {
			cfg.VisualRange.Palette = "rainbow"
		}, `surface 3D chart visual range palette "rainbow" is not supported`},
		"range paint": {func(cfg *Surface3DConfig) {
			cfg.Series[0].Style.Color = "red"
		}, "surface 3D chart visual range conflicts with series paint"},
		"legend": {func(cfg *Surface3DConfig) {
			cfg.Options.Legend = &LegendOptions{}
		}, "surface 3D chart legend is not supported"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := validSurface3DConfig()
			test.mutate(&cfg)
			var output bytes.Buffer
			err := Surface3D(cfg).Render(context.Background(), &output)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Render() error = %v, want %q", err, test.want)
			}
		})
	}
}

func validSurface3DConfig() Surface3DConfig {
	return Surface3DConfig{
		Label: "basic surface3D example",
		Series: []Surface3DSeries{{
			Name: "surface3d",
			Points: []Point3D{
				{X: -1, Y: -1, Z: 0},
				{X: 0, Y: -1, Z: 0},
			},
		}},
		VisualRange: &Surface3DVisualRange{
			Min: -3, Max: 3, Calculable: Bool(true), Palette: Surface3DPaletteColdToWarm,
		},
		DataSummary: Surface3DDataSummary{Formula: "z = sin(x × π) × sin(y × π)"},
		Options:     ChartOptions{Title: &TitleOptions{Text: "basic surface3D example"}},
	}
}

func renderSurface3D(t *testing.T, instance Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	return output.String()
}
