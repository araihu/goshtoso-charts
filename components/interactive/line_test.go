package interactive

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"
	"time"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
)

func TestLineRendersConfiguredChart(t *testing.T) {
	t.Parallel()
	instance := Line(LineConfig{
		Label:   "Weekly traffic",
		Caption: "Requests by service.",
		XAxis:   []string{"Mon", "Tue"},
		Series: []LineSeries{
			{
				Name:    "API",
				Data:    []LineData{{Value: 12}, {Value: 18}},
				Options: SeriesOptions{Smooth: Bool(true)},
			},
		},
		Width:         "720px",
		Height:        "360px",
		Options:       ChartOptions{Title: &TitleOptions{Text: "Traffic"}},
		SeriesOptions: SeriesOptions{Label: &LabelOptions{Show: Bool(true)}},
	})

	if instance.Kind() != chartcomponents.KindInteractiveLine {
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

func TestLineRendersTypedTemporalAxisAndExactValues(t *testing.T) {
	t.Parallel()
	minimum := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)
	instance := Line(LineConfig{
		Label: "Temporal values", Caption: "UTC evidence.",
		TimeAxis: &LineTimeAxis{Minimum: minimum, Values: []time.Time{
			time.Date(2025, time.February, 0, 0, 0, 0, 0, time.FixedZone("other", -3*60*60)),
			time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC),
		}},
		Series: []LineSeries{{Name: "Category A", Data: []LineData{{Value: 107}, {Value: 118}}}},
		Options: ChartOptions{
			Title:   &TitleOptions{Text: "temporal X axis", Subtitle: "time.Date as X axis values"},
			Tooltip: &TooltipOptions{Show: Bool(true), Trigger: "axis"},
			YAxis:   &AxisOptions{Min: Float(0), Max: Float(200)},
		},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`"type":"time"`, `"min":"2025-01-01T00:00:00Z"`, `"splitNumber":4`, `"hideOverlap":true`, `"showMinLabel":true`, `"showMaxLabel":true`, `"min":0,"max":200`,
		`"trigger":"axis"`, `"value":["2025-01-31T03:00:00Z",107]`,
		"Exact time and values", "UTC timestamps.", "2025-01-31T03:00:00Z", "2025-02-01T00:00:00Z", ">118</td>",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered temporal markup missing %q", want)
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
				Series: []LineSeries{{Name: "API", Data: []LineData{{Value: 12}}}},
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
				Series: []LineSeries{{Data: []LineData{{Value: 12}}}},
			},
			wantError: "line chart series 0 name is required",
		},
		"misaligned series": {
			cfg: LineConfig{
				Label:  "Traffic",
				XAxis:  []string{"Mon", "Tue"},
				Series: []LineSeries{{Name: "API", Data: []LineData{{Value: 12}}}},
			},
			wantError: `line chart series "API" has 1 data points for 2 x-axis values`,
		},
		"nonfinite value": {
			cfg:       LineConfig{Label: "Traffic", XAxis: []string{"Mon"}, Series: []LineSeries{{Name: "API", Data: []LineData{{Value: math.Inf(1)}}}}},
			wantError: `line chart series "API" data point 0 value must be finite`,
		},
		"mixed axes": {
			cfg:       LineConfig{Label: "Traffic", XAxis: []string{"Mon"}, TimeAxis: &LineTimeAxis{Minimum: time.Now(), Values: []time.Time{time.Now()}}, Series: []LineSeries{{Name: "API", Data: []LineData{{Value: 12}}}}},
			wantError: "line chart x axis and time axis are mutually exclusive",
		},
		"missing time minimum": {
			cfg:       LineConfig{Label: "Traffic", TimeAxis: &LineTimeAxis{Values: []time.Time{time.Now()}}, Series: []LineSeries{{Name: "API", Data: []LineData{{Value: 12}}}}},
			wantError: "line chart time axis minimum is required",
		},
		"time before minimum": {
			cfg:       LineConfig{Label: "Traffic", TimeAxis: &LineTimeAxis{Minimum: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC), Values: []time.Time{time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)}}, Series: []LineSeries{{Name: "API", Data: []LineData{{Value: 12}}}}},
			wantError: "line chart time axis value 0 precedes minimum",
		},
		"duplicate time": {
			cfg:       LineConfig{Label: "Traffic", TimeAxis: &LineTimeAxis{Minimum: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), Values: []time.Time{time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)}}, Series: []LineSeries{{Name: "API", Data: []LineData{{Value: 12}, {Value: 13}}}}},
			wantError: "line chart time axis values must be strictly chronological",
		},
		"time live data": {
			cfg:       LineConfig{Label: "Traffic", TimeAxis: &LineTimeAxis{Minimum: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), Values: []time.Time{time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)}}, Live: &LiveData{URL: "/events"}, Series: []LineSeries{{Name: "API", Data: []LineData{{Value: 12}}}}},
			wantError: "line chart live data supports categorical x axis only",
		},
		"negative time split number": {
			cfg:       LineConfig{Label: "Traffic", TimeAxis: &LineTimeAxis{Minimum: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), Values: []time.Time{time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)}, SplitNumber: -1}, Series: []LineSeries{{Name: "API", Data: []LineData{{Value: 12}}}}},
			wantError: "line chart time axis split number must be nonnegative",
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
