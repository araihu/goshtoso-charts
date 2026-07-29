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
		`"trigger":"axis"`, `"value":["2025-01-31T03:00:00Z",107]`, `data-line-time-exact-values`,
		"Exact time and values", "UTC timestamps.", "2025-01-31T03:00:00Z", "2025-02-01T00:00:00Z", ">118</td>",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered temporal markup missing %q", want)
		}
	}
}

func TestLineRendersTypedNumericalAxisAndPiecewiseScale(t *testing.T) {
	t.Parallel()
	instance := Line(LineConfig{
		Label: "Numerical observations", Caption: "Ordered numeric coordinates.",
		ValueAxis: &LineValueAxis{Values: []float64{0, 1, 2}},
		Series: []LineSeries{{
			Name:    "Category A",
			Data:    []LineData{{Value: 107}, {Value: 112}, {Value: 118}},
			Options: SeriesOptions{Symbol: "triangle", SymbolSize: 10, AreaStyle: &AreaStyle{}},
		}},
		Options: ChartOptions{YAxis: &AxisOptions{Max: Float(200)}},
		VisualScale: &LineVisualScale{
			Dimension: LineVisualDimensionX,
			Pieces: []LineVisualPiece{
				{GreaterThan: Float(0)},
				{LessThan: Float(0)},
				{GreaterThan: Float(1), LessThan: Float(7)},
				{GreaterThan: Float(10), LessThan: Float(15)},
			},
		},
	})

	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`"type":"value"`, `"value":[0,107]`, `"value":[2,118]`,
		`"visualMap":[{"type":"piecewise"`, `"dimension":"0"`, `"gt":0`, `"lt":0`, `"lt":7`, `"gt":1`, `"lt":15`, `"gt":10`,
		`"symbol":"triangle"`, `"symbolSize":10`, `"areaStyle":{}`,
		"Exact x and values", ">0</th>", ">118</td>",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered numerical markup missing %q", want)
		}
	}
}

func TestLineRendersStatisticalAndCoordinateReferences(t *testing.T) {
	t.Parallel()
	instance := Line(LineConfig{
		Label: "Annotated inventory",
		XAxis: []string{"Apple", "Banana", "Peach", "Lemon", "Pear", "Cherry"},
		Series: []LineSeries{{
			Name: "Category A",
			Data: []LineData{{Value: 120}, {Value: 132}, {Value: 101}, {Value: 134}, {Value: 90}, {Value: 230}},
			References: LineReferences{
				Points: []LinePointReference{
					{Name: "Maximum", Statistic: LineStatisticMaximum},
					{Name: "Average", Statistic: LineStatisticAverage},
					{Name: "Minimum", Statistic: LineStatisticMinimum},
				},
				Lines: []LineGuideReference{
					{Name: "Average", Statistic: LineStatisticAverage},
					{Name: "Danger level", Start: &LineCoordinate{X: 2, Y: 10}, End: &LineCoordinate{X: 4, Y: 50}},
					{Name: "Line of no return", X: Float(5)},
				},
				Areas:       []LineRangeReference{{Name: "In stock", StartX: 2, EndX: 4}},
				ShowLabels:  Bool(true),
				StartSymbol: "square",
				EndSymbol:   "circle",
				SymbolSize:  10,
			},
		}},
	})

	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`"markPoint":{"data":[{"name":"Maximum","type":"max"},{"name":"Average","type":"average"},{"name":"Minimum","type":"min"}]`,
		`"markLine":{"data":[{"name":"Average","type":"average"}`,
		`[{"name":"Danger level","coord":[2,10]},{"coord":[4,50]}]`,
		`{"name":"Line of no return","xAxis":5}`,
		`"symbol":["square","circle"]`, `"symbolSize":10`,
		`"markArea":{"data":[[{"name":"In stock","xAxis":2},{"xAxis":4}]]`,
		`"show":true`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered reference markup missing %q", want)
		}
	}
}

func TestLineRendersCategoricalExactValues(t *testing.T) {
	t.Parallel()
	instance := Line(LineConfig{
		Label: "Categorical values", XAxis: []string{"Apple", "Banana"},
		Series: []LineSeries{{Name: "Category A", Data: []LineData{{Value: 12}, {Value: 18}}}},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{"Exact category and values", "Apple", "Banana", "Category A", ">18</td>"} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered categorical evidence missing %q", want)
		}
	}
}

func TestLineDoesNotRenderStaleExactValuesForLiveData(t *testing.T) {
	t.Parallel()
	instance := Line(LineConfig{
		Label: "Live categorical values", XAxis: []string{"Mon"},
		Series: []LineSeries{{Name: "Category A", Data: []LineData{{Value: 12}}}},
		Live:   &LiveData{URL: "/events"},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if strings.Contains(output.String(), "data-line-exact-values") {
		t.Fatal("live Line rendered a static exact-value table that would become stale")
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
			wantError: "line chart category, time, and value axes are mutually exclusive",
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
		"unordered value axis": {
			cfg:       LineConfig{Label: "Traffic", ValueAxis: &LineValueAxis{Values: []float64{2, 1}}, Series: []LineSeries{{Name: "API", Data: []LineData{{Value: 12}, {Value: 13}}}}},
			wantError: "line chart value axis values must be strictly increasing",
		},
		"invalid series step": {
			cfg:       LineConfig{Label: "Traffic", XAxis: []string{"Mon"}, Series: []LineSeries{{Name: "API", Data: []LineData{{Value: 12}}, Options: SeriesOptions{Step: "late"}}}},
			wantError: `line chart series "API": step must be start, middle, or end`,
		},
		"invalid point symbol": {
			cfg:       LineConfig{Label: "Traffic", XAxis: []string{"Mon"}, Series: []LineSeries{{Name: "API", Data: []LineData{{Value: 12, Symbol: "star"}}}}},
			wantError: `line chart series "API" data point 0 symbol "star" is not supported`,
		},
		"mixed reference modes": {
			cfg: LineConfig{Label: "Traffic", XAxis: []string{"Mon"}, Series: []LineSeries{{
				Name: "API", Data: []LineData{{Value: 12}},
				References: LineReferences{Lines: []LineGuideReference{{Name: "Mixed", Statistic: LineStatisticAverage, X: Float(1)}}},
			}}},
			wantError: `line chart series "API": guide reference 0 requires exactly one reference mode`,
		},
		"invalid visual piece": {
			cfg: LineConfig{Label: "Traffic", XAxis: []string{"Mon"}, Series: []LineSeries{{Name: "API", Data: []LineData{{Value: 12}}}}, VisualScale: &LineVisualScale{
				Dimension: LineVisualDimensionX, Pieces: []LineVisualPiece{{GreaterThan: Float(7), LessThan: Float(1)}},
			}},
			wantError: "line chart visual scale piece 0 lower bound must be less than upper bound",
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
