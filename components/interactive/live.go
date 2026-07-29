package interactive

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
)

// LiveData configures an SSE source for chart types that support full-snapshot
// updates. Event defaults to the standard message event.
type LiveData struct {
	URL   string
	Event string
}

// CartesianSnapshot is the renderer-neutral SSE payload accepted by live Bar
// and Line components. Every series must contain one value per category.
type CartesianSnapshot struct {
	Categories []string                  `json:"categories"`
	Series     []CartesianSnapshotSeries `json:"series"`
}

// CartesianSnapshotSeries is one named series in a categorical live snapshot.
type CartesianSnapshotSeries struct {
	Name   string    `json:"name"`
	Values []float64 `json:"values"`
}

type liveConfig struct {
	URL   string
	Event string
	Shape string
}

func validateLiveData(live *LiveData) error {
	if live != nil && strings.TrimSpace(live.URL) == "" {
		return fmt.Errorf("live data URL is required")
	}
	return nil
}

func cartesianLiveConfig(live *LiveData) *liveConfig {
	if live == nil {
		return nil
	}
	event := strings.TrimSpace(live.Event)
	if event == "" {
		event = "message"
	}
	return &liveConfig{URL: strings.TrimSpace(live.URL), Event: event, Shape: "cartesian"}
}

func liveAttributes(live *liveConfig) templ.Attributes {
	if live == nil {
		return nil
	}
	return templ.Attributes{
		"data-goshtoso-charts-live-url":   live.URL,
		"data-goshtoso-charts-live-event": live.Event,
		"data-goshtoso-charts-live-shape": live.Shape,
	}
}
