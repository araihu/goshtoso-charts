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

func TestPieRendersDonutRoseChart(t *testing.T) {
	t.Parallel()
	instance := Pie(PieConfig{
		Label: "Incident states", Caption: "Incidents grouped by state.",
		Series: []PieSeries{{
			Name: "Incidents", InnerRadius: 30, OuterRadius: 70,
			RoseMode: PieRoseArea, PadAngle: 2,
			Data:    []PieData{{Name: "Open", Value: 12}, {Name: "Closed", Value: 28}},
			Options: SeriesOptions{Label: &LabelOptions{Show: Bool(true)}},
		}},
		Width: "720px", Height: "360px",
		Options:       ChartOptions{Title: &TitleOptions{Text: "Incident split"}},
		SeriesOptions: SeriesOptions{Animation: Bool(false)},
		Style:         charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
	})

	if instance.Kind() != chartcomponents.KindInteractivePie {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		"Incident states", "Incidents grouped by state.", "width:720px;height:360px",
		`"name":"Incidents"`, `"name":"Open","value":12`, `"radius":["30%","70%"]`,
		`"roseType":"area"`, `"padAngle":2`, `"show":true`, `"animation":false`,
		`"text":"Incident split"`, `"color":["#123456","#ff8a3d"`,
		"goshtoso-charts-palette-araihu min-h-80",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestPieUsesDefaultRadiusAndPalette(t *testing.T) {
	t.Parallel()
	instance := Pie(PieConfig{
		Label: "Traffic", Series: []PieSeries{{
			Name: "Sources", Data: []PieData{{Name: "Direct", Value: 4}},
		}},
		Style: charttheme.Style{Palette: charttheme.PalettePastel},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{`"radius":["0%","75%"]`, `"color":["#93c5fd","#fca5a5"`} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestPieRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	validSeries := PieSeries{Name: "States", Data: []PieData{{Name: "Open", Value: 1}}}
	tests := map[string]struct {
		cfg       PieConfig
		wantError string
	}{
		"missing label":       {PieConfig{Series: []PieSeries{validSeries}}, "pie chart label is required"},
		"missing series":      {PieConfig{Label: "States"}, "pie chart series is required"},
		"missing series name": {PieConfig{Label: "States", Series: []PieSeries{{Data: validSeries.Data}}}, "pie chart series 0 name is required"},
		"missing data":        {PieConfig{Label: "States", Series: []PieSeries{{Name: "States"}}}, `pie chart series "States" data is required`},
		"bad inner radius":    {PieConfig{Label: "States", Series: []PieSeries{{Name: "States", InnerRadius: -1, Data: validSeries.Data}}}, `pie chart series "States" inner radius must be between 0 and 100`},
		"bad outer radius":    {PieConfig{Label: "States", Series: []PieSeries{{Name: "States", OuterRadius: 101, Data: validSeries.Data}}}, `pie chart series "States" outer radius must be between 0 and 100`},
		"crossed radii":       {PieConfig{Label: "States", Series: []PieSeries{{Name: "States", InnerRadius: 60, OuterRadius: 50, Data: validSeries.Data}}}, `pie chart series "States" inner radius must be less than outer radius`},
		"bad rose mode":       {PieConfig{Label: "States", Series: []PieSeries{{Name: "States", RoseMode: "flower", Data: validSeries.Data}}}, `pie chart series "States" rose mode "flower" is not supported`},
		"negative pad":        {PieConfig{Label: "States", Series: []PieSeries{{Name: "States", PadAngle: -1, Data: validSeries.Data}}}, `pie chart series "States" pad angle must be a finite nonnegative value`},
		"missing data name":   {PieConfig{Label: "States", Series: []PieSeries{{Name: "States", Data: []PieData{{Value: 1}}}}}, `pie chart series "States" data point 0 name is required`},
		"negative value":      {PieConfig{Label: "States", Series: []PieSeries{{Name: "States", Data: []PieData{{Name: "Open", Value: -1}}}}}, `pie chart series "States" data point "Open" value must be a finite nonnegative value`},
		"nonfinite value":     {PieConfig{Label: "States", Series: []PieSeries{{Name: "States", Data: []PieData{{Name: "Open", Value: math.NaN()}}}}}, `pie chart series "States" data point "Open" value must be a finite nonnegative value`},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			err := Pie(test.cfg).Render(context.Background(), &output)
			if err == nil || err.Error() != test.wantError {
				t.Fatalf("Render() error = %v, want %q", err, test.wantError)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %d bytes for invalid config", output.Len())
			}
		})
	}
}
