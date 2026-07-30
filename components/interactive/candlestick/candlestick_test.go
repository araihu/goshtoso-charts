package candlestick_test

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
	interactive "github.com/araihu/goshtoso-charts/components/interactive"
	interactivecandlestick "github.com/araihu/goshtoso-charts/components/interactive/candlestick"
)

var chartIDPattern = regexp.MustCompile(`goecharts_([A-Za-z0-9]{12})`)

func TestCandlestickPreservesLegacyRenderContract(t *testing.T) {
	t.Parallel()
	cfg := validConfig()
	var legacyConfig interactive.CandlestickConfig = cfg
	var canonicalConfig interactivecandlestick.Config = legacyConfig

	canonical := interactivecandlestick.Candlestick(canonicalConfig)
	legacy := interactive.Candlestick(legacyConfig)
	if canonical.Kind() != chartcomponents.KindInteractiveCandlestick || canonical.Kind() != legacy.Kind() {
		t.Fatalf("canonical Kind() = %q, legacy Kind() = %q", canonical.Kind(), legacy.Kind())
	}

	canonicalMarkup := render(t, canonical)
	legacyMarkup := render(t, legacy)
	if canonicalMarkup != legacyMarkup {
		t.Fatalf("canonical render differs from legacy render\ncanonical: %s\nlegacy: %s", canonicalMarkup, legacyMarkup)
	}
	digest := sha256.Sum256([]byte(canonicalMarkup))
	if got, want := hex.EncodeToString(digest[:]), "83cce5c8539173b3c5eb0e014648f260d65947acb1792389f63439920e2a5298"; got != want {
		t.Fatalf("normalized render SHA-256 = %s, want %s", got, want)
	}
}

func TestCandlestickPreservesLegacyValidation(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		mutate    func(*interactivecandlestick.Config)
		wantError string
	}{
		"label": {
			mutate:    func(cfg *interactivecandlestick.Config) { cfg.Label = "" },
			wantError: "candlestick label is required",
		},
		"ohlc low": {
			mutate:    func(cfg *interactivecandlestick.Config) { cfg.Series[0].Data[0].Low = 2400 },
			wantError: `candlestick series "Prices" candle 0 low must not exceed open or close`,
		},
		"zoom order": {
			mutate: func(cfg *interactivecandlestick.Config) {
				cfg.DataZoom = []interactivecandlestick.DataZoom{{StartPercent: 80, EndPercent: 20}}
			},
			wantError: "candlestick data zoom 0 start must not exceed end",
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cfg := validConfig()
			test.mutate(&cfg)
			canonicalError := renderError(interactivecandlestick.Candlestick(cfg))
			legacyError := renderError(interactive.Candlestick(cfg))
			if canonicalError != test.wantError {
				t.Fatalf("canonical validation error = %q, want %q", canonicalError, test.wantError)
			}
			if canonicalError != legacyError {
				t.Fatalf("canonical validation error = %q, legacy = %q", canonicalError, legacyError)
			}
		})
	}
}

func TestCandlestickSpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()
	const canonicalPackage = "github.com/araihu/goshtoso-charts/components/interactive/candlestick"
	canonicalTypes := []reflect.Type{
		reflect.TypeOf(interactivecandlestick.Config{}),
		reflect.TypeOf(interactivecandlestick.Candle{}),
		reflect.TypeOf(interactivecandlestick.Series{}),
		reflect.TypeOf(interactivecandlestick.SeriesOptions{}),
		reflect.TypeOf(interactivecandlestick.DirectionStyle{}),
		reflect.TypeOf(interactivecandlestick.MarkOptions{}),
		reflect.TypeOf(interactivecandlestick.DataZoomType("")),
		reflect.TypeOf(interactivecandlestick.DataZoomAxis("")),
		reflect.TypeOf(interactivecandlestick.DataZoom{}),
	}
	for _, typ := range canonicalTypes {
		if got := typ.PkgPath(); got != canonicalPackage {
			t.Errorf("%s PkgPath() = %q, want %q", typ, got, canonicalPackage)
		}
	}

	compatibilityTypes := []reflect.Type{
		reflect.TypeOf(interactive.CandlestickConfig{}),
		reflect.TypeOf(interactive.Candle{}),
		reflect.TypeOf(interactive.CandlestickSeries{}),
		reflect.TypeOf(interactive.CandlestickSeriesOptions{}),
		reflect.TypeOf(interactive.CandlestickDirectionStyle{}),
		reflect.TypeOf(interactive.CandlestickMarkOptions{}),
		reflect.TypeOf(interactive.CandlestickDataZoomType("")),
		reflect.TypeOf(interactive.CandlestickDataZoomAxis("")),
		reflect.TypeOf(interactive.CandlestickDataZoom{}),
	}
	for index, typ := range compatibilityTypes {
		if typ != canonicalTypes[index] {
			t.Errorf("compatibility type %s is not identical to canonical %s", typ, canonicalTypes[index])
		}
	}

	configType := reflect.TypeOf(interactivecandlestick.Config{})
	options, ok := configType.FieldByName("Options")
	if !ok || options.Type != reflect.TypeOf(chart.ChartOptions{}) {
		t.Fatalf("Config.Options = %v/%t, want chart.ChartOptions", options.Type, ok)
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
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "candlestick.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse parent facade: %v", err)
	}
	if len(file.Imports) != 1 || strings.Trim(file.Imports[0].Path.Value, `"`) != "github.com/araihu/goshtoso-charts/components/interactive/candlestick" {
		t.Fatalf("parent imports = %v, want only canonical Candlestick package", file.Imports)
	}

	functions := 0
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			functions++
			if declaration.Name.Name != "Candlestick" || declaration.Body == nil || len(declaration.Body.List) != 1 {
				t.Errorf("parent function %s is not the single Candlestick forwarder", declaration.Name.Name)
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
		t.Errorf("parent functions = %d, want only Candlestick", functions)
	}
}

func validConfig() interactivecandlestick.Config {
	return interactivecandlestick.Config{
		Label: "Market prices", Categories: []string{"2018/1/24", "2018/1/25"},
		Series: []interactivecandlestick.Series{{Name: "Prices", Data: []interactivecandlestick.Candle{
			{Open: 2320.26, Close: 2320.26, Low: 2287.3, High: 2362.94},
			{Open: 2300, Close: 2291.3, Low: 2288.26, High: 2308.38},
		}}},
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
	_ func(interactivecandlestick.Config) chart.Instance = interactivecandlestick.Candlestick
	_ func(interactive.CandlestickConfig) chart.Instance = interactive.Candlestick
	_ func(interactivecandlestick.Config) chart.Instance = interactive.Candlestick
	_ func(interactive.CandlestickConfig) chart.Instance = interactivecandlestick.Candlestick
)
