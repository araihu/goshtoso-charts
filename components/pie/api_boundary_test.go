package pie

import (
	"reflect"
	"strings"
	"testing"
)

const forbiddenAPIPackagePrefix = "github.com/go-analyze/charts"

func TestPiePublicAPIDoesNotLeakRendererTypes(t *testing.T) {
	t.Parallel()
	seen := map[reflect.Type]bool{}
	for _, value := range []any{
		Slice{}, VariantPie, TitleOptions{}, LegendOptions{}, RadiusOptions{}, RadiusScaleArea,
		LabelOptions{}, LabelPlacementInside, CenterOptions{}, CenterContentTotal,
		ValueFormatHumanized, Padding{}, Config{}, Instance{},
		Pie,
	} {
		assertNoRendererType(t, reflect.TypeOf(value), seen)
	}
}

func assertNoRendererType(t *testing.T, value reflect.Type, seen map[reflect.Type]bool) {
	t.Helper()
	if value == nil || seen[value] {
		return
	}
	seen[value] = true
	if strings.HasPrefix(value.PkgPath(), forbiddenAPIPackagePrefix) {
		t.Fatalf("public API leaks renderer type %s", value)
	}
	switch value.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array:
		assertNoRendererType(t, value.Elem(), seen)
	case reflect.Map:
		assertNoRendererType(t, value.Key(), seen)
		assertNoRendererType(t, value.Elem(), seen)
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			if value.Field(index).PkgPath == "" {
				assertNoRendererType(t, value.Field(index).Type, seen)
			}
		}
	case reflect.Func:
		for index := 0; index < value.NumIn(); index++ {
			assertNoRendererType(t, value.In(index), seen)
		}
		for index := 0; index < value.NumOut(); index++ {
			assertNoRendererType(t, value.Out(index), seen)
		}
	}
}
