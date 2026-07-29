package interactive

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func TestCandlestickRendersTypedOHLCOptionsAndAccessibleTable(t *testing.T) {
	t.Parallel()
	cfg := validCandlestickConfig()
	cfg.Caption = "Two trading sessions."
	cfg.Width, cfg.Height = "900px", "500px"
	cfg.DataZoom = []CandlestickDataZoom{
		{Type: CandlestickDataZoomInside, StartPercent: 50, EndPercent: 100},
		{StartPercent: 50, EndPercent: 100},
		{Axis: CandlestickDataZoomYAxis, StartPercent: 50, EndPercent: 100},
	}
	cfg.Series[0].Options = CandlestickSeriesOptions{
		Rise:        CandlestickDirectionStyle{Color: "#ec0000", Class: "price-rise", BorderColor: "#8a0000"},
		Fall:        CandlestickDirectionStyle{Color: "#00da3c", Class: "price-fall", BorderColor: "#008f28"},
		BorderWidth: 2, BarWidth: "60%", BarMinWidth: "4px", BarMaxWidth: "20",
		Marks: CandlestickMarkOptions{Highest: true, Lowest: true, ShowLabel: Bool(true)},
	}
	cfg.Options = ChartOptions{
		Title:    &TitleOptions{Text: "Candlestick example"},
		Legend:   &LegendOptions{Show: Bool(true)},
		Tooltip:  &TooltipOptions{Show: Bool(true), Trigger: "axis"},
		XAxis:    &AxisOptions{Type: "category", SplitNumber: 20},
		YAxis:    &AxisOptions{Type: "value", Scale: Bool(true)},
		Controls: chartcontrol.Options{Fullscreen: true},
		Export:   &chartcontrol.ExportOptions{Filename: "candlestick-example"},
	}
	cfg.Style = charttheme.Style{Palette: charttheme.PaletteAraiHu, Class: "caller-class"}
	cfg.RootAttrs = templ.Attributes{"id": "prices", "data-purpose": "ohlc"}

	instance := Candlestick(cfg)
	if instance.Kind() != chartcomponents.KindInteractiveCandlestick {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	markup := renderCandlestick(t, instance)
	for _, want := range []string{
		`aria-label="Market prices"`, `id="prices"`, `data-purpose="ohlc"`,
		`goshtoso-charts-palette-araihu goshtoso-charts-interactive-candlestick caller-class`,
		`style="width:900px;height:500px;"`, `"type":"candlestick"`,
		`"2018/1/24","2018/1/25"`, `"value":[2320.26,2320.26,2287.3,2362.94]`,
		`"name":"Prices"`, `"text":"Candlestick example"`, `"trigger":"axis"`,
		`"splitNumber":20`, `"scale":true`,
		`"type":"inside","start":50,"end":100,"xAxisIndex":[0]`,
		`"start":50,"end":100,"xAxisIndex":[0]`,
		`"start":50,"end":100,"orient":"vertical","yAxisIndex":[0]`,
		`"barWidth":"60%"`, `"barMinWidth":"4px"`, `"barMaxWidth":"20"`,
		`"name":"highest value","type":"max","valueDim":"highest"`,
		`"name":"lowest value","type":"min","valueDim":"lowest"`,
		`"color":"#ec0000","color0":"#00da3c"`, `"borderColor":"#8a0000","borderColor0":"#008f28","borderWidth":2`,
		`data-goshtoso-charts-candlestick-styles=`, `price-rise`, `price-fall`,
		`Two trading sessions.`, `>Exact OHLC values</summary>`,
		`scope="col">Open</th>`, `scope="col">Close</th>`, `scope="col">Low</th>`, `scope="col">High</th>`,
		`>2018/1/24</th>`, `>2320.26</td>`, `>Rise</td>`, `>price-rise</td>`,
		`>2018/1/25</th>`, `>2291.3</td>`, `>Fall</td>`, `>price-fall</td>`,
		`data-goshtoso-chart-expand`,
		`-fullscreen-action`, `exportFromMenu($el, &#34;png&#34;)`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestCandlestickDefaultsToSemanticThemeAndSharedWrapper(t *testing.T) {
	t.Parallel()
	markup := renderCandlestick(t, Candlestick(validCandlestickConfig()))
	for _, want := range []string{
		`width:100%;height:500px`, `goshtoso-charts-interactive-candlestick`,
		`aspect-ratio: 9 / 5`, `--color-chart-increasing`, `--color-chart-decreasing`,
		`series.type === "candlestick"`, `data-goshtoso-chart-expand`,
		`exportFromMenu($el, &#34;png&#34;)`, `>rise</td>`, `>fall</td>`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("default markup missing %q", want)
		}
	}
	for _, unwanted := range []string{
		`-fullscreen-action"`,
		`echarts.dispose`, `-chart-expand-export"`,
	} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("default markup contains %q", unwanted)
		}
	}
	if strings.Count(markup, `data-goshtoso-charts-theme-runtime`) != 1 {
		t.Fatal("Candlestick did not use exactly one shared theme runtime")
	}
}

func TestCandlestickBoundsExactValueTable(t *testing.T) {
	t.Parallel()
	cfg := validCandlestickConfig()
	cfg.Categories = make([]string, maxCandlestickDetailRows+5)
	cfg.Series[0].Data = make([]Candle, len(cfg.Categories))
	for index := range cfg.Categories {
		cfg.Categories[index] = "Day"
		cfg.Series[0].Data[index] = Candle{Open: 1, Close: 2, Low: 0, High: 3}
	}
	rows := candlestickDetailRows(cfg, maxCandlestickDetailRows)
	if len(rows.Rows) != maxCandlestickDetailRows || rows.Omitted != 5 {
		t.Fatalf("bounded rows = %d omitted = %d", len(rows.Rows), rows.Omitted)
	}
}

func TestCandlestickRejectsInvalidDataAndOptions(t *testing.T) {
	t.Parallel()
	tests := map[string]func(*CandlestickConfig){
		"label":          func(cfg *CandlestickConfig) { cfg.Label = "" },
		"categories":     func(cfg *CandlestickConfig) { cfg.Categories = nil },
		"blank category": func(cfg *CandlestickConfig) { cfg.Categories[0] = "" },
		"series":         func(cfg *CandlestickConfig) { cfg.Series = nil },
		"series name":    func(cfg *CandlestickConfig) { cfg.Series[0].Name = "" },
		"duplicate name": func(cfg *CandlestickConfig) { cfg.Series = append(cfg.Series, cfg.Series[0]) },
		"series data":    func(cfg *CandlestickConfig) { cfg.Series[0].Data = nil },
		"alignment":      func(cfg *CandlestickConfig) { cfg.Series[0].Data = cfg.Series[0].Data[:1] },
		"finite":         func(cfg *CandlestickConfig) { cfg.Series[0].Data[0].Open = math.NaN() },
		"low open":       func(cfg *CandlestickConfig) { cfg.Series[0].Data[0].Low = 2400 },
		"high close":     func(cfg *CandlestickConfig) { cfg.Series[0].Data[0].High = 2200 },
		"border width":   func(cfg *CandlestickConfig) { cfg.Series[0].Options.BorderWidth = -1 },
		"bar width":      func(cfg *CandlestickConfig) { cfg.Series[0].Options.BarWidth = "wide" },
		"tooltip":        func(cfg *CandlestickConfig) { cfg.Options.Tooltip = &TooltipOptions{Trigger: "pixel"} },
		"legend":         func(cfg *CandlestickConfig) { cfg.Options.Legend = &LegendOptions{Orient: "diagonal"} },
		"x axis":         func(cfg *CandlestickConfig) { cfg.Options.XAxis = &AxisOptions{Type: "time"} },
		"y axis":         func(cfg *CandlestickConfig) { cfg.Options.YAxis = &AxisOptions{Type: "category"} },
		"axis bounds":    func(cfg *CandlestickConfig) { cfg.Options.YAxis = &AxisOptions{Min: Float(2), Max: Float(1)} },
		"zoom type":      func(cfg *CandlestickConfig) { cfg.DataZoom = []CandlestickDataZoom{{Type: "wheel", EndPercent: 100}} },
		"zoom axis":      func(cfg *CandlestickConfig) { cfg.DataZoom = []CandlestickDataZoom{{Axis: "z", EndPercent: 100}} },
		"zoom bounds": func(cfg *CandlestickConfig) {
			cfg.DataZoom = []CandlestickDataZoom{{StartPercent: -1, EndPercent: 100}}
		},
		"zoom order":          func(cfg *CandlestickConfig) { cfg.DataZoom = []CandlestickDataZoom{{StartPercent: 80, EndPercent: 20}} },
		"axis split number":   func(cfg *CandlestickConfig) { cfg.Options.XAxis = &AxisOptions{SplitNumber: -1} },
		"axis label interval": func(cfg *CandlestickConfig) { cfg.Options.XAxis = &AxisOptions{LabelInterval: Int(-1)} },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := validCandlestickConfig()
			mutate(&cfg)
			var output bytes.Buffer
			if err := Candlestick(cfg).Render(context.Background(), &output); err == nil {
				t.Fatal("Render() error = nil")
			}
		})
	}
}

func validCandlestickConfig() CandlestickConfig {
	return CandlestickConfig{
		Label:      "Market prices",
		Categories: []string{"2018/1/24", "2018/1/25"},
		Series: []CandlestickSeries{{
			Name: "Prices",
			Data: []Candle{
				{Open: 2320.26, Close: 2320.26, Low: 2287.3, High: 2362.94},
				{Open: 2300, Close: 2291.3, Low: 2288.26, High: 2308.38},
			},
		}},
	}
}

func renderCandlestick(t *testing.T, instance Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return output.String()
}
