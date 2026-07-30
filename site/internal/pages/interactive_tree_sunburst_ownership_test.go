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

	sunburstCode := interactiveSunburstCode()
	for _, want := range []string{
		`{Name: "parent-0", Value: 0.81`, `{Name: "child-0", Value: 0.34}`,
		`{Name: "parent-1", Value: 0.62`, `{Name: "child-1", Value: 0.57}`,
		`{Name: "parent-2", Value: 0.45`, `{Name: "child-2", Value: 0.73}`,
		`{Name: "parent-3", Value: 0.93`, `{Name: "child-3", Value: 0.28}`,
		`{Name: "parent-4", Value: 0.38`, `{Name: "child-4", Value: 0.66}`,
		`{Name: "parent-5", Value: 0.71`, `{Name: "child-5", Value: 0.49}`,
		`{Name: "parent-6", Value: 0.54`, `{Name: "child-6", Value: 0.87}`,
	} {
		if !strings.Contains(sunburstCode, want) {
			t.Errorf("Sunburst snippet elides pinned hierarchy entry %q", want)
		}
	}
	if strings.Contains(sunburstCode, "parent-1 through") {
		t.Error("Sunburst snippet retains an elision comment instead of the pinned seven-pair dataset")
	}
}
