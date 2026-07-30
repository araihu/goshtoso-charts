package treemap_test

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

	"github.com/a-h/templ"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/interactive"
	interactivetreemap "github.com/araihu/goshtoso-charts/components/interactive/treemap"
)

var treemapIDPattern = regexp.MustCompile(`id="([A-Za-z]{12})"`)

func TestTreemapSpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()
	const wantPackage = "github.com/araihu/goshtoso-charts/components/interactive/treemap"
	canonicalTypes := []reflect.Type{
		reflect.TypeOf(interactivetreemap.Navigation("")),
		reflect.TypeOf(interactivetreemap.Roam(0)),
		reflect.TypeOf(interactivetreemap.Breadcrumb{}),
		reflect.TypeOf(interactivetreemap.NodeStyle{}),
		reflect.TypeOf(interactivetreemap.ColorRange{}),
		reflect.TypeOf(interactivetreemap.Level{}),
		reflect.TypeOf(interactivetreemap.Config{}),
		reflect.TypeOf(interactivetreemap.Node{}),
	}
	for _, typ := range canonicalTypes {
		if got := typ.PkgPath(); got != wantPackage {
			t.Errorf("%s PkgPath() = %q, want %q", typ, got, wantPackage)
		}
	}
	compatibilityTypes := []reflect.Type{
		reflect.TypeOf(interactive.TreemapNavigation("")),
		reflect.TypeOf(interactive.TreemapRoam(0)),
		reflect.TypeOf(interactive.TreemapBreadcrumb{}),
		reflect.TypeOf(interactive.TreemapNodeStyle{}),
		reflect.TypeOf(interactive.TreemapColorRange{}),
		reflect.TypeOf(interactive.TreemapLevel{}),
		reflect.TypeOf(interactive.TreemapConfig{}),
		reflect.TypeOf(interactive.TreemapNode{}),
	}
	for index, typ := range compatibilityTypes {
		if typ != canonicalTypes[index] {
			t.Errorf("compatibility type %s is not identical to canonical %s", typ, canonicalTypes[index])
		}
	}
	if interactive.TreemapNavigationDrillDown != interactivetreemap.NavigationDrillDown ||
		interactive.TreemapNavigationDisabled != interactivetreemap.NavigationDisabled ||
		interactive.TreemapRoamDisabled != interactivetreemap.RoamDisabled || interactive.TreemapRoamEnabled != interactivetreemap.RoamEnabled {
		t.Fatal("legacy Treemap constants do not preserve canonical values")
	}

	configType := reflect.TypeOf(interactivetreemap.Config{})
	options, _ := configType.FieldByName("Options")
	labelOptions, _ := configType.FieldByName("LabelOptions")
	upperLabel, _ := configType.FieldByName("UpperLabel")
	rootAttrs, _ := configType.FieldByName("RootAttrs")
	levelUpperLabel, _ := reflect.TypeOf(interactivetreemap.Level{}).FieldByName("UpperLabel")
	if options.Type != reflect.TypeOf(chart.ChartOptions{}) || labelOptions.Type != reflect.TypeOf((*chart.LabelOptions)(nil)) ||
		upperLabel.Type != reflect.TypeOf((*chart.LabelOptions)(nil)) || levelUpperLabel.Type != reflect.TypeOf((*chart.LabelOptions)(nil)) ||
		rootAttrs.Type != reflect.TypeOf(templ.Attributes{}) {
		t.Fatalf("shared fields have wrong ownership: Options=%v LabelOptions=%v UpperLabel=%v Level.UpperLabel=%v RootAttrs=%v", options.Type, labelOptions.Type, upperLabel.Type, levelUpperLabel.Type, rootAttrs.Type)
	}
}

func TestTreemapFacadePreservesNavigationRenderAndValidation(t *testing.T) {
	t.Parallel()
	cfg := interactivetreemap.Config{
		Label: "Hierarchy", Nodes: []*interactivetreemap.Node{{Name: "directory", Children: []*interactivetreemap.Node{{Name: "file", Value: 1}}}},
		Navigation: interactivetreemap.NavigationDisabled, Roam: interactivetreemap.RoamEnabled,
	}
	canonical := normalizedRender(t, interactivetreemap.Treemap(cfg))
	legacy := normalizedRender(t, interactive.Treemap(cfg))
	if canonical != legacy {
		t.Fatal("canonical Treemap render differs from compatibility facade")
	}
	for _, token := range []string{`"nodeClick":false`, `"roam":true`, `>Exact hierarchy and values</summary>`, `directory / file`} {
		if !strings.Contains(canonical, token) {
			t.Errorf("canonical Treemap render lacks preserved token %s", token)
		}
	}
	invalid := cfg
	invalid.Navigation = "link"
	canonicalError := renderError(interactivetreemap.Treemap(invalid))
	legacyError := renderError(interactive.Treemap(invalid))
	if canonicalError != `treemap chart navigation "link" is not supported` || canonicalError != legacyError {
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
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "treemap.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse parent facade: %v", err)
	}
	if len(file.Imports) != 1 || strings.Trim(file.Imports[0].Path.Value, `"`) != "github.com/araihu/goshtoso-charts/components/interactive/treemap" {
		t.Fatalf("parent imports = %v, want only canonical Treemap package", file.Imports)
	}
	functions := 0
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			functions++
			if declaration.Name.Name != "Treemap" || declaration.Body == nil || len(declaration.Body.List) != 1 {
				t.Errorf("parent function %s is not the single Treemap forwarder", declaration.Name.Name)
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
		t.Errorf("parent functions = %d, want only Treemap", functions)
	}
}

func TestTreemapTemplateRuntimeAndUpstreamProvenance(t *testing.T) {
	t.Parallel()
	for _, oldPath := range []string{filepath.Join("..", "treemap.templ"), filepath.Join("..", "treemap_templ.go")} {
		if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
			t.Errorf("compatibility parent unexpectedly retains %s", oldPath)
		}
	}
	checks := []struct {
		path       string
		normalize  *strings.Replacer
		wantDigest string
	}{
		{"treemap.templ", strings.NewReplacer("package treemap", "package interactive"), "0afd301a38bdde2f84106d607b084645b7dd3db0d476c98586f0e93a2a0f6622"},
		{"treemap_templ.go", strings.NewReplacer("package treemap", "package interactive", "components/interactive/treemap/treemap.templ", "components/interactive/treemap.templ"), "f81fe22c6a06707cd08b25cc5b6134b85f7aac1596060c8948a7a0284f5e813b"},
	}
	for _, check := range checks {
		contents, err := os.ReadFile(check.path)
		if err != nil {
			t.Fatalf("read %s: %v", check.path, err)
		}
		digest := sha256.Sum256([]byte(check.normalize.Replace(string(contents))))
		if got := hex.EncodeToString(digest[:]); got != check.wantDigest {
			t.Errorf("normalized %s SHA-256 = %s, want base %s", check.path, got, check.wantDigest)
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
	if !bytes.Contains(attributions, []byte("examples/treemap.go")) || !bytes.Contains(attributions, []byte("bda428480a82d6d77ebb9fa939cf8d52528453dd")) {
		t.Fatal("central upstream provenance no longer identifies the pinned Treemap example")
	}
}

func normalizedRender(t *testing.T, instance chart.Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error: %v", err)
	}
	match := treemapIDPattern.FindStringSubmatch(output.String())
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
	_ func(interactivetreemap.Config) chart.Instance = interactivetreemap.Treemap
	_ func(interactive.TreemapConfig) chart.Instance = interactive.Treemap
	_ func(interactivetreemap.Config) chart.Instance = interactive.Treemap
	_ func(interactive.TreemapConfig) chart.Instance = interactivetreemap.Treemap
)
