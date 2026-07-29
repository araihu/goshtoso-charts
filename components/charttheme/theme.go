// Package charttheme provides chart-specific palettes that complement Goshtoso themes.
package charttheme

import (
	"strconv"
	"strings"
)

// Palette selects a built-in chart palette.
type Palette string

const (
	// PaletteAuto follows a chart-aware theme when one exists and otherwise uses Bold.
	PaletteAuto Palette = ""
	// PaletteAraiHu complements AraiHu's lime accent with warm and contrasting hues.
	PaletteAraiHu Palette = "araihu"
	// PaletteBold is the high-contrast fallback palette.
	PaletteBold Palette = "bold"
	// PaletteNeutral is the achromatic fallback palette.
	PaletteNeutral Palette = "neutral"
	// PalettePastel is the low-saturation fallback palette.
	PalettePastel Palette = "pastel"
	// PaletteStatus maps the first three series to positive, warning, and negative states.
	PaletteStatus Palette = "status"
)

// Style controls chart palette and root classes.
//
// Colors replace the selected palette in order. Class is appended to the
// chart's root figure and may contain Tailwind utility classes or an
// application-defined class that overrides --color-chart-* tokens.
type Style struct {
	Palette Palette
	Colors  []string
	Class   string
}

var palettes = map[Palette][]string{
	PaletteAraiHu:  {"#4d7c0f", "#ff8a3d", "#ef476f", "#4cc9f0", "#9b5de5", "#ffd166", "#f15bb5", "#00b8a9"},
	PaletteBold:    {"#2563eb", "#dc2626", "#16a34a", "#d97706", "#7c3aed", "#0891b2", "#db2777", "#65a30d"},
	PaletteNeutral: {"#111827", "#374151", "#4b5563", "#6b7280", "#9ca3af", "#d1d5db", "#78716c", "#a8a29e"},
	PalettePastel:  {"#93c5fd", "#fca5a5", "#86efac", "#fcd34d", "#c4b5fd", "#67e8f9", "#f9a8d4", "#bef264"},
	PaletteStatus:  {"#16a34a", "#d97706", "#dc2626", "#2563eb", "#7c3aed", "#0891b2", "#db2777", "#65a30d"},
}

// ResolvedColors returns a copy of the initial literal palette. Explicit
// Style.Colors take precedence. Auto starts with Bold; interactive components
// replace it with computed theme tokens after the private runtime initializes.
func (style Style) ResolvedColors() []string {
	palette := style.Palette
	if palette == PaletteAuto {
		palette = PaletteBold
	}
	colors, ok := palettes[palette]
	if !ok {
		colors = palettes[PaletteBold]
	}
	resolved := append([]string(nil), colors...)
	for index, color := range style.Colors {
		if strings.TrimSpace(color) == "" {
			continue
		}
		if index < len(resolved) {
			resolved[index] = color
		} else {
			resolved = append(resolved, color)
		}
	}
	return resolved
}

// RootClasses composes required chart palette classes with caller classes.
func (style Style) RootClasses(base string) string {
	palette := style.Palette
	if palette == PaletteAuto {
		palette = "auto"
	}
	parts := []string{strings.TrimSpace(base), "goshtoso-charts-palette", "goshtoso-charts-palette-" + string(palette)}
	if class := strings.TrimSpace(style.Class); class != "" {
		parts = append(parts, class)
	}
	return strings.Join(parts, " ")
}

// SeriesColor returns an explicit color when supplied, otherwise a CSS token.
func (style Style) SeriesColor(index int) string {
	if index >= 0 && index < len(style.Colors) && strings.TrimSpace(style.Colors[index]) != "" {
		return style.Colors[index]
	}
	return "var(--color-chart-series-" + strconv.Itoa(index%8+1) + ")"
}
