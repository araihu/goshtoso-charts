package line_test

import (
	"bytes"
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	interactive "github.com/araihu/goshtoso-charts/components/interactive"
	"github.com/araihu/goshtoso-charts/components/interactive/line"
)

func TestCanonicalAndCompatibilityPathsRenderIdentically(t *testing.T) {
	t.Parallel()
	minimum := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)

	tests := map[string]line.Config{
		"categorical live": {
			Label: "Live traffic", XAxis: []string{"Mon", "Tue"},
			Series: []line.Series{{Name: "API", Data: []line.Data{{Value: 12}, {Value: 18}}}},
			Live:   &interactive.LiveData{URL: "/events", Event: "traffic"},
		},
		"time axis": {
			Label: "Temporal traffic",
			TimeAxis: &line.TimeAxis{Minimum: minimum, Values: []time.Time{
				minimum.Add(time.Hour), minimum.Add(2 * time.Hour),
			}},
			Series: []line.Series{{Name: "API", Data: []line.Data{{Value: 12}, {Value: 18}}}},
		},
		"value axis with scale and references": {
			Label: "Measured traffic", ValueAxis: &line.ValueAxis{Values: []float64{0, 1}},
			Series: []line.Series{{
				Name: "API", Data: []line.Data{{Value: 12}, {Value: 18}},
				References: line.References{
					Points: []line.PointReference{{Name: "Maximum", Statistic: line.StatisticMaximum}},
					Lines:  []line.GuideReference{{Name: "Average", Statistic: line.StatisticAverage}},
					Areas:  []line.RangeReference{{Name: "Window", StartX: 0, EndX: 1}},
				},
			}},
			VisualScale: &line.VisualScale{
				Dimension: line.VisualDimensionY,
				Pieces:    []line.VisualPiece{{GreaterThan: interactive.Float(10)}},
			},
		},
	}

	for name, cfg := range tests {
		cfg := cfg
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var compatibilityConfig interactive.LineConfig = cfg
			var canonicalConfig line.Config = compatibilityConfig

			canonical := line.Line(canonicalConfig)
			compatibility := interactive.Line(compatibilityConfig)
			if canonical.Kind() != chartcomponents.KindInteractiveLine || canonical.Kind() != compatibility.Kind() {
				t.Fatalf("Kind() canonical = %q, compatibility = %q", canonical.Kind(), compatibility.Kind())
			}

			canonicalMarkup := render(t, canonical)
			compatibilityMarkup := render(t, compatibility)
			if normalizedMarkup(canonicalMarkup) != normalizedMarkup(compatibilityMarkup) {
				t.Fatalf("canonical and compatibility markup differ\ncanonical:\n%s\ncompatibility:\n%s", canonicalMarkup, compatibilityMarkup)
			}
		})
	}
}

func TestCanonicalAndCompatibilityPathsPreserveValidation(t *testing.T) {
	t.Parallel()
	cfg := line.Config{
		Label:  "Invalid traffic",
		XAxis:  []string{"Mon"},
		Series: []line.Series{{Name: "API", Data: []line.Data{{Value: 12, Symbol: "star"}}}},
	}

	var canonicalOutput, compatibilityOutput bytes.Buffer
	canonicalError := line.Line(cfg).Render(context.Background(), &canonicalOutput)
	compatibilityError := interactive.Line(cfg).Render(context.Background(), &compatibilityOutput)
	if canonicalError == nil || compatibilityError == nil {
		t.Fatalf("Render() errors canonical = %v, compatibility = %v", canonicalError, compatibilityError)
	}
	if canonicalError.Error() != compatibilityError.Error() {
		t.Fatalf("Render() error canonical = %q, compatibility = %q", canonicalError, compatibilityError)
	}
	if canonicalOutput.Len() != 0 || compatibilityOutput.Len() != 0 {
		t.Fatalf("invalid render wrote canonical = %d bytes, compatibility = %d bytes", canonicalOutput.Len(), compatibilityOutput.Len())
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

func TestLineSpecificTypesAreOwnedByCanonicalPackage(t *testing.T) {
	t.Parallel()
	const canonicalPackage = "github.com/araihu/goshtoso-charts/components/interactive/line"

	canonicalTypes := []reflect.Type{
		reflect.TypeOf(line.Config{}),
		reflect.TypeOf(line.TimeAxis{}),
		reflect.TypeOf(line.ValueAxis{}),
		reflect.TypeOf(line.VisualDimension("")),
		reflect.TypeOf(line.VisualScale{}),
		reflect.TypeOf(line.VisualPiece{}),
		reflect.TypeOf(line.Statistic("")),
		reflect.TypeOf(line.Coordinate{}),
		reflect.TypeOf(line.PointReference{}),
		reflect.TypeOf(line.GuideReference{}),
		reflect.TypeOf(line.RangeReference{}),
		reflect.TypeOf(line.References{}),
		reflect.TypeOf(line.Series{}),
		reflect.TypeOf(line.Data{}),
	}
	for _, typ := range canonicalTypes {
		if got := typ.PkgPath(); got != canonicalPackage {
			t.Errorf("%s PkgPath() = %q, want %q", typ, got, canonicalPackage)
		}
	}

	compatibilityTypes := []reflect.Type{
		reflect.TypeOf(interactive.LineConfig{}),
		reflect.TypeOf(interactive.LineTimeAxis{}),
		reflect.TypeOf(interactive.LineValueAxis{}),
		reflect.TypeOf(interactive.LineVisualDimension("")),
		reflect.TypeOf(interactive.LineVisualScale{}),
		reflect.TypeOf(interactive.LineVisualPiece{}),
		reflect.TypeOf(interactive.LineStatistic("")),
		reflect.TypeOf(interactive.LineCoordinate{}),
		reflect.TypeOf(interactive.LinePointReference{}),
		reflect.TypeOf(interactive.LineGuideReference{}),
		reflect.TypeOf(interactive.LineRangeReference{}),
		reflect.TypeOf(interactive.LineReferences{}),
		reflect.TypeOf(interactive.LineSeries{}),
		reflect.TypeOf(interactive.LineData{}),
	}
	for index, typ := range compatibilityTypes {
		if typ != canonicalTypes[index] {
			t.Errorf("compatibility type %s is not identical to canonical %s", typ, canonicalTypes[index])
		}
	}

	configType := reflect.TypeOf(line.Config{})
	options, _ := configType.FieldByName("Options")
	seriesOptions, _ := configType.FieldByName("SeriesOptions")
	live, _ := configType.FieldByName("Live")
	if options.Type != reflect.TypeOf(chart.ChartOptions{}) || seriesOptions.Type != reflect.TypeOf(chart.SeriesOptions{}) || live.Type != reflect.TypeOf((*chart.LiveData)(nil)) {
		t.Fatalf("shared Config fields are not owned by chart: Options=%v SeriesOptions=%v Live=%v", options.Type, seriesOptions.Type, live.Type)
	}
}

func TestCompatibilityParentContainsOnlyAliasesConstantsAndForwarder(t *testing.T) {
	t.Parallel()
	files := token.NewFileSet()
	file, err := parser.ParseFile(files, filepath.Join("..", "line.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse parent facade: %v", err)
	}
	if len(file.Imports) != 1 || strings.Trim(file.Imports[0].Path.Value, `"`) != "github.com/araihu/goshtoso-charts/components/interactive/line" {
		t.Fatalf("parent imports = %v, want only canonical Line package", file.Imports)
	}

	functions := 0
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			functions++
			if declaration.Name.Name != "Line" || declaration.Body == nil || len(declaration.Body.List) != 1 {
				t.Errorf("parent function %s is not the single Line forwarder", declaration.Name.Name)
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
		t.Errorf("parent functions = %d, want only Line", functions)
	}
}

func render(t *testing.T, instance interactive.Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return output.String()
}

func normalizedMarkup(markup string) string {
	const idPrefix = `goecharts_`
	start := strings.Index(markup, idPrefix)
	if start < 0 {
		return markup
	}
	start += len(idPrefix)
	const chartIDLength = 12
	if len(markup) < start+chartIDLength {
		return markup
	}
	id := markup[start : start+chartIDLength]
	return strings.ReplaceAll(markup, id, "chart-id")
}
