package echarts

import (
	"bytes"
	"context"
	"strings"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func TestLineRendersConfiguredChart(t *testing.T) {
	t.Parallel()
	instance := Line(LineConfig{
		Label:   "Weekly traffic",
		Caption: "Requests by service.",
		XAxis:   []string{"Mon", "Tue"},
		Series: []LineSeries{
			{
				Name: "API",
				Data: []opts.LineData{{Value: 12}, {Value: 18}},
				Options: []charts.SeriesOpts{
					charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}),
				},
			},
		},
		Width:  "720px",
		Height: "360px",
		GlobalOptions: []charts.GlobalOpts{
			charts.WithTitleOpts(opts.Title{Title: "Traffic"}),
		},
		SeriesOptions: []charts.SeriesOpts{
			charts.WithLabelOpts(opts.Label{Show: opts.Bool(true)}),
		},
	})

	if instance.Kind() != chartcomponents.KindEChartsLine {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		"Weekly traffic",
		"Requests by service.",
		"width:720px;height:360px",
		`"Mon","Tue"`,
		`"name":"API"`,
		`"smooth":true`,
		`"show":true`,
		`"text":"Traffic"`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestLineRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		cfg       LineConfig
		wantError string
	}{
		"missing x axis": {
			cfg: LineConfig{
				Label:  "Traffic",
				Series: []LineSeries{{Name: "API", Data: []opts.LineData{{Value: 12}}}},
			},
			wantError: "line chart x axis is required",
		},
		"missing series": {
			cfg:       LineConfig{Label: "Traffic", XAxis: []string{"Mon"}},
			wantError: "line chart series is required",
		},
		"missing series name": {
			cfg: LineConfig{
				Label:  "Traffic",
				XAxis:  []string{"Mon"},
				Series: []LineSeries{{Data: []opts.LineData{{Value: 12}}}},
			},
			wantError: "line chart series 0 name is required",
		},
		"misaligned series": {
			cfg: LineConfig{
				Label:  "Traffic",
				XAxis:  []string{"Mon", "Tue"},
				Series: []LineSeries{{Name: "API", Data: []opts.LineData{{Value: 12}}}},
			},
			wantError: `line chart series "API" has 1 data points for 2 x-axis values`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			err := Line(test.cfg).Render(context.Background(), &output)
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
