package interactive

import (
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"testing"

	internalinteractive "github.com/araihu/goshtoso-charts/components/internal/interactive"
)

func TestPercentageHelpersRemainThinPrivateAdapterWrappers(t *testing.T) {
	t.Parallel()

	for _, value := range []float64{0, 12.5, 100} {
		if got, want := percentage(value), internalinteractive.Percentage(value); got != want {
			t.Errorf("percentage(%g) = %q, want private adapter %q", value, got, want)
		}
		if got, want := validPercentage(value), internalinteractive.ValidPercentage(value); got != want {
			t.Errorf("validPercentage(%g) = %t, want private adapter %t", value, got, want)
		}
	}
	for _, value := range []float64{-1, 101, math.NaN(), math.Inf(1)} {
		if validPercentage(value) || internalinteractive.ValidPercentage(value) {
			t.Errorf("invalid percentage %g was accepted", value)
		}
	}

	file, err := parser.ParseFile(token.NewFileSet(), "options.go", nil, 0)
	if err != nil {
		t.Fatalf("parse compatibility helpers: %v", err)
	}
	wantTargets := map[string]string{
		"percentage":      "Percentage",
		"validPercentage": "ValidPercentage",
	}
	seen := make(map[string]bool, len(wantTargets))
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || wantTargets[function.Name.Name] == "" {
			continue
		}
		seen[function.Name.Name] = true
		if function.Body == nil || len(function.Body.List) != 1 {
			t.Errorf("%s is not a single-statement wrapper", function.Name.Name)
			continue
		}
		statement, ok := function.Body.List[0].(*ast.ReturnStmt)
		if !ok || len(statement.Results) != 1 {
			t.Errorf("%s does not directly return private adapter result", function.Name.Name)
			continue
		}
		call, ok := statement.Results[0].(*ast.CallExpr)
		if !ok {
			t.Errorf("%s does not directly call private adapter", function.Name.Name)
			continue
		}
		selector, selectorOK := call.Fun.(*ast.SelectorExpr)
		if !selectorOK {
			t.Errorf("%s does not directly select private adapter helper", function.Name.Name)
			continue
		}
		packageName, packageOK := selector.X.(*ast.Ident)
		if !packageOK || packageName.Name != "internalinteractive" || selector.Sel.Name != wantTargets[function.Name.Name] {
			t.Errorf("%s does not directly forward to internalinteractive.%s", function.Name.Name, wantTargets[function.Name.Name])
		}
	}
	for name := range wantTargets {
		if !seen[name] {
			t.Errorf("compatibility helper %s is missing", name)
		}
	}
}

func TestPercentageHelperDependentsRemainOnParentWrappers(t *testing.T) {
	t.Parallel()

	wants := map[string]map[string]bool{
		"sunburst.go":    {"percentage": true, "validPercentage": true},
		"theme_river.go": {"percentage": true, "validPercentage": true},
		"wordcloud.go":   {"percentage": true},
	}
	for filename, required := range wants {
		file, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", filename, err)
		}
		seen := make(map[string]bool, len(required))
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			identifier, ok := call.Fun.(*ast.Ident)
			if ok && required[identifier.Name] {
				seen[identifier.Name] = true
			}
			return true
		})
		for helper := range required {
			if !seen[helper] {
				t.Errorf("%s no longer uses compatibility helper %s", filename, helper)
			}
		}
	}
}
