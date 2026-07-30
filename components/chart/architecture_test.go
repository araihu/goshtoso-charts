package chart_test

import (
	"sort"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

const (
	chartPackage       = modulePath + "/components/chart"
	internalPackage    = modulePath + "/components/internal/interactive"
	parentFacade       = modulePath + "/components/interactive"
	childPackagePrefix = parentFacade + "/"
	barPackage         = childPackagePrefix + "bar"
	linePackage        = childPackagePrefix + "line"
	scatterPackage     = childPackagePrefix + "scatter"
	candlestickPackage = childPackagePrefix + "candlestick"
	heatmapPackage     = childPackagePrefix + "heatmap"
	piePackage         = childPackagePrefix + "pie"
	radarPackage       = childPackagePrefix + "radar"
)

func TestChartFoundationPackageDAG(t *testing.T) {
	t.Parallel()

	loaded, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedImports,
	}, chartPackage, internalPackage, parentFacade, barPackage, linePackage, scatterPackage, candlestickPackage, heatmapPackage, piePackage, radarPackage)
	if err != nil {
		t.Fatalf("load chart foundation packages: %v", err)
	}
	if len(loaded) != 10 {
		t.Fatalf("loaded packages = %d, want 10", len(loaded))
	}

	byPath := make(map[string]*packages.Package, len(loaded))
	for _, pkg := range loaded {
		if len(pkg.Errors) != 0 {
			t.Fatalf("load %s: %v", pkg.PkgPath, pkg.Errors)
		}
		byPath[pkg.PkgPath] = pkg
	}
	for _, path := range []string{chartPackage, internalPackage, parentFacade, barPackage, linePackage, scatterPackage, candlestickPackage, heatmapPackage, piePackage, radarPackage} {
		if byPath[path] == nil {
			t.Fatalf("foundation package %s was not loaded", path)
		}
	}

	assertNoImports(t, byPath[internalPackage], func(path string) bool {
		return path == parentFacade || strings.HasPrefix(path, childPackagePrefix)
	}, "parent or child interactive facade")
	assertNoImports(t, byPath[chartPackage], func(path string) bool {
		return path == internalPackage || path == parentFacade || strings.HasPrefix(path, childPackagePrefix) || isRenderingEnginePackage(path)
	}, "internal implementation, interactive facade, or rendering engine")
	assertNoImports(t, byPath[parentFacade], func(path string) bool {
		return strings.HasPrefix(path, childPackagePrefix) && path != barPackage && path != linePackage && path != scatterPackage && path != candlestickPackage && path != heatmapPackage && path != piePackage && path != radarPackage
	}, "child interactive package")
	for _, childPackage := range []string{barPackage, linePackage, scatterPackage, candlestickPackage, heatmapPackage, piePackage, radarPackage} {
		if byPath[parentFacade].Imports[childPackage] == nil {
			t.Errorf("%s does not import migrated canonical package %s", parentFacade, childPackage)
		}
		assertNoImports(t, byPath[childPackage], func(path string) bool {
			return path == parentFacade
		}, "compatibility parent")
	}
}

func assertNoImports(t *testing.T, pkg *packages.Package, forbidden func(string) bool, description string) {
	t.Helper()
	var leaks []string
	for path := range pkg.Imports {
		if forbidden(path) {
			leaks = append(leaks, path)
		}
	}
	sort.Strings(leaks)
	if len(leaks) != 0 {
		t.Errorf("%s imports forbidden %s: %s", pkg.PkgPath, description, strings.Join(leaks, ", "))
	}
}
