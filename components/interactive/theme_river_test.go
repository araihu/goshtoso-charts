package interactive

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func TestThemeRiverRendersTypedAlignedStreamsAndExactValues(t *testing.T) {
	t.Parallel()
	start, end, bottom := 4.0, 12.0, 10.0
	cfg := validThemeRiverConfig()
	cfg.Caption = "Six aligned temporal streams."
	cfg.Layout = ThemeRiverLayout{BottomPercent: &bottom}
	cfg.BoundaryGap = ThemeRiverBoundaryGap{StartPercent: &start, EndPercent: &end}
	cfg.LabelOptions = &LabelOptions{Show: Bool(true), Position: "inside", Color: "#ffffff", FontSize: 11}
	cfg.Style = charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#654321"}, Class: "caller-class"}
	cfg.Streams[1].Color = "#123456"
	cfg.RootAttrs = templ.Attributes{"id": "river", "data-purpose": "temporal"}
	cfg.Options = ChartOptions{
		Title:     &TitleOptions{Text: "ThemeRiver-SingleAxis-Time"},
		Legend:    &LegendOptions{Show: Bool(true), Top: "top"},
		Tooltip:   &TooltipOptions{Show: Bool(true), Trigger: "axis"},
		Animation: Bool(false),
		Controls:  chartcontrol.Options{Fullscreen: true, Collapsible: true},
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
		`data-goshtoso-chart-expand`, `data-goshtoso-chart-control="collapse"`,
		`data-goshtoso-chart-control="fullscreen"`, `data-goshtoso-chart-export="png"`,
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
		`data-goshtoso-chart-expand`, `data-goshtoso-chart-export="png"`,
		`goshtoso-charts-theme-river`, `aspect-ratio: 9 / 5`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("default markup missing %q", want)
		}
	}
	for _, unwanted := range []string{`data-goshtoso-chart-control="collapse"`, `data-goshtoso-chart-control="fullscreen"`, `__goshtosoChartsThemeRiverRuntime`, `echarts.dispose`} {
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
	streams := []ThemeRiverStream{{Name: "DQ", Points: make([]ThemeRiverPoint, maxThemeRiverDetailRows+5)}}
	for index := range streams[0].Points {
		streams[0].Points[index] = ThemeRiverPoint{Time: time.Date(2015, 1, 1+index, 0, 0, 0, 0, time.UTC), Value: float64(index)}
	}
	rows := themeRiverDetailRows(streams, maxThemeRiverDetailRows)
	if len(rows.Rows) != maxThemeRiverDetailRows || rows.Omitted != 5 {
		t.Fatalf("bounded rows = %d omitted = %d", len(rows.Rows), rows.Omitted)
	}
}

func TestThemeRiverRejectsInvalidDataAndOptions(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		mutate func(*ThemeRiverConfig)
		want   string
	}{
		"missing label":        {func(cfg *ThemeRiverConfig) { cfg.Label = " " }, "theme river chart label is required"},
		"missing streams":      {func(cfg *ThemeRiverConfig) { cfg.Streams = nil }, "theme river chart streams are required"},
		"missing stream name":  {func(cfg *ThemeRiverConfig) { cfg.Streams[0].Name = "" }, "theme river chart stream 0 name is required"},
		"duplicate stream":     {func(cfg *ThemeRiverConfig) { cfg.Streams[1].Name = "DQ" }, `theme river chart stream "DQ" is duplicated`},
		"missing points":       {func(cfg *ThemeRiverConfig) { cfg.Streams[0].Points = nil }, `theme river chart stream "DQ" points are required`},
		"point count mismatch": {func(cfg *ThemeRiverConfig) { cfg.Streams[1].Points = cfg.Streams[1].Points[:1] }, `theme river chart stream "TY" has 1 points for 2 aligned dates`},
		"missing time":         {func(cfg *ThemeRiverConfig) { cfg.Streams[0].Points[0].Time = time.Time{} }, `theme river chart stream "DQ" point 0 time is required`},
		"nonfinite value":      {func(cfg *ThemeRiverConfig) { cfg.Streams[0].Points[0].Value = math.NaN() }, `theme river chart stream "DQ" point 0 value must be finite`},
		"negative value":       {func(cfg *ThemeRiverConfig) { cfg.Streams[0].Points[0].Value = -1 }, `theme river chart stream "DQ" point 0 value must be nonnegative`},
		"unsorted dates":       {func(cfg *ThemeRiverConfig) { cfg.Streams[0].Points[1].Time = cfg.Streams[0].Points[0].Time }, `theme river chart stream "DQ" dates must be strictly increasing`},
		"unaligned dates": {func(cfg *ThemeRiverConfig) {
			cfg.Streams[1].Points[1].Time = cfg.Streams[1].Points[1].Time.Add(time.Hour)
		}, `theme river chart stream "TY" point 1 date is not aligned`},
		"partial gap": {func(cfg *ThemeRiverConfig) { cfg.BoundaryGap.StartPercent = Float(2) }, "theme river chart boundary gap requires both start and end percentages"},
		"bad gap": {func(cfg *ThemeRiverConfig) {
			cfg.BoundaryGap = ThemeRiverBoundaryGap{StartPercent: Float(2), EndPercent: Float(101)}
		}, "theme river chart boundary gap must be between 0 and 100"},
		"bad layout":     {func(cfg *ThemeRiverConfig) { cfg.Layout.BottomPercent = Float(-1) }, "theme river chart layout bottom percentage must be between 0 and 100"},
		"bad label":      {func(cfg *ThemeRiverConfig) { cfg.LabelOptions = &LabelOptions{FontSize: -1} }, "theme river chart label font size must be nonnegative"},
		"Cartesian axis": {func(cfg *ThemeRiverConfig) { cfg.Options.XAxis = &AxisOptions{} }, "theme river chart Cartesian axes are not supported"},
		"reserved attr":  {func(cfg *ThemeRiverConfig) { cfg.RootAttrs = templ.Attributes{"role": "group"} }, `theme river chart root attribute "role" is reserved`},
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
		})
	}
}

func validThemeRiverConfig() ThemeRiverConfig {
	d0 := time.Date(2015, time.November, 8, 0, 0, 0, 0, time.UTC)
	d1 := d0.AddDate(0, 0, 1)
	return ThemeRiverConfig{
		Label: "ThemeRiver-SingleAxis-Time",
		Streams: []ThemeRiverStream{
			{Name: "DQ", Class: "stream-dq", Points: []ThemeRiverPoint{{Time: d0, Value: 10}, {Time: d1, Value: 15}}},
			{Name: "TY", Class: "stream-ty", Points: []ThemeRiverPoint{{Time: d0, Value: 35}, {Time: d1, Value: 36}}},
		},
	}
}

func renderThemeRiver(t *testing.T, instance Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	return output.String()
}
