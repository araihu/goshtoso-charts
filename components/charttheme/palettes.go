package charttheme

import (
	"fmt"
	"strings"
)

const chartSeriesSlots = 12

// Theme identifiers follow Goshtoso v0.1.1's exact 15-theme catalog in
// all-themes.css plus the released Arai Hû organization theme from
// goshtoso-app-shells v0.1.0. Keep order stable: generated CSS and coverage
// tests use this as the canonical 16-theme chart contract.
var themePalettes = []themePalette{
	palette("araihu",
		[8]string{"#4d7c0f", "#ff8a3d", "#ef476f", "#4cc9f0", "#9b5de5", "#ffd166", "#f15bb5", "#00b8a9"},
		[3]string{"#38bdf8", "#f59e0b", "#e11d48"},
		[8]string{"#c7ff4a", "#ff8a3d", "#ef476f", "#4cc9f0", "#9b5de5", "#ffd166", "#f15bb5", "#00b8a9"},
		[3]string{"#0e7490", "#fbbf24", "#fb7185"}),
	palette("goshtoso",
		[8]string{"#2172a3", "#4b7d1c", "#3a4b5f", "#0f766e", "#c2410c", "#6d28d9", "#be123c", "#a16207"},
		[3]string{"#3bcef7", "#d3d75b", "#dc2626"},
		[8]string{"#3bcef7", "#d3d75b", "#a0b1c5", "#5eead4", "#fdba74", "#c4b5fd", "#fda4af", "#fcd34d"},
		[3]string{"#0f766e", "#f59e0b", "#f87171"}),
	palette("arctic",
		[8]string{"#1d4ed8", "#4338ca", "#0e7490", "#0f766e", "#6d28d9", "#0369a1", "#047857", "#be123c"},
		[3]string{"#67e8f9", "#f59e0b", "#dc2626"},
		[8]string{"#93c5fd", "#a5b4fc", "#67e8f9", "#5eead4", "#c4b5fd", "#7dd3fc", "#6ee7b7", "#fda4af"},
		[3]string{"#0e7490", "#fbbf24", "#f87171"}),
	palette("high-contrast",
		[8]string{"#0c4a6e", "#312e81", "#991b1b", "#166534", "#854d0e", "#581c87", "#155e75", "#831843"},
		[3]string{"#155e75", "#a16207", "#991b1b"},
		[8]string{"#7dd3fc", "#a5b4fc", "#fca5a5", "#86efac", "#fde047", "#d8b4fe", "#67e8f9", "#f9a8d4"},
		[3]string{"#67e8f9", "#fde047", "#fca5a5"}),
	palette("minimal",
		[8]string{"#111827", "#404040", "#1d4ed8", "#047857", "#a16207", "#6d28d9", "#0e7490", "#be123c"},
		[3]string{"#cbd5e1", "#64748b", "#111827"},
		[8]string{"#f9fafb", "#d4d4d4", "#93c5fd", "#6ee7b7", "#fcd34d", "#c4b5fd", "#67e8f9", "#fda4af"},
		[3]string{"#1f2937", "#9ca3af", "#f3f4f6"}),
	palette("modern",
		[8]string{"#171717", "#404040", "#0369a1", "#0f766e", "#c2410c", "#6d28d9", "#be123c", "#4d7c0f"},
		[3]string{"#bae6fd", "#0284c7", "#172554"},
		[8]string{"#fafafa", "#d4d4d4", "#7dd3fc", "#5eead4", "#fdba74", "#c4b5fd", "#fda4af", "#bef264"},
		[3]string{"#0c4a6e", "#38bdf8", "#bae6fd"}),
	palette("neo-brutalism",
		[8]string{"#7e22ce", "#4d7c0f", "#1d4ed8", "#b91c1c", "#a16207", "#0e7490", "#a21caf", "#047857"},
		[3]string{"#22d3ee", "#eab308", "#dc2626"},
		[8]string{"#c084fc", "#bef264", "#93c5fd", "#fca5a5", "#fde047", "#67e8f9", "#f0abfc", "#6ee7b7"},
		[3]string{"#0e7490", "#fde047", "#f87171"}),
	palette("halloween",
		[8]string{"#c2410c", "#7e22ce", "#4d7c0f", "#15803d", "#a16207", "#b91c1c", "#a21caf", "#0f766e"},
		[3]string{"#84cc16", "#f97316", "#7e22ce"},
		[8]string{"#a3e635", "#c026d3", "#fdba74", "#c4b5fd", "#86efac", "#fcd34d", "#fca5a5", "#67e8f9"},
		[3]string{"#4d7c0f", "#fb923c", "#e879f9"}),
	palette("zombie",
		[8]string{"#c2410c", "#7e22ce", "#4d7c0f", "#15803d", "#be123c", "#0f766e", "#a16207", "#1d4ed8"},
		[3]string{"#84cc16", "#f97316", "#7e22ce"},
		[8]string{"#fdba74", "#a3e635", "#c4b5fd", "#86efac", "#fda4af", "#5eead4", "#fcd34d", "#93c5fd"},
		[3]string{"#65a30d", "#fb923c", "#c084fc"}),
	palette("pastel",
		[8]string{"#be123c", "#c2410c", "#0369a1", "#15803d", "#7e22ce", "#0e7490", "#be185d", "#a16207"},
		[3]string{"#bae6fd", "#fde68a", "#fda4af"},
		[8]string{"#fda4af", "#fdba74", "#7dd3fc", "#86efac", "#c4b5fd", "#67e8f9", "#f9a8d4", "#fde68a"},
		[3]string{"#0c4a6e", "#fcd34d", "#fda4af"}),
	palette("90s",
		[8]string{"#7e22ce", "#0369a1", "#0e7490", "#0f766e", "#a16207", "#be185d", "#4d7c0f", "#c2410c"},
		[3]string{"#06b6d4", "#eab308", "#db2777"},
		[8]string{"#d8b4fe", "#7dd3fc", "#67e8f9", "#5eead4", "#fde047", "#f9a8d4", "#bef264", "#fdba74"},
		[3]string{"#0e7490", "#fde047", "#f472b6"}),
	palette("christmas",
		[8]string{"#b91c1c", "#047857", "#a16207", "#15803d", "#be123c", "#0f766e", "#475569", "#c2410c"},
		[3]string{"#10b981", "#f59e0b", "#dc2626"},
		[8]string{"#fca5a5", "#6ee7b7", "#fcd34d", "#86efac", "#fda4af", "#5eead4", "#cbd5e1", "#fdba74"},
		[3]string{"#047857", "#fbbf24", "#f87171"}),
	palette("prototype",
		[8]string{"#111827", "#404040", "#334155", "#3f3f46", "#44403c", "#374151", "#525252", "#475569"},
		[3]string{"#e5e7eb", "#6b7280", "#111827"},
		[8]string{"#f9fafb", "#d4d4d4", "#cbd5e1", "#d4d4d8", "#d6d3d1", "#d1d5db", "#a3a3a3", "#94a3b8"},
		[3]string{"#1f2937", "#9ca3af", "#f3f4f6"}),
	palette("news",
		[8]string{"#0369a1", "#18181b", "#b91c1c", "#15803d", "#a16207", "#6d28d9", "#0f766e", "#be123c"},
		[3]string{"#bae6fd", "#0284c7", "#0c4a6e"},
		[8]string{"#7dd3fc", "#fafafa", "#fca5a5", "#86efac", "#fcd34d", "#c4b5fd", "#5eead4", "#fda4af"},
		[3]string{"#0c4a6e", "#38bdf8", "#bae6fd"}),
	palette("industrial",
		[8]string{"#a16207", "#1c1917", "#1d4ed8", "#c2410c", "#0f766e", "#b91c1c", "#3f3f46", "#4d7c0f"},
		[3]string{"#475569", "#d97706", "#b91c1c"},
		[8]string{"#fcd34d", "#d6d3d1", "#93c5fd", "#fdba74", "#5eead4", "#fca5a5", "#d4d4d8", "#bef264"},
		[3]string{"#64748b", "#fbbf24", "#f87171"}),
	palette("dracula",
		[8]string{"#6d28d9", "#be185d", "#0e7490", "#15803d", "#a16207", "#c2410c", "#b91c1c", "#44475a"},
		[3]string{"#8be9fd", "#f1fa8c", "#ff5555"},
		[8]string{"#bd93f9", "#ff79c6", "#8be9fd", "#50fa7b", "#f1fa8c", "#ffb86c", "#ff5555", "#f8f8f2"},
		[3]string{"#6272a4", "#f1fa8c", "#ff5555"}),
}

type themePalette struct {
	ID          string
	Light, Dark themePaletteMode
}

type themePaletteMode struct {
	Series [8]string
	Scale  [3]string
}

func palette(id string, lightSeries [8]string, lightScale [3]string, darkSeries [8]string, darkScale [3]string) themePalette {
	return themePalette{id, themePaletteMode{lightSeries, lightScale}, themePaletteMode{darkSeries, darkScale}}
}

var generatedThemeCSS = renderThemePalettes(themePalettes)

func renderThemePalettes(themes []themePalette) string {
	var css strings.Builder
	css.WriteString("/* Generated from themePalettes in palettes.go. */\n")
	for _, theme := range themes {
		writeThemeRule(&css, theme.ID, false, theme.Light)
		writeThemeRule(&css, theme.ID, true, theme.Dark)
	}
	return css.String()
}

func writeThemeRule(css *strings.Builder, id string, dark bool, mode themePaletteMode) {
	modeName := "light"
	if dark {
		modeName = "dark"
	}
	fmt.Fprintf(css, "/* goshtoso-charts-theme:%s:%s */\n", id, modeName)
	selectors := themeSelectors(id, dark)
	css.WriteString(strings.Join(selectors, ",\n"))
	css.WriteString(" {\n")
	for _, token := range resolvedThemeTokens(id, dark, mode) {
		fmt.Fprintf(css, "  --color-chart-%s: %s;\n", token.Name, token.Value)
	}
	css.WriteString("}\n")
}

func themeSelectors(id string, dark bool) []string {
	paletteClass := ".goshtoso-charts-palette-auto"
	selectors := []string{}
	if id == "araihu" {
		if dark {
			selectors = append(selectors, `:where(.dark) :where(.goshtoso-charts-palette-araihu)`, `:where(.dark.goshtoso-charts-palette-araihu)`)
		} else {
			selectors = append(selectors, `:where(.goshtoso-charts-palette-araihu)`)
		}
	}
	if dark {
		return append(selectors,
			fmt.Sprintf(`:where(.dark) :where([data-theme="%s"]) :where(%s)`, id, paletteClass),
			fmt.Sprintf(`:where(.dark[data-theme="%s"]) :where(%s)`, id, paletteClass),
			fmt.Sprintf(`:where([data-theme="%s"]) :where(.dark) :where(%s)`, id, paletteClass),
			fmt.Sprintf(`:where(.dark[data-theme="%s"]%s)`, id, paletteClass),
		)
	}
	return append(selectors,
		fmt.Sprintf(`:where([data-theme="%s"]) :where(%s)`, id, paletteClass),
		fmt.Sprintf(`:where([data-theme="%s"]%s)`, id, paletteClass),
	)
}

type themeToken struct{ Name, Value string }

func resolvedThemeTokens(themeID string, dark bool, mode themePaletteMode) []themeToken {
	surface, surfaceAlt := "var(--color-surface, #ffffff)", "var(--color-surface-alt, var(--color-surface, #ffffff))"
	outline, text := "var(--color-outline, #cbd5e1)", "var(--color-on-surface, #334155)"
	strong := "var(--color-on-surface-strong, var(--color-on-surface, #0f172a))"
	muted := "var(--color-on-surface-muted, color-mix(in srgb, var(--color-on-surface, #334155) 72%, var(--color-surface, #ffffff) 28%))"
	stateStrong := "var(--color-on-surface-strong, #111827)"
	patternMix := "88%"
	patternOutlineMix := "12%"
	if dark {
		surface, surfaceAlt = "var(--color-surface-dark, #111827)", "var(--color-surface-dark-alt, var(--color-surface-dark, #111827))"
		outline, text = "var(--color-outline-dark, #475569)", "var(--color-on-surface-dark, #cbd5e1)"
		strong = "var(--color-on-surface-dark-strong, var(--color-on-surface-dark, #f8fafc))"
		muted = "var(--color-on-surface-dark-muted, color-mix(in srgb, var(--color-on-surface-dark, #cbd5e1) 72%, var(--color-surface-dark, #111827) 28%))"
		stateStrong = "var(--color-on-surface-dark-strong, #f9fafb)"
		patternMix = "82%"
		patternOutlineMix = "18%"
	} else if themeID == "minimal" {
		// Minimal intentionally removes generic surface outlines. Charts still need
		// an opaque grid and pattern boundary, so reuse its control outline.
		outline = "var(--color-control-outline, var(--color-outline-strong, #737373))"
	}
	tokens := []themeToken{
		{"surface", surface}, {"surface-alt", surfaceAlt}, {"outline", outline},
		{"grid", fmt.Sprintf("color-mix(in srgb, %s 65%%, %s 35%%)", outline, surface)},
		{"axis", fmt.Sprintf("color-mix(in srgb, %s 82%%, %s 18%%)", outline, text)},
		{"foreground", text}, {"foreground-strong", strong}, {"foreground-muted", muted},
		// text-* stays as a compatibility spelling for existing charts and callers.
		{"text", "var(--color-chart-foreground)"}, {"text-strong", "var(--color-chart-foreground-strong)"}, {"text-muted", "var(--color-chart-foreground-muted)"},
		{"pattern-text", "var(--color-chart-text-strong)"},
		{"pattern-surface", fmt.Sprintf("color-mix(in srgb, var(--color-chart-surface) %s, var(--color-chart-outline) %s)", patternMix, patternOutlineMix)},
		{"pattern-outline", "var(--color-chart-outline)"},
	}
	for index, color := range mode.Series {
		tokens = append(tokens, themeToken{fmt.Sprintf("series-%d", index+1), color})
	}
	tokens = append(tokens, derivedSeriesTokens()...)
	tokens = append(tokens,
		themeToken{"scale-low", mode.Scale[0]}, themeToken{"scale-mid", mode.Scale[1]}, themeToken{"scale-high", mode.Scale[2]},
	)
	tokens = append(tokens, semanticTokens(stateStrong)...)
	tokens = append(tokens, sequentialTokens()...)
	tokens = append(tokens, divergingTokens()...)
	tokens = append(tokens,
		themeToken{"increasing", "var(--color-chart-success)"},
		themeToken{"decreasing", "var(--color-chart-danger)"},
		themeToken{"bollinger-upper", mode.Series[0]}, themeToken{"bollinger-middle", mode.Series[3]}, themeToken{"bollinger-lower", mode.Series[4]},
	)
	return tokens
}

func derivedSeriesTokens() []themeToken {
	return []themeToken{
		{"series-9", "color-mix(in srgb, var(--color-chart-series-1) 72%, var(--color-chart-series-5) 28%)"},
		{"series-10", "color-mix(in srgb, var(--color-chart-series-2) 70%, var(--color-chart-series-6) 30%)"},
		{"series-11", "color-mix(in srgb, var(--color-chart-series-3) 68%, var(--color-chart-series-7) 32%)"},
		{"series-12", "color-mix(in srgb, var(--color-chart-series-4) 66%, var(--color-chart-series-8) 34%)"},
	}
}

func semanticTokens(strong string) []themeToken {
	semantic := func(name string) themeToken {
		return themeToken{name, fmt.Sprintf("color-mix(in srgb, var(--color-%s) 78%%, %s 22%%)", name, strong)}
	}
	return []themeToken{
		semantic("success"),
		semantic("warning"),
		semantic("danger"),
		semantic("info"),
	}
}

func sequentialTokens() []themeToken {
	return []themeToken{
		{"sequential-1", "color-mix(in srgb, var(--color-chart-series-1) 20%, var(--color-chart-surface) 80%)"},
		{"sequential-2", "color-mix(in srgb, var(--color-chart-series-1) 40%, var(--color-chart-surface) 60%)"},
		{"sequential-3", "color-mix(in srgb, var(--color-chart-series-1) 60%, var(--color-chart-surface) 40%)"},
		{"sequential-4", "color-mix(in srgb, var(--color-chart-series-1) 80%, var(--color-chart-surface) 20%)"},
		{"sequential-5", "var(--color-chart-series-1)"},
	}
}

func divergingTokens() []themeToken {
	return []themeToken{
		{"diverging-1", "var(--color-chart-scale-low)"},
		{"diverging-2", "color-mix(in srgb, var(--color-chart-scale-low) 50%, var(--color-chart-scale-mid) 50%)"},
		{"diverging-3", "var(--color-chart-scale-mid)"},
		{"diverging-4", "color-mix(in srgb, var(--color-chart-scale-mid) 50%, var(--color-chart-scale-high) 50%)"},
		{"diverging-5", "var(--color-chart-scale-high)"},
	}
}
