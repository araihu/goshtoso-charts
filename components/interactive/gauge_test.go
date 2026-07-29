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
		`data-goshtoso-charts-gauge-scale=`, `&#34;token&#34;:&#34;low&#34;`, `&#34;token&#34;:&#34;high&#34;`,
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

func TestGaugeRendersLiquidVariantWithNormalizedExactReadings(t *testing.T) {
	t.Parallel()
	instance := Gauge(GaugeConfig{
		Label: "basic liquid example", Caption: "Three bounded wave readings.",
		Variant: GaugeVariantLiquid, Min: 20, Max: 60,
		Series: []GaugeSeries{{
			Name: "liquid", Data: []GaugeData{
				{Name: "Lower wave", Value: 32},
				{Name: "Middle wave", Value: 36},
				{Name: "Upper wave", Value: 40},
			},
		}},
		Liquid: GaugeLiquidTreatment{
			Shape:             GaugeLiquidShapeDiamond,
			WaveLengthPercent: Float(48), AmplitudePercent: Float(12), PhaseDegrees: Float(90),
			Animate: Bool(true), Direction: GaugeLiquidDirectionLeft,
			Outline:    &GaugeLiquidOutline{Show: Bool(true), Width: 4, Class: "text-outline"},
			Background: &GaugeLiquidBackground{Color: "#f8fafc", BorderWidth: 2, BorderClass: "text-border"},
			Label:      &GaugeLiquidLabel{Show: Bool(true), FontSize: 24, Class: "text-label"},
			Style:      &GaugeLiquidStyle{Class: "text-wave", Opacity: Float(.72)},
		},
		Options: ChartOptions{Title: &TitleOptions{Text: "basic liquid example"}},
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
		`width:100%;height:500px`,
		`"type":"liquidFill"`, `"value":0.3`, `"value":0.4`, `"value":0.5`,
		`"shape":"diamond"`, `"waveLength":"48%"`, `"amplitude":"12%"`,
		`"phase":1.5707963267948966`, `"waveAnimation":true`, `"direction":"left"`,
		`data-goshtoso-charts-liquid=`, `&#34;class&#34;:&#34;text-wave&#34;`,
		`data-goshtoso-gauge-liquid-values`, `Lower wave`, `32`, `Middle wave`, `36`, `Upper wave`, `40`,
		`Range: 20 to 60`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered liquid gauge missing %q", want)
		}
	}
	for _, unwanted := range []string{"go-echarts", "echarts-liquidfill", `"type":"gauge"`} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("rendered liquid gauge exposes private renderer %q", unwanted)
		}
	}
}

func TestGaugeLiquidCanonicalValuesMechanicallyNormalize(t *testing.T) {
	t.Parallel()
	values := []GaugeData{{Name: "Wave 1", Value: .3}, {Name: "Wave 2", Value: .4}, {Name: "Wave 3", Value: .5}}
	normalized := normalizeGaugeLiquidData(values, 0, 1)
	want := []float64{.3, .4, .5}
	if len(normalized) != len(want) {
		t.Fatalf("normalized values = %d, want %d", len(normalized), len(want))
	}
	for index := range want {
		if normalized[index].Value != want[index] {
			t.Fatalf("normalized value %d = %#v, want %g", index, normalized[index].Value, want[index])
		}
	}
}

func TestGaugeRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		config    GaugeConfig
		wantError string
	}{
		"missing label":        {config: GaugeConfig{}, wantError: "gauge chart label is required"},
		"unsupported variant":  {config: GaugeConfig{Label: "Gauge", Variant: "dial"}, wantError: `gauge chart variant "dial" is not supported`},
		"invalid range":        {config: GaugeConfig{Label: "Gauge", Min: 100, Max: 50}, wantError: "gauge chart minimum must be less than maximum"},
		"missing series":       {config: GaugeConfig{Label: "Gauge"}, wantError: "gauge chart series is required"},
		"missing series name":  {config: GaugeConfig{Label: "Gauge", Series: []GaugeSeries{{Data: []GaugeData{{Name: "Reading", Value: 1}}}}}, wantError: "gauge chart series 0 name is required"},
		"missing series data":  {config: GaugeConfig{Label: "Gauge", Series: []GaugeSeries{{Name: "Sensor"}}}, wantError: `gauge chart series "Sensor" data is required`},
		"missing data name":    {config: GaugeConfig{Label: "Gauge", Series: []GaugeSeries{{Name: "Sensor", Data: []GaugeData{{Value: 1}}}}}, wantError: `gauge chart series "Sensor" data point 0 name is required`},
		"non-finite value":     {config: GaugeConfig{Label: "Gauge", Series: []GaugeSeries{{Name: "Sensor", Data: []GaugeData{{Name: "Reading", Value: math.NaN()}}}}}, wantError: `gauge chart series "Sensor" data point "Reading" value must be finite`},
		"out of range":         {config: GaugeConfig{Label: "Gauge", Series: []GaugeSeries{{Name: "Sensor", Data: []GaugeData{{Name: "Reading", Value: 101}}}}}, wantError: `gauge chart series "Sensor" data point "Reading" value must be between 0 and 100`},
		"bad scale mode":       {config: GaugeConfig{Label: "Gauge", Scale: GaugeScale{Mode: "rainbow"}}, wantError: `gauge chart scale mode "rainbow" is not supported`},
		"single missing paint": {config: GaugeConfig{Label: "Gauge", Scale: GaugeScale{Mode: GaugeScaleSingleColor}}, wantError: `gauge chart single-color scale requires exactly one color or class`},
		"custom too few stops": {config: GaugeConfig{Label: "Gauge", Scale: GaugeScale{Mode: GaugeScaleCustom, Stops: []GaugeScaleStop{{Value: 100, Color: "red"}}}}, wantError: `gauge chart custom scale requires at least two stops`},
		"custom duplicate":     {config: GaugeConfig{Label: "Gauge", Scale: GaugeScale{Mode: GaugeScaleCustom, Stops: []GaugeScaleStop{{Value: 50, Color: "blue"}, {Value: 50, Color: "red"}, {Value: 100, Color: "orange"}}}}, wantError: `gauge chart scale stops must be strictly increasing`},
		"custom empty class":   {config: GaugeConfig{Label: "Gauge", Scale: GaugeScale{Mode: GaugeScaleCustom, Stops: []GaugeScaleStop{{Value: 50, Class: " "}, {Value: 100, Color: "red"}}}}, wantError: `gauge chart scale stop 0 requires exactly one color or class`},
		"custom final max":     {config: GaugeConfig{Label: "Gauge", Scale: GaugeScale{Mode: GaugeScaleCustom, Stops: []GaugeScaleStop{{Value: 50, Color: "blue"}, {Value: 90, Color: "red"}}}}, wantError: `gauge chart final scale stop must equal maximum`},
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

func TestGaugeRejectsInvalidLiquidContract(t *testing.T) {
	t.Parallel()
	valid := func() GaugeConfig {
		return GaugeConfig{
			Label: "Liquid gauge", Variant: GaugeVariantLiquid,
			Series: []GaugeSeries{{Name: "liquid", Data: []GaugeData{{Name: "Reading", Value: 50}}}},
		}
	}
	tests := map[string]struct {
		mutate    func(*GaugeConfig)
		wantError string
	}{
		"multiple series":  {func(cfg *GaugeConfig) { cfg.Series = append(cfg.Series, cfg.Series[0]) }, "liquid gauge treatment requires exactly one series"},
		"progress":         {func(cfg *GaugeConfig) { cfg.Series[0].Progress = &GaugeProgressOptions{} }, "liquid gauge treatment does not accept progress options"},
		"pointer":          {func(cfg *GaugeConfig) { cfg.Series[0].ShowPointer = Bool(false) }, "liquid gauge treatment does not accept pointer options"},
		"scale":            {func(cfg *GaugeConfig) { cfg.Scale.Reverse = true }, "liquid gauge treatment does not accept dial scale options"},
		"axis":             {func(cfg *GaugeConfig) { cfg.Options.XAxis = &AxisOptions{} }, "liquid gauge treatment does not accept Cartesian axes"},
		"legend":           {func(cfg *GaugeConfig) { cfg.Options.Legend = &LegendOptions{} }, "liquid gauge treatment does not accept a legend"},
		"axis tooltip":     {func(cfg *GaugeConfig) { cfg.Options.Tooltip = &TooltipOptions{Trigger: "axis"} }, `liquid gauge tooltip trigger "axis" is not supported`},
		"line option":      {func(cfg *GaugeConfig) { cfg.SeriesOptions.LineStyle = &LineStyle{} }, "liquid gauge line style is not supported"},
		"bar option":       {func(cfg *GaugeConfig) { cfg.Series[0].Options.BarWidth = "20px" }, "liquid gauge bar width is not supported"},
		"shape":            {func(cfg *GaugeConfig) { cfg.Liquid.Shape = "hexagon" }, `liquid gauge shape "hexagon" is not supported`},
		"length nonfinite": {func(cfg *GaugeConfig) { cfg.Liquid.WaveLengthPercent = Float(math.NaN()) }, "liquid gauge wave length percentage must be finite and between 1 and 100"},
		"length range":     {func(cfg *GaugeConfig) { cfg.Liquid.WaveLengthPercent = Float(101) }, "liquid gauge wave length percentage must be finite and between 1 and 100"},
		"amplitude":        {func(cfg *GaugeConfig) { cfg.Liquid.AmplitudePercent = Float(-1) }, "liquid gauge amplitude percentage must be finite and between 0 and 100"},
		"phase":            {func(cfg *GaugeConfig) { cfg.Liquid.PhaseDegrees = Float(math.Inf(1)) }, "liquid gauge phase must be finite and between -360 and 360 degrees"},
		"direction":        {func(cfg *GaugeConfig) { cfg.Liquid.Direction = "up" }, `liquid gauge direction "up" is not supported`},
		"outline paint":    {func(cfg *GaugeConfig) { cfg.Liquid.Outline = &GaugeLiquidOutline{Color: "red", Class: "text-red"} }, "liquid gauge outline requires at most one color or class"},
		"outline width":    {func(cfg *GaugeConfig) { cfg.Liquid.Outline = &GaugeLiquidOutline{Width: 65} }, "liquid gauge outline width must be finite and between 0 and 64 pixels"},
		"background paint": {func(cfg *GaugeConfig) { cfg.Liquid.Background = &GaugeLiquidBackground{Color: " "} }, "liquid gauge background color must not be blank"},
		"background border": {func(cfg *GaugeConfig) {
			cfg.Liquid.Background = &GaugeLiquidBackground{BorderColor: "red", BorderClass: "text-red"}
		}, "liquid gauge background border requires at most one color or class"},
		"label size":         {func(cfg *GaugeConfig) { cfg.Liquid.Label = &GaugeLiquidLabel{FontSize: 257} }, "liquid gauge label font size must be between 0 and 256 pixels"},
		"label paint":        {func(cfg *GaugeConfig) { cfg.Liquid.Label = &GaugeLiquidLabel{Color: "red", Class: "text-red"} }, "liquid gauge label requires at most one color or class"},
		"wave opacity":       {func(cfg *GaugeConfig) { cfg.Liquid.Style = &GaugeLiquidStyle{Opacity: Float(1.1)} }, "liquid gauge wave opacity must be finite and between 0 and 1"},
		"wave paint":         {func(cfg *GaugeConfig) { cfg.Liquid.Style = &GaugeLiquidStyle{Color: "red", Class: "text-red"} }, "liquid gauge wave requires at most one color or class"},
		"liquid on standard": {func(cfg *GaugeConfig) { cfg.Variant = GaugeVariantStandard; cfg.Liquid.Shape = GaugeLiquidShapePin }, "standard gauge treatment does not accept liquid options"},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cfg := valid()
			test.mutate(&cfg)
			var output bytes.Buffer
			err := Gauge(cfg).Render(context.Background(), &output)
			if err == nil || err.Error() != test.wantError {
				t.Fatalf("Render() error = %v, want %q", err, test.wantError)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %d bytes for invalid config", output.Len())
			}
		})
	}
}

func TestGaugeMapsReverseCustomAndSingleScale(t *testing.T) {
	t.Parallel()
	base := GaugeConfig{Label: "Gauge", Series: []GaugeSeries{{Name: "Sensor", Data: []GaugeData{{Name: "Reading", Value: 50}}}}}
	base.Scale = GaugeScale{Mode: GaugeScaleCustom, Reverse: true, Stops: []GaugeScaleStop{{Value: 40, Class: "text-cold"}, {Value: 100, Color: "#ff0000"}}}
	var output bytes.Buffer
	if err := Gauge(base).Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`&#34;reverse&#34;:true`, `&#34;position&#34;:0.4`, `&#34;class&#34;:&#34;text-cold&#34;`, `&#34;color&#34;:&#34;#ff0000&#34;`} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("custom scale missing %q", want)
		}
	}
	base.Scale = GaugeScale{Mode: GaugeScaleSingleColor, Class: "text-accent"}
	output.Reset()
	if err := Gauge(base).Render(context.Background(), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `&#34;mode&#34;:&#34;single-color&#34;`) || !strings.Contains(output.String(), `&#34;class&#34;:&#34;text-accent&#34;`) {
		t.Fatalf("single scale missing: %s", output.String())
	}
}
