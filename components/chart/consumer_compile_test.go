package chart_test

import (
	"context"
	"io"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/interactive"
	interactivebar "github.com/araihu/goshtoso-charts/components/interactive/bar"
	interactiveline "github.com/araihu/goshtoso-charts/components/interactive/line"
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
	_ = canonicalChildConsumers
	_ = customExtensionConsumer
)
