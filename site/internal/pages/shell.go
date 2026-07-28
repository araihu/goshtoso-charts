package pages

import (
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
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
		Head:          brand.Head(),
		EnableTOC:     active == "heartbeat" || active == "line" || active == "bar" || active == "pie" || active == "interactive-bar" || active == "interactive-line",
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
			{ID: "overview", Label: "Overview", Href: "/", Icon: sidebarOverviewIcon()},
			{ID: "attributions", Label: "Attributions", Href: "/attributions", Icon: sidebarAttributionsIcon()},
		},
		SectionsTitle: "Documentation",
		Sections: []sidebar.Section{
			{
				Title: "Static / Vector",
				Items: []sidebar.Item{
					{ID: "heartbeat", Label: "Heartbeat", Href: "/components/heartbeat"},
					{ID: "line", Label: "Line chart", Href: "/components/line"},
					{ID: "bar", Label: "Bar chart", Href: "/components/bar"},
					{ID: "pie", Label: "Pie chart", Href: "/components/pie"},
				},
			},
			{
				Title: "Interactive / Cartesian",
				Items: []sidebar.Item{
					{ID: "interactive-bar", Label: "Bar", Href: "/components/interactive/bar"},
					{ID: "interactive-line", Label: "Line", Href: "/components/interactive/line"},
					{ID: "interactive-scatter", Label: "Scatter", Href: "/components/interactive/scatter"},
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
				},
			},
			{
				Title: "Examples",
				Items: []sidebar.Item{
					{ID: "status-page", Label: "Status page", Href: "/examples/status-page"},
				},
			},
		},
		SearchPlaceholder: "Search docs...",
	}
}
