package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChartControlGuideRendersSubmittedExamplesAndNativeFallback(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/docs/chart-controls?static_present=1&static_mode=disabled&static_stroke=7&interactive_present=1&interactive_orientation=horizontal&interactive_scale=150&palette_present=1&chart_palette=custom&palette_color_1=%23123456&palette_color_2=%23234567&palette_color_3=%23345678&palette_color_4=%23456789", nil)
	New().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		`data-chart-control-example="static"`, `data-chart-control-mode="disabled"`, `data-chart-control-stroke="7"`, `data-chart-control-area="off"`,
		`data-goshtoso-chart-wrapper-mode="disabled"`, `Applied: disabled wrapper, 7 px stroke, area off.`,
		`data-chart-control-example="interactive"`, `data-chart-control-orientation="horizontal"`, `data-chart-control-scale="150"`, `data-chart-control-labels="off"`,
		`Applied: horizontal, 150% scale, labels off.`, `Fourteen exact values at 150% scale`,
		`method="get" action="/docs/chart-controls#static-chart-control-example"`,
		`method="get" action="/docs/chart-controls#interactive-chart-control-example"`,
		`data-chart-control-example="palette-grid"`, `data-chart-palette="custom"`,
		`Applied: custom palette #123456, #234567, #345678, #456789.`,
		`method="get" action="/docs/chart-controls#palette-chart-control-example"`,
		`#123456`, `#234567`, `#345678`, `#456789`,
		`Static form and chart · templ`, `Interactive form and chart · templ`, `Palette form and four charts · templ`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("submitted guide missing %q", want)
		}
	}
}

func TestChartControlGuideReportsInvalidRequestValuesAndRestoresDefaults(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/docs/chart-controls?static_present=1&static_mode=raw&static_stroke=99&interactive_present=1&interactive_orientation=raw&interactive_scale=125&palette_present=1&chart_palette=custom&palette_color_1=red&palette_color_2=%23abcdef&palette_color_3=bad&palette_color_4=", nil)
	New().ServeHTTP(recorder, request)
	body := recorder.Body.String()
	for _, want := range []string{
		`role="alert"`, `data-chart-control-errors="static"`, `data-chart-control-errors="interactive"`, `data-chart-control-errors="palette-grid"`,
		`enabled was restored`, `3 was restored`, `vertical was restored`, `100% was restored`,
		`data-chart-control-mode="enabled"`, `data-chart-control-stroke="3"`,
		`data-chart-control-orientation="vertical"`, `data-chart-control-scale="100"`,
		`Color 1 must use six-digit hexadecimal notation`, `#0e7490 was restored`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("invalid guide response missing %q", want)
		}
	}
}

func TestChartControlGuideReturnsOnlyRecognizedHTMXExampleTarget(t *testing.T) {
	t.Parallel()
	tests := []struct {
		target   string
		query    string
		want     string
		unwanted string
	}{
		{
			target:   "static-chart-control-example",
			query:    "static_present=1&static_mode=omitted&static_stroke=6&static_area=on",
			want:     `data-chart-control-mode="omitted"`,
			unwanted: `data-chart-control-example="interactive"`,
		},
		{
			target:   "interactive-chart-control-example",
			query:    "interactive_present=1&interactive_orientation=horizontal&interactive_scale=50&interactive_labels=on",
			want:     `data-chart-control-scale="50"`,
			unwanted: `data-chart-control-example="static"`,
		},
		{
			target:   "palette-chart-control-example",
			query:    "palette_present=1&chart_palette=status",
			want:     `data-chart-palette="status"`,
			unwanted: `data-chart-control-example="static"`,
		},
	}
	for _, test := range tests {
		t.Run(test.target, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/docs/chart-controls?"+test.query, nil)
			request.Header.Set("HX-Request", "true")
			request.Header.Set("HX-Target", test.target)
			New().ServeHTTP(recorder, request)
			body := recorder.Body.String()
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d", recorder.Code)
			}
			if !strings.Contains(body, wantOuterID(test.target)) || !strings.Contains(body, test.want) {
				t.Errorf("HTMX response missing target or state: %s", body)
			}
			for _, unwanted := range []string{"<html", `data-chart-controls-guide`, `data-chart-control-source=`, test.unwanted} {
				if strings.Contains(body, unwanted) {
					t.Errorf("HTMX example response contains %q", unwanted)
				}
			}
		})
	}
}

func wantOuterID(target string) string { return `id="` + target + `"` }

func TestChartControlGuideKeepsNormalHTMXNavigationFragment(t *testing.T) {
	t.Parallel()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/docs/chart-controls", nil)
	request.Header.Set("HX-Request", "true")
	request.Header.Set("HX-Target", "main-content")
	New().ServeHTTP(recorder, request)
	body := recorder.Body.String()
	for _, want := range []string{`id="main-content"`, `data-chart-controls-guide`, `data-chart-control-examples`} {
		if !strings.Contains(body, want) {
			t.Errorf("navigation fragment missing %q", want)
		}
	}
}
