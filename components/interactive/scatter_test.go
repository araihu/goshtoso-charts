package interactive

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func TestScatterRendersCategoryChart(t *testing.T) {
	t.Parallel()
	instance := Scatter(ScatterConfig{
		Label: "Release quality", Caption: "Defects by release.", XAxis: []string{"v1", "v2"},
		Series: []ScatterSeries{{
			Name: "Defects", Data: []ScatterData{{Value: 8}, {Value: 3}},
			Options: SeriesOptions{Symbol: "diamond", SymbolSize: 18},
		}},
		Width: "720px", Height: "360px",
		Options: ChartOptions{Title: &TitleOptions{Text: "Quality"}},
		Style:   charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
	})

	if instance.Kind() != chartcomponents.KindInteractiveScatter {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		"Release quality", "Defects by release.", "width:720px;height:360px", `"v1","v2"`,
		`"name":"Defects"`, `"symbol":"diamond"`, `"symbolSize":18`, `"text":"Quality"`,
		`"color":["#123456","#ff8a3d"`, "goshtoso-charts-palette-araihu min-h-80",
		`data-scatter-exact-values`, `Release quality exact scatter values`,
		`<th class="px-3 py-2 font-semibold" scope="row">v1</th><td class="px-3 py-2">Defects</td><td class="px-3 py-2 tabular-nums">8</td>`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestScatterRendersValueAxisCoordinates(t *testing.T) {
	t.Parallel()
	instance := Scatter(ScatterConfig{
		Label: "Latency and load", XAxisType: CartesianAxisValue,
		Series: []ScatterSeries{{Name: "Nodes", Data: []ScatterData{
			{Name: "api-1", X: 1.5, Y: 8}, {Name: "api-2", X: 3, Y: 13},
		}}},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{`"type":"value"`, `"value":[1.5,8]`, `"value":[3,13]`, `>api-1</th>`, `>1.5</td>`, `>13</td>`} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestScatterRendersEffectVariantWithSameKind(t *testing.T) {
	t.Parallel()
	instance := Scatter(ScatterConfig{
		Label: "Release impact", Variant: ScatterVariantEffect,
		XAxis:  []string{"v1", "v2"},
		Series: []ScatterSeries{{Name: "Impact", Data: []ScatterData{{Value: 35}, {Value: 91}}}},
		Ripple: &RippleOptions{Period: 4, Scale: 8, BrushType: "stroke"},
	})
	if instance.Kind() != chartcomponents.KindInteractiveScatter {
		t.Fatalf("Kind() = %q, want unified scatter kind", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{`"type":"effectScatter"`, `"period":4`, `"scale":8`, `"brushType":"stroke"`} {
		if !strings.Contains(markup, want) {
			t.Errorf("effect variant markup missing %q", want)
		}
	}
}

func TestScatterRendersPerSeriesRippleOverrides(t *testing.T) {
	t.Parallel()
	instance := Scatter(ScatterConfig{
		Label: "Player actions", Variant: ScatterVariantEffect, XAxis: []string{"Kobe"},
		Series: []ScatterSeries{
			{Name: "Dunk", Data: []ScatterData{{Value: 94}}, Ripple: &RippleOptions{Period: 4, Scale: 10, BrushType: "stroke"}},
			{Name: "Shoot", Data: []ScatterData{{Value: 59}}, Ripple: &RippleOptions{Period: 3, Scale: 6, BrushType: "fill"}},
		},
		Ripple: &RippleOptions{Period: 8, Scale: 2, BrushType: "stroke"},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`"name":"Dunk"`, `"period":4,"scale":10,"brushType":"stroke"`,
		`"name":"Shoot"`, `"period":3,"scale":6,"brushType":"fill"`,
		`data-scatter-exact-values`, `2 points across 2 series.`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("effect variant markup missing %q", want)
		}
	}
	if strings.Contains(markup, `"period":8,"scale":2`) {
		t.Error("shared ripple was not overridden by per-series ripple")
	}
}

func TestInteractiveScatterPropagatesSharedWrapperLifecycle(t *testing.T) {
	t.Parallel()
	base := ScatterConfig{
		Label: "Scores", XAxis: []string{"Swimming"},
		Series: []ScatterSeries{{Name: "Category A", Data: []ScatterData{{Value: 81}}}},
	}
	base.Options.Controls.Mode = chartcontrol.WrapperModeDisabled
	var disabled bytes.Buffer
	if err := Scatter(base).Render(context.Background(), &disabled); err != nil {
		t.Fatalf("disabled Render() error = %v", err)
	}
	if !strings.Contains(disabled.String(), `data-goshtoso-chart-wrapper-mode="disabled"`) ||
		!strings.Contains(disabled.String(), `data-goshtoso-chart-actions-fieldset disabled aria-disabled="true"`) ||
		!strings.Contains(disabled.String(), `_echarts_instance_`) {
		t.Fatal("interactive Scatter did not propagate disabled wrapper mode")
	}

	base.Options.Controls.Mode = chartcontrol.WrapperModeHidden
	var hidden bytes.Buffer
	if err := Scatter(base).Render(context.Background(), &hidden); err != nil {
		t.Fatalf("hidden Render() error = %v", err)
	}
	if !strings.Contains(hidden.String(), `data-goshtoso-chart-wrapper-mode="hidden"`) ||
		!strings.Contains(hidden.String(), `hidden inert aria-hidden="true"`) {
		t.Fatal("interactive Scatter did not propagate hidden wrapper mode")
	}

	base.Options.Controls.Mode = chartcontrol.WrapperModeOmitted
	var omitted bytes.Buffer
	if err := Scatter(base).Render(context.Background(), &omitted); err != nil {
		t.Fatalf("omitted Render() error = %v", err)
	}
	if strings.Contains(omitted.String(), `class="goshtoso-charts-control-wrapper"`) || !strings.Contains(omitted.String(), `goshtoso-charts-interactive`) || !strings.Contains(omitted.String(), `_echarts_instance_`) {
		t.Fatal("interactive Scatter did not propagate omitted wrapper mode")
	}
}

func TestScatterRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		cfg       ScatterConfig
		wantError string
	}{
		"unsupported axis":             {cfg: ScatterConfig{XAxisType: "time"}, wantError: `scatter chart x axis type "time" is not supported`},
		"missing categories":           {cfg: ScatterConfig{}, wantError: "scatter chart x axis is required for category mode"},
		"categories in value mode":     {cfg: ScatterConfig{XAxisType: CartesianAxisValue, XAxis: []string{"A"}}, wantError: "scatter chart x axis categories are not allowed for value mode"},
		"missing series":               {cfg: ScatterConfig{XAxis: []string{"A"}}, wantError: "scatter chart series is required"},
		"missing name":                 {cfg: ScatterConfig{XAxis: []string{"A"}, Series: []ScatterSeries{{Data: []ScatterData{{Value: 1}}}}}, wantError: "scatter chart series 0 name is required"},
		"missing data":                 {cfg: ScatterConfig{XAxis: []string{"A"}, Series: []ScatterSeries{{Name: "Events"}}}, wantError: `scatter chart series "Events" data is required`},
		"misaligned categories":        {cfg: ScatterConfig{XAxis: []string{"A", "B"}, Series: []ScatterSeries{{Name: "Events", Data: []ScatterData{{Value: 1}}}}}, wantError: `scatter chart series "Events" has 1 data points for 2 x-axis categories`},
		"invalid coordinate":           {cfg: ScatterConfig{XAxisType: CartesianAxisValue, Series: []ScatterSeries{{Name: "Events", Data: []ScatterData{{X: math.NaN(), Y: 1}}}}}, wantError: `scatter chart series "Events" data point 0 must contain a numeric [x, y] coordinate`},
		"unsupported variant":          {cfg: ScatterConfig{Variant: "pulse"}, wantError: `scatter chart variant "pulse" is not supported`},
		"ripple without effect":        {cfg: ScatterConfig{Ripple: &RippleOptions{}}, wantError: "scatter chart ripple requires the effect variant"},
		"series ripple without effect": {cfg: ScatterConfig{XAxis: []string{"A"}, Series: []ScatterSeries{{Name: "Events", Data: []ScatterData{{Value: 1}}, Ripple: &RippleOptions{Period: 4}}}}, wantError: `scatter chart series "Events" ripple requires the effect variant`},
		"negative ripple period":       {cfg: ScatterConfig{Variant: ScatterVariantEffect, Ripple: &RippleOptions{Period: -1}}, wantError: "scatter chart shared ripple: period must be finite and nonnegative"},
		"invalid ripple brush":         {cfg: ScatterConfig{Variant: ScatterVariantEffect, XAxis: []string{"A"}, Series: []ScatterSeries{{Name: "Events", Data: []ScatterData{{Value: 1}}, Ripple: &RippleOptions{BrushType: "dash"}}}}, wantError: `scatter chart series "Events" ripple: brush type "dash" is not supported`},
		"invalid point symbol":         {cfg: ScatterConfig{XAxis: []string{"A"}, Series: []ScatterSeries{{Name: "Events", Data: []ScatterData{{Value: 1, Symbol: "star"}}}}}, wantError: `scatter chart series "Events" data point 0 symbol "star" is not supported`},
		"negative point symbol size":   {cfg: ScatterConfig{XAxis: []string{"A"}, Series: []ScatterSeries{{Name: "Events", Data: []ScatterData{{Value: 1, SymbolSize: -1}}}}}, wantError: `scatter chart series "Events" data point 0 symbol size must be nonnegative`},
		"invalid series symbol":        {cfg: ScatterConfig{XAxis: []string{"A"}, Series: []ScatterSeries{{Name: "Events", Data: []ScatterData{{Value: 1}}, Options: SeriesOptions{Symbol: "star"}}}}, wantError: `scatter chart series "Events" symbol "star" is not supported`},
		"invalid shared symbol":        {cfg: ScatterConfig{SeriesOptions: SeriesOptions{Symbol: "star"}}, wantError: `scatter chart shared series options symbol "star" is not supported`},
		"negative shared symbol size":  {cfg: ScatterConfig{SeriesOptions: SeriesOptions{SymbolSize: -1}}, wantError: `scatter chart shared series options symbol size must be nonnegative`},
		"effect per-point symbol":      {cfg: ScatterConfig{Variant: ScatterVariantEffect, XAxis: []string{"A"}, Series: []ScatterSeries{{Name: "Events", Data: []ScatterData{{Value: 1, Symbol: "diamond"}}}}}, wantError: `scatter chart series "Events" data point 0 per-point symbol presentation is unsupported for the effect variant; use series options`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			err := Scatter(test.cfg).Render(context.Background(), &output)
			if err == nil || err.Error() != test.wantError {
				t.Fatalf("Render() error = %v, want %q", err, test.wantError)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %d bytes for invalid config", output.Len())
			}
		})
	}
}
