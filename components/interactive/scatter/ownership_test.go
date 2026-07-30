package scatter_test

import (
	"reflect"
	"testing"

	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/interactive"
	interactivescatter "github.com/araihu/goshtoso-charts/components/interactive/scatter"
)

func TestScatterSpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()

	wantPackage := "github.com/araihu/goshtoso-charts/components/interactive/scatter"
	scatterTypes := []reflect.Type{
		reflect.TypeOf(interactivescatter.Config{}),
		reflect.TypeOf(interactivescatter.Series{}),
		reflect.TypeOf(interactivescatter.Data{}),
		reflect.TypeOf(interactivescatter.Variant("")),
		reflect.TypeOf(interactivescatter.AxisType("")),
	}
	for _, scatterType := range scatterTypes {
		if got := scatterType.PkgPath(); got != wantPackage {
			t.Errorf("%s PkgPath() = %q, want %q", scatterType, got, wantPackage)
		}
	}

	aliases := []struct {
		legacy    reflect.Type
		canonical reflect.Type
	}{
		{reflect.TypeOf(interactive.ScatterConfig{}), reflect.TypeOf(interactivescatter.Config{})},
		{reflect.TypeOf(interactive.ScatterSeries{}), reflect.TypeOf(interactivescatter.Series{})},
		{reflect.TypeOf(interactive.ScatterData{}), reflect.TypeOf(interactivescatter.Data{})},
		{reflect.TypeOf(interactive.ScatterVariant("")), reflect.TypeOf(interactivescatter.Variant(""))},
		{reflect.TypeOf(interactive.CartesianAxisType("")), reflect.TypeOf(interactivescatter.AxisType(""))},
	}
	for _, alias := range aliases {
		if alias.legacy != alias.canonical {
			t.Errorf("legacy type %s is not exact alias of %s", alias.legacy, alias.canonical)
		}
	}
}

var (
	_ func(interactivescatter.Config) chart.Instance = interactivescatter.Scatter
	_ func(interactivescatter.Config) chart.Instance = interactive.Scatter
	_ func(interactive.ScatterConfig) chart.Instance = interactivescatter.Scatter
)
