package pie

import (
	"reflect"
	"testing"

	chart "github.com/go-analyze/charts"
)

var upstreamPieValues = []float64{1048, 735, 580, 484, 300}
var upstreamPieNames = []string{"Search Engine", "Direct", "Email", "Union Ads", "Video Ads"}

func TestPieMapsPinnedBasicPresentation(t *testing.T) {
	t.Parallel()
	cfg := Config{
		Label: "Pie Chart", Title: TitleOptions{Text: "Pie Chart", Subtitle: "(Fake Data)", Placement: PlacementCenter, FontSize: 16, SubtitleFontSize: 10},
		Legend:  LegendOptions{Orientation: LegendVertical, LeftPercent: 80, VerticalPlacement: VerticalPlacementBottom, FontSize: 10},
		Padding: Padding{Top: 20, Right: 20, Bottom: 20, Left: 20}, Slices: upstreamSlices(), Width: 600, Height: 400,
	}
	options := pieOptions(cfg)
	if options.Title.Text != "Pie Chart" || options.Title.Subtext != "(Fake Data)" || options.Title.Offset.Left != "center" || options.Title.FontStyle.FontSize != 16 || options.Title.SubtextFontStyle.FontSize != 10 {
		t.Fatalf("title = %#v", options.Title)
	}
	if options.Padding.Top != 20 || options.Padding.Right != 20 || options.Padding.Bottom != 20 || options.Padding.Left != 20 || options.Legend.Vertical == nil || !*options.Legend.Vertical || options.Legend.Offset.Left != "80%" || options.Legend.Offset.Top != "bottom" || options.Legend.FontStyle.FontSize != 10 {
		t.Fatalf("padding/legend = %#v / %#v", options.Padding, options.Legend)
	}
	if got := pieSeriesValues(options.SeriesList); !reflect.DeepEqual(got, upstreamPieValues) {
		t.Fatalf("values = %v, want %v", got, upstreamPieValues)
	}
}

func TestPieMapsAreaScaledRadiiAndSegmentGap(t *testing.T) {
	t.Parallel()
	radius := pieOptions(Config{Label: "Area-scaled pie", Slices: upstreamSlices(), Radius: RadiusOptions{OuterPixels: 120, Scale: RadiusScaleArea}})
	if got, want := pieSeriesRadii(radius.SeriesList), []string{"120", "101", "90", "82", "65"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("area-scaled radii = %v, want %v", got, want)
	}
	gap := pieOptions(Config{Label: "Pie gap", Slices: upstreamSlices(), SegmentGap: 16, Legend: LegendOptions{Hidden: true}})
	if gap.SegmentGap != 16 || gap.Legend.Show == nil || *gap.Legend.Show {
		t.Fatalf("gap/legend = %g / %#v", gap.SegmentGap, gap.Legend)
	}
}

func TestDoughnutMapsPinnedStyleTreatments(t *testing.T) {
	t.Parallel()
	outside := doughnutOptions(Config{
		Label: "Labels Outside", Variant: VariantDoughnut, Slices: upstreamSlices(), SegmentGap: 24,
		Title: TitleOptions{Text: "Labels Outside", Placement: PlacementCenter}, Legend: LegendOptions{Hidden: true},
		Padding: Padding{Top: 10, Right: 10, Bottom: 15, Left: 10},
	})
	if outside.SegmentGap != 24 || outside.Legend.Show == nil || *outside.Legend.Show || outside.Padding.Bottom != 15 {
		t.Fatalf("outside treatment = %#v", outside)
	}

	inside := doughnutOptions(Config{
		Label: "Labels Inside", Variant: VariantDoughnut, Slices: upstreamSlices(), InnerRadiusPercent: 80,
		Labels: LabelOptions{Placement: LabelPlacementInside},
	})
	if inside.CenterValues != "labels" || inside.RadiusCenter != "40%" {
		t.Fatalf("inside treatment center/radius = %q / %q", inside.CenterValues, inside.RadiusCenter)
	}

	total := doughnutOptions(Config{
		Label: "Legend", Variant: VariantDoughnut, Slices: upstreamSlices(), InnerRadiusPercent: 80, SegmentGap: 8,
		Labels: LabelOptions{Hidden: true}, Center: CenterOptions{Content: CenterContentTotal, Prefix: "Total Response: ", Format: ValueFormatHumanized, Decimals: 2, FontSize: 12},
		Legend: LegendOptions{VerticalPlacement: VerticalPlacementBottom, Overlay: true},
	})
	if total.CenterValues != "sum" || total.RadiusCenter != "32%" || total.SegmentGap != 8 || total.CenterValuesFontStyle.FontSize != 12 || total.Legend.OverlayChart == nil || !*total.Legend.OverlayChart {
		t.Fatalf("total treatment = %#v", total)
	}
	if got := total.ValueFormatter(3147); got != "Total Response: 3.15k" {
		t.Fatalf("center formatter = %q", got)
	}
	for index, series := range total.SeriesList {
		if series.Label.Show == nil || *series.Label.Show {
			t.Errorf("series %d label visibility = %v, want hidden", index, series.Label.Show)
		}
	}
}

func upstreamSlices() []Slice {
	result := make([]Slice, len(upstreamPieValues))
	for index := range upstreamPieValues {
		result[index] = Slice{Name: upstreamPieNames[index], Value: upstreamPieValues[index]}
	}
	return result
}

func pieSeriesValues(series chart.PieSeriesList) []float64 {
	result := make([]float64, len(series))
	for index := range series {
		result[index] = series[index].Value
	}
	return result
}

func pieSeriesRadii(series chart.PieSeriesList) []string {
	result := make([]string, len(series))
	for index := range series {
		result[index] = series[index].Radius
	}
	return result
}
