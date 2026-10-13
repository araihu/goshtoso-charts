package tree

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

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

var (
	treeChartIDPattern = regexp.MustCompile(`id="([A-Za-z]{12})"`)
	treeScriptPattern  = regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`)
)

func TestTreeNormalizedRenderHashes(t *testing.T) {
	t.Parallel()
	variants := []struct {
		name         string
		config       Config
		fullDigest   string
		scriptDigest string
		shellDigest  string
	}{
		{
			name:       "default layered",
			config:     Config{Label: "Hierarchy", Roots: []*Node{{Name: "Root", Children: []*Node{{Name: "Leaf"}}}}},
		fullDigest: "99f0a04281060a4b60cf2256f3846cc0d533568decaefc4ac2eb69f10c2b5f48", scriptDigest: "909d3868304d08732ca65b3d5dae009013d302cd18d8c1d21eff3c58ce9a975f", shellDigest: "e17ab212d714cf92491f0ab4b3633bdb3e6b6030909d5aea604e959d510c1a23",
		},
		{
			name: "collapsed navigation and wrapper",
			config: Config{
				Label: "Service ownership", Caption: "Teams and their services.",
				Roots: []*Node{{
					Name: "Platform", Value: 12, Symbol: SymbolRoundedRectangle, SymbolSize: 24,
					ItemStyle: &chart.ItemStyle{Color: "#abcdef"}, LineStyle: &chart.LineStyle{Color: "#123456", Width: 2},
					Children: []*Node{{Name: "Runtime", Value: 7, Collapsed: chart.Bool(true)}, {Name: "Data", Value: 5, Children: []*Node{{Name: "Database", Value: 3}}}},
				}},
				Orientation: OrientationTopToBottom, Roam: RoamEnabled,
				ExpandAndCollapse: chart.Bool(false), InitialDepth: chart.Int(0),
				NodeLabel: &chart.LabelOptions{Show: chart.Bool(true), Position: "top"}, LeafLabel: &chart.LabelOptions{Show: chart.Bool(true), Position: "bottom"},
				Symbol: SymbolDiamond, SymbolSize: 18, Insets: Insets{Left: "8%", Right: "8%", Top: "12%", Bottom: "12%"},
				Width: "720px", Height: "420px",
				Options: chart.ChartOptions{Title: &chart.TitleOptions{Text: "Ownership"}, Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "ownership"}},
				Style:   charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#654321"}, Class: "overflow-x-auto"},
			},
		fullDigest: "686f3b3e0614d3224710d8ef52afa0a19f944e286c896d516082c98db2dd4bef", scriptDigest: "4f7f59059675696651cf44240216013fbca441564d1c4eaa92a36642c5064ca5", shellDigest: "1d0d57b4eebe38ce80a0f9dbfb8da248d62ef841a93d04a200181398b60562df",
		},
		{
			name: "radial expanded wrapper omitted",
			config: Config{
				Label: "Organization", Roots: []*Node{{Name: "Company", Collapsed: chart.Bool(false), Children: []*Node{{Name: "Team"}}}},
				Layout: LayoutRadial, InitialDepth: chart.Int(-1), ExpandAndCollapse: chart.Bool(true),
				Options: chart.ChartOptions{Animation: chart.Bool(false), Controls: chartcontrol.Options{Mode: chartcontrol.WrapperModeOmitted}, Export: &chartcontrol.ExportOptions{Filename: "organization"}},
				Style:   charttheme.Style{Palette: charttheme.PalettePastel, Class: "caller-tree"},
			},
			fullDigest: "8b4fc7ef84f5c9db05681efd844a420d7e8d6acd0daf84c74bcbd47a9273ff53", scriptDigest: "1518d525b8607ec8c560c92352608035dae09a21f653f373256b9871d48abc56", shellDigest: "0f0fb3724440b69e203d1e4d61a094290ad142abe717def0ba14714dc3b9df00",
		},
	}
	for _, variant := range variants {
		variant := variant
		t.Run(variant.name, func(t *testing.T) {
			t.Parallel()
			rendered := normalizedTreeRender(t, Tree(variant.config))
			assertTreeDigest(t, "full", rendered, variant.fullDigest)
			assertTreeDigest(t, "scripts", strings.Join(treeScriptPattern.FindAllString(rendered, -1), "\n"), variant.scriptDigest)
			assertTreeDigest(t, "shell", treeScriptPattern.ReplaceAllString(rendered, "<script></script>"), variant.shellDigest)
		})
	}
}

func normalizedTreeRender(t *testing.T, instance Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	match := treeChartIDPattern.FindStringSubmatch(output.String())
	if len(match) != 2 {
		t.Fatalf("rendered markup lacks chart ID: %s", output.String())
	}
	return strings.ReplaceAll(output.String(), match[1], "CHARTID")
}

func assertTreeDigest(t *testing.T, name, value, want string) {
	t.Helper()
	digest := sha256.Sum256([]byte(value))
	if got := hex.EncodeToString(digest[:]); got != want {
		t.Errorf("%s SHA-256 = %s, want %s", name, got, want)
	}
}

func TestTreeRendersConfiguredHierarchy(t *testing.T) {
	t.Parallel()
	instance := Tree(Config{
		Label: "Service ownership", Caption: "Teams and their services.",
		Roots: []*Node{{
			Name: "Platform", Value: 12, Symbol: SymbolRoundedRectangle, SymbolSize: 24,
			ItemStyle: &chart.ItemStyle{Color: "#abcdef"}, LineStyle: &chart.LineStyle{Color: "#123456", Width: 2},
			Children: []*Node{
				{Name: "Runtime", Value: 7, Collapsed: chart.Bool(true)},
				{Name: "Data", Value: 5, Children: []*Node{{Name: "Database", Value: 3}}},
			},
		}},
		Orientation: OrientationTopToBottom, Roam: RoamEnabled,
		ExpandAndCollapse: chart.Bool(false), InitialDepth: chart.Int(0),
		NodeLabel: &chart.LabelOptions{Show: chart.Bool(true), Position: "top"},
		LeafLabel: &chart.LabelOptions{Show: chart.Bool(true), Position: "bottom"},
		Symbol:    SymbolDiamond, SymbolSize: 18,
		Insets: Insets{Left: "8%", Right: "8%", Top: "12%", Bottom: "12%"},
		Width:  "720px", Height: "420px",
		Options: chart.ChartOptions{
			Title:    &chart.TitleOptions{Text: "Ownership"},
			Controls: chartcontrol.Options{Fullscreen: true},
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
		`-fullscreen-action`,
		`data-goshtoso-chart-expand`, `exportFromMenu($el, &#34;png&#34;)`,
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
	if !strings.Contains(markup, `-chart-expand-export"`) || !strings.Contains(markup, `>Copy</span>`) {
		t.Fatal("interactive tree omitted the Copy SplitButton")
	}
}

func TestTreeRendersRadialAndExpandedVariantsWithSameKind(t *testing.T) {
	t.Parallel()
	instance := Tree(Config{
		Label: "Organization", Roots: []*Node{{Name: "Company"}},
		Layout: LayoutRadial, InitialDepth: chart.Int(-1),
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
		orientation Orientation
		rendered    string
	}{
		"left to right": {OrientationLeftToRight, "LR"},
		"right to left": {OrientationRightToLeft, "RL"},
		"top to bottom": {OrientationTopToBottom, "TB"},
		"bottom to top": {OrientationBottomToTop, "BT"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			instance := Tree(Config{
				Label: "Hierarchy", Roots: []*Node{{Name: "Root"}},
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
	valid := func() Config {
		return Config{Label: "Hierarchy", Roots: []*Node{{Name: "Root", Children: []*Node{{Name: "Leaf"}}}}}
	}
	cycle := &Node{Name: "Cycle"}
	cycle.Children = []*Node{cycle}
	shared := &Node{Name: "Shared"}
	deep := &Node{Name: "depth-0"}
	cursor := deep
	for depth := 1; depth <= maxTreeDepth+1; depth++ {
		cursor.Children = []*Node{{Name: "depth-" + strconv.Itoa(depth)}}
		cursor = cursor.Children[0]
	}
	tests := map[string]struct {
		mutate    func() Config
		wantError string
	}{
		"missing label":   {func() Config { cfg := valid(); cfg.Label = " "; return cfg }, "tree chart label is required"},
		"bad layout":      {func() Config { cfg := valid(); cfg.Layout = "cluster"; return cfg }, `tree chart layout "cluster" is not supported`},
		"bad orientation": {func() Config { cfg := valid(); cfg.Orientation = "sideways"; return cfg }, `tree chart orientation "sideways" is not supported`},
		"radial orientation": {func() Config {
			cfg := valid()
			cfg.Layout = LayoutRadial
			cfg.Orientation = OrientationTopToBottom
			return cfg
		}, "tree chart orientation requires layered layout"},
		"bad roam":            {func() Config { cfg := valid(); cfg.Roam = 2; return cfg }, "tree chart roam mode 2 is not supported"},
		"bad initial depth":   {func() Config { cfg := valid(); cfg.InitialDepth = chart.Int(-2); return cfg }, "tree chart initial depth must be -1 or nonnegative"},
		"bad chart symbol":    {func() Config { cfg := valid(); cfg.Symbol = "star"; return cfg }, `tree chart symbol "star" is not supported`},
		"negative chart size": {func() Config { cfg := valid(); cfg.SymbolSize = -1; return cfg }, "tree chart symbol size must be nonnegative"},
		"missing roots":       {func() Config { cfg := valid(); cfg.Roots = nil; return cfg }, "tree chart roots are required"},
		"nil root":            {func() Config { cfg := valid(); cfg.Roots[0] = nil; return cfg }, "tree chart root 0 is nil"},
		"empty node name":     {func() Config { cfg := valid(); cfg.Roots[0].Children[0].Name = " "; return cfg }, `tree chart node "Root" child 0 name is required`},
		"nonfinite value":     {func() Config { cfg := valid(); cfg.Roots[0].Value = math.NaN(); return cfg }, `tree chart node "Root" value must be finite`},
		"bad node symbol":     {func() Config { cfg := valid(); cfg.Roots[0].Symbol = "hexagon"; return cfg }, `tree chart node "Root" symbol "hexagon" is not supported`},
		"negative node size":  {func() Config { cfg := valid(); cfg.Roots[0].SymbolSize = -1; return cfg }, `tree chart node "Root" symbol size must be nonnegative`},
		"nil child":           {func() Config { cfg := valid(); cfg.Roots[0].Children[0] = nil; return cfg }, `tree chart node "Root" child 0 is nil`},
		"cycle":               {func() Config { cfg := valid(); cfg.Roots = []*Node{cycle}; return cfg }, `tree chart node "Cycle" child 0 contains a cycle`},
		"shared node": {func() Config {
			cfg := valid()
			cfg.Roots = []*Node{{Name: "A", Children: []*Node{shared}}, {Name: "B", Children: []*Node{shared}}}
			return cfg
		}, `tree chart node "B" child 0 reuses a node`},
		"too deep": {func() Config { cfg := valid(); cfg.Roots = []*Node{deep}; return cfg }, `exceeds maximum depth 256`},
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
