// Package interactive renders typed interactive charts inside Goshtoso figures.
package interactive

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/render"
)

type renderConfig struct {
	Label     string
	Caption   string
	Chart     render.Renderer
	Style     charttheme.Style
	Live      *liveConfig
	Animation *bool
	RootAttrs templ.Attributes
	Details   templ.Component
	Controls  chartcontrol.Options
	Export    *chartcontrol.ExportOptions
	// AxisLabelIntervals restore renderer-neutral integers after the private
	// renderer serializes its string-only interval field.
	AxisLabelIntervals []int
	// ThemeSeriesItems lists series indexes whose renderer-specific item style
	// is library-managed. Explicit caller item styles are never listed.
	ThemeSeriesItems []int
	// GaugeScale is private JSON consumed by the shared theme runtime.
	GaugeScale string
	// ExplicitVisualMapColors keeps caller-selected scale colors authoritative.
	ExplicitVisualMapColors bool
	// CandlestickStyles is private JSON consumed by the shared theme runtime.
	CandlestickStyles string
	// Liquid is private paint metadata consumed by the shared theme runtime.
	Liquid string
	// GeoGeometryPaint and GeoSeriesPaints are private paint metadata consumed
	// by the shared theme runtime.
	GeoGeometryPaint string
	GeoSeriesPaints  string
	// Scatter3DPaints and Scatter3DColdToWarm are private paint metadata.
	Scatter3DPaints     string
	Scatter3DColdToWarm bool
	// Bar3DPaints and Bar3DColdToWarm are private paint metadata.
	Bar3DPaints     string
	Bar3DColdToWarm bool
	// ScriptReplacements restore values that the private renderer cannot
	// distinguish from omitted zero values.
	ScriptReplacements []scriptReplacement
}

type scriptReplacement struct {
	Old string
	New string
}

func animationPreference(value *bool) string {
	if value == nil {
		return "default"
	}
	return strconv.FormatBool(*value)
}

func themeSeriesItems(indexes []int) string {
	values := make([]string, len(indexes))
	for index, value := range indexes {
		values[index] = strconv.Itoa(value)
	}
	return strings.Join(values, ",")
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
	for _, interval := range instance.cfg.AxisLabelIntervals {
		sentinel := strconv.Quote(axisLabelIntervalSentinel(interval))
		snippet.Script = strings.ReplaceAll(snippet.Script, sentinel, strconv.Itoa(interval))
	}
	for _, replacement := range instance.cfg.ScriptReplacements {
		snippet.Script = strings.ReplaceAll(snippet.Script, replacement.Old, replacement.New)
	}
	chart := interactiveTemplate(instance.cfg, templ.Raw(snippet.Element), templ.Raw(snippet.Script))
	return chartcontrol.Wrapper(chartcontrol.WrapperConfig{
		Label: instance.cfg.Label, Controls: instance.cfg.Controls, Export: instance.cfg.Export,
		Capability: chartcontrol.ExportCapabilityInteractiveRaster,
	}, chart).Render(ctx, writer)
}

var _ chartcomponents.Component = Instance{}
