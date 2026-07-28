package interactive

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
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
	for _, want := range []string{`"type":"value"`, `"value":[1.5,8]`, `"value":[3,13]`} {
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

func TestScatterRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		cfg       ScatterConfig
		wantError string
	}{
		"unsupported axis":         {cfg: ScatterConfig{XAxisType: "time"}, wantError: `scatter chart x axis type "time" is not supported`},
		"missing categories":       {cfg: ScatterConfig{}, wantError: "scatter chart x axis is required for category mode"},
		"categories in value mode": {cfg: ScatterConfig{XAxisType: CartesianAxisValue, XAxis: []string{"A"}}, wantError: "scatter chart x axis categories are not allowed for value mode"},
		"missing series":           {cfg: ScatterConfig{XAxis: []string{"A"}}, wantError: "scatter chart series is required"},
		"missing name":             {cfg: ScatterConfig{XAxis: []string{"A"}, Series: []ScatterSeries{{Data: []ScatterData{{Value: 1}}}}}, wantError: "scatter chart series 0 name is required"},
		"missing data":             {cfg: ScatterConfig{XAxis: []string{"A"}, Series: []ScatterSeries{{Name: "Events"}}}, wantError: `scatter chart series "Events" data is required`},
		"misaligned categories":    {cfg: ScatterConfig{XAxis: []string{"A", "B"}, Series: []ScatterSeries{{Name: "Events", Data: []ScatterData{{Value: 1}}}}}, wantError: `scatter chart series "Events" has 1 data points for 2 x-axis categories`},
		"invalid coordinate":       {cfg: ScatterConfig{XAxisType: CartesianAxisValue, Series: []ScatterSeries{{Name: "Events", Data: []ScatterData{{X: math.NaN(), Y: 1}}}}}, wantError: `scatter chart series "Events" data point 0 must contain a numeric [x, y] coordinate`},
		"unsupported variant":      {cfg: ScatterConfig{Variant: "pulse"}, wantError: `scatter chart variant "pulse" is not supported`},
		"ripple without effect":    {cfg: ScatterConfig{Ripple: &RippleOptions{}}, wantError: "scatter chart ripple requires the effect variant"},
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
