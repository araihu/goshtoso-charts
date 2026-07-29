package pages

import (
	"reflect"
	"testing"

	"github.com/araihu/goshtoso-charts/components/line"
)

func TestStaticLineBasicExamplePreservesGapAndFiveSeries(t *testing.T) {
	t.Parallel()
	cfg := sampleBasicLine()
	if cfg.Title.Text != "Line" || cfg.Title.FontSize != 16 || cfg.StrokeWidth != 1.2 || cfg.Symbol.Shape != line.SymbolCircle || cfg.Legend.Padding.Left != 100 || cfg.Width != 600 || cfg.Height != 400 {
		t.Fatalf("basic presentation drifted: %#v", cfg)
	}
	wantNames := []string{"Email", "Union Ads", "Video Ads", "Direct", "Search Engine"}
	if len(cfg.Series) != 5 {
		t.Fatalf("series count = %d", len(cfg.Series))
	}
	for index, name := range wantNames {
		if cfg.Series[index].Name != name {
			t.Errorf("series %d name = %q, want %q", index, cfg.Series[index].Name, name)
		}
	}
	if len(cfg.Series[0].Points) != 7 || !cfg.Series[0].Points[3].Missing || cfg.Series[0].Points[2].Value != 101 || cfg.Series[0].Points[4].Value != 90 {
		t.Fatalf("Email gap drifted: %#v", cfg.Series[0].Points)
	}
}

func TestStaticLineSymbolAndSmoothExamplesPreserveTreatments(t *testing.T) {
	t.Parallel()
	symbols := sampleSymbolLine()
	wantSymbols := []line.SymbolShape{line.SymbolCircle, line.SymbolDiamond, line.SymbolSquare, line.SymbolDot}
	if len(symbols.Series) != len(wantSymbols) {
		t.Fatalf("symbol series count = %d", len(symbols.Series))
	}
	for index := range wantSymbols {
		if symbols.Series[index].Symbol.Shape != wantSymbols[index] {
			t.Errorf("series %d symbol = %q", index, symbols.Series[index].Symbol.Shape)
		}
	}
	smooth := sampleSmoothLine()
	if !smooth.Legend.Hidden || smooth.Symbol.Shape != line.SymbolNone || smooth.StrokeWidth != 4 || smooth.SmoothingTension != .9 {
		t.Fatalf("smooth treatment drifted: %#v", smooth)
	}
}

func TestStaticLineMarkedAndStackedExamplesPreserveStatistics(t *testing.T) {
	t.Parallel()
	marked := sampleMarkedLine()
	for index, series := range marked.Series {
		if !series.References.Average || !series.References.Maximum || series.References.Format != line.ValueFormatHumanized || series.References.Decimals != 1 {
			t.Errorf("marked series %d references = %#v", index, series.References)
		}
	}
	stacked := sampleStackedLine()
	if !stacked.Stacked || stacked.YAxes[0].Title != "A+B+C Sum" || stacked.YAxes[0].TitleFontSize != 12 || stacked.YAxes[0].LabelFontSize != 8 || stacked.XAxis.BoundaryGap == nil || *stacked.XAxis.BoundaryGap {
		t.Fatalf("stacked treatment drifted: %#v", stacked)
	}
	if len(stacked.Series) != 3 || !reflect.DeepEqual(stacked.Series[0].Values, []float64{1.9, 23.2, 25.6, 102.6, 142.2, 32.6, 20, 2.3}) {
		t.Fatalf("stacked dataset drifted: %#v", stacked.Series)
	}
	if !stacked.Series[0].Labels.TrailingZeros {
		t.Fatal("stacked data labels must preserve one decimal place")
	}
}

func TestStaticLineBoundaryGapComparisonPreservesSharedDataset(t *testing.T) {
	t.Parallel()
	withGap := sampleBoundaryGapLine(true)
	withoutGap := sampleBoundaryGapLine(false)
	if withGap.XAxis.BoundaryGap == nil || !*withGap.XAxis.BoundaryGap || withoutGap.XAxis.BoundaryGap == nil || *withoutGap.XAxis.BoundaryGap {
		t.Fatalf("boundary-gap flags = %v / %v", withGap.XAxis.BoundaryGap, withoutGap.XAxis.BoundaryGap)
	}
	if withGap.Title.Text != "Boundary Gap" || withoutGap.Title.Text != "Boundary Gap Disabled" || !reflect.DeepEqual(withGap.Series, withoutGap.Series) {
		t.Fatalf("boundary-gap comparison drifted")
	}
}

func TestStaticLineCustomLensExamplePreservesDenseGapsAndAnnotations(t *testing.T) {
	t.Parallel()
	cfg := sampleCustomLensLine()
	if len(cfg.Labels) != 451 || cfg.Labels[0] != "60mm" || cfg.Labels[450] != "510mm" || len(cfg.Series) != 4 || len(cfg.Annotations) != 11 {
		t.Fatalf("custom shape = %d labels, %d series, %d annotations", len(cfg.Labels), len(cfg.Series), len(cfg.Annotations))
	}
	if !cfg.Series[0].Points[0].Missing || !cfg.Series[0].Points[39].Missing || cfg.Series[0].Points[40].Value != 4.5 || cfg.Series[0].Points[91].Value != 5 || cfg.Series[0].Points[194].Value != 5.6 || cfg.Series[0].Points[303].Value != 6.3 || cfg.Series[0].Points[412].Value != 7.1 || !cfg.Series[0].Points[441].Missing {
		t.Fatalf("100-500mm lens transition values drifted")
	}
	if cfg.XAxis.LabelCount != 10 || cfg.XAxis.LabelRotation != 45 || cfg.YAxes[0].Min == nil || *cfg.YAxes[0].Min != 1.4 || cfg.YAxes[0].Max == nil || *cfg.YAxes[0].Max != 8 || !cfg.YAxes[0].Hidden {
		t.Fatalf("custom axes drifted: %#v / %#v", cfg.XAxis, cfg.YAxes)
	}
}

func TestStaticLineGradientLabelsPreserveDatasetAndSemanticScale(t *testing.T) {
	t.Parallel()
	cfg := sampleGradientLabelLine()
	if cfg.Title.Subtext != "Cold = Low Values, Warm = High Values" || !cfg.Legend.Hidden || cfg.Width != 800 || cfg.Height != 500 {
		t.Fatalf("gradient presentation drifted: %#v", cfg)
	}
	want := []float64{20, 15, 35, 40, 10, 55, 25, 45, 30, 50}
	if len(cfg.Series) != 1 || !reflect.DeepEqual(cfg.Series[0].Values, want) || cfg.Series[0].Labels.ColorScale != line.LabelColorScaleColdToWarm {
		t.Fatalf("gradient dataset drifted: %#v", cfg.Series)
	}
}
