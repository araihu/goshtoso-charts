package heatmap_test

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
	interactiveheatmap "github.com/araihu/goshtoso-charts/components/interactive/heatmap"
)

func TestHeatMapSpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()

	wantPackage := "github.com/araihu/goshtoso-charts/components/interactive/heatmap"
	heatMapTypes := []reflect.Type{
		reflect.TypeOf(interactiveheatmap.Config{}),
		reflect.TypeOf(interactiveheatmap.Calendar{}),
		reflect.TypeOf(interactiveheatmap.ValueRange{}),
		reflect.TypeOf(interactiveheatmap.Series{}),
		reflect.TypeOf(interactiveheatmap.Data{}),
		reflect.TypeOf(interactiveheatmap.Coordinate("")),
	}
	for _, heatMapType := range heatMapTypes {
		if got := heatMapType.PkgPath(); got != wantPackage {
			t.Errorf("%s PkgPath() = %q, want %q", heatMapType, got, wantPackage)
		}
	}

	aliases := []struct {
		legacy    reflect.Type
		canonical reflect.Type
	}{
		{reflect.TypeOf(interactive.HeatMapConfig{}), reflect.TypeOf(interactiveheatmap.Config{})},
		{reflect.TypeOf(interactive.HeatMapCalendar{}), reflect.TypeOf(interactiveheatmap.Calendar{})},
		{reflect.TypeOf(interactive.HeatMapValueRange{}), reflect.TypeOf(interactiveheatmap.ValueRange{})},
		{reflect.TypeOf(interactive.HeatMapSeries{}), reflect.TypeOf(interactiveheatmap.Series{})},
		{reflect.TypeOf(interactive.HeatMapData{}), reflect.TypeOf(interactiveheatmap.Data{})},
		{reflect.TypeOf(interactive.HeatMapCoordinate("")), reflect.TypeOf(interactiveheatmap.Coordinate(""))},
	}
	for _, alias := range aliases {
		if alias.legacy != alias.canonical {
			t.Errorf("legacy type %s is not exact alias of %s", alias.legacy, alias.canonical)
		}
	}

	if interactive.HeatMapCoordinateCartesian != interactiveheatmap.CoordinateCartesian ||
		interactive.HeatMapCoordinateCalendar != interactiveheatmap.CoordinateCalendar {
		t.Fatal("legacy HeatMap constants do not preserve canonical child values")
	}

	configType := reflect.TypeOf(interactiveheatmap.Config{})
	options, _ := configType.FieldByName("Options")
	seriesOptions, _ := configType.FieldByName("SeriesOptions")
	calendarType := reflect.TypeOf(interactiveheatmap.Calendar{})
	calendarOptions, _ := calendarType.FieldByName("Options")
	if options.Type != reflect.TypeOf(chart.ChartOptions{}) ||
		seriesOptions.Type != reflect.TypeOf(chart.SeriesOptions{}) ||
		calendarOptions.Type != reflect.TypeOf(chart.CalendarOptions{}) {
		t.Fatalf("shared fields are not owned by chart: Options=%v SeriesOptions=%v Calendar.Options=%v", options.Type, seriesOptions.Type, calendarOptions.Type)
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
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "heatmap.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse parent facade: %v", err)
	}
	if len(file.Imports) != 1 || strings.Trim(file.Imports[0].Path.Value, `"`) != "github.com/araihu/goshtoso-charts/components/interactive/heatmap" {
		t.Fatalf("parent imports = %v, want only canonical HeatMap package", file.Imports)
	}

	functions := 0
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			functions++
			if declaration.Name.Name != "HeatMap" || declaration.Body == nil || len(declaration.Body.List) != 1 {
				t.Errorf("parent function %s is not the single HeatMap forwarder", declaration.Name.Name)
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
		t.Errorf("parent functions = %d, want only HeatMap", functions)
	}
}

var (
	_ func(interactiveheatmap.Config) chart.Instance = interactiveheatmap.HeatMap
	_ func(interactiveheatmap.Config) chart.Instance = interactive.HeatMap
	_ func(interactive.HeatMapConfig) chart.Instance = interactiveheatmap.HeatMap
)
