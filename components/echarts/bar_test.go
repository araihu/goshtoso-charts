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

func TestBarRendersConfiguredChart(t *testing.T) {
	t.Parallel()
	instance := Bar(BarConfig{
		Label:   "Quarterly revenue",
		Caption: "Revenue by product.",
		XAxis:   []string{"Q1", "Q2"},
		Series: []BarSeries{
			{
				Name: "Hardware",
				Data: []opts.BarData{{Value: 12}, {Value: 18}},
				Options: []charts.SeriesOpts{
					charts.WithBarChartOpts(opts.BarChart{Stack: "revenue"}),
				},
			},
		},
		Width:  "720px",
		Height: "360px",
		GlobalOptions: []charts.GlobalOpts{
			charts.WithTitleOpts(opts.Title{Title: "Revenue"}),
		},
		SeriesOptions: []charts.SeriesOpts{
			charts.WithLabelOpts(opts.Label{Show: opts.Bool(true)}),
		},
	})

	if instance.Kind() != chartcomponents.KindEChartsBar {
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
		`"text":"Revenue"`,
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
				Series: []BarSeries{{Name: "Hardware", Data: []opts.BarData{{Value: 12}}}},
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
				Series: []BarSeries{{Data: []opts.BarData{{Value: 12}}}},
			},
			wantError: "bar chart series 0 name is required",
		},
		"misaligned series": {
			cfg: BarConfig{
				Label:  "Revenue",
				XAxis:  []string{"Q1", "Q2"},
				Series: []BarSeries{{Name: "Hardware", Data: []opts.BarData{{Value: 12}}}},
			},
			wantError: `bar chart series "Hardware" has 1 values for 2 x-axis categories`,
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
