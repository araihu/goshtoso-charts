package interactive

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
)

func TestParentPercentageHelperRemainsThinPrivateAdapterWrapper(t *testing.T) {
	t.Parallel()

	for _, value := range []float64{0, 12.5, 100} {
		if got, want := percentage(value), internalinteractive.Percentage(value); got != want {
			t.Errorf("percentage(%g) = %q, want private adapter %q", value, got, want)
		}
	}

	file, err := parser.ParseFile(token.NewFileSet(), "options.go", nil, 0)
	if err != nil {
		t.Fatalf("parse compatibility helpers: %v", err)
	}
	seen := false
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if function.Name.Name == "validPercentage" {
			t.Error("unused parent validPercentage helper remains after Sunburst migration")
		}
		if function.Name.Name != "percentage" {
			continue
		}
		seen = true
		if function.Body == nil || len(function.Body.List) != 1 {
			t.Error("percentage is not a single-statement wrapper")
			continue
		}
		statement, ok := function.Body.List[0].(*ast.ReturnStmt)
		if !ok || len(statement.Results) != 1 {
			t.Error("percentage does not directly return private adapter result")
			continue
		}
		call, ok := statement.Results[0].(*ast.CallExpr)
		if !ok {
			t.Error("percentage does not directly call private adapter")
			continue
		}
		selector, selectorOK := call.Fun.(*ast.SelectorExpr)
		if !selectorOK {
			t.Error("percentage does not directly select private adapter helper")
			continue
		}
		packageName, packageOK := selector.X.(*ast.Ident)
		if !packageOK || packageName.Name != "internalinteractive" || selector.Sel.Name != "Percentage" {
			t.Error("percentage does not directly forward to internalinteractive.Percentage")
		}
	}
	if !seen {
		t.Error("compatibility percentage helper is missing")
	}
}

func TestParentPercentageHelperOnlyServesUnmigratedWordCloud(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read compatibility parent: %v", err)
	}
	callers := make(map[string]int)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" || strings.HasSuffix(entry.Name(), "_test.go") || entry.Name() == "options.go" {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), entry.Name(), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", entry.Name(), err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			identifier, ok := call.Fun.(*ast.Ident)
			if ok && (identifier.Name == "percentage" || identifier.Name == "validPercentage") {
				callers[entry.Name()+":"+identifier.Name]++
			}
			return true
		})
	}
	want := map[string]int{"wordcloud.go:percentage": 2}
	if !reflectStringIntMapEqual(callers, want) {
		t.Fatalf("parent percentage helper callers = %v, want only unmigrated WordCloud %v", callers, want)
	}
}

func TestMigratedPackagesUsePrivatePercentageHelpersDirectly(t *testing.T) {
	t.Parallel()

	for name, path := range map[string]string{
		"Sunburst":   "sunburst/sunburst.go",
		"ThemeRiver": "themeriver/themeriver.go",
	} {
		name, path := name, path
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				t.Fatalf("parse canonical %s: %v", name, err)
			}
			wants := map[string]bool{"Percentage": false, "ValidPercentage": false}
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				if identifier, ok := call.Fun.(*ast.Ident); ok && (identifier.Name == "percentage" || identifier.Name == "validPercentage") {
					t.Errorf("canonical %s calls parent compatibility helper %s", name, identifier.Name)
				}
				selector, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				if _, wanted := wants[selector.Sel.Name]; !wanted {
					return true
				}
				packageName, ok := selector.X.(*ast.Ident)
				if ok && packageName.Name == "internalinteractive" {
					wants[selector.Sel.Name] = true
				}
				return true
			})
			for helper, seen := range wants {
				if !seen {
					t.Errorf("canonical %s does not call internalinteractive.%s", name, helper)
				}
			}
		})
	}
}

func reflectStringIntMapEqual(left, right map[string]int) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}
