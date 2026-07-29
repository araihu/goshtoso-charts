package pages

import (
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/araihu/goshtoso-charts/components/interactive"
)

const (
	interactivePieUpstreamPath     = "examples/pie.go"
	interactivePieUpstreamRevision = "bda428480a82d6d77ebb9fa939cf8d52528453dd"
	interactivePieUpstreamSHA256   = "a59bb6f11818d4175d033f025f00a58e6a191eff5acf30f0e0cd5f98cd493ada"
)

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
