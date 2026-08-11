package parallel

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

var parallelChartIDPattern = regexp.MustCompile(`id="([A-Za-z]{12})"`)

func TestParallelNormalizedRenderHashes(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		config Config
		want   string
	}{
		"mixed defaults": {
			config: validParallelConfig(),
			want:   "c33a995cbeefbd74370d95fbe942f19a3152624994440cd5834a1e9a330ee0e7",
		},
		"typed layout and presentation": {
			config: Config{
				Label: "Air quality profiles", Caption: "Daily measurements grouped by city.",
				Dimensions: []Dimension{
					{Name: "Date", Range: &Range{Min: chart.Float(1), Max: chart.Float(31)}, Inverse: true, NameLocation: NameStart, Label: AxisLabel{Show: chart.Bool(true), Rotate: 20, Margin: 9, FontSize: 11}, Line: AxisLine{Show: chart.Bool(true), Color: "#334155", Width: 2, Type: "dashed"}},
					{Name: "Level", Categories: []string{"Good", "Moderate", "Heavily"}},
				},
				Series: []Series{{
					Name: "Beijing", Class: "city-north", Color: "#123456",
					Options:      SeriesOptions{Line: LineOptions{Width: 2, Type: "solid", Opacity: chart.Float(0.8)}, Smooth: chart.Bool(true), InactiveOpacity: chart.Float(0.12)},
					Observations: []Observation{{Name: "Day 1", Values: []Value{Number(1), Category("Moderate")}, Class: "moderate"}, {Name: "Day 2", Values: []Value{Number(2), Category("Good")}, Color: "#abcdef"}},
				}},
				Layout: Layout{LeftPercent: chart.Float(15), RightPercent: chart.Float(13), TopPercent: chart.Float(20), BottomPercent: chart.Float(10)},
				Width:  "100%", Height: "500px",
				Options: chart.ChartOptions{Title: &chart.TitleOptions{Text: "Multi Series"}, Legend: &chart.LegendOptions{Show: chart.Bool(true)}, Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "air-quality-profiles"}},
				Style:   charttheme.Style{Palette: charttheme.PaletteAraiHu, Class: "max-w-full"}, RootAttrs: templ.Attributes{"id": "air-quality", "data-chart-purpose": "multivariate"},
			},
			want: "996d53e4bdb1dfd3de01ca34c33f4db9831b345e620e570b8fad9738a564fef3",
		},
		"log and middle labels": {
			config: Config{
				Label: "Log profile",
				Dimensions: []Dimension{
					{Name: "Latency", Range: &Range{Min: chart.Float(0.1), Max: chart.Float(100)}, Scale: ScaleLog, NameLocation: NameMiddle},
					{Name: "Throughput", Range: &Range{Min: chart.Float(1), Max: chart.Float(1000)}, Scale: ScaleLog, NameLocation: NameEnd},
				},
				Series: []Series{{Name: "Service", Observations: []Observation{{Name: "Current", Values: []Value{Number(10), Number(100)}}}}},
			},
			want: "1ed2be295a2b385c86eb1e34032f75c261242423da2d992b63bcb106b75fb2e8",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			if err := Parallel(test.config).Render(context.Background(), &output); err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			markup := output.String()
			match := parallelChartIDPattern.FindStringSubmatch(markup)
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

func TestParallelRendersTypedDimensionsSeriesLayoutAndExactValues(t *testing.T) {
	t.Parallel()
	instance := Parallel(Config{
		Label:   "Air quality profiles",
		Caption: "Daily measurements grouped by city.",
		Dimensions: []Dimension{
			{
				Name: "Date", Range: &Range{Min: chart.Float(1), Max: chart.Float(31)}, Inverse: true, NameLocation: NameStart,
				Label: AxisLabel{Show: chart.Bool(true), Rotate: 20, Margin: 9, FontSize: 11},
				Line:  AxisLine{Show: chart.Bool(true), Color: "#334155", Width: 2, Type: "dashed"},
			},
			{Name: "Level", Categories: []string{"Good", "Moderate", "Heavily"}},
		},
		Series: []Series{{
			Name: "Beijing", Class: "city-north", Color: "#123456",
			Options: SeriesOptions{
				Line:   LineOptions{Width: 2, Type: "solid", Opacity: chart.Float(0.8)},
				Smooth: chart.Bool(true), InactiveOpacity: chart.Float(0.12),
			},
			Observations: []Observation{
				{Name: "Day 1", Values: []Value{Number(1), Category("Moderate")}, Class: "moderate"},
				{Name: "Day 2", Values: []Value{Number(2), Category("Good")}, Color: "#abcdef"},
			},
		}},
		Layout: Layout{LeftPercent: chart.Float(15), RightPercent: chart.Float(13), TopPercent: chart.Float(20), BottomPercent: chart.Float(10)},
		Width:  "100%", Height: "500px",
		Options: chart.ChartOptions{
			Title: &chart.TitleOptions{Text: "Multi Series"}, Legend: &chart.LegendOptions{Show: chart.Bool(true)},
			Controls: chartcontrol.Options{Fullscreen: true},
			Export:   &chartcontrol.ExportOptions{Filename: "air-quality-profiles"},
		},
		Style:     charttheme.Style{Palette: charttheme.PaletteAraiHu, Class: "max-w-full"},
		RootAttrs: templ.Attributes{"id": "air-quality", "data-chart-purpose": "multivariate"},
	})

	if instance.Kind() != chartcomponents.KindInteractiveParallel {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	markup := renderParallel(t, instance)
	for _, want := range []string{
		`role="img"`, `aria-label="Air quality profiles"`, `id="air-quality"`, `data-chart-purpose="multivariate"`,
		`style="width:100%;height:500px;"`, `"parallel":{"left":"15%","top":"20%","right":"13%","bottom":"10%"}`,
		`"parallelAxis":[{"dim":0,"name":"Date","min":1,"max":31,"inverse":true,"nameLocation":"start"`,
		`"axisLabel":{"show":true,"rotate":20,"margin":9,"fontSize":11}`,
		`"axisLine":{"show":true,"lineStyle":{"color":"#334155","width":2,"type":"dashed"}}`,
		`{"dim":1,"name":"Level","type":"category","data":["Good","Moderate","Heavily"]}`,
		`"name":"Beijing","type":"parallel","className":"city-north","inactiveOpacity":0.12`,
		`"smooth":true`, `"lineStyle":{"color":"#123456","width":2,"type":"solid","opacity":0.8}`,
		`"name":"Day 1","value":[1,"Moderate"],"className":"moderate"`,
		`"name":"Day 2","value":[2,"Good"],"lineStyle":{"color":"#abcdef"}`,
		`"text":"Multi Series"`, `"show":true`, `data-goshtoso-charts-theme-runtime`,
		`data-goshtoso-chart-expand`, `exportFromMenu($el, &#34;png&#34;)`, `-fullscreen-action`,
		`>Exact observations and values</summary>`, `scope="col">Series</th>`, `scope="col">Observation</th>`,
		`scope="col">Date</th>`, `scope="col">Level</th>`, `scope="col">Semantic class</th>`,
		`>Beijing</th>`, `>Day 1</td>`, `>Moderate</td>`, `>city-north moderate</td>`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestParallelDefaultsToExpandAndPNGWithoutOptionalControls(t *testing.T) {
	t.Parallel()
	markup := renderParallel(t, Parallel(validParallelConfig()))
	for _, want := range []string{`data-goshtoso-chart-expand`, `exportFromMenu($el, &#34;png&#34;)`} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	for _, unwanted := range []string{`-fullscreen-action"`} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("rendered markup unexpectedly contains %q", unwanted)
		}
	}
}

func TestParallelColorOverrideDoesNotImplyTransparentLine(t *testing.T) {
	t.Parallel()
	cfg := validParallelConfig()
	cfg.Series[0].Color = "#123456"
	markup := renderParallel(t, Parallel(cfg))
	if !strings.Contains(markup, `"lineStyle":{"color":"#123456"}`) {
		t.Fatal("explicit series color was not serialized")
	}
	if strings.Contains(markup, `"lineStyle":{"color":"#123456","opacity":0}`) {
		t.Fatal("omitted opacity made explicit series color transparent")
	}
}

func TestParallelRejectsInvalidContracts(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		mutate    func(*Config)
		wantError string
	}{
		"label":               {func(cfg *Config) { cfg.Label = " " }, "parallel chart label is required"},
		"dimensions":          {func(cfg *Config) { cfg.Dimensions = cfg.Dimensions[:1] }, "parallel chart requires at least two dimensions"},
		"dimension name":      {func(cfg *Config) { cfg.Dimensions[0].Name = "" }, "dimension 0 name is required"},
		"duplicate dimension": {func(cfg *Config) { cfg.Dimensions[1].Name = cfg.Dimensions[0].Name }, `dimension name "Value" is duplicated`},
		"range":               {func(cfg *Config) { cfg.Dimensions[0].Range = &Range{Min: chart.Float(2), Max: chart.Float(1)} }, "numeric range min must be less than max"},
		"category range":      {func(cfg *Config) { cfg.Dimensions[1].Range = &Range{Min: chart.Float(0), Max: chart.Float(1)} }, "categorical dimensions cannot define a numeric range"},
		"duplicate category":  {func(cfg *Config) { cfg.Dimensions[1].Categories = []string{"A", "A"} }, `category "A" is duplicated`},
		"bad scale":           {func(cfg *Config) { cfg.Dimensions[0].Scale = "power" }, `scale "power" is not supported`},
		"log category":        {func(cfg *Config) { cfg.Dimensions[1].Scale = ScaleLog }, "categorical dimensions cannot use a log scale"},
		"log number": {func(cfg *Config) {
			cfg.Dimensions[0].Scale = ScaleLog
			cfg.Dimensions[0].Range = &Range{Min: chart.Float(0.1), Max: chart.Float(10)}
			cfg.Series[0].Observations[0].Values[0] = Number(0)
		}, "log value must be positive"},
		"axis rotation":    {func(cfg *Config) { cfg.Dimensions[0].Label.Rotate = 91 }, "label rotation must be between -90 and 90"},
		"axis width":       {func(cfg *Config) { cfg.Dimensions[0].Line.Width = -1 }, "axis line width must be finite and nonnegative"},
		"series":           {func(cfg *Config) { cfg.Series = nil }, "parallel chart series is required"},
		"series name":      {func(cfg *Config) { cfg.Series[0].Name = "" }, "series 0 name is required"},
		"observations":     {func(cfg *Config) { cfg.Series[0].Observations = nil }, `series "Series" observations are required`},
		"observation name": {func(cfg *Config) { cfg.Series[0].Observations[0].Name = "" }, `observation 0 name is required`},
		"alignment": {func(cfg *Config) {
			cfg.Series[0].Observations[0].Values = cfg.Series[0].Observations[0].Values[:1]
		}, "has 1 values for 2 dimensions"},
		"numeric type":     {func(cfg *Config) { cfg.Series[0].Observations[0].Values[0] = Category("A") }, "value must be numeric"},
		"category type":    {func(cfg *Config) { cfg.Series[0].Observations[0].Values[1] = Number(1) }, "value must be categorical"},
		"membership":       {func(cfg *Config) { cfg.Series[0].Observations[0].Values[1] = Category("C") }, `category "C" is not declared`},
		"finite":           {func(cfg *Config) { cfg.Series[0].Observations[0].Values[0] = Number(math.NaN()) }, "numeric value must be finite"},
		"out of range":     {func(cfg *Config) { cfg.Series[0].Observations[0].Values[0] = Number(11) }, "exceeds maximum 10"},
		"line opacity":     {func(cfg *Config) { cfg.Series[0].Options.Line.Opacity = chart.Float(1.1) }, "line opacity must be between 0 and 1"},
		"inactive opacity": {func(cfg *Config) { cfg.Series[0].Options.InactiveOpacity = chart.Float(-0.1) }, "inactive opacity must be between 0 and 1"},
		"layout":           {func(cfg *Config) { cfg.Layout.LeftPercent = chart.Float(101) }, "left inset must be between 0 and 100 percent"},
		"layout total":     {func(cfg *Config) { cfg.Layout.LeftPercent, cfg.Layout.RightPercent = chart.Float(50), chart.Float(50) }, "horizontal insets must total less than 100 percent"},
		"root attrs":       {func(cfg *Config) { cfg.RootAttrs = templ.Attributes{"role": "group"} }, `root attribute "role" is reserved`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := validParallelConfig()
			test.mutate(&cfg)
			var output bytes.Buffer
			err := Parallel(cfg).Render(context.Background(), &output)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("Render() error = %v, want containing %q", err, test.wantError)
			}
		})
	}
}

func TestParallelExactValuesAreBounded(t *testing.T) {
	t.Parallel()
	cfg := validParallelConfig()
	cfg.Series[0].Observations = make([]Observation, maxParallelDetailRows+2)
	for index := range cfg.Series[0].Observations {
		cfg.Series[0].Observations[index] = Observation{
			Name:   "row-" + strconv.Itoa(index),
			Values: []Value{Number(float64(index % 10)), Category("A")},
		}
	}
	markup := renderParallel(t, Parallel(cfg))
	if !strings.Contains(markup, "2 additional observations omitted") {
		t.Fatal("bounded exact-value table did not report omitted rows")
	}
}

func validParallelConfig() Config {
	return Config{
		Label: "Parallel",
		Dimensions: []Dimension{
			{Name: "Value", Range: &Range{Min: chart.Float(0), Max: chart.Float(10)}},
			{Name: "Class", Categories: []string{"A", "B"}},
		},
		Series: []Series{{
			Name:         "Series",
			Observations: []Observation{{Name: "Row", Values: []Value{Number(5), Category("A")}}},
		}},
	}
}

func renderParallel(t *testing.T, instance Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return output.String()
}
