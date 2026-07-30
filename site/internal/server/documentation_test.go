package server

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"strings"
	"testing"
)

var componentDocumentationRoutes = map[string]string{
	"/components/bar": "bar", "/components/line": "line", "/components/pie": "pie", "/components/scatter": "scatter", "/components/radar": "radar", "/components/candlestick": "candlestick", "/components/funnel": "funnel", "/components/heatmap": "heatmap", "/components/table": "table", "/components/violin": "violin",
	"/components/interactive/bar": "interactive", "/components/interactive/line": "interactive", "/components/interactive/scatter": "interactive", "/components/interactive/scatter-3d": "interactive", "/components/interactive/bar-3d": "interactive", "/components/interactive/surface-3d": "interactive", "/components/interactive/line-3d": "interactive", "/components/interactive/pie": "interactive", "/components/interactive/radar": "interactive", "/components/interactive/heatmap": "interactive", "/components/interactive/boxplot": "interactive", "/components/interactive/candlestick": "interactive", "/components/interactive/gauge": "interactive", "/components/interactive/funnel": "interactive", "/components/interactive/graph": "interactive", "/components/interactive/sankey": "interactive", "/components/interactive/tree": "interactive/tree", "/components/interactive/sunburst": "interactive/sunburst", "/components/interactive/treemap": "interactive", "/components/interactive/parallel": "interactive", "/components/interactive/theme-river": "interactive", "/components/interactive/word-cloud": "interactive", "/components/interactive/map": "interactive", "/components/interactive/geo": "interactive",
}

func TestEveryComponentRouteUsesGuidanceAndGoAPIFooter(t *testing.T) {
	handler := New()
	assertDocumentationRoutesMatchNavigation(t, handler)
	for path, packageName := range componentDocumentationRoutes {
		t.Run(path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d", recorder.Code)
			}
			body := recorder.Body.String()
			for marker, want := range map[string]int{
				"data-visualization-guidance": 1,
				"data-go-api-reference":       1,
				"data-shared-chart-guidance":  1,
			} {
				if got := strings.Count(body, marker); got != want {
					t.Errorf("%s count = %d, want %d", marker, got, want)
				}
			}
			for _, heading := range []string{">Purpose</dt>", ">Use when</dt>", ">Avoid when</dt>", ">Equivalent data</dt>"} {
				if got := strings.Count(body, heading); got != 1 {
					t.Errorf("visualization guidance %s count = %d, want one", heading, got)
				}
			}
			if lastFooter, lastPreview := strings.LastIndex(body, "data-go-api-reference"), strings.LastIndex(body, "data-component-preview"); lastFooter <= lastPreview {
				t.Errorf("Go API footer must follow every chart preview: footer=%d preview=%d", lastFooter, lastPreview)
			}
			for _, forbidden := range []string{"PRIMITIVE", "KIND", "CONFIGURATION", "Component contract", "contract"} {
				if strings.Contains(body, forbidden) {
					t.Errorf("retains %q", forbidden)
				}
			}
			packagePath := "github.com/araihu/goshtoso-charts/components/" + packageName
			url := "https://pkg.go.dev/github.com/araihu/goshtoso-charts@v0.0.1/components/" + packageName
			for _, want := range []string{"data-go-api-reference", `data-go-api-version="v0.0.1"`, "Go API", "v0.0.1", "The examples above cover behavior and composition.", "pkg.go.dev is the canonical reference", packagePath, url, "Open v0.0.1 API"} {
				if !strings.Contains(body, want) {
					t.Errorf("Go API footer missing %q", want)
				}
			}
			for _, want := range []string{`data-shared-chart-guidance`, `href="/docs/chart-controls"`, `href="/docs/chart-modes"`, "Enabled, disabled, hidden, and omitted wrapper behavior", "static/vector and interactive capabilities"} {
				if !strings.Contains(body, want) {
					t.Errorf("shared chart guidance missing %q", want)
				}
			}
		})
	}
}

func assertDocumentationRoutesMatchNavigation(t *testing.T, handler http.Handler) {
	t.Helper()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/bar", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("navigation source status = %d", recorder.Code)
	}
	pattern := regexp.MustCompile(`href="(/components/(?:interactive/)?[a-z0-9-]+)"`)
	set := map[string]struct{}{}
	for _, match := range pattern.FindAllStringSubmatch(recorder.Body.String(), -1) {
		set[match[1]] = struct{}{}
	}
	navigationRoutes := make([]string, 0, len(set))
	for path := range set {
		navigationRoutes = append(navigationRoutes, path)
	}
	documentedRoutes := make([]string, 0, len(componentDocumentationRoutes))
	for path := range componentDocumentationRoutes {
		documentedRoutes = append(documentedRoutes, path)
	}
	sort.Strings(navigationRoutes)
	sort.Strings(documentedRoutes)
	if strings.Join(navigationRoutes, "\n") != strings.Join(documentedRoutes, "\n") {
		t.Fatalf("documentation route matrix differs from current chart navigation\nnavigation:\n%s\ndocumentation:\n%s", strings.Join(navigationRoutes, "\n"), strings.Join(documentedRoutes, "\n"))
	}
}

func TestComponentIntroductionsAndTopGuidanceAvoidImplementationBoilerplate(t *testing.T) {
	for path := range componentDocumentationRoutes {
		t.Run(path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d", recorder.Code)
			}
			introduction, guidance := visibleIntroductionAndGuidance(t, recorder.Body.String())
			for section, text := range map[string]string{"introduction": introduction, "top guidance": guidance} {
				for _, forbidden := range []string{"ssr", "svg", "renderer", "go-native", "browser runtime", "renderer-neutral", "typed", "component contract", "primitive", "backing engine", "reusable"} {
					if strings.Contains(strings.ToLower(text), forbidden) {
						t.Errorf("%s retains implementation boilerplate %q: %q", section, forbidden, text)
					}
				}
			}
		})
	}
}

func visibleIntroductionAndGuidance(t *testing.T, body string) (string, string) {
	t.Helper()
	introductionStart := strings.Index(body, `<p data-component-description`)
	if introductionStart < 0 {
		t.Fatal("component introduction not found")
	}
	introductionEnd := strings.Index(body[introductionStart:], "</p>")
	if introductionEnd < 0 {
		t.Fatal("component introduction does not close")
	}
	introduction := body[introductionStart : introductionStart+introductionEnd]
	purpose := strings.Index(body, ">Purpose</dt>")
	if purpose < 0 {
		t.Fatal("top visualization guidance not found")
	}
	guidanceStart := strings.LastIndex(body[:purpose], "<section")
	guidanceEnd := strings.Index(body[purpose:], "</section>")
	if guidanceStart < 0 || guidanceEnd < 0 {
		t.Fatal("top visualization guidance is incomplete")
	}
	return introduction, body[guidanceStart : purpose+guidanceEnd+len("</section>")]
}

func TestSunburstDocumentationExplainsNonVisualHierarchyAccess(t *testing.T) {
	recorder := httptest.NewRecorder()
	New().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/components/interactive/sunburst", nil))
	body := recorder.Body.String()
	for _, want := range []string{"shallow hierarchy", "Deep hierarchies", "keyboard navigation", "path-and-value table"} {
		if !strings.Contains(body, want) {
			t.Errorf("sunburst guidance missing %q", want)
		}
	}
	if strings.Contains(body, "Hierarchy contract") {
		t.Error("sunburst retains hierarchy contract")
	}
}
