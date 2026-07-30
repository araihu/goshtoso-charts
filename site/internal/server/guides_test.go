package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChartGuidesDocumentActualRendererNeutralCapabilities(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path     string
		marker   string
		required []string
	}{
		{
			path:   "/docs/chart-modes",
			marker: "data-chart-modes-guide",
			required: []string{
				"Static/vector and interactive charts", "Capability comparison", "Inline SVG", "browser runtime initializes a canvas",
				"SVG and browser-rasterized PNG", "PNG snapshots", "Categorical Bar and Line", "strict nonce-only script policy",
				"data-chart-mode-comparison", "data-guide-api-link=\"line\"", "data-guide-api-link=\"interactive\"", "data-guide-api-link=\"dependencies\"",
				"data-chart-mode-limits", "Shared wrapper boundary", "Omitted removes only the wrapper", "Features remain chart-specific",
			},
		},
		{
			path:   "/docs/chart-controls",
			marker: "data-chart-controls-guide",
			required: []string{
				"Chart controls", "Enabled by default and kept as the primary action", "Disabled by default", "Not part of the public control API",
				"same chart DOM", "one stacked Expand dropdown", "flattened into one accessible overflow menu", "SVG and PNG", "Interactive canvas",
				"data-guide-api-link=\"chartcontrol\"", "Accessibility and status", "Wrapper lifecycle", "WrapperModeEnabled",
				"WrapperModeDisabled", "WrapperModeHidden", "WrapperModeOmitted", "data-wrapper-mode-comparison",
				"goshtoso-charts:set-wrapper-mode", "goshtoso-charts:wrapper-mode-change", "previousMode", "focusReturn",
				"HTMX swaps", "No-JavaScript behavior", "Omitted skips wrapper-only export validation",
				"window.__goshtosoChartsControls.setWrapperMode", "returns false", "htmx:load", "htmx:afterSwap", "MutationObserver",
				"Unknown modes", "Unsupported export", "Caller responsibilities",
				"actionless non-omitted wrapper", "versioned same-origin external runtime", "Omitted mode alone suppresses that runtime",
				"one closed state", "data-wrapper-dom-contract", "data-goshtoso-chart-wrapper-mode", "observable state, not a mutation API",
				"Disabled affects shared wrapper actions only", "enabled-to-omitted and omitted-to-non-omitted", "preserves any action that was already unavailable",
				"Change charts with form values", "data-chart-control-examples", "data-chart-control-example=\"static\"", "data-chart-control-example=\"interactive\"",
				"Static form and chart · templ", "Interactive form and chart · templ", "Apply static controls", "Apply interactive controls",
				"Native select controls keep lifecycle and orientation editable", "form, range, checkbox, button, and source presentation use Goshtoso components",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d", recorder.Code)
			}
			body := recorder.Body.String()
			for _, want := range append([]string{test.marker, "data-chart-guide-nav", `aria-current="page"`, `data-go-api-version="v0.0.1"`, "pkg.go.dev is the canonical reference"}, test.required...) {
				if !strings.Contains(body, want) {
					t.Errorf("guide missing %q", want)
				}
			}
			for _, forbidden := range []string{"go-echarts", "Apache ECharts", "go-analyze/charts", "KindInteractive", "renderer type"} {
				if strings.Contains(body, forbidden) {
					t.Errorf("guide exposes backing implementation %q", forbidden)
				}
			}
		})
	}
}

func TestGettingStartedMakesModeAndWrapperChoiceExplicit(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	body := recorder.Body.String()
	for _, want := range []string{
		"Choose delivery and wrapper behavior",
		`href="/docs/chart-modes"`,
		`href="/docs/chart-controls"`,
		"Static/vector and interactive charts solve different jobs",
		"enabled, disabled, hidden, or omitted",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("Getting Started missing %q", want)
		}
	}
}

func TestChartGuideHTMXResponseKeepsTitleSidebarAndTOCHeadings(t *testing.T) {
	t.Parallel()
	request := httptest.NewRequest(http.MethodGet, "/docs/chart-controls", nil)
	request.Header.Set("HX-Request", "true")
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, request)
	body := recorder.Body.String()
	if strings.Contains(body, "<html") {
		t.Fatal("HTMX response contains a complete document")
	}
	for _, want := range []string{
		`<title>Chart controls · Goshtoso Charts</title>`, `id="main-content"`, `id="componentdocshell-sidebar-content"`,
		`hx-swap-oob`, `data-sidebar-icon="chart-controls"`, `aria-current="page"`, `id="responsive-actions"`, `data-toc-heading`,
		`id="live-form-examples"`, `id="wrapper-lifecycle"`, `id="client-transitions"`, `id="htmx-swaps"`, `id="state-guarantees"`, `id="no-javascript"`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("HTMX guide response missing %q", want)
		}
	}
}
