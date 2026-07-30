package pages

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestTreemapAndWordCloudSiteUseCanonicalChildPackages(t *testing.T) {
	t.Parallel()

	for name, config := range map[string]struct {
		value any
		path  string
	}{
		"Treemap":   {value: sampleInteractiveTreemap(), path: "github.com/araihu/goshtoso-charts/components/interactive/treemap"},
		"WordCloud": {value: sampleInteractiveWordCloud("star shape", "star"), path: "github.com/araihu/goshtoso-charts/components/interactive/wordcloud"},
	} {
		if got := reflect.TypeOf(config.value).PkgPath(); got != config.path {
			t.Errorf("site %s config package = %q, want %q", name, got, config.path)
		}
	}

	for name, check := range map[string]struct {
		code      string
		canonical []string
		legacy    []string
	}{
		"Treemap": {
			code: interactiveTreemapCode(),
			canonical: []string{
				"@interactivetreemap.Treemap(interactivetreemap.Config{", "[]*interactivetreemap.Node",
				"interactivetreemap.NavigationDrillDown", "interactivetreemap.RoamEnabled",
				"interactivetreemap.Breadcrumb", "interactivetreemap.NodeStyle", "chart.LabelOptions", "chart.Int",
			},
			legacy: []string{"interactive.Treemap", "interactive.TreemapConfig", "interactive.TreemapNode"},
		},
		"WordCloud": {
			code: interactiveWordCloudCode(),
			canonical: []string{
				"@interactivewordcloud.WordCloud(interactivewordcloud.Config{", "interactivewordcloud.Series",
				"[]interactivewordcloud.Word", "interactivewordcloud.SeriesOptions",
				"interactivewordcloud.ShapeStar", "interactivewordcloud.SizeRange",
			},
			legacy: []string{"interactive.WordCloud", "interactive.WordCloudConfig", "interactive.WordCloudSeries"},
		},
	} {
		for _, want := range check.canonical {
			if !strings.Contains(check.code, want) {
				t.Errorf("%s snippet missing canonical API %q", name, want)
			}
		}
		for _, legacy := range check.legacy {
			if strings.Contains(check.code, legacy) {
				t.Errorf("%s snippet still teaches compatibility facade %q", name, legacy)
			}
		}
	}

	pageSource, err := os.ReadFile("pages.templ")
	if err != nil {
		t.Fatalf("read page template: %v", err)
	}
	page := string(pageSource)
	for _, want := range []string{
		"interactivetreemap.Treemap(sampleInteractiveTreemap())",
		"interactivewordcloud.WordCloud(sampleInteractiveWordCloud",
		`goAPIReference("interactive/treemap")`,
		`goAPIReference("interactive/wordcloud")`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("page template missing canonical ownership marker %q", want)
		}
	}
}
