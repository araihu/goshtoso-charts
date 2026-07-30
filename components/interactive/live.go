package interactive

import (
	"github.com/araihu/goshtoso-charts/components/chart"
	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
)

type LiveData = chart.LiveData
type CartesianSnapshot = chart.CartesianSnapshot
type CartesianSnapshotSeries = chart.CartesianSnapshotSeries
type liveConfig = internalinteractive.LiveConfig

func validateLiveData(live *LiveData) error {
	return internalinteractive.ValidateLiveData(live)
}

func cartesianLiveConfig(live *LiveData) *liveConfig {
	return internalinteractive.CartesianLiveConfig(live)
}
