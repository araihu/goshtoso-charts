package pages

import (
	"os"
	"strings"
	"testing"
)

func TestStaticBarPageDocumentsDecisionsAccessibilityAndCanonicalAPI(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("pages.templ")
	if err != nil {
		t.Fatalf("read pages.templ: %v", err)
	}
	page := string(source)
	start := strings.Index(page, "templ barContent")
	end := strings.Index(page[start:], "templ PiePage")
	if start < 0 || end < 0 {
		t.Fatal("cannot isolate static Bar page")
	}
	barPage := page[start : start+end]
	for _, want := range []string{
		"AbovePreview: visualizationGuide(", "rank", "continuous", "grouped", "stacked", "horizontal", "exact",
		`chartDocumentation(`, `"bar"`,
	} {
		if !strings.Contains(barPage, want) {
			t.Errorf("static Bar docs missing %q", want)
		}
	}
	if !strings.Contains(strings.ToLower(barPage), "static/vector") {
		t.Error("static Bar docs do not explain static/vector delivery")
	}
	for _, forbidden := range []string{"componentContract(", "Component contract", "Primitive", "Kind", "Configuration", "go-analyze", "go-echarts", "Collapse", "Fullscreen", "Export"} {
		if strings.Contains(barPage, forbidden) {
			t.Errorf("static Bar docs retain redundant or wrapper-specific copy %q", forbidden)
		}
	}
}
