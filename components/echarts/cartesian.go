package echarts

import "reflect"

// CartesianAxisType selects how scatter-family x values are interpreted.
type CartesianAxisType string

const (
	// CartesianAxisCategory uses XAxis as an ordered category list. It is the default.
	CartesianAxisCategory CartesianAxisType = "category"
	// CartesianAxisValue reads numeric x/y coordinates from each data value.
	CartesianAxisValue CartesianAxisType = "value"
)

func resolvedCartesianAxisType(axisType CartesianAxisType) CartesianAxisType {
	if axisType == "" {
		return CartesianAxisCategory
	}
	return axisType
}

func isCartesianCoordinate(value any) bool {
	reflected := reflect.ValueOf(value)
	if !reflected.IsValid() || (reflected.Kind() != reflect.Array && reflected.Kind() != reflect.Slice) || reflected.Len() != 2 {
		return false
	}
	for index := 0; index < reflected.Len(); index++ {
		switch reflected.Index(index).Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
		default:
			return false
		}
	}
	return true
}
