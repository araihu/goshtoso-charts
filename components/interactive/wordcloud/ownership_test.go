package wordcloud_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
	"github.com/araihu/goshtoso-charts/components/interactive"
	interactivewordcloud "github.com/araihu/goshtoso-charts/components/interactive/wordcloud"
)

var (
	chartIDPattern = regexp.MustCompile(`id="([A-Za-z]{12})"`)
	scriptPattern  = regexp.MustCompile(`(?s)<script[^>]*>.*?</script>`)
)

func TestWordCloudSpecificTypesHaveCanonicalChildIdentity(t *testing.T) {
	t.Parallel()
	const wantPackage = "github.com/araihu/goshtoso-charts/components/interactive/wordcloud"
	canonicalTypes := []reflect.Type{
		reflect.TypeOf(interactivewordcloud.Shape("")),
		reflect.TypeOf(interactivewordcloud.HorizontalPosition("")),
		reflect.TypeOf(interactivewordcloud.VerticalPosition("")),
		reflect.TypeOf(interactivewordcloud.Word{}),
		reflect.TypeOf(interactivewordcloud.SizeRange{}),
		reflect.TypeOf(interactivewordcloud.Rotation{}),
		reflect.TypeOf(interactivewordcloud.Layout{}),
		reflect.TypeOf(interactivewordcloud.SeriesOptions{}),
		reflect.TypeOf(interactivewordcloud.Series{}),
		reflect.TypeOf(interactivewordcloud.Config{}),
	}
	for _, typ := range canonicalTypes {
		if got := typ.PkgPath(); got != wantPackage {
			t.Errorf("%s PkgPath() = %q, want %q", typ, got, wantPackage)
		}
	}
	compatibilityTypes := []reflect.Type{
		reflect.TypeOf(interactive.WordCloudShape("")),
		reflect.TypeOf(interactive.WordCloudHorizontalPosition("")),
		reflect.TypeOf(interactive.WordCloudVerticalPosition("")),
		reflect.TypeOf(interactive.Word{}),
		reflect.TypeOf(interactive.WordCloudSizeRange{}),
		reflect.TypeOf(interactive.WordCloudRotation{}),
		reflect.TypeOf(interactive.WordCloudLayout{}),
		reflect.TypeOf(interactive.WordCloudSeriesOptions{}),
		reflect.TypeOf(interactive.WordCloudSeries{}),
		reflect.TypeOf(interactive.WordCloudConfig{}),
	}
	for index, typ := range compatibilityTypes {
		if typ != canonicalTypes[index] {
			t.Errorf("compatibility type %s is not identical to canonical %s", typ, canonicalTypes[index])
		}
	}
	if interactive.WordCloudShapeCircle != interactivewordcloud.ShapeCircle ||
		interactive.WordCloudShapeCardioid != interactivewordcloud.ShapeCardioid ||
		interactive.WordCloudShapeDiamond != interactivewordcloud.ShapeDiamond ||
		interactive.WordCloudShapeSquare != interactivewordcloud.ShapeSquare ||
		interactive.WordCloudShapeTriangleForward != interactivewordcloud.ShapeTriangleForward ||
		interactive.WordCloudShapeTriangle != interactivewordcloud.ShapeTriangle ||
		interactive.WordCloudShapePentagon != interactivewordcloud.ShapePentagon ||
		interactive.WordCloudShapeStar != interactivewordcloud.ShapeStar ||
		interactive.WordCloudHorizontalDefault != interactivewordcloud.HorizontalDefault ||
		interactive.WordCloudHorizontalLeft != interactivewordcloud.HorizontalLeft ||
		interactive.WordCloudHorizontalCenter != interactivewordcloud.HorizontalCenter ||
		interactive.WordCloudHorizontalRight != interactivewordcloud.HorizontalRight ||
		interactive.WordCloudVerticalDefault != interactivewordcloud.VerticalDefault ||
		interactive.WordCloudVerticalTop != interactivewordcloud.VerticalTop ||
		interactive.WordCloudVerticalCenter != interactivewordcloud.VerticalCenter ||
		interactive.WordCloudVerticalBottom != interactivewordcloud.VerticalBottom {
		t.Fatal("compatibility constants differ from canonical constants")
	}
	configType := reflect.TypeOf(interactivewordcloud.Config{})
	options, _ := configType.FieldByName("Options")
	if options.Type != reflect.TypeOf(chart.ChartOptions{}) {
		t.Fatalf("Config.Options = %v, want chart.ChartOptions", options.Type)
	}
}

func TestWordCloudFacadePreservesRenderingRuntimeValidationAndBaseHashes(t *testing.T) {
	t.Parallel()
	words := make([]interactivewordcloud.Word, 105)
	for index := range words {
		words[index] = interactivewordcloud.Word{Name: fmt.Sprintf("word-%03d", index), Value: float64(index), Class: "bulk"}
	}
	variants := []struct {
		name         string
		cfg          interactivewordcloud.Config
		fullDigest   string
		scriptDigest string
		shellDigest  string
	}{
		{
			name:       "default",
			cfg:        interactivewordcloud.Config{Label: "basic WordCloud example", Series: interactivewordcloud.Series{Name: "wordcloud", Words: []interactivewordcloud.Word{{Name: "Sam S Club", Value: 10000}, {Name: "Macys", Value: 6181}}}},
			fullDigest: "0d0d585ad682813887a5dcb422b982c324abdd71bd2375f500cbf1a27deb214c", scriptDigest: "d834ce3759bc960f6ea5212e8d55d1395cdbebf74a0fddadc595060170f729f7", shellDigest: "6c8cf4116f29169b0d1b8ec0888b47009741453851e99c0351cccb50e8c4de67",
		},
		{
			name: "configured",
			cfg: interactivewordcloud.Config{
				Label: "basic WordCloud example", Caption: "Twenty weighted search terms.",
				Series: interactivewordcloud.Series{Name: "wordcloud", Words: []interactivewordcloud.Word{{Name: "Sam S Club", Value: 10000, Class: "retail-anchor"}, {Name: "Macys", Value: 6181, Class: "retail", Color: "#123456"}}, Options: interactivewordcloud.SeriesOptions{
					Shape: interactivewordcloud.ShapeStar, SizeRange: &interactivewordcloud.SizeRange{Min: 14, Max: 80}, Rotation: &interactivewordcloud.Rotation{MinDegrees: -45, MaxDegrees: 45, StepDegrees: 15}, GridSize: chart.Int(9), DrawOutOfBound: chart.Bool(true), LayoutAnimation: chart.Bool(false),
					Layout: interactivewordcloud.Layout{Horizontal: interactivewordcloud.HorizontalCenter, Vertical: interactivewordcloud.VerticalCenter, WidthPercent: chart.Float(75), HeightPercent: chart.Float(80)},
				}},
				Options: chart.ChartOptions{Title: &chart.TitleOptions{Text: "basic WordCloud example"}, Tooltip: &chart.TooltipOptions{Show: chart.Bool(true), Trigger: "item"}, Animation: chart.Bool(false), Controls: chartcontrol.Options{Fullscreen: true}, Export: &chartcontrol.ExportOptions{Filename: "word-cloud"}},
				Style:   charttheme.Style{Palette: charttheme.PaletteAraiHu, Colors: []string{"#654321"}, Class: "caller-class"}, RootAttrs: templ.Attributes{"id": "search-terms", "data-purpose": "weighted-terms"},
			},
			fullDigest: "482fe131c1dcf1ed2dcf7c5bdc0318f97c92231a5461656aea9f684fa81a0da2", scriptDigest: "1eac8da7ab74d00c5502c9476596e852c52bd2f58f82a8dc95366791c102d28e", shellDigest: "878d6f564c47d8dc6f1dd14dc32e22aba62f92423b793555bf1572f0ed0241bd",
		},
		{
			name: "bounded-wrapper-omitted",
			cfg: interactivewordcloud.Config{
				Label: "Large cloud", Caption: "Bounded exact values.", Series: interactivewordcloud.Series{Name: "bulk", Words: words, Options: interactivewordcloud.SeriesOptions{Shape: interactivewordcloud.ShapeDiamond, Layout: interactivewordcloud.Layout{Horizontal: interactivewordcloud.HorizontalLeft, Vertical: interactivewordcloud.VerticalBottom, WidthPercent: chart.Float(100), HeightPercent: chart.Float(50)}}},
				Width: "720px", Height: "420px", Options: chart.ChartOptions{Controls: chartcontrol.Options{Mode: chartcontrol.WrapperModeOmitted}}, Style: charttheme.Style{Palette: charttheme.PalettePastel, Class: "caller-wordcloud"}, RootAttrs: templ.Attributes{"data-owner": "consumer"},
			},
			fullDigest: "11e81fd325b32a719b918b8614c94b67bd64d016fe4609dd55846f4033b55d54", scriptDigest: "4a8b1400a7808e5d1c18cf898affc53c89d4f5abc452da57b0b465a028855d65", shellDigest: "a464292b8479769a7ab9d44f4eaecd82d7eeb2a6dda88095490c264d2d7801e1",
		},
	}
	for _, variant := range variants {
		variant := variant
		t.Run(variant.name, func(t *testing.T) {
			t.Parallel()
			canonical := normalizedRender(t, interactivewordcloud.WordCloud(variant.cfg))
			legacy := normalizedRender(t, interactive.WordCloud(variant.cfg))
			if canonical != legacy {
				t.Fatal("canonical WordCloud render differs from compatibility facade")
			}
			assertDigest(t, "full", canonical, variant.fullDigest)
			assertDigest(t, "scripts", strings.Join(scriptPattern.FindAllString(canonical, -1), "\n"), variant.scriptDigest)
			assertDigest(t, "shell", scriptPattern.ReplaceAllString(canonical, "<script></script>"), variant.shellDigest)
		})
	}

	invalid := interactivewordcloud.Config{Label: "Cloud", Series: interactivewordcloud.Series{Name: "words", Words: []interactivewordcloud.Word{{Name: "word", Value: 1}}, Options: interactivewordcloud.SeriesOptions{Shape: "oval"}}}
	canonicalError := renderError(interactivewordcloud.WordCloud(invalid))
	legacyError := renderError(interactive.WordCloud(invalid))
	if canonicalError != `word cloud chart shape "oval" is not supported` || canonicalError != legacyError {
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
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join("..", "wordcloud.go"), nil, 0)
	if err != nil {
		t.Fatalf("parse parent facade: %v", err)
	}
	if len(file.Imports) != 1 || strings.Trim(file.Imports[0].Path.Value, `"`) != "github.com/araihu/goshtoso-charts/components/interactive/wordcloud" {
		t.Fatalf("parent imports = %v, want only canonical Word Cloud package", file.Imports)
	}
	functions := 0
	for _, declaration := range file.Decls {
		switch declaration := declaration.(type) {
		case *ast.FuncDecl:
			functions++
			if declaration.Name.Name != "WordCloud" || declaration.Body == nil || len(declaration.Body.List) != 1 {
				t.Errorf("parent function %s is not the single WordCloud forwarder", declaration.Name.Name)
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
		t.Errorf("parent functions = %d, want only WordCloud", functions)
	}
}

func TestWordCloudTemplateRuntimeHooksAndUpstreamProvenance(t *testing.T) {
	t.Parallel()
	for _, oldPath := range []string{filepath.Join("..", "wordcloud.templ"), filepath.Join("..", "wordcloud_templ.go")} {
		if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
			t.Errorf("compatibility parent unexpectedly retains %s", oldPath)
		}
	}
	checks := []struct {
		path       string
		normalize  *strings.Replacer
		wantDigest string
	}{
		{"wordcloud.templ", strings.NewReplacer("package wordcloud", "package interactive", "values valueRows", "values wordCloudValueRows"), "0d69b119b2cd5c836ee6d8acdb4cb43585539afa7a77e318c2210f6dece8ad27"},
		{"wordcloud_templ.go", strings.NewReplacer("package wordcloud", "package interactive", "components/interactive/wordcloud/wordcloud.templ", "components/interactive/wordcloud.templ", "values valueRows", "values wordCloudValueRows"), "0fcd4e022384257d01448cf4e30ffc8cdefffe05ffe1f598dd568f17349898b7"},
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
		filepath.Join("..", "..", "..", "assets", "js", "runtime", "word-cloud", "2.1.0", "runtime.min.js"): "7b6f0d55971d9de5913120c7ce6342f3551efd00b4a1df8a50f08385bb25f155",
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
	for _, want := range []string{"examples/wordcloud.go", "bda428480a82d6d77ebb9fa939cf8d52528453dd", "echarts-wordcloud", "v2.1.0"} {
		if !bytes.Contains(attributions, []byte(want)) {
			t.Errorf("central upstream/runtime provenance missing %q", want)
		}
	}
	coverage, err := os.ReadFile(filepath.Join("..", "..", "..", "docs", "upstream-example-coverage.md"))
	if err != nil {
		t.Fatalf("read central upstream coverage: %v", err)
	}
	for _, want := range []string{"Interactive Treemap and WordCloud ownership", "examples/wordcloud.go", "bda428480a82d6d77ebb9fa939cf8d52528453dd", "components/interactive/wordcloud.WordCloud"} {
		if !bytes.Contains(coverage, []byte(want)) {
			t.Errorf("central WordCloud coverage missing %q", want)
		}
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

func assertDigest(t *testing.T, name, value, want string) {
	t.Helper()
	digest := sha256.Sum256([]byte(value))
	if got := hex.EncodeToString(digest[:]); got != want {
		t.Errorf("%s SHA-256 = %s, want base %s", name, got, want)
	}
}

var (
	_ func(interactivewordcloud.Config) chart.Instance = interactivewordcloud.WordCloud
	_ func(interactive.WordCloudConfig) chart.Instance = interactive.WordCloud
	_ func(interactivewordcloud.Config) chart.Instance = interactive.WordCloud
	_ func(interactive.WordCloudConfig) chart.Instance = interactivewordcloud.WordCloud
)
