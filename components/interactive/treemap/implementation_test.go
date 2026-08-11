package treemap

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

var (
	treemapChartIDPattern = regexp.MustCompile(`id="([A-Za-z]{12})"`)
	treemapScriptPattern  = regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`)
)

func TestTreemapNormalizedRenderHashes(t *testing.T) {
	t.Parallel()
	variants := []struct {
		name         string
		config       Config
		fullDigest   string
		scriptDigest string
		shellDigest  string
	}{
		{
			name: "default hierarchy", config: Config{Label: "Hierarchy", Nodes: []*Node{{Name: "file", Value: 1}}},
			fullDigest: "c93aaaa70d47a1b0239a7c675c542752c9b792d3e855d6a6076493e47eeb59fb", scriptDigest: "9948b422ce6329716efd43501ed522a3d45a6c9a6398b45cb44fc4ce1b3aecf8", shellDigest: "ea8f4ae51f093d245c14eb61fa411f766613375607afda3d4bfaed49356829b4",
		},
		{
			name: "configured hierarchy navigation and exact values",
			config: Config{
				Label: "Basic treemap example", Caption: "File system usage.",
				Nodes:      []*Node{{Name: "directory", Class: "directory", Children: []*Node{{Name: "file", Value: 1000, Class: "large-file", Color: "#123456"}}}},
				Navigation: NavigationDrillDown, Roam: RoamEnabled,
				LabelOptions: &chart.LabelOptions{Show: chart.Bool(true), Position: "inside", Color: "#ffffff", FontSize: 11},
				UpperLabel:   &chart.LabelOptions{Show: chart.Bool(true), FontSize: 12}, Breadcrumb: &Breadcrumb{Show: chart.Bool(true), Height: 24, ItemGap: 8},
				NodeStyle: NodeStyle{BorderColor: "#ffffff", BorderWidth: 1, GapWidth: 1}, LeafDepth: chart.Int(2),
				Levels: []Level{{UpperLabel: &chart.LabelOptions{Show: chart.Bool(true)}, NodeStyle: NodeStyle{GapWidth: 1}, ColorSaturation: &ColorRange{Min: 0.35, Max: 0.5}}},
				Width:  "100%", Height: "500px", Options: chart.ChartOptions{Title: &chart.TitleOptions{Text: "Basic treemap example"}, Animation: chart.Bool(false), Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "basic-treemap"}},
				Style: charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#654321"}, Class: "rounded-radius max-w-full"}, RootAttrs: templ.Attributes{"id": "basic-treemap", "data-chart-purpose": "hierarchy"},
			},
			fullDigest: "aeb7ad0eb307bdc0241817b90349a861fde7a3c5ff19b9e65a913124aef70a04", scriptDigest: "971df31efe1f5b00d2abaa615aaa116785b01ccb8679668697926b76d74d91cf", shellDigest: "377a4a4eb53dd9f03db7d1764c12d7e6a6d0bfa894841610fec86da610d04d30",
		},
		{
			name: "navigation disabled wrapper omitted",
			config: Config{
				Label: "Fixed hierarchy", Nodes: []*Node{{Name: "file", Value: 1}}, Navigation: NavigationDisabled, Roam: RoamDisabled,
				Options: chart.ChartOptions{Controls: chartcontrol.Options{Mode: chartcontrol.WrapperModeOmitted}, Export: &chartcontrol.ExportOptions{Filename: "fixed-hierarchy"}},
				Style:   charttheme.Style{Palette: charttheme.PalettePastel, Class: "caller-treemap"}, RootAttrs: templ.Attributes{"data-owner": "consumer"},
			},
			fullDigest: "7dcf5a57b8cb6abeb2a0a8f169039c6cb5cdd1a24076e3a9104b8a2637625cc9", scriptDigest: "761c9e863a7e76e60276bb6fd35e14c8039b6ad50e1b8bbca41bfb67b15a145d", shellDigest: "704af68ae99542879d17052c83d944d0490806380ace0e246a4a470dba2997e7",
		},
	}
	for _, variant := range variants {
		variant := variant
		t.Run(variant.name, func(t *testing.T) {
			t.Parallel()
			rendered := normalizedTreemapRender(t, Treemap(variant.config))
			assertTreemapDigest(t, "full", rendered, variant.fullDigest)
			assertTreemapDigest(t, "scripts", strings.Join(treemapScriptPattern.FindAllString(rendered, -1), "\n"), variant.scriptDigest)
			assertTreemapDigest(t, "shell", treemapScriptPattern.ReplaceAllString(rendered, "<script></script>"), variant.shellDigest)
		})
	}
}

func normalizedTreemapRender(t *testing.T, instance Instance) string {
	t.Helper()
	markup := renderTreemap(t, instance)
	match := treemapChartIDPattern.FindStringSubmatch(markup)
	if len(match) != 2 {
		t.Fatalf("rendered markup lacks chart ID: %s", markup)
	}
	return strings.ReplaceAll(markup, match[1], "CHARTID")
}

func assertTreemapDigest(t *testing.T, name, value, want string) {
	t.Helper()
	digest := sha256.Sum256([]byte(value))
	if got := hex.EncodeToString(digest[:]); got != want {
		t.Errorf("%s SHA-256 = %s, want %s", name, got, want)
	}
}

func TestTreemapRendersTypedHierarchyNavigationLevelsAndExactValues(t *testing.T) {
	t.Parallel()
	instance := Treemap(Config{
		Label:   "Basic treemap example",
		Caption: "File system usage. Select a directory to focus it; use the breadcrumb to return.",
		Nodes: []*Node{
			{
				Name: "d1", Class: "directory",
				Children: []*Node{{Name: "f1", Value: 1000, Class: "large-file", Color: "#123456"}},
			},
			{Name: "f1", Value: 450, Class: "file"},
		},
		Navigation: NavigationDrillDown,
		Roam:       RoamEnabled,
		LabelOptions: &chart.LabelOptions{
			Show: chart.Bool(true), Position: "inside", Color: "#ffffff", FontSize: 11,
		},
		UpperLabel: &chart.LabelOptions{Show: chart.Bool(true), FontSize: 12},
		Breadcrumb: &Breadcrumb{Show: chart.Bool(true), Height: 24, ItemGap: 8},
		NodeStyle:  NodeStyle{BorderColor: "#ffffff", BorderWidth: 1, GapWidth: 1},
		LeafDepth:  chart.Int(2),
		Levels: []Level{
			{
				UpperLabel: &chart.LabelOptions{Show: chart.Bool(true)},
				NodeStyle:  NodeStyle{BorderColor: "#777777", BorderWidth: 1, GapWidth: 1},
			},
			{
				NodeStyle: NodeStyle{BorderColor: "#666666", BorderWidth: 2, GapWidth: 1},
			},
			{
				NodeStyle:       NodeStyle{GapWidth: 1},
				ColorSaturation: &ColorRange{Min: 0.35, Max: 0.5},
			},
		},
		Width:  "100%",
		Height: "500px",
		Options: chart.ChartOptions{
			Title:     &chart.TitleOptions{Text: "Basic treemap example", Subtitle: "File system usage", Left: "center"},
			Legend:    &chart.LegendOptions{Show: chart.Bool(false)},
			Tooltip:   &chart.TooltipOptions{Show: chart.Bool(true), Trigger: "item"},
			Animation: chart.Bool(false),
		},
		Style: charttheme.Style{
			Palette: charttheme.PaletteAraiHu,
			Colors:  []string{"#654321"},
			Class:   "rounded-radius max-w-full",
		},
		RootAttrs: templ.Attributes{"id": "basic-treemap", "data-chart-purpose": "hierarchy"},
	})

	if instance.Kind() != chartcomponents.KindInteractiveTreemap {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	markup := renderTreemap(t, instance)
	for _, want := range []string{
		`<figure class="goshtoso-charts-interactive goshtoso-charts-palette goshtoso-charts-palette-araihu rounded-radius max-w-full"`,
		`role="img"`, `aria-label="Basic treemap example"`, `id="basic-treemap"`, `data-chart-purpose="hierarchy"`,
		`style="width:100%;height:500px;"`,
		`"type":"treemap"`, `"nodeClick":"zoomToNode"`, `"roam":true`, `"leafDepth":2`,
		`"breadcrumb":{"show":true,"height":24,"itemGap":8}`,
		`"upperLabel":{"show":true,"fontSize":12}`,
		`"name":"d1","className":"directory","children":[`,
		`"name":"f1","value":1000,"className":"large-file","itemStyle":{"color":"#123456"}`,
		`"name":"f1","value":450,"className":"file"`,
		`"label":{"show":true`, `"fontSize":11`, `"position":"inside"`, `"color":"#ffffff"`,
		`"itemStyle":{"borderColor":"#ffffff","borderWidth":1,"gapWidth":1}`,
		`"levels":[{"upperLabel":{"show":true},"itemStyle":{"borderColor":"#777777","borderWidth":1,"gapWidth":1}}`,
		`"colorSaturation":[0.35,0.5]`,
		`"text":"Basic treemap example"`, `"subtext":"File system usage"`, `"left":"center"`,
		`"animation":false`, `"color":["#654321","#ff8a3d"`,
		`data-goshtoso-charts-theme-runtime`, `File system usage. Select a directory`,
		`data-goshtoso-chart-expand`, `exportFromMenu($el, &#34;png&#34;)`,
		`data-goshtoso-charts-theme-series-items=""`,
		`<details class="mt-3 max-w-full`, `>Exact hierarchy and values</summary>`,
		`scope="col">Path</th>`, `scope="col">Parent</th>`, `scope="col">Value</th>`, `scope="col">Class</th>`,
		`d1 / f1`, `>d1</td>`, `>1000</td>`, `>large-file</td>`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestTreemapUsesSharedRuntimeWithoutComponentSpecificRuntime(t *testing.T) {
	t.Parallel()
	markup := renderTreemap(t, Treemap(Config{
		Label: "Hierarchy", Nodes: []*Node{{Name: "file", Value: 1}},
	}))
	if strings.Count(markup, `data-goshtoso-charts-theme-runtime`) != 1 {
		t.Fatalf("shared theme runtime count = %d, want 1", strings.Count(markup, `data-goshtoso-charts-theme-runtime`))
	}
	if strings.Contains(markup, `__goshtosoChartsTreemapRuntime`) {
		t.Fatal("treemap rendered duplicated component-specific runtime")
	}
}

func TestTreemapKeepsNavigationAndRoamVariantsOnOneKind(t *testing.T) {
	t.Parallel()
	instance := Treemap(Config{
		Label:      "Fixed hierarchy",
		Nodes:      []*Node{{Name: "file", Value: 1}},
		Navigation: NavigationDisabled,
		Roam:       RoamDisabled,
	})
	if instance.Kind() != chartcomponents.KindInteractiveTreemap {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	markup := renderTreemap(t, instance)
	for _, want := range []string{`"nodeClick":false`, `"roam":false`, `data-goshtoso-charts-theme-series-items=""`} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestTreemapRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	valid := func() Config {
		return Config{
			Label: "Hierarchy",
			Nodes: []*Node{{
				Name:     "directory",
				Children: []*Node{{Name: "file", Value: 1}},
			}},
		}
	}
	cycle := &Node{Name: "cycle"}
	cycle.Children = []*Node{cycle}
	shared := &Node{Name: "shared", Value: 1}
	deep := &Node{Name: "depth-0"}
	cursor := deep
	for depth := 1; depth <= maxTreemapDepth+1; depth++ {
		cursor.Children = []*Node{{Name: "depth-" + strconv.Itoa(depth)}}
		cursor = cursor.Children[0]
	}
	tests := map[string]struct {
		mutate    func() Config
		wantError string
	}{
		"missing label":   {func() Config { cfg := valid(); cfg.Label = " "; return cfg }, "treemap chart label is required"},
		"missing nodes":   {func() Config { cfg := valid(); cfg.Nodes = nil; return cfg }, "treemap chart nodes are required"},
		"nil root":        {func() Config { cfg := valid(); cfg.Nodes[0] = nil; return cfg }, "treemap chart root 0 is nil"},
		"empty node":      {func() Config { cfg := valid(); cfg.Nodes[0].Name = " "; return cfg }, "treemap chart root 0 name is required"},
		"negative value":  {func() Config { cfg := valid(); cfg.Nodes[0].Children[0].Value = -1; return cfg }, `treemap chart node "file" value must be nonnegative`},
		"nonfinite value": {func() Config { cfg := valid(); cfg.Nodes[0].Children[0].Value = math.NaN(); return cfg }, `treemap chart node "file" value must be finite`},
		"parent value":    {func() Config { cfg := valid(); cfg.Nodes[0].Value = 2; return cfg }, `treemap chart parent node "directory" value must be zero; child values determine parent area`},
		"nil child":       {func() Config { cfg := valid(); cfg.Nodes[0].Children[0] = nil; return cfg }, `treemap chart node "directory" child 0 is nil`},
		"cycle":           {func() Config { cfg := valid(); cfg.Nodes = []*Node{cycle}; return cfg }, `treemap chart node "cycle" child 0 contains a cycle`},
		"shared node": {func() Config {
			cfg := valid()
			cfg.Nodes = []*Node{{Name: "a", Children: []*Node{shared}}, {Name: "b", Children: []*Node{shared}}}
			return cfg
		}, `treemap chart node "b" child 0 reuses a node`},
		"too deep":       {func() Config { cfg := valid(); cfg.Nodes = []*Node{deep}; return cfg }, "exceeds maximum depth 256"},
		"bad navigation": {func() Config { cfg := valid(); cfg.Navigation = "link"; return cfg }, `treemap chart navigation "link" is not supported`},
		"bad roam":       {func() Config { cfg := valid(); cfg.Roam = 2; return cfg }, "treemap chart roam mode 2 is not supported"},
		"bad leaf depth": {func() Config { cfg := valid(); cfg.LeafDepth = chart.Int(0); return cfg }, "treemap chart leaf depth must be positive"},
		"bad label size": {func() Config { cfg := valid(); cfg.LabelOptions = &chart.LabelOptions{FontSize: -1}; return cfg }, "treemap chart label font size must be nonnegative"},
		"bad breadcrumb height": {func() Config {
			cfg := valid()
			cfg.Breadcrumb = &Breadcrumb{Height: -1}
			return cfg
		}, "treemap chart breadcrumb height must be nonnegative"},
		"bad border width": {func() Config {
			cfg := valid()
			cfg.NodeStyle.BorderWidth = -1
			return cfg
		}, "treemap chart border width must be nonnegative"},
		"bad saturation": {func() Config {
			cfg := valid()
			cfg.Levels = []Level{{ColorSaturation: &ColorRange{Min: 0.8, Max: 0.2}}}
			return cfg
		}, "treemap chart level 0 color saturation range must be between 0 and 1 with min not greater than max"},
		"reserved role": {func() Config {
			cfg := valid()
			cfg.RootAttrs = templ.Attributes{"role": "presentation"}
			return cfg
		}, `treemap chart root attribute "role" is reserved`},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			err := Treemap(test.mutate()).Render(context.Background(), &output)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("Render() error = %v, want containing %q", err, test.wantError)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %d bytes for invalid config", output.Len())
			}
		})
	}
}

func renderTreemap(t *testing.T, instance Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return output.String()
}
