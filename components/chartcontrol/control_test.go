package chartcontrol_test

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
)

func TestOptionsPublicAPIContainsSharedControlsAndWrapperMode(t *testing.T) {
	t.Parallel()
	typ := reflect.TypeOf(chartcontrol.Options{})
	fields := make([]string, typ.NumField())
	for index := 0; index < typ.NumField(); index++ {
		fields[index] = typ.Field(index).Name
	}
	if got, want := strings.Join(fields, ","), "Fullscreen,Expand,Mode"; got != want {
		t.Fatalf("chartcontrol.Options fields = %q, want %q", got, want)
	}
}

func render(t *testing.T, cfg chartcontrol.WrapperConfig) string {
	t.Helper()
	var output bytes.Buffer
	if err := chartcontrol.Wrapper(cfg, templ.Raw(`<figure><svg width="320" height="160"></svg></figure>`)).Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return output.String()
}

func TestDefaultWrapperEnablesExpandAndCapabilityDerivedExport(t *testing.T) {
	t.Parallel()
	markup := render(t, chartcontrol.WrapperConfig{
		Label:      "Latency",
		Capability: chartcontrol.ExportCapabilityStaticSVG,
	})
	for _, want := range []string{
		`class="goshtoso-charts-control-wrapper"`, `data-goshtoso-chart-expand`,
		`data-goshtoso-chart-wrapper-mode="enabled"`, `data-goshtoso-chart-actions-fieldset`,
		`data-goshtoso-chart-actions`, `data-split-button`,
		`goshtoso-charts-controls-neutral`, `data-goshtoso-chart-control-tone="neutral"`,
		`role="dialog"`, `aria-modal="true"`, `x-trap.inert.noscroll`,
		`id="latency-chart-expand-export"`,
		`Expand`, `Copy`, `<span class="block">Download SVG</span>`, `<span class="block">Download PNG</span>`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("default wrapper missing %q", want)
		}
	}
	for _, unwanted := range []string{`Collapse`, `>Close</button>`, `data-goshtoso-chart-secondary-actions`, `data-goshtoso-chart-overflow`} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("default wrapper unexpectedly contains %q", unwanted)
		}
	}
}

func TestControlIconsUseGoshtosoIconpack(t *testing.T) {
	t.Parallel()
	markup := render(t, chartcontrol.WrapperConfig{
		Label:      "Latency",
		Controls:   chartcontrol.Options{Fullscreen: true},
		Capability: chartcontrol.ExportCapabilityStaticSVG,
	})
	const expandSymbol = `<use href="/charts/assets/charticons/sprite.svg#heroicons-optimized-24-outline-arrows-pointing-out"></use>`
	const collapseSymbol = `<use href="/charts/assets/charticons/sprite.svg#heroicons-optimized-24-outline-arrows-pointing-in"></use>`
	for _, id := range []string{
		"latency-chart-expand-primary-action",
		"latency-chart-expand-fullscreen-action",
	} {
		start := strings.Index(markup, `id="`+id+`"`)
		if start < 0 {
			t.Fatalf("chart controls missing action %q", id)
		}
		end := strings.Index(markup[start:], `</button>`)
		if end < 0 {
			t.Fatalf("chart controls action %q is missing its closing button", id)
		}
		if !strings.Contains(markup[start:start+end], expandSymbol) {
			t.Errorf("chart control %q does not use arrows-pointing-out", id)
		}
	}
	collapseStart := strings.Index(markup, `id="latency-chart-expand-collapse-action"`)
	if collapseStart < 0 {
		t.Fatal("chart controls missing collapse action")
	}
	collapseEnd := strings.Index(markup[collapseStart:], `</button>`)
	if collapseEnd < 0 || !strings.Contains(markup[collapseStart:collapseStart+collapseEnd], collapseSymbol) {
		t.Error("collapse control does not use arrows-pointing-in")
	}
	downloadAction := `id="latency-chart-expand-export-svg-action"`
	if start := strings.Index(markup, downloadAction); start < 0 || !strings.Contains(markup[start:], `heroicons-optimized-24-outline-arrow-down-tray`) {
		t.Error("chart export action does not use arrow-down-tray")
	}
}

func TestDefaultControlsShareNeutralToneAndStableDimensions(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	if err := chartcontrol.Styles().Render(context.Background(), &output); err != nil {
		t.Fatalf("Styles().Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`--goshtoso-chart-control-height: 2.75rem`,
		`--goshtoso-chart-control-width: 7.5rem`,
		`.goshtoso-charts-controls-neutral > [data-split-button] > button`,
		`background-color: var(--color-surface-alt)`,
		`.dark .goshtoso-charts-controls-neutral`,
		`.goshtoso-charts-expand-panel > div:nth-child(3)`,
		`display: none`,
		`[role="menuitem"]`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("default control styling missing %q", want)
		}
	}
}

func TestWrapperModesHaveClosedStableValues(t *testing.T) {
	t.Parallel()
	tests := map[chartcontrol.WrapperMode]string{
		chartcontrol.WrapperModeEnabled:  "",
		chartcontrol.WrapperModeDisabled: "disabled",
		chartcontrol.WrapperModeHidden:   "hidden",
		chartcontrol.WrapperModeOmitted:  "omitted",
	}
	for mode, want := range tests {
		if got := string(mode); got != want {
			t.Errorf("WrapperMode value = %q, want %q", got, want)
		}
	}
}

func TestExplicitEnabledWrapperModeRoundTripsClientState(t *testing.T) {
	t.Parallel()
	markup := render(t, chartcontrol.WrapperConfig{
		Label:      "Latency",
		Controls:   chartcontrol.Options{Mode: chartcontrol.WrapperMode("enabled")},
		Capability: chartcontrol.ExportCapabilityInteractiveRaster,
	})
	if !strings.Contains(markup, `data-goshtoso-chart-wrapper-mode="enabled"`) {
		t.Fatalf("explicit enabled mode did not round-trip: %s", markup)
	}
}

func TestDisabledWrapperKeepsChartVisibleAndActionsNativelyInert(t *testing.T) {
	t.Parallel()
	markup := render(t, chartcontrol.WrapperConfig{
		Label:      "Latency",
		Controls:   chartcontrol.Options{Mode: chartcontrol.WrapperModeDisabled, Fullscreen: true},
		Capability: chartcontrol.ExportCapabilityStaticSVG,
	})
	for _, want := range []string{
		`data-goshtoso-chart-wrapper-mode="disabled"`,
		`data-goshtoso-chart-actions-fieldset disabled aria-disabled="true"`,
		`<figure><svg width="320" height="160"></svg></figure>`,
		`src="/charts/assets/js/controls/6/controls.js"`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("disabled wrapper missing %q", want)
		}
	}
	if strings.Contains(markup, `data-goshtoso-chart-export-pixel-ratio="1" hidden inert aria-hidden="true"`) {
		t.Fatal("disabled wrapper hid visible chart DOM")
	}
}

func TestHiddenWrapperRetainsDOMRuntimeAndInertAccessibilityState(t *testing.T) {
	t.Parallel()
	markup := render(t, chartcontrol.WrapperConfig{
		Label:      "Latency",
		Controls:   chartcontrol.Options{Mode: chartcontrol.WrapperModeHidden},
		Capability: chartcontrol.ExportCapabilityInteractiveRaster,
	})
	for _, want := range []string{
		`data-goshtoso-chart-wrapper-mode="hidden"`, ` hidden inert aria-hidden="true"`,
		`<figure><svg width="320" height="160"></svg></figure>`,
		`src="/charts/assets/js/controls/6/controls.js"`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("hidden wrapper missing %q", want)
		}
	}
}

func TestOmittedWrapperIsExactChartPassthroughBeforeExportResolution(t *testing.T) {
	t.Parallel()
	const chartMarkup = `<figure data-chart-only><svg width="320" height="160"></svg></figure>`
	var output bytes.Buffer
	err := chartcontrol.Wrapper(chartcontrol.WrapperConfig{
		Label:    "Latency",
		Controls: chartcontrol.Options{Mode: chartcontrol.WrapperModeOmitted},
		Export: &chartcontrol.ExportOptions{
			Formats:    []chartcontrol.ExportFormat{"not-a-format"},
			PixelRatio: -1,
		},
	}, templ.Raw(chartMarkup)).Render(context.Background(), &output)
	if err != nil {
		t.Fatalf("omitted Render() error = %v", err)
	}
	if got := output.String(); got != chartMarkup {
		t.Fatalf("omitted markup = %q, want exact chart passthrough %q", got, chartMarkup)
	}
}

func TestUnknownWrapperModeFailsBeforeRendering(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	err := chartcontrol.Wrapper(chartcontrol.WrapperConfig{
		Label: "Latency", Controls: chartcontrol.Options{Mode: chartcontrol.WrapperMode("collapsed")},
	}, templ.Raw(`<figure></figure>`)).Render(context.Background(), &output)
	if err == nil || !strings.Contains(err.Error(), `chart wrapper mode "collapsed" is unsupported`) {
		t.Fatalf("Render() error = %v", err)
	}
	if output.Len() != 0 {
		t.Fatalf("unknown mode wrote %q", output.String())
	}
}

func TestNonOmittedModesStillValidateExportDeterministically(t *testing.T) {
	t.Parallel()
	for _, mode := range []chartcontrol.WrapperMode{
		chartcontrol.WrapperModeEnabled,
		chartcontrol.WrapperModeDisabled,
		chartcontrol.WrapperModeHidden,
	} {
		var output bytes.Buffer
		err := chartcontrol.Wrapper(chartcontrol.WrapperConfig{
			Label: "Latency", Controls: chartcontrol.Options{Mode: mode},
			Capability: chartcontrol.ExportCapabilityInteractiveRaster,
			Export:     &chartcontrol.ExportOptions{Formats: []chartcontrol.ExportFormat{chartcontrol.ExportSVG}},
		}, templ.Raw(`<figure></figure>`)).Render(context.Background(), &output)
		if err == nil || !strings.Contains(err.Error(), "svg export is unsupported") {
			t.Errorf("mode %q Render() error = %v", mode, err)
		}
	}
}

func TestExpandSplitButtonOwnsFullscreenMenu(t *testing.T) {
	t.Parallel()
	markup := render(t, chartcontrol.WrapperConfig{
		Label:      "Latency",
		Controls:   chartcontrol.Options{Fullscreen: true},
		Capability: chartcontrol.ExportCapabilityInteractiveRaster,
	})
	for _, want := range []string{
		`data-split-button`,
		`id="latency-chart-expand-stacked"`,
		`id="latency-chart-expand-primary-action"`,
		`id="latency-chart-expand-fullscreen-action"`,
		`class="goshtoso-charts-expand-control goshtoso-charts-hidden-expand-modal"`,
		`window.__goshtosoChartsControls.expandFromMenu($el)`,
		`window.__goshtosoChartsControls.toggleFullscreen($el)`,
		`window.__goshtosoChartsControls.collapseFullscreen($el)`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("combined primary action missing %q", want)
		}
	}
	if !strings.Contains(markup, `id="latency-chart-expand-collapse-action"`) || !strings.Contains(markup, `hidden`) {
		t.Fatal("combined Expand control did not render a hidden Collapse action")
	}
}

func TestFullscreenWorksWithoutExpandAndRendersHiddenCollapseControl(t *testing.T) {
	t.Parallel()
	markup := render(t, chartcontrol.WrapperConfig{
		Label:      "Latency",
		Controls:   chartcontrol.Options{Fullscreen: true, Expand: chartcontrol.Bool(false)},
		Capability: chartcontrol.ExportCapabilityStaticSVG,
		Export:     &chartcontrol.ExportOptions{Disabled: true},
	})
	for _, want := range []string{
		`id="latency-chart-expand-fullscreen-action"`,
		`window.__goshtosoChartsControls.toggleFullscreen($el)`,
		`goshtoso-charts-control-button`,
		`id="latency-chart-expand-collapse-action"`,
		`data-goshtoso-chart-control="collapse"`,
		`window.__goshtosoChartsControls.collapseFullscreen($el)`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("fullscreen markup missing %q", want)
		}
	}
}

func TestSplitButtonsKeepExpandAndExportActionsConnected(t *testing.T) {
	t.Parallel()
	markup := render(t, chartcontrol.WrapperConfig{
		Label:      "Latency",
		Controls:   chartcontrol.Options{Fullscreen: true},
		Capability: chartcontrol.ExportCapabilityStaticSVG,
	})
	for _, want := range []string{
		`id="latency-chart-expand-stacked"`,
		`id="latency-chart-expand-primary-action"`,
		`id="latency-chart-expand-fullscreen-action"`,
		`id="latency-chart-expand-export"`,
		`id="latency-chart-expand-export-copy-action"`,
		`id="latency-chart-expand-export-svg-action"`,
		`id="latency-chart-expand-export-png-action"`,
		`Download SVG`, `Download PNG`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("responsive overflow missing %q", want)
		}
	}
	if !strings.Contains(markup, `id="latency-chart-expand-collapse-action"`) {
		t.Fatal("SplitButton wrapper is missing its hidden Collapse action")
	}
}

func TestExportFormatsAreCapabilityGated(t *testing.T) {
	t.Parallel()
	staticMarkup := render(t, chartcontrol.WrapperConfig{
		Label:      "Latency",
		Capability: chartcontrol.ExportCapabilityStaticSVG,
		Export:     &chartcontrol.ExportOptions{Filename: "Availability / Latency"},
	})
	for _, want := range []string{
		`id="latency-chart-expand-export"`, `Copy`,
		`<span class="block">Download SVG</span>`, `<span class="block">Download PNG</span>`,
		`data-goshtoso-chart-export-filename="availability-latency"`,
	} {
		if !strings.Contains(staticMarkup, want) {
			t.Errorf("static export markup missing %q", want)
		}
	}

	interactiveMarkup := render(t, chartcontrol.WrapperConfig{
		Label:      "Latency",
		Capability: chartcontrol.ExportCapabilityInteractiveRaster,
		Export:     &chartcontrol.ExportOptions{Filename: "Latency"},
	})
	for _, want := range []string{
		`id="latency-chart-expand-export"`, `Copy`, `<span class="block">Download PNG</span>`,
		`window.__goshtosoChartsControls.copyFromMenu($el)`,
	} {
		if !strings.Contains(interactiveMarkup, want) {
			t.Errorf("interactive raster capability missing %q", want)
		}
	}
	if strings.Contains(interactiveMarkup, "Download SVG") {
		t.Fatal("interactive raster capability rendered unsupported SVG download")
	}
}

func TestActionlessWrapperKeepsLifecycleRuntimeWithoutRenderingActions(t *testing.T) {
	t.Parallel()
	markup := render(t, chartcontrol.WrapperConfig{
		Label: "Latency", Controls: chartcontrol.Options{Expand: chartcontrol.Bool(false)},
		Export: &chartcontrol.ExportOptions{Disabled: true}, Capability: chartcontrol.ExportCapabilityStaticSVG,
	})
	for _, unwanted := range []string{"data-goshtoso-chart-expand", "data-goshtoso-chart-actions", "data-goshtoso-chart-actions-fieldset"} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("opted-out wrapper contains %q", unwanted)
		}
	}
	const runtime = `src="/charts/assets/js/controls/6/controls.js"`
	if got := strings.Count(markup, runtime); got != 1 {
		t.Fatalf("actionless wrapper lifecycle runtime count = %d, want exactly one", got)
	}
}

func TestUnsupportedExplicitExportFormatFails(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	err := chartcontrol.Wrapper(chartcontrol.WrapperConfig{
		Label:      "Latency",
		Capability: chartcontrol.ExportCapabilityInteractiveRaster,
		Export: &chartcontrol.ExportOptions{
			Formats: []chartcontrol.ExportFormat{chartcontrol.ExportSVG},
		},
	}, templ.Raw(`<figure></figure>`)).Render(context.Background(), &output)
	if err == nil || !strings.Contains(err.Error(), "svg export is unsupported") {
		t.Fatalf("Render() error = %v", err)
	}
}

func TestInteractiveTransparentBackgroundFailsCapabilityGate(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	err := chartcontrol.Wrapper(chartcontrol.WrapperConfig{
		Label:      "Latency",
		Capability: chartcontrol.ExportCapabilityInteractiveRaster,
		Export: &chartcontrol.ExportOptions{
			Background: chartcontrol.ExportBackgroundTransparent,
		},
	}, templ.Raw(`<figure></figure>`)).Render(context.Background(), &output)
	if err == nil || !strings.Contains(err.Error(), "transparent export background is unsupported") {
		t.Fatalf("Render() error = %v", err)
	}
}

func TestSafeFilenameIsDeterministic(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		" Availability / Latency 2026 ": "availability-latency-2026",
		"../../":                        "goshtoso-chart",
		"São Paulo — latency":           "s-o-paulo-latency",
	}
	for input, want := range tests {
		if got := chartcontrol.SafeFilename(input); got != want {
			t.Errorf("SafeFilename(%q) = %q, want %q", input, got, want)
		}
	}
}
