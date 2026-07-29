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

func TestGeoRendersTypedCoordinatesRippleThemeAndExactValues(t *testing.T) {
	t.Parallel()
	cfg := validGeoConfig()
	cfg.Caption = "National coordinate values."
	cfg.GeometryPaint = GeoPaint{Class: "land-area"}
	cfg.Series[0].Color = "#123456"
	cfg.Series[0].Points[0].Class = "capital-point"
	cfg.Series[0].Options = SeriesOptions{Label: &LabelOptions{Show: Bool(true)}, SymbolSize: 18}
	cfg.Options = ChartOptions{
		Title: &TitleOptions{Text: "basic geo example"}, Tooltip: &TooltipOptions{Show: Bool(true), Trigger: "item"},
		Animation: Bool(false), Controls: chartcontrol.Options{Fullscreen: true},
		Export: &chartcontrol.ExportOptions{Filename: "national-geo"},
	}
	cfg.Style = charttheme.Style{Palette: charttheme.PaletteAraiHu, Class: "caller-class"}
	cfg.RootAttrs = templ.Attributes{"id": "coordinate-values", "data-purpose": "geography"}

	instance := Geo(cfg)
	if instance.Kind() != chartcomponents.KindInteractiveGeo {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	markup := renderGeo(t, instance)
	for _, want := range []string{
		`class="goshtoso-charts-interactive goshtoso-charts-palette goshtoso-charts-palette-araihu goshtoso-charts-geo caller-class"`,
		`role="img"`, `aria-label="basic geo example"`, `id="coordinate-values"`, `data-purpose="geography"`,
		`style="width:100%;height:500px;"`, `"type":"effectScatter"`, `"coordinateSystem":"geo"`, `"map":"brazil"`,
		`"period":4`, `"scale":6`, `"brushType":"stroke"`, `"show":true`, `"symbolSize":18`,
		`{"name":"Manaus","value":[-60.02,-3.12,81],"className":"capital-point"}`,
		`data-goshtoso-charts-geo-geometry-paint="{&#34;class&#34;:&#34;land-area&#34;}"`,
		`data-goshtoso-charts-geo-series-paints="[{&#34;color&#34;:&#34;#123456&#34;}]"`,
		`National coordinate values.`, `>Exact coordinate values</summary>`, `scope="col">Longitude</th>`,
		`scope="row">Manaus</th>`, `>-60.02</td>`, `>-3.12</td>`, `>81</td>`, `>capital-point</td>`,
		`data-goshtoso-chart-expand`,
		`-fullscreen-action`, `exportFromMenu($el, &#34;png&#34;)`,
		`var resizeObserver = window.ResizeObserver ? new ResizeObserver`,
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

func TestGeoRendersSaoPauloScatterWithThemeVisualRange(t *testing.T) {
	t.Parallel()
	cfg := GeoConfig{
		Label: "São Paulo cities", Geometry: GeoGeometrySaoPaulo,
		VisualRange: &GeoVisualRange{Min: 0, Max: 100, Calculable: Bool(true)},
		Series: []GeoSeries{{
			Name: "geo", Kind: GeoSeriesScatter,
			Points: []GeoPoint{
				{Name: "São Paulo", Longitude: -46.63, Latitude: -23.55, Value: 12},
				{Name: "Campinas", Longitude: -47.06, Latitude: -22.91, Value: 76, Color: "#abcdef"},
				{Name: "Ribeirão Preto", Longitude: -47.81, Latitude: -21.18, Value: 41},
			},
		}},
	}
	markup := renderGeo(t, Geo(cfg))
	for _, want := range []string{
		`"type":"scatter"`, `"map":"brazil-sao-paulo"`, `"calculable":true`, `"max":100`,
		`"value":[-46.63,-23.55,12]`, `"value":[-47.06,-22.91,76]`, `"value":[-47.81,-21.18,41]`,
		`"sourceColor":"#abcdef","itemStyle":{"color":"#abcdef"}`,
		`data-goshtoso-charts-explicit-visual-map-colors="false"`,
		`if (!explicitVisualMapColors) visualMap.inRange = { color: [scaleLow, scaleMid, scaleHigh] }`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestGeoEveryExportedPaintFieldChangesPrivateRendererInput(t *testing.T) {
	t.Parallel()
	cfg := validGeoConfig()
	cfg.GeometryPaint = GeoPaint{Color: "#111111"}
	cfg.Series[0].Color = "#222222"
	cfg.Series[0].Points[0].Color = "#333333"
	markup := renderGeo(t, Geo(cfg))
	for _, want := range []string{
		`"itemStyle":{"areaColor":"#111111"}`,
		`data-goshtoso-charts-geo-geometry-paint="{&#34;color&#34;:&#34;#111111&#34;}"`,
		`data-goshtoso-charts-geo-series-paints="[{&#34;color&#34;:&#34;#222222&#34;}]"`,
		`"sourceColor":"#333333","itemStyle":{"color":"#333333"}`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("Color paint markup missing %q", want)
		}
	}

	cfg = validGeoConfig()
	cfg.GeometryPaint = GeoPaint{Class: "land-class"}
	cfg.Series[0].Class = "series-class"
	cfg.Series[0].Points[0].Class = "point-class"
	markup = renderGeo(t, Geo(cfg))
	for _, want := range []string{
		`data-goshtoso-charts-geo-geometry-paint="{&#34;class&#34;:&#34;land-class&#34;}"`,
		`data-goshtoso-charts-geo-series-paints="[{&#34;class&#34;:&#34;series-class&#34;}]"`,
		`"className":"point-class"`,
		`classColorOrFallback(figure, geoGeometryPaint.class, surfaceAlt)`,
		`classColorOrFallback(figure, geoSeriesPaint.class, palette[index % palette.length])`,
		`item.className ? classColorOrFallback(figure, item.className, geoSeriesColor)`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("Class paint markup missing %q", want)
		}
	}
}

func TestGeoDefaultsToSharedExpandDirectPNGAndResponsiveSize(t *testing.T) {
	t.Parallel()
	markup := renderGeo(t, Geo(validGeoConfig()))
	for _, want := range []string{
		`data-goshtoso-chart-expand`, `exportFromMenu($el, &#34;png&#34;)`,
		`style="width:100%;height:500px;"`, `margin-inline: auto`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("default markup missing %q", want)
		}
	}
	for _, unwanted := range []string{`-fullscreen-action"`, `echarts.dispose`} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("default markup contains %q", unwanted)
		}
	}
}

func TestGeoRejectsInvalidDataAndOptions(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		mutate func(*GeoConfig)
		want   string
	}{
		"missing label":        {func(cfg *GeoConfig) { cfg.Label = " " }, "geo chart label is required"},
		"invalid geometry":     {func(cfg *GeoConfig) { cfg.Geometry = "moon" }, `geo chart geometry "moon" is not supported`},
		"missing series":       {func(cfg *GeoConfig) { cfg.Series = nil }, "geo chart series are required"},
		"missing series name":  {func(cfg *GeoConfig) { cfg.Series[0].Name = "" }, "geo chart series 0 name is required"},
		"invalid series kind":  {func(cfg *GeoConfig) { cfg.Series[0].Kind = "heat" }, `geo chart series "geo" kind "heat" is not supported`},
		"missing points":       {func(cfg *GeoConfig) { cfg.Series[0].Points = nil }, `geo chart series "geo" points are required`},
		"missing point name":   {func(cfg *GeoConfig) { cfg.Series[0].Points[0].Name = "" }, `geo chart series "geo" point 0 name is required`},
		"duplicate point":      {func(cfg *GeoConfig) { cfg.Series[0].Points = append(cfg.Series[0].Points, cfg.Series[0].Points[0]) }, `geo chart series "geo" point "Manaus" is duplicated`},
		"invalid longitude":    {func(cfg *GeoConfig) { cfg.Series[0].Points[0].Longitude = 181 }, `geo chart point "Manaus" longitude must be finite and within [-180, 180]`},
		"invalid latitude":     {func(cfg *GeoConfig) { cfg.Series[0].Points[0].Latitude = math.NaN() }, `geo chart point "Manaus" latitude must be finite and within [-90, 90]`},
		"outside geometry":     {func(cfg *GeoConfig) { cfg.Series[0].Points[0].Longitude = -20 }, `geo chart point "Manaus" is outside selected geometry bounds`},
		"nonfinite value":      {func(cfg *GeoConfig) { cfg.Series[0].Points[0].Value = math.Inf(1) }, `geo chart point "Manaus" value must be finite`},
		"ripple on scatter":    {func(cfg *GeoConfig) { cfg.Series[0].Kind = GeoSeriesScatter }, `geo chart series "geo" ripple requires effect scatter`},
		"zero ripple period":   {func(cfg *GeoConfig) { cfg.Series[0].Ripple.Period = 0 }, `geo chart series "geo" ripple period must be positive and finite`},
		"zero ripple scale":    {func(cfg *GeoConfig) { cfg.Series[0].Ripple.Scale = 0 }, `geo chart series "geo" ripple scale must be positive and finite`},
		"invalid ripple brush": {func(cfg *GeoConfig) { cfg.Series[0].Ripple.BrushType = "dots" }, `geo chart series "geo" ripple brush type "dots" is not supported`},
		"reversed range":       {func(cfg *GeoConfig) { cfg.VisualRange = &GeoVisualRange{Min: 10, Max: 1} }, "geo chart visual range minimum must be less than maximum"},
		"one range color":      {func(cfg *GeoConfig) { cfg.VisualRange = &GeoVisualRange{Min: 0, Max: 1, Colors: []string{"red"}} }, "geo chart visual range colors require at least two values"},
		"geometry paint clash": {func(cfg *GeoConfig) { cfg.GeometryPaint = GeoPaint{Color: "red", Class: "land"} }, "geo chart geometry color and class are mutually exclusive"},
		"series paint clash":   {func(cfg *GeoConfig) { cfg.Series[0].Color, cfg.Series[0].Class = "red", "cities" }, `geo chart series "geo" color and class are mutually exclusive`},
		"series option clash": {func(cfg *GeoConfig) {
			cfg.Series[0].Class, cfg.Series[0].Options.ItemStyle = "cities", &ItemStyle{Color: "red"}
		}, `geo chart series "geo" paint and item-style color are mutually exclusive`},
		"point paint clash": {func(cfg *GeoConfig) { cfg.Series[0].Points[0].Color, cfg.Series[0].Points[0].Class = "red", "capital" }, `geo chart point "Manaus" color and class are mutually exclusive`},
		"invalid tooltip":   {func(cfg *GeoConfig) { cfg.Options.Tooltip = &TooltipOptions{Trigger: "axis"} }, `geo chart tooltip trigger "axis" is not supported`},
		"legend":            {func(cfg *GeoConfig) { cfg.Options.Legend = &LegendOptions{} }, "geo chart legend is not supported"},
		"Cartesian axis":    {func(cfg *GeoConfig) { cfg.Options.XAxis = &AxisOptions{} }, "geo chart Cartesian axes are not supported"},
		"reserved attr":     {func(cfg *GeoConfig) { cfg.RootAttrs = templ.Attributes{"role": "group"} }, `geo chart root attribute "role" is reserved`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := validGeoConfig()
			test.mutate(&cfg)
			var output bytes.Buffer
			err := Geo(cfg).Render(context.Background(), &output)
			if err == nil || err.Error() != test.want {
				t.Fatalf("Render() error = %v, want %q", err, test.want)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %d bytes for invalid config", output.Len())
			}
		})
	}
}

func validGeoConfig() GeoConfig {
	return GeoConfig{
		Label: "basic geo example",
		Series: []GeoSeries{{
			Name: "geo", Kind: GeoSeriesEffectScatter,
			Points: []GeoPoint{
				{Name: "Manaus", Longitude: -60.02, Latitude: -3.12, Value: 81},
				{Name: "Recife", Longitude: -34.88, Latitude: -8.05, Value: 27},
			},
			Ripple: &RippleOptions{Period: 4, Scale: 6, BrushType: "stroke"},
		}},
	}
}

func renderGeo(t *testing.T, instance Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	return output.String()
}
