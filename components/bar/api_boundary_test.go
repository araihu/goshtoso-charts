package bar_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/components/bar"
)

func TestPublicAPIDoesNotExposeRendererTypes(t *testing.T) {
	t.Parallel()
	publicTypes := []reflect.Type{
		reflect.TypeOf(bar.Orientation("")), reflect.TypeOf(bar.Padding{}),
		reflect.TypeOf(bar.ValueFormat("")), reflect.TypeOf(bar.DataLabelPosition("")), reflect.TypeOf(bar.DataLabelOptions{}),
		reflect.TypeOf(bar.GeometryOptions{}), reflect.TypeOf(bar.LegendPlacement("")), reflect.TypeOf(bar.LegendOptions{}),
		reflect.TypeOf(bar.ValueAxisOptions{}), reflect.TypeOf(bar.ReferenceStyle{}), reflect.TypeOf(bar.References{}),
		reflect.TypeOf(bar.Series{}), reflect.TypeOf(bar.Config{}),
		reflect.TypeOf(bar.Instance{}), reflect.TypeOf(bar.Bar),
	}
	seen := make(map[reflect.Type]bool)
	for _, publicType := range publicTypes {
		assertRendererNeutral(t, publicType, seen)
	}
}

func assertRendererNeutral(t *testing.T, publicType reflect.Type, seen map[reflect.Type]bool) {
	t.Helper()
	if publicType == nil || seen[publicType] {
		return
	}
	seen[publicType] = true
	if strings.HasPrefix(publicType.PkgPath(), "github.com/go-analyze/charts") {
		t.Errorf("public API exposes renderer type %s", publicType)
		return
	}
	switch publicType.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array:
		assertRendererNeutral(t, publicType.Elem(), seen)
	case reflect.Map:
		assertRendererNeutral(t, publicType.Key(), seen)
		assertRendererNeutral(t, publicType.Elem(), seen)
	case reflect.Struct:
		for index := 0; index < publicType.NumField(); index++ {
			if field := publicType.Field(index); field.IsExported() {
				assertRendererNeutral(t, field.Type, seen)
			}
		}
	case reflect.Func:
		for index := 0; index < publicType.NumIn(); index++ {
			assertRendererNeutral(t, publicType.In(index), seen)
		}
		for index := 0; index < publicType.NumOut(); index++ {
			assertRendererNeutral(t, publicType.Out(index), seen)
		}
	}
}
