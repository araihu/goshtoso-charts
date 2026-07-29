package candlestick_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/components/candlestick"
)

func TestPublicAPIDoesNotExposeRendererTypes(t *testing.T) {
	t.Parallel()
	publicTypes := []reflect.Type{
		reflect.TypeOf(candlestick.Axis{}), reflect.TypeOf(candlestick.Config{}), reflect.TypeOf(candlestick.Datum{}),
		reflect.TypeOf(candlestick.AggregationOptions{}),
		reflect.TypeOf(candlestick.CandleStyle{}), reflect.TypeOf(candlestick.Options{}), reflect.TypeOf(candlestick.Padding{}),
		reflect.TypeOf(candlestick.TrendLine{}), reflect.TypeOf(candlestick.TrendType("")),
		reflect.TypeOf(candlestick.PatternOptions{}), reflect.TypeOf(candlestick.PatternResult{}), reflect.TypeOf(candlestick.PatternLabelStyle{}),
		reflect.TypeOf(candlestick.PatternSelection("")), reflect.TypeOf(candlestick.PatternType("")), reflect.TypeOf(candlestick.PatternLabelText("")), reflect.TypeOf(candlestick.CloseReferenceType("")),
		reflect.TypeOf(candlestick.Instance{}), reflect.TypeOf(candlestick.Candlestick), reflect.TypeOf(candlestick.DetectPatterns),
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
