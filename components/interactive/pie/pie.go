// Package pie provides the canonical interactive pie, donut, and rose API.
//
// Pie-specific types and implementation live here. Variants remain options on
// one component. Shared renderer-neutral options remain in components/chart,
// while components/interactive preserves its legacy aliases and constructor.
package pie

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	sharedchart "github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

// Instance is the renderer-neutral chart instance returned by Pie.
type Instance = sharedchart.Instance

// RoseMode controls Nightingale (rose) rendering.
type RoseMode string

const (
	// RoseNone renders sectors with one shared radius.
	RoseNone RoseMode = ""
	// RoseRadius maps values to sector radius while retaining proportional angles.
	RoseRadius RoseMode = "radius"
	// RoseArea gives sectors equal angles and maps values to sector area.
	RoseArea RoseMode = "area"
)

// LabelContent controls typed sector-label content.
type LabelContent string

const (
	// LabelDefault preserves the renderer's default label content.
	LabelDefault LabelContent = ""
	// LabelNameAndValue shows each sector name and exact value.
	LabelNameAndValue LabelContent = "name-value"
)

// TooltipContent controls typed item-tooltip content.
type TooltipContent string

const (
	// TooltipDefault preserves the renderer's default item tooltip.
	TooltipDefault TooltipContent = ""
	// TooltipNameAndShare shows the sector name and its percentage share.
	TooltipNameAndShare TooltipContent = "name-share"
)

// AutoEmphasisOptions cycles emphasis across one series. The private
// runtime stops the cycle for reduced-motion users.
type AutoEmphasisOptions struct {
	SeriesIndex          int
	IntervalMilliseconds int
	ShowTooltip          *bool
}

// Center places one pie series as percentages of chart width and height.
// Nil Center preserves the chart's centered default.
type Center struct {
	X float64
	Y float64
}

// Config describes an accessible, browser-rendered pie chart.
//
// Values must be application-owned because the browser renderer serializes them.
type Config struct {
	Label          string
	Caption        string
	Series         []Series
	Width          string
	Height         string
	Options        sharedchart.ChartOptions
	SeriesOptions  sharedchart.SeriesOptions
	Style          charttheme.Style
	TooltipContent TooltipContent
	AutoEmphasis   *AutoEmphasisOptions
}

// Series describes one named pie or donut series. InnerRadius and
// OuterRadius are percentages in the inclusive range 0..100. OuterRadius
// defaults to 75 when zero.
type Series struct {
	Name         string
	Data         []Data
	InnerRadius  float64
	OuterRadius  float64
	Center       *Center
	RoseMode     RoseMode
	LabelContent LabelContent
	PadAngle     float64
	// Selectable allows readers to toggle one or more sectors.
	Selectable bool
	Options    sharedchart.SeriesOptions
}

// Data describes one nonnegative named sector. ItemStyle, Label, and
// Tooltip provide typed per-sector customization.
type Data struct {
	Name      string
	Value     float64
	ItemStyle *sharedchart.ItemStyle
	Label     *sharedchart.LabelOptions
	Tooltip   *sharedchart.TooltipOptions
	Selected  bool
}

// Pie builds a reusable interactive pie component.
func Pie(cfg Config) Instance {
	if err := validateConfig(cfg); err != nil {
		return internalinteractive.Invalid(chartcomponents.KindInteractivePie, err)
	}

	chart := charts.NewPie()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, internalinteractive.ChartGlobalOptions(cfg.Options)...)
	if formatter := pieTooltipFormatter(cfg.TooltipContent); formatter != "" {
		tooltip := opts.Tooltip{}
		if cfg.Options.Tooltip != nil {
			tooltip = internalinteractive.RendererTooltip(cfg.Options.Tooltip)
		}
		tooltip.Formatter = types.FuncStr(formatter)
		globalOptions = append(globalOptions, charts.WithTooltipOpts(tooltip))
	}
	// Explicit component colors remain authoritative over escape-hatch options.
	if len(cfg.Style.Colors) > 0 {
		globalOptions = append(globalOptions, charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors())))
	}
	if cfg.Width != "" || cfg.Height != "" {
		globalOptions = append([]charts.GlobalOpts{
			charts.WithInitializationOpts(opts.Initialization{Width: cfg.Width, Height: cfg.Height}),
		}, globalOptions...)
	}
	chart.SetGlobalOptions(globalOptions...)

	for _, series := range cfg.Series {
		outerRadius := series.OuterRadius
		if outerRadius == 0 {
			outerRadius = 75
		}
		options := make([]charts.SeriesOpts, 0, 1+len(internalinteractive.ChartSeriesOptions(cfg.SeriesOptions))+len(internalinteractive.ChartSeriesOptions(series.Options)))
		pieOptions := opts.PieChart{
			Radius:   []string{internalinteractive.Percentage(series.InnerRadius), internalinteractive.Percentage(outerRadius)},
			RoseType: string(series.RoseMode),
			PadAngle: series.PadAngle,
		}
		if series.Center != nil {
			pieOptions.Center = []string{internalinteractive.Percentage(series.Center.X), internalinteractive.Percentage(series.Center.Y)}
		}
		options = append(options, charts.WithPieChartOpts(pieOptions))
		options = append(options, internalinteractive.MergeSeriesOptions(cfg.SeriesOptions, series.Options)...)
		if series.Selectable {
			options = append(options, func(value *charts.SingleSeries) {
				value.SelectedMode = opts.Bool(true)
			})
		}
		if formatter := pieLabelFormatter(series.LabelContent); formatter != "" {
			options = append(options, func(value *charts.SingleSeries) {
				if value.Label == nil {
					value.Label = &opts.Label{}
				}
				value.Label.Formatter = types.FuncStr(formatter)
			})
		}

		data := make([]opts.PieData, len(series.Data))
		for index, sector := range series.Data {
			data[index] = opts.PieData{Name: sector.Name, Value: sector.Value}
			if sector.Selected {
				data[index].Selected = opts.Bool(true)
			}
			if sector.ItemStyle != nil {
				style := internalinteractive.RendererItemStyle(sector.ItemStyle)
				data[index].ItemStyle = &style
			}
			if sector.Label != nil {
				label := internalinteractive.RendererLabel(sector.Label)
				data[index].Label = &label
			}
			if sector.Tooltip != nil {
				tooltip := internalinteractive.RendererTooltip(sector.Tooltip)
				data[index].Tooltip = &tooltip
			}
		}
		chart.AddSeries(series.Name, data, options...)
	}

	return internalinteractive.New(chartcomponents.KindInteractivePie, internalinteractive.RenderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export, ResponsiveWidth: internalinteractive.ResponsiveWidth(cfg.Width),
		Details: pieExactValues(pieDetailRows(cfg.Series)), PieAutoEmphasis: pieAutoEmphasisMetadata(cfg.AutoEmphasis),
	})
}

func pieLabelFormatter(content LabelContent) string {
	if content == LabelNameAndValue {
		return "{b}: {c}"
	}
	return ""
}

func pieTooltipFormatter(content TooltipContent) string {
	if content == TooltipNameAndShare {
		return "{b}: {d}%"
	}
	return ""
}

func pieAutoEmphasisMetadata(value *AutoEmphasisOptions) string {
	if value == nil {
		return ""
	}
	interval := value.IntervalMilliseconds
	if interval == 0 {
		interval = 1000
	}
	showTooltip := true
	if value.ShowTooltip != nil {
		showTooltip = *value.ShowTooltip
	}
	encoded, _ := json.Marshal(struct {
		SeriesIndex int  `json:"seriesIndex"`
		Interval    int  `json:"interval"`
		ShowTooltip bool `json:"showTooltip"`
	}{SeriesIndex: value.SeriesIndex, Interval: interval, ShowTooltip: showTooltip})
	return string(encoded)
}

type pieValueRow struct {
	Series string
	Name   string
	Value  string
	Share  string
}

func pieDetailRows(series []Series) []pieValueRow {
	rows := make([]pieValueRow, 0)
	for _, current := range series {
		total := 0.0
		for _, sector := range current.Data {
			total += sector.Value
		}
		for _, sector := range current.Data {
			share := "0%"
			if total > 0 {
				share = strconv.FormatFloat(sector.Value/total*100, 'f', 1, 64) + "%"
			}
			rows = append(rows, pieValueRow{
				Series: current.Name,
				Name:   sector.Name,
				Value:  strconv.FormatFloat(sector.Value, 'f', -1, 64),
				Share:  share,
			})
		}
	}
	return rows
}

func validateConfig(cfg Config) error {
	if err := internalinteractive.ValidateChartOptions(cfg.Options); err != nil {
		return err
	}
	if cfg.Label == "" {
		return fmt.Errorf("pie chart label is required")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("pie chart series is required")
	}
	if cfg.TooltipContent != TooltipDefault && cfg.TooltipContent != TooltipNameAndShare {
		return fmt.Errorf("pie chart tooltip content %q is not supported", cfg.TooltipContent)
	}
	if cfg.AutoEmphasis != nil {
		if cfg.AutoEmphasis.SeriesIndex < 0 || cfg.AutoEmphasis.SeriesIndex >= len(cfg.Series) {
			return fmt.Errorf("pie chart auto emphasis series index must identify a configured series")
		}
		if cfg.AutoEmphasis.IntervalMilliseconds < 0 {
			return fmt.Errorf("pie chart auto emphasis interval must be nonnegative")
		}
	}
	if err := validateSeriesOptions(cfg.SeriesOptions); err != nil {
		return err
	}
	for seriesIndex, series := range cfg.Series {
		if series.Name == "" {
			return fmt.Errorf("pie chart series %d name is required", seriesIndex)
		}
		if len(series.Data) == 0 {
			return fmt.Errorf("pie chart series %q data is required", series.Name)
		}
		if !internalinteractive.ValidPercentage(series.InnerRadius) {
			return fmt.Errorf("pie chart series %q inner radius must be between 0 and 100", series.Name)
		}
		if !internalinteractive.ValidPercentage(series.OuterRadius) {
			return fmt.Errorf("pie chart series %q outer radius must be between 0 and 100", series.Name)
		}
		outerRadius := series.OuterRadius
		if outerRadius == 0 {
			outerRadius = 75
		}
		if series.InnerRadius >= outerRadius {
			return fmt.Errorf("pie chart series %q inner radius must be less than outer radius", series.Name)
		}
		if series.RoseMode != RoseNone && series.RoseMode != RoseRadius && series.RoseMode != RoseArea {
			return fmt.Errorf("pie chart series %q rose mode %q is not supported", series.Name, series.RoseMode)
		}
		if series.Center != nil {
			if !internalinteractive.ValidPercentage(series.Center.X) {
				return fmt.Errorf("pie chart series %q center x must be between 0 and 100", series.Name)
			}
			if !internalinteractive.ValidPercentage(series.Center.Y) {
				return fmt.Errorf("pie chart series %q center y must be between 0 and 100", series.Name)
			}
		}
		if series.LabelContent != LabelDefault && series.LabelContent != LabelNameAndValue {
			return fmt.Errorf("pie chart series %q label content %q is not supported", series.Name, series.LabelContent)
		}
		if math.IsNaN(series.PadAngle) || math.IsInf(series.PadAngle, 0) || series.PadAngle < 0 {
			return fmt.Errorf("pie chart series %q pad angle must be a finite nonnegative value", series.Name)
		}
		if err := validateSeriesOptions(series.Options); err != nil {
			return err
		}
		for dataIndex, sector := range series.Data {
			if sector.Name == "" {
				return fmt.Errorf("pie chart series %q data point %d name is required", series.Name, dataIndex)
			}
			if math.IsNaN(sector.Value) || math.IsInf(sector.Value, 0) || sector.Value < 0 {
				return fmt.Errorf("pie chart series %q data point %q value must be a finite nonnegative value", series.Name, sector.Name)
			}
			if err := validatePieItemStyle(sector.ItemStyle); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateSeriesOptions(options sharedchart.SeriesOptions) error {
	if err := validatePieItemStyle(options.ItemStyle); err != nil {
		return err
	}
	if options.Emphasis != nil {
		return validatePieItemStyle(options.Emphasis.ItemStyle)
	}
	return nil
}

func validatePieItemStyle(style *sharedchart.ItemStyle) error {
	if style != nil && style.ShadowBlur < 0 {
		return fmt.Errorf("pie chart item shadow blur must be nonnegative")
	}
	return nil
}
