package interactive

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

const maxThemeRiverDetailRows = 96

// ThemeRiverConfig describes an accessible temporal stream graph.
//
// Every stream must provide the same strictly increasing dates. Label names the
// figure; Caption remains visible. Stream classes and exact values remain
// available in the bounded adjacent table.
type ThemeRiverConfig struct {
	Label        string
	Caption      string
	Streams      []ThemeRiverStream
	Layout       ThemeRiverLayout
	BoundaryGap  ThemeRiverBoundaryGap
	LabelOptions *LabelOptions
	Width        string
	Height       string
	Options      ChartOptions
	Style        charttheme.Style
	RootAttrs    templ.Attributes
}

// ThemeRiverStream is one named temporal layer.
//
// Class is a renderer-neutral semantic classification retained in the exact
// value table. Color overrides the corresponding theme palette entry.
type ThemeRiverStream struct {
	Name   string
	Class  string
	Color  string
	Points []ThemeRiverPoint
}

// ThemeRiverPoint is one typed value at an instant.
type ThemeRiverPoint struct {
	Time  time.Time
	Value float64
}

// ThemeRiverLayout positions the temporal axis inside the chart.
type ThemeRiverLayout struct {
	LeftPercent   *float64
	RightPercent  *float64
	TopPercent    *float64
	BottomPercent *float64
}

// ThemeRiverBoundaryGap reserves percentage space before and after the river.
// Zero values preserve the renderer default.
type ThemeRiverBoundaryGap struct {
	StartPercent *float64
	EndPercent   *float64
}

// ThemeRiver builds a reusable interactive temporal stream graph.
func ThemeRiver(cfg ThemeRiverConfig) Instance {
	if err := validateThemeRiverConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractiveThemeRiver, err)
	}

	style := themeRiverStyle(cfg)
	chart := charts.NewThemeRiver()
	width, height := cfg.Width, cfg.Height
	if width == "" {
		width = "100%"
	}
	if height == "" {
		height = "500px"
	}
	globalOptions := []charts.GlobalOpts{
		charts.WithInitializationOpts(opts.Initialization{Width: width, Height: height}),
		charts.WithColorsOpts(opts.Colors(style.ResolvedColors())),
	}
	globalOptions = append(globalOptions, chartGlobalOptions(cfg.Options)...)
	axis := opts.SingleAxis{Type: "time"}
	if cfg.Layout.LeftPercent != nil {
		axis.Left = percentage(*cfg.Layout.LeftPercent)
	}
	if cfg.Layout.RightPercent != nil {
		axis.Right = percentage(*cfg.Layout.RightPercent)
	}
	if cfg.Layout.TopPercent != nil {
		axis.Top = percentage(*cfg.Layout.TopPercent)
	}
	if cfg.Layout.BottomPercent != nil {
		axis.Bottom = percentage(*cfg.Layout.BottomPercent)
	} else {
		axis.Bottom = "10%"
	}
	globalOptions = append(globalOptions, charts.WithSingleAxisOpts(axis))
	chart.SetGlobalOptions(globalOptions...)

	data := make([]opts.ThemeRiverData, 0, len(cfg.Streams)*len(cfg.Streams[0].Points))
	for _, stream := range cfg.Streams {
		for _, point := range stream.Points {
			data = append(data, opts.ThemeRiverData{
				Date: themeRiverTime(point.Time), Value: point.Value, Name: stream.Name,
			})
		}
	}
	seriesOptions := make([]charts.SeriesOpts, 0, 2)
	if cfg.LabelOptions != nil {
		seriesOptions = append(seriesOptions, charts.WithLabelOpts(rendererLabel(cfg.LabelOptions)))
	}
	if cfg.Options.Animation != nil {
		seriesOptions = append(seriesOptions, charts.WithAnimationOpts(opts.Animation{Animation: opts.Bool(*cfg.Options.Animation)}))
	}
	chart.AddSeries(cfg.Label, data, seriesOptions...)

	replacements := []scriptReplacement{}
	if cfg.BoundaryGap.StartPercent != nil {
		gap, _ := json.Marshal([]string{percentage(*cfg.BoundaryGap.StartPercent), percentage(*cfg.BoundaryGap.EndPercent)})
		replacements = append(replacements, scriptReplacement{
			Old: `"type":"themeRiver"`,
			New: `"type":"themeRiver","boundaryGap":` + string(gap),
		})
	}
	return newInstance(chartcomponents.KindInteractiveThemeRiver, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: style,
		Animation: cfg.Options.Animation, RootAttrs: cfg.RootAttrs,
		Controls: cfg.Options.Controls, Export: cfg.Options.Export,
		Details:            themeRiverExactValues(themeRiverDetailRows(cfg.Streams, maxThemeRiverDetailRows)),
		ScriptReplacements: replacements,
	})
}

func themeRiverStyle(cfg ThemeRiverConfig) charttheme.Style {
	style := cfg.Style
	style.Class = strings.TrimSpace("goshtoso-charts-theme-river " + style.Class)
	hasStreamColor := false
	for _, stream := range cfg.Streams {
		if strings.TrimSpace(stream.Color) != "" {
			hasStreamColor = true
			break
		}
	}
	if !hasStreamColor {
		return style
	}
	colors := style.ResolvedColors()
	for index, stream := range cfg.Streams {
		if strings.TrimSpace(stream.Color) == "" {
			continue
		}
		for len(colors) <= index {
			colors = append(colors, style.SeriesColor(len(colors)))
		}
		colors[index] = stream.Color
	}
	style.Colors = colors
	return style
}

func themeRiverTime(value time.Time) string {
	if value.Hour() == 0 && value.Minute() == 0 && value.Second() == 0 && value.Nanosecond() == 0 {
		return value.Format("2006/01/02")
	}
	return value.Format(time.RFC3339Nano)
}

func validateThemeRiverConfig(cfg ThemeRiverConfig) error {
	if strings.TrimSpace(cfg.Label) == "" {
		return fmt.Errorf("theme river chart label is required")
	}
	if len(cfg.Streams) == 0 {
		return fmt.Errorf("theme river chart streams are required")
	}
	if cfg.Options.XAxis != nil || cfg.Options.YAxis != nil {
		return fmt.Errorf("theme river chart Cartesian axes are not supported")
	}
	for name, value := range map[string]*float64{
		"layout left": cfg.Layout.LeftPercent, "layout right": cfg.Layout.RightPercent,
		"layout top": cfg.Layout.TopPercent, "layout bottom": cfg.Layout.BottomPercent,
	} {
		if value != nil && !validPercentage(*value) {
			return fmt.Errorf("theme river chart %s percentage must be between 0 and 100", name)
		}
	}
	if (cfg.BoundaryGap.StartPercent == nil) != (cfg.BoundaryGap.EndPercent == nil) {
		return fmt.Errorf("theme river chart boundary gap requires both start and end percentages")
	}
	if cfg.BoundaryGap.StartPercent != nil &&
		(!validPercentage(*cfg.BoundaryGap.StartPercent) || !validPercentage(*cfg.BoundaryGap.EndPercent)) {
		return fmt.Errorf("theme river chart boundary gap must be between 0 and 100")
	}
	if cfg.LabelOptions != nil && cfg.LabelOptions.FontSize < 0 {
		return fmt.Errorf("theme river chart label font size must be nonnegative")
	}
	for attribute := range cfg.RootAttrs {
		for _, reserved := range []string{"class", "role", "aria-label"} {
			if strings.EqualFold(attribute, reserved) {
				return fmt.Errorf("theme river chart root attribute %q is reserved", attribute)
			}
		}
	}
	names := make(map[string]bool, len(cfg.Streams))
	var aligned []time.Time
	for streamIndex, stream := range cfg.Streams {
		name := strings.TrimSpace(stream.Name)
		if name == "" {
			return fmt.Errorf("theme river chart stream %d name is required", streamIndex)
		}
		if names[name] {
			return fmt.Errorf("theme river chart stream %q is duplicated", name)
		}
		names[name] = true
		if len(stream.Points) == 0 {
			return fmt.Errorf("theme river chart stream %q points are required", name)
		}
		if streamIndex == 0 {
			aligned = make([]time.Time, len(stream.Points))
		} else if len(stream.Points) != len(aligned) {
			return fmt.Errorf("theme river chart stream %q has %d points for %d aligned dates", name, len(stream.Points), len(aligned))
		}
		for pointIndex, point := range stream.Points {
			if point.Time.IsZero() {
				return fmt.Errorf("theme river chart stream %q point %d time is required", name, pointIndex)
			}
			if !finiteNumber(point.Value) {
				return fmt.Errorf("theme river chart stream %q point %d value must be finite", name, pointIndex)
			}
			if point.Value < 0 {
				return fmt.Errorf("theme river chart stream %q point %d value must be nonnegative", name, pointIndex)
			}
			if pointIndex > 0 && !point.Time.After(stream.Points[pointIndex-1].Time) {
				return fmt.Errorf("theme river chart stream %q dates must be strictly increasing", name)
			}
			if streamIndex == 0 {
				aligned[pointIndex] = point.Time
			} else if !point.Time.Equal(aligned[pointIndex]) {
				return fmt.Errorf("theme river chart stream %q point %d date is not aligned", name, pointIndex)
			}
		}
	}
	return validateChartOptions(cfg.Options)
}

type themeRiverValueRow struct {
	Date, Stream, Class, Value string
}

type themeRiverValueRows struct {
	Rows    []themeRiverValueRow
	Omitted int
}

func themeRiverDetailRows(streams []ThemeRiverStream, limit int) themeRiverValueRows {
	total := 0
	for _, stream := range streams {
		total += len(stream.Points)
	}
	rows := make([]themeRiverValueRow, 0, min(total, limit))
	for _, stream := range streams {
		for _, point := range stream.Points {
			if len(rows) == limit {
				return themeRiverValueRows{Rows: rows, Omitted: total - len(rows)}
			}
			rows = append(rows, themeRiverValueRow{
				Date: themeRiverTime(point.Time), Stream: stream.Name,
				Class: stream.Class, Value: fmt.Sprintf("%g", point.Value),
			})
		}
	}
	return themeRiverValueRows{Rows: rows}
}
