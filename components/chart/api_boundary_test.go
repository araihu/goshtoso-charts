package chart_test

import (
	"fmt"
	"go/token"
	"go/types"
	"sort"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

const modulePath = "github.com/araihu/goshtoso-charts"

var publicBoundaryPackages = []string{
	modulePath + "/components/chart",
	modulePath + "/components/interactive",
	modulePath + "/components/interactive/bar",
	modulePath + "/components/interactive/line",
	modulePath + "/components/interactive/scatter",
}

func TestChartFoundationPublicAPIDoesNotExposeImplementationTypes(t *testing.T) {
	t.Parallel()

	loaded := loadTypes(t, publicBoundaryPackages...)
	var leaks []string
	for _, pkg := range loaded {
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			if !token.IsExported(name) {
				continue
			}
			object := scope.Lookup(name)
			walkPublicType(pkg.PkgPath+"."+name, object.Type(), make(map[types.Type]bool), &leaks)
		}
	}

	sort.Strings(leaks)
	if len(leaks) != 0 {
		t.Fatalf("public API exposes internal or rendering-engine types:\n  %s", strings.Join(leaks, "\n  "))
	}
}

func loadTypes(t *testing.T, patterns ...string) []*packages.Package {
	t.Helper()
	loaded, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedImports | packages.NeedDeps | packages.NeedTypes,
	}, patterns...)
	if err != nil {
		t.Fatalf("load public packages: %v", err)
	}
	if len(loaded) != len(patterns) {
		t.Fatalf("loaded packages = %d, want %d", len(loaded), len(patterns))
	}
	for _, pkg := range loaded {
		if len(pkg.Errors) != 0 {
			t.Fatalf("load %s: %v", pkg.PkgPath, pkg.Errors)
		}
	}
	return loaded
}

func walkPublicType(path string, value types.Type, seen map[types.Type]bool, leaks *[]string) {
	if value == nil || seen[value] {
		return
	}
	seen[value] = true

	switch current := value.(type) {
	case *types.Alias:
		recordForbiddenPackage(path, current.Obj().Pkg(), leaks)
		walkPublicType(path, types.Unalias(current), seen, leaks)
	case *types.Named:
		recordForbiddenPackage(path, current.Obj().Pkg(), leaks)
		for index := 0; index < current.TypeArgs().Len(); index++ {
			walkPublicType(path, current.TypeArgs().At(index), seen, leaks)
		}
		walkPublicType(path, current.Underlying(), seen, leaks)
		for index := 0; index < current.NumMethods(); index++ {
			method := current.Method(index)
			if method.Exported() {
				walkPublicType(path+"."+method.Name(), method.Type(), seen, leaks)
			}
		}
	case *types.Pointer:
		walkPublicType(path, current.Elem(), seen, leaks)
	case *types.Array:
		walkPublicType(path, current.Elem(), seen, leaks)
	case *types.Slice:
		walkPublicType(path, current.Elem(), seen, leaks)
	case *types.Map:
		walkPublicType(path+" key", current.Key(), seen, leaks)
		walkPublicType(path+" value", current.Elem(), seen, leaks)
	case *types.Chan:
		walkPublicType(path, current.Elem(), seen, leaks)
	case *types.Struct:
		for index := 0; index < current.NumFields(); index++ {
			field := current.Field(index)
			if field.Exported() {
				walkPublicType(path+"."+field.Name(), field.Type(), seen, leaks)
			}
		}
	case *types.Signature:
		walkTuple(path+" parameter", current.Params(), seen, leaks)
		walkTuple(path+" result", current.Results(), seen, leaks)
		for index := 0; index < current.TypeParams().Len(); index++ {
			walkPublicType(path, current.TypeParams().At(index).Constraint(), seen, leaks)
		}
	case *types.Interface:
		current.Complete()
		for index := 0; index < current.NumMethods(); index++ {
			method := current.Method(index)
			if method.Exported() {
				walkPublicType(path+"."+method.Name(), method.Type(), seen, leaks)
			}
		}
		for index := 0; index < current.NumEmbeddeds(); index++ {
			walkPublicType(path, current.EmbeddedType(index), seen, leaks)
		}
	case *types.TypeParam:
		walkPublicType(path, current.Constraint(), seen, leaks)
	case *types.Union:
		for index := 0; index < current.Len(); index++ {
			walkPublicType(path, current.Term(index).Type(), seen, leaks)
		}
	}
}

func walkTuple(path string, tuple *types.Tuple, seen map[types.Type]bool, leaks *[]string) {
	for index := 0; index < tuple.Len(); index++ {
		walkPublicType(fmt.Sprintf("%s %d", path, index), tuple.At(index).Type(), seen, leaks)
	}
}

func recordForbiddenPackage(path string, pkg *types.Package, leaks *[]string) {
	if pkg == nil {
		return
	}
	packagePath := pkg.Path()
	if isInternalPackage(packagePath) || isRenderingEnginePackage(packagePath) {
		*leaks = append(*leaks, fmt.Sprintf("%s reaches %s", path, packagePath))
	}
}

func isInternalPackage(packagePath string) bool {
	return strings.Contains(packagePath, "/internal/") || strings.HasSuffix(packagePath, "/internal")
}

func isRenderingEnginePackage(packagePath string) bool {
	return strings.HasPrefix(packagePath, "github.com/go-echarts/") ||
		packagePath == "github.com/go-analyze/charts" ||
		strings.HasPrefix(packagePath, "github.com/go-analyze/charts/")
}
