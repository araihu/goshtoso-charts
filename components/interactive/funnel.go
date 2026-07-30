package interactive

import interactivefunnel "github.com/araihu/goshtoso-charts/components/interactive/funnel"

// FunnelOrder is the compatibility name for funnel.Order.
type FunnelOrder = interactivefunnel.Order

const (
	// FunnelOrderDescending renders the largest stage first. It is the default.
	FunnelOrderDescending FunnelOrder = interactivefunnel.OrderDescending
	// FunnelOrderAscending renders the smallest stage first.
	FunnelOrderAscending FunnelOrder = interactivefunnel.OrderAscending
	// FunnelOrderData preserves caller data order.
	FunnelOrderData FunnelOrder = interactivefunnel.OrderData
)

// FunnelConfig is the compatibility name for funnel.Config.
type FunnelConfig = interactivefunnel.Config

// FunnelSeries is the compatibility name for funnel.Series.
type FunnelSeries = interactivefunnel.Series

// FunnelData is the compatibility name for funnel.Data.
type FunnelData = interactivefunnel.Data

// Funnel forwards to the canonical funnel package.
func Funnel(cfg FunnelConfig) Instance { return interactivefunnel.Funnel(cfg) }
