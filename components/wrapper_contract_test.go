package components_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/bar"
	"github.com/araihu/goshtoso-charts/components/candlestick"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/funnel"
	"github.com/araihu/goshtoso-charts/components/heatmap"
	"github.com/araihu/goshtoso-charts/components/interactive"
	interactiveboxplot "github.com/araihu/goshtoso-charts/components/interactive/boxplot"
	interactivegauge "github.com/araihu/goshtoso-charts/components/interactive/gauge"
	interactiveheatmap "github.com/araihu/goshtoso-charts/components/interactive/heatmap"
	interactiveradar "github.com/araihu/goshtoso-charts/components/interactive/radar"
	"github.com/araihu/goshtoso-charts/components/line"
	"github.com/araihu/goshtoso-charts/components/pie"
	"github.com/araihu/goshtoso-charts/components/radar"
	"github.com/araihu/goshtoso-charts/components/scatter"
	charttable "github.com/araihu/goshtoso-charts/components/table"
	"github.com/araihu/goshtoso-charts/components/violin"
	"golang.org/x/tools/go/packages"
)

type wrapperConfigContract struct {
	kind        components.Kind
	config      reflect.Type
	interactive bool
}

var wrapperConfigContracts = []wrapperConfigContract{
	{components.KindLineChart, reflect.TypeOf(line.Config{}), false},
	{components.KindBarChart, reflect.TypeOf(bar.Config{}), false},
	{components.KindPieChart, reflect.TypeOf(pie.Config{}), false},
	{components.KindScatterChart, reflect.TypeOf(scatter.Config{}), false},
	{components.KindRadarChart, reflect.TypeOf(radar.Config{}), false},
	{components.KindCandlestickChart, reflect.TypeOf(candlestick.Config{}), false},
	{components.KindHeatMapChart, reflect.TypeOf(heatmap.Config{}), false},
	{components.KindFunnelChart, reflect.TypeOf(funnel.Config{}), false},
	{components.KindTable, reflect.TypeOf(charttable.Config{}), false},
	{components.KindViolinChart, reflect.TypeOf(violin.Config{}), false},
	{components.KindInteractiveBar, reflect.TypeOf(interactive.BarConfig{}), true},
	{components.KindInteractiveLine, reflect.TypeOf(interactive.LineConfig{}), true},
	{components.KindInteractiveScatter, reflect.TypeOf(interactive.ScatterConfig{}), true},
	{components.KindInteractiveScatter3D, reflect.TypeOf(interactive.Scatter3DConfig{}), true},
	{components.KindInteractiveBar3D, reflect.TypeOf(interactive.Bar3DConfig{}), true},
	{components.KindInteractiveSurface3D, reflect.TypeOf(interactive.Surface3DConfig{}), true},
	{components.KindInteractiveLine3D, reflect.TypeOf(interactive.Line3DConfig{}), true},
	{components.KindInteractivePie, reflect.TypeOf(interactive.PieConfig{}), true},
	{components.KindInteractiveRadar, reflect.TypeOf(interactiveradar.Config{}), true},
	{components.KindInteractiveHeatMap, reflect.TypeOf(interactiveheatmap.Config{}), true},
	{components.KindInteractiveBoxPlot, reflect.TypeOf(interactiveboxplot.Config{}), true},
	{components.KindInteractiveGauge, reflect.TypeOf(interactivegauge.Config{}), true},
	{components.KindInteractiveFunnel, reflect.TypeOf(interactive.FunnelConfig{}), true},
	{components.KindInteractiveGraph, reflect.TypeOf(interactive.GraphConfig{}), true},
	{components.KindInteractiveSankey, reflect.TypeOf(interactive.SankeyConfig{}), true},
	{components.KindInteractiveTree, reflect.TypeOf(interactive.TreeConfig{}), true},
	{components.KindInteractiveSunburst, reflect.TypeOf(interactive.SunburstConfig{}), true},
	{components.KindInteractiveTreemap, reflect.TypeOf(interactive.TreemapConfig{}), true},
	{components.KindInteractiveParallel, reflect.TypeOf(interactive.ParallelConfig{}), true},
	{components.KindInteractiveThemeRiver, reflect.TypeOf(interactive.ThemeRiverConfig{}), true},
	{components.KindInteractiveCandlestick, reflect.TypeOf(interactive.CandlestickConfig{}), true},
	{components.KindInteractiveWordCloud, reflect.TypeOf(interactive.WordCloudConfig{}), true},
	{components.KindInteractiveMap, reflect.TypeOf(interactive.MapConfig{}), true},
	{components.KindInteractiveGeo, reflect.TypeOf(interactive.GeoConfig{}), true},
}

func TestEveryPublicChartConfigSharesOneWrapperContract(t *testing.T) {
	t.Parallel()

	controlsType := reflect.TypeOf(chartcontrol.Options{})
	exportType := reflect.TypeOf((*chartcontrol.ExportOptions)(nil))
	chartOptionsType := reflect.TypeOf(interactive.ChartOptions{})
	if field, ok := chartOptionsType.FieldByName("Controls"); !ok || field.Type != controlsType {
		t.Fatalf("interactive.ChartOptions.Controls = %v/%t, want %v", field.Type, ok, controlsType)
	}
	if field, ok := chartOptionsType.FieldByName("Export"); !ok || field.Type != exportType {
		t.Fatalf("interactive.ChartOptions.Export = %v/%t, want %v", field.Type, ok, exportType)
	}

	covered := make(map[components.Kind]bool, len(wrapperConfigContracts))
	for _, contract := range wrapperConfigContracts {
		contract := contract
		t.Run(string(contract.kind), func(t *testing.T) {
			if covered[contract.kind] {
				t.Fatalf("duplicate wrapper contract for %q", contract.kind)
			}
			covered[contract.kind] = true
			if contract.interactive {
				field, ok := contract.config.FieldByName("Options")
				if !ok || field.Type != chartOptionsType {
					t.Fatalf("%s.Options = %v/%t, want interactive.ChartOptions", contract.config, field.Type, ok)
				}
				return
			}
			controls, ok := contract.config.FieldByName("Controls")
			if !ok || controls.Type != controlsType {
				t.Fatalf("%s.Controls = %v/%t, want chartcontrol.Options", contract.config, controls.Type, ok)
			}
			export, ok := contract.config.FieldByName("Export")
			if !ok || export.Type != exportType {
				t.Fatalf("%s.Export = %v/%t, want *chartcontrol.ExportOptions", contract.config, export.Type, ok)
			}
		})
	}

	allKinds := components.AllKinds()
	if len(covered) != len(allKinds) {
		t.Fatalf("wrapper contract covers %d chart kinds, want all %d", len(covered), len(allKinds))
	}
	for _, kind := range allKinds {
		if !covered[kind] {
			t.Errorf("chart kind %q has no shared wrapper config contract", kind)
		}
	}
}

func TestEveryChartRenderPathPropagatesSharedWrapperFields(t *testing.T) {
	t.Parallel()

	staticFiles := []string{
		"bar/component.go", "candlestick/component.go", "funnel/component.go", "heatmap/component.go",
		"line/component.go", "pie/component.go", "radar/component.go", "scatter/component.go",
		"table/component.go", "violin/component.go",
	}
	for _, filename := range staticFiles {
		assertWrapperLiteral(t, filename, "static-svg")
	}
	assertWrapperLiteral(t, "internal/interactive/component.go", "interactive-raster")
	assertConstructorPropagatesSharedWrapperFields(t, "interactive/bar/bar.go", "Bar")
	assertConstructorPropagatesSharedWrapperFields(t, "interactive/boxplot/boxplot.go", "BoxPlot")
	assertConstructorPropagatesSharedWrapperFields(t, "interactive/candlestick/candlestick.go", "Candlestick")
	assertConstructorPropagatesSharedWrapperFields(t, "interactive/gauge/gauge.go", "Gauge")
	assertConstructorPropagatesSharedWrapperFields(t, "interactive/heatmap/heatmap.go", "HeatMap")
	assertConstructorPropagatesSharedWrapperFields(t, "interactive/line/line.go", "Line")
	assertConstructorPropagatesSharedWrapperFields(t, "interactive/pie/pie.go", "Pie")
	assertConstructorPropagatesSharedWrapperFields(t, "interactive/radar/radar.go", "Radar")
	assertConstructorPropagatesSharedWrapperFields(t, "interactive/scatter/scatter.go", "Scatter")

	loaded, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedSyntax,
		Dir:  "interactive",
	}, ".")
	if err != nil {
		t.Fatalf("load interactive constructors: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("load interactive constructors: got %d packages, want 1", len(loaded))
	}
	if len(loaded[0].Errors) > 0 {
		t.Fatalf("load interactive constructors: %v", loaded[0].Errors)
	}
	var missing []string
	for _, file := range loaded[0].Syntax {
		for _, declaration := range file.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Recv != nil || !function.Name.IsExported() || function.Type.Results == nil || len(function.Type.Results.List) != 1 {
				continue
			}
			result, ok := function.Type.Results.List[0].Type.(*ast.Ident)
			if !ok || result.Name != "Instance" || function.Type.Params == nil || len(function.Type.Params.List) != 1 {
				continue
			}
			parameter, ok := function.Type.Params.List[0].Type.(*ast.Ident)
			if !ok || !strings.HasSuffix(parameter.Name, "Config") {
				continue
			}
			if function.Name.Name == "Bar" || function.Name.Name == "BoxPlot" || function.Name.Name == "Candlestick" || function.Name.Name == "Gauge" || function.Name.Name == "HeatMap" || function.Name.Name == "Line" || function.Name.Name == "Pie" || function.Name.Name == "Radar" || function.Name.Name == "Scatter" {
				continue // Physical ownership is covered in the child implementations above.
			}
			if !containsSelectorPath(function.Body, "cfg", "Options", "Controls") || !containsSelectorPath(function.Body, "cfg", "Options", "Export") {
				missing = append(missing, function.Name.Name)
			}
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		t.Fatalf("interactive constructors do not propagate ChartOptions Controls/Export: %v", missing)
	}
}

func assertConstructorPropagatesSharedWrapperFields(t *testing.T, filename, name string) {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", filename, err)
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Name.Name != name {
			continue
		}
		if !containsSelectorPath(function.Body, "cfg", "Options", "Controls") || !containsSelectorPath(function.Body, "cfg", "Options", "Export") {
			t.Fatalf("%s constructor %s does not propagate ChartOptions Controls/Export", filename, name)
		}
		return
	}
	t.Fatalf("%s constructor %s not found", filename, name)
}

func assertWrapperLiteral(t *testing.T, filename, capability string) {
	t.Helper()
	files, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", filename, err)
	}
	var wrappers int
	ast.Inspect(files, func(node ast.Node) bool {
		literal, ok := node.(*ast.CompositeLit)
		if !ok {
			return true
		}
		selector, ok := literal.Type.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "WrapperConfig" {
			return true
		}
		wrappers++
		fields := make(map[string]ast.Expr)
		for _, element := range literal.Elts {
			pair, ok := element.(*ast.KeyValueExpr)
			key, keyOK := pair.Key.(*ast.Ident)
			if ok && keyOK {
				fields[key.Name] = pair.Value
			}
		}
		if !selectorEndsWith(fields["Controls"], "Controls") || !selectorEndsWith(fields["Export"], "Export") {
			t.Errorf("%s wrapper does not propagate Controls and Export", filename)
		}
		capabilitySelector, ok := fields["Capability"].(*ast.SelectorExpr)
		if !ok || !strings.Contains(strings.ToLower(capabilitySelector.Sel.Name), strings.ReplaceAll(capability, "-", "")) {
			t.Errorf("%s wrapper capability = %T, want %s", filename, fields["Capability"], capability)
		}
		return true
	})
	if wrappers != 1 {
		t.Fatalf("%s has %d chartcontrol.WrapperConfig literals, want 1", filename, wrappers)
	}
}

func selectorEndsWith(expression ast.Expr, name string) bool {
	selector, ok := expression.(*ast.SelectorExpr)
	return ok && selector.Sel.Name == name
}

func containsSelectorPath(node ast.Node, path ...string) bool {
	found := false
	ast.Inspect(node, func(candidate ast.Node) bool {
		selector, ok := candidate.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		parts := selectorPath(selector)
		if reflect.DeepEqual(parts, path) {
			found = true
			return false
		}
		return true
	})
	return found
}

func selectorPath(expression ast.Expr) []string {
	switch value := expression.(type) {
	case *ast.Ident:
		return []string{value.Name}
	case *ast.SelectorExpr:
		return append(selectorPath(value.X), value.Sel.Name)
	default:
		return nil
	}
}
