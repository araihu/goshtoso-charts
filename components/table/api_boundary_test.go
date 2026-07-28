package table_test

import (
	"reflect"
	"strings"
	"testing"

	charttable "github.com/araihu/goshtoso-charts/components/table"
)

func TestPublicAPIDoesNotExposeRendererTypes(t *testing.T) {
	t.Parallel()
	publicTypes := []reflect.Type{
		reflect.TypeOf(charttable.Alignment("")), reflect.TypeOf(charttable.Column{}), reflect.TypeOf(charttable.Padding{}),
		reflect.TypeOf(charttable.Colors{}), reflect.TypeOf(charttable.Cell{}), reflect.TypeOf(charttable.CellAppearance{}),
		reflect.TypeOf(charttable.CellStyler(nil)), reflect.TypeOf(charttable.Config{}), reflect.TypeOf(charttable.Instance{}),
		reflect.TypeOf(charttable.Table),
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
