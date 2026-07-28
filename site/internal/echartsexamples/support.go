package echartsexamples

import (
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/render"
	"github.com/go-echarts/go-echarts/v2/types"
)

const supportExamplesSource = "https://github.com/go-echarts/examples/blob/master/examples/"

// SupportExamples ports the upstream theme, renderer, and page-layout
// examples. The renderer and page examples create deterministic chart
// descriptors: Goshtoso mounts RenderSnippet output inside its own document,
// whereas go-echarts' custom renderer and components.Page only render complete
// documents and do not implement RenderSnippet.
var SupportExamples = []Example{
	{Slug: "theme-chalk", Title: "Chalk theme", Group: "Support", Source: supportExamplesSource + "themes.go", Build: func() render.Renderer { return barWithTheme(types.ThemeChalk) }},
	{Slug: "theme-essos", Title: "Essos theme", Group: "Support", Source: supportExamplesSource + "themes.go", Build: func() render.Renderer { return barWithTheme(types.ThemeEssos) }},
	{Slug: "theme-infographic", Title: "Infographic theme", Group: "Support", Source: supportExamplesSource + "themes.go", Build: func() render.Renderer { return barWithTheme(types.ThemeInfographic) }},
	{Slug: "theme-macarons", Title: "Macarons theme", Group: "Support", Source: supportExamplesSource + "themes.go", Build: func() render.Renderer { return barWithTheme(types.ThemeMacarons) }},
	{Slug: "theme-purple-passion", Title: "Purple Passion theme", Group: "Support", Source: supportExamplesSource + "themes.go", Build: func() render.Renderer { return barWithTheme(types.ThemePurplePassion) }},
	{Slug: "theme-roma", Title: "Roma theme", Group: "Support", Source: supportExamplesSource + "themes.go", Build: func() render.Renderer { return barWithTheme(types.ThemeRoma) }},
	{Slug: "theme-romantic", Title: "Romantic theme", Group: "Support", Source: supportExamplesSource + "themes.go", Build: func() render.Renderer { return barWithTheme(types.ThemeRomantic) }},
	{Slug: "theme-shine", Title: "Shine theme", Group: "Support", Source: supportExamplesSource + "themes.go", Build: func() render.Renderer { return barWithTheme(types.ThemeShine) }},
	{Slug: "theme-vintage", Title: "Vintage theme", Group: "Support", Source: supportExamplesSource + "themes.go", Build: func() render.Renderer { return barWithTheme(types.ThemeVintage) }},
	{Slug: "theme-walden", Title: "Walden theme", Group: "Support", Source: supportExamplesSource + "themes.go", Build: func() render.Renderer { return barWithTheme(types.ThemeWalden) }},
	{Slug: "theme-westeros", Title: "Westeros theme", Group: "Support", Source: supportExamplesSource + "themes.go", Build: func() render.Renderer { return barWithTheme(types.ThemeWesteros) }},
	{Slug: "theme-wonderland", Title: "Wonderland theme", Group: "Support", Source: supportExamplesSource + "themes.go", Build: func() render.Renderer { return barWithTheme(types.ThemeWonderland) }},
	{Slug: "renderer-custom-template", Title: "Custom renderer template", Group: "Support", Source: supportExamplesSource + "renderer.go", Build: func() render.Renderer { return rendererDescriptor("Custom renderer template") }},
	{Slug: "renderer-snippets", Title: "Renderer snippets", Group: "Support", Source: supportExamplesSource + "renderer.go", Build: func() render.Renderer { return rendererDescriptor("Renderer snippets") }},
	{Slug: "page-center-layout", Title: "Page center layout", Group: "Support", Source: supportExamplesSource + "page_center_layout.go", Build: func() render.Renderer { return pageLayoutDescriptor("center") }},
	{Slug: "page-flex-layout", Title: "Page flex layout", Group: "Support", Source: supportExamplesSource + "page_flex_layout.go", Build: func() render.Renderer { return pageLayoutDescriptor("flex") }},
	{Slug: "page-none-layout", Title: "Page no layout", Group: "Support", Source: supportExamplesSource + "page_none_layout.go", Build: func() render.Renderer { return pageLayoutDescriptor("none") }},
}

func barWithTheme(theme string) *charts.Bar {
	bar := charts.NewBar()
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: theme}),
		charts.WithTitleOpts(opts.Title{Title: "bar with " + theme + " theme"}),
	)
	bar.SetXAxis(weeks).AddSeries("Category A", barData(0)).AddSeries("Category B", barData(18))
	return bar
}

func rendererDescriptor(title string) *charts.Bar {
	bar := titledBar(title)
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: title, Subtitle: "Rendered as a Goshtoso chart snippet"}))
	return bar
}

// pageLayoutDescriptor preserves each upstream Page layout selection as a
// renderable demo. A components.Page cannot be used here because its renderer
// emits a complete HTML document and its RenderSnippet method panics.
func pageLayoutDescriptor(layout string) *charts.Bar {
	bar := titledBar("Page " + layout + " layout")
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{Title: "Page " + layout + " layout", Subtitle: "Static descriptor; Goshtoso owns page layout"}))
	return bar
}
