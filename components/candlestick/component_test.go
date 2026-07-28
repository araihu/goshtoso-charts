package candlestick

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func validConfig() Config {
	return Config{
		Label: "Seven-day stock price", Title: "Candlestick Chart", SeriesName: "Stock Price",
		Data: []Datum{
			{Label: "Day 1", Open: 100, High: 110, Low: 95, Close: 105},
			{Label: "Day 2", Open: 105, High: 115, Low: 100, Close: 112},
			{Label: "Day 3", Open: 112, High: 118, Low: 108, Close: 115},
			{Label: "Day 4", Open: 115, High: 120, Low: 104, Close: 108},
		},
	}
}

func TestCandlestickRendersAccessibleSSRAndExactValues(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.Caption = "Seven daily OHLC observations."
	cfg.XAxis = Axis{Title: "Day"}
	cfg.YAxis = Axis{Title: "Price"}
	cfg.Style = charttheme.Style{Palette: charttheme.PaletteAraiHu, Class: "mx-auto"}
	cfg.RootAttrs = templ.Attributes{"id": "stock-price", "data-chart-purpose": "ohlc"}

	instance := Candlestick(cfg)
	if instance.Kind() != chartcomponents.KindCandlestickChart {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`<figure class="goshtoso-charts-candlestick goshtoso-charts-palette goshtoso-charts-palette-araihu mx-auto" role="img" aria-label="Seven-day stock price"`,
		`id="stock-price"`, `data-chart-purpose="ohlc"`, "<svg", "Candlestick Chart", "Stock Price", "Day", "Price",
		"Seven daily OHLC observations.", "Exact OHLC values", "Increase means close", "Day 4", "Decrease", "115", "120", "104", "108",
		"var(--color-chart-increasing)", "var(--color-chart-decreasing)", "min-width: 37.5rem",
		`data-goshtoso-chart-expand`, `data-goshtoso-chart-export-menu`,
		`>SVG</button>`, `>PNG</button>`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	for _, unwanted := range []string{"rgb(145,204,117)", "rgb(238,102,102)"} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("rendered markup contains %q", unwanted)
		}
	}
}

func TestCandlestickMapsTypedOHLCAndPresentation(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	cfg.XAxis.Title = "Session"
	cfg.YAxis.Title = "Value"
	options := candlestickOptions(cfg)
	if options.Title.Text != "Candlestick Chart" || options.XAxis.Title != "Session" || options.YAxis[0].Title != "Value" {
		t.Fatalf("titles not mapped: %#v", options)
	}
	if got := options.XAxis.Labels; len(got) != 4 || got[3] != "Day 4" {
		t.Fatalf("labels = %v", got)
	}
	if got := options.SeriesList[0]; got.Name != "Stock Price" || got.Data[3].Open != 115 || got.Data[3].High != 120 || got.Data[3].Low != 104 || got.Data[3].Close != 108 {
		t.Fatalf("series = %#v", got)
	}
	if options.CandleWidth != 0.8 || options.ShowWicks != nil {
		t.Fatalf("upstream defaults changed: width=%v wicks=%v", options.CandleWidth, options.ShowWicks)
	}
}

func TestCandlestickDefaultsToUpstreamDimensions(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	if cfg.width() != 600 || cfg.height() != 400 {
		t.Fatalf("dimensions = %dx%d, want 600x400", cfg.width(), cfg.height())
	}
}

func TestCandlestickValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{name: "label", edit: func(c *Config) { c.Label = "" }, want: "label is required"},
		{name: "series", edit: func(c *Config) { c.SeriesName = "" }, want: "series name is required"},
		{name: "data", edit: func(c *Config) { c.Data = nil }, want: "at least one datum"},
		{name: "datum label", edit: func(c *Config) { c.Data[0].Label = "" }, want: "datum 1 needs a label"},
		{name: "duplicate label", edit: func(c *Config) { c.Data[1].Label = "Day 1" }, want: `datum label "Day 1" is duplicated`},
		{name: "open finite", edit: func(c *Config) { c.Data[0].Open = math.NaN() }, want: `datum "Day 1" open must be finite`},
		{name: "high finite", edit: func(c *Config) { c.Data[0].High = math.Inf(1) }, want: `datum "Day 1" high must be finite`},
		{name: "low above open", edit: func(c *Config) { c.Data[0].Low = 101 }, want: "low must be less than or equal"},
		{name: "low above close", edit: func(c *Config) { c.Data[0].Low = 106 }, want: "low must be less than or equal"},
		{name: "high below open", edit: func(c *Config) { c.Data[0].High = 99 }, want: "high must be greater than or equal"},
		{name: "high below close", edit: func(c *Config) { c.Data[0].High = 104 }, want: "high must be greater than or equal"},
		{name: "width", edit: func(c *Config) { c.Width = -1 }, want: "width cannot be negative"},
		{name: "height", edit: func(c *Config) { c.Height = -1 }, want: "height cannot be negative"},
		{name: "root attr", edit: func(c *Config) { c.RootAttrs = templ.Attributes{"ARIA-Label": "override"} }, want: `root attribute "ARIA-Label" is reserved`},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			cfg := validConfig()
			cfg.Data = append([]Datum(nil), cfg.Data...)
			test.edit(&cfg)
			err := cfg.validate()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validate() error = %v, want %q", err, test.want)
			}
		})
	}
}
