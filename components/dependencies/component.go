// Package dependencies renders the browser runtime required by interactive
// Goshtoso Charts components.
package dependencies

import (
	"context"
	"io"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
)

// Instance is a renderable Goshtoso Charts dependency set.
type Instance struct {
	config config
}

// Dependencies returns the runtime dependency set. It uses the embedded local
// asset by default; pass WithCDN to opt into third-party delivery.
func Dependencies(options ...Option) Instance {
	return Instance{config: newConfig(options)}
}

// Kind identifies the dependency component.
func (Instance) Kind() chartcomponents.Kind { return chartcomponents.KindDependencies }

// Render writes the runtime script tag.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	cfg := instance.config
	if !cfg.initialized {
		cfg = newConfig(nil)
	}
	cfg.nonce = templ.GetNonce(ctx)
	return dependenciesTemplate(cfg).Render(ctx, writer)
}

var _ chartcomponents.Component = Instance{}
