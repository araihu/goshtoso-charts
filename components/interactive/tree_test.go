package interactive

import (
	"bytes"
	"context"
	"math"
	"strconv"
	"strings"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func TestTreeRendersConfiguredHierarchy(t *testing.T) {
	t.Parallel()
	instance := Tree(TreeConfig{
		Label: "Service ownership", Caption: "Teams and their services.",
		Roots: []*TreeNode{{
			Name: "Platform", Value: 12, Symbol: TreeSymbolRoundedRectangle, SymbolSize: 24,
			ItemStyle: &ItemStyle{Color: "#abcdef"}, LineStyle: &LineStyle{Color: "#123456", Width: 2},
			Children: []*TreeNode{
				{Name: "Runtime", Value: 7, Collapsed: Bool(true)},
				{Name: "Data", Value: 5, Children: []*TreeNode{{Name: "Database", Value: 3}}},
			},
		}},
		Orientation: TreeOrientationTopToBottom, Roam: TreeRoamEnabled,
		ExpandAndCollapse: Bool(false), InitialDepth: Int(0),
		NodeLabel: &LabelOptions{Show: Bool(true), Position: "top"},
		LeafLabel: &LabelOptions{Show: Bool(true), Position: "bottom"},
		Symbol:    TreeSymbolDiamond, SymbolSize: 18,
		Insets: TreeInsets{Left: "8%", Right: "8%", Top: "12%", Bottom: "12%"},
		Width:  "720px", Height: "420px",
		Options: ChartOptions{
			Title:    &TitleOptions{Text: "Ownership"},
			Controls: chartcontrol.Options{Fullscreen: true, Collapsible: true},
			Export:   &chartcontrol.ExportOptions{Filename: "ownership"},
		},
		Style: charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#654321"}, Class: "overflow-x-auto"},
	})

	if instance.Kind() != chartcomponents.KindInteractiveTree {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		"Service ownership", "Teams and their services.", "width:720px;height:420px",
		`"type":"tree"`, `"layout":"orthogonal"`, `"orient":"TB"`, `"roam":true`,
		`"expandAndCollapse":false`, `"initialTreeDepth":0`,
		`"label":{"show":true,"position":"top"}`, `"leaves":{"label":{"show":true,"position":"bottom"}}`,
		`"name":"Platform","value":12`, `"symbol":"roundRect"`, `"symbolSize":24`,
		`"name":"Runtime","value":7,"collapsed":true`, `"name":"Database","value":3`,
		`"symbol":"diamond"`, `"symbolSize":18`, `"left":"8%"`, `"right":"8%"`,
		`"text":"Ownership"`, `"color":["#654321","#ff8a3d"`,
		`"#abcdef"`, `"#123456"`,
		"goshtoso-charts-palette-araihu overflow-x-auto",
		`data-goshtoso-chart-control="fullscreen"`, `data-goshtoso-chart-control="collapse"`,
		`data-goshtoso-chart-expand`, `data-goshtoso-chart-export="png"`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	if strings.Contains(markup, strconv.Itoa(treeInitialDepthZeroSentinel)) {
		t.Fatal("rendered markup leaked initial-depth sentinel")
	}
	if strings.Contains(markup, `data-goshtoso-chart-export="svg"`) {
		t.Fatal("interactive tree exposed unsupported SVG export")
	}
	if strings.Contains(markup, `data-goshtoso-chart-export-menu`) {
		t.Fatal("interactive tree rendered dropdown for one format")
	}
}

func TestTreeRendersRadialAndExpandedVariantsWithSameKind(t *testing.T) {
	t.Parallel()
	instance := Tree(TreeConfig{
		Label: "Organization", Roots: []*TreeNode{{Name: "Company"}},
		Layout: TreeLayoutRadial, InitialDepth: Int(-1),
	})
	if instance.Kind() != chartcomponents.KindInteractiveTree {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	for _, want := range []string{`"layout":"radial"`, `"orient":"LR"`, `"initialTreeDepth":-1`, `"left":"14%"`, `"right":"14%"`, `"top":"12%"`, `"bottom":"12%"`, `"animationDuration":150`, `"animationDurationUpdate":100`} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestTreeMapsLayeredOrientationVariants(t *testing.T) {
	t.Parallel()
	for name, test := range map[string]struct {
		orientation TreeOrientation
		rendered    string
	}{
		"left to right": {TreeOrientationLeftToRight, "LR"},
		"right to left": {TreeOrientationRightToLeft, "RL"},
		"top to bottom": {TreeOrientationTopToBottom, "TB"},
		"bottom to top": {TreeOrientationBottomToTop, "BT"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			instance := Tree(TreeConfig{
				Label: "Hierarchy", Roots: []*TreeNode{{Name: "Root"}},
				Orientation: test.orientation,
			})
			var output bytes.Buffer
			if err := instance.Render(context.Background(), &output); err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if want := `"orient":"` + test.rendered + `"`; !strings.Contains(output.String(), want) {
				t.Errorf("rendered markup missing %q", want)
			}
		})
	}
}

func TestTreeRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	valid := func() TreeConfig {
		return TreeConfig{Label: "Hierarchy", Roots: []*TreeNode{{Name: "Root", Children: []*TreeNode{{Name: "Leaf"}}}}}
	}
	cycle := &TreeNode{Name: "Cycle"}
	cycle.Children = []*TreeNode{cycle}
	shared := &TreeNode{Name: "Shared"}
	deep := &TreeNode{Name: "depth-0"}
	cursor := deep
	for depth := 1; depth <= maxTreeDepth+1; depth++ {
		cursor.Children = []*TreeNode{{Name: "depth-" + strconv.Itoa(depth)}}
		cursor = cursor.Children[0]
	}
	tests := map[string]struct {
		mutate    func() TreeConfig
		wantError string
	}{
		"missing label":   {func() TreeConfig { cfg := valid(); cfg.Label = " "; return cfg }, "tree chart label is required"},
		"bad layout":      {func() TreeConfig { cfg := valid(); cfg.Layout = "cluster"; return cfg }, `tree chart layout "cluster" is not supported`},
		"bad orientation": {func() TreeConfig { cfg := valid(); cfg.Orientation = "sideways"; return cfg }, `tree chart orientation "sideways" is not supported`},
		"radial orientation": {func() TreeConfig {
			cfg := valid()
			cfg.Layout = TreeLayoutRadial
			cfg.Orientation = TreeOrientationTopToBottom
			return cfg
		}, "tree chart orientation requires layered layout"},
		"bad roam":            {func() TreeConfig { cfg := valid(); cfg.Roam = 2; return cfg }, "tree chart roam mode 2 is not supported"},
		"bad initial depth":   {func() TreeConfig { cfg := valid(); cfg.InitialDepth = Int(-2); return cfg }, "tree chart initial depth must be -1 or nonnegative"},
		"bad chart symbol":    {func() TreeConfig { cfg := valid(); cfg.Symbol = "star"; return cfg }, `tree chart symbol "star" is not supported`},
		"negative chart size": {func() TreeConfig { cfg := valid(); cfg.SymbolSize = -1; return cfg }, "tree chart symbol size must be nonnegative"},
		"missing roots":       {func() TreeConfig { cfg := valid(); cfg.Roots = nil; return cfg }, "tree chart roots are required"},
		"nil root":            {func() TreeConfig { cfg := valid(); cfg.Roots[0] = nil; return cfg }, "tree chart root 0 is nil"},
		"empty node name":     {func() TreeConfig { cfg := valid(); cfg.Roots[0].Children[0].Name = " "; return cfg }, `tree chart node "Root" child 0 name is required`},
		"nonfinite value":     {func() TreeConfig { cfg := valid(); cfg.Roots[0].Value = math.NaN(); return cfg }, `tree chart node "Root" value must be finite`},
		"bad node symbol":     {func() TreeConfig { cfg := valid(); cfg.Roots[0].Symbol = "hexagon"; return cfg }, `tree chart node "Root" symbol "hexagon" is not supported`},
		"negative node size":  {func() TreeConfig { cfg := valid(); cfg.Roots[0].SymbolSize = -1; return cfg }, `tree chart node "Root" symbol size must be nonnegative`},
		"nil child":           {func() TreeConfig { cfg := valid(); cfg.Roots[0].Children[0] = nil; return cfg }, `tree chart node "Root" child 0 is nil`},
		"cycle":               {func() TreeConfig { cfg := valid(); cfg.Roots = []*TreeNode{cycle}; return cfg }, `tree chart node "Cycle" child 0 contains a cycle`},
		"shared node": {func() TreeConfig {
			cfg := valid()
			cfg.Roots = []*TreeNode{{Name: "A", Children: []*TreeNode{shared}}, {Name: "B", Children: []*TreeNode{shared}}}
			return cfg
		}, `tree chart node "B" child 0 reuses a node`},
		"too deep": {func() TreeConfig { cfg := valid(); cfg.Roots = []*TreeNode{deep}; return cfg }, `exceeds maximum depth 256`},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			err := Tree(test.mutate()).Render(context.Background(), &output)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("Render() error = %v, want containing %q", err, test.wantError)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %d bytes for invalid config", output.Len())
			}
		})
	}
}
