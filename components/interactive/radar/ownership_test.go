package radar_test

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
	interactiveradar "github.com/araihu/goshtoso-charts/components/interactive/radar"
)

func TestRadarSpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()

	const wantPackage = "github.com/araihu/goshtoso-charts/components/interactive/radar"
	canonicalTypes := []reflect.Type{
		reflect.TypeOf(interactiveradar.Config{}),
		reflect.TypeOf(interactiveradar.Shape("")),
		reflect.TypeOf(interactiveradar.CoordinateOptions{}),
		reflect.TypeOf(interactiveradar.SplitLineOptions{}),
		reflect.TypeOf(interactiveradar.Indicator{}),
		reflect.TypeOf(interactiveradar.Series{}),
		reflect.TypeOf(interactiveradar.Data{}),
	}
	for _, typ := range canonicalTypes {
		if got := typ.PkgPath(); got != wantPackage {
			t.Errorf("%s PkgPath() = %q, want %q", typ, got, wantPackage)
		}
	}

	compatibilityTypes := []reflect.Type{
		reflect.TypeOf(interactive.RadarConfig{}),
		reflect.TypeOf(interactive.RadarShape("")),
		reflect.TypeOf(interactive.RadarCoordinateOptions{}),
		reflect.TypeOf(interactive.RadarSplitLineOptions{}),
		reflect.TypeOf(interactive.RadarIndicator{}),
		reflect.TypeOf(interactive.RadarSeries{}),
		reflect.TypeOf(interactive.RadarData{}),
	}
	for index, typ := range compatibilityTypes {
		if typ != canonicalTypes[index] {
			t.Errorf("compatibility type %s is not identical to canonical %s", typ, canonicalTypes[index])
		}
	}

	if interactive.RadarShapeDefault != interactiveradar.ShapeDefault ||
		interactive.RadarShapePolygon != interactiveradar.ShapePolygon ||
		interactive.RadarShapeCircle != interactiveradar.ShapeCircle {
		t.Fatal("legacy Radar constants do not preserve canonical child values")
	}

	configType := reflect.TypeOf(interactiveradar.Config{})
	options, _ := configType.FieldByName("Options")
	seriesOptions, _ := configType.FieldByName("SeriesOptions")
	seriesType := reflect.TypeOf(interactiveradar.Series{})
	perSeriesOptions, _ := seriesType.FieldByName("Options")
	splitLineType := reflect.TypeOf(interactiveradar.SplitLineOptions{})
	lineStyle, _ := splitLineType.FieldByName("Style")
	if options.Type != reflect.TypeOf(chart.ChartOptions{}) ||
		seriesOptions.Type != reflect.TypeOf(chart.SeriesOptions{}) ||
		perSeriesOptions.Type != reflect.TypeOf(chart.SeriesOptions{}) ||
		lineStyle.Type != reflect.TypeOf((*chart.LineStyle)(nil)) {
		t.Fatalf("shared fields are not owned by chart: Options=%v SeriesOptions=%v Series.Options=%v SplitLine.Style=%v", options.Type, seriesOptions.Type, perSeriesOptions.Type, lineStyle.Type)
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
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "radar.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse parent facade: %v", err)
	}
	if len(file.Imports) != 1 || strings.Trim(file.Imports[0].Path.Value, `"`) != "github.com/araihu/goshtoso-charts/components/interactive/radar" {
		t.Fatalf("parent imports = %v, want only canonical Radar package", file.Imports)
	}

	functions := 0
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			functions++
			if declaration.Name.Name != "Radar" || declaration.Body == nil || len(declaration.Body.List) != 1 {
				t.Errorf("parent function %s is not the single Radar forwarder", declaration.Name.Name)
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
		t.Errorf("parent functions = %d, want only Radar", functions)
	}
}

var (
	_ func(interactiveradar.Config) chart.Instance = interactiveradar.Radar
	_ func(interactiveradar.Config) chart.Instance = interactive.Radar
	_ func(interactive.RadarConfig) chart.Instance = interactiveradar.Radar
)
