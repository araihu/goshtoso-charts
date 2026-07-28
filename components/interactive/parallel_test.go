package interactive

import (
	"bytes"
	"context"
	"math"
	"strings"
	"testing"

	"github.com/a-h/templ"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

func TestParallelRendersTypedDimensionsSeriesLayoutAndExactValues(t *testing.T) {
	t.Parallel()
	instance := Parallel(ParallelConfig{
		Label:   "Air quality profiles",
		Caption: "Daily measurements grouped by city.",
		Dimensions: []ParallelDimension{
			{
				Name: "Date", Range: &ParallelRange{Min: Float(1), Max: Float(31)}, Inverse: true, NameLocation: ParallelNameStart,
				Label: ParallelAxisLabel{Show: Bool(true), Rotate: 20, Margin: 9, FontSize: 11},
				Line:  ParallelAxisLine{Show: Bool(true), Color: "#334155", Width: 2, Type: "dashed"},
			},
			{Name: "Level", Categories: []string{"Good", "Moderate", "Heavily"}},
		},
		Series: []ParallelSeries{{
			Name: "Beijing", Class: "city-north", Color: "#123456",
			Options: ParallelSeriesOptions{
				Line:   ParallelLineOptions{Width: 2, Type: "solid", Opacity: Float(0.8)},
				Smooth: Bool(true), InactiveOpacity: Float(0.12),
			},
			Observations: []ParallelObservation{
				{Name: "Day 1", Values: []ParallelValue{ParallelNumber(1), ParallelCategory("Moderate")}, Class: "moderate"},
				{Name: "Day 2", Values: []ParallelValue{ParallelNumber(2), ParallelCategory("Good")}, Color: "#abcdef"},
			},
		}},
		Layout: ParallelLayout{LeftPercent: Float(15), RightPercent: Float(13), TopPercent: Float(20), BottomPercent: Float(10)},
		Width:  "100%", Height: "500px",
		Options: ChartOptions{
			Title: &TitleOptions{Text: "Multi Series"}, Legend: &LegendOptions{Show: Bool(true)},
			Controls: chartcontrol.Options{Fullscreen: true, Collapsible: true},
			Export:   &chartcontrol.ExportOptions{Filename: "air-quality-profiles"},
		},
		Style:     charttheme.Style{Palette: charttheme.PaletteAraiHu, Class: "max-w-full"},
		RootAttrs: templ.Attributes{"id": "air-quality", "data-chart-purpose": "multivariate"},
	})

	if instance.Kind() != chartcomponents.KindInteractiveParallel {
		t.Fatalf("Kind() = %q", instance.Kind())
	}
	markup := renderParallel(t, instance)
	for _, want := range []string{
		`role="img"`, `aria-label="Air quality profiles"`, `id="air-quality"`, `data-chart-purpose="multivariate"`,
		`style="width:100%;height:500px;"`, `"parallel":{"left":"15%","top":"20%","right":"13%","bottom":"10%"}`,
		`"parallelAxis":[{"dim":0,"name":"Date","min":1,"max":31,"inverse":true,"nameLocation":"start"`,
		`"axisLabel":{"show":true,"rotate":20,"margin":9,"fontSize":11}`,
		`"axisLine":{"show":true,"lineStyle":{"color":"#334155","width":2,"type":"dashed"}}`,
		`{"dim":1,"name":"Level","type":"category","data":["Good","Moderate","Heavily"]}`,
		`"name":"Beijing","type":"parallel","className":"city-north","inactiveOpacity":0.12`,
		`"smooth":true`, `"lineStyle":{"color":"#123456","width":2,"type":"solid","opacity":0.8}`,
		`"name":"Day 1","value":[1,"Moderate"],"className":"moderate"`,
		`"name":"Day 2","value":[2,"Good"],"lineStyle":{"color":"#abcdef"}`,
		`"text":"Multi Series"`, `"show":true`, `data-goshtoso-charts-theme-runtime`,
		`data-goshtoso-chart-expand`, `data-goshtoso-chart-export="png"`, `data-goshtoso-chart-control="fullscreen"`, `data-goshtoso-chart-control="collapse"`,
		`>Exact observations and values</summary>`, `scope="col">Series</th>`, `scope="col">Observation</th>`,
		`scope="col">Date</th>`, `scope="col">Level</th>`, `scope="col">Semantic class</th>`,
		`>Beijing</th>`, `>Day 1</td>`, `>Moderate</td>`, `>city-north moderate</td>`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
}

func TestParallelDefaultsToExpandAndPNGWithoutOptionalControls(t *testing.T) {
	t.Parallel()
	markup := renderParallel(t, Parallel(validParallelConfig()))
	for _, want := range []string{`data-goshtoso-chart-expand`, `data-goshtoso-chart-export="png"`} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	for _, unwanted := range []string{`data-goshtoso-chart-control="fullscreen"`, `data-goshtoso-chart-control="collapse"`} {
		if strings.Contains(markup, unwanted) {
			t.Errorf("rendered markup unexpectedly contains %q", unwanted)
		}
	}
}

func TestParallelColorOverrideDoesNotImplyTransparentLine(t *testing.T) {
	t.Parallel()
	cfg := validParallelConfig()
	cfg.Series[0].Color = "#123456"
	markup := renderParallel(t, Parallel(cfg))
	if !strings.Contains(markup, `"lineStyle":{"color":"#123456"}`) {
		t.Fatal("explicit series color was not serialized")
	}
	if strings.Contains(markup, `"lineStyle":{"color":"#123456","opacity":0}`) {
		t.Fatal("omitted opacity made explicit series color transparent")
	}
}

func TestParallelRejectsInvalidContracts(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		mutate    func(*ParallelConfig)
		wantError string
	}{
		"label":               {func(cfg *ParallelConfig) { cfg.Label = " " }, "parallel chart label is required"},
		"dimensions":          {func(cfg *ParallelConfig) { cfg.Dimensions = cfg.Dimensions[:1] }, "parallel chart requires at least two dimensions"},
		"dimension name":      {func(cfg *ParallelConfig) { cfg.Dimensions[0].Name = "" }, "dimension 0 name is required"},
		"duplicate dimension": {func(cfg *ParallelConfig) { cfg.Dimensions[1].Name = cfg.Dimensions[0].Name }, `dimension name "Value" is duplicated`},
		"range":               {func(cfg *ParallelConfig) { cfg.Dimensions[0].Range = &ParallelRange{Min: Float(2), Max: Float(1)} }, "numeric range min must be less than max"},
		"category range":      {func(cfg *ParallelConfig) { cfg.Dimensions[1].Range = &ParallelRange{Min: Float(0), Max: Float(1)} }, "categorical dimensions cannot define a numeric range"},
		"duplicate category":  {func(cfg *ParallelConfig) { cfg.Dimensions[1].Categories = []string{"A", "A"} }, `category "A" is duplicated`},
		"bad scale":           {func(cfg *ParallelConfig) { cfg.Dimensions[0].Scale = "power" }, `scale "power" is not supported`},
		"log category":        {func(cfg *ParallelConfig) { cfg.Dimensions[1].Scale = ParallelScaleLog }, "categorical dimensions cannot use a log scale"},
		"log number": {func(cfg *ParallelConfig) {
			cfg.Dimensions[0].Scale = ParallelScaleLog
			cfg.Dimensions[0].Range = &ParallelRange{Min: Float(0.1), Max: Float(10)}
			cfg.Series[0].Observations[0].Values[0] = ParallelNumber(0)
		}, "log value must be positive"},
		"axis rotation":    {func(cfg *ParallelConfig) { cfg.Dimensions[0].Label.Rotate = 91 }, "label rotation must be between -90 and 90"},
		"axis width":       {func(cfg *ParallelConfig) { cfg.Dimensions[0].Line.Width = -1 }, "axis line width must be finite and nonnegative"},
		"series":           {func(cfg *ParallelConfig) { cfg.Series = nil }, "parallel chart series is required"},
		"series name":      {func(cfg *ParallelConfig) { cfg.Series[0].Name = "" }, "series 0 name is required"},
		"observations":     {func(cfg *ParallelConfig) { cfg.Series[0].Observations = nil }, `series "Series" observations are required`},
		"observation name": {func(cfg *ParallelConfig) { cfg.Series[0].Observations[0].Name = "" }, `observation 0 name is required`},
		"alignment": {func(cfg *ParallelConfig) {
			cfg.Series[0].Observations[0].Values = cfg.Series[0].Observations[0].Values[:1]
		}, "has 1 values for 2 dimensions"},
		"numeric type":     {func(cfg *ParallelConfig) { cfg.Series[0].Observations[0].Values[0] = ParallelCategory("A") }, "value must be numeric"},
		"category type":    {func(cfg *ParallelConfig) { cfg.Series[0].Observations[0].Values[1] = ParallelNumber(1) }, "value must be categorical"},
		"membership":       {func(cfg *ParallelConfig) { cfg.Series[0].Observations[0].Values[1] = ParallelCategory("C") }, `category "C" is not declared`},
		"finite":           {func(cfg *ParallelConfig) { cfg.Series[0].Observations[0].Values[0] = ParallelNumber(math.NaN()) }, "numeric value must be finite"},
		"out of range":     {func(cfg *ParallelConfig) { cfg.Series[0].Observations[0].Values[0] = ParallelNumber(11) }, "exceeds maximum 10"},
		"line opacity":     {func(cfg *ParallelConfig) { cfg.Series[0].Options.Line.Opacity = Float(1.1) }, "line opacity must be between 0 and 1"},
		"inactive opacity": {func(cfg *ParallelConfig) { cfg.Series[0].Options.InactiveOpacity = Float(-0.1) }, "inactive opacity must be between 0 and 1"},
		"layout":           {func(cfg *ParallelConfig) { cfg.Layout.LeftPercent = Float(101) }, "left inset must be between 0 and 100 percent"},
		"layout total":     {func(cfg *ParallelConfig) { cfg.Layout.LeftPercent, cfg.Layout.RightPercent = Float(50), Float(50) }, "horizontal insets must total less than 100 percent"},
		"root attrs":       {func(cfg *ParallelConfig) { cfg.RootAttrs = templ.Attributes{"role": "group"} }, `root attribute "role" is reserved`},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := validParallelConfig()
			test.mutate(&cfg)
			var output bytes.Buffer
			err := Parallel(cfg).Render(context.Background(), &output)
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("Render() error = %v, want containing %q", err, test.wantError)
			}
		})
	}
}

func TestParallelExactValuesAreBounded(t *testing.T) {
	t.Parallel()
	cfg := validParallelConfig()
	cfg.Series[0].Observations = make([]ParallelObservation, maxParallelDetailRows+2)
	for index := range cfg.Series[0].Observations {
		cfg.Series[0].Observations[index] = ParallelObservation{
			Name:   "row-" + string(rune('A'+index)),
			Values: []ParallelValue{ParallelNumber(float64(index % 10)), ParallelCategory("A")},
		}
	}
	markup := renderParallel(t, Parallel(cfg))
	if !strings.Contains(markup, "2 additional observations omitted") {
		t.Fatal("bounded exact-value table did not report omitted rows")
	}
}

func validParallelConfig() ParallelConfig {
	return ParallelConfig{
		Label: "Parallel",
		Dimensions: []ParallelDimension{
			{Name: "Value", Range: &ParallelRange{Min: Float(0), Max: Float(10)}},
			{Name: "Class", Categories: []string{"A", "B"}},
		},
		Series: []ParallelSeries{{
			Name:         "Series",
			Observations: []ParallelObservation{{Name: "Row", Values: []ParallelValue{ParallelNumber(5), ParallelCategory("A")}}},
		}},
	}
}

func renderParallel(t *testing.T, instance Instance) string {
	t.Helper()
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	return output.String()
}
