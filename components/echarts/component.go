// Package echarts renders trusted go-echarts examples inside Goshtoso figures.
package echarts

import (
	"context"
	"fmt"
	"io"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/go-echarts/go-echarts/v2/render"
)

// Config describes a trusted browser-rendered go-echarts chart.
//
// Chart must be constructed from application-owned Go values. go-echarts emits
// executable JavaScript, so never pass request- or user-controlled values here.
type Config struct {
	Label   string
	Caption string
	Chart   render.Renderer
}

// Instance is a renderable interactive go-echarts figure.
type Instance struct{ cfg Config }

// EChart returns a Goshtoso-wrapped go-echarts chart.
func EChart(cfg Config) Instance { return Instance{cfg: cfg} }

// Kind identifies the interactive chart boundary.
func (Instance) Kind() chartcomponents.Kind { return chartcomponents.KindInteractiveECharts }

// Render writes a figure containing go-echarts' element and initialization
// script. The site must serve its pinned local ECharts asset before this output.
func (instance Instance) Render(ctx context.Context, writer io.Writer) error {
	if instance.cfg.Label == "" {
		return fmt.Errorf("interactive chart label is required")
	}
	if instance.cfg.Chart == nil {
		return fmt.Errorf("interactive chart is required")
	}
	snippet := instance.cfg.Chart.RenderSnippet()
	return echartTemplate(instance.cfg, templ.Raw(snippet.Element), templ.Raw(snippet.Script)).Render(ctx, writer)
}

var _ chartcomponents.Component = Instance{}
