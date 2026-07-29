package interactive

import (
	"fmt"
	"math"
	"strconv"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

// PieRoseMode controls Nightingale (rose) rendering.
type PieRoseMode string

const (
	// PieRoseNone renders sectors with one shared radius.
	PieRoseNone PieRoseMode = ""
	// PieRoseRadius maps values to sector radius while retaining proportional angles.
	PieRoseRadius PieRoseMode = "radius"
	// PieRoseArea gives sectors equal angles and maps values to sector area.
	PieRoseArea PieRoseMode = "area"
)

// PieLabelContent controls typed sector-label content.
type PieLabelContent string

const (
	// PieLabelDefault preserves the renderer's default label content.
	PieLabelDefault PieLabelContent = ""
	// PieLabelNameAndValue shows each sector name and exact value.
	PieLabelNameAndValue PieLabelContent = "name-value"
)

// PieCenter places one pie series as percentages of chart width and height.
// Nil Center preserves the chart's centered default.
type PieCenter struct {
	X float64
	Y float64
}

// PieConfig describes an accessible, browser-rendered pie chart.
//
// Values must be application-owned because the browser renderer serializes them.
type PieConfig struct {
	Label         string
	Caption       string
	Series        []PieSeries
	Width         string
	Height        string
	Options       ChartOptions
	SeriesOptions SeriesOptions
	Style         charttheme.Style
}

// PieSeries describes one named pie or donut series. InnerRadius and
// OuterRadius are percentages in the inclusive range 0..100. OuterRadius
// defaults to 75 when zero.
type PieSeries struct {
	Name         string
	Data         []PieData
	InnerRadius  float64
	OuterRadius  float64
	Center       *PieCenter
	RoseMode     PieRoseMode
	LabelContent PieLabelContent
	PadAngle     float64
	Options      SeriesOptions
}

// PieData describes one nonnegative named sector. ItemStyle, Label, and
// Tooltip provide typed per-sector customization.
type PieData struct {
	Name      string
	Value     float64
	ItemStyle *ItemStyle
	Label     *LabelOptions
	Tooltip   *TooltipOptions
}

// Pie builds a reusable interactive pie component.
func Pie(cfg PieConfig) Instance {
	if err := validatePieConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindInteractivePie, err)
	}

	chart := charts.NewPie()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, chartGlobalOptions(cfg.Options)...)
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
		options := make([]charts.SeriesOpts, 0, 1+len(chartSeriesOptions(cfg.SeriesOptions))+len(chartSeriesOptions(series.Options)))
		pieOptions := opts.PieChart{
			Radius:   []string{percentage(series.InnerRadius), percentage(outerRadius)},
			RoseType: string(series.RoseMode),
			PadAngle: series.PadAngle,
		}
		if series.Center != nil {
			pieOptions.Center = []string{percentage(series.Center.X), percentage(series.Center.Y)}
		}
		options = append(options, charts.WithPieChartOpts(pieOptions))
		options = append(options, mergeSeriesOptions(cfg.SeriesOptions, series.Options)...)
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
			if sector.ItemStyle != nil {
				style := rendererItemStyle(sector.ItemStyle)
				data[index].ItemStyle = &style
			}
			if sector.Label != nil {
				label := rendererLabel(sector.Label)
				data[index].Label = &label
			}
			if sector.Tooltip != nil {
				tooltip := rendererTooltip(sector.Tooltip)
				data[index].Tooltip = &tooltip
			}
		}
		chart.AddSeries(series.Name, data, options...)
	}

	return newInstance(chartcomponents.KindInteractivePie, renderConfig{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style, Animation: cfg.Options.Animation, Controls: cfg.Options.Controls, Export: cfg.Options.Export, ResponsiveWidth: responsiveWidth(cfg.Width),
		Details: pieExactValues(pieDetailRows(cfg.Series)),
	})
}

func pieLabelFormatter(content PieLabelContent) string {
	if content == PieLabelNameAndValue {
		return "{b}: {c}"
	}
	return ""
}

type pieValueRow struct {
	Series string
	Name   string
	Value  string
	Share  string
}

func pieDetailRows(series []PieSeries) []pieValueRow {
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

func percentage(value float64) string { return fmt.Sprintf("%g%%", value) }

func validatePieConfig(cfg PieConfig) error {
	if cfg.Label == "" {
		return fmt.Errorf("pie chart label is required")
	}
	if len(cfg.Series) == 0 {
		return fmt.Errorf("pie chart series is required")
	}
	for seriesIndex, series := range cfg.Series {
		if series.Name == "" {
			return fmt.Errorf("pie chart series %d name is required", seriesIndex)
		}
		if len(series.Data) == 0 {
			return fmt.Errorf("pie chart series %q data is required", series.Name)
		}
		if !validPercentage(series.InnerRadius) {
			return fmt.Errorf("pie chart series %q inner radius must be between 0 and 100", series.Name)
		}
		if !validPercentage(series.OuterRadius) {
			return fmt.Errorf("pie chart series %q outer radius must be between 0 and 100", series.Name)
		}
		outerRadius := series.OuterRadius
		if outerRadius == 0 {
			outerRadius = 75
		}
		if series.InnerRadius >= outerRadius {
			return fmt.Errorf("pie chart series %q inner radius must be less than outer radius", series.Name)
		}
		if series.RoseMode != PieRoseNone && series.RoseMode != PieRoseRadius && series.RoseMode != PieRoseArea {
			return fmt.Errorf("pie chart series %q rose mode %q is not supported", series.Name, series.RoseMode)
		}
		if series.Center != nil {
			if !validPercentage(series.Center.X) {
				return fmt.Errorf("pie chart series %q center x must be between 0 and 100", series.Name)
			}
			if !validPercentage(series.Center.Y) {
				return fmt.Errorf("pie chart series %q center y must be between 0 and 100", series.Name)
			}
		}
		if series.LabelContent != PieLabelDefault && series.LabelContent != PieLabelNameAndValue {
			return fmt.Errorf("pie chart series %q label content %q is not supported", series.Name, series.LabelContent)
		}
		if math.IsNaN(series.PadAngle) || math.IsInf(series.PadAngle, 0) || series.PadAngle < 0 {
			return fmt.Errorf("pie chart series %q pad angle must be a finite nonnegative value", series.Name)
		}
		for dataIndex, sector := range series.Data {
			if sector.Name == "" {
				return fmt.Errorf("pie chart series %q data point %d name is required", series.Name, dataIndex)
			}
			if math.IsNaN(sector.Value) || math.IsInf(sector.Value, 0) || sector.Value < 0 {
				return fmt.Errorf("pie chart series %q data point %q value must be a finite nonnegative value", series.Name, sector.Name)
			}
		}
	}
	return nil
}

func validPercentage(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 100
}
