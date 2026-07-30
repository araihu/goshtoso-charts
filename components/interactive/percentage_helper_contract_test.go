package interactive

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParentPercentageHelpersAreRemovedWithNoRemainingCallers(t *testing.T) {
	t.Parallel()

	options, err := parser.ParseFile(token.NewFileSet(), "options.go", nil, 0)
	if err != nil {
		t.Fatalf("parse compatibility helpers: %v", err)
	}
	for _, declaration := range options.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if ok && (function.Name.Name == "percentage" || function.Name.Name == "validPercentage") {
			t.Errorf("obsolete parent helper %s remains after Word Cloud migration", function.Name.Name)
		}
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read compatibility parent: %v", err)
	}
	callers := make(map[string]struct{})
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" || strings.HasSuffix(entry.Name(), "_test.go") {
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
				callers[entry.Name()+":"+identifier.Name] = struct{}{}
			}
			return true
		})
	}
	if len(callers) != 0 {
		t.Fatalf("parent percentage helper callers remain: %v", callers)
	}
}

func TestMigratedPackagesUsePrivatePercentageHelpersDirectly(t *testing.T) {
	t.Parallel()

	for name, path := range map[string]string{
		"Sunburst":   "sunburst/sunburst.go",
		"ThemeRiver": "themeriver/themeriver.go",
		"WordCloud":  "wordcloud/wordcloud.go",
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
