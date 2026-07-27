package pages

import (
	"strings"

	"github.com/araihu/goshtoso-app-shells/catalogshell"
	selectfield "github.com/araihu/goshtoso/components/select"
	"github.com/araihu/goshtoso/components/sidebar"
)

func tocID(title string) string {
	return strings.ToLower(strings.ReplaceAll(title, " ", "-"))
}

func shellConfig() catalogshell.Config {
	return catalogshell.Config{
		Brand: catalogshell.Brand{
			Name:    "Goshtoso Charts",
			HomeURL: "/",
		},
		Navigation: catalogshell.Navigation{
			Items: []sidebar.Item{
				{ID: "overview", Label: "Overview", Href: "/"},
			},
			SectionsTitle: "Catalog",
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
			SearchPlaceholder: "Search catalog",
		},
		Themes: []selectfield.Option{
			{Value: "goshtoso", Label: "Goshtoso", Selected: true},
			{Value: "minimal", Label: "Minimal"},
		},
		RepositoryURL:      "https://github.com/araihu/goshtoso-charts",
		PersistPreferences: true,
	}
}
