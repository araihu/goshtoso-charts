package interactive

import interactivewordcloud "github.com/araihu/goshtoso-charts/components/interactive/wordcloud"

// WordCloudShape is retained as an alias for compatibility.
type WordCloudShape = interactivewordcloud.Shape

const (
	WordCloudShapeCircle          WordCloudShape = interactivewordcloud.ShapeCircle
	WordCloudShapeCardioid        WordCloudShape = interactivewordcloud.ShapeCardioid
	WordCloudShapeDiamond         WordCloudShape = interactivewordcloud.ShapeDiamond
	WordCloudShapeSquare          WordCloudShape = interactivewordcloud.ShapeSquare
	WordCloudShapeTriangleForward WordCloudShape = interactivewordcloud.ShapeTriangleForward
	WordCloudShapeTriangle        WordCloudShape = interactivewordcloud.ShapeTriangle
	WordCloudShapePentagon        WordCloudShape = interactivewordcloud.ShapePentagon
	WordCloudShapeStar            WordCloudShape = interactivewordcloud.ShapeStar
)

// WordCloudHorizontalPosition is retained as an alias for compatibility.
type WordCloudHorizontalPosition = interactivewordcloud.HorizontalPosition

const (
	WordCloudHorizontalDefault WordCloudHorizontalPosition = interactivewordcloud.HorizontalDefault
	WordCloudHorizontalLeft    WordCloudHorizontalPosition = interactivewordcloud.HorizontalLeft
	WordCloudHorizontalCenter  WordCloudHorizontalPosition = interactivewordcloud.HorizontalCenter
	WordCloudHorizontalRight   WordCloudHorizontalPosition = interactivewordcloud.HorizontalRight
)

// WordCloudVerticalPosition is retained as an alias for compatibility.
type WordCloudVerticalPosition = interactivewordcloud.VerticalPosition

const (
	WordCloudVerticalDefault WordCloudVerticalPosition = interactivewordcloud.VerticalDefault
	WordCloudVerticalTop     WordCloudVerticalPosition = interactivewordcloud.VerticalTop
	WordCloudVerticalCenter  WordCloudVerticalPosition = interactivewordcloud.VerticalCenter
	WordCloudVerticalBottom  WordCloudVerticalPosition = interactivewordcloud.VerticalBottom
)

// Word is retained as an alias for compatibility.
type Word = interactivewordcloud.Word

// WordCloudSizeRange is retained as an alias for compatibility.
type WordCloudSizeRange = interactivewordcloud.SizeRange

// WordCloudRotation is retained as an alias for compatibility.
type WordCloudRotation = interactivewordcloud.Rotation

// WordCloudLayout is retained as an alias for compatibility.
type WordCloudLayout = interactivewordcloud.Layout

// WordCloudSeriesOptions is retained as an alias for compatibility.
type WordCloudSeriesOptions = interactivewordcloud.SeriesOptions

// WordCloudSeries is retained as an alias for compatibility.
type WordCloudSeries = interactivewordcloud.Series

// WordCloudConfig is retained as an alias for compatibility.
type WordCloudConfig = interactivewordcloud.Config

// WordCloud forwards to the canonical child-package implementation.
func WordCloud(cfg WordCloudConfig) Instance { return interactivewordcloud.WordCloud(cfg) }
