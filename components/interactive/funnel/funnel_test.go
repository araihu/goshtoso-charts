package funnel

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"regexp"
	"strings"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

var funnelIDPattern = regexp.MustCompile(`id="([A-Za-z]{12})"`)

func TestFunnelNormalizedRenderHashes(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		config Config
		want   string
	}{
		"default descending": {
			config: Config{
				Label: "Pipeline", Series: []Series{{Name: "Pipeline", Data: []Data{{Name: "Lead", Value: 100}, {Name: "Won", Value: 15}}}},
			},
			want: "e61cfdcffe26c8cd53fdbcf6a4c67d21bbb5cd49bbad582d5e988bcbe6b93314",
		},
		"ascending custom": {
			config: Config{
				Label: "Checkout funnel", Caption: "Visitors progressing through checkout.", Order: OrderAscending,
				Series: []Series{
					{Name: "Checkout", Data: []Data{{Name: "Visit", Value: 100}, {Name: "Payment", Value: 24}}, Options: chart.SeriesOptions{Animation: chart.Bool(false)}},
					{Name: "Previous", Data: []Data{{Name: "Visit", Value: 90}, {Name: "Payment", Value: 18}}},
				},
				Width: "720px", Height: "360px",
				Options: chart.ChartOptions{Title: &chart.TitleOptions{Text: "Conversion"}}, SeriesOptions: chart.SeriesOptions{Label: &chart.LabelOptions{Show: chart.Bool(true), Position: "left"}},
				Style: charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
			},
			want: "8d8a74d2c86fba47a5a601c8bd5ca785b18739057da93ccf58acce0ce06bb1ce",
		},
		"data order wrapper": {
			config: Config{
				Label: "Exact stages", Order: OrderData,
				Series:  []Series{{Name: "Observed", Data: []Data{{Name: "Middle", Value: 37.5}, {Name: "First", Value: 0}, {Name: "Last", Value: 92.25}}, Options: chart.SeriesOptions{Label: &chart.LabelOptions{Show: chart.Bool(false)}}}},
				Options: chart.ChartOptions{Animation: chart.Bool(false), Controls: chartcontrol.Options{Mode: chartcontrol.WrapperModeOmitted}, Export: &chartcontrol.ExportOptions{Filename: "exact-funnel"}},
				Style:   charttheme.Style{Palette: charttheme.PalettePastel, Class: "caller-funnel"},
			},
			want: "68804d580d388508c0379a0de17a542a23be0ed4063130644cb498abca030436",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			markup := render(t, Funnel(test.config))
			match := funnelIDPattern.FindStringSubmatch(markup)
			if len(match) != 2 {
				t.Fatalf("rendered markup lacks chart ID: %s", markup)
			}
			normalized := strings.ReplaceAll(markup, match[1], "CHARTID")
			digest := sha256.Sum256([]byte(normalized))
			if got := hex.EncodeToString(digest[:]); got != test.want {
				t.Fatalf("normalized render SHA-256 = %s, want %s", got, test.want)
			}
		})
	}
}

func TestFunnelRendersConfiguredChart(t *testing.T) {
	t.Parallel()
	instance := Funnel(Config{
		Label: "Checkout funnel", Caption: "Visitors progressing through checkout.",
		Order: OrderAscending,
		Series: []Series{{
			Name:    "Checkout",
			Data:    []Data{{Name: "Visit", Value: 100}, {Name: "Payment", Value: 24}},
			Options: chart.SeriesOptions{Animation: chart.Bool(false)},
		}},
		Width: "720px", Height: "360px",
		Options:       chart.ChartOptions{Title: &chart.TitleOptions{Text: "Conversion"}},
		SeriesOptions: chart.SeriesOptions{Label: &chart.LabelOptions{Show: chart.Bool(true), Position: "left"}},
		Style:         charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
	})

	if instance.Kind() != chartcomponents.KindInteractiveFunnel {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	markup := render(t, instance)
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
	rows := detailRows([]Series{
		{Name: "First", Data: []Data{{Name: "Visit", Value: 31}, {Name: "Add", Value: 37.5}}},
		{Name: "Second", Data: []Data{{Name: "Deal", Value: 0}}},
	})
	want := []valueRow{
		{Series: "First", Stage: "Visit", Value: "31"},
		{Series: "First", Stage: "Add", Value: "37.5"},
		{Series: "Second", Stage: "Deal", Value: "0"},
	}
	if len(rows) != len(want) {
		t.Fatalf("detailRows() length = %d, want %d", len(rows), len(want))
	}
	for index := range want {
		if rows[index] != want[index] {
			t.Errorf("detailRows()[%d] = %#v, want %#v", index, rows[index], want[index])
		}
	}
}

func TestFunnelRendersDefaultAndDataOrders(t *testing.T) {
	t.Parallel()
	for name, order := range map[string]Order{"default": OrderDescending, "data": OrderData} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			instance := Funnel(Config{
				Label: "Pipeline", Order: order,
				Series: []Series{{Name: "Pipeline", Data: []Data{{Name: "Lead", Value: 2}}}},
				Style:  charttheme.Style{Palette: charttheme.PalettePastel},
			})
			markup := render(t, instance)
			wantOrder := `"sort":"descending"`
			if order == OrderData {
				wantOrder = `"sort":"none"`
			}
			for _, want := range []string{wantOrder, `"color":["#93c5fd","#fca5a5"`} {
				if !strings.Contains(markup, want) {
					t.Errorf("rendered markup missing %q", want)
				}
			}
		})
	}
}

func TestFunnelRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	valid := func() Config {
		return Config{Label: "Pipeline", Series: []Series{{
			Name: "Pipeline", Data: []Data{{Name: "Lead", Value: 10}},
		}}}
	}
	tests := map[string]struct {
		mutate    func() Config
		wantError string
	}{
		"missing label":       {func() Config { cfg := valid(); cfg.Label = ""; return cfg }, "funnel chart label is required"},
		"bad order":           {func() Config { cfg := valid(); cfg.Order = "sideways"; return cfg }, `funnel chart order "sideways" is not supported`},
		"missing series":      {func() Config { cfg := valid(); cfg.Series = nil; return cfg }, "funnel chart series is required"},
		"missing series name": {func() Config { cfg := valid(); cfg.Series[0].Name = ""; return cfg }, "funnel chart series 0 name is required"},
		"missing data":        {func() Config { cfg := valid(); cfg.Series[0].Data = nil; return cfg }, `funnel chart series "Pipeline" data is required`},
		"missing data name":   {func() Config { cfg := valid(); cfg.Series[0].Data[0].Name = ""; return cfg }, `funnel chart series "Pipeline" data point 0 name is required`},
		"negative value":      {func() Config { cfg := valid(); cfg.Series[0].Data[0].Value = -1; return cfg }, `funnel chart series "Pipeline" data point "Lead" value must be a finite nonnegative value`},
		"nonfinite value":     {func() Config { cfg := valid(); cfg.Series[0].Data[0].Value = math.NaN(); return cfg }, `funnel chart series "Pipeline" data point "Lead" value must be a finite nonnegative value`},
		"invalid shared option": {func() Config {
			cfg := valid()
			cfg.Options.Legend = &chart.LegendOptions{Padding: &chart.EdgeInsets{Left: -1}}
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

func render(t *testing.T, instance chart.Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return output.String()
}
