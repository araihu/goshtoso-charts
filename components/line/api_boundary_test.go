package line_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/components/line"
)

func TestPublicAPIDoesNotExposeRendererTypes(t *testing.T) {
	t.Parallel()
	publicTypes := []reflect.Type{
		reflect.TypeOf(line.Title{}), reflect.TypeOf(line.Placement("")), reflect.TypeOf(line.SymbolShape("")), reflect.TypeOf(line.Symbol{}),
		reflect.TypeOf(line.Point{}), reflect.TypeOf(line.ValueFormat("")), reflect.TypeOf(line.LabelColorScale("")),
		reflect.TypeOf(line.DataLabelOptions{}), reflect.TypeOf(line.ReferenceStyle{}), reflect.TypeOf(line.References{}), reflect.TypeOf(line.TextAnnotation{}),
		reflect.TypeOf(line.AreaOptions{}), reflect.TypeOf(line.CategoryAxisOptions{}),
		reflect.TypeOf(line.LegendOptions{}), reflect.TypeOf(line.Padding{}), reflect.TypeOf(line.Axis{}), reflect.TypeOf(line.Series{}),
		reflect.TypeOf(line.Config{}), reflect.TypeOf(line.Instance{}), reflect.TypeOf(line.Line),
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
