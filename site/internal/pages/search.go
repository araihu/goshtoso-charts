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
			entries = append(entries, searchEntry{
				Label:    item.Label,
				Href:     item.Href,
				Category: category,
				Terms:    strings.ToLower(strings.Join([]string{item.Label, category, item.ID}, " ")),
			})
		}
		entries = appendSearchItems(entries, item.Items, category)
	}
	return entries
}
