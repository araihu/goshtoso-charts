package gauge_test

import (
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
	interactivegauge "github.com/araihu/goshtoso-charts/components/interactive/gauge"
)

func TestGaugeSpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()
	const canonicalPackage = "github.com/araihu/goshtoso-charts/components/interactive/gauge"
	canonicalTypes := []reflect.Type{
		reflect.TypeOf(interactivegauge.Variant("")),
		reflect.TypeOf(interactivegauge.LiquidShape("")),
		reflect.TypeOf(interactivegauge.LiquidDirection("")),
		reflect.TypeOf(interactivegauge.LiquidTreatment{}),
		reflect.TypeOf(interactivegauge.LiquidOutline{}),
		reflect.TypeOf(interactivegauge.LiquidBackground{}),
		reflect.TypeOf(interactivegauge.LiquidLabel{}),
		reflect.TypeOf(interactivegauge.LiquidStyle{}),
		reflect.TypeOf(interactivegauge.Config{}),
		reflect.TypeOf(interactivegauge.ScaleMode("")),
		reflect.TypeOf(interactivegauge.Scale{}),
		reflect.TypeOf(interactivegauge.ScaleStop{}),
		reflect.TypeOf(interactivegauge.Series{}),
		reflect.TypeOf(interactivegauge.ProgressOptions{}),
		reflect.TypeOf(interactivegauge.Data{}),
	}
	for _, typ := range canonicalTypes {
		if got := typ.PkgPath(); got != canonicalPackage {
			t.Errorf("%s PkgPath() = %q, want %q", typ, got, canonicalPackage)
		}
	}
	compatibilityTypes := []reflect.Type{
		reflect.TypeOf(interactive.GaugeVariant("")),
		reflect.TypeOf(interactive.GaugeLiquidShape("")),
		reflect.TypeOf(interactive.GaugeLiquidDirection("")),
		reflect.TypeOf(interactive.GaugeLiquidTreatment{}),
		reflect.TypeOf(interactive.GaugeLiquidOutline{}),
		reflect.TypeOf(interactive.GaugeLiquidBackground{}),
		reflect.TypeOf(interactive.GaugeLiquidLabel{}),
		reflect.TypeOf(interactive.GaugeLiquidStyle{}),
		reflect.TypeOf(interactive.GaugeConfig{}),
		reflect.TypeOf(interactive.GaugeScaleMode("")),
		reflect.TypeOf(interactive.GaugeScale{}),
		reflect.TypeOf(interactive.GaugeScaleStop{}),
		reflect.TypeOf(interactive.GaugeSeries{}),
		reflect.TypeOf(interactive.GaugeProgressOptions{}),
		reflect.TypeOf(interactive.GaugeData{}),
	}
	for index, typ := range compatibilityTypes {
		if typ != canonicalTypes[index] {
			t.Errorf("compatibility type %s is not identical to canonical %s", typ, canonicalTypes[index])
		}
	}

	if interactive.GaugeVariantStandard != interactivegauge.VariantStandard ||
		interactive.GaugeVariantProgress != interactivegauge.VariantProgress ||
		interactive.GaugeVariantLiquid != interactivegauge.VariantLiquid ||
		interactive.GaugeLiquidShapeCircle != interactivegauge.LiquidShapeCircle ||
		interactive.GaugeLiquidShapeRect != interactivegauge.LiquidShapeRect ||
		interactive.GaugeLiquidShapeRoundRect != interactivegauge.LiquidShapeRoundRect ||
		interactive.GaugeLiquidShapeTriangle != interactivegauge.LiquidShapeTriangle ||
		interactive.GaugeLiquidShapeDiamond != interactivegauge.LiquidShapeDiamond ||
		interactive.GaugeLiquidShapePin != interactivegauge.LiquidShapePin ||
		interactive.GaugeLiquidShapeArrow != interactivegauge.LiquidShapeArrow ||
		interactive.GaugeLiquidDirectionRight != interactivegauge.LiquidDirectionRight ||
		interactive.GaugeLiquidDirectionLeft != interactivegauge.LiquidDirectionLeft ||
		interactive.GaugeScaleThermal != interactivegauge.ScaleThermal ||
		interactive.GaugeScaleCustom != interactivegauge.ScaleCustom ||
		interactive.GaugeScaleSingleColor != interactivegauge.ScaleSingleColor {
		t.Fatal("legacy Gauge constants do not preserve canonical values")
	}

	configType := reflect.TypeOf(interactivegauge.Config{})
	options, _ := configType.FieldByName("Options")
	seriesOptions, _ := configType.FieldByName("SeriesOptions")
	seriesType := reflect.TypeOf(interactivegauge.Series{})
	perSeriesOptions, _ := seriesType.FieldByName("Options")
	if options.Type != reflect.TypeOf(chart.ChartOptions{}) || seriesOptions.Type != reflect.TypeOf(chart.SeriesOptions{}) || perSeriesOptions.Type != reflect.TypeOf(chart.SeriesOptions{}) {
		t.Fatalf("shared fields are not owned by chart: Options=%v SeriesOptions=%v Series.Options=%v", options.Type, seriesOptions.Type, perSeriesOptions.Type)
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
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "gauge.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse parent facade: %v", err)
	}
	if len(file.Imports) != 1 || strings.Trim(file.Imports[0].Path.Value, `"`) != "github.com/araihu/goshtoso-charts/components/interactive/gauge" {
		t.Fatalf("parent imports = %v, want only canonical Gauge package", file.Imports)
	}
	functions := 0
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			functions++
			if declaration.Name.Name != "Gauge" || declaration.Body == nil || len(declaration.Body.List) != 1 {
				t.Errorf("parent function %s is not the single Gauge forwarder", declaration.Name.Name)
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
		t.Errorf("parent functions = %d, want only Gauge", functions)
	}
}

func TestGaugeTemplateProvenanceMatchesMovedSource(t *testing.T) {
	t.Parallel()
	for _, oldPath := range []string{filepath.Join("..", "gauge.templ"), filepath.Join("..", "gauge_templ.go")} {
		if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
			t.Errorf("compatibility parent unexpectedly retains %s", oldPath)
		}
	}
	checks := []struct {
		path       string
		normalize  *strings.Replacer
		wantDigest string
	}{
		{"gauge.templ", strings.NewReplacer("package interactive", "package gauge", "gaugeLiquidExactValues", "liquidExactValues", "GaugeData", "Data"), "9353afb66900fd69a687e0fb8f3e836dbab142916b16365be9e90ca5601241a2"},
		{"gauge_templ.go", strings.NewReplacer("package interactive", "package gauge", "components/interactive/gauge.templ", "components/interactive/gauge/gauge.templ", "gaugeLiquidExactValues", "liquidExactValues", "GaugeData", "Data"), "63506236ad077eb932f3f5ab34d84b6331d0e6872630873b2a7214957d72f9b5"},
	}
	for _, check := range checks {
		contents, err := os.ReadFile(check.path)
		if err != nil {
			t.Fatalf("read %s: %v", check.path, err)
		}
		normalized := check.normalize.Replace(string(contents))
		if check.path == "gauge_templ.go" {
			normalized = regexpColumn.ReplaceAllString(normalized, "Col: COL")
		}
		digest := sha256.Sum256([]byte(normalized))
		if got := hex.EncodeToString(digest[:]); got != check.wantDigest {
			t.Errorf("normalized %s SHA-256 = %s, want %s", check.path, got, check.wantDigest)
		}
	}
}

var regexpColumn = regexp.MustCompile(`Col: [0-9]+`)

var (
	_ func(interactivegauge.Config) chart.Instance = interactivegauge.Gauge
	_ func(interactive.GaugeConfig) chart.Instance = interactive.Gauge
	_ func(interactivegauge.Config) chart.Instance = interactive.Gauge
	_ func(interactive.GaugeConfig) chart.Instance = interactivegauge.Gauge
)
