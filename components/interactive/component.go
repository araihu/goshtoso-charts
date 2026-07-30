// Package interactive provides legacy-compatible interactive chart APIs.
//
// Chart-specific packages are becoming canonical owners in phases. This
// parent package remains a source-compatible facade while shared rendering
// behavior lives in components/internal/interactive.
package interactive

import (
	"github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
)

// Instance is the renderer-neutral chart instance shared by all chart families.
type Instance = chart.Instance

type renderConfig = internalinteractive.RenderConfig
type scriptReplacement = internalinteractive.ScriptReplacement

func newInstance(kind components.Kind, cfg renderConfig) Instance {
	return internalinteractive.New(kind, cfg)
}

func newInvalidInstance(kind components.Kind, err error) Instance {
	return internalinteractive.Invalid(kind, err)
}

func responsiveWidth(value string) bool {
	return internalinteractive.ResponsiveWidth(value)
}
