package interactive

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"
	"time"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func TestHeatMapRendersCartesianChart(t *testing.T) {
	t.Parallel()
	instance := HeatMap(HeatMapConfig{
		Label: "Weekly traffic", Caption: "Requests by weekday and hour.",
		XAxis: []string{"08:00", "09:00"}, YAxis: []string{"Mon", "Tue"},
		ValueRange: HeatMapValueRange{Min: 0, Max: 20},
		Series: []HeatMapSeries{{
			Name:    "Requests",
			Data:    []HeatMapData{{X: 0, Y: 0, Value: 5}, {X: 1, Y: 1, Value: 18}},
			Options: SeriesOptions{Label: &LabelOptions{Show: Bool(true)}},
		}},
		Width: "720px", Height: "360px",
		Options:       ChartOptions{Title: &TitleOptions{Text: "Traffic density"}},
		SeriesOptions: SeriesOptions{Animation: Bool(false)},
		Style:         charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
	})

	if instance.Kind() != chartcomponents.KindInteractiveHeatMap {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	markup := renderHeatMap(t, instance)
	for _, want := range []string{
		"Weekly traffic", "Requests by weekday and hour.", "width:720px;height:360px",
		`"data":["08:00","09:00"]`, `"data":["Mon","Tue"]`,
		`"name":"Requests"`, `"value":[0,0,5]`, `"value":[1,1,18]`,
		`"max":20`, `"color":["#123456","#ff8a3d"`,
		`"left":"52","right":"0","bottom":"56","containLabel":true`,
		`"left":"8","bottom":"24"`,
		`"show":true`, `"animation":false`, `"text":"Traffic density"`,
		`data-goshtoso-charts-explicit-visual-map-colors="true"`,
		"goshtoso-charts-palette-araihu min-h-80",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestHeatMapRendersCalendarChart(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.July, 31, 0, 0, 0, 0, time.UTC)
	instance := HeatMap(HeatMapConfig{
		Label: "July activity", Coordinate: HeatMapCoordinateCalendar,
		Calendar: &HeatMapCalendar{
			Start: start, End: end,
			Options: CalendarOptions{Top: "80", CellSize: "20"},
		},
		ValueRange: HeatMapValueRange{Min: 0, Max: 10},
		Series: []HeatMapSeries{{
			Name: "Commits",
			Data: []HeatMapData{{Date: start, Value: 3}, {Date: end, Value: 8}},
		}},
		Style: charttheme.Style{Palette: charttheme.PalettePastel},
	})

	markup := renderHeatMap(t, instance)
	for _, want := range []string{
		`"calendar"`, `"range":["2026-07-01","2026-07-31"]`, `"cellSize":"20"`,
		`"coordinateSystem":"calendar"`, `"value":["2026-07-01",3]`,
		`"value":["2026-07-31",8]`, `"color":["#93c5fd","#fca5a5"`,
		`"left":"8","bottom":"24"`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	if strings.Contains(markup, `"containLabel":true`) {
		t.Error("calendar heatmap unexpectedly rendered Cartesian grid layout")
	}
}

func TestHeatMapRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	start := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.July, 31, 0, 0, 0, 0, time.UTC)
	validCartesian := func() HeatMapConfig {
		return HeatMapConfig{
			Label: "Load", XAxis: []string{"A"}, YAxis: []string{"B"},
			ValueRange: HeatMapValueRange{Min: 0, Max: 10},
			Series:     []HeatMapSeries{{Name: "Load", Data: []HeatMapData{{Value: 5}}}},
		}
	}
	validCalendar := func() HeatMapConfig {
		return HeatMapConfig{
			Label: "Load", Coordinate: HeatMapCoordinateCalendar,
			Calendar:   &HeatMapCalendar{Start: start, End: end},
			ValueRange: HeatMapValueRange{Min: 0, Max: 10},
			Series:     []HeatMapSeries{{Name: "Load", Data: []HeatMapData{{Date: start, Value: 5}}}},
		}
	}

	tests := map[string]struct {
		mutate    func() HeatMapConfig
		wantError string
	}{
		"missing label":           {func() HeatMapConfig { cfg := validCartesian(); cfg.Label = ""; return cfg }, "heatmap chart label is required"},
		"unsupported coordinate":  {func() HeatMapConfig { cfg := validCartesian(); cfg.Coordinate = "polar"; return cfg }, `heatmap chart coordinate "polar" is not supported`},
		"invalid value range":     {func() HeatMapConfig { cfg := validCartesian(); cfg.ValueRange.Max = 0; return cfg }, "heatmap chart value range must contain finite min and max with min less than max"},
		"nonfinite range":         {func() HeatMapConfig { cfg := validCartesian(); cfg.ValueRange.Max = math.Inf(1); return cfg }, "heatmap chart value range must contain finite min and max with min less than max"},
		"renderer range overflow": {func() HeatMapConfig { cfg := validCartesian(); cfg.ValueRange.Max = math.MaxFloat64; return cfg }, "heatmap chart value range exceeds renderer limits"},
		"missing Cartesian axes":  {func() HeatMapConfig { cfg := validCartesian(); cfg.XAxis = nil; return cfg }, "heatmap chart x and y axes are required for Cartesian coordinates"},
		"calendar in Cartesian mode": {func() HeatMapConfig {
			cfg := validCartesian()
			cfg.Calendar = &HeatMapCalendar{Start: start, End: end}
			return cfg
		}, "heatmap chart calendar is not allowed for Cartesian coordinates"},
		"axes in calendar mode":  {func() HeatMapConfig { cfg := validCalendar(); cfg.XAxis = []string{"bad"}; return cfg }, "heatmap chart category axes are not allowed for calendar coordinates"},
		"missing calendar":       {func() HeatMapConfig { cfg := validCalendar(); cfg.Calendar = nil; return cfg }, "heatmap chart calendar is required for calendar coordinates"},
		"missing calendar dates": {func() HeatMapConfig { cfg := validCalendar(); cfg.Calendar.Start = time.Time{}; return cfg }, "heatmap chart calendar start and end dates are required"},
		"reversed calendar": {func() HeatMapConfig {
			cfg := validCalendar()
			cfg.Calendar.Start, cfg.Calendar.End = end, start
			return cfg
		}, "heatmap chart calendar start must not follow end"},
		"missing series":               {func() HeatMapConfig { cfg := validCartesian(); cfg.Series = nil; return cfg }, "heatmap chart series is required"},
		"missing series name":          {func() HeatMapConfig { cfg := validCartesian(); cfg.Series[0].Name = ""; return cfg }, "heatmap chart series 0 name is required"},
		"missing data":                 {func() HeatMapConfig { cfg := validCartesian(); cfg.Series[0].Data = nil; return cfg }, `heatmap chart series "Load" data is required`},
		"nonfinite value":              {func() HeatMapConfig { cfg := validCartesian(); cfg.Series[0].Data[0].Value = math.NaN(); return cfg }, `heatmap chart series "Load" data point 0 value must be finite`},
		"value outside range":          {func() HeatMapConfig { cfg := validCartesian(); cfg.Series[0].Data[0].Value = 11; return cfg }, `heatmap chart series "Load" data point 0 value is outside the configured range`},
		"Cartesian index outside axes": {func() HeatMapConfig { cfg := validCartesian(); cfg.Series[0].Data[0].X = 1; return cfg }, `heatmap chart series "Load" data point 0 category indexes are outside the axes`},
		"missing calendar data date":   {func() HeatMapConfig { cfg := validCalendar(); cfg.Series[0].Data[0].Date = time.Time{}; return cfg }, `heatmap chart series "Load" data point 0 date is required`},
		"date outside calendar": {func() HeatMapConfig {
			cfg := validCalendar()
			cfg.Series[0].Data[0].Date = end.AddDate(0, 0, 1)
			return cfg
		}, `heatmap chart series "Load" data point 0 date is outside the calendar range`},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			err := HeatMap(test.mutate()).Render(context.Background(), &output)
			if err == nil || err.Error() != test.wantError {
				t.Fatalf("Render() error = %v, want %q", err, test.wantError)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %d bytes for invalid config", output.Len())
			}
		})
	}
}

func renderHeatMap(t *testing.T, instance Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return output.String()
}
