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

func sampleDenseScatter() scatter.Config {
	return scatter.Config{
		Label:      "Dense sample populations",
		Caption:    "One to three observations per population and sample index.",
		Categories: []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"},
		Options:    scatter.Options{Size: 2},
		Series: []scatter.Series{
			{
				Name: "One",
				Points: []scatter.Point{
					{Category: "0", Value: 42}, {Category: "1", Value: 44},
					{Category: "2", Value: 43}, {Category: "2", Value: 47},
					{Category: "3", Value: 45}, {Category: "4", Value: 46}, {Category: "4", Value: 49},
					{Category: "5", Value: 52}, {Category: "6", Value: 50}, {Category: "6", Value: 54},
					{Category: "7", Value: 55}, {Category: "8", Value: 57}, {Category: "8", Value: 61},
					{Category: "9", Value: 60}, {Category: "10", Value: 62}, {Category: "10", Value: 66},
					{Category: "10", Value: 68}, {Category: "11", Value: 65},
				},
			},
			{
				Name: "Two",
				Points: []scatter.Point{
					{Category: "0", Value: 88}, {Category: "1", Value: 91},
					{Category: "2", Value: 89}, {Category: "2", Value: 94},
					{Category: "3", Value: 96}, {Category: "4", Value: 93}, {Category: "4", Value: 99},
					{Category: "5", Value: 102}, {Category: "6", Value: 100}, {Category: "6", Value: 106},
					{Category: "7", Value: 109}, {Category: "8", Value: 107}, {Category: "8", Value: 113},
					{Category: "9", Value: 116}, {Category: "10", Value: 114}, {Category: "10", Value: 121},
					{Category: "10", Value: 124}, {Category: "11", Value: 119},
				},
			},
			{
				Name: "Three",
				Points: []scatter.Point{
					{Category: "0", Value: 136}, {Category: "1", Value: 142},
					{Category: "2", Value: 139}, {Category: "2", Value: 148},
					{Category: "3", Value: 151}, {Category: "4", Value: 147}, {Category: "4", Value: 156},
					{Category: "5", Value: 160}, {Category: "6", Value: 157}, {Category: "6", Value: 166},
					{Category: "7", Value: 171}, {Category: "8", Value: 168}, {Category: "8", Value: 177},
					{Category: "9", Value: 181}, {Category: "10", Value: 178}, {Category: "10", Value: 188},
					{Category: "10", Value: 193}, {Category: "11", Value: 186},
				},
			},
		},
	}
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
  Label: "Dense sample populations",
  Categories: []string{"0", "1", "2", "3", "4"},
  Options: scatter.Options{Size: 2},
  Series: []scatter.Series{
    {
      Name: "One",
      Points: []scatter.Point{
        {Category: "0", Value: 42},
        {Category: "1", Value: 44},
        {Category: "2", Value: 43},
        {Category: "2", Value: 47},
        {Category: "3", Value: 45},
        {Category: "4", Value: 46},
        {Category: "4", Value: 49},
      },
    },
    {
      Name: "Two",
      Points: []scatter.Point{
        {Category: "0", Value: 88},
        {Category: "1", Value: 91},
        {Category: "2", Value: 89},
        {Category: "2", Value: 94},
        {Category: "3", Value: 96},
        {Category: "4", Value: 93},
        {Category: "4", Value: 99},
      },
    },
  },
})`
}
