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

func TestScatter3DRendersSourceSemanticsAndExactPoints(t *testing.T) {
	t.Parallel()
	cfg := Scatter3DConfig{
		Label: "basic Scatter3D example", Caption: "Deterministic points.",
		Series: []Scatter3DSeries{{Name: "scatter3d", Points: []Point3D{
			{Name: "point-01", X: 10, Y: 20, Z: 30},
			{Name: "point-02", X: 40, Y: 50, Z: 60, Symbol: "diamond", SymbolSize: 14},
		}}},
		VisualRange: &Scatter3DVisualRange{Max: 100, Calculable: Bool(true), Palette: Scatter3DPaletteColdToWarm},
		Options:     ChartOptions{Title: &TitleOptions{Text: "basic Scatter3D example"}},
		Style:       charttheme.Style{Class: "caller-class"},
		RootAttrs:   templ.Attributes{"id": "scatter-space"},
	}
	instance := Scatter3D(cfg)
	if instance.Kind() != chartcomponents.KindInteractiveScatter3D {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	markup := renderScatter3D(t, instance)
	for _, want := range []string{
		`goshtoso-charts-scatter-3d caller-class`, `role="img"`, `aria-label="basic Scatter3D example"`, `id="scatter-space"`,
		`style="width:100%;height:38rem;"`, `"type":"scatter3D"`, `"coordinateSystem":"cartesian3D"`,
		`"calculable":true`, `"max":100`, `"color":["#313695","#4575b4","#74add1","#abd9e9","#e0f3f8","#fee090","#fdae61","#f46d43","#d73027","#a50026"]`,
		`{"name":"point-01","value":[10,20,30]}`, `"symbol":"diamond","symbolSize":14`,
		`data-goshtoso-charts-scatter3d-cold-to-warm="true"`, `Exact 3D point values`, `2 points across 1 series.`,
		`>point-01</th>`, `>30</td>`, `data-goshtoso-chart-expand`, `data-goshtoso-chart-export="png"`,
		`var resizeObserver = window.ResizeObserver ? new ResizeObserver`, `series.type === "scatter3D"`,
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

func TestScatter3DRendersAxesAndPointStyles(t *testing.T) {
	t.Parallel()
	cfg := Scatter3DConfig{
		Label: "user-defined item style",
		Axes: &Scatter3DAxes{
			X: Scatter3DAxis{Name: "MY-X-AXIS", Show: Bool(true)},
			Y: Scatter3DAxis{Name: "MY-Y-AXIS"},
			Z: Scatter3DAxis{Name: "MY-Z-AXIS"},
		},
		Series: []Scatter3DSeries{{Name: "scatter3d", Options: Scatter3DSeriesOptions{Class: "series-class"}, Points: []Point3D{
			{Name: "point1", X: 10, Y: 10, Z: 10, Color: "green"},
			{Name: "point2", X: 15, Y: 15, Z: 15, Color: "blue"},
			{Name: "point3", X: 20, Y: 20, Z: 20, Color: "red"},
		}}},
	}
	markup := renderScatter3D(t, Scatter3D(cfg))
	for _, want := range []string{
		`"xAxis3D":{"show":true,"name":"MY-X-AXIS"}`, `"yAxis3D":{"name":"MY-Y-AXIS"}`, `"zAxis3D":{"name":"MY-Z-AXIS"}`,
		`"name":"point1","value":[10,10,10],"sourceColor":"green","itemStyle":{"color":"green"}`,
		`"name":"point2","value":[15,15,15],"sourceColor":"blue","itemStyle":{"color":"blue"}`,
		`"name":"point3","value":[20,20,20],"sourceColor":"red","itemStyle":{"color":"red"}`,
		`data-goshtoso-charts-scatter3d-paints="[{&#34;class&#34;:&#34;series-class&#34;}]"`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestScatter3DRejectsInvalidDataAndConflicts(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		mutate func(*Scatter3DConfig)
		want   string
	}{
		"label":       {func(cfg *Scatter3DConfig) { cfg.Label = " " }, "scatter 3D chart label is required"},
		"series":      {func(cfg *Scatter3DConfig) { cfg.Series = nil }, "scatter 3D chart series is required"},
		"series name": {func(cfg *Scatter3DConfig) { cfg.Series[0].Name = "" }, "scatter 3D chart series 0 name is required"},
		"points":      {func(cfg *Scatter3DConfig) { cfg.Series[0].Points = nil }, `scatter 3D chart series "points" points are required`},
		"point name":  {func(cfg *Scatter3DConfig) { cfg.Series[0].Points[0].Name = "" }, `scatter 3D chart series "points" point 0 name is required`},
		"coordinate":  {func(cfg *Scatter3DConfig) { cfg.Series[0].Points[0].X = math.NaN() }, `scatter 3D chart series "points" point "p" coordinates must be finite`},
		"value":       {func(cfg *Scatter3DConfig) { cfg.Series[0].Points[0].Value = Float(math.Inf(1)) }, `scatter 3D chart series "points" point "p" value must be finite`},
		"paint": {func(cfg *Scatter3DConfig) {
			cfg.Series[0].Points[0].Color, cfg.Series[0].Points[0].Class = "red", "danger"
		}, `scatter 3D chart series "points" point "p" color and class are mutually exclusive`},
		"axis name": {func(cfg *Scatter3DConfig) {
			cfg.Axes = &Scatter3DAxes{Y: Scatter3DAxis{Name: "Y"}, Z: Scatter3DAxis{Name: "Z"}}
		}, "scatter 3D chart x axis name is required"},
		"range": {func(cfg *Scatter3DConfig) {
			cfg.VisualRange = &Scatter3DVisualRange{Min: 2, Max: 1, Colors: []string{"a", "b"}}
		}, "scatter 3D chart visual range minimum must not exceed maximum"},
		"cold custom": {func(cfg *Scatter3DConfig) {
			cfg.VisualRange = &Scatter3DVisualRange{Palette: Scatter3DPaletteColdToWarm, Colors: []string{"a", "b"}}
		}, "scatter 3D chart cold-to-warm palette conflicts with custom colors"},
		"range point paint": {func(cfg *Scatter3DConfig) {
			cfg.Series[0].Points[0].Color = "red"
			cfg.VisualRange = &Scatter3DVisualRange{Colors: []string{"a", "b"}}
		}, "scatter 3D chart visual range conflicts with point paint"},
		"legend":    {func(cfg *Scatter3DConfig) { cfg.Options.Legend = &LegendOptions{} }, "scatter 3D chart legend is not supported"},
		"2D axis":   {func(cfg *Scatter3DConfig) { cfg.Options.XAxis = &AxisOptions{} }, "scatter 3D chart Cartesian axes must use Axes"},
		"root attr": {func(cfg *Scatter3DConfig) { cfg.RootAttrs = templ.Attributes{"role": "group"} }, `scatter 3D chart root attribute "role" is reserved`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := validScatter3DConfig()
			test.mutate(&cfg)
			var output bytes.Buffer
			err := Scatter3D(cfg).Render(context.Background(), &output)
			if err == nil || err.Error() != test.want {
				t.Fatalf("Render() error = %v, want %q", err, test.want)
			}
		})
	}
}

func validScatter3DConfig() Scatter3DConfig {
	return Scatter3DConfig{Label: "scatter", Series: []Scatter3DSeries{{Name: "points", Points: []Point3D{{Name: "p", X: 1, Y: 2, Z: 3}}}}}
}

func renderScatter3D(t *testing.T, instance Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	return output.String()
}
