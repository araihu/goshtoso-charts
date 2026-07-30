package boxplot

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func TestBoxPlotRendersMultipleTypedSeries(t *testing.T) {
	t.Parallel()
	instance := BoxPlot(Config{
		Label: "Latency distribution", Caption: "Five-number summaries by environment.",
		Categories: []string{"Development", "Production"},
		Series: []Series{
			{Name: "Current", Data: []Data{
				{Name: "dev-current", Min: 10, Q1: 20, Median: 30, Q3: 45, Max: 70},
				{Name: "prod-current", Min: 15, Q1: 25, Median: 35, Q3: 50, Max: 80, ItemStyle: &chart.ItemStyle{Color: "#abcdef"}},
			}},
			{Name: "Previous", Data: []Data{
				{Min: 12, Q1: 22, Median: 32, Q3: 48, Max: 75},
				{Min: 18, Q1: 28, Median: 40, Q3: 55, Max: 90},
			}, Options: chart.SeriesOptions{Label: &chart.LabelOptions{Show: chart.Bool(true)}}},
		},
		Width: "720px", Height: "360px",
		Options:       chart.ChartOptions{Title: &chart.TitleOptions{Text: "Latency"}},
		SeriesOptions: chart.SeriesOptions{ItemStyle: &chart.ItemStyle{BorderWidth: 2}},
		Style:         charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
	})

	if instance.Kind() != chartcomponents.KindInteractiveBoxPlot {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		"Latency distribution", "Five-number summaries by environment.", "width:720px;height:360px",
		`"Development","Production"`, `"name":"Current"`, `"name":"Previous"`,
		`"name":"dev-current","value":[10,20,30,45,70]`, `"value":[15,25,35,50,80]`,
		`"borderWidth":2`, `"show":true`, `"text":"Latency"`, `"color":["#123456","#ff8a3d"`,
		"goshtoso-charts-palette-araihu min-h-80",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestBoxPlotUsesFallbackPalette(t *testing.T) {
	t.Parallel()
	instance := BoxPlot(Config{
		Label: "Distribution", Categories: []string{"A"},
		Series: []Series{{Name: "Samples", Data: []Data{{Min: 1, Q1: 2, Median: 3, Q3: 4, Max: 5}}}},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	for _, want := range []string{
		`"color":["#2563eb","#dc2626"`,
		`"itemStyle":{"color":"#2563eb","borderColor":"#2563eb"}`,
		`data-goshtoso-charts-theme-series-items="0"`,
		"goshtoso-charts-palette-auto",
	} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestBoxPlotRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	validData := []Data{{Min: 1, Q1: 2, Median: 3, Q3: 4, Max: 5}}
	tests := map[string]struct {
		cfg       Config
		wantError string
	}{
		"missing label":       {cfg: Config{}, wantError: "box plot label is required"},
		"missing categories":  {cfg: Config{Label: "Distribution"}, wantError: "box plot categories are required"},
		"empty category":      {cfg: Config{Label: "Distribution", Categories: []string{""}}, wantError: "box plot category 0 name is required"},
		"missing series":      {cfg: Config{Label: "Distribution", Categories: []string{"A"}}, wantError: "box plot series is required"},
		"missing series name": {cfg: Config{Label: "Distribution", Categories: []string{"A"}, Series: []Series{{Data: validData}}}, wantError: "box plot series 0 name is required"},
		"missing data":        {cfg: Config{Label: "Distribution", Categories: []string{"A"}, Series: []Series{{Name: "Samples"}}}, wantError: `box plot series "Samples" data is required`},
		"category mismatch":   {cfg: Config{Label: "Distribution", Categories: []string{"A", "B"}, Series: []Series{{Name: "Samples", Data: validData}}}, wantError: `box plot series "Samples" has 1 summaries for 2 categories`},
		"nonfinite summary":   {cfg: Config{Label: "Distribution", Categories: []string{"A"}, Series: []Series{{Name: "Samples", Data: []Data{{Min: 1, Q1: 2, Median: math.NaN(), Q3: 4, Max: 5}}}}}, wantError: `box plot series "Samples" summary 0 values must be finite`},
		"unordered summary":   {cfg: Config{Label: "Distribution", Categories: []string{"A"}, Series: []Series{{Name: "Samples", Data: []Data{{Min: 1, Q1: 4, Median: 3, Q3: 5, Max: 6}}}}}, wantError: `box plot series "Samples" summary 0 values must be ordered min, q1, median, q3, max`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			err := BoxPlot(test.cfg).Render(context.Background(), &output)
			if err == nil || err.Error() != test.wantError {
				t.Fatalf("Render() error = %v, want %q", err, test.wantError)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %d bytes for invalid config", output.Len())
			}
		})
	}
}
