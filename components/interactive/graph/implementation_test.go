package graph

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math"
	"regexp"
	"strings"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

var graphChartIDPattern = regexp.MustCompile(`id="([A-Za-z]{12})"`)

func TestGraphNormalizedRenderHashes(t *testing.T) {
	t.Parallel()
	x, y := 10.0, 20.0
	tests := []struct {
		name   string
		config Config
		want   string
	}{
		{name: "force defaults", config: Config{Label: "Network", Nodes: []Node{{Name: "only"}}}, want: "552555471008948e30acf95868b7addb2df5bd147c251c5f556c283ebeb13a94"},
		{name: "configured force and wrapper", config: Config{
			Label: "Service dependencies", Caption: "Runtime calls between services.",
			Nodes: []Node{
				{Name: "api", Value: 12, X: &x, Y: &y, Category: "service", Size: 28, Fixed: chart.Bool(true), ItemStyle: &chart.ItemStyle{Color: "#abcdef"}},
				{Name: "database", Value: 8, Category: "storage", Size: 20},
			},
			Links: []Link{{Source: "api", Target: "database", Value: 3, LineStyle: &chart.LineStyle{Width: 2}}},
			Categories: []Category{
				{Name: "service", Label: &chart.LabelOptions{Show: chart.Bool(true), Position: "right"}},
				{Name: "storage", ItemStyle: &chart.ItemStyle{Color: "#654321"}},
			},
			Roam: RoamEnabled, Force: &ForceOptions{InitialLayout: ForceInitialLayoutCircular, Repulsion: 8000, Gravity: 0.2, EdgeLength: 80},
			Draggable: chart.Bool(true), FocusNodeAdjacency: chart.Bool(true), Width: "720px", Height: "360px",
			Options:       chart.ChartOptions{Title: &chart.TitleOptions{Text: "Topology"}, Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "service dependencies"}},
			SeriesOptions: chart.SeriesOptions{Label: &chart.LabelOptions{Show: chart.Bool(true)}, LineStyle: &chart.LineStyle{Width: 1}},
			Style:         charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
		}, want: "018de7354cafa45ad3928d2c4c07fc79053af1f83e37cb58ac5727e7b0fc9b16"},
		{name: "fixed coordinates", config: Config{
			Label: "Fixed network", Layout: LayoutNone,
			Nodes: []Node{{Name: "left", X: &x, Y: &y, Fixed: chart.Bool(true)}, {Name: "right", X: chart.Float(30), Y: chart.Float(40)}},
			Links: []Link{{Source: "left", Target: "right", Value: 2}},
		}, want: "7fcf95a8441dd5b246b7203fbd30b563a7507648617caf7679f988aec9ba675a"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			digest := sha256.Sum256([]byte(normalizedGraphRender(t, Graph(test.config))))
			if got := hex.EncodeToString(digest[:]); got != test.want {
				t.Errorf("normalized render SHA-256 = %s, want %s", got, test.want)
			}
		})
	}
}

func normalizedGraphRender(t *testing.T, instance Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	match := graphChartIDPattern.FindStringSubmatch(output.String())
	if len(match) != 2 {
		t.Fatalf("rendered markup lacks chart ID: %s", output.String())
	}
	return strings.ReplaceAll(output.String(), match[1], "CHARTID")
}

func TestGraphRendersConfiguredForceGraph(t *testing.T) {
	t.Parallel()
	x, y := 10.0, 20.0
	instance := Graph(Config{
		Label: "Service dependencies", Caption: "Runtime calls between services.",
		Nodes: []Node{
			{Name: "api", Value: 12, X: &x, Y: &y, Category: "service", Size: 28, Fixed: chart.Bool(true), ItemStyle: &chart.ItemStyle{Color: "#abcdef"}},
			{Name: "database", Value: 8, Category: "storage", Size: 20},
		},
		Links: []Link{{Source: "api", Target: "database", Value: 3, LineStyle: &chart.LineStyle{Width: 2}}},
		Categories: []Category{
			{Name: "service", Label: &chart.LabelOptions{Show: chart.Bool(true), Position: "right"}},
			{Name: "storage", ItemStyle: &chart.ItemStyle{Color: "#654321"}},
		},
		Roam: RoamEnabled, Force: &ForceOptions{InitialLayout: ForceInitialLayoutCircular, Repulsion: 8000, Gravity: 0.2, EdgeLength: 80},
		Draggable: chart.Bool(true), FocusNodeAdjacency: chart.Bool(true),
		Width: "720px", Height: "360px",
		Options:       chart.ChartOptions{Title: &chart.TitleOptions{Text: "Topology"}},
		SeriesOptions: chart.SeriesOptions{Label: &chart.LabelOptions{Show: chart.Bool(true)}, LineStyle: &chart.LineStyle{Width: 1}},
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
	for name, layout := range map[string]Layout{"circular": LayoutCircular, "coordinate": LayoutNone} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			instance := Graph(Config{Label: "Network", Layout: layout, Nodes: []Node{{Name: "only"}}})
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
	valid := func() Config {
		return Config{
			Label: "Network", Nodes: []Node{{Name: "api", Category: "service"}, {Name: "db"}},
			Links: []Link{{Source: "api", Target: "db"}}, Categories: []Category{{Name: "service"}},
		}
	}
	tests := map[string]struct {
		mutate    func() Config
		wantError string
	}{
		"missing label":      {func() Config { cfg := valid(); cfg.Label = ""; return cfg }, "graph chart label is required"},
		"unsupported layout": {func() Config { cfg := valid(); cfg.Layout = "grid"; return cfg }, `graph chart layout "grid" is not supported`},
		"unsupported roam":   {func() Config { cfg := valid(); cfg.Roam = 2; return cfg }, "graph chart roam mode 2 is not supported"},
		"force with other layout": {func() Config {
			cfg := valid()
			cfg.Layout = LayoutCircular
			cfg.Force = &ForceOptions{}
			return cfg
		}, "graph chart force options require force layout"},
		"bad force init":        {func() Config { cfg := valid(); cfg.Force = &ForceOptions{InitialLayout: "grid"}; return cfg }, `graph chart force initial layout "grid" is not supported`},
		"bad force repulsion":   {func() Config { cfg := valid(); cfg.Force = &ForceOptions{Repulsion: math.NaN()}; return cfg }, "graph chart force repulsion must be finite and nonnegative"},
		"bad force gravity":     {func() Config { cfg := valid(); cfg.Force = &ForceOptions{Gravity: 2}; return cfg }, "graph chart force gravity must be finite and between 0 and 1"},
		"bad force edge length": {func() Config { cfg := valid(); cfg.Force = &ForceOptions{EdgeLength: -1}; return cfg }, "graph chart force edge length must be finite and nonnegative"},
		"missing category name": {func() Config { cfg := valid(); cfg.Categories[0].Name = ""; return cfg }, "graph chart category 0 name is required"},
		"duplicate category": {func() Config {
			cfg := valid()
			cfg.Categories = append(cfg.Categories, Category{Name: "service"})
			return cfg
		}, `graph chart category "service" is duplicated`},
		"missing nodes":        {func() Config { cfg := valid(); cfg.Nodes = nil; return cfg }, "graph chart nodes are required"},
		"missing node name":    {func() Config { cfg := valid(); cfg.Nodes[0].Name = ""; return cfg }, "graph chart node 0 name is required"},
		"duplicate node":       {func() Config { cfg := valid(); cfg.Nodes[1].Name = "api"; return cfg }, `graph chart node "api" is duplicated`},
		"nonfinite node value": {func() Config { cfg := valid(); cfg.Nodes[0].Value = math.Inf(1); return cfg }, `graph chart node "api" value must be finite`},
		"unknown category":     {func() Config { cfg := valid(); cfg.Nodes[0].Category = "queue"; return cfg }, `graph chart node "api" category "queue" is not defined`},
		"missing link source":  {func() Config { cfg := valid(); cfg.Links[0].Source = ""; return cfg }, "graph chart link 0 source is required"},
		"unknown link source":  {func() Config { cfg := valid(); cfg.Links[0].Source = "web"; return cfg }, `graph chart link 0 source "web" is not defined`},
		"missing link target":  {func() Config { cfg := valid(); cfg.Links[0].Target = ""; return cfg }, "graph chart link 0 target is required"},
		"unknown link target":  {func() Config { cfg := valid(); cfg.Links[0].Target = "cache"; return cfg }, `graph chart link 0 target "cache" is not defined`},
		"nonfinite link value": {func() Config { cfg := valid(); cfg.Links[0].Value = math.NaN(); return cfg }, "graph chart link 0 value must be finite"},
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
