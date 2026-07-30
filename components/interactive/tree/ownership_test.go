package tree_test

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
	"github.com/araihu/goshtoso-charts/components/interactive"
	interactivetree "github.com/araihu/goshtoso-charts/components/interactive/tree"
)

var treeIDPattern = regexp.MustCompile(`id="([A-Za-z]{12})"`)

func TestTreeSpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()
	const wantPackage = "github.com/araihu/goshtoso-charts/components/interactive/tree"
	canonicalTypes := []reflect.Type{
		reflect.TypeOf(interactivetree.Layout("")),
		reflect.TypeOf(interactivetree.Orientation("")),
		reflect.TypeOf(interactivetree.Roam(0)),
		reflect.TypeOf(interactivetree.Symbol("")),
		reflect.TypeOf(interactivetree.Insets{}),
		reflect.TypeOf(interactivetree.Config{}),
		reflect.TypeOf(interactivetree.Node{}),
	}
	for _, typ := range canonicalTypes {
		if got := typ.PkgPath(); got != wantPackage {
			t.Errorf("%s PkgPath() = %q, want %q", typ, got, wantPackage)
		}
	}
	compatibilityTypes := []reflect.Type{
		reflect.TypeOf(interactive.TreeLayout("")),
		reflect.TypeOf(interactive.TreeOrientation("")),
		reflect.TypeOf(interactive.TreeRoam(0)),
		reflect.TypeOf(interactive.TreeSymbol("")),
		reflect.TypeOf(interactive.TreeInsets{}),
		reflect.TypeOf(interactive.TreeConfig{}),
		reflect.TypeOf(interactive.TreeNode{}),
	}
	for index, typ := range compatibilityTypes {
		if typ != canonicalTypes[index] {
			t.Errorf("compatibility type %s is not identical to canonical %s", typ, canonicalTypes[index])
		}
	}

	if interactive.TreeLayoutLayered != interactivetree.LayoutLayered || interactive.TreeLayoutRadial != interactivetree.LayoutRadial ||
		interactive.TreeOrientationLeftToRight != interactivetree.OrientationLeftToRight || interactive.TreeOrientationRightToLeft != interactivetree.OrientationRightToLeft ||
		interactive.TreeOrientationTopToBottom != interactivetree.OrientationTopToBottom || interactive.TreeOrientationBottomToTop != interactivetree.OrientationBottomToTop ||
		interactive.TreeRoamDisabled != interactivetree.RoamDisabled || interactive.TreeRoamEnabled != interactivetree.RoamEnabled ||
		interactive.TreeSymbolCircle != interactivetree.SymbolCircle || interactive.TreeSymbolRectangle != interactivetree.SymbolRectangle ||
		interactive.TreeSymbolRoundedRectangle != interactivetree.SymbolRoundedRectangle || interactive.TreeSymbolTriangle != interactivetree.SymbolTriangle ||
		interactive.TreeSymbolDiamond != interactivetree.SymbolDiamond || interactive.TreeSymbolPin != interactivetree.SymbolPin ||
		interactive.TreeSymbolArrow != interactivetree.SymbolArrow || interactive.TreeSymbolNone != interactivetree.SymbolNone {
		t.Fatal("legacy Tree constants do not preserve canonical values")
	}

	configType := reflect.TypeOf(interactivetree.Config{})
	options, _ := configType.FieldByName("Options")
	nodeLabel, _ := configType.FieldByName("NodeLabel")
	leafLabel, _ := configType.FieldByName("LeafLabel")
	nodeType := reflect.TypeOf(interactivetree.Node{})
	itemStyle, _ := nodeType.FieldByName("ItemStyle")
	lineStyle, _ := nodeType.FieldByName("LineStyle")
	if options.Type != reflect.TypeOf(chart.ChartOptions{}) || nodeLabel.Type != reflect.TypeOf((*chart.LabelOptions)(nil)) ||
		leafLabel.Type != reflect.TypeOf((*chart.LabelOptions)(nil)) || itemStyle.Type != reflect.TypeOf((*chart.ItemStyle)(nil)) ||
		lineStyle.Type != reflect.TypeOf((*chart.LineStyle)(nil)) {
		t.Fatalf("shared fields have wrong ownership: Options=%v NodeLabel=%v LeafLabel=%v ItemStyle=%v LineStyle=%v", options.Type, nodeLabel.Type, leafLabel.Type, itemStyle.Type, lineStyle.Type)
	}
}

func TestTreeFacadePreservesCollapseNavigationRenderAndValidation(t *testing.T) {
	t.Parallel()
	initialDepth := 0
	collapsed := true
	expand := false
	cfg := interactivetree.Config{
		Label: "Hierarchy",
		Roots: []*interactivetree.Node{{Name: "Root", Children: []*interactivetree.Node{{Name: "Collapsed", Collapsed: &collapsed, Children: []*interactivetree.Node{{Name: "Hidden"}}}}}},
		Roam:  interactivetree.RoamEnabled, ExpandAndCollapse: &expand, InitialDepth: &initialDepth,
	}
	canonical := normalizedRender(t, interactivetree.Tree(cfg))
	legacy := normalizedRender(t, interactive.Tree(cfg))
	if canonical != legacy {
		t.Fatal("canonical Tree render differs from compatibility facade")
	}
	for _, token := range []string{`"collapsed":true`, `"roam":true`, `"expandAndCollapse":false`, `"initialTreeDepth":0`, `"animationDuration":150`, `"animationDurationUpdate":100`} {
		if !strings.Contains(canonical, token) {
			t.Errorf("canonical Tree render lacks preserved collapse/navigation token %s", token)
		}
	}
	if strings.Contains(canonical, "-2147483648") {
		t.Fatal("canonical Tree render leaked the zero-depth sentinel")
	}
	invalid := cfg
	invalid.Layout = "cluster"
	canonicalError := renderError(interactivetree.Tree(invalid))
	legacyError := renderError(interactive.Tree(invalid))
	if canonicalError != `tree chart layout "cluster" is not supported` || canonicalError != legacyError {
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
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "tree.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse parent facade: %v", err)
	}
	if len(file.Imports) != 1 || strings.Trim(file.Imports[0].Path.Value, `"`) != "github.com/araihu/goshtoso-charts/components/interactive/tree" {
		t.Fatalf("parent imports = %v, want only canonical Tree package", file.Imports)
	}
	functions := 0
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			functions++
			if declaration.Name.Name != "Tree" || declaration.Body == nil || len(declaration.Body.List) != 1 {
				t.Errorf("parent function %s is not the single Tree forwarder", declaration.Name.Name)
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
		t.Errorf("parent functions = %d, want only Tree", functions)
	}
}

func TestTreeUsesUnchangedSharedTemplateRuntimeAndCentralProvenance(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"tree.templ", "tree_templ.go", filepath.Join("..", "tree.templ"), filepath.Join("..", "tree_templ.go")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Tree unexpectedly owns component-specific template artifact %s", path)
		}
	}
	for path, want := range map[string]string{
		filepath.Join("..", "..", "internal", "interactive", "interactive.templ"):    "96137171bb2e6cb69372a59596963d0f4e7e6f87f3079863ccc92f3f59f8680a",
		filepath.Join("..", "..", "internal", "interactive", "interactive_templ.go"): "9e83e969108d203fe38fce502a21fed7c0f85d150cd8978d4ca76e596d278808",
		filepath.Join("..", "..", "internal", "interactive", "live_runtime.go"):      "52feab4a14c172ffe212fb95b98e0363293ad0ad253ae60e8caea22caf7f2a4b",
		filepath.Join("..", "..", "internal", "interactive", "theme_runtime.go"):     "07607b72118cf2e2e1cc71d81ec3c64789dd2f053ff0f9282a2e41f92cbf24ae",
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
	if !bytes.Contains(attributions, []byte("examples/tree.go")) || !bytes.Contains(attributions, []byte("bda428480a82d6d77ebb9fa939cf8d52528453dd")) {
		t.Fatal("central provenance no longer identifies the pinned interactive examples source")
	}
	coverage, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "upstream-example-coverage.md"))
	if err != nil {
		t.Fatalf("read central coverage ledger: %v", err)
	}
	for _, token := range []string{"components/interactive/tree.Tree", "examples/tree.go", "bda428480a82d6d77ebb9fa939cf8d52528453dd"} {
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
	match := treeIDPattern.FindStringSubmatch(output.String())
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

var (
	_ func(interactivetree.Config) chart.Instance = interactivetree.Tree
	_ func(interactive.TreeConfig) chart.Instance = interactive.Tree
	_ func(interactivetree.Config) chart.Instance = interactive.Tree
	_ func(interactive.TreeConfig) chart.Instance = interactivetree.Tree
)
