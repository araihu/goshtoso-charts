package pages

import (
	"time"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	interactive "github.com/araihu/goshtoso-charts/components/interactive"
)

const (
	interactiveLineTimeUpstreamPath     = "examples/line.go (lineTime and generateLineItemsTwoAxis)"
	interactiveLineTimeUpstreamRevision = "bda428480a82d6d77ebb9fa939cf8d52528453dd"
)

func sampleInteractiveLineTime() interactive.LineConfig {
	axis := make([]time.Time, 0, 50)
	for offset := 0; offset < 50; offset++ {
		// Day zero intentionally normalizes to January 31, as in source sample.
		axis = append(axis, time.Date(2025, time.February, offset, 0, 0, 0, 0, time.UTC))
	}
	return interactive.LineConfig{
		Label: "Temporal X axis", Caption: "Fifty deterministic UTC time/value observations.",
		TimeAxis: &interactive.LineTimeAxis{
			Values: axis, Minimum: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		Series: []interactive.LineSeries{{Name: "Category A", Data: deterministicLineTimeData(50)}},
		Options: interactive.ChartOptions{
			Title:    &interactive.TitleOptions{Text: "temporal X axis", Subtitle: "time.Date as X axis values"},
			Legend:   &interactive.LegendOptions{Bottom: "0"},
			Tooltip:  &interactive.TooltipOptions{Show: interactive.Bool(true), Trigger: "axis"},
			YAxis:    &interactive.AxisOptions{Min: interactive.Float(0), Max: interactive.Float(200)},
			Controls: chartcontrol.Options{Fullscreen: true},
			Export:   &chartcontrol.ExportOptions{Filename: "temporal-x-axis"},
		},
	}
}

func deterministicLineTimeData(count int) []interactive.LineData {
	values := make([]interactive.LineData, count)
	state := uint32(7)
	for index := range values {
		state = state*1664525 + 1013904223
		values[index] = interactive.LineData{Value: float64(100 + state%20)}
	}
	return values
}

func interactiveLineTimeCode() string {
	return `@interactive.Line(interactive.LineConfig{
  Label: "Temporal X axis",
	  TimeAxis: &interactive.LineTimeAxis{
	    Minimum: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
    Values: []time.Time{time.Date(2025, time.February, 0, 0, 0, 0, 0, time.UTC)},
  },
  Series: []interactive.LineSeries{{Name: "Category A", Data: []interactive.LineData{{Value: 107}}}},
})`
}
