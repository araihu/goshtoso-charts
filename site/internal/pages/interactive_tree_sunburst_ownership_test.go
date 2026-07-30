package pages

import (
	"reflect"
	"strings"
	"testing"
)

func TestTreeAndSunburstSiteUseCanonicalChildPackages(t *testing.T) {
	t.Parallel()

	for name, config := range map[string]struct {
		value any
		path  string
	}{
		"Tree":     {value: sampleInteractiveTree(), path: "github.com/araihu/goshtoso-charts/components/interactive/tree"},
		"Sunburst": {value: sampleInteractiveSunburst(), path: "github.com/araihu/goshtoso-charts/components/interactive/sunburst"},
	} {
		if got := reflect.TypeOf(config.value).PkgPath(); got != config.path {
			t.Errorf("site %s config package = %q, want %q", name, got, config.path)
		}
	}

	for name, check := range map[string]struct {
		code      string
		canonical []string
		legacy    string
	}{
		"Tree": {
			code:      interactiveTreeCode(),
			canonical: []string{"@interactivetree.Tree(interactivetree.Config{", "[]*interactivetree.Node", "interactivetree.OrientationLeftToRight", "interactivetree.Insets", "chart.Int", "chart.LabelOptions"},
			legacy:    "interactive.Tree",
		},
		"Sunburst": {
			code:      interactiveSunburstCode(),
			canonical: []string{"@interactivesunburst.Sunburst(interactivesunburst.Config{", "[]*interactivesunburst.Node", "interactivesunburst.NavigationDrillDown", "interactivesunburst.SortDescending", "chart.LabelOptions", "charttheme.Style"},
			legacy:    "interactive.Sunburst",
		},
	} {
		for _, want := range check.canonical {
			if !strings.Contains(check.code, want) {
				t.Errorf("%s snippet missing canonical API %q", name, want)
			}
		}
		if strings.Contains(check.code, check.legacy) {
			t.Errorf("%s snippet still teaches compatibility facade", name)
		}
	}
}
