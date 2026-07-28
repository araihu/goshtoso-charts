package echarts

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func TestBoxPlotRendersMultipleTypedSeries(t *testing.T) {
	t.Parallel()
	instance := BoxPlot(BoxPlotConfig{
		Label: "Latency distribution", Caption: "Five-number summaries by environment.",
		Categories: []string{"Development", "Production"},
		Series: []BoxPlotSeries{
			{Name: "Current", Data: []BoxPlotData{
				{Name: "dev-current", Min: 10, Q1: 20, Median: 30, Q3: 45, Max: 70},
				{Name: "prod-current", Min: 15, Q1: 25, Median: 35, Q3: 50, Max: 80, ItemStyle: &opts.ItemStyle{Color: "#abcdef"}},
			}},
			{Name: "Previous", Data: []BoxPlotData{
				{Min: 12, Q1: 22, Median: 32, Q3: 48, Max: 75},
				{Min: 18, Q1: 28, Median: 40, Q3: 55, Max: 90},
			}, Options: []charts.SeriesOpts{charts.WithLabelOpts(opts.Label{Show: opts.Bool(true)})}},
		},
		Width: "720px", Height: "360px",
		GlobalOptions: []charts.GlobalOpts{charts.WithTitleOpts(opts.Title{Title: "Latency"}), charts.WithColorsOpts(opts.Colors{"#000000"})},
		SeriesOptions: []charts.SeriesOpts{charts.WithItemStyleOpts(opts.ItemStyle{BorderWidth: 2})},
		Style:         charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
	})

	if instance.Kind() != chartcomponents.KindEChartsBoxPlot {
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
	if strings.Contains(markup, "#000000") {
		t.Error("explicit Style.Colors did not override global color options")
	}
}

func TestBoxPlotUsesFallbackPalette(t *testing.T) {
	t.Parallel()
	instance := BoxPlot(BoxPlotConfig{
		Label: "Distribution", Categories: []string{"A"},
		Series: []BoxPlotSeries{{Name: "Samples", Data: []BoxPlotData{{Min: 1, Q1: 2, Median: 3, Q3: 4, Max: 5}}}},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	for _, want := range []string{`"color":["#2563eb","#dc2626"`, "goshtoso-charts-palette-auto"} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestBoxPlotRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	validData := []BoxPlotData{{Min: 1, Q1: 2, Median: 3, Q3: 4, Max: 5}}
	tests := map[string]struct {
		cfg       BoxPlotConfig
		wantError string
	}{
		"missing label":       {cfg: BoxPlotConfig{}, wantError: "box plot label is required"},
		"missing categories":  {cfg: BoxPlotConfig{Label: "Distribution"}, wantError: "box plot categories are required"},
		"empty category":      {cfg: BoxPlotConfig{Label: "Distribution", Categories: []string{""}}, wantError: "box plot category 0 name is required"},
		"missing series":      {cfg: BoxPlotConfig{Label: "Distribution", Categories: []string{"A"}}, wantError: "box plot series is required"},
		"missing series name": {cfg: BoxPlotConfig{Label: "Distribution", Categories: []string{"A"}, Series: []BoxPlotSeries{{Data: validData}}}, wantError: "box plot series 0 name is required"},
		"missing data":        {cfg: BoxPlotConfig{Label: "Distribution", Categories: []string{"A"}, Series: []BoxPlotSeries{{Name: "Samples"}}}, wantError: `box plot series "Samples" data is required`},
		"category mismatch":   {cfg: BoxPlotConfig{Label: "Distribution", Categories: []string{"A", "B"}, Series: []BoxPlotSeries{{Name: "Samples", Data: validData}}}, wantError: `box plot series "Samples" has 1 summaries for 2 categories`},
		"nonfinite summary":   {cfg: BoxPlotConfig{Label: "Distribution", Categories: []string{"A"}, Series: []BoxPlotSeries{{Name: "Samples", Data: []BoxPlotData{{Min: 1, Q1: 2, Median: math.NaN(), Q3: 4, Max: 5}}}}}, wantError: `box plot series "Samples" summary 0 values must be finite`},
		"unordered summary":   {cfg: BoxPlotConfig{Label: "Distribution", Categories: []string{"A"}, Series: []BoxPlotSeries{{Name: "Samples", Data: []BoxPlotData{{Min: 1, Q1: 4, Median: 3, Q3: 5, Max: 6}}}}}, wantError: `box plot series "Samples" summary 0 values must be ordered min, q1, median, q3, max`},
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
