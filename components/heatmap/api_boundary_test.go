package heatmap_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/components/heatmap"
)

func TestPublicAPIDoesNotExposeRendererTypes(t *testing.T) {
	t.Parallel()
	types := []reflect.Type{
		reflect.TypeOf(heatmap.Axis{}), reflect.TypeOf(heatmap.Cell{}), reflect.TypeOf(heatmap.Config{}),
		reflect.TypeOf(heatmap.Gradient{}), reflect.TypeOf(heatmap.GradientStop{}), reflect.TypeOf(heatmap.Instance{}),
		reflect.TypeOf(heatmap.ValueRange{}), reflect.TypeOf(heatmap.HeatMap),
	}
	seen := map[reflect.Type]bool{}
	for _, publicType := range types {
		assertNeutral(t, publicType, seen)
	}
}

func assertNeutral(t *testing.T, value reflect.Type, seen map[reflect.Type]bool) {
	t.Helper()
	if value == nil || seen[value] {
		return
	}
	seen[value] = true
	if strings.HasPrefix(value.PkgPath(), "github.com/go-analyze/charts") {
		t.Errorf("public API exposes renderer type %s", value)
		return
	}
	switch value.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array:
		assertNeutral(t, value.Elem(), seen)
	case reflect.Map:
		assertNeutral(t, value.Key(), seen)
		assertNeutral(t, value.Elem(), seen)
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			if value.Field(index).IsExported() {
				assertNeutral(t, value.Field(index).Type, seen)
			}
		}
	case reflect.Func:
		for index := 0; index < value.NumIn(); index++ {
			assertNeutral(t, value.In(index), seen)
		}
		for index := 0; index < value.NumOut(); index++ {
			assertNeutral(t, value.Out(index), seen)
		}
	}
}
