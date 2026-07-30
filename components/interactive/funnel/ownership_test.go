package funnel_test

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
	interactivefunnel "github.com/araihu/goshtoso-charts/components/interactive/funnel"
)

var chartIDPattern = regexp.MustCompile(`id="([A-Za-z]{12})"`)

func TestFunnelSpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()

	const wantPackage = "github.com/araihu/goshtoso-charts/components/interactive/funnel"
	canonicalTypes := []reflect.Type{
		reflect.TypeOf(interactivefunnel.Order("")),
		reflect.TypeOf(interactivefunnel.Config{}),
		reflect.TypeOf(interactivefunnel.Series{}),
		reflect.TypeOf(interactivefunnel.Data{}),
	}
	for _, typ := range canonicalTypes {
		if got := typ.PkgPath(); got != wantPackage {
			t.Errorf("%s PkgPath() = %q, want %q", typ, got, wantPackage)
		}
	}

	compatibilityTypes := []reflect.Type{
		reflect.TypeOf(interactive.FunnelOrder("")),
		reflect.TypeOf(interactive.FunnelConfig{}),
		reflect.TypeOf(interactive.FunnelSeries{}),
		reflect.TypeOf(interactive.FunnelData{}),
	}
	for index, typ := range compatibilityTypes {
		if typ != canonicalTypes[index] {
			t.Errorf("compatibility type %s is not identical to canonical %s", typ, canonicalTypes[index])
		}
	}

	if interactive.FunnelOrderDescending != interactivefunnel.OrderDescending ||
		interactive.FunnelOrderAscending != interactivefunnel.OrderAscending ||
		interactive.FunnelOrderData != interactivefunnel.OrderData {
		t.Fatal("legacy Funnel constants do not preserve canonical child values")
	}

	configType := reflect.TypeOf(interactivefunnel.Config{})
	options, _ := configType.FieldByName("Options")
	seriesOptions, _ := configType.FieldByName("SeriesOptions")
	seriesType := reflect.TypeOf(interactivefunnel.Series{})
	perSeriesOptions, _ := seriesType.FieldByName("Options")
	if options.Type != reflect.TypeOf(chart.ChartOptions{}) ||
		seriesOptions.Type != reflect.TypeOf(chart.SeriesOptions{}) ||
		perSeriesOptions.Type != reflect.TypeOf(chart.SeriesOptions{}) {
		t.Fatalf("shared fields are not owned by chart: Options=%v SeriesOptions=%v Series.Options=%v", options.Type, seriesOptions.Type, perSeriesOptions.Type)
	}
}

func TestFunnelFacadePreservesCanonicalRender(t *testing.T) {
	t.Parallel()
	cfg := interactivefunnel.Config{
		Label: "Checkout funnel", Order: interactivefunnel.OrderAscending,
		Series: []interactivefunnel.Series{{Name: "Checkout", Data: []interactivefunnel.Data{{Name: "Visit", Value: 100}, {Name: "Payment", Value: 24}}}},
	}
	canonical := normalizedRender(t, interactivefunnel.Funnel(cfg))
	legacy := normalizedRender(t, interactive.Funnel(cfg))
	if canonical != legacy {
		t.Fatal("canonical Funnel render differs from compatibility facade")
	}

	invalid := interactivefunnel.Config{Label: "Pipeline", Order: "sideways", Series: []interactivefunnel.Series{{Name: "Pipeline", Data: []interactivefunnel.Data{{Name: "Lead", Value: 10}}}}}
	canonicalError := renderError(interactivefunnel.Funnel(invalid))
	legacyError := renderError(interactive.Funnel(invalid))
	if canonicalError != `funnel chart order "sideways" is not supported` || canonicalError != legacyError {
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
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "funnel.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse parent facade: %v", err)
	}
	if len(file.Imports) != 1 || strings.Trim(file.Imports[0].Path.Value, `"`) != "github.com/araihu/goshtoso-charts/components/interactive/funnel" {
		t.Fatalf("parent imports = %v, want only canonical Funnel package", file.Imports)
	}

	functions := 0
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			functions++
			if declaration.Name.Name != "Funnel" || declaration.Body == nil || len(declaration.Body.List) != 1 {
				t.Errorf("parent function %s is not the single Funnel forwarder", declaration.Name.Name)
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
		t.Errorf("parent functions = %d, want only Funnel", functions)
	}
}

func TestFunnelTemplateAndRuntimeProvenance(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"funnel.templ", "funnel_templ.go"} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("canonical child lacks %s: %v", path, err)
		}
	}
	for _, path := range []string{filepath.Join("..", "funnel.templ"), filepath.Join("..", "funnel_templ.go")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("compatibility parent unexpectedly retains %s", path)
		}
	}
	generated, err := os.ReadFile("funnel_templ.go")
	if err != nil {
		t.Fatalf("read generated child template: %v", err)
	}
	if !bytes.Contains(generated, []byte("package funnel")) || !bytes.Contains(generated, []byte("components/interactive/funnel/funnel.templ")) {
		t.Fatal("generated child template does not record canonical package and source path")
	}

	for path, want := range map[string]string{
		filepath.Join("..", "..", "internal", "interactive", "theme_runtime.go"):     "07607b72118cf2e2e1cc71d81ec3c64789dd2f053ff0f9282a2e41f92cbf24ae",
		filepath.Join("..", "..", "internal", "interactive", "interactive.templ"):    "96137171bb2e6cb69372a59596963d0f4e7e6f87f3079863ccc92f3f59f8680a",
		filepath.Join("..", "..", "internal", "interactive", "interactive_templ.go"): "9e83e969108d203fe38fce502a21fed7c0f85d150cd8978d4ca76e596d278808",
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

	coverage, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "upstream-example-coverage.md"))
	if err != nil {
		t.Fatalf("read upstream provenance: %v", err)
	}
	if !bytes.Contains(coverage, []byte("## Interactive Funnel")) || !bytes.Contains(coverage, []byte("Source file: `examples/funnel.go`")) {
		t.Fatal("central upstream provenance no longer identifies the authoritative Funnel example")
	}
}

func render(t *testing.T, instance chart.Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return output.String()
}

func normalizedRender(t *testing.T, instance chart.Instance) string {
	t.Helper()
	markup := render(t, instance)
	match := chartIDPattern.FindStringSubmatch(markup)
	if len(match) != 2 {
		t.Fatalf("rendered markup lacks chart ID: %s", markup)
	}
	return strings.ReplaceAll(markup, match[1], "CHARTID")
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
	_ func(interactivefunnel.Config) chart.Instance = interactivefunnel.Funnel
	_ func(interactive.FunnelConfig) chart.Instance = interactive.Funnel
	_ func(interactivefunnel.Config) chart.Instance = interactive.Funnel
	_ func(interactive.FunnelConfig) chart.Instance = interactivefunnel.Funnel
)
