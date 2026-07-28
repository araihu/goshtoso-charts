package interactive

import (
	"bytes"
	"context"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func TestTreemapRendersTypedHierarchyNavigationLevelsAndExactValues(t *testing.T) {
	t.Parallel()
	instance := Treemap(TreemapConfig{
		Label:   "Basic treemap example",
		Caption: "File system usage. Select a directory to focus it; use the breadcrumb to return.",
		Nodes: []*TreemapNode{
			{
				Name: "d1", Class: "directory",
				Children: []*TreemapNode{{Name: "f1", Value: 1000, Class: "large-file", Color: "#123456"}},
			},
			{Name: "f1", Value: 450, Class: "file"},
		},
		Navigation: TreemapNavigationDrillDown,
		Roam:       TreemapRoamEnabled,
		LabelOptions: &LabelOptions{
			Show: Bool(true), Position: "inside", Color: "#ffffff", FontSize: 11,
		},
		UpperLabel: &LabelOptions{Show: Bool(true), FontSize: 12},
		Breadcrumb: &TreemapBreadcrumb{Show: Bool(true), Height: 24, ItemGap: 8},
		NodeStyle:  TreemapNodeStyle{BorderColor: "#ffffff", BorderWidth: 1, GapWidth: 1},
		LeafDepth:  Int(2),
		Levels: []TreemapLevel{
			{
				UpperLabel: &LabelOptions{Show: Bool(true)},
				NodeStyle:  TreemapNodeStyle{BorderColor: "#777777", BorderWidth: 1, GapWidth: 1},
			},
			{
				NodeStyle: TreemapNodeStyle{BorderColor: "#666666", BorderWidth: 2, GapWidth: 1},
			},
			{
				NodeStyle:       TreemapNodeStyle{GapWidth: 1},
				ColorSaturation: &TreemapColorRange{Min: 0.35, Max: 0.5},
			},
		},
		Width:  "100%",
		Height: "500px",
		Options: ChartOptions{
			Title:     &TitleOptions{Text: "Basic treemap example", Subtitle: "File system usage", Left: "center"},
			Legend:    &LegendOptions{Show: Bool(false)},
			Tooltip:   &TooltipOptions{Show: Bool(true), Trigger: "item"},
			Animation: Bool(false),
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
		`data-goshtoso-chart-expand`, `data-goshtoso-chart-export="png"`,
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
	markup := renderTreemap(t, Treemap(TreemapConfig{
		Label: "Hierarchy", Nodes: []*TreemapNode{{Name: "file", Value: 1}},
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
	instance := Treemap(TreemapConfig{
		Label:      "Fixed hierarchy",
		Nodes:      []*TreemapNode{{Name: "file", Value: 1}},
		Navigation: TreemapNavigationDisabled,
		Roam:       TreemapRoamDisabled,
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
	valid := func() TreemapConfig {
		return TreemapConfig{
			Label: "Hierarchy",
			Nodes: []*TreemapNode{{
				Name:     "directory",
				Children: []*TreemapNode{{Name: "file", Value: 1}},
			}},
		}
	}
	cycle := &TreemapNode{Name: "cycle"}
	cycle.Children = []*TreemapNode{cycle}
	shared := &TreemapNode{Name: "shared", Value: 1}
	deep := &TreemapNode{Name: "depth-0"}
	cursor := deep
	for depth := 1; depth <= maxTreemapDepth+1; depth++ {
		cursor.Children = []*TreemapNode{{Name: "depth-" + strconv.Itoa(depth)}}
		cursor = cursor.Children[0]
	}
	tests := map[string]struct {
		mutate    func() TreemapConfig
		wantError string
	}{
		"missing label":   {func() TreemapConfig { cfg := valid(); cfg.Label = " "; return cfg }, "treemap chart label is required"},
		"missing nodes":   {func() TreemapConfig { cfg := valid(); cfg.Nodes = nil; return cfg }, "treemap chart nodes are required"},
		"nil root":        {func() TreemapConfig { cfg := valid(); cfg.Nodes[0] = nil; return cfg }, "treemap chart root 0 is nil"},
		"empty node":      {func() TreemapConfig { cfg := valid(); cfg.Nodes[0].Name = " "; return cfg }, "treemap chart root 0 name is required"},
		"negative value":  {func() TreemapConfig { cfg := valid(); cfg.Nodes[0].Children[0].Value = -1; return cfg }, `treemap chart node "file" value must be nonnegative`},
		"nonfinite value": {func() TreemapConfig { cfg := valid(); cfg.Nodes[0].Children[0].Value = math.NaN(); return cfg }, `treemap chart node "file" value must be finite`},
		"parent value":    {func() TreemapConfig { cfg := valid(); cfg.Nodes[0].Value = 2; return cfg }, `treemap chart parent node "directory" value must be zero; child values determine parent area`},
		"nil child":       {func() TreemapConfig { cfg := valid(); cfg.Nodes[0].Children[0] = nil; return cfg }, `treemap chart node "directory" child 0 is nil`},
		"cycle":           {func() TreemapConfig { cfg := valid(); cfg.Nodes = []*TreemapNode{cycle}; return cfg }, `treemap chart node "cycle" child 0 contains a cycle`},
		"shared node": {func() TreemapConfig {
			cfg := valid()
			cfg.Nodes = []*TreemapNode{{Name: "a", Children: []*TreemapNode{shared}}, {Name: "b", Children: []*TreemapNode{shared}}}
			return cfg
		}, `treemap chart node "b" child 0 reuses a node`},
		"too deep":       {func() TreemapConfig { cfg := valid(); cfg.Nodes = []*TreemapNode{deep}; return cfg }, "exceeds maximum depth 256"},
		"bad navigation": {func() TreemapConfig { cfg := valid(); cfg.Navigation = "link"; return cfg }, `treemap chart navigation "link" is not supported`},
		"bad roam":       {func() TreemapConfig { cfg := valid(); cfg.Roam = 2; return cfg }, "treemap chart roam mode 2 is not supported"},
		"bad leaf depth": {func() TreemapConfig { cfg := valid(); cfg.LeafDepth = Int(0); return cfg }, "treemap chart leaf depth must be positive"},
		"bad label size": {func() TreemapConfig { cfg := valid(); cfg.LabelOptions = &LabelOptions{FontSize: -1}; return cfg }, "treemap chart label font size must be nonnegative"},
		"bad breadcrumb height": {func() TreemapConfig {
			cfg := valid()
			cfg.Breadcrumb = &TreemapBreadcrumb{Height: -1}
			return cfg
		}, "treemap chart breadcrumb height must be nonnegative"},
		"bad border width": {func() TreemapConfig {
			cfg := valid()
			cfg.NodeStyle.BorderWidth = -1
			return cfg
		}, "treemap chart border width must be nonnegative"},
		"bad saturation": {func() TreemapConfig {
			cfg := valid()
			cfg.Levels = []TreemapLevel{{ColorSaturation: &TreemapColorRange{Min: 0.8, Max: 0.2}}}
			return cfg
		}, "treemap chart level 0 color saturation range must be between 0 and 1 with min not greater than max"},
		"reserved role": {func() TreemapConfig {
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
