package candlestick

import "math"

var patternSelections = map[PatternSelection][]PatternType{
	PatternSelectionAll: {
		PatternTypeBullishEngulfing, PatternTypeBearishEngulfing, PatternTypeHammer, PatternTypeMorningStar, PatternTypeEveningStar, PatternTypeShootingStar,
		PatternTypeDarkCloudCover, PatternTypeDragonflyDoji, PatternTypeGravestoneDoji, PatternTypeBearishMarubozu, PatternTypeBullishMarubozu, PatternTypePiercingLine,
		PatternTypeDoji, PatternTypeInvertedHammer,
	},
	PatternSelectionCore:    {PatternTypeBullishEngulfing, PatternTypeBearishEngulfing, PatternTypeHammer, PatternTypeShootingStar, PatternTypeMorningStar, PatternTypeEveningStar},
	PatternSelectionBullish: {PatternTypeHammer, PatternTypeInvertedHammer, PatternTypeDragonflyDoji, PatternTypeBullishMarubozu, PatternTypeBullishEngulfing, PatternTypePiercingLine, PatternTypeMorningStar},
	PatternSelectionBearish: {PatternTypeShootingStar, PatternTypeGravestoneDoji, PatternTypeBearishMarubozu, PatternTypeBearishEngulfing, PatternTypeDarkCloudCover, PatternTypeEveningStar},
	PatternSelectionReversal: {
		PatternTypeHammer, PatternTypeShootingStar, PatternTypeDragonflyDoji, PatternTypeGravestoneDoji,
		PatternTypeBullishEngulfing, PatternTypeBearishEngulfing, PatternTypePiercingLine, PatternTypeDarkCloudCover,
		PatternTypeMorningStar, PatternTypeEveningStar,
	},
	PatternSelectionTrend: {PatternTypeBullishMarubozu, PatternTypeBearishMarubozu},
}

var patternNames = map[PatternType]string{
	PatternTypeDoji: "Doji", PatternTypeHammer: "Hammer", PatternTypeInvertedHammer: "Inverted Hammer", PatternTypeShootingStar: "Shooting Star",
	PatternTypeGravestoneDoji: "Gravestone Doji", PatternTypeDragonflyDoji: "Dragonfly Doji", PatternTypeBullishMarubozu: "Bullish Marubozu", PatternTypeBearishMarubozu: "Bearish Marubozu",
	PatternTypeBullishEngulfing: "Bullish Engulfing", PatternTypeBearishEngulfing: "Bearish Engulfing", PatternTypePiercingLine: "Piercing Line", PatternTypeDarkCloudCover: "Dark Cloud Cover",
	PatternTypeMorningStar: "Morning Star", PatternTypeEveningStar: "Evening Star",
}

// DetectPatterns returns deterministic, renderer-neutral pattern results.
func DetectPatterns(data []Datum, options PatternOptions) ([]PatternResult, error) {
	if err := validatePatternOptions(options); err != nil {
		return nil, err
	}
	if options.Selection == "" && len(options.Enabled) == 0 {
		return nil, nil
	}
	for index, datum := range data {
		if !finite(datum.Open) || !finite(datum.High) || !finite(datum.Low) || !finite(datum.Close) || datum.Low > datum.Open || datum.Low > datum.Close || datum.High < datum.Open || datum.High < datum.Close {
			return nil, fmtPatternDatumError(index)
		}
	}
	var results []PatternResult
	patterns := options.Enabled
	if options.Selection != "" {
		patterns = patternSelections[options.Selection]
	}
	for index := range data {
		for _, kind := range patterns {
			if patternMatches(kind, data, index, options) {
				results = append(results, PatternResult{Index: index, Label: data[index].Label, Type: kind, Name: patternNames[kind]})
			}
		}
	}
	return results, nil
}

func fmtPatternDatumError(index int) error { return patternDatumError(index) }

type patternDatumError int

func (e patternDatumError) Error() string {
	return "candlestick chart pattern datum " + itoa(int(e)+1) + " is invalid"
}
func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := [20]byte{}
	i := len(digits)
	for value > 0 {
		i--
		digits[i] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[i:])
}

func patternMatches(kind PatternType, data []Datum, i int, options PatternOptions) bool {
	switch kind {
	case PatternTypeDoji:
		return doji(data[i], options)
	case PatternTypeHammer:
		return hammer(data[i], options)
	case PatternTypeInvertedHammer:
		return invertedHammer(data[i], options)
	case PatternTypeShootingStar:
		return shootingStar(data[i], options)
	case PatternTypeGravestoneDoji:
		return gravestone(data[i], options)
	case PatternTypeDragonflyDoji:
		return dragonfly(data[i], options)
	case PatternTypeBullishMarubozu:
		return marubozu(data[i], true, options)
	case PatternTypeBearishMarubozu:
		return marubozu(data[i], false, options)
	case PatternTypeBullishEngulfing:
		return i > 0 && engulfing(data[i-1], data[i], true, options)
	case PatternTypeBearishEngulfing:
		return i > 0 && engulfing(data[i-1], data[i], false, options)
	case PatternTypePiercingLine:
		return i > 0 && piercing(data[i-1], data[i])
	case PatternTypeDarkCloudCover:
		return i > 0 && darkCloud(data[i-1], data[i])
	case PatternTypeMorningStar:
		return i > 1 && morningStar(data[i-2], data[i-1], data[i])
	case PatternTypeEveningStar:
		return i > 1 && eveningStar(data[i-2], data[i-1], data[i])
	default:
		return false
	}
}

func body(d Datum) float64  { return math.Abs(d.Close - d.Open) }
func lower(d Datum) float64 { return math.Min(d.Open, d.Close) - d.Low }
func upper(d Datum) float64 { return d.High - math.Max(d.Open, d.Close) }
func dojiThreshold(options PatternOptions) float64 {
	if options.DojiThreshold > 0 {
		return options.DojiThreshold
	}
	return .05
}
func shadowRatio(options PatternOptions) float64 {
	if options.ShadowRatio > 0 {
		return options.ShadowRatio
	}
	return 2
}
func shadowTolerance(options PatternOptions) float64 {
	if options.ShadowTolerance > 0 {
		return options.ShadowTolerance
	}
	return .01
}
func engulfingMinSize(options PatternOptions) float64 {
	if options.EngulfingMinSize > 0 {
		return options.EngulfingMinSize
	}
	return 1
}
func doji(d Datum, options PatternOptions) bool {
	r := d.High - d.Low
	return r != 0 && body(d)/r <= dojiThreshold(options)
}
func hammer(d Datum, options PatternOptions) bool {
	return lower(d) >= shadowRatio(options)*body(d) && upper(d) <= lower(d)*.3
}
func invertedHammer(d Datum, options PatternOptions) bool {
	return upper(d) >= shadowRatio(options)*body(d) && lower(d) <= upper(d)*.3
}
func shootingStar(d Datum, options PatternOptions) bool {
	r := d.High - d.Low
	return r != 0 && invertedHammer(d, options) && (math.Min(d.Open, d.Close)-d.Low)/r <= .33
}
func gravestone(d Datum, options PatternOptions) bool {
	mid := (d.Open + d.Close) / 2
	return doji(d, options) && d.High-mid >= shadowRatio(options)*body(d) && mid-d.Low <= (d.High-mid)*.3
}
func dragonfly(d Datum, options PatternOptions) bool {
	mid := (d.Open + d.Close) / 2
	return doji(d, options) && mid-d.Low >= shadowRatio(options)*body(d) && d.High-mid <= (mid-d.Low)*.3
}
func marubozu(d Datum, bullish bool, options PatternOptions) bool {
	total := d.High - d.Low
	return total != 0 && body(d) != 0 && (upper(d)+lower(d))/total <= shadowTolerance(options) && (d.Close > d.Open) == bullish
}
func engulfing(prev, current Datum, bullish bool, options PatternOptions) bool {
	return math.Max(current.Open, current.Close) > math.Max(prev.Open, prev.Close) && math.Min(current.Open, current.Close) < math.Min(prev.Open, prev.Close) && body(current) >= engulfingMinSize(options)*body(prev) && (prev.Close < prev.Open) == bullish && (current.Close > current.Open) == bullish
}
func piercing(prev, current Datum) bool {
	return prev.Close < prev.Open && current.Close > current.Open && current.Open < prev.Close && current.Close > (prev.Open+prev.Close)/2 && current.Close < prev.Open
}
func darkCloud(prev, current Datum) bool {
	return prev.Close > prev.Open && current.Close < current.Open && current.Open > prev.Close && current.Close < (prev.Open+prev.Close)/2 && current.Close > prev.Open
}
func morningStar(first, second, third Datum) bool {
	return first.Close < first.Open && body(second) <= body(first)*.3 && second.Open < first.Close && third.Close > third.Open && third.Open > math.Max(second.Open, second.Close) && third.Close > (first.Open+first.Close)/2 && body(third) >= body(first)*.5
}
func eveningStar(first, second, third Datum) bool {
	return first.Close > first.Open && body(second) <= body(first)*.3 && second.Open > first.Close && third.Close < third.Open && third.Open < math.Min(second.Open, second.Close) && third.Close < (first.Open+first.Close)/2 && body(third) >= body(first)*.5
}
