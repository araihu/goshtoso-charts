package scatter_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/interactive"
	interactivescatter "github.com/araihu/goshtoso-charts/components/interactive/scatter"
)

func TestScatterSpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()

	wantPackage := "github.com/araihu/goshtoso-charts/components/interactive/scatter"
	scatterTypes := []reflect.Type{
		reflect.TypeOf(interactivescatter.Config{}),
		reflect.TypeOf(interactivescatter.Series{}),
		reflect.TypeOf(interactivescatter.Data{}),
		reflect.TypeOf(interactivescatter.Variant("")),
		reflect.TypeOf(interactivescatter.AxisType("")),
	}
	for _, scatterType := range scatterTypes {
		if got := scatterType.PkgPath(); got != wantPackage {
			t.Errorf("%s PkgPath() = %q, want %q", scatterType, got, wantPackage)
		}
	}

	aliases := []struct {
		legacy    reflect.Type
		canonical reflect.Type
	}{
		{reflect.TypeOf(interactive.ScatterConfig{}), reflect.TypeOf(interactivescatter.Config{})},
		{reflect.TypeOf(interactive.ScatterSeries{}), reflect.TypeOf(interactivescatter.Series{})},
		{reflect.TypeOf(interactive.ScatterData{}), reflect.TypeOf(interactivescatter.Data{})},
		{reflect.TypeOf(interactive.ScatterVariant("")), reflect.TypeOf(interactivescatter.Variant(""))},
		{reflect.TypeOf(interactive.CartesianAxisType("")), reflect.TypeOf(interactivescatter.AxisType(""))},
	}
	for _, alias := range aliases {
		if alias.legacy != alias.canonical {
			t.Errorf("legacy type %s is not exact alias of %s", alias.legacy, alias.canonical)
		}
	}

	if interactive.ScatterVariantStandard != interactivescatter.VariantStandard ||
		interactive.ScatterVariantEffect != interactivescatter.VariantEffect ||
		interactive.CartesianAxisCategory != interactivescatter.AxisCategory ||
		interactive.CartesianAxisValue != interactivescatter.AxisValue {
		t.Fatal("legacy Scatter constants do not preserve canonical child values")
	}

	configType := reflect.TypeOf(interactivescatter.Config{})
	options, _ := configType.FieldByName("Options")
	seriesOptions, _ := configType.FieldByName("SeriesOptions")
	ripple, _ := configType.FieldByName("Ripple")
	if options.Type != reflect.TypeOf(chart.ChartOptions{}) ||
		seriesOptions.Type != reflect.TypeOf(chart.SeriesOptions{}) ||
		ripple.Type != reflect.TypeOf((*chart.RippleOptions)(nil)) {
		t.Fatalf("shared Config fields are not owned by chart: Options=%v SeriesOptions=%v Ripple=%v", options.Type, seriesOptions.Type, ripple.Type)
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
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "scatter.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse parent facade: %v", err)
	}
	if len(file.Imports) != 1 || strings.Trim(file.Imports[0].Path.Value, `"`) != "github.com/araihu/goshtoso-charts/components/interactive/scatter" {
		t.Fatalf("parent imports = %v, want only canonical Scatter package", file.Imports)
	}

	functions := 0
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			functions++
			if declaration.Name.Name != "Scatter" || declaration.Body == nil || len(declaration.Body.List) != 1 {
				t.Errorf("parent function %s is not the single Scatter forwarder", declaration.Name.Name)
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
		t.Errorf("parent functions = %d, want only Scatter", functions)
	}
}

var (
	_ func(interactivescatter.Config) chart.Instance = interactivescatter.Scatter
	_ func(interactivescatter.Config) chart.Instance = interactive.Scatter
	_ func(interactive.ScatterConfig) chart.Instance = interactivescatter.Scatter
)
