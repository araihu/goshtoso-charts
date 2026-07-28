package echarts

import (
	"bytes"
	"context"
	"strings"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func TestEffectScatterRendersCategoryChart(t *testing.T) {
	t.Parallel()
	instance := EffectScatter(EffectScatterConfig{
		Label: "Player impact", Caption: "Ripple size shows impact.",
		XAxis: []string{"Kobe", "Jordan"},
		Series: []EffectScatterSeries{{
			Name: "Dunk", Data: []opts.EffectScatterData{{Value: 88}, {Value: 96}},
			Options: []charts.SeriesOpts{charts.WithRippleEffectOpts(opts.RippleEffect{Period: 4, Scale: 10, BrushType: "stroke"})},
		}},
		Width: "720px", Height: "360px",
		GlobalOptions: []charts.GlobalOpts{
			charts.WithTitleOpts(opts.Title{Title: "Impact"}),
			charts.WithColorsOpts(opts.Colors{"#000000"}),
		},
		SeriesOptions: []charts.SeriesOpts{charts.WithEffectScatterChartOpts(opts.EffectScatterChart{Symbol: "diamond", SymbolSize: 18})},
		Style:         charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
	})

	if instance.Kind() != chartcomponents.KindEChartsEffectScatter {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		"Player impact", "Ripple size shows impact.", "width:720px;height:360px",
		`"Kobe","Jordan"`, `"name":"Dunk"`, `"period":4`, `"scale":10`,
		`"brushType":"stroke"`, `"symbol":"diamond"`, `"symbolSize":18`,
		`"text":"Impact"`, `"color":["#123456","#ff8a3d"`,
		"goshtoso-charts-palette-araihu min-h-80",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	if strings.Contains(markup, "#000000") {
		t.Error("explicit Style.Colors did not override global color options")
	}
}

func TestEffectScatterRendersValueAxisCoordinates(t *testing.T) {
	t.Parallel()
	instance := EffectScatter(EffectScatterConfig{
		Label: "Coordinate impact", XAxisType: CartesianAxisValue,
		Series: []EffectScatterSeries{{
			Name: "Events",
			Data: []opts.EffectScatterData{{Name: "first", Value: [2]float64{1.5, 8}}, {Name: "second", Value: [2]float64{3, 13}}},
		}},
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

func TestEffectScatterRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		cfg       EffectScatterConfig
		wantError string
	}{
		"unsupported axis":         {cfg: EffectScatterConfig{XAxisType: "time"}, wantError: `effect scatter chart x axis type "time" is not supported`},
		"missing categories":       {cfg: EffectScatterConfig{}, wantError: "effect scatter chart x axis is required for category mode"},
		"categories in value mode": {cfg: EffectScatterConfig{XAxisType: CartesianAxisValue, XAxis: []string{"A"}}, wantError: "effect scatter chart x axis categories are not allowed for value mode"},
		"missing series":           {cfg: EffectScatterConfig{XAxis: []string{"A"}}, wantError: "effect scatter chart series is required"},
		"missing name":             {cfg: EffectScatterConfig{XAxis: []string{"A"}, Series: []EffectScatterSeries{{Data: []opts.EffectScatterData{{Value: 1}}}}}, wantError: "effect scatter chart series 0 name is required"},
		"missing data":             {cfg: EffectScatterConfig{XAxis: []string{"A"}, Series: []EffectScatterSeries{{Name: "Events"}}}, wantError: `effect scatter chart series "Events" data is required`},
		"misaligned categories":    {cfg: EffectScatterConfig{XAxis: []string{"A", "B"}, Series: []EffectScatterSeries{{Name: "Events", Data: []opts.EffectScatterData{{Value: 1}}}}}, wantError: `effect scatter chart series "Events" has 1 data points for 2 x-axis categories`},
		"invalid coordinate":       {cfg: EffectScatterConfig{XAxisType: CartesianAxisValue, Series: []EffectScatterSeries{{Name: "Events", Data: []opts.EffectScatterData{{Value: []any{1, "two"}}}}}}, wantError: `effect scatter chart series "Events" data point 0 must contain a numeric [x, y] coordinate`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			err := EffectScatter(test.cfg).Render(context.Background(), &output)
			if err == nil || err.Error() != test.wantError {
				t.Fatalf("Render() error = %v, want %q", err, test.wantError)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %d bytes for invalid config", output.Len())
			}
		})
	}
}
