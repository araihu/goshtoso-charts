package echartsexamples

import (
	"strings"
	"testing"
)

func TestSupportExamplesBuildRenderableSnippet(t *testing.T) {
	t.Parallel()
	if len(SupportExamples) != 17 {
		t.Fatalf("support entries = %d, want 17", len(SupportExamples))
	}

	for _, example := range SupportExamples {
		snippet := example.Build().RenderSnippet()
		if !strings.Contains(snippet.Script, "echarts.init") || !strings.Contains(snippet.Element, "id=") {
			t.Fatalf("%s did not render a chart snippet", example.Slug)
		}
	}
}

func TestPageLayoutDescriptorsRemainSnippetCompatible(t *testing.T) {
	t.Parallel()
	for _, layout := range []string{"center", "flex", "none"} {
		snippet := pageLayoutDescriptor(layout).RenderSnippet()
		if !strings.Contains(snippet.Option, "Page "+layout+" layout") {
			t.Fatalf("%s layout descriptor missing from option", layout)
		}
	}
}
