// Package wordcloud provides the canonical interactive weighted-word-cloud API.
//
// Word Cloud-specific types and implementation live here; shared
// renderer-neutral options remain in components/chart.
package wordcloud

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

const maxDetailRows = 100

// Instance is the renderer-neutral chart instance returned by WordCloud.
type Instance = chart.Instance

// Shape selects a built-in word-cloud silhouette.
type Shape string

const (
	ShapeCircle          Shape = ""
	ShapeCardioid        Shape = "cardioid"
	ShapeDiamond         Shape = "diamond"
	ShapeSquare          Shape = "square"
	ShapeTriangleForward Shape = "triangle-forward"
	ShapeTriangle        Shape = "triangle"
	ShapePentagon        Shape = "pentagon"
	ShapeStar            Shape = "star"
)

// HorizontalPosition aligns the cloud inside its renderer surface.
type HorizontalPosition string

const (
	HorizontalDefault HorizontalPosition = ""
	HorizontalLeft    HorizontalPosition = "left"
	HorizontalCenter  HorizontalPosition = "center"
	HorizontalRight   HorizontalPosition = "right"
)

// VerticalPosition aligns the cloud inside its renderer surface.
type VerticalPosition string

const (
	VerticalDefault VerticalPosition = ""
	VerticalTop     VerticalPosition = "top"
	VerticalCenter  VerticalPosition = "center"
	VerticalBottom  VerticalPosition = "bottom"
)

// Word is one named weighted word. Class provides color-independent semantics;
// Color optionally overrides its theme-palette color.
type Word struct {
	Name  string
	Value float64
	Class string
	Color string
}

// SizeRange maps weights to font sizes in CSS pixels.
type SizeRange struct {
	Min float64
	Max float64
}

// Rotation configures the inclusive rotation range and step in degrees.
type Rotation struct {
	MinDegrees  float64
	MaxDegrees  float64
	StepDegrees float64
}

// Layout positions and sizes the cloud inside the renderer surface.
// Nil percentages preserve renderer defaults.
type Layout struct {
	Horizontal    HorizontalPosition
	Vertical      VerticalPosition
	WidthPercent  *float64
	HeightPercent *float64
}

// SeriesOptions controls one word-cloud layout without exposing its renderer.
type SeriesOptions struct {
	Shape           Shape
	SizeRange       *SizeRange
	Rotation        *Rotation
	GridSize        *int
	DrawOutOfBound  *bool
	LayoutAnimation *bool
	Layout          Layout
}

// Series is one named, typed collection of weighted words.
type Series struct {
	Name    string
	Words   []Word
	Options SeriesOptions
}

// Config describes an accessible interactive word cloud.
type Config struct {
	Label     string
	Caption   string
	Series    Series
	Width     string
	Height    string
	Options   chart.ChartOptions
	Style     charttheme.Style
	RootAttrs templ.Attributes
}

type rendererTextStyle struct {
	Color string `json:"color,omitempty"`
}

type rendererWord struct {
	Name        string             `json:"name"`
	Value       float64            `json:"value"`
	ClassName   string             `json:"className,omitempty"`
	SourceColor string             `json:"sourceColor,omitempty"`
	TextStyle   *rendererTextStyle `json:"textStyle,omitempty"`
}

// WordCloud builds a reusable interactive word-cloud component.
func WordCloud(cfg Config) Instance {
	if err := validateConfig(cfg); err != nil {
		return internalinteractive.Invalid(chartcomponents.KindInteractiveWordCloud, err)
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
	wordCloudChart := charts.NewWordCloud()
	globalOptions := []charts.GlobalOpts{
		charts.WithInitializationOpts(opts.Initialization{Width: width, Height: height}),
		charts.WithColorsOpts(opts.Colors(style.ResolvedColors())),
	}
	globalOptions = append(globalOptions, internalinteractive.ChartGlobalOptions(cfg.Options)...)
	wordCloudChart.SetGlobalOptions(globalOptions...)

	seriesOptions := make([]charts.SeriesOpts, 0, 1)
	private := opts.WordCloudChart{Shape: string(cfg.Series.Options.Shape)}
	if size := cfg.Series.Options.SizeRange; size != nil {
		private.SizeRange = []float32{float32(size.Min), float32(size.Max)}
	}
	if rotation := cfg.Series.Options.Rotation; rotation != nil {
		private.RotationRange = []float32{float32(rotation.MinDegrees), float32(rotation.MaxDegrees)}
	}
	seriesOptions = append(seriesOptions, charts.WithWorldCloudChartOpts(private))
	wordCloudChart.AddSeries(cfg.Series.Name, make([]opts.WordCloudData, len(cfg.Series.Words)), seriesOptions...)

	renderedWords := make([]rendererWord, len(cfg.Series.Words))
	for index, word := range cfg.Series.Words {
		renderedWords[index] = rendererWord{Name: word.Name, Value: word.Value, ClassName: word.Class, SourceColor: word.Color}
		if word.Color != "" {
			renderedWords[index].TextStyle = &rendererTextStyle{Color: word.Color}
		}
	}
	wordCloudChart.MultiSeries[0].Data = renderedWords
	// Remove the private adapter's random color callback. Shared theme runtime
	// assigns deterministic palette, semantic-class, or caller colors instead.
	wordCloudChart.MultiSeries[0].TextStyle = &opts.TextStyle{Color: style.ResolvedColors()[0]}

	return internalinteractive.New(chartcomponents.KindInteractiveWordCloud, internalinteractive.RenderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: wordCloudChart, Style: style, ResponsiveWidth: internalinteractive.ResponsiveWidth(cfg.Width),
		Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		RootAttrs: cfg.RootAttrs, Details: wordCloudExactValues(detailRows(cfg.Series.Words, maxDetailRows)),
		ScriptReplacements: scriptReplacements(cfg.Series.Options),
	})
}

func scriptReplacements(options SeriesOptions) []internalinteractive.ScriptReplacement {
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
		fields = append(fields, fmt.Sprintf(`"width":%q`, internalinteractive.Percentage(*options.Layout.WidthPercent)))
	}
	if options.Layout.HeightPercent != nil {
		fields = append(fields, fmt.Sprintf(`"height":%q`, internalinteractive.Percentage(*options.Layout.HeightPercent)))
	}
	if len(fields) == 0 {
		return nil
	}
	return []internalinteractive.ScriptReplacement{{Old: `"type":"wordCloud"`, New: `"type":"wordCloud",` + strings.Join(fields, ",")}}
}

func validateConfig(cfg Config) error {
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
		if !internalinteractive.FiniteNumber(word.Value) {
			return fmt.Errorf("word cloud chart word %q value must be finite", name)
		}
		if word.Value < 0 {
			return fmt.Errorf("word cloud chart word %q value must be nonnegative", name)
		}
	}
	validShapes := map[Shape]bool{
		ShapeCircle: true, ShapeCardioid: true, ShapeDiamond: true,
		ShapeSquare: true, ShapeTriangleForward: true, ShapeTriangle: true,
		ShapePentagon: true, ShapeStar: true,
	}
	if !validShapes[cfg.Series.Options.Shape] {
		return fmt.Errorf("word cloud chart shape %q is not supported", cfg.Series.Options.Shape)
	}
	if size := cfg.Series.Options.SizeRange; size != nil {
		if !internalinteractive.FiniteNumber(size.Min) || !internalinteractive.FiniteNumber(size.Max) || size.Min <= 0 || size.Max <= 0 {
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
		if !internalinteractive.FiniteNumber(rotation.MinDegrees) || !internalinteractive.FiniteNumber(rotation.MaxDegrees) || !internalinteractive.FiniteNumber(rotation.StepDegrees) {
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
	validHorizontal := map[HorizontalPosition]bool{HorizontalDefault: true, HorizontalLeft: true, HorizontalCenter: true, HorizontalRight: true}
	if !validHorizontal[cfg.Series.Options.Layout.Horizontal] {
		return fmt.Errorf("word cloud chart horizontal layout %q is not supported", cfg.Series.Options.Layout.Horizontal)
	}
	validVertical := map[VerticalPosition]bool{VerticalDefault: true, VerticalTop: true, VerticalCenter: true, VerticalBottom: true}
	if !validVertical[cfg.Series.Options.Layout.Vertical] {
		return fmt.Errorf("word cloud chart vertical layout %q is not supported", cfg.Series.Options.Layout.Vertical)
	}
	for name, value := range map[string]*float64{"width": cfg.Series.Options.Layout.WidthPercent, "height": cfg.Series.Options.Layout.HeightPercent} {
		if value != nil && (!internalinteractive.ValidPercentage(*value) || *value < 1) {
			return fmt.Errorf("word cloud chart layout %s percentage must be between 1 and 100", name)
		}
	}
	return internalinteractive.ValidateChartOptions(cfg.Options)
}

type valueRow struct {
	Word, Value, Class string
}

type valueRows struct {
	Rows    []valueRow
	Omitted int
}

func detailRows(words []Word, limit int) valueRows {
	rows := make([]valueRow, 0, min(len(words), limit))
	for _, word := range words {
		if len(rows) == limit {
			return valueRows{Rows: rows, Omitted: len(words) - len(rows)}
		}
		rows = append(rows, valueRow{Word: word.Name, Value: fmt.Sprintf("%g", word.Value), Class: word.Class})
	}
	return valueRows{Rows: rows}
}
