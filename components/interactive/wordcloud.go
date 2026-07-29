package interactive

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

const maxWordCloudDetailRows = 100

// WordCloudShape selects a built-in word-cloud silhouette.
type WordCloudShape string

const (
	WordCloudShapeCircle          WordCloudShape = ""
	WordCloudShapeCardioid        WordCloudShape = "cardioid"
	WordCloudShapeDiamond         WordCloudShape = "diamond"
	WordCloudShapeSquare          WordCloudShape = "square"
	WordCloudShapeTriangleForward WordCloudShape = "triangle-forward"
	WordCloudShapeTriangle        WordCloudShape = "triangle"
	WordCloudShapePentagon        WordCloudShape = "pentagon"
	WordCloudShapeStar            WordCloudShape = "star"
)

// WordCloudHorizontalPosition aligns the cloud inside its renderer surface.
type WordCloudHorizontalPosition string

const (
	WordCloudHorizontalDefault WordCloudHorizontalPosition = ""
	WordCloudHorizontalLeft    WordCloudHorizontalPosition = "left"
	WordCloudHorizontalCenter  WordCloudHorizontalPosition = "center"
	WordCloudHorizontalRight   WordCloudHorizontalPosition = "right"
)

// WordCloudVerticalPosition aligns the cloud inside its renderer surface.
type WordCloudVerticalPosition string

const (
	WordCloudVerticalDefault WordCloudVerticalPosition = ""
	WordCloudVerticalTop     WordCloudVerticalPosition = "top"
	WordCloudVerticalCenter  WordCloudVerticalPosition = "center"
	WordCloudVerticalBottom  WordCloudVerticalPosition = "bottom"
)

// Word is one named weighted word. Class provides color-independent semantics;
// Color optionally overrides its theme-palette color.
type Word struct {
	Name  string
	Value float64
	Class string
	Color string
}

// WordCloudSizeRange maps weights to font sizes in CSS pixels.
type WordCloudSizeRange struct {
	Min float64
	Max float64
}

// WordCloudRotation configures the inclusive rotation range and step in degrees.
type WordCloudRotation struct {
	MinDegrees  float64
	MaxDegrees  float64
	StepDegrees float64
}

// WordCloudLayout positions and sizes the cloud inside the renderer surface.
// Nil percentages preserve renderer defaults.
type WordCloudLayout struct {
	Horizontal    WordCloudHorizontalPosition
	Vertical      WordCloudVerticalPosition
	WidthPercent  *float64
	HeightPercent *float64
}

// WordCloudSeriesOptions controls one word-cloud layout without exposing its renderer.
type WordCloudSeriesOptions struct {
	Shape           WordCloudShape
	SizeRange       *WordCloudSizeRange
	Rotation        *WordCloudRotation
	GridSize        *int
	DrawOutOfBound  *bool
	LayoutAnimation *bool
	Layout          WordCloudLayout
}

// WordCloudSeries is one named, typed collection of weighted words.
type WordCloudSeries struct {
	Name    string
	Words   []Word
	Options WordCloudSeriesOptions
}

// WordCloudConfig describes an accessible interactive word cloud.
type WordCloudConfig struct {
	Label     string
	Caption   string
	Series    WordCloudSeries
	Width     string
	Height    string
	Options   ChartOptions
	Style     charttheme.Style
	RootAttrs templ.Attributes
}

type rendererWordCloudTextStyle struct {
	Color string `json:"color,omitempty"`
}

type rendererWord struct {
	Name        string                      `json:"name"`
	Value       float64                     `json:"value"`
	ClassName   string                      `json:"className,omitempty"`
	SourceColor string                      `json:"sourceColor,omitempty"`
	TextStyle   *rendererWordCloudTextStyle `json:"textStyle,omitempty"`
}

// WordCloud builds a reusable interactive word-cloud component.
func WordCloud(cfg WordCloudConfig) Instance {
	if err := validateWordCloudConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveWordCloud, err)
	}

	style := cfg.Style
	style.Class = strings.TrimSpace("goshtoso-charts-word-cloud " + style.Class)
	width, height := cfg.Width, cfg.Height
	if width == "" {
		width = "100%"
	}
	if height == "" {
		height = "500px"
	}
	chart := charts.NewWordCloud()
	globalOptions := []charts.GlobalOpts{
		charts.WithInitializationOpts(opts.Initialization{Width: width, Height: height}),
		charts.WithColorsOpts(opts.Colors(style.ResolvedColors())),
	}
	globalOptions = append(globalOptions, chartGlobalOptions(cfg.Options)...)
	chart.SetGlobalOptions(globalOptions...)

	seriesOptions := make([]charts.SeriesOpts, 0, 1)
	private := opts.WordCloudChart{Shape: string(cfg.Series.Options.Shape)}
	if size := cfg.Series.Options.SizeRange; size != nil {
		private.SizeRange = []float32{float32(size.Min), float32(size.Max)}
	}
	if rotation := cfg.Series.Options.Rotation; rotation != nil {
		private.RotationRange = []float32{float32(rotation.MinDegrees), float32(rotation.MaxDegrees)}
	}
	seriesOptions = append(seriesOptions, charts.WithWorldCloudChartOpts(private))
	chart.AddSeries(cfg.Series.Name, make([]opts.WordCloudData, len(cfg.Series.Words)), seriesOptions...)

	renderedWords := make([]rendererWord, len(cfg.Series.Words))
	for index, word := range cfg.Series.Words {
		renderedWords[index] = rendererWord{Name: word.Name, Value: word.Value, ClassName: word.Class, SourceColor: word.Color}
		if word.Color != "" {
			renderedWords[index].TextStyle = &rendererWordCloudTextStyle{Color: word.Color}
		}
	}
	chart.MultiSeries[0].Data = renderedWords
	// Remove the private adapter's random color callback. Shared theme runtime
	// assigns deterministic palette, semantic-class, or caller colors instead.
	chart.MultiSeries[0].TextStyle = &opts.TextStyle{Color: style.ResolvedColors()[0]}

	return newInstance(chartcomponents.KindInteractiveWordCloud, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: style, ResponsiveWidth: responsiveWidth(cfg.Width),
		Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		RootAttrs: cfg.RootAttrs, Details: wordCloudExactValues(wordCloudDetailRows(cfg.Series.Words, maxWordCloudDetailRows)),
		ScriptReplacements: wordCloudScriptReplacements(cfg.Series.Options),
	})
}

func wordCloudScriptReplacements(options WordCloudSeriesOptions) []scriptReplacement {
	fields := make([]string, 0, 8)
	if rotation := options.Rotation; rotation != nil {
		fields = append(fields, fmt.Sprintf(`"rotationStep":%g`, rotation.StepDegrees))
	}
	if options.GridSize != nil {
		fields = append(fields, fmt.Sprintf(`"gridSize":%d`, *options.GridSize))
	}
	if options.DrawOutOfBound != nil {
		fields = append(fields, fmt.Sprintf(`"drawOutOfBound":%t`, *options.DrawOutOfBound))
	}
	if options.LayoutAnimation != nil {
		fields = append(fields, fmt.Sprintf(`"layoutAnimation":%t`, *options.LayoutAnimation))
	}
	if options.Layout.Horizontal != "" {
		fields = append(fields, fmt.Sprintf(`"left":%q`, options.Layout.Horizontal))
	}
	if options.Layout.Vertical != "" {
		fields = append(fields, fmt.Sprintf(`"top":%q`, options.Layout.Vertical))
	}
	if options.Layout.WidthPercent != nil {
		fields = append(fields, fmt.Sprintf(`"width":%q`, percentage(*options.Layout.WidthPercent)))
	}
	if options.Layout.HeightPercent != nil {
		fields = append(fields, fmt.Sprintf(`"height":%q`, percentage(*options.Layout.HeightPercent)))
	}
	if len(fields) == 0 {
		return nil
	}
	return []scriptReplacement{{Old: `"type":"wordCloud"`, New: `"type":"wordCloud",` + strings.Join(fields, ",")}}
}

func validateWordCloudConfig(cfg WordCloudConfig) error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("word cloud chart label is required")
	}
	if strings.TrimSpace(cfg.Series.Name) == "" {
		return fmt.Errorf("word cloud chart series name is required")
	}
	if len(cfg.Series.Words) == 0 {
		return fmt.Errorf("word cloud chart words are required")
	}
	if cfg.Options.Legend != nil {
		return fmt.Errorf("word cloud chart legend is not supported")
	}
	if cfg.Options.XAxis != nil || cfg.Options.YAxis != nil {
		return fmt.Errorf("word cloud chart Cartesian axes are not supported")
	}
	if tooltip := cfg.Options.Tooltip; tooltip != nil && tooltip.Trigger != "" && tooltip.Trigger != "item" {
		return fmt.Errorf("word cloud chart tooltip trigger %q is not supported", tooltip.Trigger)
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("word cloud chart root attribute %q is reserved", attribute)
			}
		}
	}
	names := make(map[string]bool, len(cfg.Series.Words))
	for index, word := range cfg.Series.Words {
		name := strings.TrimSpace(word.Name)
		if name == "" {
			return fmt.Errorf("word cloud chart word %d name is required", index)
		}
		if names[name] {
			return fmt.Errorf("word cloud chart word %q is duplicated", name)
		}
		names[name] = true
		if !finiteNumber(word.Value) {
			return fmt.Errorf("word cloud chart word %q value must be finite", name)
		}
		if word.Value < 0 {
			return fmt.Errorf("word cloud chart word %q value must be nonnegative", name)
		}
	}
	validShapes := map[WordCloudShape]bool{
		WordCloudShapeCircle: true, WordCloudShapeCardioid: true, WordCloudShapeDiamond: true,
		WordCloudShapeSquare: true, WordCloudShapeTriangleForward: true, WordCloudShapeTriangle: true,
		WordCloudShapePentagon: true, WordCloudShapeStar: true,
	}
	if !validShapes[cfg.Series.Options.Shape] {
		return fmt.Errorf("word cloud chart shape %q is not supported", cfg.Series.Options.Shape)
	}
	if size := cfg.Series.Options.SizeRange; size != nil {
		if !finiteNumber(size.Min) || !finiteNumber(size.Max) || size.Min <= 0 || size.Max <= 0 {
			return fmt.Errorf("word cloud chart size range must be finite positive values")
		}
		if size.Min > size.Max {
			return fmt.Errorf("word cloud chart size range minimum must not exceed maximum")
		}
		if size.Min > 512 || size.Max > 512 {
			return fmt.Errorf("word cloud chart size range must be between 1 and 512 pixels")
		}
	}
	if rotation := cfg.Series.Options.Rotation; rotation != nil {
		if !finiteNumber(rotation.MinDegrees) || !finiteNumber(rotation.MaxDegrees) || !finiteNumber(rotation.StepDegrees) {
			return fmt.Errorf("word cloud chart rotation values must be finite")
		}
		if rotation.MinDegrees > rotation.MaxDegrees {
			return fmt.Errorf("word cloud chart rotation minimum must not exceed maximum")
		}
		if rotation.MinDegrees < -180 || rotation.MaxDegrees > 180 {
			return fmt.Errorf("word cloud chart rotation range must stay between -180 and 180 degrees")
		}
		if rotation.StepDegrees <= 0 {
			return fmt.Errorf("word cloud chart rotation step must be positive")
		}
	}
	if grid := cfg.Series.Options.GridSize; grid != nil && (*grid < 1 || *grid > 64) {
		return fmt.Errorf("word cloud chart grid size must be between 1 and 64")
	}
	validHorizontal := map[WordCloudHorizontalPosition]bool{WordCloudHorizontalDefault: true, WordCloudHorizontalLeft: true, WordCloudHorizontalCenter: true, WordCloudHorizontalRight: true}
	if !validHorizontal[cfg.Series.Options.Layout.Horizontal] {
		return fmt.Errorf("word cloud chart horizontal layout %q is not supported", cfg.Series.Options.Layout.Horizontal)
	}
	validVertical := map[WordCloudVerticalPosition]bool{WordCloudVerticalDefault: true, WordCloudVerticalTop: true, WordCloudVerticalCenter: true, WordCloudVerticalBottom: true}
	if !validVertical[cfg.Series.Options.Layout.Vertical] {
		return fmt.Errorf("word cloud chart vertical layout %q is not supported", cfg.Series.Options.Layout.Vertical)
	}
	for name, value := range map[string]*float64{"width": cfg.Series.Options.Layout.WidthPercent, "height": cfg.Series.Options.Layout.HeightPercent} {
		if value != nil && (!finiteNumber(*value) || *value < 1 || *value > 100) {
			return fmt.Errorf("word cloud chart layout %s percentage must be between 1 and 100", name)
		}
	}
	return validateChartOptions(cfg.Options)
}

type wordCloudValueRow struct {
	Word, Value, Class string
}

type wordCloudValueRows struct {
	Rows    []wordCloudValueRow
	Omitted int
}

func wordCloudDetailRows(words []Word, limit int) wordCloudValueRows {
	rows := make([]wordCloudValueRow, 0, min(len(words), limit))
	for _, word := range words {
		if len(rows) == limit {
			return wordCloudValueRows{Rows: rows, Omitted: len(words) - len(rows)}
		}
		rows = append(rows, wordCloudValueRow{Word: word.Name, Value: fmt.Sprintf("%g", word.Value), Class: word.Class})
	}
	return wordCloudValueRows{Rows: rows}
}
