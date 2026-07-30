package pages

import (
	"reflect"
	"strings"
	"testing"
)

func TestParallelAndThemeRiverSiteUseCanonicalChildPackages(t *testing.T) {
	t.Parallel()

	for name, config := range map[string]struct {
		value any
		path  string
	}{
		"Parallel":   {value: sampleInteractiveParallel(), path: "github.com/araihu/goshtoso-charts/components/interactive/parallel"},
		"ThemeRiver": {value: sampleInteractiveThemeRiver(), path: "github.com/araihu/goshtoso-charts/components/interactive/themeriver"},
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
		"Parallel": {
			code:      interactiveParallelCode(),
			canonical: []string{"@interactiveparallel.Parallel(interactiveparallel.Config{", "[]interactiveparallel.Dimension", "[]interactiveparallel.Series", "chart.ChartOptions"},
			legacy:    "interactive.Parallel",
		},
		"ThemeRiver": {
			code:      interactiveThemeRiverCode(),
			canonical: []string{"@interactivethemeriver.ThemeRiver(interactivethemeriver.Config{", "[]interactivethemeriver.Stream", "[]interactivethemeriver.Point", "chart.ChartOptions"},
			legacy:    "interactive.ThemeRiver",
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
