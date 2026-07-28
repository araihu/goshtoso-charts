// Package chartcontrol provides renderer-neutral controls shared by every chart.
package chartcontrol

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/dropdown"
)

// Options configures chart controls. Expand defaults on; fullscreen and collapse default off.
type Options struct {
	Fullscreen  bool
	Collapsible bool
	// Expand opens the same chart instance in a large Goshtoso modal. Nil enables it.
	Expand *bool
}

// ExportFormat identifies a verified browser download format.
type ExportFormat string

const (
	// ExportSVG downloads a scalable vector artifact.
	ExportSVG ExportFormat = "svg"
	// ExportPNG downloads a portable network graphics artifact.
	ExportPNG ExportFormat = "png"
)

// ExportBackground controls whether exported pixels include the chart surface.
type ExportBackground string

const (
	// ExportBackgroundOpaque resolves the current chart surface into the artifact.
	ExportBackgroundOpaque ExportBackground = ""
	// ExportBackgroundTransparent preserves transparent pixels.
	ExportBackgroundTransparent ExportBackground = "transparent"
)

// ExportOptions customizes or disables capability-derived exports, which default on.
//
// Filename is normalized to a deterministic filesystem-safe basename. Formats
// defaults to every format proven for the component. PixelRatio defaults to 1.
type ExportOptions struct {
	Filename   string
	Formats    []ExportFormat
	Background ExportBackground
	PixelRatio float64
	// Disabled opts out of the capability-derived export control, which defaults on.
	Disabled bool
}

// ExportCapability identifies formats proven for a component implementation.
// It describes output capability without exposing the backing chart engine.
type ExportCapability string

const (
	// ExportCapabilityStaticSVG supports direct SVG and browser-rasterized PNG.
	ExportCapabilityStaticSVG ExportCapability = "static-svg"
	// ExportCapabilityInteractiveRaster supports PNG snapshots of live charts.
	ExportCapabilityInteractiveRaster ExportCapability = "interactive-raster"
)

// WrapperConfig configures the shared renderer-neutral chart wrapper.
type WrapperConfig struct {
	Label      string
	Controls   Options
	Export     *ExportOptions
	Capability ExportCapability
}

type instance struct {
	cfg   WrapperConfig
	chart templ.Component
}

// Wrapper preserves chart DOM while adding enabled controls around it.
func Wrapper(cfg WrapperConfig, chart templ.Component) templ.Component {
	return instance{cfg: cfg, chart: chart}
}

func (instance instance) Render(ctx context.Context, writer io.Writer) error {
	if instance.chart == nil {
		return fmt.Errorf("chart control wrapper content is required")
	}
	formats, err := resolvedFormats(instance.cfg.Export, instance.cfg.Capability)
	if err != nil {
		return err
	}
	return wrapperTemplate(instance.cfg, formats, instance.chart).Render(ctx, writer)
}

func resolvedFormats(options *ExportOptions, capability ExportCapability) ([]ExportFormat, error) {
	if options != nil && options.Disabled {
		return nil, nil
	}
	var formats []ExportFormat
	if options != nil {
		formats = options.Formats
	}
	if len(formats) == 0 {
		switch capability {
		case ExportCapabilityStaticSVG:
			formats = []ExportFormat{ExportSVG, ExportPNG}
		case ExportCapabilityInteractiveRaster:
			formats = []ExportFormat{ExportPNG}
		default:
			return nil, nil
		}
	}
	seen := make(map[ExportFormat]bool, len(formats))
	resolved := make([]ExportFormat, 0, len(formats))
	for _, format := range formats {
		if seen[format] {
			continue
		}
		switch format {
		case ExportSVG:
			if capability != ExportCapabilityStaticSVG {
				return nil, fmt.Errorf("svg export is unsupported for this chart")
			}
		case ExportPNG:
			if capability != ExportCapabilityStaticSVG && capability != ExportCapabilityInteractiveRaster {
				return nil, fmt.Errorf("png export is unsupported for this chart")
			}
		default:
			return nil, fmt.Errorf("chart export format %q is unsupported", format)
		}
		seen[format] = true
		resolved = append(resolved, format)
	}
	if options == nil {
		return resolved, nil
	}
	if options.Background != ExportBackgroundOpaque && options.Background != ExportBackgroundTransparent {
		return nil, fmt.Errorf("chart export background %q is unsupported", options.Background)
	}
	if options.Background == ExportBackgroundTransparent && capability != ExportCapabilityStaticSVG {
		return nil, fmt.Errorf("transparent export background is unsupported for this chart")
	}
	if options.PixelRatio < 0 {
		return nil, fmt.Errorf("chart export pixel ratio must not be negative")
	}
	return resolved, nil
}

var unsafeFilename = regexp.MustCompile(`[^a-z0-9_-]+`)

// SafeFilename returns the deterministic basename used for chart downloads.
func SafeFilename(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = unsafeFilename.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-_")
	if len(value) > 80 {
		value = strings.TrimRight(value[:80], "-_")
	}
	if value == "" {
		return "goshtoso-chart"
	}
	return value
}

func exportFilename(cfg WrapperConfig) string {
	if cfg.Export == nil || strings.TrimSpace(cfg.Export.Filename) == "" {
		return SafeFilename(cfg.Label)
	}
	return SafeFilename(cfg.Export.Filename)
}

func exportBackground(cfg WrapperConfig) string {
	if cfg.Export != nil && cfg.Export.Background == ExportBackgroundTransparent {
		return string(ExportBackgroundTransparent)
	}
	return "opaque"
}

func exportPixelRatio(cfg WrapperConfig) string {
	if cfg.Export == nil || cfg.Export.PixelRatio == 0 {
		return "1"
	}
	return fmt.Sprintf("%g", cfg.Export.PixelRatio)
}

func hasRuntime(cfg WrapperConfig, formats []ExportFormat) bool {
	return cfg.Controls.Fullscreen || cfg.Controls.Collapsible || expandEnabled(cfg.Controls) || len(formats) > 0
}

// Bool returns a pointer for explicit opt-out settings whose zero value is enabled.
func Bool(value bool) *bool { return &value }

func expandEnabled(options Options) bool {
	return options.Expand == nil || *options.Expand
}

func expandID(cfg WrapperConfig) string {
	return SafeFilename(cfg.Label) + "-chart-expand"
}

func exportFormatLabel(format ExportFormat) string {
	return strings.ToUpper(string(format))
}

func exportItems(formats []ExportFormat) []dropdown.Item {
	items := make([]dropdown.Item, len(formats))
	for index, format := range formats {
		items[index] = dropdown.Item{
			Label: exportFormatLabel(format),
			OnClick: fmt.Sprintf(
				`window.__goshtosoChartsControls.exportFromMenu($el, %q); isOpen = false; openedWithKeyboard = false`,
				format,
			),
		}
	}
	return items
}
