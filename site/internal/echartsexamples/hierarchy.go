package echartsexamples

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/render"
)

const hierarchyExamplesSource = "https://github.com/go-echarts/examples/blob/master/examples/"

// HierarchyExamples ports hierarchy, relationship, and multidimensional examples
// from github.com/go-echarts/examples. Fixture-backed upstream examples use the
// deterministic values below, so rendering never depends on a local fixture.
var HierarchyExamples = []Example{
	{Slug: "tree", Title: "Tree", Group: "Hierarchy & flow", Source: hierarchyExamplesSource + "tree.go", Build: func() render.Renderer { return treeBase() }},
	{Slug: "treemap", Title: "Treemap", Group: "Hierarchy & flow", Source: hierarchyExamplesSource + "treemap.go", Build: func() render.Renderer { return treeMapBase() }},
	{Slug: "sunburst", Title: "Sunburst", Group: "Hierarchy & flow", Source: hierarchyExamplesSource + "sunburst.go", Build: func() render.Renderer { return sunburstBase() }},
	{Slug: "graph-force", Title: "Force graph", Group: "Hierarchy & flow", Source: hierarchyExamplesSource + "graph.go", Build: func() render.Renderer { return graphBase() }},
	{Slug: "graph-circular", Title: "Circular graph", Group: "Hierarchy & flow", Source: hierarchyExamplesSource + "graph.go", Build: func() render.Renderer { return graphCircle() }},
	{Slug: "graph-dependencies", Title: "Dependency graph", Group: "Hierarchy & flow", Source: hierarchyExamplesSource + "graph.go", Build: func() render.Renderer { return graphNpmDep() }},
	{Slug: "sankey", Title: "Sankey", Group: "Hierarchy & flow", Source: hierarchyExamplesSource + "sankey.go", Build: func() render.Renderer { return sankeyBase() }},
	{Slug: "sankey-energy", Title: "Energy Sankey", Group: "Hierarchy & flow", Source: hierarchyExamplesSource + "sankey.go", Build: func() render.Renderer { return graphEnergy() }},
	{Slug: "parallel", Title: "Parallel", Group: "Hierarchy & flow", Source: hierarchyExamplesSource + "parallel.go", Build: func() render.Renderer { return parallelBase() }},
	{Slug: "parallel-component", Title: "Parallel component", Group: "Hierarchy & flow", Source: hierarchyExamplesSource + "parallel.go", Build: func() render.Renderer { return parallelComponent() }},
	{Slug: "parallel-multi", Title: "Parallel multi-series", Group: "Hierarchy & flow", Source: hierarchyExamplesSource + "parallel.go", Build: func() render.Renderer { return parallelMulti() }},
	{Slug: "theme-river", Title: "Theme river", Group: "Hierarchy & flow", Source: hierarchyExamplesSource + "themeriver.go", Build: func() render.Renderer { return themeRiverTime() }},
}

func treeBase() *charts.Tree {
	tree := charts.NewTree()
	tree.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Basic tree example"}))
	tree.AddSeries("tree", []opts.TreeData{{Name: "Root", Children: []*opts.TreeData{
		{Name: "Node 1", Children: []*opts.TreeData{{Name: "Child 1"}}},
		{Name: "Node 2", Children: []*opts.TreeData{{Name: "Child 1"}, {Name: "Child 2"}, {Name: "Child 3"}}},
		{Name: "Node 3", Collapsed: opts.Bool(true), Children: []*opts.TreeData{{Name: "Child 1"}, {Name: "Child 2"}, {Name: "Child 3"}}},
	}}}).SetSeriesOptions(charts.WithTreeOpts(opts.TreeChart{Layout: "orthogonal", Orient: "LR", InitialTreeDepth: -1, Leaves: &opts.TreeLeaves{Label: &opts.Label{Show: opts.Bool(true), Position: "right"}}}), charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Position: "top"}))
	return tree
}

func treeMapBase() *charts.TreeMap {
	treeMap := charts.NewTreeMap()
	treeMap.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Basic treemap example", Subtitle: "File system usage", Left: "center"}), charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}))
	data := []opts.TreeMapNode{
		{Name: "documents", Children: []opts.TreeMapNode{{Name: "proposal.pdf", Value: 1000}}},
		{Name: "media", Children: []opts.TreeMapNode{{Name: "image-1", Value: 100}, {Name: "image-2", Value: 300}, {Name: "image-3", Value: 200}}},
		{Name: "cache", Children: []opts.TreeMapNode{{Name: "cache-1", Value: 8}, {Name: "cache-2", Value: 12}, {Name: "cache-3", Value: 16}, {Name: "cache-4", Value: 10}}},
		{Name: "readme", Value: 450},
	}
	treeMap.AddSeries("Root FS", data).SetSeriesOptions(charts.WithTreeMapOpts(opts.TreeMapChart{Animation: opts.Bool(true), Roam: opts.Bool(true), UpperLabel: &opts.UpperLabel{Show: opts.Bool(true)}}), charts.WithItemStyleOpts(opts.ItemStyle{BorderColor: "#fff"}), charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Position: "inside", Color: "white"}))
	return treeMap
}

func sunburstBase() *charts.Sunburst {
	sunburst := charts.NewSunburst()
	sunburst.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Basic sunburst example"}))
	sunburst.AddSeries("sunburst", []opts.SunBurstData{
		{Name: "parent-0", Value: 8, Children: []*opts.SunBurstData{{Name: "child-0", Value: 5}}},
		{Name: "parent-1", Value: 6, Children: []*opts.SunBurstData{{Name: "child-1", Value: 4}}},
		{Name: "parent-2", Value: 7, Children: []*opts.SunBurstData{{Name: "child-2", Value: 6}}},
		{Name: "parent-3", Value: 5, Children: []*opts.SunBurstData{{Name: "child-3", Value: 3}}},
		{Name: "parent-4", Value: 9, Children: []*opts.SunBurstData{{Name: "child-4", Value: 7}}},
	})
	return sunburst
}

var graphNodes = []opts.GraphNode{{Name: "Node1"}, {Name: "Node2"}, {Name: "Node3"}, {Name: "Node4"}, {Name: "Node5"}, {Name: "Node6"}, {Name: "Node7"}, {Name: "Node8"}}

func completeGraphLinks(nodes []opts.GraphNode) []opts.GraphLink {
	links := make([]opts.GraphLink, 0, len(nodes)*len(nodes))
	for _, source := range nodes {
		for _, target := range nodes {
			links = append(links, opts.GraphLink{Source: source.Name, Target: target.Name})
		}
	}
	return links
}

func graphBase() *charts.Graph {
	graph := charts.NewGraph()
	graph.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Basic graph example"}))
	graph.AddSeries("graph", graphNodes, completeGraphLinks(graphNodes), charts.WithGraphChartOpts(opts.GraphChart{Force: &opts.GraphForce{Repulsion: 8000}}))
	return graph
}

func graphCircle() *charts.Graph {
	graph := charts.NewGraph()
	graph.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Circular layout"}))
	graph.AddSeries("graph", graphNodes, completeGraphLinks(graphNodes)).SetSeriesOptions(charts.WithGraphChartOpts(opts.GraphChart{Force: &opts.GraphForce{Repulsion: 8000}, Layout: "circular"}), charts.WithLabelOpts(opts.Label{Show: opts.Bool(true), Position: "right"}))
	return graph
}

func graphNpmDep() *charts.Graph {
	graph := charts.NewGraph()
	graph.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "npm dependencies demo"}))
	nodes := []opts.GraphNode{{Name: "goshtoso"}, {Name: "templ"}, {Name: "go-echarts"}, {Name: "render"}, {Name: "assets"}}
	links := []opts.GraphLink{{Source: "goshtoso", Target: "templ"}, {Source: "goshtoso", Target: "go-echarts"}, {Source: "go-echarts", Target: "render"}, {Source: "goshtoso", Target: "assets"}}
	graph.AddSeries("graph", nodes, links).SetSeriesOptions(charts.WithGraphChartOpts(opts.GraphChart{Layout: "none", Roam: opts.Bool(true), FocusNodeAdjacency: opts.Bool(true)}), charts.WithEmphasisOpts(opts.Emphasis{Label: &opts.Label{Show: opts.Bool(true), Color: "black", Position: "left"}}), charts.WithLineStyleOpts(opts.LineStyle{Curveness: 0.3}))
	return graph
}

func sankeyBase() *charts.Sankey {
	sankey := charts.NewSankey()
	sankey.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Sankey basic example"}))
	nodes := []opts.SankeyNode{{Name: "category1"}, {Name: "category2"}, {Name: "category3"}, {Name: "category4"}, {Name: "category5"}, {Name: "category6"}}
	links := []opts.SankeyLink{{Source: "category1", Target: "category2", Value: 10}, {Source: "category2", Target: "category3", Value: 15}, {Source: "category3", Target: "category4", Value: 20}, {Source: "category5", Target: "category6", Value: 25}}
	sankey.AddSeries("sankey", nodes, links, charts.WithLabelOpts(opts.Label{Show: opts.Bool(true)}))
	return sankey
}

func graphEnergy() *charts.Sankey {
	sankey := charts.NewSankey()
	sankey.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Sankey energy example"}))
	nodes := []opts.SankeyNode{{Name: "Solar"}, {Name: "Wind"}, {Name: "Grid"}, {Name: "Homes"}, {Name: "Industry"}}
	links := []opts.SankeyLink{{Source: "Solar", Target: "Grid", Value: 30}, {Source: "Wind", Target: "Grid", Value: 45}, {Source: "Grid", Target: "Homes", Value: 40}, {Source: "Grid", Target: "Industry", Value: 35}}
	sankey.AddSeries("sankey", nodes, links).SetSeriesOptions(charts.WithLineStyleOpts(opts.LineStyle{Color: "source", Curveness: 0.5}), charts.WithLabelOpts(opts.Label{Show: opts.Bool(true)}))
	return sankey
}

var parallelAxes = []opts.ParallelAxis{
	{Dim: 0, Name: "Date", Inverse: opts.Bool(true), Max: 31, NameLocation: "start"},
	{Dim: 1, Name: "AQI"}, {Dim: 2, Name: "PM2.5"}, {Dim: 3, Name: "PM10"}, {Dim: 4, Name: "CO"}, {Dim: 5, Name: "NO2"}, {Dim: 6, Name: "SO2"},
	{Dim: 7, Name: "Level", Type: "category", Data: []string{"Good", "Moderate", "Lightly", "Moderately", "Heavily", "Severely"}},
}

var (
	parallelDataBJ = [][]interface{}{{1, 55, 9, 56, 0.46, 18, 6, "Moderate"}, {2, 25, 11, 21, 0.65, 34, 9, "Good"}, {3, 56, 7, 63, 0.3, 14, 5, "Moderate"}, {4, 82, 58, 90, 1.77, 68, 33, "Moderate"}}
	parallelDataGZ = [][]interface{}{{1, 26, 37, 27, 1.163, 27, 13, "Good"}, {2, 85, 62, 71, 1.195, 60, 8, "Moderate"}, {3, 78, 38, 74, 1.363, 37, 7, "Moderate"}, {4, 21, 21, 36, 0.634, 40, 9, "Good"}}
	parallelDataSH = [][]interface{}{{1, 91, 45, 125, 0.82, 34, 23, "Moderate"}, {2, 65, 27, 78, 0.86, 45, 29, "Moderate"}, {3, 83, 60, 84, 1.09, 73, 27, "Moderate"}, {4, 109, 81, 121, 1.28, 68, 51, "Lightly"}}
)

func parallelItems(data [][]interface{}) []opts.ParallelData {
	items := make([]opts.ParallelData, 0, len(data))
	for _, value := range data {
		items = append(items, opts.ParallelData{Value: value})
	}
	return items
}

func parallelBase() *charts.Parallel {
	parallel := charts.NewParallel()
	parallel.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Basic parallel example"}), charts.WithParallelAxisList(parallelAxes))
	parallel.AddSeries("Beijing", parallelItems(parallelDataBJ))
	return parallel
}

func parallelComponent() *charts.Parallel {
	parallel := charts.NewParallel()
	parallel.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "With component"}), charts.WithParallelComponentOpts(opts.ParallelComponent{Left: "15%", Right: "13%", Bottom: "10%", Top: "20%"}), charts.WithParallelAxisList(parallelAxes))
	parallel.AddSeries("Beijing", parallelItems(parallelDataBJ))
	return parallel
}

func parallelMulti() *charts.Parallel {
	parallel := charts.NewParallel()
	parallel.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Multi series"}), charts.WithParallelAxisList(parallelAxes))
	parallel.AddSeries("Beijing", parallelItems(parallelDataBJ)).AddSeries("Guangzhou", parallelItems(parallelDataGZ)).AddSeries("Shanghai", parallelItems(parallelDataSH))
	return parallel
}

func themeRiverTime() *charts.ThemeRiver {
	river := charts.NewThemeRiver()
	river.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "ThemeRiver single-axis time"}), charts.WithSingleAxisOpts(opts.SingleAxis{Type: "time", Bottom: "10%"}), charts.WithTooltipOpts(opts.Tooltip{Trigger: "axis"}))
	river.AddSeries("themeRiver", []opts.ThemeRiverData{
		{"2015/11/08", 10, "DQ"}, {"2015/11/09", 15, "DQ"}, {"2015/11/10", 35, "DQ"}, {"2015/11/11", 38, "DQ"}, {"2015/11/12", 22, "DQ"},
		{"2015/11/08", 35, "TY"}, {"2015/11/09", 36, "TY"}, {"2015/11/10", 37, "TY"}, {"2015/11/11", 22, "TY"}, {"2015/11/12", 24, "TY"},
		{"2015/11/08", 21, "SS"}, {"2015/11/09", 25, "SS"}, {"2015/11/10", 27, "SS"}, {"2015/11/11", 23, "SS"}, {"2015/11/12", 24, "SS"},
		{"2015/11/08", 10, "QG"}, {"2015/11/09", 15, "QG"}, {"2015/11/10", 35, "QG"}, {"2015/11/11", 38, "QG"}, {"2015/11/12", 22, "QG"},
	})
	return river
}
