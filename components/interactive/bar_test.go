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

func TestBarRendersConfiguredChart(t *testing.T) {
	t.Parallel()
	instance := Bar(BarConfig{
		Label:   "Quarterly revenue",
		Caption: "Revenue by product.",
		XAxis:   []string{"Q1", "Q2"},
		Series: []BarSeries{
			{
				Name:    "Hardware",
				Data:    []BarData{{Value: 12}, {Value: 18}},
				Options: SeriesOptions{Stack: "revenue"},
			},
		},
		Width:  "720px",
		Height: "360px",
		Options: ChartOptions{
			Title:     &TitleOptions{Text: "Revenue"},
			XAxis:     &AxisOptions{LabelInterval: Int(5), ShowFirstLabel: Bool(true), ShowLastLabel: Bool(true)},
			YAxis:     &AxisOptions{Min: Float(0), Max: Float(20), Show: Bool(false)},
			Animation: Bool(false),
		},
		SeriesOptions: SeriesOptions{Label: &LabelOptions{Show: Bool(true)}},
		Style:         charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
	})

	if instance.Kind() != chartcomponents.KindInteractiveBar {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		"Quarterly revenue",
		"Revenue by product.",
		"width:720px;height:360px",
		`"Q1","Q2"`,
		`"name":"Hardware"`,
		`"stack":"revenue"`,
		`"show":true`,
		`data-goshtoso-charts-explicit-animation="false"`,
		`"yAxis":[{"show":false,"min":0,"max":20}]`,
		`"axisLabel":{"interval":5,"showMinLabel":true,"showMaxLabel":true}`,
		`"text":"Revenue"`,
		`"color":["#123456","#ff8a3d"`,
		"goshtoso-charts-palette-araihu min-h-80",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestBarRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		cfg       BarConfig
		wantError string
	}{
		"missing x axis": {
			cfg: BarConfig{
				Label:  "Revenue",
				Series: []BarSeries{{Name: "Hardware", Data: []BarData{{Value: 12}}}},
			},
			wantError: "bar chart x axis is required",
		},
		"missing series": {
			cfg:       BarConfig{Label: "Revenue", XAxis: []string{"Q1"}},
			wantError: "bar chart series is required",
		},
		"missing series name": {
			cfg: BarConfig{
				Label:  "Revenue",
				XAxis:  []string{"Q1"},
				Series: []BarSeries{{Data: []BarData{{Value: 12}}}},
			},
			wantError: "bar chart series 0 name is required",
		},
		"misaligned series": {
			cfg: BarConfig{
				Label:  "Revenue",
				XAxis:  []string{"Q1", "Q2"},
				Series: []BarSeries{{Name: "Hardware", Data: []BarData{{Value: 12}}}},
			},
			wantError: `bar chart series "Hardware" has 1 values for 2 x-axis categories`,
		},
		"nonfinite value": {
			cfg:       BarConfig{Label: "Revenue", XAxis: []string{"Q1"}, Series: []BarSeries{{Name: "Hardware", Data: []BarData{{Value: math.NaN()}}}}},
			wantError: `bar chart series "Hardware" data point 0 value must be finite`,
		},
		"negative label interval": {
			cfg: BarConfig{
				Label: "Revenue", XAxis: []string{"Q1"}, Series: []BarSeries{{Name: "Hardware", Data: []BarData{{Value: 12}}}},
				Options: ChartOptions{XAxis: &AxisOptions{LabelInterval: Int(-1)}},
			},
			wantError: "x axis label interval must be nonnegative",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			err := Bar(test.cfg).Render(context.Background(), &output)
			if err == nil {
				t.Fatal("Render() error = nil")
			}
			if err.Error() != test.wantError {
				t.Fatalf("Render() error = %q, want %q", err, test.wantError)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %d bytes for invalid config", output.Len())
			}
		})
	}
}
