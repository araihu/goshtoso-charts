package main

import (
	"context"
	"os"

	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/line"
)

func main() {
	chart := line.Line(line.Config{
		Label:    "Actionless chart",
		Labels:   []string{"Mon", "Tue", "Wed"},
		Series:   []line.Series{{Name: "Value", Values: []float64{1, 3, 2}}},
		Width:    320,
		Height:   160,
		Controls: chartcontrol.Options{Expand: chartcontrol.Bool(false)},
		Export:   &chartcontrol.ExportOptions{Disabled: true},
	})
	if err := chart.Render(context.Background(), os.Stdout); err != nil {
		panic(err)
	}
}
