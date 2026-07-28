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

func TestGaugeRendersStandardVariantWithDefaults(t *testing.T) {
	t.Parallel()
	instance := Gauge(GaugeConfig{
		Label: "Project progress", Caption: "Current completion percentage.",
		Series: []GaugeSeries{{Name: "Project A", Data: []GaugeData{{Name: "Work progress", Value: 43}}}},
		Width:  "720px", Height: "360px",
		Options: ChartOptions{Title: &TitleOptions{Text: "Delivery"}},
		Style:   charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
	})

	if instance.Kind() != chartcomponents.KindInteractiveGauge {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		"Project progress", "Current completion percentage.", "width:720px;height:360px",
		`"name":"Project A"`, `"name":"Work progress","value":43`,
		`"max":100`, `"text":"Delivery"`, `"color":["#123456","#ff8a3d"`,
		"goshtoso-charts-palette-araihu min-h-80",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestGaugeRendersProgressVariantWithSameKind(t *testing.T) {
	t.Parallel()
	instance := Gauge(GaugeConfig{
		Label: "Temperature", Variant: GaugeVariantProgress, Min: -40, Max: 60,
		Series: []GaugeSeries{{
			Name: "Sensor", Data: []GaugeData{{Name: "Current", Value: 21.5}},
			Progress: &GaugeProgressOptions{Width: 18},
		}},
	})
	if instance.Kind() != chartcomponents.KindInteractiveGauge {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`"min":-40`, `"max":60`, `"progress":{"show":true,"width":18,"roundCap":true,"clip":true}`,
		`"pointer":{"show":false}`, `"value":21.5`,
		`data-goshtoso-charts-theme-series-items="0"`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestGaugeRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		config    GaugeConfig
		wantError string
	}{
		"missing label":       {config: GaugeConfig{}, wantError: "gauge chart label is required"},
		"unsupported variant": {config: GaugeConfig{Label: "Gauge", Variant: "dial"}, wantError: `gauge chart variant "dial" is not supported`},
		"invalid range":       {config: GaugeConfig{Label: "Gauge", Min: 100, Max: 50}, wantError: "gauge chart minimum must be less than maximum"},
		"missing series":      {config: GaugeConfig{Label: "Gauge"}, wantError: "gauge chart series is required"},
		"missing series name": {config: GaugeConfig{Label: "Gauge", Series: []GaugeSeries{{Data: []GaugeData{{Name: "Reading", Value: 1}}}}}, wantError: "gauge chart series 0 name is required"},
		"missing series data": {config: GaugeConfig{Label: "Gauge", Series: []GaugeSeries{{Name: "Sensor"}}}, wantError: `gauge chart series "Sensor" data is required`},
		"missing data name":   {config: GaugeConfig{Label: "Gauge", Series: []GaugeSeries{{Name: "Sensor", Data: []GaugeData{{Value: 1}}}}}, wantError: `gauge chart series "Sensor" data point 0 name is required`},
		"non-finite value":    {config: GaugeConfig{Label: "Gauge", Series: []GaugeSeries{{Name: "Sensor", Data: []GaugeData{{Name: "Reading", Value: math.NaN()}}}}}, wantError: `gauge chart series "Sensor" data point "Reading" value must be finite`},
		"out of range":        {config: GaugeConfig{Label: "Gauge", Series: []GaugeSeries{{Name: "Sensor", Data: []GaugeData{{Name: "Reading", Value: 101}}}}}, wantError: `gauge chart series "Sensor" data point "Reading" value must be between 0 and 100`},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			err := Gauge(test.config).Render(context.Background(), &output)
			if err == nil || err.Error() != test.wantError {
				t.Fatalf("Render() error = %v, want %q", err, test.wantError)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %d bytes for invalid config", output.Len())
			}
		})
	}
}
