package boxplot_test

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
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/araihu/goshtoso-charts/components/interactive"
	interactiveboxplot "github.com/araihu/goshtoso-charts/components/interactive/boxplot"
)

var chartIDPattern = regexp.MustCompile(`id="([A-Za-z]{12})"`)

func TestBoxPlotVariantsPreserveLegacyRenderContracts(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		cfg  interactiveboxplot.Config
		hash string
	}{
		"single-default": {
			cfg: interactiveboxplot.Config{
				Label: "Distribution", Categories: []string{"A"},
				Series: []interactiveboxplot.Series{{Name: "Samples", Data: []interactiveboxplot.Data{{Min: 1, Q1: 2, Median: 3, Q3: 4, Max: 5}}}},
			},
			hash: "b213621e8180ea7e33be96e6c571b73cf55e81c334e4e7d50746ef11e50f5379",
		},
		"multiple-custom": {
			cfg: interactiveboxplot.Config{
				Label: "Latency distribution", Caption: "Five-number summaries by environment.",
				Categories: []string{"Development", "Production"},
				Series: []interactiveboxplot.Series{
					{Name: "Current", Data: []interactiveboxplot.Data{
						{Name: "dev-current", Min: 10, Q1: 20, Median: 30, Q3: 45, Max: 70},
						{Name: "prod-current", Min: 15, Q1: 25, Median: 35, Q3: 50, Max: 80, ItemStyle: &chart.ItemStyle{Color: "#abcdef"}},
					}},
					{Name: "Previous", Data: []interactiveboxplot.Data{
						{Min: 12, Q1: 22, Median: 32, Q3: 48, Max: 75},
						{Min: 18, Q1: 28, Median: 40, Q3: 55, Max: 90},
					}, Options: chart.SeriesOptions{Label: &chart.LabelOptions{Show: chart.Bool(true)}}},
				},
				Width: "720px", Height: "360px",
				Options:       chart.ChartOptions{Title: &chart.TitleOptions{Text: "Latency"}},
				SeriesOptions: chart.SeriesOptions{ItemStyle: &chart.ItemStyle{BorderWidth: 2}},
				Style:         charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#123456"}, Class: "min-h-80"},
			},
			hash: "719c5bf0f1fc481cf202660e3453d0bab735d3f276805f71e1595e9431a1835d",
		},
		"point-overrides": {
			cfg: interactiveboxplot.Config{
				Label: "Custom summaries", Categories: []string{"Before", "After"},
				Series: []interactiveboxplot.Series{{Name: "Latency", Data: []interactiveboxplot.Data{
					{Name: "before", Min: -5, Q1: 0, Median: 2, Q3: 3, Max: 8, Label: &chart.LabelOptions{Show: chart.Bool(true), Position: "top", Color: "#112233", FontSize: 12}, ItemStyle: &chart.ItemStyle{Color: "#223344", BorderColor: "#334455", BorderWidth: 1.5, Opacity: chart.Float(0.75)}, Emphasis: &chart.EmphasisOptions{ItemStyle: &chart.ItemStyle{Color: "#445566"}}, Tooltip: &chart.TooltipOptions{Show: chart.Bool(true), Trigger: "item"}},
					{Name: "after", Min: -4, Q1: 1, Median: 2, Q3: 5, Max: 9},
				}, Options: chart.SeriesOptions{ItemStyle: &chart.ItemStyle{Color: "#556677"}}}},
				Options: chart.ChartOptions{Animation: chart.Bool(false), XAxis: &chart.AxisOptions{Name: "Period", LabelInterval: chart.Int(1)}},
				Style:   charttheme.Style{Palette: charttheme.PalettePastel, Class: "caller-boxplot"},
			},
			hash: "9a48e7d0173150aa33ce79860ea8207a7c29766294bb81f86b356e0357077707",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var legacyConfig interactive.BoxPlotConfig = test.cfg
			var canonicalConfig interactiveboxplot.Config = legacyConfig
			canonical := interactiveboxplot.BoxPlot(canonicalConfig)
			legacy := interactive.BoxPlot(legacyConfig)
			if canonical.Kind() != legacy.Kind() {
				t.Fatalf("canonical Kind() = %q, legacy Kind() = %q", canonical.Kind(), legacy.Kind())
			}
			canonicalMarkup := render(t, canonical)
			legacyMarkup := render(t, legacy)
			if canonicalMarkup != legacyMarkup {
				t.Fatal("canonical render differs from legacy render")
			}
			digest := sha256.Sum256([]byte(canonicalMarkup))
			if got := hex.EncodeToString(digest[:]); got != test.hash {
				t.Fatalf("normalized render SHA-256 = %s, want %s", got, test.hash)
			}
		})
	}
}

func TestBoxPlotPreservesLegacyValidation(t *testing.T) {
	t.Parallel()
	cfg := interactiveboxplot.Config{Label: "Distribution", Categories: []string{"A"}, Series: []interactiveboxplot.Series{{Name: "Samples", Data: []interactiveboxplot.Data{{Min: 1, Q1: 4, Median: 3, Q3: 5, Max: 6}}}}}
	canonicalError := renderError(interactiveboxplot.BoxPlot(cfg))
	legacyError := renderError(interactive.BoxPlot(cfg))
	const want = `box plot series "Samples" summary 0 values must be ordered min, q1, median, q3, max`
	if canonicalError != want || canonicalError != legacyError {
		t.Fatalf("canonical validation error = %q, legacy = %q, want %q", canonicalError, legacyError, want)
	}
}

func TestBoxPlotSpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()
	const canonicalPackage = "github.com/araihu/goshtoso-charts/components/interactive/boxplot"
	canonicalTypes := []reflect.Type{
		reflect.TypeOf(interactiveboxplot.Config{}),
		reflect.TypeOf(interactiveboxplot.Series{}),
		reflect.TypeOf(interactiveboxplot.Data{}),
	}
	for _, typ := range canonicalTypes {
		if got := typ.PkgPath(); got != canonicalPackage {
			t.Errorf("%s PkgPath() = %q, want %q", typ, got, canonicalPackage)
		}
	}
	compatibilityTypes := []reflect.Type{
		reflect.TypeOf(interactive.BoxPlotConfig{}),
		reflect.TypeOf(interactive.BoxPlotSeries{}),
		reflect.TypeOf(interactive.BoxPlotData{}),
	}
	for index, typ := range compatibilityTypes {
		if typ != canonicalTypes[index] {
			t.Errorf("compatibility type %s is not identical to canonical %s", typ, canonicalTypes[index])
		}
	}
	configType := reflect.TypeOf(interactiveboxplot.Config{})
	options, _ := configType.FieldByName("Options")
	seriesOptions, _ := configType.FieldByName("SeriesOptions")
	seriesType := reflect.TypeOf(interactiveboxplot.Series{})
	seriesField, _ := seriesType.FieldByName("Options")
	dataType := reflect.TypeOf(interactiveboxplot.Data{})
	sharedDataFields := map[string]reflect.Type{
		"Label": reflect.TypeOf((*chart.LabelOptions)(nil)), "ItemStyle": reflect.TypeOf((*chart.ItemStyle)(nil)),
		"Emphasis": reflect.TypeOf((*chart.EmphasisOptions)(nil)), "Tooltip": reflect.TypeOf((*chart.TooltipOptions)(nil)),
	}
	if options.Type != reflect.TypeOf(chart.ChartOptions{}) || seriesOptions.Type != reflect.TypeOf(chart.SeriesOptions{}) || seriesField.Type != reflect.TypeOf(chart.SeriesOptions{}) {
		t.Fatalf("shared option fields are not owned by chart: Config.Options=%v Config.SeriesOptions=%v Series.Options=%v", options.Type, seriesOptions.Type, seriesField.Type)
	}
	for name, want := range sharedDataFields {
		field, _ := dataType.FieldByName(name)
		if field.Type != want {
			t.Errorf("Data.%s = %v, want %v", name, field.Type, want)
		}
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
		for _, declaration := range file.Decls {
			switch declaration := declaration.(type) {
			case *ast.FuncDecl:
				if declaration.Name.IsExported() && strings.Contains(strings.ToLower(declaration.Name.Name), "echarts") {
					t.Errorf("%s exports renderer-named function %s", entry.Name(), declaration.Name.Name)
				}
			case *ast.GenDecl:
				for _, spec := range declaration.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok && typeSpec.Name.IsExported() && strings.Contains(strings.ToLower(typeSpec.Name.Name), "echarts") {
						t.Errorf("%s exports renderer-named type %s", entry.Name(), typeSpec.Name.Name)
					}
				}
			}
		}
	}
}

func TestCompatibilityParentContainsOnlyAliasesAndForwarder(t *testing.T) {
	t.Parallel()
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "boxplot.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse parent facade: %v", err)
	}
	if len(file.Imports) != 1 || strings.Trim(file.Imports[0].Path.Value, `"`) != "github.com/araihu/goshtoso-charts/components/interactive/boxplot" {
		t.Fatalf("parent imports = %v, want only canonical BoxPlot package", file.Imports)
	}
	functions := 0
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			functions++
			if declaration.Name.Name != "BoxPlot" || declaration.Body == nil || len(declaration.Body.List) != 1 {
				t.Errorf("parent function %s is not the single BoxPlot forwarder", declaration.Name.Name)
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
		t.Errorf("parent functions = %d, want only BoxPlot", functions)
	}
}

func TestBoxPlotTemplateProvenanceRemainsInSharedPrivateShell(t *testing.T) {
	t.Parallel()
	for _, directory := range []string{"."} {
		entries, err := os.ReadDir(directory)
		if err != nil {
			t.Fatalf("read %s: %v", directory, err)
		}
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".templ") || strings.HasSuffix(entry.Name(), "_templ.go") {
				t.Errorf("BoxPlot ownership directory %s unexpectedly contains template %s", directory, entry.Name())
			}
		}
	}
	for _, path := range []string{filepath.Join("..", "boxplot.templ"), filepath.Join("..", "boxplot_templ.go")} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("compatibility parent unexpectedly retains BoxPlot template %s", path)
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
}

func render(t *testing.T, instance chart.Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	if !chartIDPattern.MatchString(markup) {
		t.Fatalf("rendered markup lacks chart ID: %s", markup)
	}
	match := chartIDPattern.FindStringSubmatch(markup)
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
	_ func(interactiveboxplot.Config) chart.Instance = interactiveboxplot.BoxPlot
	_ func(interactive.BoxPlotConfig) chart.Instance = interactive.BoxPlot
	_ func(interactiveboxplot.Config) chart.Instance = interactive.BoxPlot
	_ func(interactive.BoxPlotConfig) chart.Instance = interactiveboxplot.BoxPlot
)
