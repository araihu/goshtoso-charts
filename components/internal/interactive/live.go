package interactive

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chart"
)

// LiveConfig is private renderer metadata derived from public live options.
type LiveConfig struct {
	URL   string
	Event string
	Shape string
}

// ValidateLiveData validates shared live source configuration.
func ValidateLiveData(live *chart.LiveData) error {
	if live != nil && strings.TrimSpace(live.URL) == "" {
		return fmt.Errorf("live data URL is required")
	}
	return nil
}

// CartesianLiveConfig derives private Cartesian runtime metadata.
func CartesianLiveConfig(live *chart.LiveData) *LiveConfig {
	if live == nil {
		return nil
	}
	event := strings.TrimSpace(live.Event)
	if event == "" {
		event = "message"
	}
	return &LiveConfig{URL: strings.TrimSpace(live.URL), Event: event, Shape: "cartesian"}
}

func liveAttributes(live *LiveConfig) templ.Attributes {
	if live == nil {
		return nil
	}
	return templ.Attributes{
		"data-goshtoso-charts-live-url":   live.URL,
		"data-goshtoso-charts-live-event": live.Event,
		"data-goshtoso-charts-live-shape": live.Shape,
	}
}
