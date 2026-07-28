package chartcontrol_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
)

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
		`role="dialog"`, `aria-modal="true"`, `x-trap.inert.noscroll`,
		`data-goshtoso-chart-export-menu`, `Export`, `>SVG</button>`, `>PNG</button>`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("default wrapper missing %q", want)
		}
	}
	for _, unwanted := range []string{`data-goshtoso-chart-control="collapse"`, `data-goshtoso-chart-control="fullscreen"`} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("default wrapper unexpectedly contains %q", unwanted)
		}
	}
}

func TestFullscreenAndCollapsibleAreIndependent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		options    chartcontrol.Options
		fullscreen bool
		collapse   bool
	}{
		{name: "fullscreen", options: chartcontrol.Options{Fullscreen: true, Expand: chartcontrol.Bool(false)}, fullscreen: true},
		{name: "collapsible", options: chartcontrol.Options{Collapsible: true, Expand: chartcontrol.Bool(false)}, collapse: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			markup := render(t, chartcontrol.WrapperConfig{
				Label:      "Latency",
				Controls:   test.options,
				Capability: chartcontrol.ExportCapabilityStaticSVG,
				Export:     &chartcontrol.ExportOptions{Disabled: true},
			})
			if got := strings.Contains(markup, `data-goshtoso-chart-control="fullscreen"`); got != test.fullscreen {
				t.Fatalf("fullscreen control presence = %t", got)
			}
			if got := strings.Contains(markup, `data-goshtoso-chart-control="collapse"`); got != test.collapse {
				t.Fatalf("collapse control presence = %t", got)
			}
			if test.fullscreen {
				for _, want := range []string{`aria-pressed="false"`, `aria-label="Enter fullscreen for Latency"`} {
					if !strings.Contains(markup, want) {
						t.Errorf("fullscreen markup missing %q", want)
					}
				}
			}
			if test.collapse {
				for _, want := range []string{`aria-expanded="true"`, `aria-label="Collapse Latency"`} {
					if !strings.Contains(markup, want) {
						t.Errorf("collapse markup missing %q", want)
					}
				}
			}
		})
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
		`data-goshtoso-chart-export-menu`, `Export`,
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
	if strings.Contains(interactiveMarkup, `data-goshtoso-chart-export-menu`) {
		t.Fatal("single interactive format rendered a dropdown")
	}
	if !strings.Contains(interactiveMarkup, `data-goshtoso-chart-export="png"`) {
		t.Fatal("interactive raster capability omitted PNG export")
	}
}

func TestExpandAndExportCanBeDisabledIndependently(t *testing.T) {
	t.Parallel()
	markup := render(t, chartcontrol.WrapperConfig{
		Label: "Latency", Controls: chartcontrol.Options{Expand: chartcontrol.Bool(false)},
		Export: &chartcontrol.ExportOptions{Disabled: true}, Capability: chartcontrol.ExportCapabilityStaticSVG,
	})
	for _, unwanted := range []string{"data-goshtoso-chart-expand", "data-goshtoso-chart-export-menu", `data-goshtoso-chart-export="`, "controls.js"} {
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
