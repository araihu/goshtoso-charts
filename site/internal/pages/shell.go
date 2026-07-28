package pages

import (
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	"github.com/araihu/goshtoso-charts/components/dependencies"
	"github.com/araihu/goshtoso-charts/site/internal/brand"
	"github.com/araihu/goshtoso/components/sidebar"
)

func tocID(title string) string {
	return strings.ToLower(strings.ReplaceAll(title, " ", "-"))
}

func shellPage(title string, active string, content templ.Component) componentdocshell.Page {
	return componentdocshell.Page{
		Title:         title,
		DocumentTitle: title + " · Goshtoso Charts",
		Description:   "Static vector and interactive chart components for Goshtoso applications.",
		Active:        active,
		Content:       content,
		Head:          templ.Join(brand.Head(), dependencies.Dependencies()),
		EnableTOC:     active == "getting-started" || active == "line" || active == "bar" || active == "pie" || active == "scatter" || active == "radar" || active == "candlestick" || active == "funnel" || active == "heatmap" || active == "table" || active == "violin" || active == "interactive-bar" || active == "interactive-line" || active == "interactive-candlestick" || active == "interactive-tree" || active == "interactive-sunburst" || active == "interactive-treemap" || active == "interactive-parallel" || active == "interactive-theme-river" || active == "interactive-word-cloud" || active == "interactive-map" || active == "live-availability",
	}
}

func shellConfig() componentdocshell.Config {
	navigation := shellNavigation()
	navigation.SearchSlot = docsSearch(searchEntries(navigation))
	return componentdocshell.Config{
		Brand: componentdocshell.Brand{
			Name:       "Charts",
			HomeURL:    "/",
			Logo:       brand.Logo(),
			FaviconURL: brand.IconURL(),
		},
		Navigation:    navigation,
		Appearance:    componentdocshell.AppearanceConfig{PersistPreferences: true},
		Interactions:  componentdocshell.InteractionConfig{EnableHTMX: true},
		BodyEnd:       docsSearchRuntime(),
		RepositoryURL: "https://github.com/araihu/goshtoso-charts",
	}
}

func shellNavigation() componentdocshell.Navigation {
	return componentdocshell.Navigation{
		Items: []sidebar.Item{
			{ID: "getting-started", Label: "Getting started", Href: "/", Icon: sidebarGettingStartedIcon()},
			{ID: "attributions", Label: "Attributions", Href: "/attributions", Icon: sidebarAttributionsIcon()},
		},
		SectionsTitle: "Documentation",
		Sections: []sidebar.Section{
			{
				Title: "Static / Vector",
				Items: []sidebar.Item{
					{ID: "line", Label: "Line chart", Href: "/components/line"},
					{ID: "bar", Label: "Bar chart", Href: "/components/bar"},
					{ID: "pie", Label: "Pie chart", Href: "/components/pie"},
					{ID: "scatter", Label: "Scatter chart", Href: "/components/scatter"},
					{ID: "radar", Label: "Radar chart", Href: "/components/radar"},
					{ID: "candlestick", Label: "Candlestick", Href: "/components/candlestick"},
					{ID: "funnel", Label: "Funnel chart", Href: "/components/funnel"},
					{ID: "heatmap", Label: "Heat map", Href: "/components/heatmap"},
					{ID: "table", Label: "Table", Href: "/components/table"},
					{ID: "violin", Label: "Violin chart", Href: "/components/violin"},
				},
			},
			{
				Title: "Interactive / Cartesian",
				Items: []sidebar.Item{
					{ID: "interactive-bar", Label: "Bar", Href: "/components/interactive/bar"},
					{ID: "interactive-line", Label: "Line", Href: "/components/interactive/line"},
					{ID: "interactive-scatter", Label: "Scatter", Href: "/components/interactive/scatter"},
					{ID: "interactive-candlestick", Label: "Candlestick", Href: "/components/interactive/candlestick"},
					{ID: "interactive-heatmap", Label: "Heatmap", Href: "/components/interactive/heatmap"},
				},
			},
			{
				Title: "Interactive / Statistical",
				Items: []sidebar.Item{
					{ID: "interactive-pie", Label: "Pie", Href: "/components/interactive/pie"},
					{ID: "interactive-radar", Label: "Radar", Href: "/components/interactive/radar"},
					{ID: "interactive-boxplot", Label: "Box plot", Href: "/components/interactive/boxplot"},
					{ID: "interactive-gauge", Label: "Gauge", Href: "/components/interactive/gauge"},
					{ID: "interactive-funnel", Label: "Funnel", Href: "/components/interactive/funnel"},
					{ID: "interactive-parallel", Label: "Parallel coordinates", Href: "/components/interactive/parallel"},
					{ID: "interactive-theme-river", Label: "Theme river", Href: "/components/interactive/theme-river"},
					{ID: "interactive-word-cloud", Label: "Word cloud", Href: "/components/interactive/word-cloud"},
				},
			},
			{
				Title: "Interactive / Geographic",
				Items: []sidebar.Item{
					{ID: "interactive-map", Label: "Map", Href: "/components/interactive/map"},
				},
			},
			{
				Title: "Interactive / Relationships",
				Items: []sidebar.Item{
					{ID: "interactive-graph", Label: "Graph", Href: "/components/interactive/graph"},
					{ID: "interactive-sankey", Label: "Sankey", Href: "/components/interactive/sankey"},
					{ID: "interactive-tree", Label: "Tree", Href: "/components/interactive/tree"},
					{ID: "interactive-sunburst", Label: "Sunburst", Href: "/components/interactive/sunburst"},
					{ID: "interactive-treemap", Label: "Treemap", Href: "/components/interactive/treemap"},
				},
			},
			{
				Title: "Examples",
				Items: []sidebar.Item{
					{ID: "live-availability", Label: "Live availability", Href: "/examples/live-availability"},
				},
			},
		},
		SearchPlaceholder: "Search docs...",
	}
}
