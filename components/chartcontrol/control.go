// Package chartcontrol provides renderer-neutral controls shared by every chart.
package chartcontrol

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso/components/button"
	"github.com/araihu/goshtoso/components/dropdown"
	"github.com/araihu/goshtoso/components/splitbutton"
)

// WrapperMode controls server-rendered wrapper lifecycle and its initial client state.
//
// Enabled, disabled, and hidden wrappers retain their chart DOM and can transition
// client-side. Omitted renders only the chart and cannot transition until an
// application or HTMX response renders a wrapper around it.
type WrapperMode string

const (
	// WrapperModeEnabled renders the wrapper and enables every configured action.
	WrapperModeEnabled WrapperMode = ""
	// WrapperModeDisabled keeps the wrapper and chart visible while disabling its actions.
	WrapperModeDisabled WrapperMode = "disabled"
	// WrapperModeHidden keeps the initialized wrapper and chart DOM hidden and inert.
	WrapperModeHidden WrapperMode = "hidden"
	// WrapperModeOmitted renders the chart without wrapper DOM or wrapper runtime.
	WrapperModeOmitted WrapperMode = "omitted"
)

const wrapperModeEnabledLiteral WrapperMode = "enabled"

// Options configures chart controls and the shared wrapper lifecycle. Expand
// defaults on; fullscreen defaults off; Mode defaults to WrapperModeEnabled.
type Options struct {
	Fullscreen bool
	// Expand opens the same chart instance in a large Goshtoso modal. Nil enables it.
	Expand *bool
	// Mode selects the wrapper's initial server-rendered lifecycle state.
	Mode WrapperMode
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

// Wrapper preserves chart DOM and its client lifecycle while adding configured
// controls. WrapperModeOmitted is the only mode that skips wrapper markup and
// the shared lifecycle runtime.
func Wrapper(cfg WrapperConfig, chart templ.Component) templ.Component {
	return instance{cfg: cfg, chart: chart}
}

func (instance instance) Render(ctx context.Context, writer io.Writer) error {
	if err := validateWrapperMode(instance.cfg.Controls.Mode); err != nil {
		return err
	}
	if instance.chart == nil {
		return fmt.Errorf("chart control wrapper content is required")
	}
	if instance.cfg.Controls.Mode == WrapperModeOmitted {
		return instance.chart.Render(ctx, writer)
	}
	formats, err := resolvedFormats(instance.cfg.Export, instance.cfg.Capability)
	if err != nil {
		return err
	}
	return wrapperTemplate(instance.cfg, formats, instance.chart).Render(ctx, writer)
}

func validateWrapperMode(mode WrapperMode) error {
	switch mode {
	case WrapperModeEnabled, wrapperModeEnabledLiteral, WrapperModeDisabled, WrapperModeHidden, WrapperModeOmitted:
		return nil
	default:
		return fmt.Errorf("chart wrapper mode %q is unsupported", mode)
	}
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

func wrapperMode(options Options) string {
	if options.Mode == WrapperModeEnabled || options.Mode == wrapperModeEnabledLiteral {
		return "enabled"
	}
	return string(options.Mode)
}

func wrapperDisabled(options Options) bool {
	return options.Mode == WrapperModeDisabled
}

func wrapperHidden(options Options) bool {
	return options.Mode == WrapperModeHidden
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

func expandSplitButton(cfg WrapperConfig) splitbutton.Config {
	return splitbutton.Config{
		ID:        expandID(cfg) + "-stacked",
		Primary:   expandAction(expandID(cfg) + "-primary-action"),
		MenuLabel: "Expand " + cfg.Label,
		Sections: []dropdown.Section{{Items: []dropdown.Item{
			fullscreenItem(cfg),
		}}},
		MenuAlign: dropdown.AlignEnd,
		Tone:      button.ToneAlternate,
		Size:      button.SizeMedium,
		RootClass: "goshtoso-charts-control-split",
	}
}

func expandAction(id string) splitbutton.Action {
	return splitbutton.Action{
		ID:      id,
		Label:   "Expand",
		Icon:    expandIcon(),
		OnClick: `window.__goshtosoChartsControls.expandFromMenu($el)`,
	}
}

func fullscreenID(cfg WrapperConfig) string {
	return expandID(cfg) + "-fullscreen-action"
}

func fullscreenItem(cfg WrapperConfig) dropdown.Item {
	return dropdown.Item{
		ID:      fullscreenID(cfg),
		Label:   "Fullscreen",
		Icon:    fullscreenIcon(),
		OnClick: `window.__goshtosoChartsControls.toggleFullscreen($el)`,
	}
}

func exportSplitButton(cfg WrapperConfig, formats []ExportFormat) splitbutton.Config {
	return splitbutton.Config{
		ID:        expandID(cfg) + "-export",
		Primary:   copyAction(cfg),
		MenuLabel: "Export " + cfg.Label,
		Sections:  []dropdown.Section{{Items: exportItems(cfg, formats)}},
		MenuAlign: dropdown.AlignEnd,
		Tone:      button.ToneAlternate,
		Size:      button.SizeMedium,
		RootClass: "goshtoso-charts-control-split",
	}
}

func copyAction(cfg WrapperConfig) splitbutton.Action {
	return splitbutton.Action{
		ID:      expandID(cfg) + "-export-copy-action",
		Label:   "Copy",
		Icon:    copyIcon(),
		Tooltip: "Copy " + cfg.Label + " as PNG",
		OnClick: `window.__goshtosoChartsControls.copyFromMenu($el)`,
	}
}

func exportItems(cfg WrapperConfig, formats []ExportFormat) []dropdown.Item {
	items := make([]dropdown.Item, len(formats))
	for index, format := range formats {
		items[index] = dropdown.Item{
			ID:      expandID(cfg) + "-export-" + string(format) + "-action",
			Label:   "Download " + exportFormatLabel(format),
			Icon:    downloadIcon(),
			Tooltip: "Download " + cfg.Label + " as " + exportFormatLabel(format),
			OnClick: fmt.Sprintf(
				`window.__goshtosoChartsControls.exportFromMenu($el, %q); isOpen = false; openedWithKeyboard = false`,
				format,
			),
		}
	}
	return items
}
