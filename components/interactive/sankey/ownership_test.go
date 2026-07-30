package sankey_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/araihu/goshtoso-charts/components/interactive"
	interactivesankey "github.com/araihu/goshtoso-charts/components/interactive/sankey"
)

var (
	chartIDPattern = regexp.MustCompile(`id="([A-Za-z]{12})"`)
	scriptPattern  = regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`)
)

func TestSankeySpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()
	const wantPackage = "github.com/araihu/goshtoso-charts/components/interactive/sankey"
	canonicalTypes := []reflect.Type{
		reflect.TypeOf(interactivesankey.Orientation("")),
		reflect.TypeOf(interactivesankey.Alignment("")),
		reflect.TypeOf(interactivesankey.Layout{}),
		reflect.TypeOf(interactivesankey.Config{}),
		reflect.TypeOf(interactivesankey.Series{}),
		reflect.TypeOf(interactivesankey.Node{}),
		reflect.TypeOf(interactivesankey.Link{}),
	}
	for _, typ := range canonicalTypes {
		if got := typ.PkgPath(); got != wantPackage {
			t.Errorf("%s PkgPath() = %q, want %q", typ, got, wantPackage)
		}
	}
	compatibilityTypes := []reflect.Type{
		reflect.TypeOf(interactive.SankeyOrientation("")),
		reflect.TypeOf(interactive.SankeyAlignment("")),
		reflect.TypeOf(interactive.SankeyLayout{}),
		reflect.TypeOf(interactive.SankeyConfig{}),
		reflect.TypeOf(interactive.SankeySeries{}),
		reflect.TypeOf(interactive.SankeyNode{}),
		reflect.TypeOf(interactive.SankeyLink{}),
	}
	for index, typ := range compatibilityTypes {
		if typ != canonicalTypes[index] {
			t.Errorf("compatibility type %s is not identical to canonical %s", typ, canonicalTypes[index])
		}
	}
	if interactive.SankeyOrientationHorizontal != interactivesankey.OrientationHorizontal ||
		interactive.SankeyOrientationVertical != interactivesankey.OrientationVertical ||
		interactive.SankeyAlignmentJustify != interactivesankey.AlignmentJustify ||
		interactive.SankeyAlignmentLeft != interactivesankey.AlignmentLeft ||
		interactive.SankeyAlignmentRight != interactivesankey.AlignmentRight {
		t.Fatal("compatibility constants differ from canonical constants")
	}

	configType := reflect.TypeOf(interactivesankey.Config{})
	options, _ := configType.FieldByName("Options")
	seriesOptions, _ := configType.FieldByName("SeriesOptions")
	seriesType := reflect.TypeOf(interactivesankey.Series{})
	perSeriesOptions, _ := seriesType.FieldByName("Options")
	nodeType := reflect.TypeOf(interactivesankey.Node{})
	itemStyle, _ := nodeType.FieldByName("ItemStyle")
	if options.Type != reflect.TypeOf(chart.ChartOptions{}) ||
		seriesOptions.Type != reflect.TypeOf(chart.SeriesOptions{}) ||
		perSeriesOptions.Type != reflect.TypeOf(chart.SeriesOptions{}) ||
		itemStyle.Type != reflect.TypeOf((*chart.ItemStyle)(nil)) {
		t.Fatalf("shared fields are not chart-owned: Options=%v SeriesOptions=%v Series.Options=%v Node.ItemStyle=%v", options.Type, seriesOptions.Type, perSeriesOptions.Type, itemStyle.Type)
	}
}

func TestSankeyFacadePreservesCanonicalRenderValidationAndBaseHashes(t *testing.T) {
	t.Parallel()
	depth := 1
	variants := []struct {
		name         string
		cfg          interactivesankey.Config
		fullDigest   string
		scriptDigest string
		shellDigest  string
	}{
		{
			name: "default-layout",
			cfg: interactivesankey.Config{
				Label: "Flow", Series: []interactivesankey.Series{{Name: "Flow", Nodes: []interactivesankey.Node{{Name: "Input"}, {Name: "Output"}}, Links: []interactivesankey.Link{{Source: "Input", Target: "Output", Value: 1}}}},
			},
			fullDigest: "f812067fb3b2e9e15b6853f747d9438c6d62aab38256d0af0a448368bc6b989e", scriptDigest: "d4ef9962adfeb13d0e87c871ac032d89b16ee4eb1c09ea7faef8653e8377aae6", shellDigest: "9722e2317acbea48fbea8eba8d47a023015110d121f39eff4aeed4e25450a58e",
		},
		{
			name: "vertical-custom",
			cfg: interactivesankey.Config{
				Label: "Energy flow", Caption: "Energy moving from generation to demand.",
				Layout: interactivesankey.Layout{Orientation: interactivesankey.OrientationVertical, Alignment: interactivesankey.AlignmentRight, NodeWidth: 24, NodeGap: 12},
				Series: []interactivesankey.Series{{Name: "Energy", Nodes: []interactivesankey.Node{{Name: "Solar", ItemStyle: &chart.ItemStyle{Color: "#f59e0b"}}, {Name: "Homes", Depth: &depth}}, Links: []interactivesankey.Link{{Source: "Solar", Target: "Homes", Value: 42.5}}, Options: chart.SeriesOptions{Animation: chart.Bool(false)}}},
				Width:  "720px", Height: "420px", Options: chart.ChartOptions{Title: &chart.TitleOptions{Text: "Energy balance"}}, SeriesOptions: chart.SeriesOptions{Label: &chart.LabelOptions{Show: chart.Bool(true), Position: "right"}, LineStyle: &chart.LineStyle{Color: "source", Opacity: chart.Float(0.6)}},
				Style: charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-96"},
			},
			fullDigest: "5b44d8a6b4f0e4f626dbae4237dbc6f73096aa2400ce3c69f4e3e677ad5bc552", scriptDigest: "26f1fd9e40e44c2f7506cee8ce8ccf9304677917595cc3bc1c1a051fa15e7f03", shellDigest: "f148720c24108e1d96fc1e189380edaceada288ed0597bc436b93fd79e8b6a57",
		},
		{
			name: "multiple-wrapper-omitted",
			cfg: interactivesankey.Config{
				Label: "Material flows", Layout: interactivesankey.Layout{Alignment: interactivesankey.AlignmentLeft},
				Series: []interactivesankey.Series{
					{Name: "Primary", Nodes: []interactivesankey.Node{{Name: "Raw"}, {Name: "Built"}, {Name: "Waste"}}, Links: []interactivesankey.Link{{Source: "Raw", Target: "Built", Value: 7.25}, {Source: "Raw", Target: "Waste", Value: 0}}},
					{Name: "Secondary", Nodes: []interactivesankey.Node{{Name: "Recovered"}, {Name: "Reused"}}, Links: []interactivesankey.Link{{Source: "Recovered", Target: "Reused", Value: 2.5}}, Options: chart.SeriesOptions{Label: &chart.LabelOptions{Show: chart.Bool(false)}}},
				},
				Options: chart.ChartOptions{Animation: chart.Bool(false), Controls: chartcontrol.Options{Mode: chartcontrol.WrapperModeOmitted}, Export: &chartcontrol.ExportOptions{Filename: "material-flows"}}, SeriesOptions: chart.SeriesOptions{LineStyle: &chart.LineStyle{Opacity: chart.Float(0.4)}},
				Style: charttheme.Style{Palette: charttheme.PalettePastel, Class: "caller-sankey"},
			},
			fullDigest: "4910031ab7ea5165dae6c812ba4d60d6aea26060a7b35b03efa4b36b0f7f36fd", scriptDigest: "3af9e02be276e6be363915ee4a64225e475bbbba5738d65cbfd0236f520d0268", shellDigest: "365eb2ea473e14518059bc464042ed6aba4b1028ec953aaac09356dbade497a2",
		},
	}
	for _, variant := range variants {
		variant := variant
		t.Run(variant.name, func(t *testing.T) {
			t.Parallel()
			canonical := normalizedRender(t, interactivesankey.Sankey(variant.cfg))
			legacy := normalizedRender(t, interactive.Sankey(variant.cfg))
			if canonical != legacy {
				t.Fatal("canonical Sankey render differs from compatibility facade")
			}
			assertDigest(t, "full", canonical, variant.fullDigest)
			assertDigest(t, "scripts", strings.Join(scriptPattern.FindAllString(canonical, -1), "\n"), variant.scriptDigest)
			assertDigest(t, "shell", scriptPattern.ReplaceAllString(canonical, "<script></script>"), variant.shellDigest)
		})
	}

	invalid := interactivesankey.Config{
		Label: "Flow",
		Series: []interactivesankey.Series{{
			Name:  "Flow",
			Nodes: []interactivesankey.Node{{Name: "Input"}, {Name: "Output"}},
			Links: []interactivesankey.Link{{Source: "Input", Target: "Missing", Value: 1}},
		}},
	}
	canonicalError := renderError(interactivesankey.Sankey(invalid))
	legacyError := renderError(interactive.Sankey(invalid))
	if canonicalError != `sankey chart series "Flow" link 0 target "Missing" does not name a node` || canonicalError != legacyError {
		t.Fatalf("canonical validation error = %q, legacy = %q", canonicalError, legacyError)
	}
}

func TestCanonicalPackageDoesNotImportCompatibilityParentOrExportRendererNames(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package directory: %v", err)
	}
	files := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, entry.Name(), nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", entry.Name(), err)
		}
		for _, imported := range file.Imports {
			if strings.Trim(imported.Path.Value, `"`) == "github.com/araihu/goshtoso-charts/components/interactive" {
				t.Errorf("%s imports compatibility parent %s", entry.Name(), imported.Path.Value)
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			identifier, ok := node.(*ast.Ident)
			if ok && token.IsExported(identifier.Name) && strings.Contains(strings.ToLower(identifier.Name), "echarts") {
				t.Errorf("%s exports renderer-named identifier %s", entry.Name(), identifier.Name)
			}
			return true
		})
	}
}

func TestCompatibilityParentContainsOnlyAliasesConstantsAndForwarder(t *testing.T) {
	t.Parallel()
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "sankey.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse parent facade: %v", err)
	}
	if len(file.Imports) != 1 || strings.Trim(file.Imports[0].Path.Value, `"`) != "github.com/araihu/goshtoso-charts/components/interactive/sankey" {
		t.Fatalf("parent imports = %v, want only canonical Sankey package", file.Imports)
	}
	functions := 0
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			functions++
			if declaration.Name.Name != "Sankey" || declaration.Body == nil || len(declaration.Body.List) != 1 {
				t.Errorf("parent function %s is not the single Sankey forwarder", declaration.Name.Name)
			}
		case *ast.GenDecl:
			switch declaration.Tok {
			case token.IMPORT, token.CONST:
			case token.TYPE:
				for _, spec := range declaration.Specs {
					typeSpec := spec.(*ast.TypeSpec)
					if !typeSpec.Assign.IsValid() {
						t.Errorf("parent type %s is not an alias", typeSpec.Name.Name)
					}
				}
			default:
				t.Errorf("parent contains forbidden %s declaration", declaration.Tok)
			}
		default:
			t.Errorf("parent contains unexpected declaration %T", declaration)
		}
	}
	if functions != 1 {
		t.Errorf("parent functions = %d, want only Sankey", functions)
	}
}

func TestSankeyUsesUnchangedSharedTemplateRuntimeAndCentralProvenance(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"sankey.templ", "sankey_templ.go", filepath.Join("..", "sankey.templ"), filepath.Join("..", "sankey_templ.go")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Sankey unexpectedly owns component-specific template artifact %s", path)
		}
	}
	for path, want := range map[string]string{
		filepath.Join("..", "..", "internal", "interactive", "theme_runtime.go"):     "07607b72118cf2e2e1cc71d81ec3c64789dd2f053ff0f9282a2e41f92cbf24ae",
		filepath.Join("..", "..", "internal", "interactive", "interactive.templ"):    "96137171bb2e6cb69372a59596963d0f4e7e6f87f3079863ccc92f3f59f8680a",
		filepath.Join("..", "..", "internal", "interactive", "interactive_templ.go"): "9e83e969108d203fe38fce502a21fed7c0f85d150cd8978d4ca76e596d278808",
	} {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read provenance file %s: %v", path, err)
		}
		digest := sha256.Sum256(contents)
		if got := hex.EncodeToString(digest[:]); got != want {
			t.Errorf("provenance file %s SHA-256 = %s, want %s", path, got, want)
		}
	}

	for path, required := range map[string][]string{
		filepath.Join("..", "..", "..", "site", "internal", "pages", "attributions.go"): {
			"examples/sankey.go",
			"bda428480a82d6d77ebb9fa939cf8d52528453dd",
		},
		filepath.Join("..", "..", "..", "docs", "upstream-example-coverage.md"): {
			"components/interactive/sankey.Sankey",
			"examples/sankey.go",
			"bda428480a82d6d77ebb9fa939cf8d52528453dd",
		},
	} {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read central provenance file %s: %v", path, err)
		}
		for _, token := range required {
			if !bytes.Contains(contents, []byte(token)) {
				t.Errorf("central provenance file %s lacks %q", path, token)
			}
		}
	}
}

func normalizedRender(t *testing.T, instance chart.Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	match := chartIDPattern.FindStringSubmatch(output.String())
	if len(match) != 2 {
		t.Fatalf("rendered markup lacks chart ID: %s", output.String())
	}
	return strings.ReplaceAll(output.String(), match[1], "CHARTID")
}

func renderError(instance chart.Instance) string {
	var output bytes.Buffer
	err := instance.Render(context.Background(), &output)
	if err == nil {
		return ""
	}
	if output.Len() != 0 {
		return "validation wrote output"
	}
	return err.Error()
}

func assertDigest(t *testing.T, name, value, want string) {
	t.Helper()
	digest := sha256.Sum256([]byte(value))
	if got := hex.EncodeToString(digest[:]); got != want {
		t.Errorf("%s SHA-256 = %s, want base %s", name, got, want)
	}
}

var (
	_ func(interactivesankey.Config) chart.Instance = interactivesankey.Sankey
	_ func(interactive.SankeyConfig) chart.Instance = interactive.Sankey
	_ func(interactivesankey.Config) chart.Instance = interactive.Sankey
	_ func(interactive.SankeyConfig) chart.Instance = interactivesankey.Sankey
)
