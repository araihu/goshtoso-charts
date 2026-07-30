package pages

import (
	"reflect"
	"strings"
	"testing"
)

func TestGraphAndSankeySiteUseCanonicalChildPackages(t *testing.T) {
	t.Parallel()

	for name, config := range map[string]struct {
		value any
		path  string
	}{
		"Graph":  {value: sampleInteractiveGraph(), path: "github.com/araihu/goshtoso-charts/components/interactive/graph"},
		"Sankey": {value: sampleInteractiveSankey(), path: "github.com/araihu/goshtoso-charts/components/interactive/sankey"},
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
		"Graph": {
			code:      interactiveGraphCode(),
			canonical: []string{"@interactivegraph.Graph(interactivegraph.Config{", "[]interactivegraph.Node", "[]interactivegraph.Link", "chart.ChartOptions"},
			legacy:    "interactive.Graph",
		},
		"Sankey": {
			code:      interactiveSankeyCode(),
			canonical: []string{"@interactivesankey.Sankey(interactivesankey.Config{", "[]interactivesankey.Series", "[]interactivesankey.Node", "[]interactivesankey.Link", "chart.ChartOptions"},
			legacy:    "interactive.Sankey",
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
