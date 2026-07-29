package pages

import (
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/araihu/goshtoso-charts/components/interactive"
)

const (
	interactivePieUpstreamPath     = "examples/pie.go"
	interactivePieUpstreamRevision = "bda428480a82d6d77ebb9fa939cf8d52528453dd"
	interactivePieUpstreamSHA256   = "a59bb6f11818d4175d033f025f00a58e6a191eff5acf30f0e0cd5f98cd493ada"

	pieCoverageExample = "example"
)

type interactivePieCoverageEntry struct {
	Name      string
	Status    string
	Treatment string
}

func interactivePieUpstreamCoverage() []interactivePieCoverageEntry {
	return []interactivePieCoverageEntry{
		{Name: "pieBase", Status: pieCoverageExample, Treatment: "basic seasonal distribution"},
		{Name: "pieShowLabel", Status: pieCoverageExample, Treatment: "visible name and value labels"},
		{Name: "pieRadius", Status: pieCoverageExample, Treatment: "donut radii"},
		{Name: "pieRadiusWithPadAngle", Status: pieCoverageExample, Treatment: "sector padding, offset center, vertical legend, hidden labels, and share tooltip"},
		{Name: "pieRoseArea", Status: pieCoverageExample, Treatment: "area rose"},
		{Name: "pieRoseRadius", Status: pieCoverageExample, Treatment: "radius rose"},
		{Name: "pieRoseAreaRadius", Status: pieCoverageExample, Treatment: "side-by-side area and radius roses"},
		{Name: "pieInPie", Status: pieCoverageExample, Treatment: "nested non-overlapping roses"},
		{Name: "pieWithDispatchAction", Status: pieCoverageExample, Treatment: "typed rotating emphasis and item tooltip"},
	}
}

type interactivePieSource struct {
	Path   string
	SHA256 string
	Scope  string
}

func interactivePieSupplementarySources() []interactivePieSource {
	return []interactivePieSource{
		{Path: "examples/page_center_layout.go", SHA256: "106456904719dfacfb13adcc1b9e66df83cf28a5a801539bad4d1958554166c9", Scope: "centered page-layout reference"},
		{Path: "examples/page_flex_layout.go", SHA256: "3113b7bdf78a2365ae62502fe86ab001f3ff3034b1d77752c693e95b28a0fd68", Scope: "flex page-layout reference"},
		{Path: "examples/page_none_layout.go", SHA256: "ce38424de2ffeb919661e536c7f44921de098ae14643d4f2975d8e72296c32f8", Scope: "unmanaged page-layout reference"},
	}
}

var interactivePieSeasons = []string{"Spring", "Summer", "Autumn", "Winter"}

// interactivePieData replaces upstream ambient randomness with fixed values
// from a local seed-1 sequence. Names, [0,100) domain, series call order, and
// chart geometry remain unchanged. The upstream "Autumn " typo is trimmed.
func interactivePieData(values [4]float64) []interactive.PieData {
	data := make([]interactive.PieData, len(interactivePieSeasons))
	for index, name := range interactivePieSeasons {
		data[index] = interactive.PieData{Name: name, Value: values[index]}
	}
	return data
}

func interactivePieOptions(title, filename string) interactive.ChartOptions {
	options := controlledOptions(title, filename)
	options.Legend = &interactive.LegendOptions{Left: "center", Bottom: "0"}
	return options
}

func sampleInteractivePie() interactive.PieConfig {
	return interactive.PieConfig{
		Label: "Basic seasonal pie", Caption: "Four seasonal values shown as shares of one total.",
		Series: []interactive.PieSeries{{
			Name: "Seasons", Data: interactivePieData([4]float64{81, 87, 47, 59}),
		}},
		Width: "100%", Height: "420px",
		Options: interactivePieOptions("Basic pie example", "basic-seasonal-pie"),
		Style:   charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractivePieLabels() interactive.PieConfig {
	return interactive.PieConfig{
		Label: "Seasonal values with labels", Caption: "Each sector names its season and exact value; the adjacent table also gives shares.",
		Series: []interactive.PieSeries{{
			Name: "Seasons", LabelContent: interactive.PieLabelNameAndValue,
			Data:    interactivePieData([4]float64{81, 18, 25, 40}),
			Options: interactive.SeriesOptions{Label: &interactive.LabelOptions{Show: interactive.Bool(true)}},
		}},
		Width: "100%", Height: "420px", Options: interactivePieOptions("label options", "seasonal-pie-labels"),
		Style: charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractivePieRadius() interactive.PieConfig {
	return interactive.PieConfig{
		Label: "Seasonal donut", Caption: "A 40% inner radius opens the center while preserving proportional sector angles.",
		Series: []interactive.PieSeries{{
			Name: "Seasons", InnerRadius: 40, OuterRadius: 75, LabelContent: interactive.PieLabelNameAndValue,
			Data:    interactivePieData([4]float64{56, 0, 94, 11}),
			Options: interactive.SeriesOptions{Label: &interactive.LabelOptions{Show: interactive.Bool(true)}},
		}},
		Width: "100%", Height: "420px", Options: interactivePieOptions("radius options", "seasonal-donut"),
		Style: charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractivePiePadded() interactive.PieConfig {
	options := interactivePieOptions("radius and pad-angle options", "padded-seasonal-donut")
	options.Legend = &interactive.LegendOptions{
		Orient: "vertical", Right: "20%", Top: "center",
		Padding: &interactive.EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
	}
	options.Tooltip = &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "item"}
	return interactive.PieConfig{
		Label: "Offset padded donut", Caption: "Five-degree sector gaps, an offset center, and a right-side legend preserve the upstream composition.",
		Series: []interactive.PieSeries{{
			Name: "Seasons", InnerRadius: 40, OuterRadius: 75, Center: &interactive.PieCenter{X: 40, Y: 50}, PadAngle: 5,
			Data:    interactivePieData([4]float64{62, 89, 28, 74}),
			Options: interactive.SeriesOptions{Label: &interactive.LabelOptions{Show: interactive.Bool(false)}},
		}},
		Width: "100%", Height: "420px", Options: options, TooltipContent: interactive.PieTooltipNameAndShare,
		Style: charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractivePieRoseArea() interactive.PieConfig {
	return interactive.PieConfig{
		Label: "Rose area", Caption: "Equal angles map seasonal values to sector area.",
		Series: []interactive.PieSeries{{
			Name: "Area", InnerRadius: 40, OuterRadius: 75, RoseMode: interactive.PieRoseArea,
			LabelContent: interactive.PieLabelNameAndValue,
			Data:         interactivePieData([4]float64{11, 45, 37, 6}),
			Options:      interactive.SeriesOptions{Label: &interactive.LabelOptions{Show: interactive.Bool(true)}},
		}},
		Width: "100%", Height: "420px",
		Options: interactivePieOptions("Rose area", "rose-area"),
		Style:   charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractivePieRoseRadius() interactive.PieConfig {
	return interactive.PieConfig{
		Label: "Rose radius", Caption: "Proportional angles remain while seasonal values also change radius.",
		Series: []interactive.PieSeries{{
			Name: "Radius", InnerRadius: 30, OuterRadius: 75, RoseMode: interactive.PieRoseRadius,
			LabelContent: interactive.PieLabelNameAndValue,
			Data:         interactivePieData([4]float64{95, 66, 28, 58}),
			Options:      interactive.SeriesOptions{Label: &interactive.LabelOptions{Show: interactive.Bool(true)}},
		}},
		Width: "100%", Height: "420px",
		Options: interactivePieOptions("Rose radius", "rose-radius"),
		Style:   charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractivePieNested() interactive.PieConfig {
	return interactive.PieConfig{
		Label: "Nested seasonal pie", Caption: "A thin outer area ring surrounds an inner radius rose at one shared center.",
		Series: []interactive.PieSeries{
			{
				Name: "Outer area", InnerRadius: 50, OuterRadius: 55, RoseMode: interactive.PieRoseArea,
				LabelContent: interactive.PieLabelNameAndValue,
				Data:         interactivePieData([4]float64{87, 31, 29, 56}),
				Options:      interactive.SeriesOptions{Label: &interactive.LabelOptions{Show: interactive.Bool(true)}},
			},
			{
				Name: "Inner radius", OuterRadius: 45, RoseMode: interactive.PieRoseRadius,
				Data: interactivePieData([4]float64{37, 31, 85, 26}),
			},
		},
		Width: "100%", Height: "440px",
		Options: interactivePieOptions("Pie in pie", "nested-seasonal-pie"),
		Style:   charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractivePiePairedRoses() interactive.PieConfig {
	return interactive.PieConfig{
		Label: "Area and radius roses", Caption: "Two centers compare equal-angle area encoding with proportional-angle radius encoding.",
		Series: []interactive.PieSeries{
			{
				Name: "Area", InnerRadius: 30, OuterRadius: 75, Center: &interactive.PieCenter{X: 25, Y: 50}, RoseMode: interactive.PieRoseArea,
				LabelContent: interactive.PieLabelNameAndValue, Data: interactivePieData([4]float64{47, 47, 87, 88}),
				Options: interactive.SeriesOptions{Label: &interactive.LabelOptions{Show: interactive.Bool(true)}},
			},
			{
				Name: "Radius", InnerRadius: 30, OuterRadius: 75, Center: &interactive.PieCenter{X: 75, Y: 50}, RoseMode: interactive.PieRoseRadius,
				LabelContent: interactive.PieLabelNameAndValue, Data: interactivePieData([4]float64{90, 15, 41, 8}),
				Options: interactive.SeriesOptions{Label: &interactive.LabelOptions{Show: interactive.Bool(true)}},
			},
		},
		Width: "100%", Height: "440px", Options: interactivePieOptions("roseType(area) vs roseType(radius)", "paired-seasonal-roses"),
		Style: charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractivePieAutoEmphasis() interactive.PieConfig {
	options := interactivePieOptions("dispatch action", "rotating-seasonal-emphasis")
	options.Title.Right = "40%"
	options.Legend = &interactive.LegendOptions{Orient: "vertical", Left: "left", Top: "center"}
	options.Tooltip = &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "item"}
	return interactive.PieConfig{
		Label: "Rotating seasonal emphasis", Caption: "One sector at a time receives emphasis and a tooltip; reduced-motion preference stops the cycle.",
		Series: []interactive.PieSeries{{
			Name: "Seasons", OuterRadius: 55, Center: &interactive.PieCenter{X: 50, Y: 60},
			LabelContent: interactive.PieLabelNameAndValue, Data: interactivePieData([4]float64{13, 90, 94, 63}),
			Options: interactive.SeriesOptions{
				Label:    &interactive.LabelOptions{Show: interactive.Bool(true)},
				Emphasis: &interactive.EmphasisOptions{ItemStyle: &interactive.ItemStyle{ShadowBlur: 10, ShadowColor: "rgba(0, 0, 0, 0.5)"}},
			},
		}},
		Width: "100%", Height: "440px", Options: options, TooltipContent: interactive.PieTooltipNameAndShare,
		AutoEmphasis: &interactive.PieAutoEmphasisOptions{IntervalMilliseconds: 1000},
		Style:        charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}

func sampleInteractivePieSelected() interactive.PieConfig {
	options := interactivePieOptions("selected sector", "selected-seasonal-sector")
	options.Controls = chartcontrol.Options{Fullscreen: true}
	return interactive.PieConfig{
		Label: "Selectable seasonal donut", Caption: "Spring starts selected; readers can toggle any sector without changing the underlying values.",
		Series: []interactive.PieSeries{{
			Name: "Seasons", InnerRadius: 35, OuterRadius: 70, Selectable: true,
			Data: []interactive.PieData{
				{Name: "Spring", Value: 81, Selected: true}, {Name: "Summer", Value: 18},
				{Name: "Autumn", Value: 25}, {Name: "Winter", Value: 40},
			},
		}},
		Width: "100%", Height: "420px", Options: options, Style: charttheme.Style{Class: "max-w-5xl mx-auto"},
	}
}
