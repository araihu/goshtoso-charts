package interactive

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func TestGraphRendersConfiguredForceGraph(t *testing.T) {
	t.Parallel()
	x, y := 10.0, 20.0
	instance := Graph(GraphConfig{
		Label: "Service dependencies", Caption: "Runtime calls between services.",
		Nodes: []Node{
			{Name: "api", Value: 12, X: &x, Y: &y, Category: "service", Size: 28, Fixed: Bool(true), ItemStyle: &ItemStyle{Color: "#abcdef"}},
			{Name: "database", Value: 8, Category: "storage", Size: 20},
		},
		Links: []Link{{Source: "api", Target: "database", Value: 3, LineStyle: &LineStyle{Width: 2}}},
		Categories: []Category{
			{Name: "service", Label: &LabelOptions{Show: Bool(true), Position: "right"}},
			{Name: "storage", ItemStyle: &ItemStyle{Color: "#654321"}},
		},
		Roam: GraphRoamEnabled, Force: &ForceOptions{InitialLayout: ForceInitialLayoutCircular, Repulsion: 8000, Gravity: 0.2, EdgeLength: 80},
		Draggable: Bool(true), FocusNodeAdjacency: Bool(true),
		Width: "720px", Height: "360px",
		Options:       ChartOptions{Title: &TitleOptions{Text: "Topology"}},
		SeriesOptions: SeriesOptions{Label: &LabelOptions{Show: Bool(true)}, LineStyle: &LineStyle{Width: 1}},
		Style:         charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
	})

	if instance.Kind() != chartcomponents.KindInteractiveGraph {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		"Service dependencies", "Runtime calls between services.", "width:720px;height:360px",
		`"name":"Service dependencies"`, `"layout":"force"`, `"roam":true`, `"draggable":true`, `"focusNodeAdjacency":true`,
		`"force":{"initLayout":"circular","repulsion":8000,"gravity":0.2,"edgeLength":80}`,
		`"name":"api","x":10,"y":20,"value":12,"fixed":true,"category":"service","symbolSize":28`,
		`"source":"api","target":"database","value":3`, `"name":"storage"`,
		`"text":"Topology"`, `"color":["#123456","#ff8a3d"`, "goshtoso-charts-palette-araihu min-h-80",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestGraphRendersCircularAndCoordinateLayoutsWithSameKind(t *testing.T) {
	t.Parallel()
	for name, layout := range map[string]GraphLayout{"circular": GraphLayoutCircular, "coordinate": GraphLayoutNone} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			instance := Graph(GraphConfig{Label: "Network", Layout: layout, Nodes: []Node{{Name: "only"}}})
			if instance.Kind() != chartcomponents.KindInteractiveGraph {
				t.Fatalf("Kind() = %q", instance.Kind())
			}
			var output bytes.Buffer
			if err := instance.Render(context.Background(), &output); err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			if want := `"layout":"` + string(layout) + `"`; !strings.Contains(output.String(), want) {
				t.Errorf("rendered markup missing %q", want)
			}
		})
	}
}

func TestGraphRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	valid := func() GraphConfig {
		return GraphConfig{
			Label: "Network", Nodes: []Node{{Name: "api", Category: "service"}, {Name: "db"}},
			Links: []Link{{Source: "api", Target: "db"}}, Categories: []Category{{Name: "service"}},
		}
	}
	tests := map[string]struct {
		mutate    func() GraphConfig
		wantError string
	}{
		"missing label":      {func() GraphConfig { cfg := valid(); cfg.Label = ""; return cfg }, "graph chart label is required"},
		"unsupported layout": {func() GraphConfig { cfg := valid(); cfg.Layout = "grid"; return cfg }, `graph chart layout "grid" is not supported`},
		"unsupported roam":   {func() GraphConfig { cfg := valid(); cfg.Roam = 2; return cfg }, "graph chart roam mode 2 is not supported"},
		"force with other layout": {func() GraphConfig {
			cfg := valid()
			cfg.Layout = GraphLayoutCircular
			cfg.Force = &ForceOptions{}
			return cfg
		}, "graph chart force options require force layout"},
		"bad force init":        {func() GraphConfig { cfg := valid(); cfg.Force = &ForceOptions{InitialLayout: "grid"}; return cfg }, `graph chart force initial layout "grid" is not supported`},
		"bad force repulsion":   {func() GraphConfig { cfg := valid(); cfg.Force = &ForceOptions{Repulsion: math.NaN()}; return cfg }, "graph chart force repulsion must be finite and nonnegative"},
		"bad force gravity":     {func() GraphConfig { cfg := valid(); cfg.Force = &ForceOptions{Gravity: 2}; return cfg }, "graph chart force gravity must be finite and between 0 and 1"},
		"bad force edge length": {func() GraphConfig { cfg := valid(); cfg.Force = &ForceOptions{EdgeLength: -1}; return cfg }, "graph chart force edge length must be finite and nonnegative"},
		"missing category name": {func() GraphConfig { cfg := valid(); cfg.Categories[0].Name = ""; return cfg }, "graph chart category 0 name is required"},
		"duplicate category": {func() GraphConfig {
			cfg := valid()
			cfg.Categories = append(cfg.Categories, Category{Name: "service"})
			return cfg
		}, `graph chart category "service" is duplicated`},
		"missing nodes":        {func() GraphConfig { cfg := valid(); cfg.Nodes = nil; return cfg }, "graph chart nodes are required"},
		"missing node name":    {func() GraphConfig { cfg := valid(); cfg.Nodes[0].Name = ""; return cfg }, "graph chart node 0 name is required"},
		"duplicate node":       {func() GraphConfig { cfg := valid(); cfg.Nodes[1].Name = "api"; return cfg }, `graph chart node "api" is duplicated`},
		"nonfinite node value": {func() GraphConfig { cfg := valid(); cfg.Nodes[0].Value = math.Inf(1); return cfg }, `graph chart node "api" value must be finite`},
		"unknown category":     {func() GraphConfig { cfg := valid(); cfg.Nodes[0].Category = "queue"; return cfg }, `graph chart node "api" category "queue" is not defined`},
		"missing link source":  {func() GraphConfig { cfg := valid(); cfg.Links[0].Source = ""; return cfg }, "graph chart link 0 source is required"},
		"unknown link source":  {func() GraphConfig { cfg := valid(); cfg.Links[0].Source = "web"; return cfg }, `graph chart link 0 source "web" is not defined`},
		"missing link target":  {func() GraphConfig { cfg := valid(); cfg.Links[0].Target = ""; return cfg }, "graph chart link 0 target is required"},
		"unknown link target":  {func() GraphConfig { cfg := valid(); cfg.Links[0].Target = "cache"; return cfg }, `graph chart link 0 target "cache" is not defined`},
		"nonfinite link value": {func() GraphConfig { cfg := valid(); cfg.Links[0].Value = math.NaN(); return cfg }, "graph chart link 0 value must be finite"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			err := Graph(test.mutate()).Render(context.Background(), &output)
			if err == nil || err.Error() != test.wantError {
				t.Fatalf("Render() error = %v, want %q", err, test.wantError)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %d bytes for invalid config", output.Len())
			}
		})
	}
}
