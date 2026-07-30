package heatmap

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"regexp"
	"strings"
	"testing"
	"time"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

var heatMapIDPattern = regexp.MustCompile(`id="([A-Za-z]{12})"`)

func TestHeatMapNormalizedRenderHashes(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.July, 31, 0, 0, 0, 0, time.UTC)
	tests := map[string]struct {
		cfg  Config
		want string
	}{
		"Cartesian": {
			cfg: Config{
				Label: "Weekly traffic", Caption: "Requests by weekday and hour.",
				XAxis: []string{"08:00", "09:00"}, YAxis: []string{"Mon", "Tue"},
				ValueRange: ValueRange{Min: 0, Max: 20, Calculable: chart.Bool(true)},
				SplitArea:  chart.Bool(true),
				Series: []Series{{
					Name:    "Requests",
					Data:    []Data{{X: 0, Y: 0, Value: 25}, {X: 1, Y: 0, Value: 0}, {X: 1, Y: 1, Missing: true}},
					Options: chart.SeriesOptions{Label: &chart.LabelOptions{Show: chart.Bool(true)}},
				}},
				Width: "720px", Height: "360px",
				Options: chart.ChartOptions{
					Title:    &chart.TitleOptions{Text: "Traffic density"},
					Controls: chartcontrol.Options{Fullscreen: true, Mode: chartcontrol.WrapperModeDisabled},
					Export:   &chartcontrol.ExportOptions{Filename: "traffic-density", Formats: []chartcontrol.ExportFormat{chartcontrol.ExportPNG}, PixelRatio: 2},
				},
				SeriesOptions: chart.SeriesOptions{Animation: chart.Bool(false)},
				Style:         charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
			},
			want: "9e8ff5c2d5070271ede721b2d1066471e3510f3687d03bfb8cf238ae5e5b6935",
		},
		"Calendar": {
			cfg: Config{
				Label: "July activity", Coordinate: CoordinateCalendar,
				Calendar: &Calendar{
					Start: start, End: end,
					Options: chart.CalendarOptions{
						Top: "80", CellSize: "20", Orient: "horizontal",
						CellStyle: &chart.ItemStyle{Color: "#112233", BorderColor: "#445566", BorderWidth: 0.5},
						DayLabel:  &chart.CalendarLabelOptions{Margin: 4, Position: "left", FontSize: 9},
					},
				},
				ValueRange: ValueRange{Min: 0, Max: 10},
				Series:     []Series{{Name: "Commits", Data: []Data{{Date: start, Value: 3}, {Date: end, Value: 8}}}},
				Options:    chart.ChartOptions{Controls: chartcontrol.Options{Mode: chartcontrol.WrapperModeHidden}, Export: &chartcontrol.ExportOptions{Disabled: true}},
				Style:      charttheme.Style{Palette: charttheme.PalettePastel},
			},
			want: "c73bcca26e8646a9cabbd7ece6c547c643db3aefd681faea09d51b90a5ca70e3",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			markup := renderHeatMap(t, HeatMap(test.cfg))
			match := heatMapIDPattern.FindStringSubmatch(markup)
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

func TestHeatMapRendersCartesianChart(t *testing.T) {
	t.Parallel()
	instance := HeatMap(Config{
		Label: "Weekly traffic", Caption: "Requests by weekday and hour.",
		XAxis: []string{"08:00", "09:00"}, YAxis: []string{"Mon", "Tue"},
		ValueRange: ValueRange{Min: 0, Max: 20, Calculable: chart.Bool(true)},
		SplitArea:  chart.Bool(true),
		Series: []Series{{
			Name:    "Requests",
			Data:    []Data{{X: 0, Y: 0, Value: 25}, {X: 1, Y: 0, Value: 0}, {X: 1, Y: 1, Missing: true}},
			Options: chart.SeriesOptions{Label: &chart.LabelOptions{Show: chart.Bool(true)}},
		}},
		Width: "720px", Height: "360px",
		Options:       chart.ChartOptions{Title: &chart.TitleOptions{Text: "Traffic density"}},
		SeriesOptions: chart.SeriesOptions{Animation: chart.Bool(false)},
		Style:         charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
	})

	if instance.Kind() != chartcomponents.KindInteractiveHeatMap {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	markup := renderHeatMap(t, instance)
	for _, want := range []string{
		"Weekly traffic", "Requests by weekday and hour.", "width:720px;height:360px",
		`"data":["08:00","09:00"]`, `"data":["Mon","Tue"]`,
		`"name":"Requests"`, `"value":[0,0,25]`, `"value":[1,0,0]`, `"value":[1,1,"-"]`,
		`"max":20`, `"color":["#123456","#ff8a3d"`,
		`"left":"52","right":"0","bottom":"56","containLabel":true`,
		`"left":"8","bottom":"24"`, `"calculable":true`, `"splitArea":{"show":true}`,
		`"show":true`, `"animation":false`, `"text":"Traffic density"`,
		`data-goshtoso-charts-explicit-visual-map-colors="true"`,
		"goshtoso-charts-palette-araihu min-h-80", `data-heatmap-exact-values`,
		`Weekly traffic exact heatmap values`, `data-heatmap-missing="true"`, `No data`,
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
	instance := HeatMap(Config{
		Label: "July activity", Coordinate: CoordinateCalendar,
		Calendar: &Calendar{
			Start: start, End: end,
			Options: chart.CalendarOptions{
				Top: "80", CellSize: "20", Orient: "horizontal",
				CellStyle: &chart.ItemStyle{Color: "#112233", BorderColor: "#445566", BorderWidth: 0.5},
				DayLabel:  &chart.CalendarLabelOptions{Margin: 4, Position: "left", FontSize: 9},
			},
		},
		ValueRange: ValueRange{Min: 0, Max: 10},
		Series: []Series{{
			Name: "Commits",
			Data: []Data{{Date: start, Value: 3}, {Date: end, Value: 8}},
		}},
		Style: charttheme.Style{Palette: charttheme.PalettePastel},
	})

	markup := renderHeatMap(t, instance)
	for _, want := range []string{
		`"calendar"`, `"range":["2026-07-01","2026-07-31"]`, `"cellSize":"20"`, `"orient":"horizontal"`,
		`"itemStyle":{"color":"#112233","borderColor":"#445566","borderWidth":0.5`, `"dayLabel":{"margin":4,"position":"left","fontSize":9}`,
		`"coordinateSystem":"calendar"`, `"value":["2026-07-01",3]`,
		`"value":["2026-07-31",8]`, `"color":["#93c5fd","#fca5a5"`,
		`"right":"0","bottom":"24"`, `July activity exact heatmap values`, `tabindex="0" role="region"`,
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
	validCartesian := func() Config {
		return Config{
			Label: "Load", XAxis: []string{"A"}, YAxis: []string{"B"},
			ValueRange: ValueRange{Min: 0, Max: 10},
			Series:     []Series{{Name: "Load", Data: []Data{{Value: 5}}}},
		}
	}
	validCalendar := func() Config {
		return Config{
			Label: "Load", Coordinate: CoordinateCalendar,
			Calendar:   &Calendar{Start: start, End: end},
			ValueRange: ValueRange{Min: 0, Max: 10},
			Series:     []Series{{Name: "Load", Data: []Data{{Date: start, Value: 5}}}},
		}
	}

	tests := map[string]struct {
		mutate    func() Config
		wantError string
	}{
		"missing label":           {func() Config { cfg := validCartesian(); cfg.Label = ""; return cfg }, "heatmap chart label is required"},
		"unsupported coordinate":  {func() Config { cfg := validCartesian(); cfg.Coordinate = "polar"; return cfg }, `heatmap chart coordinate "polar" is not supported`},
		"invalid value range":     {func() Config { cfg := validCartesian(); cfg.ValueRange.Max = 0; return cfg }, "heatmap chart value range must contain finite min and max with min less than max"},
		"nonfinite range":         {func() Config { cfg := validCartesian(); cfg.ValueRange.Max = math.Inf(1); return cfg }, "heatmap chart value range must contain finite min and max with min less than max"},
		"renderer range overflow": {func() Config { cfg := validCartesian(); cfg.ValueRange.Max = math.MaxFloat64; return cfg }, "heatmap chart value range exceeds renderer limits"},
		"missing Cartesian axes":  {func() Config { cfg := validCartesian(); cfg.XAxis = nil; return cfg }, "heatmap chart x and y axes are required for Cartesian coordinates"},
		"calendar in Cartesian mode": {func() Config {
			cfg := validCartesian()
			cfg.Calendar = &Calendar{Start: start, End: end}
			return cfg
		}, "heatmap chart calendar is not allowed for Cartesian coordinates"},
		"axes in calendar mode": {func() Config { cfg := validCalendar(); cfg.XAxis = []string{"bad"}; return cfg }, "heatmap chart category axes are not allowed for calendar coordinates"},
		"split area in calendar mode": {func() Config {
			cfg := validCalendar()
			cfg.SplitArea = chart.Bool(true)
			return cfg
		}, "heatmap chart split area is not allowed for calendar coordinates"},
		"missing calendar":       {func() Config { cfg := validCalendar(); cfg.Calendar = nil; return cfg }, "heatmap chart calendar is required for calendar coordinates"},
		"missing calendar dates": {func() Config { cfg := validCalendar(); cfg.Calendar.Start = time.Time{}; return cfg }, "heatmap chart calendar start and end dates are required"},
		"reversed calendar": {func() Config {
			cfg := validCalendar()
			cfg.Calendar.Start, cfg.Calendar.End = end, start
			return cfg
		}, "heatmap chart calendar start must not follow end"},
		"missing series":               {func() Config { cfg := validCartesian(); cfg.Series = nil; return cfg }, "heatmap chart series is required"},
		"missing series name":          {func() Config { cfg := validCartesian(); cfg.Series[0].Name = ""; return cfg }, "heatmap chart series 0 name is required"},
		"missing data":                 {func() Config { cfg := validCartesian(); cfg.Series[0].Data = nil; return cfg }, `heatmap chart series "Load" data is required`},
		"nonfinite value":              {func() Config { cfg := validCartesian(); cfg.Series[0].Data[0].Value = math.NaN(); return cfg }, `heatmap chart series "Load" data point 0 value must be finite`},
		"Cartesian index outside axes": {func() Config { cfg := validCartesian(); cfg.Series[0].Data[0].X = 1; return cfg }, `heatmap chart series "Load" data point 0 category indexes are outside the axes`},
		"missing calendar data date":   {func() Config { cfg := validCalendar(); cfg.Series[0].Data[0].Date = time.Time{}; return cfg }, `heatmap chart series "Load" data point 0 date is required`},
		"date outside calendar": {func() Config {
			cfg := validCalendar()
			cfg.Series[0].Data[0].Date = end.AddDate(0, 0, 1)
			return cfg
		}, `heatmap chart series "Load" data point 0 date is outside the calendar range`},
		"negative calendar cell border": {func() Config {
			cfg := validCalendar()
			cfg.Calendar.Options.CellStyle = &chart.ItemStyle{BorderWidth: -0.5}
			return cfg
		}, "heatmap chart calendar cell border width must be finite and nonnegative"},
		"nonfinite calendar cell border": {func() Config {
			cfg := validCalendar()
			cfg.Calendar.Options.CellStyle = &chart.ItemStyle{BorderWidth: math.NaN()}
			return cfg
		}, "heatmap chart calendar cell border width must be finite and nonnegative"},
		"calendar cell border renderer overflow": {func() Config {
			cfg := validCalendar()
			cfg.Calendar.Options.CellStyle = &chart.ItemStyle{BorderWidth: math.MaxFloat64}
			return cfg
		}, "heatmap chart calendar cell border width exceeds renderer limits"},
		"calendar cell opacity out of range": {func() Config {
			cfg := validCalendar()
			cfg.Calendar.Options.CellStyle = &chart.ItemStyle{Opacity: chart.Float(1.5)}
			return cfg
		}, "heatmap chart calendar cell opacity must be between 0 and 1"},
		"negative calendar label margin": {func() Config {
			cfg := validCalendar()
			cfg.Calendar.Options.MonthLabel = &chart.CalendarLabelOptions{Margin: -1}
			return cfg
		}, "heatmap chart calendar month label margin and font size must be finite and nonnegative"},
		"nonfinite calendar label margin": {func() Config {
			cfg := validCalendar()
			cfg.Calendar.Options.MonthLabel = &chart.CalendarLabelOptions{Margin: math.Inf(1)}
			return cfg
		}, "heatmap chart calendar month label margin and font size must be finite and nonnegative"},
		"unsupported calendar label position": {func() Config {
			cfg := validCalendar()
			cfg.Calendar.Options.YearLabel = &chart.CalendarLabelOptions{Position: "center"}
			return cfg
		}, `heatmap chart calendar year label position "center" is not supported`},
		"deterministic calendar label order": {func() Config {
			cfg := validCalendar()
			cfg.Calendar.Options.DayLabel = &chart.CalendarLabelOptions{Margin: -1}
			cfg.Calendar.Options.MonthLabel = &chart.CalendarLabelOptions{Margin: -1}
			cfg.Calendar.Options.YearLabel = &chart.CalendarLabelOptions{Margin: -1}
			return cfg
		}, "heatmap chart calendar day label margin and font size must be finite and nonnegative"},
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
