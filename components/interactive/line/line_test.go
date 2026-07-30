package line_test

import (
	"bytes"
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	chartcomponents "github.com/araihu/goshtoso-charts/components"
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

func TestCanonicalPackageDoesNotExposeRendererNamesOrImports(t *testing.T) {
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
			if strings.Contains(imported.Path.Value, "go-echarts") {
				t.Errorf("%s imports renderer package %s", entry.Name(), imported.Path.Value)
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
