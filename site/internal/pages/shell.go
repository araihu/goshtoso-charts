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

func eChartsLayoutClass(layout string) string {
	switch layout {
	case "center":
		return "mx-auto grid max-w-5xl gap-4"
	case "flex":
		return "flex flex-wrap gap-4"
	case "none":
		return "space-y-4"
	default:
		return "space-y-4"
	}
}

func shellPage(title string, active string, content templ.Component) componentdocshell.Page {
	return componentdocshell.Page{
		Title:         title,
		DocumentTitle: title + " · Goshtoso Charts",
		Description:   "Server-rendered chart components for Goshtoso applications.",
		Active:        active,
		Content:       content,
		Head:          brand.Head(),
		EnableTOC:     active == "heartbeat" || active == "line" || active == "bar" || active == "pie" || active == "echarts-bar" || active == "echarts-line" || active == "go-echarts",
	}
}

func shellConfig() componentdocshell.Config {
	return componentdocshell.Config{
		Brand: componentdocshell.Brand{
			Name:       "Charts",
			HomeURL:    "/",
			Logo:       brand.Logo(),
			FaviconURL: brand.IconURL(),
		},
		Navigation: componentdocshell.Navigation{
			Items: []sidebar.Item{
				{ID: "overview", Label: "Overview", Href: "/"},
			},
			SectionsTitle: "Documentation",
			Sections: []sidebar.Section{
				{
					Title: "Server-rendered",
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
						{ID: "echarts-bar", Label: "Bar", Href: "/components/echarts/bar"},
						{ID: "echarts-line", Label: "Line", Href: "/components/echarts/line"},
					},
				},
				{
					Title: "Examples",
					Items: []sidebar.Item{
						{ID: "status-page", Label: "Status page", Href: "/examples/status-page"},
						{ID: "go-echarts", Label: "go-echarts catalog", Href: "/examples/go-echarts"},
					},
				},
			},
			SearchPlaceholder: "Search docs...",
		},
		Appearance:    componentdocshell.AppearanceConfig{PersistPreferences: true},
		Interactions:  componentdocshell.InteractionConfig{EnableHTMX: true},
		RepositoryURL: "https://github.com/araihu/goshtoso-charts",
	}
}
