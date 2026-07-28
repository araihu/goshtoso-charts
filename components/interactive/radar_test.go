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

func TestRadarRendersTypedChart(t *testing.T) {
	t.Parallel()
	instance := Radar(RadarConfig{
		Label: "Service profile", Caption: "Current and target service health.",
		Indicators: []RadarIndicator{{Name: "Availability", Max: 100}, {Name: "Latency", Max: 500}, {Name: "Capacity", Max: 200}},
		Series: []RadarSeries{{
			Name:    "Profile",
			Data:    []RadarData{{Name: "Current", Values: []float64{99.9, 180, 120}}, {Name: "Target", Values: []float64{100, 100, 160}}},
			Options: SeriesOptions{AreaStyle: &AreaStyle{Opacity: Float(0.2)}},
		}},
		Width: "720px", Height: "360px",
		Options:       ChartOptions{Title: &TitleOptions{Text: "Health"}},
		SeriesOptions: SeriesOptions{LineStyle: &LineStyle{Width: 2}},
		Style:         charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
	})

	if instance.Kind() != chartcomponents.KindInteractiveRadar {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		"Service profile", "Current and target service health.", "width:720px;height:360px",
		`"name":"Availability","max":100`, `"name":"Latency","max":500`,
		`"name":"Profile"`, `"name":"Current","value":[99.9,180,120]`, `"name":"Target","value":[100,100,160]`,
		`"width":2`, `"opacity":0.2`, `"text":"Health"`, `"color":["#123456","#ff8a3d"`,
		"goshtoso-charts-palette-araihu min-h-80",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestRadarUsesFallbackPalette(t *testing.T) {
	t.Parallel()
	instance := Radar(RadarConfig{
		Label:      "Fallback profile",
		Indicators: []RadarIndicator{{Name: "A", Max: 10}, {Name: "B", Max: 20}},
		Series:     []RadarSeries{{Name: "Profile", Data: []RadarData{{Name: "Current", Values: []float64{4, 8}}}}},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{`"color":["#2563eb","#dc2626"`, "goshtoso-charts-palette-auto"} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestRadarRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	validIndicators := []RadarIndicator{{Name: "A", Max: 10}, {Name: "B", Max: 20}}
	validData := []RadarData{{Name: "Current", Values: []float64{1, 2}}}
	tests := map[string]struct {
		cfg       RadarConfig
		wantError string
	}{
		"missing label":          {cfg: RadarConfig{}, wantError: "radar chart label is required"},
		"missing indicators":     {cfg: RadarConfig{Label: "Profile"}, wantError: "radar chart indicators are required"},
		"missing indicator name": {cfg: RadarConfig{Label: "Profile", Indicators: []RadarIndicator{{Max: 10}}}, wantError: "radar chart indicator 0 name is required"},
		"nonpositive maximum":    {cfg: RadarConfig{Label: "Profile", Indicators: []RadarIndicator{{Name: "A"}}}, wantError: `radar chart indicator "A" maximum must be positive`},
		"nonfinite maximum":      {cfg: RadarConfig{Label: "Profile", Indicators: []RadarIndicator{{Name: "A", Max: float32(math.Inf(1))}}}, wantError: `radar chart indicator "A" maximum must be positive`},
		"missing series":         {cfg: RadarConfig{Label: "Profile", Indicators: validIndicators}, wantError: "radar chart series is required"},
		"missing series name":    {cfg: RadarConfig{Label: "Profile", Indicators: validIndicators, Series: []RadarSeries{{Data: validData}}}, wantError: "radar chart series 0 name is required"},
		"missing series data":    {cfg: RadarConfig{Label: "Profile", Indicators: validIndicators, Series: []RadarSeries{{Name: "Profile"}}}, wantError: `radar chart series "Profile" data is required`},
		"missing data name":      {cfg: RadarConfig{Label: "Profile", Indicators: validIndicators, Series: []RadarSeries{{Name: "Profile", Data: []RadarData{{Values: []float64{1, 2}}}}}}, wantError: `radar chart series "Profile" data point 0 name is required`},
		"empty vector":           {cfg: RadarConfig{Label: "Profile", Indicators: validIndicators, Series: []RadarSeries{{Name: "Profile", Data: []RadarData{{Name: "Current"}}}}}, wantError: `radar chart series "Profile" data "Current" values are required`},
		"misaligned vector":      {cfg: RadarConfig{Label: "Profile", Indicators: validIndicators, Series: []RadarSeries{{Name: "Profile", Data: []RadarData{{Name: "Current", Values: []float64{1}}}}}}, wantError: `radar chart series "Profile" data "Current" has 1 values for 2 indicators`},
		"nonfinite vector":       {cfg: RadarConfig{Label: "Profile", Indicators: validIndicators, Series: []RadarSeries{{Name: "Profile", Data: []RadarData{{Name: "Current", Values: []float64{1, math.NaN()}}}}}}, wantError: `radar chart series "Profile" data "Current" value 1 must be finite`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			err := Radar(test.cfg).Render(context.Background(), &output)
			if err == nil || err.Error() != test.wantError {
				t.Fatalf("Render() error = %v, want %q", err, test.wantError)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %d bytes for invalid config", output.Len())
			}
		})
	}
}
