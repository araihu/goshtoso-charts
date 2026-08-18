package bar_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chart"
	interactive "github.com/araihu/goshtoso-charts/components/interactive"
	interactivebar "github.com/araihu/goshtoso-charts/components/interactive/bar"
)

var chartIDPattern = regexp.MustCompile(`id="([A-Za-z]{12})"`)

func TestBarPreservesLegacyRenderContract(t *testing.T) {
	t.Parallel()
	showLabels := true
	cfg := interactivebar.Config{
		Label:       "Quarterly revenue",
		Caption:     "Revenue by product.",
		XAxis:       []string{"Q1", "Q2"},
		Orientation: interactivebar.OrientationHorizontal,
		Zoom:        &interactivebar.Zoom{Mode: interactivebar.ZoomSlider, StartPercent: 10, EndPercent: 80},
		Series: []interactivebar.Series{{
			Name: "Hardware",
			Data: []interactivebar.Data{{Value: 12}, {Value: 18}},
			References: interactivebar.References{
				Points: []interactivebar.PointReference{
					{Name: "Maximum", Statistic: interactivebar.StatisticMaximum},
					{Name: "Q1 target", Coordinate: &interactivebar.Coordinate{Category: "Q1", Value: 12}},
				},
				Lines:      []interactivebar.GuideReference{{Name: "Average", Statistic: interactivebar.StatisticAverage}},
				ShowLabels: &showLabels,
			},
		}},
		Options: interactive.ChartOptions{Animation: interactive.Bool(false)},
	}

	var legacyConfig interactive.BarConfig = cfg
	var canonicalConfig interactivebar.Config = legacyConfig
	canonical := interactivebar.Bar(canonicalConfig)
	legacy := interactive.Bar(legacyConfig)
	if canonical.Kind() != chartcomponents.KindInteractiveBar || canonical.Kind() != legacy.Kind() {
		t.Fatalf("canonical Kind() = %q, legacy Kind() = %q", canonical.Kind(), legacy.Kind())
	}

	canonicalMarkup := render(t, canonical)
	legacyMarkup := render(t, legacy)
	if canonicalMarkup != legacyMarkup {
		t.Fatalf("canonical render differs from legacy render\ncanonical: %s\nlegacy: %s", canonicalMarkup, legacyMarkup)
	}
	digest := sha256.Sum256([]byte(canonicalMarkup))
	if got, want := hex.EncodeToString(digest[:]), "fe5c84d2ae98234af7e725e603bbdccd8b4a6644ac62ad5fec67b4fe93963f89"; got != want {
		t.Fatalf("normalized render SHA-256 = %s, want %s", got, want)
	}
}

func TestBarPreservesLegacyValidation(t *testing.T) {
	t.Parallel()
	invalid := interactivebar.Config{
		Label:       "Revenue",
		XAxis:       []string{"Q1"},
		Series:      []interactivebar.Series{{Name: "Hardware", Data: []interactivebar.Data{{Value: 12}}}},
		Orientation: interactivebar.Orientation("diagonal"),
	}

	canonicalError := renderError(interactivebar.Bar(invalid))
	legacyError := renderError(interactive.Bar(invalid))
	if canonicalError != `bar chart orientation "diagonal" is not supported` {
		t.Fatalf("canonical validation error = %q", canonicalError)
	}
	if canonicalError != legacyError {
		t.Fatalf("canonical validation error = %q, legacy = %q", canonicalError, legacyError)
	}
}

func TestBarExportsConciseChartSpecificConstants(t *testing.T) {
	t.Parallel()
	_ = []interactivebar.Orientation{interactivebar.OrientationVertical, interactivebar.OrientationHorizontal}
	_ = []interactivebar.ZoomMode{interactivebar.ZoomInside, interactivebar.ZoomSlider}
	_ = []interactivebar.Statistic{interactivebar.StatisticMinimum, interactivebar.StatisticMaximum, interactivebar.StatisticAverage}
}

func TestBarSpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()

	wantPackage := "github.com/araihu/goshtoso-charts/components/interactive/bar"
	barTypes := []reflect.Type{
		reflect.TypeOf(interactivebar.Config{}),
		reflect.TypeOf(interactivebar.Series{}),
		reflect.TypeOf(interactivebar.Data{}),
		reflect.TypeOf(interactivebar.Orientation("")),
		reflect.TypeOf(interactivebar.ZoomMode("")),
		reflect.TypeOf(interactivebar.Zoom{}),
		reflect.TypeOf(interactivebar.Statistic("")),
		reflect.TypeOf(interactivebar.Coordinate{}),
		reflect.TypeOf(interactivebar.PointReference{}),
		reflect.TypeOf(interactivebar.GuideReference{}),
		reflect.TypeOf(interactivebar.References{}),
	}
	for _, barType := range barTypes {
		if got := barType.PkgPath(); got != wantPackage {
			t.Errorf("%s PkgPath() = %q, want %q", barType, got, wantPackage)
		}
	}

	compatibilityTypes := []reflect.Type{
		reflect.TypeOf(interactive.BarConfig{}),
		reflect.TypeOf(interactive.BarSeries{}),
		reflect.TypeOf(interactive.BarData{}),
		reflect.TypeOf(interactive.BarOrientation("")),
		reflect.TypeOf(interactive.BarZoomMode("")),
		reflect.TypeOf(interactive.BarZoom{}),
		reflect.TypeOf(interactive.BarStatistic("")),
		reflect.TypeOf(interactive.BarCoordinate{}),
		reflect.TypeOf(interactive.BarPointReference{}),
		reflect.TypeOf(interactive.BarGuideReference{}),
		reflect.TypeOf(interactive.BarReferences{}),
	}
	for index, compatibilityType := range compatibilityTypes {
		if compatibilityType != barTypes[index] {
			t.Errorf("compatibility type %s is not identical to canonical %s", compatibilityType, barTypes[index])
		}
	}
}

func TestCompatibilityParentContainsOnlyAliasesConstantsAndForwarder(t *testing.T) {
	t.Parallel()
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "bar.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse parent facade: %v", err)
	}
	if len(file.Imports) != 1 || strings.Trim(file.Imports[0].Path.Value, `"`) != "github.com/araihu/goshtoso-charts/components/interactive/bar" {
		t.Fatalf("parent imports = %v, want only canonical Bar package", file.Imports)
	}

	functions := 0
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			functions++
			if declaration.Name.Name != "Bar" || declaration.Body == nil || len(declaration.Body.List) != 1 {
				t.Errorf("parent function %s is not the single Bar forwarder", declaration.Name.Name)
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
		t.Errorf("parent functions = %d, want only Bar", functions)
	}
}

var (
	_ func(interactivebar.Config) chart.Instance = interactivebar.Bar
	_ func(interactivebar.Config) chart.Instance = interactive.Bar
	_ func(interactive.BarConfig) chart.Instance = interactivebar.Bar
)

func render(t *testing.T, instance interactive.Instance) string {
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
	return strings.ReplaceAll(markup, match[1], "CHARTID")
}

func renderError(instance interactive.Instance) string {
	var output bytes.Buffer
	err := instance.Render(context.Background(), &output)
	if err == nil {
		return ""
	}
	return err.Error()
}
