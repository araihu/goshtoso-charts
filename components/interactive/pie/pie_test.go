package pie_test

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

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	interactive "github.com/araihu/goshtoso-charts/components/interactive"
	interactivepie "github.com/araihu/goshtoso-charts/components/interactive/pie"
)

var chartIDPattern = regexp.MustCompile(`goecharts_([A-Za-z0-9]{12})`)

func TestPieVariantsPreserveLegacyRenderContracts(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		cfg  interactivepie.Config
		hash string
	}{
		"standard": {
			cfg: interactivepie.Config{
				Label: "Seasons", Series: []interactivepie.Series{{Name: "Weather", Data: []interactivepie.Data{{Name: "Spring", Value: 20}, {Name: "Summer", Value: 30}}}},
			},
			hash: "c4421526e80d0fc3b49cc2426540f3e2aad8467c1f855ed9a3333ce75d4b52af",
		},
		"donut": {
			cfg: interactivepie.Config{
				Label: "Donut", Caption: "Share by state.", Series: []interactivepie.Series{{Name: "States", InnerRadius: 40, OuterRadius: 75, LabelContent: interactivepie.LabelNameAndValue, Data: []interactivepie.Data{{Name: "Open", Value: 12}, {Name: "Closed", Value: 28}}}},
				TooltipContent: interactivepie.TooltipNameAndShare, Style: charttheme.Style{Palette: charttheme.PalettePastel},
			},
			hash: "a2e172286949ae6ca08bc7bca91ae8fb11de106d8ae3410801489221ae94e5de",
		},
		"rose": {
			cfg: interactivepie.Config{
				Label: "Rose", Width: "720px", Height: "360px", Series: []interactivepie.Series{{Name: "Incidents", InnerRadius: 30, OuterRadius: 70, Center: &interactivepie.Center{X: 25, Y: 50}, RoseMode: interactivepie.RoseArea, PadAngle: 2, Selectable: true, LabelContent: interactivepie.LabelNameAndValue, Data: []interactivepie.Data{{Name: "Open", Value: 12, Selected: true}, {Name: "Closed", Value: 28}}}},
				Options: chart.ChartOptions{Animation: chart.Bool(false)}, AutoEmphasis: &interactivepie.AutoEmphasisOptions{SeriesIndex: 0, IntervalMilliseconds: 1250}, Style: charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}},
			},
			hash: "4ed54edf548763b2f34a63f5c7abbecc818021a94d7b6ceb7ba7fa1b42b0b561",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var legacyConfig interactive.PieConfig = test.cfg
			var canonicalConfig interactivepie.Config = legacyConfig
			canonical := interactivepie.Pie(canonicalConfig)
			legacy := interactive.Pie(legacyConfig)
			if canonical.Kind() != chartcomponents.KindInteractivePie || canonical.Kind() != legacy.Kind() {
				t.Fatalf("canonical Kind() = %q, legacy Kind() = %q", canonical.Kind(), legacy.Kind())
			}
			canonicalMarkup := render(t, canonical)
			legacyMarkup := render(t, legacy)
			if canonicalMarkup != legacyMarkup {
				t.Fatalf("canonical render differs from legacy render\ncanonical: %s\nlegacy: %s", canonicalMarkup, legacyMarkup)
			}
			digest := sha256.Sum256([]byte(canonicalMarkup))
			if got := hex.EncodeToString(digest[:]); got != test.hash {
				t.Fatalf("normalized render SHA-256 = %s, want %s", got, test.hash)
			}
		})
	}
}

func TestPiePreservesLegacyValidation(t *testing.T) {
	t.Parallel()
	cfg := interactivepie.Config{
		Label: "States", Series: []interactivepie.Series{{Name: "States", RoseMode: interactivepie.RoseMode("flower"), Data: []interactivepie.Data{{Name: "Open", Value: 1}}}},
	}
	canonicalError := renderError(interactivepie.Pie(cfg))
	legacyError := renderError(interactive.Pie(cfg))
	const want = `pie chart series "States" rose mode "flower" is not supported`
	if canonicalError != want {
		t.Fatalf("canonical validation error = %q, want %q", canonicalError, want)
	}
	if canonicalError != legacyError {
		t.Fatalf("canonical validation error = %q, legacy = %q", canonicalError, legacyError)
	}
}

func TestPieSpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()
	const canonicalPackage = "github.com/araihu/goshtoso-charts/components/interactive/pie"
	canonicalTypes := []reflect.Type{
		reflect.TypeOf(interactivepie.RoseMode("")),
		reflect.TypeOf(interactivepie.LabelContent("")),
		reflect.TypeOf(interactivepie.TooltipContent("")),
		reflect.TypeOf(interactivepie.AutoEmphasisOptions{}),
		reflect.TypeOf(interactivepie.Center{}),
		reflect.TypeOf(interactivepie.Config{}),
		reflect.TypeOf(interactivepie.Series{}),
		reflect.TypeOf(interactivepie.Data{}),
	}
	for _, typ := range canonicalTypes {
		if got := typ.PkgPath(); got != canonicalPackage {
			t.Errorf("%s PkgPath() = %q, want %q", typ, got, canonicalPackage)
		}
	}
	compatibilityTypes := []reflect.Type{
		reflect.TypeOf(interactive.PieRoseMode("")),
		reflect.TypeOf(interactive.PieLabelContent("")),
		reflect.TypeOf(interactive.PieTooltipContent("")),
		reflect.TypeOf(interactive.PieAutoEmphasisOptions{}),
		reflect.TypeOf(interactive.PieCenter{}),
		reflect.TypeOf(interactive.PieConfig{}),
		reflect.TypeOf(interactive.PieSeries{}),
		reflect.TypeOf(interactive.PieData{}),
	}
	for index, typ := range compatibilityTypes {
		if typ != canonicalTypes[index] {
			t.Errorf("compatibility type %s is not identical to canonical %s", typ, canonicalTypes[index])
		}

	}
	configType := reflect.TypeOf(interactivepie.Config{})
	options, _ := configType.FieldByName("Options")
	seriesOptions, _ := configType.FieldByName("SeriesOptions")
	if options.Type != reflect.TypeOf(chart.ChartOptions{}) || seriesOptions.Type != reflect.TypeOf(chart.SeriesOptions{}) {
		t.Fatalf("shared Config fields are not owned by chart: Options=%v SeriesOptions=%v", options.Type, seriesOptions.Type)
	}
	if interactive.PieRoseNone != interactivepie.RoseNone ||
		interactive.PieRoseRadius != interactivepie.RoseRadius ||
		interactive.PieRoseArea != interactivepie.RoseArea ||
		interactive.PieLabelDefault != interactivepie.LabelDefault ||
		interactive.PieLabelNameAndValue != interactivepie.LabelNameAndValue ||
		interactive.PieTooltipDefault != interactivepie.TooltipDefault ||
		interactive.PieTooltipNameAndShare != interactivepie.TooltipNameAndShare {
		t.Fatal("legacy Pie constants do not preserve canonical child values")
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
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "pie.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse parent facade: %v", err)
	}
	if len(file.Imports) != 1 || strings.Trim(file.Imports[0].Path.Value, `"`) != "github.com/araihu/goshtoso-charts/components/interactive/pie" {
		t.Fatalf("parent imports = %v, want only canonical Pie package", file.Imports)
	}
	functions := 0
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			functions++
			if declaration.Name.Name != "Pie" || declaration.Body == nil || len(declaration.Body.List) != 1 {
				t.Errorf("parent function %s is not the single Pie forwarder", declaration.Name.Name)
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
		t.Errorf("parent functions = %d, want only Pie", functions)
	}
}

func render(t *testing.T, instance chart.Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	match := chartIDPattern.FindStringSubmatch(markup)
	if len(match) != 2 {
		t.Fatalf("rendered markup lacks chart ID: %s", markup)
	}
	return strings.ReplaceAll(markup, match[1], "chart-id")
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
	_ func(interactivepie.Config) chart.Instance = interactivepie.Pie
	_ func(interactive.PieConfig) chart.Instance = interactive.Pie
	_ func(interactivepie.Config) chart.Instance = interactive.Pie
	_ func(interactive.PieConfig) chart.Instance = interactivepie.Pie
)
