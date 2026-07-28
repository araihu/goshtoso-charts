// Package interactive renders typed interactive charts inside Goshtoso figures.
package interactive

import (
	"context"
	"fmt"
	"io"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/render"
)

type renderConfig struct {
	Label   string
	Caption string
	Chart   render.Renderer
	Style   charttheme.Style
}

// Instance is a renderable interactive chart figure.
type Instance struct {
	cfg  renderConfig
	kind chartcomponents.Kind
	err  error
}

func newInstance(kind chartcomponents.Kind, cfg renderConfig) Instance {
	return Instance{cfg: cfg, kind: kind}
}

func newInvalidInstance(kind chartcomponents.Kind, err error) Instance {
	return Instance{kind: kind, err: err}
}

// Kind identifies the interactive chart boundary.
func (instance Instance) Kind() chartcomponents.Kind { return instance.kind }

// Render writes a figure containing the chart element and initialization script.
// The site must serve its pinned local chart runtime before this output.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	if instance.err != nil {
		return instance.err
	}
	if instance.cfg.Label == "" {
		return fmt.Errorf("interactive chart label is required")
	}
	if instance.cfg.Chart == nil {
		return fmt.Errorf("interactive chart is required")
	}
	snippet := instance.cfg.Chart.RenderSnippet()
	return interactiveTemplate(instance.cfg, templ.Raw(snippet.Element), templ.Raw(snippet.Script)).Render(ctx, writer)
}

var _ chartcomponents.Component = Instance{}
