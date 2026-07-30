package bar_test

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestPublicAPIIsRendererNeutral(t *testing.T) {
	t.Parallel()
	loaded, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes,
		Dir:  ".",
	}, ".")
	if err != nil {
		t.Fatalf("load package: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("load package: packages=%d, want 1", len(loaded))
	}
	if len(loaded[0].Errors) != 0 {
		t.Fatalf("load package: errors=%v", loaded[0].Errors)
	}

	for _, name := range loaded[0].Types.Scope().Names() {
		if !token.IsExported(name) {
			continue
		}
		typeName := types.TypeString(loaded[0].Types.Scope().Lookup(name).Type(), packagePath)
		if hasRendererType(loaded[0].Types.Scope().Lookup(name).Type(), make(map[types.Type]bool)) {
			t.Errorf("public API %s exposes renderer type through %s", name, typeName)
		}
		if strings.Contains(strings.ToLower(name), "echarts") {
			t.Errorf("public API name %s exposes renderer implementation", name)
		}
	}

	for _, file := range loaded[0].Syntax {
		ast.Inspect(file, func(node ast.Node) bool {
			importSpec, ok := node.(*ast.ImportSpec)
			if ok {
				path, err := strconv.Unquote(importSpec.Path.Value)
				if err != nil {
					t.Fatalf("unquote import %s: %v", importSpec.Path.Value, err)
				}
				if path == "github.com/araihu/goshtoso-charts/components/interactive" {
					t.Errorf("canonical Bar package imports parent compatibility facade")
				}
			}
			declaration, ok := node.(*ast.GenDecl)
			if ok && declaration.Doc != nil && strings.Contains(strings.ToLower(declaration.Doc.Text()), "echarts") {
				t.Errorf("public package documentation exposes renderer implementation: %q", declaration.Doc.Text())
			}
			function, ok := node.(*ast.FuncDecl)
			if ok && function.Name.IsExported() && function.Doc != nil && strings.Contains(strings.ToLower(function.Doc.Text()), "echarts") {
				t.Errorf("public API documentation for %s exposes renderer implementation", function.Name.Name)
			}
			return true
		})
	}
}

func hasRendererType(value types.Type, seen map[types.Type]bool) bool {
	if value == nil || seen[value] {
		return false
	}
	seen[value] = true

	switch current := value.(type) {
	case *types.Alias:
		return hasRendererType(types.Unalias(current), seen)
	case *types.Named:
		object := current.Obj()
		if object.Pkg() != nil {
			path := object.Pkg().Path()
			if strings.HasPrefix(path, "github.com/go-echarts/") {
				return true
			}
			if !strings.HasPrefix(path, "github.com/araihu/goshtoso-charts/") {
				return false
			}
		}
		if hasRendererType(current.Underlying(), seen) {
			return true
		}
		for index := 0; index < current.NumMethods(); index++ {
			method := current.Method(index)
			if method.Exported() && hasRendererType(method.Type(), seen) {
				return true
			}
		}
	case *types.Pointer:
		return hasRendererType(current.Elem(), seen)
	case *types.Array:
		return hasRendererType(current.Elem(), seen)
	case *types.Slice:
		return hasRendererType(current.Elem(), seen)
	case *types.Map:
		return hasRendererType(current.Key(), seen) || hasRendererType(current.Elem(), seen)
	case *types.Chan:
		return hasRendererType(current.Elem(), seen)
	case *types.Struct:
		for index := 0; index < current.NumFields(); index++ {
			field := current.Field(index)
			if field.Exported() && hasRendererType(field.Type(), seen) {
				return true
			}
		}
	case *types.Signature:
		return tupleHasRendererType(current.Params(), seen) || tupleHasRendererType(current.Results(), seen)
	case *types.Interface:
		for index := 0; index < current.NumExplicitMethods(); index++ {
			method := current.ExplicitMethod(index)
			if method.Exported() && hasRendererType(method.Type(), seen) {
				return true
			}
		}
	}
	return false
}

func tupleHasRendererType(tuple *types.Tuple, seen map[types.Type]bool) bool {
	for index := 0; index < tuple.Len(); index++ {
		if hasRendererType(tuple.At(index).Type(), seen) {
			return true
		}
	}
	return false
}

func packagePath(pkg *types.Package) string { return pkg.Path() }
