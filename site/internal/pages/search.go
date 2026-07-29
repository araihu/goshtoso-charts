package pages

import (
	"strings"

	"github.com/araihu/goshtoso-app-shells/componentdocshell"
	"github.com/araihu/goshtoso/components/sidebar"
)

type searchEntry struct {
	Label    string
	Href     string
	Category string
	Terms    string
}

var supplementalSearchTerms = map[string]string{
	"getting-started": "choose chart mode wrapper controls static vector interactive",
	"chart-modes":     "static vector interactive capabilities use cases delivery runtime print",
	"chart-controls":  "wrapper lifecycle enabled disabled hidden omitted client events javascript alpine htmx export svg png fullscreen",
}

func searchEntries(navigation componentdocshell.Navigation) []searchEntry {
	entries := make([]searchEntry, 0, len(navigation.Items)+len(navigation.Sections)*2)
	entries = appendSearchItems(entries, navigation.Items, "General")
	for _, section := range navigation.Sections {
		entries = appendSearchItems(entries, section.Items, section.Title)
	}
	return entries
}

func appendSearchItems(entries []searchEntry, items []sidebar.Item, category string) []searchEntry {
	for _, item := range items {
		if !item.Disabled && item.Href != "" {
			extraTerms := supplementalSearchTerms[item.ID]
			entries = append(entries, searchEntry{
				Label:    item.Label,
				Href:     item.Href,
				Category: category,
				Terms:    strings.ToLower(strings.TrimSpace(strings.Join([]string{item.Label, category, item.ID, extraTerms}, " "))),
			})
		}
		entries = appendSearchItems(entries, item.Items, category)
	}
	return entries
}
