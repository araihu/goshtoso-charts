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
				Options: SeriesOptions{Stack: "revenue", BarWidth: "35"},
				References: BarReferences{
					Points: []BarPointReference{
						{Name: "Maximum", Statistic: BarStatisticMaximum},
						{Name: "Q1 target", Coordinate: &BarCoordinate{Category: "Q1", Value: 12}, Label: &LabelOptions{Show: Bool(true), Position: "inside"}},
					},
					Lines:      []BarGuideReference{{Name: "Average", Statistic: BarStatisticAverage}},
					ShowLabels: Bool(true),
				},
			},
		},
		Orientation: BarOrientationHorizontal,
		Zoom:        &BarZoom{Mode: BarZoomSlider, StartPercent: 10, EndPercent: 80},
		Width:       "720px",
		Height:      "360px",
		Options: ChartOptions{
			Title:     &TitleOptions{Text: "Revenue"},
			XAxis:     &AxisOptions{LabelInterval: Int(5), ShowFirstLabel: Bool(true), ShowLastLabel: Bool(true), LabelPrefix: "$", LabelSuffix: " total"},
			YAxis:     &AxisOptions{Min: Float(0), Max: Float(20), Show: Bool(false)},
			Animation: Bool(false),
		},
		SeriesOptions: SeriesOptions{Label: &LabelOptions{Show: Bool(true)}, BarGap: "150%"},
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
		`"barWidth":"35"`,
		`"barGap":"150%"`,
		`"dataZoom":[{"type":"slider","start":10,"end":80,"yAxisIndex":0}]`,
		`"formatter":"${value} total"`,
		`"markPoint":{"data":[{"name":"Maximum","type":"max"},{"name":"Q1 target","coord":["Q1",12]`,
		`"markLine":{"data":[{"name":"Average","type":"average"}]`,
		`"markLine":{"data":[{"name":"Average","type":"average"}],"label":{"show":true}}`,
		`"show":true`,
		`data-goshtoso-charts-explicit-animation="false"`,
		`"yAxis":[{"show":false,`,
		`"min":0,"max":20`,
		`"axisLabel":{"interval":5,`,
		`"showMinLabel":true,"showMaxLabel":true`,
		`"text":"Revenue"`,
		`"color":["#123456","#ff8a3d"`,
		"goshtoso-charts-palette-araihu min-h-80",
		`data-bar-exact-values`,
		`Quarterly revenue exact category values`,
		`<th class="px-3 py-2 font-semibold" scope="row">Q1</th><td class="px-3 py-2">Hardware</td><td class="px-3 py-2 tabular-nums">12</td>`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestInteractiveBarPropagatesSharedWrapperLifecycle(t *testing.T) {
	t.Parallel()
	base := BarConfig{
		Label: "Revenue", XAxis: []string{"Q1"},
		Series: []BarSeries{{Name: "Hardware", Data: []BarData{{Value: 12}}}},
	}
	base.Options.Controls.Mode = chartcontrol.WrapperModeDisabled
	var disabled bytes.Buffer
	if err := Bar(base).Render(context.Background(), &disabled); err != nil {
		t.Fatalf("disabled Render() error = %v", err)
	}
	if !strings.Contains(disabled.String(), `data-goshtoso-chart-wrapper-mode="disabled"`) ||
		!strings.Contains(disabled.String(), `data-goshtoso-chart-actions-fieldset disabled aria-disabled="true"`) ||
		!strings.Contains(disabled.String(), `_echarts_instance_`) {
		t.Fatalf("interactive chart did not propagate disabled wrapper mode")
	}

	base.Options.Controls.Mode = chartcontrol.WrapperModeOmitted
	var omitted bytes.Buffer
	if err := Bar(base).Render(context.Background(), &omitted); err != nil {
		t.Fatalf("omitted Render() error = %v", err)
	}
	if strings.Contains(omitted.String(), `class="goshtoso-charts-control-wrapper"`) ||
		!strings.Contains(omitted.String(), `goshtoso-charts-interactive`) ||
		!strings.Contains(omitted.String(), `_echarts_instance_`) {
		t.Fatalf("interactive chart did not propagate omitted wrapper mode")
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
		"renderer placeholder in label suffix": {
			cfg: BarConfig{
				Label: "Revenue", XAxis: []string{"Q1"}, Series: []BarSeries{{Name: "Hardware", Data: []BarData{{Value: 12}}}},
				Options: ChartOptions{YAxis: &AxisOptions{LabelSuffix: " {series}"}},
			},
			wantError: "y axis label prefix and suffix must be literal text without braces",
		},
		"invalid orientation": {
			cfg: BarConfig{
				Label: "Revenue", XAxis: []string{"Q1"}, Series: []BarSeries{{Name: "Hardware", Data: []BarData{{Value: 12}}}},
				Orientation: BarOrientation("diagonal"),
			},
			wantError: `bar chart orientation "diagonal" is not supported`,
		},
		"invalid zoom mode": {
			cfg: BarConfig{
				Label: "Revenue", XAxis: []string{"Q1"}, Series: []BarSeries{{Name: "Hardware", Data: []BarData{{Value: 12}}}},
				Zoom: &BarZoom{Mode: BarZoomMode("wheel"), StartPercent: 10, EndPercent: 80},
			},
			wantError: `bar chart zoom mode "wheel" is not supported`,
		},
		"invalid zoom range": {
			cfg: BarConfig{
				Label: "Revenue", XAxis: []string{"Q1"}, Series: []BarSeries{{Name: "Hardware", Data: []BarData{{Value: 12}}}},
				Zoom: &BarZoom{Mode: BarZoomInside, StartPercent: 80, EndPercent: 10},
			},
			wantError: "bar chart zoom range must satisfy 0 <= start < end <= 100",
		},
		"unknown point category": {
			cfg: BarConfig{
				Label: "Revenue", XAxis: []string{"Q1"}, Series: []BarSeries{{Name: "Hardware", Data: []BarData{{Value: 12}}, References: BarReferences{Points: []BarPointReference{{Name: "Target", Coordinate: &BarCoordinate{Category: "Q2", Value: 12}}}}}},
			},
			wantError: `bar chart series "Hardware": point reference 0 category "Q2" is not on the category axis`,
		},
		"ambiguous point reference": {
			cfg: BarConfig{
				Label: "Revenue", XAxis: []string{"Q1"}, Series: []BarSeries{{Name: "Hardware", Data: []BarData{{Value: 12}}, References: BarReferences{Points: []BarPointReference{{Name: "Target", Statistic: BarStatisticMaximum, Coordinate: &BarCoordinate{Category: "Q1", Value: 12}}}}}},
			},
			wantError: `bar chart series "Hardware": point reference 0 requires exactly one reference mode`,
		},
		"invalid guide statistic": {
			cfg: BarConfig{
				Label: "Revenue", XAxis: []string{"Q1"}, Series: []BarSeries{{Name: "Hardware", Data: []BarData{{Value: 12}}, References: BarReferences{Lines: []BarGuideReference{{Name: "Median", Statistic: BarStatistic("median")}}}}},
			},
			wantError: `bar chart series "Hardware": guide reference 0 statistic "median" is not supported`,
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

func TestLiveBarDoesNotRenderStaleExactValues(t *testing.T) {
	t.Parallel()
	instance := Bar(BarConfig{
		Label:  "Live categories",
		XAxis:  []string{"Q1"},
		Series: []BarSeries{{Name: "Hardware", Data: []BarData{{Value: 12}}}},
		Live:   &LiveData{URL: "/events", Event: "chart"},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if strings.Contains(output.String(), "data-bar-exact-values") {
		t.Fatal("live Bar rendered a static exact-value table that would become stale")
	}
}
