package pages

import (
	"reflect"
	"strings"
	"testing"
)

func TestInteractiveGaugeSiteUsesCanonicalChildPackage(t *testing.T) {
	t.Parallel()

	config := sampleInteractiveGauge()
	if got := reflect.TypeOf(config).PkgPath(); got != "github.com/araihu/goshtoso-charts/components/interactive/gauge" {
		t.Fatalf("site Gauge config package = %q", got)
	}
	for name, code := range map[string]string{
		"progress": interactiveChartGaugeCode(),
		"liquid":   interactiveGaugeLiquidCode(),
	} {
		for _, want := range []string{"@interactivegauge.Gauge(interactivegauge.Config{", "[]interactivegauge.Series", "[]interactivegauge.Data"} {
			if !strings.Contains(code, want) {
				t.Errorf("%s Gauge snippet missing canonical API %q", name, want)
			}
		}
		if strings.Contains(code, "interactive.Gauge") {
			t.Errorf("%s Gauge snippet still teaches compatibility facade", name)
		}
	}
}
