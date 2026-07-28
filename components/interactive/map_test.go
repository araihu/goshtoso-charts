package interactive

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func TestMapRendersTypedRegionsResourceScaleThemeAndExactValues(t *testing.T) {
	t.Parallel()
	cfg := validMapConfig()
	cfg.Caption = "Nine provincial values."
	cfg.Variant = MapVariantScale
	cfg.ShowLabels = Bool(true)
	cfg.Scale = &MapScale{Min: 0, Max: 150, Calculable: Bool(true), Colors: []string{"#50a3ba", "#eac736", "#d94e5d"}}
	cfg.Series.Regions[0].Class = "capital-region"
	cfg.Series.Regions[1].Color = "#123456"
	cfg.Style = charttheme.Style{Palette: charttheme.PaletteAraiHu, Class: "caller-class"}
	cfg.RootAttrs = templ.Attributes{"id": "regional-values", "data-purpose": "geography"}
	cfg.Options = ChartOptions{
		Title: &TitleOptions{Text: "VisualMap"}, Tooltip: &TooltipOptions{Show: Bool(true), Trigger: "item"},
		Animation: Bool(false), Controls: chartcontrol.Options{Fullscreen: true, Collapsible: true},
		Export: &chartcontrol.ExportOptions{Filename: "regional-values"},
	}

	instance := Map(cfg)
	if instance.Kind() != chartcomponents.KindInteractiveMap {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	markup := renderMap(t, instance)
	for _, want := range []string{
		`class="goshtoso-charts-interactive goshtoso-charts-palette goshtoso-charts-palette-araihu goshtoso-charts-map caller-class"`,
		`role="img"`, `aria-label="basic map example"`, `id="regional-values"`, `data-purpose="geography"`,
		`style="width:100%;height:500px;"`, `"type":"map"`, `"map":"china"`, `"show":true`,
		`"calculable":true`, `"max":150`, `"color":["#50a3ba","#eac736","#d94e5d"]`,
		`{"name":"北京","value":101,"className":"capital-region"}`,
		`{"name":"上海","value":72,"sourceColor":"#123456","itemStyle":{"color":"#123456"}}`,
		`data-goshtoso-charts-explicit-visual-map-colors="true"`, `series.type === "map"`,
		`Nine provincial values.`, `>Exact region values</summary>`, `scope="col">Region</th>`, `>北京</th>`, `>101</td>`, `>capital-region</td>`,
		`data-goshtoso-chart-expand`, `data-goshtoso-chart-control="collapse"`, `data-goshtoso-chart-control="fullscreen"`, `data-goshtoso-chart-export="png"`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	for _, unwanted := range []string{"go-echarts", "Apache ECharts", "echarts.dispose"} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("rendered markup contains %q", unwanted)
		}
	}
}

func TestMapVariantsPreserveOneKindAndRegionalGeometry(t *testing.T) {
	t.Parallel()
	for _, variant := range []MapVariant{MapVariantBasic, MapVariantLabels, MapVariantScale, MapVariantRegional, MapVariantTheme} {
		cfg := validMapConfig()
		cfg.Variant = variant
		markup := renderMap(t, Map(cfg))
		if variant == MapVariantRegional && !strings.Contains(markup, `"map":"广东"`) {
			t.Errorf("regional variant missing Guangdong resource")
		}
		if Map(cfg).Kind() != chartcomponents.KindInteractiveMap {
			t.Errorf("variant %q changed component identity", variant)
		}
	}
}

func TestMapDefaultsToSharedExpandDirectPNGAndResponsiveSize(t *testing.T) {
	t.Parallel()
	markup := renderMap(t, Map(validMapConfig()))
	for _, want := range []string{`data-goshtoso-chart-expand`, `data-goshtoso-chart-export="png"`, `aspect-ratio: 9 / 5`, `style="width:100%;height:500px;"`, `var resizeObserver = window.ResizeObserver ? new ResizeObserver`} {
		if !strings.Contains(markup, want) {
			t.Errorf("default markup missing %q", want)
		}
	}
	for _, unwanted := range []string{`data-goshtoso-chart-control="collapse"`, `data-goshtoso-chart-control="fullscreen"`, `echarts.dispose`} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("default markup contains %q", unwanted)
		}
	}
}

func TestMapRejectsInvalidDataAndOptions(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		mutate func(*MapConfig)
		want   string
	}{
		"missing label":       {func(cfg *MapConfig) { cfg.Label = " " }, "map chart label is required"},
		"missing series":      {func(cfg *MapConfig) { cfg.Series.Name = " " }, "map chart series name is required"},
		"missing regions":     {func(cfg *MapConfig) { cfg.Series.Regions = nil }, "map chart regions are required"},
		"invalid geometry":    {func(cfg *MapConfig) { cfg.Geometry = "moon" }, `map chart geometry "moon" is not supported`},
		"invalid variant":     {func(cfg *MapConfig) { cfg.Variant = "globe" }, `map chart variant "globe" is not supported`},
		"missing region name": {func(cfg *MapConfig) { cfg.Series.Regions[0].Name = "" }, "map chart region 0 name is required"},
		"duplicate region":    {func(cfg *MapConfig) { cfg.Series.Regions[1].Name = "北京" }, `map chart region "北京" is duplicated`},
		"nonfinite value":     {func(cfg *MapConfig) { cfg.Series.Regions[0].Value = math.NaN() }, `map chart region "北京" value must be finite`},
		"reversed scale":      {func(cfg *MapConfig) { cfg.Scale = &MapScale{Min: 10, Max: 1} }, "map chart scale minimum must not exceed maximum"},
		"one scale color":     {func(cfg *MapConfig) { cfg.Scale = &MapScale{Colors: []string{"red"}} }, "map chart scale colors require at least two values"},
		"invalid tooltip":     {func(cfg *MapConfig) { cfg.Options.Tooltip = &TooltipOptions{Trigger: "axis"} }, `map chart tooltip trigger "axis" is not supported`},
		"legend":              {func(cfg *MapConfig) { cfg.Options.Legend = &LegendOptions{} }, "map chart legend is not supported"},
		"Cartesian axis":      {func(cfg *MapConfig) { cfg.Options.XAxis = &AxisOptions{} }, "map chart Cartesian axes are not supported"},
		"reserved attr":       {func(cfg *MapConfig) { cfg.RootAttrs = templ.Attributes{"role": "group"} }, `map chart root attribute "role" is reserved`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := validMapConfig()
			test.mutate(&cfg)
			var output bytes.Buffer
			err := Map(cfg).Render(context.Background(), &output)
			if err == nil || err.Error() != test.want {
				t.Fatalf("Render() error = %v, want %q", err, test.want)
			}
		})
	}
}

func validMapConfig() MapConfig {
	return MapConfig{Label: "basic map example", Series: MapSeries{Name: "map", Regions: []MapRegion{{Name: "北京", Value: 101}, {Name: "上海", Value: 72}}}}
}

func renderMap(t *testing.T, instance Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	return output.String()
}
