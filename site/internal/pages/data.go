package pages

import (
	"github.com/araihu/goshtoso-charts/components/bar"
	"github.com/araihu/goshtoso-charts/components/line"
	"github.com/araihu/goshtoso-charts/components/pie"
	"github.com/araihu/goshtoso-charts/components/scatter"
	"github.com/araihu/goshtoso/components/codeblock"
)

func gettingStartedCodeBlock(language, label, code string) codeblock.Config {
	return codeblock.Config{Language: language, Label: label, Code: code}
}

const (
	gettingStartedInstallCode = `go get github.com/araihu/goshtoso-charts`
	gettingStartedAssetsCode  = `import chartassets "github.com/araihu/goshtoso-charts/assets"

mux.Handle("GET "+chartassets.Prefix, chartassets.Handler())`
	gettingStartedDependenciesCode = `import "github.com/araihu/goshtoso-charts/components/dependencies"

templ Layout() {
  <head>
    @dependencies.Dependencies()
  </head>
}`
	gettingStartedStaticCode = `@line.Line(line.Config{
  Label: "Request latency",
  Labels: []string{"Mon", "Tue", "Wed"},
  Series: []line.Series{{Name: "p95 (ms)", Values: []float64{42, 47, 44}}},
})`
	gettingStartedInteractiveCode = `@interactive.Bar(interactive.BarConfig{
  Label: "Deployments",
  XAxis: []string{"Mon", "Tue", "Wed"},
  Series: []interactive.BarSeries{{
    Name: "Production",
    Data: []interactive.BarData{{Value: 3}, {Value: 5}, {Value: 4}},
  }},
})`
)

func sampleLatency() line.Config {
	return line.Config{
		Label:   "HTTPS monitor latency in milliseconds",
		Caption: "Median latency, last seven checks.",
		Labels:  []string{"08:00", "08:01", "08:02", "08:03", "08:04", "08:05", "08:06"},
		Series:  []line.Series{{Name: "Latency (ms)", Values: []float64{42, 47, 900, 51, 2_000, 44, 46}}},
	}
}

func sampleDeployments() bar.Config {
	return bar.Config{
		Label:   "Deployments by environment",
		Caption: "Successful and failed deployments this week.",
		Labels:  []string{"Development", "Staging", "Production"},
		Series: []bar.Series{
			{Name: "Successful", Values: []float64{18, 12, 9}},
			{Name: "Failed", Values: []float64{1, 2, 1}},
		},
		Stacked: true,
	}
}

func sampleObservationStates() pie.Config {
	return pie.Config{
		Label:   "Observation states",
		Caption: "Most recent 100 retained monitor observations.",
		Slices: []pie.Slice{
			{Name: "Up", Value: 94},
			{Name: "Degraded", Value: 4},
			{Name: "Down", Value: 2},
		},
	}
}

func sampleBasicScatter() scatter.Config {
	return scatter.Config{
		Label:      "Scatter series by day",
		Caption:    "Five series across Monday through Sunday; the missing Thursday point in Email preserves the upstream null value.",
		Categories: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
		Options:    scatter.Options{Size: 4},
		Series: []scatter.Series{
			{
				Name:   "Email",
				Points: scatterPoints([]string{"Mon", "Tue", "Wed", "Fri", "Sat", "Sun"}, []float64{120, 132, 101, 90, 230, 210}),
			},
			{
				Name:   "Union Ads",
				Points: scatterPoints([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}, []float64{220, 182, 191, 234, 290, 330, 310}),
			},
			{Name: "Video Ads", Points: scatterPoints([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}, []float64{150, 232, 201, 154, 190, 330, 410})},
			{Name: "Direct", Points: scatterPoints([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}, []float64{320, 332, 301, 334, 390, 330, 320})},
			{Name: "Search Engine", Points: scatterPoints([]string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}, []float64{820, 932, 901, 934, 1290, 1330, 1320})},
		},
	}
}

func scatterPoints(categories []string, values []float64) []scatter.Point {
	points := make([]scatter.Point, len(categories))
	for index := range categories {
		points[index] = scatter.Point{Category: categories[index], Value: values[index]}
	}
	return points
}

func lineCode() string {
	return `@line.Line(line.Config{
  Label: "HTTPS monitor latency in milliseconds",
  Labels: []string{"08:00", "08:01", "08:02"},
  Series: []line.Series{{Name: "Latency (ms)", Values: []float64{42, 47, 51}}},
})`
}

func barCode() string {
	return `@bar.Bar(bar.Config{
  Label: "Deployments by environment",
  Labels: []string{"Development", "Staging", "Production"},
  Series: []bar.Series{
    {Name: "Successful", Values: []float64{18, 12, 9}},
    {Name: "Failed", Values: []float64{1, 2, 1}},
  },
  Stacked: true,
})`
}

func pieCode() string {
	return `@pie.Pie(pie.Config{
  Label: "Observation states",
  Slices: []pie.Slice{
    {Name: "Up", Value: 94},
    {Name: "Degraded", Value: 4},
    {Name: "Down", Value: 2},
  },
})`
}

func scatterCode() string {
	return `@scatter.Scatter(scatter.Config{
  Label: "Scatter series by day",
  Categories: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
  Options: scatter.Options{Size: 4},
  Series: []scatter.Series{
    {
      Name: "Email",
      Points: []scatter.Point{
        {Category: "Mon", Value: 120}, {Category: "Tue", Value: 132},
        {Category: "Wed", Value: 101}, {Category: "Fri", Value: 90},
        {Category: "Sat", Value: 230}, {Category: "Sun", Value: 210},
      },
    },
    {
      Name: "Union Ads",
      Points: []scatter.Point{
        {Category: "Mon", Value: 220}, {Category: "Tue", Value: 182},
        {Category: "Wed", Value: 191}, {Category: "Thu", Value: 234},
        {Category: "Fri", Value: 290}, {Category: "Sat", Value: 330},
        {Category: "Sun", Value: 310},
      },
    },
    // Video Ads, Direct, and Search Engine retain the same upstream values.
  },
})`
}
