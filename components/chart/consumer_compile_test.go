package chart_test

import (
	"context"
	"io"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/interactive"
	interactivebar "github.com/araihu/goshtoso-charts/components/interactive/bar"
	interactivecandlestick "github.com/araihu/goshtoso-charts/components/interactive/candlestick"
	interactiveheatmap "github.com/araihu/goshtoso-charts/components/interactive/heatmap"
	interactiveline "github.com/araihu/goshtoso-charts/components/interactive/line"
	interactivepie "github.com/araihu/goshtoso-charts/components/interactive/pie"
	interactivescatter "github.com/araihu/goshtoso-charts/components/interactive/scatter"
)

// These snippets are compile contracts for old facade consumers, canonical
// child-package consumers, and renderer-neutral custom component consumers.
// Their import list intentionally contains no rendering engine package.

func oldFacadeConsumer() chart.Instance {
	return interactive.Bar(interactive.BarConfig{
		Label:  "Weekly deployments",
		XAxis:  []string{"Mon", "Tue"},
		Series: []interactive.BarSeries{{Name: "Production", Data: []interactive.BarData{{Value: 3}, {Value: 5}}}},
	})
}

func oldFacadeLineConsumer() chart.Instance {
	return interactive.Line(interactive.LineConfig{
		Label:  "Weekly latency",
		XAxis:  []string{"Mon", "Tue"},
		Series: []interactive.LineSeries{{Name: "p95 (ms)", Data: []interactive.LineData{{Value: 42}, {Value: 47}}}},
	})
}

func oldFacadeScatterConsumer() chart.Instance {
	return interactive.Scatter(interactive.ScatterConfig{
		Label:  "Weekly throughput",
		XAxis:  []string{"Mon", "Tue"},
		Series: []interactive.ScatterSeries{{Name: "Requests", Data: []interactive.ScatterData{{Value: 42}, {Value: 47}}}},
	})
}

func oldFacadeCandlestickConsumer() chart.Instance {
	return interactive.Candlestick(interactive.CandlestickConfig{
		Label:      "Daily prices",
		Categories: []string{"Mon"},
		Series:     []interactive.CandlestickSeries{{Name: "Price", Data: []interactive.Candle{{Open: 10, Close: 11, Low: 9, High: 12}}}},
	})
}

func oldFacadeHeatMapConsumer() chart.Instance {
	return interactive.HeatMap(interactive.HeatMapConfig{
		Label:      "Deployment activity",
		XAxis:      []string{"Mon"},
		YAxis:      []string{"Morning"},
		ValueRange: interactive.HeatMapValueRange{Min: 0, Max: 10},
		Series:     []interactive.HeatMapSeries{{Name: "Deployments", Data: []interactive.HeatMapData{{Value: 3}}}},
	})
}

func oldFacadePieConsumer() chart.Instance {
	return interactive.Pie(interactive.PieConfig{
		Label:  "Deployment outcomes",
		Series: []interactive.PieSeries{{Name: "Outcome", Data: []interactive.PieData{{Name: "Passed", Value: 8}, {Name: "Failed", Value: 2}}}},
	})
}

func canonicalChildConsumers() []chart.Instance {
	return []chart.Instance{
		interactivebar.Bar(interactivebar.Config{
			Label:  "Weekly deployments",
			XAxis:  []string{"Mon", "Tue"},
			Series: []interactivebar.Series{{Name: "Production", Data: []interactivebar.Data{{Value: 3}, {Value: 5}}}},
		}),
		interactiveline.Line(interactiveline.Config{
			Label:  "Weekly latency",
			XAxis:  []string{"Mon", "Tue"},
			Series: []interactiveline.Series{{Name: "p95 (ms)", Data: []interactiveline.Data{{Value: 42}, {Value: 47}}}},
		}),
		interactivescatter.Scatter(interactivescatter.Config{
			Label:  "Weekly throughput",
			XAxis:  []string{"Mon", "Tue"},
			Series: []interactivescatter.Series{{Name: "Requests", Data: []interactivescatter.Data{{Value: 42}, {Value: 47}}}},
		}),
		interactivecandlestick.Candlestick(interactivecandlestick.Config{
			Label:      "Daily prices",
			Categories: []string{"Mon"},
			Series:     []interactivecandlestick.Series{{Name: "Price", Data: []interactivecandlestick.Candle{{Open: 10, Close: 11, Low: 9, High: 12}}}},
		}),
		interactiveheatmap.HeatMap(interactiveheatmap.Config{
			Label:      "Deployment activity",
			XAxis:      []string{"Mon"},
			YAxis:      []string{"Morning"},
			ValueRange: interactiveheatmap.ValueRange{Min: 0, Max: 10},
			Series:     []interactiveheatmap.Series{{Name: "Deployments", Data: []interactiveheatmap.Data{{Value: 3}}}},
		}),
		interactivepie.Pie(interactivepie.Config{
			Label:  "Deployment outcomes",
			Series: []interactivepie.Series{{Name: "Outcome", Data: []interactivepie.Data{{Name: "Passed", Value: 8}, {Name: "Failed", Value: 2}}}},
		}),
	}
}

type customConsumerComponent struct{}

func (customConsumerComponent) Kind() chartcomponents.Kind { return "custom-chart" }

func (customConsumerComponent) Render(context.Context, io.Writer) error { return nil }

func customExtensionConsumer() chart.Instance {
	return chart.NewInstance(customConsumerComponent{})
}

var (
	_ = oldFacadeConsumer
	_ = oldFacadeLineConsumer
	_ = oldFacadeScatterConsumer
	_ = oldFacadeCandlestickConsumer
	_ = oldFacadeHeatMapConsumer
	_ = oldFacadePieConsumer
	_ = canonicalChildConsumers
	_ = customExtensionConsumer
)
