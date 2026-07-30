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

func TestFunnelRendersConfiguredChart(t *testing.T) {
	t.Parallel()
	instance := Funnel(FunnelConfig{
		Label: "Checkout funnel", Caption: "Visitors progressing through checkout.",
		Order: FunnelOrderAscending,
		Series: []FunnelSeries{{
			Name:    "Checkout",
			Data:    []FunnelData{{Name: "Visit", Value: 100}, {Name: "Payment", Value: 24}},
			Options: SeriesOptions{Animation: Bool(false)},
		}},
		Width: "720px", Height: "360px",
		Options:       ChartOptions{Title: &TitleOptions{Text: "Conversion"}},
		SeriesOptions: SeriesOptions{Label: &LabelOptions{Show: Bool(true), Position: "left"}},
		Style:         charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
	})

	if instance.Kind() != chartcomponents.KindInteractiveFunnel {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		"Checkout funnel", "Visitors progressing through checkout.", "width:720px;height:360px",
		`"name":"Checkout"`, `"name":"Visit","value":100`, `"name":"Payment","value":24`,
		`"sort":"ascending"`, `"show":true`, `"position":"left"`, `"animation":false`,
		`"text":"Conversion"`, `"color":["#123456","#ff8a3d"`,
		"goshtoso-charts-palette-araihu min-h-80", "Exact funnel values",
		`aria-label="Checkout funnel exact funnel values"`, "Checkout", "Visit", "100", "Payment", "24",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestFunnelDetailRowsPreserveSeriesStageOrderAndExactValues(t *testing.T) {
	t.Parallel()
	rows := funnelDetailRows([]FunnelSeries{
		{Name: "First", Data: []FunnelData{{Name: "Visit", Value: 31}, {Name: "Add", Value: 37.5}}},
		{Name: "Second", Data: []FunnelData{{Name: "Deal", Value: 0}}},
	})
	want := []funnelValueRow{
		{Series: "First", Stage: "Visit", Value: "31"},
		{Series: "First", Stage: "Add", Value: "37.5"},
		{Series: "Second", Stage: "Deal", Value: "0"},
	}
	if len(rows) != len(want) {
		t.Fatalf("funnelDetailRows() length = %d, want %d", len(rows), len(want))
	}
	for index := range want {
		if rows[index] != want[index] {
			t.Errorf("funnelDetailRows()[%d] = %#v, want %#v", index, rows[index], want[index])
		}
	}
}

func TestFunnelRendersDefaultAndDataOrders(t *testing.T) {
	t.Parallel()
	for name, order := range map[string]FunnelOrder{"default": FunnelOrderDescending, "data": FunnelOrderData} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			instance := Funnel(FunnelConfig{
				Label: "Pipeline", Order: order,
				Series: []FunnelSeries{{Name: "Pipeline", Data: []FunnelData{{Name: "Lead", Value: 2}}}},
				Style:  charttheme.Style{Palette: charttheme.PalettePastel},
			})
			var output bytes.Buffer
			if err := instance.Render(context.Background(), &output); err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			wantOrder := `"sort":"descending"`
			if order == FunnelOrderData {
				wantOrder = `"sort":"none"`
			}
			for _, want := range []string{wantOrder, `"color":["#93c5fd","#fca5a5"`} {
				if !strings.Contains(output.String(), want) {
					t.Errorf("rendered markup missing %q", want)
				}
			}
		})
	}
}

func TestFunnelRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	valid := func() FunnelConfig {
		return FunnelConfig{Label: "Pipeline", Series: []FunnelSeries{{
			Name: "Pipeline", Data: []FunnelData{{Name: "Lead", Value: 10}},
		}}}
	}
	tests := map[string]struct {
		mutate    func() FunnelConfig
		wantError string
	}{
		"missing label":       {func() FunnelConfig { cfg := valid(); cfg.Label = ""; return cfg }, "funnel chart label is required"},
		"bad order":           {func() FunnelConfig { cfg := valid(); cfg.Order = "sideways"; return cfg }, `funnel chart order "sideways" is not supported`},
		"missing series":      {func() FunnelConfig { cfg := valid(); cfg.Series = nil; return cfg }, "funnel chart series is required"},
		"missing series name": {func() FunnelConfig { cfg := valid(); cfg.Series[0].Name = ""; return cfg }, "funnel chart series 0 name is required"},
		"missing data":        {func() FunnelConfig { cfg := valid(); cfg.Series[0].Data = nil; return cfg }, `funnel chart series "Pipeline" data is required`},
		"missing data name":   {func() FunnelConfig { cfg := valid(); cfg.Series[0].Data[0].Name = ""; return cfg }, `funnel chart series "Pipeline" data point 0 name is required`},
		"negative value":      {func() FunnelConfig { cfg := valid(); cfg.Series[0].Data[0].Value = -1; return cfg }, `funnel chart series "Pipeline" data point "Lead" value must be a finite nonnegative value`},
		"nonfinite value":     {func() FunnelConfig { cfg := valid(); cfg.Series[0].Data[0].Value = math.NaN(); return cfg }, `funnel chart series "Pipeline" data point "Lead" value must be a finite nonnegative value`},
		"invalid shared option": {func() FunnelConfig {
			cfg := valid()
			cfg.Options.Legend = &LegendOptions{Padding: &EdgeInsets{Left: -1}}
			return cfg
		}, "legend padding must be nonnegative"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			err := Funnel(test.mutate()).Render(context.Background(), &output)
			if err == nil || err.Error() != test.wantError {
				t.Fatalf("Render() error = %v, want %q", err, test.wantError)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %d bytes for invalid config", output.Len())
			}
		})
	}
}
