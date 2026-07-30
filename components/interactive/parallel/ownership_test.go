package parallel_test

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

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/interactive"
	interactiveparallel "github.com/araihu/goshtoso-charts/components/interactive/parallel"
)

func TestParallelSpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()
	const canonicalPackage = "github.com/araihu/goshtoso-charts/components/interactive/parallel"
	canonicalTypes := []reflect.Type{
		reflect.TypeOf(interactiveparallel.Scale("")),
		reflect.TypeOf(interactiveparallel.NameLocation("")),
		reflect.TypeOf(interactiveparallel.Range{}),
		reflect.TypeOf(interactiveparallel.AxisLabel{}),
		reflect.TypeOf(interactiveparallel.AxisLine{}),
		reflect.TypeOf(interactiveparallel.Dimension{}),
		reflect.TypeOf(interactiveparallel.Layout{}),
		reflect.TypeOf(interactiveparallel.LineOptions{}),
		reflect.TypeOf(interactiveparallel.SeriesOptions{}),
		reflect.TypeOf(interactiveparallel.Value{}),
		reflect.TypeOf(interactiveparallel.Observation{}),
		reflect.TypeOf(interactiveparallel.Series{}),
		reflect.TypeOf(interactiveparallel.Config{}),
	}
	for _, typ := range canonicalTypes {
		if got := typ.PkgPath(); got != canonicalPackage {
			t.Errorf("%s PkgPath() = %q, want %q", typ, got, canonicalPackage)
		}
	}
	compatibilityTypes := []reflect.Type{
		reflect.TypeOf(interactive.ParallelScale("")),
		reflect.TypeOf(interactive.ParallelNameLocation("")),
		reflect.TypeOf(interactive.ParallelRange{}),
		reflect.TypeOf(interactive.ParallelAxisLabel{}),
		reflect.TypeOf(interactive.ParallelAxisLine{}),
		reflect.TypeOf(interactive.ParallelDimension{}),
		reflect.TypeOf(interactive.ParallelLayout{}),
		reflect.TypeOf(interactive.ParallelLineOptions{}),
		reflect.TypeOf(interactive.ParallelSeriesOptions{}),
		reflect.TypeOf(interactive.ParallelValue{}),
		reflect.TypeOf(interactive.ParallelObservation{}),
		reflect.TypeOf(interactive.ParallelSeries{}),
		reflect.TypeOf(interactive.ParallelConfig{}),
	}
	for index, typ := range compatibilityTypes {
		if typ != canonicalTypes[index] {
			t.Errorf("compatibility type %s is not identical to canonical %s", typ, canonicalTypes[index])
		}
	}

	if interactive.ParallelScaleLinear != interactiveparallel.ScaleLinear ||
		interactive.ParallelScaleLog != interactiveparallel.ScaleLog ||
		interactive.ParallelNameEnd != interactiveparallel.NameEnd ||
		interactive.ParallelNameStart != interactiveparallel.NameStart ||
		interactive.ParallelNameMiddle != interactiveparallel.NameMiddle {
		t.Fatal("legacy Parallel constants do not preserve canonical values")
	}

	configType := reflect.TypeOf(interactiveparallel.Config{})
	options, _ := configType.FieldByName("Options")
	rootAttrs, _ := configType.FieldByName("RootAttrs")
	if options.Type != reflect.TypeOf(chart.ChartOptions{}) || rootAttrs.Type != reflect.TypeOf(templ.Attributes{}) {
		t.Fatalf("shared fields have wrong ownership: Options=%v RootAttrs=%v", options.Type, rootAttrs.Type)
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

func TestCompatibilityParentContainsOnlyAliasesConstantsAndForwarders(t *testing.T) {
	t.Parallel()
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "parallel.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse parent facade: %v", err)
	}
	if len(file.Imports) != 1 || strings.Trim(file.Imports[0].Path.Value, `"`) != "github.com/araihu/goshtoso-charts/components/interactive/parallel" {
		t.Fatalf("parent imports = %v, want only canonical Parallel package", file.Imports)
	}
	wantFunctions := map[string]bool{"Parallel": false, "ParallelNumber": false, "ParallelCategory": false}
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			if _, ok := wantFunctions[declaration.Name.Name]; !ok || declaration.Body == nil || len(declaration.Body.List) != 1 {
				t.Errorf("parent function %s is not an allowed single-statement forwarder", declaration.Name.Name)
			} else {
				wantFunctions[declaration.Name.Name] = true
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
	for name, seen := range wantFunctions {
		if !seen {
			t.Errorf("parent forwarder %s is missing", name)
		}
	}
}

func TestParallelTemplateProvenanceMatchesMovedSource(t *testing.T) {
	t.Parallel()
	for _, oldPath := range []string{filepath.Join("..", "parallel.templ"), filepath.Join("..", "parallel_templ.go")} {
		if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
			t.Errorf("compatibility parent unexpectedly retains %s", oldPath)
		}
	}
	checks := []struct {
		path       string
		normalize  *strings.Replacer
		wantDigest string
	}{
		{"parallel.templ", strings.NewReplacer("package interactive", "package parallel", "ParallelDimension", "Dimension"), "766eee3fbe773fee2997938ef4f7972721f205491a08835f8faf3d477540f4b1"},
		{"parallel_templ.go", strings.NewReplacer("package interactive", "package parallel", "components/interactive/parallel.templ", "components/interactive/parallel/parallel.templ", "ParallelDimension", "Dimension"), "c0ae62216fa9bbe47845caedf0f15f1dc84b014d0ae6573819035eb1cb1b4e3e"},
	}
	for _, check := range checks {
		contents, err := os.ReadFile(check.path)
		if err != nil {
			t.Fatalf("read %s: %v", check.path, err)
		}
		normalized := check.normalize.Replace(string(contents))
		if check.path == "parallel_templ.go" {
			normalized = parallelColumn.ReplaceAllString(normalized, "Col: COL")
		}
		digest := sha256.Sum256([]byte(normalized))
		if got := hex.EncodeToString(digest[:]); got != check.wantDigest {
			t.Errorf("normalized %s SHA-256 = %s, want %s", check.path, got, check.wantDigest)
		}
	}
}

var parallelColumn = regexp.MustCompile(`Col: [0-9]+`)

var (
	_ func(interactiveparallel.Config) chart.Instance = interactiveparallel.Parallel
	_ func(interactive.ParallelConfig) chart.Instance = interactive.Parallel
	_ func(interactiveparallel.Config) chart.Instance = interactive.Parallel
	_ func(interactive.ParallelConfig) chart.Instance = interactiveparallel.Parallel
	_ func(float64) interactiveparallel.Value         = interactiveparallel.Number
	_ func(float64) interactive.ParallelValue         = interactive.ParallelNumber
	_ func(string) interactiveparallel.Value          = interactiveparallel.Category
	_ func(string) interactive.ParallelValue          = interactive.ParallelCategory
)
