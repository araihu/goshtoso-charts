package funnel_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/components/funnel"
)

func TestPublicAPIDoesNotExposeRendererTypes(t *testing.T) {
	t.Parallel()
	publicTypes := []reflect.Type{
		reflect.TypeOf(funnel.LabelMode("")), reflect.TypeOf(funnel.LegendOrientation("")), reflect.TypeOf(funnel.LegendPlacement("")),
		reflect.TypeOf(funnel.Padding{}), reflect.TypeOf(funnel.Legend{}), reflect.TypeOf(funnel.Options{}),
		reflect.TypeOf(funnel.Stage{}), reflect.TypeOf(funnel.Config{}), reflect.TypeOf(funnel.Instance{}),
		reflect.TypeOf(funnel.Funnel),
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
