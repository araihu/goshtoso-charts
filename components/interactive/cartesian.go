package interactive

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
