package graph_test

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
	interactivegraph "github.com/araihu/goshtoso-charts/components/interactive/graph"
)

var graphIDPattern = regexp.MustCompile(`id="([A-Za-z]{12})"`)

func TestGraphSpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()
	const wantPackage = "github.com/araihu/goshtoso-charts/components/interactive/graph"
	canonicalTypes := []reflect.Type{
		reflect.TypeOf(interactivegraph.Layout("")),
		reflect.TypeOf(interactivegraph.Roam(0)),
		reflect.TypeOf(interactivegraph.ForceInitialLayout("")),
		reflect.TypeOf(interactivegraph.ForceOptions{}),
		reflect.TypeOf(interactivegraph.Config{}),
		reflect.TypeOf(interactivegraph.Node{}),
		reflect.TypeOf(interactivegraph.Link{}),
		reflect.TypeOf(interactivegraph.Category{}),
	}
	for _, typ := range canonicalTypes {
		if got := typ.PkgPath(); got != wantPackage {
			t.Errorf("%s PkgPath() = %q, want %q", typ, got, wantPackage)
		}
	}
	compatibilityTypes := []reflect.Type{
		reflect.TypeOf(interactive.GraphLayout("")),
		reflect.TypeOf(interactive.GraphRoam(0)),
		reflect.TypeOf(interactive.ForceInitialLayout("")),
		reflect.TypeOf(interactive.ForceOptions{}),
		reflect.TypeOf(interactive.GraphConfig{}),
		reflect.TypeOf(interactive.Node{}),
		reflect.TypeOf(interactive.Link{}),
		reflect.TypeOf(interactive.Category{}),
	}
	for index, typ := range compatibilityTypes {
		if typ != canonicalTypes[index] {
			t.Errorf("compatibility type %s is not identical to canonical %s", typ, canonicalTypes[index])
		}
	}

	if interactive.GraphLayoutForce != interactivegraph.LayoutForce ||
		interactive.GraphLayoutNone != interactivegraph.LayoutNone ||
		interactive.GraphLayoutCircular != interactivegraph.LayoutCircular ||
		interactive.GraphRoamDisabled != interactivegraph.RoamDisabled ||
		interactive.GraphRoamEnabled != interactivegraph.RoamEnabled ||
		interactive.ForceInitialLayoutNone != interactivegraph.ForceInitialLayoutNone ||
		interactive.ForceInitialLayoutCircular != interactivegraph.ForceInitialLayoutCircular {
		t.Fatal("legacy Graph constants do not preserve canonical values")
	}

	configType := reflect.TypeOf(interactivegraph.Config{})
	options, _ := configType.FieldByName("Options")
	seriesOptions, _ := configType.FieldByName("SeriesOptions")
	nodeItemStyle, _ := reflect.TypeOf(interactivegraph.Node{}).FieldByName("ItemStyle")
	linkLineStyle, _ := reflect.TypeOf(interactivegraph.Link{}).FieldByName("LineStyle")
	categoryItemStyle, _ := reflect.TypeOf(interactivegraph.Category{}).FieldByName("ItemStyle")
	categoryLabel, _ := reflect.TypeOf(interactivegraph.Category{}).FieldByName("Label")
	if options.Type != reflect.TypeOf(chart.ChartOptions{}) || seriesOptions.Type != reflect.TypeOf(chart.SeriesOptions{}) ||
		nodeItemStyle.Type != reflect.TypeOf((*chart.ItemStyle)(nil)) || linkLineStyle.Type != reflect.TypeOf((*chart.LineStyle)(nil)) ||
		categoryItemStyle.Type != reflect.TypeOf((*chart.ItemStyle)(nil)) || categoryLabel.Type != reflect.TypeOf((*chart.LabelOptions)(nil)) {
		t.Fatalf("shared fields have wrong ownership: Options=%v SeriesOptions=%v Node.ItemStyle=%v Link.LineStyle=%v Category.ItemStyle=%v Category.Label=%v", options.Type, seriesOptions.Type, nodeItemStyle.Type, linkLineStyle.Type, categoryItemStyle.Type, categoryLabel.Type)
	}
}

func TestGraphFacadePreservesCanonicalRenderAndValidation(t *testing.T) {
	t.Parallel()
	cfg := interactivegraph.Config{
		Label:      "Network",
		Nodes:      []interactivegraph.Node{{Name: "api", Category: "service"}, {Name: "db"}},
		Links:      []interactivegraph.Link{{Source: "api", Target: "db"}},
		Categories: []interactivegraph.Category{{Name: "service"}},
	}
	canonical := normalizedRender(t, interactivegraph.Graph(cfg))
	legacy := normalizedRender(t, interactive.Graph(cfg))
	if canonical != legacy {
		t.Fatal("canonical Graph render differs from compatibility facade")
	}
	invalid := cfg
	invalid.Layout = "grid"
	canonicalError := renderError(interactivegraph.Graph(invalid))
	legacyError := renderError(interactive.Graph(invalid))
	if canonicalError != `graph chart layout "grid" is not supported` || canonicalError != legacyError {
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
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "graph.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse parent facade: %v", err)
	}
	if len(file.Imports) != 1 || strings.Trim(file.Imports[0].Path.Value, `"`) != "github.com/araihu/goshtoso-charts/components/interactive/graph" {
		t.Fatalf("parent imports = %v, want only canonical Graph package", file.Imports)
	}
	functions := 0
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			functions++
			if declaration.Name.Name != "Graph" || declaration.Body == nil || len(declaration.Body.List) != 1 {
				t.Errorf("parent function %s is not the single Graph forwarder", declaration.Name.Name)
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
		t.Errorf("parent functions = %d, want only Graph", functions)
	}
}

func TestGraphSharedTemplateRuntimeAndUpstreamProvenance(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"graph.templ", "graph_templ.go", filepath.Join("..", "graph.templ"), filepath.Join("..", "graph_templ.go")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Graph unexpectedly owns component-specific template %s", path)
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
			t.Fatalf("read shared provenance file %s: %v", path, err)
		}
		digest := sha256.Sum256(contents)
		if got := hex.EncodeToString(digest[:]); got != want {
			t.Errorf("shared provenance file %s SHA-256 = %s, want %s", path, got, want)
		}
	}
	attributions, err := os.ReadFile(filepath.Join("..", "..", "..", "site", "internal", "pages", "attributions.go"))
	if err != nil {
		t.Fatalf("read central upstream provenance: %v", err)
	}
	if !bytes.Contains(attributions, []byte("examples/graph.go")) || !bytes.Contains(attributions, []byte("bda428480a82d6d77ebb9fa939cf8d52528453dd")) {
		t.Fatal("central upstream provenance no longer identifies the pinned interactive examples source")
	}
}

func normalizedRender(t *testing.T, instance chart.Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	match := graphIDPattern.FindStringSubmatch(output.String())
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
	_ func(interactivegraph.Config) chart.Instance = interactivegraph.Graph
	_ func(interactive.GraphConfig) chart.Instance = interactive.Graph
	_ func(interactivegraph.Config) chart.Instance = interactive.Graph
	_ func(interactive.GraphConfig) chart.Instance = interactivegraph.Graph
)
