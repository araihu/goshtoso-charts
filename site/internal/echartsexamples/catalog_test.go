package echartsexamples

import (
	"strings"
	"testing"
)

func TestCatalogPortsEveryCurrent2DExampleVariation(t *testing.T) {
	t.Parallel()
	examples := All()
	if len(examples) != 121 {
		t.Fatalf("catalog entries = %d, want 121", len(examples))
	}
	seen := make(map[string]bool, len(examples))
	for _, example := range examples {
		if example.Slug == "" || example.Title == "" || example.Group == "" || example.Build == nil {
			t.Fatalf("invalid catalog entry: %+v", example)
		}
		if seen[example.Slug] {
			t.Fatalf("duplicate catalog slug %q", example.Slug)
		}
		seen[example.Slug] = true
		if !strings.HasPrefix(example.Source, "https://github.com/go-echarts/examples/blob/master/examples/") {
			t.Fatalf("%s source = %q", example.Slug, example.Source)
		}
		snippet := example.Build().RenderSnippet()
		if !strings.Contains(snippet.Script, "echarts.init") || !strings.Contains(snippet.Element, "id=") {
			t.Fatalf("%s did not render a go-echarts snippet", example.Slug)
		}
	}
}
