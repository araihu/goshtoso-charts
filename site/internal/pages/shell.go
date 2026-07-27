package pages

import (
	"strings"

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	"github.com/araihu/goshtoso/components/sidebar"
)

func tocID(title string) string {
	return strings.ToLower(strings.ReplaceAll(title, " ", "-"))
}

func shellPage(title string, active string, content templ.Component) componentdocshell.Page {
	return componentdocshell.Page{
		Title:       title,
		Description: "Server-rendered chart components for Goshtoso applications.",
		Active:      active,
		Content:     content,
		EnableTOC:   active == "heartbeat" || active == "line",
	}
}

func shellConfig() componentdocshell.Config {
	return componentdocshell.Config{
		Brand: componentdocshell.GoshtosoBrand("Goshtoso Charts", "/", ""),
		Navigation: componentdocshell.Navigation{
			Items: []sidebar.Item{
				{ID: "overview", Label: "Overview", Href: "/"},
			},
			SectionsTitle: "Documentation",
			Sections: []sidebar.Section{
				{
					Title: "Components",
					Items: []sidebar.Item{
						{ID: "heartbeat", Label: "Heartbeat", Href: "/components/heartbeat"},
						{ID: "line", Label: "Line chart", Href: "/components/line"},
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
		},
		Appearance:    componentdocshell.AppearanceConfig{PersistPreferences: true},
		Interactions:  componentdocshell.InteractionConfig{EnableHTMX: true},
		RepositoryURL: "https://github.com/araihu/goshtoso-charts",
	}
}
