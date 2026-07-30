package themeriver

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

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

var themeRiverIDPattern = regexp.MustCompile(`id="([A-Za-z]{12})"`)

func TestThemeRiverNormalizedRenderHashes(t *testing.T) {
	t.Parallel()
	d0 := time.Date(2015, time.November, 8, 0, 0, 0, 0, time.UTC)
	d1 := d0.AddDate(0, 0, 1)
	t0 := time.Date(2015, time.November, 8, 9, 30, 0, 123, time.FixedZone("UTC-3", -3*60*60))
	t1 := t0.Add(6 * time.Hour)
	start, end, bottom := 4.0, 12.0, 10.0
	left, right, top, bottom2 := 1.5, 2.5, 3.5, 4.5
	tests := map[string]struct {
		config Config
		want   string
	}{
		"default aligned": {
			config: validThemeRiverConfig(),
			want:   "ac532ba4d25c57a6a5e61ab4fb5cc998964a1b8e2425d2020834f4d576b4a610",
		},
		"custom boundary wrapper": {
			config: Config{
				Label: "ThemeRiver-SingleAxis-Time", Caption: "Six aligned temporal streams.",
				Streams: []Stream{
					{Name: "DQ", Class: "stream-dq", Points: []Point{{Time: d0, Value: 10}, {Time: d1, Value: 15}}},
					{Name: "TY", Class: "stream-ty", Color: "#123456", Points: []Point{{Time: d0, Value: 35}, {Time: d1, Value: 36}}},
				},
				Layout: Layout{BottomPercent: &bottom}, BoundaryGap: BoundaryGap{StartPercent: &start, EndPercent: &end},
				LabelOptions: &chart.LabelOptions{Show: chart.Bool(true), Position: "inside", Color: "#ffffff", FontSize: 11},
				Options:      chart.ChartOptions{Title: &chart.TitleOptions{Text: "ThemeRiver-SingleAxis-Time"}, Legend: &chart.LegendOptions{Show: chart.Bool(true), Top: "top"}, Tooltip: &chart.TooltipOptions{Show: chart.Bool(true), Trigger: "axis"}, Animation: chart.Bool(false), Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "theme-river"}},
				Style:        charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#654321"}, Class: "caller-class"}, RootAttrs: templ.Attributes{"id": "river"},
			},
			want: "1d3c898e7ad28652b4c3cbe6f7ba374d4db494a5ec409330676f7073ba55516b",
		},
		"timestamp layout omitted": {
			config: Config{
				Label: "Hourly streams", Caption: "Exact instants.", Width: "760px", Height: "340px",
				Streams: []Stream{
					{Name: "Alpha", Color: "#abcdef", Points: []Point{{Time: t0, Value: 0}, {Time: t1, Value: 1.25}}},
					{Name: "Beta", Points: []Point{{Time: t0, Value: 12.5}, {Time: t1, Value: 20}}},
				},
				Layout:  Layout{LeftPercent: &left, RightPercent: &right, TopPercent: &top, BottomPercent: &bottom2},
				Options: chart.ChartOptions{Animation: chart.Bool(false), Controls: chartcontrol.Options{Mode: chartcontrol.WrapperModeOmitted}, Export: &chartcontrol.ExportOptions{Filename: "hourly-streams"}},
				Style:   charttheme.Style{Palette: charttheme.PalettePastel, Class: "caller-river"}, RootAttrs: templ.Attributes{"data-purpose": "temporal"},
			},
			want: "7d360e273333fcc918397f369c41d776c8401ffd4095158c13f5e86bac264550",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			markup := renderThemeRiver(t, ThemeRiver(test.config))
			match := themeRiverIDPattern.FindStringSubmatch(markup)
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

func TestThemeRiverRendersTypedAlignedStreamsAndExactValues(t *testing.T) {
	t.Parallel()
	start, end, bottom := 4.0, 12.0, 10.0
	cfg := validThemeRiverConfig()
	cfg.Caption = "Six aligned temporal streams."
	cfg.Layout = Layout{BottomPercent: &bottom}
	cfg.BoundaryGap = BoundaryGap{StartPercent: &start, EndPercent: &end}
	cfg.LabelOptions = &chart.LabelOptions{Show: chart.Bool(true), Position: "inside", Color: "#ffffff", FontSize: 11}
	cfg.Style = charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#654321"}, Class: "caller-class"}
	cfg.Streams[1].Color = "#123456"
	cfg.RootAttrs = templ.Attributes{"id": "river", "data-purpose": "temporal"}
	cfg.Options = chart.ChartOptions{
		Title:     &chart.TitleOptions{Text: "ThemeRiver-SingleAxis-Time"},
		Legend:    &chart.LegendOptions{Show: chart.Bool(true), Top: "top"},
		Tooltip:   &chart.TooltipOptions{Show: chart.Bool(true), Trigger: "axis"},
		Animation: chart.Bool(false),
		Controls:  chartcontrol.Options{Fullscreen: true},
		Export:    &chartcontrol.ExportOptions{Filename: "theme-river"},
	}

	instance := ThemeRiver(cfg)
	if instance.Kind() != chartcomponents.KindInteractiveThemeRiver {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	markup := renderThemeRiver(t, instance)
	for _, want := range []string{
		`class="goshtoso-charts-interactive goshtoso-charts-palette goshtoso-charts-palette-araihu goshtoso-charts-theme-river caller-class"`,
		`role="img"`, `aria-label="ThemeRiver-SingleAxis-Time"`, `id="river"`, `data-purpose="temporal"`,
		`style="width:100%;height:500px;"`, `"type":"themeRiver","boundaryGap":["4%","12%"]`,
		`["2015/11/08",10,"DQ"]`, `["2015/11/09",15,"DQ"]`,
		`["2015/11/08",35,"TY"]`, `"singleAxis":{`, `"bottom":"10%"`, `"type":"time"`,
		`"text":"ThemeRiver-SingleAxis-Time"`, `"trigger":"axis"`, `"animation":false`,
		`"color":["#654321","#123456"`, `"label":{"show":true`, `"position":"inside"`,
		`data-goshtoso-charts-theme-runtime`, `singleAxis: repeat(current.singleAxis, axis)`,
		`var resizeObserver = window.ResizeObserver ? new ResizeObserver`,
		`Six aligned temporal streams.`, `>Exact stream values</summary>`,
		`scope="col">Date</th>`, `scope="col">Stream</th>`, `scope="col">Value</th>`, `scope="col">Class</th>`,
		`>2015/11/08</th>`, `>DQ</td>`, `>10</td>`, `>stream-dq</td>`,
		`data-goshtoso-chart-expand`, `-fullscreen-action`, `exportFromMenu($el, &#34;png&#34;)`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestThemeRiverDefaultsToSharedExpandAndDirectPNG(t *testing.T) {
	t.Parallel()
	markup := renderThemeRiver(t, ThemeRiver(validThemeRiverConfig()))
	for _, want := range []string{
		`data-goshtoso-chart-expand`, `exportFromMenu($el, &#34;png&#34;)`,
		`goshtoso-charts-theme-river`, `aspect-ratio: 9 / 5`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("default markup missing %q", want)
		}
	}
	for _, unwanted := range []string{`-fullscreen-action"`, `__goshtosoChartsThemeRiverRuntime`, `echarts.dispose`} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("default markup contains %q", unwanted)
		}
	}
	if strings.Count(markup, `data-goshtoso-charts-theme-runtime`) != 1 {
		t.Fatal("ThemeRiver did not use exactly one shared theme runtime")
	}
}

func TestThemeRiverBoundsExactValueTable(t *testing.T) {
	t.Parallel()
	streams := []Stream{{Name: "DQ", Points: make([]Point, maxThemeRiverDetailRows+5)}}
	for index := range streams[0].Points {
		streams[0].Points[index] = Point{Time: time.Date(2015, 1, 1+index, 0, 0, 0, 0, time.UTC), Value: float64(index)}
	}
	rows := themeRiverDetailRows(streams, maxThemeRiverDetailRows)
	if len(rows.Rows) != maxThemeRiverDetailRows || rows.Omitted != 5 {
		t.Fatalf("bounded rows = %d omitted = %d", len(rows.Rows), rows.Omitted)
	}
}

func TestThemeRiverRejectsInvalidDataAndOptions(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		mutate func(*Config)
		want   string
	}{
		"missing label":        {func(cfg *Config) { cfg.Label = " " }, "theme river chart label is required"},
		"missing streams":      {func(cfg *Config) { cfg.Streams = nil }, "theme river chart streams are required"},
		"missing stream name":  {func(cfg *Config) { cfg.Streams[0].Name = "" }, "theme river chart stream 0 name is required"},
		"duplicate stream":     {func(cfg *Config) { cfg.Streams[1].Name = "DQ" }, `theme river chart stream "DQ" is duplicated`},
		"missing points":       {func(cfg *Config) { cfg.Streams[0].Points = nil }, `theme river chart stream "DQ" points are required`},
		"point count mismatch": {func(cfg *Config) { cfg.Streams[1].Points = cfg.Streams[1].Points[:1] }, `theme river chart stream "TY" has 1 points for 2 aligned dates`},
		"missing time":         {func(cfg *Config) { cfg.Streams[0].Points[0].Time = time.Time{} }, `theme river chart stream "DQ" point 0 time is required`},
		"nonfinite value":      {func(cfg *Config) { cfg.Streams[0].Points[0].Value = math.NaN() }, `theme river chart stream "DQ" point 0 value must be finite`},
		"negative value":       {func(cfg *Config) { cfg.Streams[0].Points[0].Value = -1 }, `theme river chart stream "DQ" point 0 value must be nonnegative`},
		"unsorted dates":       {func(cfg *Config) { cfg.Streams[0].Points[1].Time = cfg.Streams[0].Points[0].Time }, `theme river chart stream "DQ" dates must be strictly increasing`},
		"unaligned dates": {func(cfg *Config) {
			cfg.Streams[1].Points[1].Time = cfg.Streams[1].Points[1].Time.Add(time.Hour)
		}, `theme river chart stream "TY" point 1 date is not aligned`},
		"partial gap": {func(cfg *Config) { cfg.BoundaryGap.StartPercent = chart.Float(2) }, "theme river chart boundary gap requires both start and end percentages"},
		"bad gap": {func(cfg *Config) {
			cfg.BoundaryGap = BoundaryGap{StartPercent: chart.Float(2), EndPercent: chart.Float(101)}
		}, "theme river chart boundary gap must be between 0 and 100"},
		"bad layout":     {func(cfg *Config) { cfg.Layout.BottomPercent = chart.Float(-1) }, "theme river chart layout bottom percentage must be between 0 and 100"},
		"bad label":      {func(cfg *Config) { cfg.LabelOptions = &chart.LabelOptions{FontSize: -1} }, "theme river chart label font size must be nonnegative"},
		"Cartesian axis": {func(cfg *Config) { cfg.Options.XAxis = &chart.AxisOptions{} }, "theme river chart Cartesian axes are not supported"},
		"reserved attr":  {func(cfg *Config) { cfg.RootAttrs = templ.Attributes{"role": "group"} }, `theme river chart root attribute "role" is reserved`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := validThemeRiverConfig()
			test.mutate(&cfg)
			var output bytes.Buffer
			err := ThemeRiver(cfg).Render(context.Background(), &output)
			if err == nil || err.Error() != test.want {
				t.Fatalf("Render() error = %v, want %q", err, test.want)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %d bytes for invalid config", output.Len())
			}
		})
	}
}

func validThemeRiverConfig() Config {
	d0 := time.Date(2015, time.November, 8, 0, 0, 0, 0, time.UTC)
	d1 := d0.AddDate(0, 0, 1)
	return Config{
		Label: "ThemeRiver-SingleAxis-Time",
		Streams: []Stream{
			{Name: "DQ", Class: "stream-dq", Points: []Point{{Time: d0, Value: 10}, {Time: d1, Value: 15}}},
			{Name: "TY", Class: "stream-ty", Points: []Point{{Time: d0, Value: 35}, {Time: d1, Value: 36}}},
		},
	}
}

func renderThemeRiver(t *testing.T, instance chart.Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	return output.String()
}
