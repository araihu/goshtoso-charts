package themeriver_test

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
	"time"

	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/interactive"
	interactivethemeriver "github.com/araihu/goshtoso-charts/components/interactive/themeriver"
)

var chartIDPattern = regexp.MustCompile(`id="([A-Za-z]{12})"`)

func TestThemeRiverSpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()
	const wantPackage = "github.com/araihu/goshtoso-charts/components/interactive/themeriver"
	canonicalTypes := []reflect.Type{
		reflect.TypeOf(interactivethemeriver.Config{}),
		reflect.TypeOf(interactivethemeriver.Stream{}),
		reflect.TypeOf(interactivethemeriver.Point{}),
		reflect.TypeOf(interactivethemeriver.Layout{}),
		reflect.TypeOf(interactivethemeriver.BoundaryGap{}),
	}
	for _, typ := range canonicalTypes {
		if got := typ.PkgPath(); got != wantPackage {
			t.Errorf("%s PkgPath() = %q, want %q", typ, got, wantPackage)
		}
	}
	compatibilityTypes := []reflect.Type{
		reflect.TypeOf(interactive.ThemeRiverConfig{}),
		reflect.TypeOf(interactive.ThemeRiverStream{}),
		reflect.TypeOf(interactive.ThemeRiverPoint{}),
		reflect.TypeOf(interactive.ThemeRiverLayout{}),
		reflect.TypeOf(interactive.ThemeRiverBoundaryGap{}),
	}
	for index, typ := range compatibilityTypes {
		if typ != canonicalTypes[index] {
			t.Errorf("compatibility type %s is not identical to canonical %s", typ, canonicalTypes[index])
		}
	}
	configType := reflect.TypeOf(interactivethemeriver.Config{})
	options, _ := configType.FieldByName("Options")
	labelOptions, _ := configType.FieldByName("LabelOptions")
	if options.Type != reflect.TypeOf(chart.ChartOptions{}) || labelOptions.Type != reflect.TypeOf((*chart.LabelOptions)(nil)) {
		t.Fatalf("shared fields are not owned by chart: Options=%v LabelOptions=%v", options.Type, labelOptions.Type)
	}
}

func TestThemeRiverFacadePreservesCanonicalRenderAndValidation(t *testing.T) {
	t.Parallel()
	d0 := time.Date(2015, time.November, 8, 0, 0, 0, 0, time.UTC)
	d1 := d0.AddDate(0, 0, 1)
	cfg := interactivethemeriver.Config{
		Label:   "ThemeRiver-SingleAxis-Time",
		Streams: []interactivethemeriver.Stream{{Name: "DQ", Points: []interactivethemeriver.Point{{Time: d0, Value: 10}, {Time: d1, Value: 15}}}},
	}
	canonical := normalizedRender(t, interactivethemeriver.ThemeRiver(cfg))
	legacy := normalizedRender(t, interactive.ThemeRiver(cfg))
	if canonical != legacy {
		t.Fatal("canonical ThemeRiver render differs from compatibility facade")
	}
	invalid := cfg
	invalid.Streams[0].Points[1].Time = invalid.Streams[0].Points[0].Time
	canonicalError := renderError(interactivethemeriver.ThemeRiver(invalid))
	legacyError := renderError(interactive.ThemeRiver(invalid))
	if canonicalError != `theme river chart stream "DQ" dates must be strictly increasing` || canonicalError != legacyError {
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

func TestCompatibilityParentContainsOnlyAliasesAndForwarder(t *testing.T) {
	t.Parallel()
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "theme_river.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse parent facade: %v", err)
	}
	if len(file.Imports) != 1 || strings.Trim(file.Imports[0].Path.Value, `"`) != "github.com/araihu/goshtoso-charts/components/interactive/themeriver" {
		t.Fatalf("parent imports = %v, want only canonical ThemeRiver package", file.Imports)
	}
	functions := 0
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			functions++
			if declaration.Name.Name != "ThemeRiver" || declaration.Body == nil || len(declaration.Body.List) != 1 {
				t.Errorf("parent function %s is not the single ThemeRiver forwarder", declaration.Name.Name)
			}
		case *ast.GenDecl:
			switch declaration.Tok {
			case token.IMPORT:
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
		t.Errorf("parent functions = %d, want only ThemeRiver", functions)
	}
}

func TestThemeRiverTemplateRuntimeAndUpstreamProvenance(t *testing.T) {
	t.Parallel()
	for _, oldPath := range []string{filepath.Join("..", "theme_river.templ"), filepath.Join("..", "theme_river_templ.go")} {
		if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
			t.Errorf("compatibility parent unexpectedly retains %s", oldPath)
		}
	}
	checks := []struct {
		path       string
		normalize  *strings.Replacer
		wantDigest string
	}{
		{"themeriver.templ", strings.NewReplacer("package themeriver", "package interactive"), "b0994f8a3e5dff363bdd365c9cb8bde976c8e7bf2e1a2692a2cfc3ea7ab28c7a"},
		{"themeriver_templ.go", strings.NewReplacer("package themeriver", "package interactive", "components/interactive/themeriver/themeriver.templ", "components/interactive/theme_river.templ"), "57ef1b0f0dad2cac39f62e73020ba9ff8095aa8587f925c12db4361cfcfc97b0"},
	}
	for _, check := range checks {
		contents, err := os.ReadFile(check.path)
		if err != nil {
			t.Fatalf("read %s: %v", check.path, err)
		}
		digest := sha256.Sum256([]byte(check.normalize.Replace(string(contents))))
		if got := hex.EncodeToString(digest[:]); got != check.wantDigest {
			t.Errorf("normalized %s SHA-256 = %s, want %s", check.path, got, check.wantDigest)
		}
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
	attributions, err := os.ReadFile(filepath.Join("..", "..", "..", "site", "internal", "pages", "attributions.go"))
	if err != nil {
		t.Fatalf("read central upstream provenance: %v", err)
	}
	if !bytes.Contains(attributions, []byte("examples/themeriver.go")) || !bytes.Contains(attributions, []byte("bda428480a82d6d77ebb9fa939cf8d52528453dd")) {
		t.Fatal("central upstream provenance no longer identifies the pinned ThemeRiver example")
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

var (
	_ func(interactivethemeriver.Config) chart.Instance = interactivethemeriver.ThemeRiver
	_ func(interactive.ThemeRiverConfig) chart.Instance = interactive.ThemeRiver
	_ func(interactivethemeriver.Config) chart.Instance = interactive.ThemeRiver
	_ func(interactive.ThemeRiverConfig) chart.Instance = interactivethemeriver.ThemeRiver
)
