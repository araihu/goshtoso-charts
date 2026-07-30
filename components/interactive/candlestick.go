package interactive

import interactivecandlestick "github.com/araihu/goshtoso-charts/components/interactive/candlestick"

// CandlestickConfig is the compatibility name for candlestick.Config.
type CandlestickConfig = interactivecandlestick.Config

// Candle is the compatibility name for candlestick.Candle.
type Candle = interactivecandlestick.Candle

// CandlestickSeries is the compatibility name for candlestick.Series.
type CandlestickSeries = interactivecandlestick.Series

// CandlestickSeriesOptions is the compatibility name for candlestick.SeriesOptions.
type CandlestickSeriesOptions = interactivecandlestick.SeriesOptions

// CandlestickDirectionStyle is the compatibility name for candlestick.DirectionStyle.
type CandlestickDirectionStyle = interactivecandlestick.DirectionStyle

// CandlestickMarkOptions is the compatibility name for candlestick.MarkOptions.
type CandlestickMarkOptions = interactivecandlestick.MarkOptions

// CandlestickDataZoomType is the compatibility name for candlestick.DataZoomType.
type CandlestickDataZoomType = interactivecandlestick.DataZoomType

const (
	CandlestickDataZoomSlider CandlestickDataZoomType = interactivecandlestick.DataZoomSlider
	CandlestickDataZoomInside CandlestickDataZoomType = interactivecandlestick.DataZoomInside
)

// CandlestickDataZoomAxis is the compatibility name for candlestick.DataZoomAxis.
type CandlestickDataZoomAxis = interactivecandlestick.DataZoomAxis

const (
	CandlestickDataZoomXAxis CandlestickDataZoomAxis = interactivecandlestick.DataZoomXAxis
	CandlestickDataZoomYAxis CandlestickDataZoomAxis = interactivecandlestick.DataZoomYAxis
)

// CandlestickDataZoom is the compatibility name for candlestick.DataZoom.
type CandlestickDataZoom = interactivecandlestick.DataZoom

// Candlestick forwards to the canonical candlestick package.
func Candlestick(cfg CandlestickConfig) Instance { return interactivecandlestick.Candlestick(cfg) }
