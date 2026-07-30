package charttheme

import (
	"bytes"
	"context"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestStylePrecedenceAndClasses(t *testing.T) {
	style := Style{Palette: PalettePastel, Colors: []string{"#123456"}, Class: "ring-2 custom-chart"}
	if got := style.SeriesColor(0); got != "#123456" {
		t.Fatalf("explicit color = %q", got)
	}
	if got := style.SeriesColor(1); got != "var(--color-chart-series-2)" {
		t.Fatalf("fallback token = %q", got)
	}
	if got := style.SeriesColor(11); got != "var(--color-chart-series-12)" {
		t.Fatalf("dense fallback token = %q", got)
	}
	if got := style.SeriesColor(12); got != "var(--color-chart-series-1)" {
		t.Fatalf("wrapped fallback token = %q", got)
	}
	if got := style.RootClasses("chart"); got != "chart goshtoso-charts-palette goshtoso-charts-palette-pastel ring-2 custom-chart" {
		t.Fatalf("classes = %q", got)
	}
	colors := style.ResolvedColors()
	if len(colors) != 8 || colors[0] != "#123456" || colors[1] != "#fca5a5" {
		t.Fatalf("resolved colors = %#v", colors)
	}
}

func TestAutoUsesBoldLiteralFallback(t *testing.T) {
	colors := (Style{}).ResolvedColors()
	if len(colors) != 8 || colors[0] != "#2563eb" {
		t.Fatalf("auto fallback = %#v", colors)
	}
	if got := (Style{}).RootClasses("chart"); got != "chart goshtoso-charts-palette goshtoso-charts-palette-auto" {
		t.Fatalf("auto classes = %q", got)
	}
	unknown := Style{Palette: Palette("future-theme")}
	if got := unknown.RootClasses("chart"); got != "chart goshtoso-charts-palette goshtoso-charts-palette-bold" {
		t.Fatalf("unknown palette classes = %q", got)
	}
	if got := unknown.ResolvedColors(); !reflect.DeepEqual(got, palettes[PaletteBold]) {
		t.Fatalf("unknown palette fallback = %#v", got)
	}
}

func TestStatusPaletteUsesSemanticOrderAndRootClass(t *testing.T) {
	t.Parallel()
	style := Style{Palette: PaletteStatus, Colors: []string{"#123456"}, Class: "custom-status-chart"}
	colors := style.ResolvedColors()
	if len(colors) != 8 || colors[0] != "#123456" || colors[1] != "#d97706" || colors[2] != "#dc2626" {
		t.Fatalf("status colors = %#v", colors)
	}
	if got := style.RootClasses("chart"); got != "chart goshtoso-charts-palette goshtoso-charts-palette-status custom-status-chart" {
		t.Fatalf("status classes = %q", got)
	}
}

func TestStylesExposeLightAndDarkSemanticChartTokens(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	if err := Styles().Render(context.Background(), &output); err != nil {
		t.Fatalf("Styles().Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`--color-chart-surface: var(--color-surface)`,
		`--color-chart-series-1: var(--color-blue-600, #2563eb)`,
		`--goshtoso-charts-series-1: var(--color-chart-series-1)`,
		`--goshtoso-charts-surface: var(--color-chart-surface)`,
		`--color-chart-surface-alt: var(--color-surface-alt, var(--color-surface))`,
		`--color-chart-outline: var(--color-outline)`,
		`--color-chart-grid: color-mix`,
		`--color-chart-axis: color-mix`,
		`--color-chart-foreground: var(--color-on-surface)`,
		`--color-chart-foreground-strong: var(--color-on-surface-strong`,
		`--color-chart-foreground-muted: var(--color-on-surface-muted`,
		`--color-chart-text: var(--color-chart-foreground)`,
		`--color-chart-text-strong: var(--color-chart-foreground-strong)`,
		`--color-chart-text-muted: var(--color-chart-foreground-muted)`,
		`--color-chart-scale-low:`,
		`--color-chart-scale-mid:`,
		`--color-chart-scale-high:`,
		`--color-chart-scale-low: var(--color-cyan-300, #67e8f9)`,
		`--color-chart-scale-mid: var(--color-amber-400, #fbbf24)`,
		`--color-chart-scale-high: var(--color-red-600, #dc2626)`,
		`--color-chart-scale-low: var(--color-cyan-700, #0e7490)`,
		`--color-chart-scale-high: var(--color-red-400, #f87171)`,
		`--color-chart-series-12: color-mix`,
		`--color-chart-success: color-mix(in srgb, var(--color-success`,
		`--color-chart-warning: color-mix(in srgb, var(--color-warning`,
		`--color-chart-danger: color-mix(in srgb, var(--color-danger`,
		`--color-chart-info: color-mix(in srgb, var(--color-info`,
		`--color-chart-sequential-1: color-mix`,
		`--color-chart-sequential-5: var(--color-chart-series-1)`,
		`--color-chart-diverging-1: var(--color-chart-scale-low)`,
		`--color-chart-diverging-5: var(--color-chart-scale-high)`,
		`--goshtoso-charts-series-12: var(--color-chart-series-12)`,
		`--goshtoso-charts-success: var(--color-chart-success)`,
		`.goshtoso-charts-palette-status`,
		`--color-chart-series-1: var(--color-chart-success)`,
		`--color-chart-series-2: var(--color-chart-warning)`,
		`--color-chart-series-3: var(--color-chart-danger)`,
		`:where(.dark) :where(.goshtoso-charts-palette-status)`,
		`:where(.dark) :where(.goshtoso-charts-palette)`,
		`--color-chart-surface: var(--color-surface-dark)`,
		`--color-chart-foreground-strong: var(--color-on-surface-dark-strong`,
		`--color-chart-foreground-muted: var(--color-on-surface-dark-muted`,
		`:where([data-theme="araihu"]) :where(.goshtoso-charts-palette-auto)`,
		`--color-chart-series-1: #4d7c0f`,
		`--color-chart-series-1: #c7ff4a`,
		`--color-chart-scale-low: #38bdf8`,
		`--color-chart-scale-mid: #f59e0b`,
		`--color-chart-scale-high: #e11d48`,
		`--color-chart-scale-high: #fb7185`,
		`--color-chart-scale-mid: var(--color-amber-200`,
		`.goshtoso-charts-scatter__viewport`,
		`min-width: 36rem`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("theme styles missing %q", want)
		}
	}
}

func TestAraiHuAutoPaletteSplitsLightAndDarkSeriesOneWithoutChangingSemanticPalettes(t *testing.T) {
	t.Parallel()
	var output strings.Builder
	if err := Styles().Render(context.Background(), &output); err != nil {
		t.Fatalf("Styles().Render() error = %v", err)
	}
	markup := output.String()
	lightStart := strings.Index(markup, `/* goshtoso-charts-theme:araihu:light */`)
	darkStart := strings.Index(markup, `/* goshtoso-charts-theme:araihu:dark */`)
	if lightStart == -1 || darkStart == -1 || darkStart <= lightStart {
		t.Fatal("AraiHu light/dark palette rules are missing or unordered")
	}
	lightRule := markup[lightStart:darkStart]
	darkRuleEnd := strings.Index(markup[darkStart:], `}`)
	if darkRuleEnd == -1 {
		t.Fatal("AraiHu dark palette rule is incomplete")
	}
	darkRule := markup[darkStart : darkStart+darkRuleEnd]
	if !strings.Contains(lightRule, `--color-chart-series-1: #4d7c0f`) || strings.Contains(lightRule, `--color-chart-series-1: #c7ff4a`) {
		t.Fatalf("AraiHu light series-1 token is not contrast-safe: %q", lightRule)
	}
	if !strings.Contains(darkRule, `--color-chart-series-1: #c7ff4a`) || strings.Contains(darkRule, `--color-chart-series-1: #4d7c0f`) {
		t.Fatalf("AraiHu dark series-1 token does not restore bright lime: %q", darkRule)
	}
	for _, unchanged := range []string{
		`--color-chart-series-2: #ff8a3d`,
		`--color-chart-scale-low: #38bdf8`,
		`--color-chart-scale-mid: #f59e0b`,
		`--color-chart-scale-high: #e11d48`,
		`--color-chart-success: color-mix(in srgb, var(--color-success)`,
		`--color-chart-warning: color-mix(in srgb, var(--color-warning)`,
		`--color-chart-danger: color-mix(in srgb, var(--color-danger)`,
	} {
		if !strings.Contains(markup, unchanged) {
			t.Errorf("unrelated semantic palette token changed or missing: %q", unchanged)
		}
	}
}

func TestThemePaletteCatalogIsExactCompleteAndUnique(t *testing.T) {
	t.Parallel()
	wantIDs := []string{
		"araihu", "goshtoso", "arctic", "high-contrast", "minimal", "modern",
		"neo-brutalism", "halloween", "zombie", "pastel", "90s", "christmas",
		"prototype", "news", "industrial", "dracula",
	}
	gotIDs := make([]string, 0, len(themePalettes))
	seenIDs := map[string]bool{}
	seenModes := map[string]string{}
	for _, theme := range themePalettes {
		if seenIDs[theme.ID] {
			t.Fatalf("duplicate theme %q", theme.ID)
		}
		seenIDs[theme.ID] = true
		gotIDs = append(gotIDs, theme.ID)
		for name, mode := range map[string]themePaletteMode{"light": theme.Light, "dark": theme.Dark} {
			seenColors := map[string]bool{}
			for index, color := range mode.Series {
				if strings.TrimSpace(color) == "" {
					t.Fatalf("%s:%s series-%d is blank", theme.ID, name, index+1)
				}
				if seenColors[color] {
					t.Fatalf("%s:%s repeats categorical color %q", theme.ID, name, color)
				}
				seenColors[color] = true
			}
			for index, color := range mode.Scale {
				if strings.TrimSpace(color) == "" {
					t.Fatalf("%s:%s scale-%d is blank", theme.ID, name, index)
				}
			}
			signature := strings.Join(mode.Series[:], "|") + ":" + strings.Join(mode.Scale[:], "|")
			if previous, exists := seenModes[name+":"+signature]; exists {
				t.Fatalf("%s:%s reuses %s palette", theme.ID, name, previous)
			}
			seenModes[name+":"+signature] = theme.ID
		}
	}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("theme IDs = %v, want %v", gotIDs, wantIDs)
	}
}

func TestGeneratedThemeCSSCoversFullContractAndTemplateIsCurrent(t *testing.T) {
	t.Parallel()
	contract := resolvedThemeTokens(themePalettes[0].ID, false, themePalettes[0].Light)
	wantNames := []string{
		"surface", "surface-alt", "outline", "grid", "axis",
		"foreground", "foreground-strong", "foreground-muted",
		"text", "text-strong", "text-muted",
		"pattern-text", "pattern-surface", "pattern-outline",
		"series-1", "series-2", "series-3", "series-4", "series-5", "series-6",
		"series-7", "series-8", "series-9", "series-10", "series-11", "series-12",
		"scale-low", "scale-mid", "scale-high",
		"success", "warning", "danger", "info",
		"sequential-1", "sequential-2", "sequential-3", "sequential-4", "sequential-5",
		"diverging-1", "diverging-2", "diverging-3", "diverging-4", "diverging-5",
		"increasing", "decreasing", "bollinger-upper", "bollinger-middle", "bollinger-lower",
	}
	if len(contract) != len(wantNames) {
		t.Fatalf("chart token contract size = %d, want %d", len(contract), len(wantNames))
	}
	gotNames := make([]string, len(contract))
	seenNames := map[string]bool{}
	for index, token := range contract {
		gotNames[index] = token.Name
		if seenNames[token.Name] {
			t.Fatalf("duplicate chart token %q", token.Name)
		}
		seenNames[token.Name] = true
	}
	if !reflect.DeepEqual(gotNames, wantNames) {
		t.Fatalf("chart token names = %v, want %v", gotNames, wantNames)
	}
	for _, theme := range themePalettes {
		for _, mode := range []struct {
			name   string
			dark   bool
			values themePaletteMode
		}{{"light", false, theme.Light}, {"dark", true, theme.Dark}} {
			marker := "/* goshtoso-charts-theme:" + theme.ID + ":" + mode.name + " */"
			start := strings.Index(generatedThemeCSS, marker)
			if start == -1 {
				t.Fatalf("missing generated rule %s:%s", theme.ID, mode.name)
			}
			end := strings.Index(generatedThemeCSS[start:], "}\n")
			if end == -1 {
				t.Fatalf("incomplete generated rule %s:%s", theme.ID, mode.name)
			}
			rule := generatedThemeCSS[start : start+end]
			resolved := resolvedThemeTokens(theme.ID, mode.dark, mode.values)
			seen := map[string]bool{}
			for _, token := range resolved {
				if seen[token.Name] {
					t.Fatalf("%s:%s duplicates %s", theme.ID, mode.name, token.Name)
				}
				seen[token.Name] = true
				if strings.TrimSpace(token.Value) == "" {
					t.Fatalf("%s:%s has blank %s", theme.ID, mode.name, token.Name)
				}
				if !strings.Contains(rule, "--color-chart-"+token.Name+":") {
					t.Errorf("%s:%s missing %s", theme.ID, mode.name, token.Name)
				}
			}
			if len(seen) != len(wantNames) {
				t.Fatalf("%s:%s token count = %d, want %d", theme.ID, mode.name, len(seen), len(wantNames))
			}
		}
	}
	var output strings.Builder
	if err := Styles().Render(context.Background(), &output); err != nil {
		t.Fatalf("Styles().Render() error = %v", err)
	}
	if !strings.Contains(output.String(), generatedThemeCSS) {
		t.Fatal("styles_templ.go drifted from generated theme CSS injection")
	}
}

func TestKnownThemesMapNamedSemanticsAndDistinctLightDarkModes(t *testing.T) {
	t.Parallel()
	for _, theme := range themePalettes {
		light := tokenValues(resolvedThemeTokens(theme.ID, false, theme.Light))
		dark := tokenValues(resolvedThemeTokens(theme.ID, true, theme.Dark))
		for _, semantic := range []string{"success", "warning", "danger", "info"} {
			want := "var(--color-" + semantic + ")"
			if !strings.Contains(light[semantic], want) || !strings.Contains(dark[semantic], want) {
				t.Errorf("%s semantic %s is not mapped by name: light=%q dark=%q", theme.ID, semantic, light[semantic], dark[semantic])
			}
		}
		for _, token := range []string{"surface", "axis", "foreground", "series-1", "scale-low"} {
			if light[token] == dark[token] {
				t.Errorf("%s %s does not distinguish light and dark modes: %q", theme.ID, token, light[token])
			}
		}
	}
}

func TestFallbackAndTailwindVariableDiscoveryAreDocumented(t *testing.T) {
	t.Parallel()
	var output strings.Builder
	if err := Styles().Render(context.Background(), &output); err != nil {
		t.Fatalf("Styles().Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`--color-chart-success: color-mix(in srgb, var(--color-success, #16a34a)`,
		`--color-chart-info: color-mix(in srgb, var(--color-info, #0891b2)`,
		`--color-chart-series-12:`,
		`--color-chart-diverging-5:`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("fallback CSS missing %q", want)
		}
	}
	readme, err := os.ReadFile("../../README.md")
	if err != nil {
		t.Fatalf("read README: %v", err)
	}
	for _, want := range []string{
		`text-(--color-chart-series-1)`,
		`bg-(--color-chart-surface)`,
		`text-(--color-chart-success)`,
		`bg-(--color-chart-diverging-5)`,
	} {
		if !strings.Contains(string(readme), want) {
			t.Errorf("Tailwind source-discovery documentation missing %q", want)
		}
	}
}

func tokenValues(tokens []themeToken) map[string]string {
	values := make(map[string]string, len(tokens))
	for _, token := range tokens {
		values[token.Name] = token.Value
	}
	return values
}

func TestMinimalLightChartOutlineUsesOpaqueControlBoundary(t *testing.T) {
	t.Parallel()
	values := map[string]string{}
	for _, token := range resolvedThemeTokens("minimal", false, themePaletteMode{}) {
		values[token.Name] = token.Value
	}
	if got := values["outline"]; got != "var(--color-control-outline, var(--color-outline-strong, #737373))" {
		t.Fatalf("minimal light chart outline = %q", got)
	}
}

func TestThemeSelectorsKeepCallerTokenOverridesAuthoritative(t *testing.T) {
	t.Parallel()
	for _, theme := range themePalettes {
		for _, dark := range []bool{false, true} {
			for _, selector := range themeSelectors(theme.ID, dark) {
				if !strings.HasPrefix(selector, ":where(") {
					t.Errorf("theme selector has nonzero specificity: %q", selector)
				}
			}
		}
	}
	if strings.Contains(generatedThemeCSS, "!important") {
		t.Fatal("generated palette CSS must not defeat caller overrides")
	}
}

func TestStylesCenterInteractiveRendererHostWithoutSizingIt(t *testing.T) {
	t.Parallel()
	var output strings.Builder
	if err := Styles().Render(context.Background(), &output); err != nil {
		t.Fatalf("Styles().Render() error = %v", err)
	}

	markup := output.String()
	containerRuleStart := strings.Index(markup, `.goshtoso-charts-interactive > .container`)
	if containerRuleStart == -1 {
		t.Fatal("shared interactive renderer container selector is missing")
	}
	containerRuleEnd := strings.Index(markup[containerRuleStart:], `}`)
	if containerRuleEnd == -1 {
		t.Fatal("shared interactive renderer container rule is incomplete")
	}
	containerRule := markup[containerRuleStart : containerRuleStart+containerRuleEnd]
	if !strings.Contains(containerRule, `overflow-x: auto`) {
		t.Fatalf("shared renderer container must contain narrow-screen overflow: %q", containerRule)
	}

	ruleStart := strings.Index(markup, `.goshtoso-charts-interactive > .container > .item`)
	if ruleStart == -1 {
		t.Fatal("shared interactive renderer-host selector is missing")
	}
	ruleEnd := strings.Index(markup[ruleStart:], `}`)
	if ruleEnd == -1 {
		t.Fatal("shared interactive renderer-host rule is incomplete")
	}
	rule := markup[ruleStart : ruleStart+ruleEnd]
	if !strings.Contains(rule, `margin-inline: auto`) {
		t.Fatalf("shared renderer-host rule does not center the host: %q", rule)
	}
	for _, forbidden := range []string{`width:`, `max-width:`, `min-width:`} {
		if strings.Contains(rule, forbidden) {
			t.Fatalf("shared renderer-host rule must preserve configured dimensions; found %q in %q", forbidden, rule)
		}
	}
}
