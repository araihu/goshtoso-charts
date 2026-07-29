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

func TestPieRendersDonutRoseChart(t *testing.T) {
	t.Parallel()
	instance := Pie(PieConfig{
		Label: "Incident states", Caption: "Incidents grouped by state.",
		Series: []PieSeries{{
			Name: "Incidents", InnerRadius: 30, OuterRadius: 70,
			RoseMode: PieRoseArea, PadAngle: 2, Center: &PieCenter{X: 25, Y: 50},
			Selectable:   true,
			LabelContent: PieLabelNameAndValue,
			Data:         []PieData{{Name: "Open", Value: 12, Selected: true}, {Name: "Closed", Value: 28}},
			Options: SeriesOptions{
				Label:    &LabelOptions{Show: Bool(true)},
				Emphasis: &EmphasisOptions{ItemStyle: &ItemStyle{ShadowBlur: 10, ShadowColor: "rgba(0,0,0,0.5)"}},
			},
		}},
		Width: "720px", Height: "360px",
		Options: ChartOptions{
			Title:  &TitleOptions{Text: "Incident split"},
			Legend: &LegendOptions{Orient: "vertical", Right: "20%", Padding: &EdgeInsets{Top: 1, Right: 2, Bottom: 3, Left: 4}},
		},
		TooltipContent: PieTooltipNameAndShare,
		AutoEmphasis:   &PieAutoEmphasisOptions{SeriesIndex: 0, IntervalMilliseconds: 1250},
		SeriesOptions:  SeriesOptions{Animation: Bool(false)},
		Style:          charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
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
		`"roseType":"area"`, `"padAngle":2`, `"center":["25%","50%"]`,
		`"formatter":"{b}: {c}"`, `"formatter":"{b}: {d}%"`, `"show":true`, `"animation":false`,
		`"selectedMode":true`, `"selected":true`, `"padding":[1,2,3,4]`,
		`"shadowBlur":10`, `"shadowColor":"rgba(0,0,0,0.5)"`,
		`data-goshtoso-charts-pie-auto-emphasis="{&#34;seriesIndex&#34;:0,&#34;interval&#34;:1250,&#34;showTooltip&#34;:true}"`,
		`"text":"Incident split"`, `"color":["#123456","#ff8a3d"`,
		"goshtoso-charts-palette-araihu min-h-80", "Exact pie values", "Open", "12", "30%",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestPiePreservesEverySharedWrapperMode(t *testing.T) {
	t.Parallel()
	tests := map[chartcontrol.WrapperMode][]string{
		chartcontrol.WrapperModeEnabled:  {`data-goshtoso-chart-wrapper-mode="enabled"`},
		chartcontrol.WrapperModeDisabled: {`data-goshtoso-chart-wrapper-mode="disabled"`, `data-goshtoso-chart-actions-fieldset disabled aria-disabled="true"`},
		chartcontrol.WrapperModeHidden:   {`data-goshtoso-chart-wrapper-mode="hidden"`, `hidden inert aria-hidden="true"`},
		chartcontrol.WrapperModeOmitted:  {`goshtoso-charts-interactive`, `Exact pie values`},
	}
	for mode, wants := range tests {
		mode, wants := mode, wants
		t.Run(string(mode), func(t *testing.T) {
			t.Parallel()
			instance := Pie(PieConfig{
				Label: "Wrapper Pie", Series: []PieSeries{{Name: "Series", Data: []PieData{{Name: "A", Value: 1}}}},
				Options: ChartOptions{Controls: chartcontrol.Options{Mode: mode}},
			})
			var output bytes.Buffer
			if err := instance.Render(context.Background(), &output); err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			markup := output.String()
			for _, want := range wants {
				if !strings.Contains(markup, want) {
					t.Errorf("mode %q markup missing %q", mode, want)
				}
			}
			if mode == chartcontrol.WrapperModeOmitted && strings.Contains(markup, `<div class="goshtoso-charts-control-wrapper"`) {
				t.Error("omitted mode rendered wrapper DOM")
			}
		})
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
		"bad center x":        {PieConfig{Label: "States", Series: []PieSeries{{Name: "States", Center: &PieCenter{X: -1, Y: 50}, Data: validSeries.Data}}}, `pie chart series "States" center x must be between 0 and 100`},
		"bad center y":        {PieConfig{Label: "States", Series: []PieSeries{{Name: "States", Center: &PieCenter{X: 50, Y: 101}, Data: validSeries.Data}}}, `pie chart series "States" center y must be between 0 and 100`},
		"bad label content":   {PieConfig{Label: "States", Series: []PieSeries{{Name: "States", LabelContent: "raw", Data: validSeries.Data}}}, `pie chart series "States" label content "raw" is not supported`},
		"negative pad":        {PieConfig{Label: "States", Series: []PieSeries{{Name: "States", PadAngle: -1, Data: validSeries.Data}}}, `pie chart series "States" pad angle must be a finite nonnegative value`},
		"bad tooltip content": {PieConfig{Label: "States", TooltipContent: "raw", Series: []PieSeries{validSeries}}, `pie chart tooltip content "raw" is not supported`},
		"bad legend padding":  {PieConfig{Label: "States", Options: ChartOptions{Legend: &LegendOptions{Padding: &EdgeInsets{Left: -1}}}, Series: []PieSeries{validSeries}}, `legend padding must be nonnegative`},
		"bad emphasis index":  {PieConfig{Label: "States", AutoEmphasis: &PieAutoEmphasisOptions{SeriesIndex: 1}, Series: []PieSeries{validSeries}}, `pie chart auto emphasis series index must identify a configured series`},
		"bad emphasis delay":  {PieConfig{Label: "States", AutoEmphasis: &PieAutoEmphasisOptions{IntervalMilliseconds: -1}, Series: []PieSeries{validSeries}}, `pie chart auto emphasis interval must be nonnegative`},
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

func TestPieRendersNestedSeriesAtOneCenter(t *testing.T) {
	t.Parallel()
	instance := Pie(PieConfig{
		Label: "Nested seasons",
		Series: []PieSeries{
			{Name: "Outer", InnerRadius: 50, OuterRadius: 55, RoseMode: PieRoseArea, Data: []PieData{{Name: "Spring", Value: 81}}},
			{Name: "Inner", OuterRadius: 45, RoseMode: PieRoseRadius, Data: []PieData{{Name: "Spring", Value: 56}}},
		},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{`"name":"Outer"`, `"radius":["50%","55%"]`, `"roseType":"area"`, `"name":"Inner"`, `"radius":["0%","45%"]`, `"roseType":"radius"`} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	if strings.Contains(markup, `"center"`) {
		t.Error("default-center nested series must preserve renderer default center")
	}
}
