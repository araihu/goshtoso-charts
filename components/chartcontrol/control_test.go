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

func TestOptionsPublicAPIContainsOnlyExpandAndFullscreen(t *testing.T) {
	t.Parallel()
	typ := reflect.TypeOf(chartcontrol.Options{})
	fields := make([]string, typ.NumField())
	for index := 0; index < typ.NumField(); index++ {
		fields[index] = typ.Field(index).Name
	}
	if got, want := strings.Join(fields, ","), "Fullscreen,Expand"; got != want {
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
		`data-goshtoso-action-group`, `data-action-group-primary`,
		`role="dialog"`, `aria-modal="true"`, `x-trap.inert.noscroll`,
		`id="latency-chart-expand-export"`, `data-action-group-overflow`,
		`More Latency chart actions`, `Export`, `>SVG</button>`, `>PNG</button>`,
		`data-action-group-overflow-counts="3"`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("default wrapper missing %q", want)
		}
	}
	for _, unwanted := range []string{`Collapse`, `data-goshtoso-chart-secondary-actions`, `data-goshtoso-chart-overflow`} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("default wrapper unexpectedly contains %q", unwanted)
		}
	}
}

func TestExpandAndFullscreenShareOnePrimaryDropdown(t *testing.T) {
	t.Parallel()
	markup := render(t, chartcontrol.WrapperConfig{
		Label:      "Latency",
		Controls:   chartcontrol.Options{Fullscreen: true},
		Capability: chartcontrol.ExportCapabilityInteractiveRaster,
	})
	for _, want := range []string{
		`goshtoso-charts-stacked-primary`,
		`id="latency-chart-expand-stacked"`,
		`id="latency-chart-expand-action"`,
		`id="latency-chart-expand-fullscreen-action"`,
		`class="goshtoso-charts-expand-control goshtoso-charts-hidden-expand-modal"`,
		`window.__goshtosoChartsControls.expandFromMenu($el)`,
		`window.__goshtosoChartsControls.toggleFullscreen($el)`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("combined primary action missing %q", want)
		}
	}
	if strings.Contains(markup, `Collapse`) {
		t.Fatal("combined Expand menu rendered an adjacent fullscreen peer")
	}
}

func TestFullscreenWorksWithoutExpandAndCollapseIsAbsentFromPublicMarkup(t *testing.T) {
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
		`data-action-group-primary`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("fullscreen markup missing %q", want)
		}
	}
	if strings.Contains(markup, "Collapse") {
		t.Fatal("removed Collapse control leaked into markup")
	}
}

func TestActionGroupFlattensStackedActionsAndExportFormats(t *testing.T) {
	t.Parallel()
	markup := render(t, chartcontrol.WrapperConfig{
		Label:      "Latency",
		Controls:   chartcontrol.Options{Fullscreen: true},
		Capability: chartcontrol.ExportCapabilityStaticSVG,
	})
	for _, want := range []string{
		`data-goshtoso-action-group`,
		`data-action-group-secondary`,
		`data-action-group-overflow`,
		`data-action-group-overflow-counts="3,3"`,
		`id="latency-chart-expand-action-overflow"`,
		`id="latency-chart-expand-fullscreen-action-overflow"`,
		`id="latency-chart-expand-export-svg-action-overflow"`,
		`id="latency-chart-expand-export-png-action-overflow"`,
		`aria-label="More Latency chart actions"`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("responsive overflow missing %q", want)
		}
	}
	if strings.Contains(markup, "Collapse") {
		t.Fatal("ActionGroup contains removed Collapse action")
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
		`id="latency-chart-expand-export"`, `Export`,
		`>SVG</button>`, `>PNG</button>`,
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
	if strings.Contains(interactiveMarkup, `id="latency-chart-expand-export"`) {
		t.Fatal("single interactive format rendered a dropdown")
	}
	if !strings.Contains(interactiveMarkup, `window.__goshtosoChartsControls.exportFromMenu($el, &#34;png&#34;)`) {
		t.Fatal("interactive raster capability omitted PNG export")
	}
}

func TestExpandAndExportCanBeDisabledIndependently(t *testing.T) {
	t.Parallel()
	markup := render(t, chartcontrol.WrapperConfig{
		Label: "Latency", Controls: chartcontrol.Options{Expand: chartcontrol.Bool(false)},
		Export: &chartcontrol.ExportOptions{Disabled: true}, Capability: chartcontrol.ExportCapabilityStaticSVG,
	})
	for _, unwanted := range []string{"data-goshtoso-chart-expand", "data-goshtoso-action-group", "controls.js"} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("opted-out wrapper contains %q", unwanted)
		}
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
