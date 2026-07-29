package pie

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/a-h/templ"
	chartassets "github.com/araihu/goshtoso-charts/assets"
	chartcomponents "github.com/araihu/goshtoso-charts/components"
	"github.com/araihu/goshtoso-charts/components/chartcontrol"
	"github.com/araihu/goshtoso-charts/components/charttheme"
)

const upstreamDoughnutExample = "examples/1-Painter/doughnut_chart-1-basic/main.go@1fe31b06b8a82e00df877ff4417a75858547c1c2"

func TestPieRendersSSRAccessibleSVG(t *testing.T) {
	t.Parallel()
	instance := Pie(Config{
		Label: "Deployments by status", Caption: "This week.",
		Slices: []Slice{{Name: "Successful", Value: 14}, {Name: "Failed", Value: 2}},
	})
	if instance.Kind() != chartcomponents.KindPieChart {
		t.Fatalf("Kind() = %q, want %q", instance.Kind(), chartcomponents.KindPieChart)
	}
	var output bytes.Buffer
	if err := instance.Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{"<figure class=\"goshtoso-charts-pie goshtoso-charts-palette goshtoso-charts-palette-auto\" role=\"img\" aria-label=\"Deployments by status\"", "<svg", "This week.", "var(--color-chart-series-1)", "var(--color-chart-surface)", "var(--font-paragraph), sans-serif"} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	if !strings.Contains(markup, `src="`+chartassets.ControlRuntimeURL+`"`) {
		t.Errorf("SSR chart missing shared controls runtime")
	}
	if got := strings.Count(markup, "<script"); got != 1 {
		t.Errorf("SSR chart script count = %d, want shared controls runtime only", got)
	}
}

func TestPiePreservesSharedFourModeWrapperContract(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name string
		mode chartcontrol.WrapperMode
		want string
	}{
		{name: "enabled", mode: chartcontrol.WrapperModeEnabled, want: `data-goshtoso-chart-wrapper-mode="enabled"`},
		{name: "disabled", mode: chartcontrol.WrapperModeDisabled, want: `data-goshtoso-chart-wrapper-mode="disabled"`},
		{name: "hidden", mode: chartcontrol.WrapperModeHidden, want: `data-goshtoso-chart-wrapper-mode="hidden"`},
		{name: "omitted", mode: chartcontrol.WrapperModeOmitted},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer
			err := Pie(Config{Label: "Wrapper mode Pie", Slices: []Slice{{Name: "A", Value: 1}}, Controls: chartcontrol.Options{Mode: testCase.mode}}).Render(context.Background(), &output)
			if err != nil {
				t.Fatalf("Render() error = %v", err)
			}
			markup := output.String()
			if !strings.Contains(markup, `<figure class="goshtoso-charts-pie`) || !strings.Contains(markup, "<svg") {
				t.Fatal("wrapper mode lost chart DOM")
			}
			if testCase.mode == chartcontrol.WrapperModeOmitted {
				for _, forbidden := range []string{"data-goshtoso-chart-wrapper", chartassets.ControlRuntimeURL} {
					if strings.Contains(markup, forbidden) {
						t.Errorf("omitted wrapper retains %q", forbidden)
					}
				}
				return
			}
			if !strings.Contains(markup, testCase.want) || !strings.Contains(markup, chartassets.ControlRuntimeURL) {
				t.Errorf("wrapper mode markup missing %q or runtime", testCase.want)
			}
		})
	}
}

func TestPieDefaultSVGCompatibilityHash(t *testing.T) {
	t.Parallel()
	svg, err := renderSVG(Config{
		Label: "Deployments by status", Caption: "This week.",
		Slices: []Slice{{Name: "Successful", Value: 14}, {Name: "Failed", Value: 2}},
	})
	if err != nil {
		t.Fatalf("renderSVG() error = %v", err)
	}
	const preExtensionSHA256 = "46c6ae21c6c14963fcb3a5320f05257a803f983b7c8d0a3ddf122a9a0fcc7d1f"
	if got := fmtHash(svg); got != preExtensionSHA256 {
		t.Fatalf("default Pie SVG SHA-256 = %s, want pre-extension %s", got, preExtensionSHA256)
	}
}

func TestDoughnutPreservesPinnedExampleDataAndLayout(t *testing.T) {
	t.Parallel()
	cfg := upstreamDoughnutConfig()
	if upstreamDoughnutExample == "" {
		t.Fatal("upstream source path must remain recorded")
	}
	options := doughnutOptions(cfg)
	gotValues := make([]float64, len(options.SeriesList))
	for index, series := range options.SeriesList {
		gotValues[index] = series.Value
	}
	if want := []float64{1048, 735, 580, 484, 300}; !reflect.DeepEqual(gotValues, want) {
		t.Fatalf("doughnut values = %v, want %v", gotValues, want)
	}
	if want := []string{"Search Engine", "Direct", "Email", "Union Ads", "Video Ads"}; !reflect.DeepEqual(options.Legend.SeriesNames, want) {
		t.Fatalf("legend names = %v, want %v", options.Legend.SeriesNames, want)
	}
	if options.Title.Text != "Doughnut Chart" || options.Title.Subtext != "(Fake Data)" {
		t.Fatalf("title = %#v, want pinned title and subtitle", options.Title)
	}
	if options.Title.Offset.Left != "center" {
		t.Fatalf("title offset = %#v, want horizontally centered", options.Title.Offset)
	}
	if options.Title.FontStyle.FontSize != 16 || options.Title.SubtextFontStyle.FontSize != 10 {
		t.Fatalf("title sizes = %v/%v, want 16/10", options.Title.FontStyle.FontSize, options.Title.SubtextFontStyle.FontSize)
	}
	if options.Padding.Top != 20 || options.Padding.Right != 20 || options.Padding.Bottom != 20 || options.Padding.Left != 20 {
		t.Fatalf("padding = %#v, want 20 on every edge", options.Padding)
	}
	if options.Legend.Vertical == nil || !*options.Legend.Vertical || options.Legend.Offset.Left != "80%" || options.Legend.Offset.Top != "bottom" || options.Legend.FontStyle.FontSize != 10 {
		t.Fatalf("legend = %#v, want vertical at left 80%% and bottom with size 10", options.Legend)
	}
	if options.RadiusCenter != "24%" {
		t.Fatalf("inner radius = %q, want 24%% diameter (60%% of outer ring)", options.RadiusCenter)
	}
	if cfg.width() != 600 || cfg.height() != 400 {
		t.Fatalf("dimensions = %dx%d, want 600x400", cfg.width(), cfg.height())
	}
}

func TestDoughnutRendersAccessibleExactSummaryAndCallerOverrides(t *testing.T) {
	t.Parallel()
	cfg := upstreamDoughnutConfig()
	cfg.Style = charttheme.Style{Colors: []string{"#123456"}, Class: "caller-root"}
	cfg.RootAttrs = templ.Attributes{"data-chart": "doughnut"}
	cfg.Slices[0].Color = "#abcdef"
	cfg.Slices[0].Class = "search-slice"
	var output bytes.Buffer
	if err := Pie(cfg).Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	markup := output.String()
	for _, want := range []string{
		`class="goshtoso-charts-pie goshtoso-charts-palette goshtoso-charts-palette-auto caller-root"`,
		`data-chart="doughnut"`,
		`aria-label="Doughnut Chart exact slice values"`,
		`Search Engine`, `1048`, `33.30155703844932%`,
		`Direct`, `735`, `Email`, `580`, `Union Ads`, `484`, `Video Ads`, `300`,
		`fill:#abcdef`, `class="search-slice"`,
		`preserveAspectRatio="xMidYMid meet"`,
	} {
		if !strings.Contains(markup, want) {
			t.Errorf("rendered markup missing %q", want)
		}
	}
	if got := strings.Count(markup, `<tr class="border-t border-outline dark:border-outline-dark`); got != 5 {
		t.Errorf("exact slice row count = %d, want 5", got)
	}
}

func TestPieStrictValidation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{name: "variant", cfg: Config{Label: "Pie", Variant: "ring"}, want: "variant"},
		{name: "negative width", cfg: Config{Label: "Pie", Width: -1}, want: "width"},
		{name: "nonfinite radius", cfg: Config{Label: "Pie", Variant: VariantDoughnut, InnerRadiusPercent: math.NaN()}, want: "inner radius"},
		{name: "pie radius", cfg: Config{Label: "Pie", InnerRadiusPercent: 50}, want: "requires doughnut"},
		{name: "radius upper bound", cfg: Config{Label: "Pie", Variant: VariantDoughnut, InnerRadiusPercent: 100}, want: "below 100"},
		{name: "title placement", cfg: Config{Label: "Pie", Title: TitleOptions{Placement: "left"}}, want: "title placement"},
		{name: "legend orientation", cfg: Config{Label: "Pie", Legend: LegendOptions{Orientation: "diagonal"}}, want: "legend orientation"},
		{name: "legend left", cfg: Config{Label: "Pie", Legend: LegendOptions{LeftPercent: 101}}, want: "legend left"},
		{name: "outer radius", cfg: Config{Label: "Pie", Radius: RadiusOptions{OuterPixels: math.NaN()}}, want: "outer radius"},
		{name: "radius scale", cfg: Config{Label: "Pie", Radius: RadiusOptions{OuterPixels: 120, Scale: "diameter"}}, want: "radius scale"},
		{name: "area radius needs outer", cfg: Config{Label: "Pie", Radius: RadiusOptions{Scale: RadiusScaleArea}}, want: "requires an outer radius"},
		{name: "area radius doughnut", cfg: Config{Label: "Pie", Variant: VariantDoughnut, Radius: RadiusOptions{OuterPixels: 120, Scale: RadiusScaleArea}}, want: "requires pie variant"},
		{name: "label placement", cfg: Config{Label: "Pie", Labels: LabelOptions{Placement: "edge"}}, want: "label placement"},
		{name: "inside label pie", cfg: Config{Label: "Pie", Labels: LabelOptions{Placement: LabelPlacementInside}}, want: "require doughnut"},
		{name: "inside and hidden", cfg: Config{Label: "Pie", Variant: VariantDoughnut, Labels: LabelOptions{Hidden: true, Placement: LabelPlacementInside}}, want: "cannot be hidden"},
		{name: "center total pie", cfg: Config{Label: "Pie", Center: CenterOptions{Content: CenterContentTotal}}, want: "requires doughnut"},
		{name: "center format", cfg: Config{Label: "Pie", Variant: VariantDoughnut, Center: CenterOptions{Content: CenterContentTotal, Format: "currency"}}, want: "value format"},
		{name: "center formatting without total", cfg: Config{Label: "Pie", Variant: VariantDoughnut, Center: CenterOptions{Prefix: "Total: "}}, want: "requires center total"},
		{name: "inside and total", cfg: Config{Label: "Pie", Variant: VariantDoughnut, Labels: LabelOptions{Placement: LabelPlacementInside}, Center: CenterOptions{Content: CenterContentTotal}}, want: "cannot be combined"},
		{name: "segment gap", cfg: Config{Label: "Pie", SegmentGap: -1}, want: "segment gap"},
		{name: "padding", cfg: Config{Label: "Pie", Padding: Padding{Left: -1}}, want: "padding"},
		{name: "reserved root", cfg: Config{Label: "Pie", RootAttrs: templ.Attributes{"role": "presentation"}}, want: "reserved"},
		{name: "duplicate", cfg: Config{Label: "Pie", Slices: []Slice{{Name: "A", Value: 1}, {Name: "A", Value: 2}}}, want: "duplicated"},
		{name: "nonfinite value", cfg: Config{Label: "Pie", Slices: []Slice{{Name: "A", Value: math.Inf(1)}}}, want: "finite"},
	}
	for _, testCase := range tests {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if _, err := renderSVG(testCase.cfg); err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("renderSVG() error = %v, want containing %q", err, testCase.want)
			}
		})
	}
}

func TestPieRejectsNegativeSlice(t *testing.T) {
	t.Parallel()
	if _, err := renderSVG(Config{Label: "Deployments", Slices: []Slice{{Name: "Successful", Value: -1}}}); err == nil {
		t.Fatal("renderSVG() error = nil, want validation error")
	}
}

func TestPieRendersNoDataState(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer
	if err := Pie(Config{Label: "Deployment outcomes"}).Render(context.Background(), &output); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if !strings.Contains(output.String(), "No data in this period.") {
		t.Fatalf("rendered markup missing explicit no-data caption: %s", output.String())
	}
}

func upstreamDoughnutConfig() Config {
	return Config{
		Label: "Doughnut Chart", Variant: VariantDoughnut, InnerRadiusPercent: 60,
		Title: TitleOptions{
			Text: "Doughnut Chart", Subtitle: "(Fake Data)",
			Placement: PlacementCenter, FontSize: 16, SubtitleFontSize: 10,
		},
		Legend: LegendOptions{
			Orientation: LegendVertical, LeftPercent: 80,
			VerticalPlacement: VerticalPlacementBottom, FontSize: 10,
		},
		Padding: Padding{Top: 20, Right: 20, Bottom: 20, Left: 20},
		Slices: []Slice{
			{Name: "Search Engine", Value: 1048},
			{Name: "Direct", Value: 735},
			{Name: "Email", Value: 580},
			{Name: "Union Ads", Value: 484},
			{Name: "Video Ads", Value: 300},
		},
		Width: 600, Height: 400,
	}
}

func fmtHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return fmt.Sprintf("%x", sum)
}
