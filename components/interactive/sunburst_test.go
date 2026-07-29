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

func TestSunburstRendersTypedHierarchyOptionsAndExactValues(t *testing.T) {
	t.Parallel()
	instance := Sunburst(SunburstConfig{
		Label: "Basic sunburst example", Caption: "Seven parent and child pairs.",
		Nodes: []*SunburstNode{
			{
				Name: "parent-0", Value: 0.81,
				ItemStyle: &ItemStyle{Color: "#123456", BorderColor: "#ffffff", BorderWidth: 2},
				Label:     &LabelOptions{Show: Bool(true), Position: "inside", Color: "#fedcba"},
				Children:  []*SunburstNode{{Name: "child-0", Value: 0.34}},
			},
			{Name: "parent-1", Value: 0, Children: []*SunburstNode{{Name: "child-1", Value: 0.57}}},
		},
		Navigation:        SunburstNavigationDrillDown,
		Sort:              SunburstSortAscending,
		LabelOptions:      &LabelOptions{Show: Bool(true), Position: "inside", FontSize: 10},
		ItemStyle:         &ItemStyle{BorderColor: "#eeeeee", BorderWidth: 1},
		ShowLabelsForZero: Bool(true),
		InnerRadius:       16, OuterRadius: 88,
		Width: "100%", Height: "32rem",
		Options: ChartOptions{Title: &TitleOptions{Text: "Basic sunburst example"}, Animation: Bool(false)},
		Style: charttheme.Style{
			Palette: charttheme.PaletteAraiHu,
			Colors:  []string{"#654321"},
			Class:   "rounded-radius max-w-full",
		},
		RootAttrs: templ.Attributes{"id": "basic-sunburst", "data-chart-purpose": "hierarchy"},
	})

	if instance.Kind() != chartcomponents.KindInteractiveSunburst {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	markup := renderSunburst(t, instance)
	for _, want := range []string{
		`<figure class="goshtoso-charts-interactive goshtoso-charts-palette goshtoso-charts-palette-araihu rounded-radius max-w-full"`,
		`role="img"`, `aria-label="Basic sunburst example"`, `id="basic-sunburst"`, `data-chart-purpose="hierarchy"`,
		`width:100%;height:32rem`,
		`"type":"sunburst"`, `"nodeClick":"rootToNode"`, `"sort":"asc"`, `"renderLabelForZeroData":true`,
		`"radius":["16%","88%"]`,
		`"name":"parent-0","value":0.81`, `"name":"child-0","value":0.34`,
		`"name":"parent-1","value":0`, `"name":"child-1","value":0.57`,
		`"label":{"show":true,"fontSize":10,"position":"inside"}`, `"itemStyle":{"color":"#123456","borderColor":"#ffffff","borderWidth":2}`,
		`"text":"Basic sunburst example"`, `"animation":false`, `"color":["#654321","#ff8a3d"`,
		`data-goshtoso-charts-theme-runtime`, `Seven parent and child pairs.`,
		`data-goshtoso-chart-expand`, `exportFromMenu($el, &#34;png&#34;)`, `>Export</span>`,
		`<details class="mt-3 max-w-full`, `>Exact hierarchy and values</summary>`,
		`scope="col">Path</th>`, `scope="col">Parent</th>`, `scope="col">Value</th>`,
		`parent-0 / child-0`, `>parent-0</td>`, `>0.34</td>`, `>0</td>`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestSunburstKeepsNavigationAndOrderingVariantsOnOneKind(t *testing.T) {
	t.Parallel()
	instance := Sunburst(SunburstConfig{
		Label: "Fixed hierarchy", Nodes: []*SunburstNode{{Name: "root", Value: 1}},
		Navigation: SunburstNavigationDisabled, Sort: SunburstSortInput,
	})
	if instance.Kind() != chartcomponents.KindInteractiveSunburst {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	markup := renderSunburst(t, instance)
	for _, want := range []string{`"nodeClick":false`, `"sort":null`} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestSunburstDefaultsToDrillDownAndDescendingOrder(t *testing.T) {
	t.Parallel()
	markup := renderSunburst(t, Sunburst(SunburstConfig{
		Label: "Hierarchy", Nodes: []*SunburstNode{{Name: "root", Value: 1}},
	}))
	for _, want := range []string{`"nodeClick":"rootToNode"`, `"sort":"desc"`, `"radius":["0%","75%"]`} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing default %q", want)
		}
	}
}

func TestSunburstRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	valid := func() SunburstConfig {
		return SunburstConfig{Label: "Hierarchy", Nodes: []*SunburstNode{{Name: "parent", Value: 1, Children: []*SunburstNode{{Name: "child", Value: 0.5}}}}}
	}
	cycle := &SunburstNode{Name: "cycle", Value: 1}
	cycle.Children = []*SunburstNode{cycle}
	shared := &SunburstNode{Name: "shared", Value: 1}
	deep := &SunburstNode{Name: "depth-0", Value: 1}
	cursor := deep
	for depth := 1; depth <= maxSunburstDepth+1; depth++ {
		cursor.Children = []*SunburstNode{{Name: "depth-" + strconv.Itoa(depth), Value: 1}}
		cursor = cursor.Children[0]
	}
	tests := map[string]struct {
		mutate    func() SunburstConfig
		wantError string
	}{
		"missing label":   {func() SunburstConfig { cfg := valid(); cfg.Label = " "; return cfg }, "sunburst chart label is required"},
		"missing nodes":   {func() SunburstConfig { cfg := valid(); cfg.Nodes = nil; return cfg }, "sunburst chart nodes are required"},
		"nil root":        {func() SunburstConfig { cfg := valid(); cfg.Nodes[0] = nil; return cfg }, "sunburst chart root 0 is nil"},
		"empty node":      {func() SunburstConfig { cfg := valid(); cfg.Nodes[0].Name = " "; return cfg }, "sunburst chart root 0 name is required"},
		"negative value":  {func() SunburstConfig { cfg := valid(); cfg.Nodes[0].Value = -1; return cfg }, `sunburst chart node "parent" value must be nonnegative`},
		"nonfinite value": {func() SunburstConfig { cfg := valid(); cfg.Nodes[0].Value = math.NaN(); return cfg }, `sunburst chart node "parent" value must be finite`},
		"nil child":       {func() SunburstConfig { cfg := valid(); cfg.Nodes[0].Children[0] = nil; return cfg }, `sunburst chart node "parent" child 0 is nil`},
		"cycle":           {func() SunburstConfig { cfg := valid(); cfg.Nodes = []*SunburstNode{cycle}; return cfg }, `sunburst chart node "cycle" child 0 contains a cycle`},
		"shared node": {func() SunburstConfig {
			cfg := valid()
			cfg.Nodes = []*SunburstNode{{Name: "a", Value: 1, Children: []*SunburstNode{shared}}, {Name: "b", Value: 1, Children: []*SunburstNode{shared}}}
			return cfg
		}, `sunburst chart node "b" child 0 reuses a node`},
		"too deep":       {func() SunburstConfig { cfg := valid(); cfg.Nodes = []*SunburstNode{deep}; return cfg }, "exceeds maximum depth 256"},
		"bad navigation": {func() SunburstConfig { cfg := valid(); cfg.Navigation = "zoom"; return cfg }, `sunburst chart navigation "zoom" is not supported`},
		"bad sort":       {func() SunburstConfig { cfg := valid(); cfg.Sort = "alphabetical"; return cfg }, `sunburst chart sort "alphabetical" is not supported`},
		"bad label size": {func() SunburstConfig { cfg := valid(); cfg.LabelOptions = &LabelOptions{FontSize: -1}; return cfg }, "sunburst chart label font size must be nonnegative"},
		"bad node label": {func() SunburstConfig { cfg := valid(); cfg.Nodes[0].Label = &LabelOptions{FontSize: -1}; return cfg }, `sunburst chart node "parent" label font size must be nonnegative`},
		"bad inner radius": {func() SunburstConfig {
			cfg := valid()
			cfg.InnerRadius = -1
			return cfg
		}, "sunburst chart inner radius must be between 0 and 100"},
		"bad outer radius": {func() SunburstConfig {
			cfg := valid()
			cfg.OuterRadius = 101
			return cfg
		}, "sunburst chart outer radius must be between 0 and 100"},
		"inverted radii": {func() SunburstConfig {
			cfg := valid()
			cfg.InnerRadius = 80
			cfg.OuterRadius = 70
			return cfg
		}, "sunburst chart inner radius must be less than outer radius"},
		"reserved role": {func() SunburstConfig {
			cfg := valid()
			cfg.RootAttrs = templ.Attributes{"role": "presentation"}
			return cfg
		}, `sunburst chart root attribute "role" is reserved`},
		"reserved label": {func() SunburstConfig {
			cfg := valid()
			cfg.RootAttrs = templ.Attributes{"Aria-Label": "override"}
			return cfg
		}, `sunburst chart root attribute "Aria-Label" is reserved`},
		"reserved class": {func() SunburstConfig {
			cfg := valid()
			cfg.RootAttrs = templ.Attributes{"class": "override"}
			return cfg
		}, `sunburst chart root attribute "class" is reserved`},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			err := Sunburst(test.mutate()).Render(context.Background(), &output)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("Render() error = %v, want containing %q", err, test.wantError)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %d bytes for invalid config", output.Len())
			}
		})
	}
}

func renderSunburst(t *testing.T, instance Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return output.String()
}
