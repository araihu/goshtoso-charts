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

func TestSankeyRendersConfiguredChart(t *testing.T) {
	t.Parallel()
	depth := 1
	instance := Sankey(SankeyConfig{
		Label: "Energy flow", Caption: "Energy moving from generation to demand.",
		Layout: SankeyLayout{Orientation: SankeyOrientationVertical, Alignment: SankeyAlignmentRight, NodeWidth: 24, NodeGap: 12},
		Series: []SankeySeries{{
			Name: "Energy",
			Nodes: []SankeyNode{
				{Name: "Solar", ItemStyle: &ItemStyle{Color: "#f59e0b"}},
				{Name: "Homes", Depth: &depth},
			},
			Links:   []SankeyLink{{Source: "Solar", Target: "Homes", Value: 42.5}},
			Options: SeriesOptions{Animation: Bool(false)},
		}},
		Width: "720px", Height: "420px",
		Options:       ChartOptions{Title: &TitleOptions{Text: "Energy balance"}},
		SeriesOptions: SeriesOptions{Label: &LabelOptions{Show: Bool(true), Position: "right"}, LineStyle: &LineStyle{Color: "source", Opacity: Float(0.6)}},
		Style:         charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-96"},
	})

	if instance.Kind() != chartcomponents.KindInteractiveSankey {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		"Energy flow", "Energy moving from generation to demand.", "width:720px;height:420px",
		`"name":"Energy"`, `"name":"Solar","itemStyle":{"color":"#f59e0b"}`, `"name":"Homes","depth":1`,
		`"source":"Solar","target":"Homes","value":42.5`, `"orient":"vertical"`, `"nodeAlign":"right"`,
		`"nodeWidth":24`, `"nodeGap":12`, `"show":true`, `"position":"right"`, `"color":"source"`, `"opacity":0.6`,
		`"animation":false`, `"text":"Energy balance"`, `"color":["#123456","#ff8a3d"`,
		"goshtoso-charts-palette-araihu min-h-96",
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestSankeyRendersDefaultLayout(t *testing.T) {
	t.Parallel()
	instance := Sankey(validSankeyConfig())

	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	for _, want := range []string{`"orient":"horizontal"`, `"nodeAlign":"justify"`} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestSankeyRejectsInvalidDataContract(t *testing.T) {
	t.Parallel()
	negativeDepth := -1
	tests := map[string]struct {
		mutate    func(SankeyConfig) SankeyConfig
		wantError string
	}{
		"missing label":       {func(cfg SankeyConfig) SankeyConfig { cfg.Label = ""; return cfg }, "sankey chart label is required"},
		"bad orientation":     {func(cfg SankeyConfig) SankeyConfig { cfg.Layout.Orientation = "diagonal"; return cfg }, `sankey chart orientation "diagonal" is not supported`},
		"bad alignment":       {func(cfg SankeyConfig) SankeyConfig { cfg.Layout.Alignment = "center"; return cfg }, `sankey chart alignment "center" is not supported`},
		"negative node width": {func(cfg SankeyConfig) SankeyConfig { cfg.Layout.NodeWidth = -1; return cfg }, "sankey chart node width must be nonnegative"},
		"negative node gap":   {func(cfg SankeyConfig) SankeyConfig { cfg.Layout.NodeGap = -1; return cfg }, "sankey chart node gap must be nonnegative"},
		"missing series":      {func(cfg SankeyConfig) SankeyConfig { cfg.Series = nil; return cfg }, "sankey chart series is required"},
		"missing series name": {func(cfg SankeyConfig) SankeyConfig { cfg.Series[0].Name = ""; return cfg }, "sankey chart series 0 name is required"},
		"missing nodes":       {func(cfg SankeyConfig) SankeyConfig { cfg.Series[0].Nodes = nil; return cfg }, `sankey chart series "Flow" nodes are required`},
		"missing node name":   {func(cfg SankeyConfig) SankeyConfig { cfg.Series[0].Nodes[0].Name = ""; return cfg }, `sankey chart series "Flow" node 0 name is required`},
		"duplicate node":      {func(cfg SankeyConfig) SankeyConfig { cfg.Series[0].Nodes[1].Name = "Input"; return cfg }, `sankey chart series "Flow" node "Input" is duplicated`},
		"negative depth":      {func(cfg SankeyConfig) SankeyConfig { cfg.Series[0].Nodes[0].Depth = &negativeDepth; return cfg }, `sankey chart series "Flow" node "Input" depth must be nonnegative`},
		"missing links":       {func(cfg SankeyConfig) SankeyConfig { cfg.Series[0].Links = nil; return cfg }, `sankey chart series "Flow" links are required`},
		"missing source":      {func(cfg SankeyConfig) SankeyConfig { cfg.Series[0].Links[0].Source = ""; return cfg }, `sankey chart series "Flow" link 0 source is required`},
		"missing target":      {func(cfg SankeyConfig) SankeyConfig { cfg.Series[0].Links[0].Target = ""; return cfg }, `sankey chart series "Flow" link 0 target is required`},
		"unknown source":      {func(cfg SankeyConfig) SankeyConfig { cfg.Series[0].Links[0].Source = "Other"; return cfg }, `sankey chart series "Flow" link 0 source "Other" does not name a node`},
		"unknown target":      {func(cfg SankeyConfig) SankeyConfig { cfg.Series[0].Links[0].Target = "Other"; return cfg }, `sankey chart series "Flow" link 0 target "Other" does not name a node`},
		"negative value":      {func(cfg SankeyConfig) SankeyConfig { cfg.Series[0].Links[0].Value = -1; return cfg }, `sankey chart series "Flow" link 0 value must be a finite nonnegative value`},
		"nonfinite value":     {func(cfg SankeyConfig) SankeyConfig { cfg.Series[0].Links[0].Value = math.NaN(); return cfg }, `sankey chart series "Flow" link 0 value must be a finite nonnegative value`},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			instance := Sankey(test.mutate(validSankeyConfig()))
			if instance.Kind() != chartcomponents.KindInteractiveSankey {
				t.Fatalf("Kind() = %q", instance.Kind())
			}
			var output bytes.Buffer
			err := instance.Render(context.Background(), &output)
			if err == nil || err.Error() != test.wantError {
				t.Fatalf("Render() error = %v, want %q", err, test.wantError)
			}
			if output.Len() != 0 {
				t.Fatalf("Render() wrote %d bytes for invalid config", output.Len())
			}
		})
	}
}

func validSankeyConfig() SankeyConfig {
	return SankeyConfig{
		Label: "Flow",
		Series: []SankeySeries{{
			Name:  "Flow",
			Nodes: []SankeyNode{{Name: "Input"}, {Name: "Output"}},
			Links: []SankeyLink{{Source: "Input", Target: "Output", Value: 1}},
		}},
	}
}
