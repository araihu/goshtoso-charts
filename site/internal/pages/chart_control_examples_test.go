package pages

import (
	"bytes"
	"context"
	"math"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	interactive "github.com/araihu/goshtoso-charts/components/interactive"
)

func TestChartControlExamplesReuseCentralizedUpstreamTreatments(t *testing.T) {
	t.Parallel()
	if areaLineUpstreamPath != "examples/1-Painter/line_chart-5-area/main.go" {
		t.Fatalf("static source path = %q", areaLineUpstreamPath)
	}
	if interactiveBarUpstreamPath != "examples/bar.go" {
		t.Fatalf("interactive source path = %q", interactiveBarUpstreamPath)
	}

	examples := defaultChartControlExamples()
	staticConfig := staticChartControlConfig(examples.static)
	areaConfig := sampleAreaLine()
	if !reflect.DeepEqual(staticConfig.Labels, areaConfig.Labels) || !reflect.DeepEqual(staticConfig.Series, areaConfig.Series) || staticConfig.Area != areaConfig.Area {
		t.Fatal("static control example drifted from centralized filled-area treatment")
	}
	interactiveConfig := interactiveChartControlConfig(examples.interactive)
	barConfig := sampleInteractiveBar()
	if !reflect.DeepEqual(interactiveConfig.XAxis, barConfig.XAxis) || !reflect.DeepEqual(interactiveConfig.Series, barConfig.Series) {
		t.Fatal("interactive control example drifted from centralized basic Bar treatment")
	}

	paletteLine := paletteLineConfig(examples.palette)
	if !reflect.DeepEqual(paletteLine.Labels, areaConfig.Labels) || !reflect.DeepEqual(paletteLine.Series, areaConfig.Series) || paletteLine.Area != areaConfig.Area {
		t.Fatal("palette line drifted from centralized filled-area treatment")
	}
	paletteBar := paletteBarConfig(examples.palette)
	upstreamStaticBar := sampleBasicBar()
	if !reflect.DeepEqual(paletteBar.Labels, upstreamStaticBar.Labels[:6]) || !reflect.DeepEqual(paletteBar.Series[0].Values, upstreamStaticBar.Series[0].Values[:6]) || !reflect.DeepEqual(paletteBar.Series[1].Values, upstreamStaticBar.Series[1].Values[:6]) {
		t.Fatal("palette bar drifted from centralized grouped Bar treatment")
	}
	palettePie, upstreamPie := palettePieConfig(examples.palette), sampleBasicPie()
	if !reflect.DeepEqual(palettePie.Slices, upstreamPie.Slices) {
		t.Fatal("palette pie drifted from centralized basic Pie treatment")
	}
	paletteHeatMap, upstreamHeatMap := paletteHeatMapConfig(examples.palette), sampleBasicHeatMap()
	if !reflect.DeepEqual(paletteHeatMap.Rows, upstreamHeatMap.Rows) || paletteHeatMap.ValueRange != upstreamHeatMap.ValueRange {
		t.Fatal("palette heat map drifted from centralized basic Heat Map treatment")
	}
}

func TestChartControlExamplesParseClosedValuesAndDriveTypedConfigs(t *testing.T) {
	t.Parallel()

	defaults := ParseChartControlExamples(nil)
	if defaults.static.ModeValue != "enabled" || defaults.static.StrokeWidth != 3 || !defaults.static.Area {
		t.Fatalf("static defaults = %#v", defaults.static)
	}
	if defaults.interactive.Orientation != interactive.BarOrientationVertical || defaults.interactive.Scale != 100 || !defaults.interactive.ShowLabels {
		t.Fatalf("interactive defaults = %#v", defaults.interactive)
	}
	if defaults.palette.Palette != charttheme.PaletteAuto || defaults.palette.PaletteValue != "theme" || defaults.palette.Custom || defaults.palette.Colors != defaultPaletteChartColors {
		t.Fatalf("palette defaults = %#v", defaults.palette)
	}

	values := url.Values{
		"static_present":          {"1"},
		"static_mode":             {"disabled"},
		"static_stroke":           {"7"},
		"interactive_present":     {"1"},
		"interactive_orientation": {"horizontal"},
		"interactive_scale":       {"150"},
		"palette_present":         {"1"},
		"chart_palette":           {"custom"},
		"palette_color_1":         {"#123456"},
		"palette_color_2":         {"#234567"},
		"palette_color_3":         {"#345678"},
		"palette_color_4":         {"#456789"},
	}
	examples := ParseChartControlExamples(values)
	if examples.static.Mode != chartcontrol.WrapperModeDisabled || examples.static.StrokeWidth != 7 || examples.static.Area {
		t.Fatalf("static parsed state = %#v", examples.static)
	}
	staticConfig := staticChartControlConfig(examples.static)
	if staticConfig.Controls.Mode != chartcontrol.WrapperModeDisabled || staticConfig.StrokeWidth != 7 || staticConfig.Area.Enabled {
		t.Fatalf("static chart config = %#v", staticConfig)
	}
	if examples.interactive.Orientation != interactive.BarOrientationHorizontal || examples.interactive.Scale != 150 || examples.interactive.ShowLabels {
		t.Fatalf("interactive parsed state = %#v", examples.interactive)
	}
	base := interactiveChartControlConfig(defaults.interactive).Series[0].Data[0].Value
	interactiveConfig := interactiveChartControlConfig(examples.interactive)
	if interactiveConfig.Orientation != interactive.BarOrientationHorizontal || *interactiveConfig.SeriesOptions.Label.Show {
		t.Fatalf("interactive chart config = %#v", interactiveConfig)
	}
	want := math.Round(base*1.5*10) / 10
	if got := interactiveConfig.Series[0].Data[0].Value; got != want {
		t.Fatalf("scaled value = %v, want %v", got, want)
	}
	if !examples.palette.Custom || examples.palette.PaletteValue != "custom" || examples.palette.Colors != [4]string{"#123456", "#234567", "#345678", "#456789"} {
		t.Fatalf("palette parsed state = %#v", examples.palette)
	}
	lineConfig := paletteLineConfig(examples.palette)
	if !reflect.DeepEqual(lineConfig.Style.Colors, []string{"#123456", "#234567", "#345678", "#456789"}) {
		t.Fatalf("palette line colors = %#v", lineConfig.Style.Colors)
	}
	barColors := paletteBarConfig(examples.palette).Style.Colors
	if !reflect.DeepEqual(barColors, lineConfig.Style.Colors) {
		t.Fatalf("palette bar colors = %#v", barColors)
	}
	pieColors := palettePieConfig(examples.palette).Style.Colors
	if !reflect.DeepEqual(pieColors, []string{"#123456", "#234567", "#345678", "#456789", "#123456"}) {
		t.Fatalf("palette pie colors = %#v", pieColors)
	}
	heatStops := paletteHeatMapConfig(examples.palette).Gradient.Stops
	if got := []string{heatStops[0].Color, heatStops[1].Color, heatStops[2].Color, heatStops[3].Color}; !reflect.DeepEqual(got, lineConfig.Style.Colors) {
		t.Fatalf("palette heat-map colors = %#v", got)
	}
}

func TestChartControlExamplesAllowInitialCustomPaletteWithoutSubmittedColors(t *testing.T) {
	t.Parallel()
	examples := ParseChartControlExamples(url.Values{
		"palette_present": {"1"},
		"chart_palette":   {"custom"},
	})
	if !examples.palette.Custom || len(examples.palette.Errors) != 0 {
		t.Fatalf("initial custom palette = %#v", examples.palette)
	}
	if examples.palette.Colors != defaultPaletteChartColors {
		t.Fatalf("initial custom colors = %#v, want %#v", examples.palette.Colors, defaultPaletteChartColors)
	}
}

func TestChartControlExamplesRejectUnknownOrOutOfRangeValues(t *testing.T) {
	t.Parallel()
	examples := ParseChartControlExamples(url.Values{
		"static_present":          {"1"},
		"static_mode":             {"sideways"},
		"static_stroke":           {"20"},
		"interactive_present":     {"1"},
		"interactive_orientation": {"diagonal"},
		"interactive_scale":       {"125"},
		"palette_present":         {"1"},
		"chart_palette":           {"custom"},
		"palette_color_1":         {"red"},
		"palette_color_2":         {"#ABCDEF"},
		"palette_color_3":         {"#12345g"},
		"palette_color_4":         {""},
	})
	if examples.static.ModeValue != "enabled" || examples.static.StrokeWidth != 3 || len(examples.static.Errors) != 2 {
		t.Fatalf("invalid static fallback = %#v", examples.static)
	}
	if examples.interactive.Orientation != interactive.BarOrientationVertical || examples.interactive.Scale != 100 || len(examples.interactive.Errors) != 2 {
		t.Fatalf("invalid interactive fallback = %#v", examples.interactive)
	}
	if !examples.palette.Custom || examples.palette.Colors[1] != "#abcdef" || examples.palette.Colors[0] != defaultPaletteChartColors[0] || len(examples.palette.Errors) != 3 {
		t.Fatalf("invalid palette fallback = %#v", examples.palette)
	}
}

func TestChartControlExampleSourceMatchesRenderedControlContract(t *testing.T) {
	t.Parallel()
	templateBytes, err := os.ReadFile("chart_control_examples.templ")
	if err != nil {
		t.Fatal(err)
	}
	paletteTemplateBytes, err := os.ReadFile("chart_control_palette_example.templ")
	if err != nil {
		t.Fatal(err)
	}
	templateSource := string(templateBytes) + string(paletteTemplateBytes)
	if strings.Contains(templateSource, "<script") {
		t.Fatal("chart-control examples add inline script instead of site-owned HTMX behavior")
	}

	tests := []struct {
		name            string
		snippet         string
		templateMarkers []string
		snippetMarkers  []string
	}{
		{
			name:    "static",
			snippet: staticChartControlSource,
			templateMarkers: []string{
				`name="static_present"`, `name="static_mode"`, `Name: "static_stroke"`, `Name: "static_area"`,
				`staticChartControlTarget`, `line.Line(staticChartControlConfig(state))`, `button.WithType("submit")`,
			},
			snippetMarkers: []string{
				`StaticChartControlsHandler`, `templ StaticChartControls`, `static_present`, `static_mode`, `static_stroke`, `static_area`,
				`#static-chart-control-example`, `line.Line(staticConfig(state))`, `WrapperModeDisabled`, `WrapperModeHidden`, `WrapperModeOmitted`,
				`[]float64{120, 132, 101, 134, 90, 230, 210}`, `button.WithType("submit")`,
				`Area: area`, `XAxis: line.CategoryAxisOptions{BoundaryGap: &noGap}`, `YAxes: []line.Axis{{Min: &minimum}}`,
				`Caption: fmt.Sprintf`, `role="alert"`, `aria-live="polite"`,
			},
		},
		{
			name:    "interactive",
			snippet: interactiveChartControlSource,
			templateMarkers: []string{
				`name="interactive_present"`, `name="interactive_orientation"`, `Name: "interactive_scale"`, `Name: "interactive_labels"`,
				`interactiveChartControlTarget`, `interactive.Bar(interactiveChartControlConfig(state))`, `button.WithType("submit")`,
			},
			snippetMarkers: []string{
				`InteractiveChartControlsHandler`, `templ InteractiveChartControls`, `interactive_present`, `interactive_orientation`, `interactive_scale`, `interactive_labels`,
				`#interactive-chart-control-example`, `interactive.Bar(interactiveConfig(state))`, `BarOrientationHorizontal`,
				`rand.New(rand.NewSource(seed))`, `interactive.Bool(state.Labels)`, `button.WithType("submit")`,
				`float64(source.Intn(300)) * float64(scale) / 100`, `math.Round(scaled * 10) / 10`, `Title: &interactive.TitleOptions{Text: "Weekly categories"}`,
				`Tooltip: &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "axis"}`, `role="alert"`, `aria-live="polite"`,
			},
		},
		{
			name:    "palette-grid",
			snippet: paletteChartControlSource,
			templateMarkers: []string{
				`name="palette_present"`, `name="chart_palette"`, `name={ fmt.Sprintf("palette_color_%d", index+1) }`,
				`paletteChartControlTarget`, `line.Line(paletteLineConfig(state))`, `bar.Bar(paletteBarConfig(state))`,
				`pie.Pie(palettePieConfig(state))`, `heatmap.HeatMap(paletteHeatMapConfig(state))`, `lg:grid-cols-2`,
			},
			snippetMarkers: []string{
				`PaletteChartsHandler`, `templ PaletteCharts`, `palette_present`, `chart_palette`, `palette_color_%d`, `customDefaults`,
				`#palette-chart-control-example`, `charttheme.PaletteAraiHu`, `charttheme.PaletteStatus`, `validHex`,
				`@line.Line(lineConfig(state))`, `@bar.Bar(barConfig(state))`, `@pie.Pie(pieConfig(state))`, `@heatmap.HeatMap(heatMapConfig(state))`,
				`heatmap.Gradient{Stops: []heatmap.GradientStop`, `grid gap-5 lg:grid-cols-2`, `Palette: state.Palette`,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, marker := range test.templateMarkers {
				if !strings.Contains(templateSource, marker) {
					t.Errorf("rendered template missing %q", marker)
				}
			}
			for _, marker := range test.snippetMarkers {
				if !strings.Contains(test.snippet, marker) {
					t.Errorf("copyable source missing %q", marker)
				}
			}
		})
	}
	if got := strings.Count(paletteChartControlSource, `Controls: chartcontrol.Options{Fullscreen: true}`); got != 4 {
		t.Errorf("copyable palette source configures fullscreen controls %d times, want 4", got)
	}
}

func TestChartControlExamplesRenderGoshtosoFormsChartsAndCodeBlocks(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	if err := chartControlExamples(defaultChartControlExamples()).Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	body := output.String()
	for _, want := range []string{
		`data-chart-control-examples`, `id="static-chart-control-form"`, `id="interactive-chart-control-form"`,
		`id="palette-chart-control-form"`, `data-chart-palette="theme"`, `data-palette-chart-grid`,
		`hx-get="/docs/chart-controls"`, `hx-swap="outerHTML focus-scroll:false"`, `hx-trigger="change"`,
		`data-chart-control-preview="static"`, `data-chart-control-preview="interactive"`,
		`aria-label="Controlled weekly email line"`, `aria-label="Controlled weekly category bar"`,
		`aria-label="Shared-palette line and area"`, `aria-label="Shared-palette grouped bars"`,
		`aria-label="Shared-palette channel shares"`, `aria-label="Shared-palette heat map"`,
		`id="static-chart-control-source"`, `id="interactive-chart-control-source"`,
		`id="palette-chart-control-source"`, `aria-label="Copy Palette form and four charts · templ code"`,
		`aria-label="Copy Static form and chart · templ code"`, `aria-label="Copy Interactive form and chart · templ code"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered examples missing %q", want)
		}
	}
	for _, forbidden := range []string{"go-echarts", "Apache ECharts", "go-analyze/charts"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("rendered examples expose backing implementation %q", forbidden)
		}
	}
	if got := strings.Count(body, "goshtoso-charts-palette-auto"); got < 4 {
		t.Errorf("default palette grid has %d theme-token chart classes, want at least 4", got)
	}
	for _, token := range []string{"var(--color-chart-series-1)", "var(--color-chart-scale-low)", "var(--color-chart-scale-mid)", "var(--color-chart-scale-high)"} {
		if !strings.Contains(body, token) {
			t.Errorf("default palette grid missing theme token %q", token)
		}
	}
	if got := strings.Count(body, `type="color" name="palette_color_`); got != 4 {
		t.Errorf("palette color input count = %d, want 4", got)
	}
}

func TestChartControlExampleTargetIsClosed(t *testing.T) {
	t.Parallel()
	examples := defaultChartControlExamples()
	for _, target := range []string{staticChartControlTarget, interactiveChartControlTarget, paletteChartControlTarget} {
		if component, ok := ChartControlExampleForTarget(examples, target); !ok || component == nil {
			t.Errorf("target %q was not recognized", target)
		}
	}
	if component, ok := ChartControlExampleForTarget(examples, "main-content"); ok || component != nil {
		t.Fatal("guide target escaped closed example target set")
	}
}
