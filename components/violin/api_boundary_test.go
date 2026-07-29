package violin_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/components/violin"
)

func TestPublicAPIDoesNotExposeRendererTypes(t *testing.T) {
	t.Parallel()
	types := []reflect.Type{
		reflect.TypeOf(violin.Axis{}), reflect.TypeOf(violin.Config{}), reflect.TypeOf(violin.Distribution{}),
		reflect.TypeOf(violin.Instance{}), reflect.TypeOf(violin.MarkLines{}), reflect.TypeOf(violin.Normalization("")),
		reflect.TypeOf(violin.Padding{}), reflect.TypeOf(violin.Series{}), reflect.TypeOf(violin.Violin),
		reflect.TypeOf(violin.Statistics{}),
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
