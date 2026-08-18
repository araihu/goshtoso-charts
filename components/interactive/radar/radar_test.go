package radar

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

var radarIDPattern = regexp.MustCompile(`goecharts_([A-Za-z0-9]{12})`)

func TestRadarNormalizedRenderHashes(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		config Config
		want   string
	}{
		"default polygon": {
			config: Config{
				Label:      "Fallback profile",
				Indicators: []Indicator{{Name: "A", Max: 10}, {Name: "B", Max: 20}},
				Series:     []Series{{Name: "Profile", Data: []Data{{Name: "Current", Values: []float64{4, 8}}}}},
			},
		want: "f02bbabd061a59c2572ac6500cd9b3e9ecd3462a6308b62b655dd6ee2b2d9c08",
		},
		"explicit polygon": {
			config: Config{
				Label: "Service profile", Caption: "Current and target service health.",
				Indicators: []Indicator{{Name: "Availability", Max: 100}, {Name: "Latency", Max: 500}, {Name: "Capacity", Max: 200}},
				Coordinate: CoordinateOptions{Shape: ShapePolygon, SplitNumber: 4},
				Series: []Series{{
					Name: "Profile", Data: []Data{{Name: "Current", Values: []float64{99.9, 180, 120}}, {Name: "Target", Values: []float64{100, 100, 160}}},
					Options: chart.SeriesOptions{AreaStyle: &chart.AreaStyle{Opacity: chart.Float(0.2)}},
				}},
				Width: "720px", Height: "360px",
				Options:       chart.ChartOptions{Title: &chart.TitleOptions{Text: "Health"}, Animation: chart.Bool(false)},
				SeriesOptions: chart.SeriesOptions{LineStyle: &chart.LineStyle{Width: 2}},
				Style:         charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
			},
		want: "c1f02681940f219b49e3174a7e9dba8bfd284f96dc028df73d5466c123403117",
		},
		"circle": {
			config: Config{
				Label: "Daily air quality profiles", Caption: "Two cities across six pollutant indicators.",
				Indicators: []Indicator{
					{Name: "AQI", Max: 300}, {Name: "PM2.5", Max: 250}, {Name: "PM10", Max: 300},
					{Name: "CO", Max: 5}, {Name: "NO2", Max: 200}, {Name: "SO2", Max: 100},
				},
				Coordinate: CoordinateOptions{
					Shape: ShapeCircle, SplitNumber: 5, SplitArea: chart.Bool(true),
					SplitLine: &SplitLineOptions{Show: chart.Bool(true), Style: &chart.LineStyle{Width: 1, Type: "dashed", Opacity: chart.Float(0.1)}},
				},
				Series: []Series{
					{Name: "Beijing", Data: []Data{{Name: "Day 1", Values: []float64{55, 9, 56, 0.46, 18, 6}}}},
					{Name: "Guangzhou", Data: []Data{{Name: "Day 1", Values: []float64{26, 37, 27, 1.163, 27, 13}}}},
				},
				Options: chart.ChartOptions{
					Legend:   &chart.LegendOptions{Show: chart.Bool(true), Left: "center", Bottom: "5px", SelectionMode: chart.LegendSelectionSingle},
					Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "daily-air-quality"},
				},
				SeriesOptions: chart.SeriesOptions{LineStyle: &chart.LineStyle{Width: 1, Opacity: chart.Float(0.5)}, AreaStyle: &chart.AreaStyle{Opacity: chart.Float(0.1)}},
				Width:         "100%", Height: "480px", Style: charttheme.Style{Palette: charttheme.PaletteAraiHu, Class: "max-w-5xl mx-auto"},
			},
		want: "576c3d298ce3b78a4ac808cdff64cc739f6e5289e3db63325e4d8adb7b5e6703",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			if err := Radar(test.config).Render(context.Background(), &output); err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			markup := output.String()
			match := radarIDPattern.FindStringSubmatch(markup)
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

func TestRadarRendersTypedChart(t *testing.T) {
	t.Parallel()
	instance := Radar(Config{
		Label: "Service profile", Caption: "Current and target service health.",
		Indicators: []Indicator{{Name: "Availability", Max: 100}, {Name: "Latency", Max: 500}, {Name: "Capacity", Max: 200}},
		Series: []Series{{
			Name:    "Profile",
			Data:    []Data{{Name: "Current", Values: []float64{99.9, 180, 120}}, {Name: "Target", Values: []float64{100, 100, 160}}},
			Options: chart.SeriesOptions{AreaStyle: &chart.AreaStyle{Opacity: chart.Float(0.2)}},
		}},
		Width: "720px", Height: "360px",
		Options:       chart.ChartOptions{Title: &chart.TitleOptions{Text: "Health"}},
		SeriesOptions: chart.SeriesOptions{LineStyle: &chart.LineStyle{Width: 2}},
		Style:         charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
	})

	if instance.Kind() != chartcomponents.KindInteractiveRadar {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		"Service profile", "Current and target service health.", "width:720px;height:360px",
		`"name":"Availability","max":100`, `"name":"Latency","max":500`,
		`"name":"Profile"`, `"name":"Current","value":[99.9,180,120]`, `"name":"Target","value":[100,100,160]`,
		`"width":2`, `"opacity":0.2`, `"text":"Health"`, `"color":["#123456","#ff8a3d"`,
		"goshtoso-charts-palette-araihu min-h-80",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestRadarUsesFallbackPalette(t *testing.T) {
	t.Parallel()
	instance := Radar(Config{
		Label:      "Fallback profile",
		Indicators: []Indicator{{Name: "A", Max: 10}, {Name: "B", Max: 20}},
		Series:     []Series{{Name: "Profile", Data: []Data{{Name: "Current", Values: []float64{4, 8}}}}},
	})
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{`"color":["#2563eb","#dc2626"`, "goshtoso-charts-palette-auto"} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestRadarRendersTypedCoordinateLegendAndExactValueTreatments(t *testing.T) {
	t.Parallel()
	instance := Radar(Config{
		Label: "Daily air quality profiles", Caption: "Two cities across six pollutant indicators.",
		Indicators: []Indicator{
			{Name: "AQI", Max: 300}, {Name: "PM2.5", Max: 250}, {Name: "PM10", Max: 300},
			{Name: "CO", Max: 5}, {Name: "NO2", Max: 200}, {Name: "SO2", Max: 100},
		},
		Coordinate: CoordinateOptions{
			Shape: ShapeCircle, SplitNumber: 5, SplitArea: chart.Bool(true),
			SplitLine: &SplitLineOptions{Show: chart.Bool(true), Style: &chart.LineStyle{Width: 1, Type: "dashed", Opacity: chart.Float(0.1)}},
		},
		Series: []Series{
			{Name: "Beijing", Data: []Data{{Name: "Day 1", Values: []float64{55, 9, 56, 0.46, 18, 6}}}},
			{Name: "Guangzhou", Data: []Data{{Name: "Day 1", Values: []float64{26, 37, 27, 1.163, 27, 13}}}},
		},
		Options: chart.ChartOptions{
			Legend:   &chart.LegendOptions{Show: chart.Bool(true), Left: "center", Bottom: "5px", SelectionMode: chart.LegendSelectionSingle},
			Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "daily-air-quality"},
		},
		SeriesOptions: chart.SeriesOptions{LineStyle: &chart.LineStyle{Width: 1, Opacity: chart.Float(0.5)}, AreaStyle: &chart.AreaStyle{Opacity: chart.Float(0.1)}},
		Width:         "100%", Height: "480px", Style: charttheme.Style{Palette: charttheme.PaletteAraiHu, Class: "max-w-5xl mx-auto"},
	})

	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`"shape":"circle"`, `"splitNumber":5`, `"splitArea":{"show":true}`,
		`"splitLine":{"show":true,"lineStyle":{"width":1,"type":"dashed","opacity":0.1}}`,
		`"selectedMode":"single"`, `"name":"Beijing"`, `"name":"Guangzhou"`,
		`data-radar-exact-values`, `Daily air quality profiles exact radar values`,
		`<th scope="col" class="px-3 py-2 font-semibold">AQI</th>`,
		`<td class="px-3 py-2">Beijing</td>`, `<th scope="row" class="px-3 py-2 font-semibold">Day 1</th>`,
		`<td class="px-3 py-2 tabular-nums">0.46</td>`,
		`data-goshtoso-chart-wrapper-mode="enabled"`, `Download Daily air quality profiles as PNG`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	if got := strings.Count(markup, `data-radar-value-row`); got != 2 {
		t.Errorf("exact-value row count = %d, want 2", got)
	}
}

func TestRadarPropagatesWrapperLifecycleModes(t *testing.T) {
	t.Parallel()
	config := Config{
		Label:      "Profile",
		Indicators: []Indicator{{Name: "A", Max: 10}, {Name: "B", Max: 20}},
		Series:     []Series{{Name: "Series", Data: []Data{{Name: "Observation", Values: []float64{4, 8}}}}},
	}
	for _, mode := range []chartcontrol.WrapperMode{chartcontrol.WrapperModeEnabled, chartcontrol.WrapperModeDisabled, chartcontrol.WrapperModeHidden, chartcontrol.WrapperModeOmitted} {
		config.Options.Controls.Mode = mode
		var output bytes.Buffer
		if err := Radar(config).Render(context.Background(), &output); err != nil {
			t.Fatalf("mode %q Render() error = %v", mode, err)
		}
		markup := output.String()
		if mode == chartcontrol.WrapperModeOmitted {
			if strings.Contains(markup, `data-goshtoso-chart-wrapper=`) || !strings.Contains(markup, `data-radar-exact-values`) {
				t.Errorf("omitted wrapper markup = %q", markup)
			}
			continue
		}
		wantMode := string(mode)
		if mode == chartcontrol.WrapperModeEnabled {
			wantMode = "enabled"
		}
		if !strings.Contains(markup, `data-goshtoso-chart-wrapper-mode="`+wantMode+`"`) {
			t.Errorf("mode %q not propagated", mode)
		}
	}
}

func TestRadarRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	validIndicators := []Indicator{{Name: "A", Max: 10}, {Name: "B", Max: 20}}
	validData := []Data{{Name: "Current", Values: []float64{1, 2}}}
	tests := map[string]struct {
		cfg       Config
		wantError string
	}{
		"missing label":          {cfg: Config{}, wantError: "radar chart label is required"},
		"missing indicators":     {cfg: Config{Label: "Profile"}, wantError: "radar chart indicators are required"},
		"missing indicator name": {cfg: Config{Label: "Profile", Indicators: []Indicator{{Max: 10}}}, wantError: "radar chart indicator 0 name is required"},
		"nonpositive maximum":    {cfg: Config{Label: "Profile", Indicators: []Indicator{{Name: "A"}}}, wantError: `radar chart indicator "A" maximum must be positive`},
		"nonfinite maximum":      {cfg: Config{Label: "Profile", Indicators: []Indicator{{Name: "A", Max: float32(math.Inf(1))}}}, wantError: `radar chart indicator "A" maximum must be positive`},
		"missing series":         {cfg: Config{Label: "Profile", Indicators: validIndicators}, wantError: "radar chart series is required"},
		"missing series name":    {cfg: Config{Label: "Profile", Indicators: validIndicators, Series: []Series{{Data: validData}}}, wantError: "radar chart series 0 name is required"},
		"missing series data":    {cfg: Config{Label: "Profile", Indicators: validIndicators, Series: []Series{{Name: "Profile"}}}, wantError: `radar chart series "Profile" data is required`},
		"missing data name":      {cfg: Config{Label: "Profile", Indicators: validIndicators, Series: []Series{{Name: "Profile", Data: []Data{{Values: []float64{1, 2}}}}}}, wantError: `radar chart series "Profile" data point 0 name is required`},
		"empty vector":           {cfg: Config{Label: "Profile", Indicators: validIndicators, Series: []Series{{Name: "Profile", Data: []Data{{Name: "Current"}}}}}, wantError: `radar chart series "Profile" data "Current" values are required`},
		"misaligned vector":      {cfg: Config{Label: "Profile", Indicators: validIndicators, Series: []Series{{Name: "Profile", Data: []Data{{Name: "Current", Values: []float64{1}}}}}}, wantError: `radar chart series "Profile" data "Current" has 1 values for 2 indicators`},
		"nonfinite vector":       {cfg: Config{Label: "Profile", Indicators: validIndicators, Series: []Series{{Name: "Profile", Data: []Data{{Name: "Current", Values: []float64{1, math.NaN()}}}}}}, wantError: `radar chart series "Profile" data "Current" value 1 must be finite`},
		"unsupported shape":      {cfg: Config{Label: "Profile", Indicators: validIndicators, Series: []Series{{Name: "Profile", Data: validData}}, Coordinate: CoordinateOptions{Shape: "star"}}, wantError: `radar chart shape "star" is not supported`},
		"negative split number":  {cfg: Config{Label: "Profile", Indicators: validIndicators, Series: []Series{{Name: "Profile", Data: validData}}, Coordinate: CoordinateOptions{SplitNumber: -1}}, wantError: "radar chart split number must be nonnegative"},
		"invalid line opacity":   {cfg: Config{Label: "Profile", Indicators: validIndicators, Series: []Series{{Name: "Profile", Data: validData}}, Coordinate: CoordinateOptions{SplitLine: &SplitLineOptions{Style: &chart.LineStyle{Opacity: chart.Float(1.1)}}}}, wantError: "radar chart split-line opacity must be between 0 and 1"},
		"invalid line width":     {cfg: Config{Label: "Profile", Indicators: validIndicators, Series: []Series{{Name: "Profile", Data: validData}}, Coordinate: CoordinateOptions{SplitLine: &SplitLineOptions{Style: &chart.LineStyle{Width: -1}}}}, wantError: "radar chart split-line width must be nonnegative"},
		"invalid line type":      {cfg: Config{Label: "Profile", Indicators: validIndicators, Series: []Series{{Name: "Profile", Data: validData}}, Coordinate: CoordinateOptions{SplitLine: &SplitLineOptions{Style: &chart.LineStyle{Type: "wave"}}}}, wantError: `radar chart split-line type "wave" is not supported`},
		"invalid legend mode":    {cfg: Config{Label: "Profile", Indicators: validIndicators, Series: []Series{{Name: "Profile", Data: validData}}, Options: chart.ChartOptions{Legend: &chart.LegendOptions{SelectionMode: "exclusive"}}}, wantError: `legend selection mode "exclusive" is not supported`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			err := Radar(test.cfg).Render(context.Background(), &output)
			if err == nil || err.Error() != test.wantError {
				t.Fatalf("Render() error = %v, want %q", err, test.wantError)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %d bytes for invalid config", output.Len())
			}
		})
	}
}
