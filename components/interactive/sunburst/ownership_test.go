package sunburst_test

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

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/araihu/goshtoso-charts/components/interactive"
	interactivesunburst "github.com/araihu/goshtoso-charts/components/interactive/sunburst"
)

var (
	chartIDPattern = regexp.MustCompile(`id="([A-Za-z]{12})"`)
	scriptPattern  = regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`)
)

func TestSunburstSpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()
	const wantPackage = "github.com/araihu/goshtoso-charts/components/interactive/sunburst"
	canonicalTypes := []reflect.Type{
		reflect.TypeOf(interactivesunburst.Navigation("")),
		reflect.TypeOf(interactivesunburst.Sort("")),
		reflect.TypeOf(interactivesunburst.Config{}),
		reflect.TypeOf(interactivesunburst.Node{}),
	}
	for _, typ := range canonicalTypes {
		if got := typ.PkgPath(); got != wantPackage {
			t.Errorf("%s PkgPath() = %q, want %q", typ, got, wantPackage)
		}
	}
	compatibilityTypes := []reflect.Type{
		reflect.TypeOf(interactive.SunburstNavigation("")),
		reflect.TypeOf(interactive.SunburstSort("")),
		reflect.TypeOf(interactive.SunburstConfig{}),
		reflect.TypeOf(interactive.SunburstNode{}),
	}
	for index, typ := range compatibilityTypes {
		if typ != canonicalTypes[index] {
			t.Errorf("compatibility type %s is not identical to canonical %s", typ, canonicalTypes[index])
		}
	}
	if interactive.SunburstNavigationDrillDown != interactivesunburst.NavigationDrillDown ||
		interactive.SunburstNavigationDisabled != interactivesunburst.NavigationDisabled ||
		interactive.SunburstSortDescending != interactivesunburst.SortDescending ||
		interactive.SunburstSortAscending != interactivesunburst.SortAscending ||
		interactive.SunburstSortInput != interactivesunburst.SortInput {
		t.Fatal("compatibility constants differ from canonical constants")
	}

	configType := reflect.TypeOf(interactivesunburst.Config{})
	options, _ := configType.FieldByName("Options")
	labelOptions, _ := configType.FieldByName("LabelOptions")
	itemStyle, _ := configType.FieldByName("ItemStyle")
	nodeType := reflect.TypeOf(interactivesunburst.Node{})
	nodeLabel, _ := nodeType.FieldByName("Label")
	nodeItemStyle, _ := nodeType.FieldByName("ItemStyle")
	if options.Type != reflect.TypeOf(chart.ChartOptions{}) ||
		labelOptions.Type != reflect.TypeOf((*chart.LabelOptions)(nil)) ||
		itemStyle.Type != reflect.TypeOf((*chart.ItemStyle)(nil)) ||
		nodeLabel.Type != reflect.TypeOf((*chart.LabelOptions)(nil)) ||
		nodeItemStyle.Type != reflect.TypeOf((*chart.ItemStyle)(nil)) {
		t.Fatalf("shared fields are not chart-owned: Options=%v LabelOptions=%v ItemStyle=%v Node.Label=%v Node.ItemStyle=%v", options.Type, labelOptions.Type, itemStyle.Type, nodeLabel.Type, nodeItemStyle.Type)
	}
}

func TestSunburstFacadePreservesNavigationSortRenderingValidationAndBaseHashes(t *testing.T) {
	t.Parallel()
	variants := []struct {
		name         string
		cfg          interactivesunburst.Config
		fullDigest   string
		scriptDigest string
		shellDigest  string
	}{
		{
			name: "default-navigation-sort",
			cfg: interactivesunburst.Config{
				Label: "Hierarchy", Nodes: []*interactivesunburst.Node{{Name: "root", Value: 1}},
			},
			fullDigest: "b8d64af3a92633226f37fd43acb678b66d1a714546bd0e0172b5ffdb4b28922b", scriptDigest: "9ecd51fc5549c9f222d30112aac121917615f2f836e760e4f8a5c7d471c62244", shellDigest: "1967c6e077666a3227a8b82170d69a921488db3fa6d1337039d717f54235490e",
		},
		{
			name: "configured-hierarchy",
			cfg: interactivesunburst.Config{
				Label: "Basic sunburst example", Caption: "Seven parent and child pairs.",
				Nodes: []*interactivesunburst.Node{
					{Name: "parent-0", Value: 0.81, ItemStyle: &chart.ItemStyle{Color: "#123456", BorderColor: "#ffffff", BorderWidth: 2}, Label: &chart.LabelOptions{Show: chart.Bool(true), Position: "inside", Color: "#fedcba"}, Children: []*interactivesunburst.Node{{Name: "child-0", Value: 0.34}}},
					{Name: "parent-1", Value: 0, Children: []*interactivesunburst.Node{{Name: "child-1", Value: 0.57}}},
				},
				Navigation: interactivesunburst.NavigationDrillDown, Sort: interactivesunburst.SortAscending,
				LabelOptions: &chart.LabelOptions{Show: chart.Bool(true), Position: "inside", FontSize: 10}, ItemStyle: &chart.ItemStyle{BorderColor: "#eeeeee", BorderWidth: 1}, ShowLabelsForZero: chart.Bool(true),
				InnerRadius: 16, OuterRadius: 88, Width: "100%", Height: "32rem",
				Options: chart.ChartOptions{Title: &chart.TitleOptions{Text: "Basic sunburst example"}, Animation: chart.Bool(false)}, Style: charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#654321"}, Class: "rounded-radius max-w-full"},
				RootAttrs: templ.Attributes{"id": "basic-sunburst", "data-chart-purpose": "hierarchy"},
			},
			fullDigest: "fd94c3ef45a9fd3033414d2f8668ec9dd58cfcd1fca13c4236e7e9dacab95bf4", scriptDigest: "3eb4a4c40ef4d8e1776002b793c059d55fcbac15b6e82a35a13acec0eb3da7e9", shellDigest: "900bacd029217b6be0429404f1e4ebece7fbb3a9fe2416390ed91ff0eecf8f12",
		},
		{
			name: "fixed-input-wrapper-omitted",
			cfg: interactivesunburst.Config{
				Label: "Fixed hierarchy", Caption: "Input order and zero values.",
				Nodes:      []*interactivesunburst.Node{{Name: "zero", Value: 0}, {Name: "branch", Value: 3, Children: []*interactivesunburst.Node{{Name: "leaf", Value: 1.5}}}},
				Navigation: interactivesunburst.NavigationDisabled, Sort: interactivesunburst.SortInput, OuterRadius: 100,
				Options: chart.ChartOptions{Controls: chartcontrol.Options{Mode: chartcontrol.WrapperModeOmitted}, Export: &chartcontrol.ExportOptions{Filename: "fixed-hierarchy"}},
				Style:   charttheme.Style{Palette: charttheme.PalettePastel, Class: "caller-sunburst"}, RootAttrs: templ.Attributes{"data-owner": "consumer"},
			},
			fullDigest: "7a18bb5a926e1f7c2919270b801a81c8173dff5668ec36413d8f0997284f9745", scriptDigest: "876f2695e08168f639a987c347de482bbe1b340750213eaf3508d7a9b068cf7e", shellDigest: "266cebeb904299687c7cd6b4d4ac4e31a21c5a06627f9556f4971611465c04fb",
		},
	}
	for _, variant := range variants {
		variant := variant
		t.Run(variant.name, func(t *testing.T) {
			t.Parallel()
			canonical := normalizedRender(t, interactivesunburst.Sunburst(variant.cfg))
			legacy := normalizedRender(t, interactive.Sunburst(variant.cfg))
			if canonical != legacy {
				t.Fatal("canonical Sunburst render differs from compatibility facade")
			}
			assertDigest(t, "full", canonical, variant.fullDigest)
			assertDigest(t, "scripts", strings.Join(scriptPattern.FindAllString(canonical, -1), "\n"), variant.scriptDigest)
			assertDigest(t, "shell", scriptPattern.ReplaceAllString(canonical, "<script></script>"), variant.shellDigest)
		})
	}

	invalid := interactivesunburst.Config{Label: "Hierarchy", Nodes: []*interactivesunburst.Node{{Name: "root", Value: 1}}, Sort: "alphabetical"}
	canonicalError := renderError(interactivesunburst.Sunburst(invalid))
	legacyError := renderError(interactive.Sunburst(invalid))
	if canonicalError != `sunburst chart sort "alphabetical" is not supported` || canonicalError != legacyError {
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
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "sunburst.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse parent facade: %v", err)
	}
	if len(file.Imports) != 1 || strings.Trim(file.Imports[0].Path.Value, `"`) != "github.com/araihu/goshtoso-charts/components/interactive/sunburst" {
		t.Fatalf("parent imports = %v, want only canonical Sunburst package", file.Imports)
	}
	functions := 0
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			functions++
			if declaration.Name.Name != "Sunburst" || declaration.Body == nil || len(declaration.Body.List) != 1 {
				t.Errorf("parent function %s is not the single Sunburst forwarder", declaration.Name.Name)
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
		t.Errorf("parent functions = %d, want only Sunburst", functions)
	}
}

func TestSunburstTemplateRuntimeAndUpstreamProvenance(t *testing.T) {
	t.Parallel()
	for _, oldPath := range []string{filepath.Join("..", "sunburst.templ"), filepath.Join("..", "sunburst_templ.go")} {
		if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
			t.Errorf("compatibility parent unexpectedly retains %s", oldPath)
		}
	}
	checks := []struct {
		path       string
		normalize  *strings.Replacer
		wantDigest string
	}{
		{"sunburst.templ", strings.NewReplacer("package sunburst", "package interactive", "[]valueRow", "[]sunburstValueRow"), "21d74e9158e25ad202dcd079b8191b6cf69be6fc07812861875a6167f88322bc"},
		{"sunburst_templ.go", strings.NewReplacer("package sunburst", "package interactive", "components/interactive/sunburst/sunburst.templ", "components/interactive/sunburst.templ", "[]valueRow", "[]sunburstValueRow"), "0e0957600b5bbcc5fec2b34fe2de11e12c8b37e202b3933ce15478fb937065e3"},
	}
	for _, check := range checks {
		contents, err := os.ReadFile(check.path)
		if err != nil {
			t.Fatalf("read %s: %v", check.path, err)
		}
		digest := sha256.Sum256([]byte(check.normalize.Replace(string(contents))))
		if got := hex.EncodeToString(digest[:]); got != check.wantDigest {
			t.Errorf("normalized %s SHA-256 = %s, want base %s", check.path, got, check.wantDigest)
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
	attributions, err := os.ReadFile(filepath.Join("..", "..", "..", "site", "internal", "pages", "attributions.go"))
	if err != nil {
		t.Fatalf("read central upstream provenance: %v", err)
	}
	if !bytes.Contains(attributions, []byte("examples/sunburst.go")) || !bytes.Contains(attributions, []byte("bda428480a82d6d77ebb9fa939cf8d52528453dd")) {
		t.Fatal("central upstream provenance no longer identifies the pinned Sunburst example")
	}
	coverage, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "upstream-example-coverage.md"))
	if err != nil {
		t.Fatalf("read central coverage ledger: %v", err)
	}
	for _, token := range []string{"components/interactive/sunburst.Sunburst", "examples/sunburst.go", "bda428480a82d6d77ebb9fa939cf8d52528453dd"} {
		if !bytes.Contains(coverage, []byte(token)) {
			t.Errorf("central coverage ledger lacks %q", token)
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
	_ func(interactivesunburst.Config) chart.Instance = interactivesunburst.Sunburst
	_ func(interactive.SunburstConfig) chart.Instance = interactive.Sunburst
	_ func(interactivesunburst.Config) chart.Instance = interactive.Sunburst
	_ func(interactive.SunburstConfig) chart.Instance = interactivesunburst.Sunburst
)
