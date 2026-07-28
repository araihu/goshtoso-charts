package pages

type attribution struct {
	Name       string
	Version    string
	ProjectURL string
	License    string
	LicenseURL string
	UsedFor    string
}

var foundationAttributions = []attribution{
	{Name: "Goshtoso", Version: "v0.0.13", ProjectURL: "https://github.com/araihu/goshtoso", License: "MIT", LicenseURL: "https://github.com/araihu/goshtoso/blob/v0.0.13/LICENSE", UsedFor: "UI components, theme tokens, and shared browser assets."},
	{Name: "Goshtoso App Shells", Version: "commit 4c4aa5ae787e", ProjectURL: "https://github.com/araihu/goshtoso-app-shells", License: "MIT", LicenseURL: "https://github.com/araihu/goshtoso-app-shells/blob/4c4aa5ae787e/LICENSE", UsedFor: "Documentation shell, categorized navigation, page structure, and responsive layout."},
	{Name: "templ", Version: "v0.3.1020", ProjectURL: "https://github.com/a-h/templ", License: "MIT", LicenseURL: "https://github.com/a-h/templ/blob/v0.3.1020/LICENSE", UsedFor: "Type-safe Go templates for chart components and documentation pages."},
}

var chartAttributions = []attribution{
	{Name: "go-analyze/charts", Version: "v0.6.0", ProjectURL: "https://github.com/go-analyze/charts", License: "MIT", LicenseURL: "https://github.com/go-analyze/charts/blob/v0.6.0/LICENSE", UsedFor: "Go-native SVG rendering behind static/vector components. Documentation adapts examples/1-Painter/scatter_chart-1-basic/main.go and examples/1-Painter/radar_chart-1-basic/main.go."},
	{Name: "go-echarts", Version: "v2.7.2", ProjectURL: "https://github.com/go-echarts/go-echarts", License: "MIT", LicenseURL: "https://github.com/go-echarts/go-echarts/blob/v2.7.2/LICENSE", UsedFor: "Private Go adapter behind renderer-neutral interactive components. Sunburst documentation adapts examples/sunburst.go."},
}

var runtimeAttributions = []attribution{
	{Name: "Apache ECharts", Version: "v5.4.3", ProjectURL: "https://echarts.apache.org/", License: "Apache-2.0", LicenseURL: "https://github.com/apache/echarts/blob/5.4.3/LICENSE", UsedFor: "Pinned browser runtime bundled for local interactive-chart rendering."},
}
