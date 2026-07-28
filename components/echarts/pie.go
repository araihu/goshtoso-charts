package echarts

import (
	"fmt"
	"math"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
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

// PieConfig describes an accessible, browser-rendered pie chart.
//
// Values must be application-owned. go-echarts serializes chart values into
// executable JavaScript.
type PieConfig struct {
	Label         string
	Caption       string
	Series        []PieSeries
	Width         string
	Height        string
	GlobalOptions []charts.GlobalOpts
	SeriesOptions []charts.SeriesOpts
	Style         charttheme.Style
}

// PieSeries describes one named pie or donut series. InnerRadius and
// OuterRadius are percentages in the inclusive range 0..100. OuterRadius
// defaults to 75 when zero.
type PieSeries struct {
	Name        string
	Data        []PieData
	InnerRadius float64
	OuterRadius float64
	RoseMode    PieRoseMode
	PadAngle    float64
	Options     []charts.SeriesOpts
}

// PieData describes one nonnegative named sector. ItemStyle, Label, and
// Tooltip provide typed per-sector customization.
type PieData struct {
	Name      string
	Value     float64
	ItemStyle *opts.ItemStyle
	Label     *opts.Label
	Tooltip   *opts.Tooltip
}

// Pie builds a reusable interactive pie component.
func Pie(cfg PieConfig) Instance {
	if err := validatePieConfig(cfg); err != nil {
		return newInvalidInstance(chartcomponents.KindEChartsPie, err)
	}

	chart := charts.NewPie()
	globalOptions := []charts.GlobalOpts{charts.WithColorsOpts(opts.Colors(cfg.Style.ResolvedColors()))}
	globalOptions = append(globalOptions, cfg.GlobalOptions...)
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
		options := make([]charts.SeriesOpts, 0, 1+len(cfg.SeriesOptions)+len(series.Options))
		options = append(options, charts.WithPieChartOpts(opts.PieChart{
			Radius:   []string{percentage(series.InnerRadius), percentage(outerRadius)},
			RoseType: string(series.RoseMode),
			PadAngle: series.PadAngle,
		}))
		options = append(options, cfg.SeriesOptions...)
		options = append(options, series.Options...)

		data := make([]opts.PieData, len(series.Data))
		for index, sector := range series.Data {
			data[index] = opts.PieData{
				Name: sector.Name, Value: sector.Value, ItemStyle: sector.ItemStyle,
				Label: sector.Label, Tooltip: sector.Tooltip,
			}
		}
		chart.AddSeries(series.Name, data, options...)
	}

	return newInstance(chartcomponents.KindEChartsPie, Config{
		Label: cfg.Label, Caption: cfg.Caption, Chart: chart, Style: cfg.Style,
	})
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
