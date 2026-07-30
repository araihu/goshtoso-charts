package pages

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/components/funnel"
)

func TestStaticFunnelUpstreamCoverageIsExhaustiveAndPinned(t *testing.T) {
	t.Parallel()
	if staticFunnelUpstreamRevision != "1fe31b06b8a82e00df877ff4417a75858547c1c2" {
		t.Fatalf("static Funnel upstream revision = %q", staticFunnelUpstreamRevision)
	}
	wantCoverage := []staticFunnelCoverageEntry{
		{Path: "examples/1-Painter/funnel_chart-1-basic/main.go", SHA256: "a54875c89cd0be43fa7c0614520a11c489562bbc72f8c577a314eb3f24f75a6d", Treatment: "Basic seven-stage funnel"},
		{Path: "examples/2-OptionFunc/funnel_chart-1-basic/main.go", SHA256: "332ffeb340e1236170f41a3a93a46499897af75df261ed60f7ac01dd35ae4893", Treatment: "Same seven-stage funnel through option functions"},
	}
	if got := staticFunnelUpstreamCoverage(); !reflect.DeepEqual(got, wantCoverage) {
		t.Fatalf("static Funnel coverage = %#v, want %#v", got, wantCoverage)
	}
	wantSpans := []staticFunnelSourceSpan{
		{Path: "examples/1-Painter/funnel_chart-1-basic/main.go", Lines: "14-22", Name: "writeFile", SHA256: "9a3e255a47b40c36123b225677519f09a62c7f0f38cd1341515d30ce687f1fc5", Role: "Filesystem output helper; no chart behavior"},
		{Path: "examples/1-Painter/funnel_chart-1-basic/main.go", Lines: "24-46", Name: "main", SHA256: "82bd67be63d88d6ef2312890964ed595ff9a12990ee89bc4edc85e4bc93f9892", Role: "Seven-stage data, title, legend padding, and 600x400 presentation"},
		{Path: "examples/2-OptionFunc/funnel_chart-1-basic/main.go", Lines: "14-22", Name: "writeFile", SHA256: "9a3e255a47b40c36123b225677519f09a62c7f0f38cd1341515d30ce687f1fc5", Role: "Filesystem output helper; no chart behavior"},
		{Path: "examples/2-OptionFunc/funnel_chart-1-basic/main.go", Lines: "24-44", Name: "main", SHA256: "1289983e007306c2d3b1d44136aaa35ec56691a5d05f502a3c718d6d356b69a3", Role: "Duplicate seven-stage presentation through option functions"},
		{Path: "examples/2-OptionFunc/web-1/main.go", Lines: "145-547", Name: "indexHandler", SHA256: "c9970d3ac7aab849623793deee832f1c1fec5569c4964e3b84b61857424b8068", Role: "Mixed-family web composition containing the supplementary Funnel literal"},
		{Path: "examples/2-OptionFunc/web-1/main.go", Lines: "450-487", Name: "Funnel literal", SHA256: "f2a990a30d9a89a7d982bc79522c1fb11b2611173ac54c31a19382a543b1ba45", Role: "Five-stage supplementary dataset and default presentation"},
	}
	if got := staticFunnelSourceSpans(); !reflect.DeepEqual(got, wantSpans) {
		t.Fatalf("static Funnel source spans = %#v, want %#v", got, wantSpans)
	}
	wantAPISources := []staticFunnelAPISource{
		{Path: "funnel_chart.go", SHA256: "d8c1d8e5ee84ae0f534da3046bd7d43973f7afbf9298e8b8d86c4200f0db6cfb"},
		{Path: "series.go", SHA256: "953f4e5d555701348ebcb8eb0bfe1753a6df56eb4f94a86403c6dc6cecf79217"},
		{Path: "series_label.go", SHA256: "d7b176bea3679542e878c4c5703db3711d1e258b64efb7d19614a16fc5722611"},
		{Path: "title.go", SHA256: "e85f6a0fe2e8fd7c253ac226d780164beb0cc7214e94979241d1e9fbce824b26"},
		{Path: "legend.go", SHA256: "eaad1144ff1c5af84049a2968e952caba2dac3970558d542c18c6a98ba6d515b"},
		{Path: "chart_option.go", SHA256: "0b298fcd45fab6bbe476514d90e5107d65d32bd8ba985f2411e79ec88fd2b858"},
	}
	if got := staticFunnelAPISources(); !reflect.DeepEqual(got, wantAPISources) {
		t.Fatalf("static Funnel API sources = %#v, want %#v", got, wantAPISources)
	}
}

func TestStaticFunnelExamplesPreservePinnedDataAndPresentation(t *testing.T) {
	t.Parallel()
	basic := sampleBasicFunnel()
	compact := sampleCompactFunnel()
	wantBasic := []funnel.Stage{
		{Label: "Show", Value: 100}, {Label: "Click", Value: 80}, {Label: "Visit", Value: 60},
		{Label: "Inquiry", Value: 40}, {Label: "Order", Value: 20}, {Label: "Pay", Value: 10},
		{Label: "Cancel", Value: 2},
	}
	wantCompact := append([]funnel.Stage(nil), wantBasic[:5]...)
	if !reflect.DeepEqual(basic.Stages, wantBasic) || basic.Title != "Funnel" || basic.Options.Legend.Padding.Left != 100 || basic.Width != 0 || basic.Height != 0 {
		t.Fatalf("basic Funnel treatment drifted: %#v", basic)
	}
	if !reflect.DeepEqual(compact.Stages, wantCompact) || compact.Title != "Funnel" || compact.Options != (funnel.Options{}) || compact.Width != 0 || compact.Height != 0 {
		t.Fatalf("compact Funnel treatment drifted: %#v", compact)
	}
	if basic.RootAttrs["data-static-funnel-source"] != "a54875c89cd0be43" || compact.RootAttrs["data-static-funnel-source"] != "f2a990a30d9a89a7" {
		t.Fatalf("Funnel source markers = %#v / %#v", basic.RootAttrs, compact.RootAttrs)
	}
}

func TestStaticFunnelSnippetsMatchBothPreviews(t *testing.T) {
	t.Parallel()
	for _, want := range []string{
		`Label: "Basic funnel"`, `Title: "Funnel"`, `{Label: "Cancel", Value: 2}`, `Padding{Left: 100}`,
	} {
		if !strings.Contains(funnelCode(), want) {
			t.Errorf("basic Funnel snippet missing %q", want)
		}
	}
	for _, want := range []string{
		`Label: "Compact five-stage funnel"`, `Title: "Funnel"`, `{Label: "Order", Value: 20}`,
	} {
		if !strings.Contains(compactFunnelCode(), want) {
			t.Errorf("compact Funnel snippet missing %q", want)
		}
	}
}

func TestStaticFunnelAttributionAndLedgerAreComplete(t *testing.T) {
	t.Parallel()
	if len(chartAttributions) == 0 {
		t.Fatal("chart attributions are empty")
	}
	attribution := chartAttributions[0]
	if attribution.Name != "go-analyze/charts" || !strings.Contains(attribution.Version, staticFunnelUpstreamRevision) {
		t.Fatalf("static attribution source = %#v", attribution)
	}
	for _, want := range []string{
		"examples/1-Painter/funnel_chart-1-basic/main.go",
		"examples/2-OptionFunc/funnel_chart-1-basic/main.go",
		"examples/2-OptionFunc/web-1/main.go",
	} {
		if !strings.Contains(attribution.UsedFor, want) {
			t.Errorf("static Funnel attribution missing %q", want)
		}
	}
	data, err := os.ReadFile("../../../docs/upstream-example-coverage.md")
	if err != nil {
		t.Fatalf("read coverage ledger: %v", err)
	}
	section := canonicalLedgerSection(t, string(data), "## Static/vector Funnel")
	for _, entry := range staticFunnelUpstreamCoverage() {
		if count := strings.Count(section, "`"+entry.Path+"`"); count != 1 {
			t.Errorf("static Funnel coverage row %q occurs %d times, want 1", entry.Path, count)
		}
		if !strings.Contains(section, "`"+entry.SHA256+"`") {
			t.Errorf("static Funnel row %q missing SHA-256 %s", entry.Path, entry.SHA256)
		}
	}
	for _, span := range staticFunnelSourceSpans() {
		for _, want := range []string{span.Lines, "`" + span.SHA256 + "`"} {
			if !strings.Contains(section, want) {
				t.Errorf("static Funnel span %s %s missing %q", span.Path, span.Name, want)
			}
		}
	}
	for _, source := range staticFunnelAPISources() {
		for _, want := range []string{"`" + source.Path + "`", "`" + source.SHA256 + "`"} {
			if !strings.Contains(section, want) {
				t.Errorf("static Funnel API source %s missing %q", source.Path, want)
			}
		}
	}
	for _, want := range []string{
		"both dedicated Funnel-family files", "one distinct", "outside the two-file dedicated coverage denominator",
		"the second example on the page", "Unsupported dedicated Funnel-family behaviors: none", "typed renderer-neutral configuration",
	} {
		if !strings.Contains(section, want) {
			t.Errorf("static Funnel ledger missing boundary statement %q", want)
		}
	}
}

func TestStaticFunnelPageDocumentsDecisionsAccessibilityAndCanonicalAPI(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("pages.templ")
	if err != nil {
		t.Fatalf("read pages.templ: %v", err)
	}
	page := string(source)
	start := strings.Index(page, "templ funnelContent")
	if start < 0 {
		t.Fatal("cannot isolate static Funnel page")
	}
	end := strings.Index(page[start:], "templ heatMapContent")
	if end < 0 {
		t.Fatal("cannot isolate static Funnel page")
	}
	funnelPage := page[start : start+end]
	for _, want := range []string{
		"AbovePreview: visualizationGuide(", "fixed process", "increase, branch, recur", "adjacent disclosure",
		"Compact five-stage sequence", "shorter ordered sequence", "Static/vector behavior", "inline SVG", "print", "without JavaScript",
		"shared chart controls", "chart modes", `chartDocumentation(`, `"funnel"`,
	} {
		if !strings.Contains(funnelPage, want) {
			t.Errorf("static Funnel docs missing %q", want)
		}
	}
	for _, forbidden := range []string{"Component contract", "Primitive", "Kind", "Configuration", "go-analyze", "go-echarts", "infrastructure", "operations"} {
		if strings.Contains(funnelPage, forbidden) {
			t.Errorf("static Funnel docs retain redundant or renderer-specific copy %q", forbidden)
		}
	}
}
