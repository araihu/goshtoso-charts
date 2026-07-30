package wordcloud

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func TestWordCloudRendersTypedWordsOptionsAndExactValues(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Caption = "Twenty weighted search terms."
	cfg.Series.Options = SeriesOptions{
		Shape:     ShapeStar,
		SizeRange: &SizeRange{Min: 14, Max: 80},
		Rotation:  &Rotation{MinDegrees: -45, MaxDegrees: 45, StepDegrees: 15},
		GridSize:  chart.Int(9), DrawOutOfBound: chart.Bool(true), LayoutAnimation: chart.Bool(false),
		Layout: Layout{
			Horizontal: HorizontalCenter, Vertical: VerticalCenter,
			WidthPercent: chart.Float(75), HeightPercent: chart.Float(80),
		},
	}
	cfg.Series.Words[0].Class = "retail-anchor"
	cfg.Series.Words[1].Class = "retail"
	cfg.Series.Words[1].Color = "#123456"
	cfg.Style = charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#654321"}, Class: "caller-class"}
	cfg.RootAttrs = templ.Attributes{"id": "search-terms", "data-purpose": "weighted-terms"}
	cfg.Options = chart.ChartOptions{
		Title:     &chart.TitleOptions{Text: "basic WordCloud example"},
		Tooltip:   &chart.TooltipOptions{Show: chart.Bool(true), Trigger: "item"},
		Animation: chart.Bool(false), Controls: chartcontrol.Options{Fullscreen: true},
		Export: &chartcontrol.ExportOptions{Filename: "word-cloud"},
	}

	instance := WordCloud(cfg)
	if instance.Kind() != chartcomponents.KindInteractiveWordCloud {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	markup := renderWordCloud(t, instance)
	for _, want := range []string{
		`class="goshtoso-charts-interactive goshtoso-charts-palette goshtoso-charts-palette-araihu goshtoso-charts-word-cloud caller-class"`,
		`role="img"`, `aria-label="basic WordCloud example"`, `id="search-terms"`, `data-purpose="weighted-terms"`,
		`style="width:100%;height:500px;"`, `"type":"wordCloud"`, `"shape":"star"`,
		`"sizeRange":[14,80]`, `"rotationRange":[-45,45]`, `"rotationStep":15`, `"gridSize":9`,
		`"drawOutOfBound":true`, `"layoutAnimation":false`, `"left":"center"`, `"top":"center"`,
		`"width":"75%"`, `"height":"80%"`, `"trigger":"item"`, `"animation":false`,
		`{"name":"Sam S Club","value":10000,"className":"retail-anchor"}`,
		`{"name":"Macys","value":6181,"className":"retail","sourceColor":"#123456","textStyle":{"color":"#123456"}}`,
		`data-goshtoso-charts-theme-runtime`, `series.type === "wordCloud"`,
		`Twenty weighted search terms.`, `>Exact word values</summary>`, `scope="col">Word</th>`,
		`scope="col">Value</th>`, `scope="col">Class</th>`, `>Sam S Club</th>`, `>10000</td>`, `>retail-anchor</td>`,
		`data-goshtoso-chart-expand`,
		`-fullscreen-action`, `exportFromMenu($el, &#34;png&#34;)`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	for _, unwanted := range []string{"Math.random", "go-echarts", "echarts-wordcloud"} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("rendered markup contains %q", unwanted)
		}
	}
}

func TestWordCloudDefaultsToSharedExpandDirectPNGAndOneRuntime(t *testing.T) {
	t.Parallel()
	markup := renderWordCloud(t, WordCloud(validConfig()))
	for _, want := range []string{
		`data-goshtoso-chart-expand`, `exportFromMenu($el, &#34;png&#34;)`,
		`goshtoso-charts-word-cloud`, `aspect-ratio: 9 / 5`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("default markup missing %q", want)
		}
	}
	for _, unwanted := range []string{`-fullscreen-action"`, `__goshtosoChartsWordCloudRuntime`, `echarts.dispose`, `Math.random`} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("default markup contains %q", unwanted)
		}
	}
	if strings.Count(markup, `data-goshtoso-charts-theme-runtime`) != 1 {
		t.Fatal("WordCloud did not use exactly one shared theme runtime")
	}
}

func TestWordCloudBoundsExactValueList(t *testing.T) {
	t.Parallel()
	words := make([]Word, maxDetailRows+5)
	for index := range words {
		words[index] = Word{Name: strings.Repeat("w", index+1), Value: float64(index)}
	}
	rows := detailRows(words, maxDetailRows)
	if len(rows.Rows) != maxDetailRows || rows.Omitted != 5 {
		t.Fatalf("bounded rows = %d omitted = %d", len(rows.Rows), rows.Omitted)
	}
}

func TestWordCloudRejectsInvalidDataAndOptions(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		mutate func(*Config)
		want   string
	}{
		"missing label":        {func(cfg *Config) { cfg.Label = " " }, "word cloud chart label is required"},
		"missing series name":  {func(cfg *Config) { cfg.Series.Name = " " }, "word cloud chart series name is required"},
		"missing words":        {func(cfg *Config) { cfg.Series.Words = nil }, "word cloud chart words are required"},
		"missing word name":    {func(cfg *Config) { cfg.Series.Words[0].Name = "" }, "word cloud chart word 0 name is required"},
		"duplicate word":       {func(cfg *Config) { cfg.Series.Words[1].Name = "Sam S Club" }, `word cloud chart word "Sam S Club" is duplicated`},
		"nonfinite value":      {func(cfg *Config) { cfg.Series.Words[0].Value = math.NaN() }, `word cloud chart word "Sam S Club" value must be finite`},
		"negative value":       {func(cfg *Config) { cfg.Series.Words[0].Value = -1 }, `word cloud chart word "Sam S Club" value must be nonnegative`},
		"invalid shape":        {func(cfg *Config) { cfg.Series.Options.Shape = "oval" }, `word cloud chart shape "oval" is not supported`},
		"invalid size minimum": {func(cfg *Config) { cfg.Series.Options.SizeRange = &SizeRange{Min: 0, Max: 80} }, "word cloud chart size range must be finite positive values"},
		"reversed size range":  {func(cfg *Config) { cfg.Series.Options.SizeRange = &SizeRange{Min: 80, Max: 14} }, "word cloud chart size range minimum must not exceed maximum"},
		"oversize range":       {func(cfg *Config) { cfg.Series.Options.SizeRange = &SizeRange{Min: 14, Max: 513} }, "word cloud chart size range must be between 1 and 512 pixels"},
		"nonfinite rotation": {func(cfg *Config) {
			cfg.Series.Options.Rotation = &Rotation{MinDegrees: math.Inf(-1), MaxDegrees: 45, StepDegrees: 15}
		}, "word cloud chart rotation values must be finite"},
		"reversed rotation": {func(cfg *Config) {
			cfg.Series.Options.Rotation = &Rotation{MinDegrees: 45, MaxDegrees: -45, StepDegrees: 15}
		}, "word cloud chart rotation minimum must not exceed maximum"},
		"invalid rotation step": {func(cfg *Config) {
			cfg.Series.Options.Rotation = &Rotation{MinDegrees: -45, MaxDegrees: 45}
		}, "word cloud chart rotation step must be positive"},
		"rotation outside bounds": {func(cfg *Config) {
			cfg.Series.Options.Rotation = &Rotation{MinDegrees: -181, MaxDegrees: 45, StepDegrees: 15}
		}, "word cloud chart rotation range must stay between -180 and 180 degrees"},
		"invalid grid":              {func(cfg *Config) { cfg.Series.Options.GridSize = chart.Int(0) }, "word cloud chart grid size must be between 1 and 64"},
		"invalid horizontal layout": {func(cfg *Config) { cfg.Series.Options.Layout.Horizontal = "middle" }, `word cloud chart horizontal layout "middle" is not supported`},
		"invalid vertical layout":   {func(cfg *Config) { cfg.Series.Options.Layout.Vertical = "middle" }, `word cloud chart vertical layout "middle" is not supported`},
		"invalid width":             {func(cfg *Config) { cfg.Series.Options.Layout.WidthPercent = chart.Float(101) }, "word cloud chart layout width percentage must be between 1 and 100"},
		"invalid tooltip":           {func(cfg *Config) { cfg.Options.Tooltip = &chart.TooltipOptions{Trigger: "axis"} }, `word cloud chart tooltip trigger "axis" is not supported`},
		"legend":                    {func(cfg *Config) { cfg.Options.Legend = &chart.LegendOptions{} }, "word cloud chart legend is not supported"},
		"Cartesian axis":            {func(cfg *Config) { cfg.Options.XAxis = &chart.AxisOptions{} }, "word cloud chart Cartesian axes are not supported"},
		"reserved attr":             {func(cfg *Config) { cfg.RootAttrs = templ.Attributes{"role": "group"} }, `word cloud chart root attribute "role" is reserved`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := validConfig()
			test.mutate(&cfg)
			var output bytes.Buffer
			err := WordCloud(cfg).Render(context.Background(), &output)
			if err == nil || err.Error() != test.want {
				t.Fatalf("Render() error = %v, want %q", err, test.want)
			}
		})
	}
}

func validConfig() Config {
	return Config{
		Label: "basic WordCloud example",
		Series: Series{
			Name: "wordcloud",
			Words: []Word{
				{Name: "Sam S Club", Value: 10000},
				{Name: "Macys", Value: 6181},
			},
		},
	}
}

func renderWordCloud(t *testing.T, instance Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	return output.String()
}
