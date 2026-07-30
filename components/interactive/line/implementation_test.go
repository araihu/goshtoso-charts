package line

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"
	"time"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
)

type ChartOptions = chart.ChartOptions
type SeriesOptions = chart.SeriesOptions
type TitleOptions = chart.TitleOptions
type LabelOptions = chart.LabelOptions
type AreaStyle = chart.AreaStyle
type AxisOptions = chart.AxisOptions
type TooltipOptions = chart.TooltipOptions
type LiveData = chart.LiveData

var Bool = chart.Bool
var Float = chart.Float

func TestLineRendersConfiguredChart(t *testing.T) {
	t.Parallel()
	instance := Line(Config{
		Label:   "Weekly traffic",
		Caption: "Requests by service.",
		XAxis:   []string{"Mon", "Tue"},
		Series: []Series{
			{
				Name:    "API",
				Data:    []Data{{Value: 12}, {Value: 18}},
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
	instance := Line(Config{
		Label: "Temporal values", Caption: "UTC evidence.",
		TimeAxis: &TimeAxis{Minimum: minimum, Values: []time.Time{
			time.Date(2025, time.February, 0, 0, 0, 0, 0, time.FixedZone("other", -3*60*60)),
			time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC),
		}},
		Series: []Series{{Name: "Category A", Data: []Data{{Value: 107}, {Value: 118}}}},
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
	instance := Line(Config{
		Label: "Numerical observations", Caption: "Ordered numeric coordinates.",
		ValueAxis: &ValueAxis{Values: []float64{0, 1, 2}},
		Series: []Series{{
			Name:    "Category A",
			Data:    []Data{{Value: 107}, {Value: 112}, {Value: 118}},
			Options: SeriesOptions{Symbol: "triangle", SymbolSize: 10, AreaStyle: &AreaStyle{}},
		}},
		Options: ChartOptions{YAxis: &AxisOptions{Max: Float(200)}},
		VisualScale: &VisualScale{
			Dimension: VisualDimensionX,
			Pieces: []VisualPiece{
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
	instance := Line(Config{
		Label: "Annotated inventory",
		XAxis: []string{"Apple", "Banana", "Peach", "Lemon", "Pear", "Cherry"},
		Series: []Series{{
			Name: "Category A",
			Data: []Data{{Value: 120}, {Value: 132}, {Value: 101}, {Value: 134}, {Value: 90}, {Value: 230}},
			References: References{
				Points: []PointReference{
					{Name: "Maximum", Statistic: StatisticMaximum},
					{Name: "Average", Statistic: StatisticAverage},
					{Name: "Minimum", Statistic: StatisticMinimum},
				},
				Lines: []GuideReference{
					{Name: "Average", Statistic: StatisticAverage},
					{Name: "Danger level", Start: &Coordinate{X: 2, Y: 10}, End: &Coordinate{X: 4, Y: 50}},
					{Name: "Line of no return", X: Float(5)},
				},
				Areas:       []RangeReference{{Name: "In stock", StartX: 2, EndX: 4}},
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
	instance := Line(Config{
		Label: "Categorical values", XAxis: []string{"Apple", "Banana"},
		Series: []Series{{Name: "Category A", Data: []Data{{Value: 12}, {Value: 18}}}},
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
	instance := Line(Config{
		Label: "Live categorical values", XAxis: []string{"Mon"},
		Series: []Series{{Name: "Category A", Data: []Data{{Value: 12}}}},
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
		cfg       Config
		wantError string
	}{
		"missing x axis": {
			cfg: Config{
				Label:  "Traffic",
				Series: []Series{{Name: "API", Data: []Data{{Value: 12}}}},
			},
			wantError: "line chart x axis is required",
		},
		"missing series": {
			cfg:       Config{Label: "Traffic", XAxis: []string{"Mon"}},
			wantError: "line chart series is required",
		},
		"missing series name": {
			cfg: Config{
				Label:  "Traffic",
				XAxis:  []string{"Mon"},
				Series: []Series{{Data: []Data{{Value: 12}}}},
			},
			wantError: "line chart series 0 name is required",
		},
		"misaligned series": {
			cfg: Config{
				Label:  "Traffic",
				XAxis:  []string{"Mon", "Tue"},
				Series: []Series{{Name: "API", Data: []Data{{Value: 12}}}},
			},
			wantError: `line chart series "API" has 1 data points for 2 x-axis values`,
		},
		"nonfinite value": {
			cfg:       Config{Label: "Traffic", XAxis: []string{"Mon"}, Series: []Series{{Name: "API", Data: []Data{{Value: math.Inf(1)}}}}},
			wantError: `line chart series "API" data point 0 value must be finite`,
		},
		"mixed axes": {
			cfg:       Config{Label: "Traffic", XAxis: []string{"Mon"}, TimeAxis: &TimeAxis{Minimum: time.Now(), Values: []time.Time{time.Now()}}, Series: []Series{{Name: "API", Data: []Data{{Value: 12}}}}},
			wantError: "line chart category, time, and value axes are mutually exclusive",
		},
		"missing time minimum": {
			cfg:       Config{Label: "Traffic", TimeAxis: &TimeAxis{Values: []time.Time{time.Now()}}, Series: []Series{{Name: "API", Data: []Data{{Value: 12}}}}},
			wantError: "line chart time axis minimum is required",
		},
		"time before minimum": {
			cfg:       Config{Label: "Traffic", TimeAxis: &TimeAxis{Minimum: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC), Values: []time.Time{time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)}}, Series: []Series{{Name: "API", Data: []Data{{Value: 12}}}}},
			wantError: "line chart time axis value 0 precedes minimum",
		},
		"duplicate time": {
			cfg:       Config{Label: "Traffic", TimeAxis: &TimeAxis{Minimum: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), Values: []time.Time{time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)}}, Series: []Series{{Name: "API", Data: []Data{{Value: 12}, {Value: 13}}}}},
			wantError: "line chart time axis values must be strictly chronological",
		},
		"time live data": {
			cfg:       Config{Label: "Traffic", TimeAxis: &TimeAxis{Minimum: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), Values: []time.Time{time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)}}, Live: &LiveData{URL: "/events"}, Series: []Series{{Name: "API", Data: []Data{{Value: 12}}}}},
			wantError: "line chart live data supports categorical x axis only",
		},
		"negative time split number": {
			cfg:       Config{Label: "Traffic", TimeAxis: &TimeAxis{Minimum: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), Values: []time.Time{time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)}, SplitNumber: -1}, Series: []Series{{Name: "API", Data: []Data{{Value: 12}}}}},
			wantError: "line chart time axis split number must be nonnegative",
		},
		"unordered value axis": {
			cfg:       Config{Label: "Traffic", ValueAxis: &ValueAxis{Values: []float64{2, 1}}, Series: []Series{{Name: "API", Data: []Data{{Value: 12}, {Value: 13}}}}},
			wantError: "line chart value axis values must be strictly increasing",
		},
		"invalid series step": {
			cfg:       Config{Label: "Traffic", XAxis: []string{"Mon"}, Series: []Series{{Name: "API", Data: []Data{{Value: 12}}, Options: SeriesOptions{Step: "late"}}}},
			wantError: `line chart series "API": step must be start, middle, or end`,
		},
		"invalid point symbol": {
			cfg:       Config{Label: "Traffic", XAxis: []string{"Mon"}, Series: []Series{{Name: "API", Data: []Data{{Value: 12, Symbol: "star"}}}}},
			wantError: `line chart series "API" data point 0 symbol "star" is not supported`,
		},
		"mixed reference modes": {
			cfg: Config{Label: "Traffic", XAxis: []string{"Mon"}, Series: []Series{{
				Name: "API", Data: []Data{{Value: 12}},
				References: References{Lines: []GuideReference{{Name: "Mixed", Statistic: StatisticAverage, X: Float(1)}}},
			}}},
			wantError: `line chart series "API": guide reference 0 requires exactly one reference mode`,
		},
		"invalid visual piece": {
			cfg: Config{Label: "Traffic", XAxis: []string{"Mon"}, Series: []Series{{Name: "API", Data: []Data{{Value: 12}}}}, VisualScale: &VisualScale{
				Dimension: VisualDimensionX, Pieces: []VisualPiece{{GreaterThan: Float(7), LessThan: Float(1)}},
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
